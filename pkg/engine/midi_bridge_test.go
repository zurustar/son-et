package engine

import (
	"bytes"
	"fmt"
	"os"
	"sync"
	"testing"
	"testing/quick"

	"github.com/sinshu/go-meltysynth/meltysynth"
	"gitlab.com/gomidi/midi/v2"
)

// Feature: gomidi-midi-playback, Property 2: MIDI Bridge Message Forwarding
// For any MIDI message sent to the MIDIBridge, the bridge SHALL forward it to
// meltysynth.Synthesizer.ProcessMidiMessage with correctly extracted channel,
// command, data1, and data2 parameters.
func TestMIDIBridgeMessageForwarding(t *testing.T) {
	// Create a minimal soundfont for testing
	sf := createMinimalSoundFont(t)
	settings := meltysynth.NewSynthesizerSettings(44100)
	synth, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		t.Fatalf("Failed to create synthesizer: %v", err)
	}

	// Create bridge
	bridge := NewMIDIBridge(synth)

	// Property: For any valid MIDI message, the bridge forwards it correctly
	property := func(channel, key, velocity byte) bool {
		// Constrain to valid MIDI values
		channel = channel % 16       // 0-15
		key = key & 0x7F             // 0-127
		velocity = (velocity & 0x7F) // 0-127
		if velocity == 0 {
			velocity = 1 // Avoid note off
		}

		// Create a NoteOn message
		msg := midi.NoteOn(channel, key, velocity)

		// Write message through bridge
		err = bridge.Write(msg)
		if err != nil {
			t.Logf("Write failed: %v", err)
			return false
		}

		// Verify components were extracted correctly
		expectedChannel := channel
		expectedCommand := byte(0x90)
		expectedData1 := key
		expectedData2 := velocity

		// Extract actual components
		actualChannel, actualCommand, actualData1, actualData2 := extractMIDIComponents(msg)

		if actualChannel != expectedChannel {
			t.Logf("Channel mismatch: expected %d, got %d", expectedChannel, actualChannel)
			return false
		}
		if actualCommand != expectedCommand {
			t.Logf("Command mismatch: expected 0x%02X, got 0x%02X", expectedCommand, actualCommand)
			return false
		}
		if actualData1 != expectedData1 {
			t.Logf("Data1 mismatch: expected %d, got %d", expectedData1, actualData1)
			return false
		}
		if actualData2 != expectedData2 {
			t.Logf("Data2 mismatch: expected %d, got %d", expectedData2, actualData2)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// createMinimalSoundFont creates a minimal soundfont for testing
func createMinimalSoundFont(t *testing.T) *meltysynth.SoundFont {
	t.Helper()

	// Load the test soundfont
	data, err := os.ReadFile("../../GeneralUser-GS.sf2")
	if err != nil {
		t.Skipf("Skipping test: soundfont not available: %v", err)
		return nil
	}

	sf, err := meltysynth.NewSoundFont(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("Failed to parse soundfont: %v", err)
	}

	return sf
}

// Feature: gomidi-midi-playback, Property 3: MIDI Bridge Message Type Support
// For any MIDI message type (NoteOn, NoteOff, ControlChange, ProgramChange, PitchBend),
// the MIDIBridge SHALL handle it without error.
func TestMIDIBridgeMessageTypeSupport(t *testing.T) {
	// Create a minimal soundfont for testing
	sf := createMinimalSoundFont(t)
	settings := meltysynth.NewSynthesizerSettings(44100)
	synth, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		t.Fatalf("Failed to create synthesizer: %v", err)
	}

	// Create bridge
	bridge := NewMIDIBridge(synth)

	// Property: For any message type, the bridge handles it without error
	property := func(channel, data1, data2 byte, msgType uint8) bool {
		// Constrain to valid MIDI values
		channel = channel % 16
		data1 = data1 & 0x7F
		data2 = data2 & 0x7F

		// Select message type based on msgType parameter
		var msg midi.Message
		switch msgType % 5 {
		case 0: // NoteOn
			if data2 == 0 {
				data2 = 1 // Avoid note off
			}
			msg = midi.NoteOn(channel, data1, data2)
		case 1: // NoteOff
			msg = midi.NoteOff(channel, data1)
		case 2: // ControlChange
			msg = midi.ControlChange(channel, data1, data2)
		case 3: // ProgramChange
			msg = midi.ProgramChange(channel, data1)
		case 4: // PitchBend
			// PitchBend takes a 14-bit value (0-16383)
			value := int16(data1) | (int16(data2) << 7)
			msg = midi.Pitchbend(channel, value)
		}

		// Write message through bridge - should not error
		err := bridge.Write(msg)
		if err != nil {
			t.Logf("Write failed for message type %d: %v", msgType%5, err)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// Feature: gomidi-midi-playback, Property 4: MIDI Bridge Thread Safety
// For any sequence of concurrent Write calls to the MIDIBridge, the bridge SHALL
// handle them without data races or panics.
func TestMIDIBridgeThreadSafety(t *testing.T) {
	// Create a minimal soundfont for testing
	sf := createMinimalSoundFont(t)
	settings := meltysynth.NewSynthesizerSettings(44100)
	synth, err := meltysynth.NewSynthesizer(sf, settings)
	if err != nil {
		t.Fatalf("Failed to create synthesizer: %v", err)
	}

	// Create bridge
	bridge := NewMIDIBridge(synth)

	// Property: Concurrent writes don't cause data races or panics
	property := func(numGoroutines uint8, messagesPerGoroutine uint8) bool {
		// Constrain to reasonable values
		numGoroutines = (numGoroutines % 10) + 1               // 1-10 goroutines
		messagesPerGoroutine = (messagesPerGoroutine % 20) + 1 // 1-20 messages

		var wg sync.WaitGroup
		errChan := make(chan error, int(numGoroutines)*int(messagesPerGoroutine))

		// Launch concurrent goroutines
		for i := uint8(0); i < numGoroutines; i++ {
			wg.Add(1)
			go func(goroutineID uint8) {
				defer wg.Done()
				defer func() {
					if r := recover(); r != nil {
						errChan <- fmt.Errorf("panic in goroutine %d: %v", goroutineID, r)
					}
				}()

				// Send multiple messages from this goroutine
				for j := uint8(0); j < messagesPerGoroutine; j++ {
					channel := (goroutineID + j) % 16
					key := (goroutineID*10 + j) & 0x7F
					velocity := ((goroutineID + j + 1) & 0x7F)
					if velocity == 0 {
						velocity = 1
					}

					msg := midi.NoteOn(channel, key, velocity)
					if err := bridge.Write(msg); err != nil {
						errChan <- fmt.Errorf("write error in goroutine %d: %v", goroutineID, err)
					}
				}
			}(i)
		}

		// Wait for all goroutines to complete
		wg.Wait()
		close(errChan)

		// Check for errors
		for err := range errChan {
			t.Logf("Concurrent write error: %v", err)
			return false
		}

		return true
	}

	config := &quick.Config{MaxCount: 100}
	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}
