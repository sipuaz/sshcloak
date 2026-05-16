package metadata_test

import (
	"os"
	"testing"

	"github.com/sipuaz/sshcloak/internal/metadata"
)

type fileStub struct {
	storage map[string][]byte
}

func newFileStub() *fileStub {
	return &fileStub{storage: make(map[string][]byte)}
}

func (s *fileStub) Read(path string) ([]byte, error) {
	data, ok := s.storage[path]
	if !ok {
		return nil, os.ErrNotExist
	}
	return data, nil
}

func (s *fileStub) Write(path string, data []byte, perm os.FileMode) error {
	s.storage[path] = data
	return nil
}

func (s *fileStub) AtomicWrite(path string, data []byte, perm os.FileMode) error {
	return s.Write(path, data, perm)
}

func (s *fileStub) Append(path string, data []byte) error {
	current, _ := s.Read(path)
	s.storage[path] = append(current, data...)
	return nil
}

func (s *fileStub) Exists(path string) bool {
	_, ok := s.storage[path]
	return ok
}

func TestStoreAddTagsCreatesAndSorts(t *testing.T) {
	files := newFileStub()
	store := metadata.NewStore(files, "meta.yaml")

	if err := store.AddTags("prod", "web", "production", "web"); err != nil {
		t.Fatalf("AddTags() error: %v", err)
	}

	tags, err := store.Tags("prod")
	if err != nil {
		t.Fatalf("Tags() error: %v", err)
	}
	if len(tags) != 2 || tags[0] != "production" || tags[1] != "web" {
		t.Fatalf("Tags() = %v, want [production web]", tags)
	}
}

func TestStoreRemoveTagsDeletesEmptyHostEntry(t *testing.T) {
	files := newFileStub()
	store := metadata.NewStore(files, "meta.yaml")

	if err := store.AddTags("prod", "web"); err != nil {
		t.Fatalf("AddTags() error: %v", err)
	}
	if err := store.RemoveTags("prod", "web"); err != nil {
		t.Fatalf("RemoveTags() error: %v", err)
	}

	tags, err := store.Tags("prod")
	if err != nil {
		t.Fatalf("Tags() error: %v", err)
	}
	if len(tags) != 0 {
		t.Fatalf("Tags() = %v, want empty", tags)
	}
}

func TestStoreTaggedHostsReturnsSortedMatches(t *testing.T) {
	files := newFileStub()
	store := metadata.NewStore(files, "meta.yaml")

	if err := store.AddTags("db", "production"); err != nil {
		t.Fatalf("AddTags(db) error: %v", err)
	}
	if err := store.AddTags("app", "production", "web"); err != nil {
		t.Fatalf("AddTags(app) error: %v", err)
	}

	hosts, err := store.TaggedHosts("production")
	if err != nil {
		t.Fatalf("TaggedHosts() error: %v", err)
	}
	if len(hosts) != 2 || hosts[0] != "app" || hosts[1] != "db" {
		t.Fatalf("TaggedHosts() = %v, want [app db]", hosts)
	}
}

func TestStoreOrphansFiltersUnknownHosts(t *testing.T) {
	files := newFileStub()
	store := metadata.NewStore(files, "meta.yaml")

	if err := store.AddTags("known", "production"); err != nil {
		t.Fatalf("AddTags(known) error: %v", err)
	}
	if err := store.AddTags("orphan", "stale"); err != nil {
		t.Fatalf("AddTags(orphan) error: %v", err)
	}

	orphans, err := store.Orphans(func(label string) (bool, error) {
		return label == "known", nil
	})
	if err != nil {
		t.Fatalf("Orphans() error: %v", err)
	}
	if len(orphans) != 1 || orphans[0] != "orphan" {
		t.Fatalf("Orphans() = %v, want [orphan]", orphans)
	}
}
