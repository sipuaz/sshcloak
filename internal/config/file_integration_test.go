//go:build integration

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

// testFilePath returns a per-test temp file path so each test is isolated and
// self-contained with no dependency on a pre-existing resource file.
func testFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "ssh_mock_config")
}

func TestFileHandler_Read(t *testing.T) {
	path := testFilePath(t)
	want := []byte("Host test\n  HostName test.example.com\n")
	if err := os.WriteFile(path, want, 0600); err != nil {
		t.Fatalf("setup write: %v", err)
	}

	handler := config.NewFileHandler()
	data, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if string(data) != string(want) {
		t.Fatalf("Read() = %q, want %q", data, want)
	}
}

func TestFileHandler_Write(t *testing.T) {
	path := testFilePath(t)
	testData := []byte("Test data for write")

	handler := config.NewFileHandler()
	if err := handler.Write(path, testData, 0600); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	data, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read() after Write() error: %v", err)
	}
	if string(data) != string(testData) {
		t.Fatalf("Write() stored %q, want %q", data, testData)
	}
}

func TestFileHandler_Append(t *testing.T) {
	path := testFilePath(t)
	initial := []byte("initial\n")
	appended := []byte("appended\n")

	handler := config.NewFileHandler()
	if err := handler.Write(path, initial, 0600); err != nil {
		t.Fatalf("Write() setup error: %v", err)
	}
	if err := handler.Append(path, appended); err != nil {
		t.Fatalf("Append() error: %v", err)
	}

	data, err := handler.Read(path)
	if err != nil {
		t.Fatalf("Read() after Append() error: %v", err)
	}
	want := string(initial) + string(appended)
	if string(data) != want {
		t.Fatalf("Append() result = %q, want %q", data, want)
	}
}

func TestFileHandler_Exists(t *testing.T) {
	path := testFilePath(t)

	handler := config.NewFileHandler()
	if handler.Exists(path) {
		t.Fatal("Exists() returned true for a file that has not been created yet")
	}

	if err := os.WriteFile(path, []byte("x"), 0600); err != nil {
		t.Fatalf("setup write: %v", err)
	}
	if !handler.Exists(path) {
		t.Fatal("Exists() returned false for an existing file")
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
