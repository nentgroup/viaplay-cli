---
title: Manage viaplay-cli configuration and team config
description: Initialise, inspect, pull, edit, and validate global and team configuration.
---

# viaplay-cli: configuration management

Use `vip config` to set up and maintain viaplay-cli's own configuration: the main
config file, personal (per-user) config, team config folders (environments, rulesets,
secrets definitions, template overrides), and an optional shared read-only config source
repository that teams pull from.

## When to use this

Trigger this workflow when the user asks to:
- "set up viaplay-cli for the first time" / "initialize vip config"
- "scaffold a new team's config"
- "what's my default team/organization"
- "pull the latest shared team config / hooks"
- "edit the team config" / "where is the team config stored"
- "validate my config"

## Workflow

### First-time setup
```
vip config init                                   # interactive org/team selection
vip config init --team <team> --organization <org> # non-interactive
```
Requires GitHub auth first (`vip auth login`). This creates the main config file,
personal config, and (if team+org given) scaffolds team config, auto-configuring the
shared source repo if the `<org>/vip-shared-configs` convention exists.

### Scaffold just a team config folder (no GitHub auth required)
```
vip config init team <name> --organization <org> [--override]
```

### Configure the shared, read-only config source
```
vip config init source --organization <org> [--repository <url>] [--branch <name>] [--root <path>]
```

### Pull shared config
```
vip config pull                       # default team + hooks
vip config pull team <name> --organization <org>
vip config pull hooks
vip config pull all
```

### Inspect / edit
```
vip config get [key]                  # e.g. vip config get default_team
vip config paths                      # show config/teams/cache paths
vip config path <team|user|hooks|templates>
vip config edit [main|team|user|hooks|templates]
```

### Validate
```
vip config validate                   # active config + default team, if set
vip config validate --all-teams
```

## Guardrails

- Never run `vip config init --override` (or team/source variants) without confirming
  with the user first — it replaces existing generated files.
- Confirm `--organization` explicitly rather than guessing; several subcommands fall back
  to `default_organization`, which may not match the org the user actually means for this
  task.
- When scaffolding or pulling team config, tell the user the resulting path
  (`vip config path team --team <team>`) so they know where to review changes before
  committing them to a shared config repo.
