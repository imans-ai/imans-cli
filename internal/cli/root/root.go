package root

import (
	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/auth"
	"github.com/imans-ai/imans-cli/internal/cli/flags"
	"github.com/imans-ai/imans-cli/internal/cli/products"
	"github.com/imans-ai/imans-cli/internal/cli/productvariants"
	"github.com/imans-ai/imans-cli/internal/cli/profile"
	"github.com/imans-ai/imans-cli/internal/cli/salesorderclassifications"
	"github.com/imans-ai/imans-cli/internal/cli/salesorderitems"
	"github.com/imans-ai/imans-cli/internal/cli/salesorders"
	"github.com/imans-ai/imans-cli/internal/cli/workspace"
	"github.com/imans-ai/imans-cli/internal/output"
)

func New(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "imans",
		Short:         "Imans public API CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	flags.AddPersistentFlags(cmd)

	cmd.AddCommand(newVersionCommand(app))
	cmd.AddCommand(auth.New(app))
	cmd.AddCommand(profile.New(app))
	cmd.AddCommand(workspace.New(app))
	cmd.AddCommand(products.New(app))
	cmd.AddCommand(productvariants.New(app))
	cmd.AddCommand(salesorders.New(app))
	cmd.AddCommand(salesorderitems.New(app))
	cmd.AddCommand(salesorderclassifications.New(app))
	return cmd
}

func newVersionCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Show CLI version metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			if opts.JSON {
				return printer.PrintJSON(app.Version)
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "version", Value: app.Version.Version},
				{Key: "commit", Value: app.Version.Commit},
				{Key: "build_date", Value: app.Version.BuildDate},
				{Key: "schema_version", Value: app.Version.SchemaVersion},
				{Key: "goos", Value: app.Version.GOOS},
				{Key: "goarch", Value: app.Version.GOARCH},
			})
		},
	}
}
