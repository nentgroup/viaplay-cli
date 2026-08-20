# Template Commands

The `template` command group provides tools for managing local template copies and testing templates.

## Template Management

### `vip template list`

List template copies currently stored locally for reuse.

### `vip template info`

Show details about local template storage, including size and file statistics.

### `vip template update`

Fetch or refresh all local template copies from their configured sources.

### `vip template prune [--days <n>]`

Remove local template copies that have not been used recently.

- `--days <n>`: prune template copies older than the specified number of days (default: `30`)

### `vip template clean`

Remove all local template copies. Prompts for confirmation before deleting.

## Template Test

Test and validate templates locally without creating GitHub repositories or authenticating with GitHub.

### Overview

- **Purpose**: Test templates in isolation during development or CI pipelines
- **Authentication**: No GitHub authentication required
- **Configuration**: Works independently of your Viaplay CLI configuration
- **Output**: Always outputs to a temporary directory

### Usage

```bash
vip template test [flags]
```

### Required Flags

| Flag | Description |
|------|-------------|
| `--template-path` | Local path to a template directory |

### Optional Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--json` | Output results in JSON format for scripting | `false` |
| `--force` | Force refresh of local template copies | `false` |
| `--name` | Project name for template variables | `test-project` |
| `--owner` | Project owner for template variables | `test-owner` |

### Examples

#### Local Template Management

```bash
vip template list
vip template info
vip template update
vip template prune --days 60
vip template clean
```

#### Basic Template Testing

Test a template with all defaults (outputs to a temporary directory):

```bash
# Test a local template and output to a temporary directory
vip template test --template-path ./path/to/my-template
```

The command will output the path to the temporary directory where your files are scaffolded.

#### JSON Output for Scripts

```bash
vip template test \
  --template-path ./path/to/my-template \
  --json
```

This will output a JSON object containing:
- `success`: Boolean indicating if the operation succeeded
- `outputPath`: The path to the output directory
- `templatePath`: The path to the template source
- `error`: Any error message (only present if there was an error)

#### Custom Template Variables

```bash
vip template test \
  --template-path ./path/to/my-template \
  --name awesome-project \
  --owner my-team
```

### Use Cases

#### Local Template Development


```bash
# Edit your template files
vim my-template/files/src/main.go

# Test the template instantly (outputs to temp dir)
vip template test --template-path ./my-template

# Verify the output (use the path printed in the output)
cat /tmp/vip-template-test-20250903-120145-a1b2c3/src/main.go
```

#### CI/CD Pipeline Testing with JSON


```bash
# Example CI step
JSON_OUTPUT=$(vip template test --template-path ./templates/go-service --name test-service --json)

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
vip template test \
  --template-path ./my-template \
  --name custom-name \
  --owner custom-owner
```

### Notes

- The command always creates a unique temporary directory in your OS's standard temp location
- The temporary directory includes a timestamp in the format `vip-template-test-YYYYMMDD-HHMMSS-XXXXX`
- The command outputs the path to the temporary directory for easy access
- JSON output is available for scripting and CI/CD pipeline integration
- Template variables are minimal by default and may need customization for complex templates
