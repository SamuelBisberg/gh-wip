package ai

import (
	"context"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/SamuelBisberg/gh-wip/pkg/config"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		command  string
		wantOK   bool
		wantName string
	}{
		{name: "claude", driver: config.DriverClaude, wantOK: true, wantName: "Claude"},
		{name: "copilot", driver: config.DriverCopilot, wantOK: true, wantName: "Copilot"},
		{name: "custom with command", driver: config.DriverCustom, command: "my-script", wantOK: true, wantName: "custom"},
		{name: "custom without command", driver: config.DriverCustom, command: "", wantOK: false},
		{name: "custom with blank command", driver: config.DriverCustom, command: "   ", wantOK: false},
		{name: "none", driver: config.DriverNone, wantOK: false},
		{name: "unknown driver", driver: "bogus", wantOK: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{}
			cfg.AI.Driver = tt.driver
			cfg.AI.CustomCommand = tt.command

			provider, ok := New(cfg)
			if ok != tt.wantOK {
				t.Fatalf("New() ok = %v, want %v", ok, tt.wantOK)
			}
			if !ok {
				if provider != nil {
					t.Errorf("New() provider = %v, want nil when ok is false", provider)
				}
				return
			}
			if provider.Name() != tt.wantName {
				t.Errorf("New() provider.Name() = %q, want %q", provider.Name(), tt.wantName)
			}
		})
	}
}

func TestTruncateDiff(t *testing.T) {
	t.Run("under the limit is unchanged", func(t *testing.T) {
		diff := "short diff"
		if got := truncateDiff(diff); got != diff {
			t.Errorf("truncateDiff() = %q, want %q", got, diff)
		}
	})

	t.Run("over the limit is truncated with a suffix", func(t *testing.T) {
		diff := strings.Repeat("a", maxDiffChars+100)
		got := truncateDiff(diff)
		wantPrefix := strings.Repeat("a", maxDiffChars)
		if !strings.HasPrefix(got, wantPrefix) {
			t.Error("truncateDiff() did not preserve the first maxDiffChars characters")
		}
		if !strings.HasSuffix(got, "... (diff truncated)") {
			t.Errorf("truncateDiff() = %q, want it to end with the truncation marker", got)
		}
		if len(got) != maxDiffChars+len("\n... (diff truncated)") {
			t.Errorf("truncateDiff() length = %d, want %d", len(got), maxDiffChars+len("\n... (diff truncated)"))
		}
	})
}

func TestBuildPrompt(t *testing.T) {
	prompt := buildPrompt("+ added line")
	if !strings.Contains(prompt, "+ added line") {
		t.Errorf("buildPrompt() = %q, want it to contain the diff", prompt)
	}
	if !strings.Contains(prompt, "commit message") {
		t.Errorf("buildPrompt() = %q, want it to ask for a commit message", prompt)
	}
}

func TestCleanMessage(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain text", in: "fix the bug", want: "fix the bug"},
		{name: "surrounding whitespace", in: "  fix the bug\n\n", want: "fix the bug"},
		{name: "wrapped in double quotes", in: `"fix the bug"`, want: "fix the bug"},
		{name: "wrapped in backticks", in: "`fix the bug`", want: "fix the bug"},
		{name: "empty input", in: "   ", want: ""},
		{name: "only quotes", in: `""`, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := cleanMessage(tt.in); got != tt.want {
				t.Errorf("cleanMessage(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestRunCapture(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo hello")
		got, err := runCapture(context.Background(), cmd, "test")
		if err != nil {
			t.Fatalf("runCapture() error = %v", err)
		}
		if got != "hello" {
			t.Errorf("runCapture() = %q, want %q", got, "hello")
		}
	})

	t.Run("empty output is an error", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "true")
		_, err := runCapture(context.Background(), cmd, "test")
		if err == nil || !strings.Contains(err.Error(), "empty summary") {
			t.Errorf("runCapture() error = %v, want an empty-summary error", err)
		}
	})

	t.Run("non-zero exit surfaces stderr", func(t *testing.T) {
		cmd := exec.Command("sh", "-c", "echo boom >&2; exit 1")
		_, err := runCapture(context.Background(), cmd, "test")
		if err == nil || !strings.Contains(err.Error(), "boom") {
			t.Errorf("runCapture() error = %v, want it to contain %q", err, "boom")
		}
	})

	t.Run("context timeout", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		cmd := exec.CommandContext(ctx, "sh", "-c", "sleep 1")
		_, err := runCapture(ctx, cmd, "test")
		if err == nil || !strings.Contains(err.Error(), "timed out") {
			t.Errorf("runCapture() error = %v, want a timeout error", err)
		}
	})
}
