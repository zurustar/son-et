package engine

import (
	"sync/atomic"
	"time"
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

	// TODO: Update VM execution (Phase 3)
	// TODO: Update audio (Phase 5)

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
