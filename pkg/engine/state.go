package engine

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
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
	draggedWindowID   int    // ID of window being dragged (0 = none)
	globalWindowTitle string // Default title for next window opened

	// Execution state
	sequencers    []*Sequencer                   // Active sequences
	eventHandlers []*EventHandler                // Registered event handlers
	functions     map[string]*FunctionDefinition // User-defined functions (lowercase names)
	nextSeqID     int                            // Next sequence ID to assign
	nextGroupID   int                            // Next group ID to assign
	nextHandlerID int                            // Next event handler ID to assign
	nextPicID     int                            // Next picture ID to assign
	nextWinID     int                            // Next window ID to assign
	nextCastID    int                            // Next cast ID to assign
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
	ID         int           // Unique picture ID
	Image      *ebiten.Image // The actual image data (Ebiten native)
	BackBuffer *ebiten.Image // Double buffer for cast rendering
	Width      int           // Image width
	Height     int           // Image height
}

// Window represents a virtual window on the desktop.
type Window struct {
	ID        int         // Unique window ID
	PictureID int         // Picture to display
	X         int         // Position X
	Y         int         // Position Y
	Width     int         // Window width
	Height    int         // Window height
	PicX      int         // Picture offset X
	PicY      int         // Picture offset Y
	Caption   string      // Window caption (title bar)
	Visible   bool        // Is window visible
	Color     color.Color // Background color (drawn behind picture content)

	// Drag state
	IsDragging  bool // True if window is being dragged
	DragOffsetX int  // Mouse offset from window X when drag started
	DragOffsetY int  // Mouse offset from window Y when drag started
}

// Cast represents a sprite (movable image element).
type Cast struct {
	ID               int           // Unique cast ID
	PictureID        int           // Picture to display
	WindowID         int           // Parent window
	X                int           // Position X (relative to window)
	Y                int           // Position Y (relative to window)
	SrcX             int           // Source clipping X
	SrcY             int           // Source clipping Y
	Width            int           // Clipping width
	Height           int           // Clipping height
	TransparentColor int           // Transparent color (0xRRGGBB format, -1 = no transparency)
	Visible          bool          // Is cast visible
	ProcessedImage   *ebiten.Image // Pre-processed image with transparency applied (cached)
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
		nextPicID:     0, // Start from 0 for FILLY compatibility
		nextWinID:     0, // Start from 0 for FILLY compatibility
		nextCastID:    1,
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

	// Convert to Ebiten image immediately
	ebitenImg := ebiten.NewImageFromImage(img)
	bounds := img.Bounds()

	// Create picture
	pic := &Picture{
		ID:     e.nextPicID,
		Image:  ebitenImg,
		Width:  bounds.Dx(),
		Height: bounds.Dy(),
	}

	// Store picture
	e.pictures[pic.ID] = pic
	e.nextPicID++

	if e.debugLevel >= 2 {
		fmt.Printf("[DEBUG] LoadPicture: %s -> ID=%d, size=(%dx%d), bounds=%v\n",
			filename, pic.ID, pic.Width, pic.Height, bounds)
	}

	return pic.ID, nil
}

// CreatePicture creates an empty image buffer with the specified dimensions.
// Returns the picture ID.
func (e *EngineState) CreatePicture(width, height int) int {
	// Create empty Ebiten image
	img := ebiten.NewImage(width, height)

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

// EnsureRGBA converts an image to *image.RGBA if it isn't already.
// This is needed for drawing operations that require direct pixel access.
func (e *EngineState) EnsureRGBA(img image.Image) *image.RGBA {
	// If already RGBA, return as-is
	if rgba, ok := img.(*image.RGBA); ok {
		return rgba
	}

	// Convert to RGBA
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}
	return rgba
}

// MovePicture copies pixels from source picture to destination picture with transparency.
// Parameters:
//
//	srcID: source picture ID
//	srcX, srcY: source rectangle top-left corner
//	srcW, srcH: source rectangle dimensions
//	dstID: destination picture ID
//	dstX, dstY: destination position
//	mode: transfer mode (0=normal, 1=transparent, 2=scene change)
//
// Returns an error if source or destination doesn't exist.
func (e *EngineState) MovePicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, mode int) error {
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

		// Create new larger Ebiten image
		newImg := ebiten.NewImage(newW, newH)

		// Copy old content using Ebiten
		opts := &ebiten.DrawImageOptions{}
		newImg.DrawImage(dstPic.Image, opts)

		// Update picture
		dstPic.Image = newImg
		dstPic.Width = newW
		dstPic.Height = newH
	}

	// Clip source rectangle to source image bounds
	// If srcW or srcH exceed the source image, we need to clip them
	actualSrcW := srcW
	actualSrcH := srcH

	// Ensure we don't read beyond source image bounds
	if srcX+srcW > srcPic.Width {
		actualSrcW = srcPic.Width - srcX
		if actualSrcW < 0 {
			actualSrcW = 0
		}
	}
	if srcY+srcH > srcPic.Height {
		actualSrcH = srcPic.Height - srcY
		if actualSrcH < 0 {
			actualSrcH = 0
		}
	}

	if e.debugLevel >= 2 {
		fmt.Printf("[DEBUG] MovePicture: src=%d (%d,%d,%d,%d) actual=(%d,%d) dst=%d (%d,%d)\n",
			srcID, srcX, srcY, srcW, srcH, actualSrcW, actualSrcH, dstID, dstX, dstY)
	}

	// If nothing to copy, return early
	if actualSrcW <= 0 || actualSrcH <= 0 {
		return nil
	}

	// Use Ebiten's SubImage and DrawImage for efficient copying with alpha blending
	srcRect := image.Rect(srcX, srcY, srcX+actualSrcW, srcY+actualSrcH)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(dstX), float64(dstY))
	dstPic.Image.DrawImage(subImg, opts)

	return nil
}

// MoveSPicture copies and scales pixels from source picture to destination picture with transparency.
// Uses Ebiten's scaling for efficient rendering.
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

		// Create new larger Ebiten image
		newImg := ebiten.NewImage(newW, newH)

		// Copy old content
		opts := &ebiten.DrawImageOptions{}
		newImg.DrawImage(dstPic.Image, opts)

		// Update picture
		dstPic.Image = newImg
		dstPic.Width = newW
		dstPic.Height = newH
	}

	// Extract source region
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// Calculate scale factors
	scaleX := float64(dstW) / float64(srcW)
	scaleY := float64(dstH) / float64(srcH)

	// Draw with scaling
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(scaleX, scaleY)
	opts.GeoM.Translate(float64(dstX), float64(dstY))
	dstPic.Image.DrawImage(subImg, opts)

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

		// Create new larger Ebiten image
		newImg := ebiten.NewImage(newW, newH)

		// Copy old content
		opts := &ebiten.DrawImageOptions{}
		newImg.DrawImage(dstPic.Image, opts)

		// Update picture
		dstPic.Image = newImg
		dstPic.Width = newW
		dstPic.Height = newH
	}

	// Extract source region
	srcRect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// Horizontal flip using GeoM
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(-1, 1)                                 // Flip horizontally
	opts.GeoM.Translate(float64(dstX+srcW), float64(dstY)) // Translate to destination
	dstPic.Image.DrawImage(subImg, opts)

	return nil
}

// OpenWindow creates a new window and returns its ID.
// Parameters:
//
//	picID: picture to display in the window
//	x, y: window position on virtual desktop
//	width, height: window dimensions (0,0 = use picture size)
//	picX, picY: offset into the picture to display
//	bgColor: background color (0xRRGGBB format)
func (e *EngineState) OpenWindow(picID, x, y, width, height, picX, picY, bgColor int) int {
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

	// Convert background color from 0xRRGGBB format to color.Color
	// The background color is stored in the window and drawn by the renderer
	// BEFORE the picture content, so transparent areas in the picture show the background
	var winColor color.Color
	if bgColor >= 0 {
		r := uint8((bgColor >> 16) & 0xFF)
		g := uint8((bgColor >> 8) & 0xFF)
		b := uint8(bgColor & 0xFF)
		winColor = color.RGBA{R: r, G: g, B: b, A: 255}
	} else {
		// Default to white background
		winColor = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	}

	// Create window
	win := &Window{
		ID:        e.nextWinID,
		PictureID: picID,
		X:         x,
		Y:         y,
		Width:     width,
		Height:    height,
		PicX:      -picX,               // Inverted for legacy compatibility
		PicY:      -picY,               // Inverted for legacy compatibility
		Caption:   e.globalWindowTitle, // Use global title if set
		Visible:   true,
		Color:     winColor,
	}

	// Store window
	e.windows[win.ID] = win
	e.nextWinID++

	// Clear global title after using it
	e.globalWindowTitle = ""

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
//	picID: new picture ID
//	x, y: new window position
//	width, height: new window dimensions (0 = keep current)
//	picX, picY: new picture offset
func (e *EngineState) MoveWindow(id, picID, x, y, width, height, picX, picY int) error {
	win := e.windows[id]
	if win == nil {
		return fmt.Errorf("window %d not found", id)
	}

	// Update picture ID
	win.PictureID = picID

	// Update position (0,0 is valid, so always update)
	win.X = x
	win.Y = y

	// Update size only if non-zero (0 means keep current size)
	if width != 0 {
		win.Width = width
	}
	if height != 0 {
		win.Height = height
	}

	// Update picture offset
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
// If the window doesn't exist yet, sets the global default title for the next window.
func (e *EngineState) SetWindowCaption(id int, caption string) error {
	win := e.windows[id]
	if win == nil {
		// Window doesn't exist yet - set global title for next window
		e.globalWindowTitle = caption
		return nil
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
			// Title bar is at the top of the window, after the border
			titleBarX := win.X + BorderThickness
			titleBarY := win.Y + BorderThickness
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

// PutCast creates a new cast (sprite) and returns its ID.
// In FILLY, PutCast not only creates a cast object but also draws it directly
// to the destination picture.
// Parameters:
//
//	destPicID: destination picture ID (where to draw the cast)
//	picID: source picture to display
//	x, y: position on destination picture
//	srcX, srcY: source clipping position in picture
//	width, height: clipping dimensions
//	transparentColor: color to treat as transparent (0xRRGGBB format, -1 = no transparency)
//
// Note: The cast is associated with the window that uses destPicID.
func (e *EngineState) PutCast(destPicID, picID, x, y, srcX, srcY, width, height, transparentColor int) int {
	// Get source picture
	srcPic := e.pictures[picID]
	if srcPic == nil {
		if e.debugLevel >= 1 {
			fmt.Printf("[ERROR] PutCast: source picture %d not found\n", picID)
		}
		return 0
	}

	// Pre-process transparency if needed
	var processedImage *ebiten.Image
	if transparentColor >= 0 {
		processedImage = e.createTransparentImage(srcPic, srcX, srcY, width, height, transparentColor)
	}

	// Create cast
	cast := &Cast{
		ID:               e.nextCastID,
		PictureID:        picID,
		WindowID:         destPicID, // Store destPicID as WindowID for now
		X:                x,
		Y:                y,
		SrcX:             srcX,
		SrcY:             srcY,
		Width:            width,
		Height:           height,
		TransparentColor: transparentColor,
		Visible:          true,
		ProcessedImage:   processedImage, // Store pre-processed image
	}

	// Store cast
	e.casts[cast.ID] = cast
	e.nextCastID++

	// Draw cast to destination picture immediately (FILLY behavior)
	e.drawCastToPicture(cast, destPicID, transparentColor)

	return cast.ID
}

// createTransparentImage creates a new Ebiten image with transparency applied.
// This is done once at PutCast time to avoid repeated pixel-by-pixel processing.
func (e *EngineState) createTransparentImage(srcPic *Picture, srcX, srcY, width, height, transparentColor int) *ebiten.Image {
	// In headless mode, skip transparency processing
	// ReadPixels cannot be called before game starts
	if e.headlessMode {
		// Return a simple SubImage without transparency processing
		srcRect := image.Rect(srcX, srcY, srcX+width, srcY+height)
		return srcPic.Image.SubImage(srcRect).(*ebiten.Image)
	}

	// Create temporary RGBA image for pixel processing
	processedImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// Convert 0xRRGGBB to RGB components
	tr := uint8((transparentColor >> 16) & 0xFF)
	tg := uint8((transparentColor >> 8) & 0xFF)
	tb := uint8(transparentColor & 0xFF)

	// Read pixels from Ebiten image
	// Note: This is still pixel-by-pixel but only done once at PutCast time
	for sy := 0; sy < height; sy++ {
		for sx := 0; sx < width; sx++ {
			srcPixelX := srcX + sx
			srcPixelY := srcY + sy

			// Check bounds
			if srcPixelX < 0 || srcPixelX >= srcPic.Width || srcPixelY < 0 || srcPixelY >= srcPic.Height {
				// Out of bounds - make transparent
				processedImg.Set(sx, sy, color.RGBA{0, 0, 0, 0})
				continue
			}

			// Get source pixel from Ebiten image
			c := srcPic.Image.At(srcPixelX, srcPixelY)
			r, g, b, a := c.RGBA()

			// Convert from 16-bit to 8-bit
			r8 := uint8(r >> 8)
			g8 := uint8(g >> 8)
			b8 := uint8(b >> 8)

			// Check if matches transparent color
			if r8 == tr && g8 == tg && b8 == tb {
				// Make fully transparent
				processedImg.Set(sx, sy, color.RGBA{0, 0, 0, 0})
			} else {
				// Keep original color with alpha
				processedImg.Set(sx, sy, color.RGBA{r8, g8, b8, uint8(a >> 8)})
			}
		}
	}

	// Convert to Ebiten image
	return ebiten.NewImageFromImage(processedImg)
}

// drawCastToPicture draws a cast directly to a picture using Ebiten.
// This is used by PutCast to "bake" casts into pictures.
// transparentColor: color to treat as transparent (0xRRGGBB format, -1 = no transparency)
func (e *EngineState) drawCastToPicture(cast *Cast, destPicID int, transparentColor int) {
	if e.debugLevel >= 2 {
		fmt.Printf("[DEBUG] drawCastToPicture: cast %d (pic %d) -> dest pic %d at (%d,%d) clip=(%d,%d,%d,%d) transparent=0x%X\n",
			cast.ID, cast.PictureID, destPicID, cast.X, cast.Y, cast.SrcX, cast.SrcY, cast.Width, cast.Height, transparentColor)
	}

	// Get source picture
	srcPic := e.pictures[cast.PictureID]
	if srcPic == nil {
		if e.debugLevel >= 1 {
			fmt.Printf("[ERROR] drawCastToPicture: source picture %d not found\n", cast.PictureID)
		}
		return
	}

	// Get destination picture
	destPic := e.pictures[destPicID]
	if destPic == nil {
		if e.debugLevel >= 1 {
			fmt.Printf("[ERROR] drawCastToPicture: destination picture %d not found\n", destPicID)
		}
		return
	}

	// If we have a pre-processed image with transparency, use it
	if cast.ProcessedImage != nil {
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(cast.X), float64(cast.Y))
		destPic.Image.DrawImage(cast.ProcessedImage, opts)
		return
	}

	// Otherwise, extract the clipped region and draw it
	srcRect := image.Rect(cast.SrcX, cast.SrcY, cast.SrcX+cast.Width, cast.SrcY+cast.Height)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(cast.X), float64(cast.Y))
	destPic.Image.DrawImage(subImg, opts)
}

// GetCast retrieves a cast by ID.
// Returns nil if the cast doesn't exist.
func (e *EngineState) GetCast(id int) *Cast {
	return e.casts[id]
}

// MoveCast updates cast position and optionally clipping.
// In FILLY, MoveCast uses double-buffering to prevent cast accumulation:
// 1. Clears BackBuffer to transparent
// 2. Redraws ALL casts onto BackBuffer
// 3. Swaps BackBuffer with main Image
// This ensures casts are redrawn cleanly each frame without accumulating.
// Parameters:
//
//	id: cast ID
//	x, y: new position relative to window
//	srcX, srcY: new source clipping position (optional, -1 = no change)
//	width, height: new clipping dimensions (optional, -1 = no change)
func (e *EngineState) MoveCast(id, x, y, srcX, srcY, width, height int) error {
	cast := e.casts[id]
	if cast == nil {
		return fmt.Errorf("cast %d not found", id)
	}

	// Update position
	cast.X = x
	cast.Y = y

	// Check if clipping parameters changed (animation frame change)
	clippingChanged := false
	if srcX >= 0 && srcX != cast.SrcX {
		cast.SrcX = srcX
		clippingChanged = true
	}
	if srcY >= 0 && srcY != cast.SrcY {
		cast.SrcY = srcY
		clippingChanged = true
	}
	if width >= 0 && width != cast.Width {
		cast.Width = width
		clippingChanged = true
	}
	if height >= 0 && height != cast.Height {
		cast.Height = height
		clippingChanged = true
	}

	// If clipping changed and we have transparency, regenerate ProcessedImage
	if clippingChanged && cast.TransparentColor >= 0 {
		srcPic := e.pictures[cast.PictureID]
		if srcPic != nil {
			cast.ProcessedImage = e.createTransparentImage(srcPic, cast.SrcX, cast.SrcY, cast.Width, cast.Height, cast.TransparentColor)
		}
	}

	// IMPORTANT: In FILLY, MoveCast uses double-buffering to prevent cast accumulation
	// cast.WindowID actually stores destPicID
	destPicID := cast.WindowID

	destPic := e.pictures[destPicID]
	if destPic == nil {
		return fmt.Errorf("destination picture %d not found", destPicID)
	}

	// Initialize BackBuffer if needed
	if destPic.BackBuffer == nil {
		destPic.BackBuffer = ebiten.NewImage(destPic.Width, destPic.Height)
	}

	// CRITICAL: Copy current Image content to BackBuffer FIRST
	// This preserves any MovePic drawings that were done to the main Image
	// Then we'll draw all casts on top of this content
	destPic.BackBuffer.Clear()
	opts := &ebiten.DrawImageOptions{}
	destPic.BackBuffer.DrawImage(destPic.Image, opts)

	// Redraw ALL casts that belong to this destination picture onto BackBuffer
	for _, c := range e.GetCasts() {
		if c.WindowID == destPicID && c.Visible {
			// Draw cast directly to BackBuffer
			e.drawCastToImage(c, destPic.BackBuffer, c.TransparentColor)
		}
	}

	// Swap BackBuffer with main Image (double buffering)
	temp := destPic.Image
	destPic.Image = destPic.BackBuffer
	destPic.BackBuffer = temp

	return nil
}

// drawCastToImage draws a cast directly to a specific Ebiten image buffer.
// This is used by MoveCast for double buffering.
func (e *EngineState) drawCastToImage(cast *Cast, destImg *ebiten.Image, transparentColor int) {
	if e.debugLevel >= 2 {
		fmt.Printf("[DEBUG] drawCastToImage: cast %d (pic %d) at (%d,%d) clip=(%d,%d,%d,%d) transparent=0x%X\n",
			cast.ID, cast.PictureID, cast.X, cast.Y, cast.SrcX, cast.SrcY, cast.Width, cast.Height, transparentColor)
	}

	// If we have a pre-processed image with transparency, use it directly
	if cast.ProcessedImage != nil {
		// Use the pre-processed image (transparency already applied)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(cast.X), float64(cast.Y))
		destImg.DrawImage(cast.ProcessedImage, opts)
		return
	}

	// Otherwise, fall back to original logic (for casts without transparency)
	// Get source picture
	srcPic := e.pictures[cast.PictureID]
	if srcPic == nil {
		if e.debugLevel >= 1 {
			fmt.Printf("[ERROR] drawCastToImage: source picture %d not found\n", cast.PictureID)
		}
		return
	}

	// Extract the clipped region from source
	srcRect := image.Rect(cast.SrcX, cast.SrcY, cast.SrcX+cast.Width, cast.SrcY+cast.Height)
	subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

	// Draw to destination
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(cast.X), float64(cast.Y))
	destImg.DrawImage(subImg, opts)
}

// redrawAllCastsOnWindow redraws all casts on a window to its picture.
// This is called by MoveCast to update the destination picture.
func (e *EngineState) redrawAllCastsOnWindow(windowID int) {
	// Get the window to find its picture
	window := e.windows[windowID]
	if window == nil {
		return
	}

	destPicID := window.PictureID

	// Clear the destination picture first (or restore base image)
	// For now, we'll just redraw all casts on top
	// TODO: Implement proper base image restoration if needed

	// Redraw all casts for this window
	for _, cast := range e.GetCastsByWindow(windowID) {
		e.drawCastToPicture(cast, destPicID, cast.TransparentColor)
	}
}

// DeleteCast removes a cast and releases its resources.
func (e *EngineState) DeleteCast(id int) {
	delete(e.casts, id)
}

// GetCasts returns all casts in creation order (z-order).
// This is used for rendering casts in the correct order.
func (e *EngineState) GetCasts() []*Cast {
	// Collect all casts
	casts := make([]*Cast, 0, len(e.casts))
	for _, cast := range e.casts {
		casts = append(casts, cast)
	}

	// Sort by ID (creation order)
	for i := 0; i < len(casts); i++ {
		for j := i + 1; j < len(casts); j++ {
			if casts[i].ID > casts[j].ID {
				casts[i], casts[j] = casts[j], casts[i]
			}
		}
	}

	return casts
}

// GetCastsByWindow returns all casts for a specific window in creation order.
func (e *EngineState) GetCastsByWindow(windowID int) []*Cast {
	// Collect casts for this window
	casts := make([]*Cast, 0)
	for _, cast := range e.casts {
		if cast.WindowID == windowID {
			casts = append(casts, cast)
		}
	}

	// Sort by ID (creation order)
	for i := 0; i < len(casts); i++ {
		for j := i + 1; j < len(casts); j++ {
			if casts[i].ID > casts[j].ID {
				casts[i], casts[j] = casts[j], casts[i]
			}
		}
	}

	return casts
}
