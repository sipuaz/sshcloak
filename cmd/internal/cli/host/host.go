// Package host implements the "sshcloak host" command group and the manager
// injection mechanism used by each sub-command.
package host

import (
	"github.com/sipuaz/sshcloak/internal/config"
	"github.com/sipuaz/sshcloak/internal/metadata"
	"github.com/spf13/cobra"
)

// SetManager stores the manager on the root command's annotation map so it
// survives across the PersistentPreRunE / RunE boundary.
// cobra.Command carries an Annotations map[string]string which is not suitable
// for non-string values, so we use the command's parent chain and a package-
// level pointer instead (safe because commands run sequentially in a CLI).
var sharedManager *config.Manager

// sharedMetadata is injected by the root command and backs host tagging.
var sharedMetadata *metadata.Store

// SetManager is called by root.PersistentPreRunE to inject the manager before
// any sub-command runs.
func SetManager(_ *cobra.Command, mgr *config.Manager) {
	sharedManager = mgr
}

// getManager retrieves the injected manager, panicking if SetManager was never
// called (which would indicate a wiring bug, not a user error).
func getManager() *config.Manager {
	if sharedManager == nil {
		panic("sshcloak: manager not initialised — this is a bug")
	}
	return sharedManager
}

// GetManager returns the injected manager for sibling command packages that
// need to share the same application context.
func GetManager() *config.Manager {
	return getManager()
}

// SetMetadataStore injects the sidecar metadata store before sub-commands run.
func SetMetadataStore(_ *cobra.Command, store *metadata.Store) {
	sharedMetadata = store
}

// getMetadata retrieves the injected metadata store, panicking on wiring bugs.
func getMetadata() *metadata.Store {
	if sharedMetadata == nil {
		panic("sshcloak: metadata store not initialised — this is a bug")
	}
	return sharedMetadata
}

// NewHostCmd returns the parent "host" command.  It has no Run of its own;
// running "sshcloak host" without a sub-command prints usage automatically.
func NewHostCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "host",
		Short: "Manage SSH host entries",
		Long:  "Add, list, inspect, edit, and remove SSH host entries in the sshcloak-managed config.",
	}

	cmd.AddCommand(
		newAddCmd(),
		newListCmd(),
		newGetCmd(),
		newTagCmd(),
		newEditCmd(),
		newRemoveCmd(),
	)

	return cmd
}
