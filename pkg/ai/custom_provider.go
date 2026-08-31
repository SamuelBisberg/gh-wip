package ai

import (
	"bytes"
	"context"
	"fmt"
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
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("custom AI command timed out")
		}
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("custom AI command: %s", msg)
	}

	message := cleanMessage(stdout.String())
	if message == "" {
		return "", fmt.Errorf("custom AI command returned an empty summary")
	}
	return message, nil
}
