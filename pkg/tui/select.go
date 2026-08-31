package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"

	"github.com/SamuelBisberg/gh-wip/pkg/git"
)

// ErrAborted is returned by SelectWipBranch/ConfirmDelete/SelectWipBranches/
// ConfirmPermanentDelete when the user cancels the prompt (e.g. Ctrl+C).
var ErrAborted = huh.ErrUserAborted

// wipBranchLabel formats a WIP branch for display in a picker: its name,
// age, commit subject, and author.
func wipBranchLabel(b git.WipBranch) string {
	return fmt.Sprintf("%s  -  %s  -  %s (%s)", b.Name, humanize.Time(b.When), b.Subject, b.Author)
}

// SelectWipBranch prompts the user to pick one of the given WIP branches,
// showing each one's age, author, and commit subject. Typing filters the
// list, so this doubles as a fuzzy-ish finder without shelling out to fzf.
func SelectWipBranch(branches []git.WipBranch) (git.WipBranch, error) {
	byName := make(map[string]git.WipBranch, len(branches))
	options := make([]huh.Option[string], 0, len(branches))
	for _, b := range branches {
		byName[b.Name] = b
		options = append(options, huh.NewOption(wipBranchLabel(b), b.Name))
	}

	var chosen string
	err := huh.NewSelect[string]().
		Title("Select a WIP branch to pull").
		Options(options...).
		Filtering(true).
		Height(min(len(options)+2, 15)).
		Value(&chosen).
		Run()
	if err != nil {
		return git.WipBranch{}, err
	}
	return byName[chosen], nil
}

// SelectWipBranches prompts the user to check off any number of the given
// WIP branches, for bulk operations like cleanup. Space toggles a branch,
// enter submits the selection; press "/" to filter first if the list is
// long. Returns an empty slice, not an error, if the user submits without
// checking anything.
//
// Unlike SelectWipBranch, filtering does NOT start active: a checklist is
// driven by arrow keys and space, and huh only lets space toggle a checkbox
// while its filter input isn't focused, so starting filtered (as
// SelectWipBranch does for its fuzzy-finder feel) would silently eat every
// space keystroke as filter text instead of checking anything.
func SelectWipBranches(branches []git.WipBranch) ([]git.WipBranch, error) {
	byName := make(map[string]git.WipBranch, len(branches))
	options := make([]huh.Option[string], 0, len(branches))
	for _, b := range branches {
		byName[b.Name] = b
		options = append(options, huh.NewOption(wipBranchLabel(b), b.Name))
	}

	var chosen []string
	err := huh.NewMultiSelect[string]().
		Title("Select WIP branches to delete").
		Options(options...).
		Height(min(len(options)+2, 15)).
		Value(&chosen).
		Run()
	if err != nil {
		return nil, err
	}

	result := make([]git.WipBranch, 0, len(chosen))
	for _, name := range chosen {
		result = append(result, byName[name])
	}
	return result, nil
}

// ConfirmDelete asks whether the remote WIP branch should be deleted now
// that it has been pulled.
func ConfirmDelete(branch string) (bool, error) {
	var yes bool
	err := huh.NewConfirm().
		Title(fmt.Sprintf("Delete remote branch %s now that it's merged?", branch)).
		Affirmative("Delete").
		Negative("Keep").
		Value(&yes).
		Run()
	if err != nil {
		return false, err
	}
	return yes, nil
}

// ConfirmPermanentDelete requires the user to type the literal text "yes"
// before a bulk, unrecoverable deletion proceeds. A typed confirmation
// (rather than a Y/n toggle) makes it much harder to blow through by
// reflexively hitting enter, since deleting several branches at once can't
// be undone the way a single accidental `pull --delete` can be recovered
// from (the local WIP branch is usually still around).
func ConfirmPermanentDelete(branches []git.WipBranch) (bool, error) {
	var input string
	err := huh.NewInput().
		Title(fmt.Sprintf("Type yes to permanently delete %d branch(es), or leave blank to cancel:", len(branches))).
		Value(&input).
		Run()
	if err != nil {
		return false, err
	}
	return input == "yes", nil
}
