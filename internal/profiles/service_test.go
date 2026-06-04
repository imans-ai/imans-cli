package profiles

import (
	"errors"
	"testing"

	"github.com/imans-ai/imans-cli/internal/config"
	"github.com/imans-ai/imans-cli/internal/secrets"
)

// fakeSecrets is an in-memory secrets.Store for exercising removal ordering.
type fakeSecrets struct {
	data      map[string]string
	deleteErr error
}

func (f *fakeSecrets) Get(name string) (string, error) {
	v, ok := f.data[name]
	if !ok {
		return "", secrets.ErrNotFound
	}
	return v, nil
}

func (f *fakeSecrets) Set(name, value string) error {
	if f.data == nil {
		f.data = map[string]string{}
	}
	f.data[name] = value
	return nil
}

func (f *fakeSecrets) Delete(name string) error {
	if f.deleteErr != nil {
		return f.deleteErr
	}
	if _, ok := f.data[name]; !ok {
		return secrets.ErrNotFound
	}
	delete(f.data, name)
	return nil
}

func newSvc(t *testing.T, fs *fakeSecrets) *Service {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	cm, err := config.NewManager("imans")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return NewService(cm, fs)
}

func TestRemoveDeletesTokenAndProfile(t *testing.T) {
	fs := &fakeSecrets{data: map[string]string{}}
	svc := newSvc(t, fs)
	if err := svc.Save("acme", config.Profile{BaseURL: "https://api.imans.ai/"}, true); err != nil {
		t.Fatalf("Save: %v", err)
	}
	_ = fs.Set("acme", "tok")

	if err := svc.Remove("acme"); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, ok := fs.data["acme"]; ok {
		t.Fatal("token was not deleted")
	}
	if _, err := svc.Show("acme"); err == nil {
		t.Fatal("profile entry should be gone")
	}
}

func TestRemoveKeepsProfileWhenTokenDeleteFails(t *testing.T) {
	// Token deletion is attempted first; if it fails, the profile entry must
	// remain so the user can retry rather than ending up with an orphaned token.
	fs := &fakeSecrets{data: map[string]string{"acme": "tok"}, deleteErr: errors.New("keyring locked")}
	svc := newSvc(t, fs)
	if err := svc.Save("acme", config.Profile{}, true); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := svc.Remove("acme"); err == nil {
		t.Fatal("expected an error when the token delete fails")
	}
	if _, err := svc.Show("acme"); err != nil {
		t.Fatalf("profile should remain after a failed token delete: %v", err)
	}
	if _, ok := fs.data["acme"]; !ok {
		t.Fatal("token should be untouched after a failed delete")
	}
}

func TestRemoveIgnoresMissingToken(t *testing.T) {
	// A profile whose token is already gone should still be removable.
	fs := &fakeSecrets{data: map[string]string{}}
	svc := newSvc(t, fs)
	if err := svc.Save("acme", config.Profile{}, true); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := svc.Remove("acme"); err != nil {
		t.Fatalf("Remove with missing token should succeed: %v", err)
	}
	if _, err := svc.Show("acme"); err == nil {
		t.Fatal("profile entry should be gone")
	}
}
