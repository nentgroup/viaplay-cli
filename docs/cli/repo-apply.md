# repo apply Command

The `vip repo apply` command applies team configuration to an existing GitHub repository.

---

## Subcommands

- `vip repo apply` — Apply bulk configuration from team or personal config
- `vip repo apply secret` — Apply a single secret to an existing repository
- `vip repo apply variable` — Apply a single variable to an existing repository

## Repo Apply Command

### Usage

```bash
vip repo apply [owner/]<repo-name> [flags]
```

### Flags

- `--team`: Team name for loading configuration (uses `default_team` from config if omitted)
- `--owner`: Repository owner (required if not specified in the repo argument)
- `--only`: Comma-separated list of scopes to apply (`envs,rulesets,secrets,repo-secrets`)
- `--skip`: Comma-separated list of scopes to skip
- `--repo-secrets`: JSON string containing repository-specific secrets or variables
- `--secrets-file`: Path to a JSON file containing repository-specific secrets or variables
- `--dry-run`: Show what would be applied without making changes

### Examples

```bash
# Apply all configured scopes
vip repo apply myorg/myrepo --team platform

# Apply only secrets
vip repo apply myorg/myrepo --only secrets --team platform

# Apply repo-specific secrets only
vip repo apply myorg/myrepo --only repo-secrets --secrets-file secrets.json

# Dry run
vip repo apply myorg/myrepo --dry-run --team platform
```

---

## Repo Apply Secret Command

### Usage

```bash
vip repo apply secret [owner/]<repo-name> <secret-name> [flags]
```

### Flags

- `--team`: Team name for loading configuration
- `--owner`: Repository owner (required if not specified in the repo argument)
- `--env`: Target environment for the secret or variable (leave empty for repository scope)
- `--from`: Direct source: `keyring`, `env`, `file`, or `stdin`
- `--type`: Value type: `secret` or `variable`
- `--value`: Literal value, or source selector when used with `--from env|file|keyring`
- `--dry-run`: Show what would be applied without making changes

### Resolution Order

Without `--from` or `--value`, the command resolves the named entry from team config first, then falls back to a keyring entry with the same name.

### Examples

```bash
# Apply a team-configured secret
vip repo apply secret myorg/myrepo MY_SECRET --team platform

# Apply to a specific environment
vip repo apply secret myorg/myrepo API_KEY --env staging --team platform

# Apply a variable from an environment variable
vip repo apply secret myorg/myrepo SERVICE_URL --type variable --from env --value SERVICE_URL

# Apply a secret from a file path
vip repo apply secret myorg/myrepo TOKEN --from file --value ./token.txt

# Apply a literal one-off value
vip repo apply secret myorg/myrepo TEMP_SECRET --value "temporary-value"

# Apply from stdin
printf %s "super-secret" | vip repo apply secret myorg/myrepo TOKEN --from stdin
```

---

## Repo Apply Variable Command

### Usage

```bash
vip repo apply variable [owner/]<repo-name> <name> [flags]
```

### Flags

- `--team`: Team name for loading configuration
- `--owner`: Repository owner (required if not specified in the repo argument)
- `--env`: Target environment for the variable (leave empty for repository scope)
- `--from`: Direct source: `keyring`, `env`, `file`, or `stdin`
- `--value`: Literal value, or source selector when used with `--from env|file|keyring`
- `--dry-run`: Show what would be applied without making changes

### Resolution Order

Without `--from` or `--value`, the command resolves the named entry from team config first, then falls back to a keyring entry with the same name.

### Examples

```bash
# Apply a team-configured variable
vip repo apply variable myorg/myrepo SERVICE_URL --team platform

# Apply to a specific environment
vip repo apply variable myorg/myrepo API_URL --env staging --team platform

# Apply a variable from an environment variable
vip repo apply variable myorg/myrepo SERVICE_URL --from env --value SERVICE_URL

# Apply a variable from a file path
vip repo apply variable myorg/myrepo BUILD_SHA --from file --value ./build-sha.txt
```

---

## Notes

- Use `vip repo apply` for bulk configuration from `envs/`, `rulesets/`, `secrets.yaml`, and optional repo-specific JSON.
- Use `vip repo apply secret` when you only need to update one secret or variable.
- `--dry-run` shows scope and source resolution without writing to GitHub.
- Secrets and variables use GitHub environment scope when `--env` is provided.
