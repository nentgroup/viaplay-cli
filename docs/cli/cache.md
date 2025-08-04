# cache Command

Manage the local cache of project templates.

---

## Subcommands & Flags

### `vip cache update`
Update all templates in the cache to their latest versions.
- No arguments or flags.

### `vip cache prune [--days <n>]`
Remove templates from the cache that haven't been used in a specified number of days.
- `--days <n>`: Prune templates older than the specified number of days (default: 30).

### `vip cache clean`
Remove all template files from the cache.
- No arguments or flags. Prompts for confirmation before deleting.

### `vip cache list`
List all templates currently stored in the cache.
- No arguments or flags.

### `vip cache info`
Show detailed information about the template cache, including size and statistics.
- No arguments or flags.

---

## Update the Template Cache

Refresh the local cache of remote templates:

```
vip cache update
```

This ensures you have the latest version of all templates.

---

## Clear the Template Cache

Remove all cached templates (if supported):

```
vip cache clear
```

This can help resolve issues with outdated or corrupted templates.

---

## Examples

```bash
vip cache update
vip cache prune --days 60
vip cache clean
vip cache list
vip cache info
```

---

Caching improves performance and allows offline use of previously downloaded templates.
