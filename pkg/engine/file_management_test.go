package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// TestCopyFile tests file copying
// Requirement 28.1: WHEN CopyFile is called, THE Runtime SHALL copy a file from source to destination path
func TestCopyFile(t *testing.T) {
	// Create a temporary source file
	srcFile := "test_src.txt"
	dstFile := "test_dst.txt"
	defer os.Remove(srcFile)
	defer os.Remove(dstFile)

	// Write test content
	content := []byte("Hello, World!")
	err := os.WriteFile(srcFile, content, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Copy the file
	result := CopyFile(srcFile, dstFile)
	if result != 0 {
		t.Errorf("CopyFile failed with result %d", result)
	}

	// Verify destination file exists and has same content
	dstContent, err := os.ReadFile(dstFile)
	if err != nil {
		t.Fatalf("Failed to read destination file: %v", err)
	}

	if string(dstContent) != string(content) {
		t.Errorf("Content mismatch: expected '%s', got '%s'", content, dstContent)
	}
}

// TestDelFile tests file deletion
// Requirement 28.2: WHEN DelFile is called, THE Runtime SHALL delete the specified file
func TestDelFile(t *testing.T) {
	// Create a temporary file
	testFile := "test_delete.txt"
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Delete the file
	result := DelFile(testFile)
	if result != 0 {
		t.Errorf("DelFile failed with result %d", result)
	}

	// Verify file no longer exists
	if _, err := os.Stat(testFile); !os.IsNotExist(err) {
		t.Error("File still exists after deletion")
	}
}

// TestIsExist tests file existence checking
// Requirement 28.3: WHEN IsExist is called, THE Runtime SHALL return whether the specified file exists
func TestIsExist(t *testing.T) {
	// Test with existing file
	testFile := "test_exist.txt"
	err := os.WriteFile(testFile, []byte("test"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(testFile)

	result := IsExist(testFile)
	if result != 1 {
		t.Errorf("IsExist returned %d for existing file, expected 1", result)
	}

	// Test with non-existing file
	result = IsExist("nonexistent_file.txt")
	if result != 0 {
		t.Errorf("IsExist returned %d for non-existing file, expected 0", result)
	}
}

// TestMkDir tests directory creation
// Requirement 28.4: WHEN MkDir is called, THE Runtime SHALL create the specified directory
func TestMkDir(t *testing.T) {
	testDir := "test_mkdir"
	defer os.RemoveAll(testDir)

	// Create directory
	result := MkDir(testDir)
	if result != 0 {
		t.Errorf("MkDir failed with result %d", result)
	}

	// Verify directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
}

// TestMkDir_Nested tests creating nested directories
// Requirement 28.4: WHEN MkDir is called, THE Runtime SHALL create the specified directory
func TestMkDir_Nested(t *testing.T) {
	testDir := filepath.Join("test_mkdir_nested", "subdir", "subsubdir")
	defer os.RemoveAll("test_mkdir_nested")

	// Create nested directories
	result := MkDir(testDir)
	if result != 0 {
		t.Errorf("MkDir failed with result %d", result)
	}

	// Verify nested directory exists
	info, err := os.Stat(testDir)
	if err != nil {
		t.Fatalf("Nested directory was not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("Created path is not a directory")
	}
}

// TestRmDir tests directory removal
// Requirement 28.5: WHEN RmDir is called, THE Runtime SHALL remove the specified directory
func TestRmDir(t *testing.T) {
	testDir := "test_rmdir"

	// Create directory
	err := os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}

	// Remove directory
	result := RmDir(testDir)
	if result != 0 {
		t.Errorf("RmDir failed with result %d", result)
	}

	// Verify directory no longer exists
	if _, err := os.Stat(testDir); !os.IsNotExist(err) {
		t.Error("Directory still exists after removal")
	}
}

// TestChDir tests changing working directory
// Requirement 28.6: WHEN ChDir is called, THE Runtime SHALL change the current working directory
func TestChDir(t *testing.T) {
	// Get original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir) // Restore original directory

	// Create a test directory
	testDir := "test_chdir"
	os.RemoveAll(testDir) // Clean up any existing directory
	err = os.Mkdir(testDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create test directory: %v", err)
	}
	defer os.RemoveAll(testDir)

	// Change to test directory
	result := ChDir(testDir)
	if result != 0 {
		t.Errorf("ChDir failed with result %d", result)
	}

	// Verify we're in the new directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	expectedDir := filepath.Join(originalDir, testDir)
	if currentDir != expectedDir {
		t.Errorf("ChDir did not change directory: expected %s, got %s", expectedDir, currentDir)
	}
}

// TestGetCWD tests getting current working directory
// Requirement 28.7: WHEN GetCWD is called, THE Runtime SHALL return the current working directory path
func TestGetCWD(t *testing.T) {
	// Get expected directory using os.Getwd
	expected, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Get directory using GetCWD
	result := GetCWD()

	if result != expected {
		t.Errorf("GetCWD returned %s, expected %s", result, expected)
	}
}

// TestFileOperations_Integration tests a sequence of file operations
// Requirements 28.1-28.7: Integration test for file operations
func TestFileOperations_Integration(t *testing.T) {
	// Get original directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	// Create a test directory structure
	testDir := "test_integration"
	defer os.RemoveAll(testDir)

	// 1. Create directory
	if result := MkDir(testDir); result != 0 {
		t.Fatal("Failed to create test directory")
	}

	// 2. Change to directory
	if result := ChDir(testDir); result != 0 {
		t.Fatal("Failed to change to test directory")
	}

	// 3. Verify we're in the right place
	cwd := GetCWD()
	if !filepath.IsAbs(cwd) || filepath.Base(cwd) != testDir {
		t.Errorf("Not in expected directory: %s", cwd)
	}

	// 4. Create a file
	testFile := "test.txt"
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// 5. Check file exists
	if result := IsExist(testFile); result != 1 {
		t.Error("File should exist")
	}

	// 6. Copy file
	copyFile := "test_copy.txt"
	if result := CopyFile(testFile, copyFile); result != 0 {
		t.Fatal("Failed to copy file")
	}

	// 7. Verify copy exists
	if result := IsExist(copyFile); result != 1 {
		t.Error("Copied file should exist")
	}

	// 8. Delete original
	if result := DelFile(testFile); result != 0 {
		t.Fatal("Failed to delete file")
	}

	// 9. Verify original is gone
	if result := IsExist(testFile); result != 0 {
		t.Error("Original file should not exist")
	}

	// 10. Change back to original directory
	if result := ChDir(originalDir); result != 0 {
		t.Fatal("Failed to change back to original directory")
	}
}
