package cli

import (
	"fmt"
	"io"

	"github.com/imans-ai/imans-cli/internal/apperrors"
	"github.com/imans-ai/imans-cli/internal/client"
	"github.com/imans-ai/imans-cli/internal/config"
	"github.com/imans-ai/imans-cli/internal/output"
	"github.com/imans-ai/imans-cli/internal/profiles"
	"github.com/imans-ai/imans-cli/internal/secrets"
	"github.com/imans-ai/imans-cli/internal/version"
)

type IOStreams struct {
	In     io.Reader
	Out    io.Writer
	ErrOut io.Writer
}

type App struct {
	IO       IOStreams
	Version  version.Info
	Config   *config.Manager
	Secrets  secrets.Store
	Profiles *profiles.Service
}

func New(io IOStreams) (*App, error) {
	configManager, err := config.NewManager("imans")
	if err != nil {
		return nil, err
	}
	secretStore := secrets.NewStore("imans", configManager.ConfigDir())
	return &App{
		IO:       io,
		Version:  version.Current(),
		Config:   configManager,
		Secrets:  secretStore,
		Profiles: profiles.NewService(configManager, secretStore),
	}, nil
}

func (a *App) Printer(jsonOutput, quiet bool) *output.Printer {
	return &output.Printer{Out: a.IO.Out, ErrOut: a.IO.ErrOut, JSON: jsonOutput, Quiet: quiet}
}

func (a *App) APIClient(profileName string, debug bool) (*client.Client, profiles.Entry, error) {
	entry, err := a.Profiles.Resolve(profileName)
	if err != nil {
		return nil, profiles.Entry{}, err
	}
	token, err := a.Secrets.Get(entry.Name)
	if err != nil {
		if err == secrets.ErrNotFound {
			return nil, profiles.Entry{}, apperrors.New(apperrors.ExitUsage, "profile token not found in secret storage")
		}
		return nil, profiles.Entry{}, err
	}
	apiClient, err := client.New(client.Options{
		BaseURL:   entry.Profile.BaseURL,
		Token:     token,
		UserAgent: fmt.Sprintf("imans-cli/%s", a.Version.Version),
		Debug:     debug,
		ErrOut:    a.IO.ErrOut,
	})
	if err != nil {
		return nil, profiles.Entry{}, err
	}
	return apiClient, entry, nil
}
