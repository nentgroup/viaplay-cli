# Rulesets and Configuration

Rulesets in viaplay-cli allow you to enforce policies and automate repository settings, such as branch protection, required reviews, and more. These rulesets are based on GitHub's native ruleset feature and must be defined as JSON files.

---

## What is a Ruleset?

A ruleset is a JSON file that defines repository rules, such as:
- Branch protection (e.g., require PR reviews, status checks)
- Commit signing requirements
- Push restrictions
- Environment protection rules

> **Note:** Only JSON format is supported for rulesets. YAML is not supported.

---

## Using Rulesets

Rulesets can be applied during project creation or later using the CLI:

```
vip create --repo-name my-service --apply-rulesets
```

Or manually:

```
vip config apply-rulesets
```

---

## Configuring Rulesets

- Place your ruleset files in the config directory (e.g., `~/.config/viaplay/teams/<team>/rulesets/`).
- Reference them in your config or pass via CLI options.
- Edit rulesets to match your team's policies.

---

## Example Ruleset (GitHub JSON)

```json
{
  "branch_protection": {
    "required_status_checks": ["ci/test", "lint"],
    "enforce_admins": true,
    "required_pull_request_reviews": 1
  },
  "require_signed_commits": true
}
```
