package productvariants

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
	cmd := &cobra.Command{Use: "product-variants", Short: "List product variants"}
	cmd.AddCommand(newListCommand(app))
	return cmd
}

func newListCommand(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List product variants",
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			query := url.Values{}
			search, _ := cmd.Flags().GetString("search")
			status, _ := cmd.Flags().GetString("status")
			productID, _ := cmd.Flags().GetInt("product-id")
			isBundle, _ := cmd.Flags().GetString("is-bundle")
			common.SetIfNotEmpty(query, "search", search)
			common.SetIfNotEmpty(query, "status", status)
			common.SetIfNotEmpty(query, "is_bundle", isBundle)
			if productID > 0 {
				query.Set("product_id", strconv.Itoa(productID))
			}
			common.ApplyPagination(cmd, query)

			ctx := context.Background()
			if common.WantAll(cmd) {
				items, count, err := apiClient.ProductVariantsAll(ctx, query)
				if err != nil {
					return err
				}
				if opts.JSON {
					return printer.PrintJSON(common.Combined(items, count))
				}
				return printTable(app, printer, items)
			}

			page, err := apiClient.ProductVariants(ctx, query)
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(page)
			}
			return printTable(app, printer, page.Results)
		},
	}
	cmd.Flags().String("search", "", "Search by SKU, GTIN, or name")
	cmd.Flags().Int("product-id", 0, "Filter by owning product id")
	cmd.Flags().String("status", "", "Comma-separated statuses")
	cmd.Flags().String("is-bundle", "", "Filter by bundle variants: true or false")
	common.AddPaginationFlags(cmd)
	return cmd
}

func printTable(app *cli.App, printer *output.Printer, items []generated.ProductVariant) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(app.IO.Out, common.NoResults("product variants"))
		return err
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			strconv.Itoa(item.ID),
			item.SKU,
			item.Name,
			item.Status,
			item.CurrentPrice,
			common.BoolString(item.IsBundle),
		})
	}
	return printer.PrintTable([]string{"ID", "SKU", "NAME", "STATUS", "PRICE", "BUNDLE"}, rows)
}
