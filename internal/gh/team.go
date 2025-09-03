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
	// and make it lowercase to match GitHub's behaviour
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

// GetTeamID fetches the ID for a team by its name within an organization
func (ghc *GitHubClient) GetTeamID(org, teamName string) (int64, error) {
	// Parameter validation
	if org == "" {
		return 0, fmt.Errorf("organization name is required")
	}
	if teamName == "" {
		return 0, fmt.Errorf("team name is required")
	}

	// GitHub API uses "slug" format for team names in the URL, so we convert spaces to hyphens
	// and make it lowercase to match GitHub's behaviour
	teamSlug := strings.ToLower(strings.ReplaceAll(teamName, " ", "-"))

	// Try to fetch the team information
	team, _, err := ghc.client.Teams.GetTeamBySlug(ghc.ctx, org, teamSlug)
	if err != nil {
		// Check if this is a 404 error and provide a more user-friendly message
		if strings.Contains(err.Error(), "404") {
			return 0, fmt.Errorf("team '%s' not found in organization '%s'", teamName, org)
		}
		return 0, fmt.Errorf("failed to fetch team '%s' from organization '%s': %w", teamName, org, err)
	}

	// Return the team ID
	return team.GetID(), nil
}

// TeamInfo contains basic information about a GitHub team
type TeamInfo struct {
	ID          int64
	Name        string
	Slug        string
	Description string
}

// ListTeams returns a list of teams in the given organization
func (ghc *GitHubClient) ListTeams(org string) ([]*TeamInfo, error) {
	// Parameter validation
	if org == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	var allTeams []*TeamInfo

	opts := &github.ListOptions{
		PerPage: 100, // Maximum number of teams per page
	}

	for {
		teams, resp, err := ghc.client.Teams.ListTeams(ghc.ctx, org, opts)
		if err != nil {
			// Check if this is a 404 error and provide a more user-friendly message
			if strings.Contains(err.Error(), "404") {
				return nil, fmt.Errorf("organization '%s' not found or you don't have access to it", org)
			}
			return nil, fmt.Errorf("failed to list teams for organization '%s': %w", org, err)
		}

		// Convert github.Team objects to our TeamInfo struct
		for _, team := range teams {
			allTeams = append(allTeams, &TeamInfo{
				ID:          team.GetID(),
				Name:        team.GetName(),
				Slug:        team.GetSlug(),
				Description: team.GetDescription(),
			})
		}

		// Break if we've processed all pages
		if resp.NextPage == 0 {
			break
		}

		// Update options for the next page
		opts.Page = resp.NextPage
	}

	return allTeams, nil
}
