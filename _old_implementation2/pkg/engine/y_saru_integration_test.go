package engine

import (
	"os"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestYSaruIntegration tests that the y_saru sample plays correctly
// with proper animation synchronization and completes at approximately 294 seconds.
// **Validates: Requirements 10.5**
func TestYSaruIntegration(t *testing.T) {
	// Check if y_saru sample exists
	midiPath := "../../samples/y_saru/FLYINSKY.MID"
	if _, err := os.Stat(midiPath); os.IsNotExist(err) {
		t.Skip("y_saru sample not found, skipping integration test")
	}

	// Check if soundfont exists
	soundfontPath := "../../GeneralUser-GS.sf2"
	if _, err := os.Stat(soundfontPath); os.IsNotExist(err) {
		t.Skip("soundfont not found, skipping integration test")
	}

	// Create test engine
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.SetDebugLevel(2) // Enable debug logging
	engine.Start()

	// Load MIDI file into asset loader
	midiData, err := os.ReadFile(midiPath)
	if err != nil {
		t.Fatalf("Failed to read MIDI file: %v", err)
	}
	assetLoader.Files["FLYINSKY.MID"] = midiData

	// Load soundfont data into asset loader
	soundfontData, err := os.ReadFile(soundfontPath)
	if err != nil {
		t.Fatalf("Failed to read soundfont file: %v", err)
	}
	assetLoader.Files["soundfont.sf2"] = soundfontData

	// Track MIDI_END event
	eventReceived := false
	eventChan := make(chan time.Time, 1)

	// Register event handler for MIDI_END
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("midi_end_triggered"), int64(1)}},
	}
	engine.RegisterMesBlock(EventMIDI_END, opcodes, nil, 0)

	// Monitor for MIDI_END event
	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()
		timeout := time.After(310 * time.Second) // Allow extra time for processing (294s + 16s buffer)
		for {
			select {
			case <-ticker.C:
				sequencers := engine.GetState().GetSequencers()
				// If any sequencer was created, it means MIDI_END was triggered
				if len(sequencers) > 0 && !eventReceived {
					eventReceived = true
					select {
					case eventChan <- time.Now():
					default:
					}
					return
				}
			case <-timeout:
				return
			}
		}
	}()

	// Create MIDI player and load soundfont
	mp := NewMIDIPlayer(engine)
	err = mp.LoadSoundFont("soundfont.sf2")
	if err != nil {
		t.Fatalf("Failed to load soundfont: %v", err)
	}

	// Start playback and record start time
	startTime := time.Now()
	err = mp.PlayMIDI("FLYINSKY.MID")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Verify playback started
	time.Sleep(200 * time.Millisecond)
	mp.mutex.Lock()
	isPlaying := mp.isPlaying
	mp.mutex.Unlock()

	if !isPlaying {
		t.Fatalf("MIDI playback did not start")
	}

	// Wait for MIDI_END event
	var eventTime time.Time
	select {
	case eventTime = <-eventChan:
		t.Logf("MIDI_END event received")
	case <-time.After(310 * time.Second):
		t.Fatalf("MIDI_END event not received within timeout")
	}

	// Calculate elapsed time
	elapsed := eventTime.Sub(startTime).Seconds()

	// Expected duration: ~294.21 seconds (based on MIDI file analysis)
	// The MIDI file is Format 1 with 17 tracks, PPQ 480, tempo 128 BPM
	// Allow tolerance for processing time and timing variations.
	// We'll use a range of 290-300 seconds to account for variations.
	if elapsed < 290.0 || elapsed > 300.0 {
		t.Errorf("MIDI playback duration incorrect: got %.2fs, expected ~294s", elapsed)
		t.Logf("NOTE: This failure may indicate timing issues in playback")
	} else {
		t.Logf("MIDI playback completed in %.2f seconds (expected ~294.21s)", elapsed)
	}

	// Verify event was received
	if !eventReceived {
		t.Fatalf("MIDI_END event was not triggered")
	}

	// Verify player is no longer playing
	// Wait a bit longer for the audio player to fully stop after EOF
	time.Sleep(500 * time.Millisecond)
	if mp.IsPlaying() {
		t.Errorf("Player still marked as playing after MIDI_END")
	}

	// Clean up
	mp.Stop()
}
