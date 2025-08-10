package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
)

// GitHub API scope constants
const (
	// ScopeRepo grants read/write access to code, commit statuses, etc.
	ScopeRepo = "repo"

	// ScopeDeleteRepo allows deletion of repositories
	ScopeDeleteRepo = "delete_repo"

	// ScopeReadOrg grants read-only access to organisation data
	ScopeReadOrg = "read:org"

	// DefaultScopes combines the default scopes needed for viaplay-cli
	DefaultScopes = ScopeRepo + " " + ScopeDeleteRepo + " " + ScopeReadOrg
)

type GitHubClient struct {
	ctx    context.Context
	client *github.Client
}

// NewGitHubClient creates a new GitHubClient
func NewGitHubClient(token string) *GitHubClient {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	client := github.NewClient(oauth2.NewClient(ctx, ts))
	return &GitHubClient{
		ctx:    ctx,
		client: client,
	}
}

// GetAuthenticatedUser returns the username of the authenticated GitHub user
func (ghc *GitHubClient) GetAuthenticatedUser() (string, error) {
	user, _, err := ghc.client.Users.Get(ghc.ctx, "")
	if err != nil {
		return "", err
	}
	return user.GetLogin(), nil
}
