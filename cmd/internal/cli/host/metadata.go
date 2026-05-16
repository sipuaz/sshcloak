package host

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// joinTags renders a sorted tag list for table and key/value output.
func joinTags(values []string) string {
	if len(values) == 0 {
		return "-"
	}
	return strings.Join(values, ",")
}

// warnOrphanedMetadata prints a non-fatal warning for metadata labels that do
// not exist anywhere in the SSH config tree.
func warnOrphanedMetadata(cmd *cobra.Command) {
	orphans, err := getMetadata().Orphans(getManager().HasHost)
	if err != nil || len(orphans) == 0 {
		return
	}
	fmt.Fprintf(cmd.ErrOrStderr(), "warning: orphaned metadata for hosts: %s\n", strings.Join(orphans, ", "))
}
