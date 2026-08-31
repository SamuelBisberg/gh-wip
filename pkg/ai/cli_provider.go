package ai

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
)

// cliProvider drives a headless-capable local AI CLI (e.g. `claude -p` or
// `copilot -p ... -s`) that accepts a single prompt argument and prints its
// answer to stdout.
type cliProvider struct {
	name string
	bin  string
	args func(prompt string) []string
}

func (p *cliProvider) Name() string { return p.name }

func (p *cliProvider) Available() bool {
	_, err := exec.LookPath(p.bin)
	return err == nil
}

func (p *cliProvider) Summarize(ctx context.Context, diff string) (string, error) {
	cmd := exec.CommandContext(ctx, p.bin, p.args(buildPrompt(diff))...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return "", fmt.Errorf("%s timed out", p.name)
		}
		msg := stderr.String()
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("%s: %s", p.name, msg)
	}

	message := cleanMessage(stdout.String())
	if message == "" {
		return "", fmt.Errorf("%s returned an empty summary", p.name)
	}
	return message, nil
}
