package config

import (
	"testing"
)

func TestMatchPatterns(t *testing.T) {
	tests := []struct {
		name     string
		patterns []string
		hostname string
		want     bool
	}{
		{
			name:     "positive glob matches",
			patterns: []string{"*.example.com"},
			hostname: "foo.example.com",
			want:     true,
		},
		{
			name:     "positive glob no match",
			patterns: []string{"*.example.com"},
			hostname: "foo.other.com",
			want:     false,
		},
		{
			name:     "star wildcard matches everything",
			patterns: []string{"*"},
			hostname: "anything.at.all",
			want:     true,
		},
		{
			name:     "empty pattern list",
			patterns: []string{},
			hostname: "example.com",
			want:     false,
		},
		{
			name:     "negation only — negated host excluded",
			patterns: []string{"!excluded.com"},
			hostname: "excluded.com",
			want:     false,
		},
		{
			name:     "negation only — non-negated host still no match (no positive pattern)",
			patterns: []string{"!excluded.com"},
			hostname: "other.com",
			want:     false,
		},
		{
			name:     "positive and negation — negated host excluded",
			patterns: []string{"*.example.com", "!bad.example.com"},
			hostname: "bad.example.com",
			want:     false,
		},
		{
			name:     "positive and negation — non-negated host matches",
			patterns: []string{"*.example.com", "!bad.example.com"},
			hostname: "good.example.com",
			want:     true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchPatterns(tc.patterns, tc.hostname)
			if got != tc.want {
				t.Errorf("matchPatterns(%v, %q) = %t, want %t", tc.patterns, tc.hostname, got, tc.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// matchConditions
// ---------------------------------------------------------------------------

func TestMatchConditions(t *testing.T) {
	tests := []struct {
		name  string
		conds []Condition
		ctx   MatchContext
		want  bool
	}{
		{
			name:  "empty slice — no conditions evaluated → false",
			conds: []Condition{},
			ctx:   MatchContext{},
			want:  false,
		},
		{
			name:  "nil slice — no conditions evaluated → false",
			conds: nil,
			ctx:   MatchContext{},
			want:  false,
		},
		{
			name:  "Match All — always true",
			conds: []Condition{{Keyword: "All"}},
			ctx:   MatchContext{},
			want:  true,
		},
		{
			name:  "all recognized keywords pass",
			conds: []Condition{{Keyword: "User", Value: "alice"}, {Keyword: "Host", Value: "*.prod"}},
			ctx:   MatchContext{User: "alice", Host: "web.prod"},
			want:  true,
		},
		{
			name:  "one failing keyword — returns false",
			conds: []Condition{{Keyword: "User", Value: "alice"}, {Keyword: "Host", Value: "*.prod"}},
			ctx:   MatchContext{User: "alice", Host: "dev.local"},
			want:  false,
		},
		{
			name:  "all unrecognized keywords — evaluated=false → returns false",
			conds: []Condition{{Keyword: "Exec", Value: "true"}, {Keyword: "Tagged", Value: "x"}},
			ctx:   MatchContext{},
			want:  false,
		},
		{
			name:  "unrecognized mixed with recognized passing",
			conds: []Condition{{Keyword: "Exec", Value: "true"}, {Keyword: "User", Value: "alice"}},
			ctx:   MatchContext{User: "alice"},
			want:  true,
		},
		{
			name:  "LocalUser condition passes",
			conds: []Condition{{Keyword: "LocalUser", Value: "simone"}},
			ctx:   MatchContext{LocalUser: "simone"},
			want:  true,
		},
		{
			name:  "LocalUser condition fails",
			conds: []Condition{{Keyword: "LocalUser", Value: "simone"}},
			ctx:   MatchContext{LocalUser: "other"},
			want:  false,
		},
		{
			name:  "keyword matching is case-insensitive",
			conds: []Condition{{Keyword: "USER", Value: "alice"}},
			ctx:   MatchContext{User: "alice"},
			want:  true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := matchConditions(tc.conds, tc.ctx)
			if got != tc.want {
				t.Errorf("matchConditions(%v, %v) = %t, want %t", tc.conds, tc.ctx, got, tc.want)
			}
		})
	}
}
