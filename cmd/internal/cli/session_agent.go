package cli

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/sipuaz/sshcloak/internal/session"
)

// newSessionAgentCmd starts the hidden local session-agent process.
func newSessionAgentCmd() *cobra.Command {
	var socketPath string
	var ttl time.Duration

	cmd := &cobra.Command{
		Use:    "__session-agent",
		Hidden: true,
		Args:   cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			if socketPath == "" {
				return errors.New("session agent: --socket is required")
			}
			if ttl <= 0 {
				return errors.New("session agent: --ttl must be > 0")
			}
			if err := session.RunAgent(socketPath, ttl); err != nil {
				return fmt.Errorf("session agent: %w", err)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&socketPath, "socket", "", "unix socket path")
	cmd.Flags().DurationVar(&ttl, "ttl", 15*time.Minute, "cache ttl")
	return cmd
}
