# config Command

The `vip config` command manages global, team, and project-specific configuration for viaplay-cli. Configuration files control default values, paths, team settings, and feature toggles.

---

## Subcommands & Flags

### `vip config init`
Initialises the main config file and (optionally) team configs.

**Flags:**
- `--team <team>`: Initialise a config for the specified team (creates `~/.config/viaplay/teams/<team>/config.yaml`).

### `vip config init team <name>`
Scaffold a team configuration folder with starter files for environments, rulesets, and secrets.

**Flags:**
- `--organization`, `-o`: Organization name for the team config folder. Falls back to `default_organization` when configured.
- `--override`: Replace existing starter files if they already exist.

### `vip config get [key]`
Get a config value (or all values if no key is provided).

**Arguments:**
- `key` (optional): The config key to retrieve (e.g., `default_team`).

### `vip config edit [main|team|user|hooks|templates]`
Open the main config file or a resolved config directory in your editor.

When no target is provided, it opens the main config file.

**Flags:**
- `--team`, `-t`: Team name for `edit team` (falls back to `default_team`)
- `--organization`, `-o`: Organization name for `edit team` (falls back to `default_organization`)
- `--user`, `-u`: Username for `edit user`

### `vip config paths`
Show config, teams, and cache paths. No flags or arguments.

### `vip config path <team|user|hooks|templates>`
Print one resolved path for scripting or quick navigation.

**Flags:**
- `--team`, `-t`: Team name for `path team` (falls back to `default_team`)
- `--organization`, `-o`: Organization name for `path team` (falls back to `default_organization`)
- `--user`, `-u`: Username for `path user`

### `vip config validate`
Validate the active config file and optionally team or personal config directories.

By default it validates the active config file, and if `default_team` is configured it also validates that team directory.

**Flags:**
- `--team`, `-t`: Team name to validate (falls back to `default_team`)
- `--organization`, `-o`: Organization name for team validation (falls back to `default_organization`)
- `--user`, `-u`: Username for personal config validation
- `--all-teams`: Validate every discovered team config directory

---

## Example Config File

Below is an example of a real viaplay-cli config file (see `internal/config/blueprints/config.yaml`):

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
    service: github@github.com/nentgroup/go-service-template.git
  node:
    service: github@github.com/nentgroup/node-service-template.git
  # Add more language/type templates as needed
  # rust:
  #   service: local@/path/to/your/rust-service-template

config_dir: "~/.config/viaplay"
teams_dir: "~/.config/viaplay/teams"
global_dir: "~/.config/viaplay/global"
default_branch: "main"
github_api: "https://api.github.com"
container_registry: "ecr"
debug: false
```

---

## Example Usage

```bash
vip config init
vip config init --team myteam
vip config init team myteam --organization nentgroup
vip config get default_account
vip config edit
vip config edit team --team myteam --organization nentgroup
vip config path team --team myteam --organization nentgroup
vip config path hooks
vip config validate
vip config validate --all-teams
vip config paths
```

---

For more details, see the [Rulesets and Configuration](../rulesets.md) section.
