# Getting Started

## Installation

### From source (recommended)

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

---

## First run

### 1 — Initialise the vault

The vault is an age-encrypted file that stores SSH passwords.  
Create it with a passphrase you will remember:

```
$ sshcloak vault init
Enter vault passphrase:
Confirm passphrase:
vault initialised
```

The vault is written to `~/.ssh/sshcloak/vault.age`.

### 2 — Add a host

```
$ sshcloak host add myserver \
    --hostname 203.0.113.10 \
    --user alice \
    --port 2222
host "myserver" added
```

sshcloak appends one `Include ~/.ssh/sshcloak/config` line to `~/.ssh/config`  
the first time you add a host (if it is not already present).

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
| Vault | `~/.ssh/sshcloak/vault.age` | age-encrypted password store |

All paths can be overridden with global flags — see `sshcloak --help`.

---

## Uninstalling

```bash
sudo make uninstall          # removes /usr/local/bin/sshcloak
rm -rf ~/.ssh/sshcloak       # removes managed config + vault
```

Then manually remove the `Include ~/.ssh/sshcloak/config` line from `~/.ssh/config`.
