package salesorders

import (
	"context"
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
	cmd := &cobra.Command{Use: "sales-orders", Short: "Read sales order data"}
	cmd.AddCommand(newListCommand(app))
	cmd.AddCommand(newGetCommand(app))
	return cmd
}

func newListCommand(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List sales orders",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			query := url.Values{}
			common.SetIfNotEmpty(query, "order_date_from", mustGetString(cmd, "order-date-from"))
			common.SetIfNotEmpty(query, "order_date_to", mustGetString(cmd, "order-date-to"))
			common.SetIfNotEmpty(query, "order_status", mustGetString(cmd, "order-status"))
			common.SetIfNotEmpty(query, "classification_id", intFlagString(cmd, "classification-id"))
			common.SetIfNotEmpty(query, "customer_id", intFlagString(cmd, "customer-id"))
			common.SetIfNotEmpty(query, "sales_agent_id", intFlagString(cmd, "sales-agent-id"))
			common.SetIfNotEmpty(query, "search", mustGetString(cmd, "search"))
			common.ApplyPagination(cmd, query)

			ctx := context.Background()
			if common.WantAll(cmd) {
				items, count, err := apiClient.SalesOrdersAll(ctx, query)
				if err != nil {
					return err
				}
				if opts.JSON {
					return printer.PrintJSON(common.Combined(items, count))
				}
				return printSalesOrdersTable(printer, items)
			}

			page, err := apiClient.SalesOrders(ctx, query)
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(page)
			}
			return printSalesOrdersTable(printer, page.Results)
		},
	}
	cmd.Flags().String("order-date-from", "", "Filter by order_date >= YYYY-MM-DD")
	cmd.Flags().String("order-date-to", "", "Filter by order_date <= YYYY-MM-DD")
	cmd.Flags().String("order-status", "", "Comma-separated order statuses")
	cmd.Flags().Int("classification-id", 0, "Filter by classification id")
	cmd.Flags().Int("customer-id", 0, "Filter by customer id")
	cmd.Flags().Int("sales-agent-id", 0, "Filter by sales agent id")
	cmd.Flags().String("search", "", "Search by order number or invoice number")
	common.AddPaginationFlags(cmd)
	return cmd
}

func newGetCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get one sales order",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			item, err := apiClient.SalesOrder(context.Background(), args[0])
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(item)
			}
			return printer.PrintKeyValues([]output.KeyValue{
				{Key: "id", Value: strconv.Itoa(item.ID)},
				{Key: "order_number", Value: item.OrderNumber},
				{Key: "invoice_number", Value: item.InvoiceNumber},
				{Key: "status", Value: item.OrderStatus},
				{Key: "classification", Value: classificationName(item.OrderClassification)},
				{Key: "customer_id", Value: strconv.Itoa(item.CustomerID)},
				{Key: "sales_agent_id", Value: salesAgentID(item.SalesAgent)},
				{Key: "order_date", Value: item.OrderDate},
				{Key: "expected_delivery_date", Value: item.ExpectedDeliveryDate},
				{Key: "total_amount", Value: item.TotalAmount},
				{Key: "product_total_amount", Value: item.ProductTotalAmount},
				{Key: "is_accountable", Value: common.BoolString(item.IsAccountable)},
				{Key: "created_at", Value: item.CreatedAt},
				{Key: "updated_at", Value: item.UpdatedAt},
			})
		},
	}
}

func printSalesOrdersTable(printer *output.Printer, items []generated.SalesOrder) error {
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			strconv.Itoa(item.ID),
			item.OrderNumber,
			item.InvoiceNumber,
			item.OrderStatus,
			item.OrderDate,
			item.TotalAmount,
			classificationName(item.OrderClassification),
		})
	}
	return printer.PrintTable([]string{"ID", "ORDER", "INVOICE", "STATUS", "ORDER DATE", "TOTAL", "CLASSIFICATION"}, rows)
}

func classificationName(item *generated.EmbeddedClassification) string {
	if item == nil {
		return ""
	}
	return item.Name
}

func salesAgentID(item *generated.EmbeddedSalesAgent) string {
	if item == nil {
		return ""
	}
	return strconv.Itoa(item.ID)
}

func mustGetString(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}

func intFlagString(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetInt(name)
	if value <= 0 {
		return ""
	}
	return strconv.Itoa(value)
}
