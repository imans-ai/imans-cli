package salesorderitems

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
	cmd := &cobra.Command{Use: "sales-order-items", Short: "List sales order items"}
	cmd.AddCommand(newListCommand(app))
	return cmd
}

func newListCommand(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sales order items",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			query := url.Values{}
			common.SetIfNotEmpty(query, "order_id", intFlagString(cmd, "order-id"))
			common.SetIfNotEmpty(query, "product_id", intFlagString(cmd, "product-id"))
			common.ApplyPagination(cmd, query)

			ctx := context.Background()
			if common.WantAll(cmd) {
				items, count, err := apiClient.SalesOrderItemsAll(ctx, query)
				if err != nil {
					return err
				}
				if opts.JSON {
					return printer.PrintJSON(common.Combined(items, count))
				}
				return printTable(app, printer, items)
			}
			page, err := apiClient.SalesOrderItems(ctx, query)
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(page)
			}
			return printTable(app, printer, page.Results)
		},
	}
	cmd.Flags().Int("order-id", 0, "Filter by order id")
	cmd.Flags().Int("product-id", 0, "Filter by product variant id")
	common.AddPaginationFlags(cmd)
	return cmd
}

func printTable(app *cli.App, printer *output.Printer, items []generated.SalesOrderItem) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(app.IO.Out, common.NoResults("sales order items"))
		return err
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			strconv.Itoa(item.ID),
			strconv.Itoa(item.OrderID),
			strconv.Itoa(item.ProductID),
			item.Quantity,
			item.TotalAmount,
			item.ProfitTotalAmount,
		})
	}
	return printer.PrintTable([]string{"ID", "ORDER", "PRODUCT", "QTY", "TOTAL", "PROFIT"}, rows)
}

func intFlagString(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetInt(name)
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}
