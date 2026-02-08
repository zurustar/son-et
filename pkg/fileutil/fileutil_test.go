package fileutil

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindFileCaseInsensitive(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()

	// Create test files with various cases
	testFiles := []string{
		"TestFile.txt",
		"UPPERCASE.WAV",
		"lowercase.mid",
		"MixedCase.BMP",
	}

	for _, filename := range testFiles {
		path := filepath.Join(tmpDir, filename)
		if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	tests := []struct {
		name          string
		searchName    string
		shouldFind    bool
		expectedMatch string
	}{
		{
			name:          "exact match",
			searchName:    "TestFile.txt",
			shouldFind:    true,
			expectedMatch: "TestFile.txt",
		},
		{
			name:          "lowercase search for mixed case file",
			searchName:    "testfile.txt",
			shouldFind:    true,
			expectedMatch: "TestFile.txt",
		},
		{
			name:          "uppercase search for mixed case file",
			searchName:    "TESTFILE.TXT",
			shouldFind:    true,
			expectedMatch: "TestFile.txt",
		},
		{
			name:          "mixed case search for uppercase file",
			searchName:    "Uppercase.wav",
			shouldFind:    true,
			expectedMatch: "UPPERCASE.WAV",
		},
		{
			name:          "uppercase search for lowercase file",
			searchName:    "LOWERCASE.MID",
			shouldFind:    true,
			expectedMatch: "lowercase.mid",
		},
		{
			name:       "file not found",
			searchName: "nonexistent.txt",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := FindFileCaseInsensitive(tmpDir, tt.searchName)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("Expected to find file, but got error: %v", err)
					return
				}

				actualFilename := filepath.Base(path)
				if actualFilename != tt.expectedMatch {
					t.Errorf("Expected filename %s, got %s", tt.expectedMatch, actualFilename)
				}

				// Verify the file actually exists
				if _, err := os.Stat(path); err != nil {
					t.Errorf("Returned path does not exist: %s", path)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for non-existent file, but got path: %s", path)
				}
			}
		})
	}
}


