// Package audio provides audio-related components for the FILLY virtual machine.
// This file implements the WAV Player for WAV file playback using Ebitengine/audio.
package audio

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// WAV-related errors
var (
	// ErrWAVFileNotFound is returned when the WAV file cannot be found.
	// Requirement 5.4: When WAV file is not found, system logs error and continues execution.
	ErrWAVFileNotFound = errors.New("WAV file not found")

	// ErrWAVInvalidFormat is returned when the WAV file has an invalid format.
	// Requirement 5.5: When WAV file is corrupted, system logs error and continues execution.
	ErrWAVInvalidFormat = errors.New("invalid WAV file format")
)

// WAVPlayer handles WAV file playback using Ebitengine/audio.
// It supports multiple simultaneous playback streams with automatic mixing.
//
// Requirement 5.1: When PlayWAVE(filename) is called, system starts playback of specified WAV file.
// Requirement 5.2: When multiple PlayWAVE calls are made, system plays all WAV files simultaneously.
// Requirement 5.3: System supports standard WAV file formats (PCM, 8-bit, 16-bit).
// Requirement 5.6: System mixes multiple WAV streams into a single audio output.
type WAVPlayer struct {
	// Ebitengine/audio context (shared with MIDI player)
	audioCtx *audio.Context

	// Active players - Ebitengine/audio handles automatic mixing
	// Requirement 5.6: System mixes multiple WAV streams into a single audio output.
	players []*audio.Player

	// State
	muted bool

	// Mutex for thread-safe access
	mu sync.Mutex
}

// NewWAVPlayer creates a new WAV player with the specified audio context.
// The audio context should be shared with other audio components (e.g., MIDI player)
// to enable automatic mixing by Ebitengine/audio.
//
// Parameters:
//   - audioCtx: Ebitengine audio context (can be nil, will be created if needed)
//
// Returns:
//   - *WAVPlayer: The initialized WAV player
func NewWAVPlayer(audioCtx *audio.Context) *WAVPlayer {
	// Create audio context if not provided
	if audioCtx == nil {
		audioCtx = audio.NewContext(SampleRate)
	}

	return &WAVPlayer{
		audioCtx: audioCtx,
		players:  make([]*audio.Player, 0),
		muted:    false,
	}
}

// Play starts playback of the specified WAV file.
// Multiple WAV files can be played simultaneously; Ebitengine/audio handles mixing.
//
// Requirement 5.1: When PlayWAVE(filename) is called, system starts playback of specified WAV file.
// Requirement 5.2: When multiple PlayWAVE calls are made, system plays all WAV files simultaneously.
// Requirement 5.3: System supports standard WAV file formats (PCM, 8-bit, 16-bit).
// Requirement 5.4: When WAV file is not found, system logs error and continues execution.
// Requirement 5.5: When WAV file is corrupted, system logs error and continues execution.
//
// Parameters:
//   - filename: Path to the WAV file to play
//
// Returns:
//   - error: Error if the file cannot be loaded or played (caller should log and continue)
func (wp *WAVPlayer) Play(filename string) error {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Clean up finished players before adding new ones
	wp.cleanupFinishedPlayers()

	// Load WAV file
	// Requirement 5.4: When WAV file is not found, system logs error and continues execution.
	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: %s", ErrWAVFileNotFound, filename)
		}
		return fmt.Errorf("failed to read WAV file: %w", err)
	}

	// Decode WAV file
	// Requirement 5.3: System supports standard WAV file formats (PCM, 8-bit, 16-bit).
	// Requirement 5.5: When WAV file is corrupted, system logs error and continues execution.
	stream, err := wav.DecodeWithSampleRate(SampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrWAVInvalidFormat, err)
	}

	// Create audio player
	// Requirement 5.2: When multiple PlayWAVE calls are made, system plays all WAV files simultaneously.
	// Requirement 5.6: System mixes multiple WAV streams into a single audio output.
	// Ebitengine/audio automatically mixes multiple players
	player, err := wp.audioCtx.NewPlayer(stream)
	if err != nil {
		return fmt.Errorf("failed to create audio player: %w", err)
	}

	// Set volume based on muted state
	if wp.muted {
		player.SetVolume(0)
	}

	// Start playback
	player.Play()

	// Add to active players list
	wp.players = append(wp.players, player)

	return nil
}

// SetMuted sets the muted state of the WAV player.
// When muted, all current and future WAV playback will be silent.
//
// Requirement 12.2: When headless mode is enabled, system mutes all audio output.
//
// Parameters:
//   - muted: true to mute, false to unmute
func (wp *WAVPlayer) SetMuted(muted bool) {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	wp.muted = muted

	// Update volume for all active players
	for _, player := range wp.players {
		if player != nil {
			if muted {
				player.SetVolume(0)
			} else {
				player.SetVolume(1)
			}
		}
	}
}

// IsMuted returns whether the WAV player is muted.
func (wp *WAVPlayer) IsMuted() bool {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	return wp.muted
}

// StopAll stops all active WAV playback.
func (wp *WAVPlayer) StopAll() {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	for _, player := range wp.players {
		if player != nil {
			player.Close()
		}
	}
	wp.players = make([]*audio.Player, 0)
}

// GetActivePlayerCount returns the number of active WAV players.
// This is useful for testing and debugging.
func (wp *WAVPlayer) GetActivePlayerCount() int {
	wp.mu.Lock()
	defer wp.mu.Unlock()

	// Clean up finished players first
	wp.cleanupFinishedPlayers()

	return len(wp.players)
}

// cleanupFinishedPlayers removes players that have finished playing.
// Must be called with wp.mu held.
func (wp *WAVPlayer) cleanupFinishedPlayers() {
	activePlayers := make([]*audio.Player, 0, len(wp.players))
	for _, player := range wp.players {
		if player != nil && player.IsPlaying() {
			activePlayers = append(activePlayers, player)
		} else if player != nil {
			// Close finished player to release resources
			player.Close()
		}
	}
	wp.players = activePlayers
}

// Update is called from the game loop to perform periodic cleanup.
// This removes finished players from the active list.
func (wp *WAVPlayer) Update() {
	wp.mu.Lock()
	defer wp.mu.Unlock()
	wp.cleanupFinishedPlayers()
}

// GetAudioContext returns the audio context used by this player.
// This can be used to share the context with other audio components.
func (wp *WAVPlayer) GetAudioContext() *audio.Context {
	return wp.audioCtx
}
