// Package gh provides GitHub API integration for viaplay-cli.
package gh

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
)

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
	// and make it lowercase to match GitHub's behavior
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

// ListTeams returns a list of teams in the given organization
func (ghc *GitHubClient) ListTeams(org string) ([]*TeamInfo, error) {
	if org == "" {
		return nil, fmt.Errorf("organization name is required")
	}

	// Fetch all teams from the organization
	opts := &github.ListOptions{
		PerPage: 100, // Maximum number of teams per page
	}

	var allTeams []*TeamInfo

	for {
		teams, resp, err := ghc.client.Teams.ListTeams(ghc.ctx, org, opts)
		if err != nil {
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

// TeamInfo contains basic information about a GitHub team
type TeamInfo struct {
	ID          int64
	Name        string
	Slug        string
	Description string
}
