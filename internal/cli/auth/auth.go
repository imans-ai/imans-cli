package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"

	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/internal/apperrors"
	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/common"
	"github.com/imans-ai/imans-cli/internal/cli/flags"
	"github.com/imans-ai/imans-cli/internal/client"
	"github.com/imans-ai/imans-cli/internal/config"
	"github.com/imans-ai/imans-cli/internal/output"
)

const defaultBaseURL = "https://api.imans.ai/"

func New(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{Use: "auth", Short: "Manage API tokens and profile authentication"}
	cmd.AddCommand(newAddCommand(app))
	cmd.AddCommand(newTestCommand(app))
	cmd.AddCommand(newRemoveCommand(app))
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

			baseURL, _ := cmd.Flags().GetString("base-url")
			baseURL = normalizeBaseURL(baseURL)

			token, err := readToken(cmd, app.IO.In, app.IO.Out)
			if err != nil {
				return err
			}

			if err := app.Secrets.Set(alias, token); err != nil {
				return apperrors.Wrap(apperrors.ExitGeneric, "failed to store token securely", err)
			}

			apiClient, err := client.New(client.Options{
				BaseURL:   baseURL,
				Token:     token,
				UserAgent: fmt.Sprintf("imans-cli/%s", app.Version.Version),
				Debug:     opts.Debug,
				ErrOut:    app.IO.ErrOut,
			})
			if err != nil {
				_ = app.Secrets.Delete(alias)
				return err
			}

			ctx := context.Background()
			workspace, err := apiClient.Workspace(ctx)
			if err != nil {
				_ = app.Secrets.Delete(alias)
				return err
			}

			setActive, _ := cmd.Flags().GetBool("set-active")
			profile := config.Profile{
				BaseURL:       baseURL,
				WorkspaceCode: workspace.WorkspaceCode,
				WorkspaceName: workspace.Name,
				DefaultOutput: "text",
			}
			if err := app.Profiles.Save(alias, profile, setActive); err != nil {
				_ = app.Secrets.Delete(alias)
				return err
			}
			entry, err := app.Profiles.Show(alias)
			if err != nil {
				return err
			}

			printer := app.Printer(opts.JSON, opts.Quiet)
			common.WarnOnVersionMismatch(ctx, app.Version.SchemaVersion, printer, apiClient)

			duplicates, err := app.Profiles.DuplicateWorkspaceAliases(baseURL, workspace.WorkspaceCode, alias)
			if err == nil && len(duplicates) > 0 {
				printer.Warnf("Warning: workspace %s is also saved as %s", workspace.Name, strings.Join(duplicates, ", "))
			}

			if opts.JSON {
				return printer.PrintJSON(map[string]any{
					"profile":   alias,
					"base_url":  baseURL,
					"workspace": workspace,
					"active":    entry.Active,
				})
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "profile", Value: alias},
				{Key: "base_url", Value: baseURL},
				{Key: "workspace_name", Value: workspace.Name},
				{Key: "workspace_code", Value: workspace.WorkspaceCode},
				{Key: "active", Value: common.BoolString(entry.Active)},
				{Key: "status", Value: workspace.Status},
			})
		},
	}
	cmd.Flags().String("base-url", defaultBaseURL, "Imans API base URL")
	cmd.Flags().String("token", "", "API token value")
	cmd.Flags().String("token-env", "", "Read the token from a named environment variable")
	cmd.Flags().Bool("token-stdin", false, "Read the token from stdin")
	cmd.Flags().Bool("set-active", false, "Set this profile as the active default")
	return cmd
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
