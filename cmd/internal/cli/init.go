// Package cli wires together the Cobra command tree and the shared application
// context that carries the config.Manager instance across sub-commands.
package cli

import (
	"fmt"
	"os"

	"github.com/sipuaz/sshcloak/cmd/internal/cli/host"
	"github.com/sipuaz/sshcloak/cmd/internal/cli/vault"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newInitCmd returns the top-level "sshcloak init" command.
// It bootstraps both the SSH include file and the encrypted vault in one step.
func newInitCmd() *cobra.Command {
	var appendMode bool

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialise sshcloak for the current user",
		Long: `Create the sshcloak include line in ~/.ssh/config and initialise the
age-encrypted vault used to store SSH passwords.

By default the Include directive is placed at the top of ~/.ssh/config so
sshcloak host entries take precedence (first-match-wins).  Use --append to
place it at the bottom instead.

This is the recommended first command after installation.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			passphrase, err := readPassphraseConfirm("Enter sshcloak vault passphrase: ", "Confirm passphrase: ")
			if err != nil {
				return err
			}

			manager := host.GetManager()
			if err := manager.EnsureInclude(!appendMode); err != nil {
				return fmt.Errorf("init: %w", err)
			}

			store := vault.GetStore()
			if err := store.Initialize(passphrase); err != nil {
				return fmt.Errorf("init: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "sshcloak initialised")
			return nil
		},
	}

	cmd.Flags().BoolVar(&appendMode, "append", false,
		"place the Include directive at the bottom of ~/.ssh/config instead of the top")

	return cmd
}

// readPassphrase prompts the user for a single passphrase without echo.
func readPassphrase(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(raw), nil
}

// readPassphraseConfirm prompts twice and returns an error if the values differ.
func readPassphraseConfirm(prompt, confirmPrompt string) (string, error) {
	first, err := readPassphrase(prompt)
	if err != nil {
		return "", err
	}
	second, err := readPassphrase(confirmPrompt)
	if err != nil {
		return "", err
	}
	if first != second {
		return "", fmt.Errorf("passphrases do not match")
	}
	return first, nil
}
