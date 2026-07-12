package vault

import (
	"fmt"

	"github.com/spf13/cobra"
)

// newInitCmd returns "sshcloak vault init".
// It unlocks (or creates) the vault with a passphrase supplied interactively,
// ensuring the vault file exists and is valid before any password commands run.
func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialise the encrypted vault",
		Long: `Create the vault file if it does not exist, or verify an existing one.

The passphrase is read from the terminal (no echo).  Running vault init is
optional: the first "password set" call will also create the vault.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			passphrase, err := readPassphraseConfirm("Enter vault passphrase: ", "Confirm passphrase: ")
			if err != nil {
				return err
			}

			store := getStore()
			if err := store.Initialize(passphrase); err != nil {
				return fmt.Errorf("vault init: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "vault initialised")
			return nil
		},
	}
}

// newRotateCmd returns "sshcloak vault rotate-passphrase".
func newRotateCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "rotate-passphrase",
		Short: "Re-encrypt the vault with a new passphrase",
		Long: `Unlock the vault with the current passphrase and immediately re-encrypt it
with a new one.  Both passphrases are read interactively.`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			store, err := UnlockStore("Current passphrase: ")
			if err != nil {
				return fmt.Errorf("vault rotate-passphrase: %w", err)
			}
			defer store.Lock()

			newPass, err := readPassphraseConfirm("New passphrase: ", "Confirm new passphrase: ")
			if err != nil {
				return err
			}

			if err := store.RotatePassphrase(newPass); err != nil {
				return fmt.Errorf("vault rotate-passphrase: %w", err)
			}
			if err := LockSession(); err != nil {
				return fmt.Errorf("vault rotate-passphrase: lock session cache: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), "vault passphrase rotated")
			return nil
		},
	}
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
