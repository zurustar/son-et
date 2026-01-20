package engine

import (
	"testing"
)

func TestNewMIDIPlayer(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	if engine.midiPlayer == nil {
		t.Fatal("MIDI player not created")
	}

	if engine.midiPlayer.audioContext == nil {
		t.Error("Audio context not initialized")
	}
}

func TestLoadSoundFont_FileNotFound(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	err := engine.LoadSoundFont("nonexistent.sf2")
	if err == nil {
		t.Error("Expected error for nonexistent soundfont")
	}
}

func TestPlayMIDI_NoSoundFont(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	err := engine.PlayMIDI("test.mid")
	if err == nil {
		t.Error("Expected error when playing MIDI without soundfont")
	}
}

func TestPlayMIDI_FileNotFound(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Load a dummy soundfont (will fail parsing, but that's ok for this test)
	assetLoader.Files["test.sf2"] = []byte("dummy")

	err := engine.PlayMIDI("nonexistent.mid")
	if err == nil {
		t.Error("Expected error for nonexistent MIDI file")
	}
}

func TestStopMIDI(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	// Should not panic even if nothing is playing
	engine.StopMIDI()
}

func TestIsMIDIPlaying_NotPlaying(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	engine := NewEngine(renderer, assetLoader, imageDecoder)

	if engine.IsMIDIPlaying() {
		t.Error("MIDI should not be playing initially")
	}
}

func TestParseMIDITempo_InvalidData(t *testing.T) {
	// Test with too short data
	_, _, err := parseMIDITempo([]byte{0x00, 0x01})
	if err == nil {
		t.Error("Expected error for too short MIDI data")
	}

	// Test with invalid header
	invalidHeader := make([]byte, 14)
	copy(invalidHeader, "XXXX")
	_, _, err = parseMIDITempo(invalidHeader)
	if err == nil {
		t.Error("Expected error for invalid MIDI header")
	}
}

func TestParseMIDITempo_ValidHeader(t *testing.T) {
	// Create a minimal valid MIDI header
	data := make([]byte, 14)
	copy(data[0:4], "MThd") // Header chunk
	data[4] = 0x00          // Length MSB
	data[5] = 0x00          // Length
	data[6] = 0x00          // Length
	data[7] = 0x06          // Length LSB (6 bytes)
	data[8] = 0x00          // Format MSB
	data[9] = 0x00          // Format LSB (0 = single track)
	data[10] = 0x00         // Tracks MSB
	data[11] = 0x01         // Tracks LSB (1 track)
	data[12] = 0x01         // PPQ MSB
	data[13] = 0xE0         // PPQ LSB (480)

	tempoMap, ppq, err := parseMIDITempo(data)
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if ppq != 480 {
		t.Errorf("Expected PPQ 480, got %d", ppq)
	}

	if len(tempoMap) == 0 {
		t.Error("Expected at least one tempo event (default)")
	}

	// Check default tempo (120 BPM = 500000 microseconds per beat)
	if tempoMap[0].MicrosPerBeat != 500000 {
		t.Errorf("Expected default tempo 500000, got %d", tempoMap[0].MicrosPerBeat)
	}
}

func TestReadVarInt(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected int
		bytes    int
	}{
		{"single byte", []byte{0x00}, 0, 1},
		{"single byte max", []byte{0x7F}, 127, 1},
		{"two bytes", []byte{0x81, 0x00}, 128, 2},
		{"two bytes max", []byte{0xFF, 0x7F}, 16383, 2},
		{"three bytes", []byte{0x81, 0x80, 0x00}, 16384, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, bytesRead := readVarInt(tt.data)
			if value != tt.expected {
				t.Errorf("Expected value %d, got %d", tt.expected, value)
			}
			if bytesRead != tt.bytes {
				t.Errorf("Expected %d bytes read, got %d", tt.bytes, bytesRead)
			}
		})
	}
}

func TestCalculateMIDILength(t *testing.T) {
	// Create a minimal MIDI file with one track
	data := make([]byte, 14)
	copy(data[0:4], "MThd")
	data[7] = 0x06
	data[11] = 0x01
	data[12] = 0x01
	data[13] = 0xE0 // PPQ = 480

	// Add a track chunk
	trackData := []byte{
		'M', 'T', 'r', 'k', // Track header
		0x00, 0x00, 0x00, 0x04, // Track length = 4 bytes
		0x00, // Delta time = 0
		0x90, // Note on
		0x00, // Delta time = 0
		0xFF, // Meta event (end of track)
	}
	data = append(data, trackData...)

	length := calculateMIDILength(data, 480)
	if length < 0 {
		t.Errorf("Expected non-negative length, got %d", length)
	}
}
