package common

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"github.com/imans-ai/imans-cli/internal/output"
)

const (
	FlagPage     = "page"
	FlagPageSize = "page-size"
	FlagAll      = "all"
)

type AllPayload[T any] struct {
	Items      []T `json:"items"`
	Pagination struct {
		Count   int `json:"count"`
		Fetched int `json:"fetched"`
	} `json:"pagination"`
}

func Combined[T any](items []T, count int) AllPayload[T] {
	out := AllPayload[T]{Items: items}
	out.Pagination.Count = count
	out.Pagination.Fetched = len(items)
	return out
}

func AddPaginationFlags(cmd *cobra.Command) {
	cmd.Flags().Int(FlagPage, 0, "Fetch a specific page")
	cmd.Flags().Int(FlagPageSize, 0, "Set page size (API max 200)")
	cmd.Flags().Bool(FlagAll, false, "Fetch all pages")
}

func ApplyPagination(cmd *cobra.Command, values url.Values) {
	page, _ := cmd.Flags().GetInt(FlagPage)
	pageSize, _ := cmd.Flags().GetInt(FlagPageSize)
	if page > 0 {
		values.Set("page", strconv.Itoa(page))
	}
	if pageSize > 0 {
		values.Set("page_size", strconv.Itoa(pageSize))
	}
}

func WantAll(cmd *cobra.Command) bool {
	all, _ := cmd.Flags().GetBool(FlagAll)
	return all
}

func SetIfNotEmpty(values url.Values, key, value string) {
	trimmed := strings.TrimSpace(value)
	if trimmed != "" {
		values.Set(key, trimmed)
	}
}

func BoolString(value bool) string {
	if value {
		return "true"
	}
	return "false"
}

func PtrIntString(value *int) string {
	if value == nil {
		return ""
	}
	return strconv.Itoa(*value)
}

func NoResults(resource string) string {
	return fmt.Sprintf("No %s found.", resource)
}

func WarnOnVersionMismatch(ctx context.Context, schemaVersion string, printer *output.Printer, apiClient interface {
	ContractVersion(context.Context) (string, error)
}) {
	serverVersion, err := apiClient.ContractVersion(ctx)
	if err != nil || serverVersion == "" || schemaVersion == "" || schemaVersion == "dev" {
		return
	}
	if serverVersion != schemaVersion {
		printer.Warnf("Warning: server API version %s differs from CLI schema version %s", serverVersion, schemaVersion)
	}
}
