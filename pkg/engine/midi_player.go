package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

const (
	// Audio sample rate for MIDI playback
	MIDISampleRate = 44100
)

var (
	// Global audio context (Ebiten allows only one)
	globalAudioContext *audio.Context
	audioContextMutex  sync.Mutex
)

// getAudioContext returns the global audio context, creating it if necessary.
func getAudioContext() *audio.Context {
	audioContextMutex.Lock()
	defer audioContextMutex.Unlock()

	if globalAudioContext == nil {
		globalAudioContext = audio.NewContext(MIDISampleRate)
	}
	return globalAudioContext
}

// TempoEvent represents a tempo change in a MIDI file.
type TempoEvent struct {
	Tick          int // MIDI tick when tempo changes
	MicrosPerBeat int // Microseconds per quarter note
}

// MIDIPlayer manages MIDI playback with tick generation.
type MIDIPlayer struct {
	audioContext *audio.Context
	soundFont    *meltysynth.SoundFont
	player       *audio.Player
	stream       *MIDIStream
	mutex        sync.Mutex
	isPlaying    bool
	engine       *Engine
}

// NewMIDIPlayer creates a new MIDI player.
func NewMIDIPlayer(engine *Engine) *MIDIPlayer {
	return &MIDIPlayer{
		audioContext: getAudioContext(),
		engine:       engine,
	}
}

// LoadSoundFont loads a SoundFont (.sf2) file for MIDI synthesis.
func (mp *MIDIPlayer) LoadSoundFont(filename string) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	var data []byte
	var err error

	// Check if filename is an absolute path
	if filepath.IsAbs(filename) {
		// Load directly from filesystem (for auto-loaded SoundFonts)
		data, err = os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to load soundfont %s: %w", filename, err)
		}
	} else {
		// Load via AssetLoader (for project-relative paths)
		data, err = mp.engine.state.assetLoader.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to load soundfont %s: %w", filename, err)
		}
	}

	// Parse SoundFont
	sf, err := meltysynth.NewSoundFont(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse soundfont %s: %w", filename, err)
	}

	mp.soundFont = sf
	mp.engine.logger.LogInfo("Loaded SoundFont: %s", filename)
	return nil
}

// PlayMIDI starts MIDI playback.
// Returns immediately (non-blocking).
func (mp *MIDIPlayer) PlayMIDI(filename string) error {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.soundFont == nil {
		return fmt.Errorf("no soundfont loaded - call LoadSoundFont first")
	}

	// Stop current playback if any
	if mp.player != nil {
		mp.player.Close()
		mp.player = nil
	}

	// Load MIDI file via AssetLoader
	data, err := mp.engine.state.assetLoader.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load MIDI file %s: %w", filename, err)
	}

	// Parse MIDI file
	midiFile, err := meltysynth.NewMidiFile(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse MIDI file %s: %w", filename, err)
	}

	// Parse tempo map and PPQ
	tempoMap, ppq, err := parseMIDITempo(data)
	if err != nil {
		mp.engine.logger.LogError("Failed to parse MIDI tempo: %v, using defaults", err)
		tempoMap = []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
		ppq = 480
	}

	// Calculate total ticks
	totalTicks := calculateMIDILength(data, ppq)

	// Calculate duration in seconds
	microsPerBeat := 500000 // Default 120 BPM
	if len(tempoMap) > 0 {
		microsPerBeat = tempoMap[0].MicrosPerBeat
	}
	bpm := 60000000 / microsPerBeat
	durationSeconds := float64(totalTicks) / float64(ppq) * float64(microsPerBeat) / 1000000.0

	mp.engine.logger.LogInfo("MIDI file: %s, PPQ: %d, Total ticks: %d, BPM: %d, Duration: %.2fs",
		filename, ppq, totalTicks, bpm, durationSeconds)

	// Create synthesizer
	settings := meltysynth.NewSynthesizerSettings(MIDISampleRate)
	synthesizer, err := meltysynth.NewSynthesizer(mp.soundFont, settings)
	if err != nil {
		return fmt.Errorf("failed to create synthesizer: %w", err)
	}

	// Create sequencer
	sequencer := meltysynth.NewMidiFileSequencer(synthesizer)
	sequencer.Play(midiFile, false) // loop=false

	// Create tick generator
	tickGen := NewWallClockTickGenerator(MIDISampleRate, ppq, tempoMap)

	// Create MIDI stream
	mp.stream = &MIDIStream{
		sequencer:     sequencer,
		tickGenerator: tickGen,
		startTime:     time.Now(),
		totalTicks:    totalTicks,
		engine:        mp.engine,
		lastTick:      -1,
	}

	// Create audio player
	mp.player, err = mp.audioContext.NewPlayer(mp.stream)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %w", err)
	}

	// Mute in headless mode
	if mp.engine.IsHeadless() {
		mp.player.SetVolume(0)
		mp.engine.logger.LogInfo("MIDI audio muted (headless mode)")
	}

	// Start playback in goroutine (non-blocking)
	mp.isPlaying = true
	go func() {
		mp.player.Play()
		mp.engine.logger.LogInfo("MIDI playback started: %s", filename)
	}()

	return nil
}

// Stop stops MIDI playback.
func (mp *MIDIPlayer) Stop() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if mp.player != nil {
		mp.player.Close()
		mp.player = nil
	}
	mp.isPlaying = false
}

// IsPlaying returns whether MIDI is currently playing.
func (mp *MIDIPlayer) IsPlaying() bool {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Check if we have a player and it's marked as playing
	if !mp.isPlaying || mp.player == nil {
		return false
	}

	// Check if the audio player is still playing
	// Note: player.IsPlaying() returns false when the stream has ended
	return mp.player.IsPlaying()
}

// UpdateHeadless updates MIDI tick in headless mode.
// This is called from the main loop when running in headless mode.
func (mp *MIDIPlayer) UpdateHeadless() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	if !mp.isPlaying || mp.stream == nil {
		return
	}

	// Calculate current tick based on wall-clock time (in MIDI PPQ units)
	elapsed := time.Since(mp.stream.startTime).Seconds()
	currentMIDITick := mp.stream.tickGenerator.CalculateTickFromTime(elapsed)

	// Convert MIDI ticks to FILLY ticks (32nd note resolution)
	// 1 quarter note = 8 FILLY ticks (32nd notes)
	// 1 quarter note = PPQ MIDI ticks
	// Therefore: FILLY tick = MIDI tick * 8 / PPQ
	ppq := mp.stream.tickGenerator.ppq
	currentTick := currentMIDITick * 8 / ppq

	// Calculate how many ticks have advanced
	ticksAdvanced := currentTick - mp.stream.lastTick
	if ticksAdvanced <= 0 {
		return
	}

	// Check if MIDI has ended (compare MIDI ticks)
	if currentMIDITick >= mp.stream.totalTicks {
		if !mp.stream.endReported {
			mp.stream.endReported = true
			mp.engine.logger.LogInfo("MIDI playback completed at MIDI tick %d", currentMIDITick)
			// Trigger MIDI_END event
			mp.engine.TriggerEvent(EventMIDI_END, &EventData{})
			// Mark MIDI player as not playing
			mp.isPlaying = false
		}
		return
	}

	// Update MIDI sequences with the number of ticks advanced (in FILLY ticks)
	mp.engine.UpdateMIDISequences(ticksAdvanced)

	// Update the last delivered tick (in FILLY ticks)
	mp.stream.tickGenerator.SetLastDeliveredTick(currentMIDITick)
	mp.stream.lastTick = currentTick
}

// MIDIStream implements io.Reader for MIDI audio streaming.
type MIDIStream struct {
	sequencer     *meltysynth.MidiFileSequencer
	tickGenerator *WallClockTickGenerator
	startTime     time.Time
	totalTicks    int
	engine        *Engine
	lastTick      int
	endReported   bool
	mutex         sync.Mutex
}

// Read generates audio samples and delivers MIDI ticks.
func (ms *MIDIStream) Read(p []byte) (int, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Calculate how many samples to render
	sampleCount := len(p) / 4 // 2 channels * 2 bytes per sample

	// Render audio samples
	left := make([]float32, sampleCount)
	right := make([]float32, sampleCount)
	ms.sequencer.Render(left, right)

	// Convert float32 samples to int16 and interleave
	for i := 0; i < sampleCount; i++ {
		// Clamp and convert left channel
		leftSample := int16(left[i] * 32767)
		binary.LittleEndian.PutUint16(p[i*4:], uint16(leftSample))

		// Clamp and convert right channel
		rightSample := int16(right[i] * 32767)
		binary.LittleEndian.PutUint16(p[i*4+2:], uint16(rightSample))
	}

	// Calculate current tick based on wall-clock time (in MIDI PPQ units)
	elapsed := time.Since(ms.startTime).Seconds()
	currentMIDITick := ms.tickGenerator.CalculateTickFromTime(elapsed)

	// Convert MIDI ticks to FILLY ticks (32nd note resolution)
	// 1 quarter note = 8 FILLY ticks (32nd notes)
	// 1 quarter note = PPQ MIDI ticks
	// Therefore: FILLY tick = MIDI tick * 8 / PPQ
	ppq := ms.tickGenerator.ppq
	currentTick := currentMIDITick * 8 / ppq

	// Calculate how many ticks have advanced
	ticksAdvanced := currentTick - ms.lastTick
	if ticksAdvanced > 0 {
		ms.engine.logger.LogDebug("MIDIStream.Read: ticksAdvanced=%d (currentTick=%d, lastTick=%d, currentMIDITick=%d)",
			ticksAdvanced, currentTick, ms.lastTick, currentMIDITick)

		// Check if MIDI has ended (compare MIDI ticks)
		if currentMIDITick >= ms.totalTicks {
			if !ms.endReported {
				ms.endReported = true
				ms.engine.logger.LogInfo("MIDI playback completed at MIDI tick %d", currentMIDITick)
				// Trigger MIDI_END event
				ms.engine.TriggerEvent(EventMIDI_END, &EventData{})
				// Mark MIDI player as not playing
				if ms.engine.midiPlayer != nil {
					ms.engine.midiPlayer.mutex.Lock()
					ms.engine.midiPlayer.isPlaying = false
					ms.engine.midiPlayer.mutex.Unlock()
				}
			}
		} else {
			// Update MIDI sequences with the number of ticks advanced (in FILLY ticks)
			ms.engine.UpdateMIDISequences(ticksAdvanced)
		}

		// Update the last delivered tick (in FILLY ticks)
		ms.tickGenerator.SetLastDeliveredTick(currentMIDITick)
		ms.lastTick = currentTick
	}

	return len(p), nil
}

// parseMIDITempo extracts tempo events and PPQ from MIDI file data.
func parseMIDITempo(data []byte) ([]TempoEvent, int, error) {
	if len(data) < 14 {
		return nil, 480, fmt.Errorf("MIDI data too short")
	}

	// Check header chunk (MThd)
	if string(data[0:4]) != "MThd" {
		return nil, 480, fmt.Errorf("invalid MIDI header")
	}

	// Extract PPQ from time division (bytes 12-13)
	timeDivision := int(data[12])<<8 | int(data[13])
	ppq := 480
	if timeDivision&0x8000 == 0 {
		ppq = timeDivision
	}

	// Default tempo: 120 BPM (500000 microseconds per beat)
	events := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}

	// Scan tracks for tempo events
	offset := 14
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		// Check for track chunk (MTrk)
		if string(data[offset:offset+4]) != "MTrk" {
			offset += 4
			continue
		}

		// Read track length
		trackLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 |
			int(data[offset+6])<<8 | int(data[offset+7])
		offset += 8
		trackEnd := offset + trackLen
		if trackEnd > len(data) {
			trackEnd = len(data)
		}

		// Parse track events
		currentTick := 0
		pos := offset
		for pos < trackEnd {
			// Read delta time
			delta, n := readVarInt(data[pos:])
			pos += n
			currentTick += delta

			if pos >= trackEnd {
				break
			}

			// Read event type
			eventType := data[pos]
			pos++

			// Check for meta event (0xFF)
			if eventType == 0xFF {
				if pos >= trackEnd {
					break
				}
				metaType := data[pos]
				pos++

				// Read meta event length
				length, n := readVarInt(data[pos:])
				pos += n

				// Check for Set Tempo event (0x51)
				if metaType == 0x51 && length == 3 && pos+3 <= trackEnd {
					microsPerBeat := int(data[pos])<<16 | int(data[pos+1])<<8 | int(data[pos+2])
					events = append(events, TempoEvent{
						Tick:          currentTick,
						MicrosPerBeat: microsPerBeat,
					})
				}

				pos += length
			} else {
				// Skip other events
				// This is a simplified parser - just skip to next event
				if pos < trackEnd {
					pos++
				}
			}
		}

		offset = trackEnd
	}

	return events, ppq, nil
}

// readVarInt reads a variable-length integer from MIDI data.
func readVarInt(data []byte) (int, int) {
	value := 0
	bytesRead := 0

	for i := 0; i < len(data) && i < 4; i++ {
		b := data[i]
		bytesRead++
		value = (value << 7) | int(b&0x7F)
		if b&0x80 == 0 {
			break
		}
	}

	return value, bytesRead
}

// calculateMIDILength calculates the total number of ticks in a MIDI file
// by finding the last event in all tracks.
func calculateMIDILength(data []byte, ppq int) int {
	if len(data) < 14 {
		return 0
	}

	// Skip header
	offset := 14
	maxTick := 0

	// Parse all tracks
	for offset < len(data) {
		if offset+8 > len(data) {
			break
		}

		chunkType := string(data[offset : offset+4])
		chunkLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
		offset += 8

		if chunkType == "MTrk" {
			if offset+chunkLen > len(data) {
				break // Invalid chunk length
			}
			trackData := data[offset : offset+chunkLen]
			trackOffset := 0
			currentTick := 0

			// Parse all events in this track
			for trackOffset < len(trackData) {
				// Read delta time
				deltaTime, consumed := readVarInt(trackData[trackOffset:])
				if consumed == 0 {
					break
				}
				trackOffset += consumed
				currentTick += deltaTime

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
