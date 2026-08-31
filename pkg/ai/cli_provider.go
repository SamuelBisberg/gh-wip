package ai

import (
	"context"
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
	return runCapture(ctx, cmd, p.name)
}
