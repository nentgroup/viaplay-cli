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

// RepositoryExists checks if a repository with the given name already exists
func (g *GitHubClient) RepositoryExists(owner, repo string) (bool, error) {
	// Use the GitHub API to check if the repository exists
	_, resp, err := g.client.Repositories.Get(g.ctx, owner, repo)

	// If we got a 404, the repository doesn't exist
	if resp != nil && resp.StatusCode == 404 {
		return false, nil
	}

	// If we got a different error, something went wrong with the API call
	if err != nil && resp == nil {
		return false, fmt.Errorf("failed to check if repository exists: %w", err)
	}

	// If we get here with a non-nil error but it's not a 404, treat it as a special case
	if err != nil {
		// Check if it's a rate limit error or other specific GitHub API error
		if _, ok := err.(*github.RateLimitError); ok {
			return false, fmt.Errorf("GitHub API rate limit exceeded: %w", err)
		}

		// For authentication errors, check the status code
		if resp != nil && resp.StatusCode == 401 {
			return false, fmt.Errorf("not authorized to access this repository: %w", err)
		}

		// Fall back to checking the status code from the response
		if resp != nil && resp.StatusCode != 404 {
			return true, nil // Repository likely exists but we have limited access
		}

		return false, fmt.Errorf("error checking repository: %w", err)
	}

	// If we got no error, the repository exists
	return true, nil
}
