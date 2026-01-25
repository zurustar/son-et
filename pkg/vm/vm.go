// Package vm provides the virtual machine for executing FILLY script OpCodes.
// It implements an event-driven execution model with support for:
// - OpCode execution
// - Event handling (TIME, MIDI_TIME, MIDI_END, mouse events)
// - Scope management (global and local variables)
// - Built-in function registry
// - Audio system integration
// - Headless mode for testing
// - Timeout functionality
package vm

import (
	"context"
	"fmt"
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/zurustar/son-et/pkg/compiler"
	"github.com/zurustar/son-et/pkg/logger"
)

// MaxStackDepth is the maximum call stack depth before stack overflow.
// Requirement 20.7: System maintains maximum stack depth of 1000 frames.
const MaxStackDepth = 1000

// VM represents the virtual machine that executes OpCode instructions.
// It manages the execution state, scopes, event system, and audio system.
//
// Requirement 8.1: When VM receives OpCode sequence, system executes each OpCode in order.
// Requirement 14.1: System runs main event loop that processes events and executes OpCode.
type VM struct {
	// OpCode execution
	opcodes []compiler.OpCode
	pc      int // Program counter

	// Scope management
	globalScope *Scope
	localScope  *Scope
	callStack   []*StackFrame

	// User-defined functions
	functions map[string]*FunctionDef

	// Built-in functions
	builtins map[string]BuiltinFunc

	// Event system
	eventQueue      *EventQueue
	handlerRegistry *HandlerRegistry
	eventDispatcher *EventDispatcher
	currentHandler  *EventHandler // Currently executing handler (for del_me)

	// Audio system interface (to avoid import cycle)
	audioSystem AudioSystemInterface

	// Step execution state (for step() blocks outside handlers)
	// Requirement 6.1: When OpSetStep is executed, system initializes step counter.
	stepCounter int

	// Execution control
	running bool
	mu      sync.RWMutex

	// Configuration
	headless      bool
	timeout       time.Duration
	soundFontPath string
	titlePath     string // Base path for resolving relative file paths

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Logger
	log *slog.Logger
}

// AudioSystemInterface defines the interface for audio system operations.
// This interface is used to avoid import cycles between vm and vm/audio packages.
type AudioSystemInterface interface {
	PlayMIDI(filename string) error
	PlayWAVE(filename string) error
	SetMuted(muted bool)
	Update()
	Shutdown()
	StartTimer()
	StopTimer()
	IsMIDIPlaying() bool
	IsTimerRunning() bool
}

// FunctionDef represents a user-defined function.
type FunctionDef struct {
	Name       string
	Parameters []FunctionParam
	Body       []compiler.OpCode
}

// FunctionParam represents a function parameter.
type FunctionParam struct {
	Name       string
	Type       string
	IsArray    bool
	Default    any
	HasDefault bool
}

// StackFrame represents a call stack frame for function calls.
// Requirement 20.1: When function is called, system pushes new stack frame.
// Requirement 20.2: When function returns, system pops stack frame.
type StackFrame struct {
	FunctionName string
	LocalScope   *Scope
	ReturnPC     int
	ReturnValue  any
}

// BuiltinFunc is the signature for built-in functions.
// Built-in functions receive the VM instance and arguments, and return a value and error.
type BuiltinFunc func(vm *VM, args []any) (any, error)

// Option is a functional option for configuring the VM.
type Option func(*VM)

// WithHeadless enables headless mode (no GUI, muted audio).
// Requirement 12.1: When headless mode is enabled, system initializes audio system.
// Requirement 12.2: When headless mode is enabled, system mutes all audio output.
func WithHeadless(headless bool) Option {
	return func(vm *VM) {
		vm.headless = headless
	}
}

// WithTimeout sets the execution timeout.
// Requirement 13.1: When timeout is specified, system terminates execution after specified duration.
func WithTimeout(timeout time.Duration) Option {
	return func(vm *VM) {
		vm.timeout = timeout
	}
}

// WithLogger sets a custom logger.
func WithLogger(log *slog.Logger) Option {
	return func(vm *VM) {
		vm.log = log
	}
}

// WithSoundFont sets the SoundFont file path for MIDI playback.
// Requirement 4.9: When SoundFont file is provided, system uses it for MIDI synthesis.
func WithSoundFont(path string) Option {
	return func(vm *VM) {
		vm.soundFontPath = path
	}
}

// WithTitlePath sets the base path for resolving relative file paths.
// This is used for loading audio and image files relative to the title directory.
func WithTitlePath(path string) Option {
	return func(vm *VM) {
		vm.titlePath = path
	}
}

// New creates a new VM instance with the given OpCodes and options.
// It initializes the global scope, built-in functions, and applies configuration options.
//
// Parameters:
//   - opcodes: The compiled OpCode sequence to execute
//   - opts: Optional configuration options (headless, timeout, logger, soundFont)
//
// Returns:
//   - *VM: The initialized VM instance
func New(opcodes []compiler.OpCode, opts ...Option) *VM {
	ctx, cancel := context.WithCancel(context.Background())

	vm := &VM{
		opcodes:         opcodes,
		pc:              0,
		globalScope:     NewScope(nil),
		localScope:      nil,
		callStack:       make([]*StackFrame, 0, 64),
		functions:       make(map[string]*FunctionDef),
		builtins:        make(map[string]BuiltinFunc),
		eventQueue:      NewEventQueue(),
		handlerRegistry: NewHandlerRegistry(),
		currentHandler:  nil,
		audioSystem:     nil,
		running:         false,
		headless:        false,
		timeout:         0,
		soundFontPath:   "",
		ctx:             ctx,
		cancel:          cancel,
		log:             logger.GetLogger(),
	}

	// Initialize event dispatcher
	vm.eventDispatcher = NewEventDispatcher(vm.eventQueue, vm.handlerRegistry, vm)

	// Apply options
	for _, opt := range opts {
		opt(vm)
	}

	// Register default built-in functions
	vm.registerDefaultBuiltins()

	return vm
}

// registerDefaultBuiltins registers the default built-in functions.
func (vm *VM) registerDefaultBuiltins() {
	// del_me: Remove the currently executing handler
	// Requirement 10.3: When del_me is called, system removes current event handler.
	vm.RegisterBuiltinFunction("del_me", func(v *VM, args []any) (any, error) {
		if v.currentHandler != nil {
			v.currentHandler.Remove()
			v.log.Debug("del_me called", "handler", v.currentHandler.ID)
		}
		return nil, nil
	})

	// del_us: Same as del_me
	// Requirement 10.5: When del_us is called, system removes current event handler.
	vm.RegisterBuiltinFunction("del_us", func(v *VM, args []any) (any, error) {
		if v.currentHandler != nil {
			v.currentHandler.Remove()
			v.log.Debug("del_us called", "handler", v.currentHandler.ID)
		}
		return nil, nil
	})

	// del_all: Remove all registered handlers
	// Requirement 10.4: When del_all is called, system removes all event handlers.
	vm.RegisterBuiltinFunction("del_all", func(v *VM, args []any) (any, error) {
		v.handlerRegistry.UnregisterAll()
		v.log.Debug("del_all called")
		return nil, nil
	})

	// PlayMIDI: Play a MIDI file
	// Requirement 10.1: When PlayMIDI is called, system calls MIDI playback function.
	vm.RegisterBuiltinFunction("PlayMIDI", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			v.log.Error("PlayMIDI requires filename argument")
			return nil, nil
		}
		filename, ok := args[0].(string)
		if !ok {
			v.log.Error("PlayMIDI filename must be string", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}
		if err := v.PlayMIDI(filename); err != nil {
			// Requirement 11.2: When file is not found, system logs error and continues execution.
			v.log.Error("PlayMIDI failed", "filename", filename, "error", err)
			return nil, nil
		}
		v.log.Debug("PlayMIDI called", "filename", filename)
		return nil, nil
	})

	// PlayWAVE: Play a WAV file
	// Requirement 10.2: When PlayWAVE is called, system calls WAV playback function.
	vm.RegisterBuiltinFunction("PlayWAVE", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			v.log.Error("PlayWAVE requires filename argument")
			return nil, nil
		}
		filename, ok := args[0].(string)
		if !ok {
			v.log.Error("PlayWAVE filename must be string", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}
		if err := v.PlayWAVE(filename); err != nil {
			// Requirement 5.4: When WAV file is not found, system logs error and continues execution.
			// Requirement 5.5: When WAV file is corrupted, system logs error and continues execution.
			v.log.Error("PlayWAVE failed", "filename", filename, "error", err)
			return nil, nil
		}
		v.log.Debug("PlayWAVE called", "filename", filename)
		return nil, nil
	})

	// end_step: End the current step block
	// Requirement 6.7: When end_step is called, system terminates step block execution.
	// Requirement 10.6: When end_step is called, system terminates current step block.
	vm.RegisterBuiltinFunction("end_step", func(v *VM, args []any) (any, error) {
		if v.currentHandler != nil {
			// Reset the handler's step counter and wait counter
			v.currentHandler.StepCounter = 0
			v.currentHandler.WaitCounter = 0
			// Reset PC to end of handler to stop execution
			v.currentHandler.CurrentPC = len(v.currentHandler.OpCodes)
			v.log.Debug("end_step called", "handler", v.currentHandler.ID)
		} else {
			// Reset VM's step counter
			v.stepCounter = 0
			v.log.Debug("end_step called (no handler)")
		}
		return nil, nil
	})

	// Wait: Wait for specified number of events
	// Requirement 17.1: When Wait(n) is called, system pauses execution for n events.
	// Requirement 17.2: When Wait(n) is called in mes(TIME) handler, system waits for n TIME events.
	// Requirement 17.3: When Wait(n) is called in mes(MIDI_TIME) handler, system waits for n MIDI_TIME events.
	// Requirement 17.4: When Wait(0) is called, system continues execution immediately.
	// Requirement 17.5: When Wait(n) is called with n<0, system treats it as Wait(0).
	// Requirement 17.6: System maintains separate wait counter for each handler.
	vm.RegisterBuiltinFunction("Wait", func(v *VM, args []any) (any, error) {
		// Get wait count from arguments
		waitCount := 1 // Default to 1 if not specified
		if len(args) >= 1 {
			if wc, ok := toInt64(args[0]); ok {
				waitCount = int(wc)
			} else if f, fok := toFloat64(args[0]); fok {
				waitCount = int(f)
			}
		}

		// Requirement 17.4: When Wait(0) is called, system continues execution immediately.
		// Requirement 17.5: When Wait(n) is called with n<0, system treats it as Wait(0).
		if waitCount <= 0 {
			v.log.Debug("Wait: wait count <= 0, continuing immediately", "waitCount", waitCount)
			return nil, nil
		}

		// If a handler is currently executing, set its WaitCounter
		// Requirement 17.6: System maintains separate wait counter for each handler.
		if v.currentHandler != nil {
			v.currentHandler.WaitCounter = waitCount
			v.log.Debug("Wait: handler wait counter set", "handler", v.currentHandler.ID, "waitCount", waitCount)
			// Return a wait marker to signal the handler should pause
			return &waitMarker{WaitCount: waitCount}, nil
		}

		// If no handler is executing (e.g., in main code), we can't really wait
		// Log a warning and continue
		v.log.Warn("Wait called outside of event handler, ignoring", "waitCount", waitCount)
		return nil, nil
	})

	// ExitTitle: Terminate the program
	// Requirement 10.7: When ExitTitle is called, system terminates program.
	// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
	// Requirement 15.2: When ExitTitle is called, system closes all windows.
	// Requirement 15.3: When ExitTitle is called, system cleans up all resources.
	// Requirement 15.4: When ExitTitle is called, system terminates event loop.
	// Requirement 15.7: System provides graceful shutdown mechanism.
	vm.RegisterBuiltinFunction("ExitTitle", func(v *VM, args []any) (any, error) {
		v.log.Info("ExitTitle called, initiating graceful shutdown")

		// Stop all audio playback
		// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
		if v.audioSystem != nil {
			v.audioSystem.StopTimer()
			v.audioSystem.Shutdown()
		}

		// Remove all event handlers
		v.handlerRegistry.UnregisterAll()

		// Stop the VM
		// Requirement 15.4: When ExitTitle is called, system terminates event loop.
		v.Stop()

		return nil, nil
	})

	// Dummy functions for unimplemented graphics features
	// These will be properly implemented in a separate spec for graphics system

	// LoadPic: Load a picture (dummy implementation)
	vm.RegisterBuiltinFunction("LoadPic", func(v *VM, args []any) (any, error) {
		v.log.Debug("LoadPic called (dummy implementation)", "args", args)
		return nil, nil
	})

	// CreatePic: Create a picture (dummy implementation)
	vm.RegisterBuiltinFunction("CreatePic", func(v *VM, args []any) (any, error) {
		v.log.Debug("CreatePic called (dummy implementation)", "args", args)
		return nil, nil
	})

	// MovePic: Move a picture (dummy implementation)
	vm.RegisterBuiltinFunction("MovePic", func(v *VM, args []any) (any, error) {
		v.log.Debug("MovePic called (dummy implementation)", "args", args)
		return nil, nil
	})

	// DelPic: Delete a picture (dummy implementation)
	vm.RegisterBuiltinFunction("DelPic", func(v *VM, args []any) (any, error) {
		v.log.Debug("DelPic called (dummy implementation)", "args", args)
		return nil, nil
	})

	// OpenWin: Open a window (dummy implementation)
	vm.RegisterBuiltinFunction("OpenWin", func(v *VM, args []any) (any, error) {
		v.log.Debug("OpenWin called (dummy implementation)", "args", args)
		return nil, nil
	})

	// CloseWin: Close a window (dummy implementation)
	vm.RegisterBuiltinFunction("CloseWin", func(v *VM, args []any) (any, error) {
		v.log.Debug("CloseWin called (dummy implementation)", "args", args)
		return nil, nil
	})

	// MoveWin: Move a window (dummy implementation)
	vm.RegisterBuiltinFunction("MoveWin", func(v *VM, args []any) (any, error) {
		v.log.Debug("MoveWin called (dummy implementation)", "args", args)
		return nil, nil
	})

	// PutCast: Put a cast (dummy implementation)
	vm.RegisterBuiltinFunction("PutCast", func(v *VM, args []any) (any, error) {
		v.log.Debug("PutCast called (dummy implementation)", "args", args)
		return nil, nil
	})

	// MoveCast: Move a cast (dummy implementation)
	vm.RegisterBuiltinFunction("MoveCast", func(v *VM, args []any) (any, error) {
		v.log.Debug("MoveCast called (dummy implementation)", "args", args)
		return nil, nil
	})

	// DelCast: Delete a cast (dummy implementation)
	vm.RegisterBuiltinFunction("DelCast", func(v *VM, args []any) (any, error) {
		v.log.Debug("DelCast called (dummy implementation)", "args", args)
		return nil, nil
	})

	// TextWrite: Write text (dummy implementation)
	vm.RegisterBuiltinFunction("TextWrite", func(v *VM, args []any) (any, error) {
		v.log.Debug("TextWrite called (dummy implementation)", "args", args)
		return nil, nil
	})

	// DrawRect: Draw a rectangle (dummy implementation)
	vm.RegisterBuiltinFunction("DrawRect", func(v *VM, args []any) (any, error) {
		v.log.Debug("DrawRect called (dummy implementation)", "args", args)
		return nil, nil
	})

	// DrawLine: Draw a line (dummy implementation)
	vm.RegisterBuiltinFunction("DrawLine", func(v *VM, args []any) (any, error) {
		v.log.Debug("DrawLine called (dummy implementation)", "args", args)
		return nil, nil
	})

	// FillRect: Fill a rectangle (dummy implementation)
	vm.RegisterBuiltinFunction("FillRect", func(v *VM, args []any) (any, error) {
		v.log.Debug("FillRect called (dummy implementation)", "args", args)
		return nil, nil
	})

	// SetColor: Set color (dummy implementation)
	vm.RegisterBuiltinFunction("SetColor", func(v *VM, args []any) (any, error) {
		v.log.Debug("SetColor called (dummy implementation)", "args", args)
		return nil, nil
	})
}

// RegisterBuiltinFunction registers a built-in function with the given name.
// Requirement 10.9: System provides registry of built-in functions.
// Requirement 10.10: System allows registration of custom built-in functions.
func (vm *VM) RegisterBuiltinFunction(name string, fn BuiltinFunc) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	vm.builtins[name] = fn
}

// Run starts the VM execution loop.
// It processes OpCodes sequentially until completion, stop, or timeout.
// After initial OpCode execution, it enters an event loop to process events.
//
// Requirement 14.1: System runs main event loop that processes events and executes OpCode.
// Requirement 14.2: When event queue is empty, system waits for next event.
// Requirement 13.1: When timeout is specified, system terminates execution after specified duration.
//
// Returns:
//   - error: Any error that occurred during execution
func (vm *VM) Run() error {
	vm.mu.Lock()
	if vm.running {
		vm.mu.Unlock()
		return fmt.Errorf("VM is already running")
	}
	vm.running = true
	vm.mu.Unlock()

	defer func() {
		vm.mu.Lock()
		vm.running = false
		vm.mu.Unlock()
	}()

	// Set up timeout if specified
	// Requirement 13.1: When timeout is specified, system terminates execution after specified duration.
	if vm.timeout > 0 {
		var timeoutCancel context.CancelFunc
		vm.ctx, timeoutCancel = context.WithTimeout(vm.ctx, vm.timeout)
		defer timeoutCancel()
	}

	vm.log.Info("VM started", "opcode_count", len(vm.opcodes), "headless", vm.headless, "timeout", vm.timeout)

	// First pass: collect function definitions
	if err := vm.collectFunctionDefinitions(); err != nil {
		return fmt.Errorf("failed to collect function definitions: %w", err)
	}

	// Call main function if it exists
	// This is the entry point for FILLY scripts
	if mainFunc, ok := vm.functions["main"]; ok {
		vm.log.Info("Calling main function")
		if _, err := vm.callUserFunction(mainFunc, []any{}); err != nil {
			vm.log.Error("main function execution failed", "error", err)
			return fmt.Errorf("main function execution failed: %w", err)
		}
		vm.log.Info("main function completed")
	}

	// Execute initial OpCodes (main function)
	// Requirement 8.1: When VM receives OpCode sequence, system executes each OpCode in order.
	for vm.pc < len(vm.opcodes) {
		// Check for cancellation (timeout or stop)
		select {
		case <-vm.ctx.Done():
			if vm.ctx.Err() == context.DeadlineExceeded {
				// Requirement 13.3: When timeout expires, system logs timeout message.
				vm.log.Info("VM execution timed out")
				return nil
			}
			vm.log.Info("VM execution cancelled")
			return nil
		default:
		}

		// Execute current OpCode
		opcode := vm.opcodes[vm.pc]
		_, err := vm.Execute(opcode)
		if err != nil {
			// Log error but continue execution for non-fatal errors
			// Requirement 11.8: System continues execution after non-fatal errors.
			vm.log.Error("OpCode execution error", "pc", vm.pc, "cmd", opcode.Cmd, "error", err)
		}

		vm.pc++
	}

	vm.log.Info("VM initial execution completed, entering event loop")

	// Enter event loop
	// Requirement 14.1: System runs main event loop that processes events and executes OpCode.
	// Requirement 15.6: When main function completes, system continues event processing.
	return vm.runEventLoop()
}

// runEventLoop runs the main event loop.
// It processes events from the queue and dispatches them to registered handlers.
//
// Requirement 14.1: System runs main event loop that processes events and executes OpCode.
// Requirement 14.2: When event queue is empty, system waits for next event.
// Requirement 14.3: When events are available, system processes them in order.
// Requirement 14.4: When OpCode execution is in progress, system continues until wait point.
// Requirement 14.5: When wait point is reached, system returns control to event loop.
// Requirement 14.6: System maintains balance between event processing and OpCode execution.
func (vm *VM) runEventLoop() error {
	// If no handlers are registered, exit immediately
	// This allows simple scripts without event handlers to complete
	if vm.handlerRegistry.Count() == 0 {
		vm.log.Info("No event handlers registered, exiting event loop")
		return nil
	}

	vm.log.Info("Event loop started", "handler_count", vm.handlerRegistry.Count())

	for {
		// Check for cancellation (timeout or stop)
		select {
		case <-vm.ctx.Done():
			if vm.ctx.Err() == context.DeadlineExceeded {
				// Requirement 13.3: When timeout expires, system logs timeout message.
				vm.log.Info("Event loop timed out")
				return nil
			}
			vm.log.Info("Event loop cancelled")
			return nil
		default:
		}

		// Update audio system to generate MIDI_TIME and MIDI_END events
		// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
		// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
		vm.UpdateAudio()

		// Process events from the queue
		// Requirement 14.3: When events are available, system processes them in order.
		processed, err := vm.eventDispatcher.ProcessOne()
		if err != nil {
			vm.log.Error("Event processing error", "error", err)
		}

		// If no events were processed, check if we should continue
		if !processed {
			// Requirement 14.2: When event queue is empty, system waits for next event.
			// In a real implementation, this would wait for events from the timer,
			// audio system, or input system. For now, we just check if there are
			// any handlers left.
			if vm.handlerRegistry.Count() == 0 {
				vm.log.Info("All handlers removed, exiting event loop")
				return nil
			}

			// Small sleep to prevent busy-waiting
			// In a real implementation with Ebitengine, this would be handled
			// by the game loop's Update() method
			time.Sleep(1 * time.Millisecond)
		}
	}
}

// collectFunctionDefinitions scans OpCodes for function definitions and registers them.
func (vm *VM) collectFunctionDefinitions() error {
	for _, opcode := range vm.opcodes {
		if opcode.Cmd == compiler.OpDefineFunction {
			if err := vm.registerFunction(opcode); err != nil {
				return err
			}
		}
	}
	return nil
}

// registerFunction registers a function definition from an OpDefineFunction OpCode.
func (vm *VM) registerFunction(opcode compiler.OpCode) error {
	if len(opcode.Args) < 3 {
		return fmt.Errorf("OpDefineFunction requires 3 arguments, got %d", len(opcode.Args))
	}

	name, ok := opcode.Args[0].(string)
	if !ok {
		return fmt.Errorf("function name must be string, got %T", opcode.Args[0])
	}

	// Parse parameters
	var params []FunctionParam
	if paramsRaw, ok := opcode.Args[1].([]any); ok {
		for _, p := range paramsRaw {
			if paramMap, ok := p.(map[string]any); ok {
				param := FunctionParam{
					Name:    paramMap["name"].(string),
					Type:    paramMap["type"].(string),
					IsArray: paramMap["isArray"].(bool),
				}
				if defaultVal, hasDefault := paramMap["default"]; hasDefault {
					param.Default = defaultVal
					param.HasDefault = true
				}
				params = append(params, param)
			}
		}
	}

	// Get body OpCodes
	body, ok := opcode.Args[2].([]compiler.OpCode)
	if !ok {
		return fmt.Errorf("function body must be []OpCode, got %T", opcode.Args[2])
	}

	vm.functions[name] = &FunctionDef{
		Name:       name,
		Parameters: params,
		Body:       body,
	}

	vm.log.Debug("Function registered", "name", name, "params", len(params))
	return nil
}

// Stop stops the VM execution.
// Requirement 15.4: When ExitTitle is called, system terminates event loop.
func (vm *VM) Stop() {
	vm.mu.Lock()
	defer vm.mu.Unlock()

	if vm.running {
		vm.cancel()
		vm.log.Info("VM stop requested")
	}
}

// Execute executes a single OpCode and returns the result.
// This is the main dispatch method that routes OpCodes to their handlers.
//
// Requirement 8.1: When VM receives OpCode sequence, system executes each OpCode in order.
//
// Parameters:
//   - opcode: The OpCode to execute
//
// Returns:
//   - any: The result of the OpCode execution (may be nil)
//   - error: Any error that occurred during execution
func (vm *VM) Execute(opcode compiler.OpCode) (any, error) {
	vm.log.Debug("Executing OpCode", "cmd", opcode.Cmd, "pc", vm.pc)

	switch opcode.Cmd {
	case compiler.OpAssign:
		return vm.executeAssign(opcode)
	case compiler.OpArrayAssign:
		return vm.executeArrayAssign(opcode)
	case compiler.OpCall:
		return vm.executeCall(opcode)
	case compiler.OpBinaryOp:
		return vm.executeBinaryOp(opcode)
	case compiler.OpUnaryOp:
		return vm.executeUnaryOp(opcode)
	case compiler.OpArrayAccess:
		return vm.executeArrayAccess(opcode)
	case compiler.OpIf:
		return vm.executeIf(opcode)
	case compiler.OpFor:
		return vm.executeFor(opcode)
	case compiler.OpWhile:
		return vm.executeWhile(opcode)
	case compiler.OpSwitch:
		return vm.executeSwitch(opcode)
	case compiler.OpBreak:
		return vm.executeBreak(opcode)
	case compiler.OpContinue:
		return vm.executeContinue(opcode)
	case compiler.OpRegisterEventHandler:
		return vm.executeRegisterEventHandler(opcode)
	case compiler.OpWait:
		return vm.executeWait(opcode)
	case compiler.OpSetStep:
		return vm.executeSetStep(opcode)
	case compiler.OpDefineFunction:
		// Function definitions are processed in collectFunctionDefinitions
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown OpCode command: %s", opcode.Cmd)
	}
}

// IsRunning returns whether the VM is currently running.
func (vm *VM) IsRunning() bool {
	vm.mu.RLock()
	defer vm.mu.RUnlock()
	return vm.running
}

// GetGlobalScope returns the global scope.
func (vm *VM) GetGlobalScope() *Scope {
	return vm.globalScope
}

// GetCurrentScope returns the current scope (local if in function, global otherwise).
func (vm *VM) GetCurrentScope() *Scope {
	if vm.localScope != nil {
		return vm.localScope
	}
	return vm.globalScope
}

// PushStackFrame pushes a new stack frame for a function call.
// Requirement 20.1: When function is called, system pushes new stack frame.
// Requirement 20.6: System detects stack overflow and reports error.
// Requirement 20.7: System maintains maximum stack depth of 1000 frames.
func (vm *VM) PushStackFrame(functionName string, localScope *Scope) error {
	if len(vm.callStack) >= MaxStackDepth {
		// Requirement 20.8: When stack overflow occurs, system logs error and terminates execution.
		return fmt.Errorf("stack overflow: maximum depth %d exceeded", MaxStackDepth)
	}

	frame := &StackFrame{
		FunctionName: functionName,
		LocalScope:   localScope,
		ReturnPC:     vm.pc,
	}
	vm.callStack = append(vm.callStack, frame)
	vm.localScope = localScope

	vm.log.Debug("Stack frame pushed", "function", functionName, "depth", len(vm.callStack))
	return nil
}

// PopStackFrame pops the current stack frame after a function returns.
// Requirement 20.2: When function returns, system pops stack frame.
func (vm *VM) PopStackFrame() (*StackFrame, error) {
	if len(vm.callStack) == 0 {
		return nil, fmt.Errorf("cannot pop from empty call stack")
	}

	frame := vm.callStack[len(vm.callStack)-1]
	vm.callStack = vm.callStack[:len(vm.callStack)-1]

	// Restore previous local scope
	if len(vm.callStack) > 0 {
		vm.localScope = vm.callStack[len(vm.callStack)-1].LocalScope
	} else {
		vm.localScope = nil
	}

	vm.log.Debug("Stack frame popped", "function", frame.FunctionName, "depth", len(vm.callStack))
	return frame, nil
}

// GetStackDepth returns the current call stack depth.
func (vm *VM) GetStackDepth() int {
	return len(vm.callStack)
}

// breakSignal is a special type to signal a break from a loop.
type breakSignal struct{}

// continueSignal is a special type to signal a continue in a loop.
type continueSignal struct{}

// waitMarker is a special type to signal that execution should pause and wait for events.
// Requirement 6.2: When OpWait is executed, system pauses execution until next event.
type waitMarker struct {
	WaitCount int // Number of events to wait for
}

// executeIf executes an OpIf OpCode.
// OpIf evaluates a condition and executes the appropriate branch.
// Args: [condition, thenBlock []OpCode, elseBlock []OpCode]
//
// Requirement 8.5: When OpIf is executed, system evaluates condition and executes appropriate branch.
func (vm *VM) executeIf(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpIf requires at least 2 arguments, got %d", len(opcode.Args))
	}

	// Evaluate the condition
	conditionVal, err := vm.evaluateValue(opcode.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate condition: %w", err)
	}

	condition := toBool(conditionVal)
	vm.log.Debug("If condition evaluated", "condition", condition)

	if condition {
		// Execute then block
		thenBlock, ok := opcode.Args[1].([]compiler.OpCode)
		if !ok {
			return nil, fmt.Errorf("OpIf then block must be []OpCode, got %T", opcode.Args[1])
		}
		return vm.executeBlock(thenBlock)
	} else if len(opcode.Args) >= 3 {
		// Execute else block if present
		elseBlock, ok := opcode.Args[2].([]compiler.OpCode)
		if !ok {
			return nil, fmt.Errorf("OpIf else block must be []OpCode, got %T", opcode.Args[2])
		}
		return vm.executeBlock(elseBlock)
	}

	return nil, nil
}

// executeFor executes an OpFor OpCode.
// OpFor executes a for loop with init, condition, post, and body.
// Args: [initBlock []OpCode, condition, postBlock []OpCode, bodyBlock []OpCode]
//
// Requirement 8.6: When OpFor is executed, system executes loop with init, condition, increment.
func (vm *VM) executeFor(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 4 {
		return nil, fmt.Errorf("OpFor requires 4 arguments, got %d", len(opcode.Args))
	}

	// Execute init block
	if initBlock, ok := opcode.Args[0].([]compiler.OpCode); ok && len(initBlock) > 0 {
		if _, err := vm.executeBlock(initBlock); err != nil {
			return nil, fmt.Errorf("failed to execute for init: %w", err)
		}
	}

	// Get body and post blocks
	bodyBlock, ok := opcode.Args[3].([]compiler.OpCode)
	if !ok {
		return nil, fmt.Errorf("OpFor body must be []OpCode, got %T", opcode.Args[3])
	}

	postBlock, _ := opcode.Args[2].([]compiler.OpCode)

	// Loop
	var lastResult any
	for {
		// Check condition (if present)
		if opcode.Args[1] != nil {
			conditionVal, err := vm.evaluateValue(opcode.Args[1])
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate for condition: %w", err)
			}
			if !toBool(conditionVal) {
				break
			}
		}

		// Execute body
		result, err := vm.executeBlock(bodyBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to execute for body: %w", err)
		}

		// Check for break/continue signals
		if _, isBreak := result.(*breakSignal); isBreak {
			// Requirement 8.9: When OpBreak is executed, system exits current loop.
			break
		}
		if _, isContinue := result.(*continueSignal); isContinue {
			// Requirement 8.10: When OpContinue is executed, system skips to next loop iteration.
			// Continue to post block
		} else {
			lastResult = result
		}

		// Execute post block
		if len(postBlock) > 0 {
			if _, err := vm.executeBlock(postBlock); err != nil {
				return nil, fmt.Errorf("failed to execute for post: %w", err)
			}
		}
	}

	return lastResult, nil
}

// executeWhile executes an OpWhile OpCode.
// OpWhile executes a while loop with condition and body.
// Args: [condition, bodyBlock []OpCode]
//
// Requirement 8.7: When OpWhile is executed, system executes loop while condition is true.
func (vm *VM) executeWhile(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpWhile requires 2 arguments, got %d", len(opcode.Args))
	}

	bodyBlock, ok := opcode.Args[1].([]compiler.OpCode)
	if !ok {
		return nil, fmt.Errorf("OpWhile body must be []OpCode, got %T", opcode.Args[1])
	}

	var lastResult any
	for {
		// Check condition
		if opcode.Args[0] != nil {
			conditionVal, err := vm.evaluateValue(opcode.Args[0])
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate while condition: %w", err)
			}
			if !toBool(conditionVal) {
				break
			}
		}

		// Execute body
		result, err := vm.executeBlock(bodyBlock)
		if err != nil {
			return nil, fmt.Errorf("failed to execute while body: %w", err)
		}

		// Check for break/continue signals
		if _, isBreak := result.(*breakSignal); isBreak {
			// Requirement 8.9: When OpBreak is executed, system exits current loop.
			break
		}
		if _, isContinue := result.(*continueSignal); isContinue {
			// Requirement 8.10: When OpContinue is executed, system skips to next loop iteration.
			continue
		}

		lastResult = result
	}

	return lastResult, nil
}

// executeSwitch executes an OpSwitch OpCode.
// OpSwitch evaluates a value and executes the matching case.
// Args: [value, cases []any, defaultBlock []OpCode]
// Each case is a map[string]any with "value" and "body" keys.
//
// Requirement 8.8: When OpSwitch is executed, system evaluates value and executes matching case.
func (vm *VM) executeSwitch(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpSwitch requires at least 2 arguments, got %d", len(opcode.Args))
	}

	// Evaluate the switch value
	switchVal, err := vm.evaluateValue(opcode.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate switch value: %w", err)
	}

	vm.log.Debug("Switch value evaluated", "value", switchVal)

	// Get cases
	cases, ok := opcode.Args[1].([]any)
	if !ok {
		return nil, fmt.Errorf("OpSwitch cases must be []any, got %T", opcode.Args[1])
	}

	// Find matching case
	for _, c := range cases {
		caseClause, ok := c.(map[string]any)
		if !ok {
			continue
		}

		// Evaluate case value
		caseVal, err := vm.evaluateValue(caseClause["value"])
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate case value: %w", err)
		}

		// Compare values
		if vm.valuesEqual(switchVal, caseVal) {
			// Execute case body
			caseBody, ok := caseClause["body"].([]compiler.OpCode)
			if !ok {
				return nil, fmt.Errorf("case body must be []OpCode, got %T", caseClause["body"])
			}
			return vm.executeBlock(caseBody)
		}
	}

	// No matching case, execute default if present
	if len(opcode.Args) >= 3 && opcode.Args[2] != nil {
		defaultBlock, ok := opcode.Args[2].([]compiler.OpCode)
		if !ok {
			return nil, fmt.Errorf("OpSwitch default block must be []OpCode, got %T", opcode.Args[2])
		}
		return vm.executeBlock(defaultBlock)
	}

	return nil, nil
}

// executeBreak executes an OpBreak OpCode.
// OpBreak signals a break from the current loop.
// Args: []
//
// Requirement 8.9: When OpBreak is executed, system exits current loop.
func (vm *VM) executeBreak(_ compiler.OpCode) (any, error) {
	vm.log.Debug("Break executed")
	return &breakSignal{}, nil
}

// executeContinue executes an OpContinue OpCode.
// OpContinue signals a continue to the next loop iteration.
// Args: []
//
// Requirement 8.10: When OpContinue is executed, system skips to next loop iteration.
func (vm *VM) executeContinue(_ compiler.OpCode) (any, error) {
	vm.log.Debug("Continue executed")
	return &continueSignal{}, nil
}

// executeBlock executes a block of OpCodes and returns the last result.
// It handles break, continue, and wait signals by propagating them up.
func (vm *VM) executeBlock(opcodes []compiler.OpCode) (any, error) {
	var lastResult any
	for _, op := range opcodes {
		result, err := vm.Execute(op)
		if err != nil {
			// Log error but continue execution for non-fatal errors
			// Requirement 11.8: System continues execution after non-fatal errors.
			vm.log.Error("OpCode execution error in block", "cmd", op.Cmd, "error", err)
			continue
		}

		// Check for break/continue signals - propagate them up
		if _, isBreak := result.(*breakSignal); isBreak {
			return result, nil
		}
		if _, isContinue := result.(*continueSignal); isContinue {
			return result, nil
		}

		// Check for return marker - propagate it up
		if _, isReturn := result.(*returnMarker); isReturn {
			return result, nil
		}

		// Check for wait marker - propagate it up
		// Requirement 6.2: When OpWait is executed, system pauses execution until next event.
		if _, isWait := result.(*waitMarker); isWait {
			return result, nil
		}

		lastResult = result
	}
	return lastResult, nil
}

// valuesEqual compares two values for equality.
// It handles type coercion for numeric comparisons.
func (vm *VM) valuesEqual(a, b any) bool {
	// Handle nil cases
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Try numeric comparison
	aInt, aIsInt := toInt64(a)
	bInt, bIsInt := toInt64(b)
	if aIsInt && bIsInt {
		return aInt == bInt
	}

	aFloat, aIsFloat := toFloat64(a)
	bFloat, bIsFloat := toFloat64(b)
	if aIsFloat && bIsFloat {
		return aFloat == bFloat
	}

	// String comparison
	aStr, aIsStr := a.(string)
	bStr, bIsStr := b.(string)
	if aIsStr && bIsStr {
		return aStr == bStr
	}

	// Default: direct comparison
	return a == b
}

// executeRegisterEventHandler executes an OpRegisterEventHandler OpCode.
// OpRegisterEventHandler registers an event handler for mes() blocks.
// Args: [eventType string, bodyBlock []OpCode]
//
// Requirement 2.1: When OpRegisterEventHandler is executed, system registers handler for specified event type.
// Requirement 2.2: When mes(TIME) handler is registered, system calls it on each timer tick.
// Requirement 2.3: When mes(MIDI_TIME) handler is registered, system calls it on each MIDI tick.
// Requirement 2.4: When mes(MIDI_END) handler is registered, system calls it when MIDI playback completes.
// Requirement 2.5: When mes(LBDOWN) handler is registered, system calls it on left mouse button press.
// Requirement 2.6: When mes(RBDOWN) handler is registered, system calls it on right mouse button press.
// Requirement 2.7: When mes(RBDBLCLK) handler is registered, system calls it on right mouse button double-click.
// Requirement 2.8: When handler is registered inside another handler, system supports nested handler registration.
func (vm *VM) executeRegisterEventHandler(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpRegisterEventHandler requires 2 arguments, got %d", len(opcode.Args))
	}

	// Get event type
	eventTypeStr, ok := opcode.Args[0].(string)
	if !ok {
		return nil, fmt.Errorf("OpRegisterEventHandler event type must be string, got %T", opcode.Args[0])
	}

	// Convert string to EventType
	eventType := EventType(eventTypeStr)

	// Validate event type
	switch eventType {
	case EventTIME, EventMIDI_TIME, EventMIDI_END, EventLBDOWN, EventRBDOWN, EventRBDBLCLK:
		// Valid event type
	default:
		return nil, fmt.Errorf("unknown event type: %s", eventTypeStr)
	}

	// Get body OpCodes
	bodyOpcodes, ok := opcode.Args[1].([]compiler.OpCode)
	if !ok {
		return nil, fmt.Errorf("OpRegisterEventHandler body must be []OpCode, got %T", opcode.Args[1])
	}

	// Create and register the handler
	handler := NewEventHandler("", eventType, bodyOpcodes, vm)
	id := vm.handlerRegistry.Register(handler)

	vm.log.Debug("Event handler registered", "id", id, "eventType", eventType, "opcodeCount", len(bodyOpcodes))

	// Automatically start timer when TIME event handler is registered
	// Requirement 2.2: When mes(TIME) handler is registered, system calls it on each timer tick.
	if eventType == EventTIME {
		vm.StartTimer()
		vm.log.Debug("Timer started automatically for TIME event handler")
	}

	return id, nil
}

func (vm *VM) executeWait(opcode compiler.OpCode) (any, error) {
	// Requirement 6.2: When OpWait OpCode is executed, system pauses execution until next event.
	// Requirement 6.3: When event occurs during step execution, system proceeds to next step.

	// Get the comma count from arguments (number of commas in the step block)
	commaCount := 1 // Default to 1 if not specified
	if len(opcode.Args) >= 1 {
		waitValue, err := vm.evaluateValue(opcode.Args[0])
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate comma count: %w", err)
		}

		if wc, ok := toInt64(waitValue); ok {
			commaCount = int(wc)
		} else if f, fok := toFloat64(waitValue); fok {
			commaCount = int(f)
		}
	}

	// Requirement 6.10: When step(n) is called with n=0, system executes immediately without waiting.
	// Requirement 17.4: When Wait(0) is called, system continues execution immediately.
	// Requirement 17.5: When Wait(n) is called with n<0, system treats it as Wait(0).
	if commaCount <= 0 {
		vm.log.Debug("OpWait: comma count <= 0, continuing immediately", "commaCount", commaCount)
		return nil, nil
	}

	// If a handler is currently executing, calculate wait count based on step value
	if vm.currentHandler != nil {
		// StepCounter holds the step value (n from step(n))
		// Each comma waits for stepValue TIME events
		// Since TIME events are generated every 50ms, step(n) with one comma waits n × 50ms
		// For example: step(65) with ,, means wait for 65 × 2 = 130 TIME events = 6500ms
		stepValue := vm.currentHandler.StepCounter
		if stepValue <= 0 {
			stepValue = 1 // Default to 1 event per comma if not set
		}

		waitCount := commaCount * stepValue
		vm.currentHandler.WaitCounter = waitCount
		vm.log.Debug("OpWait: handler wait counter set", "handler", vm.currentHandler.ID, "commas", commaCount, "stepValue", stepValue, "totalWaitCount", waitCount)
		// Return a wait marker to signal the handler should pause
		return &waitMarker{WaitCount: waitCount}, nil
	}

	// If no handler is executing (e.g., in main code), we can't really wait
	// Log a warning and continue
	vm.log.Warn("OpWait called outside of event handler, ignoring", "commaCount", commaCount)
	return nil, nil
}

func (vm *VM) executeSetStep(opcode compiler.OpCode) (any, error) {
	// Requirement 6.1: When OpSetStep OpCode is executed, system initializes step counter with specified count.
	// The step count represents the number of TIME events to wait per comma in step() blocks.
	// Since TIME events are generated every 50ms, step(n) means each comma waits n × 50ms.
	// For example, step(65) means each comma waits 65 × 50ms = 3250ms.
	if len(opcode.Args) < 1 {
		return nil, fmt.Errorf("OpSetStep requires 1 argument, got %d", len(opcode.Args))
	}

	// Evaluate the step value (number of TIME events per comma)
	stepValue, err := vm.evaluateValue(opcode.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate step value: %w", err)
	}

	// Convert to int (number of TIME events)
	stepCount, ok := toInt64(stepValue)
	if !ok {
		// Try to convert from float
		if f, fok := toFloat64(stepValue); fok {
			stepCount = int64(f)
		} else {
			vm.log.Warn("OpSetStep: invalid step value, using 0", "value", stepValue)
			stepCount = 0
		}
	}

	// If a handler is currently executing, set its step count
	// StepCounter stores the number of TIME events to wait per comma
	// Otherwise, store in the VM for later use
	if vm.currentHandler != nil {
		vm.currentHandler.StepCounter = int(stepCount)
		vm.log.Debug("OpSetStep: handler step count set", "handler", vm.currentHandler.ID, "eventsPerComma", stepCount)
	} else {
		vm.stepCounter = int(stepCount)
		vm.log.Debug("OpSetStep: VM step count set", "eventsPerComma", stepCount)
	}

	return nil, nil
}

// GetEventQueue returns the event queue.
func (vm *VM) GetEventQueue() *EventQueue {
	return vm.eventQueue
}

// GetHandlerRegistry returns the handler registry.
func (vm *VM) GetHandlerRegistry() *HandlerRegistry {
	return vm.handlerRegistry
}

// GetEventDispatcher returns the event dispatcher.
func (vm *VM) GetEventDispatcher() *EventDispatcher {
	return vm.eventDispatcher
}

// GetCurrentHandler returns the currently executing handler.
func (vm *VM) GetCurrentHandler() *EventHandler {
	return vm.currentHandler
}

// SetCurrentHandler sets the currently executing handler.
func (vm *VM) SetCurrentHandler(handler *EventHandler) {
	vm.currentHandler = handler
}

// GetStepCounter returns the VM's step counter.
// This is used when no handler is executing.
// Requirement 6.1: When OpSetStep is executed, system initializes step counter.
func (vm *VM) GetStepCounter() int {
	return vm.stepCounter
}

// SetStepCounter sets the VM's step counter.
func (vm *VM) SetStepCounter(count int) {
	vm.stepCounter = count
}

// QueueEvent adds an event to the event queue.
func (vm *VM) QueueEvent(event *Event) {
	vm.eventQueue.Push(event)
}

// ProcessEvents processes all pending events in the queue.
func (vm *VM) ProcessEvents() error {
	return vm.eventDispatcher.ProcessQueue()
}

// SetAudioSystem sets the audio system.
// This allows external initialization of the audio system to avoid import cycles.
//
// Parameters:
//   - audioSys: The audio system implementing AudioSystemInterface
func (vm *VM) SetAudioSystem(audioSys AudioSystemInterface) {
	vm.audioSystem = audioSys

	// Mute audio in headless mode
	// Requirement 12.2: When headless mode is enabled, system mutes all audio output.
	if vm.headless && audioSys != nil {
		audioSys.SetMuted(true)
	}

	vm.log.Info("Audio system set", "muted", vm.headless)
}

// GetAudioSystem returns the audio system.
// Returns nil if the audio system has not been initialized.
func (vm *VM) GetAudioSystem() AudioSystemInterface {
	return vm.audioSystem
}

// UpdateAudio updates the audio system.
// This should be called from the game loop to generate MIDI_TIME events.
//
// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
func (vm *VM) UpdateAudio() {
	if vm.audioSystem != nil {
		vm.audioSystem.Update()
	}
}

// ShutdownAudio shuts down the audio system and releases resources.
//
// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
// Requirement 15.3: When ExitTitle is called, system cleans up all resources.
func (vm *VM) ShutdownAudio() {
	if vm.audioSystem != nil {
		vm.audioSystem.Shutdown()
		vm.log.Info("Audio system shut down")
	}
}

// PlayMIDI plays a MIDI file through the audio system.
//
// Requirement 4.1: When PlayMIDI(filename) is called, system starts playback of specified MIDI file.
func (vm *VM) PlayMIDI(filename string) error {
	if vm.audioSystem == nil {
		return fmt.Errorf("audio system not initialized")
	}

	// Resolve relative path using titlePath
	fullPath := vm.resolveFilePath(filename)
	return vm.audioSystem.PlayMIDI(fullPath)
}

// PlayWAVE plays a WAV file through the audio system.
//
// Requirement 5.1: When PlayWAVE(filename) is called, system starts playback of specified WAV file.
func (vm *VM) PlayWAVE(filename string) error {
	if vm.audioSystem == nil {
		return fmt.Errorf("audio system not initialized")
	}

	// Resolve relative path using titlePath
	fullPath := vm.resolveFilePath(filename)
	return vm.audioSystem.PlayWAVE(fullPath)
}

// resolveFilePath resolves a relative file path using the title path.
// If the path is already absolute, it is returned as-is.
func (vm *VM) resolveFilePath(filename string) string {
	// If path is absolute, return as-is
	if filepath.IsAbs(filename) {
		return filename
	}

	// If titlePath is set, join with it
	if vm.titlePath != "" {
		return filepath.Join(vm.titlePath, filename)
	}

	// Otherwise, return as-is (relative to current directory)
	return filename
}

// StartTimer starts the timer for TIME event generation.
//
// Requirement 3.1: System generates TIME events periodically.
func (vm *VM) StartTimer() {
	if vm.audioSystem != nil {
		vm.audioSystem.StartTimer()
	}
}

// StopTimer stops the timer.
func (vm *VM) StopTimer() {
	if vm.audioSystem != nil {
		vm.audioSystem.StopTimer()
	}
}

// GetSoundFontPath returns the configured SoundFont path.
func (vm *VM) GetSoundFontPath() string {
	return vm.soundFontPath
}

// IsHeadless returns whether the VM is running in headless mode.
func (vm *VM) IsHeadless() bool {
	return vm.headless
}
