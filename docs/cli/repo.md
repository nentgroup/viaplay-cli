# repo Command

The `vip repo` command manages GitHub repositories without scaffolding code.

---

## Subcommands

- `vip repo create` — Create a GitHub repository without code scaffolding.
- `vip repo apply` — Apply environments, rulesets, secrets, and variables to an existing repository.
  - Includes `vip repo apply secret` and `vip repo apply variable` for one-off updates.

## Repo Create Command

### Usage

```bash
vip repo create [owner/]<repo-name> [flags]
```

### Flags

- `--description`: Repository description
- `--public`: Create a public repository (overrides --private)
- `--private`, `-p`: Create a private repository (overrides default visibility)
- `--team`: Team name for configs
- `--apply-envs`: Apply environment configs from team settings
- `--apply-rulesets`: Apply ruleset configs from team settings
- `--apply-secrets`: Apply secret configs from team settings
- `--repo-secrets`: JSON string with repository-specific secrets
- `--secrets-file`: Path to JSON file with repository-specific secrets
- `--cleanup-on-error`: Clean up resources on error
- `--verbose`: Enable verbose output (prints detailed progress and debug info)

---

## Examples

```bash
# Create a private repository with default settings
vip repo create nentgroup/myrepo --team platform

# Create a public repository
vip repo create nentgroup/mypublicrepo --public --team frontend

# Force a private repository (overrides default_visibility if set to "public")
vip repo create nentgroup/privaterepo --private

# Create a repository with a specific description
vip repo create nentgroup/myrepo --description "This is my custom repository description"

# Create a repository and apply team secrets
vip repo create nentgroup/myrepo --team platform --apply-secrets

# Create a repository in your personal GitHub account
vip repo create username/myrepo --description "Personal repository"
```

---

## Repository Visibility

By default, repositories are created with the visibility defined in your `default_visibility` config setting (defaults to "private"). You can override this with:

- `--public` to force a public repository
- `--private` (or `-p`) to force a private repository 

If both flags are specified, `--public` takes precedence.

---

## Repository Location

The command takes one argument in the format `[owner/]<repo-name>`:
- `owner/repo-name` creates the repository in the specified organization or user account
- `repo-name` (without owner) uses the default organization from your config, or your personal GitHub account

---

## Repo Apply Command

The `vip repo apply` command applies team configuration to an existing GitHub repository, reusing the same configuration as `repo create` without creating a new repository.

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

## Repo Apply Notes

- Use `vip repo apply` for bulk configuration from `envs/`, `rulesets/`, `secrets.yaml`, and optional repo-specific JSON.
- Use `vip repo apply secret` when you only need to update one secret or variable.
- `--dry-run` shows scope and source resolution without writing to GitHub.
- Secrets and variables use GitHub environment scope when `--env` is provided.

---

See `vip repo create --help` and `vip repo apply --help` for more details.

