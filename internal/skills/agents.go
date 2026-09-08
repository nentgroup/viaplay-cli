package skills

import (
	"fmt"
	"os"
	"path/filepath"
)

// InstallScope selects where a skill pack should be installed.
type InstallScope string

// Supported install scopes.
const (
	ScopeProject InstallScope = "project"
	ScopeGlobal  InstallScope = "global"
)

// Agent describes an AI coding agent that can consume an installed skill
// pack. All supported agents implement the open Agent Skills standard
// (https://agentskills.io): a skill is a directory named after the skill,
// containing a SKILL.md file with name/description frontmatter. Agents differ
// only in which root directory they discover skills from.
type Agent struct {
	// ID is the stable identifier used on the command line, e.g. "claude".
	ID string
	// Name is a short human-readable name.
	Name string
	// projectRoot is the skills root directory relative to a project, e.g.
	// ".claude/skills" or ".agents/skills".
	projectRoot string
	// globalHomeRoot is the skills root directory relative to the user's
	// home directory, e.g. ".claude/skills" or ".agents/skills".
	globalHomeRoot string
}

// Agents is the ordered list of all known agent targets. Several agents
// share the same on-disk convention (".agents/skills"), which is
// intentional: installing once for one of them satisfies the others too.
var Agents = []Agent{
	{
		ID:             "claude",
		Name:           "Claude Code",
		projectRoot:    filepath.Join(".claude", "skills"),
		globalHomeRoot: filepath.Join(".claude", "skills"),
	},
	{
		// GitHub Copilot (CLI, cloud agent, VS Code and JetBrains agent
		// mode) discovers project skills under .github/skills, .claude/skills,
		// or .agents/skills, and personal skills under ~/.copilot/skills or
		// ~/.agents/skills. We standardise on the shared .agents/skills
		// convention. See
		// https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/add-skills
		ID:             "copilot",
		Name:           "GitHub Copilot",
		projectRoot:    filepath.Join(".agents", "skills"),
		globalHomeRoot: filepath.Join(".agents", "skills"),
	},
	{
		// Cursor discovers skills under .agents/skills or .cursor/skills
		// (project), and ~/.agents/skills or ~/.cursor/skills (global). We
		// standardise on the shared .agents/skills convention. See
		// https://cursor.com/docs/skills
		ID:             "cursor",
		Name:           "Cursor",
		projectRoot:    filepath.Join(".agents", "skills"),
		globalHomeRoot: filepath.Join(".agents", "skills"),
	},
	{
		// Codex CLI discovers skills under .agents/skills (repo scope) and
		// ~/.agents/skills (user scope). See
		// https://developers.openai.com/codex/skills
		ID:             "codex",
		Name:           "OpenAI Codex CLI",
		projectRoot:    filepath.Join(".agents", "skills"),
		globalHomeRoot: filepath.Join(".agents", "skills"),
	},
}

// FindAgent returns the agent with the given ID, or false if it doesn't exist.
func FindAgent(id string) (Agent, bool) {
	for _, a := range Agents {
		if a.ID == id {
			return a, true
		}
	}
	return Agent{}, false
}

// AgentIDs returns the IDs of all known agents, in declaration order.
func AgentIDs() []string {
	ids := make([]string, 0, len(Agents))
	for _, a := range Agents {
		ids = append(ids, a.ID)
	}
	return ids
}

// ResolvePath resolves the install target path for a pack, given the desired
// scope and an optional explicit override.
func (a Agent) ResolvePath(packID string, scope InstallScope, targetDir string) (string, error) {
	if targetDir != "" {
		return filepath.Join(targetDir, "vip-"+packID, "SKILL.md"), nil
	}

	switch scope {
	case ScopeGlobal:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to resolve home directory: %w", err)
		}
		return filepath.Join(home, a.globalHomeRoot, "vip-"+packID, "SKILL.md"), nil
	default:
		return filepath.Join(a.projectRoot, "vip-"+packID, "SKILL.md"), nil
	}
}
