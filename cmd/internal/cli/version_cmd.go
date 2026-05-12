// Package cli wires together the Cobra command tree and the shared application
// context that carries the config.Manager instance across sub-commands.
package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newVersionCmd returns the explicit "sshcloak version" command.
// Cobra already exposes -v/--version; this command makes the version visible
// as a normal sub-command too.
func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the sshcloak version",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			fmt.Fprintf(cmd.OutOrStdout(), "sshcloak version %s\n", Version)
			return nil
		},
	}
}
