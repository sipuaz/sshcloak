# sshcloak — Project Context Summary

## What is sshcloak?

`sshcloak` is a Go CLI tool that wraps the standard `ssh` command transparently, injecting passwords and passphrases silently from the OS keyring so that the user never has to type credentials manually. It also acts as a host manager, reading and writing `~/.ssh/config` natively.

The project will be open source, licensed under **GPL-3.0**, and published on GitHub once considered mature enough for public use.

---

## Why it exists — the problem

The user evaluated existing SSH connection managers:

- **Termius** — polished but subscription-based, poor backup story, proprietary config format
- **Tabby** — buggy, complicated interactions
- **Ásbrú Connection Manager** — good feature set but no package for Ubuntu 26.04, install failed
- **MobaXterm / XShell / Royal TSX** — Windows or macOS only, not viable on Linux
- **Plain `~/.ssh/config`** — loved by the user but lacks credential storage

The user wants the simplicity of `~/.ssh/config` as the single source of truth, combined with transparent password injection so that typing `ssh myserver` just works — no prompts, no GUI required.

---

## User profile

- **OS:** Ubuntu 26.04 LTS
- **Shell:** zsh
- **Programming languages:** Java, TypeScript, Python, Go (proficient in all)
- **Preferred language for this project:** Go
- **Style:** Unix philosophy, config-as-files, no proprietary lock-in, open source friendly
- **Server auth mix:** mainly password-based login, some key-based

---

## Core requirements (as emerged from conversation)

1. **Transparent `ssh` wrapper** — the user types `ssh myserver` as always. A shell function in `.zshrc` intercepts it and delegates to `sshcloak`, which injects the password silently if one is stored, then calls the real `/usr/bin/ssh`.

2. **`~/.ssh/config` as the single source of truth** — `sshcloak` reads and writes standard SSH config natively. No parallel proprietary config file.

3. **Host manager** — ability to add, remove, list, and edit hosts in `~/.ssh/config` via CLI commands.

4. **Credential storage via OS keyring abstraction** — passwords are stored in the OS keyring (GNOME Keyring on Linux, macOS Keychain on macOS), never in plaintext. The keyring backend is abstracted behind an interface so the tool works cross-platform. A fallback encrypted-file backend should exist for headless/CI environments.

5. **No external password manager required** — the user explicitly does not want to rely on KeePass, Bitwarden, or similar tools. Credentials live inside the tool's own storage.

6. **Easy backup** — since `~/.ssh/config` is a plain text file, and credentials live in the OS keyring (exportable via `secret-tool`), backup is trivial and portable.

7. **Shell initialisation** — `sshcloak` should provide an `init` command that emits a shell snippet (similar to `eval "$(sshcloak init)"`), injecting the `ssh()` wrapper function into the user's shell without manual editing.

8. **Isolation** - `~/.ssh/config` MUST NOT be directly written except for the import statement of `~/.ssh/sshcloak/config`. This ensures isolation of the as-is configs. `sshcloak` must write/updated/delete its entries inside `~/.ssh/sshcloak/config` only, while ensuring read operations from `~/.ssh/config` and its import statements.

---

## Technical decisions made

| Decision | Choice | Reason |
|---|---|---|
| Language | Go | Single static binary, cross-compilation, strong typing, good SSH config library available |
| CLI framework | `cobra` + `viper` | De-facto standard for Go CLIs |
| Licence | GPL-3.0 | Motivates contributors to give back modifications |
| Name | `sshcloak` | Describes transparent, invisible credential handling |

---

## Proposed CLI surface (draft)

```
sshcloak host add <label> --hostname <ip/domain> --user <user> --port <port>
sshcloak host remove <label>
sshcloak host list
sshcloak host edit <label> --hostname <new>

sshcloak password set <label>
sshcloak password delete <label>

sshcloak connect <label>     # called internally by the ssh() shell wrapper
sshcloak init                # prints shell snippet for .zshrc/.bashrc
```

---

## Shell wrapper behaviour

A shell function named `ssh` is injected into the user's shell via `sshcloak init`:

```bash
ssh() {
    sshcloak connect "$@"
}
```

`sshcloak connect` logic:
1. Parse the host label from arguments
2. Look up credentials in the OS keyring for that label
3. If password found → `exec sshpass -p <password> /usr/bin/ssh "$@"`
4. If not found (key-based host) → `exec /usr/bin/ssh "$@"`

Using `exec` ensures TTY, signals, and terminal behaviour are identical to native SSH.

---

## README tagline (chosen)

> `ssh myserver`. That's it. No password prompt, no proprietary config, no subscription. Just your `~/.ssh/config` the way you like it, with credentials silently handled by your OS keyring.