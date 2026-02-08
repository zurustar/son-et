package audio

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/zurustar/son-et/pkg/fileutil"
)

func TestReadSoundFontFS_NilFallback(t *testing.T) {
	// Test that nil FileSystem falls back to os.ReadFile
	t.Run("returns error for non-existent file with nil fs", func(t *testing.T) {
		_, err := ReadSoundFontFS(nil, "/nonexistent/path/soundfont.sf2")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("reads existing file with nil fs", func(t *testing.T) {
		// Find the actual SoundFont file
		sfPath := findTestSoundFont(t)
		if sfPath == "" {
			t.Skip("SoundFont file not found, skipping test")
		}

		data, err := ReadSoundFontFS(nil, sfPath)
		if err != nil {
			t.Fatalf("Failed to read SoundFont: %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected non-empty data")
		}

		// Verify it's a valid SoundFont (starts with RIFF)
		if len(data) < 4 || string(data[0:4]) != "RIFF" {
			t.Error("Data does not appear to be a valid SoundFont file")
		}
	})
}

func TestReadSoundFontFS_WithFileSystem(t *testing.T) {
	t.Run("returns error for non-existent file with FileSystem", func(t *testing.T) {
		// Create a RealFS for testing
		tmpDir := t.TempDir()
		fs := fileutil.NewRealFS(tmpDir)

		_, err := ReadSoundFontFS(fs, "nonexistent.sf2")
		if err == nil {
			t.Error("Expected error for non-existent file")
		}
	})

	t.Run("reads file through FileSystem", func(t *testing.T) {
		sfPath := findTestSoundFont(t)
		if sfPath == "" {
			t.Skip("SoundFont file not found, skipping test")
		}

		// Get absolute path and directory
		absPath, err := filepath.Abs(sfPath)
		if err != nil {
			t.Fatalf("Failed to get absolute path: %v", err)
		}
		dir := filepath.Dir(absPath)
		base := filepath.Base(absPath)

		// Create FileSystem with the directory as base
		fs := fileutil.NewRealFS(dir)

		data, err := ReadSoundFontFS(fs, base)
		if err != nil {
			t.Fatalf("Failed to read SoundFont through FileSystem: %v", err)
		}

		if len(data) == 0 {
			t.Error("Expected non-empty data")
		}
	})
}

func TestLoadSoundFontFS_NilFallback(t *testing.T) {
	t.Run("loads and parses SoundFont with nil fs", func(t *testing.T) {
		sfPath := findTestSoundFont(t)
		if sfPath == "" {
			t.Skip("SoundFont file not found, skipping test")
		}

		sf, err := LoadSoundFontFS(nil, sfPath)
		if err != nil {
			t.Fatalf("Failed to load SoundFont: %v", err)
		}

		if sf == nil {
			t.Error("Expected non-nil SoundFont")
		}
	})

	t.Run("returns error for invalid SoundFont", func(t *testing.T) {
		// Create a temporary file with invalid content
		tmpDir := t.TempDir()
		invalidPath := filepath.Join(tmpDir, "invalid.sf2")
		if err := os.WriteFile(invalidPath, []byte("not a soundfont"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		_, err := LoadSoundFontFS(nil, invalidPath)
		if err == nil {
			t.Error("Expected error for invalid SoundFont")
		}
	})
}

// findTestSoundFont searches for a SoundFont file for testing
func findTestSoundFont(t *testing.T) string {
	t.Helper()

	paths := []string{
		"../../../GeneralUser-GS.sf2",
		"../../GeneralUser-GS.sf2",
		"GeneralUser-GS.sf2",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}
