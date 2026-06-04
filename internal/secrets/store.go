package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	keyring "github.com/zalando/go-keyring"
)

const insecureSecretsEnv = "IMANS_INSECURE_FILE_SECRETS"

var ErrNotFound = stderrors.New("secret not found")

type Store interface {
	Get(name string) (string, error)
	Set(name string, value string) error
	Delete(name string) error
}

// NewStore selects a secret backend.
//
// Resolution order:
//   - IMANS_INSECURE_FILE_SECRETS=1 forces a plaintext file backend (development only).
//   - macOS and Windows always use the OS keychain/credential backend.
//   - Linux uses the Secret Service keyring when one is available and unlocked,
//     and otherwise transparently falls back to an encrypted on-disk file so the
//     CLI still works on headless servers, containers, and WSL.
//
// Backend selection is lazy: it happens on the first Get/Set/Delete so commands
// that never touch secrets (version, profile list, ...) pay no probe cost.
func NewStore(appName, configDir string) Store {
	service := appName + "-cli"
	if os.Getenv(insecureSecretsEnv) == "1" {
		return newFileStore(filepath.Join(configDir, "secrets.json"))
	}
	return &lazyStore{
		resolve: func() Store {
			if runtime.GOOS == "linux" && !keyringAvailable(service) {
				return newEncryptedFileStore(filepath.Join(configDir, "secrets.enc"), appName)
			}
			return &keyringStore{service: service}
		},
	}
}

// keyringAvailable reports whether a working Secret Service keyring is reachable
// by performing a throwaway write/delete probe.
func keyringAvailable(service string) bool {
	const probeKey = "__imans_keyring_probe__"
	if err := keyring.Set(service, probeKey, "1"); err != nil {
		return false
	}
	_ = keyring.Delete(service, probeKey)
	return true
}

// lazyStore resolves its concrete backend once, on first use.
type lazyStore struct {
	resolve func() Store
	once    sync.Once
	backend Store
}

func (s *lazyStore) get() Store {
	s.once.Do(func() { s.backend = s.resolve() })
	return s.backend
}

func (s *lazyStore) Get(name string) (string, error) { return s.get().Get(name) }
func (s *lazyStore) Set(name, value string) error    { return s.get().Set(name, value) }
func (s *lazyStore) Delete(name string) error        { return s.get().Delete(name) }

type keyringStore struct {
	service string
}

func (s *keyringStore) Get(name string) (string, error) {
	value, err := keyring.Get(s.service, name)
	if err != nil {
		if stderrors.Is(err, keyring.ErrNotFound) {
			return "", ErrNotFound
		}
		return "", fmt.Errorf("secure keyring unavailable: %w", err)
	}
	return value, nil
}

func (s *keyringStore) Set(name string, value string) error {
	if err := keyring.Set(s.service, name, value); err != nil {
		return fmt.Errorf("secure keyring unavailable: %w", err)
	}
	return nil
}

func (s *keyringStore) Delete(name string) error {
	if err := keyring.Delete(s.service, name); err != nil {
		if stderrors.Is(err, keyring.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("secure keyring unavailable: %w", err)
	}
	return nil
}

// fileStore persists secrets as a plaintext JSON map. Used only when the user
// explicitly opts in via IMANS_INSECURE_FILE_SECRETS=1.
type fileStore struct {
	path  string
	codec codec
}

// codec converts the secret map to and from the bytes written on disk.
type codec interface {
	encode(map[string]string) ([]byte, error)
	decode([]byte) (map[string]string, error)
}

func newFileStore(path string) *fileStore {
	return &fileStore{path: path, codec: plaintextCodec{}}
}

func newEncryptedFileStore(path, appName string) *fileStore {
	return &fileStore{path: path, codec: aesCodec{key: deriveKey(appName)}}
}

func (s *fileStore) Get(name string) (string, error) {
	data, err := s.load()
	if err != nil {
		return "", err
	}
	value, ok := data[name]
	if !ok {
		return "", ErrNotFound
	}
	return value, nil
}

func (s *fileStore) Set(name string, value string) error {
	data, err := s.load()
	if err != nil {
		return err
	}
	data[name] = value
	return s.save(data)
}

func (s *fileStore) Delete(name string) error {
	data, err := s.load()
	if err != nil {
		return err
	}
	if _, ok := data[name]; !ok {
		return ErrNotFound
	}
	delete(data, name)
	return s.save(data)
}

func (s *fileStore) load() (map[string]string, error) {
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}
	out, err := s.codec.decode(raw)
	if err != nil {
		// Unreadable contents (e.g. the machine identity changed) are treated as
		// an empty store so the user can simply log in again to overwrite it.
		return map[string]string{}, nil
	}
	if out == nil {
		out = map[string]string{}
	}
	return out, nil
}

func (s *fileStore) save(data map[string]string) error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0o755); err != nil {
		return err
	}
	encoded, err := s.codec.encode(data)
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, encoded, 0o600)
}

type plaintextCodec struct{}

func (plaintextCodec) encode(data map[string]string) ([]byte, error) {
	return json.MarshalIndent(data, "", "  ")
}

func (plaintextCodec) decode(raw []byte) (map[string]string, error) {
	var out map[string]string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// aesCodec encrypts the secret map with AES-256-GCM. The key is derived from a
// stable per-machine, per-user identity, so the file is not plaintext and is not
// portable to other machines or users. It is not a substitute for an OS keyring:
// anything able to run as this user on this machine can recompute the key.
type aesCodec struct {
	key [32]byte
}

func (c aesCodec) encode(data map[string]string) ([]byte, error) {
	plaintext, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	sealed := gcm.Seal(nonce, nonce, plaintext, nil)
	encoded := make([]byte, base64.StdEncoding.EncodedLen(len(sealed)))
	base64.StdEncoding.Encode(encoded, sealed)
	return encoded, nil
}

func (c aesCodec) decode(raw []byte) (map[string]string, error) {
	sealed := make([]byte, base64.StdEncoding.DecodedLen(len(raw)))
	n, err := base64.StdEncoding.Decode(sealed, raw)
	if err != nil {
		return nil, err
	}
	sealed = sealed[:n]
	block, err := aes.NewCipher(c.key[:])
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	if len(sealed) < gcm.NonceSize() {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce, body := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, body, nil)
	if err != nil {
		return nil, err
	}
	var out map[string]string
	if err := json.Unmarshal(plaintext, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// deriveKey produces a 32-byte key bound to the current machine and user.
func deriveKey(appName string) [32]byte {
	seed := machineID()
	if u, err := user.Current(); err == nil {
		seed += ":" + u.Uid
	}
	return sha256.Sum256([]byte("imans-cli-secret-key:" + appName + ":" + seed))
}

func machineID() string {
	for _, path := range []string{"/etc/machine-id", "/var/lib/dbus/machine-id"} {
		if raw, err := os.ReadFile(path); err == nil {
			if id := strings.TrimSpace(string(raw)); id != "" {
				return id
			}
		}
	}
	if host, err := os.Hostname(); err == nil && host != "" {
		return host
	}
	return "imans-fallback-identity"
}
