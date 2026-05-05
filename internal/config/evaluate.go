package config

import "strings"

// Resolve returns the effective directives for a hostname.
// It implements first-match-wins: once a key is set, later blocks
// cannot override it (except for multi-valued keys like IdentityFile).
func (cfg *Config) Resolve(hostname string, ctx MatchContext) map[string][]string {
	return cloneResolved(cfg.resolve(hostname, ctx))
}

func (cfg *Config) resolve(hostname string, ctx MatchContext) map[string][]string {
	cacheKey := resolveKey{hostname: hostname, ctx: ctx}

	cfg.mu.RLock()
	if cached, ok := cfg.resolvedByKey[cacheKey]; ok {
		cfg.mu.RUnlock()
		return cached
	}
	cfg.mu.RUnlock()

	result := cfg.computeResolved(hostname, ctx)

	cfg.mu.Lock()
	defer cfg.mu.Unlock()

	if cfg.resolvedByKey == nil {
		cfg.resolvedByKey = make(map[resolveKey]map[string][]string)
	}
	if _, exists := cfg.resolvedByKey[cacheKey]; !exists {
		cfg.resolvedByKey[cacheKey] = result
	}
	return cfg.resolvedByKey[cacheKey]
}

func (cfg *Config) computeResolved(hostname string, ctx MatchContext) map[string][]string {
	result := make(map[string][]string)

	for _, block := range cfg.Blocks {
		if !blockMatches(block, hostname, ctx) {
			continue
		}
		for _, d := range block.Directives {
			lower := strings.ToLower(d.Key)
			if isMultiValued(lower) {
				result[lower] = append(result[lower], d.Values...)
			} else if _, seen := result[lower]; !seen {
				result[lower] = cloneValues(d.Values)
			}
		}
	}

	return result
}

// Get is a convenience wrapper returning the first value for a key.
// It holds the read lock while extracting the scalar value so the
// internal cached map reference never escapes to callers.
func (cfg *Config) Get(hostname, key string, ctx MatchContext) (string, bool) {
	cfg.mu.RLock()
	cached, ok := cfg.resolvedByKey[resolveKey{hostname: hostname, ctx: ctx}]
	if ok {
		vals := cached[strings.ToLower(key)]
		cfg.mu.RUnlock()
		if len(vals) == 0 {
			return EMPTY_STR, false
		}
		return vals[0], true
	}
	cfg.mu.RUnlock()

	// cache miss — compute outside any lock, then store under write lock
	result := cfg.computeResolved(hostname, ctx)

	cfg.mu.Lock()
	defer cfg.mu.Unlock()
	if cfg.resolvedByKey == nil {
		cfg.resolvedByKey = make(map[resolveKey]map[string][]string)
	}
	if _, exists := cfg.resolvedByKey[resolveKey{hostname: hostname, ctx: ctx}]; !exists {
		cfg.resolvedByKey[resolveKey{hostname: hostname, ctx: ctx}] = result
	}
	vals := result[strings.ToLower(key)]
	if len(vals) == 0 {
		return EMPTY_STR, false
	}
	return vals[0], true
}

func cloneResolved(resolved map[string][]string) map[string][]string {
	clone := make(map[string][]string, len(resolved))
	for key, values := range resolved {
		clone[key] = cloneValues(values)
	}
	return clone
}

func cloneValues(values []string) []string {
	return append([]string(nil), values...)
}

// isMultiValued returns true if the given key is a multi-valued directive.
func isMultiValued(key string) bool {
	return directives[key]
}
