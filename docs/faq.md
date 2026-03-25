# FAQ — viaplay-cli

## General

**What is viaplay-cli?**

viaplay-cli is a command-line tool for scaffolding projects, managing GitHub repositories, and enforcing team standards (rulesets, secrets, environments) in a secure and automated way.

**Which platforms are supported?**

viaplay-cli works on macOS, Linux, and Windows.

**What languages and project types are supported?**

You can scaffold Go, Rust, TypeScript, and other project types, depending on the templates you configure.

---

## Authentication & GitHub

**How does authentication work?**

viaplay-cli uses GitHub's device flow for authentication. Your token is stored securely in your OS keyring.

**How do I log out?**

```bash
vip auth logout
```

**How do I check if I'm authenticated?**

```bash
vip auth status
vip auth whoami
```

---

## Secrets

**Where are secrets stored?**

Secret values are stored in your operating system's keyring. The config files only reference secret names and metadata.

**How do I add a secret?**

```bash
vip secrets set <name>
```

You will be prompted for the value securely.

**Can I use secrets in templates or CI/CD YAML?**

No. Secrets can only be referenced in config files or via the `--repo-secrets` flag during project/repo creation. They are not available in templates or CI/CD YAML.

---

## Teams & Configuration

**How do I set up a team?**

Use `vip config init --team <team>` to scaffold a team config. Place your team secrets, rulesets, and environments in the corresponding team directory.

**Can I override team settings for a specific project?**

You can provide repository-specific secrets using the `--repo-secrets` flag or a YAML file, but all other settings are 
managed at the team/global level.

---

## Troubleshooting

**I get an error about missing GitHub client ID. What do I do?**

Set the `GITHUB_CLIENT_ID` environment variable or build viaplay-cli with an embedded client ID.

**My secret isn't being injected. Why?**

Make sure the secret exists in your keyring and is referenced correctly in your config or `--repo-secrets` JSON.

**How do I update templates?**

Run `vip cache update` to refresh cached templates.

---

## Comparison to Other Tools

**Why use viaplay-cli instead of tools like Yeoman, Cookiecutter, or Plop?**

While tools like Yeoman, Cookiecutter, and Plop are great for generic project scaffolding, viaplay-cli focuses on teams and organizations that also need:

- **GitHub integration:** Automated repository creation, configuration, and management (rulesets, environments, secrets) via the GitHub API.
- **Team and org standards:** Enforce team-specific rulesets, secrets, and environment configs out of the box.
- **Secure secrets management:** Store and inject secrets using your OS keyring, not just in template files.
- **Unified workflow:** Combine project scaffolding, repository setup, and configuration in a single CLI.
- **YAML/JSON-driven config:** Centralise and version your team/project settings.

If you only need file scaffolding, generic tools may be enough. If you also want automated repo creation, configuration, and team standards, viaplay-cli covers that.

---

For more help, see the [documentation](README.md) or open an issue on [GitHub](https://github.com/nentgroup/viaplay-cli).
