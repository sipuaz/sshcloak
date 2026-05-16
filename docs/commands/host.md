# sshcloak host

Manage SSH host entries in the sshcloak-managed include file.

---

## Synopsis

```
sshcloak host <subcommand> [options]
```

---

## Subcommands

### host add

```
sshcloak host add <label> [flags]
```

**Description**

Add a new `Host` block to `~/.ssh/sshcloak/config`.  The `label` becomes the
SSH alias (the value after `Host` in the config file).  Wildcards are not
permitted as labels.

If `~/.ssh/config` does not yet contain an `Include` directive for the managed
file it is added automatically.

**Options**

| Flag | Type | Description |
|---|---|---|
| `--hostname` | string | Remote hostname or IP address (`HostName`) |
| `--user` | string | Remote login username (`User`) |
| `--port` | string | Remote port (`Port`) |
| `--identity-file` | string (repeatable) | Path to a private key (`IdentityFile`) |
| `--extra KEY=VALUE` | string (repeatable) | Any additional SSH directive |

**Examples**

```bash
# Minimal — alias only
sshcloak host add bastion

# Full example
sshcloak host add prod \
    --hostname 203.0.113.5 \
    --user deploy \
    --port 22 \
    --identity-file ~/.ssh/id_ed25519 \
    --extra ProxyJump=bastion \
    --extra ServerAliveInterval=30
```

---

### host list

```
sshcloak host list
sshcloak host list --tag production
```

**Description**

Print a table of all managed host entries.  Wildcard blocks (e.g. `Host *`)
are excluded from the output.  When `--tag` is provided, only hosts whose
metadata sidecar entry includes that tag are shown.

**Output columns**

| Column | Description |
|---|---|
| `LABEL` | SSH alias |
| `HOSTNAME` | `HostName` value |
| `USER` | `User` value |
| `PORT` | `Port` value |
| `TAGS` | Comma-separated tags from `~/.ssh/sshcloak/meta.yaml` |

**Example**

```
LABEL     HOSTNAME       USER    PORT  TAGS
bastion   203.0.113.1    ops     22    edge,jump
prod      203.0.113.5    deploy  22    production,web
```

---

### host get

```
sshcloak host get <label>
```

**Description**

Print all directives for a single host entry, one per line.

**Example**

```bash
$ sshcloak host get prod
Label:          prod
HostName:       203.0.113.5
User:           deploy
Port:           22
Tags:           production,web
IdentityFile:   ~/.ssh/id_ed25519
ProxyJump:      bastion
```

---

### host tag add

```
sshcloak host tag add <label> <tag> [<tag>...]
```

**Description**

Add one or more arbitrary tags to a host label in the sshcloak metadata sidecar
file at `~/.ssh/sshcloak/meta.yaml`.  The host itself must already exist in the
SSH config tree.

**Example**

```bash
sshcloak host tag add prod production web europe
```

---

### host tag remove

```
sshcloak host tag remove <label> <tag> [<tag>...]
```

**Description**

Remove one or more tags from a host label.  When the last tag is removed, the
sidecar entry is deleted automatically.

**Example**

```bash
sshcloak host tag remove prod europe
```

---

### host tag list

```
sshcloak host tag list <label>
```

**Description**

Print the tags currently associated with one host label, one tag per line.

**Example**

```bash
sshcloak host tag list prod
production
web
```

---

### host edit

```
sshcloak host edit <label> [flags]
```

**Description**

Update one or more directives of an existing host entry.  Uses merge-patch
semantics: only flags that are explicitly supplied overwrite the current value.
Omitted flags leave the existing directive unchanged.

To clear a single-valued field pass an empty string:

```bash
sshcloak host edit prod --port ""
```

To fully replace the `IdentityFile` list supply all desired paths via repeated
`--identity-file` flags.

Extra directives are merged: new keys overwrite existing ones, unmentioned keys
are preserved.

**Options**

Same flags as [`host add`](#host-add).

**Example**

```bash
# Change only the port
sshcloak host edit prod --port 2222

# Add a proxy jump without touching other fields
sshcloak host edit prod --extra ProxyJump=bastion
```

---

### host remove

```
sshcloak host remove <label>
sshcloak host rm <label>
sshcloak host delete <label>
```

**Description**

Delete the `Host` block for `<label>` from the managed include file.  
The root `~/.ssh/config` is not modified.

**Example**

```bash
sshcloak host remove prod
```

---

## Global flags

These flags are inherited from the root command and apply to all `host` subcommands.

| Flag | Default | Description |
|---|---|---|
| `--config` | `~/.ssh/config` | Path to the root SSH config file |
| `--managed-config` | `~/.ssh/sshcloak/config` | Path to the sshcloak-managed include file |
| `--meta` | `~/.ssh/sshcloak/meta.yaml` | Path to the sshcloak host metadata sidecar file |

---

## Exit status

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Host not found, already exists, or other error |

---

## See also

- [vault](vault.md) — manage the encrypted vault
- [password](password.md) — store SSH passwords
