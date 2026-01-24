package engine

import (
	"os"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestKuma2Integration tests that the kuma2 sample plays correctly
// and terminates at approximately 18-19 seconds.
// **Validates: Requirements 10.4**
func TestKuma2Integration(t *testing.T) {
	// Check if kuma2 sample exists
	midiPath := "../../samples/kuma2/KUMA.MID"
	if _, err := os.Stat(midiPath); os.IsNotExist(err) {
		t.Skip("kuma2 sample not found, skipping integration test")
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
	assetLoader.Files["KUMA.MID"] = midiData

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
		timeout := time.After(25 * time.Second) // Allow extra time for processing
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
	err = mp.PlayMIDI("KUMA.MID")
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
	case <-time.After(25 * time.Second):
		t.Fatalf("MIDI_END event not received within timeout")
	}

	// Calculate elapsed time
	elapsed := eventTime.Sub(startTime).Seconds()

	// Expected duration: ~18.02 seconds (based on MIDI file analysis with tempo changes)
	// The MIDI file has:
	// - Tempo 0: tick=0, BPM=120 (500000 microsPerBeat)
	// - Tempo 1: tick=1890, BPM=75 (800000 microsPerBeat)
	// - Total ticks: 11520, PPQ: 480
	// Calculated duration:
	//   Segment 1 (0-1890): 1890/480 * 500000/1000000 = 1.97s
	//   Segment 2 (1890-11520): 9630/480 * 800000/1000000 = 16.05s
	//   Total: 18.02s
	// Allow tolerance for processing time and timing variations (±1 second).
	if elapsed < 17.0 || elapsed > 19.0 {
		t.Errorf("MIDI playback duration incorrect: got %.2fs, expected ~18.02s", elapsed)
		t.Logf("NOTE: This failure may indicate that tempo changes are not being handled correctly in playback")
	} else {
		t.Logf("MIDI playback completed in %.2f seconds (expected ~18.02s)", elapsed)
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
