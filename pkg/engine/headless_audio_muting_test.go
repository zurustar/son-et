package engine

import (
	"bytes"
	"os"
	"testing"

	"github.com/sinshu/go-meltysynth/meltysynth"
)

// TestHeadlessModeAudioMuting tests that audio volume is set to 0 in headless mode.
// **Validates: Requirements 6.1**
func TestHeadlessModeAudioMuting(t *testing.T) {
	// Load soundfont for testing
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}

	// Create test engine in headless mode
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start() // Initialize context

	// Create a minimal MIDI file
	midiData := createTestMIDIFile(1, 10)

	// Add MIDI file to asset loader
	assetLoader.Files["test.mid"] = midiData

	// Create MIDI player and load soundfont
	mp := NewMIDIPlayer(engine)
	mp.soundFont = sf

	// Start playback
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Verify audio is muted in headless mode
	if mp.player == nil {
		t.Fatalf("Expected audio player to be created")
	}

	volume := mp.player.Volume()
	if volume != 0 {
		t.Errorf("Expected audio volume to be 0 in headless mode, but got %f", volume)
	}

	// Stop playback
	mp.Stop()
}
