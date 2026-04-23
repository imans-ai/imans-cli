package config

import (
	"testing"
)

func TestManagerSaveAndLoad(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	manager, err := NewManager("imans")
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}

	cfg := &File{
		ActiveProfile: "acme",
		Profiles: map[string]Profile{
			"acme": {
				BaseURL:       "https://api.imans.ai/",
				WorkspaceCode: "workspace-1",
				WorkspaceName: "Acme",
			},
		},
	}
	if err := manager.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.ActiveProfile != "acme" {
		t.Fatalf("ActiveProfile = %q", loaded.ActiveProfile)
	}
	if loaded.Profiles["acme"].WorkspaceName != "Acme" {
		t.Fatalf("WorkspaceName = %q", loaded.Profiles["acme"].WorkspaceName)
	}
}
