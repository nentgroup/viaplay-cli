package gh

import (
	"github.com/google/go-github/v74/github"
)

// CreateEnvironment creates or updates an environment for a repo using go-github
func (ghc *GitHubClient) CreateEnvironment(owner, repo, envName string) error {
	_, _, err := ghc.client.Repositories.CreateUpdateEnvironment(ghc.ctx, owner, repo, envName,
		&github.CreateUpdateEnvironment{})
	return err
}
