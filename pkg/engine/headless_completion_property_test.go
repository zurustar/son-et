package engine

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Feature: gomidi-midi-playback, Property 10: Headless Mode Completion Event
// **Validates: Requirements 6.3**
// For any MIDI playback in headless mode, the MIDI_Player SHALL trigger the
// MIDI_END event when playback completes.
func TestHeadlessModeCompletionEvent(t *testing.T) {
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

	// Create a minimal MIDI file (1 note, short duration)
	midiData := createTestMIDIFile(1, 10)

	// Add MIDI file to asset loader
	assetLoader.Files["test.mid"] = midiData

	// Track MIDI_END event
	eventReceived := false
	eventChan := make(chan bool, 1)

	// Register event handler for MIDI_END
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("midi_end_triggered"), int64(1)}},
	}
	engine.RegisterMesBlock(EventMIDI_END, opcodes, nil, 0)

	// Monitor sequencers to detect when MIDI_END event is triggered
	go func() {
		ticker := time.NewTicker(50 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(5 * time.Second)
		for {
			select {
			case <-ticker.C:
				sequencers := engine.GetState().GetSequencers()
				// Look for a sequencer that was created by the MIDI_END event
				for _, seq := range sequencers {
					if seq.GetVariable("midi_end_triggered") == int64(1) {
						eventReceived = true
						select {
						case eventChan <- true:
						default:
						}
						return
					}
				}
			case <-timeout:
				return
			}
		}
	}()

	// Create MIDI player and load soundfont
	mp := NewMIDIPlayer(engine)
	mp.soundFont = sf

	// Start playback
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Verify audio is muted in headless mode
	if mp.player != nil && mp.player.Volume() != 0 {
		t.Errorf("Expected audio to be muted in headless mode, but volume is %f", mp.player.Volume())
	}

	// Wait for MIDI_END event with timeout
	// The MIDI file is very short (1 note, 20 ticks at 480 PPQ = ~0.025 seconds at 120 BPM)
	// But we need to account for goroutine scheduling and processing time
	select {
	case <-eventChan:
		// Event received successfully
		t.Logf("MIDI_END event received successfully in headless mode")
	case <-time.After(5 * time.Second):
		t.Fatalf("MIDI_END event not received within timeout in headless mode")
	}

	// Verify event was received
	if !eventReceived {
		t.Fatalf("MIDI_END event was not triggered in headless mode")
	}

	// Verify player is no longer playing (give it a moment to update state)
	time.Sleep(100 * time.Millisecond)
	if mp.IsPlaying() {
		t.Errorf("Player still marked as playing after completion in headless mode")
	}
}
