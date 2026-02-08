package app

import (
	"embed"
	"os"
	"path/filepath"
	"testing"
)

func TestFindSoundFont_ExternalFile(t *testing.T) {
	// Create a temporary directory with a SoundFont file
	tmpDir := t.TempDir()
	sfPath := filepath.Join(tmpDir, DefaultSoundFontName)

	// Create a dummy SoundFont file
	if err := os.WriteFile(sfPath, []byte("RIFF....sfbk"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Change to temp directory for the test
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	t.Run("finds SoundFont in current directory", func(t *testing.T) {
		// Empty embed.FS (no embedded files)
		var emptyFS embed.FS

		result := findSoundFont(emptyFS, "", false)
		if result == nil {
			t.Fatal("Expected to find SoundFont in current directory")
		}
		if result.IsEmbedded {
			t.Error("Expected external file, got embedded")
		}
		if result.FileSystem != nil {
			t.Error("Expected nil FileSystem for external file")
		}
		if result.Path != DefaultSoundFontName {
			t.Errorf("Expected path %s, got %s", DefaultSoundFontName, result.Path)
		}
	})
}

func TestFindSoundFont_TitlePath(t *testing.T) {
	// Create a temporary directory structure
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "title")
	os.MkdirAll(titleDir, 0755)

	sfPath := filepath.Join(titleDir, DefaultSoundFontName)
	if err := os.WriteFile(sfPath, []byte("RIFF....sfbk"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	t.Run("finds SoundFont in title directory", func(t *testing.T) {
		var emptyFS embed.FS

		result := findSoundFont(emptyFS, titleDir, false)
		if result == nil {
			t.Fatal("Expected to find SoundFont in title directory")
		}
		if result.IsEmbedded {
			t.Error("Expected external file, got embedded")
		}
		if result.Path != sfPath {
			t.Errorf("Expected path %s, got %s", sfPath, result.Path)
		}
	})
}

func TestFindSoundFont_NotFound(t *testing.T) {
	t.Run("returns nil when no SoundFont found", func(t *testing.T) {
		var emptyFS embed.FS

		// Use a non-existent directory
		result := findSoundFont(emptyFS, "/nonexistent/path", false)
		if result != nil {
			t.Error("Expected nil when no SoundFont found")
		}
	})
}

func TestFindSoundFont_Priority(t *testing.T) {
	// Create temp directories
	tmpDir := t.TempDir()
	titleDir := filepath.Join(tmpDir, "title")
	os.MkdirAll(titleDir, 0755)

	// Create SoundFont in both current dir and title dir
	currentSF := filepath.Join(tmpDir, DefaultSoundFontName)
	titleSF := filepath.Join(titleDir, DefaultSoundFontName)

	os.WriteFile(currentSF, []byte("RIFF-current"), 0644)
	os.WriteFile(titleSF, []byte("RIFF-title"), 0644)

	// Change to temp directory
	originalDir, _ := os.Getwd()
	defer os.Chdir(originalDir)
	os.Chdir(tmpDir)

	t.Run("current directory has priority over title directory", func(t *testing.T) {
		var emptyFS embed.FS

		result := findSoundFont(emptyFS, titleDir, false)
		if result == nil {
			t.Fatal("Expected to find SoundFont")
		}
		// Current directory should be found first
		if result.Path != DefaultSoundFontName {
			t.Errorf("Expected current directory SoundFont, got %s", result.Path)
		}
	})
}
