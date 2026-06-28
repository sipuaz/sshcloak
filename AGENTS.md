# AGENTS.md

This file is the canonical operating guide for AI agents in this repository.

## 1. Repository Purpose and Type

- **Type:** Go CLI application (`sshcloak`) with documentation site (MkDocs).
- **Purpose:** Manage SSH hosts and encrypted credentials while keeping OpenSSH-native workflows (`ssh myhost`) and config compatibility.
- **Primary outputs:** `sshcloak` binary and published docs (`docs/`, `mkdocs.yml`).

## 2. Directory Responsibility Map

- `cmd/`: CLI entrypoint and command wiring (`host`, `vault`, `password`, `connect`, `init`, `version`).
- `internal/config/`: SSH config parsing, evaluation, rendering, and file persistence logic.
- `internal/keyring/`: Encrypted vault domain model, crypto, unlock/lock lifecycle, persistence.
- `internal/metadata/`: Sidecar metadata storage (for host tags and related metadata).
- `docs/`: User-facing documentation and command references.
- `.github/workflows/`: CI, version tagging, release, and docs publish automation.
- `Makefile`: Canonical local build/test/docs commands.
- `VERSION`: Release version source for version-tagging workflow.

## 3. Branch and Contribution Model

- Contribution flow documented in `CONTRIBUTING.md`:
  - Fork repo.
  - Create branch from `main`.
  - Open PR against `develop` (maintainer-confirmed policy).
- Workflow evidence:
  - CI runs on PRs targeting `main` and `develop` (`.github/workflows/ci.yml`).
  - `VERSION` tagging workflow runs on pushes to `main` and `develop` (`.github/workflows/version.yml`).
  - Dev pre-release publishes from `develop` (`.github/workflows/dev-release.yml`).
  - Stable release publishes from pushed tags `v*` (`.github/workflows/release.yml`).
- **Merge strategy confirmed by maintainer:** only `squash` or `rebase` merges are allowed.
- Required approvals and branch protection defaults are not yet documented in-repo.
- **Release track confirmed by maintainer:** `develop` is for alpha/pre-release work; `main` is for stable releases.

## 4. Core Operating Principles for Agents

- Keep changes narrowly scoped to one logical change per PR.
- Add/update tests for new behavior.
- Update docs when commands/flags/public behavior change.
- Prefer clear, auditable code over cleverness (security-adjacent project).
- Follow Go conventions:
  - `gofmt` and `go vet`.
  - Exported symbols require doc comments.
  - Wrap and propagate errors; do not silently swallow errors.
- Security handling:
  - Never log or print secrets.
  - Handle secrets as `[]byte` where feasible and clear memory after use.
  - Validate inputs at CLI/file boundaries.
- File persistence:
  - Use atomic write pattern for config/vault writes (temp file in same dir, chmod, rename).

## 5. Domain-Specific Conventions

- SSH config isolation model:
  - `~/.ssh/config` should only carry/include `~/.ssh/sshcloak/config`.
  - sshcloak-managed host CRUD belongs in managed config, not direct root-config rewrites.
- Include precedence behavior:
  - `sshcloak init` places include at top by default so sshcloak entries win with SSH first-match semantics (`docs/getting-started.md`).
- Vault behavior:
  - Vault stored at `~/.ssh/sshcloak/vault.age` by default.
  - age+scrypt encryption via `internal/keyring/crypto.go`.
  - File permissions and atomic persistence are security-critical (`0600` vault file, `0700` vault dir in `internal/keyring/file_vault.go`).
- Password injection behavior:
  - `connect` uses `SSHPASS` env var with `sshpass -e` when a password exists, otherwise plain `ssh` fallback (`docs/commands/connect.md`).

## 6. High-Care Files and Why

- `internal/keyring/file_vault.go`: Secret lifecycle, lock/unlock state, passphrase zeroing, vault write semantics.
- `internal/keyring/crypto.go`: Encryption/decryption correctness and vault format compatibility.
- `internal/config/file.go`: Atomic write guarantees for config safety.
- `.github/workflows/version.yml`: Version format/monotonicity logic and automated tag creation.
- `VERSION`: Directly affects tagging and release behavior.
- `docs/getting-started.md` + `docs/commands/*.md`: User-facing contract for CLI behavior; must stay aligned with implementation.

## 7. Validation Before Completion

Required PR checks (from `.github/pull_request_template.md`):

- `make test`
- `make test-integration`
- `make docs-build`

Additional local quality checks for code changes:

- `gofmt` on touched files.
- `go vet ./...`
- Update docs under `docs/` and `README.md` when command/flag/public behavior changes.

## 8. Commit and PR Summary Templates

### Commit message template (Conventional Commits)

```text
<type>(<scope>): <short summary>

[optional body]

[optional footer: Closes #<number>]
```

Common types in this repo: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`.

### PR summary template

```text
## Summary
<what changed and why>

## Type of change
- [ ] Bug fix
- [ ] New feature
- [ ] Documentation
- [ ] Refactor
- [ ] Chore

## Verification
- [ ] make test
- [ ] make test-integration
- [ ] make docs-build

## Notes
<reviewer context, risks, follow-ups>
```

## 9. Known Gaps Requiring Human Clarification

- Review and merge governance is not documented in-repo.
  - **Clarify:** Required approvals? Required status checks?
- Security response policy is currently limited to “report privately” (`SECURITY.md`) and does not define expected timelines.
  - **Clarify (expanded):** What explicit security-response expectations should agents communicate?
    - **Time to first acknowledgment:** max time between submission and first maintainer response confirming receipt.
    - **Triage severity assignment timeline:** max time to classify severity/impact after acknowledgment.
    - **Coordinated disclosure / embargo window:** **90 days** private by default before public disclosure (unless maintainer and reporter agree otherwise).
    - **Progress update cadence:** how often maintainers update the reporter while investigation/fix is ongoing (for example every 7 days).
