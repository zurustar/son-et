package engine

import (
	"image"
	"sync"
)

// EngineState holds all runtime state for the FILLY engine.
// It is the central data structure that contains graphics, audio, and execution state.
type EngineState struct {
	// Graphics state
	pictures map[int]*Picture // Picture ID -> Picture
	windows  map[int]*Window  // Window ID -> Window
	casts    map[int]*Cast    // Cast ID -> Cast

	// Execution state
	sequencers []*Sequencer // Active sequences
	tickCount  int64        // Global tick counter

	// Dependencies (injected)
	renderer     Renderer     // Rendering abstraction
	assetLoader  AssetLoader  // Asset loading abstraction
	imageDecoder ImageDecoder // Image decoding abstraction

	// Synchronization
	renderMutex sync.Mutex // Protects graphics state from concurrent access

	// Configuration
	headlessMode bool // True if running without GUI
	debugLevel   int  // 0=errors, 1=info, 2=debug
}

// Picture represents a loaded or created image.
type Picture struct {
	ID     int         // Unique picture ID
	Image  image.Image // The actual image data
	Width  int         // Image width
	Height int         // Image height
}

// Window represents a virtual window on the desktop.
type Window struct {
	ID        int    // Unique window ID
	PictureID int    // Picture to display
	X         int    // Position X
	Y         int    // Position Y
	Width     int    // Window width
	Height    int    // Window height
	PicX      int    // Picture offset X
	PicY      int    // Picture offset Y
	Caption   string // Window caption (title bar)
	Visible   bool   // Is window visible
}

// Cast represents a sprite (movable image element).
type Cast struct {
	ID        int  // Unique cast ID
	PictureID int  // Picture to display
	WindowID  int  // Parent window
	X         int  // Position X (relative to window)
	Y         int  // Position Y (relative to window)
	SrcX      int  // Source clipping X
	SrcY      int  // Source clipping Y
	Width     int  // Clipping width
	Height    int  // Clipping height
	Visible   bool // Is cast visible
}

// NewEngineState creates a new engine state with the given dependencies.
func NewEngineState(renderer Renderer, assetLoader AssetLoader, imageDecoder ImageDecoder) *EngineState {
	return &EngineState{
		pictures:     make(map[int]*Picture),
		windows:      make(map[int]*Window),
		casts:        make(map[int]*Cast),
		sequencers:   make([]*Sequencer, 0),
		tickCount:    0,
		renderer:     renderer,
		assetLoader:  assetLoader,
		imageDecoder: imageDecoder,
		headlessMode: false,
		debugLevel:   0,
	}
}

// SetHeadlessMode enables or disables headless mode.
func (e *EngineState) SetHeadlessMode(enabled bool) {
	e.headlessMode = enabled
}

// SetDebugLevel sets the debug logging level.
// 0=errors only, 1=info, 2=debug
func (e *EngineState) SetDebugLevel(level int) {
	e.debugLevel = level
}

// GetTickCount returns the current global tick count.
func (e *EngineState) GetTickCount() int64 {
	return e.tickCount
}

// IncrementTick increments the global tick counter.
func (e *EngineState) IncrementTick() {
	e.tickCount++
}
