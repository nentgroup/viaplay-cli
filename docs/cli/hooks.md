# Hooks Command

The `hooks` command group helps you inspect, validate, run, and scaffold post-install hooks defined in template configuration.

---

## Subcommands & Flags

### `vip hooks list [<language>/<type>]`

List the configured post-install hook commands and scripts for a template.
The `Preview` column is a best-effort rendering using the current directory,
config defaults, and any override flags you provide. It is not the exact runtime
resolution from `vip project create`.

**Flags:**
- `--language`: Template language (falls back to `default_language`)
- `--type`: Template type (falls back to `default_type`)
- `--path`: Project directory used to derive default template variables (default: current directory)
- `--name`: Project name override for template variable rendering
- `--owner`: Repository owner override for template variable rendering
- `--team`: Team override for template variable rendering
- `--description`: Project description override for template variable rendering

### `vip hooks doctor [<language>/<type>]`

Validate hook rendering and script resolution for a template.

Checks include:
- command rendering
- script existence
- script executability
- script content rendering

Uses the same flags as `vip hooks list`.

### `vip hooks run [<language>/<type>]`

Run post-install hooks for a template against a target directory.

**Flags:**
- `--language`: Template language (falls back to `default_language`)
- `--type`: Template type (falls back to `default_type`)
- `--path`: Target project directory (default: current directory)
- `--name`: Project name override for template variable rendering
- `--owner`: Repository owner override for template variable rendering
- `--team`: Team override for template variable rendering
- `--description`: Project description override for template variable rendering

### `vip hooks init`

Create the global hooks directory and a sample executable script.

**Flags:**
- `--override`: Replace the sample script if it already exists

---

## Examples

```bash
vip hooks init
vip hooks list go/service
vip hooks doctor --language go --type service --name test
vip hooks doctor go/service
vip hooks run go/service --path .
vip hooks run go/service --path ./my-service --name my-service --team gecko
```
