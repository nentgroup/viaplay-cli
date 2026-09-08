---
title: Explore, test, and debug templates and hooks
description: List/inspect/test local template copies and preview or validate post-install hooks.
---

# viaplay-cli: template & hook exploration

Use `vip template` and `vip hooks` to discover available templates, test a template in
isolation before scaffolding a real project, and preview or validate post-install hooks —
all without creating a GitHub repository or needing GitHub authentication.

## When to use this

Trigger this workflow when the user asks to:
- "what templates are available" / "what options does template X support"
- "test my template changes without creating a real project"
- "why did my post-install hook fail" / "preview the hooks for this template"

## Template commands

```
vip template list                          # locally cached template copies + storage stats
vip template update                        # refresh all cached templates from source
vip template prune --days <n>              # remove copies unused for n days (default 30)
vip template clean                         # remove all cached templates (prompts to confirm)

vip template inspect <source> [--json]     # list manifest options/variables, no scaffolding
vip template inspect github@org/repo --json

vip template test <source> [--name <n>] [--owner <o>] [--json] \
  [--set key=value ...] [--no-input]
```
`<source>` accepts the same formats as `project create`'s `--template-source`: a GitHub
address (`github.com/owner/repo`, `https://github.com/owner/repo`, `github@owner/repo[@ref]`,
or the SSH shorthand `git@owner/repo[@ref]` — no host needed, it expands to GitHub), a local
path, or an explicit `local@`/`url@`/`git@` source. `--template-path` on `template test` is a
deprecated alias for `<source>` kept only for backwards compatibility — always use the
positional form in new usage.

`template inspect` and `template test` always clone remote sources fresh into a throwaway
temporary location for that single invocation — they never read from or write to the shared
template cache (`vip template list`/`update`/`prune`/`clean`), so they're safe to run
repeatedly and can't corrupt or go stale relative to a real `project create` run. `template
test` always scaffolds into a fresh temporary directory and prints its path — it never touches
a real project directory or creates a GitHub repository. Use `--json` when the result needs to
be consumed by a script or CI step.

## Hook commands

```
vip hooks list [<language>/<type>] [--team <team>] [--path <dir>]
vip hooks doctor [<language>/<type>]        # validates rendering + script existence/executability
vip hooks run [<language>/<type>] --path <dir>   # actually runs hooks against a target directory
vip hooks init [--team <team>] [--organization <org>] [--override]  # scaffold hooks/ + sample script
```
`hooks list`/`doctor` render a best-effort preview using the current directory and flags —
it is not a guarantee of the exact runtime resolution `vip project create` will use.

## Guardrails

- Prefer `template test` / `template inspect` over `project create` when the user's goal is
  just to validate or explore a template — don't create a real repository just to check
  template output.
- `hooks run` executes real commands/scripts against the target directory; confirm the
  target path with the user before running, especially if it's not an empty/test directory.
- `template clean` and `hooks init --override` are destructive/overwriting — confirm before
  running non-interactively on the user's behalf.
