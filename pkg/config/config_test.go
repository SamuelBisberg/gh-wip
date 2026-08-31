package config

import (
	"os"
	"path/filepath"
	"testing"
)

// withConfigHome points os.UserConfigDir at a fresh temp directory so tests
// never touch the real user's config.
func withConfigHome(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	return dir
}

func TestPath(t *testing.T) {
	dir := withConfigHome(t)

	got, err := Path()
	if err != nil {
		t.Fatalf("Path() error = %v", err)
	}
	want := filepath.Join(dir, "gh", "wip.json")
	if got != want {
		t.Errorf("Path() = %q, want %q", got, want)
	}
}

func TestLoad(t *testing.T) {
	t.Run("missing file returns defaults", func(t *testing.T) {
		withConfigHome(t)

		cfg, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if cfg.AI.Driver != DriverNone {
			t.Errorf("AI.Driver = %q, want %q", cfg.AI.Driver, DriverNone)
		}
		if cfg.UI.Color != ColorAuto {
			t.Errorf("UI.Color = %q, want %q", cfg.UI.Color, ColorAuto)
		}
	})

	t.Run("existing file overrides defaults", func(t *testing.T) {
		withConfigHome(t)

		cfg := defaults()
		if err := cfg.Set("ai.driver", DriverClaude); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
		if err := cfg.Save(); err != nil {
			t.Fatalf("Save() error = %v", err)
		}

		got, err := Load()
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}
		if got.AI.Driver != DriverClaude {
			t.Errorf("AI.Driver = %q, want %q", got.AI.Driver, DriverClaude)
		}
		// Fields absent from the persisted JSON must still fall back to
		// their default rather than zero-valuing.
		if got.UI.Color != ColorAuto {
			t.Errorf("UI.Color = %q, want %q", got.UI.Color, ColorAuto)
		}
	})

	t.Run("invalid JSON returns an error", func(t *testing.T) {
		dir := withConfigHome(t)

		path := filepath.Join(dir, "gh", "wip.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatalf("MkdirAll() error = %v", err)
		}
		if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
			t.Fatalf("WriteFile() error = %v", err)
		}

		if _, err := Load(); err == nil {
			t.Error("Load() error = nil, want an error for invalid JSON")
		}
	})
}

func TestConfig_Save(t *testing.T) {
	dir := withConfigHome(t)

	cfg := defaults()
	if err := cfg.Set("pull.autoDelete", "true"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := cfg.Save(); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	path := filepath.Join(dir, "gh", "wip.json")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat(%q) error = %v", path, err)
	}
	if perm := info.Mode().Perm(); perm != 0o600 {
		t.Errorf("file mode = %o, want %o", perm, 0o600)
	}

	reloaded, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if !reloaded.Pull.AutoDelete {
		t.Error("Pull.AutoDelete = false, want true after reload")
	}
}

func TestConfig_Get(t *testing.T) {
	cfg := defaults()
	cfg.AI.CustomCommand = "my-command"
	cfg.Pull.AutoDelete = true

	tests := []struct {
		name string
		key  string
		want string
	}{
		{"ai.driver", "ai.driver", DriverNone},
		{"ai.customCommand", "ai.customCommand", "my-command"},
		{"pull.autoDelete", "pull.autoDelete", "true"},
		{"ui.color", "ui.color", ColorAuto},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := cfg.Get(tt.key)
			if err != nil {
				t.Fatalf("Get(%q) error = %v", tt.key, err)
			}
			if got != tt.want {
				t.Errorf("Get(%q) = %q, want %q", tt.key, got, tt.want)
			}
		})
	}

	t.Run("unknown key", func(t *testing.T) {
		if _, err := cfg.Get("nonexistent.key"); err == nil {
			t.Error("Get() error = nil, want an error for an unknown key")
		}
	})
}

func TestConfig_Set(t *testing.T) {
	tests := []struct {
		name    string
		key     string
		value   string
		wantErr bool
	}{
		{"valid ai.driver", "ai.driver", DriverCopilot, false},
		{"invalid ai.driver", "ai.driver", "chatgpt", true},
		{"any ai.customCommand", "ai.customCommand", "echo hi", false},
		{"valid pull.autoDelete true", "pull.autoDelete", "true", false},
		{"valid pull.autoDelete false", "pull.autoDelete", "false", false},
		{"invalid pull.autoDelete", "pull.autoDelete", "yes", true},
		{"valid ui.color", "ui.color", ColorNever, false},
		{"invalid ui.color", "ui.color", "rainbow", true},
		{"unknown key", "nonexistent.key", "x", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaults()
			err := cfg.Set(tt.key, tt.value)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Set(%q, %q) error = %v, wantErr %v", tt.key, tt.value, err, tt.wantErr)
			}
			if err != nil {
				return
			}
			got, getErr := cfg.Get(tt.key)
			if getErr != nil {
				t.Fatalf("Get(%q) error = %v", tt.key, getErr)
			}
			if got != tt.value {
				t.Errorf("after Set, Get(%q) = %q, want %q", tt.key, got, tt.value)
			}
		})
	}
}

func TestKeys(t *testing.T) {
	got := Keys()
	want := []string{"ai.driver", "ai.customCommand", "pull.autoDelete", "ui.color"}
	if len(got) != len(want) {
		t.Fatalf("Keys() = %v, want %v", got, want)
	}
	for i, k := range want {
		if got[i] != k {
			t.Errorf("Keys()[%d] = %q, want %q", i, got[i], k)
		}
	}
}
