package salesorderclassifications

import (
	"context"
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/api/generated"
	"github.com/imans-ai/imans-cli/internal/cli"
	"github.com/imans-ai/imans-cli/internal/cli/common"
	"github.com/imans-ai/imans-cli/internal/cli/flags"
	"github.com/imans-ai/imans-cli/internal/output"
)

func New(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{Use: "sales-order-classifications", Short: "Read sales order classifications"}
	cmd.AddCommand(newListCommand(app))
	cmd.AddCommand(newGetCommand(app))
	return cmd
}

func newListCommand(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sales order classifications",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			ctx := context.Background()
			if common.WantAll(cmd) {
				items, count, err := apiClient.SalesOrderClassificationsAll(ctx, nil)
				if err != nil {
					return err
				}
				if opts.JSON {
					return printer.PrintJSON(common.Combined(items, count))
				}
				return printTable(app, printer, items)
			}
			query := mapPagination(cmd)
			page, err := apiClient.SalesOrderClassifications(ctx, query)
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(page)
			}
			return printTable(app, printer, page.Results)
		},
	}
	common.AddPaginationFlags(cmd)
	return cmd
}

func newGetCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get one sales order classification",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			item, err := apiClient.SalesOrderClassification(context.Background(), args[0])
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(item)
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "id", Value: strconv.Itoa(item.ID)},
				{Key: "name", Value: item.Name},
				{Key: "parent", Value: common.PtrIntString(item.Parent)},
				{Key: "sales_channel", Value: common.PtrIntString(item.SalesChannel)},
			})
		},
	}
}

func printTable(app *cli.App, printer *output.Printer, items []generated.SalesOrderClassification) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(app.IO.Out, common.NoResults("sales order classifications"))
		return err
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			strconv.Itoa(item.ID),
			item.Name,
			common.PtrIntString(item.Parent),
			common.PtrIntString(item.SalesChannel),
		})
	}
	return printer.PrintTable([]string{"ID", "NAME", "PARENT", "SALES CHANNEL"}, rows)
}

func mapPagination(cmd *cobra.Command) url.Values {
	values := url.Values{}
	page, _ := cmd.Flags().GetInt(common.FlagPage)
	pageSize, _ := cmd.Flags().GetInt(common.FlagPageSize)
	if page > 0 {
		values.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		values.Set("page_size", strconv.Itoa(pageSize))
	}
	return values
}
