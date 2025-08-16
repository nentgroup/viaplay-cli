// Package cli provides the command-line interface for viaplay-cli.
// It defines all commands, flags, and user interactions for the CLI application.
package cli

import (
	"strings"
)

// parseOwnerRepoArg parses the owner/repo argument in the format "owner/repo"
// and sets the RepoOwner and RepoName fields in the options struct.
func parseOwnerRepoArg(arg string, opts *CreateCommandOptions) {
	// Split by slash, allowing for either owner/repo or just repo formats
	parts := strings.SplitN(arg, "/", 2)

	// If two parts, we have owner/repo format
	if len(parts) == 2 {
		opts.RepoOwner = parts[0]
		opts.RepoName = parts[1]
		return
	}

	// Otherwise, it's just the repo name (owner defaults to config or GitHub username)
	opts.RepoName = parts[0]
	return
}
