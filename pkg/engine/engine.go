package engine

import (
	"errors"
	"image/color"
	"sync/atomic"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// ErrTerminated is returned when the engine is terminated.
var ErrTerminated = errors.New("engine terminated")

// Engine is the main FILLY engine that coordinates execution.
type Engine struct {
	state             *EngineState
	logger            *Logger
	midiPlayer        *MIDIPlayer
	wavPlayer         *WAVPlayer
	textRenderer      *TextRenderer
	drawingContext    *DrawingContext
	headless          bool
	programTerminated atomic.Bool
	timeout           time.Duration
	startTime         time.Time
}

// NewEngine creates a new FILLY engine.
func NewEngine(renderer Renderer, assetLoader AssetLoader, imageDecoder ImageDecoder) *Engine {
	state := NewEngineState(renderer, assetLoader, imageDecoder)
	logger := NewLogger(DebugLevelError)

	engine := &Engine{
		state:    state,
		logger:   logger,
		headless: false,
		timeout:  0,
	}

	// Create MIDI player
	engine.midiPlayer = NewMIDIPlayer(engine)

	// Create WAV player
	engine.wavPlayer = NewWAVPlayer(engine)

	// Create text renderer
	engine.textRenderer = NewTextRenderer(engine)

	// Create drawing context
	engine.drawingContext = NewDrawingContext()

	return engine
}

// SetHeadless enables or disables headless mode.
func (e *Engine) SetHeadless(enabled bool) {
	e.headless = enabled
	e.state.SetHeadlessMode(enabled)
	if enabled {
		e.logger.LogInfo("Headless mode enabled")
	}
}

// IsHeadless returns whether headless mode is enabled.
func (e *Engine) IsHeadless() bool {
	return e.headless
}

// SetTimeout sets the execution timeout.
// A timeout of 0 means no timeout.
func (e *Engine) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
	if timeout > 0 {
		e.logger.LogInfo("Timeout set to %v", timeout)
	}
}

// SetDebugLevel sets the logging debug level.
func (e *Engine) SetDebugLevel(level DebugLevel) {
	e.logger.SetLevel(level)
	e.state.SetDebugLevel(int(level))
}

// Start initializes the engine and starts execution.
func (e *Engine) Start() {
	e.startTime = time.Now()
	e.programTerminated.Store(false)
	e.logger.LogInfo("Engine started")
}

// Terminate sets the termination flag.
func (e *Engine) Terminate() {
	if !e.programTerminated.Load() {
		e.programTerminated.Store(true)
		e.logger.LogInfo("Engine termination requested")
	}
}

// IsTerminated returns whether the engine has been terminated.
func (e *Engine) IsTerminated() bool {
	return e.programTerminated.Load()
}

// CheckTermination checks if the engine should terminate.
// Returns true if termination is requested or timeout exceeded.
func (e *Engine) CheckTermination() bool {
	// Check termination flag
	if e.programTerminated.Load() {
		return true
	}

	// Check timeout
	if e.timeout > 0 {
		elapsed := time.Since(e.startTime)
		if elapsed >= e.timeout {
			e.logger.LogInfo("Timeout exceeded: %v", elapsed)
			e.Terminate()
			return true
		}
	}

	return false
}

// Update performs one engine tick (called at 60 FPS).
func (e *Engine) Update() error {
	// Check termination before execution
	if e.CheckTermination() {
		// Return a termination error to stop the game loop
		return ErrTerminated
	}

	// Increment tick counter
	e.state.IncrementTick()

	// Update VM execution
	if err := e.UpdateVM(); err != nil {
		return err
	}

	// Check if all sequences have completed
	if e.AllSequencesComplete() {
		e.logger.LogInfo("All sequences completed, terminating")
		e.Terminate()
		return ErrTerminated
	}

	// TODO: Update audio (Phase 5)

	return nil
}

// UpdateVM processes one tick for all active sequences.
// Each active sequence that is not waiting executes one OpCode.
func (e *Engine) UpdateVM() error {
	vm := NewVM(e.state, e, e.logger)

	// Process all active sequences
	for _, seq := range e.state.GetSequencers() {
		// Skip inactive sequences
		if !seq.IsActive() {
			continue
		}

		// Skip completed sequences
		if seq.IsComplete() {
			continue
		}

		// Handle waiting sequences
		if seq.IsWaiting() {
			seq.DecrementWait()
			continue
		}

		// Get current command
		cmd := seq.GetCurrentCommand()
		if cmd == nil {
			continue
		}

		// Execute command
		if err := vm.ExecuteOp(seq, *cmd); err != nil {
			e.logger.LogError("VM execution error at seq %d, pc %d: %v", seq.GetID(), seq.GetPC(), err)
			return err
		}

		// Advance program counter
		seq.IncrementPC()
	}

	return nil
}

// Render renders the current frame.
func (e *Engine) Render() {
	if e.headless {
		e.logger.LogDebug("Render (headless)")
		return
	}

	// Renderer is set externally (in main.go)
	// For now, just log
	e.logger.LogDebug("Render frame %d", e.state.GetTickCount())
}

// Shutdown performs cleanup and releases resources.
func (e *Engine) Shutdown() {
	e.logger.LogInfo("Engine shutdown")
	// TODO: Cleanup resources (Phase 3+)
}

// GetState returns the engine state (for testing).
func (e *Engine) GetState() *EngineState {
	return e.state
}

// GetLogger returns the logger (for testing).
func (e *Engine) GetLogger() *Logger {
	return e.logger
}

// WinInfo returns information about the virtual desktop.
// index 0: desktop width (1280)
// index 1: desktop height (720)
// Returns 0 for invalid indices.
func (e *Engine) WinInfo(index int) int {
	switch index {
	case 0:
		return e.state.GetDesktopWidth()
	case 1:
		return e.state.GetDesktopHeight()
	default:
		return 0
	}
}

// RegisterSequence registers a new sequence with the engine.
// If groupID is 0, a new group ID is allocated.
// Returns the sequence ID.
func (e *Engine) RegisterSequence(seq *Sequencer, groupID int) int {
	if groupID == 0 {
		groupID = e.state.AllocateGroupID()
	}

	seqID := e.state.RegisterSequence(seq, groupID)
	e.logger.LogDebug("Registered sequence %d (group %d, mode %d)", seqID, groupID, seq.GetMode())

	return seqID
}

// DeleteMe deactivates the current sequence (del_me).
// This is called from within a sequence to terminate itself.
func (e *Engine) DeleteMe(seqID int) {
	e.state.DeactivateSequence(seqID)
	e.logger.LogDebug("Deactivated sequence %d (del_me)", seqID)
}

// DeleteUs deactivates all sequences in a group (del_us).
func (e *Engine) DeleteUs(groupID int) {
	e.state.DeactivateGroup(groupID)
	e.logger.LogDebug("Deactivated group %d (del_us)", groupID)
}

// DeleteAll deactivates all sequences (del_all).
func (e *Engine) DeleteAll() {
	e.state.DeactivateAll()
	e.logger.LogDebug("Deactivated all sequences (del_all)")
}

// CleanupSequences removes inactive sequences from the list.
func (e *Engine) CleanupSequences() {
	e.state.CleanupInactiveSequences()
}

// RegisterMesBlock registers a mes() block (event handler).
// For TIME mode, this blocks until the sequence completes.
// For other modes, this returns immediately.
func (e *Engine) RegisterMesBlock(eventType EventType, opcodes []interpreter.OpCode, parent *Sequencer, userID int) int {
	// Determine timing mode based on event type
	var mode TimingMode
	if eventType == EventTIME {
		mode = TIME
	} else {
		mode = MIDI_TIME
	}

	// Register as event handler (stores OpCode template, not Sequencer)
	handlerID := e.state.RegisterEventHandler(eventType, opcodes, mode, parent, userID)
	e.logger.LogDebug("Registered mes(%s) handler %d (user ID: %d)", eventType.String(), handlerID, userID)

	// For TIME mode, execute immediately and block
	if eventType == EventTIME {
		// Create a new sequencer for immediate execution
		seq := NewSequencer(opcodes, mode, parent)
		seqID := e.RegisterSequence(seq, 0)
		e.logger.LogDebug("TIME mode: executing sequence %d (blocking)", seqID)

		// TODO: Block until sequence completes (requires main loop integration)
		// This will be verified in Task 7.3.8
		// For now, just register it - blocking behavior depends on how
		// the main loop calls UpdateVM() while waiting
	}

	return handlerID
}

// TriggerEvent triggers all event handlers for a given event type.
// Event parameters are passed via EventData.
// Each trigger creates a new Sequencer instance from the handler's OpCode template.
func (e *Engine) TriggerEvent(eventType EventType, data *EventData) {
	handlers := e.state.GetEventHandlers(eventType)
	e.logger.LogDebug("Triggering %s event (%d handlers)", eventType.String(), len(handlers))

	for _, handler := range handlers {
		// Create a NEW sequencer instance from the handler's template
		seq := NewSequencer(handler.Commands, handler.Mode, handler.Parent)

		// Set event parameters in the NEW sequencer's scope
		if data != nil {
			seq.SetVariable("MesP1", int64(data.MesP1))
			seq.SetVariable("MesP2", int64(data.MesP2))
			seq.SetVariable("MesP3", int64(data.MesP3))
			seq.SetVariable("MesP4", int64(data.MesP4))
		}

		// Register the NEW sequencer for execution
		e.RegisterSequence(seq, 0)
	}
}

// TriggerUserEvent triggers all USER event handlers for a specific user ID.
// Each trigger creates a new Sequencer instance from the handler's OpCode template.
func (e *Engine) TriggerUserEvent(userID int, data *EventData) {
	handlers := e.state.GetUserEventHandlers(userID)
	e.logger.LogDebug("Triggering USER event %d (%d handlers)", userID, len(handlers))

	for _, handler := range handlers {
		// Create a NEW sequencer instance from the handler's template
		seq := NewSequencer(handler.Commands, handler.Mode, handler.Parent)

		// Set event parameters in the NEW sequencer's scope
		if data != nil {
			seq.SetVariable("MesP1", int64(data.MesP1))
			seq.SetVariable("MesP2", int64(data.MesP2))
			seq.SetVariable("MesP3", int64(data.MesP3))
			seq.SetVariable("MesP4", int64(data.MesP4))
		}

		// Register the NEW sequencer for execution
		e.RegisterSequence(seq, 0)
	}
}

// CallMainFunction calls the main() function if it exists.
// This should be called after the initial script has been loaded and executed
// (so that function definitions are registered).
func (e *Engine) CallMainFunction() error {
	// Look up main function
	mainFunc, ok := e.state.GetFunction("main")
	if !ok {
		e.logger.LogInfo("No main() function found, skipping automatic execution")
		return nil
	}

	e.logger.LogInfo("Calling main() function")

	// Create a new sequencer for main() execution (TIME mode, no parent)
	mainSeq := NewSequencer(mainFunc.Body, TIME, nil)

	// Register the main sequence
	seqID := e.RegisterSequence(mainSeq, 0)

	e.logger.LogDebug("Registered main() sequence %d", seqID)

	return nil
}

// ExecuteTopLevel executes top-level opcodes synchronously.
// This is used to register function definitions before starting the main execution loop.
func (e *Engine) ExecuteTopLevel(opcodes []interpreter.OpCode) error {
	e.logger.LogDebug("Executing %d top-level opcodes", len(opcodes))

	// Create a temporary sequencer for top-level execution
	seq := NewSequencer(opcodes, TIME, nil)
	vm := NewVM(e.state, e, e.logger)

	// Execute all opcodes synchronously
	for !seq.IsComplete() {
		cmd := seq.GetCurrentCommand()
		if cmd == nil {
			break
		}

		if err := vm.ExecuteOp(seq, *cmd); err != nil {
			return err
		}

		seq.IncrementPC()
	}

	e.logger.LogDebug("Top-level execution complete")
	return nil
}

// AllSequencesComplete checks if all sequences have completed execution.
// Returns true if there are no active sequences or all sequences are complete.
func (e *Engine) AllSequencesComplete() bool {
	sequencers := e.state.GetSequencers()

	// If no sequences exist, consider it complete
	if len(sequencers) == 0 {
		e.logger.LogDebug("AllSequencesComplete: no sequences exist")
		return true
	}

	// Check if all sequences are either inactive or complete
	activeCount := 0
	completeCount := 0
	for _, seq := range sequencers {
		if seq.IsActive() {
			activeCount++
			if !seq.IsComplete() {
				e.logger.LogDebug("AllSequencesComplete: sequence %d is active and not complete (pc=%d/%d, waiting=%v)",
					seq.GetID(), seq.GetPC(), len(seq.commands), seq.IsWaiting())
				return false
			} else {
				completeCount++
			}
		}
	}

	e.logger.LogDebug("AllSequencesComplete: %d active sequences, %d complete", activeCount, completeCount)
	return true
}

// LoadPic loads an image file and returns its picture ID.
// Returns 0 on error.
func (e *Engine) LoadPic(filename string) int {
	picID, err := e.state.LoadPicture(filename)
	if err != nil {
		e.logger.LogError("LoadPic failed: %v", err)
		return 0
	}
	e.logger.LogDebug("Loaded picture %d: %s", picID, filename)
	return picID
}

// CreatePic creates an empty image buffer.
// Returns the picture ID.
func (e *Engine) CreatePic(width, height int) int {
	picID := e.state.CreatePicture(width, height)
	e.logger.LogDebug("Created picture %d: %dx%d", picID, width, height)
	return picID
}

// DelPic deletes a picture and releases its resources.
func (e *Engine) DelPic(picID int) {
	e.state.DeletePicture(picID)
	e.logger.LogDebug("Deleted picture %d", picID)
}

// PicWidth returns the width of a picture.
func (e *Engine) PicWidth(picID int) int {
	return e.state.GetPictureWidth(picID)
}

// PicHeight returns the height of a picture.
func (e *Engine) PicHeight(picID int) int {
	return e.state.GetPictureHeight(picID)
}

// MovePic copies pixels from source picture to destination picture with transparency.
func (e *Engine) MovePic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, mode int) {
	err := e.state.MovePicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, mode)
	if err != nil {
		e.logger.LogError("MovePic failed: %v", err)
		return
	}
	e.logger.LogDebug("MovePic: src=%d (%d,%d,%d,%d) -> dst=%d (%d,%d) mode=%d",
		srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, mode)
}

// MoveSPic copies and scales pixels from source picture to destination picture with transparency.
func (e *Engine) MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH int) {
	err := e.state.MoveSPicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)
	if err != nil {
		e.logger.LogError("MoveSPic failed: %v", err)
		return
	}
	e.logger.LogDebug("MoveSPic: src=%d (%d,%d,%d,%d) -> dst=%d (%d,%d,%d,%d)",
		srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)
}

// ReversePic copies and horizontally flips pixels from source picture to destination picture.
func (e *Engine) ReversePic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY int) {
	err := e.state.ReversePicture(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)
	if err != nil {
		e.logger.LogError("ReversePic failed: %v", err)
		return
	}
	e.logger.LogDebug("ReversePic: src=%d (%d,%d,%d,%d) -> dst=%d (%d,%d) [flipped]",
		srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)
}

// OpenWin creates a new window and returns its ID.
func (e *Engine) OpenWin(picID, x, y, width, height, picX, picY, color int) int {
	winID := e.state.OpenWindow(picID, x, y, width, height, picX, picY, color)
	// Get the actual window to log its real dimensions
	win := e.state.GetWindow(winID)
	if win != nil {
		e.logger.LogDebug("Opened window %d: pic=%d pos=(%d,%d) size=(%dx%d) picOffset=(%d,%d) color=0x%X",
			winID, win.PictureID, win.X, win.Y, win.Width, win.Height, win.PicX, win.PicY, color)
	}
	return winID
}

// MoveWin updates window properties.
func (e *Engine) MoveWin(id, picID, x, y, width, height, picX, picY int) {
	err := e.state.MoveWindow(id, picID, x, y, width, height, picX, picY)
	if err != nil {
		e.logger.LogError("MoveWin failed: %v", err)
		return
	}
	e.logger.LogDebug("Moved window %d: pic=%d pos=(%d,%d) size=(%dx%d) picOffset=(%d,%d)",
		id, picID, x, y, width, height, picX, picY)
}

// CloseWin closes a window.
func (e *Engine) CloseWin(id int) {
	e.state.CloseWindow(id)
	e.logger.LogDebug("Closed window %d", id)
}

// CloseWinAll closes all windows.
func (e *Engine) CloseWinAll() {
	e.state.CloseAllWindows()
	e.logger.LogDebug("Closed all windows")
}

// CapTitle sets the caption (title) of a window.
func (e *Engine) CapTitle(id int, caption string) {
	err := e.state.SetWindowCaption(id, caption)
	if err != nil {
		e.logger.LogError("CapTitle failed: %v", err)
		return
	}
	e.logger.LogDebug("Set window %d caption: %q", id, caption)
}

// GetPicNo returns the picture ID associated with a window.
func (e *Engine) GetPicNo(id int) int {
	return e.state.GetWindowPictureID(id)
}

// StartWindowDrag initiates dragging a window at the given mouse position.
// Returns the window ID that started dragging, or 0 if no window was clicked.
func (e *Engine) StartWindowDrag(mouseX, mouseY int) int {
	winID := e.state.StartWindowDrag(mouseX, mouseY)
	if winID != 0 {
		e.logger.LogDebug("Started dragging window %d at mouse position (%d,%d)", winID, mouseX, mouseY)
	}
	return winID
}

// UpdateWindowDrag updates the position of the dragged window.
// Returns true if a window was updated.
func (e *Engine) UpdateWindowDrag(mouseX, mouseY int) bool {
	updated := e.state.UpdateWindowDrag(mouseX, mouseY)
	if updated {
		winID := e.state.GetDraggedWindowID()
		e.logger.LogDebug("Updated dragged window %d to mouse position (%d,%d)", winID, mouseX, mouseY)
	}
	return updated
}

// StopWindowDrag stops dragging the current window.
func (e *Engine) StopWindowDrag() {
	winID := e.state.GetDraggedWindowID()
	if winID != 0 {
		e.logger.LogDebug("Stopped dragging window %d", winID)
	}
	e.state.StopWindowDrag()
}

// GetDraggedWindowID returns the ID of the window currently being dragged.
func (e *Engine) GetDraggedWindowID() int {
	return e.state.GetDraggedWindowID()
}

// PutCast creates a new cast (sprite) and returns its ID.
func (e *Engine) PutCast(windowID, picID, x, y, srcX, srcY, width, height int) int {
	castID := e.state.PutCast(windowID, picID, x, y, srcX, srcY, width, height)
	e.logger.LogDebug("Created cast %d: win=%d pic=%d pos=(%d,%d) clip=(%d,%d,%d,%d)",
		castID, windowID, picID, x, y, srcX, srcY, width, height)
	return castID
}

// MoveCast updates cast position and optionally clipping.
func (e *Engine) MoveCast(id, x, y, srcX, srcY, width, height int) {
	err := e.state.MoveCast(id, x, y, srcX, srcY, width, height)
	if err != nil {
		e.logger.LogError("MoveCast failed: %v", err)
		return
	}
	e.logger.LogDebug("Moved cast %d: pos=(%d,%d) clip=(%d,%d,%d,%d)",
		id, x, y, srcX, srcY, width, height)
}

// DelCast removes a cast.
func (e *Engine) DelCast(id int) {
	e.state.DeleteCast(id)
	e.logger.LogDebug("Deleted cast %d", id)
}

// LoadSoundFont loads a SoundFont (.sf2) file for MIDI synthesis.
func (e *Engine) LoadSoundFont(filename string) error {
	return e.midiPlayer.LoadSoundFont(filename)
}

// PlayMIDI starts MIDI playback.
// Returns immediately (non-blocking).
func (e *Engine) PlayMIDI(filename string) error {
	return e.midiPlayer.PlayMIDI(filename)
}

// StopMIDI stops MIDI playback.
func (e *Engine) StopMIDI() {
	e.midiPlayer.Stop()
}

// IsMIDIPlaying returns whether MIDI is currently playing.
func (e *Engine) IsMIDIPlaying() bool {
	return e.midiPlayer.IsPlaying()
}

// PlayWAVE plays a WAV file asynchronously.
func (e *Engine) PlayWAVE(filename string) error {
	return e.wavPlayer.PlayWAVE(filename)
}

// LoadRsc preloads a WAV file into memory.
// Returns a resource ID for use with PlayRsc.
func (e *Engine) LoadRsc(filename string) (int, error) {
	return e.wavPlayer.LoadRsc(filename)
}

// PlayRsc plays a preloaded WAV resource.
func (e *Engine) PlayRsc(resourceID int) error {
	return e.wavPlayer.PlayRsc(resourceID)
}

// DelRsc deletes a preloaded WAV resource.
func (e *Engine) DelRsc(resourceID int) {
	e.wavPlayer.DelRsc(resourceID)
}

// StopAllWAV stops all active WAV players.
func (e *Engine) StopAllWAV() {
	e.wavPlayer.StopAll()
}

// CleanupWAV removes finished WAV players.
func (e *Engine) CleanupWAV() {
	e.wavPlayer.Cleanup()
}

// SetFont sets the current font for text rendering.
func (e *Engine) SetFont(size int, name string, charset int) {
	e.textRenderer.SetFont(size, name, charset)
}

// TextColor sets the text color.
func (e *Engine) TextColor(r, g, b int) {
	e.textRenderer.SetTextColor(r, g, b)
}

// BgColor sets the background color for text.
func (e *Engine) BgColor(r, g, b int) {
	e.textRenderer.SetBgColor(r, g, b)
}

// BackMode sets the background mode (0=transparent, 1=opaque).
func (e *Engine) BackMode(mode int) {
	e.textRenderer.SetBackMode(mode)
}

// TextWrite draws text on a picture.
func (e *Engine) TextWrite(text string, picID, x, y int) error {
	return e.textRenderer.TextWrite(text, picID, x, y)
}

// MeasureText returns the width and height of text in pixels.
func (e *Engine) MeasureText(text string) (int, int) {
	return e.textRenderer.MeasureText(text)
}

// SetLineSize sets the line width for drawing operations.
func (e *Engine) SetLineSize(size int) {
	e.drawingContext.SetLineSize(size)
}

// SetPaintColor sets the drawing color.
func (e *Engine) SetPaintColor(colorValue int) {
	// Convert integer color (0xRRGGBB) to color.Color
	r := uint8((colorValue >> 16) & 0xFF)
	g := uint8((colorValue >> 8) & 0xFF)
	b := uint8(colorValue & 0xFF)
	e.drawingContext.SetPaintColor(color.RGBA{R: r, G: g, B: b, A: 255})
}

// SetROP sets the raster operation mode.
func (e *Engine) SetROP(mode int) {
	e.drawingContext.SetROP(ROPMode(mode))
}

// DrawLine draws a line on a picture.
func (e *Engine) DrawLine(picID, x1, y1, x2, y2 int) error {
	pic := e.state.GetPicture(picID)
	if pic == nil {
		return NewRuntimeError("DrawLine", "", "Picture %d not found", picID)
	}

	// Ensure picture is RGBA
	rgba := e.state.EnsureRGBA(pic.Image)
	e.drawingContext.DrawLine(rgba, x1, y1, x2, y2)

	e.logger.LogDebug("DrawLine: pic=%d (%d,%d) -> (%d,%d)", picID, x1, y1, x2, y2)
	return nil
}

// DrawCircle draws a circle on a picture.
func (e *Engine) DrawCircle(picID, x, y, radius, fillMode int) error {
	pic := e.state.GetPicture(picID)
	if pic == nil {
		return NewRuntimeError("DrawCircle", "", "Picture %d not found", picID)
	}

	// Ensure picture is RGBA
	rgba := e.state.EnsureRGBA(pic.Image)
	e.drawingContext.DrawCircle(rgba, x, y, radius, fillMode)

	e.logger.LogDebug("DrawCircle: pic=%d center=(%d,%d) radius=%d fill=%d", picID, x, y, radius, fillMode)
	return nil
}

// DrawRect draws a rectangle on a picture.
func (e *Engine) DrawRect(picID, x1, y1, x2, y2, fillMode int) error {
	pic := e.state.GetPicture(picID)
	if pic == nil {
		return NewRuntimeError("DrawRect", "", "Picture %d not found", picID)
	}

	// Ensure picture is RGBA
	rgba := e.state.EnsureRGBA(pic.Image)
	e.drawingContext.DrawRect(rgba, x1, y1, x2, y2, fillMode)

	e.logger.LogDebug("DrawRect: pic=%d (%d,%d) -> (%d,%d) fill=%d", picID, x1, y1, x2, y2, fillMode)
	return nil
}

// GetColor returns the color of a pixel at (x, y) on a picture.
// Returns the color as an integer (0xRRGGBB).
func (e *Engine) GetColor(picID, x, y int) (int, error) {
	pic := e.state.GetPicture(picID)
	if pic == nil {
		return 0, NewRuntimeError("GetColor", "", "Picture %d not found", picID)
	}

	// Ensure picture is RGBA
	rgba := e.state.EnsureRGBA(pic.Image)
	c := GetColor(rgba, x, y)

	// Convert color to integer (0xRRGGBB)
	r, g, b, _ := c.RGBA()
	colorValue := int((r>>8)<<16 | (g>>8)<<8 | (b >> 8))

	e.logger.LogDebug("GetColor: pic=%d (%d,%d) = 0x%06X", picID, x, y, colorValue)
	return colorValue, nil
}
