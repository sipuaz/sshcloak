package config

import (
	"errors"
	"strings"
	"testing"
)

func TestTokenize(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantTokens []token
		wantErr    bool
		errMsg     string
	}{
		{
			name:  "key whitespace value",
			input: "User alice",
			wantTokens: []token{
				{key: "User", values: []string{"alice"}, line: 1, raw: "User alice"},
			},
		},
		{
			name:  "key=value",
			input: "User=alice",
			wantTokens: []token{
				{key: "User", values: []string{"alice"}, line: 1, raw: "User=alice"},
			},
		},
		{
			name:  "inline comment stripped",
			input: "User alice # this is ignored",
			wantTokens: []token{
				{key: "User", values: []string{"alice"}, line: 1, raw: "User alice # this is ignored"},
			},
		},
		{
			name:  "blank lines skipped",
			input: "\n  \nUser alice\n\n",
			wantTokens: []token{
				{key: "User", values: []string{"alice"}, line: 3, raw: "User alice"},
			},
		},
		{
			name:  "quoted single value with spaces",
			input: `User "alice bob"`,
			wantTokens: []token{
				{key: "User", values: []string{"alice bob"}, line: 1, raw: `User "alice bob"`},
			},
		},
		{
			name:  "quoted multi-value (IdentityFile)",
			input: `IdentityFile "/path/with spaces/key" /simple/path`,
			wantTokens: []token{
				{key: "IdentityFile", values: []string{"/path/with spaces/key", "/simple/path"}, line: 1, raw: `IdentityFile "/path/with spaces/key" /simple/path`},
			},
		},
		{
			name:    "unclosed quote in multi-valued directive → ParseError",
			input:   `IdentityFile "unclosed`,
			wantErr: true,
			errMsg:  "unclosed double quote",
		},
		{
			name:    "missing value → ParseError",
			input:   "Host",
			wantErr: true,
			errMsg:  "missing value",
		},
		{
			name:    "missing value with equals → ParseError",
			input:   "User=",
			wantErr: true,
			errMsg:  "missing value",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tokenize(strings.NewReader(tc.input))
			if tc.wantErr {
				if err == nil {
					t.Fatalf("tokenize() want error containing %q, got nil", tc.errMsg)
				}
				if !strings.Contains(err.Error(), tc.errMsg) {
					t.Errorf("tokenize() error = %q, want it to contain %q", err.Error(), tc.errMsg)
				}
				return
			}
			if err != nil {
				t.Fatalf("tokenize() unexpected error: %v", err)
			}
			if len(got) != len(tc.wantTokens) {
				t.Fatalf("tokenize() returned %d tokens, want %d", len(got), len(tc.wantTokens))
			}
			for i, want := range tc.wantTokens {
				g := got[i]
				if g.key != want.key {
					t.Errorf("token[%d].key = %q, want %q", i, g.key, want.key)
				}
				if g.line != want.line {
					t.Errorf("token[%d].line = %d, want %d", i, g.line, want.line)
				}
				if g.raw != want.raw {
					t.Errorf("token[%d].raw = %q, want %q", i, g.raw, want.raw)
				}
				if len(g.values) != len(want.values) {
					t.Errorf("token[%d].values = %v, want %v", i, g.values, want.values)
					continue
				}
				for j, v := range want.values {
					if g.values[j] != v {
						t.Errorf("token[%d].values[%d] = %q, want %q", i, j, g.values[j], v)
					}
				}
			}
		})
	}
}

func TestSplitFields(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    []string
		wantErr bool
	}{
		{
			name:  "unquoted tokens",
			input: "foo bar baz",
			want:  []string{"foo", "bar", "baz"},
		},
		{
			name:  "quoted token with spaces inside",
			input: `"foo bar"`,
			want:  []string{"foo bar"},
		},
		{
			name:  "mixed quoted and unquoted",
			input: `"foo bar" baz`,
			want:  []string{"foo bar", "baz"},
		},
		{
			name:  "empty string",
			input: "",
			want:  nil,
		},
		{
			name:    "unclosed quote → error",
			input:   `"unclosed`,
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := splitFields(tc.input, 1, tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("splitFields() want error, got nil")
				}
				var pe *ParseError
				if !errors.As(err, &pe) {
					t.Errorf("splitFields() error is %T, want *ParseError", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitFields() unexpected error: %v", err)
			}
			if len(got) != len(tc.want) {
				t.Fatalf("splitFields() = %v, want %v", got, tc.want)
			}
			for i, v := range tc.want {
				if got[i] != v {
					t.Errorf("splitFields()[%d] = %q, want %q", i, got[i], v)
				}
			}
		})
	}
}
