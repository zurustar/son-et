// Package audio provides audio-related components for the FILLY virtual machine.
// This file implements the MIDI Player for MIDI file playback using go-meltysynth
// and Ebitengine/audio.
package audio

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/sinshu/go-meltysynth/meltysynth"
	"github.com/zurustar/son-et/pkg/vm"
)

// SampleRate is the audio sample rate used for MIDI synthesis.
const SampleRate = 44100

// ErrNoSoundFont is returned when no SoundFont file is provided.
// Requirement 4.10: When SoundFont is not provided, system reports error.
var ErrNoSoundFont = errors.New("SoundFont file is required for MIDI playback")

// ErrSoundFontNotFound is returned when the SoundFont file cannot be found.
var ErrSoundFontNotFound = errors.New("SoundFont file not found")

// ErrMIDIFileNotFound is returned when the MIDI file cannot be found.
var ErrMIDIFileNotFound = errors.New("MIDI file not found")

// ErrMIDIInvalidFormat is returned when the MIDI file has an invalid format.
var ErrMIDIInvalidFormat = errors.New("invalid MIDI file format")

// MIDIStream implements io.Reader for Ebitengine/audio.
// It renders audio samples from the MIDI sequencer.
//
// Requirement 4.8: System uses software synthesizer to render MIDI audio.
type MIDIStream struct {
	sequencer   *meltysynth.MidiFileSequencer
	sampleCount int64
	stopped     bool
	mu          sync.Mutex
}

// Read implements io.Reader interface for MIDIStream.
// It renders audio samples from the sequencer and converts them to int16 format.
func (s *MIDIStream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if stream has been stopped
	if s.stopped || s.sequencer == nil {
		// Return silence (zeros)
		for i := range p {
			p[i] = 0
		}
		return len(p), nil
	}

	// Calculate number of samples (16-bit stereo = 4 bytes per sample)
	samples := len(p) / 4
	if samples == 0 {
		return 0, nil
	}

	// Allocate buffers for left and right channels
	left := make([]float32, samples)
	right := make([]float32, samples)

	// Render audio from sequencer
	s.sequencer.Render(left, right)
	s.sampleCount += int64(samples)

	// Convert float32 to int16 interleaved stereo
	for i := range samples {
		l := int16(clamp(left[i], -1, 1) * 32767)
		r := int16(clamp(right[i], -1, 1) * 32767)
		binary.LittleEndian.PutUint16(p[i*4:], uint16(l))
		binary.LittleEndian.PutUint16(p[i*4+2:], uint16(r))
	}

	return len(p), nil
}

// Stop marks the stream as stopped, causing Read to return silence.
func (s *MIDIStream) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.stopped = true
}

// GetSampleCount returns the total number of samples rendered.
func (s *MIDIStream) GetSampleCount() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sampleCount
}

// Reset resets the sample count.
func (s *MIDIStream) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sampleCount = 0
}

// clamp restricts a value to the range [min, max].
func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// TempoEvent represents a tempo change in a MIDI file.
// Requirement 4.2: When MIDI playback starts, system extracts tempo information from MIDI file.
type TempoEvent struct {
	Tick          int // MIDI tick position
	MicrosPerBeat int // Microseconds per quarter note
}

// TickCalculator calculates MIDI ticks from sample count considering tempo changes.
// Requirement 18.1: When MIDI file contains tempo change events, system detects them.
// Requirement 18.2: When tempo change event is encountered, system updates MIDI_TIME event interval.
type TickCalculator struct {
	ppq           int          // Ticks per quarter note (from MIDI header)
	tempoMap      []TempoEvent // List of tempo change events
	sampleAtTempo []int64      // Pre-calculated sample count at each tempo change
}

// NewTickCalculator creates a new TickCalculator with the given PPQ and tempo map.
func NewTickCalculator(ppq int, tempoMap []TempoEvent) *TickCalculator {
	tc := &TickCalculator{
		ppq:      ppq,
		tempoMap: tempoMap,
	}
	tc.precalculate()
	return tc
}

// precalculate computes the sample count at each tempo change point.
// This allows efficient conversion from samples to ticks.
func (tc *TickCalculator) precalculate() {
	tc.sampleAtTempo = make([]int64, len(tc.tempoMap))
	if len(tc.tempoMap) == 0 {
		return
	}

	tc.sampleAtTempo[0] = 0

	for i := 1; i < len(tc.tempoMap); i++ {
		prevTempo := tc.tempoMap[i-1]
		currTempo := tc.tempoMap[i]

		// Ticks in previous tempo segment
		ticksInSegment := currTempo.Tick - prevTempo.Tick

		// Samples per tick at previous tempo
		// 1 quarter note = ppq ticks = microsPerBeat microseconds
		// 1 tick = microsPerBeat / ppq microseconds
		// samples per tick = sampleRate * microsPerBeat / ppq / 1000000
		samplesPerTick := float64(SampleRate) * float64(prevTempo.MicrosPerBeat) / float64(tc.ppq) / 1000000.0

		tc.sampleAtTempo[i] = tc.sampleAtTempo[i-1] + int64(float64(ticksInSegment)*samplesPerTick)
	}
}

// TickFromSamples converts sample count to MIDI tick (PPQ units).
// Requirement 18.4: System maintains accurate synchronization across tempo changes.
func (tc *TickCalculator) TickFromSamples(samples int64) int {
	if len(tc.tempoMap) == 0 {
		return 0
	}

	// Find which tempo segment we're in
	segmentIdx := 0
	for i := len(tc.tempoMap) - 1; i >= 0; i-- {
		if samples >= tc.sampleAtTempo[i] {
			segmentIdx = i
			break
		}
	}

	tempo := tc.tempoMap[segmentIdx]
	samplesIntoSegment := samples - tc.sampleAtTempo[segmentIdx]

	// Convert samples to ticks
	samplesPerTick := float64(SampleRate) * float64(tempo.MicrosPerBeat) / float64(tc.ppq) / 1000000.0
	if samplesPerTick <= 0 {
		return tempo.Tick
	}
	ticksIntoSegment := int(float64(samplesIntoSegment) / samplesPerTick)

	return tempo.Tick + ticksIntoSegment
}

// FillyTickFromSamples converts sample count to FILLY tick (32nd note units).
// In FILLY, 1 quarter note = 8 ticks (32nd notes).
func (tc *TickCalculator) FillyTickFromSamples(samples int64) int {
	midiTick := tc.TickFromSamples(samples)
	// 1 quarter note = ppq MIDI ticks = 8 FILLY ticks (32nd notes)
	if tc.ppq == 0 {
		return 0
	}
	return midiTick * 8 / tc.ppq
}

// GetPPQ returns the ticks per quarter note.
func (tc *TickCalculator) GetPPQ() int {
	return tc.ppq
}

// GetTempoMap returns the tempo map.
func (tc *TickCalculator) GetTempoMap() []TempoEvent {
	return tc.tempoMap
}

// MIDIPlayer handles MIDI file playback using go-meltysynth and Ebitengine/audio.
//
// Requirement 4.1: When PlayMIDI(filename) is called, system starts playback of specified MIDI file.
// Requirement 4.7: System supports Standard MIDI File (SMF) format.
// Requirement 4.8: System uses software synthesizer to render MIDI audio.
// Requirement 4.9: When SoundFont file is provided, system uses it for MIDI synthesis.
type MIDIPlayer struct {
	// go-meltysynth components
	soundFont *meltysynth.SoundFont
	synth     *meltysynth.Synthesizer
	sequencer *meltysynth.MidiFileSequencer

	// Ebitengine/audio components
	audioCtx *audio.Context
	player   *audio.Player
	stream   *MIDIStream

	// Tempo management
	tickCalc *TickCalculator

	// Event generation (will be used in task 5.5)
	eventQueue *vm.EventQueue
	lastTick   int

	// State
	playing       bool
	muted         bool
	duration      time.Duration
	soundFontPath string
	currentFile   string

	mu sync.RWMutex
}

// NewMIDIPlayer creates a new MIDI player with the specified SoundFont.
//
// Requirement 4.9: When SoundFont file is provided, system uses it for MIDI synthesis.
// Requirement 4.10: When SoundFont is not provided, system reports error.
//
// Parameters:
//   - soundFontPath: Path to the SoundFont (.sf2) file
//   - audioCtx: Ebitengine audio context (can be nil, will be created if needed)
//   - eventQueue: Event queue for MIDI_TIME events (can be nil for basic playback)
//
// Returns:
//   - *MIDIPlayer: The initialized MIDI player
//   - error: Error if SoundFont cannot be loaded
func NewMIDIPlayer(soundFontPath string, audioCtx *audio.Context, eventQueue *vm.EventQueue) (*MIDIPlayer, error) {
	// Requirement 4.10: When SoundFont is not provided, system reports error.
	if soundFontPath == "" {
		return nil, ErrNoSoundFont
	}

	// Load SoundFont file
	sf2Data, err := os.ReadFile(soundFontPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: %s", ErrSoundFontNotFound, soundFontPath)
		}
		return nil, fmt.Errorf("failed to read SoundFont file: %w", err)
	}

	// Parse SoundFont
	// Requirement 4.9: When SoundFont file is provided, system uses it for MIDI synthesis.
	soundFont, err := meltysynth.NewSoundFont(bytes.NewReader(sf2Data))
	if err != nil {
		return nil, fmt.Errorf("failed to parse SoundFont: %w", err)
	}

	// Create audio context if not provided
	if audioCtx == nil {
		audioCtx = audio.NewContext(SampleRate)
	}

	// Create synthesizer
	// Requirement 4.8: System uses software synthesizer to render MIDI audio.
	settings := meltysynth.NewSynthesizerSettings(SampleRate)
	synth, err := meltysynth.NewSynthesizer(soundFont, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to create synthesizer: %w", err)
	}

	return &MIDIPlayer{
		soundFont:     soundFont,
		synth:         synth,
		audioCtx:      audioCtx,
		eventQueue:    eventQueue,
		soundFontPath: soundFontPath,
		playing:       false,
		muted:         false,
	}, nil
}

// Play starts playback of the specified MIDI file.
//
// Requirement 4.1: When PlayMIDI(filename) is called, system starts playback of specified MIDI file.
// Requirement 4.7: System supports Standard MIDI File (SMF) format.
//
// Parameters:
//   - filename: Path to the MIDI file to play
//
// Returns:
//   - error: Error if the file cannot be loaded or played
func (mp *MIDIPlayer) Play(filename string) error {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	// Stop any current playback (will be enhanced in task 5.7)
	mp.stopInternal()

	// Find file with case-insensitive search (Windows 3.1 compatibility)
	actualPath, err := FindFileInsensitive(filename)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrMIDIFileNotFound, filename)
	}

	// Load MIDI file
	midiData, err := os.ReadFile(actualPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrMIDIFileNotFound, filename)
		}
		return fmt.Errorf("failed to read MIDI file: %w", err)
	}

	// Parse MIDI file
	// Requirement 4.7: System supports Standard MIDI File (SMF) format.
	midi, err := meltysynth.NewMidiFile(bytes.NewReader(midiData))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrMIDIInvalidFormat, err)
	}

	// Extract tempo map and PPQ
	// Requirement 4.2: When MIDI playback starts, system extracts tempo information from MIDI file.
	tempoMap, ppq := ParseMIDITempoMap(midiData)
	mp.tickCalc = NewTickCalculator(ppq, tempoMap)

	// Create sequencer and start playback
	mp.sequencer = meltysynth.NewMidiFileSequencer(mp.synth)
	mp.sequencer.Play(midi, false) // false = don't loop

	// Get duration
	mp.duration = midi.GetLength()

	// Create stream
	mp.stream = &MIDIStream{sequencer: mp.sequencer}

	// Create audio player
	player, err := mp.audioCtx.NewPlayer(mp.stream)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %w", err)
	}
	mp.player = player

	// Set volume based on muted state
	if mp.muted {
		mp.player.SetVolume(0)
	}

	// Start playback
	mp.player.Play()
	mp.playing = true
	mp.currentFile = filename
	mp.lastTick = 0

	return nil
}

// Stop stops the current MIDI playback.
func (mp *MIDIPlayer) Stop() {
	mp.mu.Lock()
	defer mp.mu.Unlock()
	mp.stopInternal()
}

// stopInternal stops playback without acquiring the lock.
// Must be called with mp.mu held.
// Requirement 4.6: When another MIDI is playing and PlayMIDI is called,
// system stops the previous MIDI and starts the new one.
func (mp *MIDIPlayer) stopInternal() {
	// Stop the stream first to prevent further reads
	if mp.stream != nil {
		mp.stream.Stop()
	}
	if mp.player != nil {
		mp.player.Close()
		mp.player = nil
	}
	mp.sequencer = nil
	mp.stream = nil
	mp.playing = false
	mp.currentFile = ""
	mp.lastTick = 0
}

// IsPlaying returns whether MIDI is currently playing.
func (mp *MIDIPlayer) IsPlaying() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.playing
}

// SetMuted sets the muted state of the MIDI player.
// When muted, audio is not output but MIDI_TIME events are still generated.
//
// Requirement 12.2: When headless mode is enabled, system mutes all audio output.
func (mp *MIDIPlayer) SetMuted(muted bool) {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	mp.muted = muted
	if mp.player != nil {
		if muted {
			mp.player.SetVolume(0)
		} else {
			mp.player.SetVolume(1)
		}
	}
}

// IsMuted returns whether the MIDI player is muted.
func (mp *MIDIPlayer) IsMuted() bool {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.muted
}

// GetDuration returns the duration of the current MIDI file.
func (mp *MIDIPlayer) GetDuration() time.Duration {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.duration
}

// GetPosition returns the current playback position.
func (mp *MIDIPlayer) GetPosition() time.Duration {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if mp.player == nil {
		return 0
	}
	return mp.player.Position()
}

// GetCurrentTick returns the current MIDI tick position.
func (mp *MIDIPlayer) GetCurrentTick() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if mp.player == nil || mp.tickCalc == nil {
		return 0
	}

	position := mp.player.Position()
	samples := int64(position.Seconds() * float64(SampleRate))
	return mp.tickCalc.TickFromSamples(samples)
}

// GetCurrentFillyTick returns the current FILLY tick position (32nd note units).
func (mp *MIDIPlayer) GetCurrentFillyTick() int {
	mp.mu.RLock()
	defer mp.mu.RUnlock()

	if mp.player == nil || mp.tickCalc == nil {
		return 0
	}

	position := mp.player.Position()
	samples := int64(position.Seconds() * float64(SampleRate))
	return mp.tickCalc.FillyTickFromSamples(samples)
}

// GetTickCalculator returns the tick calculator for the current MIDI file.
func (mp *MIDIPlayer) GetTickCalculator() *TickCalculator {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.tickCalc
}

// GetCurrentFile returns the path of the currently playing MIDI file.
func (mp *MIDIPlayer) GetCurrentFile() string {
	mp.mu.RLock()
	defer mp.mu.RUnlock()
	return mp.currentFile
}

// Update is called from the game loop to check playback status and generate MIDI_TIME events.
//
// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
// Requirement 4.4: When MIDI tempo is 120 BPM with resolution 480 ticks per beat, system generates MIDI_TIME event every 1.04ms.
// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
//
// The method:
// 1. Gets the current audio playback position using player.Position()
// 2. Converts the position to FILLY ticks using TickCalculator
// 3. If the tick has advanced since the last update, generates MIDI_TIME events
// 4. Pushes the events to the event queue
// 5. When playback completes (position >= duration), generates MIDI_END event
func (mp *MIDIPlayer) Update() {
	mp.mu.Lock()
	defer mp.mu.Unlock()

	if !mp.playing || mp.player == nil {
		return
	}

	// Get current playback position
	position := mp.player.Position()

	// Check if playback has finished
	// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
	if position >= mp.duration {
		mp.playing = false

		// Generate MIDI_END event if event queue is configured
		if mp.eventQueue != nil {
			event := vm.NewEvent(vm.EventMIDI_END)
			mp.eventQueue.Push(event)
		}
		return
	}

	// Generate MIDI_TIME events if tick has advanced
	// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
	if mp.tickCalc != nil && mp.eventQueue != nil {
		// Convert position to samples
		samples := int64(position.Seconds() * float64(SampleRate))

		// Get current FILLY tick (32nd note units)
		currentTick := mp.tickCalc.FillyTickFromSamples(samples)

		// Generate MIDI_TIME events for each tick that has passed
		// Requirement 4.4: System generates MIDI_TIME events at the correct interval
		for tick := mp.lastTick + 1; tick <= currentTick; tick++ {
			event := vm.NewEventWithParams(vm.EventMIDI_TIME, map[string]any{
				"Tick": tick,
			})
			mp.eventQueue.Push(event)
		}

		// Update last tick
		mp.lastTick = currentTick
	}
}

// ParseMIDITempoMap extracts all tempo events and PPQ from MIDI data.
// Requirement 4.2: When MIDI playback starts, system extracts tempo information from MIDI file.
// Requirement 18.1: When MIDI file contains tempo change events, system detects them.
func ParseMIDITempoMap(data []byte) ([]TempoEvent, int) {
	ppq := 480 // Default PPQ
	var events []TempoEvent

	// Check MIDI header
	if len(data) < 14 || string(data[0:4]) != "MThd" {
		// Return default tempo (120 BPM = 500000 microseconds per beat)
		return []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}, ppq
	}

	// Extract PPQ from header (time division)
	timeDivision := int(data[12])<<8 | int(data[13])
	if timeDivision&0x8000 == 0 {
		ppq = timeDivision
	}

	// Scan all tracks for tempo events
	offset := 14
	for offset < len(data) {
		if offset+8 > len(data) || string(data[offset:offset+4]) != "MTrk" {
			break
		}

		trackLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
		trackEnd := offset + 8 + trackLen
		pos := offset + 8
		currentTick := 0
		lastStatus := byte(0)

		for pos < trackEnd {
			// Read delta time (variable length)
			delta, n := readVarLen(data[pos:])
			pos += n
			currentTick += delta

			if pos >= trackEnd {
				break
			}

			eventByte := data[pos]

			// Handle running status
			if eventByte < 0x80 {
				eventByte = lastStatus
			} else {
				pos++
				if eventByte >= 0x80 && eventByte < 0xF0 {
					lastStatus = eventByte
				}
			}

			if eventByte == 0xFF { // Meta event
				if pos >= trackEnd {
					break
				}
				metaType := data[pos]
				pos++
				length, n := readVarLen(data[pos:])
				pos += n

				// Tempo event (0x51)
				if metaType == 0x51 && length == 3 && pos+3 <= trackEnd {
					microsPerBeat := int(data[pos])<<16 | int(data[pos+1])<<8 | int(data[pos+2])
					events = append(events, TempoEvent{Tick: currentTick, MicrosPerBeat: microsPerBeat})
				}
				pos += length
			} else if eventByte == 0xF0 || eventByte == 0xF7 { // SysEx
				length, n := readVarLen(data[pos:])
				pos += n + length
			} else if eventByte >= 0x80 {
				// Channel messages
				if eventByte >= 0xC0 && eventByte < 0xE0 {
					pos++ // 1 data byte
				} else {
					pos += 2 // 2 data bytes
				}
			}
		}
		offset = trackEnd
	}

	// Ensure we have at least one tempo event at tick 0
	if len(events) == 0 {
		events = []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // Default 120 BPM
	} else if events[0].Tick > 0 {
		// Insert default tempo at tick 0 if first tempo event is not at tick 0
		events = append([]TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}, events...)
	}

	return events, ppq
}

// readVarLen reads a variable-length quantity from MIDI data.
func readVarLen(data []byte) (int, int) {
	value := 0
	n := 0
	for i := 0; i < len(data) && i < 4; i++ {
		n++
		value = (value << 7) | int(data[i]&0x7F)
		if data[i]&0x80 == 0 {
			break
		}
	}
	return value, n
}
