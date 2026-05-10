# sshcloak vault

Manage the age-encrypted secret vault.

---

## Synopsis

```
sshcloak vault <subcommand>
```

---

## Description

The vault is a single file encrypted with [age](https://age-encryption.org/)
using scrypt passphrase derivation.  It stores all SSH passwords managed by
sshcloak.  The default location is `~/.ssh/sshcloak/vault.age`.

The vault is locked at rest.  Every command that reads or writes secrets
prompts for the passphrase, unlocks the vault in memory, performs its
operation, then locks the vault again before exiting.

---

## Subcommands

### vault init

```
sshcloak vault init
```

**Description**

Create the vault file if it does not exist, or verify an existing one by
performing an unlock/lock round-trip.

The passphrase is prompted twice for confirmation (no echo).

Running `vault init` is optional.  The first `password set` call also creates
the vault automatically.

**Example**

```
$ sshcloak vault init
Enter vault passphrase:
Confirm passphrase:
vault initialised
```

---

### vault rotate-passphrase

```
sshcloak vault rotate-passphrase
```

**Description**

Unlock the vault with the current passphrase and immediately re-encrypt it
with a new one.  The vault contents are not modified — only the encryption
key changes.

Both the current and the new passphrase are read interactively.  The new
passphrase is prompted twice for confirmation.

**Example**

```
$ sshcloak vault rotate-passphrase
Current passphrase:
New passphrase:
Confirm new passphrase:
vault passphrase rotated
```

---

## Global flags

| Flag | Default | Description |
|---|---|---|
| `--vault` | `~/.ssh/sshcloak/vault.age` | Path to the encrypted vault file |

---

## Exit status

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Wrong passphrase, I/O error, or passphrases did not match |

---

## Security notes

- The vault uses age's scrypt recipient, which is intentionally slow to
  resist brute-force attacks.
- The passphrase is never written to disk or echoed to the terminal.
- The in-memory passphrase bytes are zeroed before the process exits.

---

## See also

- [password](password.md) — store and retrieve SSH passwords
- [host](host.md) — manage SSH host entries
