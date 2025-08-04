# config Command

The `vip config` command manages global, team, and project-specific configuration for viaplay-cli. Configuration files control default values, paths, team settings, and feature toggles.

---

## Subcommands & Flags

### `vip config init`
Initializes the main config file and (optionally) team configs.

**Flags:**
- `--team <team>`: Initialize a config for the specified team (creates `~/.config/viaplay/teams/<team>/config.yaml`).

### `vip config get [key]`
Get a config value (or all values if no key is provided).

**Arguments:**
- `key` (optional): The config key to retrieve (e.g., `default_team`).

### `vip config set <key> <value>`
Set a config value.

**Arguments:**
- `key`: The config key to set (e.g., `default_team`).
- `value`: The value to assign.

### `vip config paths`
Show config, teams, and cache paths. No flags or arguments.

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

## Example Usage

```bash
vip config init
vip config init --team myteam
vip config get default_account
vip config set default_team myteam
vip config paths
```

---

## Best Practices

- Use `vip config set` to update values without editing YAML manually.
- Keep team configs versioned for consistency.
- Use `vip config paths` to troubleshoot or verify config locations.

---

For more details, see the [Rulesets and Configuration](../rulesets.md) section.
