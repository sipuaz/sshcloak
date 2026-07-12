# sshcloak - Global Context

## Project Identity
sshcloak is a Go-based CLI tool (using Cobra) that acts as a transparent SSH credential manager. It intercepts `ssh` calls, injects passwords from an `age`-encrypted vault via `sshpass`, and manages `~/.ssh/config` host entries via a dedicated include file (`~/.ssh/sshcloak/config`).

## Core Architecture
- `cmd/internal/cli/`: Contains all Cobra command wiring grouped by domain (`connect`, `host`, `vault`, `password`). CLI state and flags are managed directly via `cobra.Command` (no global Viper state).
- `internal/config`: A pure-Go, zero-dependency SSH config parser and AST evaluator.
- `internal/keyring`: Manages the `age`-encrypted YAML vault (`vault.age`).
- `internal/metadata`: Manages the sidecar YAML file for host tags (`meta.yaml`).

*(Note: Domain-specific rules are automatically applied based on the file types you edit.)*