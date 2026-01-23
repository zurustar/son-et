package engine

import (
	"bytes"
	"testing"

	"gitlab.com/gomidi/midi/v2/smf"
)

// createTestMIDIFileWithTempo creates a minimal MIDI file with specified tempo events
func createTestMIDIFileWithTempo(tempos []struct {
	tick int
	bpm  float64
}) []byte {
	var buf bytes.Buffer

	// Write MThd header
	buf.Write([]byte("MThd"))
	buf.Write([]byte{0x00, 0x00, 0x00, 0x06}) // Header length: 6 bytes
	buf.Write([]byte{0x00, 0x00})             // Format: 0 (single track)
	buf.Write([]byte{0x00, 0x01})             // Number of tracks: 1
	buf.Write([]byte{0x01, 0xE0})             // Time division: 480 PPQ

	// Build track data
	var trackData bytes.Buffer

	lastTick := 0
	for _, tempo := range tempos {
		// Calculate delta time
		delta := tempo.tick - lastTick
		trackData.Write(encodeVarInt(delta))

		// Write tempo meta event
		microsPerBeat := int(60000000 / tempo.bpm)
		trackData.Write([]byte{0xFF, 0x51, 0x03}) // Meta event: Set Tempo
		trackData.Write([]byte{
			byte(microsPerBeat >> 16),
			byte(microsPerBeat >> 8),
			byte(microsPerBeat),
		})

		lastTick = tempo.tick
	}

	// Add a note to make it a valid MIDI file
	trackData.Write([]byte{0x00})             // Delta time: 0
	trackData.Write([]byte{0x90, 0x3C, 0x40}) // Note On
	trackData.Write([]byte{0x10})             // Delta time: 16
	trackData.Write([]byte{0x80, 0x3C, 0x00}) // Note Off

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

// encodeVarInt encodes an integer as a variable-length quantity
func encodeVarInt(value int) []byte {
	if value == 0 {
		return []byte{0}
	}

	var result []byte
	for value > 0 {
		b := byte(value & 0x7F)
		value >>= 7
		if len(result) > 0 {
			b |= 0x80
		}
		result = append([]byte{b}, result...)
	}
	return result
}

// TestExtractTempoMap_DefaultTempo verifies that a MIDI file with no tempo events
// gets the default 120 BPM tempo (500000 microseconds per beat).
func TestExtractTempoMap_DefaultTempo(t *testing.T) {
	// Create a simple MIDI file with no tempo events
	midiData := createTestMIDIFileWithTempo(nil)

	// Parse the MIDI file
	smfData, err := smf.ReadFrom(bytes.NewReader(midiData))
	if err != nil {
		t.Fatalf("Failed to parse MIDI file: %v", err)
	}

	// Extract tempo map
	tempoMap, err := extractTempoMap(smfData, 480)
	if err != nil {
		t.Fatalf("Failed to extract tempo map: %v", err)
	}

	// Verify default tempo is present
	if len(tempoMap) != 1 {
		t.Errorf("Expected 1 tempo event (default), got %d", len(tempoMap))
	}

	if tempoMap[0].Tick != 0 {
		t.Errorf("Expected default tempo at tick 0, got tick %d", tempoMap[0].Tick)
	}

	if tempoMap[0].MicrosPerBeat != 500000 {
		t.Errorf("Expected default tempo 500000 µs/beat (120 BPM), got %d", tempoMap[0].MicrosPerBeat)
	}
}

// TestExtractTempoMap_SingleTempoChange verifies that a MIDI file with one tempo change
// is correctly extracted.
func TestExtractTempoMap_SingleTempoChange(t *testing.T) {
	// Create a MIDI file with a tempo change at tick 0
	tempos := []struct {
		tick int
		bpm  float64
	}{
		{0, 140.0},
	}
	midiData := createTestMIDIFileWithTempo(tempos)

	// Parse the MIDI file
	smfData, err := smf.ReadFrom(bytes.NewReader(midiData))
	if err != nil {
		t.Fatalf("Failed to parse MIDI file: %v", err)
	}

	// Extract tempo map
	tempoMap, err := extractTempoMap(smfData, 480)
	if err != nil {
		t.Fatalf("Failed to extract tempo map: %v", err)
	}

	// Verify we have default tempo + 1 tempo change
	if len(tempoMap) != 2 {
		t.Errorf("Expected 2 tempo events (default + 1 change), got %d", len(tempoMap))
	}

	// Verify default tempo
	if tempoMap[0].Tick != 0 || tempoMap[0].MicrosPerBeat != 500000 {
		t.Errorf("Expected default tempo at tick 0 with 500000 µs/beat, got tick %d with %d µs/beat",
			tempoMap[0].Tick, tempoMap[0].MicrosPerBeat)
	}

	// Verify tempo change (140 BPM = 428571 microseconds per beat)
	expectedMicros := 428571
	if tempoMap[1].Tick != 0 {
		t.Errorf("Expected tempo change at tick 0, got tick %d", tempoMap[1].Tick)
	}
	if tempoMap[1].MicrosPerBeat != expectedMicros {
		t.Errorf("Expected tempo 140 BPM (%d µs/beat), got %d µs/beat",
			expectedMicros, tempoMap[1].MicrosPerBeat)
	}
}

// TestExtractTempoMap_MultipleTempoChanges verifies that a MIDI file with multiple
// tempo changes is correctly extracted.
func TestExtractTempoMap_MultipleTempoChanges(t *testing.T) {
	// Create a MIDI file with multiple tempo changes
	tempos := []struct {
		tick int
		bpm  float64
	}{
		{0, 120.0},
		{480, 140.0},
		{960, 100.0},
	}
	midiData := createTestMIDIFileWithTempo(tempos)

	// Parse the MIDI file
	smfData, err := smf.ReadFrom(bytes.NewReader(midiData))
	if err != nil {
		t.Fatalf("Failed to parse MIDI file: %v", err)
	}

	// Extract tempo map
	tempoMap, err := extractTempoMap(smfData, 480)
	if err != nil {
		t.Fatalf("Failed to extract tempo map: %v", err)
	}

	// Verify we have default tempo + 3 tempo changes
	if len(tempoMap) != 4 {
		t.Errorf("Expected 4 tempo events (default + 3 changes), got %d", len(tempoMap))
	}

	// Verify tempo events
	expectedTempos := []struct {
		tick   int
		micros int
	}{
		{0, 500000},   // Default 120 BPM
		{0, 500000},   // 120 BPM
		{480, 428571}, // 140 BPM
		{960, 600000}, // 100 BPM
	}

	for i, expected := range expectedTempos {
		if tempoMap[i].Tick != expected.tick {
			t.Errorf("Tempo event %d: expected tick %d, got %d", i, expected.tick, tempoMap[i].Tick)
		}

		if tempoMap[i].MicrosPerBeat != expected.micros {
			t.Errorf("Tempo event %d: expected %d µs/beat, got %d µs/beat",
				i, expected.micros, tempoMap[i].MicrosPerBeat)
		}
	}
}

// TestWallClockTickGenerator_WithTempoMap verifies that the tick generator
// correctly calculates ticks with a tempo map.
func TestWallClockTickGenerator_WithTempoMap(t *testing.T) {
	// Create a tempo map with two tempo changes
	// 0-480 ticks: 120 BPM (500000 µs/beat)
	// 480+ ticks: 140 BPM (428571 µs/beat)
	tempoMap := []TempoEvent{
		{Tick: 0, MicrosPerBeat: 500000},   // 120 BPM
		{Tick: 480, MicrosPerBeat: 428571}, // 140 BPM
	}

	ppq := 480
	tg := NewWallClockTickGenerator(44100, ppq, tempoMap)

	// Test 1: At 0.5 seconds at 120 BPM
	// 120 BPM = 2 beats per second
	// 0.5 seconds = 1 beat = 480 ticks (at PPQ=480)
	elapsed := 0.5
	tick := tg.CalculateTickFromTime(elapsed)
	expectedTick := 480
	if tick != expectedTick {
		t.Errorf("At %.2fs: expected tick %d, got %d", elapsed, expectedTick, tick)
	}

	// Test 2: At 1.0 seconds (2 beats at 120 BPM), we should be at tick 960
	// But the tempo changes at tick 480 (0.5s), so:
	// 0-0.5s: 480 ticks at 120 BPM
	// 0.5-1.0s: 0.5s at 140 BPM = 0.5 / (428571/1000000) = 1.167 beats = 560 ticks
	// Total: 480 + 560 = 1040 ticks
	elapsed = 1.0
	tick = tg.CalculateTickFromTime(elapsed)
	expectedTick = 1040
	tolerance := 5 // Allow small rounding error
	if tick < expectedTick-tolerance || tick > expectedTick+tolerance {
		t.Errorf("At %.2fs: expected tick ~%d, got %d", elapsed, expectedTick, tick)
	}

	// Test 3: At 1.5 seconds
	// 0-0.5s: 480 ticks at 120 BPM
	// 0.5-1.5s: 1.0s at 140 BPM = 1.0 / (428571/1000000) = 2.333 beats = 1120 ticks
	// Total: 480 + 1120 = 1600 ticks
	elapsed = 1.5
	tick = tg.CalculateTickFromTime(elapsed)
	expectedTick = 1600
	tolerance = 5 // Allow small rounding error
	if tick < expectedTick-tolerance || tick > expectedTick+tolerance {
		t.Errorf("At %.2fs: expected tick ~%d, got %d", elapsed, expectedTick, tick)
	}
}
