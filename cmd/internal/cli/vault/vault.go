// Package vault implements the "sshcloak vault" command group and the vault
// store injection mechanism used by password sub-commands.
package vault

import (
	"time"

	"github.com/spf13/cobra"

	"github.com/sipuaz/sshcloak/internal/keyring"
	"github.com/sipuaz/sshcloak/internal/session"
)

// sharedStore is the vault store instance injected by root.PersistentPreRunE.
// CLI commands run sequentially, so a package-level pointer is safe here.
var sharedStore *keyring.FileVaultStore
var sharedSessionConfig SessionConfig

// SessionConfig controls sudo-like vault unlock caching behavior.
type SessionConfig struct {
	Enabled bool
	TTL     time.Duration
	Client  *session.Client
}

// SetStore is called by root.PersistentPreRunE to inject the vault store before
// any sub-command that needs it runs.
func SetStore(_ *cobra.Command, store *keyring.FileVaultStore) {
	sharedStore = store
}

// SetSession injects session-caching dependencies before sub-commands run.
func SetSession(_ *cobra.Command, config SessionConfig) {
	sharedSessionConfig = config
}

// getStore returns the injected vault store, panicking on a wiring bug.
func getStore() *keyring.FileVaultStore {
	if sharedStore == nil {
		panic("sshcloak: vault store not initialised — this is a bug")
	}
	return sharedStore
}

// GetStore is the exported accessor used by sibling command packages (e.g.
// the password package) that share the same injected vault store.
func GetStore() *keyring.FileVaultStore {
	return getStore()
}

// NewVaultCmd returns the parent "vault" command.  It has no Run of its own.
func NewVaultCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "vault",
		Short: "Manage the encrypted secret vault",
		Long:  "Initialise, unlock, and rotate the passphrase of the sshcloak encrypted vault.",
	}

	cmd.AddCommand(
		newInitCmd(),
		newRotateCmd(),
	)

	return cmd
}
