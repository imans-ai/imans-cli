package client

import (
	"net/http"
	"strings"
	"testing"

	"github.com/imans-ai/imans-cli/internal/apperrors"
)

func TestParseAPIError5xxIsActionable(t *testing.T) {
	header := http.Header{}
	header.Set("X-Cloud-Trace-Context", "0123456789abcdef0123456789abcdef/12345;o=1")

	// A 500 commonly returns an HTML page with no JSON detail.
	apiErr := parseAPIError(500, []byte("<!doctype html><h1>Server Error (500)</h1>"), header)

	if apiErr.Status != 500 {
		t.Fatalf("Status = %d, want 500", apiErr.Status)
	}
	if apiErr.TraceID != "0123456789abcdef0123456789abcdef" {
		t.Fatalf("TraceID = %q, want the bare trace id", apiErr.TraceID)
	}
	if strings.Contains(apiErr.Error(), "<") {
		t.Fatalf("Error() leaked HTML: %q", apiErr.Error())
	}

	// The formatted output users see must explain it's server-side and carry
	// the trace ID, and the exit code must map to ExitServer.
	formatted := apperrors.Format(apiErr)
	if !strings.Contains(formatted, "server error (HTTP 500)") {
		t.Fatalf("missing server-error message:\n%s", formatted)
	}
	if !strings.Contains(formatted, "not your request") {
		t.Fatalf("missing reassurance line:\n%s", formatted)
	}
	if !strings.Contains(formatted, "Trace ID: 0123456789abcdef0123456789abcdef") {
		t.Fatalf("missing trace id:\n%s", formatted)
	}
	if code := apperrors.ExitCode(apiErr); code != apperrors.ExitServer {
		t.Fatalf("ExitCode = %d, want ExitServer(%d)", code, apperrors.ExitServer)
	}
}

func TestParseAPIErrorPrefersJSONDetail(t *testing.T) {
	// A 4xx with a JSON detail keeps the server-provided message.
	apiErr := parseAPIError(403, []byte(`{"detail":"Invalid token"}`), nil)
	if apiErr.Error() != "Invalid token" {
		t.Fatalf("Error() = %q, want \"Invalid token\"", apiErr.Error())
	}
	if apiErr.TraceID != "" {
		t.Fatalf("TraceID = %q, want empty with nil header", apiErr.TraceID)
	}
}

func TestTraceIDFromHeader(t *testing.T) {
	cases := []struct {
		name string
		set  map[string]string
		want string
	}{
		{"cloud trace with span", map[string]string{"X-Cloud-Trace-Context": "abc123/999;o=1"}, "abc123"},
		{"cloud trace bare", map[string]string{"X-Cloud-Trace-Context": "abc123"}, "abc123"},
		{"cf-ray fallback", map[string]string{"CF-Ray": "abcd1234ef567890-XYZ"}, "abcd1234ef567890-XYZ"},
		{"cloud trace wins over cf-ray", map[string]string{"X-Cloud-Trace-Context": "abc123;o=1", "CF-Ray": "ray"}, "abc123"},
		{"none", map[string]string{}, ""},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := http.Header{}
			for k, v := range tc.set {
				h.Set(k, v)
			}
			if got := traceIDFromHeader(h); got != tc.want {
				t.Fatalf("traceIDFromHeader = %q, want %q", got, tc.want)
			}
		})
	}
}
