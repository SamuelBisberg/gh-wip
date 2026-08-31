// Package ai generates human-readable summaries of a WIP diff by shelling
// out to a locally installed AI CLI. Every provider is best-effort: callers
// are expected to fall back to a plain timestamp message on any error.
package ai

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/SamuelBisberg/gh-wip/pkg/config"
)

// Timeout bounds how long a single summarization attempt may take before
// gh-wip gives up and falls back to a default commit message.
const Timeout = 45 * time.Second

// maxDiffChars caps how much diff text is sent to a provider, keeping
// prompts well within reach of any model's context window.
const maxDiffChars = 12000

// Provider generates a commit-message summary from a diff.
type Provider interface {
	// Name is the human-readable driver name, used in status messages.
	Name() string
	// Available reports whether the provider's backing CLI is usable.
	Available() bool
	// Summarize returns a concise commit message describing diff.
	Summarize(ctx context.Context, diff string) (string, error)
}

// New returns the configured provider, or ok=false if AI summaries are
// disabled (driver "none", or "custom" with no command configured).
func New(cfg *config.Config) (provider Provider, ok bool) {
	switch cfg.AI.Driver {
	case config.DriverClaude:
		return &cliProvider{name: "Claude", bin: "claude", args: func(p string) []string { return []string{"-p", p} }}, true
	case config.DriverCopilot:
		return &cliProvider{name: "Copilot", bin: "copilot", args: func(p string) []string { return []string{"-p", p, "-s"} }}, true
	case config.DriverCustom:
		if strings.TrimSpace(cfg.AI.CustomCommand) == "" {
			return nil, false
		}
		return &customProvider{command: cfg.AI.CustomCommand}, true
	default:
		return nil, false
	}
}

func truncateDiff(diff string) string {
	if len(diff) <= maxDiffChars {
		return diff
	}
	return diff[:maxDiffChars] + "\n... (diff truncated)"
}

func buildPrompt(diff string) string {
	return fmt.Sprintf(
		"Write a concise, conventional git commit message for the following "+
			"work-in-progress diff. Output ONLY the commit message: a summary "+
			"line under 72 characters, optionally followed by a short body. "+
			"Do not wrap it in quotes or markdown.\n\n%s",
		truncateDiff(diff),
	)
}

// cleanMessage strips common wrapping (quotes, code fences) that models
// sometimes add despite being asked not to.
func cleanMessage(s string) string {
	s = strings.TrimSpace(s)
	s = strings.Trim(s, "`")
	s = strings.Trim(s, "\"")
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	return s
}
