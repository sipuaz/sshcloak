# connect

Open an SSH session with automatic password injection.

## Synopsis

```
sshcloak connect <label> [-- <ssh-args>...]
```

## Description

`sshcloak connect` retrieves the stored password for `<label>` from the encrypted
vault and opens an SSH session to that host without requiring you to type the
password manually.

Internally the command executes:

```
sshpass -e ssh <label> [ssh-args...]
```

The password is passed to `sshpass` via the `SSHPASS` environment variable so
it **does not appear in the process list** (unlike `sshpass -p`).

Because `sshcloak connect` replaces the current process with `sshpass` +
`ssh` (via `execve`), the SSH session inherits the terminal directly — PTY
allocation, signal forwarding, and terminal resize all work as expected.

## Prerequisites

`sshpass` must be installed and on your `$PATH`:

| Distribution | Command |
|---|---|
| Debian / Ubuntu | `apt install sshpass` |
| Fedora / RHEL | `dnf install sshpass` |
| macOS (Homebrew) | `brew install hudochenkov/sshpass/sshpass` |

## Arguments

| Argument | Description |
|---|---|
| `<label>` | The sshcloak host label whose password should be injected |
| `-- <ssh-args>` | Optional extra arguments forwarded verbatim to `ssh` |

## Examples

Basic connection:

```bash
sshcloak connect prod
```

With extra ssh flags (port forwarding, X11 forwarding):

```bash
sshcloak connect prod -- -X -L 8080:localhost:80
```

## Workflow

```
sshcloak connect prod
```

1. Prompts for the vault passphrase (no echo).
2. Unlocks the encrypted vault.
3. Retrieves the password stored under label `prod`.
4. Execs: `sshpass -e ssh prod` with `SSHPASS=<password>` in the environment.

## Notes

- A password must already be stored with `sshcloak password set <label>` before
  `connect` can inject it.
- The host entry must be resolvable by SSH — either via the sshcloak-managed
  include file or your existing `~/.ssh/config`.
- `connect` does not fall back to interactive password entry if no vault record
  exists; it exits with an error so that `sshpass` is never invoked without a
  known password.
- To use key-based authentication without password injection, connect with
  plain `ssh <label>` as usual.
