# How Templates Work

viaplay-cli uses project templates to scaffold new repositories. Templates can be local directories or remote Git repositories. During project creation, template files are copied and variables are replaced with values you provide.

---

## Template Sources

- **Remote Git:** Use a Git URL (e.g., `git@github.com:nentgroup/rust-http-service-template-2.git`).
- **Local Directory:** Use a local path for custom templates.

---

## Template Structure

A template typically contains:
- Project files (README, source code, configs)
- Placeholders for variables (see [Template Variables](templates.md))

---

## Raw Files (no rendering)

Add the `.raw` suffix to any template file you want copied verbatim. The file is copied as-is and the `.raw` suffix is stripped in the generated project.

Example: `template.go.tmpl.raw` is copied to `template.go.tmpl` without any rendering.

---

## Using a Template

Specify the template source when creating a project:

```bash
# Basic usage
vip project create my-service --language rust --type service --template-source git@github.com:nentgroup/rust-http-service-template-2.git

# With explicit organization/owner
vip project create myorg/my-service --language rust --type service --template-source git@github.com:nentgroup/rust-http-service-template-2.git
```

Or use a local template:

```bash
# Basic usage
vip project create my-service --language rust --type service --template-source ~/my-templates/rust-service

# With explicit organization/owner
vip project create myorg/my-service --language rust --type service --template-source ~/my-templates/rust-service
```

---

## Customizing Templates

- Add or remove files as needed.
- Use variable placeholders in any file.

---

## Template Caching

Remote templates are cached locally for faster reuse. Use `vip cache update` to refresh the cache.

---

For more on template variables, see [Template Variables](templates.md).
