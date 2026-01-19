package engine

import (
	"bytes"
	"encoding/binary"
	"testing"
)

// createTestMIDI creates a minimal MIDI file with specified tempo values
func createTestMIDI(tempoValues []int) []byte {
	buf := new(bytes.Buffer)

	// Write MIDI header (MThd)
	buf.WriteString("MThd")
	binary.Write(buf, binary.BigEndian, uint32(6))   // Header length
	binary.Write(buf, binary.BigEndian, uint16(0))   // Format 0
	binary.Write(buf, binary.BigEndian, uint16(1))   // 1 track
	binary.Write(buf, binary.BigEndian, uint16(480)) // PPQ = 480

	// Write track header (MTrk)
	trackBuf := new(bytes.Buffer)

	// Write tempo events
	for i, tempo := range tempoValues {
		// Delta time (0 for first event, 100 for subsequent)
		if i == 0 {
			trackBuf.WriteByte(0x00) // Delta time = 0
		} else {
			trackBuf.WriteByte(0x64) // Delta time = 100
		}

		// Meta event: Set Tempo (FF 51 03)
		trackBuf.WriteByte(0xFF) // Meta event
		trackBuf.WriteByte(0x51) // Set Tempo
		trackBuf.WriteByte(0x03) // Length = 3 bytes

		// Write tempo value (microseconds per beat, 3 bytes big-endian)
		trackBuf.WriteByte(byte(tempo >> 16))
		trackBuf.WriteByte(byte(tempo >> 8))
		trackBuf.WriteByte(byte(tempo))
	}

	// End of track
	trackBuf.WriteByte(0x00) // Delta time
	trackBuf.WriteByte(0xFF) // Meta event
	trackBuf.WriteByte(0x2F) // End of Track
	trackBuf.WriteByte(0x00) // Length = 0

	// Write track chunk
	buf.WriteString("MTrk")
	binary.Write(buf, binary.BigEndian, uint32(trackBuf.Len()))
	buf.Write(trackBuf.Bytes())

	return buf.Bytes()
}

func TestParseMidiTempo_ValidTempo(t *testing.T) {
	// Test with valid tempo (120 BPM = 500000 microseconds per beat)
	midiData := createTestMIDI([]int{500000})

	events, ppq, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if ppq != 480 {
		t.Errorf("Expected PPQ=480, got %d", ppq)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 tempo events (default + explicit), got %d", len(events))
	}

	// Check the explicit tempo event
	if events[1].MicrosPerBeat != 500000 {
		t.Errorf("Expected tempo 500000, got %d", events[1].MicrosPerBeat)
	}
}

func TestParseMidiTempo_InvalidTempoZero(t *testing.T) {
	// Test with invalid tempo (0 microseconds per beat)
	midiData := createTestMIDI([]int{0})

	events, ppq, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if ppq != 480 {
		t.Errorf("Expected PPQ=480, got %d", ppq)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 tempo events (default + corrected), got %d", len(events))
	}

	// Check that invalid tempo was replaced with default 120 BPM
	if events[1].MicrosPerBeat != 500000 {
		t.Errorf("Expected invalid tempo to be replaced with 500000 (120 BPM), got %d", events[1].MicrosPerBeat)
	}
}

func TestParseMidiTempo_InvalidTempoTooFast(t *testing.T) {
	// Test with invalid tempo (too fast: 100000 microseconds per beat = 600 BPM)
	midiData := createTestMIDI([]int{100000})

	events, _, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 tempo events (default + corrected), got %d", len(events))
	}

	// Check that invalid tempo was replaced with default 120 BPM
	if events[1].MicrosPerBeat != 500000 {
		t.Errorf("Expected invalid tempo to be replaced with 500000 (120 BPM), got %d", events[1].MicrosPerBeat)
	}
}

func TestParseMidiTempo_InvalidTempoTooSlow(t *testing.T) {
	// Test with invalid tempo (too slow: 5000000 microseconds per beat = 12 BPM)
	midiData := createTestMIDI([]int{5000000})

	events, _, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 tempo events (default + corrected), got %d", len(events))
	}

	// Check that invalid tempo was replaced with default 120 BPM
	if events[1].MicrosPerBeat != 500000 {
		t.Errorf("Expected invalid tempo to be replaced with 500000 (120 BPM), got %d", events[1].MicrosPerBeat)
	}
}

func TestParseMidiTempo_MultipleTempos(t *testing.T) {
	// Test with multiple tempo changes, including one invalid
	midiData := createTestMIDI([]int{500000, 0, 600000})

	events, ppq, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if ppq != 480 {
		t.Errorf("Expected PPQ=480, got %d", ppq)
	}

	if len(events) != 4 {
		t.Fatalf("Expected 4 tempo events (default + 3 explicit), got %d", len(events))
	}

	// Check first explicit tempo (valid)
	if events[1].MicrosPerBeat != 500000 {
		t.Errorf("Expected first tempo 500000, got %d", events[1].MicrosPerBeat)
	}

	// Check second explicit tempo (invalid, should be replaced)
	if events[2].MicrosPerBeat != 500000 {
		t.Errorf("Expected invalid tempo to be replaced with 500000, got %d", events[2].MicrosPerBeat)
	}

	// Check third explicit tempo (valid)
	if events[3].MicrosPerBeat != 600000 {
		t.Errorf("Expected third tempo 600000, got %d", events[3].MicrosPerBeat)
	}
}

func TestParseMidiTempo_EdgeCaseMinValidTempo(t *testing.T) {
	// Test with minimum valid tempo (300 BPM = 200000 microseconds per beat)
	midiData := createTestMIDI([]int{200000})

	events, _, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 tempo events, got %d", len(events))
	}

	// Check that valid tempo is preserved
	if events[1].MicrosPerBeat != 200000 {
		t.Errorf("Expected tempo 200000 to be preserved, got %d", events[1].MicrosPerBeat)
	}
}

func TestParseMidiTempo_EdgeCaseMaxValidTempo(t *testing.T) {
	// Test with maximum valid tempo (20 BPM = 3000000 microseconds per beat)
	midiData := createTestMIDI([]int{3000000})

	events, _, err := parseMidiTempo(midiData)
	if err != nil {
		t.Fatalf("parseMidiTempo failed: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("Expected 2 tempo events, got %d", len(events))
	}

	// Check that valid tempo is preserved
	if events[1].MicrosPerBeat != 3000000 {
		t.Errorf("Expected tempo 3000000 to be preserved, got %d", events[1].MicrosPerBeat)
	}
}
