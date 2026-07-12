# sshcloak session

Manage sudo-like vault unlock cache.

## Synopsis

```
sshcloak session <subcommand>
```

## Description

When session caching is enabled (default), sshcloak stores the vault passphrase
in a local memory-only session agent for a limited time (`--session-ttl`).
Commands in the same shell session can reuse that unlock state without prompting
again until the TTL expires.

## Subcommands

### session status

```
sshcloak session status
```

Shows whether the current shell session has an active cached vault unlock.

### session lock

```
sshcloak session lock
```

Removes the cached vault unlock for the current shell session.

### session clear

```
sshcloak session clear
```

Clears all cached vault unlock sessions from the local session agent.

## Global flags

| Flag | Default | Description |
|---|---|---|
| `--session-ttl` | `15m` | Duration before cached vault unlock expires |
| `--no-session-cache` | `false` | Disable session cache and prompt every command |

## See also

- [vault](vault.md) — manage vault encryption and passphrase rotation
- [password](password.md) — store and retrieve host passwords
