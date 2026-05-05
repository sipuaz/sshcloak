package config_test

import (
	"strings"
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

func TestParseGlobalDirectivesCollectedBeforeFirstHostBlock(t *testing.T) {
	input := `ServerAliveInterval 60
StrictHostKeyChecking no

Host example
  User alice`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 2 {
		t.Fatalf("expected 2 blocks, got %d", len(cfg.Blocks))
	}
	// First block should be the implicit Host * carrying the global directives.
	global := cfg.Blocks[0]
	if len(global.Patterns) == 0 || global.Patterns[0] != "*" {
		t.Errorf("first block patterns = %v, want [\"*\"]", global.Patterns)
	}
	val := cfg.Resolve("example", config.MatchContext{})["serveraliveinterval"]
	if len(val) == 0 || val[0] != "60" {
		t.Errorf("ServerAliveInterval = %v, want [\"60\"]", val)
	}
}

func TestParseHostBlockWithMultiplePatterns(t *testing.T) {
	input := `Host web-01 web-02 *.staging
  User deploy`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(cfg.Blocks))
	}
	b := cfg.Blocks[0]
	if b.Type != config.BlockHost {
		t.Errorf("block type = %v, want BlockHost", b.Type)
	}
	want := []string{"web-01", "web-02", "*.staging"}
	if len(b.Patterns) != len(want) {
		t.Fatalf("patterns = %v, want %v", b.Patterns, want)
	}
	for i, p := range want {
		if b.Patterns[i] != p {
			t.Errorf("patterns[%d] = %q, want %q", i, b.Patterns[i], p)
		}
	}
}

func TestParseMatchBlockWithKeywordValuePairs(t *testing.T) {
	input := `Match User alice Host *.prod
  Port 2222`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(cfg.Blocks))
	}
	b := cfg.Blocks[0]
	if b.Type != config.BlockMatch {
		t.Errorf("block type = %v, want BlockMatch", b.Type)
	}
	if len(b.Conditions) != 2 {
		t.Fatalf("expected 2 conditions, got %d: %v", len(b.Conditions), b.Conditions)
	}
	if b.Conditions[0].Keyword != "User" || b.Conditions[0].Value != "alice" {
		t.Errorf("conditions[0] = %+v, want {User alice}", b.Conditions[0])
	}
	if b.Conditions[1].Keyword != "Host" || b.Conditions[1].Value != "*.prod" {
		t.Errorf("conditions[1] = %+v, want {Host *.prod}", b.Conditions[1])
	}
}

func TestParseMatchAll(t *testing.T) {
	input := `Match All
  ServerAliveInterval 30`

	cfg, err := config.Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 1 {
		t.Fatalf("expected 1 block, got %d", len(cfg.Blocks))
	}
	b := cfg.Blocks[0]
	if b.Type != config.BlockMatch {
		t.Errorf("block type = %v, want BlockMatch", b.Type)
	}
	if len(b.Conditions) != 1 || b.Conditions[0].Keyword != "All" {
		t.Errorf("conditions = %v, want [{All }]", b.Conditions)
	}
}

func TestParseMatchOddKeywordCountError(t *testing.T) {
	// "Match User alice Host" — 3 tokens (odd, not the "All" special case)
	input := `Match User alice Host
  Port 22`

	_, err := config.Parse(strings.NewReader(input))
	if err == nil {
		t.Fatal("Parse() want error for odd keyword count in Match block, got nil")
	}
}

func TestParseEmptyInput(t *testing.T) {
	cfg, err := config.Parse(strings.NewReader(""))
	if err != nil {
		t.Fatalf("Parse() error: %v", err)
	}
	if len(cfg.Blocks) != 0 {
		t.Errorf("expected 0 blocks for empty input, got %d", len(cfg.Blocks))
	}
}
