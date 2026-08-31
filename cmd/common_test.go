package cmd

import (
	"strings"
	"testing"

	"github.com/SamuelBisberg/gh-wip/pkg/config"
)

func TestLoadTheme(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	cfg, theme, err := loadTheme()
	if err != nil {
		t.Fatalf("loadTheme() error = %v", err)
	}
	if cfg.AI.Driver != config.DriverNone {
		t.Errorf("cfg.AI.Driver = %q, want %q", cfg.AI.Driver, config.DriverNone)
	}
	if theme == nil {
		t.Error("loadTheme() theme = nil, want a *tui.Theme")
	}
}

// isolateGH clears every source go-gh's auth package consults for a token
// or host (env vars, its on-disk config, and any real `gh` binary on PATH),
// so checkAuth's result depends only on what the test itself sets up rather
// than the machine it happens to run on.
func isolateGH(t *testing.T) {
	t.Helper()
	t.Setenv("GH_CONFIG_DIR", t.TempDir())
	t.Setenv("GH_PATH", "/nonexistent/gh")
	for _, v := range []string{"GH_HOST", "GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN", "GITHUB_ENTERPRISE_TOKEN"} {
		t.Setenv(v, "")
	}
}

func TestCheckAuth(t *testing.T) {
	t.Run("not logged in", func(t *testing.T) {
		isolateGH(t)

		err := checkAuth()
		if err == nil {
			t.Fatal("checkAuth() error = nil, want an error with no token available")
		}
		if !strings.Contains(err.Error(), "not logged in") {
			t.Errorf("checkAuth() error = %q, want it to mention %q", err.Error(), "not logged in")
		}
	})

	t.Run("logged in via GH_TOKEN", func(t *testing.T) {
		isolateGH(t)
		t.Setenv("GH_TOKEN", "test-token")

		if err := checkAuth(); err != nil {
			t.Errorf("checkAuth() error = %v, want nil with GH_TOKEN set", err)
		}
	})
}
