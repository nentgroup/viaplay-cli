---
title: viaplay-cli (vip) command overview
description: Short index of vip command groups and which skill pack or docs to use for each.
---

# viaplay-cli (vip): command overview

`vip` scaffolds projects, creates/configures GitHub repositories, and applies team
standards (rulesets, secrets, environments) for Viaplay teams. This is a short index —
prefer the more detailed skill pack listed for each area when actually performing a task.

| Task | Command group | Skill pack |
|---|---|---|
| Scaffold a new project + create its repo | `vip project` | `project-create` |
| Create/configure an existing repo (no scaffolding) | `vip repo` | `repo-manage` |
| Store/retrieve secrets in the local keyring | `vip secrets` | `secrets-manage` |
| Initialise/inspect/pull viaplay-cli config or team config | `vip config` | `config-manage` |
| Explore/test templates, preview or debug hooks | `vip template`, `vip hooks` | `template-explore` |
| Authenticate with GitHub | `vip auth` | _(see below)_ |
| Manage the local template cache (legacy) | `vip cache` | superseded by `vip template` |
| Install these skill packs for an agent | `vip skills` | _(this feature)_ |
| Show CLI version | `vip version` | _(no pack needed)_ |

## Auth (no dedicated pack — small enough to cover here)

```
vip auth login     # GitHub OAuth device flow
vip auth logout    # remove stored token
vip auth status    # show auth status + account
vip auth whoami    # print authenticated username
vip auth token     # print the stored token (sensitive — avoid printing to chat)
```
Authentication is required before any command that talks to GitHub (project/repo
creation, repo apply, team init with GitHub-derived defaults).

## General guardrails across all commands

- Never guess a team, organization, or repository name — confirm with the user first for
  anything that creates or modifies GitHub resources or writes secrets.
- Prefer `--dry-run` (where available) before applying changes the user hasn't explicitly
  confirmed.
- Never print secret/token values into chat unless the user explicitly asks and
  acknowledges they'll be visible in the conversation.
- If unsure which command group applies, run `vip --help` or `vip <group> --help` rather
  than guessing flags.
