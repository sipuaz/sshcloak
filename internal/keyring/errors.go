package keyring

import "errors"

var (
	// ErrVaultLocked reports that a vault operation requires Unlock first.
	ErrVaultLocked = errors.New("vault is locked")
	// ErrSecretNotFound reports that no secret exists for the requested label and kind.
	ErrSecretNotFound = errors.New("secret not found")
	// ErrUnsupportedSecretKind reports that the requested secret kind is unknown.
	ErrUnsupportedSecretKind = errors.New("unsupported secret kind")
	// ErrUnsupportedVaultVersion reports that an on-disk vault uses a newer format.
	ErrUnsupportedVaultVersion = errors.New("unsupported vault version")
)
