package keyring

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileVaultStore persists a vault document as an encrypted file on disk.
type FileVaultStore struct {
	path       string
	mu         sync.RWMutex
	unlocked   bool
	passphrase []byte
	document   VaultDocument
}

const (
	vaultDirectoryPerm = 0700
	vaultFilePerm      = 0600
)

// NewFileVaultStore creates a vault store bound to one encrypted file path.
func NewFileVaultStore(path string) *FileVaultStore {
	return &FileVaultStore{path: path}
}

// Unlock loads and decrypts the vault using the provided passphrase.
func (s *FileVaultStore) Unlock(passphrase string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	document, err := s.readDocumentLocked(passphrase)
	if err != nil {
		return err
	}

	s.lockBytes()
	s.passphrase = []byte(passphrase)
	s.document = document
	s.unlocked = true
	return nil
}

// Lock discards the in-memory document and clears the cached passphrase.
func (s *FileVaultStore) Lock() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.lockBytes()
	s.document = VaultDocument{}
	s.unlocked = false
}

// Get returns one secret record by label and kind.
func (s *FileVaultStore) Get(label string, kind SecretKind) (SecretRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureUnlockedLocked(); err != nil {
		return SecretRecord{}, err
	}
	record, ok := s.document.Secrets[secretKey(label, kind)]
	if !ok {
		return SecretRecord{}, fmt.Errorf("%w: %s/%s", ErrSecretNotFound, label, kind)
	}
	return cloneSecretRecord(record), nil
}

// Put inserts or updates one secret record and persists the encrypted vault.
func (s *FileVaultStore) Put(record SecretRecord) error {
	normalized, err := normalizeSecretRecord(record)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureUnlockedLocked(); err != nil {
		return err
	}

	key := secretKey(normalized.Label, normalized.Kind)
	now := time.Now().UTC()
	existing, exists := s.document.Secrets[key]
	normalized.CreatedAt = now
	normalized.UpdatedAt = now
	if exists {
		normalized.CreatedAt = existing.CreatedAt
		if normalized.CreatedAt.IsZero() {
			normalized.CreatedAt = now
		}
	}
	s.document.Secrets[key] = cloneSecretRecord(normalized)
	return s.persistLocked()
}

// Delete removes one secret record and persists the encrypted vault.
func (s *FileVaultStore) Delete(label string, kind SecretKind) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureUnlockedLocked(); err != nil {
		return err
	}

	key := secretKey(label, kind)
	if _, ok := s.document.Secrets[key]; !ok {
		return fmt.Errorf("%w: %s/%s", ErrSecretNotFound, label, kind)
	}
	delete(s.document.Secrets, key)
	return s.persistLocked()
}

// List returns every secret record sorted by label and kind.
func (s *FileVaultStore) List() ([]SecretRecord, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if err := s.ensureUnlockedLocked(); err != nil {
		return nil, err
	}
	return sortedSecretRecords(s.document), nil
}

// RotatePassphrase rewrites the vault using a new passphrase.
func (s *FileVaultStore) RotatePassphrase(newPassphrase string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ensureUnlockedLocked(); err != nil {
		return err
	}
	s.lockBytes()
	s.passphrase = []byte(newPassphrase)
	s.unlocked = true
	return s.persistLocked()
}

// readDocumentLocked loads the encrypted file or returns an empty document.
func (s *FileVaultStore) readDocumentLocked(passphrase string) (VaultDocument, error) {
	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return newVaultDocument(), nil
		}
		return VaultDocument{}, err
	}
	if len(bytesTrimSpace(data)) == 0 {
		return newVaultDocument(), nil
	}
	return decryptVaultDocument(data, passphrase)
}

// persistLocked encrypts and atomically writes the vault to disk.
func (s *FileVaultStore) persistLocked() error {
	if err := os.MkdirAll(filepath.Dir(s.path), vaultDirectoryPerm); err != nil {
		return err
	}
	encrypted, err := encryptVaultDocument(s.document, string(s.passphrase))
	if err != nil {
		return err
	}
	return atomicWriteFile(s.path, encrypted, vaultFilePerm)
}

// ensureUnlockedLocked reports ErrVaultLocked when the vault is not open.
func (s *FileVaultStore) ensureUnlockedLocked() error {
	if !s.unlocked {
		return ErrVaultLocked
	}
	if s.document.Secrets == nil {
		s.document = newVaultDocument()
	}
	if s.document.Version == 0 {
		s.document.Version = currentVaultVersion
	}
	return nil
}

// lockBytes zeroes the cached passphrase bytes before the store is relocked.
func (s *FileVaultStore) lockBytes() {
	for i := range s.passphrase {
		s.passphrase[i] = 0
	}
	s.passphrase = nil
}

// bytesTrimSpace avoids importing bytes in the file-vault implementation.
func bytesTrimSpace(data []byte) []byte {
	start := 0
	for start < len(data) {
		if data[start] != ' ' && data[start] != '\n' && data[start] != '\r' && data[start] != '\t' {
			break
		}
		start++
	}
	end := len(data)
	for end > start {
		b := data[end-1]
		if b != ' ' && b != '\n' && b != '\r' && b != '\t' {
			break
		}
		end--
	}
	return data[start:end]
}

// atomicWriteFile writes a file through a same-directory temp file and rename.
func atomicWriteFile(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)
	tmp, err := os.CreateTemp(dir, ".sshcloak-vault-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	removeTemp := true
	defer func() {
		if removeTemp {
			os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	removeTemp = false
	return nil
}
