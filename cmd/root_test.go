package cmd

import "testing"

func TestNewRootCmd(t *testing.T) {
	Version = "v1.2.3"
	t.Cleanup(func() { Version = "dev" })

	root := NewRootCmd()

	if root.Use != "wip" {
		t.Errorf("Use = %q, want %q", root.Use, "wip")
	}
	if root.Version != "v1.2.3" {
		t.Errorf("Version = %q, want %q", root.Version, "v1.2.3")
	}

	wantSubcommands := []string{"push", "pull", "config", "cleanup"}
	for _, name := range wantSubcommands {
		if root.Commands() == nil {
			t.Fatalf("root has no subcommands, want at least %v", wantSubcommands)
		}
		found := false
		for _, c := range root.Commands() {
			if c.Name() == name {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("subcommand %q not registered", name)
		}
	}
}
