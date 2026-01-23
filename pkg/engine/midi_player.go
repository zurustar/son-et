package engine

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/sinshu/go-meltysynth/meltysynth"
	"gitlab.com/gomidi/midi/v2"
	"gitlab.com/gomidi/midi/v2/smf"
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

// MIDIBridge implements midi.Writer to forward MIDI messages from gomidi to meltysynth.
type MIDIBridge struct {
	synthesizer *meltysynth.Synthesizer
	mutex       sync.Mutex
}

// NewMIDIBridge creates a new MIDI bridge.
func NewMIDIBridge(synthesizer *meltysynth.Synthesizer) *MIDIBridge {
	return &MIDIBridge{
		synthesizer: synthesizer,
	}
}

// Write implements midi.Writer interface to forward MIDI messages to meltysynth.
func (mb *MIDIBridge) Write(msg midi.Message) error {
	mb.mutex.Lock()
	defer mb.mutex.Unlock()

	// Extract MIDI message components
	channel, command, data1, data2 := extractMIDIComponents(msg)

	// Forward to meltysynth
	mb.synthesizer.ProcessMidiMessage(int32(channel), int32(command), int32(data1), int32(data2))

	return nil
}

// extractMIDIComponents extracts channel, command, data1, and data2 from a gomidi MIDI message.
func extractMIDIComponents(msg midi.Message) (channel, command, data1, data2 byte) {
	bytes := msg.Bytes()
	if len(bytes) == 0 {
		return 0, 0, 0, 0
	}

	// First byte contains status (command + channel for channel messages)
	status := bytes[0]

	// For channel messages (0x80-0xEF), extract channel and command
	if status >= 0x80 && status < 0xF0 {
		channel = status & 0x0F
		command = status & 0xF0
	} else {
		// System messages (0xF0-0xFF) - no channel
		channel = 0
		command = status
	}

	// Extract data bytes
	if len(bytes) > 1 {
		data1 = bytes[1]
	}
	if len(bytes) > 2 {
		data2 = bytes[2]
	}

	return channel, command, data1, data2
}

// MIDIPlayer manages MIDI playback with tick generation.
type MIDIPlayer struct {
	audioContext *audio.Context
	soundFont    *meltysynth.SoundFont
	player       *audio.Player
	stream       *MIDIStream
	bridge       *MIDIBridge
	stopChan     chan bool
	finishedChan chan bool
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
	if mp.stopChan != nil {
		close(mp.stopChan)
	}
	if mp.finishedChan != nil {
		close(mp.finishedChan)
	}

	// Load MIDI file via AssetLoader
	data, err := mp.engine.state.assetLoader.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load MIDI file %s: %w", filename, err)
	}

	// Parse MIDI file using gomidi
	smfData, err := smf.ReadFrom(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to parse MIDI file %s: %w", filename, err)
	}

	// Extract PPQ from time format
	ppq := 480 // default
	if mt, ok := smfData.TimeFormat.(smf.MetricTicks); ok {
		ppq = int(mt)
	}

	// Parse tempo map for tick generation
	tempoMap, err := extractTempoMap(smfData, ppq)
	if err != nil {
		mp.engine.logger.LogError("Failed to extract tempo map: %v, using defaults", err)
		tempoMap = []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
	}

	// Calculate total ticks (find the last event across all tracks)
	totalTicks := 0
	for _, track := range smfData.Tracks {
		absTick := 0
		for _, event := range track {
			absTick += int(event.Delta)
		}
		if absTick > totalTicks {
			totalTicks = absTick
		}
	}

	// Calculate duration in seconds (considering tempo changes)
	durationSeconds := 0.0
	lastTick := 0
	lastTempo := 500000 // Default 120 BPM

	for i, tempoEvent := range tempoMap {
		if i > 0 {
			// Calculate duration for the previous tempo segment
			ticksInSegment := tempoEvent.Tick - lastTick
			durationSeconds += float64(ticksInSegment) / float64(ppq) * float64(lastTempo) / 1000000.0
		}
		lastTick = tempoEvent.Tick
		lastTempo = tempoEvent.MicrosPerBeat
	}

	// Add duration for the final segment (from last tempo change to end)
	if totalTicks > lastTick {
		ticksInSegment := totalTicks - lastTick
		durationSeconds += float64(ticksInSegment) / float64(ppq) * float64(lastTempo) / 1000000.0
	}

	// For logging, use the last tempo (most representative)
	bpm := 60000000 / lastTempo

	mp.engine.logger.LogInfo("MIDI file: %s, PPQ: %d, Total ticks: %d, BPM: %d, Duration: %.2fs",
		filename, ppq, totalTicks, bpm, durationSeconds)

	// Create synthesizer
	settings := meltysynth.NewSynthesizerSettings(MIDISampleRate)
	synthesizer, err := meltysynth.NewSynthesizer(mp.soundFont, settings)
	if err != nil {
		return fmt.Errorf("failed to create synthesizer: %w", err)
	}

	// Create MIDI bridge
	mp.bridge = NewMIDIBridge(synthesizer)

	// Create channels for playback control
	mp.stopChan = make(chan bool, 1)
	mp.finishedChan = make(chan bool, 1)

	// Create tick generator
	tickGen := NewWallClockTickGenerator(MIDISampleRate, ppq, tempoMap)

	// Create MIDI stream
	mp.stream = &MIDIStream{
		synthesizer:   synthesizer,
		tickGenerator: tickGen,
		startTime:     time.Now(),
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

	// Start goroutine to play MIDI messages through the bridge
	go mp.playMIDIMessages(smfData, ppq)

	// Start audio playback
	go func() {
		// Check if context is already cancelled before starting
		select {
		case <-mp.engine.GetContext().Done():
			mp.engine.logger.LogInfo("MIDI playback cancelled before start")
			mp.mutex.Lock()
			mp.isPlaying = false
			mp.mutex.Unlock()
			return
		default:
		}

		// Check if player is valid before playing
		mp.mutex.Lock()
		player := mp.player
		mp.mutex.Unlock()

		if player == nil {
			mp.engine.logger.LogError("Audio player is nil, cannot start playback")
			mp.mutex.Lock()
			mp.isPlaying = false
			mp.mutex.Unlock()
			return
		}

		player.Play()
		mp.engine.logger.LogInfo("MIDI playback started: %s", filename)

		// Monitor context cancellation
		go func() {
			<-mp.engine.GetContext().Done()
			mp.engine.logger.LogInfo("MIDI playback cancelled by context")
			mp.mutex.Lock()
			if mp.stopChan != nil {
				select {
				case mp.stopChan <- true:
				default:
				}
			}
			if mp.player != nil {
				mp.player.Close()
				mp.player = nil
			}
			mp.isPlaying = false
			mp.mutex.Unlock()
		}()
	}()

	// Monitor finished channel
	go func() {
		<-mp.finishedChan
		mp.engine.logger.LogInfo("MIDI playback completed")
		mp.mutex.Lock()
		mp.isPlaying = false
		mp.mutex.Unlock()
		// Trigger MIDI_END event
		mp.engine.TriggerEvent(EventMIDI_END, &EventData{})
	}()

	return nil
}

// playMIDIMessages plays MIDI messages from the SMF file through the bridge with proper timing.
func (mp *MIDIPlayer) playMIDIMessages(smfData *smf.SMF, ppq int) {
	// Build a timeline of all MIDI messages across all tracks
	type timedMessage struct {
		absTick int
		message smf.Message
	}
	var timeline []timedMessage

	for _, track := range smfData.Tracks {
		absTick := 0
		for _, event := range track {
			absTick += int(event.Delta)
			msg := event.Message

			// Skip meta messages (they don't produce sound)
			if msg.IsMeta() {
				continue
			}

			// Only process playable MIDI messages
			if msg.IsPlayable() {
				timeline = append(timeline, timedMessage{
					absTick: absTick,
					message: msg,
				})
			}
		}
	}

	// Sort timeline by absolute tick
	// (Simple bubble sort since we expect relatively few messages)
	for i := 0; i < len(timeline); i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[j].absTick < timeline[i].absTick {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}

	// Get tempo map from the stream's tick generator
	mp.mutex.Lock()
	tempoMap := mp.stream.tickGenerator.tempoMap
	mp.mutex.Unlock()

	// Play messages with timing
	lastTick := 0
	for _, tm := range timeline {
		// Calculate wait time based on tick delta
		tickDelta := tm.absTick - lastTick
		if tickDelta > 0 {
			// Calculate wait duration based on tempo map
			waitDuration := mp.calculateWaitDuration(lastTick, tm.absTick, ppq, tempoMap)

			// Wait or check for stop signal
			select {
			case <-mp.stopChan:
				mp.engine.logger.LogInfo("MIDI playback stopped")
				return
			case <-time.After(waitDuration):
				// Continue
			}
		}

		// Send message to bridge (convert smf.Message to midi.Message)
		midiMsg := midi.Message(tm.message.Bytes())
		if err := mp.bridge.Write(midiMsg); err != nil {
			mp.engine.logger.LogError("Failed to write MIDI message: %v", err)
		}

		lastTick = tm.absTick
	}

	// Signal completion
	select {
	case mp.finishedChan <- true:
	default:
	}
}

// calculateWaitDuration calculates the wait duration between two MIDI ticks,
// accounting for tempo changes in the tempo map.
func (mp *MIDIPlayer) calculateWaitDuration(startTick, endTick, ppq int, tempoMap []TempoEvent) time.Duration {
	totalDuration := 0.0
	currentTick := startTick

	// Find the tempo at the start tick
	currentTempoIdx := 0
	for i := len(tempoMap) - 1; i >= 0; i-- {
		if tempoMap[i].Tick <= startTick {
			currentTempoIdx = i
			break
		}
	}

	// Calculate duration segment by segment, handling tempo changes
	for currentTick < endTick {
		currentTempo := tempoMap[currentTempoIdx]
		microsPerBeat := float64(currentTempo.MicrosPerBeat)

		// Calculate time per MIDI tick at current tempo
		// 1 quarter note = PPQ ticks
		// Time per tick = (microsPerBeat / 1000000) / PPQ seconds
		timePerTick := (microsPerBeat / 1000000.0) / float64(ppq)

		// Determine the end of this tempo segment
		segmentEndTick := endTick
		if currentTempoIdx+1 < len(tempoMap) {
			nextTempoTick := tempoMap[currentTempoIdx+1].Tick
			if nextTempoTick < endTick {
				segmentEndTick = nextTempoTick
			}
		}

		// Calculate duration for this segment
		ticksInSegment := segmentEndTick - currentTick
		segmentDuration := float64(ticksInSegment) * timePerTick
		totalDuration += segmentDuration

		// Move to next segment
		currentTick = segmentEndTick
		if currentTempoIdx+1 < len(tempoMap) && currentTick >= tempoMap[currentTempoIdx+1].Tick {
			currentTempoIdx++
		}
	}

	return time.Duration(totalDuration * float64(time.Second))
}

// Stop stops MIDI playback.
func (mp *MIDIPlayer) Stop() {
	mp.mutex.Lock()
	defer mp.mutex.Unlock()

	// Signal stop channel
	if mp.stopChan != nil {
		select {
		case mp.stopChan <- true:
		default:
		}
	}

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

	// Check for context cancellation
	select {
	case <-mp.engine.GetContext().Done():
		if mp.isPlaying {
			mp.engine.logger.LogInfo("MIDI playback stopped due to context cancellation")
			mp.isPlaying = false
		}
		return
	default:
	}

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

	// Update MIDI sequences with the number of ticks advanced (in FILLY ticks)
	mp.engine.UpdateMIDISequences(ticksAdvanced)

	// Update the last delivered tick (in FILLY ticks)
	mp.stream.tickGenerator.SetLastDeliveredTick(currentMIDITick)
	mp.stream.lastTick = currentTick
}

// MIDIStream implements io.Reader for MIDI audio streaming.
type MIDIStream struct {
	synthesizer   *meltysynth.Synthesizer
	tickGenerator *WallClockTickGenerator
	startTime     time.Time
	engine        *Engine
	lastTick      int
	mutex         sync.Mutex
}

// Read generates audio samples and delivers MIDI ticks.
func (ms *MIDIStream) Read(p []byte) (int, error) {
	ms.mutex.Lock()
	defer ms.mutex.Unlock()

	// Check for context cancellation
	select {
	case <-ms.engine.GetContext().Done():
		return 0, io.EOF
	default:
	}

	// Calculate how many samples to render
	sampleCount := len(p) / 4 // 2 channels * 2 bytes per sample

	// Render audio samples
	left := make([]float32, sampleCount)
	right := make([]float32, sampleCount)
	ms.synthesizer.Render(left, right)

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

		// Update MIDI sequences with the number of ticks advanced (in FILLY ticks)
		ms.engine.UpdateMIDISequences(ticksAdvanced)

		// Update the last delivered tick (in FILLY ticks)
		ms.tickGenerator.SetLastDeliveredTick(currentMIDITick)
		ms.lastTick = currentTick
	}

	return len(p), nil
}

// extractTempoMap extracts tempo events from a gomidi SMF file.
func extractTempoMap(smfData *smf.SMF, ppq int) ([]TempoEvent, error) {
	// Default tempo: 120 BPM (500000 microseconds per beat)
	events := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}

	// Scan all tracks for tempo events
	for _, track := range smfData.Tracks {
		absTick := 0
		for _, event := range track {
			absTick += int(event.Delta)
			msg := event.Message

			// Check for tempo meta message
			var bpm float64
			if msg.GetMetaTempo(&bpm) {
				microsPerBeat := int(60000000 / bpm)
				events = append(events, TempoEvent{
					Tick:          absTick,
					MicrosPerBeat: microsPerBeat,
				})
			}
		}
	}

	return events, nil
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
		lastStatus := byte(0)

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

			// Handle running status
			if eventType < 0x80 {
				eventType = lastStatus
			} else {
				pos++
				if eventType >= 0x80 && eventType < 0xF0 {
					lastStatus = eventType
				}
			}

			// Check for meta event (0xFF)
			if eventType == 0xFF {
				lastStatus = 0
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
			} else if eventType == 0xF0 || eventType == 0xF7 {
				// SysEx event
				lastStatus = 0
				length, n := readVarInt(data[pos:])
				pos += n + length
			} else if eventType >= 0x80 {
				// MIDI channel event
				if eventType >= 0xC0 && eventType < 0xE0 {
					// Program change or channel pressure (1 data byte)
					pos++
				} else {
					// Other events (2 data bytes)
					pos += 2
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
			lastStatus := byte(0) // Track running status

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

				// Read event byte
				eventByte := trackData[trackOffset]

				// Handle running status (if event byte < 0x80, use last status)
				if eventByte < 0x80 {
					eventByte = lastStatus
					// Don't increment trackOffset - the data byte is part of the event
				} else {
					trackOffset++
					// Update running status for channel messages (0x80-0xEF)
					if eventByte >= 0x80 && eventByte < 0xF0 {
						lastStatus = eventByte
					}
				}

				if eventByte == 0xFF {
					// Meta event (clears running status)
					lastStatus = 0
					if trackOffset >= len(trackData) {
						break
					}
					trackOffset++ // Skip meta type
					length, consumed := readVarInt(trackData[trackOffset:])
					trackOffset += consumed + length
				} else if eventByte == 0xF0 || eventByte == 0xF7 {
					// SysEx event (clears running status)
					lastStatus = 0
					length, consumed := readVarInt(trackData[trackOffset:])
					trackOffset += consumed + length
				} else if eventByte >= 0x80 {
					// MIDI channel event
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
