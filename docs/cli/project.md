# project Command

The `vip project` command is used to manage projects with GitHub integration.

---

## Subcommands

- `vip project create` — Create a new project with scaffolding and a GitHub repository.

---

## Project Create Command

### Usage

```bash
vip project create --name <project-name> [flags]
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
- `--language`: Programming language (go, typescript, etc.)
- `--type`: Project type (service, cli, etc.)
- `--template-source`: Custom template source
- `--output-dir`: Directory to create the project in
- `--binary-name`: Name of the compiled binary (for compiled languages like Go and Rust)
- `--no-repo`: Do not create a GitHub repository (only scaffold locally)
- `--no-cache`: Force update of the template cache before scaffolding the project
- `--no-hooks`: Skip execution of post-installation hooks defined in the config file
- `--verbose`: Enable verbose output (prints detailed progress and debug info)

---

## Examples

```bash
# Create a Go service with default settings (private repository by default)
vip project create --name myservice --language go --type service --team myteam --verbose

# Create a public project
vip project create --name mypublicproject --public --team myteam --verbose

# Force a private project (overrides default_visibility if set to "public")
vip project create --name privateproject --private

# Create a Rust service with custom binary name
vip project create --name myservice --language rust --type service --binary-name custom-binary

# Create a local project without a GitHub repository
vip project create --name local-app --language go --type cli --no-repo
```

---

## Repository Visibility

By default, repositories are created with the visibility defined in your `default_visibility` config setting (defaults to "private"). You can override this with:

- `--public` to force a public repository
- `--private` (or `-p`) to force a private repository 

If both flags are specified, `--public` takes precedence.

---

See `vip project create --help` for more details.
