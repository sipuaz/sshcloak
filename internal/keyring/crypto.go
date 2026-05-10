package keyring

import (
	"bytes"
	"io"

	"filippo.io/age"
	"filippo.io/age/armor"
	"go.yaml.in/yaml/v3"
)

// marshalVaultDocument serializes the logical vault document to YAML.
func marshalVaultDocument(document VaultDocument) ([]byte, error) {
	if document.Version == 0 {
		document.Version = currentVaultVersion
	}
	if document.Secrets == nil {
		document.Secrets = make(map[string]SecretRecord)
	}
	return yaml.Marshal(document)
}

// unmarshalVaultDocument parses a YAML vault document and normalizes defaults.
func unmarshalVaultDocument(data []byte) (VaultDocument, error) {
	if len(bytes.TrimSpace(data)) == 0 {
		return newVaultDocument(), nil
	}

	var document VaultDocument
	if err := yaml.Unmarshal(data, &document); err != nil {
		return VaultDocument{}, err
	}
	if document.Version == 0 {
		document.Version = currentVaultVersion
	}
	if document.Version > currentVaultVersion {
		return VaultDocument{}, ErrUnsupportedVaultVersion
	}
	if document.Secrets == nil {
		document.Secrets = make(map[string]SecretRecord)
	}
	return document, nil
}

// encryptVaultDocument encrypts the YAML representation of a vault document.
func encryptVaultDocument(document VaultDocument, passphrase string) ([]byte, error) {
	plaintext, err := marshalVaultDocument(document)
	if err != nil {
		return nil, err
	}

	recipient, err := age.NewScryptRecipient(passphrase)
	if err != nil {
		return nil, err
	}

	var encrypted bytes.Buffer
	armored := armor.NewWriter(&encrypted)
	writer, err := age.Encrypt(armored, recipient)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(plaintext); err != nil {
		writer.Close()
		armored.Close()
		return nil, err
	}
	if err := writer.Close(); err != nil {
		armored.Close()
		return nil, err
	}
	if err := armored.Close(); err != nil {
		return nil, err
	}

	return encrypted.Bytes(), nil
}

// decryptVaultDocument decrypts an age-encoded vault file into the logical document.
func decryptVaultDocument(ciphertext []byte, passphrase string) (VaultDocument, error) {
	if len(bytes.TrimSpace(ciphertext)) == 0 {
		return newVaultDocument(), nil
	}

	identity, err := age.NewScryptIdentity(passphrase)
	if err != nil {
		return VaultDocument{}, err
	}

	reader, err := age.Decrypt(armor.NewReader(bytes.NewReader(ciphertext)), identity)
	if err != nil {
		return VaultDocument{}, err
	}

	plaintext, err := io.ReadAll(reader)
	if err != nil {
		return VaultDocument{}, err
	}

	return unmarshalVaultDocument(plaintext)
}
