package host

import (
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

// newListCmd returns the "sshcloak host list" command.
// Output is a tab-aligned table sorted by label.
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all managed SSH hosts",
		Long:  "Print a tab-aligned table of every Host block in the sshcloak-managed config.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			hosts, err := getManager().ListHosts()
			if err != nil {
				return fmt.Errorf("list hosts: %w", err)
			}

			if len(hosts) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no managed hosts found")
				return nil
			}

			w := newTabWriter(cmd.OutOrStdout())
			fmt.Fprintln(w, "LABEL\tHOSTNAME\tUSER\tPORT")
			for _, h := range hosts {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", h.Label, h.HostName, h.User, h.Port)
			}
			return w.Flush()
		},
	}
}

// newTabWriter returns a tabwriter that produces clean column-aligned output.
func newTabWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
}
