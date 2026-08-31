package ai

import (
	"context"
	"strings"
	"testing"
)

func TestCliProvider_Name(t *testing.T) {
	p := &cliProvider{name: "Claude"}
	if got := p.Name(); got != "Claude" {
		t.Errorf("Name() = %q, want %q", got, "Claude")
	}
}

func TestCliProvider_Available(t *testing.T) {
	t.Run("binary on PATH", func(t *testing.T) {
		p := &cliProvider{bin: "sh"}
		if !p.Available() {
			t.Error("Available() = false, want true for `sh`")
		}
	})

	t.Run("binary not on PATH", func(t *testing.T) {
		p := &cliProvider{bin: "gh-wip-nonexistent-binary"}
		if p.Available() {
			t.Error("Available() = true, want false for a made-up binary name")
		}
	})
}

func TestCliProvider_Summarize(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := &cliProvider{
			name: "test",
			bin:  "sh",
			args: func(prompt string) []string { return []string{"-c", "echo generated summary"} },
		}
		got, err := p.Summarize(context.Background(), "some diff")
		if err != nil {
			t.Fatalf("Summarize() error = %v", err)
		}
		if got != "generated summary" {
			t.Errorf("Summarize() = %q, want %q", got, "generated summary")
		}
	})

	t.Run("builds args from the prompt", func(t *testing.T) {
		var gotPrompt string
		p := &cliProvider{
			name: "test",
			bin:  "sh",
			args: func(prompt string) []string {
				gotPrompt = prompt
				return []string{"-c", "echo ok"}
			},
		}
		if _, err := p.Summarize(context.Background(), "+ added a line"); err != nil {
			t.Fatalf("Summarize() error = %v", err)
		}
		if !strings.Contains(gotPrompt, "+ added a line") {
			t.Errorf("prompt passed to args() = %q, want it to contain the diff", gotPrompt)
		}
	})

	t.Run("failure propagates the error", func(t *testing.T) {
		p := &cliProvider{
			name: "test",
			bin:  "sh",
			args: func(prompt string) []string { return []string{"-c", "exit 1"} },
		}
		if _, err := p.Summarize(context.Background(), "diff"); err == nil {
			t.Error("Summarize() error = nil, want an error when the command fails")
		}
	})
}
