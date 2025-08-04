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

## Using a Template

Specify the template source when creating a project:

```bash
vip create --repo-name my-service --template-source git@github.com:nentgroup/rust-http-service-template-2.git
```

Or use a local template:

```bash
vip create --repo-name my-service --template-source ~/my-templates/rust-service
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
