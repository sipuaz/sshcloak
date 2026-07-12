package metadata

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"go.yaml.in/yaml/v3"

	"github.com/sipuaz/sshcloak/internal/config"
)

const (
	metaFilePerm = 0600
)

// Store persists sshcloak-owned host metadata in a YAML sidecar file.
type Store struct {
	files config.File
	path  string
}

// Document is the root YAML structure stored on disk.
type Document struct {
	Hosts map[string]HostMetadata `yaml:"hosts,omitempty"`
}

// HostMetadata holds metadata attached to one host label.
type HostMetadata struct {
	Tags []string `yaml:"tags,omitempty"`
}

// NewStore creates a metadata store bound to one sidecar path.
func NewStore(files config.File, path string) *Store {
	return &Store{files: files, path: path}
}

// Tags returns the sorted tags stored for one host label.
func (s *Store) Tags(label string) ([]string, error) {
	doc, err := s.load()
	if err != nil {
		return nil, err
	}
	meta, ok := doc.Hosts[strings.TrimSpace(label)]
	if !ok {
		return nil, nil
	}
	return append([]string(nil), meta.Tags...), nil
}

// AddTags adds one or more tags to the given host label.
func (s *Store) AddTags(label string, tags ...string) error {
	normalizedLabel := strings.TrimSpace(label)
	if normalizedLabel == "" {
		return errors.New("metadata host label cannot be empty")
	}

	normalizedTags := normalizeTags(tags)
	if len(normalizedTags) == 0 {
		return errors.New("at least one tag is required")
	}

	doc, err := s.load()
	if err != nil {
		return err
	}
	meta := doc.Hosts[normalizedLabel]
	meta.Tags = mergeTags(meta.Tags, normalizedTags)
	doc.Hosts[normalizedLabel] = meta
	return s.persist(doc)
}

// RemoveTags removes one or more tags from the given host label.
func (s *Store) RemoveTags(label string, tags ...string) error {
	normalizedLabel := strings.TrimSpace(label)
	if normalizedLabel == "" {
		return errors.New("metadata host label cannot be empty")
	}

	normalizedTags := normalizeTags(tags)
	if len(normalizedTags) == 0 {
		return errors.New("at least one tag is required")
	}

	doc, err := s.load()
	if err != nil {
		return err
	}
	meta, ok := doc.Hosts[normalizedLabel]
	if !ok {
		return nil
	}

	remaining := subtractTags(meta.Tags, normalizedTags)
	if len(remaining) == 0 {
		delete(doc.Hosts, normalizedLabel)
	} else {
		meta.Tags = remaining
		doc.Hosts[normalizedLabel] = meta
	}
	return s.persist(doc)
}

// TaggedHosts returns host labels whose metadata includes the requested tag.
func (s *Store) TaggedHosts(tag string) ([]string, error) {
	normalizedTag := strings.TrimSpace(tag)
	if normalizedTag == "" {
		return nil, errors.New("tag cannot be empty")
	}

	doc, err := s.load()
	if err != nil {
		return nil, err
	}

	hosts := make([]string, 0)
	for label, meta := range doc.Hosts {
		for _, candidate := range meta.Tags {
			if candidate == normalizedTag {
				hosts = append(hosts, label)
				break
			}
		}
	}
	sort.Strings(hosts)
	return hosts, nil
}

// Orphans returns metadata labels that do not exist in the provided host set.
func (s *Store) Orphans(exists func(string) (bool, error)) ([]string, error) {
	doc, err := s.load()
	if err != nil {
		return nil, err
	}

	orphans := make([]string, 0)
	for label := range doc.Hosts {
		ok, err := exists(label)
		if err != nil {
			return nil, err
		}
		if !ok {
			orphans = append(orphans, label)
		}
	}
	sort.Strings(orphans)
	return orphans, nil
}

func (s *Store) load() (Document, error) {
	doc := Document{Hosts: make(map[string]HostMetadata)}
	if !s.files.Exists(s.path) {
		return doc, nil
	}

	data, err := s.files.Read(s.path)
	if err != nil {
		return Document{}, err
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return doc, nil
	}
	if err := yaml.Unmarshal(data, &doc); err != nil {
		return Document{}, fmt.Errorf("metadata load: %w", err)
	}
	if doc.Hosts == nil {
		doc.Hosts = make(map[string]HostMetadata)
	}

	for label, meta := range doc.Hosts {
		normalizedLabel := strings.TrimSpace(label)
		normalizedTags := normalizeTags(meta.Tags)
		delete(doc.Hosts, label)
		if normalizedLabel == "" || len(normalizedTags) == 0 {
			continue
		}
		doc.Hosts[normalizedLabel] = HostMetadata{Tags: normalizedTags}
	}
	return doc, nil
}

func (s *Store) persist(doc Document) error {
	if err := ensureParentDir(s.path); err != nil {
		return err
	}

	if len(doc.Hosts) == 0 {
		return s.files.AtomicWrite(s.path, []byte("hosts: {}\n"), metaFilePerm)
	}

	labels := make([]string, 0, len(doc.Hosts))
	for label := range doc.Hosts {
		labels = append(labels, label)
	}
	sort.Strings(labels)

	ordered := Document{Hosts: make(map[string]HostMetadata, len(doc.Hosts))}
	for _, label := range labels {
		ordered.Hosts[label] = HostMetadata{Tags: append([]string(nil), doc.Hosts[label].Tags...)}
	}

	data, err := yaml.Marshal(ordered)
	if err != nil {
		return fmt.Errorf("metadata persist: %w", err)
	}
	return s.files.AtomicWrite(s.path, data, metaFilePerm)
}

func normalizeTags(tags []string) []string {
	seen := make(map[string]struct{})
	normalized := make([]string, 0, len(tags))
	for _, tag := range tags {
		trimmed := strings.TrimSpace(tag)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		normalized = append(normalized, trimmed)
	}
	sort.Strings(normalized)
	return normalized
}

func mergeTags(existing, added []string) []string {
	merged := append(append([]string(nil), existing...), added...)
	return normalizeTags(merged)
}

func subtractTags(existing, removed []string) []string {
	blocked := make(map[string]struct{}, len(removed))
	for _, tag := range removed {
		blocked[tag] = struct{}{}
	}
	remaining := make([]string, 0, len(existing))
	for _, tag := range existing {
		if _, ok := blocked[tag]; ok {
			continue
		}
		remaining = append(remaining, tag)
	}
	return normalizeTags(remaining)
}

func ensureParentDir(path string) error {
	directory := filepath.Dir(path)
	if directory == "" || directory == "." {
		return nil
	}
	return os.MkdirAll(directory, 0700)
}
