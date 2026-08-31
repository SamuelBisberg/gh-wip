package cmd

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/auth"

	"github.com/SamuelBisberg/gh-wip/pkg/git"
)

// preflight confirms this command can safely talk to git and GitHub: the
// current directory is inside a git repo, and gh has an active login.
func preflight() error {
	if !git.IsInsideRepo() {
		return fmt.Errorf("not a git repository")
	}
	return checkAuth()
}

// checkAuth confirms gh has an active token before we attempt any network
// operation, so a missing login surfaces as a clear message instead of a
// cryptic git push/fetch permission error.
func checkAuth() error {
	host, _ := auth.DefaultHost()
	if token, _ := auth.TokenForHost(host); token == "" {
		return fmt.Errorf("not logged in to %s - run `gh auth login` first", host)
	}
	return nil
}
