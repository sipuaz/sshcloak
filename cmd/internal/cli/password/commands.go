package password

import (
	"fmt"
	"os"

	"github.com/sipuaz/sshcloak/internal/keyring"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

// newSetCmd returns "sshcloak password set <label>".
// The password value and vault passphrase are both read from the terminal.
func newSetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "set <label>",
		Short: "Store an SSH password for a host label",
		Long: `Prompt for the vault passphrase, unlock the vault, prompt for the SSH
password for the given host label, store it, and lock the vault.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			vaultPass, err := readPassphrase("Vault passphrase: ")
			if err != nil {
				return err
			}

			store := getStore()
			if err := store.Unlock(vaultPass); err != nil {
				return fmt.Errorf("password set: %w", err)
			}
			defer store.Lock()

			sshPass, err := readPassphrase(fmt.Sprintf("SSH password for %q: ", label))
			if err != nil {
				return err
			}

			if err := store.Put(keyring.SecretRecord{
				Label: label,
				Kind:  keyring.SecretKindPassword,
				Value: sshPass,
			}); err != nil {
				return fmt.Errorf("password set: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "password stored for %q\n", label)
			return nil
		},
	}
}

// newGetCmd returns "sshcloak password get <label>".
// It unlocks the vault, prints the stored password to stdout, and locks.
// Intended for scripting; use with care as the password appears in stdout.
func newGetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "get <label>",
		Short: "Print the stored SSH password for a host label",
		Long: `Unlock the vault and print the stored password to stdout.

This command is intended for scripting.  The password is written to stdout
with a trailing newline so it can be captured in a sub-shell.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			vaultPass, err := readPassphrase("Vault passphrase: ")
			if err != nil {
				return err
			}

			store := getStore()
			if err := store.Unlock(vaultPass); err != nil {
				return fmt.Errorf("password get: %w", err)
			}
			defer store.Lock()

			record, err := store.Get(label, keyring.SecretKindPassword)
			if err != nil {
				return fmt.Errorf("password get: %w", err)
			}

			fmt.Fprintln(cmd.OutOrStdout(), record.Value)
			return nil
		},
	}
}

// newDeleteCmd returns "sshcloak password delete <label>".
func newDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "delete <label>",
		Aliases: []string{"rm", "remove"},
		Short:   "Delete the stored SSH password for a host label",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			label := args[0]

			vaultPass, err := readPassphrase("Vault passphrase: ")
			if err != nil {
				return err
			}

			store := getStore()
			if err := store.Unlock(vaultPass); err != nil {
				return fmt.Errorf("password delete: %w", err)
			}
			defer store.Lock()

			if err := store.Delete(label, keyring.SecretKindPassword); err != nil {
				return fmt.Errorf("password delete: %w", err)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "password deleted for %q\n", label)
			return nil
		},
	}
}

// newListCmd returns "sshcloak password list".
func newListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all labels that have a stored password",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			vaultPass, err := readPassphrase("Vault passphrase: ")
			if err != nil {
				return err
			}

			store := getStore()
			if err := store.Unlock(vaultPass); err != nil {
				return fmt.Errorf("password list: %w", err)
			}
			defer store.Lock()

			records, err := store.List()
			if err != nil {
				return fmt.Errorf("password list: %w", err)
			}

			out := cmd.OutOrStdout()
			found := false
			for _, r := range records {
				if r.Kind == keyring.SecretKindPassword {
					fmt.Fprintln(out, r.Label)
					found = true
				}
			}
			if !found {
				fmt.Fprintln(out, "no passwords stored")
			}
			return nil
		},
	}
}

// readPassphrase prompts on stderr and reads a passphrase without echo.
func readPassphrase(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	raw, err := term.ReadPassword(int(os.Stdin.Fd()))
	fmt.Fprintln(os.Stderr)
	if err != nil {
		return "", fmt.Errorf("reading passphrase: %w", err)
	}
	return string(raw), nil
}
