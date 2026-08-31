package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-wip/pkg/git"
	"github.com/SamuelBisberg/gh-wip/pkg/tui"
)

func newCleanupCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cleanup",
		Short: "Delete multiple wip/ branches from the remote at once",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCleanup()
		},
	}
}

func runCleanup() error {
	if err := preflight(); err != nil {
		return err
	}

	_, theme, err := loadTheme()
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

	chosen, err := tui.SelectWipBranches(branches)
	if errors.Is(err, tui.ErrAborted) {
		theme.Infof("Cancelled.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("selecting WIP branches: %w", err)
	}
	if len(chosen) == 0 {
		theme.Infof("Nothing selected.")
		return nil
	}

	confirmed, err := tui.ConfirmPermanentDelete(chosen)
	if errors.Is(err, tui.ErrAborted) {
		theme.Infof("Cancelled.")
		return nil
	}
	if err != nil {
		return fmt.Errorf("confirming deletion: %w", err)
	}
	if !confirmed {
		theme.Infof("Cancelled.")
		return nil
	}

	var failed int
	for _, b := range chosen {
		remoteErr := theme.RunWithSpinner(fmt.Sprintf("Deleting %s", b.Name), func() error {
			return git.DeleteRemoteBranch(remote, b.Name)
		})
		if remoteErr != nil {
			theme.Warningf("Couldn't delete %s: %v", b.Name, remoteErr)
			failed++
			continue
		}

		if git.LocalBranchExists(b.Name) {
			if err := git.DeleteLocalBranch(b.Name); err != nil {
				theme.Warningf("Deleted %s on %s, but couldn't remove the local branch: %v", b.Name, remote, err)
			}
		}
		theme.Successf("Deleted %s", b.Name)
	}

	if failed > 0 {
		return fmt.Errorf("failed to delete %d of %d branches", failed, len(chosen))
	}
	return nil
}
