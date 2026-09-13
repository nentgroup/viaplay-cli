# Configuration

This page covers how configuration works in `vip`, including global, team, and project-specific settings.

---

## Configuration Overview

`vip` uses YAML configuration files to control default values, paths, team settings, template sources,
rulesets, secrets, and more. There are three main types of configuration:

- **Global config:** `~/.config/viaplay/config.yaml` — User-wide defaults and settings.
- **Organization team configs:** `~/.config/viaplay/orgs/<organization>/<team>/` — Team-specific settings folder containing various configuration files like rulesets, environments, and secrets within an organization.
- **Personal user configs:** `~/.config/viaplay/users/<username>/` — User-specific settings folder containing configuration files for personal repositories.
- **Shared config source:** an optional read-only Git repository used to distribute hooks and team config directories locally via `vip config pull`.

---

## Global Config Example (config.yaml)

This file controls the default behaviour of `vip`:

```yaml
# vip Configuration
# This file controls the behaviour of the vip CLI
# See https://github.com/nentgroup/viaplay-cli for documentation

# -----------------------------------------------
# Default Settings (used when no flags are provided)
# -----------------------------------------------

# Default GitHub organization name for repositories
# This determines which organization-specific settings will be used
default_organization: ""

# Default team name to use when creating repositories
# This determines which team-specific templates to use
default_team: ""

# Default programming language for new projects
# Available options: go, typescript, rust, python
default_language: "go"

# Default project type for new projects
# Available options: service, lambda, cli, package, app
default_type: "service"

# Default repository visibility
# Options: "private" or "public"
# Use --public flag to override and create public repositories
default_visibility: "private"

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
# - GitHub repo: "git@github.com:<owner>/<repo>.git"
# - Local path: "local@/path/to/template"
# - Tarball URL: "url@https://example.com/template.tar.gz"
templates:
  # Go templates
  go:
    service:
      source: "git@github.com:nentgroup/go-service-template.git"
      hooks:
        post:
          install:
            cmd:
              - "echo 'Go service template installed successfully'"
    # Add more Go project types as needed (cli, lambda, package, etc.)

  # Node templates
  node:
    service:
      source: git@github.com:nentgroup/node-service-template.git

  # Add more language templates as needed
  # rust:
  #   service:
  #     source: git@github.com:your-org/rust-service-template.git


# -----------------------------------------------
# Shared Configuration Source
# -----------------------------------------------

# Optional read-only repository used to distribute shared team configs and hooks
config_source:
  repository: "git@github.com:nentgroup/vip-shared-configs.git"
  branch: "main"
  root: "."


# -----------------------------------------------
# Directory Configuration
# -----------------------------------------------

# Base directory for all configuration files
# Default: ~/.config/viaplay
config_dir: "~/.config/viaplay"

# Directory for organization-specific configurations
# Default: ~/.config/viaplay/orgs
orgs_dir: "~/.config/viaplay/orgs"

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

## Shared Config Source

You can keep shared hooks and team configuration in a dedicated Git repository
and pull it into your local viaplay config.

### Configuring the source

```bash
vip config init source --organization nentgroup
```

This looks for the conventional repository `nentgroup/vip-shared-configs`. You
can also configure an explicit repository URL:

```bash
vip config init source --repository git@github.com:nentgroup/vip-shared-configs.git
```

### Expected repository layout

The shared source is read-only from `vip` and can contain any of these paths:

```text
hooks/
teams/<team>/
orgs/<organization>/teams/<team>/
```

### Pulling shared config

```bash
vip config pull
vip config pull team gecko --organization nentgroup
vip config pull hooks
vip config pull all
```

- `vip config pull` pulls shared hooks and your `default_team` when configured.
- `vip config pull team <name>` pulls one team config.
- `vip config pull hooks` pulls only shared hooks.
- `vip config pull all` pulls hooks and every shared team config directory found in the source repo.

During `vip config init`, if an organization is selected and the conventional
shared config repository exists, `vip` configures it automatically and pulls
hooks plus the selected/default team config.

---

## Environments

Environments define deployment targets (such as staging, production) with their own protection rules, reviewers, and secrets. Environment files are placed in your team configuration directory under `envs/`.

For a full example and detailed options, see [Environments Configuration](envs.md).

---

## Rulesets

Rulesets define branch protection and repository rules for GitHub repositories. You can specify ruleset files in your team config directory and they will be applied automatically to new repositories.

For a full example and detailed options, see [Rulesets and Configuration](rulesets.md).

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

Hooks are automatically executed after a project is scaffolded using the `vip project create` command, unless explicitly disabled:

```bash
# Create a project and execute hooks
vip project create nentgroup/myservice --language go --type service --team myteam

# Create a project but skip executing hooks
vip project create nentgroup/myservice --language go --type service --team myteam --no-hooks
```

You can also inspect and manage hook configuration directly:

```bash
vip hooks init
vip hooks list go/service
vip hooks doctor go/service
vip hooks run go/service --path ./myservice
```

Teams can also override template sources and hooks with a `config.yaml` file inside the team config directory. When present, the team config overlays the main config for matching `<language>/<type>` entries, and **always takes precedence** over your personal config for those entries — see [Team Config Overrides](#team-config-overrides) below.

### Team Config Overrides

When a `<language>/<type>` entry exists in both your personal `~/.config/viaplay/config.yaml` and a team's `config.yaml`, the team's value always wins once merged (via `--team`/`default_team`). This lets a team enforce a standard template regardless of individual config.

`vip template add` is team-aware: if a team is configured (via `--team`, or `default_team`) and has a config directory set up (even without a `config.yaml` yet, e.g. via `vip config init team`/`vip config pull`), the entry is registered directly in that team's `config.yaml`. If no team is configured, or the team has no config directory at all, the entry is registered in your personal config as usual (see [`vip template add`](cli/template.md#vip-template-add-source)).

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
- **Log progress**: Output useful information during hook execution
- **Keep hooks focused**: Each hook should serve a specific purpose
- **Test hooks**: Make sure hooks work across different environments

---

## How to Edit Configuration

- Use `vip config edit` to open the main config file in your configured editor.
- Use `vip config edit team --team <name> --organization <org>` to jump straight into a team config directory.
- Use `vip config get <key>` to view current values.
- Use `vip config init source` to configure the shared config repository once.
- Use `vip config pull` to refresh local shared hooks and team config from that source.
- Use `vip config path <team|user|hooks|templates>` to print one resolved config path for scripts or quick navigation.
- Use `vip config validate` to check the main config file and any selected team or user config directories before applying them.
- Edit YAML files directly for advanced changes.
- Use `vip config init team <name> --organization <org>` to scaffold a new team configuration folder with starter files.

---

## Best Practices

- Use these examples as a starting point for your own configs.
- Version team configs in your team repository.
- Use project config for overrides and metadata.
- Keep secrets in your OS keyring; version ruleset and environment configs in your team repository.
- Use environment sections to manage per-environment secrets and settings.

---

For more, see the [config command reference](cli/config.md) and [GitHub Ruleset documentation](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-rulesets/about-rulesets).
