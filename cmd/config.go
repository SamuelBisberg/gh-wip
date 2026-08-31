package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-wip/pkg/config"
	"github.com/SamuelBisberg/gh-wip/pkg/tui"
)

func newConfigCmd() *cobra.Command {
	c := &cobra.Command{
		Use:   "config",
		Short: "View or change gh-wip settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList()
		},
	}
	c.AddCommand(newConfigListCmd(), newConfigGetCmd(), newConfigSetCmd())
	return c
}

func newConfigListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show all current settings",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runConfigList()
		},
	}
}

func newConfigGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <key>",
		Short: "Print the value of a single setting",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			value, err := cfg.Get(args[0])
			if err != nil {
				return err
			}
			fmt.Println(value)
			return nil
		},
	}
}

func newConfigSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <key> <value>",
		Short: "Change a setting and save it",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("loading config: %w", err)
			}
			key := args[0]
			value := strings.Join(args[1:], " ")
			if err := cfg.Set(key, value); err != nil {
				return err
			}
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("saving config: %w", err)
			}
			tui.New(cfg).Successf("Set %s = %s", key, value)
			return nil
		},
	}
}

func runConfigList() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}
	theme := tui.New(cfg)

	path, err := config.Path()
	if err == nil {
		theme.Infof("Config file: %s", path)
	}
	for _, key := range config.Keys() {
		value, _ := cfg.Get(key)
		if value == "" {
			value = theme.Faint.Render("(unset)")
		}
		fmt.Printf("  %s = %s\n", theme.Bold.Render(key), value)
	}
	return nil
}
