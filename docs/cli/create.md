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
- `--public`: Create a public repository (default is private)
- `--team`: Team name for configs
- `--apply-envs`: Apply environment configs from team settings
- `--apply-rulesets`: Apply ruleset configs from team settings
- `--apply-secrets`: Apply secret configs from team settings
- `--secrets`: JSON string with repository-specific secrets
- `--secrets-file`: Path to JSON file with repository-specific secrets
- `--verbose`: Enable verbose output (prints detailed progress and debug info)

### Project-specific Flags

- `--language`: Programming language (go, typescript, etc.)
- `--type`: Project type (service, cli, etc.)
- `--template-source`: Custom template source
- `--output-dir`: Directory to create the project in
- `--binary-name`: Name of the compiled binary (for compiled languages like Go and Rust)
- `--run-hooks`: Enable execution of post-installation hooks defined in the template (default is true)

---

## Examples

```bash
vip create project --name myservice --language go --type service --team myteam --verbose
vip create repo --name myrepo --public --team myteam --verbose
vip create project --name myservice --language rust --type service --binary-name custom-binary
```

---

See `vip create project --help` and `vip create repo --help` for more details.
