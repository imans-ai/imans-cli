package auth

import (
	"testing"

	"github.com/imans-ai/imans-cli/api/generated"
)

func TestDeriveAlias(t *testing.T) {
	cases := []struct {
		name string
		ws   generated.Workspace
		want string
	}{
		{"workspace code", generated.Workspace{WorkspaceCode: "ACME-01"}, "acme-01"},
		{"spaces and underscores", generated.Workspace{WorkspaceCode: "Acme Co_NA"}, "acme-co-na"},
		{"falls back to name", generated.Workspace{Name: "Acme Corp"}, "acme-corp"},
		{"drops non-ascii", generated.Workspace{WorkspaceCode: "Açme"}, "ame"},
		{"empty falls back to default", generated.Workspace{}, "default"},
		{"trims dashes", generated.Workspace{WorkspaceCode: "  -acme-  "}, "acme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveAlias(tc.ws); got != tc.want {
				t.Fatalf("deriveAlias(%+v) = %q, want %q", tc.ws, got, tc.want)
			}
		})
	}
}
