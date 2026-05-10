// Package password implements the "sshcloak password" command group.
package password

import (
	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/spf13/cobra"
)

// NewPasswordCmd returns the parent "password" command.
func NewPasswordCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "password",
		Short: "Manage SSH passwords in the encrypted vault",
		Long:  "Store, retrieve, and delete SSH passwords for managed host labels.",
	}

	cmd.AddCommand(
		newSetCmd(),
		newGetCmd(),
		newDeleteCmd(),
		newListCmd(),
	)

	return cmd
}

// getStore is a convenience alias so sub-commands in this package can access
// the vault store without importing the vault package internals directly.
var getStore = vault.GetStore
