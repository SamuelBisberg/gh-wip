// Package git wraps the system git binary with the small set of porcelain
// operations gh-wip needs to stash, restore, and inspect WIP branches.
package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"
)

// CommandError wraps a failed git invocation with its stderr output so
// callers can surface something more useful than a bare exit status.
type CommandError struct {
	Args   []string
	Stderr string
	Err    error
}

func (e *CommandError) Error() string {
	stderr := strings.TrimSpace(e.Stderr)
	if stderr == "" {
		return fmt.Sprintf("git %s: %v", strings.Join(e.Args, " "), e.Err)
	}
	return fmt.Sprintf("git %s: %s", strings.Join(e.Args, " "), stderr)
}

func (e *CommandError) Unwrap() error { return e.Err }

// run executes git with the given args in the current working directory,
// returning trimmed stdout. On failure it returns a *CommandError.
func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", &CommandError{Args: args, Stderr: stderr.String(), Err: err}
	}
	return strings.TrimSpace(stdout.String()), nil
}

// runCombined behaves like run but returns stdout+stderr together even on
// success, for commands like merge where progress/conflict info is useful.
func runCombined(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := strings.TrimSpace(out.String())
	if err != nil {
		return output, &CommandError{Args: args, Stderr: output, Err: err}
	}
	return output, nil
}

// IsInsideRepo reports whether the current directory is inside a git work tree.
func IsInsideRepo() bool {
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// HasUncommittedChanges reports whether the working tree has staged,
// unstaged, or untracked changes.
func HasUncommittedChanges() (bool, error) {
	out, err := run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return out != "", nil
}

// CurrentBranch returns the name of the currently checked-out branch.
func CurrentBranch() (string, error) {
	return run("rev-parse", "--abbrev-ref", "HEAD")
}

// DefaultRemote returns "origin" if it's configured, the error otherwise.
func DefaultRemote() (string, error) {
	out, err := run("remote")
	if err != nil {
		return "", err
	}
	remotes := strings.Fields(out)
	for _, r := range remotes {
		if r == "origin" {
			return "origin", nil
		}
	}
	if len(remotes) > 0 {
		return remotes[0], nil
	}
	return "", fmt.Errorf("no git remote configured")
}

// WorkingDiff returns a unified diff of everything that would be captured
// by staging and committing right now (tracked changes plus a summary of
// untracked files), suitable for feeding to an AI summarizer.
func WorkingDiff() (string, error) {
	diff, err := run("diff", "HEAD")
	if err != nil {
		return "", err
	}

	untracked, err := run("ls-files", "--others", "--exclude-standard")
	if err != nil {
		return "", err
	}
	if untracked == "" {
		return diff, nil
	}

	var b strings.Builder
	b.WriteString(diff)
	b.WriteString("\n\nUntracked files:\n")
	for _, f := range strings.Split(untracked, "\n") {
		b.WriteString("  " + f + "\n")
	}
	return b.String(), nil
}

// StageAll stages every change in the working tree, tracked and untracked.
func StageAll() error {
	_, err := run("add", "-A")
	return err
}

// Commit records a commit with the given message.
func Commit(message string) error {
	_, err := run("commit", "-m", message)
	return err
}

// CreateBranch creates and checks out a new branch from the current HEAD.
func CreateBranch(name string) error {
	_, err := run("checkout", "-b", name)
	return err
}

// Checkout switches to an existing local branch.
func Checkout(branch string) error {
	_, err := run("checkout", branch)
	return err
}

// LocalBranchExists reports whether a local branch with the given name exists.
func LocalBranchExists(name string) bool {
	_, err := run("show-ref", "--verify", "--quiet", "refs/heads/"+name)
	return err == nil
}

// Push pushes a local branch to the remote, setting up tracking.
func Push(remote, branch string) error {
	_, err := run("push", "-u", remote, branch)
	return err
}

// Fetch fetches refs from the remote.
func Fetch(remote string) error {
	_, err := run("fetch", remote)
	return err
}

// DeleteRemoteBranch deletes a branch on the remote.
func DeleteRemoteBranch(remote, branch string) error {
	_, err := run("push", remote, "--delete", branch)
	return err
}

// MergeSquashResult describes the outcome of a squash merge attempt.
type MergeSquashResult struct {
	Conflict bool
	Output   string
}

// MergeSquash squashes ref's changes into the working tree, leaving them
// staged/uncommitted rather than creating a merge commit. On conflicts it
// returns Conflict=true instead of an error, since git leaves the tree in a
// resolvable state (conflict markers, nothing to abort) for squash merges.
func MergeSquash(ref string) (*MergeSquashResult, error) {
	output, err := runCombined("merge", "--squash", ref)
	if err == nil {
		return &MergeSquashResult{Output: output}, nil
	}

	var cmdErr *CommandError
	if errors.As(err, &cmdErr) && (strings.Contains(output, "CONFLICT") || strings.Contains(output, "Automatic merge failed")) {
		return &MergeSquashResult{Conflict: true, Output: output}, nil
	}
	return nil, err
}
