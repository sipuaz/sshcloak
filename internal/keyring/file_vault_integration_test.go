//go:build integration

package keyring

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFileVaultStoreCRUD verifies the file-backed vault can store, retrieve,
// list, and delete both password and private-key secrets.
func TestFileVaultStoreCRUD(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.age")
	store := NewFileVaultStore(path)

	if err := store.Unlock("correct horse battery staple"); err != nil {
		t.Fatalf("Unlock() error: %v", err)
	}
	if err := store.Put(SecretRecord{Label: "prod", Kind: SecretKindPassword, Value: "s3cr3t"}); err != nil {
		t.Fatalf("Put(password) error: %v", err)
	}
	if err := store.Put(SecretRecord{Label: "prod", Kind: SecretKindPrivateKey, Value: "-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----"}); err != nil {
		t.Fatalf("Put(private key) error: %v", err)
	}

	password, err := store.Get("prod", SecretKindPassword)
	if err != nil {
		t.Fatalf("Get(password) error: %v", err)
	}
	if password.Value != "s3cr3t" {
		t.Fatalf("password value = %q, want %q", password.Value, "s3cr3t")
	}

	secrets, err := store.List()
	if err != nil {
		t.Fatalf("List() error: %v", err)
	}
	if len(secrets) != 2 {
		t.Fatalf("List() count = %d, want 2", len(secrets))
	}

	store.Lock()
	if err := store.Unlock("correct horse battery staple"); err != nil {
		t.Fatalf("Unlock() after lock error: %v", err)
	}
	if err := store.Delete("prod", SecretKindPassword); err != nil {
		t.Fatalf("Delete(password) error: %v", err)
	}
	secrets, err = store.List()
	if err != nil {
		t.Fatalf("List() after delete error: %v", err)
	}
	if len(secrets) != 1 || secrets[0].Kind != SecretKindPrivateKey {
		t.Fatalf("List() after delete = %+v, want one private key", secrets)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("vault file should not be empty after writes")
	}
}

// TestFileVaultStoreWrongPassphrase verifies the store refuses to unlock with
// an incorrect passphrase.
func TestFileVaultStoreWrongPassphrase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.age")
	store := NewFileVaultStore(path)
	if err := store.Unlock("one"); err != nil {
		t.Fatalf("Unlock() error: %v", err)
	}
	if err := store.Put(SecretRecord{Label: "prod", Kind: SecretKindPassword, Value: "secret"}); err != nil {
		t.Fatalf("Put() error: %v", err)
	}

	other := NewFileVaultStore(path)
	if err := other.Unlock("two"); err == nil {
		t.Fatal("Unlock() with wrong passphrase should fail")
	}
}

// TestFileVaultStoreRotatePassphrase verifies the vault can be re-encrypted
// with a new passphrase and that the old passphrase no longer works.
func TestFileVaultStoreRotatePassphrase(t *testing.T) {
	path := filepath.Join(t.TempDir(), "vault.age")
	store := NewFileVaultStore(path)
	if err := store.Unlock("old-pass"); err != nil {
		t.Fatalf("Unlock() error: %v", err)
	}
	if err := store.Put(SecretRecord{Label: "prod", Kind: SecretKindPassword, Value: "secret"}); err != nil {
		t.Fatalf("Put() error: %v", err)
	}
	if err := store.RotatePassphrase("new-pass"); err != nil {
		t.Fatalf("RotatePassphrase() error: %v", err)
	}

	oldStore := NewFileVaultStore(path)
	if err := oldStore.Unlock("old-pass"); err == nil {
		t.Fatal("old passphrase should no longer unlock the vault")
	}

	newStore := NewFileVaultStore(path)
	if err := newStore.Unlock("new-pass"); err != nil {
		t.Fatalf("Unlock(new-pass) error: %v", err)
	}
	record, err := newStore.Get("prod", SecretKindPassword)
	if err != nil {
		t.Fatalf("Get() after rotate error: %v", err)
	}
	if record.Value != "secret" {
		t.Fatalf("Get() after rotate = %q, want secret", record.Value)
	}
}
