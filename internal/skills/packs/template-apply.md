---
title: Apply template changes to an existing repo
description: Compare a template against a repo created from it and suggest updates that should be applied into the repo.
---

# vip: apply template changes to repo

Use this workflow when a template has changed and you want to bring an existing repo created from that template up to date.

## When to use this

Trigger this workflow when the user asks to:
- "apply new template changes to this repo"
- "sync this repo with the latest template"
- "bring the repo up to date with template changes"

## Required inputs

Confirm before running:
- **source template** — the local path to the template containing the new changes
- **target repo** — the local path to the repository created from that template

## Workflow

1. Verify the repo and template are compatible:
   - the target repo must have been created from the source template
   - compare template identity markers such as manifest metadata, template-specific files, or repository conventions before suggesting changes
2. Compare the template against the repo and identify template-only additions, shared-file updates, and repo-specific divergences.
3. Keep repo-local customisations intact unless the template change is clearly intended to supersede them.
4. Present the suggested changes as repo updates, grouped by file and intent.
5. If the user wants implementation, apply the updates to the repo path only, then re-evaluate the diff.

## Guardrails

- Never assume a repo is template-compatible just because it has similar files.
- Do not overwrite repo-specific secrets, environment files, or local overrides without explicit confirmation.
- If the template and repo have diverged substantially, call out the mismatch instead of forcing a merge.
