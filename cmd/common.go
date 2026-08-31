package cmd

import (
	"fmt"

	"github.com/SamuelBisberg/gh-wip/pkg/config"
	"github.com/SamuelBisberg/gh-wip/pkg/tui"
)

// loadTheme loads the persisted config and builds the matching color theme
// - the first thing every subcommand needs before it can print anything.
func loadTheme() (*config.Config, *tui.Theme, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("loading config: %w", err)
	}
	return cfg, tui.New(cfg), nil
}
