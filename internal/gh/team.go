// Package gh provides GitHub API integration for viaplay-cli.
package gh

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
)

// TeamPermission represents the permission level for a team in a repository
type TeamPermission string

const (
	// TeamPermissionPull represents read-only access
	TeamPermissionPull TeamPermission = "pull"
	// TeamPermissionTriage represents triage access (can manage issues and PRs without write access)
	TeamPermissionTriage TeamPermission = "triage"
	// TeamPermissionPush represents write access
	TeamPermissionPush TeamPermission = "push"
	// TeamPermissionMaintain represents maintain access (can manage issues, PRs, and some repository settings)
	TeamPermissionMaintain TeamPermission = "maintain"
	// TeamPermissionAdmin represents admin access (full control of the repository)
	TeamPermissionAdmin TeamPermission = "admin"
)

// AddTeamToRepository adds a team to a repository with the specified permission
func (ghc *GitHubClient) AddTeamToRepository(org, repo, team string, permission TeamPermission) error {
	// Parameter validation
	if org == "" {
		return fmt.Errorf("organization name is required")
	}
	if repo == "" {
		return fmt.Errorf("repository name is required")
	}
	if team == "" {
		return fmt.Errorf("team name is required")
	}

	// GitHub API uses "slug" format for team names in the URL, so we convert spaces to hyphens
	// and make it lowercase to match GitHub's behavior
	teamSlug := strings.ToLower(strings.ReplaceAll(team, " ", "-"))

	// Add team to the repository
	_, err := ghc.client.Teams.AddTeamRepoBySlug(ghc.ctx, org, teamSlug, org, repo, &github.TeamAddTeamRepoOptions{
		Permission: string(permission),
	})
	if err != nil {
		return fmt.Errorf("failed to add team '%s' to repository '%s/%s': %w", team, org, repo, err)
	}

	return nil
}

// GetTeamID retrieves the ID of a team in the specified organization
func (ghc *GitHubClient) GetTeamID(org, team string) (int64, error) {
	// Parameter validation
	if org == "" {
		return 0, fmt.Errorf("organization name is required")
	}
	if team == "" {
		return 0, fmt.Errorf("team name is required")
	}

	// GitHub API uses "slug" format for team names in the URL, so we convert spaces to hyphens
	// and make it lowercase to match GitHub's behavior
	teamSlug := strings.ToLower(strings.ReplaceAll(team, " ", "-"))

	// Get team info
	t, _, err := ghc.client.Teams.GetTeamBySlug(ghc.ctx, org, teamSlug)
	if err != nil {
		return 0, fmt.Errorf("failed to get team '%s' in organization '%s': %w", team, org, err)
	}

	return t.GetID(), nil
}

// ListOrgTeams retrieves all teams in an organization
func (ghc *GitHubClient) ListOrgTeams(org string) ([]*github.Team, error) {
	if org == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Use ListOptions to handle pagination
	opts := &github.ListOptions{
		PerPage: 100,
	}

	var allTeams []*github.Team
	for {
		teams, resp, err := ghc.client.Teams.ListTeams(ghc.ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list teams in organization '%s': %w", org, err)
		}

		allTeams = append(allTeams, teams...)

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTeams, nil
}

// TeamInfo contains basic information about a GitHub team
type TeamInfo struct {
	ID          int64
	Name        string
	Slug        string
	Description string
}
