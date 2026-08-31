package git

import (
	"errors"
	"strings"
	"testing"
)

func TestCommandError_Error(t *testing.T) {
	tests := []struct {
		name string
		err  *CommandError
		want string
	}{
		{
			name: "with stderr",
			err:  &CommandError{Args: []string{"status"}, Stderr: "fatal: not a git repository\n"},
			want: "git status: fatal: not a git repository",
		},
		{
			name: "without stderr falls back to Err",
			err:  &CommandError{Args: []string{"push"}, Err: errors.New("boom")},
			want: "git push: boom",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.Error(); got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsInsideRepo(t *testing.T) {
	t.Run("inside a repo", func(t *testing.T) {
		newTestRepo(t)
		if !IsInsideRepo() {
			t.Error("IsInsideRepo() = false, want true")
		}
	})

	t.Run("outside a repo", func(t *testing.T) {
		chdir(t, t.TempDir())
		if IsInsideRepo() {
			t.Error("IsInsideRepo() = true, want false")
		}
	})
}

func TestStatus(t *testing.T) {
	newTestRepo(t)

	t.Run("clean tree", func(t *testing.T) {
		entries, err := Status()
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		if entries != nil {
			t.Errorf("Status() = %v, want nil", entries)
		}
	})

	t.Run("untracked and modified files", func(t *testing.T) {
		writeFile(t, "tracked.txt", "v1\n")
		runGit(t, "add", "tracked.txt")
		runGit(t, "commit", "-m", "add tracked.txt")

		writeFile(t, "tracked.txt", "v2\n")
		writeFile(t, "untracked.txt", "new\n")

		entries, err := Status()
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}

		codes := map[string]string{}
		for _, e := range entries {
			codes[e.Path] = e.Code
		}
		if codes["tracked.txt"] != " M" {
			t.Errorf("tracked.txt code = %q, want %q", codes["tracked.txt"], " M")
		}
		if codes["untracked.txt"] != "??" {
			t.Errorf("untracked.txt code = %q, want %q", codes["untracked.txt"], "??")
		}
	})
}

func TestHasUncommittedChanges(t *testing.T) {
	newTestRepo(t)

	dirty, err := HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error = %v", err)
	}
	if dirty {
		t.Error("HasUncommittedChanges() = true, want false on a clean tree")
	}

	writeFile(t, "new.txt", "x\n")

	dirty, err = HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error = %v", err)
	}
	if !dirty {
		t.Error("HasUncommittedChanges() = false, want true with an untracked file")
	}
}

func TestCurrentBranch(t *testing.T) {
	newTestRepo(t)

	got, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if got != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "main")
	}

	runGit(t, "checkout", "-b", "feature")

	got, err = CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if got != "feature" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "feature")
	}
}

func TestDefaultRemote(t *testing.T) {
	t.Run("no remotes configured", func(t *testing.T) {
		newTestRepo(t)
		if _, err := DefaultRemote(); err == nil {
			t.Error("DefaultRemote() error = nil, want an error with no remotes configured")
		}
	})

	t.Run("prefers origin", func(t *testing.T) {
		newTestRepo(t)
		runGit(t, "remote", "add", "upstream", "https://example.com/upstream.git")
		runGit(t, "remote", "add", "origin", "https://example.com/origin.git")

		got, err := DefaultRemote()
		if err != nil {
			t.Fatalf("DefaultRemote() error = %v", err)
		}
		if got != "origin" {
			t.Errorf("DefaultRemote() = %q, want %q", got, "origin")
		}
	})

	t.Run("falls back to the only remote", func(t *testing.T) {
		newTestRepo(t)
		runGit(t, "remote", "add", "upstream", "https://example.com/upstream.git")

		got, err := DefaultRemote()
		if err != nil {
			t.Fatalf("DefaultRemote() error = %v", err)
		}
		if got != "upstream" {
			t.Errorf("DefaultRemote() = %q, want %q", got, "upstream")
		}
	})
}

func TestWorkingDiff(t *testing.T) {
	t.Run("with untracked files", func(t *testing.T) {
		newTestRepo(t)
		writeFile(t, "tracked.txt", "v1\n")
		runGit(t, "add", "tracked.txt")
		runGit(t, "commit", "-m", "add tracked.txt")

		writeFile(t, "tracked.txt", "v2\n")
		writeFile(t, "untracked.txt", "new\n")

		diff, err := WorkingDiff()
		if err != nil {
			t.Fatalf("WorkingDiff() error = %v", err)
		}
		if !strings.Contains(diff, "tracked.txt") {
			t.Errorf("WorkingDiff() = %q, want it to mention tracked.txt", diff)
		}
		if !strings.Contains(diff, "Untracked files:") || !strings.Contains(diff, "untracked.txt") {
			t.Errorf("WorkingDiff() = %q, want an Untracked files section listing untracked.txt", diff)
		}
	})

	t.Run("without untracked files", func(t *testing.T) {
		newTestRepo(t)
		writeFile(t, "tracked.txt", "v1\n")
		runGit(t, "add", "tracked.txt")
		runGit(t, "commit", "-m", "add tracked.txt")
		writeFile(t, "tracked.txt", "v2\n")

		diff, err := WorkingDiff()
		if err != nil {
			t.Fatalf("WorkingDiff() error = %v", err)
		}
		if strings.Contains(diff, "Untracked files:") {
			t.Errorf("WorkingDiff() = %q, want no Untracked files section", diff)
		}
	})
}

func TestStageAll(t *testing.T) {
	newTestRepo(t)
	writeFile(t, "a.txt", "a\n")
	writeFile(t, "b.txt", "b\n")

	if err := StageAll(); err != nil {
		t.Fatalf("StageAll() error = %v", err)
	}

	entries, err := Status()
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	for _, e := range entries {
		if e.Code != "A " {
			t.Errorf("Status()[%s].Code = %q, want %q", e.Path, e.Code, "A ")
		}
	}
}

func TestCommit(t *testing.T) {
	newTestRepo(t)
	writeFile(t, "a.txt", "a\n")
	if err := StageAll(); err != nil {
		t.Fatalf("StageAll() error = %v", err)
	}

	if err := Commit("add a.txt"); err != nil {
		t.Fatalf("Commit() error = %v", err)
	}

	dirty, err := HasUncommittedChanges()
	if err != nil {
		t.Fatalf("HasUncommittedChanges() error = %v", err)
	}
	if dirty {
		t.Error("HasUncommittedChanges() = true, want false right after Commit()")
	}

	subject := strings.TrimSpace(runGit(t, "log", "-1", "--format=%s"))
	if subject != "add a.txt" {
		t.Errorf("last commit subject = %q, want %q", subject, "add a.txt")
	}
}

func TestCreateBranch(t *testing.T) {
	newTestRepo(t)

	if err := CreateBranch("feature"); err != nil {
		t.Fatalf("CreateBranch() error = %v", err)
	}

	got, err := CurrentBranch()
	if err != nil {
		t.Fatalf("CurrentBranch() error = %v", err)
	}
	if got != "feature" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "feature")
	}
}

func TestCheckout(t *testing.T) {
	newTestRepo(t)
	runGit(t, "branch", "feature")

	if err := Checkout("feature"); err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	if got, _ := CurrentBranch(); got != "feature" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "feature")
	}

	if err := Checkout("main"); err != nil {
		t.Fatalf("Checkout() error = %v", err)
	}
	if got, _ := CurrentBranch(); got != "main" {
		t.Errorf("CurrentBranch() = %q, want %q", got, "main")
	}
}

func TestLocalBranchExists(t *testing.T) {
	newTestRepo(t)
	runGit(t, "branch", "feature")

	if !LocalBranchExists("feature") {
		t.Error("LocalBranchExists(feature) = false, want true")
	}
	if LocalBranchExists("does-not-exist") {
		t.Error("LocalBranchExists(does-not-exist) = true, want false")
	}
}

func TestPush(t *testing.T) {
	newTestRepo(t)
	remoteDir := newBareRemote(t)
	runGit(t, "remote", "add", "origin", remoteDir)

	if err := Push("origin", "main"); err != nil {
		t.Fatalf("Push() error = %v", err)
	}

	out := runGitDir(t, remoteDir, "branch", "--list", "main")
	if strings.TrimSpace(out) == "" {
		t.Error("Push() did not create refs/heads/main on the remote")
	}
}

func TestFetch(t *testing.T) {
	newTestRepo(t)
	remoteDir := newBareRemote(t)
	runGit(t, "remote", "add", "origin", remoteDir)
	if err := Push("origin", "main"); err != nil {
		t.Fatalf("setup Push() error = %v", err)
	}

	// Advance the remote's main from a second clone, independently of the
	// repo under test, then confirm Fetch pulls the new tip.
	otherDir := t.TempDir()
	runGitDir(t, "", "clone", remoteDir, otherDir)
	runGitDir(t, otherDir, "config", "user.email", "test@example.com")
	runGitDir(t, otherDir, "config", "user.name", "Test")
	runGitDir(t, otherDir, "commit", "--allow-empty", "-m", "advance main")
	runGitDir(t, otherDir, "push", "origin", "main")
	remoteTip := strings.TrimSpace(runGitDir(t, otherDir, "rev-parse", "HEAD"))

	if err := Fetch("origin"); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	localTip := strings.TrimSpace(runGit(t, "rev-parse", "origin/main"))
	if localTip != remoteTip {
		t.Errorf("origin/main = %s, want %s", localTip, remoteTip)
	}
}

func TestDeleteRemoteBranch(t *testing.T) {
	newTestRepo(t)
	remoteDir := newBareRemote(t)
	runGit(t, "remote", "add", "origin", remoteDir)
	runGit(t, "checkout", "-b", "wip/test")
	if err := Push("origin", "wip/test"); err != nil {
		t.Fatalf("setup Push() error = %v", err)
	}

	if err := DeleteRemoteBranch("origin", "wip/test"); err != nil {
		t.Fatalf("DeleteRemoteBranch() error = %v", err)
	}

	out := runGitDir(t, remoteDir, "branch", "--list", "wip/test")
	if strings.TrimSpace(out) != "" {
		t.Errorf("wip/test still exists on the remote after DeleteRemoteBranch(): %q", out)
	}
}

func TestMergeSquash(t *testing.T) {
	t.Run("clean merge", func(t *testing.T) {
		newTestRepo(t)
		runGit(t, "checkout", "-b", "feature")
		writeFile(t, "feature.txt", "hi\n")
		runGit(t, "add", "feature.txt")
		runGit(t, "commit", "-m", "add feature.txt")
		runGit(t, "checkout", "main")

		result, err := MergeSquash("feature")
		if err != nil {
			t.Fatalf("MergeSquash() error = %v", err)
		}
		if result.Conflict {
			t.Fatalf("MergeSquash() Conflict = true, want false: %s", result.Output)
		}

		entries, err := Status()
		if err != nil {
			t.Fatalf("Status() error = %v", err)
		}
		found := false
		for _, e := range entries {
			if e.Path == "feature.txt" {
				found = true
			}
		}
		if !found {
			t.Error("feature.txt not staged after MergeSquash()")
		}
	})

	t.Run("conflicting merge", func(t *testing.T) {
		newTestRepo(t)
		writeFile(t, "shared.txt", "base\n")
		runGit(t, "add", "shared.txt")
		runGit(t, "commit", "-m", "add shared.txt")

		runGit(t, "checkout", "-b", "feature")
		writeFile(t, "shared.txt", "feature change\n")
		runGit(t, "commit", "-am", "change on feature")

		runGit(t, "checkout", "main")
		writeFile(t, "shared.txt", "main change\n")
		runGit(t, "commit", "-am", "change on main")

		result, err := MergeSquash("feature")
		if err != nil {
			t.Fatalf("MergeSquash() error = %v", err)
		}
		if !result.Conflict {
			t.Fatalf("MergeSquash() Conflict = false, want true: %s", result.Output)
		}
	})
}
