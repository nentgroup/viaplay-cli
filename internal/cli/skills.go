package cli

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/nentgroup/viaplay-cli/internal/output"
	"github.com/nentgroup/viaplay-cli/internal/skills"
)

// NewSkillsCommand creates the skills command with subcommands.
func NewSkillsCommand() *cobra.Command {
	skillsCmd := &cobra.Command{
		Use:   "skills",
		Short: "Install viaplay-cli skill packs for AI coding agents",
		Long: `Install "skill" packs that teach AI coding agents (Claude Code,
GitHub Copilot, Cursor, Codex CLI, ...) how to drive viaplay-cli workflows,
such as creating a project/repo for a team with the right template and
standards.

Packs are installed as SKILL.md directories following the open Agent Skills
standard (https://agentskills.io), which all supported agents implement
natively.

Examples:
  vip skills list
  vip skills install --agent claude
  vip skills install --all-agents --global
  vip skills install --agent copilot --pack project-create
`,
	}

	skillsCmd.AddCommand(newSkillsListCommand())
	skillsCmd.AddCommand(newSkillsInstallCommand())

	return skillsCmd
}

func newSkillsListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   cmdList,
		Short: "List available skill packs and supported agents",
		Long:  `List the skill packs available to install, and the AI coding agents supported as install targets.`,
		Run: func(cmd *cobra.Command, args []string) {
			output.Section("Skill packs")
			rows := make([][]string, 0, len(skills.Packs))
			for _, p := range skills.Packs {
				rows = append(rows, []string{p.ID, p.Title, p.Description})
			}
			fmt.Print(output.Table([]string{"ID", "Title", "Description"}, rows, 0))

			fmt.Println()
			output.Section("Supported agents")
			agentRows := make([][]string, 0, len(skills.Agents))
			for _, a := range skills.Agents {
				agentRows = append(agentRows, []string{a.ID, a.Name})
			}
			fmt.Print(output.Table([]string{"ID", "Name"}, agentRows, 0))
		},
	}
}

func newSkillsInstallCommand() *cobra.Command {
	var (
		agentIDs  []string
		allAgents bool
		packIDs   []string
		global    bool
		targetDir string
		force     bool
		dryRun    bool
	)

	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install skill packs for one or more AI coding agents",
		Long: `Install one or more skill packs into the SKILL.md skills directory of
one or more AI coding agents.

By default, packs are installed project-locally (e.g. .claude/skills,
.agents/skills). Use --global to install to the agent's global/user-level
directory instead (e.g. ~/.claude/skills, ~/.agents/skills).

Use --target-dir to override path resolution entirely and install directly
into a specific directory.
`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			resolvedAgentIDs, err := resolveSkillsAgentIDs(agentIDs, allAgents)
			if err != nil {
				return err
			}

			resolvedPacks, err := resolveSkillsPackIDs(packIDs)
			if err != nil {
				return err
			}

			scope := skills.ScopeProject
			if global {
				scope = skills.ScopeGlobal
			}
			opts := skills.InstallOptions{
				Scope:     scope,
				TargetDir: targetDir,
				Force:     force,
				DryRun:    dryRun,
			}

			return runSkillsInstall(resolvedAgentIDs, resolvedPacks, opts)
		},
	}

	cmd.Flags().StringSliceVar(&agentIDs, "agent", nil, "Agent(s) to install for (repeatable or comma-separated), e.g. claude,copilot,cursor")
	cmd.Flags().BoolVar(&allAgents, "all-agents", false, "Install for all supported agents")
	cmd.Flags().StringSliceVar(&packIDs, "pack", nil, "Skill pack(s) to install (defaults to all)")
	cmd.Flags().BoolVar(&global, "global", false, "Install to the agent's global/user directory instead of project-local")
	cmd.Flags().StringVar(&targetDir, "target-dir", "", "Explicit directory to install into, overriding scope resolution")
	cmd.Flags().BoolVar(&force, "force", false, "Overwrite files that already exist")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be installed without writing files")

	return cmd
}

func resolveSkillsAgentIDs(agentIDs []string, allAgents bool) ([]string, error) {
	if allAgents {
		return skills.AgentIDs(), nil
	}
	if len(agentIDs) == 0 {
		return nil, fmt.Errorf("no agent specified — use --agent <id> (%s) or --all-agents", strings.Join(skills.AgentIDs(), ", "))
	}
	for _, id := range agentIDs {
		if _, ok := skills.FindAgent(id); !ok {
			return nil, fmt.Errorf("unknown agent %q — supported agents: %s", id, strings.Join(skills.AgentIDs(), ", "))
		}
	}
	return agentIDs, nil
}

func resolveSkillsPackIDs(packIDs []string) ([]skills.Pack, error) {
	if len(packIDs) == 0 {
		return skills.Packs, nil
	}
	resolved := make([]skills.Pack, 0, len(packIDs))
	for _, id := range packIDs {
		p, ok := skills.Find(id)
		if !ok {
			return nil, fmt.Errorf("unknown skill pack %q — available packs: %s", id, strings.Join(skills.IDs(), ", "))
		}
		resolved = append(resolved, p)
	}
	return resolved, nil
}

func runSkillsInstall(agentIDs []string, packs []skills.Pack, opts skills.InstallOptions) error {
	for _, agentID := range agentIDs {
		agent, _ := skills.FindAgent(agentID)
		for _, pack := range packs {
			result, err := skills.Install(agent, pack, opts)
			if err != nil {
				output.WarningMessage(fmt.Sprintf("%s/%s: %v", agent.ID, pack.ID, err))
				continue
			}

			printSkillsInstallResult(result)
		}
	}

	return nil
}

func printSkillsInstallResult(result skills.InstallResult) {
	switch {
	case result.Skipped:
		output.InfoMessage(fmt.Sprintf("%s/%s: %s (%s)", result.Agent.ID, result.Pack.ID, result.Path, result.Note))
	case result.Written:
		output.SuccessMessage(fmt.Sprintf("%s/%s: installed to %s", result.Agent.ID, result.Pack.ID, output.Bold(result.Path)))
	default:
		output.InfoMessage(fmt.Sprintf("%s/%s: %s (%s)", result.Agent.ID, result.Pack.ID, result.Path, result.Note))
	}
}
