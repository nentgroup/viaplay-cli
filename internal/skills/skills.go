// Package skills provides embedded agent "skill" packs that teach AI coding
// agents (Claude Code, GitHub Copilot, Cursor, etc.) how to drive viaplay-cli
// workflows, and installs them into the appropriate per-agent instruction
// directories.
package skills

import (
	"embed"
	"fmt"
	"sort"
)

//go:embed packs
var packFiles embed.FS

// Pack describes one embedded skill pack.
type Pack struct {
	// ID is the stable identifier used on the command line, e.g. "project-create".
	ID string
	// Title is a short human-readable name.
	Title string
	// Description summarises what the skill teaches the agent to do.
	Description string
	// file is the embedded file name under packs/.
	file string
}

// Packs is the ordered list of all embedded skill packs.
var Packs = []Pack{
	{
		ID:          "overview",
		Title:       "Command overview",
		Description: "Short index of vip command groups and which skill pack or docs to use for each.",
		file:        "overview.md",
	},
	{
		ID:          "project-create",
		Title:       "Create project/repo",
		Description: "Scaffold a project and GitHub repository, applying team standards.",
		file:        "project-create.md",
	},
	{
		ID:          "repo-manage",
		Title:       "Manage existing repos",
		Description: "Create repos without scaffolding, and apply envs/rulesets/secrets/variables.",
		file:        "repo-manage.md",
	},
	{
		ID:          "secrets-manage",
		Title:       "Manage keyring secrets",
		Description: "Store, retrieve, list, and remove secrets in the local system keyring.",
		file:        "secrets-manage.md",
	},
	{
		ID:          "config-manage",
		Title:       "Manage vip configuration",
		Description: "Initialise, inspect, pull, edit, and validate global and team configuration.",
		file:        "config-manage.md",
	},
	{
		ID:          "template-explore",
		Title:       "Explore templates & hooks",
		Description: "List/inspect/test local template copies and preview or validate post-install hooks.",
		file:        "template-explore.md",
	},
}

// Find returns the pack with the given ID, or false if it doesn't exist.
func Find(id string) (Pack, bool) {
	for _, p := range Packs {
		if p.ID == id {
			return p, true
		}
	}
	return Pack{}, false
}

// IDs returns the sorted list of all known pack IDs.
func IDs() []string {
	ids := make([]string, 0, len(Packs))
	for _, p := range Packs {
		ids = append(ids, p.ID)
	}
	sort.Strings(ids)
	return ids
}

// Content returns the raw markdown content of the pack.
func (p Pack) Content() ([]byte, error) {
	data, err := packFiles.ReadFile("packs/" + p.file)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded skill pack %q: %w", p.ID, err)
	}
	return data, nil
}
