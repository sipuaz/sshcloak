// Package cli wires together the Cobra command tree and the shared application
// context that carries the config.Manager instance across sub-commands.
package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/host"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/password"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/sipuaz/sshcloak/internal/config"
	"github.com/sipuaz/sshcloak/internal/keyring"
	"github.com/spf13/cobra"
)

// Execute builds the root command and runs it.  It is the only entry point
// called from main.go.
func Execute() error {
	return newRootCmd().Execute()
}

// newRootCmd constructs the root cobra.Command, registers persistent flags,
// and attaches all sub-command groups.
func newRootCmd() *cobra.Command {
	var userConfigPath string
	var managedConfigPath string
	var vaultPath string

	root := &cobra.Command{
		Use:     "sshcloak",
		Short:   "Transparent SSH credential manager",
		Version: Version,
		Long: `sshcloak wraps the ssh command to inject passwords silently from the OS
keyring and manages host entries in ~/.ssh/config via a dedicated include file.`,
		// SilenceUsage prevents Cobra from printing usage on every error.
		SilenceUsage: true,
	}

	// Persistent flags available to every sub-command.
	root.PersistentFlags().StringVar(
		&userConfigPath,
		"config",
		defaultUserConfigPath(),
		"path to the root SSH config file",
	)
	root.PersistentFlags().StringVar(
		&managedConfigPath,
		"managed-config",
		defaultManagedConfigPath(),
		"path to the sshcloak-managed include file",
	)
	root.PersistentFlags().StringVar(
		&vaultPath,
		"vault",
		defaultVaultPath(),
		"path to the encrypted vault file",
	)

	// Build the shared app context after flags have been parsed.
	// PersistentPreRunE runs before every sub-command's Run.
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		mgr := config.NewManager(
			config.NewFileHandler(),
			userConfigPath,
			managedConfigPath,
		)
		host.SetManager(cmd, mgr)

		store := keyring.NewFileVaultStore(vaultPath)
		vault.SetStore(cmd, store)

		return nil
	}

	// Attach command groups.
	root.AddCommand(
		host.NewHostCmd(),
		vault.NewVaultCmd(),
		password.NewPasswordCmd(),
	)

	return root
}

// defaultUserConfigPath returns ~/.ssh/config, expanding the home directory.
func defaultUserConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		fmt.Fprintln(os.Stderr, "warning: cannot determine home directory, using relative path")
		return filepath.Join(".ssh", "config")
	}
	return filepath.Join(home, ".ssh", "config")
}

// defaultManagedConfigPath returns ~/.ssh/sshcloak/config.
func defaultManagedConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".ssh", "sshcloak", "config")
	}
	return filepath.Join(home, ".ssh", "sshcloak", "config")
}

// defaultVaultPath returns ~/.ssh/sshcloak/vault.age.
func defaultVaultPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".ssh", "sshcloak", "vault.age")
	}
	return filepath.Join(home, ".ssh", "sshcloak", "vault.age")
}
