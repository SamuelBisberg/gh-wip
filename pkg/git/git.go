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

// CommandError wraps a failed git invocation with its stderr output and
// exit code so callers can surface something more useful than a bare exit
// status, and branch on the exit code instead of matching human-readable
// (and locale-dependent) output text.
type CommandError struct {
	Args     []string
	Stderr   string
	ExitCode int
	Err      error
}

func (e *CommandError) Error() string {
	stderr := strings.TrimSpace(e.Stderr)
	if stderr == "" {
		return fmt.Sprintf("git %s: %v", strings.Join(e.Args, " "), e.Err)
	}
	return fmt.Sprintf("git %s: %s", strings.Join(e.Args, " "), stderr)
}

func newCommandError(args []string, stderr string, err error) *CommandError {
	ce := &CommandError{Args: args, Stderr: stderr, Err: err}
	var exitErr *exec.ExitError
	if errors.As(err, &exitErr) {
		ce.ExitCode = exitErr.ExitCode()
	}
	return ce
}

// run executes git with the given args in the current working directory,
// returning stdout with its trailing newline removed. On failure it returns
// a *CommandError.
//
// Only the trailing newline is trimmed, not all leading/trailing whitespace:
// several callers (Status, WorkingDiff) parse column-sensitive formats like
// `git status --porcelain` where a line can legitimately start with a space
// (e.g. " M file.txt"), and a blanket TrimSpace would eat that space when
// it's the first character of the whole output.
func run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", newCommandError(args, stderr.String(), err)
	}
	return strings.TrimRight(stdout.String(), "\n"), nil
}

// runCombined behaves like run but returns stdout+stderr together even on
// success, for commands like merge where progress/conflict info is useful.
func runCombined(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	output := strings.TrimRight(out.String(), "\n")
	if err != nil {
		return output, newCommandError(args, output, err)
	}
	return output, nil
}

// IsInsideRepo reports whether the current directory is inside a git work tree.
func IsInsideRepo() bool {
	out, err := run("rev-parse", "--is-inside-work-tree")
	return err == nil && out == "true"
}

// StatusEntry is one line of `git status --porcelain` output: a two-letter
// status code (e.g. "M ", "??", " M") and the path it applies to.
type StatusEntry struct {
	Code string
	Path string
}

// Status returns the working tree's porcelain status, one entry per staged,
// unstaged, or untracked path.
func Status() ([]StatusEntry, error) {
	out, err := run("status", "--porcelain")
	if err != nil {
		return nil, err
	}
	if out == "" {
		return nil, nil
	}

	lines := strings.Split(out, "\n")
	entries := make([]StatusEntry, 0, len(lines))
	for _, line := range lines {
		if len(line) < 4 {
			continue
		}
		entries = append(entries, StatusEntry{Code: line[:2], Path: line[3:]})
	}
	return entries, nil
}

// HasUncommittedChanges reports whether the working tree has staged,
// unstaged, or untracked changes.
func HasUncommittedChanges() (bool, error) {
	entries, err := Status()
	if err != nil {
		return false, err
	}
	return len(entries) > 0, nil
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

	entries, err := Status()
	if err != nil {
		return "", err
	}
	var untracked []string
	for _, e := range entries {
		if e.Code == "??" {
			untracked = append(untracked, e.Path)
		}
	}
	if len(untracked) == 0 {
		return diff, nil
	}

	var b strings.Builder
	b.WriteString(diff)
	b.WriteString("\n\nUntracked files:\n")
	for _, f := range untracked {
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

// DeleteLocalBranch force-deletes a local branch regardless of its merge
// status. Force is required because WIP branches are typically restored via
// a squash merge (MergeSquash), which git doesn't record as an ancestor of
// the branch it merged into, so a plain `git branch -d` would refuse them as
// "not fully merged" even after their changes have safely landed elsewhere.
func DeleteLocalBranch(name string) error {
	_, err := run("branch", "-D", name)
	return err
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
//
// Conflicts are detected via git's exit code 1 rather than matching
// human-readable ("CONFLICT ...") output text, which is both fragile
// (wording varies across git versions) and locale-dependent. Callers are
// expected to have already confirmed a clean working tree before calling
// this, which rules out git's other exit-1 case (local changes that would
// be overwritten by the merge).
func MergeSquash(ref string) (*MergeSquashResult, error) {
	output, err := runCombined("merge", "--squash", ref)
	if err == nil {
		return &MergeSquashResult{Output: output}, nil
	}

	var cmdErr *CommandError
	if errors.As(err, &cmdErr) && cmdErr.ExitCode == 1 {
		return &MergeSquashResult{Conflict: true, Output: output}, nil
	}
	return nil, err
}
