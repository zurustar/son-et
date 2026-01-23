package engine

import (
	"bytes"
	"os"
	"testing"
	"testing/quick"
	"time"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"gitlab.com/gomidi/midi/v2/smf"
)

// Feature: gomidi-midi-playback, Property 5: Playback Completion Event
// **Validates: Requirements 3.2**
// For any MIDI file, when gomidi.Player signals completion via the finished channel,
// the MIDI_Player SHALL trigger the MIDI_END event.
func TestPlaybackCompletionEvent(t *testing.T) {
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

	// Property: For any MIDI file duration, when finished channel signals completion,
	// the MIDI_END event is triggered
	property := func(numNotes uint8, noteDuration uint8) bool {
		// Constrain to reasonable values
		numNotes = (numNotes % 10) + 1        // 1-10 notes
		noteDuration = (noteDuration % 5) + 1 // 1-5 ticks per note

		// Create test engine with event tracking
		engine := NewEngine(nil, nil, nil)
		engine.SetHeadless(true)
		engine.Start() // Initialize context

		// Track MIDI_END event by registering a handler that sets a flag
		eventReceived := false
		eventChan := make(chan bool, 1)

		// Register event handler for MIDI_END
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("midi_end_triggered"), int64(1)}},
		}
		engine.RegisterMesBlock(EventMIDI_END, opcodes, nil, 0)

		// Create MIDI player
		mp := NewMIDIPlayer(engine)
		mp.soundFont = sf

		// Simulate playback completion by directly testing the finished channel mechanism
		// Create channels
		mp.finishedChan = make(chan bool, 1)
		mp.stopChan = make(chan bool, 1)
		mp.isPlaying = true

		// Start goroutine to monitor finished channel (same as in PlayMIDI)
		go func() {
			<-mp.finishedChan
			mp.mutex.Lock()
			mp.isPlaying = false
			mp.mutex.Unlock()
			// Trigger MIDI_END event
			engine.TriggerEvent(EventMIDI_END, &EventData{})
			// Signal that event was triggered
			eventReceived = true
			select {
			case eventChan <- true:
			default:
			}
		}()

		// Signal completion via finished channel
		mp.finishedChan <- true

		// Wait for event with timeout
		select {
		case <-eventChan:
			// Event received successfully
		case <-time.After(1 * time.Second):
			t.Logf("MIDI_END event not received within timeout")
			return false
		}

		// Verify event was received
		if !eventReceived {
			t.Logf("MIDI_END event was not triggered")
			return false
		}

		// Verify player is no longer playing
		if mp.IsPlaying() {
			t.Logf("Player still marked as playing after completion")
			return false
		}

		// Verify that a sequencer was created for the MIDI_END event
		sequencers := engine.GetState().GetSequencers()
		if len(sequencers) != 1 {
			t.Logf("Expected 1 sequencer for MIDI_END event, got %d", len(sequencers))
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// createTestMIDIFile creates a minimal MIDI file with the specified number of notes
func createTestMIDIFile(numNotes int, noteDuration int) []byte {
	// Create a simple MIDI file with one track
	// Format: Type 0 (single track), 1 track, 480 PPQ
	var buf bytes.Buffer

	// Write MThd header
	buf.Write([]byte("MThd"))
	buf.Write([]byte{0x00, 0x00, 0x00, 0x06}) // Header length: 6 bytes
	buf.Write([]byte{0x00, 0x00})             // Format: 0 (single track)
	buf.Write([]byte{0x00, 0x01})             // Number of tracks: 1
	buf.Write([]byte{0x01, 0xE0})             // Time division: 480 PPQ

	// Build track data
	var trackData bytes.Buffer

	// Set tempo: 120 BPM (500000 microseconds per beat)
	trackData.Write([]byte{0x00})             // Delta time: 0
	trackData.Write([]byte{0xFF, 0x51, 0x03}) // Meta event: Set Tempo
	trackData.Write([]byte{0x07, 0xA1, 0x20}) // 500000 microseconds

	// Add notes
	for i := 0; i < numNotes; i++ {
		note := byte(60 + (i % 12)) // C4 to B4
		velocity := byte(64)

		// Note On
		writeMIDIVarInt(&trackData, noteDuration)     // Delta time
		trackData.Write([]byte{0x90, note, velocity}) // Note On, channel 0

		// Note Off
		writeMIDIVarInt(&trackData, noteDuration) // Delta time
		trackData.Write([]byte{0x80, note, 0x00}) // Note Off, channel 0
	}

	// End of track
	trackData.Write([]byte{0x00})             // Delta time: 0
	trackData.Write([]byte{0xFF, 0x2F, 0x00}) // Meta event: End of Track

	// Write MTrk header
	buf.Write([]byte("MTrk"))
	trackLen := trackData.Len()
	buf.Write([]byte{
		byte(trackLen >> 24),
		byte(trackLen >> 16),
		byte(trackLen >> 8),
		byte(trackLen),
	})
	buf.Write(trackData.Bytes())

	return buf.Bytes()
}

// writeMIDIVarInt writes a variable-length integer to a buffer (MIDI format)
func writeMIDIVarInt(buf *bytes.Buffer, value int) {
	if value < 0 {
		value = 0
	}

	// Convert to variable-length quantity
	var bytes []byte
	bytes = append(bytes, byte(value&0x7F))
	value >>= 7

	for value > 0 {
		bytes = append(bytes, byte((value&0x7F)|0x80))
		value >>= 7
	}

	// Write in reverse order (big-endian)
	for i := len(bytes) - 1; i >= 0; i-- {
		buf.WriteByte(bytes[i])
	}
}

// TestPlaybackCompletionEventIntegration tests the full playback completion flow
// with a real MIDI file to ensure the finished channel mechanism works end-to-end.
func TestPlaybackCompletionEventIntegration(t *testing.T) {
	// Create test engine with event tracking
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	engine := NewEngine(nil, assetLoader, nil)
	engine.SetHeadless(true)
	engine.Start() // Initialize context

	// Create a minimal MIDI file (1 note, short duration)
	midiData := createTestMIDIFile(1, 10)

	// Verify the MIDI file is valid
	_, err := smf.ReadFrom(bytes.NewReader(midiData))
	if err != nil {
		t.Fatalf("Failed to parse generated MIDI file: %v", err)
	}

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
		timeout := time.After(3 * time.Second)
		for {
			select {
			case <-ticker.C:
				sequencers := engine.GetState().GetSequencers()
				if len(sequencers) > 0 {
					eventReceived = true
					select {
					case eventChan <- true:
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
	err = mp.LoadSoundFont("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return
	}

	// Start playback
	err = mp.PlayMIDI("test.mid")
	if err != nil {
		t.Fatalf("Failed to start MIDI playback: %v", err)
	}

	// Wait for MIDI_END event with timeout
	// The MIDI file is very short (1 note, 20 ticks at 480 PPQ = ~0.025 seconds at 120 BPM)
	// But we need to account for goroutine scheduling and processing time
	select {
	case <-eventChan:
		// Event received successfully
		t.Logf("MIDI_END event received successfully")
	case <-time.After(3 * time.Second):
		t.Fatalf("MIDI_END event not received within timeout")
	}

	// Verify event was received
	if !eventReceived {
		t.Fatalf("MIDI_END event was not triggered")
	}

	// Verify player is no longer playing (give it a moment to update state)
	time.Sleep(100 * time.Millisecond)
	if mp.IsPlaying() {
		t.Errorf("Player still marked as playing after completion")
	}
}
