package auth

import (
	"testing"

	"github.com/imans-ai/imans-cli/api/generated"
	"github.com/imans-ai/imans-cli/internal/config"
	"github.com/imans-ai/imans-cli/internal/profiles"
)

const baseURL = "https://api.example.test/"

func entry(name, code string) profiles.Entry {
	return profiles.Entry{
		Name:    name,
		Profile: config.Profile{BaseURL: baseURL, WorkspaceCode: code},
	}
}

func TestDeriveAliasPrefersNameSlug(t *testing.T) {
	cases := []struct {
		name string
		ws   generated.Workspace
		want string
	}{
		{"name slug wins over code", generated.Workspace{Name: "Acme Co", WorkspaceCode: "11111111-2222"}, "acme-co"},
		{"collapses separators", generated.Workspace{Name: "Acme   Co_NA"}, "acme-co-na"},
		{"drops non-ascii", generated.Workspace{Name: "Açme"}, "ame"},
		{"falls back to code when no name", generated.Workspace{WorkspaceCode: "SHORT-01"}, "short-01"},
		{"empty falls back to default", generated.Workspace{}, "default"},
		{"trims dashes", generated.Workspace{Name: "  -acme-  "}, "acme"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := deriveAlias(tc.ws, baseURL, nil); got != tc.want {
				t.Fatalf("deriveAlias(%+v) = %q, want %q", tc.ws, got, tc.want)
			}
		})
	}
}

func TestDeriveAliasReusesExistingWorkspaceProfile(t *testing.T) {
	existing := []profiles.Entry{entry("custom-name", "ws-code-1")}
	ws := generated.Workspace{Name: "Acme Co", WorkspaceCode: "ws-code-1"}

	// Same workspace already saved (under any alias) -> reuse that alias so a
	// re-login refreshes it instead of creating a duplicate.
	if got := deriveAlias(ws, baseURL, existing); got != "custom-name" {
		t.Fatalf("deriveAlias = %q, want existing alias %q", got, "custom-name")
	}
}

func TestDeriveAliasDisambiguatesNameCollision(t *testing.T) {
	// A different workspace already owns the "acme" slug.
	existing := []profiles.Entry{entry("acme", "11111111-aaaa")}
	ws := generated.Workspace{Name: "Acme", WorkspaceCode: "22222222-bbbb"}

	got := deriveAlias(ws, baseURL, existing)
	if got != "acme-22222222" {
		t.Fatalf("deriveAlias = %q, want disambiguated %q", got, "acme-22222222")
	}

	// And it stays deterministic: re-deriving with that profile now present
	// reuses the same alias (exact-workspace match), not a third name.
	existing = append(existing, entry(got, ws.WorkspaceCode))
	if again := deriveAlias(ws, baseURL, existing); again != got {
		t.Fatalf("re-derive = %q, want stable %q", again, got)
	}
}

func TestDeriveAliasDifferentBaseURLIsNotReused(t *testing.T) {
	// Same workspace code on a different base URL is a different workspace, so
	// the alias must not be reused — it gets disambiguated instead.
	existing := []profiles.Entry{{
		Name:    "acme",
		Profile: config.Profile{BaseURL: "https://other.example.test/", WorkspaceCode: "wscode1"},
	}}
	ws := generated.Workspace{Name: "Acme", WorkspaceCode: "wscode1"}

	got := deriveAlias(ws, baseURL, existing)
	if got == "acme" {
		t.Fatalf("deriveAlias reused alias across different base URLs: %q", got)
	}
	if got != "acme-wscode1" {
		t.Fatalf("deriveAlias = %q, want disambiguated %q", got, "acme-wscode1")
	}
}
