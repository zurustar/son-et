package engine

import (
	"embed"
	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/*
var testFS embed.FS

func TestFilesystemAssetLoader_ReadFile(t *testing.T) {
	// Create temp directory with test files
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testData := []byte("hello world")
	if err := os.WriteFile(testFile, testData, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tmpDir)

	// Test exact match
	data, err := loader.ReadFile("test.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("Expected 'hello world', got %s", string(data))
	}

	// Test case-insensitive match
	data, err = loader.ReadFile("TEST.TXT")
	if err != nil {
		t.Errorf("Case-insensitive ReadFile failed: %v", err)
	}
	if string(data) != "hello world" {
		t.Errorf("Expected 'hello world', got %s", string(data))
	}

	// Test non-existent file
	_, err = loader.ReadFile("missing.txt")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestFilesystemAssetLoader_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "exists.txt")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tmpDir)

	// Test exact match
	if !loader.Exists("exists.txt") {
		t.Error("Expected file to exist")
	}

	// Test case-insensitive match
	if !loader.Exists("EXISTS.TXT") {
		t.Error("Expected case-insensitive match to exist")
	}

	// Test non-existent file
	if loader.Exists("missing.txt") {
		t.Error("Expected file to not exist")
	}
}

func TestFilesystemAssetLoader_ListFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	files := []string{"test1.txt", "test2.txt", "data.bin"}
	for _, file := range files {
		path := filepath.Join(tmpDir, file)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	loader := NewFilesystemAssetLoader(tmpDir)

	// Test pattern matching
	matches, err := loader.ListFiles("*.txt")
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}
	if len(matches) != 2 {
		t.Errorf("Expected 2 matches, got %d", len(matches))
	}

	// Test all files
	matches, err = loader.ListFiles("*")
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}
	if len(matches) != 3 {
		t.Errorf("Expected 3 matches, got %d", len(matches))
	}
}

func TestFilesystemAssetLoader_CaseInsensitiveSubdirectory(t *testing.T) {
	tmpDir := t.TempDir()

	// Create subdirectory with file
	subDir := filepath.Join(tmpDir, "SubDir")
	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}
	testFile := filepath.Join(subDir, "File.txt")
	if err := os.WriteFile(testFile, []byte("nested"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tmpDir)

	// Test case-insensitive path
	data, err := loader.ReadFile("subdir/file.txt")
	if err != nil {
		t.Errorf("Case-insensitive subdirectory read failed: %v", err)
	}
	if string(data) != "nested" {
		t.Errorf("Expected 'nested', got %s", string(data))
	}
}

func TestEmbedFSAssetLoader_ReadFile(t *testing.T) {
	loader := NewEmbedFSAssetLoader(testFS)

	// Test reading embedded file
	data, err := loader.ReadFile("testdata/sample.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}
	if len(data) == 0 {
		t.Error("Expected non-empty data")
	}

	// Test non-existent file
	_, err = loader.ReadFile("testdata/missing.txt")
	if err == nil {
		t.Error("Expected error for missing file")
	}
}

func TestEmbedFSAssetLoader_Exists(t *testing.T) {
	loader := NewEmbedFSAssetLoader(testFS)

	// Test existing file
	if !loader.Exists("testdata/sample.txt") {
		t.Error("Expected file to exist")
	}

	// Test non-existent file
	if loader.Exists("testdata/missing.txt") {
		t.Error("Expected file to not exist")
	}
}

func TestEmbedFSAssetLoader_ListFiles(t *testing.T) {
	loader := NewEmbedFSAssetLoader(testFS)

	// Test pattern matching
	matches, err := loader.ListFiles("*.txt")
	if err != nil {
		t.Errorf("ListFiles failed: %v", err)
	}
	if len(matches) == 0 {
		t.Error("Expected at least one match")
	}
}
