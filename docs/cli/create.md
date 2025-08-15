# create Command

The `vip create` command is used to create projects and repositories with GitHub integration.

---

## Subcommands

- `vip create project` — Create a new project with scaffolding and a GitHub repository.
- `vip create repo` — Create a GitHub repository without code scaffolding.

---

## Common Flags

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

### Project-specific Flags

- `--language`: Programming language (go, typescript, etc.)
- `--type`: Project type (service, cli, etc.)
- `--template-source`: Custom template source
- `--output-dir`: Directory to create the project in
- `--binary-name`: Name of the compiled binary (for compiled languages like Go and Rust)
- `--no-repo`: Do not create a GitHub repository (only scaffold locally)
- `--no-cache`: Force update of the template cache before scaffolding the project
- `--no-hooks`: Skip execution of post-installation hooks defined in the config file

---

## Examples

```bash
# Create a service project with default settings (private repository by default)
vip create project --name myservice --language go --type service --team myteam --verbose

# Create a public repository
vip create repo --name myrepo --public --team myteam --verbose

# Force a private repository (overrides default_visibility if set to "public")
vip create project --name privateproject --private

# Create a Rust service with custom binary name
vip create project --name myservice --language rust --type service --binary-name custom-binary
```

---

## Repository Visibility

By default, repositories are created with the visibility defined in your `default_visibility` config setting (defaults to "private"). You can override this with:

- `--public` to force a public repository
- `--private` (or `-p`) to force a private repository 

If both flags are specified, `--public` takes precedence.

---

See `vip create project --help` and `vip create repo --help` for more details.
