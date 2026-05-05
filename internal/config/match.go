package config

import (
	"path" // filepath.Match uses glob semantics ssh uses too
	"strings"
)

// MatchContext holds the runtime values used when evaluating Match blocks.
type MatchContext struct {
	Host      string // destination hostname (after canonicalization)
	User      string // remote username
	LocalUser string // local OS user
}

// blockMatches returns true if the block's Host patterns or Match conditions
// apply to the given context.
func blockMatches(b *Block, hostname string, ctx MatchContext) bool {
	switch b.Type {
	case BlockHost:
		return matchPatterns(b.Patterns, hostname)
	case BlockMatch:
		return matchConditions(b.Conditions, ctx)
	}
	return false
}

// resolve returns the effective configuration for the given hostname and context.
func matchPatterns(patterns []string, hostname string) bool {
	for _, p := range patterns {
		if !strings.HasPrefix(p, NEGATION_STR) {
			continue
		}
		ok, _ := path.Match(p[1:], hostname)
		if ok {
			return false
		} // negation wins immediately
	}
	for _, p := range patterns {
		if strings.HasPrefix(p, NEGATION_STR) {
			continue
		}
		if p == STAR_STR {
			return true
		}
		ok, _ := path.Match(p, hostname)
		if ok {
			return true
		}
	}
	return false
}

// matchConditions evaluates the Match block's conditions against the context.
func matchConditions(conds []Condition, ctx MatchContext) bool {
	evaluated := false
	for _, c := range conds {
		switch strings.ToLower(c.Keyword) {
		case ALL_LOWER_STR:
			evaluated = true
			// always passes, continue
		case HOST_STR:
			evaluated = true
			if ok, _ := path.Match(c.Value, ctx.Host); !ok {
				return false
			}
		case USER_STR:
			evaluated = true
			if ok, _ := path.Match(c.Value, ctx.User); !ok {
				return false
			}
		case LOCALUSER_STR:
			evaluated = true
			if ok, _ := path.Match(c.Value, ctx.LocalUser); !ok {
				return false
			}
		default:
			// unrecognized — do not set evaluated
		}
	}
	return evaluated
}
