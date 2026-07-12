package vault

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"

	"github.com/sipuaz/sshcloak/internal/keyring"
)

// UnlockStore unlocks the shared vault store, using session cache when enabled.
func UnlockStore(prompt string) (*keyring.FileVaultStore, error) {
	store := getStore()
	if !sharedSessionConfig.Enabled || sharedSessionConfig.Client == nil {
		passphrase, err := readPassphrase(prompt)
		if err != nil {
			return nil, err
		}
		if err := store.Unlock(passphrase); err != nil {
			return nil, err
		}
		return store, nil
	}

	sessionID, err := currentSessionID()
	if err != nil {
		return nil, err
	}

	client := sharedSessionConfig.Client
	if err := client.EnsureAgent(); err != nil {
		return nil, fmt.Errorf("start session cache: %w", err)
	}

	if passphrase, found, err := client.Get(sessionID); err == nil && found {
		if err := store.Unlock(passphrase); err == nil {
			return store, nil
		}
		_ = client.Lock(sessionID)
	}

	passphrase, err := readPassphrase(prompt)
	if err != nil {
		return nil, err
	}
	if err := store.Unlock(passphrase); err != nil {
		return nil, err
	}
	if err := client.Set(sessionID, passphrase); err != nil {
		store.Lock()
		return nil, fmt.Errorf("update session cache: %w", err)
	}
	return store, nil
}

// SessionStatus reports whether the current shell has a valid cache entry.
func SessionStatus() (bool, error) {
	if !sharedSessionConfig.Enabled || sharedSessionConfig.Client == nil {
		return false, nil
	}
	sessionID, err := currentSessionID()
	if err != nil {
		return false, err
	}
	client := sharedSessionConfig.Client
	if err := client.EnsureAgent(); err != nil {
		return false, err
	}
	return client.Status(sessionID)
}

// LockSession removes cache for current shell session.
func LockSession() error {
	if !sharedSessionConfig.Enabled || sharedSessionConfig.Client == nil {
		return nil
	}
	sessionID, err := currentSessionID()
	if err != nil {
		return err
	}
	client := sharedSessionConfig.Client
	if err := client.EnsureAgent(); err != nil {
		return err
	}
	return client.Lock(sessionID)
}

// ClearSessions removes all cached entries from the local agent.
func ClearSessions() error {
	if !sharedSessionConfig.Enabled || sharedSessionConfig.Client == nil {
		return nil
	}
	client := sharedSessionConfig.Client
	if err := client.EnsureAgent(); err != nil {
		return err
	}
	return client.Clear()
}

func currentSessionID() (string, error) {
	ttyPath, err := os.Readlink("/proc/self/fd/0")
	if err != nil {
		return "", fmt.Errorf("read tty for session cache: %w", err)
	}
	return "ppid:" + strconv.Itoa(os.Getppid()) + "|tty:" + ttyPath, nil
}

func readPassphrase(prompt string) (string, error) {
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("reading passphrase: stdin is not a terminal")
	}
	fmt.Fprint(os.Stderr, prompt)
	raw, err := term.ReadPassword(fd)
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(raw), nil
}
