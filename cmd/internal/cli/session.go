package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
)

// newSessionCmd returns the "sshcloak session" command group.
func newSessionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "session",
		Short: "Manage vault unlock session cache",
		Long:  "Inspect and clear sudo-like vault unlock cache for this shell session.",
	}
	cmd.AddCommand(
		newSessionStatusCmd(),
		newSessionLockCmd(),
		newSessionClearCmd(),
	)
	return cmd
}

func newSessionStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show whether this shell has an active vault unlock cache",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			cached, err := vault.SessionStatus()
			if err != nil {
				return fmt.Errorf("session status: %w", err)
			}
			if cached {
				fmt.Fprintln(cmd.OutOrStdout(), "vault session unlocked")
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "vault session locked")
			return nil
		},
	}
}

func newSessionLockCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lock",
		Short: "Remove unlock cache for this shell session",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := vault.LockSession(); err != nil {
				return fmt.Errorf("session lock: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "vault session locked")
			return nil
		},
	}
}

func newSessionClearCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "clear",
		Short: "Clear all vault unlock cache entries",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := vault.ClearSessions(); err != nil {
				return fmt.Errorf("session clear: %w", err)
			}
			fmt.Fprintln(cmd.OutOrStdout(), "all vault sessions cleared")
			return nil
		},
	}
}
