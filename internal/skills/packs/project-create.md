---
title: Create a vip project or repo
description: Scaffold a new project and GitHub repository with vip, applying team standards.
---

# vip: create project/repo

Use the `vip` CLI to scaffold a new project and/or create and configure its GitHub
repository. Prefer this workflow over manually creating repositories or copying files by hand.

## When to use this

Trigger this workflow when the user asks to:
- "create a new repo/project for team X"
- "scaffold a Go/Node/... project for team X"
- "set up a new service/repo with our standard rulesets/secrets/environments"

## Required inputs

Before running any command, confirm (ask if missing or ambiguous):
- **team** — which team owns this repo (maps to team config under `~/.config/viaplay`). If the
  user has **no team** (a personal/one-off repo), see the "No team" guardrail below — do not
  silently let it fall back to `default_team`.
- **name** — repository/project name
- **language/type** — template to use, e.g. `go/service`, `node/api` (run `vip template list` if unsure)
- **owner** — GitHub org or user the repo should be created under (defaults to team's organization)
- **visibility** — private or public (falls back to team/default config)

Never guess the repository name or team silently — always confirm with the user first.

## Workflow

1. **Check auth**: `vip auth status`. If not authenticated, run `vip auth login` first.
2. **Discover template** (if language/type unknown): `vip template list`. If the template the
   user wants isn't registered yet (not in `vip template list`'s config), register it first with
   `vip template add <source> [--language <l> --type <t>]` — see the `template-explore` skill.
3. **Discover required template options (mandatory before every non-interactive run):**
   `vip template inspect <language>/<type> --json`
   This lists every `--set key=value` variable the template's manifest (`.vip.yaml`) exposes,
   including which are required and their defaults. Do not skip this step and rely on trial and
   error — templates commonly define custom variables (e.g. a service's `shortName`) that are
   invisible in `vip template list`/`vip hooks list` and will only surface as a runtime failure
   if you don't check first. Ask the user for values for any required option with no default.
4. **Create the project + repo** (repo name is a positional argument):
   ```
   vip project create [<owner>/]<name> --team <team> --language <language> --type <type> \
     [--public|--private] --apply-envs --apply-rulesets --apply-secrets --no-input \
     [--set key=value ...]
   ```
   This single command scaffolds the project locally, creates the GitHub repository (unless
   `--no-repo`), applies team environments/rulesets/secrets when `--apply-envs`/
   `--apply-rulesets`/`--apply-secrets` are enabled (or their `default_*`/`apply_*` config
   equivalents — **these default to `true`** in a fresh `vip` config, see the "No team" guardrail),
   and runs post-install hooks (unless `--no-hooks`).

   **Booleans need an explicit `=false` to disable:** `--apply-envs`, `--apply-rulesets`, and
   `--apply-secrets` are boolean flags. Passing the bare flag (or nothing, since they default to
   `true` from config) enables them; to disable one you must write `--apply-envs=false` etc.
   (space-separated `--apply-envs false` will NOT work with Cobra boolean flags).

   **Always pass `--no-input`** so the command never blocks waiting for terminal input. Combine
   it with the `--set key=value` values gathered in step 3. If `--no-input` is set and a required
   option has no default and no matching `--set`, the command fails with "missing required
   template option" for that one option at a time — re-check step 3's full option list rather
   than fixing them one failed run at a time.
5. **If the repo already exists** and you only need to apply team standards to it, use
   `vip repo apply <owner>/<repo> --team <team>` instead of `vip project create`.
6. **Report the summary**: the command prints a step-by-step summary (scaffold, repo creation,
   environments, rulesets, secrets, hooks) and the created repository URL. Relay any warnings
   or partial failures to the user — do not silently swallow them.

## Guardrails

- Never pass `--cleanup-on-error=false` unless the user explicitly asks to keep partial state
  after a failure.
- Never run destructive commands (e.g. deleting a repo) without explicit user confirmation.
- If `--team` is omitted, do not assume `default_team` silently for destructive/creation
  operations — confirm the resolved team with the user first.
- **No team (personal/one-off repo):** a fresh `vip` config ships with `apply_envs`,
  `apply_rulesets`, and `apply_secrets` all defaulting to `true`. If the user has no team, `vip
  project create` will still silently resolve `--team` to `default_team` and apply that team's
  environments/rulesets/secrets unless you explicitly pass
  `--apply-envs=false --apply-rulesets=false --apply-secrets=false`. Always do this when the user
  has no team, rather than letting settings apply from a team they didn't ask for.
- Always pass `--no-input` when running non-interactively, and always run `vip template inspect`
  first (step 3) to gather every required `--set key=value` upfront — don't discover them by
  trial and error.
- **Failed runs can leave an empty output directory behind.** Scaffolding creates the output
  directory before resolving template options, so if `--no-input` fails on a missing required
  option, the (empty) project directory is not automatically removed by `--cleanup-on-error`
  (cleanup only removes it once scaffolding has actually completed). After a failed run, check
  whether `--output-dir` (or `./<name>` by default) exists and is empty, and remove it yourself
  before retrying, so a stale empty directory doesn't cause a "directory already exists" error on
  the next attempt.

## Useful follow-ups

- `vip template add <source>` — register a new template under `templates.<language>.<type>`
  before using it with `--language`/`--type` (see `template-explore` skill for details)
- `vip hooks list <language>/<type>` — preview configured post-install hooks before creating
- `vip config get default_team` / `vip config get default_organization` — check defaults
- `vip config get apply_envs` / `apply_rulesets` / `apply_secrets` — check whether these default
  to `true` in the current config (they do in a fresh install)
- `vip repo apply <owner>/<repo> --team <team> --dry-run` — preview standards before applying

