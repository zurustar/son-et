package engine

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
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

	// Window drag state
	draggedWindowID int // ID of window being dragged (0 = none)

	// Execution state
	sequencers    []*Sequencer                   // Active sequences
	eventHandlers []*EventHandler                // Registered event handlers
	functions     map[string]*FunctionDefinition // User-defined functions (lowercase names)
	nextSeqID     int                            // Next sequence ID to assign
	nextGroupID   int                            // Next group ID to assign
	nextHandlerID int                            // Next event handler ID to assign
	nextPicID     int                            // Next picture ID to assign
	nextWinID     int                            // Next window ID to assign
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

	// Drag state
	IsDragging  bool // True if window is being dragged
	DragOffsetX int  // Mouse offset from window X when drag started
	DragOffsetY int  // Mouse offset from window Y when drag started
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
		nextWinID:     1,
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

// MovePicture copies pixels from source picture to destination picture with transparency.
// Parameters:
//
//	srcID: source picture ID
//	srcX, srcY: source rectangle top-left corner
//	srcW, srcH: source rectangle dimensions
//	dstID: destination picture ID
//	dstX, dstY: destination position
//
// Returns an error if source or destination doesn't exist.
func (e *EngineState) MovePicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY int) error {
	// Get source picture
	srcPic := e.pictures[srcID]
	if srcPic == nil {
		return fmt.Errorf("source picture %d not found", srcID)
	}

	// Get destination picture
	dstPic := e.pictures[dstID]
	if dstPic == nil {
		return fmt.Errorf("destination picture %d not found", dstID)
	}

	// Cannot copy to itself
	if srcID == dstID {
		return fmt.Errorf("cannot copy picture to itself (ID %d)", srcID)
	}

	// Auto-expand destination if needed
	requiredW := dstX + srcW
	requiredH := dstY + srcH
	if requiredW > dstPic.Width || requiredH > dstPic.Height {
		newW := dstPic.Width
		newH := dstPic.Height
		if requiredW > newW {
			newW = requiredW
		}
		if requiredH > newH {
			newH = requiredH
		}

		// Create new larger image
		newImg := image.NewRGBA(image.Rect(0, 0, newW, newH))

		// Copy old content
		draw.Draw(newImg, newImg.Bounds(), dstPic.Image, image.Point{}, draw.Src)

		// Update picture
		dstPic.Image = newImg
		dstPic.Width = newW
		dstPic.Height = newH
	}

	// Copy pixels with alpha blending (transparency support)
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	dstPoint := image.Point{dstX, dstY}

	// Use draw.Over for alpha blending (respects transparency)
	draw.Draw(dstPic.Image.(draw.Image), image.Rectangle{dstPoint, dstPoint.Add(srcRect.Size())},
		srcPic.Image, srcRect.Min, draw.Over)

	return nil
}

// MoveSPicture copies and scales pixels from source picture to destination picture with transparency.
// Uses nearest-neighbor scaling for simplicity.
// Parameters:
//
//	srcID: source picture ID
//	srcX, srcY: source rectangle top-left corner
//	srcW, srcH: source rectangle dimensions
//	dstID: destination picture ID
//	dstX, dstY: destination position
//	dstW, dstH: destination dimensions (scaled size)
//
// Returns an error if source or destination doesn't exist.
func (e *EngineState) MoveSPicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH int) error {
	// Get source picture
	srcPic := e.pictures[srcID]
	if srcPic == nil {
		return fmt.Errorf("source picture %d not found", srcID)
	}

	// Get destination picture
	dstPic := e.pictures[dstID]
	if dstPic == nil {
		return fmt.Errorf("destination picture %d not found", dstID)
	}

	// Cannot copy to itself
	if srcID == dstID {
		return fmt.Errorf("cannot copy picture to itself (ID %d)", srcID)
	}

	// Auto-expand destination if needed
	requiredW := dstX + dstW
	requiredH := dstY + dstH
	if requiredW > dstPic.Width || requiredH > dstPic.Height {
		newW := dstPic.Width
		newH := dstPic.Height
		if requiredW > newW {
			newW = requiredW
		}
		if requiredH > newH {
			newH = requiredH
		}

		// Create new larger image
		newImg := image.NewRGBA(image.Rect(0, 0, newW, newH))

		// Copy old content
		draw.Draw(newImg, newImg.Bounds(), dstPic.Image, image.Point{}, draw.Src)

		// Update picture
		dstPic.Image = newImg
		dstPic.Width = newW
		dstPic.Height = newH
	}

	// Nearest-neighbor scaling
	scaleX := float64(srcW) / float64(dstW)
	scaleY := float64(srcH) / float64(dstH)

	dstImg := dstPic.Image.(*image.RGBA)

	for dy := 0; dy < dstH; dy++ {
		for dx := 0; dx < dstW; dx++ {
			// Map destination pixel to source pixel
			sx := int(float64(dx) * scaleX)
			sy := int(float64(dy) * scaleY)

			// Get source pixel
			srcColor := srcPic.Image.At(srcX+sx, srcY+sy)

			// Set destination pixel (with alpha blending)
			dstImg.Set(dstX+dx, dstY+dy, srcColor)
		}
	}

	return nil
}

// ReversePicture copies and horizontally flips pixels from source picture to destination picture.
// Parameters:
//
//	srcID: source picture ID
//	srcX, srcY: source rectangle top-left corner
//	srcW, srcH: source rectangle dimensions
//	dstID: destination picture ID
//	dstX, dstY: destination position
//
// Returns an error if source or destination doesn't exist.
func (e *EngineState) ReversePicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY int) error {
	// Get source picture
	srcPic := e.pictures[srcID]
	if srcPic == nil {
		return fmt.Errorf("source picture %d not found", srcID)
	}

	// Get destination picture
	dstPic := e.pictures[dstID]
	if dstPic == nil {
		return fmt.Errorf("destination picture %d not found", dstID)
	}

	// Cannot copy to itself
	if srcID == dstID {
		return fmt.Errorf("cannot copy picture to itself (ID %d)", srcID)
	}

	// Auto-expand destination if needed
	requiredW := dstX + srcW
	requiredH := dstY + srcH
	if requiredW > dstPic.Width || requiredH > dstPic.Height {
		newW := dstPic.Width
		newH := dstPic.Height
		if requiredW > newW {
			newW = requiredW
		}
		if requiredH > newH {
			newH = requiredH
		}

		// Create new larger image
		newImg := image.NewRGBA(image.Rect(0, 0, newW, newH))

		// Copy old content
		draw.Draw(newImg, newImg.Bounds(), dstPic.Image, image.Point{}, draw.Src)

		// Update picture
		dstPic.Image = newImg
		dstPic.Width = newW
		dstPic.Height = newH
	}

	// Horizontal flip: copy pixels from right to left
	dstImg := dstPic.Image.(*image.RGBA)

	for sy := 0; sy < srcH; sy++ {
		for sx := 0; sx < srcW; sx++ {
			// Get source pixel
			srcColor := srcPic.Image.At(srcX+sx, srcY+sy)

			// Set destination pixel (flipped horizontally)
			// When sx=0, we want to write to dstX+srcW-1
			// When sx=srcW-1, we want to write to dstX
			dstImg.Set(dstX+srcW-1-sx, dstY+sy, srcColor)
		}
	}

	return nil
}

// OpenWindow creates a new window and returns its ID.
// Parameters:
//
//	picID: picture to display in the window
//	x, y: window position on virtual desktop
//	width, height: window dimensions (0,0 = use picture size)
//	picX, picY: offset into the picture to display
//	color: background color (0xRRGGBB format, currently unused)
func (e *EngineState) OpenWindow(picID, x, y, width, height, picX, picY, color int) int {
	// If width and height are both 0, use picture dimensions
	if width == 0 && height == 0 {
		if pic := e.pictures[picID]; pic != nil {
			width = pic.Width
			height = pic.Height
		} else {
			// Fallback to default size
			width = 640
			height = 480
		}
	}

	// Create window
	win := &Window{
		ID:        e.nextWinID,
		PictureID: picID,
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		PicX:      picX,
		PicY:      picY,
		Caption:   "", // No caption by default (can be set with CapTitle)
		Visible:   true,
	}

	// Store window
	e.windows[win.ID] = win
	e.nextWinID++

	return win.ID
}

// GetWindow retrieves a window by ID.
// Returns nil if the window doesn't exist.
func (e *EngineState) GetWindow(id int) *Window {
	return e.windows[id]
}

// MoveWindow updates window properties.
// Parameters:
//
//	id: window ID
//	x, y: new window position
//	width, height: new window dimensions
//	picX, picY: new picture offset
func (e *EngineState) MoveWindow(id, x, y, width, height, picX, picY int) error {
	win := e.windows[id]
	if win == nil {
		return fmt.Errorf("window %d not found", id)
	}

	win.X = x
	win.Y = y
	win.Width = width
	win.Height = height
	win.PicX = picX
	win.PicY = picY

	return nil
}

// CloseWindow closes a window and removes it.
func (e *EngineState) CloseWindow(id int) {
	delete(e.windows, id)
}

// CloseAllWindows closes all windows.
func (e *EngineState) CloseAllWindows() {
	e.windows = make(map[int]*Window)
}

// SetWindowCaption sets the caption (title) of a window.
func (e *EngineState) SetWindowCaption(id int, caption string) error {
	win := e.windows[id]
	if win == nil {
		return fmt.Errorf("window %d not found", id)
	}

	win.Caption = caption
	return nil
}

// GetWindowPictureID returns the picture ID associated with a window.
// Returns 0 if the window doesn't exist.
func (e *EngineState) GetWindowPictureID(id int) int {
	if win := e.windows[id]; win != nil {
		return win.PictureID
	}
	return 0
}

// GetWindows returns all windows in creation order.
// This is used for rendering windows in the correct z-order.
func (e *EngineState) GetWindows() []*Window {
	// Collect all windows
	windows := make([]*Window, 0, len(e.windows))
	for _, win := range e.windows {
		windows = append(windows, win)
	}

	// Sort by ID (creation order)
	// Using a simple bubble sort since the number of windows is typically small
	for i := 0; i < len(windows); i++ {
		for j := i + 1; j < len(windows); j++ {
			if windows[i].ID > windows[j].ID {
				windows[i], windows[j] = windows[j], windows[i]
			}
		}
	}

	return windows
}

// TitleBarHeight returns the height of the title bar in pixels.
// Windows with captions have a draggable title bar.
const TitleBarHeight = 20

// StartWindowDrag initiates dragging a window.
// Parameters:
//
//	mouseX, mouseY: current mouse position on virtual desktop
//
// Returns the window ID that started dragging, or 0 if no window was clicked.
func (e *EngineState) StartWindowDrag(mouseX, mouseY int) int {
	// Find the topmost window under the mouse cursor (reverse order = top to bottom)
	windows := e.GetWindows()
	for i := len(windows) - 1; i >= 0; i-- {
		win := windows[i]

		// Skip invisible windows
		if !win.Visible {
			continue
		}

		// Check if mouse is in title bar area (only if window has a caption)
		if win.Caption != "" {
			titleBarX := win.X
			titleBarY := win.Y
			titleBarWidth := win.Width
			titleBarHeight := TitleBarHeight

			if mouseX >= titleBarX && mouseX < titleBarX+titleBarWidth &&
				mouseY >= titleBarY && mouseY < titleBarY+titleBarHeight {
				// Start dragging this window
				win.IsDragging = true
				win.DragOffsetX = mouseX - win.X
				win.DragOffsetY = mouseY - win.Y
				e.draggedWindowID = win.ID
				return win.ID
			}
		}
	}

	return 0
}

// UpdateWindowDrag updates the position of the dragged window.
// Parameters:
//
//	mouseX, mouseY: current mouse position on virtual desktop
//
// Returns true if a window was updated, false if no window is being dragged.
func (e *EngineState) UpdateWindowDrag(mouseX, mouseY int) bool {
	if e.draggedWindowID == 0 {
		return false
	}

	win := e.windows[e.draggedWindowID]
	if win == nil || !win.IsDragging {
		e.draggedWindowID = 0
		return false
	}

	// Calculate new position
	newX := mouseX - win.DragOffsetX
	newY := mouseY - win.DragOffsetY

	// Constrain to virtual desktop bounds (Task 4.3.10)
	// Keep at least part of the title bar visible
	minX := -(win.Width - 50)                     // Allow dragging mostly off-screen to the left
	maxX := VirtualDesktopWidth - 50              // Keep at least 50px visible on the right
	minY := 0                                     // Don't allow dragging above the desktop
	maxY := VirtualDesktopHeight - TitleBarHeight // Keep title bar visible at bottom

	if newX < minX {
		newX = minX
	}
	if newX > maxX {
		newX = maxX
	}
	if newY < minY {
		newY = minY
	}
	if newY > maxY {
		newY = maxY
	}

	// Update window position
	win.X = newX
	win.Y = newY

	return true
}

// StopWindowDrag stops dragging the current window.
func (e *EngineState) StopWindowDrag() {
	if e.draggedWindowID != 0 {
		win := e.windows[e.draggedWindowID]
		if win != nil {
			win.IsDragging = false
		}
		e.draggedWindowID = 0
	}
}

// GetDraggedWindowID returns the ID of the window currently being dragged.
// Returns 0 if no window is being dragged.
func (e *EngineState) GetDraggedWindowID() int {
	return e.draggedWindowID
}
