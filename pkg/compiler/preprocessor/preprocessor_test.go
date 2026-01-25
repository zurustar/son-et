// Package preprocessor provides preprocessing functionality for FILLY scripts.
package preprocessor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestPreprocessorBasic tests basic preprocessor functionality.
func TestPreprocessorBasic(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "preprocessor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create test files
	mainContent := `// Main file
int x = 1
#include "helper.tfy"
main() {
    x = 2
}
`
	helperContent := `// Helper file
int y = 10
helper() {
    y = 20
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.tfy"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.tfy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "helper.tfy"), []byte(helperContent), 0644); err != nil {
		t.Fatalf("Failed to write helper.tfy: %v", err)
	}

	// Test preprocessing
	p := New(tmpDir)
	result, err := p.PreprocessFile("main.tfy")
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}

	// Check that helper content is included
	if !strings.Contains(result.Source, "int y = 10") {
		t.Errorf("Expected helper content to be included, got: %s", result.Source)
	}

	// Check that main content is present
	if !strings.Contains(result.Source, "int x = 1") {
		t.Errorf("Expected main content to be present, got: %s", result.Source)
	}

	// Check included files
	if len(result.IncludedFiles) != 2 {
		t.Errorf("Expected 2 included files, got %d: %v", len(result.IncludedFiles), result.IncludedFiles)
	}
}

// TestPreprocessorCircularReference tests circular reference detection.
func TestPreprocessorCircularReference(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "preprocessor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files with circular reference
	aContent := `// File A
#include "b.tfy"
`
	bContent := `// File B
#include "a.tfy"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "a.tfy"), []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.tfy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.tfy"), []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.tfy: %v", err)
	}

	// Test preprocessing - should detect circular reference
	p := New(tmpDir)
	_, err = p.PreprocessFile("a.tfy")
	if err == nil {
		t.Error("Expected circular reference error, got nil")
	}
	if !strings.Contains(err.Error(), "circular") {
		t.Errorf("Expected circular reference error, got: %v", err)
	}
}

// TestPreprocessorIncludeGuard tests that duplicate includes are prevented.
func TestPreprocessorIncludeGuard(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "preprocessor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create files where common.tfy is included twice
	mainContent := `// Main file
#include "a.tfy"
#include "b.tfy"
`
	aContent := `// File A
#include "common.tfy"
int a = 1
`
	bContent := `// File B
#include "common.tfy"
int b = 2
`
	commonContent := `// Common file
int common = 100
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.tfy"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.tfy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "a.tfy"), []byte(aContent), 0644); err != nil {
		t.Fatalf("Failed to write a.tfy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "b.tfy"), []byte(bContent), 0644); err != nil {
		t.Fatalf("Failed to write b.tfy: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tmpDir, "common.tfy"), []byte(commonContent), 0644); err != nil {
		t.Fatalf("Failed to write common.tfy: %v", err)
	}

	// Test preprocessing
	p := New(tmpDir)
	result, err := p.PreprocessFile("main.tfy")
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}

	// Count occurrences of "int common = 100" - should be exactly 1
	count := strings.Count(result.Source, "int common = 100")
	if count != 1 {
		t.Errorf("Expected common content to appear exactly once, got %d times", count)
	}

	// Check that all files are in the included list
	if len(result.IncludedFiles) != 4 {
		t.Errorf("Expected 4 included files, got %d: %v", len(result.IncludedFiles), result.IncludedFiles)
	}
}

// TestPreprocessorFileNotFound tests error handling for missing files.
func TestPreprocessorFileNotFound(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "preprocessor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainContent := `// Main file
#include "nonexistent.tfy"
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.tfy"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.tfy: %v", err)
	}

	// Test preprocessing - should fail with file not found
	p := New(tmpDir)
	_, err = p.PreprocessFile("main.tfy")
	if err == nil {
		t.Error("Expected file not found error, got nil")
	}
}

// TestPreprocessorNoIncludes tests preprocessing a file with no includes.
func TestPreprocessorNoIncludes(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "preprocessor_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	mainContent := `// Main file with no includes
int x = 1
main() {
    x = 2
}
`

	if err := os.WriteFile(filepath.Join(tmpDir, "main.tfy"), []byte(mainContent), 0644); err != nil {
		t.Fatalf("Failed to write main.tfy: %v", err)
	}

	// Test preprocessing
	p := New(tmpDir)
	result, err := p.PreprocessFile("main.tfy")
	if err != nil {
		t.Fatalf("PreprocessFile failed: %v", err)
	}

	// Check that content is preserved
	if !strings.Contains(result.Source, "int x = 1") {
		t.Errorf("Expected content to be preserved, got: %s", result.Source)
	}

	// Check included files
	if len(result.IncludedFiles) != 1 {
		t.Errorf("Expected 1 included file, got %d: %v", len(result.IncludedFiles), result.IncludedFiles)
	}
}

// TestExtractIncludeFilename tests the filename extraction function.
func TestExtractIncludeFilename(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`#include "helper.tfy"`, "helper.tfy"},
		{`#include <helper.tfy>`, "helper.tfy"},
		{`#include helper.tfy`, "helper.tfy"},
		{`#include "path/to/file.tfy"`, "path/to/file.tfy"},
	}

	for _, tt := range tests {
		result := extractIncludeFilename(tt.input)
		if result != tt.expected {
			t.Errorf("extractIncludeFilename(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}
