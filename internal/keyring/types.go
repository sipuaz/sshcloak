package keyring

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// SecretKind identifies the type of secret stored in the vault.
type SecretKind string

const (
	// SecretKindPassword stores an SSH password used for sshpass-style auth.
	SecretKindPassword SecretKind = "password"
	// SecretKindPassphrase stores a private-key or vault passphrase.
	SecretKindPassphrase SecretKind = "passphrase"
	// SecretKindPrivateKey stores a complete private key material blob.
	SecretKindPrivateKey SecretKind = "private_key"
)

const currentVaultVersion = 1

// SecretRecord is one typed secret entry stored in the encrypted vault.
type SecretRecord struct {
	Label     string            `yaml:"label"`
	Kind      SecretKind        `yaml:"kind"`
	Value     string            `yaml:"value"`
	Metadata  map[string]string `yaml:"metadata,omitempty"`
	CreatedAt time.Time         `yaml:"created_at"`
	UpdatedAt time.Time         `yaml:"updated_at"`
}

// VaultDocument is the logical YAML document persisted inside the encrypted file.
type VaultDocument struct {
	Version int                     `yaml:"version"`
	Secrets map[string]SecretRecord `yaml:"secrets"`
}

// newVaultDocument returns an empty, initialized vault document.
func newVaultDocument() VaultDocument {
	return VaultDocument{Version: currentVaultVersion, Secrets: make(map[string]SecretRecord)}
}

// secretKey builds the deterministic map key for one label/kind pair.
func secretKey(label string, kind SecretKind) string {
	return strings.TrimSpace(label) + "\x00" + string(kind)
}

// cloneSecretRecord returns a deep copy of one secret record.
func cloneSecretRecord(record SecretRecord) SecretRecord {
	clone := record
	if len(record.Metadata) > 0 {
		clone.Metadata = cloneMetadata(record.Metadata)
	}
	return clone
}

// cloneMetadata returns a deep copy of a metadata map.
func cloneMetadata(metadata map[string]string) map[string]string {
	clone := make(map[string]string, len(metadata))
	for key, value := range metadata {
		clone[key] = value
	}
	return clone
}

// normalizeSecretRecord validates the record and trims user-provided fields.
func normalizeSecretRecord(record SecretRecord) (SecretRecord, error) {
	label := strings.TrimSpace(record.Label)
	if label == "" {
		return SecretRecord{}, fmt.Errorf("secret label cannot be empty")
	}

	kind := SecretKind(strings.TrimSpace(string(record.Kind)))
	if !kind.isSupported() {
		return SecretRecord{}, fmt.Errorf("%w: %s", ErrUnsupportedSecretKind, kind)
	}

	value := record.Value
	if value == "" {
		return SecretRecord{}, fmt.Errorf("secret value cannot be empty")
	}

	normalized := SecretRecord{
		Label: label,
		Kind:  kind,
		Value: value,
	}

	if len(record.Metadata) > 0 {
		normalized.Metadata = make(map[string]string, len(record.Metadata))
		for key, value := range record.Metadata {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == "" {
				continue
			}
			normalized.Metadata[trimmedKey] = strings.TrimSpace(value)
		}
		if len(normalized.Metadata) == 0 {
			normalized.Metadata = nil
		}
	}

	return normalized, nil
}

// isSupported reports whether a secret kind is one of the documented values.
func (k SecretKind) isSupported() bool {
	switch k {
	case SecretKindPassword, SecretKindPassphrase, SecretKindPrivateKey:
		return true
	default:
		return false
	}
}

// sortedSecretRecords returns a deterministic copy of the document secrets.
func sortedSecretRecords(document VaultDocument) []SecretRecord {
	records := make([]SecretRecord, 0, len(document.Secrets))
	for _, record := range document.Secrets {
		records = append(records, cloneSecretRecord(record))
	}
	sort.Slice(records, func(left, right int) bool {
		if records[left].Label != records[right].Label {
			return records[left].Label < records[right].Label
		}
		return records[left].Kind < records[right].Kind
	})
	return records
}
