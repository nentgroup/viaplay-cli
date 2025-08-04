package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
)

// SetBranchProtection sets branch protection rules for a given repo/branch using go-github.
func SetBranchProtection(ctx context.Context, client *github.Client, owner, repo, branch string, protectionRequest *github.ProtectionRequest) error {
	_, _, err := client.Repositories.UpdateBranchProtection(ctx, owner, repo, branch, protectionRequest)
	return err
}

// CreateRuleset creates a ruleset for a repository or organisation.
func (ghc *GitHubClient) CreateRuleset(owner, repo string, ruleset github.RepositoryRuleset) error {
	_, _, err := ghc.client.Repositories.CreateRuleset(ghc.ctx, owner, repo, ruleset)
	return err
}
