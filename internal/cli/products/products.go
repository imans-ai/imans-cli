package products

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
	cmd := &cobra.Command{Use: "products", Short: "Read product catalog resources"}
	cmd.AddCommand(newListCommand(app))
	cmd.AddCommand(newGetCommand(app))
	return cmd
}

func newListCommand(app *cli.App) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List products",
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
			categoryID, _ := cmd.Flags().GetInt("category-id")
			brandID, _ := cmd.Flags().GetInt("brand-id")
			isVariable, _ := cmd.Flags().GetString("is-variable")
			common.SetIfNotEmpty(query, "search", search)
			common.SetIfNotEmpty(query, "status", status)
			common.SetIfNotEmpty(query, "is_variable", isVariable)
			if categoryID > 0 {
				query.Set("category_id", strconv.Itoa(categoryID))
			}
			if brandID > 0 {
				query.Set("brand_id", strconv.Itoa(brandID))
			}
			common.ApplyPagination(cmd, query)

			ctx := context.Background()
			if common.WantAll(cmd) {
				items, count, err := apiClient.ProductsAll(ctx, query)
				if err != nil {
					return err
				}
				if opts.JSON {
					return printer.PrintJSON(common.Combined(items, count))
				}
				return printProductsTable(app, printer, items)
			}

			page, err := apiClient.Products(ctx, query)
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(page)
			}
			return printProductsTable(app, printer, page.Results)
		},
	}
	cmd.Flags().String("search", "", "Search by product name or parent code")
	cmd.Flags().String("status", "", "Comma-separated product statuses")
	cmd.Flags().Int("category-id", 0, "Filter by category id")
	cmd.Flags().Int("brand-id", 0, "Filter by brand id")
	cmd.Flags().String("is-variable", "", "Filter by variable products: true or false")
	common.AddPaginationFlags(cmd)
	return cmd
}

func newGetCommand(app *cli.App) *cobra.Command {
	return &cobra.Command{
		Use:   "get <id>",
		Short: "Get one product with variants",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			opts := flags.OptionsFromCommand(cmd)
			printer := app.Printer(opts.JSON, opts.Quiet)
			apiClient, _, err := app.APIClient(opts.Profile, opts.Debug)
			if err != nil {
				return err
			}
			product, err := apiClient.Product(context.Background(), args[0])
			if err != nil {
				return err
			}
			if opts.JSON {
				return printer.PrintJSON(product)
			}
			items := []output.KeyValue{
				{Key: "id", Value: strconv.Itoa(product.ID)},
				{Key: "name", Value: product.Name},
				{Key: "parent_code", Value: product.ParentCode},
				{Key: "status", Value: product.Status},
				{Key: "category", Value: productCategoryName(product.Category)},
				{Key: "brand", Value: productBrandName(product.Brand)},
				{Key: "is_variable", Value: common.BoolString(product.IsVariable)},
				{Key: "variant_count", Value: strconv.Itoa(product.VariantCount)},
				{Key: "created_at", Value: product.CreatedAt},
				{Key: "updated_at", Value: product.UpdatedAt},
			}
			if err := printer.PrintKeyValues(items); err != nil {
				return err
			}
			if len(product.Variants) == 0 {
				return nil
			}
			_, _ = fmt.Fprintln(app.IO.Out)
			return printProductVariantsTable(printer, product.Variants)
		},
	}
}

func printProductsTable(app *cli.App, printer *output.Printer, items []generated.Product) error {
	if len(items) == 0 {
		_, err := fmt.Fprintln(app.IO.Out, common.NoResults("products"))
		return err
	}
	rows := make([][]string, 0, len(items))
	for _, item := range items {
		rows = append(rows, []string{
			strconv.Itoa(item.ID),
			item.Name,
			item.Status,
			productBrandName(item.Brand),
			productCategoryName(item.Category),
			strconv.Itoa(item.VariantCount),
		})
	}
	return printer.PrintTable([]string{"ID", "NAME", "STATUS", "BRAND", "CATEGORY", "VARIANTS"}, rows)
}

func printProductVariantsTable(printer *output.Printer, items []generated.ProductVariant) error {
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

func productCategoryName(category *generated.ProductCategory) string {
	if category == nil {
		return ""
	}
	return category.Name
}

func productBrandName(brand *generated.ProductBrand) string {
	if brand == nil {
		return ""
	}
	return brand.Name
}
