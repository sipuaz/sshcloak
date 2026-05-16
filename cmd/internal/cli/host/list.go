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
	var tag string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all managed SSH hosts",
		Long:  "Print a tab-aligned table of every Host block in the sshcloak-managed config, optionally filtered by tag.",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			hosts, err := getManager().ListHosts()
			if err != nil {
				return fmt.Errorf("list hosts: %w", err)
			}

			filter := map[string]struct{}(nil)
			if tag != "" {
				labels, err := getMetadata().TaggedHosts(tag)
				if err != nil {
					return fmt.Errorf("list hosts: %w", err)
				}
				filter = make(map[string]struct{}, len(labels))
				for _, label := range labels {
					filter[label] = struct{}{}
				}
			}

			type row struct {
				label    string
				hostname string
				user     string
				port     string
				tags     string
			}
			rows := make([]row, 0, len(hosts))
			for _, h := range hosts {
				if filter != nil {
					if _, ok := filter[h.Label]; !ok {
						continue
					}
				}
				tags, err := getMetadata().Tags(h.Label)
				if err != nil {
					return fmt.Errorf("list hosts: %w", err)
				}
				rows = append(rows, row{
					label:    h.Label,
					hostname: h.HostName,
					user:     h.User,
					port:     h.Port,
					tags:     joinTags(tags),
				})
			}

			if len(rows) == 0 {
				if tag == "" {
					fmt.Fprintln(cmd.OutOrStdout(), "no managed hosts found")
				} else {
					fmt.Fprintln(cmd.OutOrStdout(), "no matching managed hosts found")
				}
				warnOrphanedMetadata(cmd)
				return nil
			}

			w := newTabWriter(cmd.OutOrStdout())
			fmt.Fprintln(w, "LABEL\tHOSTNAME\tUSER\tPORT\tTAGS")
			for _, h := range rows {
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", h.label, h.hostname, h.user, h.port, h.tags)
			}
			if err := w.Flush(); err != nil {
				return err
			}
			warnOrphanedMetadata(cmd)
			return nil
		},
	}

	cmd.Flags().StringVar(&tag, "tag", "", "filter managed hosts by one tag")
	return cmd
}

// newTabWriter returns a tabwriter that produces clean column-aligned output.
func newTabWriter(out io.Writer) *tabwriter.Writer {
	return tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
}
