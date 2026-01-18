package engine

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

var (
	audioContext *audio.Context
	soundFont    *meltysynth.SoundFont
	midiPlayer   *audio.Player
	sequencer    *meltysynth.MidiFileSequencer
	synthesizer  *meltysynth.Synthesizer
	midiFinished bool // Flag to track if MIDI has finished playing
)

const (
	sampleRate = 44100
)

// InitializeAudio sets up the Ebiten audio context
func InitializeAudio() {
	if audioContext == nil {
		audioContext = audio.NewContext(sampleRate)
	}
}

// findAsset attempts to find an asset file case-insensitively in the embedded FS
func findAsset(name string) ([]byte, error) {
	// 1. Try exact match
	data, err := assets.ReadFile(name)
	if err == nil {
		return data, nil
	}

	// 2. Try case-insensitive scan in root
	entries, err := assets.ReadDir(".")
	if err != nil {
		return nil, err
	}

	lowerName := strings.ToLower(name)
	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == lowerName {
			return assets.ReadFile(entry.Name())
		}
	}

	return nil, fmt.Errorf("asset not found: %s", name)
}

// LoadSoundFont loads a .sf2 file for MIDI playback
func LoadSoundFont(path string) error {
	// Try loading from explicit path first
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		sf, err := meltysynth.NewSoundFont(f)
		if err != nil {
			return fmt.Errorf("failed to parse soundfont file: %v", err)
		}
		soundFont = sf
		fmt.Printf("SoundFont loaded: %s\n", path)
		return nil
	}

	// Try loading from assets
	data, errAsset := findAsset(path)
	if errAsset != nil {
		return fmt.Errorf("failed to open soundfont %s: %v", path, errAsset)
	}

	// Open from bytes
	sf, err := meltysynth.NewSoundFont(strings.NewReader(string(data)))
	if err != nil {
		return fmt.Errorf("failed to parse soundfont: %v", err)
	}
	soundFont = sf
	fmt.Printf("SoundFont loaded from assets: %s\n", path)
	return nil
}

// PlayMidiFile loads and plays a MIDI file
func PlayMidiFile(path string) {
	if soundFont == nil {
		fmt.Println("PlayMIDI Error: No SoundFont loaded. Please provide 'default.sf2'.")
		return
	}

	fmt.Printf("PlayMidiFile: Loading %s\n", path)

	// Load MIDI file
	var midiFile *meltysynth.MidiFile
	var err error

	// Read MIDI data
	var midiFileBytes []byte
	var errBytes error

	// Try file system first
	midiFileBytes, errBytes = os.ReadFile(path)
	if errBytes != nil {
		// Try AssetLoader if available
		if globalEngine != nil && globalEngine.assetLoader != nil {
			midiFileBytes, errBytes = globalEngine.assetLoader.ReadFile(path)
		} else {
			// Fallback to findAsset for embedded mode
			midiFileBytes, errBytes = findAsset(path)
		}

		if errBytes != nil {
			fmt.Printf("PlayMIDI Error: Could not find or open %s: %v\n", path, errBytes)
			return
		}
	}

	// Parse for Meltysynth
	midiFile, err = meltysynth.NewMidiFile(strings.NewReader(string(midiFileBytes)))

	if err != nil {
		fmt.Printf("PlayMIDI Error: Failed to parse MIDI: %v\n", err)
		return
	}

	// Create settings
	settings := meltysynth.NewSynthesizerSettings(sampleRate)
	synthesizer, err = meltysynth.NewSynthesizer(soundFont, settings)
	if err != nil {
		fmt.Printf("PlayMIDI Error: Failed to create synthesizer: %v\n", err)
		return
	}

	sequencer = meltysynth.NewMidiFileSequencer(synthesizer)
	sequencer.Play(midiFile, false) // loop=false (User requested one-shot)

	// Parse Tempo Map and PPQ
	tempoMap, ppq, err := parseMidiTempo(midiFileBytes)
	if err != nil {
		fmt.Printf("Warning: Parser error: %v. Using defaults.\n", err)
	}

	// Calculate initial BPM for display
	initBPM := 120.0
	if len(tempoMap) > 0 {
		initBPM = 60000000.0 / float64(tempoMap[0].MicrosPerBeat)
	}
	fmt.Printf("PlayMIDI: Sync info - InitBPM=%.2f, PPQ=%d, Events=%d\n", initBPM, ppq, len(tempoMap))

	// Create TickGenerator
	tickGen, err := NewTickGenerator(sampleRate, ppq, tempoMap)
	if err != nil {
		fmt.Printf("PlayMIDI Error: Failed to create tick generator: %v\n", err)
		return
	}

	// Calculate total ticks in MIDI file
	totalTicks := calculateMidiLength(midiFileBytes, ppq)
	fmt.Printf("PlayMIDI: Total ticks in MIDI file: %d\n", totalTicks)

	// Create a stream that reads from the synthesizer
	stream := &MidiStream{
		sequencer:     sequencer,
		tickGenerator: tickGen,
		startTime:     time.Now(),
		totalTicks:    totalTicks,
		endReported:   false,
	}

	// Create Ebiten player
	if midiPlayer != nil {
		midiPlayer.Close()
	}
	midiPlayer, err = audioContext.NewPlayer(stream)
	if err != nil {
		fmt.Printf("PlayMIDI Error: Failed to create audio player: %v\n", err)
		return
	}
	fmt.Printf("PlayMIDI: Audio player created successfully\n")

	// Set volume to 0 in headless mode
	if headlessMode {
		midiPlayer.SetVolume(0)
		fmt.Println("PlayMIDI: Audio muted (headless mode)")
	}

	// Reset MIDI finished flag
	midiFinished = false

	// Start playback in a goroutine to avoid blocking the VM execution
	// This allows the MIDI player to run concurrently with the game loop
	go func() {
		midiPlayer.Play()
		fmt.Println("PlayMIDI: Playback started")
		fmt.Printf("PlayMIDI: IsPlaying=%v\n", midiPlayer.IsPlaying())
	}()

	// Start the Conductor (legacy - now using TickGenerator)
	StartConductor(tempoMap, ppq)
}

// Global Conductor Ticker
var (
	// conductorStop chan bool // Removed in favor of Audio Sync
	// tickChannel   chan int

	// Audio Sync Globals
	globalTempoMap []TempoEvent
	globalPPQ      int = 480
	currentSamples int64
)

func StartConductor(tempoMap []TempoEvent, ppq int) {
	// Setup globals for MidiStream to access
	globalTempoMap = tempoMap
	globalPPQ = ppq
	currentSamples = 0

	// Update Engine Global
	GlobalPPQ = ppq

	// We don't start a goroutine anymore.
	// Ticks are driven by MidiStream.Read()
}

// StartConductor sets up the sync engine.
// Ticks are now generated by the audio callbacks in MidiStream.Read
func StartConductorStub() {
	// Legacy placeholder if needed
}

func parseMidiTempo(data []byte) ([]TempoEvent, int, error) {
	// Full MIDI parser to find Tempo Map
	if len(data) < 14 {
		return nil, 480, fmt.Errorf("data too short")
	}

	// Check Header chunk (MThd)
	if string(data[0:4]) != "MThd" {
		return nil, 480, fmt.Errorf("invalid midi header")
	}

	// Time Division is at offset 12 (2 bytes)
	td := int(data[12])<<8 | int(data[13])
	ppq := 480
	if td&0x8000 == 0 {
		ppq = td
	}

	// Default Tempo
	events := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM

	// Scan tracks for Set Tempo
	offset := 14
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}
		// Read Chunk Type
		if string(data[offset:offset+4]) == "MTrk" {
			chunkLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
			offset += 8
			trackEnd := offset + chunkLen
			if trackEnd > len(data) {
				trackEnd = len(data)
			}

			curr := offset
			currentTick := 0
			for curr < trackEnd {
				delta, n := readVarInt(data[curr:])
				curr += n
				currentTick += delta

				if curr >= trackEnd {
					break
				}

				status := data[curr]
				// Running status?
				if status < 0x80 {
					// Running status not supported in this simple scanner, assume properly formed SMF for meta
					// Actually Meta events always start with FF, which is > 0x80.
					// So we just scan for FF.
					// But we must skip data bytes of other events to find next delta.
					// This is hard without full parser.
					// LUCKILY: Standard MIDI Files usually put Tempo Map in Track 0 alone.
					// We will try our best.
				}

				curr++

				if status == 0xFF { // Meta Event
					typeByte := data[curr]
					curr++
					lenByte, n := readVarInt(data[curr:])
					curr += n

					if typeByte == 0x51 && lenByte == 3 {
						usPerBeat := int(data[curr])<<16 | int(data[curr+1])<<8 | int(data[curr+2])
						events = append(events, TempoEvent{Tick: currentTick, MicrosPerBeat: usPerBeat})
					}
					curr += int(lenByte)
				} else if status == 0xF0 || status == 0xF7 {
					// SysEx
					lenByte, n := readVarInt(data[curr:])
					curr += n + int(lenByte)
				} else {
					// Channel message.
					// 1 data byte: C0-DF
					// 2 data bytes: 80-BF, E0-EF
					if status >= 0x80 && status <= 0xEF {
						if (status&0xE0) == 0xC0 || (status&0xE0) == 0xD0 {
							curr += 1
						} else {
							curr += 2
						}
					} else {
						// Running status case... tough.
						// Assuming running status applies to last status.
						// Ideally we use a library for this.
						// But let's hope finding FF 51 03 is enough.
					}
				}
			}
			offset = trackEnd
		} else {
			break // Unknown chunk
		}
	}

	// Sort events by Tick just in case
	// (Assuming linear scan of Track 0 produced sorted)

	return events, ppq, nil
}

// calculateMidiLength calculates the total length of a MIDI file in ticks
// by finding the last event in all tracks
func calculateMidiLength(data []byte, ppq int) int64 {
	if len(data) < 14 {
		return 0
	}

	// Skip header
	offset := 14

	maxTick := int64(0)

	// Parse all tracks
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		chunkType := string(data[offset : offset+4])
		chunkLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
		offset += 8

		if chunkType == "MTrk" {
			trackData := data[offset : offset+chunkLen]
			trackOffset := 0
			currentTick := int64(0)

			// Parse all events in this track
			for trackOffset < len(trackData) {
				// Read delta time
				deltaTime, consumed := readVarInt(trackData[trackOffset:])
				if consumed == 0 {
					break
				}
				trackOffset += consumed
				currentTick += int64(deltaTime)

				if trackOffset >= len(trackData) {
					break
				}

				// Skip event data (we only care about timing)
				eventByte := trackData[trackOffset]
				trackOffset++

				if eventByte == 0xFF {
					// Meta event
					if trackOffset >= len(trackData) {
						break
					}
					trackOffset++ // Skip meta type
					length, consumed := readVarInt(trackData[trackOffset:])
					trackOffset += consumed + length
				} else if eventByte == 0xF0 || eventByte == 0xF7 {
					// SysEx event
					length, consumed := readVarInt(trackData[trackOffset:])
					trackOffset += consumed + length
				} else if eventByte >= 0x80 {
					// MIDI event
					if eventByte >= 0xC0 && eventByte < 0xE0 {
						// Program change or channel pressure (1 data byte)
						trackOffset++
					} else {
						// Other events (2 data bytes)
						trackOffset += 2
					}
				}
			}

			// Update max tick
			if currentTick > maxTick {
				maxTick = currentTick
			}

			offset += chunkLen
		} else {
			offset += chunkLen
		}
	}

	return maxTick
}

func readVarInt(data []byte) (int, int) {
	if len(data) == 0 {
		return 0, 0
	}
	val := 0
	for i, b := range data {
		val = (val << 7) | int(b&0x7F)
		if b&0x80 == 0 {
			return val, i + 1
		}
		if i >= 3 {
			break
		} // Max 4 bytes
	}
	return 0, 0
}

// MidiStream implements io.Reader to pipe synthesizer output to Ebiten
type MidiStream struct {
	sequencer     *meltysynth.MidiFileSequencer
	leftBuf       []float32
	rightBuf      []float32
	tickGenerator *TickGenerator
	startTime     time.Time
	totalTicks    int64 // Total ticks in MIDI file
	endReported   bool  // Flag to ensure MIDI_END is triggered only once
}

func (s *MidiStream) Read(p []byte) (n int, err error) {
	// Ebiten requests bytes. Format is usually 16bit Little Endian Stereo.
	// 4 bytes per sample (2 channels * 2 bytes).
	numSamples := len(p) / 4

	if len(s.leftBuf) < numSamples {
		s.leftBuf = make([]float32, numSamples)
		s.rightBuf = make([]float32, numSamples)
	}

	// Render samples
	s.sequencer.Render(s.leftBuf[:numSamples], s.rightBuf[:numSamples])

	// Update Clock using wall-clock time for maximum accuracy
	// This avoids cumulative drift from sample-based calculation
	if s.tickGenerator != nil {
		elapsed := time.Since(s.startTime).Seconds()

		// Calculate current tick from elapsed time using tempo map
		// This properly accounts for tempo changes
		currentTick := s.tickGenerator.CalculateTickFromTime(elapsed)

		// Check if we've reached the end of the MIDI file
		if int64(currentTick) >= s.totalTicks && !s.endReported {
			fmt.Println("MidiStream.Read: MIDI playback ended (reached totalTicks), triggering MIDI_END event")
			s.endReported = true
			midiFinished = true // Set global flag
			// Trigger MIDI_END event if handler is registered
			if midiEndHandler != nil && !midiEndTriggered {
				TriggerMidiEnd()
			}
			// Stop sending ticks
			return len(p), nil
		}

		// Notify all ticks from lastDeliveredTick+1 to currentTick
		// This ensures we don't skip any ticks even if processing is delayed
		// But don't exceed totalTicks
		endTick := int64(currentTick)
		if endTick > s.totalTicks {
			endTick = s.totalTicks
		}

		for tick := s.tickGenerator.lastDeliveredTick + 1; tick <= int(endTick); tick++ {
			NotifyTick(tick)
		}

		if currentTick > s.tickGenerator.lastDeliveredTick {
			s.tickGenerator.lastDeliveredTick = currentTick
		}
	}

	// Convert float32 [-1, 1] to int16 bytes
	for i := 0; i < numSamples; i++ {
		// Left
		valL := int16(s.leftBuf[i] * 32767)
		p[4*i] = byte(valL)
		p[4*i+1] = byte(valL >> 8)

		// Right
		valR := int16(s.rightBuf[i] * 32767)
		p[4*i+2] = byte(valR)
		p[4*i+3] = byte(valR >> 8)
	}

	return len(p), nil
}

// GetCurrentTick returns the current MIDI tick for synchronization
func GetCurrentTick() int {
	// TODO: Expose tick from sequencer
	return 0
}

// PlayWAVE plays a WAV file
func PlayWAVE(path string) {
	fmt.Printf("PlayWAVE: %s\n", path)

	// Find asset using AssetLoader
	var data []byte
	var err error

	if globalEngine != nil && globalEngine.assetLoader != nil {
		data, err = globalEngine.assetLoader.ReadFile(path)
	} else {
		// Fallback to findAsset for embedded mode
		data, err = findAsset(path)
	}

	if err != nil {
		fmt.Printf("PlayWAVE Error: File not found %s: %v\n", path, err)
		return
	}

	// Create context if needed
	InitializeAudio()

	// Decode WAV
	// Note: Ebiten's wav package is needed
	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		fmt.Printf("PlayWAVE Error: Failed to decode WAV %s: %v\n", path, err)
		return
	}

	// Create Player
	player, err := audioContext.NewPlayer(stream)
	if err != nil {
		fmt.Printf("PlayWAVE Error: Failed to create player for %s: %v\n", path, err)
		return
	}

	// Set volume to 0 in headless mode
	if headlessMode {
		player.SetVolume(0)
		fmt.Println("PlayWAVE: Audio muted (headless mode)")
	}

	player.Play()
}
