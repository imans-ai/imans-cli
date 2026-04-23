package apperrors

import (
	"testing"

	"github.com/imans-ai/imans-cli/internal/client"
)

func TestExitCodeFromAPIStatus(t *testing.T) {
	tests := []struct {
		status int
		want   int
	}{
		{status: 401, want: ExitAuth},
		{status: 403, want: ExitScope},
		{status: 404, want: ExitNotFound},
		{status: 503, want: ExitServer},
	}
	for _, tc := range tests {
		got := ExitCode(&client.APIError{Status: tc.status, Detail: "boom"})
		if got != tc.want {
			t.Fatalf("status %d => %d, want %d", tc.status, got, tc.want)
		}
	}
}
