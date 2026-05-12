package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var (
	// ErrHostNotFound reports that no managed host block matched the requested label.
	ErrHostNotFound = errors.New("host not found")
	// ErrHostAlreadyExists reports that a host label already exists in the readable config tree.
	ErrHostAlreadyExists = errors.New("host already exists")
	// ErrManagedHostUnsupported reports that a block exists but is not editable by the manager.
	ErrManagedHostUnsupported = errors.New("managed host block is not supported")
)

// Manager orchestrates CRUD operations for sshcloak-managed Host blocks.
type Manager struct {
	files             File
	userConfigPath    string
	managedConfigPath string
}

// HostSpec is the canonical representation of one managed Host block.
type HostSpec struct {
	Label         string
	HostName      string
	User          string
	Port          string
	IdentityFiles []string
	Extra         map[string][]string
}

// NewManager builds a manager bound to the user SSH config and the managed include file.
func NewManager(files File, userConfigPath, managedConfigPath string) *Manager {
	return &Manager{
		files:             files,
		userConfigPath:    userConfigPath,
		managedConfigPath: managedConfigPath,
	}
}

// EnsureInclude ensures the root SSH config contains one Include directive for the managed file.
// When authoritative is true (the default) the Include is placed at the very top of the file so
// sshcloak host blocks take precedence over any existing entries (first-match-wins semantics).
// If the Include already exists but is not at the top, it is relocated there.
// When authoritative is false the Include is appended, preserving existing precedence.
func (m *Manager) EnsureInclude(authoritative bool) error {
	if err := ensureParentDir(m.userConfigPath); err != nil {
		return err
	}

	if m.files.Exists(m.userConfigPath) {
		existing, err := m.files.Read(m.userConfigPath)
		if err != nil {
			return err
		}

		cfg, err := m.loadConfig(m.userConfigPath)
		if err != nil {
			return err
		}

		line := formatIncludeLine(m.managedConfigPath)
		alreadyPresent := configIncludesPath(cfg, m.userConfigPath, m.managedConfigPath)

		if authoritative {
			// Already at the top — nothing to do.
			if alreadyPresent && isIncludeAtTop(existing, line) {
				return nil
			}
			// Present but not at the top — strip and prepend.
			if alreadyPresent {
				stripped := removeIncludeLine(existing, line)
				return m.files.AtomicWrite(m.userConfigPath, append([]byte(line), stripped...), RWOwnerRAll)
			}
			// Not present — prepend, separating from existing content with a newline.
			suffix := existing
			if len(existing) > 0 {
				suffix = append([]byte(newLineStr), existing...)
			}
			return m.files.AtomicWrite(m.userConfigPath, append([]byte(line), suffix...), RWOwnerRAll)
		}

		// Non-authoritative: append only if not already present.
		if alreadyPresent {
			return nil
		}
		appendLine := line
		if len(existing) > 0 && existing[len(existing)-1] != newLineChar {
			appendLine = newLineStr + line
		}
		return m.files.Append(m.userConfigPath, []byte(appendLine))
	}

	return m.files.Write(m.userConfigPath, []byte(formatIncludeLine(m.managedConfigPath)), RWOwnerRAll)
}

// ListHosts returns all editable Host blocks from the managed config file.
func (m *Manager) ListHosts() ([]HostSpec, error) {
	cfg, err := m.loadManagedConfig()
	if err != nil {
		return nil, err
	}

	hosts := make([]HostSpec, 0, len(cfg.Blocks))
	for _, block := range cfg.Blocks {
		if !isEditableHostBlock(block) {
			continue
		}
		hosts = append(hosts, blockToHostSpec(block))
	}

	sort.Slice(hosts, func(left, right int) bool {
		return hosts[left].Label < hosts[right].Label
	})

	return hosts, nil
}

// GetHost returns one editable managed Host block by its exact label.
func (m *Manager) GetHost(label string) (HostSpec, error) {
	cfg, err := m.loadManagedConfig()
	if err != nil {
		return HostSpec{}, err
	}

	_, block, err := findHostBlock(cfg, label)
	if err != nil {
		return HostSpec{}, err
	}

	return blockToHostSpec(block), nil
}

// AddHost appends a new editable Host block and bootstraps the Include directive if needed.
func (m *Manager) AddHost(spec HostSpec) error {
	normalized, err := normalizeHostSpec(spec)
	if err != nil {
		return err
	}

	exists, err := m.hostExistsInConfigTree(normalized.Label)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("%w: %s", ErrHostAlreadyExists, normalized.Label)
	}

	if err := m.EnsureInclude(true); err != nil {
		return err
	}

	cfg, err := m.loadManagedConfig()
	if err != nil {
		return err
	}

	cfg.Blocks = append(cfg.Blocks, buildHostBlock(normalized))
	clearResolvedCache(cfg)
	return m.persistManagedConfig(cfg)
}

// UpdateHost replaces one editable managed Host block with the provided spec.
func (m *Manager) UpdateHost(label string, spec HostSpec) error {
	normalized, err := normalizeHostSpec(spec)
	if err != nil {
		return err
	}

	cfg, err := m.loadManagedConfig()
	if err != nil {
		return err
	}

	index, _, err := findHostBlock(cfg, label)
	if err != nil {
		return err
	}

	if normalized.Label != label {
		exists, err := m.hostExistsInConfigTree(normalized.Label)
		if err != nil {
			return err
		}
		if exists {
			return fmt.Errorf("%w: %s", ErrHostAlreadyExists, normalized.Label)
		}
	}

	cfg.Blocks[index] = buildHostBlock(normalized)
	clearResolvedCache(cfg)
	return m.persistManagedConfig(cfg)
}

// DeleteHost removes one editable managed Host block by its exact label.
func (m *Manager) DeleteHost(label string) error {
	cfg, err := m.loadManagedConfig()
	if err != nil {
		return err
	}

	index, _, err := findHostBlock(cfg, label)
	if err != nil {
		return err
	}

	cfg.Blocks = append(cfg.Blocks[:index], cfg.Blocks[index+1:]...)
	clearResolvedCache(cfg)
	return m.persistManagedConfig(cfg)
}

// loadManagedConfig reads and parses the managed include file or returns an empty config.
func (m *Manager) loadManagedConfig() (*Config, error) {
	return m.loadConfig(m.managedConfigPath)
}

// loadConfig reads and parses one config file path, treating a missing file as empty.
func (m *Manager) loadConfig(path string) (*Config, error) {
	if !m.files.Exists(path) {
		return &Config{}, nil
	}

	data, err := m.files.Read(path)
	if err != nil {
		return nil, err
	}

	cfg, err := Parse(strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// persistManagedConfig writes the current managed config in canonical form
// using an atomic rename so readers never observe a partial write.
func (m *Manager) persistManagedConfig(cfg *Config) error {
	if err := ensureParentDir(m.managedConfigPath); err != nil {
		return err
	}
	return m.files.AtomicWrite(m.managedConfigPath, renderConfig(cfg), RWOwnerRAll)
}

// hostExistsInConfigTree reports whether the label exists in the root config or any included config.
func (m *Manager) hostExistsInConfigTree(label string) (bool, error) {
	visited := make(map[string]struct{})
	return m.hostExistsInFile(m.userConfigPath, label, visited)
}

// hostExistsInFile checks a config file and its includes recursively for an exact host label.
func (m *Manager) hostExistsInFile(path, label string, visited map[string]struct{}) (bool, error) {
	resolvedPath := cleanPath(path)
	if _, seen := visited[resolvedPath]; seen {
		return false, nil
	}
	visited[resolvedPath] = struct{}{}

	cfg, err := m.loadConfig(path)
	if err != nil {
		return false, err
	}

	for _, block := range cfg.Blocks {
		if block.Type != BlockHost {
			continue
		}
		for _, pattern := range block.Patterns {
			if pattern == label {
				return true, nil
			}
		}
	}

	for _, includePath := range findIncludedPaths(cfg, path) {
		exists, err := m.hostExistsInFile(includePath, label, visited)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}

	return false, nil
}

// findHostBlock locates one editable block by label and returns its index and value.
func findHostBlock(cfg *Config, label string) (int, *Block, error) {
	const notFoundIdx = -1
	for index, block := range cfg.Blocks {
		if block.Type != BlockHost {
			continue
		}
		if len(block.Patterns) == 1 && block.Patterns[0] == label {
			if !isEditableHostBlock(block) {
				return notFoundIdx, nil, fmt.Errorf("%w: %s", ErrManagedHostUnsupported, label)
			}
			return index, block, nil
		}
	}
	return notFoundIdx, nil, fmt.Errorf("%w: %s", ErrHostNotFound, label)
}

// isEditableHostBlock reports whether a block can be safely managed by CRUD operations.
func isEditableHostBlock(block *Block) bool {
	if block == nil || block.Type != BlockHost || len(block.Patterns) != 1 {
		return false
	}
	pattern := block.Patterns[0]
	if pattern == emptyString || pattern == starStr || strings.HasPrefix(pattern, negationStr) {
		return false
	}
	return !strings.ContainsAny(pattern, jollyHostStr)
}

// buildHostBlock converts one HostSpec into a canonical editable Host block.
func buildHostBlock(spec HostSpec) *Block {
	// fixedDirectiveCount is the number of well-known directives: HostName, User, Port.
	const fixedDirectiveCount = 3
	directives := make([]*Directive, 0, fixedDirectiveCount+len(spec.IdentityFiles)+len(spec.Extra))
	appendIfPresent := func(key, value string) {
		if value == emptyString {
			return
		}
		directives = append(directives, &Directive{Key: key, Values: []string{value}})
	}

	appendIfPresent(hostNameStr, spec.HostName)
	appendIfPresent(userStr, spec.User)
	appendIfPresent(portStr, spec.Port)

	for _, identityFile := range spec.IdentityFiles {
		appendIfPresent(identityFileStr, identityFile)
	}

	extraKeys := sortedExtraKeys(spec.Extra)
	for _, key := range extraKeys {
		for _, value := range spec.Extra[key] {
			appendIfPresent(key, value)
		}
	}

	return &Block{
		Type:       BlockHost,
		Patterns:   []string{spec.Label},
		Directives: directives,
	}
}

// blockToHostSpec converts a managed Host block into its structured representation.
func blockToHostSpec(block *Block) HostSpec {
	spec := HostSpec{
		Label: block.Patterns[0],
		Extra: make(map[string][]string),
	}

	for _, directive := range block.Directives {
		lowerKey := strings.ToLower(directive.Key)
		switch lowerKey {
		case hostNameLowerStr:
			spec.HostName = firstValue(directive.Values)
		case userLowerStr:
			spec.User = firstValue(directive.Values)
		case portLowerStr:
			spec.Port = firstValue(directive.Values)
		case identityFileLowerStr:
			spec.IdentityFiles = append(spec.IdentityFiles, directive.Values...)
		default:
			spec.Extra[directive.Key] = append(spec.Extra[directive.Key], directive.Values...)
		}
	}

	if len(spec.Extra) == 0 {
		spec.Extra = nil
	}

	return spec
}

// normalizeHostSpec validates input and rewrites it into canonical manager form.
func normalizeHostSpec(spec HostSpec) (HostSpec, error) {
	label := strings.TrimSpace(spec.Label)
	if label == emptyString {
		return HostSpec{}, fmt.Errorf("host label cannot be empty")
	}
	if strings.ContainsAny(label, endTabStr) {
		return HostSpec{}, fmt.Errorf("host label cannot contain whitespace: %s", label)
	}
	if label == starStr || strings.HasPrefix(label, negationStr) || strings.ContainsAny(label, jollyHostStr) {
		return HostSpec{}, fmt.Errorf("host label must be an exact alias: %s", label)
	}

	normalized := HostSpec{
		Label:         label,
		HostName:      strings.TrimSpace(spec.HostName),
		User:          strings.TrimSpace(spec.User),
		Port:          strings.TrimSpace(spec.Port),
		IdentityFiles: trimNonEmptyValues(spec.IdentityFiles),
	}

	if len(spec.Extra) > 0 {
		normalized.Extra = make(map[string][]string, len(spec.Extra))
		for key, values := range spec.Extra {
			trimmedKey := strings.TrimSpace(key)
			if trimmedKey == emptyString {
				continue
			}
			filtered := trimNonEmptyValues(values)
			if len(filtered) == 0 {
				continue
			}
			normalized.Extra[trimmedKey] = filtered
		}
		if len(normalized.Extra) == 0 {
			normalized.Extra = nil
		}
	}

	return normalized, nil
}

// sortedExtraKeys returns deterministic ordering for extra directives.
func sortedExtraKeys(extra map[string][]string) []string {
	keys := make([]string, 0, len(extra))
	for key := range extra {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

// trimNonEmptyValues removes blank entries while preserving order.
func trimNonEmptyValues(values []string) []string {
	trimmed := make([]string, 0, len(values))
	for _, value := range values {
		candidate := strings.TrimSpace(value)
		if candidate == emptyString {
			continue
		}
		trimmed = append(trimmed, candidate)
	}
	return trimmed
}

// firstValue returns the first directive value when present.
func firstValue(values []string) string {
	if len(values) == 0 {
		return emptyString
	}
	return values[0]
}

// configIncludesPath reports whether a parsed config already references the managed file.
func configIncludesPath(cfg *Config, sourcePath, targetPath string) bool {
	for _, includePath := range findIncludedPaths(cfg, sourcePath) {
		if sameConfigPath(includePath, targetPath) {
			return true
		}
	}
	return false
}

// findIncludedPaths extracts and resolves Include directives from one parsed config file.
func findIncludedPaths(cfg *Config, sourcePath string) []string {
	if cfg == nil {
		return nil
	}

	baseDir := filepath.Dir(sourcePath)
	paths := make([]string, 0)
	seen := make(map[string]struct{})
	for _, block := range cfg.Blocks {
		for _, directive := range block.Directives {
			if !strings.EqualFold(directive.Key, "Include") {
				continue
			}
			for _, value := range directive.Values {
				for _, includePath := range expandIncludeValue(value, baseDir) {
					resolvedPath := cleanPath(includePath)
					if _, exists := seen[resolvedPath]; exists {
						continue
					}
					seen[resolvedPath] = struct{}{}
					paths = append(paths, resolvedPath)
				}
			}
		}
	}
	return paths
}

// expandIncludeValue resolves one Include directive value relative to its source file.
func expandIncludeValue(value, baseDir string) []string {
	if value == emptyString {
		return nil
	}

	resolved := value
	if !filepath.IsAbs(resolved) {
		resolved = filepath.Join(baseDir, resolved)
	}

	matches, err := filepath.Glob(resolved)
	if err != nil {
		return []string{cleanPath(resolved)}
	}
	if len(matches) == 0 {
		return []string{cleanPath(resolved)}
	}

	for index := range matches {
		matches[index] = cleanPath(matches[index])
	}
	return matches
}

// sameConfigPath compares two config paths after cleaning and absolutizing when possible.
func sameConfigPath(left, right string) bool {
	return cleanPath(left) == cleanPath(right)
}

// cleanPath returns a stable cleaned path representation for comparisons.
func cleanPath(path string) string {
	if path == emptyString {
		return emptyString
	}
	if absolute, err := filepath.Abs(path); err == nil {
		return filepath.Clean(absolute)
	}
	return filepath.Clean(path)
}

// formatIncludeLine renders one Include directive line with a trailing newline.
func formatIncludeLine(path string) string {
	return includeHeader + renderValue(path) + newLineStr
}

// isIncludeAtTop reports whether the Include line for the managed file is the
// very first non-empty, non-comment line in the raw config bytes.
func isIncludeAtTop(raw []byte, includeLine string) bool {
	for _, line := range strings.SplitAfter(string(raw), newLineStr) {
		trimmed := strings.TrimSpace(line)
		if trimmed == emptyString || strings.HasPrefix(trimmed, string(commentChar)) {
			continue
		}
		return strings.TrimSpace(line) == strings.TrimSpace(includeLine)
	}
	return false
}

// removeIncludeLine strips every occurrence of the Include directive for the
// managed file from raw config bytes, cleaning up any double blank lines left
// behind.
func removeIncludeLine(raw []byte, includeLine string) []byte {
	needle := strings.TrimSpace(includeLine)
	var out []byte
	for _, line := range strings.SplitAfter(string(raw), newLineStr) {
		if strings.TrimSpace(line) == needle {
			continue
		}
		out = append(out, []byte(line)...)
	}
	return out
}

// ensureParentDir creates the parent directory for a config file when needed.
func ensureParentDir(path string) error {
	directory := filepath.Dir(path)
	if directory == emptyString || directory == currentDir {
		return nil
	}
	return os.MkdirAll(directory, RWXOwnerRAll)
}

// clearResolvedCache invalidates the cached resolved directive map after mutations.
func clearResolvedCache(cfg *Config) {
	if cfg != nil {
		cfg.resolvedByKey = nil
	}
}
