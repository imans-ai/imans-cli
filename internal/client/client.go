package client

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/imans-ai/imans-cli/api/generated"
	"github.com/imans-ai/imans-cli/internal/transport"
)

type Options struct {
	BaseURL   string
	Token     string
	UserAgent string
	Debug     bool
	ErrOut    io.Writer
}

type Client struct {
	baseURL   *url.URL
	token     string
	userAgent string
	debug     bool
	errOut    io.Writer
	http      *http.Client
	random    *rand.Rand
}

type APIError struct {
	Status  int
	Code    string
	Detail  string
	Details []string
}

func (e *APIError) Error() string {
	if e.Detail != "" {
		return e.Detail
	}
	return fmt.Sprintf("api request failed with status %d", e.Status)
}

func (e *APIError) HTTPStatusCode() int {
	return e.Status
}

func (e *APIError) ErrorDetail() string {
	return e.Error()
}

func (e *APIError) ErrorDetails() []string {
	return e.Details
}

func New(opts Options) (*Client, error) {
	baseURL, err := url.Parse(opts.BaseURL)
	if err != nil {
		return nil, err
	}
	if opts.UserAgent == "" {
		opts.UserAgent = "imans-cli/dev"
	}
	return &Client{
		baseURL:   baseURL,
		token:     opts.Token,
		userAgent: opts.UserAgent,
		debug:     opts.Debug,
		errOut:    opts.ErrOut,
		http:      transport.NewHTTPClient(transport.Options{}),
		random:    rand.New(rand.NewSource(time.Now().UnixNano())),
	}, nil
}

func (c *Client) Workspace(ctx context.Context) (generated.Workspace, error) {
	var out generated.Workspace
	return out, c.getJSON(ctx, "v1/workspace/", nil, &out)
}

func (c *Client) Products(ctx context.Context, query url.Values) (generated.PaginatedResponse[generated.Product], error) {
	return getPage[generated.Product](ctx, c, "v1/products/", query)
}

func (c *Client) ProductsAll(ctx context.Context, query url.Values) ([]generated.Product, int, error) {
	return getAll[generated.Product](ctx, c, "v1/products/", query)
}

func (c *Client) Product(ctx context.Context, id string) (generated.ProductDetail, error) {
	var out generated.ProductDetail
	return out, c.getJSON(ctx, fmt.Sprintf("v1/products/%s/", id), nil, &out)
}

func (c *Client) ProductVariants(ctx context.Context, query url.Values) (generated.PaginatedResponse[generated.ProductVariant], error) {
	return getPage[generated.ProductVariant](ctx, c, "v1/product-variants/", query)
}

func (c *Client) ProductVariantsAll(ctx context.Context, query url.Values) ([]generated.ProductVariant, int, error) {
	return getAll[generated.ProductVariant](ctx, c, "v1/product-variants/", query)
}

func (c *Client) SalesOrders(ctx context.Context, query url.Values) (generated.PaginatedResponse[generated.SalesOrder], error) {
	return getPage[generated.SalesOrder](ctx, c, "v1/sales-orders/", query)
}

func (c *Client) SalesOrdersAll(ctx context.Context, query url.Values) ([]generated.SalesOrder, int, error) {
	return getAll[generated.SalesOrder](ctx, c, "v1/sales-orders/", query)
}

func (c *Client) SalesOrder(ctx context.Context, id string) (generated.SalesOrder, error) {
	var out generated.SalesOrder
	return out, c.getJSON(ctx, fmt.Sprintf("v1/sales-orders/%s/", id), nil, &out)
}

func (c *Client) SalesOrderItems(ctx context.Context, query url.Values) (generated.PaginatedResponse[generated.SalesOrderItem], error) {
	return getPage[generated.SalesOrderItem](ctx, c, "v1/sales-order-items/", query)
}

func (c *Client) SalesOrderItemsAll(ctx context.Context, query url.Values) ([]generated.SalesOrderItem, int, error) {
	return getAll[generated.SalesOrderItem](ctx, c, "v1/sales-order-items/", query)
}

func (c *Client) SalesOrderClassifications(ctx context.Context, query url.Values) (generated.PaginatedResponse[generated.SalesOrderClassification], error) {
	return getPage[generated.SalesOrderClassification](ctx, c, "v1/sales-order-classifications/", query)
}

func (c *Client) SalesOrderClassificationsAll(ctx context.Context, query url.Values) ([]generated.SalesOrderClassification, int, error) {
	return getAll[generated.SalesOrderClassification](ctx, c, "v1/sales-order-classifications/", query)
}

func (c *Client) SalesOrderClassification(ctx context.Context, id string) (generated.SalesOrderClassification, error) {
	var out generated.SalesOrderClassification
	return out, c.getJSON(ctx, fmt.Sprintf("v1/sales-order-classifications/%s/", id), nil, &out)
}

func (c *Client) ContractVersion(ctx context.Context) (string, error) {
	var doc struct {
		Info struct {
			Version string `json:"version"`
		} `json:"info"`
	}
	if err := c.getJSON(ctx, "documentation/v1/schema/", nil, &doc); err != nil {
		return "", err
	}
	return doc.Info.Version, nil
}

func getPage[T any](ctx context.Context, c *Client, path string, query url.Values) (generated.PaginatedResponse[T], error) {
	var out generated.PaginatedResponse[T]
	return out, c.getJSON(ctx, path, query, &out)
}

func getAll[T any](ctx context.Context, c *Client, path string, query url.Values) ([]T, int, error) {
	items := []T{}
	nextPath := path
	nextQuery := cloneValues(query)
	total := 0
	for {
		page, err := getPage[T](ctx, c, nextPath, nextQuery)
		if err != nil {
			return nil, 0, err
		}
		if total == 0 {
			total = page.Count
		}
		items = append(items, page.Results...)
		if page.Next == "" {
			break
		}
		resolved, err := c.resolvePath(page.Next)
		if err != nil {
			return nil, 0, err
		}
		nextPath = resolved.Path
		nextQuery = resolved.Query()
	}
	return items, total, nil
}

func (c *Client) getJSON(ctx context.Context, path string, query url.Values, out any) error {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		status, body, err := c.do(ctx, http.MethodGet, path, query)
		if err != nil {
			lastErr = err
			if !shouldRetryError(err) || attempt == 2 {
				return err
			}
			c.sleep(attempt)
			continue
		}
		if status >= 200 && status < 300 {
			if err := json.Unmarshal(body, out); err != nil {
				return fmt.Errorf("decode response: %w", err)
			}
			return nil
		}

		apiErr := parseAPIError(status, body)
		lastErr = apiErr
		if !shouldRetryStatus(status) || attempt == 2 {
			return apiErr
		}
		c.sleep(attempt)
	}
	return lastErr
}

func (c *Client) do(ctx context.Context, method, path string, query url.Values) (int, []byte, error) {
	requestURL, err := c.resolvePath(path)
	if err != nil {
		return 0, nil, err
	}
	if len(query) > 0 {
		requestURL.RawQuery = query.Encode()
	}

	req, err := http.NewRequestWithContext(ctx, method, requestURL.String(), nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Authorization", "Token "+c.token)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", c.userAgent)

	start := time.Now()
	resp, err := c.http.Do(req)
	if err != nil {
		c.debugLog(method, requestURL.String(), 0, time.Since(start))
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, err
	}
	c.debugLog(method, requestURL.String(), resp.StatusCode, time.Since(start))
	return resp.StatusCode, body, nil
}

func (c *Client) resolvePath(raw string) (*url.URL, error) {
	parsed, err := url.Parse(raw)
	if err != nil {
		return nil, err
	}
	if parsed.IsAbs() {
		return parsed, nil
	}
	return c.baseURL.ResolveReference(parsed), nil
}

func (c *Client) sleep(attempt int) {
	delay := time.Duration(200*(1<<attempt)+c.random.Intn(150)) * time.Millisecond
	time.Sleep(delay)
}

func (c *Client) debugLog(method, requestURL string, status int, duration time.Duration) {
	if !c.debug || c.errOut == nil {
		return
	}
	_, _ = fmt.Fprintf(c.errOut, "DEBUG %s %s status=%d latency=%s\n", method, requestURL, status, duration.Round(time.Millisecond))
}

func parseAPIError(status int, body []byte) *APIError {
	errOut := &APIError{Status: status, Detail: fmt.Sprintf("request failed with status %d", status)}
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return errOut
	}
	if detail, ok := payload["detail"].(string); ok && strings.TrimSpace(detail) != "" {
		errOut.Detail = detail
	}
	if code, ok := payload["error"].(string); ok {
		errOut.Code = code
	}
	if missing, ok := payload["missing_scopes"].([]any); ok {
		scopes := make([]string, 0, len(missing))
		for _, item := range missing {
			if value, ok := item.(string); ok {
				scopes = append(scopes, value)
			}
		}
		if len(scopes) > 0 {
			errOut.Details = append(errOut.Details, "Missing scopes: "+strings.Join(scopes, ", "))
		}
	}
	return errOut
}

func shouldRetryStatus(status int) bool {
	return status == 502 || status == 503 || status == 504
}

func shouldRetryError(err error) bool {
	var netErr net.Error
	if stderrors.As(err, &netErr) {
		return true
	}
	return false
}

func cloneValues(in url.Values) url.Values {
	if in == nil {
		return url.Values{}
	}
	out := url.Values{}
	for key, values := range in {
		for _, value := range values {
			out.Add(key, value)
		}
	}
	return out
}
