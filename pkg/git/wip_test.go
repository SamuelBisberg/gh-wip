package git

import (
	"testing"
	"time"
)

func TestNewBranchName(t *testing.T) {
	t.Run("formats as UTC", func(t *testing.T) {
		tm := time.Date(2026, 8, 31, 14, 5, 1, 0, time.UTC)
		got := NewBranchName(tm)
		want := "wip/20260831-140501"
		if got != want {
			t.Errorf("NewBranchName() = %q, want %q", got, want)
		}
	})

	t.Run("converts non-UTC input to UTC", func(t *testing.T) {
		loc := time.FixedZone("TEST", -5*3600) // UTC-5
		tm := time.Date(2026, 8, 31, 9, 5, 1, 0, loc)
		got := NewBranchName(tm)
		want := "wip/20260831-140501"
		if got != want {
			t.Errorf("NewBranchName() = %q, want %q", got, want)
		}
	})
}

func TestWipBranch_Ref(t *testing.T) {
	b := WipBranch{Name: "wip/20260831-140501"}
	got := b.Ref("origin")
	want := "origin/wip/20260831-140501"
	if got != want {
		t.Errorf("Ref() = %q, want %q", got, want)
	}
}

func TestListWipBranches(t *testing.T) {
	newTestRepo(t)
	remoteDir := newBareRemote(t)
	runGit(t, "remote", "add", "origin", remoteDir)

	// main isn't a WIP branch and must not show up in the results.
	if err := Push("origin", "main"); err != nil {
		t.Fatalf("setup Push(main) error = %v", err)
	}

	runGit(t, "checkout", "-b", "wip/20260101-000000")
	runGit(t, "commit", "--allow-empty", "-m", "first wip")
	if err := Push("origin", "wip/20260101-000000"); err != nil {
		t.Fatalf("setup Push() error = %v", err)
	}

	runGit(t, "checkout", "main")
	runGit(t, "checkout", "-b", "wip/20260102-000000")
	runGit(t, "commit", "--allow-empty", "-m", "second wip")
	if err := Push("origin", "wip/20260102-000000"); err != nil {
		t.Fatalf("setup Push() error = %v", err)
	}

	if err := Fetch("origin"); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	branches, err := ListWipBranches("origin")
	if err != nil {
		t.Fatalf("ListWipBranches() error = %v", err)
	}
	if len(branches) != 2 {
		t.Fatalf("ListWipBranches() returned %d branches, want 2: %+v", len(branches), branches)
	}

	subjects := map[string]string{}
	for _, b := range branches {
		subjects[b.Name] = b.Subject
		if b.Author != "Test" {
			t.Errorf("%s Author = %q, want %q", b.Name, b.Author, "Test")
		}
	}
	if subjects["wip/20260101-000000"] != "first wip" {
		t.Errorf("wip/20260101-000000 subject = %q, want %q", subjects["wip/20260101-000000"], "first wip")
	}
	if subjects["wip/20260102-000000"] != "second wip" {
		t.Errorf("wip/20260102-000000 subject = %q, want %q", subjects["wip/20260102-000000"], "second wip")
	}
}

func TestListWipBranches_none(t *testing.T) {
	newTestRepo(t)
	remoteDir := newBareRemote(t)
	runGit(t, "remote", "add", "origin", remoteDir)
	if err := Push("origin", "main"); err != nil {
		t.Fatalf("setup Push() error = %v", err)
	}
	if err := Fetch("origin"); err != nil {
		t.Fatalf("Fetch() error = %v", err)
	}

	branches, err := ListWipBranches("origin")
	if err != nil {
		t.Fatalf("ListWipBranches() error = %v", err)
	}
	if branches != nil {
		t.Errorf("ListWipBranches() = %v, want nil", branches)
	}
}

func TestUniqueBranchName(t *testing.T) {
	newTestRepo(t)

	t.Run("no collision", func(t *testing.T) {
		got := UniqueBranchName("wip/20260831-140501")
		want := "wip/20260831-140501"
		if got != want {
			t.Errorf("UniqueBranchName() = %q, want %q", got, want)
		}
	})

	t.Run("collision appends a numeric suffix", func(t *testing.T) {
		runGit(t, "branch", "wip/taken")
		got := UniqueBranchName("wip/taken")
		want := "wip/taken-2"
		if got != want {
			t.Errorf("UniqueBranchName() = %q, want %q", got, want)
		}
	})

	t.Run("skips over multiple collisions", func(t *testing.T) {
		runGit(t, "branch", "wip/multi")
		runGit(t, "branch", "wip/multi-2")
		got := UniqueBranchName("wip/multi")
		want := "wip/multi-3"
		if got != want {
			t.Errorf("UniqueBranchName() = %q, want %q", got, want)
		}
	})
}
