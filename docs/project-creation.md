# Project and Repository Creation


---

## Overview

viaplay-cli can scaffold new projects from templates and automate the creation of repositories on GitHub. The creation process supports applying organization settings like environments, rulesets, and secrets from team configurations.

---

## How It Works

- **Project Creation:**
  - Scaffolds a new project from a template (with your chosen language and type)
  - Creates a new GitHub repository (in an organization or your personal account)
  - Applies team/org settings (environments, rulesets, secrets)
  - Executes post-installation hooks (if defined and not skipped)

- **Repository Creation:**
  - Creates a new GitHub repository (no code scaffolding)
  - Applies team/org settings (environments, rulesets, secrets)

---

## What Gets Created?

- A new project directory with all files from the template, rendered with your variables (for project creation)
- A new GitHub repository (if enabled), with rulesets and secrets applied

---

## Examples

### Creating a Go Service with Custom Binary Name in an Organization

```bash
vip project create nentgroup/repo-name --language go --type service --team myteam --binary-name custombin
```

This will create a Go service project where the compiled binary will be named `custombin` instead of defaulting to the repository name.

### Creating a TypeScript Web Application in an Organization

```bash
vip project create nentgroup/repo-name --language typescript --type webapp --team myteam
```

### Creating a Project in Your Personal GitHub Account

```bash
vip project create username/repo-name --language go --type library
```

The command takes one argument in the format `[owner/]<repo-name>`. For details on how the owner is resolved, see the [project command reference](cli/project.md) or the [repo command reference](cli/repo.md).

---

## Learn More

- For detailed usage, flags, and subcommands, see the [project command reference](cli/project.md).
- For template details, see [Templates](templates.md).
- For team/org configuration, see [Configuration](configuration.md).

