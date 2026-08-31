package git

import (
	"os"
	"os/exec"
	"testing"
)

// newTestRepo creates a git repository in a fresh temp directory and chdirs
// the test process into it, since every function in this package operates
// on the process's current directory rather than an explicit path. It
// returns the repository's path with one empty commit on "main" already
// made, so CurrentBranch/HEAD-dependent operations have something to work
// with. Tests using it cannot run in parallel with each other, since the
// working directory is process-global.
func newTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	chdir(t, dir)

	runGit(t, "init", "-b", "main")
	runGit(t, "config", "user.email", "test@example.com")
	runGit(t, "config", "user.name", "Test")
	runGit(t, "commit", "--allow-empty", "-m", "initial commit")
	return dir
}

// newBareRemote creates a bare repository in a fresh temp directory, suitable
// for use as a push/fetch target without needing network access.
func newBareRemote(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "init", "--bare", "-b", "main", dir).CombinedOutput(); err != nil {
		t.Fatalf("git init --bare: %v\n%s", err, out)
	}
	return dir
}

// chdir switches the test process's working directory to dir and restores
// the original one on cleanup.
func chdir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd() error = %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("Chdir(%q) error = %v", dir, err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(orig); err != nil {
			t.Fatalf("restoring Chdir(%q) error = %v", orig, err)
		}
	})
}

// runGit runs git with args in the test process's current directory, failing
// the test immediately on error.
func runGit(t *testing.T, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", args...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return string(out)
}

// runGitDir behaves like runGit but runs in dir instead of the current
// directory, for driving a second repository (e.g. another clone) without
// disturbing the current one.
func runGitDir(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v (in %s): %v\n%s", args, dir, err, out)
	}
	return string(out)
}

// writeFile writes content to a path relative to the current directory,
// failing the test on error.
func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%q) error = %v", path, err)
	}
}
