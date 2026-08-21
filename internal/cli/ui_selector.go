// Package cli provides the command-line interface for viaplay-cli.
// This file contains interactive wizard prompts for configuration.
package cli

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/manifoldco/promptui"

	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
)

// runSelect shows an inline filterable list prompt and returns the selected value.
// Returns ("", nil) when the user cancels with Ctrl+C.
func runSelect(label string, items []string) (string, error) {
	searcher := func(input string, index int) bool {
		return strings.Contains(strings.ToLower(items[index]), strings.ToLower(input))
	}

	prompt := promptui.Select{
		Label:             label,
		Items:             items,
		Size:              10,
		Searcher:          searcher,
		StartInSearchMode: len(items) > 5,
		HideSelected:      false,
	}

	_, result, err := prompt.Run()
	if err != nil {
		if errors.Is(err, promptui.ErrInterrupt) || errors.Is(err, promptui.ErrEOF) {
			return "", nil // user cancelled → treat as skip
		}
		return "", fmt.Errorf("prompt failed: %w", err)
	}

	return result, nil
}

// SelectOrganizationWithBubbles presents an inline filterable prompt of the user's organizations.
func SelectOrganizationWithBubbles(ctx context.Context, ghClient *gh.GitHubClient) (string, error) {
	output.ProcessingMessage("Fetching your organizations...")

	orgs, err := ghClient.GetUserOrganizations(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to fetch organizations: %w", err)
	}

	if len(orgs) == 0 {
		return "", fmt.Errorf("you don't belong to any organizations")
	}

	// Single org — confirm automatically, no prompt needed.
	if len(orgs) == 1 {
		login := orgs[0].GetLogin()
		fmt.Printf("Organization: %s  (only one found)\n", output.Bold(login))
		return login, nil
	}

	logins := make([]string, len(orgs))
	for i, org := range orgs {
		logins[i] = org.GetLogin()
	}

	selected, err := runSelect("Organization", logins)
	if err != nil {
		return "", err
	}

	if selected != "" {
		fmt.Printf("Organization: %s\n", output.Bold(selected))
	}

	return selected, nil
}

// SelectTeamWithBubbles presents an inline filterable prompt of teams in an organization.
func SelectTeamWithBubbles(ctx context.Context, ghClient *gh.GitHubClient, orgName string) (string, error) {
	if orgName == "" {
		return "", fmt.Errorf("organization name is required")
	}

	output.ProcessingMessage(fmt.Sprintf("Fetching teams in %s...", orgName))

	teams, err := ghClient.ListTeams(ctx, orgName)
	if err != nil {
		return "", fmt.Errorf("failed to fetch teams: %w", err)
	}

	if len(teams) == 0 {
		output.InfoMessage(fmt.Sprintf("No teams found in %s — skipping team selection", orgName))
		return "", nil
	}

	// Single team — confirm automatically, no prompt needed.
	if len(teams) == 1 {
		slug := teams[0].Slug
		fmt.Printf("Team:         %s  (only one found)\n", output.Bold(slug))
		return slug, nil
	}

	slugs := make([]string, len(teams))
	for i, t := range teams {
		slugs[i] = t.Slug
	}

	label := fmt.Sprintf("Team [%s]", orgName)
	selected, err := runSelect(label, slugs)
	if err != nil {
		return "", err
	}

	if selected != "" {
		fmt.Printf("Team:         %s\n", output.Bold(selected))
	}

	return selected, nil
}
