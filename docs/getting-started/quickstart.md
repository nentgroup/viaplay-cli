# Quick Start

Get up and running with **viaplay-cli** in just a few steps!

---

## 1. Install viaplay-cli

See the [installation guide](/getting-started/installation.md) for details.

---

## 2. Authenticate with GitHub

```
vip auth login
```

---

## 3. Initialize Configuration

```
vip config init
```

This step will:
- Create the necessary configuration files
- Configure default settings based on your GitHub account
- Let you select your organization and team (if applicable)

You can also specify a team directly:

```
vip config init --team myteam
```

---

## 4. Review and Customize Configuration

Review the generated configuration files in `~/.config/viaplay/` and make any necessary adjustments:

- Global settings in `config.yaml`
- Team-specific files in the team directories
- Personal configurations

This ensures your projects follow your team's standards and practices.

For detailed information about configuration options, see the [Configuration documentation](/configuration.md).

---

## 5. Create a New Project

```
vip project create <org/repo-name> --language <lang> --type <type> [options]
```

Example:

```
vip project create nentgroup/my-service --language rust --type service --team gecko
```

Or for your personal account:

```
vip project create username/my-service --language rust --type service
```

The CLI will automatically initialize a Git repository, make an initial commit, and push it to GitHub.

---

For more details, see the [full documentation](../README.md).
