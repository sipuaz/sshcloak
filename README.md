# sshcloak

`ssh myserver`. That's it. No password prompt, no proprietary config, no subscription.  
Your `~/.ssh/config` stays exactly as you wrote it — sshcloak manages host entries through a dedicated include file and injects passwords from an age-encrypted vault.

---

## What it does

- **Host management** — add, edit, list, and remove SSH host entries without ever touching your main `~/.ssh/config` directly.
- **Encrypted credential store** — passwords are stored in a single [age](https://age-encryption.org/)-encrypted vault at `~/.ssh/sshcloak/vault.age`, protected by a passphrase with scrypt derivation.
- **Non-destructive** — only one `Include ~/.ssh/sshcloak/config` line is ever appended to `~/.ssh/config`. Everything else is untouched.
- **Single static binary** — no runtime dependencies, no daemon, no root required.

## Requirements

| Requirement | Version |
|---|---|
| Linux or macOS | — |
| OpenSSH client | ≥ 7.3 (for `Include` directive support) |
| Go _(build only)_ | ≥ 1.22 |

## Installation

If you want `sshcloak` available as a global command without sudo, install it
to your user bin directory:

```bash
make install-user
export PATH="$HOME/.local/bin:$PATH"
```

For a system-wide install:

```bash
git clone https://github.com/sipuaz/sshcloak.git
cd sshcloak
make install          # builds and copies to /usr/local/bin/sshcloak
```

Or with `go install`:

```bash
go install github.com/sipuaz/sshcloak/cmd@latest
```

## Quick start

```bash
# 1. Bootstrap sshcloak
sshcloak init

# 2. Add a host
sshcloak host add prod --hostname 203.0.113.5 --user alice --port 22

# 3. Check the version or help
sshcloak version
sshcloak --help

# 4. Store its password
sshcloak password set prod

# 5. Connect normally
ssh prod
```

## Project layout

```
sshcloak/
├── cmd/
│   ├── main.go                     # entry point
│   └── internal/cli/
│       ├── root.go                 # root command, flag wiring
│       ├── host/                   # sshcloak host *
│       ├── vault/                  # sshcloak vault *
│       └── password/               # sshcloak password *
├── internal/
│   ├── config/                     # SSH config parser + CRUD manager
│   └── keyring/                    # age-encrypted vault
├── docs/                           # full documentation (MkDocs + Material)
├── Makefile
└── mkdocs.yml
```

## Makefile targets

| Target | Description |
|---|---|
| `make build` | Compile binary to `./sshcloak` |
| `make init` | Build and bootstrap sshcloak for the current user |
| `make install-user` | Build and install to `~/.local/bin` |
| `make install` | Build and copy to `/usr/local/bin` |
| `make test` | Run unit tests |
| `make test-integration` | Run unit + integration tests |
| `make docs-serve` | Live-preview docs at http://127.0.0.1:8000 |
| `make docs-build` | Build static docs site into `site/` |
| `make docs-deploy` | Publish docs to GitHub Pages |

## Documentation

Full documentation — including command reference and architecture internals — is available in [`docs/`](docs/) and published at <https://sipuaz.github.io/sshcloak>.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines on opening issues, submitting pull requests, and the coding conventions used in this project.

## License

GPL-3.0 — see [LICENSE](LICENSE).
