package engine

import (
	"sync/atomic"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Engine is the main FILLY engine that coordinates execution.
type Engine struct {
	state             *EngineState
	logger            *Logger
	headless          bool
	programTerminated atomic.Bool
	timeout           time.Duration
	startTime         time.Time
}

// NewEngine creates a new FILLY engine.
func NewEngine(renderer Renderer, assetLoader AssetLoader, imageDecoder ImageDecoder) *Engine {
	state := NewEngineState(renderer, assetLoader, imageDecoder)
	logger := NewLogger(DebugLevelError)

	return &Engine{
		state:    state,
		logger:   logger,
		headless: false,
		timeout:  0,
	}
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
		return nil
	}

	// Increment tick counter
	e.state.IncrementTick()

	// Update VM execution
	if err := e.UpdateVM(); err != nil {
		return err
	}

	// TODO: Update audio (Phase 5)

	return nil
}

// UpdateVM processes one tick for all active sequences.
// Each active sequence that is not waiting executes one OpCode.
func (e *Engine) UpdateVM() error {
	vm := NewVM(e.state, e.logger)

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

	// TODO: Actual rendering (Phase 4)
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
