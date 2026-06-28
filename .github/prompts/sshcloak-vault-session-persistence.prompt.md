---
description: "Use when: defining a sudo-like vault password session cache for sshcloak CLI commands."
name: "Sshcloak Vault Session Persistence"
argument-hint: "Extra constraints (TTL, platform, security model, UX)"
agent: "agent"
---
Help me define a secure session-persistence mechanism for sshcloak vault password unlock, so users do not need to retype the vault password for every command in the same CLI session.

Treat this as a sudo-like model:
1. Password is requested once per CLI session (or per configurable timeout window).
2. Subsequent commands in that session reuse unlock state without storing plaintext password on disk.
3. Behavior must fail closed and keep existing security guarantees.

Repository context:
- CLI entrypoint and command wiring: `cmd/`
- Vault lifecycle and crypto: `internal/keyring/file_vault.go`, `internal/keyring/crypto.go`
- User-facing behavior docs: `docs/commands/*.md`, `docs/getting-started.md`

Your task:
1. Propose 2-3 viable designs (with tradeoffs) for session persistence in sshcloak.
2. Recommend one design and explain why it best fits this codebase.
3. Provide an implementation plan with concrete touched files/functions.
4. Specify security considerations (memory lifecycle, lock timeout, process boundaries, crash behavior).
5. Provide acceptance criteria and test plan (unit + integration).
6. Suggest required documentation updates.

If I pass extra constraints in the prompt argument, prioritize them in the design and call out any conflicts explicitly.
