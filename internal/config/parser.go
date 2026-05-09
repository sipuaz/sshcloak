package config

import (
	"io"
	"strings"
)

// Parse reads an ssh_config and returns the AST.
func Parse(r io.Reader) (*Config, error) {
	tokens, err := tokenize(r)
	if err != nil {
		return nil, err
	}

	cfg := &Config{}
	var current *Block

	// Directives before any Host/Match line go into an implicit "Host *" block.
	// It is only prepended to cfg.Blocks if it actually collects directives.
	global := &Block{Type: BlockHost, Patterns: []string{starStr}}
	current = global

	for _, tkn := range tokens {
		lower := strings.ToLower(tkn.key)

		switch lower {
		case hostStr:
			b := &Block{
				Type:     BlockHost,
				Patterns: tkn.values,
				Line:     tkn.line,
			}
			cfg.Blocks = append(cfg.Blocks, b)
			current = b

		case matchStr:
			b, err := parseMatchBlock(tkn)
			if err != nil {
				return nil, err
			}
			cfg.Blocks = append(cfg.Blocks, b)
			current = b

		default:
			current.Directives = append(current.Directives, &Directive{
				Key:    tkn.key,
				Values: tkn.values,
				Line:   tkn.line,
			})
		}
	}

	if len(global.Directives) > 0 {
		cfg.Blocks = append([]*Block{global}, cfg.Blocks...)
	}

	return cfg, nil
}

// parseMatchBlock converts a "Match" token into a Block with Conditions.
func parseMatchBlock(tkn token) (*Block, error) {
	// values: ["User", "alice", "Host", "*.prod"]  (keyword-value pairs)
	b := &Block{Type: BlockMatch, Line: tkn.line}
	vals := tkn.values
	// Should return a ParseError when len(vals) is odd and not the All special case.
	if len(vals)%2 != 0 && !(len(vals) == 1 && strings.EqualFold(vals[0], allLowerStr)) {
		return nil, &ParseError{
			Line: tkn.line,
			Msg:  "Match block must have an even number of keyword-value pairs or be 'Match All'",
		}
	}
	for i := 0; i+1 < len(vals); i += 2 {
		b.Conditions = append(b.Conditions, Condition{
			Keyword: vals[i],
			Value:   vals[i+1],
		})
	}
	// "Match All" is a special single-token case
	if len(vals) == 1 && strings.EqualFold(vals[0], allLowerStr) {
		b.Conditions = append(b.Conditions, Condition{Keyword: allStr})
	}
	return b, nil
}
