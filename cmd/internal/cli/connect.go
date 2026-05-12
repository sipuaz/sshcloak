package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/sipuaz/sshcloak/internal/keyring"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newConnectCmd returns the "sshcloak connect" command.
// It unlocks the vault, retrieves the stored password for the given host label,
// and replaces the current process with:
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
	return &cobra.Command{
		Use:   "connect <label> [-- <ssh-args>...]",
		Short: "Open an SSH session with automatic password injection",
		Long: `Unlock the vault, retrieve the stored password for <label>, and exec:
		sshpass -e ssh <label> [ssh-args...]

		The password is passed to sshpass via the SSHPASS environment variable to
		keep it out of the process list.  Any arguments after -- are forwarded
		unchanged to ssh.

		sshpass must be installed:
		  apt install sshpass
		  brew install hudochenkov/sshpass/sshpass
		  dnf install sshpass`,
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: false,
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
						"  apt   install sshpass\n" +
						"  brew  install hudochenkov/sshpass/sshpass\n" +
						"  dnf   install sshpass",
				)
			}

			sshPath, err := exec.LookPath("ssh")
			if err != nil {
				return errors.New("ssh not found in PATH")
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

			// Build the argv for sshpass.
			// argv[0] must be the program name as seen by the child.
			argv := append([]string{"sshpass", "-e", sshPath, label}, extraSSHArgs...)

			// Inject the password via environment — not via a -p flag — so it
			// does not appear in the output of `ps`.
			env := append(os.Environ(), "SSHPASS="+record.Value)

			// Replace this process entirely.  The SSH session inherits the
			// current terminal, so interactive usage (pty, signals, resize)
			// all work correctly.
			return syscall.Exec(sshpassPath, argv, env)
		},
	}
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
