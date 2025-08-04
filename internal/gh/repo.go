package gh

import (
	"fmt"

	"github.com/google/go-github/v74/github"
)

// CreateRepo creates a new GitHub repository under the given org (or user if org is empty)
func (ghc *GitHubClient) CreateRepo(repoName, org string, private bool, description string) (string, error) {
	repo := &github.Repository{
		Name:        github.Ptr(repoName),
		Private:     github.Ptr(private),
		Description: github.Ptr(description),
	}
	var createdRepo *github.Repository
	var resp *github.Response
	var err error
	if org != "" {
		createdRepo, resp, err = ghc.client.Repositories.Create(ghc.ctx, org, repo)
	} else {
		createdRepo, resp, err = ghc.client.Repositories.Create(ghc.ctx, "", repo)
	}
	if err != nil {
		return "", fmt.Errorf("failed to create repo: %w", err)
	}
	if resp.StatusCode != 201 {
		return "", fmt.Errorf("GitHub API error: %s", resp.Status)
	}
	return createdRepo.GetHTMLURL(), nil
}

// SetBranchProtection sets branch protection rules for a given repo/branch
// TODO: Implement this using github.ProtectionRequest if needed, or remove if not used
func (ghc *GitHubClient) SetBranchProtection(owner, repo, branch string, rules map[string]interface{}) error {
	return fmt.Errorf("setBranchProtection not implemented with go-github; use ruleset.go for advanced protection")
}
