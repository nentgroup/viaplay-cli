# Quick Start


---

## 1. Install vip

See the [installation guide](/getting-started/installation.md) for details.

---

## 2. Authenticate with GitHub

```
vip auth login
```

---

## 3. Initialise Configuration

```
vip config init
```

This step will:
- Create the necessary configuration files
- Configure default settings based on your GitHub account
- Let you select your organization and team (if applicable)

You can also specify a team and organization directly:

```
vip config init --team myteam -o myorg
```

---

## 4. Review and Customise Configuration

Review the generated configuration files in `~/.config/viaplay/` and make any necessary adjustments:

- Global settings in `config.yaml`
- Team-specific files in the team directories
- Personal configurations


For detailed information about configuration options, see the [Configuration documentation](/configuration.md).

---

## 5. Create a New Project

```
vip project create [owner/]<repo-name> --language <lang> --type <type> [options]
```

Example:

```
vip project create nentgroup/my-service --language rust --type service --team gecko
```

Or for your personal account:

```
vip project create username/my-service --language rust --type service
```

The CLI will automatically initialise a Git repository, make an initial commit, and push it to GitHub.

---

For more details, see the [full documentation](../README.md).
