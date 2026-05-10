# sshcloak password

Store, retrieve, and delete SSH passwords in the encrypted vault.

---

## Synopsis

```
sshcloak password <subcommand> [label]
```

---

## Description

Each password is associated with a host `label` — the same alias used in
`sshcloak host add`.  Every subcommand prompts for the vault passphrase
interactively, unlocks the vault, performs its operation, and locks the vault
before returning.

---

## Subcommands

### password set

```
sshcloak password set <label>
```

**Description**

Store an SSH password for `<label>`.  If an entry for that label already
exists it is overwritten.

Two prompts are shown in sequence (no echo on either):

1. **Vault passphrase** — unlocks the vault
2. **SSH password** — the value to store

**Example**

```
$ sshcloak password set prod
Vault passphrase:
SSH password for "prod":
password stored for "prod"
```

---

### password get

```
sshcloak password get <label>
```

**Description**

Print the stored password for `<label>` to **stdout** followed by a newline.
The vault passphrase is prompted on **stderr** so the password alone can be
captured in a sub-shell.

!!! warning "Security"
    The password is printed in plaintext.  Do not use this command in
    environments where stdout is logged or visible to other users.

**Example**

```bash
$ sshcloak password get prod
Vault passphrase:
s3cr3t!
```

Scripting:

```bash
PASS=$(sshcloak password get prod)
sshpass -p "$PASS" ssh prod
```

---

### password delete

```
sshcloak password delete <label>
sshcloak password rm <label>
sshcloak password remove <label>
```

**Description**

Remove the stored password for `<label>` from the vault.  The vault is
re-encrypted after deletion.

**Example**

```
$ sshcloak password delete prod
Vault passphrase:
password deleted for "prod"
```

---

### password list

```
sshcloak password list
```

**Description**

Print the label of every entry in the vault that has kind `password`, one per
line.  If no passwords are stored, prints `no passwords stored`.

**Example**

```
$ sshcloak password list
Vault passphrase:
bastion
prod
staging
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
| `1` | Wrong passphrase, label not found, or I/O error |

---

## See also

- [vault](vault.md) — initialise and rotate the vault passphrase
- [host](host.md) — manage SSH host entries
