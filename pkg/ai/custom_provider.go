package ai

import (
	"bytes"
	"context"
	"os/exec"
)

// customProvider runs a user-configured shell command (ai.customCommand),
// piping the raw diff to its stdin and capturing stdout as the summary.
// This is the escape hatch for any AI tool gh-wip doesn't wire up directly.
type customProvider struct {
	command string
}

func (p *customProvider) Name() string { return "custom" }

func (p *customProvider) Available() bool { return p.command != "" }

func (p *customProvider) Summarize(ctx context.Context, diff string) (string, error) {
	cmd := exec.CommandContext(ctx, "sh", "-c", p.command)
	cmd.Stdin = bytes.NewReader([]byte(truncateDiff(diff)))
	return runCapture(ctx, cmd, "custom AI command")
}
