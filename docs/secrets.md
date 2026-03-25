# Secrets

---

## What are Secrets?

Secrets are sensitive values (such as API keys, tokens, or passwords) used by your repositories or CI/CD pipelines. viaplay-cli manages secrets at the team level and injects them into repositories or environments during project creation.

---

## Where are Secrets Stored?

- **Team secrets:**
  - Configured in `~/.config/viaplay/teams/<team>/secrets.yaml` (or the global secrets file)
  - Secret values are stored in your OS keyring (e.g., macOS Keychain, Windows Credential Manager, or Linux Secret Service) — config files only reference names and metadata.

---

## Example Team Secrets File

```yaml
secrets:
  - name: GH_TOKEN
    value: ${{ secrets.GH_TOKEN }}
    env: ""
    type: secret
  - name: STAGING_API_KEY
    value: ${{ secrets.STAGING_API_KEY }}
    env: staging
    type: secret
  - name: DATABASE_URL
    value: ${{ secrets.DATABASE_URL }}
    env: production
    type: variable
```

- `env`: Target environment where the secret/variable will be created
  - When empty or omitted: Secret is created at the repository level (available to all workflows)
  - When specified (e.g., "staging", "production"): Secret is created for that specific GitHub environment only
- `type`: Either `secret` (encrypted) or `variable` (plain variable)

### Scoping Secrets to Environments

The `env` parameter lets you target secrets to specific deployment environments:

```yaml
secrets:
  - name: API_KEY
    value: ${{ secrets.DEV_API_KEY }}
    env: development
    type: secret
  - name: API_KEY
    value: ${{ secrets.PROD_API_KEY }}
    env: production
    type: secret
  - name: GLOBAL_SECRET
    value: ${{ secrets.GLOBAL_SECRET }}
    type: secret
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

## Referencing Secrets in Configuration

Use the `${{ secrets.SECRET_KEY }}` syntax to reference a secret stored in your OS keyring. The value is resolved securely at the time of repository creation or configuration.

This syntax can be used:
- In the `--repo-secrets` inline flag
- Inside YAML secrets configs (e.g., `secrets.yaml`)

Example in a secrets config:
```yaml
secrets:
  - name: API_KEY
    value: ${{ secrets.GH_TOKEN }}
    type: secret
```

Example with the CLI flag:
```bash
vip project create nentgroup/myservice --language go --type service --repo-secrets '{"secrets":[{"name":"API_KEY","value":"${{ secrets.GH_TOKEN }}","type":"secret"}]}'
```

> **Note:** When the `env` field is omitted or left empty, the secret is created at the repository level (available to all workflows). When an environment name is specified, the secret is scoped to that environment only.

## Using Template Variables in Secrets

You can reference template variables in your secret values using Go template syntax:

```yaml
secrets:
  - name: RESOURCE_PREFIX
    value: "{{.Project.Name}}-resources"
    type: variable
  - name: STACK_NAME
    value: "Dev-{{.Service.Name | pascal}}"
    type: variable
    env: dev
  - name: SERVICE_URL
    value: "https://api.example.com/{{.Repo.Name | kebab}}/v1"
    type: variable
    env: production
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

- Always use the `${{ secrets.SECRET_KEY }}` syntax to reference secrets in configuration files.
- Never commit secret values to version control.
- Use team or global secrets files to define which secrets are needed, and store the actual values in your OS keyring.
- Rotate secrets regularly and remove unused ones.

---

For more, see the [CLI secrets command reference](cli/secrets.md).
