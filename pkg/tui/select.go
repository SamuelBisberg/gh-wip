package tui

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/dustin/go-humanize"

	"github.com/SamuelBisberg/gh-wip/pkg/git"
)

// ErrAborted is returned by SelectWipBranch/ConfirmDelete when the user
// cancels the prompt (e.g. Ctrl+C).
var ErrAborted = huh.ErrUserAborted

// SelectWipBranch prompts the user to pick one of the given WIP branches,
// showing each one's age, author, and commit subject. Typing filters the
// list, so this doubles as a fuzzy-ish finder without shelling out to fzf.
func SelectWipBranch(branches []git.WipBranch) (git.WipBranch, error) {
	byName := make(map[string]git.WipBranch, len(branches))
	options := make([]huh.Option[string], 0, len(branches))
	for _, b := range branches {
		byName[b.Name] = b
		label := fmt.Sprintf("%s  —  %s  —  %s (%s)", b.Name, humanize.Time(b.When), b.Subject, b.Author)
		options = append(options, huh.NewOption(label, b.Name))
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
