package completion

import (
	"fmt"

	"github.com/spf13/cobra"
	pflag "github.com/spf13/pflag"
)

// newFlagsCmd returns the "completion flags" command.
func newFlagsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "flags <command>",
		Short: "List available flags for a sshcloak subcommand",
		Long: `List all available flags for a given sshcloak subcommand.

Output format: newline-delimited flags (one flag per line), e.g.:
  --create
  --update
  --delete
  --list

This command is called by shell completion functions; not intended for direct user interaction.`,
		Args: cobra.ExactArgs(1),
		RunE: runFlags,
	}
}

// runFlags lists all available flags for a given command.
func runFlags(cmd *cobra.Command, args []string) error {
	commandName := args[0]

	// Get the root command
	root := cmd.Root()

	// Find the target command by iterating through subcommands
	var targetCmd *cobra.Command
	if commandName == "sshcloak" {
		targetCmd = root
	} else {
		// Search for the command in root's immediate children
		for _, c := range root.Commands() {
			if c.Name() == commandName {
				targetCmd = c
				break
			}
		}
		// If not found, return silently
		if targetCmd == nil {
			return nil
		}
	}

	// Collect all flags local to this command (not inherited persistent flags)
	seenFlags := make(map[string]bool)
	output := cmd.OutOrStdout()

	// Local flags (specific to this command)
	targetCmd.Flags().VisitAll(func(f *pflag.Flag) {
		if !seenFlags[f.Name] {
			fmt.Fprintf(output, "--%s\n", f.Name)
			seenFlags[f.Name] = true
		}
	})

	return nil
}
