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

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/host"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/sipuaz/sshcloak/internal/config"
	"github.com/sipuaz/sshcloak/internal/keyring"
)

// newConnectCmd returns the "sshcloak connect" command.
// It tries to retrieve a stored password for the given host label, ensures the
// host key is trusted (prompting the user if it is new), and then replaces the
// current process with either:
//
//	sshpass -e ssh <label> [extra-ssh-args...]
//
// or, when no password is stored for <label>:
//
//	ssh <label> [extra-ssh-args...]
//
// The password is passed via the SSHPASS environment variable so it does not
// appear in the process list.
//
// Extra arguments to ssh can be appended after a double-dash separator:
//
//	sshcloak connect prod -- -X -L 8080:localhost:80
//
// When no label is given an interactive host picker is shown, optionally
// filtered with --tag.
// --debug=<1|2|3> adds -v / -vv / -vvv to the underlying ssh invocation.
func newConnectCmd() *cobra.Command {
	var debugLevel int
	var tag string

	cmd := &cobra.Command{
		Use:   "connect [<label>] [-- <ssh-args>...]",
		Short: "Open an SSH session with optional automatic password injection",
		Example: "  sshcloak connect prod\n" +
			"  sshcloak connect --tag stable\n" +
			"  sshcloak connect prod -- -X -L 8080:localhost:80",
		Long: `Retrieve the stored password for <label> when available, then exec:

		sshpass -e ssh <label> [ssh-args...]

		If no password is stored for <label>, sshcloak falls back to plain:

		ssh <label> [ssh-args...]

		When password injection is used, the password is passed to sshpass via the
		SSHPASS environment variable to keep it out of the process list. Any
		arguments after -- are forwarded unchanged to ssh.

		If the host key is not yet trusted, sshcloak fetches it via ssh-keyscan,
		displays the fingerprint, and asks for confirmation before adding it to
		~/.ssh/known_hosts — exactly as OpenSSH would.

		When no label is provided an interactive list of managed hosts is shown;
		use ↑/↓ to navigate, Enter to select, and q to cancel.  Use --tag to
		filter that picker by one metadata tag.

		sshpass is required only when a stored password exists for the selected host:
		apt install sshpass
		brew install hudochenkov/sshpass/sshpass
		dnf install sshpass`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if debugLevel < 0 || debugLevel > 3 {
				return fmt.Errorf("connect: --debug level must be between 0 and 3")
			}

			// Split positional args from extra ssh args around the -- separator.
			var labelArgs, extraSSHArgs []string
			if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
				labelArgs = args[:dashIdx]
				extraSSHArgs = args[dashIdx:]
			} else {
				labelArgs = args
			}

			// Resolve the label — interactively if not supplied.
			var label string
			if len(labelArgs) == 0 {
				mgr := host.GetManager()
				hosts, err := mgr.ListHosts()
				if err != nil {
					return fmt.Errorf("connect: list hosts: %w", err)
				}
				if tag != "" {
					labels, err := host.GetMetadata().TaggedHosts(tag)
					if err != nil {
						return fmt.Errorf("connect: filter hosts by tag: %w", err)
					}
					allowed := make(map[string]struct{}, len(labels))
					for _, label := range labels {
						allowed[label] = struct{}{}
					}
					filtered := make([]config.HostSpec, 0, len(hosts))
					for _, candidate := range hosts {
						if _, ok := allowed[candidate.Label]; ok {
							filtered = append(filtered, candidate)
						}
					}
					hosts = filtered
				}
				if len(hosts) == 0 {
					if tag == "" {
						return errors.New("connect: no managed hosts found; add one with 'sshcloak host add'")
					}
					return fmt.Errorf("connect: no managed hosts found with tag %q", tag)
				}
				label, err = pickHost(hosts)
				if err != nil {
					return err
				}
			} else {
				label = labelArgs[0]
			}

			sshPath, err := exec.LookPath("ssh")
			if err != nil {
				return errors.New("ssh not found in PATH")
			}

			// Resolve the real hostname and port from the sshcloak-managed config
			// so ssh-keyscan can reach the host directly.
			checkHostname, checkPort := resolveHostTarget(label)

			// Ensure the host key is present in known_hosts before handing off
			// to ssh/sshpass.
			if err := ensureKnownHost(checkHostname, checkPort); err != nil {
				return err
			}

			store, err := vault.UnlockStore("Vault passphrase: ")
			if err != nil {
				return fmt.Errorf("connect: unlock vault: %w", err)
			}
			// Lock zeroes the in-memory passphrase; call it on any error path.
			// On the success path the process is replaced by syscall.Exec so
			// deferred calls do not run — that is intentional and safe.

			record, err := store.Get(label, keyring.SecretKindPassword)
			usePassword := true
			password := ""
			if err != nil {
				if errors.Is(err, keyring.ErrSecretNotFound) {
					usePassword = false
				} else {
					store.Lock()
					return fmt.Errorf("connect: retrieve password: %w", err)
				}
			} else {
				password = record.Value
			}

			sshpassPath := ""
			if usePassword {
				// Locate sshpass only when password injection is needed.
				sshpassPath, err = exec.LookPath("sshpass")
				if err != nil {
					store.Lock()
					return errors.New(
						"sshpass not found in PATH\n" +
							"  apt install sshpass\n" +
							"  brew install hudochenkov/sshpass/sshpass\n" +
							"  dnf install sshpass",
					)
				}
			}

			program, argv, env, err := buildConnectExecPlan(
				sshPath,
				sshpassPath,
				label,
				debugLevel,
				extraSSHArgs,
				os.Environ(),
				password,
				usePassword,
			)
			if err != nil {
				store.Lock()
				return err
			}

			// Replace this process entirely. The SSH session inherits the
			// current terminal, so interactive usage (pty, signals, resize)
			// all work correctly.
			return syscall.Exec(program, argv, env)
		},
	}

	cmd.Flags().IntVar(&debugLevel, "debug", 0, "SSH verbosity level: 1=-v, 2=-vv, 3=-vvv")
	cmd.Flags().StringVar(&tag, "tag", "", "filter the interactive host picker by one tag when no label is specified")

	return cmd
}

// buildConnectExecPlan returns the final program, argv, and env for connect.
// In password mode it wraps ssh with sshpass and injects SSHPASS.
func buildConnectExecPlan(
	sshPath, sshpassPath, label string,
	debugLevel int,
	extraSSHArgs, baseEnv []string,
	password string,
	usePassword bool,
) (program string, argv, env []string, err error) {
	sshArgs := buildSSHInvocation(sshPath, label, debugLevel, extraSSHArgs)
	env = append([]string(nil), baseEnv...)

	if !usePassword {
		return sshPath, sshArgs, env, nil
	}
	if sshpassPath == "" {
		return "", nil, nil, errors.New("connect: sshpass path is required in password mode")
	}
	if password == "" {
		return "", nil, nil, errors.New("connect: empty password for password mode")
	}

	argv = append([]string{"sshpass", "-e"}, sshArgs...)
	env = append(env, "SSHPASS="+password)
	return sshpassPath, argv, env, nil
}

// buildSSHInvocation returns argv for direct ssh execution.
func buildSSHInvocation(sshPath, label string, debugLevel int, extraSSHArgs []string) []string {
	// StrictHostKeyChecking=yes is safe here because ensureKnownHost has already
	// verified and recorded the key.
	sshArgs := []string{sshPath, "-o", "StrictHostKeyChecking=yes"}

	// Append -v / -vv / -vvv when a debug level was requested.
	if debugLevel > 0 {
		sshArgs = append(sshArgs, "-"+strings.Repeat("v", debugLevel))
	}

	sshArgs = append(sshArgs, label)
	sshArgs = append(sshArgs, extraSSHArgs...)
	return sshArgs
}

// ANSI escape sequences used by the host picker.
const (
	ansiCursorUpFmt      = "\033[%dA\r"       // move cursor up N lines and go to column 0
	ansiEraseDown        = "\033[J"           // erase from cursor to end of screen
	ansiCursorUpEraseFmt = "\033[%dA\r\033[J" // move up N lines then erase to end of screen
	ansiColorCyan        = "\033[1;36m"       // bold cyan — used for the selected row
	ansiColorReset       = "\033[0m"          // reset all SGR attributes
)

// Unicode symbols used in the host picker prompt.
const (
	unicodeArrowUp   = "\u2191" // ↑
	unicodeArrowDown = "\u2193" // ↓
)

// Key codes used by the host picker.
const (
	keyCtrlC          = 3 // Ctrl-C (ETX)
	keyCarriageReturn = '\r'
	keyLineFeed       = '\n'
	keyEsc            = 0x1b // ESC — start of CSI escape sequences
	keyCSI            = '['  // CSI introducer that follows ESC
	keyCursorUp       = 'A'  // final byte of ESC [ A
	keyCursorDown     = 'B'  // final byte of ESC [ B
)

// pickHost displays an interactive list of managed hosts in the terminal and
// returns the label selected by the user.  Navigation: ↑/↓ arrows, Enter to
// confirm, q to cancel.  The controlling terminal (/dev/tty) is used directly
// so the picker works even when stdin/stdout are redirected.
func pickHost(hosts []config.HostSpec) (string, error) {
	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
	if err != nil {
		return "", fmt.Errorf("connect: cannot open terminal for host picker: %w", err)
	}
	defer tty.Close()

	fd := int(tty.Fd())
	oldState, err := term.MakeRaw(fd)
	if err != nil {
		return "", fmt.Errorf("connect: cannot set raw terminal mode: %w", err)
	}
	defer term.Restore(fd, oldState) //nolint:errcheck

	cursor := 0
	linesDrawn := 0

	redraw := func() {
		if linesDrawn > 0 {
			// Move cursor up to overwrite previous render.
			fmt.Fprintf(tty, ansiCursorUpFmt, linesDrawn)
		}
		// Erase from cursor to end of screen.
		fmt.Fprint(tty, ansiEraseDown)

		fmt.Fprintf(tty, "Select a host (%s %s arrows, Enter to connect, q to quit):\r\n", unicodeArrowUp, unicodeArrowDown)
		for i, h := range hosts {
			desc := h.HostName
			if h.User != "" {
				desc = h.User + "@" + h.HostName
			}
			if i == cursor {
				fmt.Fprintf(tty, "  %s> %-20s  %s%s\r\n", ansiColorCyan, h.Label, desc, ansiColorReset)
			} else {
				fmt.Fprintf(tty, "    %-20s  %s\r\n", h.Label, desc)
			}
		}
		linesDrawn = len(hosts) + 1
	}

	redraw()

	buf := make([]byte, 4)
	for {
		n, err := tty.Read(buf)
		if err != nil {
			return "", fmt.Errorf("connect: read key: %w", err)
		}

		switch {
		case n == 1 && (buf[0] == 'q' || buf[0] == 'Q' || buf[0] == keyCtrlC):
			// Clear the picker before returning.
			if linesDrawn > 0 {
				fmt.Fprintf(tty, ansiCursorUpEraseFmt, linesDrawn)
			}
			return "", errors.New("connect: cancelled")

		case n == 1 && (buf[0] == keyCarriageReturn || buf[0] == keyLineFeed):
			// Clear the picker before handing back control.
			if linesDrawn > 0 {
				fmt.Fprintf(tty, ansiCursorUpEraseFmt, linesDrawn)
			}
			return hosts[cursor].Label, nil

		case n >= 3 && buf[0] == keyEsc && buf[1] == keyCSI && buf[2] == keyCursorUp:
			if cursor > 0 {
				cursor--
				redraw()
			}

		case n >= 3 && buf[0] == keyEsc && buf[1] == keyCSI && buf[2] == keyCursorDown:
			if cursor < len(hosts)-1 {
				cursor++
				redraw()
			}
		}
	}
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
