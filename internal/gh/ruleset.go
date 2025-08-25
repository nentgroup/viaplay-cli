package gh

import (
	"github.com/google/go-github/v74/github"
)

// CreateRuleset creates or updates a GitHub repository ruleset
func (ghc *GitHubClient) CreateRuleset(owner, repo string, ruleset github.RepositoryRuleset) error {
	// Create the ruleset
	_, _, err := ghc.client.Repositories.CreateRuleset(ghc.ctx, owner, repo, ruleset)
	return err
}
