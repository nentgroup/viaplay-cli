# repo Command

The `vip repo` command is used to manage GitHub repositories without scaffolding code.

---

## Subcommands

- `vip repo create` — Create a GitHub repository without code scaffolding.

## Repo Create Command

### Usage

```bash
vip repo create <org/repo-name> [flags]
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

If both flags are specified, `--private` takes precedence.

---

## Repository Location

The command takes one argument in the format `org/repo-name` or `username/repo-name`:
- `org/repo-name` creates the repository in the specified organization
- `username/repo-name` creates the repository in your personal GitHub account

---

See `vip repo create --help` for more details.

