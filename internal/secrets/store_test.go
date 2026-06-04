package secrets

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptedFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.enc")
	store := newEncryptedFileStore(path, "imans")

	const token = "imns_live_secret_token_value"
	if err := store.Set("acme", token); err != nil {
		t.Fatalf("Set: %v", err)
	}

	got, err := store.Get("acme")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != token {
		t.Fatalf("Get returned %q, want %q", got, token)
	}

	// A second backend instance (fresh process simulation) must decrypt the
	// same file, since the key is derived from stable machine/user identity.
	reopened := newEncryptedFileStore(path, "imans")
	if got, err := reopened.Get("acme"); err != nil || got != token {
		t.Fatalf("reopened Get = %q, %v; want %q, nil", got, err, token)
	}

	if err := store.Delete("acme"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if _, err := store.Get("acme"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get after delete = %v, want ErrNotFound", err)
	}
}

func TestEncryptedFileStoreDoesNotPersistPlaintext(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.enc")
	store := newEncryptedFileStore(path, "imans")

	const token = "imns_live_secret_token_value"
	if err := store.Set("acme", token); err != nil {
		t.Fatalf("Set: %v", err)
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if bytes.Contains(raw, []byte(token)) {
		t.Fatalf("token stored as plaintext on disk")
	}
	if bytes.Contains(raw, []byte("acme")) {
		t.Fatalf("secret name stored as plaintext on disk")
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Fatalf("file mode = %o, want 600", perm)
	}
}

func TestEncryptedFileStoreUnreadableTreatedAsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "secrets.enc")
	if err := os.WriteFile(path, []byte("not-valid-ciphertext"), 0o600); err != nil {
		t.Fatalf("seed file: %v", err)
	}
	store := newEncryptedFileStore(path, "imans")

	// Corrupt/foreign contents should not error out; the user can re-login.
	if _, err := store.Get("acme"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("Get on corrupt file = %v, want ErrNotFound", err)
	}
	if err := store.Set("acme", "fresh"); err != nil {
		t.Fatalf("Set on corrupt file: %v", err)
	}
	if got, err := store.Get("acme"); err != nil || got != "fresh" {
		t.Fatalf("Get after recovery = %q, %v; want fresh, nil", got, err)
	}
}

func TestPlaintextFileStoreRoundTrip(t *testing.T) {
	dir := t.TempDir()
	store := newFileStore(filepath.Join(dir, "secrets.json"))
	if err := store.Set("acme", "tok"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, err := store.Get("acme"); err != nil || got != "tok" {
		t.Fatalf("Get = %q, %v; want tok, nil", got, err)
	}
}

func TestInsecureEnvForcesPlaintext(t *testing.T) {
	t.Setenv(insecureSecretsEnv, "1")
	store := NewStore("imans", t.TempDir())
	if _, ok := store.(*fileStore); !ok {
		t.Fatalf("with %s=1, store = %T, want *fileStore", insecureSecretsEnv, store)
	}
}
