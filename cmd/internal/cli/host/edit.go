package host

import (
	"fmt"

	"github.com/sipuaz/sshcloak/internal/config"
	"github.com/spf13/cobra"
)

// newEditCmd returns the "sshcloak host edit <label>" command.
// Only flags that are explicitly provided overwrite the current value (merge-
// patch semantics).  Omitted flags leave the existing directive unchanged.
func newEditCmd() *cobra.Command {
	var (
		hostname      string
		user          string
		port          string
		identityFiles []string
		extraPairs    []string
	)

	cmd := &cobra.Command{
		Use:   "edit <label>",
		Short: "Edit a managed SSH host",
		Long: `Update one or more directives of an existing managed Host block.

Only the flags you supply are changed; everything else is preserved.
To clear a single-valued field pass an empty string, e.g. --port "".
To fully replace the identity-file list supply all desired paths via
repeated --identity-file flags.

Extra directives follow the same KEY=VALUE syntax as "host add".`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			// Load the current state so unset flags keep their values.
			current, err := getManager().GetHost(label)
			if err != nil {
				return fmt.Errorf("edit host: %w", err)
			}

			// Merge: only overwrite fields whose flag was explicitly set.
			flags := cmd.Flags()
			merged := config.HostSpec{
				Label:         label,
				HostName:      current.HostName,
				User:          current.User,
				Port:          current.Port,
				IdentityFiles: current.IdentityFiles,
				Extra:         current.Extra,
			}

			if flags.Changed("hostname") {
				merged.HostName = hostname
			}
			if flags.Changed("user") {
				merged.User = user
			}
			if flags.Changed("port") {
				merged.Port = port
			}
			if flags.Changed("identity-file") {
				merged.IdentityFiles = identityFiles
			}
			if flags.Changed("extra") {
				extra, err := parseExtraPairs(extraPairs)
				if err != nil {
					return err
				}
				// Merge extra map: new keys overwrite, unmentioned keys are kept.
				if merged.Extra == nil {
					merged.Extra = make(map[string][]string)
				}
				for k, v := range extra {
					merged.Extra[k] = v
				}
			}

			if err := getManager().UpdateHost(label, merged); err != nil {
				return fmt.Errorf("edit host: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "host %q updated\n", label)
			return nil
		},
	}

	cmd.Flags().StringVar(&hostname, "hostname", "", "new HostName value")
	cmd.Flags().StringVar(&user, "user", "", "new User value")
	cmd.Flags().StringVar(&port, "port", "", "new Port value")
	cmd.Flags().StringArrayVar(&identityFiles, "identity-file", nil, "replace the full IdentityFile list; repeatable")
	cmd.Flags().StringArrayVar(&extraPairs, "extra", nil, "add or overwrite extra directives as KEY=VALUE; repeatable")

	return cmd
}
