package host

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newGetCmd returns the "sshcloak host get <label>" command.
// Output is one key: value line per directive, suitable for scripting.
func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <label>",
		Short: "Show details of a managed SSH host",
		Long:  "Print every directive of one managed Host block, one key: value per line.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			host, err := getManager().GetHost(label)
			if err != nil {
				return fmt.Errorf("get host: %w", err)
			}
			tags, err := getMetadata().Tags(label)
			if err != nil {
				return fmt.Errorf("get host: %w", err)
			}

			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "Label:    %s\n", host.Label)
			fmt.Fprintf(out, "HostName: %s\n", host.HostName)
			fmt.Fprintf(out, "User:     %s\n", host.User)
			fmt.Fprintf(out, "Port:     %s\n", host.Port)
			fmt.Fprintf(out, "Tags:     %s\n", joinTags(tags))

			for _, id := range host.IdentityFiles {
				fmt.Fprintf(out, "IdentityFile: %s\n", id)
			}

			for _, key := range sortedKeys(host.Extra) {
				for _, val := range host.Extra[key] {
					fmt.Fprintf(out, "%s: %s\n", key, val)
				}
			}

			return nil
		},
	}
}
