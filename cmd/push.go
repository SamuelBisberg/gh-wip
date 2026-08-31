package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/SamuelBisberg/gh-wip/pkg/ai"
	"github.com/SamuelBisberg/gh-wip/pkg/git"
)

func newPushCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "push",
		Short: "Capture uncommitted changes onto a remote wip/ branch",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush()
		},
	}
}

func runPush() error {
	if err := preflight(); err != nil {
		return err
	}

	cfg, theme, err := loadTheme()
	if err != nil {
		return err
	}

	dirty, err := git.HasUncommittedChanges()
	if err != nil {
		return fmt.Errorf("checking working tree status: %w", err)
	}
	if !dirty {
		theme.Infof("Nothing to push — working tree is clean.")
		return nil
	}

	originalBranch, err := git.CurrentBranch()
	if err != nil {
		return fmt.Errorf("determining current branch: %w", err)
	}
	// However this command exits, leave the user back on the branch they
	// started on: the WIP branch (with the commit already on it, if we got
	// that far) is the working copy from here on.
	defer func() {
		if err := git.Checkout(originalBranch); err != nil {
			theme.Warningf("Couldn't switch back to %s automatically: %v", originalBranch, err)
		}
	}()

	remote, err := git.DefaultRemote()
	if err != nil {
		return err
	}
	diff, err := git.WorkingDiff()
	if err != nil {
		return fmt.Errorf("reading changes: %w", err)
	}

	message := defaultCommitMessage(originalBranch)
	if provider, ok := ai.New(cfg); ok {
		if !provider.Available() {
			theme.Warningf("%s isn't installed or on PATH — using a default commit message.", provider.Name())
		} else {
			var summary string
			spinErr := theme.RunWithSpinner(fmt.Sprintf("Generating summary with %s", provider.Name()), func() error {
				ctx, cancel := context.WithTimeout(context.Background(), ai.Timeout)
				defer cancel()
				s, sErr := provider.Summarize(ctx, diff)
				summary = s
				return sErr
			})
			if spinErr != nil {
				theme.Warningf("AI summary failed (%v) — using a default commit message.", spinErr)
			} else {
				message = summary
			}
		}
	}

	branchName := git.UniqueBranchName(git.NewBranchName(time.Now()))

	commitErr := theme.RunWithSpinner(fmt.Sprintf("Capturing changes on %s", branchName), func() error {
		if err := git.CreateBranch(branchName); err != nil {
			return err
		}
		if err := git.StageAll(); err != nil {
			return err
		}
		return git.Commit(message)
	})
	if commitErr != nil {
		return fmt.Errorf("capturing WIP changes: %w", commitErr)
	}

	pushErr := theme.RunWithSpinner(fmt.Sprintf("Pushing %s to %s", branchName, remote), func() error {
		return git.Push(remote, branchName)
	})
	if pushErr != nil {
		theme.Warningf("Committed locally to %s, but pushing to %s failed: %v", branchName, remote, pushErr)
		theme.Infof("Your changes are safe locally. Retry with: git push -u %s %s", remote, branchName)
		return nil
	}

	theme.Successf("Captured WIP as %s/%s", remote, branchName)
	theme.Infof("Restore it later with: gh wip pull")
	return nil
}

func defaultCommitMessage(branch string) string {
	return fmt.Sprintf("WIP: snapshot of %s at %s", branch, time.Now().Format("2006-01-02 15:04:05 MST"))
}
