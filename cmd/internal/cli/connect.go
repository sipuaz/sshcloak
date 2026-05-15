package cli

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/host"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/sipuaz/sshcloak/internal/keyring"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newConnectCmd returns the "sshcloak connect" command.
// It unlocks the vault, retrieves the stored password for the given host label,
// ensures the host key is trusted (prompting the user if it is new), and then
// replaces the current process with:
//
//	sshpass -e ssh <label> [extra-ssh-args...]
//
// The password is passed via the SSHPASS environment variable so it does not
// appear in the process list.  sshpass must be installed and on PATH.
//
// Extra arguments to ssh can be appended after a double-dash separator:
//
//	sshcloak connect prod -- -X -L 8080:localhost:80
func newConnectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "connect <label> [-- <ssh-args>...]",
		Short: "Open an SSH session with automatic password injection",
		Long: `Unlock the vault, retrieve the stored password for <label>, and exec:

  sshpass -e ssh <label> [ssh-args...]

The password is passed to sshpass via the SSHPASS environment variable to
keep it out of the process list.  Any arguments after -- are forwarded
unchanged to ssh.

If the host key is not yet trusted, sshcloak fetches it via ssh-keyscan,
displays the fingerprint, and asks for confirmation before adding it to
~/.ssh/known_hosts — exactly as OpenSSH would.

sshpass must be installed:
  apt install sshpass
  brew install hudochenkov/sshpass/sshpass
  dnf install sshpass`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			// Collect any extra ssh arguments supplied after --.
			var extraSSHArgs []string
			if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
				extraSSHArgs = args[dashIdx:]
			}

			// Locate sshpass early so we fail before prompting for the passphrase.
			sshpassPath, err := exec.LookPath("sshpass")
			if err != nil {
				return errors.New(
					"sshpass not found in PATH\n" +
						"  apt install sshpass\n" +
						"  brew install hudochenkov/sshpass/sshpass\n" +
						"  dnf install sshpass",
				)
			}

			sshPath, err := exec.LookPath("ssh")
			if err != nil {
				return errors.New("ssh not found in PATH")
			}

			// Resolve the real hostname and port from the sshcloak-managed config
			// so ssh-keyscan can reach the host directly.
			checkHostname, checkPort := resolveHostTarget(label)

			// Ensure the host key is present in known_hosts before handing off
			// to sshpass (which cannot handle interactive fingerprint prompts).
			if err := ensureKnownHost(checkHostname, checkPort); err != nil {
				return err
			}

			// Prompt for the vault passphrase (reads directly from the terminal).
			vaultPass, err := readConnectPassphrase("Vault passphrase: ")
			if err != nil {
				return err
			}

			store := vault.GetStore()
			if err := store.Unlock(vaultPass); err != nil {
				return fmt.Errorf("connect: unlock vault: %w", err)
			}
			// Lock zeroes the in-memory passphrase; call it on any error path.
			// On the success path the process is replaced by syscall.Exec so
			// deferred calls do not run — that is intentional and safe.

			record, err := store.Get(label, keyring.SecretKindPassword)
			if err != nil {
				store.Lock()
				return fmt.Errorf("connect: retrieve password: %w", err)
			}

			// Build argv for sshpass.  StrictHostKeyChecking=yes is safe here
			// because ensureKnownHost has already verified and recorded the key.
			argv := append(
				[]string{"sshpass", "-e", sshPath, "-o", "StrictHostKeyChecking=yes", label},
				extraSSHArgs...,
			)

			// Inject the password via environment — not via a -p flag — so it
			// does not appear in the output of `ps`.
			env := append(os.Environ(), "SSHPASS="+record.Value)

			// Replace this process entirely.  The SSH session inherits the
			// current terminal, so interactive usage (pty, signals, resize)
			// all work correctly.
			return syscall.Exec(sshpassPath, argv, env)
		},
	}

	return cmd
}

// resolveHostTarget returns the hostname and port to use for ssh-keyscan.
// It looks up the label in the sshcloak-managed config; if the label is not
// found (e.g. a host defined only in the user's own ~/.ssh/config), it falls
// back to using the label itself as the hostname on port 22.
func resolveHostTarget(label string) (hostname, port string) {
	mgr := host.GetManager()
	spec, err := mgr.GetHost(label)
	if err != nil {
		return label, "22"
	}
	hostname = spec.HostName
	if hostname == "" {
		hostname = label
	}
	port = spec.Port
	if port == "" {
		port = "22"
	}
	return hostname, port
}

// ensureKnownHost checks whether <hostname>:<port> is already present in
// ~/.ssh/known_hosts.  If it is not, it fetches the host key via ssh-keyscan,
// displays the fingerprint in the standard OpenSSH format, and prompts the
// user for confirmation before appending the key to known_hosts.
func ensureKnownHost(hostname, port string) error {
	// ssh-keygen -F expects [host]:port for non-standard ports.
	lookupTarget := hostname
	if port != "22" {
		lookupTarget = fmt.Sprintf("[%s]:%s", hostname, port)
	}

	// Check known_hosts — exit code 0 means the key is already trusted.
	out, err := exec.Command("ssh-keygen", "-F", lookupTarget).Output()
	if err == nil && len(out) > 0 {
		return nil
	}

	// Fetch the host key(s).
	keyscanArgs := []string{"-p", port, hostname}
	keyscanOut, err := exec.Command("ssh-keyscan", keyscanArgs...).Output()
	if err != nil || len(strings.TrimSpace(string(keyscanOut))) == 0 {
		return fmt.Errorf("connect: could not fetch host key for %s:%s", hostname, port)
	}

	// Write the raw key to a temp file so ssh-keygen can compute its fingerprint.
	tmp, err := os.CreateTemp("", "sshcloak-keyscan-*")
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(keyscanOut); err != nil {
		tmp.Close()
		return fmt.Errorf("connect: %w", err)
	}
	tmp.Close()

	fingerprintOut, err := exec.Command("ssh-keygen", "-lf", tmp.Name()).Output()
	if err != nil {
		return fmt.Errorf("connect: compute fingerprint: %w", err)
	}

	// Prompt the user — mirror the OpenSSH message so it feels familiar.
	fmt.Fprintf(os.Stderr, "The authenticity of host '%s' can't be established.\n", lookupTarget)
	for _, line := range strings.Split(strings.TrimSpace(string(fingerprintOut)), "\n") {
		fmt.Fprintf(os.Stderr, "%s\n", line)
	}
	fmt.Fprint(os.Stderr, "Are you sure you want to continue connecting (yes/no)? ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	if strings.TrimSpace(strings.ToLower(answer)) != "yes" {
		return errors.New("connect: host key verification failed")
	}

	// Append the key to known_hosts.
	home, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("connect: %w", err)
	}
	knownHostsPath := filepath.Join(home, ".ssh", "known_hosts")
	f, err := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("connect: write known_hosts: %w", err)
	}
	defer f.Close()
	if _, err := f.Write(keyscanOut); err != nil {
		return fmt.Errorf("connect: write known_hosts: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Warning: Permanently added '%s' to the list of known hosts.\n", lookupTarget)
	return nil
}

// readConnectPassphrase reads a passphrase from the controlling terminal without
// echo.  It is a thin wrapper around term.ReadPassword so connect.go does not
// depend on init.go internals.
func readConnectPassphrase(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", errors.New("connect: stdin is not a terminal; cannot read vault passphrase")
	}
	fmt.Fprint(os.Stderr, prompt)
	raw, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("connect: read passphrase: %w", err)
	}
	return string(raw), nil
}
