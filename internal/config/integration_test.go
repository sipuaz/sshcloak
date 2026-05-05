//go:build integration

package config_test

import (
	"os"
	"strings"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

// ---------------------------------------------------------------------------
// Fixture 1 — typical workstation config
// ---------------------------------------------------------------------------

func TestTypicalWorkstationConfig(t *testing.T) {
	f, err := os.Open("testdata/typical.ssh_config")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	cfg, err := config.Parse(f)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	t.Run("bastion host resolves Hostname and Port", func(t *testing.T) {
		ctx := config.MatchContext{Host: "bastion"}
		hostname, ok := cfg.Get("bastion", "Hostname", ctx)
		if !ok || hostname != "bastion.example.com" {
			t.Errorf("Hostname = %q (%t), want %q", hostname, ok, "bastion.example.com")
		}
		port, ok := cfg.Get("bastion", "Port", ctx)
		if !ok || port != "2222" {
			t.Errorf("Port = %q (%t), want %q", port, ok, "2222")
		}
	})

	t.Run("prod host with deploy user gets prod User and both IdentityFiles", func(t *testing.T) {
		ctx := config.MatchContext{User: "deploy", Host: "web-01.prod"}
		user, ok := cfg.Get("web-01.prod", "User", ctx)
		if !ok || user != "produser" {
			t.Errorf("User = %q (%t), want %q", user, ok, "produser")
		}

		resolved := cfg.Resolve("web-01.prod", ctx)
		ids := resolved["identityfile"]
		wantIDs := []string{"~/.ssh/id_prod", "~/.ssh/id_deploy", "~/.ssh/id_ed25519"}
		if len(ids) != len(wantIDs) {
			t.Fatalf("IdentityFile count = %d, want %d: %v", len(ids), len(wantIDs), ids)
		}
		for i, want := range wantIDs {
			if ids[i] != want {
				t.Errorf("identityfile[%d] = %q, want %q", i, ids[i], want)
			}
		}
	})

	t.Run("unknown host returns only global defaults", func(t *testing.T) {
		ctx := config.MatchContext{Host: "unknown.local"}
		resolved := cfg.Resolve("unknown.local", ctx)

		if user := resolved["user"]; len(user) == 0 || user[0] != "defaultuser" {
			t.Errorf("user = %v, want [defaultuser]", user)
		}
		if ids := resolved["identityfile"]; len(ids) == 0 || ids[0] != "~/.ssh/id_ed25519" {
			t.Errorf("identityfile = %v, want [~/.ssh/id_ed25519]", ids)
		}
		// prod-specific directives must not be present
		if _, ok := resolved["hostname"]; ok {
			t.Error("Hostname should not be set for unknown host")
		}
	})
}

// ---------------------------------------------------------------------------
// Fixture 2 — edge cases
// ---------------------------------------------------------------------------

func TestEdgeCasesConfig(t *testing.T) {
	f, err := os.Open("testdata/edge.ssh_config")
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	defer f.Close()

	cfg, err := config.Parse(f)
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	t.Run("quoted path parsed as single value without surrounding quotes", func(t *testing.T) {
		resolved := cfg.Resolve("anyhost", config.MatchContext{Host: "anyhost"})
		ids := resolved["identityfile"]
		if len(ids) == 0 {
			t.Fatal("IdentityFile not resolved")
		}
		// Must not contain the quote characters.
		if strings.Contains(ids[0], "\"") {
			t.Errorf("IdentityFile contains quote character: %q", ids[0])
		}
		if ids[0] != "/home/user/my keys/id_rsa" {
			t.Errorf("IdentityFile = %q, want %q", ids[0], "/home/user/my keys/id_rsa")
		}
	})

	t.Run("negated pattern excludes matching hostname", func(t *testing.T) {
		// "excluded" matches the !excluded negation — Host block must not apply.
		resolved := cfg.Resolve("excluded", config.MatchContext{Host: "excluded"})
		if user := resolved["user"]; len(user) > 0 && user[0] == "exampleuser" {
			t.Error("User should not be exampleuser for excluded host")
		}
	})

	t.Run("non-excluded host in *.example.com gets User from Host block", func(t *testing.T) {
		resolved := cfg.Resolve("good.example.com", config.MatchContext{Host: "good.example.com"})
		if user := resolved["user"]; len(user) == 0 || user[0] != "exampleuser" {
			t.Errorf("user = %v, want [exampleuser]", user)
		}
	})

	t.Run("Match All block applies to every host", func(t *testing.T) {
		for _, host := range []string{"anyhost", "excluded", "good.example.com"} {
			ctx := config.MatchContext{Host: host}
			resolved := cfg.Resolve(host, ctx)
			if val := resolved["serveralivecountmax"]; len(val) == 0 || val[0] != "5" {
				t.Errorf("host %q: ServerAliveCountMax = %v, want [5]", host, val)
			}
		}
	})

	t.Run("Key=Value syntax parsed correctly", func(t *testing.T) {
		resolved := cfg.Resolve("anyhost", config.MatchContext{Host: "anyhost"})
		if val := resolved["serveraliveinterval"]; len(val) == 0 || val[0] != "30" {
			t.Errorf("ServerAliveInterval = %v, want [30]", val)
		}
	})
}
