package completion

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

// newListHostsCmd returns the "completion list-hosts" command.
func newListHostsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list-hosts",
		Short: "List all managed host aliases as JSON",
		Long: `List all managed host aliases from ~/.ssh/sshcloak/config.

Output format: one JSON object per line (JSONL), each with:
  - "host": the host alias
  - "tags": array of host tags (from metadata)

Example output:
  {"host":"prod-server","tags":["production"]}
  {"host":"dev-laptop","tags":["development"]}

This command is called by shell completion functions; not intended for direct user interaction.`,
		RunE: runListHosts,
	}
}

// listHostOutput represents one host entry in the completion output.
type listHostOutput struct {
	Host string   `json:"host"`
	Tags []string `json:"tags"`
}

// runListHosts lists all managed hosts as JSON.
func runListHosts(cmd *cobra.Command, args []string) error {
	mgr := getManager()

	hosts, err := mgr.ListHosts()
	if err != nil {
		// If config doesn't exist or can't be read, return empty list silently.
		// This prevents completion from failing when config is missing.
		return nil
	}

	// For now, tags are empty; this is a placeholder for future metadata integration.
	for _, host := range hosts {
		out := listHostOutput{
			Host: host.Label,
			Tags: []string{},
		}
		if err := json.NewEncoder(cmd.OutOrStdout()).Encode(out); err != nil {
			return fmt.Errorf("encode host: %w", err)
		}
	}

	return nil
}
