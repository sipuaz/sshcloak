// Package cli wires together the Cobra command tree and the shared application
// context that carries the config.Manager instance across sub-commands.
package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spf13/cobra"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/completion"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/host"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/password"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/sipuaz/sshcloak/internal/config"
	"github.com/sipuaz/sshcloak/internal/keyring"
	"github.com/sipuaz/sshcloak/internal/metadata"
	"github.com/sipuaz/sshcloak/internal/session"
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
	var metaPath string
	var vaultPath string
	var sessionTTL time.Duration
	var disableSessionCache bool

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
	root.PersistentFlags().StringVar(
		&metaPath,
		"meta",
		defaultMetaPath(),
		"path to the sshcloak host metadata file",
	)
	root.PersistentFlags().DurationVar(
		&sessionTTL,
		"session-ttl",
		15*time.Minute,
		"duration before cached vault unlock expires",
	)
	root.PersistentFlags().BoolVar(
		&disableSessionCache,
		"no-session-cache",
		false,
		"disable sudo-like vault unlock cache and prompt every command",
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
		completion.SetManager(cmd, mgr)
		host.SetMetadataStore(cmd, metadata.NewStore(config.NewFileHandler(), metaPath))

		store := keyring.NewFileVaultStore(vaultPath)
		vault.SetStore(cmd, store)
		socketPath, err := session.DefaultSocketPath(vaultPath)
		if err != nil {
			return fmt.Errorf("resolve session socket path: %w", err)
		}
		vault.SetSession(cmd, vault.SessionConfig{
			Enabled: !disableSessionCache,
			TTL:     sessionTTL,
			Client:  session.NewClient(socketPath, sessionTTL),
		})

		return nil
	}

	// Attach command groups.
	root.AddCommand(
		newInitCmd(),
		newConnectCmd(),
		newSessionCmd(),
		newSessionAgentCmd(),
		newVersionCmd(),
		host.NewHostCmd(),
		vault.NewVaultCmd(),
		password.NewPasswordCmd(),
		completion.NewCompletionCmd(),
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

// defaultMetaPath returns ~/.ssh/sshcloak/meta.yaml.
func defaultMetaPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".ssh", "sshcloak", "meta.yaml")
	}
	return filepath.Join(home, ".ssh", "sshcloak", "meta.yaml")
}
