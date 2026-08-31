// Package config manages gh-wip's persistent user settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

const (
	DriverNone    = "none"
	DriverClaude  = "claude"
	DriverCopilot = "copilot"
	DriverCustom  = "custom"

	ColorAuto   = "auto"
	ColorAlways = "always"
	ColorNever  = "never"
)

// Config holds all persisted gh-wip settings.
type Config struct {
	AI struct {
		Driver        string `json:"driver"`
		CustomCommand string `json:"customCommand"`
	} `json:"ai"`
	Pull struct {
		AutoDelete bool `json:"autoDelete"`
	} `json:"pull"`
	UI struct {
		Color string `json:"color"`
	} `json:"ui"`
}

func defaults() *Config {
	c := &Config{}
	c.AI.Driver = DriverNone
	c.UI.Color = ColorAuto
	return c
}

// Path returns the on-disk location of the config file, honoring
// XDG_CONFIG_HOME / os.UserConfigDir so it resolves to ~/.config/gh/wip.json
// on Linux and the platform-appropriate equivalent elsewhere.
func Path() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "gh", "wip.json"), nil
}

// Load reads the config file, returning defaults if none exists yet.
func Load() (*Config, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return defaults(), nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading config: %w", err)
	}

	cfg := defaults()
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config at %s: %w", path, err)
	}
	return cfg, nil
}

// Save writes the config file, creating its parent directory if needed.
func (c *Config) Save() error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding config: %w", err)
	}
	data = append(data, '\n')

	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("writing config at %s: %w", path, err)
	}
	return nil
}

// Get returns the string value stored at a dotted key, e.g. "ai.driver".
func (c *Config) Get(key string) (string, error) {
	switch key {
	case "ai.driver":
		return c.AI.Driver, nil
	case "ai.customCommand":
		return c.AI.CustomCommand, nil
	case "pull.autoDelete":
		return fmt.Sprintf("%t", c.Pull.AutoDelete), nil
	case "ui.color":
		return c.UI.Color, nil
	default:
		return "", fmt.Errorf("unknown config key %q", key)
	}
}

// Set validates and assigns a value to a dotted key, e.g. "ai.driver".
func (c *Config) Set(key, value string) error {
	switch key {
	case "ai.driver":
		switch value {
		case DriverNone, DriverClaude, DriverCopilot, DriverCustom:
			c.AI.Driver = value
		default:
			return fmt.Errorf("invalid ai.driver %q (want one of: none, claude, copilot, custom)", value)
		}
	case "ai.customCommand":
		c.AI.CustomCommand = value
	case "pull.autoDelete":
		switch value {
		case "true":
			c.Pull.AutoDelete = true
		case "false":
			c.Pull.AutoDelete = false
		default:
			return fmt.Errorf("invalid pull.autoDelete %q (want true or false)", value)
		}
	case "ui.color":
		switch value {
		case ColorAuto, ColorAlways, ColorNever:
			c.UI.Color = value
		default:
			return fmt.Errorf("invalid ui.color %q (want one of: auto, always, never)", value)
		}
	default:
		return fmt.Errorf("unknown config key %q", key)
	}
	return nil
}

// Keys lists every known dotted config key, in stable display order.
func Keys() []string {
	return []string{"ai.driver", "ai.customCommand", "pull.autoDelete", "ui.color"}
}
