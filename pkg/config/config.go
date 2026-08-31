// Package config manages gh-wip's persistent user settings.
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// AI driver values for the "ai.driver" setting: which summarizer, if any,
// generates the commit message for `gh wip push`.
const (
	DriverNone    = "none"
	DriverClaude  = "claude"
	DriverCopilot = "copilot"
	DriverCustom  = "custom"
)

// Color values for the "ui.color" setting: whether to force, suppress, or
// auto-detect ANSI color output.
const (
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

// settingField is one entry in the config schema: a dotted key plus how to
// read and validate/write it. Get, Set, and Keys all derive from the single
// fields table below, so adding a setting means adding one entry instead of
// touching three separate switch statements that could drift out of sync.
type settingField struct {
	key string
	get func(*Config) string
	set func(*Config, string) error
}

var fields = []settingField{
	{
		key: "ai.driver",
		get: func(c *Config) string { return c.AI.Driver },
		set: func(c *Config, v string) error {
			switch v {
			case DriverNone, DriverClaude, DriverCopilot, DriverCustom:
				c.AI.Driver = v
				return nil
			default:
				return fmt.Errorf("invalid ai.driver %q (want one of: none, claude, copilot, custom)", v)
			}
		},
	},
	{
		key: "ai.customCommand",
		get: func(c *Config) string { return c.AI.CustomCommand },
		set: func(c *Config, v string) error { c.AI.CustomCommand = v; return nil },
	},
	{
		key: "pull.autoDelete",
		get: func(c *Config) string { return fmt.Sprintf("%t", c.Pull.AutoDelete) },
		set: func(c *Config, v string) error {
			switch v {
			case "true":
				c.Pull.AutoDelete = true
			case "false":
				c.Pull.AutoDelete = false
			default:
				return fmt.Errorf("invalid pull.autoDelete %q (want true or false)", v)
			}
			return nil
		},
	},
	{
		key: "ui.color",
		get: func(c *Config) string { return c.UI.Color },
		set: func(c *Config, v string) error {
			switch v {
			case ColorAuto, ColorAlways, ColorNever:
				c.UI.Color = v
				return nil
			default:
				return fmt.Errorf("invalid ui.color %q (want one of: auto, always, never)", v)
			}
		},
	},
}

func lookupField(key string) (settingField, error) {
	for _, f := range fields {
		if f.key == key {
			return f, nil
		}
	}
	return settingField{}, fmt.Errorf("unknown config key %q", key)
}

// Get returns the string value stored at a dotted key, e.g. "ai.driver".
func (c *Config) Get(key string) (string, error) {
	f, err := lookupField(key)
	if err != nil {
		return "", err
	}
	return f.get(c), nil
}

// Set validates and assigns a value to a dotted key, e.g. "ai.driver".
func (c *Config) Set(key, value string) error {
	f, err := lookupField(key)
	if err != nil {
		return err
	}
	return f.set(c, value)
}

// Keys lists every known dotted config key, in stable display order.
func Keys() []string {
	keys := make([]string, len(fields))
	for i, f := range fields {
		keys[i] = f.key
	}
	return keys
}
