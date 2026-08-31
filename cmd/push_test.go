package cmd

import (
	"strings"
	"testing"
)

func TestDefaultCommitMessage(t *testing.T) {
	got := defaultCommitMessage("feature/foo")
	if !strings.HasPrefix(got, "WIP: snapshot of feature/foo at ") {
		t.Errorf("defaultCommitMessage() = %q, want it to start with %q", got, "WIP: snapshot of feature/foo at ")
	}
}
