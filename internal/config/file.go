package config

import (
	"os"
	"path/filepath"
)

// File abstracts filesystem operations used by the manager and config loader.
type File interface {
	Read(path string) ([]byte, error)
	Write(path string, data []byte, perm os.FileMode) error
	// AtomicWrite writes data to path via a same-directory temp file and an
	// os.Rename, guaranteeing that readers never observe a partial write.
	AtomicWrite(path string, data []byte, perm os.FileMode) error
	Append(path string, data []byte) error
	Exists(path string) bool
}

type FileHandler struct{}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (lf *FileHandler) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (lf *FileHandler) Write(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

// AtomicWrite writes data safely by writing to a temp file in the same
// directory as path and then renaming it over the target. Because the source
// and destination reside on the same filesystem, os.Rename is atomic on all
// POSIX-compliant systems. Any partial write is cleaned up via a defer.
func (lf *FileHandler) AtomicWrite(path string, data []byte, perm os.FileMode) error {
	dir := filepath.Dir(path)

	tmp, err := os.CreateTemp(dir, ".sshcloak-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	// removeTemp tracks whether we still own the temp file and must clean it up.
	removeTemp := true
	defer func() {
		if removeTemp {
			os.Remove(tmpName)
		}
	}()

	if err := tmp.Chmod(perm); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	removeTemp = false // os.Rename consumed the temp file; nothing to remove.
	return nil
}

func (lf *FileHandler) Append(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, RWOwnerRAll)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.Write(data)
	return err
}

func (lf *FileHandler) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}
