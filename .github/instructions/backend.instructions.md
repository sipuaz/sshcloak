---
applyTo: **/*.go, go.mod
---
# Backend Engineer Role

You are writing and modifying backend Go code for sshcloak. Follow these engineering invariants strictly.

## Your Toolbox (Internal APIs)
- **Tool: `FileHandler.AtomicWrite` (`internal/config`)** -> Use this for ALL file writes. It guarantees atomic `os.Rename` operations. Never write directly to `~/.ssh/config` except to prepend the `Include` directive. All CRUD operations MUST happen inside `~/.ssh/sshcloak/config`.
- **Tool: `fileStub` (`internal/config/file_test.go` and `internal/metadata/store_test.go`)** -> Use this for ALL unit tests requiring file I/O. Do not use external mocking frameworks like `gomock` or `testify`.
- **Tool: `config.Manager` (`internal/config`)** -> Use this to read, mutate, and resolve the AST of the SSH config. NEVER import third-party SSH config parsers like `github.com/kevinburke/ssh_config`.

## Security & Error Handling
- **Memory Security:** Passphrases and passwords must be handled as `[]byte` and zeroed out in memory immediately after use. NEVER log or print secrets.
- **Errors:** Always return errors. Do not `panic` unless resolving a fatal developer wiring bug. Wrap all errors with context using `fmt.Errorf("action: %w", err)`.
- **Integration Tests:** Tests that touch the real filesystem or perform real encryption must be tagged with `//go:build integration` and use `t.TempDir()`.