package engine

import (
	"bytes"
	"os"
	"testing"
	"testing/quick"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Feature: gomidi-midi-playback, Property 11: Context Cancellation Stops Playback
// **Validates: Requirements 7.3**
// For any active MIDI playback, when the engine context is cancelled,
// the MIDI_Player SHALL stop playback.
func TestContextCancellationStopsPlayback(t *testing.T) {
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

	// Property: For any MIDI playback, when context is cancelled, playback stops
	property := func(numNotes uint8, cancelDelayMs uint16) bool {
		// Constrain to reasonable values
		numNotes = (numNotes % 20) + 5             // 5-24 notes (longer playback)
		cancelDelayMs = (cancelDelayMs % 200) + 50 // 50-249ms delay before cancellation

		// Create test engine with cancellable context
		assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
		engine := NewEngine(nil, assetLoader, nil)
		engine.SetHeadless(true)
		engine.Start() // Initialize context

		// Create a MIDI file with enough notes to ensure playback is active when cancelled
		midiData := createTestMIDIFile(int(numNotes), 20) // 20 ticks per note

		// Verify the MIDI file is valid
		_, err := smf.ReadFrom(bytes.NewReader(midiData))
		if err != nil {
			t.Logf("Failed to parse generated MIDI file: %v", err)
			return false
		}

		// Add MIDI file to asset loader
		assetLoader.Files["test.mid"] = midiData

		// Create MIDI player and load soundfont
		mp := NewMIDIPlayer(engine)
		mp.soundFont = sf

		// Start playback
		err = mp.PlayMIDI("test.mid")
		if err != nil {
			t.Logf("Failed to start MIDI playback: %v", err)
			return false
		}

		// Wait a bit to ensure playback has started
		time.Sleep(10 * time.Millisecond)

		// Verify playback is active
		if !mp.IsPlaying() {
			t.Logf("Playback did not start")
			return false
		}

		// Wait for the specified delay before cancelling
		time.Sleep(time.Duration(cancelDelayMs) * time.Millisecond)

		// Cancel the context
		engine.Terminate() // This cancels the engine context

		// Wait for cancellation to propagate
		time.Sleep(100 * time.Millisecond)

		// Verify playback has stopped
		if mp.IsPlaying() {
			t.Logf("Playback did not stop after context cancellation")
			return false
		}

		// Verify the audio player was closed
		mp.mutex.Lock()
		playerClosed := mp.player == nil
		mp.mutex.Unlock()

		if !playerClosed {
			t.Logf("Audio player was not closed after context cancellation")
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// TestContextCancellationBeforePlayback tests that context cancellation
// before playback starts is handled correctly.
func TestContextCancellationBeforePlayback(t *testing.T) {
	// Create test engine with timeout of 1ms (will cancel almost immediately)
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.SetTimeout(1 * time.Millisecond)
	engine.Start()

	// Wait for context to be cancelled
	time.Sleep(10 * time.Millisecond)

	// Verify context is cancelled
	select {
	case <-engine.GetContext().Done():
		// Context is cancelled as expected
	default:
		t.Fatalf("Context should be cancelled")
	}

	// Create a minimal MIDI file
	midiData := createTestMIDIFile(5, 20)
	assetLoader.Files["test.mid"] = midiData

	// Load soundfont
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}

	// Create MIDI player
	mp := NewMIDIPlayer(engine)
	mp.soundFont = sf

	// Try to start playback with cancelled context
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("PlayMIDI failed: %v", err)
	}

	// Wait a bit for goroutines to process
	time.Sleep(100 * time.Millisecond)

	// Verify playback did not start or stopped immediately
	if mp.IsPlaying() {
		t.Errorf("Playback should not be active with cancelled context")
	}
}

// TestContextCancellationInHeadlessMode tests that context cancellation
// works correctly in headless mode.
func TestContextCancellationInHeadlessMode(t *testing.T) {
	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start()

	// Create a MIDI file
	midiData := createTestMIDIFile(10, 20)
	assetLoader.Files["test.mid"] = midiData

	// Load soundfont
	sfData, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	sf, err := meltysynth.NewSoundFont(bytes.NewReader(sfData))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}

	// Create MIDI player
	mp := NewMIDIPlayer(engine)
	mp.soundFont = sf

	// Start playback
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Wait for playback to start
	time.Sleep(10 * time.Millisecond)

	// Verify playback is active
	if !mp.IsPlaying() {
		t.Fatalf("Playback did not start")
	}

	// Call UpdateHeadless a few times to simulate headless mode operation
	for i := 0; i < 5; i++ {
		mp.UpdateHeadless()
		time.Sleep(10 * time.Millisecond)
	}

	// Cancel the context
	engine.Terminate()

	// Wait for cancellation to propagate
	time.Sleep(100 * time.Millisecond)

	// Verify playback has stopped
	if mp.IsPlaying() {
		t.Errorf("Playback did not stop after context cancellation in headless mode")
	}

	// Call UpdateHeadless again to verify it handles cancelled context gracefully
	mp.UpdateHeadless()

	// Verify it doesn't panic or cause issues
	if mp.IsPlaying() {
		t.Errorf("UpdateHeadless should not restart playback after cancellation")
	}
}
