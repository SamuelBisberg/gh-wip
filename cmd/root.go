// Package cmd wires gh-wip's cobra command tree.
package cmd

import "github.com/spf13/cobra"

// Version is injected at build time (see main.go / .github/workflows/release.yml).
var Version = "dev"

// NewRootCmd builds the `gh wip` root command and all its subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "wip",
		Short: "Capture and restore work-in-progress state via GitHub",
		Long: "gh-wip lets you quickly stash uncommitted changes onto a remote wip/ branch\n" +
			"and pull them back down later - from this machine or any other.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       Version,
	}

	root.AddCommand(newPushCmd())
	root.AddCommand(newPullCmd())
	root.AddCommand(newConfigCmd())
	root.AddCommand(newCleanupCmd())
	return root
}
