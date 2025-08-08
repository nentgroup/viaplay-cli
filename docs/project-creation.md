# Project and Repository Creation

This page explains the workflow and concepts behind project and repository creation with viaplay-cli.

---

## Overview

viaplay-cli can scaffold new projects from templates and automate the creation of repositories on GitHub. The creation process supports applying organization settings like environments, rulesets, and secrets from team configurations.

---

## How It Works

- **Project Creation:**
  - Scaffolds a new project from a template (with your chosen language and type)
  - Creates a new GitHub repository
  - Applies team/org settings (environments, rulesets, secrets)
  - Executes post-installation hooks (if defined and not skipped)
  - Optionally clones the project locally and sets up initial config files

- **Repository Creation:**
  - Creates a new GitHub repository (no code scaffolding)
  - Applies team/org settings (environments, rulesets, secrets)

---

## What Gets Created?

- A new project directory with all files from the templa  te, rendered with your variables (for project creation)
- A new GitHub repository (if enabled), with rulesets and secrets applied

---

## Examples

### Creating a Go Service with Custom Binary Name

```bash
vip create project --name my-awesome-service --language go --type service --team myteam --binary-name custombin
```

This will create a Go service project where the compiled binary will be named `custombin` instead of defaulting to the repository name.

### Creating a TypeScript Web Application

```bash
vip create project --name my-web-app --language typescript --type webapp --team myteam
```

---

## Learn More

- For detailed usage, flags, and subcommands, see the [create command reference](cli/create.md).
- For template details, see [Templates](templates.md).
- For team/org configuration, see [Configuration](configuration.md).

---

This page is a high-level overview. For exact CLI usage, always refer to the CLI Reference.
