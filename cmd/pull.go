package cmd

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-wip/pkg/git"
	"github.com/SamuelBisberg/gh-wip/pkg/tui"
)

func newPullCmd() *cobra.Command {
	var deleteAfter bool

	c := &cobra.Command{
		Use:   "pull",
		Short: "Pick a wip/ branch and merge its changes back into your working directory",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(deleteAfter)
		},
	}
	c.Flags().BoolVar(&deleteAfter, "delete", false, "delete the remote WIP branch after a successful pull, without prompting")
	return c
}

func runPull(deleteFlag bool) error {
	if err := preflight(); err != nil {
		return err
	}

	dirty, err := git.HasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("checking working tree status: %w", err)
	}
	if dirty {
		return fmt.Errorf("you have uncommitted changes — commit, stash, or run `gh wip push` before pulling a WIP branch")
	}

	cfg, theme, err := loadTheme()
	if err != nil {
		return err
	}

	remote, err := git.DefaultRemote()
	if err != nil {
		return err
	}

	if err := theme.RunWithSpinner(fmt.Sprintf("Fetching from %s", remote), func() error {
		return git.Fetch(remote)
	}); err != nil {
		return fmt.Errorf("fetching from %s: %w", remote, err)
	}

	branches, err := git.ListWipBranches(remote)
	if err != nil {
		return fmt.Errorf("listing WIP branches: %w", err)
	}
	if len(branches) == 0 {
		theme.Infof("No WIP branches found on %s.", remote)
		return nil
	}

	chosen, err := tui.SelectWipBranch(branches)
	if errors.Is(err, tui.ErrAborted) {
		theme.Infof("Cancelled.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("selecting a WIP branch: %w", err)
	}

	result, err := git.MergeSquash(chosen.Ref(remote))
	if err != nil {
		return fmt.Errorf("merging %s: %w", chosen.Name, err)
	}

	if result.Conflict {
		theme.Warningf("Merging %s produced conflicts:", chosen.Name)
		for _, line := range strings.Split(result.Output, "\n") {
			if strings.Contains(line, "CONFLICT") {
				fmt.Println("  " + line)
			}
		}
		theme.Infof("Resolve the conflicts, `git add` the fixed files, then commit when ready.")
		return nil
	}

	theme.Successf("Merged %s — changes are staged in your working directory, uncommitted.", chosen.Name)

	shouldDelete := deleteFlag || cfg.Pull.AutoDelete
	if !shouldDelete {
		ok, cErr := tui.ConfirmDelete(chosen.Name)
		if cErr != nil && !errors.Is(cErr, tui.ErrAborted) {
			return fmt.Errorf("confirming branch deletion: %w", cErr)
		}
		shouldDelete = ok
	}

	if shouldDelete {
		if err := theme.RunWithSpinner(fmt.Sprintf("Deleting remote branch %s", chosen.Name), func() error {
			return git.DeleteRemoteBranch(remote, chosen.Name)
		}); err != nil {
			theme.Warningf("Couldn't delete remote branch %s: %v", chosen.Name, err)
		} else {
			theme.Successf("Deleted remote branch %s", chosen.Name)
		}
	}

	return nil
}
