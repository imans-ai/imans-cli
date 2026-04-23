package generated

// These models are intentionally hand-maintained until schema generation is
// wired into the repository. Keep the package boundary stable so generated
// code can replace this file later.

type PaginatedResponse[T any] struct {
	Count    int    `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []T    `json:"results"`
}

type WorkspaceSettings struct {
	Timezone                           string `json:"timezone"`
	BRAutoClassifyAccountabilityByCFOP bool   `json:"br_auto_classify_accountability_by_cfop"`
	EnableProfitCalculation            bool   `json:"enable_profit_calculation"`
	ProfitCalculationMode              string `json:"profit_calculation_mode"`
}

type Workspace struct {
	WorkspaceCode string             `json:"workspace_code"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	Status        string             `json:"status"`
	CreatedAt     string             `json:"created_at"`
	Settings      *WorkspaceSettings `json:"settings"`
}

type ProductCategory struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	ParentID int    `json:"parent_id"`
}

type ProductBrand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductVariant struct {
	ID           int    `json:"id"`
	SKU          string `json:"sku"`
	Name         string `json:"name"`
	ImageURL     string `json:"image_url"`
	GTIN         string `json:"gtin"`
	GrossWeight  string `json:"gross_weight"`
	NetWeight    string `json:"net_weight"`
	Length       string `json:"length"`
	Width        string `json:"width"`
	Height       string `json:"height"`
	Status       string `json:"status"`
	IsBundle     bool   `json:"is_bundle"`
	CurrentCost  string `json:"current_cost"`
	CurrentPrice string `json:"current_price"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

type Product struct {
	ID           int              `json:"id"`
	Name         string           `json:"name"`
	ImageURL     string           `json:"image_url"`
	Category     *ProductCategory `json:"category"`
	Brand        *ProductBrand    `json:"brand"`
	ParentCode   string           `json:"parent_code"`
	IsVariable   bool             `json:"is_variable"`
	Status       string           `json:"status"`
	GrossWeight  string           `json:"gross_weight"`
	NetWeight    string           `json:"net_weight"`
	Length       string           `json:"length"`
	Width        string           `json:"width"`
	Height       string           `json:"height"`
	BROrigin     string           `json:"br_origin"`
	BRNCM        string           `json:"br_ncm"`
	BRCEST       string           `json:"br_cest"`
	VariantCount int              `json:"variant_count"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
}

type ProductDetail struct {
	Product
	Variants []ProductVariant `json:"variants"`
}

type EmbeddedClassification struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type EmbeddedSalesAgent struct {
	ID int `json:"id"`
}

type SalesOrder struct {
	ID                     int                     `json:"id"`
	OrderStatus            string                  `json:"order_status"`
	OrderClassification    *EmbeddedClassification `json:"order_classification"`
	IsAccountable          bool                    `json:"is_accountable"`
	CustomerID             int                     `json:"customer_id"`
	SalesAgent             *EmbeddedSalesAgent     `json:"sales_agent"`
	OrderNumber            string                  `json:"order_number"`
	InvoiceNumber          string                  `json:"invoice_number"`
	OrderDate              string                  `json:"order_date"`
	ExpectedDeliveryDate   string                  `json:"expected_delivery_date"`
	InvoicedDate           string                  `json:"invoiced_date"`
	ShippedDate            string                  `json:"shipped_date"`
	DeliveryDate           string                  `json:"delivery_date"`
	CancelledDate          string                  `json:"cancelled_date"`
	ReturnedDate           string                  `json:"returned_date"`
	TotalAmount            string                  `json:"total_amount"`
	ProductTotalAmount     string                  `json:"product_total_amount"`
	ShippingCostOn         string                  `json:"shipping_cost_on"`
	ShippingCostOff        string                  `json:"shipping_cost_off"`
	DiscountOn             string                  `json:"discount_on"`
	DiscountOff            string                  `json:"discount_off"`
	OtherExpensesOn        string                  `json:"other_expenses_on"`
	OtherExpensesOff       string                  `json:"other_expenses_off"`
	ProductCostTotalAmount string                  `json:"product_cost_total_amount"`
	TaxTotalAmount         string                  `json:"tax_total_amount"`
	CreatedAt              string                  `json:"created_at"`
	UpdatedAt              string                  `json:"updated_at"`
}

type SalesOrderItem struct {
	ID                     int    `json:"id"`
	OrderID                int    `json:"order_id"`
	ProductID              int    `json:"product_id"`
	Quantity               string `json:"quantity"`
	UnitPrice              string `json:"unit_price"`
	TotalAmount            string `json:"total_amount"`
	RawProductCode         string `json:"raw_product_code"`
	RawProductName         string `json:"raw_product_name"`
	ShippingCostOn         string `json:"shipping_cost_on"`
	ShippingCostOff        string `json:"shipping_cost_off"`
	DiscountOn             string `json:"discount_on"`
	DiscountOff            string `json:"discount_off"`
	OtherExpensesOn        string `json:"other_expenses_on"`
	OtherExpensesOff       string `json:"other_expenses_off"`
	ProductCostTotalAmount string `json:"product_cost_total_amount"`
	TaxTotalAmount         string `json:"tax_total_amount"`
	ProfitTotalAmount      string `json:"profit_total_amount"`
}

type SalesOrderClassification struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Parent       *int   `json:"parent"`
	SalesChannel *int   `json:"sales_channel"`
}
