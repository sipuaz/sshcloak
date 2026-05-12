package config_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

// TestManagerAddHostWritesManagedConfig verifies add, include bootstrapping, and canonical serialization.
func TestManagerAddHostWritesManagedConfig(t *testing.T) {
	files := newFileStub()
	files.storage["user.conf"] = []byte("Host legacy\n  User alice\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	err := manager.AddHost(config.HostSpec{
		Label:    "prod",
		HostName: "prod.example.com",
		User:     "deploy",
		Port:     "2222",
		IdentityFiles: []string{
			"~/.ssh/id_prod",
		},
		Extra: map[string][]string{
			"ProxyJump": {"bastion"},
		},
	})
	if err != nil {
		t.Fatalf("AddHost() error: %v", err)
	}

	userConfig := string(files.storage["user.conf"])
	if !strings.Contains(userConfig, "Include managed.conf") {
		t.Fatalf("user config missing Include directive: %q", userConfig)
	}
	if !strings.HasPrefix(userConfig, "Include managed.conf\n") {
		t.Fatalf("Include directive not at top of user config: %q", userConfig)
	}

	host, err := manager.GetHost("prod")
	if err != nil {
		t.Fatalf("GetHost() error: %v", err)
	}
	if host.HostName != "prod.example.com" {
		t.Fatalf("HostName = %q, want %q", host.HostName, "prod.example.com")
	}
	if host.User != "deploy" {
		t.Fatalf("User = %q, want %q", host.User, "deploy")
	}
	if host.Port != "2222" {
		t.Fatalf("Port = %q, want %q", host.Port, "2222")
	}
	if len(host.IdentityFiles) != 1 || host.IdentityFiles[0] != "~/.ssh/id_prod" {
		t.Fatalf("IdentityFiles = %v, want [~/.ssh/id_prod]", host.IdentityFiles)
	}
	if values := host.Extra["ProxyJump"]; len(values) != 1 || values[0] != "bastion" {
		t.Fatalf("ProxyJump = %v, want [bastion]", values)
	}
}

// TestManagerAddHostRejectsDuplicateAcrossIncludes verifies duplicate detection across included files.
func TestManagerAddHostRejectsDuplicateAcrossIncludes(t *testing.T) {
	userConfigPath := filepath.Join(string(filepath.Separator), "tmp", "user.conf")
	includedConfigPath := filepath.Join(string(filepath.Separator), "tmp", "extras", "team.conf")
	managedConfigPath := filepath.Join(string(filepath.Separator), "tmp", "managed.conf")

	files := newFileStub()
	files.storage[userConfigPath] = []byte("Include extras/team.conf\n")
	files.storage[includedConfigPath] = []byte("Host prod\n  User alice\n")

	manager := config.NewManager(files, userConfigPath, managedConfigPath)
	err := manager.AddHost(config.HostSpec{Label: "prod", HostName: "prod.example.com"})
	if !errors.Is(err, config.ErrHostAlreadyExists) {
		t.Fatalf("AddHost() error = %v, want ErrHostAlreadyExists", err)
	}
	if _, exists := files.storage[managedConfigPath]; exists {
		t.Fatal("managed config should not be written when AddHost fails")
	}
}

// TestManagerUpdateHostReplacesBlock verifies that updates rewrite the targeted block canonically.
func TestManagerUpdateHostReplacesBlock(t *testing.T) {
	files := newFileStub()
	files.storage["managed.conf"] = []byte("Host prod\n  HostName old.example.com\n  User old\n  Port 22\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	err := manager.UpdateHost("prod", config.HostSpec{
		Label:    "prod",
		HostName: "new.example.com",
		User:     "deploy",
		Port:     "2222",
		IdentityFiles: []string{
			"~/.ssh/id_prod",
			"~/.ssh/id_backup",
		},
	})
	if err != nil {
		t.Fatalf("UpdateHost() error: %v", err)
	}

	host, err := manager.GetHost("prod")
	if err != nil {
		t.Fatalf("GetHost() error: %v", err)
	}
	if host.HostName != "new.example.com" || host.User != "deploy" || host.Port != "2222" {
		t.Fatalf("updated host = %+v", host)
	}
	if len(host.IdentityFiles) != 2 {
		t.Fatalf("IdentityFiles count = %d, want 2", len(host.IdentityFiles))
	}
}

// TestManagerDeleteHostRemovesManagedBlock verifies delete removes the requested Host block only.
func TestManagerDeleteHostRemovesManagedBlock(t *testing.T) {
	files := newFileStub()
	files.storage["managed.conf"] = []byte("Host prod\n  HostName prod.example.com\n\nHost qa\n  HostName qa.example.com\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	err := manager.DeleteHost("prod")
	if err != nil {
		t.Fatalf("DeleteHost() error: %v", err)
	}

	hosts, err := manager.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts() error: %v", err)
	}
	if len(hosts) != 1 || hosts[0].Label != "qa" {
		t.Fatalf("ListHosts() = %+v, want only qa", hosts)
	}

	_, err = manager.GetHost("prod")
	if !errors.Is(err, config.ErrHostNotFound) {
		t.Fatalf("GetHost(prod) error = %v, want ErrHostNotFound", err)
	}
}

// TestManagerListHostsSkipsUnsupportedBlocks verifies CRUD list output ignores non-editable host patterns.
func TestManagerListHostsSkipsUnsupportedBlocks(t *testing.T) {
	files := newFileStub()
	files.storage["managed.conf"] = []byte("Host *.prod\n  User deploy\n\nHost app\n  HostName app.example.com\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	hosts, err := manager.ListHosts()
	if err != nil {
		t.Fatalf("ListHosts() error: %v", err)
	}
	if len(hosts) != 1 || hosts[0].Label != "app" {
		t.Fatalf("ListHosts() = %+v, want only app", hosts)
	}
}

// TestEnsureIncludeAuthoritativePrepends verifies the Include is placed at the top
// of an existing config that does not yet reference the managed file.
func TestEnsureIncludeAuthoritativePrepends(t *testing.T) {
	files := newFileStub()
	files.storage["user.conf"] = []byte("Host legacy\n  User alice\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	if err := manager.EnsureInclude(true); err != nil {
		t.Fatalf("EnsureInclude(true) error: %v", err)
	}

	got := string(files.storage["user.conf"])
	if !strings.HasPrefix(got, "Include managed.conf\n") {
		t.Errorf("expected Include at top, got:\n%s", got)
	}
	if !strings.Contains(got, "Host legacy") {
		t.Errorf("expected original content preserved, got:\n%s", got)
	}
}

// TestEnsureIncludeAuthoritativeRelocates verifies that a bottom-placed Include
// is moved to the top when EnsureInclude is called in authoritative mode.
func TestEnsureIncludeAuthoritativeRelocates(t *testing.T) {
	files := newFileStub()
	files.storage["user.conf"] = []byte("Host legacy\n  User alice\nInclude managed.conf\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	if err := manager.EnsureInclude(true); err != nil {
		t.Fatalf("EnsureInclude(true) error: %v", err)
	}

	got := string(files.storage["user.conf"])
	if !strings.HasPrefix(got, "Include managed.conf\n") {
		t.Errorf("expected Include relocated to top, got:\n%s", got)
	}
	// The Include must not appear twice.
	if strings.Count(got, "Include managed.conf") != 1 {
		t.Errorf("Include directive appears more than once:\n%s", got)
	}
	if !strings.Contains(got, "Host legacy") {
		t.Errorf("expected original content preserved, got:\n%s", got)
	}
}

// TestEnsureIncludeAuthoritativeIdempotent verifies no-op when Include is already at the top.
func TestEnsureIncludeAuthoritativeIdempotent(t *testing.T) {
	files := newFileStub()
	original := "Include managed.conf\nHost legacy\n  User alice\n"
	files.storage["user.conf"] = []byte(original)

	manager := config.NewManager(files, "user.conf", "managed.conf")
	if err := manager.EnsureInclude(true); err != nil {
		t.Fatalf("EnsureInclude(true) error: %v", err)
	}

	if got := string(files.storage["user.conf"]); got != original {
		t.Errorf("file modified unexpectedly:\ngot:  %q\nwant: %q", got, original)
	}
}

// TestEnsureIncludeNonAuthoritativeAppends verifies the Include is placed at the
// bottom when authoritative is false.
func TestEnsureIncludeNonAuthoritativeAppends(t *testing.T) {
	files := newFileStub()
	files.storage["user.conf"] = []byte("Host legacy\n  User alice\n")

	manager := config.NewManager(files, "user.conf", "managed.conf")
	if err := manager.EnsureInclude(false); err != nil {
		t.Fatalf("EnsureInclude(false) error: %v", err)
	}

	got := string(files.storage["user.conf"])
	if strings.HasPrefix(got, "Include") {
		t.Errorf("expected Include at bottom, but it is at the top:\n%s", got)
	}
	if !strings.HasSuffix(got, "Include managed.conf\n") {
		t.Errorf("expected Include at bottom, got:\n%s", got)
	}
}

// TestEnsureIncludeNonAuthoritativeIdempotent verifies no duplicate is added
// when a bottom-placed Include already exists and mode is non-authoritative.
func TestEnsureIncludeNonAuthoritativeIdempotent(t *testing.T) {
	files := newFileStub()
	original := "Host legacy\n  User alice\nInclude managed.conf\n"
	files.storage["user.conf"] = []byte(original)

	manager := config.NewManager(files, "user.conf", "managed.conf")
	if err := manager.EnsureInclude(false); err != nil {
		t.Fatalf("EnsureInclude(false) error: %v", err)
	}

	if got := string(files.storage["user.conf"]); got != original {
		t.Errorf("file modified unexpectedly:\ngot:  %q\nwant: %q", got, original)
	}
}
