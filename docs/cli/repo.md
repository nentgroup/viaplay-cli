Let's rethink the config init and try to simplyfy it. As we are logged we can be sure to use the logged in user as the default user in github.username. Actualy we dont even need the username in the config at all because we already know who is logged in.
We can also fetch all the orgs the user has access to right?
so we can do gonfig init --team team-1 --org org1 and it will init the global config, team configs for the team-1 in the org-1 and also personal configs for the logged in user. What do you think?# repo Command

The `vip repo` command is used to manage GitHub repositories without scaffolding code.

---

## Subcommands

- `vip repo create` — Create a GitHub repository without code scaffolding.

## Repo Create Command

### Usage

```bash
vip repo create --name <repo-name> [flags]
```

### Flags

- `--name` (required): Repository name
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
vip repo create --name myrepo --team platform

# Create a public repository
vip repo create --name mypublicrepo --public --team frontend

# Force a private repository (overrides default_visibility if set to "public")
vip repo create --name privaterepo --private

# Create a repository with a specific description
vip repo create --name myrepo --description "This is my custom repository description"

# Create a repository and apply team secrets
vip repo create --name myrepo --team platform --apply-secrets
```

---

## Repository Visibility

By default, repositories are created with the visibility defined in your `default_visibility` config setting (defaults to "private"). You can override this with:

- `--public` to force a public repository
- `--private` (or `-p`) to force a private repository 

If both flags are specified, `--public` takes precedence.

---

See `vip repo create --help` for more details.
---

