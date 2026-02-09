package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
)

// CreateRuleset creates or updates a GitHub repository ruleset
func (ghc *GitHubClient) CreateRuleset(ctx context.Context, owner, repo string,
	ruleset github.RepositoryRuleset,
) error {
	// Create the ruleset
	_, _, err := ghc.client.Repositories.CreateRuleset(ctx, owner, repo, ruleset)
	return err
}
