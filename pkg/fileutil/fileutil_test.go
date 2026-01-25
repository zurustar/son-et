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

func TestFindFileInPaths(t *testing.T) {
	// Create temporary directories
	tmpDir1 := t.TempDir()
	tmpDir2 := t.TempDir()

	// Create test files in different directories
	file1 := filepath.Join(tmpDir1, "file1.txt")
	file2 := filepath.Join(tmpDir2, "FILE2.TXT")

	if err := os.WriteFile(file1, []byte("test1"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	if err := os.WriteFile(file2, []byte("test2"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name       string
		paths      []string
		searchName string
		shouldFind bool
		expectedIn string
	}{
		{
			name:       "find in first directory",
			paths:      []string{tmpDir1, tmpDir2},
			searchName: "FILE1.TXT",
			shouldFind: true,
			expectedIn: tmpDir1,
		},
		{
			name:       "find in second directory",
			paths:      []string{tmpDir1, tmpDir2},
			searchName: "file2.txt",
			shouldFind: true,
			expectedIn: tmpDir2,
		},
		{
			name:       "file not found in any directory",
			paths:      []string{tmpDir1, tmpDir2},
			searchName: "nonexistent.txt",
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := FindFileInPaths(tt.paths, tt.searchName)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("Expected to find file, but got error: %v", err)
					return
				}

				if !filepath.HasPrefix(path, tt.expectedIn) {
					t.Errorf("Expected file in %s, but got %s", tt.expectedIn, path)
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

func TestResolveFilePath(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()

	// Create test file
	testFile := filepath.Join(tmpDir, "TestFile.WAV")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	tests := []struct {
		name       string
		filename   string
		baseDirs   []string
		shouldFind bool
	}{
		{
			name:       "find relative file with case-insensitive search",
			filename:   "testfile.wav",
			baseDirs:   []string{tmpDir},
			shouldFind: true,
		},
		{
			name:       "absolute path",
			filename:   testFile,
			baseDirs:   []string{},
			shouldFind: true,
		},
		{
			name:       "file not found",
			filename:   "nonexistent.wav",
			baseDirs:   []string{tmpDir},
			shouldFind: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path, err := ResolveFilePath(tt.filename, tt.baseDirs)

			if tt.shouldFind {
				if err != nil {
					t.Errorf("Expected to find file, but got error: %v", err)
					return
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
