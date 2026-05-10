package host

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newRemoveCmd returns the "sshcloak host remove <label>" command.
func newRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "remove <label>",
		Aliases: []string{"rm", "delete"},
		Short:   "Remove a managed SSH host",
		Long:    "Delete the Host block for <label> from the sshcloak-managed include file.",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			if err := getManager().DeleteHost(label); err != nil {
				return fmt.Errorf("remove host: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "host %q removed\n", label)
			return nil
		},
	}
}
