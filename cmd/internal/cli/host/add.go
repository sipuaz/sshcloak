package host

import (
	"fmt"

	"github.com/sipuaz/sshcloak/internal/config"
	"github.com/spf13/cobra"
)

// newAddCmd returns the "sshcloak host add <label>" command.
// All fields are optional except the positional label argument; a host with
// only a label and no directives is valid SSH config.
func newAddCmd() *cobra.Command {
	var (
		hostname      string
		user          string
		port          string
		identityFiles []string
		extraPairs    []string
	)

	cmd := &cobra.Command{
		Use:   "add <label>",
		Short: "Add a new managed SSH host",
		Long: `Add a new Host block to the sshcloak-managed include file.

The label must be an exact alias (no wildcards).  If ~/.ssh/config does not yet
contain an Include directive for the managed file, it is appended automatically.

Extra SSH directives can be supplied as KEY=VALUE pairs via --extra, for example:
  --extra ProxyJump=bastion --extra ServerAliveInterval=30`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			extra, err := parseExtraPairs(extraPairs)
			if err != nil {
				return err
			}

			spec := config.HostSpec{
				Label:         label,
				HostName:      hostname,
				User:          user,
				Port:          port,
				IdentityFiles: identityFiles,
				Extra:         extra,
			}

			if err := getManager().AddHost(spec); err != nil {
				return fmt.Errorf("add host: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "host %q added\n", label)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "remote hostname or IP address (HostName)")
	cmd.Flags().StringVar(&user, "user", "", "remote username (User)")
	cmd.Flags().StringVar(&port, "port", "", "remote port (Port)")
	cmd.Flags().StringArrayVar(&identityFiles, "identity-file", nil, "path to an identity file; repeatable (IdentityFile)")
	cmd.Flags().StringArrayVar(&extraPairs, "extra", nil, "additional directive as KEY=VALUE; repeatable")

	return cmd
}
