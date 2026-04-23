package secrets

import (
	"encoding/json"
	stderrors "errors"
	"fmt"
	"os"
	"path/filepath"

	keyring "github.com/zalando/go-keyring"
)

const insecureSecretsEnv = "IMANS_INSECURE_FILE_SECRETS"

var ErrNotFound = stderrors.New("secret not found")

type Store interface {
	Get(name string) (string, error)
	Set(name string, value string) error
	Delete(name string) error
}

func NewStore(appName, configDir string) Store {
	if os.Getenv(insecureSecretsEnv) == "1" {
		return &fileStore{path: filepath.Join(configDir, "secrets.json")}
	}
	return &keyringStore{service: appName + "-cli"}
}

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

type fileStore struct {
	path string
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
	data, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return map[string]string{}, nil
		}
		return nil, err
	}

	var out map[string]string
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
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
	encoded, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, encoded, 0o600)
}
