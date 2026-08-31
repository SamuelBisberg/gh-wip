package ai

import (
	"context"
	"strings"
	"testing"
)

func TestCustomProvider_Name(t *testing.T) {
	p := &customProvider{command: "echo hi"}
	if got := p.Name(); got != "custom" {
		t.Errorf("Name() = %q, want %q", got, "custom")
	}
}

func TestCustomProvider_Available(t *testing.T) {
	tests := []struct {
		name    string
		command string
		want    bool
	}{
		{name: "configured command", command: "echo hi", want: true},
		{name: "empty command", command: "", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &customProvider{command: tt.command}
			if got := p.Available(); got != tt.want {
				t.Errorf("Available() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCustomProvider_Summarize(t *testing.T) {
	t.Run("pipes the diff to stdin", func(t *testing.T) {
		p := &customProvider{command: "cat"}
		got, err := p.Summarize(context.Background(), "some diff text")
		if err != nil {
			t.Fatalf("Summarize() error = %v", err)
		}
		if got != "some diff text" {
			t.Errorf("Summarize() = %q, want %q", got, "some diff text")
		}
	})

	t.Run("empty diff produces an empty-summary error", func(t *testing.T) {
		p := &customProvider{command: "cat"}
		_, err := p.Summarize(context.Background(), "")
		if err == nil || !strings.Contains(err.Error(), "empty summary") {
			t.Errorf("Summarize() error = %v, want an empty-summary error", err)
		}
	})

	t.Run("command failure propagates", func(t *testing.T) {
		p := &customProvider{command: "exit 1"}
		if _, err := p.Summarize(context.Background(), "diff"); err == nil {
			t.Error("Summarize() error = nil, want an error when the command fails")
		}
	})
}
