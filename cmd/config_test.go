package cmd

import (
	"testing"

	"github.com/SamuelBisberg/gh-wip/pkg/config"
)

func TestRunConfigGet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	t.Run("known key returns its default", func(t *testing.T) {
		if err := runConfigGet("ai.driver"); err != nil {
			t.Errorf("runConfigGet() error = %v", err)
		}
	})

	t.Run("unknown key errors", func(t *testing.T) {
		if err := runConfigGet("nonexistent.key"); err == nil {
			t.Error("runConfigGet() error = nil, want an error for an unknown key")
		}
	})
}

func TestRunConfigSet(t *testing.T) {
	t.Run("valid value persists to disk", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		if err := runConfigSet("ai.driver", config.DriverClaude); err != nil {
			t.Fatalf("runConfigSet() error = %v", err)
		}

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("config.Load() error = %v", err)
		}
		if cfg.AI.Driver != config.DriverClaude {
			t.Errorf("persisted AI.Driver = %q, want %q", cfg.AI.Driver, config.DriverClaude)
		}
	})

	t.Run("invalid value is rejected without touching disk", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", t.TempDir())

		if err := runConfigSet("ai.driver", "not-a-real-driver"); err == nil {
			t.Fatal("runConfigSet() error = nil, want an error for an invalid value")
		}

		cfg, err := config.Load()
		if err != nil {
			t.Fatalf("config.Load() error = %v", err)
		}
		if cfg.AI.Driver != config.DriverNone {
			t.Errorf("AI.Driver = %q, want the default %q after a rejected Set", cfg.AI.Driver, config.DriverNone)
		}
	})
}

func TestRunConfigList(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	if err := runConfigList(); err != nil {
		t.Errorf("runConfigList() error = %v", err)
	}
}
