// Package audio provides audio-related components for the FILLY virtual machine.
// This file contains tests for the WAV Player.
package audio

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// TestNewWAVPlayer tests the creation of a new WAV player.
func TestNewWAVPlayer(t *testing.T) {
	// Test with shared audio context
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)
	if player == nil {
		t.Fatal("NewWAVPlayer returned nil")
	}
	if player.audioCtx == nil {
		t.Error("audioCtx should not be nil")
	}
	if player.players == nil {
		t.Error("players slice should not be nil")
	}
	if player.muted {
		t.Error("player should not be muted by default")
	}
}

// TestWAVPlayerSetMuted tests the mute functionality.
func TestWAVPlayerSetMuted(t *testing.T) {
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	// Initially not muted
	if player.IsMuted() {
		t.Error("player should not be muted initially")
	}

	// Set muted
	player.SetMuted(true)
	if !player.IsMuted() {
		t.Error("player should be muted after SetMuted(true)")
	}

	// Unmute
	player.SetMuted(false)
	if player.IsMuted() {
		t.Error("player should not be muted after SetMuted(false)")
	}
}

// TestWAVPlayerPlayFileNotFound tests error handling for missing files.
// Requirement 5.4: When WAV file is not found, system logs error and continues execution.
func TestWAVPlayerPlayFileNotFound(t *testing.T) {
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	err := player.Play("nonexistent_file.wav")
	if err == nil {
		t.Fatal("Play should return error for nonexistent file")
	}

	if !errors.Is(err, ErrWAVFileNotFound) {
		t.Errorf("expected ErrWAVFileNotFound, got: %v", err)
	}
}

// TestWAVPlayerPlayInvalidFormat tests error handling for invalid WAV files.
// Requirement 5.5: When WAV file is corrupted, system logs error and continues execution.
func TestWAVPlayerPlayInvalidFormat(t *testing.T) {
	// Create a temporary file with invalid WAV data
	tmpDir := t.TempDir()
	invalidFile := filepath.Join(tmpDir, "invalid.wav")
	err := os.WriteFile(invalidFile, []byte("not a valid wav file"), 0644)
	if err != nil {
		t.Fatalf("failed to create test file: %v", err)
	}

	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	err = player.Play(invalidFile)
	if err == nil {
		t.Fatal("Play should return error for invalid WAV file")
	}

	if !errors.Is(err, ErrWAVInvalidFormat) {
		t.Errorf("expected ErrWAVInvalidFormat, got: %v", err)
	}
}

// TestWAVPlayerStopAll tests stopping all active players.
func TestWAVPlayerStopAll(t *testing.T) {
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	// StopAll should not panic even with no active players
	player.StopAll()

	// Verify players list is empty
	if player.GetActivePlayerCount() != 0 {
		t.Error("active player count should be 0 after StopAll")
	}
}

// TestWAVPlayerGetActivePlayerCount tests the active player count.
func TestWAVPlayerGetActivePlayerCount(t *testing.T) {
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	// Initially no active players
	if count := player.GetActivePlayerCount(); count != 0 {
		t.Errorf("expected 0 active players, got %d", count)
	}
}

// TestWAVPlayerUpdate tests the Update method for cleanup.
func TestWAVPlayerUpdate(t *testing.T) {
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	// Update should not panic with no active players
	player.Update()

	// Verify players list is still valid
	if player.GetActivePlayerCount() != 0 {
		t.Error("active player count should be 0 after Update")
	}
}

// TestWAVPlayerGetAudioContext tests getting the audio context.
func TestWAVPlayerGetAudioContext(t *testing.T) {
	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	ctx := player.GetAudioContext()
	if ctx == nil {
		t.Error("GetAudioContext should not return nil")
	}
}

// findSampleWAVFile looks for a sample WAV file in the samples directory.
func findSampleWAVFile() string {
	// Try to find a sample WAV file
	samplePaths := []string{
		"../../../samples/home/BOW.WAV",
		"../../../samples/home/DOKUN.WAV",
		"../../../samples/home/SHOOT.WAV",
		"../../../samples/home/nyao.wav",
		"../../../samples/home/okaeri.wav",
	}

	for _, path := range samplePaths {
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return ""
}

// TestWAVPlayerPlayRealFile tests playing a real WAV file if available.
// This test is skipped if no sample WAV file is found.
// Requirement 5.1: When PlayWAVE(filename) is called, system starts playback of specified WAV file.
// Requirement 5.3: System supports standard WAV file formats (PCM, 8-bit, 16-bit).
func TestWAVPlayerPlayRealFile(t *testing.T) {
	sampleFile := findSampleWAVFile()
	if sampleFile == "" {
		t.Skip("No sample WAV file found, skipping real file test")
	}

	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)
	// Mute to avoid actual audio output during tests
	player.SetMuted(true)

	err := player.Play(sampleFile)
	if err != nil {
		t.Fatalf("Play failed for sample file %s: %v", sampleFile, err)
	}

	// Verify player was added
	if count := player.GetActivePlayerCount(); count != 1 {
		t.Errorf("expected 1 active player, got %d", count)
	}

	// Clean up
	player.StopAll()
}

// TestWAVPlayerMultiplePlayback tests playing multiple WAV files simultaneously.
// Requirement 5.2: When multiple PlayWAVE calls are made, system plays all WAV files simultaneously.
// Requirement 5.6: System mixes multiple WAV streams into a single audio output.
func TestWAVPlayerMultiplePlayback(t *testing.T) {
	sampleFile := findSampleWAVFile()
	if sampleFile == "" {
		t.Skip("No sample WAV file found, skipping multiple playback test")
	}

	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)
	// Mute to avoid actual audio output during tests
	player.SetMuted(true)

	// Play the same file multiple times (simulating multiple WAV playback)
	for i := 0; i < 3; i++ {
		err := player.Play(sampleFile)
		if err != nil {
			t.Fatalf("Play %d failed: %v", i+1, err)
		}
	}

	// Verify all players were added
	if count := player.GetActivePlayerCount(); count != 3 {
		t.Errorf("expected 3 active players, got %d", count)
	}

	// Clean up
	player.StopAll()

	// Verify all players were stopped
	if count := player.GetActivePlayerCount(); count != 0 {
		t.Errorf("expected 0 active players after StopAll, got %d", count)
	}
}

// TestWAVPlayerMutedPlayback tests that muted playback still works.
// Requirement 12.2: When headless mode is enabled, system mutes all audio output.
func TestWAVPlayerMutedPlayback(t *testing.T) {
	sampleFile := findSampleWAVFile()
	if sampleFile == "" {
		t.Skip("No sample WAV file found, skipping muted playback test")
	}

	audioCtx := getSharedAudioContext()
	player := NewWAVPlayer(audioCtx)

	// Set muted before playing
	player.SetMuted(true)

	err := player.Play(sampleFile)
	if err != nil {
		t.Fatalf("Play failed while muted: %v", err)
	}

	// Verify player was added even when muted
	if count := player.GetActivePlayerCount(); count != 1 {
		t.Errorf("expected 1 active player, got %d", count)
	}

	// Clean up
	player.StopAll()
}
