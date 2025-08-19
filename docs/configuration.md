# Configuration

This section explains how configuration works in viaplay-cli, including global, team, and project-specific settings. Real examples from the internal/config/blueprints directory are provided to illustrate best practices.

---

## Configuration Overview

viaplay-cli uses YAML configuration files to control default values, paths, team settings, template sources,
rulesets, secrets, and more. There are two main types of configuration:

- **Global config:** `~/.config/viaplay/config.yaml` — User-wide defaults and settings.
- **Team config:** `~/.config/viaplay/teams/<team>/config.yaml` — Team-specific settings, rulesets, and secrets.

---

## Global Config Example (config.yaml)

This file controls the default behavior of viaplay-cli:

```yaml
# viaplay-cli Configuration
# This file controls the behaviour of the viaplay-cli tool
# See https://github.com/nentgroup/viaplay-cli for documentation

# -----------------------------------------------
# Default Settings (used when no flags are provided)
# -----------------------------------------------

# Team name to use for loading configuration templates
# This determines which team-specific templates to use from ~/.config/viaplay/teams/
default_team: ""

# GitHub account/organization name to use for repositories
# For personal repos, use your GitHub username
# For org repos, use the organization name
default_account: ""

# Whether the default_account is an organization (true) or personal account (false)
# This affects repository creation behaviour
is_org: false

# Default programming language for new projects
# Available options: go, typescript, rust, python
default_language: "go"

# Default project type for new projects
# Available options: service, lambda, cli, package, app
default_type: "service"

# Default repository visibility
# When true, repositories will be created as private by default
# Use --private=false flag to override and create public repositories
default_private: true

# -----------------------------------------------
# Default Behavior Settings
# -----------------------------------------------

# Apply team environments to new repositories by default
# When true, environments from team config will be applied without needing --apply-envs flag
apply_envs: true

# Apply team secrets to new repositories by default
# When true, secrets from team config will be applied without needing --apply-secrets flag
apply_secrets: true

# Apply team rulesets to new repositories by default
# When true, rulesets from team config will be applied without needing --apply-rulesets flag
apply_rulesets: true

# Clean up resources on error by default
# When true, any created resources (repos, directories) will be deleted if an error occurs
cleanup_on_error: true

# Skip post-installation hooks by default
# When true, post-installation scripts won't run unless explicitly enabled with --hooks
no_hooks: false

# Skip repository creation by default
# When true, only scaffolds local project without creating GitHub repository
no_repo: false

# Disable caching by default
# When true, templates won't be cached
no_cache: false

# -----------------------------------------------
# Project Templates
# -----------------------------------------------

# Repository templates for each language and project type
# Format: templates.<language>.<type> = "<source>"
# Source can be:
# - GitHub repo: "git@githubc:<owner>/<repo>.git"
# - Local path: "local@/path/to/template"
# - Tarball URL: "url@https://example.com/template.tar.gz"
templates:
  # Go templates
  go:
    cli:
      source: "git@github.com:nentgroup/go-cli-template.git"
    lambda:
      source: "git@github.com:nentgroup/go-lambda-template.git"
    package:
      source: "git@github.com:nentgroup/go-package-template.git"
    service:
      source: "git@github.com:nentgroup/go-service-template.git"
      hooks:
        post:
          install:
            cmd:
              - "echo 'Go service template installed successfully'"

  # Rust templates
  rust:
    service:
        source: git@github.com:nentgroup/rust-service-template.git

  # Node templates
  node:
    service:
      source: git@github.com:nentgroup/node-service-template.git
  # Add more language templates as needed


# -----------------------------------------------
# Directory Configuration
# -----------------------------------------------

# Base directory for all configuration files
# Default: ~/.config/viaplay
config_dir: "~/.config/viaplay"

# Directory for team-specific configurations
# Default: ~/.config/viaplay/teams
teams_dir: "~/.config/viaplay/teams"

# Directory for global configurations (used as fallback if team config not found)
# Default: ~/.config/viaplay/global
global_dir: "~/.config/viaplay/global"

# -----------------------------------------------
# GitHub Configuration
# -----------------------------------------------

# Default branch name for new repositories
default_branch: "main"

# GitHub API endpoint (change for GitHub Enterprise)
# Default: https://api.github.com
github_api: "https://api.github.com"

# -----------------------------------------------
# Build and Deployment Settings
# -----------------------------------------------

# Default container registry for services
# Options: "ecr", "gcr", "dockerhub", etc.
container_registry: "ecr"

# -----------------------------------------------
# Advanced Settings
# -----------------------------------------------

# Debug mode (enables verbose logging)
debug: false
```

---

## Environment Example (environment.yaml)

Defines environment-specific settings, such as deployment policies and reviewers.

```yaml
name: staging
wait_timer: 0
reviewers:
  - type: Team
    id: {{.Org.TeamID}}
deployment_branch_policy:
  protected_branches: false
  custom_branch_policies: true
  branch_patterns:
    - name: main              # This represents a DeploymentBranchPolicyRequest
      type: branch            # Values could be "branch" or "tag"
    - name: "release/*"       # Another pattern example
      type: branch
    - name: "*"
      type: tag

```

---

## Rulesets

Rulesets define branch protection and repository rules for GitHub repositories. You can specify a ruleset file in your config and it will be applied automatically to new repositories. Example ruleset file:

```yaml
# Example GitHub branch ruleset
# This file defines a ruleset for GitHub repositories
# It supports template variables like {{.Org.TeamID}} for dynamic values

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

- For more details on GitHub rulesets, see the [GitHub Ruleset documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets).
- Reference your ruleset file in your config under the appropriate team or project section.

---

## Post-Installation Hooks

Hooks allow you to run commands or scripts after project creation. They are defined in the template configurations and run automatically unless disabled with the `--no-hooks` flag during project creation.

### Hook Configuration

Hooks are configured under the `templates` section in your config file, associated with specific language and project type combinations:

```yaml
templates:
  go:
    service:
      source: github@github.com:nentgroup/go-service-template.git
      hooks:
        post:
          install:
            cmd:
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
            cmd:
              - "npm install"
              - "npm run build"
            scripts:
              - "setup-ts-lambda.sh"
```

### Hook Types

Hooks can be defined as:

- **Commands (`cmd`)**: Shell commands executed in the project directory
- **Scripts (`scripts`)**: Executable script files that are run in the project context

### Template Variables in Hooks

When hooks are executed, they have access to the same template variables used during project scaffolding. For example, in a shell script hook:

```bash
#!/bin/bash
# This is an example post-install hook script

echo "Project name: {{.Project.Name}}"
echo "Repository: {{.Repo.Name}}"

# Set up a custom environment based on the project type
if [ "{{.Project.Type}}" = "service" ]; then
  echo "Setting up service-specific environment..."
fi
```

### Using Hooks

Hooks are automatically executed after a project is scaffolded using the `vip create project` command, unless explicitly disabled:

```bash
# Create a project and execute hooks
vip create project --name myservice --language go --type service

# Create a project but skip executing hooks
vip create project --name myservice --language go --type service --no-hooks
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
- Edit YAML files directly for advanced changes.

---

## Best Practices

- Use these examples as a starting point for your own configs.
- Version team configs in your team repository.
- Use project config for overrides and metadata.
- Keep secrets and rulesets in secure, versioned locations.
- Use environment sections to manage per-environment secrets and settings.

---

For more, see the [config command reference](cli/config.md) and [GitHub Ruleset documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets).
