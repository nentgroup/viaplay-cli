# Skills Command

The `skills` command group installs embedded "skill" packs that teach AI coding
agents (Claude Code, GitHub Copilot, Cursor, Codex CLI, ...) how to drive
viaplay-cli workflows — for example, creating a project/repo for a team with
the right template and standards — instead of the agent guessing or copying
files by hand.

Skill packs are markdown files embedded in the `vip` binary. Installing a pack
writes a `SKILL.md` file into a `vip-<pack>/` directory under the skills
directory the target agent discovers, following the open
[Agent Skills standard](https://agentskills.io) that all supported agents
implement natively.

---

## Subcommands & Flags

### `vip skills list`

List the available skill packs and the AI coding agents supported as install
targets.

### `vip skills install`

Install one or more skill packs for one or more agents.

**Flags:**
- `--agent`: Agent(s) to install for (repeatable or comma-separated), e.g. `claude,copilot,cursor`
- `--all-agents`: Install for all supported agents
- `--pack`: Skill pack(s) to install (defaults to all packs)
- `--global`: Install to the agent's global/user directory instead of project-local
- `--target-dir`: Explicit directory to install into, overriding scope resolution
- `--force`: Overwrite files that already exist
- `--dry-run`: Show what would be installed without writing files

Examples:
```
vip skills install --agent claude
vip skills install --all-agents --global
vip skills install --agent copilot --pack project-create
vip skills install --agent cursor --target-dir ./custom/skills-dir
```

---

## Install locations

Each pack is installed as `vip-<pack>/SKILL.md` under the agent's skills
directory:

| Agent   | Project-local (default) | Global (`--global`) |
|---------|--------------------------|----------------------|
| claude  | `.claude/skills/vip-<pack>/SKILL.md`  | `~/.claude/skills/vip-<pack>/SKILL.md` |
| copilot | `.agents/skills/vip-<pack>/SKILL.md`  | `~/.agents/skills/vip-<pack>/SKILL.md` |
| cursor  | `.agents/skills/vip-<pack>/SKILL.md`  | `~/.agents/skills/vip-<pack>/SKILL.md` |
| codex   | `.agents/skills/vip-<pack>/SKILL.md`  | `~/.agents/skills/vip-<pack>/SKILL.md` |

Copilot, Cursor, and Codex CLI all discover skills from the shared
`.agents/skills` (project) / `~/.agents/skills` (global) convention, so
installing once for any of them satisfies the others too — `vip skills
install --all-agents` writes that shared file only once and reports the rest
as already installed.

Existing files are left untouched unless `--force` is passed. Use `--dry-run`
to preview the resolved paths before writing anything.

---

## Available packs

- `overview` — short index of all `vip` command groups, pointing to the more detailed
  pack for each area (safe default when you don't know which pack applies yet).
- `project-create` — teaches an agent the end-to-end workflow for scaffolding
  a project and GitHub repository with `vip project create`/`vip repo apply`,
  including which inputs to confirm with the user first and which guardrails
  to respect (never guessing team/repo names, never skipping confirmation on
  destructive actions).
- `repo-manage` — create repos without scaffolding, and apply environments/
  rulesets/secrets/variables to existing repos with `vip repo create`/`vip repo apply`.
- `secrets-manage` — store/retrieve/list/remove secrets in the local system keyring
  with `vip secrets`, and guardrails against leaking secret values into chat.
- `config-manage` — initialise, inspect, pull, edit, and validate global and team
  configuration with `vip config`.
- `template-explore` — list/inspect/test local template copies and preview or
  validate post-install hooks with `vip template` and `vip hooks`, without creating
  a real project or GitHub repository.

Run `vip skills list` for the up-to-date list shipped with your installed
version of `vip`.
