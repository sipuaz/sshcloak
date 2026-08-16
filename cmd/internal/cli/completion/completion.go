// Package completion implements the "sshcloak completion" command group
// for shell completion support (bash and zsh).
package completion

import (
	"github.com/spf13/cobra"

	"github.com/sipuaz/sshcloak/internal/config"
)

// sharedManager is the config manager instance injected by root.PersistentPreRunE.
var sharedManager *config.Manager

// SetManager is called by root.PersistentPreRunE to inject the manager before
// any completion sub-command runs.
func SetManager(_ *cobra.Command, mgr *config.Manager) {
	sharedManager = mgr
}

// getManager retrieves the injected manager, panicking if SetManager was never
// called (which would indicate a wiring bug, not a user error).
func getManager() *config.Manager {
	if sharedManager == nil {
		panic("sshcloak: manager not initialised — this is a bug")
	}
	return sharedManager
}

// NewCompletionCmd returns the parent "completion" command.
func NewCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion",
		Short: "Print completion functions for bash and zsh",
		Long: `Output shell completion functions for bash and zsh.

Completion is delivered as separate scripts that can be sourced in your shell configuration:
  - Bash: source in ~/.bashrc
  - Zsh: source in ~/.zshrc

Use 'sshcloak completion list-hosts' to list managed host aliases.
Use 'sshcloak completion flags <command>' to list flags for a command.`,
		Hidden: true, // Internal command for shell completion
	}

	cmd.AddCommand(
		newListHostsCmd(),
		newFlagsCmd(),
	)

	return cmd
}
