package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestAgentSetGetStatusAndLock(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "session.sock")
	go func() {
		_ = RunAgent(socketPath, time.Minute)
	}()
	time.Sleep(200 * time.Millisecond)

	client := NewClient(socketPath, time.Minute)
	if err := client.EnsureAgent(); err != nil {
		t.Fatalf("EnsureAgent() error: %v", err)
	}

	if err := client.Set("shell-1", "vault-pass"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	got, found, err := client.Get("shell-1")
	if err != nil {
		t.Fatalf("Get() error: %v", err)
	}
	if !found {
		t.Fatal("Get() found = false, want true")
	}
	if got != "vault-pass" {
		t.Fatalf("Get() value = %q, want %q", got, "vault-pass")
	}

	active, err := client.Status("shell-1")
	if err != nil {
		t.Fatalf("Status() error: %v", err)
	}
	if !active {
		t.Fatal("Status() = false, want true")
	}

	if err := client.Lock("shell-1"); err != nil {
		t.Fatalf("Lock() error: %v", err)
	}
	_, found, err = client.Get("shell-1")
	if err != nil {
		t.Fatalf("Get() after lock error: %v", err)
	}
	if found {
		t.Fatal("Get() after lock found = true, want false")
	}
}

func TestAgentTTlExpiry(t *testing.T) {
	socketPath := filepath.Join(t.TempDir(), "session.sock")
	go func() {
		_ = RunAgent(socketPath, 150*time.Millisecond)
	}()
	time.Sleep(200 * time.Millisecond)

	client := NewClient(socketPath, time.Minute)
	if err := client.EnsureAgent(); err != nil {
		t.Fatalf("EnsureAgent() error: %v", err)
	}
	if err := client.Set("shell-2", "vault-pass"); err != nil {
		t.Fatalf("Set() error: %v", err)
	}
	time.Sleep(250 * time.Millisecond)

	_, found, err := client.Get("shell-2")
	if err != nil {
		t.Fatalf("Get() after ttl error: %v", err)
	}
	if found {
		t.Fatal("Get() after ttl found = true, want false")
	}
}

func TestDefaultSocketPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	path, err := DefaultSocketPath("/tmp/vault.age")
	if err != nil {
		t.Fatalf("DefaultSocketPath() error: %v", err)
	}
	if filepath.Dir(path) != filepath.Join(home, ".ssh", "sshcloak", "run") {
		t.Fatalf("socket dir = %q", filepath.Dir(path))
	}
	if filepath.Ext(path) != ".sock" {
		t.Fatalf("socket path = %q, want .sock extension", path)
	}
	if _, err := os.Stat(filepath.Dir(path)); err == nil {
		t.Fatal("socket directory should not be created by DefaultSocketPath")
	}
}
