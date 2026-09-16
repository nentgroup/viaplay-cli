---
title: Refresh a template from an existing repo
description: Compare a repo created from a template against the template and suggest changes that should be applied back to the template.
---

# vip: refresh template from repo

Use this workflow when a template-generated repository has drifted from the template and you want to pull those improvements back into the template itself.

## When to use this

Trigger this workflow when the user asks to:
- "refresh the template from this repo"
- "suggest template updates from an existing repo"
- "sync repo changes back into the template"

## Required inputs

Confirm before running:
- **source repo** — the local path to the repository created from the template
- **target template** — the local path to the template to update

## Workflow

1. Verify the repo and template are compatible:
   - the source repo must have been created from the target template
   - the template should be the same logical template family, not just a similar project
   - compare template identity markers such as manifest metadata, template-specific files, or repository conventions before suggesting changes
2. Compare the repo against the template and identify repo-only additions, template edits, and shared-file divergences.
3. Prefer changes that improve the template for all future projects, not repo-specific customisations.
4. Present the suggested changes as template updates, grouped by file and intent.
5. If the user wants implementation, apply the updates to the template path only, then re-evaluate the diff.

## Guardrails

- Never treat an arbitrary repo/template pair as compatible without checking they were created from the same template lineage.
- Do not copy repo-specific config, secrets, environment files, or per-project overrides into the template.
- Keep changes minimal and broadly reusable so they benefit future scaffolds.
