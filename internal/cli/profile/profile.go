package profile

import (
	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/flags"
	"github.com/imans-ai/imans-cli/internal/output"
)

func New(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{Use: "profile", Short: "Manage saved workspace profiles"}
	cmd.AddCommand(newListCommand(app))
	cmd.AddCommand(newShowCommand(app))
	cmd.AddCommand(newUseCommand(app))
	return cmd
}

func newListCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List saved profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			entries, err := app.Profiles.List()
			if err != nil {
				return err
			}
			printer := app.Printer(opts.JSON, opts.Quiet)
			if opts.JSON {
				return printer.PrintJSON(entries)
			}
			rows := make([][]string, 0, len(entries))
			for _, entry := range entries {
				active := ""
				if entry.Active {
					active = "*"
				}
				rows = append(rows, []string{
					active,
					entry.Name,
					entry.Profile.WorkspaceCode,
					entry.Profile.WorkspaceName,
					entry.Profile.BaseURL,
				})
			}
			return printer.PrintTable([]string{"ACTIVE", "PROFILE", "WORKSPACE CODE", "WORKSPACE NAME", "BASE URL"}, rows)
		},
	}
}

func newShowCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "show [profile]",
		Short: "Show one profile or the active profile",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			name := opts.Profile
			if len(args) == 1 {
				name = args[0]
			}
			entry, err := app.Profiles.Show(name)
			if err != nil {
				return err
			}
			printer := app.Printer(opts.JSON, opts.Quiet)
			if opts.JSON {
				return printer.PrintJSON(entry)
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "profile", Value: entry.Name},
				{Key: "active", Value: map[bool]string{true: "true", false: "false"}[entry.Active]},
				{Key: "base_url", Value: entry.Profile.BaseURL},
				{Key: "workspace_code", Value: entry.Profile.WorkspaceCode},
				{Key: "workspace_name", Value: entry.Profile.WorkspaceName},
				{Key: "default_output", Value: entry.Profile.DefaultOutput},
			})
		},
	}
}

func newUseCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "use <profile>",
		Short: "Set the active profile",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			if err := app.Profiles.Use(args[0]); err != nil {
				return err
			}
			printer := app.Printer(opts.JSON, opts.Quiet)
			if opts.JSON {
				return printer.PrintJSON(map[string]any{"active_profile": args[0]})
			}
			return printer.PrintKeyValues([]output.KeyValue{{Key: "active_profile", Value: args[0]}})
		},
	}
}
