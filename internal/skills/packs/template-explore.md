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
- "register/add this template so I can use it with project create" / "remove that template mapping"

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

vip template add <source> [--language <l>] [--type <t>] [--team <t>] [--force] [--skip-cache]
vip template remove <language>/<type> [--team <t>] [--keep-cache]
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

`template add` also uses an ephemeral clone (same as `inspect`) to read the template's
manifest, then registers `templates.<language>.<type>.source`. If a team is configured
(`--team`, or `default_team`) and has a config directory set up, it registers there instead;
otherwise it writes to the user's personal `~/.config/viaplay/config.yaml`. `--language`/
`--type` are read from the manifest's `metadata.language`/`metadata.type` fields if present;
pass the flags explicitly when the template has none, or to override. It fails if the mapping
already exists unless `--force` is given. It also warms the local template cache immediately
so the template is ready for `project create` right away — pass `--skip-cache` to skip this
(local sources are never cached either way). `template remove <language>/<type>` undoes this
the same team-aware way (`--team`/`default_team`, falling back to personal config) and also
removes the cached clone by default — pass `--keep-cache` to leave it in place.

`vip` validates a manifest's structure (unique/non-empty option and variable keys, a
recognised `options[].type` of `bool`/`boolean`/`select`, non-empty `select` choices, and a
compilable `validate.pattern` regex). `template inspect` still shows the manifest but lists
any issues found; `template add` refuses to register an invalid manifest; `project create`/
`template test` fail fast with the issues before scaffolding starts. When debugging a template
that isn't behaving as expected, always run `template inspect` first — validation issues
explain most unexpected behavior (e.g. a typo'd `type` silently becoming a free-text prompt).

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
