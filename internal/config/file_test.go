package config_test

import (
	"os"
	"testing"
)

// fileStub implements the File interface using an in-memory map,
// allowing unit tests to exercise code that depends on file I/O
// without touching the real filesystem.
type fileStub struct {
	storage map[string][]byte
}

func newFileStub() *fileStub {
	return &fileStub{storage: make(map[string][]byte)}
}

func (s *fileStub) Read(path string) ([]byte, error) {
	data, ok := s.storage[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (s *fileStub) Write(path string, data []byte, perm os.FileMode) error {
	s.storage[path] = data
	return nil
}

func (s *fileStub) Append(path string, data []byte) error {
	current, _ := s.Read(path)
	s.storage[path] = append(current, data...)
	return nil
}

func (s *fileStub) Exists(path string) bool {
	_, ok := s.storage[path]
	return ok
}

func TestFileStubRead(t *testing.T) {
	stub := newFileStub()
	stub.storage["config"] = []byte("test data")

	data, err := stub.Read("config")
	if err != nil {
		t.Fatalf("Read() error: %v", err)
	}
	if string(data) != "test data" {
		t.Errorf("Read() = %q, want %q", string(data), "test data")
	}

	_, err = stub.Read("nonexistent")
	if err == nil {
		t.Error("Read() on missing key should return error")
	}
}

func TestFileStubWrite(t *testing.T) {
	stub := newFileStub()
	if err := stub.Write("config", []byte("hello"), 0644); err != nil {
		t.Fatalf("Write() error: %v", err)
	}
	if string(stub.storage["config"]) != "hello" {
		t.Errorf("Write() stored %q, want %q", string(stub.storage["config"]), "hello")
	}
}

func TestFileStubAppend(t *testing.T) {
	stub := newFileStub()
	stub.storage["config"] = []byte("hello")
	if err := stub.Append("config", []byte(" world")); err != nil {
		t.Fatalf("Append() error: %v", err)
	}
	if got := string(stub.storage["config"]); got != "hello world" {
		t.Errorf("Append() = %q, want %q", got, "hello world")
	}
}

func TestFileStubExists(t *testing.T) {
	stub := newFileStub()
	stub.storage["config"] = []byte("data")

	if !stub.Exists("config") {
		t.Error("Exists() should return true for present key")
	}
	if stub.Exists("missing") {
		t.Error("Exists() should return false for absent key")
	}
}
