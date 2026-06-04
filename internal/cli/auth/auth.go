package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/api/generated"
	"github.com/imans-ai/imans-cli/internal/apperrors"
	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/common"
	"github.com/imans-ai/imans-cli/internal/cli/flags"
	"github.com/imans-ai/imans-cli/internal/client"
	"github.com/imans-ai/imans-cli/internal/config"
	"github.com/imans-ai/imans-cli/internal/output"
	"github.com/imans-ai/imans-cli/internal/profiles"
)

const defaultBaseURL = "https://api.imans.ai/"

func New(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage API tokens and profile authentication"}
	cmd.AddCommand(newAddCommand(app))
	cmd.AddCommand(newTestCommand(app))
	cmd.AddCommand(newRemoveCommand(app))
	return cmd
}

// NewLogin is the friendly, flag-light entry point exposed at the top level as
// `imans login`. It prompts for a token, validates it, saves a profile, and
// makes it active so the user can run resource commands immediately.
func NewLogin(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in with an API token and start using the CLI",
		Long: "Log in with an Imans API token.\n\n" +
			"You will be prompted to paste your token (or pass it with --token-stdin,\n" +
			"--token-env, or IMANS_TOKEN). The token is validated against the workspace\n" +
			"endpoint, stored securely, and the resulting profile is made active.\n\n" +
			"Run login again with a different workspace token to add another workspace;\n" +
			"each workspace is saved as its own profile and the most recent login becomes\n" +
			"active. Switch between them with `imans profile use <name>`.",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			// An explicit --profile names the profile; otherwise it is derived
			// from the workspace so each workspace gets a stable alias.
			res, err := connectProfile(cmd, app, opts, printer, strings.TrimSpace(opts.Profile), true)
			if err != nil {
				return err
			}

			if opts.JSON {
				return printer.PrintJSON(map[string]any{
					"profile":   res.alias,
					"base_url":  res.baseURL,
					"workspace": res.workspace,
					"active":    res.entry.Active,
					"success":   true,
				})
			}
			printer.Successf("✓ Connected to %s (%s)", res.workspace.Name, res.workspace.WorkspaceCode)
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "active profile", Value: res.alias},
				{Key: "base url", Value: res.baseURL},
				{Key: "next", Value: "imans products list"},
			})
		},
	}
	addTokenFlags(cmd)
	return cmd
}

func newAddCommand(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a token-backed profile",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			alias := strings.TrimSpace(opts.Profile)
			if alias == "" {
				return apperrors.MissingFlag(flags.FlagProfile)
			}

			setActive, _ := cmd.Flags().GetBool("set-active")
			printer := app.Printer(opts.JSON, opts.Quiet)
			res, err := connectProfile(cmd, app, opts, printer, alias, setActive)
			if err != nil {
				return err
			}

			if opts.JSON {
				return printer.PrintJSON(map[string]any{
					"profile":   res.alias,
					"base_url":  res.baseURL,
					"workspace": res.workspace,
					"active":    res.entry.Active,
				})
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "profile", Value: res.alias},
				{Key: "base_url", Value: res.baseURL},
				{Key: "workspace_name", Value: res.workspace.Name},
				{Key: "workspace_code", Value: res.workspace.WorkspaceCode},
				{Key: "active", Value: common.BoolString(res.entry.Active)},
				{Key: "status", Value: res.workspace.Status},
			})
		},
	}
	addTokenFlags(cmd)
	cmd.Flags().Bool("set-active", false, "Set this profile as the active default")
	return cmd
}

// connectResult carries the outcome of a successful login/add for output.
type connectResult struct {
	alias     string
	baseURL   string
	workspace generated.Workspace
	entry     profiles.Entry
}

// connectProfile validates a token against the workspace endpoint, then stores
// the token and saves the profile. It is shared by `auth add` and `login`.
//
// When explicitAlias is empty the alias is derived from the validated workspace,
// so each workspace maps to a stable profile name and re-running login refreshes
// that workspace's token instead of creating a duplicate.
func connectProfile(cmd *cobra.Command, app *cli.App, opts flags.Options, printer *output.Printer, explicitAlias string, setActive bool) (connectResult, error) {
	baseURL, _ := cmd.Flags().GetString("base-url")
	baseURL = normalizeBaseURL(baseURL)

	token, err := readToken(cmd, app.IO.In, app.IO.Out)
	if err != nil {
		return connectResult{}, err
	}

	apiClient, err := client.New(client.Options{
		BaseURL:   baseURL,
		Token:     token,
		UserAgent: fmt.Sprintf("imans-cli/%s", app.Version.Version),
		Debug:     opts.Debug,
		ErrOut:    app.IO.ErrOut,
	})
	if err != nil {
		return connectResult{}, err
	}

	// Validate before persisting anything so a bad token leaves no trace.
	ctx := context.Background()
	workspace, err := apiClient.Workspace(ctx)
	if err != nil {
		return connectResult{}, err
	}

	alias := explicitAlias
	if alias == "" {
		existing, _ := app.Profiles.List()
		alias = deriveAlias(workspace, baseURL, existing)
	}

	// Remember any existing token for this alias so a failed save can roll back
	// to it instead of wiping a working profile on a re-login.
	prevToken, hadPrev := "", false
	if existing, gerr := app.Secrets.Get(alias); gerr == nil && existing != "" {
		prevToken, hadPrev = existing, true
	}

	if err := app.Secrets.Set(alias, token); err != nil {
		return connectResult{}, apperrors.WithDetails(
			apperrors.Wrap(apperrors.ExitGeneric, "failed to store token securely", err),
			"Secrets use the OS keychain when available, otherwise an encrypted local file.",
			"Check that the config directory is writable. Set IMANS_INSECURE_FILE_SECRETS=1 only for local development to force plaintext storage.",
		)
	}

	profile := config.Profile{
		BaseURL:       baseURL,
		WorkspaceCode: workspace.WorkspaceCode,
		WorkspaceName: workspace.Name,
		DefaultOutput: "text",
	}
	if err := app.Profiles.Save(alias, profile, setActive); err != nil {
		// Roll back the secret: restore the prior token if we overwrote one,
		// otherwise remove the token this command just created.
		if hadPrev {
			_ = app.Secrets.Set(alias, prevToken)
		} else {
			_ = app.Secrets.Delete(alias)
		}
		return connectResult{}, err
	}
	entry, err := app.Profiles.Show(alias)
	if err != nil {
		return connectResult{}, err
	}

	common.WarnOnVersionMismatch(ctx, app.Version.SchemaVersion, printer, apiClient)

	if duplicates, derr := app.Profiles.DuplicateWorkspaceAliases(baseURL, workspace.WorkspaceCode, alias); derr == nil && len(duplicates) > 0 {
		printer.Warnf("Warning: workspace %s is also saved as %s", workspace.Name, strings.Join(duplicates, ", "))
	}

	return connectResult{alias: alias, baseURL: baseURL, workspace: workspace, entry: entry}, nil
}

func addTokenFlags(cmd *cobra.Command) {
	cmd.Flags().String("base-url", defaultBaseURL, "Imans API base URL")
	cmd.Flags().String("token", "", "API token value")
	cmd.Flags().String("token-env", "", "Read the token from a named environment variable")
	cmd.Flags().Bool("token-stdin", false, "Read the token from stdin")
}

// deriveAlias builds a stable, friendly, filesystem-safe profile name for a
// workspace. It prefers a slug of the workspace name (e.g. "Acme Co" ->
// "acme-co") because workspace codes are opaque UUIDs that are painful to type
// in `imans profile use <name>`. The code is used only as a fallback.
//
// Naming is deterministic per workspace so re-logging into the same workspace
// refreshes the same profile instead of creating a duplicate:
//   - If a profile for this exact workspace already exists, its current alias
//     is reused (regardless of how it was originally named).
//   - Otherwise the name slug is used; if that slug is already taken by a
//     different workspace, a short workspace-code suffix disambiguates it.
func deriveAlias(ws generated.Workspace, baseURL string, existing []profiles.Entry) string {
	// 1. Reuse the existing alias for this exact workspace.
	for _, entry := range existing {
		if strings.EqualFold(entry.Profile.BaseURL, baseURL) && entry.Profile.WorkspaceCode == ws.WorkspaceCode {
			return entry.Name
		}
	}

	// 2. Build a friendly base from the name, falling back to the code.
	base := slugify(ws.Name)
	if base == "" {
		base = slugify(ws.WorkspaceCode)
	}
	if base == "" {
		base = "default"
	}

	taken := make(map[string]bool, len(existing))
	for _, entry := range existing {
		taken[entry.Name] = true
	}
	if !taken[base] {
		return base
	}

	// 3. Disambiguate a name collision with a short, stable code suffix.
	candidate := base
	if suffix := shortCode(ws.WorkspaceCode); suffix != "" {
		candidate = base + "-" + suffix
	}
	unique := candidate
	for n := 2; taken[unique]; n++ {
		unique = fmt.Sprintf("%s-%d", candidate, n)
	}
	return unique
}

// slugify lowercases input and keeps only [a-z0-9], turning separators into
// single dashes and trimming leading/trailing dashes.
func slugify(s string) string {
	var b strings.Builder
	lastDash := false
	for _, r := range strings.ToLower(strings.TrimSpace(s)) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		case r == ' ' || r == '_' || r == '-':
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// shortCode returns the leading alphanumeric run of a workspace code, capped at
// 8 characters — enough to disambiguate without dragging a full UUID into the
// alias.
func shortCode(code string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(code) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			if b.Len() >= 8 {
				break
			}
		} else if b.Len() > 0 {
			break
		}
	}
	return b.String()
}

func newTestCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "test",
		Short: "Validate a saved profile against the workspace endpoint",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, entry, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			ctx := context.Background()
			workspace, err := apiClient.Workspace(ctx)
			if err != nil {
				return err
			}
			common.WarnOnVersionMismatch(ctx, app.Version.SchemaVersion, printer, apiClient)
			if opts.JSON {
				return printer.PrintJSON(map[string]any{
					"profile":   entry.Name,
					"base_url":  entry.Profile.BaseURL,
					"workspace": workspace,
					"success":   true,
				})
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "profile", Value: entry.Name},
				{Key: "base_url", Value: entry.Profile.BaseURL},
				{Key: "workspace_name", Value: workspace.Name},
				{Key: "workspace_code", Value: workspace.WorkspaceCode},
				{Key: "success", Value: "true"},
			})
		},
	}
}

func newRemoveCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "remove [profile]",
		Short: "Remove a saved profile and its token",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			name := opts.Profile
			if len(args) == 1 {
				name = args[0]
			}
			if name == "" {
				return apperrors.New(apperrors.ExitUsage, "provide a profile name or use --profile")
			}
			if err := app.Profiles.Remove(name); err != nil {
				return err
			}
			printer := app.Printer(opts.JSON, opts.Quiet)
			if opts.JSON {
				return printer.PrintJSON(map[string]any{"removed": name})
			}
			return printer.PrintKeyValues([]output.KeyValue{{Key: "removed", Value: name}})
		},
	}
}

func normalizeBaseURL(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		trimmed = defaultBaseURL
	}
	if !strings.HasSuffix(trimmed, "/") {
		trimmed += "/"
	}
	return trimmed
}

func readToken(cmd *cobra.Command, in io.Reader, out io.Writer) (string, error) {
	if token, _ := cmd.Flags().GetString("token"); strings.TrimSpace(token) != "" {
		return strings.TrimSpace(token), nil
	}
	if envName, _ := cmd.Flags().GetString("token-env"); strings.TrimSpace(envName) != "" {
		value := strings.TrimSpace(os.Getenv(envName))
		if value == "" {
			return "", fmt.Errorf("environment variable %s is empty", envName)
		}
		return value, nil
	}
	if implicit := strings.TrimSpace(os.Getenv("IMANS_TOKEN")); implicit != "" {
		return implicit, nil
	}
	if fromStdin, _ := cmd.Flags().GetBool("token-stdin"); fromStdin {
		data, err := io.ReadAll(in)
		if err != nil {
			return "", err
		}
		value := strings.TrimSpace(string(data))
		if value == "" {
			return "", fmt.Errorf("stdin did not contain a token")
		}
		return value, nil
	}

	file, ok := in.(*os.File)
	if !ok || !term.IsTerminal(int(file.Fd())) {
		return "", fmt.Errorf("token required; use --token, --token-env, --token-stdin, or IMANS_TOKEN")
	}
	_, _ = fmt.Fprint(out, "Token: ")
	data, err := term.ReadPassword(int(file.Fd()))
	_, _ = fmt.Fprintln(out)
	if err != nil {
		return "", err
	}
	value := strings.TrimSpace(string(data))
	if value == "" {
		return "", fmt.Errorf("token cannot be empty")
	}
	return value, nil
}
