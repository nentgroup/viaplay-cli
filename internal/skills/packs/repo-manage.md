---
title: Apply team standards to an existing GitHub repo
description: Apply environments, rulesets, secrets, and variables to a repository that already exists.
---

# viaplay-cli: repo management

Use `vip repo` to create or configure GitHub repositories directly, without scaffolding
project code. This is the right tool when the repository already exists (or should be
created without a template) and you need to create it and/or bring it in line with team
standards (environments, rulesets, secrets, variables).

## When to use this

Trigger this workflow when the user asks to:
- "create a repo (no scaffolding) for team X"
- "apply team standards/rulesets/secrets/environments to an existing repo"
- "sync repo Y with team X's config"
- "add/update one secret or variable on repo Y"

If the user instead wants a new project scaffolded from a template, use the
`project-create` skill pack (`vip project create`) instead — it wraps repo creation
plus scaffolding plus standards application in one step.

## Required inputs

Confirm before running (ask if missing or ambiguous):
- **repo** — `[owner/]<repo-name>` (owner defaults to `default_organization` if omitted)
- **team** — which team's config to apply (falls back to `default_team`, but confirm
  explicitly for anything that writes to GitHub)
- **scope** — which of `envs,rulesets,secrets,repo-secrets` to apply (default: all)

## Workflow

### Create a repo without scaffolding
```
vip repo create [owner/]<repo-name> --team <team> [--private|--public] [--description "..."]
```
Add `--apply-envs`, `--apply-rulesets`, `--apply-secrets` to also apply team standards
during creation.

### Apply team standards to an existing repo
```
vip repo apply [owner/]<repo-name> --team <team> [--only <scopes>] [--skip <scopes>] [--dry-run]
```
- **Always run with `--dry-run` first** when the user hasn't explicitly asked for changes
  to be applied immediately, and show them the plan before re-running without `--dry-run`.
- Use `--only envs,rulesets,secrets,repo-secrets` (comma-separated) to limit scope, or
  `--skip` to exclude specific scopes.

### Apply a single secret or variable
```
vip repo apply secret [owner/]<repo-name> <NAME> --team <team> [--env <environment>] [--dry-run]
vip repo apply variable [owner/]<repo-name> <NAME> --team <team> [--env <environment>] [--dry-run]
```
- Without `--from`/`--value`, resolution order is: team config first, then a matching
  keyring entry with the same name.
- `--from keyring|env|file|stdin` plus `--value` explicitly selects the source — use this
  when the user tells you exactly where the value should come from.
- Omit `--env` for a repository-level secret/variable; set `--env <name>` to scope it to
  a GitHub environment.

## Guardrails

- Never write secrets/variables to a repo without the user confirming the target repo and
  scope (repo vs environment) first.
- Prefer `--dry-run` before any apply operation that the user hasn't explicitly asked to
  execute immediately.
- Never print the value of a secret back to the user unless they explicitly ask to see it,
  and even then, prefer confirming this is intended (secrets are typically not meant to be
  echoed into chat transcripts).
- If `--team` is omitted, do not silently assume `default_team` for anything that writes
  to GitHub — confirm the resolved team with the user first.
