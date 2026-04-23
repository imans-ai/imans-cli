package workspace

import (
	"context"

	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/common"
	"github.com/imans-ai/imans-cli/internal/cli/flags"
	"github.com/imans-ai/imans-cli/internal/output"
)

func New(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{Use: "workspace", Short: "Read workspace metadata"}
	cmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Retrieve the active profile's workspace",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
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
				return printer.PrintJSON(workspace)
			}
			items := []output.KeyValue{
				{Key: "workspace_code", Value: workspace.WorkspaceCode},
				{Key: "name", Value: workspace.Name},
				{Key: "description", Value: workspace.Description},
				{Key: "status", Value: workspace.Status},
				{Key: "created_at", Value: workspace.CreatedAt},
			}
			if workspace.Settings != nil {
				items = append(items,
					output.KeyValue{Key: "settings.timezone", Value: workspace.Settings.Timezone},
					output.KeyValue{Key: "settings.br_auto_classify_accountability_by_cfop", Value: map[bool]string{true: "true", false: "false"}[workspace.Settings.BRAutoClassifyAccountabilityByCFOP]},
					output.KeyValue{Key: "settings.enable_profit_calculation", Value: map[bool]string{true: "true", false: "false"}[workspace.Settings.EnableProfitCalculation]},
					output.KeyValue{Key: "settings.profit_calculation_mode", Value: workspace.Settings.ProfitCalculationMode},
				)
			}
			return printer.PrintKeyValues(items)
		},
	})
	return cmd
}
