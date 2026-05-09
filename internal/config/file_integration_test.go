//go:build integration

package config_test

import (
	"testing"

	"github.com/sipuaz/sshcloak/internal/config"
)

var fileHandler = config.NewFileHandler()

const testPath = "../resources/ssh_mock_config"

func TestFileHandler_Read(t *testing.T) {
	data, err := fileHandler.Read(testPath)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if len(data) == 0 {
		t.Fatalf("File is empty")
	}
}

func TestFileHandler_Write(t *testing.T) {
	testData := []byte("Test data for write")
	err := fileHandler.Write(testPath, testData, 0644)
	if err != nil {
		t.Fatalf("Failed to write to file: %v", err)
	}

	// Verify that the data was written correctly
	data, err := fileHandler.Read(testPath)
	if err != nil {
		t.Fatalf("Failed to read file after write: %v", err)
	}
	if string(data[len(data)-len(testData):]) != string(testData) {
		t.Fatalf("Data was not written correctly")
	}
}

func TestFileHandler_Append(t *testing.T) {
	testData := []byte("Test data for append")
	err := fileHandler.Append(testPath, testData)
	if err != nil {
		t.Fatalf("Failed to append to file: %v", err)
	}

	// Verify that the data was appended correctly
	data, err := fileHandler.Read(testPath)
	if err != nil {
		t.Fatalf("Failed to read file after append: %v", err)
	}
	if string(data[len(data)-len(testData):]) != string(testData) {
		t.Fatalf("Data was not appended correctly")
	}
}

func TestFileHandler_Exists(t *testing.T) {
	if !fileHandler.Exists(testPath) {
		t.Fatalf("File should exist: %v", testPath)
	}
	if fileHandler.Exists("non_existent_file") {
		t.Fatalf("File should not exist: %v", "non_existent_file")
	}
}
