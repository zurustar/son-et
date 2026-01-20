package engine

import (
	"bytes"
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Virtual desktop dimensions (fixed at 1280×720)
const (
	VirtualDesktopWidth  = 1280
	VirtualDesktopHeight = 720
)

// FunctionDefinition represents a user-defined function.
type FunctionDefinition struct {
	Name       string               // Function name (lowercase)
	Parameters []string             // Parameter names
	Body       []interpreter.OpCode // Function body as OpCode sequence
}

// EngineState holds all runtime state for the FILLY engine.
// It is the central data structure that contains graphics, audio, and execution state.
type EngineState struct {
	// Graphics state
	pictures map[int]*Picture // Picture ID -> Picture
	windows  map[int]*Window  // Window ID -> Window
	casts    map[int]*Cast    // Cast ID -> Cast

	// Execution state
	sequencers    []*Sequencer                   // Active sequences
	eventHandlers []*EventHandler                // Registered event handlers
	functions     map[string]*FunctionDefinition // User-defined functions (lowercase names)
	nextSeqID     int                            // Next sequence ID to assign
	nextGroupID   int                            // Next group ID to assign
	nextHandlerID int                            // Next event handler ID to assign
	nextPicID     int                            // Next picture ID to assign
	tickCount     int64                          // Global tick counter

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
		pictures:      make(map[int]*Picture),
		windows:       make(map[int]*Window),
		casts:         make(map[int]*Cast),
		sequencers:    make([]*Sequencer, 0),
		eventHandlers: make([]*EventHandler, 0),
		functions:     make(map[string]*FunctionDefinition),
		nextSeqID:     1,
		nextGroupID:   1,
		nextHandlerID: 1,
		nextPicID:     1,
		tickCount:     0,
		renderer:      renderer,
		assetLoader:   assetLoader,
		imageDecoder:  imageDecoder,
		headlessMode:  false,
		debugLevel:    0,
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

// GetDesktopWidth returns the virtual desktop width (always 1280).
func (e *EngineState) GetDesktopWidth() int {
	return VirtualDesktopWidth
}

// GetDesktopHeight returns the virtual desktop height (always 720).
func (e *EngineState) GetDesktopHeight() int {
	return VirtualDesktopHeight
}

// RegisterSequence registers a new sequence with the engine.
// Returns the sequence ID.
func (e *EngineState) RegisterSequence(seq *Sequencer, groupID int) int {
	// Assign sequence ID
	seq.SetID(e.nextSeqID)
	e.nextSeqID++

	// Assign group ID (0 means no group)
	if groupID > 0 {
		seq.SetGroupID(groupID)
	}

	// Add to active sequences
	e.sequencers = append(e.sequencers, seq)

	return seq.GetID()
}

// AllocateGroupID allocates a new group ID.
func (e *EngineState) AllocateGroupID() int {
	id := e.nextGroupID
	e.nextGroupID++
	return id
}

// GetSequencers returns all active sequencers.
func (e *EngineState) GetSequencers() []*Sequencer {
	return e.sequencers
}

// DeactivateSequence deactivates a sequence by ID.
func (e *EngineState) DeactivateSequence(id int) {
	for _, seq := range e.sequencers {
		if seq.GetID() == id {
			seq.Deactivate()
			return
		}
	}
}

// DeactivateGroup deactivates all sequences in a group.
func (e *EngineState) DeactivateGroup(groupID int) {
	for _, seq := range e.sequencers {
		if seq.GetGroupID() == groupID {
			seq.Deactivate()
		}
	}
}

// DeactivateAll deactivates all sequences.
func (e *EngineState) DeactivateAll() {
	for _, seq := range e.sequencers {
		seq.Deactivate()
	}
}

// CleanupInactiveSequences removes inactive sequences from the list.
func (e *EngineState) CleanupInactiveSequences() {
	active := make([]*Sequencer, 0, len(e.sequencers))
	for _, seq := range e.sequencers {
		if seq.IsActive() {
			active = append(active, seq)
		}
	}
	e.sequencers = active
}

// RegisterEventHandler registers a new event handler.
// Returns the handler ID.
func (e *EngineState) RegisterEventHandler(eventType EventType, commands []interpreter.OpCode, mode TimingMode, parent *Sequencer, userID int) int {
	handler := &EventHandler{
		ID:        e.nextHandlerID,
		EventType: eventType,
		Commands:  commands,
		Mode:      mode,
		Parent:    parent,
		Active:    true,
		UserID:    userID,
	}

	e.nextHandlerID++
	e.eventHandlers = append(e.eventHandlers, handler)

	return handler.ID
}

// GetEventHandlers returns all active event handlers for a given event type.
func (e *EngineState) GetEventHandlers(eventType EventType) []*EventHandler {
	handlers := make([]*EventHandler, 0)
	for _, handler := range e.eventHandlers {
		if handler.Active && handler.EventType == eventType {
			handlers = append(handlers, handler)
		}
	}
	return handlers
}

// GetUserEventHandlers returns all active USER event handlers for a specific user ID.
func (e *EngineState) GetUserEventHandlers(userID int) []*EventHandler {
	handlers := make([]*EventHandler, 0)
	for _, handler := range e.eventHandlers {
		if handler.Active && handler.EventType == EventUSER && handler.UserID == userID {
			handlers = append(handlers, handler)
		}
	}
	return handlers
}

// DeactivateEventHandler deactivates an event handler by ID.
func (e *EngineState) DeactivateEventHandler(id int) {
	for _, handler := range e.eventHandlers {
		if handler.ID == id {
			handler.Active = false
			return
		}
	}
}

// CleanupInactiveEventHandlers removes inactive event handlers from the list.
func (e *EngineState) CleanupInactiveEventHandlers() {
	active := make([]*EventHandler, 0, len(e.eventHandlers))
	for _, handler := range e.eventHandlers {
		if handler.Active {
			active = append(active, handler)
		}
	}
	e.eventHandlers = active
}

// RegisterFunction registers a user-defined function.
func (e *EngineState) RegisterFunction(name string, parameters []string, body []interpreter.OpCode) {
	// Convert function name to lowercase for case-insensitive lookup
	lowerName := strings.ToLower(name)

	e.functions[lowerName] = &FunctionDefinition{
		Name:       lowerName,
		Parameters: parameters,
		Body:       body,
	}
}

// GetFunction retrieves a user-defined function by name (case-insensitive).
func (e *EngineState) GetFunction(name string) (*FunctionDefinition, bool) {
	lowerName := strings.ToLower(name)
	fn, ok := e.functions[lowerName]
	return fn, ok
}

// LoadPicture loads an image from a file and assigns it a picture ID.
// Returns the picture ID, or an error if loading fails.
func (e *EngineState) LoadPicture(filename string) (int, error) {
	// Load file data via AssetLoader
	data, err := e.assetLoader.ReadFile(filename)
	if err != nil {
		return 0, fmt.Errorf("failed to load picture %s: %w", filename, err)
	}

	// Create reader from byte slice
	reader := bytes.NewReader(data)

	// Decode image
	img, _, err := e.imageDecoder.Decode(reader)
	if err != nil {
		return 0, fmt.Errorf("failed to decode picture %s: %w", filename, err)
	}

	// Create picture
	bounds := img.Bounds()
	pic := &Picture{
		ID:     e.nextPicID,
		Image:  img,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}

	// Store picture
	e.pictures[pic.ID] = pic
	e.nextPicID++

	return pic.ID, nil
}

// CreatePicture creates an empty image buffer with the specified dimensions.
// Returns the picture ID.
func (e *EngineState) CreatePicture(width, height int) int {
	// Create empty RGBA image
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Create picture
	pic := &Picture{
		ID:     e.nextPicID,
		Image:  img,
		Width:  width,
		Height: height,
	}

	// Store picture
	e.pictures[pic.ID] = pic
	e.nextPicID++

	return pic.ID
}

// GetPicture retrieves a picture by ID.
// Returns nil if the picture doesn't exist.
func (e *EngineState) GetPicture(id int) *Picture {
	return e.pictures[id]
}

// DeletePicture removes a picture and releases its resources.
func (e *EngineState) DeletePicture(id int) {
	delete(e.pictures, id)
}

// GetPictureWidth returns the width of a picture.
// Returns 0 if the picture doesn't exist.
func (e *EngineState) GetPictureWidth(id int) int {
	if pic := e.pictures[id]; pic != nil {
		return pic.Width
	}
	return 0
}

// GetPictureHeight returns the height of a picture.
// Returns 0 if the picture doesn't exist.
func (e *EngineState) GetPictureHeight(id int) int {
	if pic := e.pictures[id]; pic != nil {
		return pic.Height
	}
	return 0
}
