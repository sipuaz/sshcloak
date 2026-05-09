package config

import (
	"sync"
)

// Config is the parsed representation of an ssh_config file.
type Config struct {
	Blocks []*Block

	mu            sync.RWMutex
	resolvedByKey map[resolveKey]map[string][]string
}

type resolveKey struct {
	hostname string
	ctx      MatchContext
}

// Block is a Host or Match stanza plus all its directives.
type Block struct {
	Type       BlockType   // BlockHost or BlockMatch
	Patterns   []string    // e.g. ["*.example.com", "!bastion"]
	Conditions []Condition // for Match blocks
	Directives []*Directive
	Line       int // source line of the Host/Match keyword
}

type BlockType int

const (
	BlockHost  BlockType = iota
	BlockMatch           // Match blocks have structured criteria
)

// Condition is one criterion inside a Match block.
// e.g. "Match User alice Host *.prod"
type Condition struct {
	Keyword string // "User", "Host", "LocalUser", "Exec", etc.
	Value   string
}

// Directive is a single key-value setting inside a block.
// Values is a slice because some keys are multi-valued
// (IdentityFile, LocalForward, etc.).
type Directive struct {
	Key    string   // canonical casing preserved, lowered for lookup
	Values []string // always at least one element
	Line   int
}
