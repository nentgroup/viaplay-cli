# secrets Command

Manage secrets for your projects and teams securely using the system keyring.

---

## Available Commands

### `vip secrets set <name>`
Store a secret in the system keyring. You will be prompted to enter the secret value securely (it will not be shown on screen).

### `vip secrets get <name>`
Retrieve a secret from the system keyring by name.

### `vip secrets list`
List all secrets stored in the system keyring.

### `vip secrets forget <name>`
Remove a secret from the system keyring by name.

---

## Examples

```bash
vip secrets set GH_TOKEN
vip secrets get GH_TOKEN
vip secrets list
vip secrets forget GH_TOKEN
```

---

Secrets are stored securely in your system keyring and can be injected into your repository or CI/CD pipelines as needed. Team secrets are managed in the team config directory, while project secrets are stored in the project config.

---

For more, see the [CLI secrets command reference](cli/secrets.md).
