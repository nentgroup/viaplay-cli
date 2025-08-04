package gh

import (
	"context"

	"github.com/google/go-github/v74/github"
	"golang.org/x/oauth2"
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
