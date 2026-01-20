package engine

import (
	"bytes"
	"fmt"
	"io"
	"sync"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

// WAVPlayer manages WAV file playback with concurrent playback support.
type WAVPlayer struct {
	engine         *Engine
	players        map[int]*audio.Player // Active players by ID
	resources      map[string][]byte     // Preloaded WAV resources
	nextPlayerID   int
	nextResourceID int
	mutex          sync.Mutex
}

// NewWAVPlayer creates a new WAV player.
func NewWAVPlayer(engine *Engine) *WAVPlayer {
	return &WAVPlayer{
		engine:    engine,
		players:   make(map[int]*audio.Player),
		resources: make(map[string][]byte),
	}
}

// PlayWAVE plays a WAV file asynchronously.
// Returns immediately, allowing concurrent playback.
func (wp *WAVPlayer) PlayWAVE(filename string) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	// Load WAV file via AssetLoader
	data, err := wp.engine.state.assetLoader.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to load WAV file %s: %w", filename, err)
	}

	// Decode WAV
	stream, err := wav.DecodeWithSampleRate(MIDISampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode WAV file %s: %w", filename, err)
	}

	// Create audio player
	player, err := getAudioContext().NewPlayer(stream)
	if err != nil {
		return fmt.Errorf("failed to create audio player for %s: %w", filename, err)
	}

	// Mute in headless mode
	if wp.engine.IsHeadless() {
		player.SetVolume(0)
	}

	// Assign player ID and store
	playerID := wp.nextPlayerID
	wp.nextPlayerID++
	wp.players[playerID] = player

	// Start playback in goroutine
	go func() {
		player.Play()
		wp.engine.logger.LogDebug("WAV playback started: %s (player %d)", filename, playerID)

		// Wait for playback to complete, then cleanup
		// Note: We can't easily detect when playback finishes with Ebiten's API,
		// so we'll rely on manual cleanup or let the player be garbage collected
	}()

	wp.engine.logger.LogInfo("Playing WAV: %s", filename)
	return nil
}

// LoadRsc preloads a WAV file into memory for fast playback.
// Returns a resource ID that can be used with PlayRsc.
func (wp *WAVPlayer) LoadRsc(filename string) (int, error) {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	// Load WAV file via AssetLoader
	data, err := wp.engine.state.assetLoader.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to load WAV resource %s: %w", filename, err)
	}

	// Assign resource ID and store
	resourceID := wp.nextResourceID
	wp.nextResourceID++
	resourceKey := fmt.Sprintf("rsc_%d", resourceID)
	wp.resources[resourceKey] = data

	wp.engine.logger.LogInfo("Loaded WAV resource: %s (ID: %d)", filename, resourceID)
	return resourceID, nil
}

// PlayRsc plays a preloaded WAV resource.
// This is faster than PlayWAVE since the file is already in memory.
func (wp *WAVPlayer) PlayRsc(resourceID int) error {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	// Get resource data
	resourceKey := fmt.Sprintf("rsc_%d", resourceID)
	data, ok := wp.resources[resourceKey]
	if !ok {
		return fmt.Errorf("resource ID %d not found", resourceID)
	}

	// Decode WAV
	stream, err := wav.DecodeWithSampleRate(MIDISampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("failed to decode WAV resource %d: %w", resourceID, err)
	}

	// Create audio player
	player, err := getAudioContext().NewPlayer(stream)
	if err != nil {
		return fmt.Errorf("failed to create audio player for resource %d: %w", resourceID, err)
	}

	// Mute in headless mode
	if wp.engine.IsHeadless() {
		player.SetVolume(0)
	}

	// Assign player ID and store
	playerID := wp.nextPlayerID
	wp.nextPlayerID++
	wp.players[playerID] = player

	// Start playback in goroutine
	go func() {
		player.Play()
		wp.engine.logger.LogDebug("WAV resource playback started: %d (player %d)", resourceID, playerID)
	}()

	wp.engine.logger.LogInfo("Playing WAV resource: %d", resourceID)
	return nil
}

// DelRsc deletes a preloaded WAV resource.
func (wp *WAVPlayer) DelRsc(resourceID int) {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	resourceKey := fmt.Sprintf("rsc_%d", resourceID)
	delete(wp.resources, resourceKey)
	wp.engine.logger.LogInfo("Deleted WAV resource: %d", resourceID)
}

// StopAll stops all active WAV players.
func (wp *WAVPlayer) StopAll() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	for id, player := range wp.players {
		player.Close()
		delete(wp.players, id)
	}
	wp.engine.logger.LogInfo("Stopped all WAV players")
}

// Cleanup removes finished players from the active list.
// This should be called periodically to prevent memory leaks.
func (wp *WAVPlayer) Cleanup() {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()

	for id, player := range wp.players {
		if !player.IsPlaying() {
			player.Close()
			delete(wp.players, id)
		}
	}
}

// GetActivePlayerCount returns the number of active WAV players.
func (wp *WAVPlayer) GetActivePlayerCount() int {
	wp.mutex.Lock()
	defer wp.mutex.Unlock()
	return len(wp.players)
}

// WAVStream wraps an io.Reader to implement io.ReadSeeker for WAV playback.
// This is needed because Ebiten's audio player requires io.ReadSeeker.
type WAVStream struct {
	data   []byte
	offset int64
}

// NewWAVStream creates a new WAV stream from data.
func NewWAVStream(data []byte) *WAVStream {
	return &WAVStream{
		data:   data,
		offset: 0,
	}
}

// Read reads data from the stream.
func (ws *WAVStream) Read(p []byte) (int, error) {
	if ws.offset >= int64(len(ws.data)) {
		return 0, io.EOF
	}

	n := copy(p, ws.data[ws.offset:])
	ws.offset += int64(n)
	return n, nil
}

// Seek seeks to a position in the stream.
func (ws *WAVStream) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64

	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = ws.offset + offset
	case io.SeekEnd:
		newOffset = int64(len(ws.data)) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative position")
	}

	ws.offset = newOffset
	return newOffset, nil
}
