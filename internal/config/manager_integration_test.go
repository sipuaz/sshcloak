//go:build integration

package config_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

// TestManagerIntegrationCRUD verifies CRUD operations against the real filesystem.
func TestManagerIntegrationCRUD(t *testing.T) {
	tempDir := os.Getenv("SSHCLOAK_TEST_DIR")
	if tempDir == "" {
		tempDir = t.TempDir()
	}
	t.Logf("test directory: %s", tempDir)
	userConfigPath := filepath.Join(tempDir, ".ssh", "config")
	managedConfigPath := filepath.Join(tempDir, ".ssh", "sshcloak", "config")

	manager := config.NewManager(config.NewFileHandler(), userConfigPath, managedConfigPath)
	err := manager.AddHost(config.HostSpec{
		Label:    "prod",
		HostName: "prod.example.com",
		User:     "deploy",
		Port:     "2222",
		Extra: map[string][]string{
			"ProxyJump": {"bastion"},
		},
	})
	if err != nil {
		t.Fatalf("AddHost() error: %v", err)
	}

	userConfigBytes, err := config.NewFileHandler().Read(userConfigPath)
	if err != nil {
		t.Fatalf("Read(user config) error: %v", err)
	}
	if !strings.Contains(string(userConfigBytes), "Include "+managedConfigPath) {
		t.Fatalf("user config missing Include directive: %q", string(userConfigBytes))
	}

	err = manager.UpdateHost("prod", config.HostSpec{
		Label:    "prod",
		HostName: "prod2.example.com",
		User:     "deploy",
		Port:     "2200",
	})
	if err != nil {
		t.Fatalf("UpdateHost() error: %v", err)
	}

	host, err := manager.GetHost("prod")
	if err != nil {
		t.Fatalf("GetHost() error: %v", err)
	}
	if host.HostName != "prod2.example.com" || host.Port != "2200" {
		t.Fatalf("updated host = %+v", host)
	}

	err = manager.DeleteHost("prod")
	if err != nil {
		t.Fatalf("DeleteHost() error: %v", err)
	}

	hosts, err := manager.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts() error: %v", err)
	}
	if len(hosts) != 0 {
		t.Fatalf("ListHosts() = %+v, want empty", hosts)
	}

	managedConfigBytes, err := config.NewFileHandler().Read(managedConfigPath)
	if err != nil {
		t.Fatalf("Read(managed config) error: %v", err)
	}
	if strings.TrimSpace(string(managedConfigBytes)) != "" {
		t.Fatalf("managed config = %q, want empty", string(managedConfigBytes))
	}
}
