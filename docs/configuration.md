# Configuration

This section explains how configuration works in viaplay-cli, including global, team, and project-specific settings. Real examples from the internal/config/blueprints directory are provided to illustrate best practices.

---

## Configuration Overview

viaplay-cli uses YAML and JSON configuration files to control default values, paths, team settings, template sources,
rulesets, secrets, and more. There are two main types of configuration:

- **Global config:** `~/.config/viaplay/config.yaml` — User-wide defaults and settings.
- **Team config:** `~/.config/viaplay/teams/<team>/config.yaml` — Team-specific settings, rulesets, and secrets.

---

## Global Config Example (config.yaml)

This file controls the default behavior of viaplay-cli. Key fields include:

- `default_team`: Team name for loading team-specific templates.
- `default_account`: GitHub username or organization.
- `is_org`: Whether the default account is an organization.
- `default_language`, `default_type`, `default_private`: Defaults for new projects.
- `templates`: Maps languages and project types to template sources (GitHub, local, or URL).
- `config_dir`, `teams_dir`, `global_dir`: Directory locations for configs.
- `default_branch`: Default branch for new repos.
- `github_api`: GitHub API endpoint.
- `container_registry`: Default container registry.
- `debug`: Enables verbose logging.

Example:
```yaml
# viaplay-cli Configuration
# This file controls the behaviour of the viaplay-cli tool
# See https://github.com/nentgroup/viaplay-cli for documentation

default_team: ""
default_account: ""
is_org: false
default_language: "go"
default_type: "service"
default_private: true
templates:
  go:
    cli: github@github.com/nentgroup/go-cli-template.git
    lambda: github@github.com/nentgroup/go-lambda-template.git
    package: github@github.com/nentgroup/go-package-template.git
    service: github@github.com/nentgroup/go-service-template.git
  rust:
    http-service: local@/Users/alescole/.config/viaplay/templates_cache/rust/http-service
  typescript:
    service: github@github.com/nentgroup/ts-service-template.git
    lambda: github@github.com/nentgroup/ts-lambda-template.git
config_dir: "~/.config/viaplay"
teams_dir: "~/.config/viaplay/teams"
global_dir: "~/.config/viaplay/global"
default_branch: "main"
github_api: "https://api.github.com"
container_registry: "ecr"
debug: false
```

---

## Environment Example (environment.json)

Defines environment-specific settings, such as deployment policies and reviewers.

```json
{
  "name": "staging",
  "wait_timer": 0,
  "reviewers": [],
  "deployment_branch_policy": {
    "protected_branches": false,
    "custom_branch_policies": true
  }
}
```

---

## Rulesets

Rulesets define branch protection and repository rules for GitHub repositories. You can specify a ruleset file in your config and it will be applied automatically to new repositories. Example ruleset file:

```json
{
  "name": "example-ruleset",
  "target": "branch",
  "enforcement": "active",
  "conditions": {
    "ref_name": {
      "include": ["refs/heads/main"],
      "exclude": []
    }
  },
  "rules": [
    {
      "rule_type": "require_pull_request",
      "parameters": {
        "required_approving_review_count": 1,
        "require_code_owner_review": true,
        "dismiss_stale_reviews_on_push": true,
        "require_last_push_approval": false,
        "allowed_merge_methods": ["squash", "rebase"]
      }
    }
  ]
}
```

- For more details on GitHub rulesets, see the [GitHub Ruleset documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets).
- Reference your ruleset file in your config under the appropriate team or project section.

---

## Post-Installation Hooks

Hooks allow you to run commands or scripts after project creation. They are defined in the template configurations and run automatically unless disabled with the `--skip-hooks` flag during project creation.

### Hook Configuration

Hooks are configured under the `templates` section in your config file, associated with specific language and project type combinations:

```yaml
templates:
  go:
    service:
      source: github@github.com/nentgroup/go-service-template.git
      hooks:
        post:
          install:
            run:
              - "go mod tidy"
              - "go generate ./..."
            scripts:
              - "setup-go-service.sh"
  typescript:
    lambda:
      source: github@github.com/nentgroup/ts-lambda-template.git
      hooks:
        post:
          install:
            run:
              - "npm install"
              - "npm run build"
            scripts:
              - "setup-ts-lambda.sh"
```

### Hook Types

Hooks can be defined as:

- **Commands (`run`)**: Shell commands executed in the project directory
- **Scripts (`scripts`)**: Executable script files that are run in the project context

### Template Variables in Hooks

When hooks are executed, they have access to the same template variables used during project scaffolding. For example, in a shell script hook:

```bash
#!/bin/bash
# This is an example post-install hook script

echo "Project name: {{{ProjectName}}}"
echo "Repository: {{{RepoURL}}}"

# Set up a custom environment based on the project type
if [ "{{{ProjectType}}}" = "service" ]; then
  echo "Setting up service-specific environment..."
fi
```

### Using Hooks

Hooks are automatically executed after a project is scaffolded using the `vip create project` command, unless explicitly disabled:

```bash
# Create a project and execute hooks
vip create project --name myservice --language go --type service

# Create a project but skip executing hooks
vip create project --name myservice --language go --type service --skip-hooks
```

### Locating Hook Scripts

Hook scripts referenced in the `scripts` section are resolved in the following order:

1. Absolute paths are used as-is
2. Relative paths are resolved from the Viaplay CLI hooks directory (`~/.config/viaplay/hooks/`)

To create a new hook script:

1. Create a script file in your hooks directory
2. Make it executable (`chmod +x myhook.sh`)
3. Reference it in your template configuration

### Best Practices for Hooks

- **Keep hooks idempotent**: Hooks should be safe to run multiple times
- **Handle errors gracefully**: Include error checking in your scripts
- **Provide progress feedback**: Output meaningful information during hook execution
- **Keep hooks focused**: Each hook should serve a specific purpose
- **Test hooks thoroughly**: Ensure hooks work across different environments

---

## How to Edit Configuration

- Use `vip config set <key> <value>` to update values.
- Use `vip config get <key>` to view current values.
- Edit YAML or JSON files directly for advanced changes.

---

## Best Practices

- Use these examples as a starting point for your own configs.
- Version team configs in your team repository.
- Use project config for overrides and metadata.
- Keep secrets and rulesets in secure, versioned locations.
- Use environment sections to manage per-environment secrets and settings.

---

For more, see the [config command reference](cli/config.md) and [GitHub Ruleset documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets).
