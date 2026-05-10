package host

import (
	"fmt"
	"sort"
	"strings"
)

// parseExtraPairs converts a slice of "KEY=VALUE" strings into the map format
// expected by config.HostSpec.Extra.  Each element must contain exactly one
// "=" separator; values may themselves contain "=".
func parseExtraPairs(pairs []string) (map[string][]string, error) {
	if len(pairs) == 0 {
		return nil, nil
	}

	extra := make(map[string][]string, len(pairs))
	for _, pair := range pairs {
		idx := strings.IndexByte(pair, '=')
		if idx < 1 {
			return nil, fmt.Errorf("invalid --extra value %q: expected KEY=VALUE", pair)
		}
		key := strings.TrimSpace(pair[:idx])
		val := strings.TrimSpace(pair[idx+1:])
		if key == "" {
			return nil, fmt.Errorf("invalid --extra value %q: key cannot be empty", pair)
		}
		extra[key] = append(extra[key], val)
	}
	return extra, nil
}

// sortedKeys returns the keys of a map in ascending order, used to produce
// deterministic output in "host get".
func sortedKeys(m map[string][]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
