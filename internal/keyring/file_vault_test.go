package keyring

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// TestSecretKindSupport verifies the supported secret kinds are recognized.
func TestSecretKindSupport(t *testing.T) {
	for _, kind := range []SecretKind{SecretKindPassword, SecretKindPassphrase, SecretKindPrivateKey} {
		if !kind.isSupported() {
			t.Fatalf("%q should be supported", kind)
		}
	}
	if SecretKind("unknown").isSupported() {
		t.Fatal("unknown kind should not be supported")
	}
}

// TestNormalizeSecretRecord trims fields, validates kinds, and copies metadata.
func TestNormalizeSecretRecord(t *testing.T) {
	record, err := normalizeSecretRecord(SecretRecord{
		Label: " prod ",
		Kind:  SecretKindPassword,
		Value: "secret",
		Metadata: map[string]string{
			" host ": " example.com ",
		},
	})
	if err != nil {
		t.Fatalf("normalizeSecretRecord() error: %v", err)
	}
	if record.Label != "prod" {
		t.Fatalf("Label = %q, want prod", record.Label)
	}
	if record.Metadata["host"] != "example.com" {
		t.Fatalf("Metadata = %v, want trimmed values", record.Metadata)
	}

	_, err = normalizeSecretRecord(SecretRecord{Label: "prod", Kind: SecretKind("bad"), Value: "x"})
	if !errors.Is(err, ErrUnsupportedSecretKind) {
		t.Fatalf("expected ErrUnsupportedSecretKind, got %v", err)
	}
}

// TestSecretRecordClone verifies records and metadata are deep-copied.
func TestSecretRecordClone(t *testing.T) {
	record := SecretRecord{Label: "prod", Kind: SecretKindPassword, Value: "secret", Metadata: map[string]string{"a": "b"}}
	clone := cloneSecretRecord(record)
	clone.Metadata["a"] = "c"
	if record.Metadata["a"] != "b" {
		t.Fatalf("metadata should be deep copied, got %v", record.Metadata)
	}
}

// TestVaultDocumentMarshalRoundTrip verifies YAML marshal/unmarshal stays stable.
func TestVaultDocumentMarshalRoundTrip(t *testing.T) {
	doc := VaultDocument{
		Version: currentVaultVersion,
		Secrets: map[string]SecretRecord{
			secretKey("prod", SecretKindPassword): {
				Label:     "prod",
				Kind:      SecretKindPassword,
				Value:     "secret",
				CreatedAt: time.Unix(10, 0).UTC(),
				UpdatedAt: time.Unix(20, 0).UTC(),
			},
		},
	}

	data, err := marshalVaultDocument(doc)
	if err != nil {
		t.Fatalf("marshalVaultDocument() error: %v", err)
	}

	decoded, err := unmarshalVaultDocument(data)
	if err != nil {
		t.Fatalf("unmarshalVaultDocument() error: %v", err)
	}
	if decoded.Version != currentVaultVersion {
		t.Fatalf("Version = %d, want %d", decoded.Version, currentVaultVersion)
	}
	if len(decoded.Secrets) != 1 {
		t.Fatalf("Secrets count = %d, want 1", len(decoded.Secrets))
	}
}

// TestEncryptDecryptVaultDocumentRoundTrip verifies passphrase encryption round-trips the YAML document.
func TestEncryptDecryptVaultDocumentRoundTrip(t *testing.T) {
	doc := newVaultDocument()
	doc.Secrets[secretKey("prod", SecretKindPassword)] = SecretRecord{
		Label: "prod",
		Kind:  SecretKindPassword,
		Value: "secret",
	}

	ciphertext, err := encryptVaultDocument(doc, "passphrase")
	if err != nil {
		t.Fatalf("encryptVaultDocument() error: %v", err)
	}
	if !strings.Contains(string(ciphertext), "BEGIN AGE ENCRYPTED FILE") {
		t.Fatalf("ciphertext should be armored age output, got %q", string(ciphertext))
	}

	decoded, err := decryptVaultDocument(ciphertext, "passphrase")
	if err != nil {
		t.Fatalf("decryptVaultDocument() error: %v", err)
	}
	if len(decoded.Secrets) != 1 {
		t.Fatalf("Secrets count = %d, want 1", len(decoded.Secrets))
	}
	if got := decoded.Secrets[secretKey("prod", SecretKindPassword)].Value; got != "secret" {
		t.Fatalf("decoded secret = %q, want %q", got, "secret")
	}
}
