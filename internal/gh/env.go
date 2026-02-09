package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
)

// CreateEnvironment creates or updates an environment for a repo using go-github
func (ghc *GitHubClient) CreateEnvironment(ctx context.Context, owner, repo, envName string,
	data *github.CreateUpdateEnvironment,
) error {
	_, _, err := ghc.client.Repositories.CreateUpdateEnvironment(ctx, owner, repo, envName, data)
	return err
}

// CreateCustomBranchPolicy creates a new custom branch pattern for environment deployments
func (ghc *GitHubClient) CreateCustomBranchPolicy(ctx context.Context, owner, repo, envName string,
	data *github.DeploymentBranchPolicyRequest,
) error {
	_, _, err := ghc.client.Repositories.CreateDeploymentBranchPolicy(ctx, owner, repo, envName, data)
	if err != nil {
		return err
	}

	return nil
}
