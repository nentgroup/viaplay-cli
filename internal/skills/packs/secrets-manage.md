---
title: Manage secrets in the system keyring
description: Store, retrieve, list, and remove secrets used by vip team configs and repo apply.
---

# vip: secrets management

Use `vip secrets` to manage secrets in the local system keyring. These secrets can be
referenced by name from `vip repo apply secret` / `vip repo apply variable` when no
explicit source is given, so storing them here lets team members push the same named
secret to a repository without embedding the value in any config file or chat message.

## When to use this

Trigger this workflow when the user asks to:
- "store/save a secret called X locally"
- "what secrets do I have saved"
- "remove/delete a saved secret"
- "push my saved GITHUB_TOKEN secret to repo Y" (combine with the `repo-manage` skill:
  `vip repo apply secret owner/repo NAME --team <team>` will resolve from the keyring if
  not found in team config)

## Commands

```
vip secrets set <NAME>            # prompts securely for the value (not echoed)
vip secrets set <NAME> --file path/to/file   # reads value from a file instead of prompting
vip secrets get <NAME>             # confirms the secret exists; add --show to print the value
vip secrets forget <NAME>          # removes the secret from the keyring
vip secrets list                   # lists secret names only (not implemented as a full list yet —
                                    # falls back to guidance for using the OS keyring utility directly)
```

## Guardrails

- **Never** type or paste a secret value directly as a CLI argument or into chat — always
  use the interactive prompt (`vip secrets set <NAME>`) or `--file` so the value never
  appears in shell history or this conversation.
- Never run `vip secrets get <NAME> --show` unless the user explicitly asks to see the
  value; without `--show` the command only confirms the secret exists.
- Confirm the exact secret name with the user before `forget` — this is a destructive,
  unrecoverable action (no confirmation prompt in the tool itself).
- Do not paste retrieved secret values back into the chat transcript unless the user
  explicitly asks for the value and understands it will be visible in the conversation.
