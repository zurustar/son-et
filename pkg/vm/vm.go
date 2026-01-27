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
	"image/color"
	"log/slog"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/zurustar/son-et/pkg/compiler"
	"github.com/zurustar/son-et/pkg/graphics"
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

	// Graphics system interface (to avoid import cycle)
	graphicsSystem GraphicsSystemInterface

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

// GraphicsSystemInterface defines the interface for graphics system operations.
// This interface is used to avoid import cycles between vm and graphics packages.
type GraphicsSystemInterface interface {
	// Picture management
	LoadPic(filename string) (int, error)
	CreatePic(width, height int) (int, error)
	CreatePicFrom(srcID int) (int, error)
	CreatePicWithSize(srcID, width, height int) (int, error)
	DelPic(id int) error
	PicWidth(id int) int
	PicHeight(id int) int

	// Picture transfer
	MovePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode int) error
	MovePicWithSpeed(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, speed int) error
	MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH int) error
	TransPic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int, transColor any) error
	ReversePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int) error

	// Window management
	OpenWin(picID int, opts ...any) (int, error)
	MoveWin(id int, opts ...any) error
	CloseWin(id int) error
	CloseWinAll()
	CapTitle(id int, title string) error
	CapTitleAll(title string)
	GetPicNo(id int) (int, error)
	GetWinByPicID(picID int) (int, error)

	// Cast management
	PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error)
	PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h int, transColor color.Color) (int, error)
	MoveCast(id int, opts ...any) error
	MoveCastWithOptions(id int, opts ...graphics.CastOption) error
	DelCast(id int) error

	// Text rendering
	TextWrite(picID, x, y int, text string) error
	SetFont(name string, size int, opts ...any) error
	SetTextColor(c any) error
	SetBgColor(c any) error
	SetBackMode(mode int) error

	// Drawing primitives
	DrawLine(picID, x1, y1, x2, y2 int) error
	DrawRect(picID, x1, y1, x2, y2, fillMode int) error
	FillRect(picID, x1, y1, x2, y2 int, c any) error
	DrawCircle(picID, x, y, radius, fillMode int) error
	SetLineSize(size int)
	SetPaintColor(c any) error
	GetColor(picID, x, y int) (int, error)

	// Virtual desktop info
	GetVirtualWidth() int
	GetVirtualHeight() int
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
			return nil, fmt.Errorf("PlayMIDI requires filename argument")
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
			return nil, fmt.Errorf("PlayWAVE requires filename argument")
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
			// ログは削除（頻繁すぎるため）
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

	// LoadPic: Load a picture
	vm.RegisterBuiltinFunction("LoadPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("LoadPic called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("LoadPic requires filename argument")
			return -1, nil
		}
		filename, ok := args[0].(string)
		if !ok {
			v.log.Error("LoadPic filename must be string", "got", fmt.Sprintf("%T", args[0]))
			return -1, nil
		}
		picID, err := v.graphicsSystem.LoadPic(filename)
		if err != nil {
			return -1, fmt.Errorf("LoadPic failed")
		}
		v.log.Debug("LoadPic called", "filename", filename, "picID", picID)
		return picID, nil
	})

	// CreatePic: Create a picture
	// Supports three patterns:
	// - CreatePic(srcPicID) - create from existing picture (same size)
	// - CreatePic(width, height) - create with specified size
	// - CreatePic(srcPicID, width, height) - create empty picture with specified size (source is for existence check only)
	vm.RegisterBuiltinFunction("CreatePic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CreatePic called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("CreatePic requires at least one argument")
			return -1, nil
		}

		// Check if it's a single argument (create from existing picture)
		if len(args) == 1 {
			srcID, ok := toInt64(args[0])
			if !ok {
				return -1, fmt.Errorf("CreatePic source picture ID must be integer")
			}
			picID, err := v.graphicsSystem.CreatePicFrom(int(srcID))
			if err != nil {
				return -1, fmt.Errorf("CreatePic (from source) failed")
			}
			v.log.Debug("CreatePic called (from source)", "srcID", srcID, "picID", picID)
			return picID, nil
		}

		// Three arguments: srcPicID, width, height
		if len(args) == 3 {
			srcID, sok := toInt64(args[0])
			width, wok := toInt64(args[1])
			height, hok := toInt64(args[2])
			if !sok || !wok || !hok {
				return -1, fmt.Errorf("CreatePic arguments must be integers")
			}
			picID, err := v.graphicsSystem.CreatePicWithSize(int(srcID), int(width), int(height))
			if err != nil {
				return -1, fmt.Errorf("CreatePic (with size) failed: %v", err)
			}
			v.log.Debug("CreatePic called (with size)", "srcID", srcID, "width", width, "height", height, "picID", picID)
			return picID, nil
		}

		// Two arguments: width and height
		width, wok := toInt64(args[0])
		height, hok := toInt64(args[1])
		if !wok || !hok {
			return -1, fmt.Errorf("CreatePic width and height must be integers")
		}
		picID, err := v.graphicsSystem.CreatePic(int(width), int(height))
		if err != nil {
			return -1, fmt.Errorf("CreatePic failed")
		}
		v.log.Debug("CreatePic called", "width", width, "height", height, "picID", picID)
		return picID, nil
	})

	// MovePic: Transfer picture region
	// MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y) - mode defaults to 0
	// MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, mode)
	// MovePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, mode, speed)
	// MovePic(src_pic, dst_pic) - copy entire picture (simplified form)
	vm.RegisterBuiltinFunction("MovePic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MovePic called but graphics system not initialized", "args", args)
			return nil, nil
		}

		// Simplified form: MovePic(src_pic, dst_pic) - copy entire picture
		if len(args) == 2 {
			srcID, sok := toInt64(args[0])
			dstID, dok := toInt64(args[1])
			if !sok || !dok {
				return nil, fmt.Errorf("MovePic: picture IDs must be integers")
			}
			// Get source picture dimensions
			srcW := v.graphicsSystem.PicWidth(int(srcID))
			srcH := v.graphicsSystem.PicHeight(int(srcID))
			if err := v.graphicsSystem.MovePic(int(srcID), 0, 0, srcW, srcH, int(dstID), 0, 0, 0); err != nil {
				v.log.Error("MovePic failed", "srcID", srcID, "dstID", dstID, "error", err)
			}
			return nil, nil
		}

		// Check for invalid argument counts (not 2, and less than 8)
		if len(args) < 8 {
			return nil, fmt.Errorf("MovePic: invalid argument count %d (expected 2, 8, 9, or 10)", len(args))
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		width, _ := toInt64(args[3])
		height, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])

		// mode is optional, defaults to 0 (normal copy)
		var mode int64 = 0
		if len(args) >= 9 {
			mode, _ = toInt64(args[8])
		}

		// Optional speed argument
		speed := 50 // default speed
		if len(args) >= 10 {
			if s, ok := toInt64(args[9]); ok {
				speed = int(s)
			}
		}

		var err error
		if len(args) >= 10 {
			err = v.graphicsSystem.MovePicWithSpeed(int(srcID), int(srcX), int(srcY), int(width), int(height),
				int(dstID), int(dstX), int(dstY), int(mode), speed)
		} else {
			err = v.graphicsSystem.MovePic(int(srcID), int(srcX), int(srcY), int(width), int(height),
				int(dstID), int(dstX), int(dstY), int(mode))
		}

		if err != nil {
			v.log.Error("MovePic failed", "error", err)
		}
		return nil, nil
	})

	// DelPic: Delete a picture
	vm.RegisterBuiltinFunction("DelPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DelPic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("DelPic requires picture ID argument")
		}
		picID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("DelPic picture ID must be integer")
			return nil, nil
		}
		if err := v.graphicsSystem.DelPic(int(picID)); err != nil {
			v.log.Error("DelPic failed", "picID", picID, "error", err)
		}
		v.log.Debug("DelPic called", "picID", picID)
		return nil, nil
	})

	// PicWidth: Get picture width
	vm.RegisterBuiltinFunction("PicWidth", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("PicWidth called but graphics system not initialized", "args", args)
			return 0, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("PicWidth requires picture ID argument")
			return 0, nil
		}
		picID, ok := toInt64(args[0])
		if !ok {
			return 0, fmt.Errorf("PicWidth picture ID must be integer")
		}
		width := v.graphicsSystem.PicWidth(int(picID))
		v.log.Debug("PicWidth called", "picID", picID, "width", width)
		return width, nil
	})

	// PicHeight: Get picture height
	vm.RegisterBuiltinFunction("PicHeight", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("PicHeight called but graphics system not initialized", "args", args)
			return 0, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("PicHeight requires picture ID argument")
			return 0, nil
		}
		picID, ok := toInt64(args[0])
		if !ok {
			return 0, fmt.Errorf("PicHeight picture ID must be integer")
		}
		height := v.graphicsSystem.PicHeight(int(picID))
		v.log.Debug("PicHeight called", "picID", picID, "height", height)
		return height, nil
	})

	// MoveSPic: Scale and transfer picture
	// MoveSPic(src_pic, src_x, src_y, src_w, src_h, dst_pic, dst_x, dst_y, dst_w, dst_h)
	vm.RegisterBuiltinFunction("MoveSPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MoveSPic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 10 {
			return nil, fmt.Errorf("MoveSPic requires 10 arguments")
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		srcW, _ := toInt64(args[3])
		srcH, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])
		dstW, _ := toInt64(args[8])
		dstH, _ := toInt64(args[9])

		if err := v.graphicsSystem.MoveSPic(int(srcID), int(srcX), int(srcY), int(srcW), int(srcH),
			int(dstID), int(dstX), int(dstY), int(dstW), int(dstH)); err != nil {
			v.log.Error("MoveSPic failed", "error", err)
		}
		v.log.Debug("MoveSPic called", "srcID", srcID, "dstID", dstID)
		return nil, nil
	})

	// TransPic: Transfer with transparency
	// TransPic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y, trans_color)
	vm.RegisterBuiltinFunction("TransPic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("TransPic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 9 {
			return nil, fmt.Errorf("TransPic requires 9 arguments")
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		width, _ := toInt64(args[3])
		height, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])
		transColor, _ := toInt64(args[8])

		if err := v.graphicsSystem.TransPic(int(srcID), int(srcX), int(srcY), int(width), int(height),
			int(dstID), int(dstX), int(dstY), int(transColor)); err != nil {
			v.log.Error("TransPic failed", "error", err)
		}
		v.log.Debug("TransPic called", "srcID", srcID, "dstID", dstID)
		return nil, nil
	})

	// ReversePic: Transfer with horizontal flip
	// ReversePic(src_pic, src_x, src_y, width, height, dst_pic, dst_x, dst_y)
	vm.RegisterBuiltinFunction("ReversePic", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("ReversePic called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 8 {
			return nil, fmt.Errorf("ReversePic requires 8 arguments")
		}

		srcID, _ := toInt64(args[0])
		srcX, _ := toInt64(args[1])
		srcY, _ := toInt64(args[2])
		width, _ := toInt64(args[3])
		height, _ := toInt64(args[4])
		dstID, _ := toInt64(args[5])
		dstX, _ := toInt64(args[6])
		dstY, _ := toInt64(args[7])

		if err := v.graphicsSystem.ReversePic(int(srcID), int(srcX), int(srcY), int(width), int(height),
			int(dstID), int(dstX), int(dstY)); err != nil {
			v.log.Error("ReversePic failed", "error", err)
		}
		v.log.Debug("ReversePic called", "srcID", srcID, "dstID", dstID)
		return nil, nil
	})

	// OpenWin: Open a window
	vm.RegisterBuiltinFunction("OpenWin", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("OpenWin called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("OpenWin requires at least picture ID argument")
		}
		picID, ok := toInt64(args[0])
		if !ok {
			return -1, fmt.Errorf("OpenWin picture ID must be integer")
		}
		// Pass remaining args as options (will be handled by GraphicsSystem)
		v.log.Debug("OpenWin args", "picID", picID, "opts", args[1:], "optsLen", len(args)-1)
		winID, err := v.graphicsSystem.OpenWin(int(picID), args[1:]...)
		if err != nil {
			return -1, fmt.Errorf("OpenWin failed")
		}
		v.log.Debug("OpenWin called", "picID", picID, "winID", winID)
		return winID, nil
	})

	// CloseWin: Close a window
	vm.RegisterBuiltinFunction("CloseWin", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CloseWin called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("CloseWin requires window ID argument")
		}
		winID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("CloseWin window ID must be integer")
			return nil, nil
		}
		if err := v.graphicsSystem.CloseWin(int(winID)); err != nil {
			v.log.Error("CloseWin failed", "winID", winID, "error", err)
		}
		v.log.Debug("CloseWin called", "winID", winID)
		return nil, nil
	})

	// MoveWin: Move a window
	vm.RegisterBuiltinFunction("MoveWin", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MoveWin called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("MoveWin requires at least window ID argument")
		}
		winID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("MoveWin window ID must be integer")
			return nil, nil
		}
		// Pass remaining args as options
		if err := v.graphicsSystem.MoveWin(int(winID), args[1:]...); err != nil {
			v.log.Error("MoveWin failed", "winID", winID, "error", err)
		}
		v.log.Debug("MoveWin called", "winID", winID)
		return nil, nil
	})

	// CloseWinAll: Close all windows
	vm.RegisterBuiltinFunction("CloseWinAll", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CloseWinAll called but graphics system not initialized")
			return nil, nil
		}
		v.graphicsSystem.CloseWinAll()
		v.log.Debug("CloseWinAll called")
		return nil, nil
	})

	// CapTitle: Set window caption
	// CapTitle(title) - set caption for ALL windows (受け入れ基準 3.1)
	// CapTitle(win_no, title) - set caption for specific window (受け入れ基準 3.3)
	// 存在しないウィンドウIDの場合はエラーを発生させない (受け入れ基準 3.4)
	// 空文字列をタイトルとして受け入れる (受け入れ基準 3.5)
	vm.RegisterBuiltinFunction("CapTitle", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("CapTitle called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("CapTitle requires at least 1 argument")
		}

		if len(args) == 1 {
			// CapTitle(title) - set caption for ALL windows (受け入れ基準 3.1)
			title, _ := args[0].(string)
			v.graphicsSystem.CapTitleAll(title)
			v.log.Debug("CapTitle called (all windows)", "title", title)
		} else {
			// CapTitle(win_no, title) - set caption for specific window (受け入れ基準 3.3)
			winID, _ := toInt64(args[0])
			title, _ := args[1].(string)
			// エラーを無視する (受け入れ基準 3.4: 存在しないウィンドウIDでもエラーを発生させない)
			_ = v.graphicsSystem.CapTitle(int(winID), title)
			v.log.Debug("CapTitle called", "winID", winID, "title", title)
		}
		return nil, nil
	})

	// GetPicNo: Get picture number associated with a window
	vm.RegisterBuiltinFunction("GetPicNo", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("GetPicNo called but graphics system not initialized", "args", args)
			return -1, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("GetPicNo requires window ID argument")
			return -1, nil
		}
		winID, ok := toInt64(args[0])
		if !ok {
			return -1, fmt.Errorf("GetPicNo window ID must be integer")
		}
		picID, err := v.graphicsSystem.GetPicNo(int(winID))
		if err != nil {
			return -1, fmt.Errorf("GetPicNo failed")
		}
		v.log.Debug("GetPicNo called", "winID", winID, "picID", picID)
		return picID, nil
	})

	// PutCast: Put a cast on a window
	// PutCast(win_no, pic_no, x, y, src_x, src_y, width, height) - 8 args
	// PutCast(pic_no, base_pic, x, y) - 4 args: simplified form (no transparency)
	// PutCast(pic_no, base_pic, x, y, transparentColor) - 5 args: with transparency
	// PutCast(pic_no, base_pic, x, y, transparentColor, ?, ?, ?, width, height, srcX, srcY) - 12 args: full
	vm.RegisterBuiltinFunction("PutCast", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("PutCast called but graphics system not initialized", "args", args)
			return -1, nil
		}

		// 12 args: PutCast(pic_no, base_pic, x, y, transparentColor, ?, ?, ?, width, height, srcX, srcY)
		if len(args) >= 12 {
			picID, _ := toInt64(args[0])
			basePic, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			transColorInt, _ := toInt64(args[4])
			// args[5-7] = unknown (ignored)
			width, _ := toInt64(args[8])
			height, _ := toInt64(args[9])
			srcX, _ := toInt64(args[10])
			srcY, _ := toInt64(args[11])

			// basePicはピクチャーIDなので、対応するウィンドウIDを逆引きする
			winID, err := v.graphicsSystem.GetWinByPicID(int(basePic))
			if err != nil {
				v.log.Warn("PutCast: no window found for base_pic", "basePic", basePic, "error", err)
				return -1, nil
			}

			// 透明色を設定
			transColor := graphics.ColorFromInt(int(transColorInt))
			castID, err := v.graphicsSystem.PutCastWithTransColor(winID, int(picID), int(x), int(y), int(srcX), int(srcY), int(width), int(height), transColor)
			if err != nil {
				v.log.Warn("PutCast failed", "error", err)
				return -1, nil
			}
			v.log.Debug("PutCast called (12 args)", "picID", picID, "basePic", basePic, "winID", winID, "x", x, "y", y, "transColor", transColorInt, "castID", castID)
			return castID, nil
		}

		// 5 args: PutCast(pic_no, base_pic, x, y, transparentColor)
		if len(args) == 5 {
			picID, _ := toInt64(args[0])
			basePic, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			transColorInt, _ := toInt64(args[4])
			w := v.graphicsSystem.PicWidth(int(picID))
			h := v.graphicsSystem.PicHeight(int(picID))

			// basePicはピクチャーIDなので、対応するウィンドウIDを逆引きする
			winID, err := v.graphicsSystem.GetWinByPicID(int(basePic))
			if err != nil {
				v.log.Warn("PutCast: no window found for base_pic", "basePic", basePic, "error", err)
				return -1, nil
			}

			// 透明色を設定
			transColor := graphics.ColorFromInt(int(transColorInt))
			castID, err := v.graphicsSystem.PutCastWithTransColor(winID, int(picID), int(x), int(y), 0, 0, w, h, transColor)
			if err != nil {
				v.log.Warn("PutCast failed", "error", err)
				return -1, nil
			}
			v.log.Debug("PutCast called (5 args)", "picID", picID, "basePic", basePic, "winID", winID, "transColor", transColorInt, "castID", castID)
			return castID, nil
		}

		// 4 args: PutCast(pic_no, base_pic, x, y) - no transparency
		// _old_implementation2の動作: キャストを作成し、かつbase_picにも画像を「焼き付ける」
		// これにより：
		// - キャストとして管理される（デバッグオーバーレイで表示される）
		// - base_picにも描画されるので、後続のMovePicで上書きできる
		// y_saruのシーン3では: PutCast(25,base_pic,0,0)で背景をキャストとして配置し、
		// その後MovePic(18,...)で爆発画像をbase_picに描画している。
		if len(args) == 4 {
			picID, _ := toInt64(args[0])
			basePic, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			w := v.graphicsSystem.PicWidth(int(picID))
			h := v.graphicsSystem.PicHeight(int(picID))

			// basePicはピクチャーIDなので、対応するウィンドウIDを逆引きする
			winID, err := v.graphicsSystem.GetWinByPicID(int(basePic))
			if err != nil {
				v.log.Warn("PutCast (4 args): no window found for base_pic", "basePic", basePic, "error", err)
				return -1, nil
			}

			// 1. キャストを作成（透明色なし）
			castID, err := v.graphicsSystem.PutCast(winID, int(picID), int(x), int(y), 0, 0, w, h)
			if err != nil {
				v.log.Warn("PutCast (4 args) failed", "error", err)
				return -1, nil
			}

			// 2. base_picにも画像を転送（焼き付け）
			// これにより後続のMovePicで描画された内容がキャストの下に隠れない
			err = v.graphicsSystem.MovePic(int(picID), 0, 0, w, h, int(basePic), int(x), int(y), 0)
			if err != nil {
				v.log.Warn("PutCast (4 args): MovePic failed", "error", err)
				// キャストは作成済みなのでエラーでも続行
			}

			v.log.Debug("PutCast called (4 args)", "picID", picID, "basePic", basePic, "winID", winID, "x", x, "y", y, "castID", castID)
			return castID, nil
		}

		// 8 args: PutCast(win_no, pic_no, x, y, src_x, src_y, width, height)
		if len(args) >= 8 {
			winID, _ := toInt64(args[0])
			picID, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			srcX, _ := toInt64(args[4])
			srcY, _ := toInt64(args[5])
			width, _ := toInt64(args[6])
			height, _ := toInt64(args[7])

			castID, err := v.graphicsSystem.PutCast(int(winID), int(picID), int(x), int(y), int(srcX), int(srcY), int(width), int(height))
			if err != nil {
				v.log.Warn("PutCast failed", "error", err)
				return -1, nil
			}
			v.log.Debug("PutCast called (8 args)", "winID", winID, "picID", picID, "castID", castID)
			return castID, nil
		}

		v.log.Warn("PutCast: invalid number of arguments", "count", len(args))
		return -1, nil
	})

	// MoveCast: Move a cast
	// MoveCast(cast_no, pic_no, x, y, ?, width, height, srcX, srcY) - 9 args (y_saru style)
	// MoveCast(cast_no, x, y) - 3 args: move position only
	// MoveCast(cast_no, x, y, src_x, src_y, width, height) - 7 args: move and change source
	// MoveCast(cast_no, pic_no, x, y) - 4 args: change picture and position
	vm.RegisterBuiltinFunction("MoveCast", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("MoveCast called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 3 {
			v.log.Warn("MoveCast requires at least 3 arguments", "count", len(args))
			return nil, nil
		}

		castID, _ := toInt64(args[0])

		// 9 args: MoveCast(cast_no, pic_no, x, y, ?, width, height, srcX, srcY)
		// This is the y_saru style
		if len(args) >= 9 {
			// args[1] = picID (ignored for now)
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])
			// args[4] = unknown/transparent color (ignored)
			width, _ := toInt64(args[5])
			height, _ := toInt64(args[6])
			srcX, _ := toInt64(args[7])
			srcY, _ := toInt64(args[8])

			v.log.Debug("MoveCast called (9 args)", "castID", castID, "x", x, "y", y, "width", width, "height", height, "srcX", srcX, "srcY", srcY)

			opts := []graphics.CastOption{
				graphics.WithCastPosition(int(x), int(y)),
				graphics.WithCastSource(int(srcX), int(srcY), int(width), int(height)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// 7 args: MoveCast(cast_no, x, y, src_x, src_y, width, height)
		if len(args) == 7 {
			x, _ := toInt64(args[1])
			y, _ := toInt64(args[2])
			srcX, _ := toInt64(args[3])
			srcY, _ := toInt64(args[4])
			width, _ := toInt64(args[5])
			height, _ := toInt64(args[6])

			opts := []graphics.CastOption{
				graphics.WithCastPosition(int(x), int(y)),
				graphics.WithCastSource(int(srcX), int(srcY), int(width), int(height)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// 4 args: MoveCast(cast_no, pic_no, x, y)
		if len(args) == 4 {
			picID, _ := toInt64(args[1])
			x, _ := toInt64(args[2])
			y, _ := toInt64(args[3])

			opts := []graphics.CastOption{
				graphics.WithCastPicID(int(picID)),
				graphics.WithCastPosition(int(x), int(y)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// 3 args: MoveCast(cast_no, x, y)
		if len(args) == 3 {
			x, _ := toInt64(args[1])
			y, _ := toInt64(args[2])

			opts := []graphics.CastOption{
				graphics.WithCastPosition(int(x), int(y)),
			}
			if err := v.graphicsSystem.MoveCastWithOptions(int(castID), opts...); err != nil {
				v.log.Warn("MoveCast failed", "castID", castID, "error", err)
			}
			return nil, nil
		}

		// Fallback: use old method
		if err := v.graphicsSystem.MoveCast(int(castID), args[1:]...); err != nil {
			v.log.Warn("MoveCast failed", "castID", castID, "error", err)
		}
		return nil, nil
	})

	// DelCast: Delete a cast
	vm.RegisterBuiltinFunction("DelCast", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DelCast called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("DelCast requires cast ID argument")
		}

		castID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("DelCast cast ID must be integer")
			return nil, nil
		}
		if err := v.graphicsSystem.DelCast(int(castID)); err != nil {
			v.log.Error("DelCast failed", "castID", castID, "error", err)
		}
		v.log.Debug("DelCast called", "castID", castID)
		return nil, nil
	})

	// TextWrite: Write text to a picture
	// TextWrite(text, pic_no, x, y)
	vm.RegisterBuiltinFunction("TextWrite", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("TextWrite called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 4 {
			return nil, fmt.Errorf("TextWrite requires 4 arguments (text, pic_no, x, y)")
			return nil, nil
		}

		text, ok := args[0].(string)
		if !ok {
			v.log.Error("TextWrite text must be string", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}
		picID, _ := toInt64(args[1])
		x, _ := toInt64(args[2])
		y, _ := toInt64(args[3])

		if err := v.graphicsSystem.TextWrite(int(picID), int(x), int(y), text); err != nil {
			v.log.Error("TextWrite failed", "error", err)
		}
		v.log.Debug("TextWrite called", "text", text, "picID", picID, "x", x, "y", y)
		return nil, nil
	})

	// SetFont: Set font for text rendering
	// SetFont(size, name, charset, italic, underline, strikeout, weight)
	vm.RegisterBuiltinFunction("SetFont", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetFont called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 2 {
			return nil, fmt.Errorf("SetFont requires at least 2 arguments (size, name)")
			return nil, nil
		}

		size, _ := toInt64(args[0])
		name, _ := args[1].(string)

		// Pass remaining args as options
		if err := v.graphicsSystem.SetFont(name, int(size), args[2:]...); err != nil {
			v.log.Error("SetFont failed", "error", err)
		}
		v.log.Debug("SetFont called", "size", size, "name", name)
		return nil, nil
	})

	// TextColor: Set text color
	// TextColor(r, g, b) or TextColor(color)
	vm.RegisterBuiltinFunction("TextColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("TextColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("TextColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetTextColor(colorInt); err != nil {
			v.log.Error("TextColor failed", "error", err)
		}
		v.log.Debug("TextColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// BgColor: Set background color for text
	// BgColor(r, g, b) or BgColor(color)
	vm.RegisterBuiltinFunction("BgColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("BgColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("BgColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetBgColor(colorInt); err != nil {
			v.log.Error("BgColor failed", "error", err)
		}
		v.log.Debug("BgColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// BackMode: Set background mode for text
	// BackMode(mode) - 0=transparent, 1=opaque
	vm.RegisterBuiltinFunction("BackMode", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("BackMode called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("BackMode requires 1 argument")
		}

		mode, _ := toInt64(args[0])
		if err := v.graphicsSystem.SetBackMode(int(mode)); err != nil {
			v.log.Error("BackMode failed", "error", err)
		}
		v.log.Debug("BackMode called", "mode", mode)
		return nil, nil
	})

	// WinInfo: Get virtual desktop information
	// WinInfo(0) - returns width
	// WinInfo(1) - returns height
	vm.RegisterBuiltinFunction("WinInfo", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("WinInfo called but graphics system not initialized", "args", args)
			// Return default values (skelton要件: 1024x768)
			if len(args) >= 1 {
				infoType, _ := toInt64(args[0])
				if infoType == 0 {
					return 1024, nil // default width
				}
				return 768, nil // default height
			}
			return 0, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("WinInfo requires 1 argument")
			return 0, nil
		}

		infoType, _ := toInt64(args[0])
		var result int
		if infoType == 0 {
			result = v.graphicsSystem.GetVirtualWidth()
		} else {
			result = v.graphicsSystem.GetVirtualHeight()
		}
		v.log.Debug("WinInfo called", "infoType", infoType, "result", result)
		return result, nil
	})

	// Debug: Set debug level (placeholder - does nothing for now)
	vm.RegisterBuiltinFunction("Debug", func(v *VM, args []any) (any, error) {
		if len(args) >= 1 {
			level, _ := toInt64(args[0])
			v.log.Debug("Debug called", "level", level)
		}
		return nil, nil
	})

	// DrawRect: Draw a rectangle
	// DrawRect(pic_no, x1, y1, x2, y2, fill_mode)
	// DrawRect(pic_no, x1, y1, x2, y2, r, g, b) - with color (fill mode 0)
	vm.RegisterBuiltinFunction("DrawRect", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DrawRect called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 6 {
			return nil, fmt.Errorf("DrawRect requires at least 6 arguments")
		}

		picID, _ := toInt64(args[0])
		x1, _ := toInt64(args[1])
		y1, _ := toInt64(args[2])
		x2, _ := toInt64(args[3])
		y2, _ := toInt64(args[4])
		fillMode, _ := toInt64(args[5])

		// If 8 arguments, treat as DrawRect with color (r, g, b)
		if len(args) >= 8 {
			r, _ := toInt64(args[5])
			g, _ := toInt64(args[6])
			b, _ := toInt64(args[7])
			colorInt := int(r)<<16 | int(g)<<8 | int(b)
			if err := v.graphicsSystem.SetPaintColor(colorInt); err != nil {
				v.log.Error("DrawRect SetPaintColor failed", "error", err)
			}
			fillMode = 0 // outline only when color is specified
		}

		if err := v.graphicsSystem.DrawRect(int(picID), int(x1), int(y1), int(x2), int(y2), int(fillMode)); err != nil {
			v.log.Error("DrawRect failed", "error", err)
		}
		v.log.Debug("DrawRect called", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2, "fillMode", fillMode)
		return nil, nil
	})

	// DrawLine: Draw a line
	// DrawLine(pic_no, x1, y1, x2, y2)
	vm.RegisterBuiltinFunction("DrawLine", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DrawLine called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 5 {
			return nil, fmt.Errorf("DrawLine requires 5 arguments")
		}

		picID, _ := toInt64(args[0])
		x1, _ := toInt64(args[1])
		y1, _ := toInt64(args[2])
		x2, _ := toInt64(args[3])
		y2, _ := toInt64(args[4])

		if err := v.graphicsSystem.DrawLine(int(picID), int(x1), int(y1), int(x2), int(y2)); err != nil {
			v.log.Error("DrawLine failed", "error", err)
		}
		v.log.Debug("DrawLine called", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2)
		return nil, nil
	})

	// FillRect: Fill a rectangle with color
	// FillRect(pic_no, x1, y1, x2, y2, color)
	vm.RegisterBuiltinFunction("FillRect", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("FillRect called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 6 {
			return nil, fmt.Errorf("FillRect requires 6 arguments")
		}

		picID, _ := toInt64(args[0])
		x1, _ := toInt64(args[1])
		y1, _ := toInt64(args[2])
		x2, _ := toInt64(args[3])
		y2, _ := toInt64(args[4])
		colorVal, _ := toInt64(args[5])

		if err := v.graphicsSystem.FillRect(int(picID), int(x1), int(y1), int(x2), int(y2), int(colorVal)); err != nil {
			v.log.Error("FillRect failed", "error", err)
		}
		v.log.Debug("FillRect called", "picID", picID, "x1", x1, "y1", y1, "x2", x2, "y2", y2)
		return nil, nil
	})

	// DrawCircle: Draw a circle
	// DrawCircle(pic_no, x, y, radius, fill_mode)
	vm.RegisterBuiltinFunction("DrawCircle", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("DrawCircle called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 5 {
			return nil, fmt.Errorf("DrawCircle requires 5 arguments")
		}

		picID, _ := toInt64(args[0])
		x, _ := toInt64(args[1])
		y, _ := toInt64(args[2])
		radius, _ := toInt64(args[3])
		fillMode, _ := toInt64(args[4])

		if err := v.graphicsSystem.DrawCircle(int(picID), int(x), int(y), int(radius), int(fillMode)); err != nil {
			v.log.Error("DrawCircle failed", "error", err)
		}
		v.log.Debug("DrawCircle called", "picID", picID, "x", x, "y", y, "radius", radius, "fillMode", fillMode)
		return nil, nil
	})

	// SetLineSize: Set line thickness
	vm.RegisterBuiltinFunction("SetLineSize", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetLineSize called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("SetLineSize requires 1 argument")
		}

		size, _ := toInt64(args[0])
		v.graphicsSystem.SetLineSize(int(size))
		v.log.Debug("SetLineSize called", "size", size)
		return nil, nil
	})

	// SetPaintColor: Set paint color
	// SetPaintColor(color) or SetPaintColor(r, g, b)
	vm.RegisterBuiltinFunction("SetPaintColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetPaintColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("SetPaintColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetPaintColor(colorInt); err != nil {
			v.log.Error("SetPaintColor failed", "error", err)
		}
		v.log.Debug("SetPaintColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// GetColor: Get pixel color at coordinates
	// GetColor(pic_no, x, y)
	vm.RegisterBuiltinFunction("GetColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("GetColor called but graphics system not initialized", "args", args)
			return 0, nil
		}
		if len(args) < 3 {
			return nil, fmt.Errorf("GetColor requires 3 arguments")
			return 0, nil
		}

		picID, _ := toInt64(args[0])
		x, _ := toInt64(args[1])
		y, _ := toInt64(args[2])

		colorVal, err := v.graphicsSystem.GetColor(int(picID), int(x), int(y))
		if err != nil {
			return 0, fmt.Errorf("GetColor failed")
		}
		v.log.Debug("GetColor called", "picID", picID, "x", x, "y", y, "color", fmt.Sprintf("0x%06X", colorVal))
		return colorVal, nil
	})

	// SetColor: Set color (alias for SetPaintColor)
	vm.RegisterBuiltinFunction("SetColor", func(v *VM, args []any) (any, error) {
		if v.graphicsSystem == nil {
			v.log.Debug("SetColor called but graphics system not initialized", "args", args)
			return nil, nil
		}
		if len(args) < 1 {
			return nil, fmt.Errorf("SetColor requires at least 1 argument")
		}

		var colorInt int
		if len(args) >= 3 {
			// RGB format
			r, _ := toInt64(args[0])
			g, _ := toInt64(args[1])
			b, _ := toInt64(args[2])
			colorInt = int(r)<<16 | int(g)<<8 | int(b)
		} else {
			// Single color value
			colorInt64, _ := toInt64(args[0])
			colorInt = int(colorInt64)
		}

		if err := v.graphicsSystem.SetPaintColor(colorInt); err != nil {
			v.log.Error("SetColor failed", "error", err)
		}
		v.log.Debug("SetColor called", "color", fmt.Sprintf("0x%06X", colorInt))
		return nil, nil
	})

	// StrPrint: Printf-style string formatting
	// Requirement 1.1: When StrPrint is called with format string and arguments, system returns formatted string.
	// Requirement 1.2: System supports %ld format specifier for decimal integers, converting to Go's %d.
	// Requirement 1.3: System supports %lx format specifier for hexadecimal, converting to Go's %x.
	// Requirement 1.4: System supports %s format specifier for strings.
	// Requirement 1.5: System supports width and padding specifiers like %03d.
	// Requirement 1.6: System converts escape sequences (\n, \t, \r) to actual control characters.
	// Requirement 1.7: When called with fewer arguments than format specifiers, system handles gracefully.
	// Requirement 1.8: When called with more arguments than format specifiers, system ignores extra arguments.
	vm.RegisterBuiltinFunction("StrPrint", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return "", nil
		}

		// Get format string
		format, ok := args[0].(string)
		if !ok {
			v.log.Error("StrPrint format must be string", "got", fmt.Sprintf("%T", args[0]))
			return "", nil
		}

		// Convert FILLY format specifiers to Go format specifiers
		// Use regex to handle width/padding specifiers like %03ld, %5lx
		// Pattern: %[flags][width][.precision]ld or %[flags][width][.precision]lx
		convertedFormat := format

		// Convert %ld variants (with optional flags, width, precision) to %d
		// Matches: %ld, %5ld, %05ld, %-5ld, %+5ld, etc.
		ldPattern := regexp.MustCompile(`%([+-]?\d*\.?\d*)ld`)
		convertedFormat = ldPattern.ReplaceAllString(convertedFormat, "%${1}d")

		// Convert %lx variants to %x
		lxPattern := regexp.MustCompile(`%([+-]?\d*\.?\d*)lx`)
		convertedFormat = lxPattern.ReplaceAllString(convertedFormat, "%${1}x")

		// Convert escape sequences to actual control characters
		convertedFormat = strings.ReplaceAll(convertedFormat, "\\n", "\n")
		convertedFormat = strings.ReplaceAll(convertedFormat, "\\t", "\t")
		convertedFormat = strings.ReplaceAll(convertedFormat, "\\r", "\r")

		// Prepare arguments for fmt.Sprintf
		formatArgs := make([]any, 0, len(args)-1)
		for i := 1; i < len(args); i++ {
			formatArgs = append(formatArgs, args[i])
		}

		// Use fmt.Sprintf to format the string
		// This handles both fewer and more arguments than format specifiers gracefully
		result := fmt.Sprintf(convertedFormat, formatArgs...)

		v.log.Debug("StrPrint called", "format", format, "result", result)
		return result, nil
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

	// Create and register the handler with the current scope
	// This allows the handler to access variables from the enclosing scope (like C blocks)
	parentScope := vm.GetCurrentScope()
	handler := NewEventHandler("", eventType, bodyOpcodes, vm, parentScope)
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
		// ログは削除（頻繁すぎるため）
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

// PushMouseEvent pushes a mouse event to the event queue.
// This implements the MouseEventPusher interface for the window package.
// 要件 14.6: マウスイベントをEbitengineから取得し、VMのイベントキューに追加する
// 要件 7.1: 左マウスボタンが押されたとき、LBDOWNイベントを生成する
// 要件 7.2: 右マウスボタンが押されたとき、RBDOWNイベントを生成する
// 要件 7.3: 右マウスボタンがダブルクリックされたとき、RBDBLCLKイベントを生成する
func (vm *VM) PushMouseEvent(eventType string, windowID, x, y int) {
	var evType EventType
	switch eventType {
	case "LBDOWN":
		evType = EventLBDOWN
	case "RBDOWN":
		evType = EventRBDOWN
	case "RBDBLCLK":
		evType = EventRBDBLCLK
	default:
		vm.log.Warn("Unknown mouse event type", "type", eventType)
		return
	}

	event := NewEventWithParams(evType, map[string]any{
		"MesP1": windowID, // ウィンドウID
		"MesP2": x,        // X座標
		"MesP3": y,        // Y座標
	})

	vm.eventQueue.Push(event)
	vm.log.Debug("Mouse event pushed", "type", eventType, "windowID", windowID, "x", x, "y", y)
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

// SetGraphicsSystem sets the graphics system.
// This allows external initialization of the graphics system to avoid import cycles.
//
// Parameters:
//   - graphicsSys: The graphics system implementing GraphicsSystemInterface
func (vm *VM) SetGraphicsSystem(graphicsSys GraphicsSystemInterface) {
	vm.graphicsSystem = graphicsSys
	vm.log.Info("Graphics system set")
}

// GetGraphicsSystem returns the graphics system.
// Returns nil if the graphics system has not been initialized.
func (vm *VM) GetGraphicsSystem() GraphicsSystemInterface {
	return vm.graphicsSystem
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
