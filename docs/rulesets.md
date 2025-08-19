# Rulesets and Configuration

Rulesets in viaplay-cli allow you to enforce policies and automate repository settings, such as branch protection, required reviews, and more. These rulesets are based on GitHub's native ruleset feature and must be defined as JSON files.

---

## What is a Ruleset?

A ruleset is a yaml file that defines repository rules, such as:
- Branch protection (e.g., require PR reviews, status checks)
- Commit signing requirements
- Push restrictions
- Environment protection rules


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

## Example Ruleset 

```yaml
name: branch-protection
target: branch
enforcement: active

# You can use template variables for dynamic values
bypass_actors:
  - actor_id: {{.Org.TeamID}}
    actor_type: Team
    bypass_mode: always

conditions:
  ref_name:
    include:
      - refs/heads/main
    exclude: []

rules:
  - type: require_pull_request
    parameters:
      required_approving_review_count: 1
      require_code_owner_review: true
      dismiss_stale_reviews_on_push: true
      require_last_push_approval: false
      allowed_merge_methods:
        - squash
        - rebase
```
