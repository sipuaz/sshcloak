package config

import (
	"os"
)

type File interface {
    Read(path string) ([]byte, error)
    Write(path string, data []byte, perm os.FileMode) error
	Append(path string, data []byte) error
    Exists(path string) bool
}

type FileHandler struct {}

func NewFileHandler() *FileHandler {
	return &FileHandler{}
}

func (lf *FileHandler) Read(path string) ([]byte, error) {
	return os.ReadFile(path)
}

func (lf *FileHandler) Write(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (lf *FileHandler) Append(path string, data []byte) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, FILE_READ_WRITE_OWNER_READ_ALL)
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