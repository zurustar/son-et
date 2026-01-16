package engine

import (
	"os"
	"path/filepath"
	"testing"
)

// TestFilesystemAssetLoader_ReadFile tests basic file reading
func TestFilesystemAssetLoader_ReadFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir := t.TempDir()

	// Create a test file
	testContent := []byte("test content")
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Create loader
	loader := NewFilesystemAssetLoader(tempDir)

	// Test reading the file
	data, err := loader.ReadFile("test.txt")
	if err != nil {
		t.Errorf("ReadFile failed: %v", err)
	}

	if string(data) != string(testContent) {
		t.Errorf("ReadFile returned wrong content: got %q, want %q", string(data), string(testContent))
	}
}

// TestFilesystemAssetLoader_ReadFile_CaseInsensitive tests case-insensitive file matching
func TestFilesystemAssetLoader_ReadFile_CaseInsensitive(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file with uppercase name
	testContent := []byte("test content")
	testFile := filepath.Join(tempDir, "TEST.BMP")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tempDir)

	// Test reading with lowercase name
	data, err := loader.ReadFile("test.bmp")
	if err != nil {
		t.Errorf("ReadFile with lowercase name failed: %v", err)
	}

	if string(data) != string(testContent) {
		t.Errorf("ReadFile returned wrong content: got %q, want %q", string(data), string(testContent))
	}
}

// TestFilesystemAssetLoader_ReadFile_MixedCase tests various case combinations
func TestFilesystemAssetLoader_ReadFile_MixedCase(t *testing.T) {
	tempDir := t.TempDir()

	testCases := []struct {
		name        string
		actualFile  string
		requestName string
		content     string
	}{
		{
			name:        "Uppercase file, lowercase request",
			actualFile:  "IMAGE.BMP",
			requestName: "image.bmp",
			content:     "uppercase content",
		},
		{
			name:        "Lowercase file, uppercase request",
			actualFile:  "picture.bmp",
			requestName: "PICTURE.BMP",
			content:     "lowercase content",
		},
		{
			name:        "Mixed case file, different case request",
			actualFile:  "MyImage.BMP",
			requestName: "myimage.bmp",
			content:     "mixed case content",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create test file
			testFile := filepath.Join(tempDir, tc.actualFile)
			if err := os.WriteFile(testFile, []byte(tc.content), 0644); err != nil {
				t.Fatalf("Failed to create test file: %v", err)
			}

			loader := NewFilesystemAssetLoader(tempDir)

			// Test reading with different case
			data, err := loader.ReadFile(tc.requestName)
			if err != nil {
				t.Errorf("ReadFile failed: %v", err)
			}

			if string(data) != tc.content {
				t.Errorf("ReadFile returned wrong content: got %q, want %q", string(data), tc.content)
			}

			// Clean up for next test
			os.Remove(testFile)
		})
	}
}

// TestFilesystemAssetLoader_ReadFile_RelativePath tests relative path handling
func TestFilesystemAssetLoader_ReadFile_RelativePath(t *testing.T) {
	tempDir := t.TempDir()

	testContent := []byte("relative path content")
	testFile := filepath.Join(tempDir, "relative.txt")
	if err := os.WriteFile(testFile, testContent, 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tempDir)

	// Test various relative path formats
	testCases := []string{
		"relative.txt",
		"./relative.txt",
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			data, err := loader.ReadFile(path)
			if err != nil {
				t.Errorf("ReadFile with path %q failed: %v", path, err)
			}

			if string(data) != string(testContent) {
				t.Errorf("ReadFile returned wrong content: got %q, want %q", string(data), string(testContent))
			}
		})
	}
}

// TestFilesystemAssetLoader_ReadFile_NotFound tests error handling for missing files
func TestFilesystemAssetLoader_ReadFile_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	loader := NewFilesystemAssetLoader(tempDir)

	// Try to read a non-existent file
	_, err := loader.ReadFile("nonexistent.txt")
	if err == nil {
		t.Error("ReadFile should return error for non-existent file")
	}
}

// TestFilesystemAssetLoader_ReadDir tests directory reading
func TestFilesystemAssetLoader_ReadDir(t *testing.T) {
	tempDir := t.TempDir()

	// Create some test files
	files := []string{"file1.txt", "file2.bmp", "file3.mid"}
	for _, file := range files {
		testFile := filepath.Join(tempDir, file)
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
	}

	loader := NewFilesystemAssetLoader(tempDir)

	// Read directory
	entries, err := loader.ReadDir(".")
	if err != nil {
		t.Fatalf("ReadDir failed: %v", err)
	}

	// Check that we got the expected number of entries
	if len(entries) != len(files) {
		t.Errorf("ReadDir returned wrong number of entries: got %d, want %d", len(entries), len(files))
	}

	// Check that all files are present
	foundFiles := make(map[string]bool)
	for _, entry := range entries {
		foundFiles[entry.Name()] = true
	}

	for _, file := range files {
		if !foundFiles[file] {
			t.Errorf("ReadDir did not return file %q", file)
		}
	}
}

// TestFilesystemAssetLoader_ReadDir_RelativePath tests directory reading with relative paths
func TestFilesystemAssetLoader_ReadDir_RelativePath(t *testing.T) {
	tempDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tempDir, "test.txt")
	if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tempDir)

	// Test various relative path formats
	testCases := []string{
		".",
		"./",
	}

	for _, path := range testCases {
		t.Run(path, func(t *testing.T) {
			entries, err := loader.ReadDir(path)
			if err != nil {
				t.Errorf("ReadDir with path %q failed: %v", path, err)
			}

			if len(entries) == 0 {
				t.Errorf("ReadDir returned no entries for path %q", path)
			}
		})
	}
}

// TestFilesystemAssetLoader_ReadDir_NotFound tests error handling for missing directories
func TestFilesystemAssetLoader_ReadDir_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	loader := NewFilesystemAssetLoader(tempDir)

	// Try to read a non-existent directory
	_, err := loader.ReadDir("nonexistent")
	if err == nil {
		t.Error("ReadDir should return error for non-existent directory")
	}
}

// TestFilesystemAssetLoader_BMP tests loading actual BMP files
func TestFilesystemAssetLoader_BMP(t *testing.T) {
	tempDir := t.TempDir()

	// Create a minimal valid BMP file (1x1 pixel, 24-bit)
	bmpData := []byte{
		// BMP Header
		0x42, 0x4D, // "BM"
		0x36, 0x00, 0x00, 0x00, // File size: 54 bytes
		0x00, 0x00, 0x00, 0x00, // Reserved
		0x36, 0x00, 0x00, 0x00, // Offset to pixel data: 54 bytes
		// DIB Header (BITMAPINFOHEADER)
		0x28, 0x00, 0x00, 0x00, // Header size: 40 bytes
		0x01, 0x00, 0x00, 0x00, // Width: 1 pixel
		0x01, 0x00, 0x00, 0x00, // Height: 1 pixel
		0x01, 0x00, // Planes: 1
		0x18, 0x00, // Bits per pixel: 24
		0x00, 0x00, 0x00, 0x00, // Compression: none
		0x00, 0x00, 0x00, 0x00, // Image size: 0 (uncompressed)
		0x00, 0x00, 0x00, 0x00, // X pixels per meter: 0
		0x00, 0x00, 0x00, 0x00, // Y pixels per meter: 0
		0x00, 0x00, 0x00, 0x00, // Colors used: 0
		0x00, 0x00, 0x00, 0x00, // Important colors: 0
		// Pixel data (BGR format, padded to 4-byte boundary)
		0xFF, 0xFF, 0xFF, 0x00, // White pixel + padding
	}

	testFile := filepath.Join(tempDir, "test.bmp")
	if err := os.WriteFile(testFile, bmpData, 0644); err != nil {
		t.Fatalf("Failed to create test BMP file: %v", err)
	}

	loader := NewFilesystemAssetLoader(tempDir)

	// Test reading the BMP file
	data, err := loader.ReadFile("test.bmp")
	if err != nil {
		t.Errorf("ReadFile failed for BMP: %v", err)
	}

	if len(data) != len(bmpData) {
		t.Errorf("ReadFile returned wrong size: got %d, want %d", len(data), len(bmpData))
	}
}
