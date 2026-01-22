package engine

import (
	"errors"
	"image/color"
	"math/rand/v2"
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
	random            *rand.Rand
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
		random:   rand.New(rand.NewPCG(uint64(time.Now().UnixNano()), uint64(time.Now().UnixNano()))),
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

	// Update MIDI in headless mode
	if e.IsHeadless() && e.midiPlayer != nil {
		e.midiPlayer.UpdateHeadless()
	}

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

	// Check termination again after execution (in case timeout occurred during update)
	if e.CheckTermination() {
		return ErrTerminated
	}

	// TODO: Update audio (Phase 5)

	return nil
}

// UpdateVM processes one tick for all active sequences.
// Each active sequence that is not waiting executes OpCodes until it hits a wait or completes.
func (e *Engine) UpdateVM() error {
	vm := NewVM(e.state, e, e.logger)

	// Process all active sequences
	for _, seq := range e.state.GetSequencers() {
		// Skip inactive sequences
		if !seq.IsActive() {
			continue
		}

		// Don't skip "completed" sequences - they should loop back to the beginning
		// mes(TIME) and mes(MIDI_TIME) blocks run continuously until explicitly terminated

		// Handle waiting sequences (only for TIME mode)
		if seq.IsWaiting() {
			// Only decrement wait for TIME mode sequences
			// MIDI_TIME sequences are decremented by UpdateMIDISequences
			if seq.GetMode() == TIME {
				seq.DecrementWait()
			}
			continue
		}

		// If sequence reached the end, loop back to beginning (unless noLoop is set)
		if seq.GetPC() >= len(seq.commands) {
			if len(seq.commands) > 0 && seq.ShouldLoop() {
				seq.SetPC(0)
				e.logger.LogDebug("Sequence %d looped back to beginning", seq.GetID())
			} else {
				// Sequence is complete and should not loop
				continue
			}
		}

		// Execute commands until we hit a wait or reach the end
		// This prevents slow execution in mes() blocks with many OpCodes
		maxOpsPerTick := 1000 // Safety limit to prevent infinite loops
		opsExecuted := 0

		for opsExecuted < maxOpsPerTick {
			// Get current command
			cmd := seq.GetCurrentCommand()
			if cmd == nil {
				break
			}

			// Execute command
			if err := vm.ExecuteOp(seq, *cmd); err != nil {
				e.logger.LogError("VM execution error at seq %d, pc %d: %v", seq.GetID(), seq.GetPC(), err)

				// For resilience, continue execution instead of stopping the engine
				// This allows games to continue running even if individual operations fail
				// The original FILLY engine was likely more permissive with errors
			}

			// Advance program counter
			seq.IncrementPC()
			opsExecuted++

			// If sequence is now waiting, stop executing for this tick
			if seq.IsWaiting() {
				break
			}

			// If sequence reached the end, stop
			if seq.GetPC() >= len(seq.commands) {
				break
			}
		}

		if opsExecuted >= maxOpsPerTick {
			e.logger.LogError("Sequence %d hit max ops per tick limit (%d)", seq.GetID(), maxOpsPerTick)
		}
	}

	return nil
}

// UpdateMIDISequences processes MIDI tick updates for MIDI_TIME sequences.
// This is called when MIDI ticks advance (from MIDI player callback).
func (e *Engine) UpdateMIDISequences(tickCount int) error {
	if tickCount <= 0 {
		return nil
	}

	e.logger.LogInfo("UpdateMIDISequences: processing %d ticks", tickCount)

	vm := NewVM(e.state, e, e.logger)

	// Log MIDI sequences status (first tick only)
	if tickCount > 0 {
		midiSeqCount := 0
		for _, seq := range e.state.GetSequencers() {
			if seq.GetMode() == MIDI_TIME && seq.IsActive() && !seq.IsComplete() {
				midiSeqCount++
			}
		}
		if midiSeqCount == 0 {
			e.logger.LogInfo("  No active MIDI_TIME sequences found")
		}
	}

	// Process each MIDI tick sequentially
	for tick := 0; tick < tickCount; tick++ {
		// Process all MIDI_TIME sequences for this tick
		for _, seq := range e.state.GetSequencers() {
			// Skip inactive sequences
			if !seq.IsActive() {
				continue
			}

			// Don't skip "completed" sequences - they should loop back to the beginning
			// mes(MIDI_TIME) blocks run continuously until explicitly terminated

			// Only process MIDI_TIME mode sequences
			if seq.GetMode() != MIDI_TIME {
				continue
			}

			// Handle waiting sequences
			if seq.IsWaiting() {
				waitBefore := seq.GetWaitCount()
				seq.DecrementWait()
				waitAfter := seq.GetWaitCount()
				if tick == 0 {
					e.logger.LogDebug("  Seq %d: wait %d -> %d", seq.GetID(), waitBefore, waitAfter)
				}
				continue
			}

			// If sequence reached the end, loop back to beginning (unless noLoop is set)
			if seq.GetPC() >= len(seq.commands) {
				if len(seq.commands) > 0 && seq.ShouldLoop() {
					seq.SetPC(0)
					if tick == 0 {
						e.logger.LogDebug("  Seq %d: looped back to beginning", seq.GetID())
					}
				} else {
					// Sequence is complete and should not loop
					continue
				}
			}

			// Execute one command if not waiting
			cmd := seq.GetCurrentCommand()
			if cmd == nil {
				if tick == 0 {
					e.logger.LogInfo("  Seq %d: no current command (PC=%d)",
						seq.GetID(), seq.GetPC())
				}
				continue
			}

			// Log command execution (first tick only)
			if tick == 0 {
				e.logger.LogInfo("  Seq %d: executing command at PC=%d, waiting=%v",
					seq.GetID(), seq.GetPC(), seq.IsWaiting())
			}

			// Execute command
			if err := vm.ExecuteOp(seq, *cmd); err != nil {
				e.logger.LogError("MIDI VM execution error at seq %d, pc %d: %v", seq.GetID(), seq.GetPC(), err)

				// For resilience, continue execution instead of stopping the engine
				// This allows games to continue running even if individual operations fail
				// The original FILLY engine was likely more permissive with errors
			}

			// Log PC after execution (first tick only)
			if tick == 0 {
				e.logger.LogInfo("  Seq %d: after execution, PC=%d, waiting=%v, waitCount=%d",
					seq.GetID(), seq.GetPC(), seq.IsWaiting(), seq.GetWaitCount())
			}

			// Advance program counter
			seq.IncrementPC()
		}
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
// For MIDI_TIME mode, this starts a MIDI-synchronized sequence immediately.
// For other modes, this registers an event handler.
func (e *Engine) RegisterMesBlock(eventType EventType, opcodes []interpreter.OpCode, parent *Sequencer, userID int) int {
	// Determine timing mode based on event type
	var mode TimingMode
	if eventType == EventTIME {
		mode = TIME
	} else {
		mode = MIDI_TIME
	}

	// For TIME and MIDI_TIME modes, execute immediately as a sequence
	if eventType == EventTIME || eventType == EventMIDI_TIME {
		// Create a new sequencer for immediate execution
		seq := NewSequencer(opcodes, mode, parent)
		seqID := e.RegisterSequence(seq, 0)
		e.logger.LogDebug("%s mode: executing sequence %d", eventType.String(), seqID)
		return seqID
	}

	// For other event types, register as event handler
	handlerID := e.state.RegisterEventHandler(eventType, opcodes, mode, parent, userID)
	e.logger.LogDebug("Registered mes(%s) handler %d (user ID: %d)", eventType.String(), handlerID, userID)
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

	// main() should not loop - it runs once to initialize and register mes() blocks
	mainSeq.SetNoLoop(true)

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

	// If no sequences exist, check if there are any event handlers waiting
	if len(sequencers) == 0 {
		// Check if there are any registered event handlers (especially MIDI_TIME)
		// If MIDI is playing and there are MIDI_TIME handlers, don't terminate
		if e.midiPlayer != nil && e.midiPlayer.IsPlaying() {
			handlers := e.state.GetEventHandlers(EventMIDI_TIME)
			if len(handlers) > 0 {
				e.logger.LogDebug("AllSequencesComplete: no active sequences, but MIDI is playing with %d MIDI_TIME handlers", len(handlers))
				return false
			}
		}
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
				// Sequence is still running
				return false
			} else {
				completeCount++
			}
		}
	}

	// If all active sequences are complete, check if MIDI is still playing
	if e.midiPlayer != nil && e.midiPlayer.IsPlaying() {
		return false
	}

	// All sequences complete
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
	e.logger.LogInfo("Created picture %d: %dx%d", picID, width, height)
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
func (e *Engine) PutCast(destPicID, picID, x, y, srcX, srcY, width, height, transparentColor int) int {
	castID := e.state.PutCast(destPicID, picID, x, y, srcX, srcY, width, height, transparentColor)
	e.logger.LogInfo("Created cast %d: destPic=%d pic=%d pos=(%d,%d) clip=(%d,%d,%d,%d) transparent=0x%X",
		castID, destPicID, picID, x, y, srcX, srcY, width, height, transparentColor)
	return castID
}

// MoveCast updates cast position and optionally clipping.
func (e *Engine) MoveCast(id, x, y, srcX, srcY, width, height int) {
	err := e.state.MoveCast(id, x, y, srcX, srcY, width, height)
	if err != nil {
		e.logger.LogError("MoveCast failed: %v", err)
		return
	}
	e.logger.LogInfo("Moved cast %d: pos=(%d,%d) clip=(%d,%d,%d,%d)",
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

// GetSysTime returns the current Unix timestamp in seconds.
func (e *Engine) GetSysTime() int64 {
	return time.Now().Unix()
}

// WhatDay returns the current day of the month (1-31).
func (e *Engine) WhatDay() int {
	return time.Now().Day()
}

// WhatTime returns a time component based on the mode.
// mode 0: hour (0-23)
// mode 1: minute (0-59)
// mode 2: second (0-59)
func (e *Engine) WhatTime(mode int) int {
	now := time.Now()
	switch mode {
	case 0:
		return now.Hour()
	case 1:
		return now.Minute()
	case 2:
		return now.Second()
	default:
		return 0
	}
}

// GetMesNo returns the sequence ID of a mes() block.
// In FILLY, this is used to query the message number (sequence ID).
// Returns the sequence ID, or 0 if not found.
func (e *Engine) GetMesNo(seqID int) int {
	// Verify the sequence exists
	for _, seq := range e.state.GetSequencers() {
		if seq.GetID() == seqID {
			return seqID
		}
	}
	return 0
}

// DelMes terminates a specific mes() block (sequence) by ID.
// This is similar to DeleteMe but can target any sequence.
func (e *Engine) DelMes(seqID int) {
	e.state.DeactivateSequence(seqID)
	e.logger.LogDebug("DelMes: deactivated sequence %d", seqID)
}

// FreezeMes pauses a mes() block (sequence) by ID.
// The sequence stops executing but remains in the sequencer list.
// Note: Current implementation uses deactivation; a proper implementation
// would add a "paused" state to Sequencer.
func (e *Engine) FreezeMes(seqID int) {
	// For now, we deactivate the sequence
	// A full implementation would add a "paused" flag to Sequencer
	e.state.DeactivateSequence(seqID)
	e.logger.LogDebug("FreezeMes: paused sequence %d", seqID)
}

// ActivateMes resumes a paused mes() block (sequence) by ID.
// Note: Current implementation is a no-op since we don't have a proper
// pause/resume mechanism yet. This would require adding a "paused" state
// to Sequencer.
func (e *Engine) ActivateMes(seqID int) {
	// For now, this is a no-op
	// A full implementation would clear the "paused" flag on Sequencer
	e.logger.LogDebug("ActivateMes: resumed sequence %d (no-op in current implementation)", seqID)
}

// PostMes sends a custom message to mes() blocks.
// This triggers event handlers registered with mes(USER, userID).
// Parameters:
//   - messageType: event type (KEY, CLICK, USER, etc.) or user ID for USER events
//   - p1, p2, p3, p4: message parameters (stored in MesP1-MesP4)
func (e *Engine) PostMes(messageType int, p1, p2, p3, p4 int) {
	e.logger.LogDebug("PostMes: messageType=%d, params=(%d,%d,%d,%d)", messageType, p1, p2, p3, p4)

	// Create event data
	data := &EventData{
		MesP1: p1,
		MesP2: p2,
		MesP3: p3,
		MesP4: p4,
	}

	// Determine event type and trigger
	// For USER events, messageType is the user ID
	// For other events, messageType maps to EventType
	switch messageType {
	case 0: // TIME mode
		e.TriggerEvent(EventTIME, data)
	case 1: // MIDI_TIME mode
		e.TriggerEvent(EventMIDI_TIME, data)
	case 2: // MIDI_END
		e.TriggerEvent(EventMIDI_END, data)
	case 3: // KEY
		e.TriggerEvent(EventKEY, data)
	case 4: // CLICK
		e.TriggerEvent(EventCLICK, data)
	case 5: // RBDOWN
		e.TriggerEvent(EventRBDOWN, data)
	case 6: // RBDBLCLK
		e.TriggerEvent(EventRBDBLCLK, data)
	default:
		// Treat as USER event with custom ID
		e.TriggerUserEvent(messageType, data)
	}
}
