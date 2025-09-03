# Template Commands

The `template` command group provides tools for working with templates in the Viaplay CLI. These commands are designed to help developers work with project templates more efficiently.

## Template Test

The `template test` command helps template developers validate and test their templates without needing to create GitHub repositories or authenticate with GitHub. This streamlines the template development process and enables automated testing of templates.

### Overview

- **Purpose**: Test templates in isolation during development or CI pipelines
- **Authentication**: No GitHub authentication required
- **Configuration**: Works independently of your Viaplay CLI configuration
- **Output**: Always outputs to a temporary directory for simple testing

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
| `--force` | Force refresh of template cache | `false` |
| `--name` | Project name for template variables | `test-project` |
| `--owner` | Project owner for template variables | `test-owner` |

### Examples

#### Basic Template Testing

Test a template with all defaults (outputs to a temporary directory):

```bash
# Test a local template and output to a temporary directory
vip template test --template-path ./path/to/my-template
```

The command will output the path to the temporary directory where your files are scaffolded.

#### Using JSON Output for Scripts

When you want to use the command in scripts and CI pipelines:

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

#### Customizing Template Variables

This example shows how to set custom template variables:

```bash
vip template test \
  --template-path ./path/to/my-template \
  --name awesome-project \
  --owner my-team
```

### Use Cases

#### Local Template Development

When developing or modifying templates, you can use this command for quick testing:

```bash
# Edit your template files
vim my-template/files/src/main.go

# Test the template instantly (outputs to temp dir)
vip template test --template-path ./my-template

# Verify the output (use the path printed in the output)
cat /tmp/vip-template-test-20250903-120145-a1b2c3/src/main.go
```

#### CI/CD Pipeline Testing with JSON

Include template validation in your CI/CD pipeline using JSON output:

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

Test how different template variables affect the output:

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
