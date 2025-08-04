# Secrets in viaplay-cli

This page explains how secrets are managed, stored, and used in viaplay-cli for teams and repositories.

---

## What are Secrets?

Secrets are sensitive values (such as API keys, tokens, or passwords) that are required by your repositories or CI/CD pipelines. viaplay-cli allows you to securely manage secrets at the team level and inject them into repositories or environments during project creation.

---

## Where are Secrets Stored?

- **Team secrets:**
  - Configured in `~/.config/viaplay/teams/<team>/secrets.json` (or the global secrets file)
  - The actual secret values are always stored securely in your operating system's keyring (e.g., macOS Keychain, Windows Credential Manager, or Linux Secret Service), not in the config files themselves.
  - The config files only reference the secret names and metadata, never the values.

---

## Example Team Secrets File

```json
{
  "secrets": [
    { "name": "GH_TOKEN", "value": "${{ secrets.GH_TOKEN }}", "env": "", "type": "secret" },
    { "name": "STAGING_API_KEY", "value": "${{ secrets.STAGING_API_KEY }}", "env": "staging", "type": "secret" },
    { "name": "DATABASE_URL", "value": "${{ secrets.DATABASE_URL }}", "env": "production", "type": "variable" }
  ]
}
```

- `env`: Target environment (empty string for repository-level, or the name of a GitHub environment)
- `type`: Either `secret` (encrypted) or `variable` (plain variable)

---

## How to Add, List, and Remove Secrets

- Add a secret:
  ```bash
  vip secrets set <name>
  ```
- List secrets:
  ```bash
  vip secrets list
  ```
- Remove a secret:
  ```bash
  vip secrets forget <name>
  ```

---

## Referencing Secrets: The Preferred Way

The preferred and most secure way to reference a secret is using the `${{ secrets.SECRET_KEY }}` syntax. This ensures that the value is resolved securely from your OS keyring at the time of repository creation or configuration.

- `${{ secrets.SECRET_KEY }}` can be used:
  - In the `--repo-secrets` inline flag (as a value or reference)
  - Inside the JSON secrets config (e.g., `secrets.json`)

Example usage in a secrets config:
```json
{
  "secrets": [
    { "name": "API_KEY", "value": "${{ secrets.GH_TOKEN }}", "type": "secret" }
  ]
}
```

Example usage with the CLI flag:
```bash
vip create project --name myservice --repo-secrets '{"secrets":[{"name":"API_KEY","value":"${{ secrets.GH_TOKEN }}","type":"secret"}]}'
```

> **Note:**
> - Do not use `${{ secrets.SECRET_KEY }}` in your CI/CD pipeline YAML files or in templates. It is only resolved by viaplay-cli during repository setup.
> - You cannot reference secrets inside templates.

---

## Best Practices

- Always use the `${{ secrets.SECRET_KEY }}` syntax to reference secrets for maximum security and portability.
- Never commit secret values to version control.
- Use team or global secrets files to configure which secrets are needed, but store the actual values in your OS keyring.
- Rotate secrets regularly and remove unused ones.

---

For more, see the [CLI secrets command reference](cli/secrets.md).
