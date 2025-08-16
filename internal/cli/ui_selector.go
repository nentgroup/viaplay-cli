// Package cli provides the command-line interface for viaplay-cli.
// This file contains Bubble Tea UI components for interactive selection.
package cli

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/nentgroup/viaplay-cli/internal/gh"
	"github.com/nentgroup/viaplay-cli/internal/output"
	"golang.org/x/term"
)

// UI styling constants
var (
	titleStyle        = lipgloss.NewStyle().Foreground(lipgloss.Color("#1E88E5")).Bold(true)
	selectedItemStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#4CAF50")).Bold(true)
	paginationStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
	helpStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Italic(true)
)

// Item represents a selectable item in a list
type Item struct {
	TitleText       string
	DescriptionText string
	Value           string
}

// FilterValue returns the value used for filtering the list
func (i Item) FilterValue() string { return i.TitleText }

// Title returns the item title for the list
func (i Item) Title() string { return i.TitleText }

// Description returns the item description for the list
func (i Item) Description() string { return i.DescriptionText }

// SelectModel represents a selectable list UI
type SelectModel struct {
	list         list.Model
	selectedItem *Item
	done         bool
	title        string
	err          error
	width        int
	height       int
}

// NewSelector creates a new selector with the given title and items
func NewSelector(title string, items []Item, width, height int) SelectModel {
	// Add a "skip" option as the first item
	skipItem := Item{
		TitleText:       "Skip selection",
		DescriptionText: "Continue without selecting",
		Value:           "",
	}
	allItems := append([]Item{skipItem}, items...)

	// Configure list display
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().Faint(true)
	delegate.ShowDescription = true

	// Convert items to list.Items
	listItems := make([]list.Item, len(allItems))
	for i, item := range allItems {
		listItems[i] = item
	}

	// Create the list
	l := list.New(listItems, delegate, width, height)
	l.Title = title
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.SetShowHelp(true) // Show help text
	l.Styles.HelpStyle = helpStyle
	l.SetFilteringEnabled(true) // Allow filtering

	return SelectModel{
		list:   l,
		title:  title,
		width:  width,
		height: height,
	}
}

// Init initializes the model
func (m SelectModel) Init() tea.Cmd {
	return nil
}

// Update handles UI events
func (m SelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(Item); ok {
				m.selectedItem = &i
				m.done = true
				return m, tea.Quit
			}
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width, msg.Height-4)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI
func (m SelectModel) View() string {
	if m.done {
		if m.selectedItem != nil && m.selectedItem.Value != "" {
			return fmt.Sprintf("Selected: %s\n", selectedItemStyle.Render(m.selectedItem.TitleText))
		}
		return "Selection skipped\n"
	}

	// Create a styled header with navigation instructions
	header := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3D85C6")).
		Padding(0, 1).
		Width(m.width - 2).
		Render("↑/↓: Navigate • Enter: Select • Ctrl+C/q: Cancel • /: Filter")

	return fmt.Sprintf("%s\n%s", header, m.list.View())
}

// GetSelected returns the selected item's value and whether the selection is done
func (m SelectModel) GetSelected() (string, bool) {
	if !m.done || m.selectedItem == nil {
		return "", false
	}
	return m.selectedItem.Value, true
}

// getTerminalSize gets the terminal dimensions
func getTerminalSize() (width, height int) {
	// Default fallback values
	width, height = 80, 20

	// Try to get the actual terminal size
	if w, h, err := term.GetSize(int(os.Stdout.Fd())); err == nil {
		width, height = w, h
	}

	return width, height
}

// RunSelector runs a bubble tea selector and returns the selected value
func RunSelector(title string, items []Item) (string, error) {
	// Get terminal dimensions
	width, height := getTerminalSize()

	// Adjust list height to be appropriate for the terminal
	listHeight := height - 6 // Leave room for title, header and status line
	if listHeight < 10 {
		listHeight = 10 // Minimum reasonable height
	}
	if listHeight > 20 {
		listHeight = 20 // Maximum reasonable height
	}

	// Adjust list width
	listWidth := width - 4 // Leave some margin
	if listWidth < 60 {
		listWidth = 60 // Minimum reasonable width
	}

	model := NewSelector(title, items, listWidth, listHeight)

	// Run the bubbletea program in full screen mode for better rendering
	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return "", fmt.Errorf("error running selection: %w", err)
	}

	if m, ok := finalModel.(SelectModel); ok {
		if selected, done := m.GetSelected(); done {
			return selected, nil
		}
	}

	return "", nil // Return empty string for "Skip" option
}

// slugify converts a string to a URL/path friendly slug (lowercase, hyphens)
func slugify(input string) string {
	// Convert to lowercase
	slug := strings.ToLower(input)

	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")

	// Remove any character that's not alphanumeric or hyphen
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	// Remove leading and trailing hyphens
	slug = strings.Trim(slug, "-")

	return slug
}

// SelectOrganizationWithBubbles presents a bubbletea list of organizations to select from
func SelectOrganizationWithBubbles(ghClient *gh.GitHubClient) (string, error) {
	// Fetch organizations the user belongs to
	orgs, err := ghClient.GetUserOrganizations()
	if err != nil {
		return "", fmt.Errorf("failed to fetch organizations: %w", err)
	}

	if len(orgs) == 0 {
		return "", fmt.Errorf("you don't belong to any organizations")
	}

	// Convert organizations to selectable items
	var items []Item
	for _, org := range orgs {
		description := ""
		if org.Description != nil {
			description = *org.Description
		}

		items = append(items, Item{
			TitleText:       *org.Login,
			DescriptionText: description,
			Value:           slugify(*org.Login),
		})
	}

	// Run the selector
	selected, err := RunSelector("Select an organization", items)
	if err != nil {
		return "", err
	}

	if selected != "" {
		fmt.Printf("Selected organization: %s\n", output.Bold(selected))
	}

	return selected, nil
}

// SelectTeamWithBubbles presents a bubbletea list of teams in an organization
func SelectTeamWithBubbles(ghClient *gh.GitHubClient, orgName string) (string, error) {
	if orgName == "" {
		return "", fmt.Errorf("organization name is required")
	}

	// Fetch teams in the organization
	teams, err := ghClient.ListOrgTeams(orgName)
	if err != nil {
		return "", fmt.Errorf("failed to fetch teams: %w", err)
	}

	if len(teams) == 0 {
		return "", fmt.Errorf("no teams found in organization %s", orgName)
	}

	// Convert teams to selectable items
	var items []Item
	for _, team := range teams {
		desc := ""
		if team.Description != nil && *team.Description != "" {
			desc = *team.Description
		}

		items = append(items, Item{
			TitleText:       *team.Name,
			DescriptionText: desc,
			Value:           slugify(*team.Name),
		})
	}

	// Run the selector
	selected, err := RunSelector(fmt.Sprintf("SELECT a team FROM %S", orgName), items)
	if err != nil {
		return "", err
	}

	if selected != "" {
		fmt.Printf("Selected team: %s\n", output.Bold(selected))
	}

	return selected, nil
}
