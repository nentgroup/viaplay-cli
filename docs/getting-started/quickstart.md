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

## 3. Create a New Project

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

---

## 4. Configure Your Project

Edit the generated config files or use CLI commands to update settings, secrets, and rulesets.

---

## 5. Push to GitHub

```
git push origin main
```

---

For more details, see the [full documentation](../README.md).
