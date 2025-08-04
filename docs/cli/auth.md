# auth Command

Authenticate with GitHub to enable repository creation, configuration, and secret management.

---

## Subcommands & Arguments

### `vip auth login`
Authenticate with your GitHub account.
- No arguments or flags.
- This will open a browser window for OAuth authentication or prompt for a token.

### `vip auth logout`
Remove your stored GitHub token from the keyring.
- No arguments or flags.
- This will delete your locally stored authentication token.

### `vip auth status`
Show authentication status and account info.
- No arguments or flags.
- Displays whether you are authenticated and shows account details if available.

### `vip auth token`
Print the stored GitHub token from the keyring (if any).
- No arguments or flags.
- Outputs the token if it exists.

### `vip auth whoami`
Print the currently authenticated GitHub username.
- No arguments or flags.
- Shows the username associated with the stored token.

---

## Examples

```bash
vip auth login
vip auth logout
vip auth status
vip auth token
vip auth whoami
```

---

Authentication is required for all operations that interact with GitHub, such as creating repositories, managing secrets, and applying rulesets.
