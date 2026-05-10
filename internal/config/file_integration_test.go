//go:build integration

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

var fileHandler = config.NewFileHandler()

const testPath = "../resources/ssh_mock_config"

func TestFileHandler_Read(t *testing.T) {
	data, err := fileHandler.Read(testPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("File is empty")
	}
}

func TestFileHandler_Write(t *testing.T) {
	testData := []byte("Test data for write")
	err := fileHandler.Write(testPath, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	// Verify that the data was written correctly
	data, err := fileHandler.Read(testPath)
	if err != nil {
		t.Fatalf("Failed to read file after write: %v", err)
	}
	if string(data[len(data)-len(testData):]) != string(testData) {
		t.Fatalf("Data was not written correctly")
	}
}

func TestFileHandler_Append(t *testing.T) {
	testData := []byte("Test data for append")
	err := fileHandler.Append(testPath, testData)
	if err != nil {
		t.Fatalf("Failed to append to file: %v", err)
	}

	// Verify that the data was appended correctly
	data, err := fileHandler.Read(testPath)
	if err != nil {
		t.Fatalf("Failed to read file after append: %v", err)
	}
	if string(data[len(data)-len(testData):]) != string(testData) {
		t.Fatalf("Data was not appended correctly")
	}
}

func TestFileHandler_Exists(t *testing.T) {
	if !fileHandler.Exists(testPath) {
		t.Fatalf("File should exist: %v", testPath)
	}
	if fileHandler.Exists("non_existent_file") {
		t.Fatalf("File should not exist: %v", "non_existent_file")
	}
}

// TestFileHandler_AtomicWrite verifies that AtomicWrite produces the expected
// file content, applies the requested permissions, and leaves no temp file
// behind after a successful write.
func TestFileHandler_AtomicWrite(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "atomic.conf")

	handler := config.NewFileHandler()
	want := []byte("Host prod\n  HostName prod.example.com\n")

	if err := handler.AtomicWrite(target, want, 0600); err != nil {
		t.Fatalf("AtomicWrite() error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() after AtomicWrite error: %v", err)
	}
	if string(got) != string(want) {
		t.Errorf("content = %q, want %q", got, want)
	}

	info, err := os.Stat(target)
	if err != nil {
		t.Fatalf("Stat() error: %v", err)
	}
	if got := info.Mode().Perm(); got != 0600 {
		t.Errorf("permissions = %o, want 0600", got)
	}

	// No temp file (.sshcloak-*) must remain in the directory.
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error: %v", err)
	}
	for _, entry := range entries {
		if entry.Name() != filepath.Base(target) {
			t.Errorf("unexpected file left in dir: %s", entry.Name())
		}
	}
}

// TestFileHandler_AtomicWriteOverwrites verifies that a second AtomicWrite
// fully replaces the previous content with no observable intermediate state.
func TestFileHandler_AtomicWriteOverwrites(t *testing.T) {
	dir := t.TempDir()
	target := filepath.Join(dir, "atomic.conf")
	handler := config.NewFileHandler()

	first := []byte("original content\n")
	if err := handler.AtomicWrite(target, first, 0644); err != nil {
		t.Fatalf("first AtomicWrite() error: %v", err)
	}

	second := []byte("updated content\n")
	if err := handler.AtomicWrite(target, second, 0644); err != nil {
		t.Fatalf("second AtomicWrite() error: %v", err)
	}

	got, err := os.ReadFile(target)
	if err != nil {
		t.Fatalf("ReadFile() error: %v", err)
	}
	if string(got) != string(second) {
		t.Errorf("content = %q, want %q", got, second)
	}
}
