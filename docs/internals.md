-.-# Internals

This page describes sshcloak's internal architecture for contributors and
advanced users.

---

## Component overview

```mermaid
graph TD
    subgraph cmd
        ROOT[root.go flags + wiring]
        HOST[host/* add list get edit remove]
        VAULT[vault/* init rotate-passphrase]
        PASS[password/* set get delete list]
    end

    subgraph internal/config
        MGR[Manager CRUD + include bootstrap]
        PARSER[Parser tokenizer → AST]
        EVAL[Evaluator first-match-wins resolve]
        FILE[FileHandler atomic read/write]
    end

    subgraph internal/keyring
        STORE[FileVaultStore lock / unlock / CRUD]
        CRYPTO[crypto.go age encrypt / decrypt]
        TYPES[types.go SecretRecord VaultDocument]
    end

    ROOT --> HOST
    ROOT --> VAULT
    ROOT --> PASS

    HOST --> MGR
    VAULT --> STORE
    PASS --> STORE

    MGR --> PARSER
    MGR --> EVAL
    MGR --> FILE

    STORE --> CRYPTO
    STORE --> TYPES
    CRYPTO -->|filippo.io/age| EXT([age library])
```

---

## SSH config manager (`internal/config`)

sshcloak parses `~/.ssh/config` with a hand-written tokenizer and builds an
AST of `Host` and `Match` blocks.  It never uses an external SSH config
library so the output is always canonical and round-trips losslessly.

### Parse pipeline

```mermaid
flowchart LR
    A[raw text] --> B[Lexer token stream]
    B --> C[Parser AST: Config / Block / Directive]
    C --> D[Evaluator Resolve hostname → directives]
    C --> E[Renderer AST → canonical text]
    E --> F[FileHandler atomic write]
```

### Isolation model

sshcloak appends exactly one line to `~/.ssh/config`:

```
Include ~/.ssh/sshcloak/config
```

All host CRUD operates exclusively on `~/.ssh/sshcloak/config`.  The root file
is never rewritten, only appended to (once).

### Atomic writes

All config and vault writes use the same pattern:

1. Write to a temp file in the **same directory** as the target.
2. `chmod` the temp file to the correct permissions.
3. `os.Rename` — atomic on POSIX systems.

This guarantees that a crash or power loss never produces a truncated or
partially-written file.

---

## Encrypted vault (`internal/keyring`)

### Vault format

The vault is a single YAML document encrypted with age and ASCII-armored
(PEM-style).  The plaintext structure:

```yaml
version: 1
secrets:
  "myserver\x00password":
    label: myserver
    kind: password
    value: s3cr3t!
    created_at: 2026-05-10T12:00:00Z
    updated_at: 2026-05-10T12:00:00Z
```

The map key is `label + "\x00" + kind`, which makes label+kind pairs unique
without nesting.

### Encryption

```mermaid
sequenceDiagram
    participant CLI
    participant FileVaultStore
    participant age

    CLI->>FileVaultStore: Unlock(passphrase)
    FileVaultStore->>age: NewScryptIdentity(passphrase)
    age-->>FileVaultStore: decrypt vault.age → VaultDocument
    FileVaultStore-->>CLI: ok

    CLI->>FileVaultStore: Put(record)
    FileVaultStore->>age: NewScryptRecipient(passphrase)
    age-->>FileVaultStore: encrypt VaultDocument → vault.age
    FileVaultStore-->>CLI: ok

    CLI->>FileVaultStore: Lock()
    FileVaultStore->>FileVaultStore: zero passphrase bytes
```

age uses scrypt with default work factors, which makes brute-force attacks
against a stolen vault file computationally expensive.

---

## Data flow: `password set`

```mermaid
sequenceDiagram
    actor User
    participant CLI as sshcloak password set
    participant Vault as FileVaultStore
    participant Disk

    User->>CLI: sshcloak password set myserver
    CLI->>User: Vault passphrase: [no echo]
    User->>CLI: ••••••••
    CLI->>Vault: Unlock(passphrase)
    Vault->>Disk: read vault.age
    Disk-->>Vault: ciphertext
    Vault->>Vault: age decrypt → VaultDocument
    Vault-->>CLI: ok

    CLI->>User: SSH password for "myserver": [no echo]
    User->>CLI: ••••••••
    CLI->>Vault: Put(SecretRecord{label, password, value})
    Vault->>Vault: age encrypt → ciphertext
    Vault->>Disk: atomic write vault.age
    Vault-->>CLI: ok

    CLI->>Vault: Lock()
    Vault->>Vault: zero passphrase bytes in memory
    CLI-->>User: password stored for "myserver"
```

---

## Dependency graph

| Package | External dependency | Purpose |
|---|---|---|
| `internal/keyring` | `filippo.io/age` | Passphrase-based encryption |
| `internal/keyring` | `go.yaml.in/yaml/v3` | Vault serialization |
| `cmd/...` | `github.com/spf13/cobra` | CLI framework |
| `cmd/...` | `golang.org/x/term` | No-echo passphrase prompting |
| `internal/config` | none | Pure Go SSH config parser |
