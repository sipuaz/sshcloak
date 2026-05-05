package config

import (
	"bufio"
	"fmt"
	"io"
	"strings"
)

// token represents a single directive in the ssh_config file, along with its line number and raw text for error reporting.
type token struct {
	key    string
	values []string
	line   int
	raw    string
}

// tokenize reads the ssh_config file and produces a list of tokens, one per directive.
func tokenize(r io.Reader) ([]token, error) {
	var tokens []token
	scanner := bufio.NewScanner(r)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		raw := scanner.Text()

		// strip inline comments and trim
		line := strings.TrimSpace(raw)
		if idx := strings.IndexByte(line, COMMENT_CHAR); idx >= 0 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == EMPTY_STR {
			continue
		}

		// split on first '=' or first whitespace run
		// ssh_config allows both "Key Value" and "Key=Value"
		var key, rest string
		if eq := strings.IndexByte(line, EQUAL_CHAR); eq > 0 {
			// check no space before '=' (that would be a value with spaces)
			before := line[:eq]
			if !strings.ContainsAny(before, END_TAB_STR) {
				key = strings.TrimSpace(before)
				rest = strings.TrimSpace(line[eq+1:])
			}
		}
		if key == EMPTY_STR {
			i := strings.IndexAny(line, END_TAB_STR)
			if i < 0 {
				key = line
			} else {
				key = line[:i]
				rest = strings.TrimSpace(line[i:])
			}
		}
		if rest == EMPTY_STR {
			return nil, &ParseError{Line: lineNum, Raw: raw, Msg: "missing value"}
		}

		values, err := splitValues(strings.ToLower(key), rest, lineNum, raw)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token{
			key:    key,
			values: values,
			line:   lineNum,
			raw:    raw,
		})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("line %d: %w", lineNum, err)
	}
	return tokens, nil
}

// splitValues splits the rest of the line into values, depending on whether the key is a multi-value directive.
func splitValues(key, rest string, lineNum int, raw string) ([]string, error) {
	if directives[key] {
		return splitFields(rest, lineNum, raw)
	}
	// Strip enclosing double quotes for single-valued directives.
	if len(rest) >= 2 && rest[0] == DOUBLE_QUOTE_CHAR && rest[len(rest)-1] == DOUBLE_QUOTE_CHAR {
		return []string{rest[1 : len(rest)-1]}, nil
	}
	return []string{rest}, nil
}

// splitFields splits s on whitespace, respecting double-quoted tokens.
// Outer quotes are stripped from each returned token.
// Returns a ParseError if a double-quoted token is not closed.
func splitFields(s string, lineNum int, raw string) ([]string, error) {
	var fields []string
	s = strings.TrimSpace(s)
	for len(s) > 0 {
		if s[0] == DOUBLE_QUOTE_CHAR {
			end := strings.IndexByte(s[1:], DOUBLE_QUOTE_CHAR)
			if end < 0 {
				return nil, &ParseError{Line: lineNum, Raw: raw, Msg: "unclosed double quote"}
			}
			fields = append(fields, s[1:end+1])
			s = strings.TrimSpace(s[end+2:])
		} else {
			i := strings.IndexAny(s, END_TAB_STR)
			if i < 0 {
				fields = append(fields, s)
				break
			}
			fields = append(fields, s[:i])
			s = strings.TrimSpace(s[i:])
		}
	}
	return fields, nil
}
