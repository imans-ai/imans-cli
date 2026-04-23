package profiles

import (
	"fmt"
	"sort"
	"strings"

	"github.com/imans-ai/imans-cli/internal/config"
	"github.com/imans-ai/imans-cli/internal/secrets"
)

type Entry struct {
	Name    string         `json:"name"`
	Active  bool           `json:"active"`
	Profile config.Profile `json:"profile"`
}

type Service struct {
	config  *config.Manager
	secrets secrets.Store
}

func NewService(config *config.Manager, secrets secrets.Store) *Service {
	return &Service{config: config, secrets: secrets}
}

func (s *Service) List() ([]Entry, error) {
	cfg, err := s.config.Load()
	if err != nil {
		return nil, err
	}

	names := make([]string, 0, len(cfg.Profiles))
	for name := range cfg.Profiles {
		names = append(names, name)
	}
	sort.Strings(names)

	entries := make([]Entry, 0, len(names))
	for _, name := range names {
		entries = append(entries, Entry{
			Name:    name,
			Active:  name == cfg.ActiveProfile,
			Profile: cfg.Profiles[name],
		})
	}
	return entries, nil
}

func (s *Service) Show(name string) (Entry, error) {
	cfg, err := s.config.Load()
	if err != nil {
		return Entry{}, err
	}
	resolved, err := resolveName(cfg, name)
	if err != nil {
		return Entry{}, err
	}
	profile := cfg.Profiles[resolved]
	return Entry{Name: resolved, Active: resolved == cfg.ActiveProfile, Profile: profile}, nil
}

func (s *Service) Use(name string) error {
	cfg, err := s.config.Load()
	if err != nil {
		return err
	}
	resolved, err := resolveName(cfg, name)
	if err != nil {
		return err
	}
	cfg.ActiveProfile = resolved
	return s.config.Save(cfg)
}

func (s *Service) Save(name string, profile config.Profile, setActive bool) error {
	cfg, err := s.config.Load()
	if err != nil {
		return err
	}
	if cfg.Profiles == nil {
		cfg.Profiles = map[string]config.Profile{}
	}
	cfg.Profiles[name] = profile
	if setActive || cfg.ActiveProfile == "" {
		cfg.ActiveProfile = name
	}
	return s.config.Save(cfg)
}

func (s *Service) Remove(name string) error {
	cfg, err := s.config.Load()
	if err != nil {
		return err
	}
	resolved, err := resolveName(cfg, name)
	if err != nil {
		return err
	}
	delete(cfg.Profiles, resolved)
	if cfg.ActiveProfile == resolved {
		cfg.ActiveProfile = ""
	}
	if err := s.config.Save(cfg); err != nil {
		return err
	}
	if err := s.secrets.Delete(resolved); err != nil && err != secrets.ErrNotFound {
		return err
	}
	return nil
}

func (s *Service) Resolve(name string) (Entry, error) {
	return s.Show(name)
}

func (s *Service) DuplicateWorkspaceAliases(baseURL, workspaceCode, exclude string) ([]string, error) {
	cfg, err := s.config.Load()
	if err != nil {
		return nil, err
	}
	out := []string{}
	for name, profile := range cfg.Profiles {
		if name == exclude {
			continue
		}
		if strings.EqualFold(profile.BaseURL, baseURL) && profile.WorkspaceCode == workspaceCode {
			out = append(out, name)
		}
	}
	sort.Strings(out)
	return out, nil
}

func resolveName(cfg *config.File, name string) (string, error) {
	if name == "" {
		if cfg.ActiveProfile == "" {
			return "", fmt.Errorf("no active profile; use --profile or run `imans profile use <name>`")
		}
		name = cfg.ActiveProfile
	}
	if _, ok := cfg.Profiles[name]; !ok {
		return "", fmt.Errorf("profile %q not found", name)
	}
	return name, nil
}
