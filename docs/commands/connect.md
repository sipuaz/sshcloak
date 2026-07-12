# connect

Open an SSH session with optional automatic password injection.

## Synopsis

```
sshcloak connect [<label>] [-- <ssh-args>...]
```

## Description

`sshcloak connect` attempts to retrieve a stored password for `<label>` from the
encrypted vault.

- If a password is found, it opens SSH through `sshpass` so password entry is automatic.
- If no password is stored for that label, it falls back to plain `ssh` (key-based auth,
  agent auth, or interactive auth continue to work as normal).

If no label is provided, sshcloak shows an interactive host picker; `--tag`
narrows that picker to hosts carrying one metadata tag.

Internally the command executes:

```
sshpass -e ssh <label> [ssh-args...]
```

or, when no stored password exists:

```
ssh <label> [ssh-args...]
```

The password is passed to `sshpass` via the `SSHPASS` environment variable so
it **does not appear in the process list** (unlike `sshpass -p`).

Because `sshcloak connect` replaces the current process with `ssh` (or
`sshpass` + `ssh`) via `execve`, the SSH session inherits the terminal
directly — PTY allocation, signal forwarding, and terminal resize all work as
expected.

## Prerequisites

`sshpass` must be installed and on your `$PATH` only when password injection is used:

| Distribution | Command |
|---|---|
| Debian / Ubuntu | `apt install sshpass` |
| Fedora / RHEL | `dnf install sshpass` |
| macOS (Homebrew) | `brew install hudochenkov/sshpass/sshpass` |

## Arguments

| Argument | Description |
|---|---|
| `<label>` | Optional sshcloak host label whose password should be injected |
| `-- <ssh-args>` | Optional extra arguments forwarded verbatim to `ssh` |

## Flags

| Flag | Description |
|---|---|
| `--tag <name>` | Filter the interactive picker by one tag when no label is provided |
| `--debug <0-3>` | Add `-v`, `-vv`, or `-vvv` to the underlying `ssh` invocation |

## Examples

Basic connection:

```bash
sshcloak connect prod
```

Interactive selection filtered by tag:

```bash
sshcloak connect --tag stable
```

With extra ssh flags (port forwarding, X11 forwarding):

```bash
sshcloak connect prod -- -X -L 8080:localhost:80
```

## Workflow

```
sshcloak connect prod
```

1. Resolves the host label directly, or via the interactive picker when no label was given.
2. If `--tag` was provided, filters the picker to hosts carrying that tag.
3. Prompts for the vault passphrase on first use in a shell session (or every command when `--no-session-cache` is set).
4. Unlocks the encrypted vault.
5. If a password is stored for the selected label, execs: `sshpass -e ssh <label>`
  with `SSHPASS=<password>` in the environment.
6. If no password is stored, falls back to plain `ssh <label>`.

## Notes

- A password stored with `sshcloak password set <label>` is optional. It enables
  automatic password injection for that label.
- Session cache is memory-only and expires after `--session-ttl` (default 15m).
- The host entry must be resolvable by SSH — either via the sshcloak-managed
  include file or your existing `~/.ssh/config`.
- When no password record exists for a label, `connect` runs plain `ssh` and
  lets your normal SSH authentication methods apply (IdentityFile, agent, etc.).
