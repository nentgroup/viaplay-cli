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

- `env`: Target environment where the secret/variable will be created
  - When empty or omitted: Secret is created at the repository level (available to all workflows)
  - When specified (e.g., "staging", "production"): Secret is created for that specific GitHub environment only
- `type`: Either `secret` (encrypted) or `variable` (plain variable)

### Scoping Secrets to Environments

The `env` parameter lets you target secrets to specific deployment environments:

```json
{
  "secrets": [
    { "name": "API_KEY", "value": "${{ secrets.DEV_API_KEY }}", "env": "development", "type": "secret" },
    { "name": "API_KEY", "value": "${{ secrets.PROD_API_KEY }}", "env": "production", "type": "secret" },
    { "name": "GLOBAL_SECRET", "value": "${{ secrets.GLOBAL_SECRET }}", "type": "secret" }
  ]
}
```

In this example:
- The same secret name `API_KEY` points to different values in different environments
- `GLOBAL_SECRET` is available to all workflows since it has no environment restriction

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

## Using Template Variables in Secrets

You can reference template variables in your secret values using Go template syntax. This is particularly useful for creating dynamic values that incorporate project information:

```json
{
  "secrets": [
    { 
      "name": "RESOURCE_PREFIX", 
      "value": "{{.Project.Name}}-resources",
      "type": "variable"
    },
    {
      "name": "STACK_NAME",
      "value": "Dev-{{.Service.Name | pascal}}",
      "type": "variable",
      "env": "dev"
    },
    {
      "name": "SERVICE_URL",
      "value": "https://api.example.com/{{.Repo.Name | kebab}}/v1",
      "type": "variable",
      "env": "production"
    }
  ]
}
```

In the example above:
- The template function `pascal` converts "my-service" to "MyService"
- The template function `kebab` ensures consistent kebab-case formatting
- Any template variable from the project can be referenced

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
