# Template Commands

The `template` command group provides tools for managing local template copies, testing templates, and inspecting manifest-defined options.

## Template Management

### `vip template list`

List template copies currently stored locally for reuse, along with overall storage statistics (size, file/directory counts, oldest/newest file).

### `vip template update`

Fetch or refresh all local template copies from their configured sources.

### `vip template prune [--days <n>]`

Remove local template copies that have not been used recently.

- `--days <n>`: prune template copies older than the specified number of days (default: `30`)

### `vip template clean`

Remove all local template copies. Prompts for confirmation before deleting.

## Registering Templates

### `vip template add <source>`

Register a template source under `templates.<language>.<type>` and cache it
locally, so it can be used with `vip project create --language <language>
--type <type>`. Registers under your team's config if one is set up
(`--team`/`default_team`), otherwise your personal config.

`<source>` accepts the same formats as `template inspect`/`template test`.
`--language`/`--type` default to the template's manifest metadata.

```bash
vip template add github.com/nentgroup/go-service-template --language go --type service
vip template add ~/my-templates/go-worker   # language/type read from .vip.yaml metadata
vip template add github.com/nentgroup/gecko-template --language go --type worker --team gecko
```

| Flag | Description |
|------|-------------|
| `--language` | Language to register the template under (overrides manifest metadata) |
| `--type` | Project type to register the template under (overrides manifest metadata) |
| `--team` | Team to register the template under (falls back to `default_team`, then personal config) |
| `--force` | Overwrite an existing `templates.<language>.<type>` entry |
| `--skip-cache` | Register without cloning it into the local cache now |

### `vip template remove <language>/<type>`

Remove a `templates.<language>.<type>` entry and its cached clone. Removes
from your team's config if `--team`/`default_team` applies, otherwise your
personal config.

```bash
vip template remove go/worker
vip template remove go/worker --team gecko
vip template remove go/worker --keep-cache
```

| Flag | Description |
|------|-------------|
| `--team` | Team to remove the template from (falls back to `default_team`, then personal config) |
| `--keep-cache` | Do not remove the locally cached clone for this language/type |

## Template Test

Test and validate templates locally without creating GitHub repositories or authenticating with GitHub.

### Overview

- **Purpose**: Test templates in isolation during development or CI pipelines
- **Authentication**: No GitHub authentication required
- **Configuration**: Works independently of your Viaplay CLI configuration
- **Output**: Always outputs to a temporary directory

### Usage

```bash
vip template test <source> [flags]
```

### Required Arguments

| Argument | Description |
|------|-------------|
| `<source>` | Template source to test — a GitHub address (e.g. `github.com/owner/repo`, `https://github.com/owner/repo`, or `github@owner/repo[@ref]`), a local path, or an explicit `local@`/`url@`/`git@` source. |

### Optional Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--template-path` | *(deprecated)* Equivalent to `<source>`; kept for backwards compatibility | none |
| `--json` | Output results in JSON format for scripting | `false` |
| `--force` | *(deprecated, no-op)* Templates are always fetched fresh | `false` |
| `--name` | Project name for template variables | `test-project` |
| `--owner` | Project owner for template variables | `test-owner` |
| `--set` | Set a template manifest option or variable using `key=value` (repeatable) | none |
| `--no-input` | Do not prompt for manifest options/variables; use `--set` values/defaults | `false` |
| `--allow-template-hooks` | Allow execution of manifest `hooks.post` commands while testing | `false` |

### Examples

#### Local Template Management

```bash
vip template list
vip template update
vip template prune --days 60
vip template clean
```

#### Basic Template Testing

Test a template with all defaults (outputs to a temporary directory):

```bash
# Test a local template and output to a temporary directory
vip template test ./path/to/my-template
```

The command will output the path to the temporary directory where your files are scaffolded.

#### JSON Output for Scripts

```bash
vip template test ./path/to/my-template --json
```

This will output a JSON object containing:
- `success`: Boolean indicating if the operation succeeded
- `outputPath`: The path to the output directory
- `templatePath`: The path to the template source
- `error`: Any error message (only present if there was an error)

#### Custom Template Variables

```bash
vip template test ./path/to/my-template \
  --name awesome-project \
  --owner my-team
```

### Use Cases

#### Local Template Development


```bash
# Edit your template files
vim my-template/files/src/main.go

# Test the template instantly (outputs to temp dir)
vip template test ./my-template

# Verify the output (use the path printed in the output)
cat /tmp/vip-template-test-20250903-120145-a1b2c3/src/main.go
```

#### CI/CD Pipeline Testing with JSON


```bash
# Example CI step
JSON_OUTPUT=$(vip template test ./templates/go-service --name test-service --json)

# Extract the output directory using jq
OUTPUT_DIR=$(echo "$JSON_OUTPUT" | jq -r .outputPath)

# Validate output structure
if [ ! -f "$OUTPUT_DIR/main.go" ]; then
  echo "Template validation failed: missing main.go"
  exit 1
fi
```

#### Debugging Template Variables


```bash
vip template test ./my-template \
  --name custom-name \
  --owner custom-owner
```

### Notes

- The command always creates a unique temporary directory in your OS's standard temp location
- The temporary directory includes a timestamp in the format `vip-template-test-YYYYMMDD-HHMMSS-XXXXX`
- The command outputs the path to the temporary directory for easy access
- JSON output is available for scripting and CI/CD pipeline integration
- Template variables are minimal by default and may need customization for complex templates
- If the template defines a `.vip.yaml` manifest, `template test` prompts
  for its options and variables interactively, unless overridden with `--set` and/or `--no-input`.
  See [Interactive Templates (Manifest)](../templates.md#interactive-templates-manifest) for
  details on manifest structure, options vs. variables, and validation.
- Manifest `hooks.post` commands are only executed when `--allow-template-hooks`
  is explicitly set.
- When hooks are detected, vip asks for an interactive double confirmation
  before executing them. With `--no-input`, hook execution is refused.

### `vip template inspect <source>`

Inspect a template and list the options/variables its manifest defines, without
scaffolding anything. Useful for discovering which `--set key=value` flags a
template supports before running `project create` or `template test`.

`<source>` is a positional argument and accepts the same template source
formats as `project create` and `template test` (GitHub address, `github@owner/repo[@ref]`, local path, etc.).

```bash
vip template inspect ./path/to/my-template
vip template inspect github@nentgroup/go-service-template --json
```

| Flag | Description |
|------|-------------|
| `--force` | Force refresh of local template copies |
| `--json` | Output the manifest as JSON |

If the template has no `.vip.yaml`, the command reports that it has no
configurable options. If the manifest is invalid (see
[Manifest Validation](../templates.md#manifest-validation)), the command still
shows the manifest but lists the validation issues alongside it — fix these
before the template can be used with `project create`/`template test`.

### Setting options with `--set` and `--no-input`

Both `vip project create` and `vip template test` accept:

- `--set key=value` (repeatable) — pre-set a manifest option or variable, skipping
  its prompt; invalid values (per `validate.pattern`) are rejected immediately
- `--no-input` — never prompt; any manifest option/variable not covered by `--set`
  uses its declared default (validated the same way)

```bash
vip project create myorg/my-service \
  --template-source github@nentgroup/go-service-template \
  --no-input --set sqs=true --set sns=false --set shortName=my-service
```
