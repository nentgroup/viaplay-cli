package gh

import (
	"github.com/google/go-github/v74/github"
)

// CreateEnvironment creates or updates an environment for a repo using go-github
func (ghc *GitHubClient) CreateEnvironment(owner, repo, envName string, data *github.CreateUpdateEnvironment) error {
	_, _, err := ghc.client.Repositories.CreateUpdateEnvironment(ghc.ctx, owner, repo, envName, data)
	return err
}

// CreateCustomBranchPolicy creates a new custom branch pattern for environment deployments
func (ghc *GitHubClient) CreateCustomBranchPolicy(owner, repo, envName string, data *github.DeploymentBranchPolicyRequest) error {
	_, _, err := ghc.client.Repositories.CreateDeploymentBranchPolicy(ghc.ctx, owner, repo, envName, data)
	if err != nil {
		return err
	}

	return nil
}
