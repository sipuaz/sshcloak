# Contributing to sshcloak

Thank you for taking the time to contribute.  This document describes the
conventions used in this project so that contributions are consistent and easy
to review.

---

## Code of conduct

Be respectful and constructive.  This project follows the
[Contributor Covenant v2.1](https://www.contributor-covenant.org/version/2/1/code_of_conduct/).

---

## Opening an issue

Before opening an issue, please:

1. **Search existing issues** — the problem or idea may already be tracked.
2. **Use the right label** — choose one of:
   - `bug` — something is broken or behaves unexpectedly
   - `enhancement` — a new feature or improvement
   - `question` — a usage question (consider the docs first)

### Bug reports

A useful bug report includes:

- sshcloak version (`sshcloak --version` or `git describe --tags`)
- OS and OpenSSH version (`ssh -V`)
- The exact command you ran
- The full error output
- What you expected to happen

### Feature requests

Describe the problem you are trying to solve, not just the solution you have
in mind.  This helps evaluate the request in the context of the project's
scope.

---

## Submitting a pull request

1. **Fork** the repository and create a branch from `main`:
   ```bash
   git checkout -b feat/my-feature
   ```

2. **Keep changes focused** — one logical change per PR.  A PR that fixes a
   bug and refactors unrelated code is harder to review.

3. **Write tests** — all new behaviour must have unit tests.  Use the
   `//go:build integration` tag for tests that touch the filesystem or
   perform real encryption.

4. **Ensure the test suite passes** before opening the PR:
   ```bash
   make test
   make test-integration
   ```

5. **Update documentation** if you add or change a command, flag, or
   public API.  Doc pages live in `docs/`.

6. **Open the PR against `main`** with a clear title and description.
   Link any related issue with `Closes #<number>`.

---

## Coding conventions

### General

- Follow standard Go style — run `gofmt` and `go vet` before committing.
- Use `golangci-lint` if available; the CI configuration is the source of
  truth for which linters are enforced.
- Prefer clarity over cleverness.  This is a security-adjacent tool; obvious
  code is easier to audit.

### Comments

- Every exported symbol (function, type, constant) must have a Go doc comment.
- Non-obvious internal logic should have an inline comment explaining *why*,
  not *what*.

### Error handling

- Return errors; do not panic except for programmer mistakes (wiring bugs,
  nil injections).
- Wrap errors with context using `fmt.Errorf("operation: %w", err)` so the
  call stack is clear in error messages.
- Do not swallow errors silently.

### Security

- Secrets (passphrases, passwords) must be handled as `[]byte` and zeroed
  before the slice is released.  Never store secrets in a `string` longer
  than necessary.
- Do not log or print secrets, even at debug level.
- Validate all inputs at system boundaries (CLI arguments, file content).

### File writes

All writes to config files and the vault must go through the atomic write
pattern (temp file in the same directory → `chmod` → `os.Rename`).  Never
write directly to the target path.

### Tests

- Unit tests use in-memory stubs; they must not touch the real filesystem or
  network.
- Integration tests (tagged `//go:build integration`) use `t.TempDir()` and
  clean up after themselves automatically.
- Table-driven tests are preferred for coverage of multiple input variants.

---

## Commit messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
<type>(<scope>): <short summary>

[optional body]

[optional footer: Closes #N]
```

Common types: `feat`, `fix`, `docs`, `test`, `refactor`, `chore`.

Examples:

```
feat(password): add password list subcommand
fix(config): handle missing home directory gracefully
docs: add internals architecture page
```

---

## Local development

```bash
# Build
make build

# Unit tests
make test

# Integration tests
make test-integration

# Live docs preview
make docs-serve
```

The docs site requires Python ≥ 3.9.  Run `make docs-install` once to set up
the virtual environment.
