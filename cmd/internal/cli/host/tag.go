package host

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newTagCmd returns the parent command for host tag operations.
func newTagCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tag",
		Short: "Manage host tags stored in the metadata sidecar",
		Long:  "Add, remove, and list arbitrary tags attached to SSH host labels via ~/.ssh/sshcloak/meta.yaml.",
	}

	cmd.AddCommand(
		newTagAddCmd(),
		newTagRemoveCmd(),
		newTagListCmd(),
	)

	return cmd
}

// newTagAddCmd returns the "sshcloak host tag add" command.
func newTagAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <label> <tag> [<tag>...]",
		Short: "Add one or more tags to a host",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]
			exists, err := getManager().HasHost(label)
			if err != nil {
				return fmt.Errorf("add tag: %w", err)
			}
			if !exists {
				return fmt.Errorf("add tag: host %q not found in SSH config", label)
			}
			if err := getMetadata().AddTags(label, args[1:]...); err != nil {
				return fmt.Errorf("add tag: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "tags added to %q\n", label)
			return nil
		},
	}
}

// newTagRemoveCmd returns the "sshcloak host tag remove" command.
func newTagRemoveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "remove <label> <tag> [<tag>...]",
		Short: "Remove one or more tags from a host",
		Args:  cobra.MinimumNArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]
			if err := getMetadata().RemoveTags(label, args[1:]...); err != nil {
				return fmt.Errorf("remove tag: %w", err)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "tags removed from %q\n", label)
			return nil
		},
	}
}

// newTagListCmd returns the "sshcloak host tag list" command.
func newTagListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list <label>",
		Short: "List all tags attached to a host",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]
			tags, err := getMetadata().Tags(label)
			if err != nil {
				return fmt.Errorf("list tags: %w", err)
			}
			if len(tags) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "no tags")
				return nil
			}
			for _, tag := range tags {
				fmt.Fprintln(cmd.OutOrStdout(), tag)
			}
			return nil
		},
	}
}
