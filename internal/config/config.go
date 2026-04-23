package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	BaseURL       string `yaml:"base_url" json:"base_url"`
	WorkspaceCode string `yaml:"workspace_code,omitempty" json:"workspace_code,omitempty"`
	WorkspaceName string `yaml:"workspace_name,omitempty" json:"workspace_name,omitempty"`
	DefaultOutput string `yaml:"default_output,omitempty" json:"default_output,omitempty"`
}

type File struct {
	ActiveProfile string             `yaml:"active_profile,omitempty" json:"active_profile,omitempty"`
	Profiles      map[string]Profile `yaml:"profiles" json:"profiles"`
}

type Manager struct {
	appName    string
	configDir  string
	configPath string
}

func NewManager(appName string) (*Manager, error) {
	baseDir, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}

	configDir := filepath.Join(baseDir, appName)
	return &Manager{
		appName:    appName,
		configDir:  configDir,
		configPath: filepath.Join(configDir, "config.yaml"),
	}, nil
}

func (m *Manager) ConfigDir() string {
	return m.configDir
}

func (m *Manager) ConfigPath() string {
	return m.configPath
}

func (m *Manager) Load() (*File, error) {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			return defaultFile(), nil
		}
		return nil, err
	}

	var cfg File
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}
	return &cfg, nil
}

func (m *Manager) Save(cfg *File) error {
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]Profile{}
	}

	if err := os.MkdirAll(m.configDir, 0o755); err != nil {
		return err
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0o600)
}

func defaultFile() *File {
	return &File{Profiles: map[string]Profile{}}
}
