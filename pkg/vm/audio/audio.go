// Package audio provides audio-related components for the FILLY virtual machine.
// This file implements the AudioSystem which integrates MIDI player, WAV player,
// and Timer into a unified audio management interface.
package audio

import (
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/zurustar/son-et/pkg/vm"
)

// AudioSystem is the main interface for audio operations in the FILLY VM.
// It manages the lifecycle of all audio components and provides a unified API
// for audio playback and control.
//
// Design: AudioSystem integrates midiPlayer, wavPlayer, timer, and audioCtx.
// The AudioSystem should be the main interface for audio operations.
// It manages the lifecycle of all audio components.
type AudioSystem struct {
	// midiPlayer handles MIDI file playback
	midiPlayer *MIDIPlayer

	// wavPlayer handles WAV file playback
	wavPlayer *WAVPlayer

	// timer generates periodic TIME events
	timer *Timer

	// audioCtx is the shared Ebitengine audio context
	audioCtx *audio.Context

	// eventQueue is the event queue for audio events
	eventQueue *vm.EventQueue

	// muted indicates whether all audio output is muted
	muted bool

	// soundFontPath is the path to the SoundFont file for MIDI playback
	soundFontPath string

	// ownsAudioCtx indicates whether this AudioSystem owns the audio context
	// (and should not create a new one)
	ownsAudioCtx bool

	// mu protects the audio system state
	mu sync.RWMutex
}

// NewAudioSystem creates a new AudioSystem with the specified SoundFont and event queue.
// The AudioSystem creates and shares a single Ebitengine audio context among all components.
//
// Design: NewAudioSystem(soundFontPath string, eventSystem *EventSystem) (*AudioSystem, error)
//
// Parameters:
//   - soundFontPath: Path to the SoundFont (.sf2) file for MIDI playback
//   - eventQueue: Event queue for TIME and MIDI_TIME events
//
// Returns:
//   - *AudioSystem: The initialized AudioSystem
//   - error: Error if initialization fails (e.g., SoundFont not found)
func NewAudioSystem(soundFontPath string, eventQueue *vm.EventQueue) (*AudioSystem, error) {
	// Create shared audio context
	// All audio components share this context for automatic mixing
	audioCtx := audio.NewContext(SampleRate)
	return NewAudioSystemWithContext(soundFontPath, eventQueue, audioCtx)
}

// NewAudioSystemWithContext creates a new AudioSystem with an existing audio context.
// This is useful for testing or when the audio context is managed externally.
//
// Parameters:
//   - soundFontPath: Path to the SoundFont (.sf2) file for MIDI playback
//   - eventQueue: Event queue for TIME and MIDI_TIME events
//   - audioCtx: Existing Ebitengine audio context to use
//
// Returns:
//   - *AudioSystem: The initialized AudioSystem
//   - error: Error if initialization fails (e.g., SoundFont not found)
func NewAudioSystemWithContext(soundFontPath string, eventQueue *vm.EventQueue, audioCtx *audio.Context) (*AudioSystem, error) {
	// Create audio context if not provided
	ownsAudioCtx := false
	if audioCtx == nil {
		audioCtx = audio.NewContext(SampleRate)
		ownsAudioCtx = true
	}

	// Create MIDI player with shared audio context
	// Requirement 4.9: When SoundFont file is provided, system uses it for MIDI synthesis.
	midiPlayer, err := NewMIDIPlayer(soundFontPath, audioCtx, eventQueue)
	if err != nil {
		return nil, err
	}

	// Create WAV player with shared audio context
	// Requirement 5.6: System mixes multiple WAV streams into a single audio output.
	wavPlayer := NewWAVPlayer(audioCtx)

	// Create timer for TIME event generation
	// Requirement 3.1: System generates TIME events periodically.
	timer := NewTimer(DefaultTimerInterval, eventQueue)

	return &AudioSystem{
		midiPlayer:    midiPlayer,
		wavPlayer:     wavPlayer,
		timer:         timer,
		audioCtx:      audioCtx,
		eventQueue:    eventQueue,
		muted:         false,
		soundFontPath: soundFontPath,
		ownsAudioCtx:  ownsAudioCtx,
	}, nil
}

// PlayMIDI starts playback of the specified MIDI file.
// If another MIDI is currently playing, it will be stopped first.
//
// Design: func (as *AudioSystem) PlayMIDI(filename string) error
//
// Requirement 4.1: When PlayMIDI(filename) is called, system starts playback of specified MIDI file.
// Requirement 4.6: When another MIDI is playing and PlayMIDI is called,
//
//	system stops the previous MIDI and starts the new one.
//
// Parameters:
//   - filename: Path to the MIDI file to play
//
// Returns:
//   - error: Error if the file cannot be loaded or played
func (as *AudioSystem) PlayMIDI(filename string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.midiPlayer == nil {
		return ErrNoSoundFont
	}

	return as.midiPlayer.Play(filename)
}

// PlayWAVE starts playback of the specified WAV file.
// Multiple WAV files can be played simultaneously.
//
// Design: func (as *AudioSystem) PlayWAVE(filename string) error
//
// Requirement 5.1: When PlayWAVE(filename) is called, system starts playback of specified WAV file.
// Requirement 5.2: When multiple PlayWAVE calls are made, system plays all WAV files simultaneously.
//
// Parameters:
//   - filename: Path to the WAV file to play
//
// Returns:
//   - error: Error if the file cannot be loaded or played
func (as *AudioSystem) PlayWAVE(filename string) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.wavPlayer == nil {
		return nil // No error, just skip if not initialized
	}

	return as.wavPlayer.Play(filename)
}

// SetMuted sets the muted state for all audio components.
// When muted, audio is not output but events (MIDI_TIME, TIME) are still generated.
//
// Design: func (as *AudioSystem) SetMuted(muted bool)
//
// Requirement 12.2: When headless mode is enabled, system mutes all audio output.
// Requirement 12.3: When headless mode is enabled, system generates MIDI_TIME events normally.
//
// Parameters:
//   - muted: true to mute all audio, false to unmute
func (as *AudioSystem) SetMuted(muted bool) {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.muted = muted

	// Mute MIDI player
	if as.midiPlayer != nil {
		as.midiPlayer.SetMuted(muted)
	}

	// Mute WAV player
	if as.wavPlayer != nil {
		as.wavPlayer.SetMuted(muted)
	}
}

// IsMuted returns whether the audio system is muted.
func (as *AudioSystem) IsMuted() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.muted
}

// Update is called from the game loop to update all audio components.
// This method should be called every frame to:
// - Generate MIDI_TIME events based on playback position
// - Generate MIDI_END events when playback completes
// - Clean up finished WAV players
//
// Design: func (as *AudioSystem) Update() // Called from game loop
//
// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
func (as *AudioSystem) Update() {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Update MIDI player (generates MIDI_TIME and MIDI_END events)
	if as.midiPlayer != nil {
		as.midiPlayer.Update()
	}

	// Update WAV player (cleanup finished players)
	if as.wavPlayer != nil {
		as.wavPlayer.Update()
	}
}

// Shutdown stops all audio playback and releases resources.
// This should be called when the VM is shutting down.
//
// Design: func (as *AudioSystem) Shutdown()
//
// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
// Requirement 15.3: When ExitTitle is called, system cleans up all resources.
func (as *AudioSystem) Shutdown() {
	as.mu.Lock()
	defer as.mu.Unlock()

	// Stop timer
	if as.timer != nil {
		as.timer.Stop()
	}

	// Stop MIDI playback
	if as.midiPlayer != nil {
		as.midiPlayer.Stop()
	}

	// Stop all WAV playback
	if as.wavPlayer != nil {
		as.wavPlayer.StopAll()
	}
}

// StartTimer starts the timer for TIME event generation.
//
// Requirement 3.1: System generates TIME events periodically.
func (as *AudioSystem) StartTimer() {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.timer != nil {
		as.timer.Start()
	}
}

// StopTimer stops the timer.
func (as *AudioSystem) StopTimer() {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.timer != nil {
		as.timer.Stop()
	}
}

// IsTimerRunning returns whether the timer is currently running.
func (as *AudioSystem) IsTimerRunning() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if as.timer == nil {
		return false
	}
	return as.timer.IsRunning()
}

// IsMIDIPlaying returns whether MIDI is currently playing.
func (as *AudioSystem) IsMIDIPlaying() bool {
	as.mu.RLock()
	defer as.mu.RUnlock()

	if as.midiPlayer == nil {
		return false
	}
	return as.midiPlayer.IsPlaying()
}

// StopMIDI stops the current MIDI playback.
func (as *AudioSystem) StopMIDI() {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.midiPlayer != nil {
		as.midiPlayer.Stop()
	}
}

// StopAllWAV stops all WAV playback.
func (as *AudioSystem) StopAllWAV() {
	as.mu.Lock()
	defer as.mu.Unlock()

	if as.wavPlayer != nil {
		as.wavPlayer.StopAll()
	}
}

// GetMIDIPlayer returns the MIDI player for advanced operations.
// This is useful for testing and debugging.
func (as *AudioSystem) GetMIDIPlayer() *MIDIPlayer {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.midiPlayer
}

// GetWAVPlayer returns the WAV player for advanced operations.
// This is useful for testing and debugging.
func (as *AudioSystem) GetWAVPlayer() *WAVPlayer {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.wavPlayer
}

// GetTimer returns the timer for advanced operations.
// This is useful for testing and debugging.
func (as *AudioSystem) GetTimer() *Timer {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.timer
}

// GetAudioContext returns the shared audio context.
// This is useful for testing and debugging.
func (as *AudioSystem) GetAudioContext() *audio.Context {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.audioCtx
}

// GetEventQueue returns the event queue.
func (as *AudioSystem) GetEventQueue() *vm.EventQueue {
	as.mu.RLock()
	defer as.mu.RUnlock()
	return as.eventQueue
}
