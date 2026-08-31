package cmd

import (
	"fmt"

	"github.com/cli/go-gh/v2/pkg/auth"
)

// checkAuth confirms gh has an active token before we attempt any network
// operation, so a missing login surfaces as a clear message instead of a
// cryptic git push/fetch permission error.
func checkAuth() error {
	host, _ := auth.DefaultHost()
	if token, _ := auth.TokenForHost(host); token == "" {
		return fmt.Errorf("not logged in to %s — run `gh auth login` first", host)
	}
	return nil
}
