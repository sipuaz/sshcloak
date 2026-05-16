# Getting Started

## Installation

### From source (recommended)

To install `sshcloak` as a global command for your user account without
sudo, use the user-local target:

```bash
git clone https://github.com/sipuaz/sshcloak.git
cd sshcloak
make install-user
export PATH="$HOME/.local/bin:$PATH"
```

For a system-wide install:

```bash
git clone https://github.com/sipuaz/sshcloak.git
cd sshcloak
make install          # builds and copies to /usr/local/bin/sshcloak
```

`make install` requires write permission to `/usr/local/bin`.  
Use `sudo make install` if you do not own that directory.

### With `go install`

```bash
go install github.com/sipuaz/sshcloak/cmd@latest
```

The binary is placed in `$GOPATH/bin` (default `~/go/bin`).  
Ensure that directory is on your `$PATH`:

```bash
export PATH="$HOME/go/bin:$PATH"   # add to ~/.zshrc or ~/.bashrc
```

You can also check the installed binary with:

```bash
sshcloak version
sshcloak --help
```

---

## First run

### 1 — Bootstrap sshcloak

The bootstrap command creates the `Include ~/.ssh/sshcloak/config` line in
`~/.ssh/config` and initialises the encrypted vault.  Create it with a
passphrase you will remember:

```
$ sshcloak init
Enter sshcloak vault passphrase:
Confirm passphrase:
sshcloak initialised
```

The vault is written to `~/.ssh/sshcloak/vault.age`.

By default `sshcloak init` places the `Include` directive at the **top** of
`~/.ssh/config` so that sshcloak host entries take precedence over any
existing entries (SSH uses first-match-wins semantics).  If the directive was
previously placed at the bottom (for example by an earlier `sshcloak host
add`), it is automatically relocated to the top.

If you want the Include directive appended instead, use `--append`:

```bash
sshcloak init --append
```

If you prefer a local shortcut during development, `make init` runs the same
bootstrap flow after building the binary.

### 2 — Add a host

```
$ sshcloak host add myserver \
    --hostname 203.0.113.10 \
    --user alice \
    --port 2222
host "myserver" added
```

sshcloak adds one `Include ~/.ssh/sshcloak/config` line at the **top** of
`~/.ssh/config` if the bootstrap command has not already done it, ensuring
sshcloak hosts are resolved with the highest priority.

### 3 — Store the password

```
$ sshcloak password set myserver
Vault passphrase:
SSH password for "myserver":
password stored for "myserver"
```

### 4 — Verify

```
$ sshcloak host get myserver
Label:    myserver
HostName: 203.0.113.10
User:     alice
Port:     2222
```

```
$ ssh myserver     # OpenSSH resolves the host via the include file
```

---

## File locations

| File | Default path | Purpose |
|---|---|---|
| Root SSH config | `~/.ssh/config` | Untouched except for one `Include` line |
| Managed config | `~/.ssh/sshcloak/config` | All sshcloak host entries |
| Metadata sidecar | `~/.ssh/sshcloak/meta.yaml` | Host tags and future sshcloak-owned metadata |
| Vault | `~/.ssh/sshcloak/vault.age` | age-encrypted password store |

All paths can be overridden with global flags — see `sshcloak --help`.

---

## Uninstalling

```bash
sudo make uninstall          # removes /usr/local/bin/sshcloak
rm -rf ~/.ssh/sshcloak       # removes managed config + vault
```

Then manually remove the `Include ~/.ssh/sshcloak/config` line from `~/.ssh/config`.
