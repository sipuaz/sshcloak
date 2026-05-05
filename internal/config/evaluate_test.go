package config_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

func TestGetCachesResolvedLookup(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{{
			Type:     config.BlockHost,
			Patterns: []string{"example"},
			Directives: []*config.Directive{{
				Key:    "User",
				Values: []string{"alice"},
			}},
		}},
	}

	user, ok := cfg.Get("example", "User", config.MatchContext{})
	if !ok || user != "alice" {
		t.Fatalf("Get() = (%q, %t), want (%q, %t)", user, ok, "alice", true)
	}

	// Mutate the underlying block after caching — Get must still return cached value.
	cfg.Blocks[0].Directives[0].Values[0] = "bob"

	user, ok = cfg.Get("example", "User", config.MatchContext{})
	if !ok || user != "alice" {
		t.Fatalf("Get() after mutation = (%q, %t), want cached (%q, %t)", user, ok, "alice", true)
	}
}

func TestResolveReturnsIndependentCopies(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{{
			Type:     config.BlockHost,
			Patterns: []string{"example"},
			Directives: []*config.Directive{{
				Key:    "IdentityFile",
				Values: []string{"id_a"},
			}},
		}},
	}

	resolved := cfg.Resolve("example", config.MatchContext{})
	resolved["identityfile"][0] = "id_b"

	again := cfg.Resolve("example", config.MatchContext{})
	if got := again["identityfile"][0]; got != "id_a" {
		t.Fatalf("Resolve() returned shared cached data: got %q, want %q", got, "id_a")
	}
}

func TestParseQuotedIdentityFile(t *testing.T) {
	input := `Host example
  IdentityFile "path with spaces/key"
  IdentityFile /unquoted/path`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	vals := cfg.Resolve("example", config.MatchContext{})["identityfile"]
	if len(vals) != 2 {
		t.Fatalf("expected 2 IdentityFile values, got %d: %v", len(vals), vals)
	}
	if vals[0] != "path with spaces/key" {
		t.Errorf("vals[0] = %q, want %q", vals[0], "path with spaces/key")
	}
	if vals[1] != "/unquoted/path" {
		t.Errorf("vals[1] = %q, want %q", vals[1], "/unquoted/path")
	}
}

func TestParseQuotedSingleValue(t *testing.T) {
	input := `Host example
  User "alice bob"`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}

	user, ok := cfg.Get("example", "User", config.MatchContext{})
	if !ok {
		t.Fatal("User not found in resolved directives")
	}
	if user != "alice bob" {
		t.Errorf("Get(User) = %q, want %q", user, "alice bob")
	}
}

func TestParseNoImplicitGlobalBlock(t *testing.T) {
	input := `Host example
  User alice`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d; implicit Host * block was added unnecessarily", len(cfg.Blocks))
	}
}

func TestParseImplicitGlobalBlockPresentWhenNeeded(t *testing.T) {
	input := `StrictHostKeyChecking no

Host example
  User alice`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 2 {
		t.Fatalf("expected 2 blocks (implicit Host* + explicit Host), got %d", len(cfg.Blocks))
	}
	if len(cfg.Blocks[0].Patterns) == 0 || cfg.Blocks[0].Patterns[0] != "*" {
		t.Errorf("first block should be implicit Host *, got patterns %v", cfg.Blocks[0].Patterns)
	}
}

// ---------------------------------------------------------------------------
// New evaluate tests
// ---------------------------------------------------------------------------

func TestFirstMatchWinsForSingleValuedDirectives(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Type:     config.BlockHost,
				Patterns: []string{"*.prod"},
				Directives: []*config.Directive{{
					Key:    "User",
					Values: []string{"produser"},
				}},
			},
			{
				Type:     config.BlockHost,
				Patterns: []string{"*"},
				Directives: []*config.Directive{{
					Key:    "User",
					Values: []string{"defaultuser"},
				}},
			},
		},
	}

	user, ok := cfg.Get("web-01.prod", "User", config.MatchContext{})
	if !ok || user != "produser" {
		t.Errorf("Get(User) = (%q, %t), want (\"produser\", true)", user, ok)
	}
}

func TestMultiValuedDirectivesAccumulate(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{
			{
				Type:     config.BlockHost,
				Patterns: []string{"example"},
				Directives: []*config.Directive{{
					Key:    "IdentityFile",
					Values: []string{"id_a", "id_b"},
				}},
			},
			{
				Type:     config.BlockHost,
				Patterns: []string{"*"},
				Directives: []*config.Directive{{
					Key:    "IdentityFile",
					Values: []string{"id_c"},
				}},
			},
		},
	}

	vals := cfg.Resolve("example", config.MatchContext{})["identityfile"]
	if len(vals) != 3 {
		t.Fatalf("expected 3 IdentityFile values (accumulation), got %d: %v", len(vals), vals)
	}
}

func TestGetMissingKeyReturnsFalse(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{{
			Type:     config.BlockHost,
			Patterns: []string{"example"},
			Directives: []*config.Directive{{
				Key:    "User",
				Values: []string{"alice"},
			}},
		}},
	}

	val, ok := cfg.Get("example", "Port", config.MatchContext{})
	if ok || val != "" {
		t.Errorf("Get(missing key) = (%q, %t), want (\"\", false)", val, ok)
	}
}

func TestGetConcurrentCallsDoNotRace(t *testing.T) {
	cfg := &config.Config{
		Blocks: []*config.Block{{
			Type:     config.BlockHost,
			Patterns: []string{"*"},
			Directives: []*config.Directive{
				{Key: "User", Values: []string{"alice"}},
				{Key: "Port", Values: []string{"22"}},
			},
		}},
	}

	const goroutines = 50
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func(n int) {
			defer wg.Done()
			host := "host"
			if n%2 == 0 {
				host = "other"
			}
			cfg.Get(host, "User", config.MatchContext{})
			cfg.Get(host, "Port", config.MatchContext{})
		}(i)
	}
	wg.Wait()
}
