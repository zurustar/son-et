package engine

import (
	"embed"
	"fmt"
	"image"
	"image/color"
	_ "image/png"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"gopkg.in/ini.v1"
)

// Debug level (env: DEBUG_LEVEL): 0=errors, 1=important ops (default), 2=all debug
var debugLevel = 1

func init() {
	if level := os.Getenv("DEBUG_LEVEL"); level != "" {
		if l, err := strconv.Atoi(level); err == nil {
			debugLevel = l
		}
	}
}

// Timing mode constants
const (
	Time     = 0 // TIME mode (frame-based timing)
	MidiTime = 1 // MIDI_TIME mode (MIDI-synchronized timing)
)

// Picture represents a loaded or created image
type Picture struct {
	ID         int
	Image      *ebiten.Image
	BackBuffer *ebiten.Image
	Width      int
	Height     int
}

// Window represents an open display window
type Window struct {
	ID         int
	Picture    int // Picture ID to display
	X, Y       int // Window position
	W, H       int // Window size
	SrcX, SrcY int // Picture Display Offset (X, Y)
	SrcW, SrcH int // Source size (unused in new logic?)
	Visible    bool
	Title      string      // Window title for decoration
	Color      color.Color // Background color
}

// Cast represents a sprite with color key transparency
type Cast struct {
	ID          int
	Picture     int         // Source picture ID
	DestPicture int         // Destination picture ID
	X, Y        int         // Position
	W, H        int         // Size
	SrcX, SrcY  int         // Added for sprite clipping
	Transparent color.Color // Transparent color (from top-left pixel)
	Visible     bool
}

// EngineState encapsulates all global state for the engine
// This allows for clean initialization, testing, and state isolation
type EngineState struct {
	// Synchronization
	renderMutex sync.Mutex // Protects Ebiten API calls and shared state

	// Dependencies (for testing and flexibility)
	assetLoader  AssetLoader
	imageDecoder ImageDecoder
	renderer     Renderer // Renderer for drawing (can be mocked for testing)

	// Global resources
	pictures      map[int]*Picture
	windows       map[int]*Window
	casts         map[int]*Cast
	castDrawOrder []int // Explicit draw order (creation order)

	// ID counters
	nextPicID   int // Start from 0 for FILLY compatibility
	nextWinID   int // Start from 0 for FILLY compatibility
	nextCastID  int
	windowOrder []int // Z-order for rendering

	// Window decoration state
	defaultWindowTitle string
	globalWindowTitle  string

	// File I/O state
	openFiles      map[int]*os.File // File handles for binary I/O
	nextFileHandle int              // Next file handle ID

	// Text rendering state
	currentFontSize  int
	currentFontName  string
	currentTextColor color.Color
	currentBgColor   color.Color
	currentBackMode  int // 0=transparent, 1=opaque
	currentFont      font.Face

	// Mock state
	MidiTime int // 1 = MIDI Sync Mode (0 = TIME Mode)
	MesP1    int
	MesP2    int
	MesP3    int
	MesP4    int

	// VM / Sequencer state
	mainSequencer *Sequencer
	vmLock        sync.Mutex
	tickCount     int64
	ticksPerStep  int
	tickLock      sync.Mutex
	midiSyncMode  bool
	targetTick    int64 // Atomic access
	GlobalPPQ     int

	// User functions
	userFuncs map[string]reflect.Value

	// Procedural execution state
	procMode      int // 0: TIME, 1: MIDI_TIME
	procStep      int // Default 6 ticks (100ms) for compat
	procWaitTicks int

	// Queued callback
	queuedCallback func()
}

// EngineStateOption is a function that configures an EngineState
type EngineStateOption func(*EngineState)

// WithAssetLoader sets a custom AssetLoader for the engine
func WithAssetLoader(loader AssetLoader) EngineStateOption {
	return func(e *EngineState) {
		e.assetLoader = loader
	}
}

// WithImageDecoder sets a custom ImageDecoder for the engine
func WithImageDecoder(decoder ImageDecoder) EngineStateOption {
	return func(e *EngineState) {
		e.imageDecoder = decoder
	}
}

// WithRenderer sets a custom Renderer for the engine
func WithRenderer(renderer Renderer) EngineStateOption {
	return func(e *EngineState) {
		e.renderer = renderer
	}
}

// NewEngineState creates a new EngineState with default values
// Optional dependencies can be injected via EngineStateOption functions
func NewEngineState(opts ...EngineStateOption) *EngineState {
	e := &EngineState{
		pictures:           make(map[int]*Picture),
		windows:            make(map[int]*Window),
		casts:              make(map[int]*Cast),
		castDrawOrder:      []int{},
		nextPicID:          0,
		nextWinID:          0,
		nextCastID:         1,
		windowOrder:        []int{},
		defaultWindowTitle: "FILLY Window",
		globalWindowTitle:  "FILLY Window",
		openFiles:          make(map[int]*os.File),
		nextFileHandle:     1,
		currentFontSize:    14,
		currentFontName:    "sans-serif",
		currentTextColor:   color.RGBA{0, 0, 0, 255},
		currentBgColor:     color.RGBA{255, 255, 255, 255},
		currentBackMode:    0,
		MidiTime:           1,
		MesP1:              0,
		MesP2:              0,
		MesP3:              0,
		MesP4:              0,
		ticksPerStep:       12,
		GlobalPPQ:          480,
		userFuncs:          make(map[string]reflect.Value),
		procMode:           0,
		procStep:           6,
		procWaitTicks:      0,
		// Default dependencies (can be overridden)
		assetLoader:  nil, // Will be set when Init is called
		imageDecoder: NewBMPImageDecoder(),
		renderer:     NewEbitenRenderer(), // Default to Ebitengine renderer
	}

	// Apply options
	for _, opt := range opts {
		opt(e)
	}

	return e
}

// Reset clears all state for test cleanup
func (e *EngineState) Reset() {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	// Clear resources
	e.pictures = make(map[int]*Picture)
	e.windows = make(map[int]*Window)
	e.casts = make(map[int]*Cast)
	e.castDrawOrder = []int{}
	e.windowOrder = []int{}

	// Reset ID counters
	e.nextPicID = 0
	e.nextWinID = 0
	e.nextCastID = 1

	// Reset text rendering state
	e.currentFontSize = 14
	e.currentFontName = "sans-serif"
	e.currentTextColor = color.RGBA{0, 0, 0, 255}
	e.currentBgColor = color.RGBA{255, 255, 255, 255}
	e.currentBackMode = 0
	e.currentFont = nil

	// Reset window decoration
	e.globalWindowTitle = e.defaultWindowTitle

	// Reset VM state
	e.vmLock.Lock()
	e.mainSequencer = nil
	e.tickCount = 0
	e.ticksPerStep = 12
	e.midiSyncMode = false
	atomic.StoreInt64(&e.targetTick, 0)
	e.vmLock.Unlock()

	// Reset user functions
	e.userFuncs = make(map[string]reflect.Value)

	// Reset procedural state
	e.procMode = 0
	e.procStep = 6
	e.procWaitTicks = 0
	e.queuedCallback = nil
}

var (
	// Global EngineState instance - used by all package-level functions
	globalEngine *EngineState

	// Legacy global variables - kept for backward compatibility
	// These are now aliases/wrappers to globalEngine
	renderMutex sync.Mutex // Protects Ebiten API calls and shared state (deprecated, use globalEngine.renderMutex)
	assets      embed.FS
	script      func() // The user script converted to Go function
	gameState   *Game

	// Global resources - deprecated, use globalEngine instead
	pictures      = make(map[int]*Picture)
	windows       = make(map[int]*Window)
	casts         = make(map[int]*Cast)
	castDrawOrder []int // Explicit draw order (creation order)

	// ID counters - deprecated, use globalEngine instead
	nextPicID   = 0 // Start from 0 for FILLY compatibility
	nextWinID   = 0 // Start from 0 for FILLY compatibility
	nextCastID  = 1
	windowOrder []int // Z-order for rendering

	// Window decoration state - deprecated, use globalEngine instead
	defaultWindowTitle = "FILLY Window"
	globalWindowTitle  = defaultWindowTitle // Initialize with default, allowing overwrite with empty string

	// Text rendering state - deprecated, use globalEngine instead
	currentFontSize  = 14
	currentFontName  = "sans-serif"
	currentTextColor = color.RGBA{0, 0, 0, 255}       // Black
	currentBgColor   = color.RGBA{255, 255, 255, 255} // White
	currentBackMode  = 0                              // 0=transparent, 1=opaque
	currentFont      font.Face                        // Current loaded font

	// Mock state - deprecated, use globalEngine instead
	// Note: MidiTime is now a constant (0=TIME, 1=MIDI_TIME)
	MesP1 = 0
	MesP2 = 0
	MesP3 = 0
	MesP4 = 0

	// Event handlers for mes() blocks
	midiEndHandler   func() // Handler for mes(MIDI_END)
	rbDownHandler    func() // Handler for mes(RBDOWN) - Right button down
	rbDblClkHandler  func() // Handler for mes(RBDBLCLK) - Right button double click
	midiEndTriggered bool   // Flag to track if MIDI_END event has been triggered

	// UI Layout Constants
	TitleBarHeight  = 24
	BorderThickness = 4
)

type Game struct {
	state      *EngineState
	renderer   Renderer
	tickCount  int
	frameImage *ebiten.Image
}

func (g *Game) Update() error {
	if g.tickCount == 0 {
		fmt.Println("Game.Update: First update call")
	}
	if g.tickCount < 5 {
		fmt.Printf("Game.Update: Update #%d\n", g.tickCount+1)
	}

	startTime := time.Now()
	// g.tickCount is frame count, distinct from global tickCount (VM ticks)
	g.tickCount++

	// Check for mouse events
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		// Right button is pressed
		if rbDownHandler != nil {
			TriggerRBDown()
		}
	}

	// Note: Double-click detection would require tracking click timing
	// For now, we'll implement a simple version that triggers on right-click
	// A full implementation would need to track the time between clicks

	if !midiSyncMode {
		// TIME MODE (Async / Frame-based)
		// Ignore MIDI-driven targetTick. Run exactly once per frame (60FPS).
		// ticksPerStep=6 -> 6 frames per step = 100ms.

		tickLock.Lock()
		tickCount++
		currentTick := int(tickCount)
		tickLock.Unlock()

		UpdateVM(currentTick)
	} else {
		// MIDI SYNC MODE
		currentTarget := atomic.LoadInt64(&targetTick)

		// Special case: Allow first command (pc==0) to execute even if targetTick is 0
		// This is necessary to execute PlayMIDI which starts the MIDI player
		vmLock.Lock()
		needsInitialExec := mainSequencer != nil && mainSequencer.pc == 0 && tickCount == 0
		vmLock.Unlock()

		if needsInitialExec && currentTarget == 0 {
			// Execute first command to bootstrap MIDI playback
			tickLock.Lock()
			tickCount = 1
			tickLock.Unlock()
			UpdateVM(1)
			return nil
		}

		// Limit catch-up to avoid spiral of death
		loops := 0

		tickLock.Lock()
		// We read global tickCount inside lock
		for tickCount < currentTarget && loops < 10 {
			tickCount++
			currentTick := int(tickCount)
			// Unlock during UpdateVM to avoid holding lock while VM runs (which might call Wait -> tickLock)
			tickLock.Unlock()

			UpdateVM(currentTick)

			tickLock.Lock() // Re-acquire for loop check
			loops++
		}
		tickLock.Unlock()
	}

	elapsed := time.Since(startTime)
	if elapsed > 5*time.Millisecond {
		// fmt.Printf("PERF: Update took %v\n", elapsed)
	}

	if g.tickCount <= 5 {
		fmt.Printf("Game.Update: Returning nil (tick #%d)\n", g.tickCount)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	if g.tickCount <= 5 {
		fmt.Printf("Game.Draw: Draw call #%d\n", g.tickCount)
	}

	// TEMPORARY: Debug nil checks
	if g == nil {
		fmt.Println("ERROR: Game is nil in Draw()")
		return
	}
	if g.state == nil {
		fmt.Println("ERROR: Game.state is nil in Draw()")
		return
	}
	if g.renderer == nil {
		fmt.Println("ERROR: Game.renderer is nil in Draw()")
		return
	}
	if screen == nil {
		fmt.Println("ERROR: screen is nil in Draw()")
		return
	}

	// TEMPORARY: Force clear screen to test if Draw() is being called
	screen.Fill(color.RGBA{0, 128, 255, 255}) // Blue background

	// Use the renderer to draw the frame
	// The renderer will handle locking and reading from state
	g.renderer.RenderFrame(screen, g.state)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int) {
	fmt.Printf("Layout called: outside=(%d,%d) -> returning (1280,720)\n", outsideWidth, outsideHeight)
	return 1280, 720
}

func Init(fs embed.FS, scriptFunc func()) {
	assets = fs
	script = scriptFunc

	// Create the global EngineState instance
	globalEngine = NewEngineState(
		WithAssetLoader(NewEmbedFSAssetLoader(fs)),
	)

	// Create the game state using the global engine
	gameState = &Game{
		state:      globalEngine,
		renderer:   NewEbitenRenderer(),
		tickCount:  0,
		frameImage: ebiten.NewImage(1280, 720),
	}

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("son-et")
	fmt.Println("Init: Window configured (1280x720)")

	// Try to load default font
	SetFont(14, "MS UI Gothic", 0)

	// Initialize Audio
	fmt.Println("=== AUDIO INITIALIZATION START ===")
	InitializeAudio()

	// SoundFont loading logic
	sfPath := "default.sf2"

	// 1. Check command line args
	for i, arg := range os.Args {
		if (arg == "-sf" || arg == "-soundfont") && i+1 < len(os.Args) {
			sfPath = os.Args[i+1]
			break
		}
	}

	// 2. If default not found and not specified, try to find a single .sf2 in CWD
	if sfPath == "default.sf2" {
		if _, err := os.Stat(sfPath); os.IsNotExist(err) {
			// Check for other .sf2 files
			matches, _ := filepath.Glob("*.sf2")
			if len(matches) == 0 {
				matches, _ = filepath.Glob("*.SF2") // Try uppercase
			}

			if len(matches) > 0 {
				sfPath = matches[0]
				if len(matches) > 1 {
					fmt.Printf("Info: Multiple SoundFonts found, using %s\n", sfPath)
				} else {
					fmt.Printf("Info: Auto-detected SoundFont: %s\n", sfPath)
				}
			}
		}
	}

	// Try to load SoundFont
	if err := LoadSoundFont(sfPath); err != nil {
		fmt.Printf("Warning: Could not load SoundFont (%s): %v\n", sfPath, err)
		fmt.Println("Please provide a .sf2 file via -sf <path> or place a .sf2 file in the directory.")
	} else {
		fmt.Printf("Success: SoundFont loaded from %s\n", sfPath)
	}
	fmt.Println("=== AUDIO INITIALIZATION END ===")
}

// InitEngineState initializes a new EngineState with the provided embed.FS
// This is useful for testing and when you want to use EngineState directly
func InitEngineState(fs embed.FS, opts ...EngineStateOption) *EngineState {
	// Create asset loader from embed.FS
	assetLoader := NewEmbedFSAssetLoader(fs)

	// Prepend the asset loader option
	allOpts := append([]EngineStateOption{WithAssetLoader(assetLoader)}, opts...)

	return NewEngineState(allOpts...)
}

// InitDirect initializes the engine for direct mode execution
// This is used by the CLI when executing TFY projects from a directory
func InitDirect(assetLoader AssetLoader, imageDecoder ImageDecoder, scriptFunc func()) {
	script = scriptFunc

	// Create the global EngineState instance with custom loaders
	globalEngine = NewEngineState(
		WithAssetLoader(assetLoader),
		WithImageDecoder(imageDecoder),
	)

	// Create the game state using the global engine
	gameState = &Game{
		state:      globalEngine,
		renderer:   NewEbitenRenderer(),
		tickCount:  0,
		frameImage: ebiten.NewImage(1280, 720),
	}

	ebiten.SetWindowSize(1280, 720)
	ebiten.SetWindowTitle("son-et")
	fmt.Println("InitDirect: Window configured (1280x720)")

	// Try to load default font
	SetFont(14, "MS UI Gothic", 0)

	// Initialize Audio
	fmt.Println("=== AUDIO INITIALIZATION START ===")
	// TEMPORARY: Disable audio for testing
	// InitializeAudio()
	fmt.Println("=== AUDIO INITIALIZATION SKIPPED (TESTING) ===")

	// SoundFont loading logic
	sfPath := "default.sf2"

	// 1. Check command line args
	for i, arg := range os.Args {
		if (arg == "-sf" || arg == "-soundfont") && i+1 < len(os.Args) {
			sfPath = os.Args[i+1]
			break
		}
	}

	// 2. If default not found and not specified, try to find a single .sf2 in CWD
	if sfPath == "default.sf2" {
		if _, err := os.Stat(sfPath); os.IsNotExist(err) {
			// Check for other .sf2 files
			matches, _ := filepath.Glob("*.sf2")
			if len(matches) == 0 {
				matches, _ = filepath.Glob("*.SF2") // Try uppercase
			}

			if len(matches) > 0 {
				sfPath = matches[0]
				if len(matches) > 1 {
					fmt.Printf("Info: Multiple SoundFonts found, using %s\n", sfPath)
				} else {
					fmt.Printf("Info: Auto-detected SoundFont: %s\n", sfPath)
				}
			}
		}
	}

	// Try to load SoundFont
	// TEMPORARY: Skip SoundFont loading for testing
	/*
		if err := LoadSoundFont(sfPath); err != nil {
			fmt.Printf("Warning: Could not load SoundFont (%s): %v\n", sfPath, err)
			fmt.Println("Please provide a .sf2 file via -sf <path> or place a .sf2 file in the directory.")
		} else {
			fmt.Printf("Success: SoundFont loaded from %s\n", sfPath)
		}
	*/
	fmt.Println("=== SOUNDFONT LOADING SKIPPED (TESTING) ===")
	fmt.Println("=== AUDIO INITIALIZATION END ===")
}

func Close() {
	fmt.Println("Engine Close")
}

func Run() {
	fmt.Printf("Run: About to call ebiten.RunGame with gameState=%v\n", gameState)
	if gameState == nil {
		fmt.Println("Run: ERROR - gameState is nil!")
		return
	}

	// Verify gameState fields
	fmt.Printf("Run: gameState.state=%v, gameState.renderer=%v\n", gameState.state, gameState.renderer)

	// Start script execution in background after a short delay
	// This allows the game loop to start first
	fmt.Println("Run: Starting goroutine for script execution")
	go func() {
		time.Sleep(100 * time.Millisecond)
		fmt.Println("Run: Executing script function")
		if script != nil {
			script()
		} else {
			fmt.Println("Run: Warning - script is nil")
		}
		fmt.Println("Run: Script function completed")
	}()

	fmt.Println("Run: Calling ebiten.RunGame...")
	fmt.Println("Run: This should block until window closes")
	if err := ebiten.RunGame(gameState); err != nil {
		fmt.Printf("Run: ebiten.RunGame returned error: %v\n", err)
		log.Fatal(err)
	}
	fmt.Println("Run: ebiten.RunGame returned normally")
}

// --- API Stubs ---

func CapTitle(args ...any) {
	fmt.Printf("CapTitle called with %d args: %v\n", len(args), args)
	for i, a := range args {
		fmt.Printf("  Arg[%d]: type=%T value=%v\n", i, a, a)
	}

	var title string
	if len(args) == 1 {
		// CapTitle(title) - Set global title for new windows
		if t, ok := args[0].(string); ok {
			if globalEngine != nil {
				globalEngine.globalWindowTitle = t
			}
			globalWindowTitle = t
			title = t
		}
	} else if len(args) >= 2 {
		// CapTitle(winID, title) - Set specific window title
		if t, ok := args[1].(string); ok {
			title = t
		}
		if id, ok := args[0].(int); ok {
			fmt.Printf("  Updating Window %d title to '%s'\n", id, title)
			if globalEngine != nil {
				if win, ok := globalEngine.windows[id]; ok {
					win.Title = title
				} else {
					fmt.Printf("  Window %d not found!\n", id)
				}
			} else if win, ok := windows[id]; ok {
				win.Title = title
			} else {
				fmt.Printf("  Window %d not found!\n", id)
			}
		}
	}

	fmt.Printf("CapTitle: Setting global/window title to: %s\n", title)
	// Also set main window title for feedback
	if title != "" {
		ebiten.SetWindowTitle(title)
	}
}

func WinInfo(mode int) int {
	// 0: Width, 1: Height
	if mode == 0 {
		return 1280
	}
	if mode == 1 {
		return 720
	}
	return 0
}

// Debug sets the debug level for logging
// level: 0=errors only, 1=important ops (default), 2=all debug
func Debug(level int) {
	debugLevel = level
	if debugLevel >= 1 {
		fmt.Printf("Debug: Set debug level to %d\n", level)
	}
}

// ExitTitle exits the program gracefully
// This is called when a FILLY script wants to terminate
func ExitTitle() {
	if debugLevel >= 1 {
		fmt.Println("ExitTitle: Terminating program")
	}
	os.Exit(0)
}

// RegisterMidiEndHandler registers a handler for MIDI_END event
func RegisterMidiEndHandler(handler func()) {
	midiEndHandler = handler
	midiEndTriggered = false
	if debugLevel >= 1 {
		fmt.Println("RegisterMidiEndHandler: Handler registered")
	}
}

// RegisterRBDownHandler registers a handler for RBDOWN (right button down) event
func RegisterRBDownHandler(handler func()) {
	rbDownHandler = handler
	if debugLevel >= 1 {
		fmt.Println("RegisterRBDownHandler: Handler registered")
	}
}

// RegisterRBDblClkHandler registers a handler for RBDBLCLK (right button double click) event
func RegisterRBDblClkHandler(handler func()) {
	rbDblClkHandler = handler
	if debugLevel >= 1 {
		fmt.Println("RegisterRBDblClkHandler: Handler registered")
	}
}

// TriggerMidiEnd triggers the MIDI_END event handler
// This should be called when MIDI playback finishes
func TriggerMidiEnd() {
	if midiEndHandler != nil && !midiEndTriggered {
		midiEndTriggered = true
		if debugLevel >= 1 {
			fmt.Println("TriggerMidiEnd: Triggering MIDI_END handler")
		}
		go midiEndHandler()
	}
}

// TriggerRBDown triggers the RBDOWN event handler
func TriggerRBDown() {
	if rbDownHandler != nil {
		if debugLevel >= 2 {
			fmt.Println("TriggerRBDown: Triggering RBDOWN handler")
		}
		go rbDownHandler()
	}
}

// TriggerRBDblClk triggers the RBDBLCLK event handler
func TriggerRBDblClk() {
	if rbDblClkHandler != nil {
		if debugLevel >= 2 {
			fmt.Println("TriggerRBDblClk: Triggering RBDBLCLK handler")
		}
		go rbDblClkHandler()
	}
}

// LoadPic loads a picture from embedded assets (EngineState method)
func (e *EngineState) LoadPic(path string) int {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()
	fmt.Printf("LoadPic: %s\n", path)

	// Check if assetLoader is available
	if e.assetLoader == nil {
		fmt.Printf("  Error: AssetLoader not initialized\n")
		return -1
	}

	// Try to read the file from embedded assets (case-insensitive)
	// Windows 3.1 compatibility: file names are case-insensitive
	entries, err := e.assetLoader.ReadDir(".")
	if err != nil {
		fmt.Printf("  Error reading assets: %v\n", err)
		return 0 // Return 0 for error
	}

	// Find matching file (case-insensitive)
	pathLower := strings.ToLower(path)
	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == pathLower {
			data, err := e.assetLoader.ReadFile(entry.Name())
			if err != nil {
				fmt.Printf("  Error reading %s: %v\n", entry.Name(), err)
				return -1
			}

			// Decode BMP using injected decoder
			img, err := e.imageDecoder.Decode(data)
			if err != nil {
				fmt.Printf("  Error decoding BMP %s: %v\n", entry.Name(), err)
				return -1
			}

			// Convert to Ebiten image
			ebitenImg := ebiten.NewImageFromImage(img)
			bounds := img.Bounds()
			width := bounds.Dx()
			height := bounds.Dy()

			// Store picture
			picID := e.nextPicID
			e.nextPicID++
			e.pictures[picID] = &Picture{
				ID:     picID,
				Image:  ebitenImg,
				Width:  width,
				Height: height,
			}

			fmt.Printf("  Loaded and decoded %s (%dx%d, ID=%d)\n", entry.Name(), width, height, picID)
			return picID
		}
	}

	fmt.Printf("  File not found in assets\n")
	return -1 // Return -1 for not found
}

// LoadPic is a backward-compatible wrapper for the global state
func LoadPic(path string) int {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return -1
	}
	return globalEngine.LoadPic(path)
}

// convertTransparentColor creates a new image with the specified color converted to transparent (alpha=0)
// convertTransparentColor creates a new image with the specified color converted to transparent (alpha=0)
// This leverages Ebitengine's native alpha channel support for efficient rendering
func convertTransparentColor(src *ebiten.Image, transparentColor color.Color) (result *ebiten.Image) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create new image with transparency
	result = ebiten.NewImage(width, height)

	// Get 8-bit RGB values of transparent color for comparison
	// Convert from 16-bit RGBA() values to 8-bit
	tr16, tg16, tb16, _ := transparentColor.RGBA()
	tr := uint8(tr16 >> 8)
	tg := uint8(tg16 >> 8)
	tb := uint8(tb16 >> 8)

	fmt.Printf("convertTransparentColor: Target color RGB=(%d,%d,%d)\n", tr, tg, tb)

	// Process pixels: convert matching color to transparent, copy others
	// Note: This may fail in test environments before game loop starts
	// In that case, we'll just return the original image
	defer func() {
		if r := recover(); r != nil {
			// If we can't read pixels (e.g., in tests), just return original
			fmt.Printf("convertTransparentColor: Panic during pixel processing, returning original image\n")
			result = src
		}
	}()

	transparentCount := 0
	opaqueCount := 0

	// Sample first few pixels for debugging
	sampleCount := 0

	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			c := src.At(px+bounds.Min.X, py+bounds.Min.Y)
			r16, g16, b16, a16 := c.RGBA()

			// Convert to 8-bit for comparison
			r := uint8(r16 >> 8)
			g := uint8(g16 >> 8)
			b := uint8(b16 >> 8)
			a := uint8(a16 >> 8)

			// Debug: Print first few pixels
			if sampleCount < 5 {
				fmt.Printf("  Pixel(%d,%d): RGB=(%d,%d,%d) A=%d\n", px, py, r, g, b, a)
				sampleCount++
			}

			// If pixel matches transparent color, skip it (leave as transparent)
			if r == tr && g == tg && b == tb {
				transparentCount++
				continue
			}

			// Copy non-transparent pixel
			result.Set(px, py, color.RGBA{
				R: r,
				G: g,
				B: b,
				A: a,
			})
			opaqueCount++
		}
	}

	fmt.Printf("  Converted: %d transparent, %d opaque pixels\n", transparentCount, opaqueCount)

	return result
}

// CreatePic creates a new picture (EngineState method)
func (e *EngineState) CreatePic(args ...any) int {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()
	fmt.Printf("CreatePic: args=%v\n", args)

	var width, height int
	var sourcePic *Picture

	// Parse arguments
	// Common patterns:
	// CreatePic(sourcePicID, width, height) - create from source
	// CreatePic(sourcePicID) - copy source
	// CreatePic(width, height) - create blank

	if len(args) >= 3 {
		// CreatePic(sourcePicID, width, height)
		if srcID, ok := args[0].(int); ok {
			sourcePic = e.pictures[srcID]
		}
		if w, ok := args[1].(int); ok {
			width = w
		}
		if h, ok := args[2].(int); ok {
			height = h
		}
	} else if len(args) == 2 {
		// CreatePic(width, height)
		if w, ok := args[0].(int); ok {
			width = w
		}
		if h, ok := args[1].(int); ok {
			height = h
		}
	} else if len(args) == 1 {
		// CreatePic(sourcePicID) - copy
		if srcID, ok := args[0].(int); ok {
			if srcID < 0 {
				return -1
			}
			sourcePic = e.pictures[srcID]
			if sourcePic != nil {
				width = sourcePic.Width
				height = sourcePic.Height
			}
		}
	}

	// Default size if not specified
	if width == 0 {
		width = 640
	}
	if height == 0 {
		height = 480
	}

	// Create new image (always empty)
	newImg := ebiten.NewImage(width, height)

	// Fill with white background (not transparent)
	// This prevents lower windows from showing through
	newImg.Fill(color.White)

	// NOTE: We do NOT copy the source image content here
	// The source is only used to determine the size
	// Content should be added via PutCast, MovePic, etc.

	// Store picture
	picID := e.nextPicID
	e.nextPicID++
	e.pictures[picID] = &Picture{
		ID:     picID,
		Image:  newImg,
		Width:  width,
		Height: height,
	}

	fmt.Printf("  Created picture ID=%d (%dx%d)\n", picID, width, height)
	return picID
}

// CreatePic is a backward-compatible wrapper for the global state
func CreatePic(args ...any) int {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return -1
	}
	return globalEngine.CreatePic(args...)
}

// OpenWin creates a new window (EngineState method)
func (e *EngineState) OpenWin(pic, x, y, w, h, picX, picY, col int) int {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	fmt.Printf("OpenWin: pic=%d at (%d,%d) size %dx%d\n", pic, x, y, w, h)

	// Use picture dimensions if not specified
	if w == 0 || h == 0 {
		if picture, ok := e.pictures[pic]; ok {
			if w == 0 {
				w = picture.Width
			}
			if h == 0 {
				h = picture.Height
			}
			fmt.Printf("OpenWin: Using picture size: %dx%d\n", w, h)
		} else {
			// Fallback to default if picture not found
			if w == 0 {
				w = 640
			}
			if h == 0 {
				h = 480
			}
			fmt.Printf("OpenWin: Picture %d not found, using default size: %dx%d\n", pic, w, h)
		}
	}

	// Debug logging when DEBUG_LEVEL >= 2
	if debugLevel >= 2 {
		fmt.Printf("OpenWin: pic=%d window at %d,%d size %dx%d source from %d,%d color=%X\n", pic, x, y, w, h, picX, picY, col)
	}

	winID := e.nextWinID
	e.nextWinID++

	// Determine title
	title := e.globalWindowTitle

	// Parse color (int 0xRRGGBB)
	r := uint8((col >> 16) & 0xFF)
	g := uint8((col >> 8) & 0xFF)
	b := uint8(col & 0xFF)
	bgColor := color.RGBA{r, g, b, 255}

	e.windows[winID] = &Window{
		ID:      winID,
		Picture: pic,
		X:       x + BorderThickness,                  // Adjust for border
		Y:       y + TitleBarHeight + BorderThickness, // Adjust for title bar and border
		W:       w,                                    // Content Width (as requested)
		H:       h,                                    // Content Height (as requested)
		SrcX:    -picX,                                // Picture Offset X (Inverted for legacy compatibility)
		SrcY:    -picY,                                // Picture Offset Y (Inverted for legacy compatibility)
		SrcW:    w,                                    // Source width (usually same as window)
		SrcH:    h,                                    // Source height (usually same as window)
		Visible: true,
		Title:   title,
		Color:   bgColor,
	}

	// Add to render order
	e.windowOrder = append(e.windowOrder, winID)

	// DEBUG: Confirm window was added
	fmt.Printf("  Window %d added. windowOrder len=%d, windows map len=%d\n",
		winID, len(e.windowOrder), len(e.windows))

	return winID
}

// OpenWin is a backward-compatible wrapper for the global state
func OpenWin(pic, x, y, w, h, picX, picY, col int) int {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return -1
	}
	return globalEngine.OpenWin(pic, x, y, w, h, picX, picY, col)
}

// CloseWin closes a window (EngineState method)
func (e *EngineState) CloseWin(winID int) {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()
	fmt.Printf("CloseWin: %d (before: %d windows)\n", winID, len(e.windows))

	// Remove from windows map
	delete(e.windows, winID)

	// Remove from render order
	for i, id := range e.windowOrder {
		if id == winID {
			e.windowOrder = append(e.windowOrder[:i], e.windowOrder[i+1:]...)
			break
		}
	}

	fmt.Printf("  After close: %d windows remaining\n", len(e.windows))
}

// CloseWin is a backward-compatible wrapper for the global state
func CloseWin(winID int) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.CloseWin(winID)
}

// CloseWinAll closes all open windows (EngineState method)
func (e *EngineState) CloseWinAll() {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	count := len(e.windows)
	if debugLevel >= 1 {
		fmt.Printf("CloseWinAll: Closing %d windows\n", count)
	}

	// Clear all windows
	e.windows = make(map[int]*Window)
	e.windowOrder = []int{}

	if debugLevel >= 1 {
		fmt.Printf("  All windows closed\n")
	}
}

// CloseWinAll is a backward-compatible wrapper for the global state
func CloseWinAll() {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.CloseWinAll()
}

// MoveWin updates window properties (EngineState method)
func (e *EngineState) MoveWin(winID, pic, x, y, w, h, picX, picY int) {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()
	fmt.Printf("MoveWin: %d\n", winID)

	win, ok := e.windows[winID]
	if !ok {
		return
	}

	// Update window properties
	win.Picture = pic
	win.X = x + BorderThickness
	win.Y = y + TitleBarHeight + BorderThickness
	win.W = w
	win.H = h
	win.SrcX = -picX // Picture Offset X (Inverted for legacy compatibility)
	win.SrcY = -picY // Picture Offset Y (Inverted for legacy compatibility)
}

// MoveWin is a backward-compatible wrapper for the global state
func MoveWin(winID, pic, x, y, w, h, picX, picY int) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.MoveWin(winID, pic, x, y, w, h, picX, picY)
}

// DelPic deletes a picture (EngineState method)
func (e *EngineState) DelPic(picID int) {
	fmt.Printf("DelPic: %d\n", picID)
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()
	delete(e.pictures, picID)
}

// DelPic is a backward-compatible wrapper for the global state
func DelPic(picID int) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.DelPic(picID)
}

func SetFont(size int, name string, weight int, args ...any) {
	// Legacy support: TFY scripts like ROBOT.TFY pass SetFont(640, "Name", 128).
	// If size is unreasonably large (> 200) and weight is a reasonable font size, swap them.
	realSize := size
	if size > 200 && weight > 0 && weight < 200 {
		fmt.Printf("SetFont: Legacy mode detected. Using arg3 (%d) as size instead of arg1 (%d)\n", weight, size)
		realSize = weight
	}

	fmt.Printf("SetFont: %s %d\n", name, realSize)

	if globalEngine != nil {
		globalEngine.currentFontSize = realSize
		globalEngine.currentFontName = name
	}

	// Also update legacy globals for backward compatibility
	currentFontSize = realSize
	currentFontName = name

	// Try to load system font
	// Common Japanese font paths on macOS
	fontPaths := []string{
		"/System/Library/Fonts/ヒラギノ明朝 ProN.ttc",
		"/System/Library/Fonts/ヒラギノ角ゴシック W3.ttc",
		"/System/Library/Fonts/ヒラギノ角ゴシック W4.ttc",
		"/Library/Fonts/Arial Unicode.ttf",
		"/System/Library/Fonts/Supplemental/Arial Unicode.ttf",
	}

	// Try each font path
	for _, fontPath := range fontPaths {
		if _, err := os.Stat(fontPath); err == nil {
			if face := loadFont(fontPath, float64(size)); face != nil {
				if globalEngine != nil {
					globalEngine.currentFont = face
				}
				currentFont = face
				fmt.Printf("  Loaded font: %s\n", fontPath)
				return
			}
		}
	}

	fmt.Println("  Warning: Could not load system font, text rendering will be limited")
}

// loadFont loads a TrueType font from file
func loadFont(path string, size float64) font.Face {
	fontData, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Printf("    Failed to read font file: %v\n", err)
		return nil
	}

	// Try to parse as a single font first
	tt, err := opentype.Parse(fontData)
	if err != nil {
		// If that fails, try as a font collection (.ttc)
		collection, err := opentype.ParseCollection(fontData)
		if err != nil {
			fmt.Printf("    Failed to parse font: %v\n", err)
			return nil
		}
		// Use the first font in the collection
		if collection.NumFonts() > 0 {
			tt, err = collection.Font(0)
			if err != nil {
				fmt.Printf("    Failed to get font from collection: %v\n", err)
				return nil
			}
		} else {
			fmt.Printf("    Font collection is empty\n")
			return nil
		}
	}

	face, err := opentype.NewFace(tt, &opentype.FaceOptions{
		Size:    size,
		DPI:     72,
		Hinting: font.HintingFull,
	})
	if err != nil {
		fmt.Printf("    Failed to create font face: %v\n", err)
		return nil
	}

	return face
}

func TextColor(r, g, b int) {
	fmt.Printf("TextColor: %d, %d, %d\n", r, g, b)
	col := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	if globalEngine != nil {
		globalEngine.currentTextColor = col
	}
	currentTextColor = col
}

func BgColor(r, g, b int) {
	fmt.Printf("BgColor: %d, %d, %d\n", r, g, b)
	col := color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	if globalEngine != nil {
		globalEngine.currentBgColor = col
	}
	currentBgColor = col
}

func BackMode(mode int) {
	fmt.Printf("BackMode: %d\n", mode)
	if globalEngine != nil {
		globalEngine.currentBackMode = mode
	}
	currentBackMode = mode
}

// TextWrite writes text to a picture (EngineState method)
func (e *EngineState) TextWrite(textStr string, pic, x, y int) {
	fmt.Printf("TextWrite: %q at %d,%d (pic=%d)\n", textStr, x, y, pic)

	// Get destination picture directly
	if pic < 0 {
		return
	}

	// PROTECT: RenderMutex
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	destPic, ok := e.pictures[pic]
	if !ok {
		fmt.Printf("  Picture ID=%d not found\n", pic)
		return
	}

	// If we have a loaded font, use it
	if e.currentFont != nil {
		// text.Draw uses Y as the baseline, so text is drawn ABOVE the Y coordinate
		// Add font size to ensure the entire text is within positive coordinates
		adjustedY := y + e.currentFontSize

		// Clear the text area first to prevent alpha blending artifacts
		// This is needed because text.Draw uses alpha blending, so drawing text
		// multiple times on the same area causes the old text to show through
		textWidth := len(textStr) * e.currentFontSize // Rough estimate
		textHeight := e.currentFontSize + 4           // Add some padding

		// Use opaque white background to completely erase previous text
		// (transparent background doesn't work well with antialiased text)
		clearColor := color.RGBA{255, 255, 255, 255} // Opaque white
		for py := 0; py < textHeight; py++ {
			for px := 0; px < textWidth; px++ {
				if x+px < destPic.Width && y+py < destPic.Height {
					destPic.Image.Set(x+px, y+py, clearColor)
				}
			}
		}

		text.Draw(destPic.Image, textStr, e.currentFont, x, adjustedY, e.currentTextColor)
		fmt.Printf("  Text drawn successfully to picture ID=%d (adjusted Y from %d to %d)\n", pic, y, adjustedY)
	} else {
		// Fallback: draw a colored rectangle as placeholder
		if e.currentBackMode == 1 {
			textWidth := len(textStr) * e.currentFontSize / 2
			textHeight := e.currentFontSize
			for py := 0; py < textHeight; py++ {
				for px := 0; px < textWidth; px++ {
					if x+px < destPic.Width && y+py < destPic.Height {
						destPic.Image.Set(x+px, y+py, e.currentTextColor)
					}
				}
			}
		}
		fmt.Printf("  Text placeholder drawn to picture ID=%d (no font loaded)\n", pic)
	}
}

// TextWrite is a backward-compatible wrapper for the global state
func TextWrite(textStr string, pic, x, y int) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.TextWrite(textStr, pic, x, y)
}

// MovePic copies pixels from source to destination (EngineState method)
func (e *EngineState) MovePic(pic, x, y, w, h, dest, dx, dy int, args ...any) {
	// Reduced logging
	// fmt.Printf("MovePic: %d -> %d at %d,%d\n", pic, dest, dx, dy)

	// Thread Safety: Lock to prevent race with Game.Draw
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	// Get source picture
	srcPic, ok := e.pictures[pic]
	if !ok {
		return
	}

	// Get destination - check windows first, then pictures
	var destImg *ebiten.Image
	var destPicID int

	if win, ok := e.windows[dest]; ok && win.Visible {
		if destPic, ok := e.pictures[win.Picture]; ok {
			destImg = destPic.Image
			destPicID = win.Picture
		}
	} else if destPic, ok := e.pictures[dest]; ok {
		destImg = destPic.Image
		destPicID = dest
	}

	if destImg != nil {
		// Check if source and destination are the same
		if pic == destPicID {
			fmt.Printf("  Warning: Cannot draw image to itself (pic=%d)\n", pic)
			return
		}

		// Clear the destination image first to prevent ghosting -> REMOVED
		// destImg.Clear()

		// Auto-expand destination if needed
		destPic := e.pictures[destPicID]
		requiredW := dx + w
		requiredH := dy + h

		if requiredW > destPic.Width || requiredH > destPic.Height {
			newW := destPic.Width
			newH := destPic.Height
			if requiredW > newW {
				newW = requiredW
			}
			if requiredH > newH {
				newH = requiredH
			}

			fmt.Printf("  Auto-expanding destination Picture ID=%d from %dx%d to %dx%d\n", destPicID, destPic.Width, destPic.Height, newW, newH)

			newImg := ebiten.NewImage(newW, newH)
			// Copy old image
			op := &ebiten.DrawImageOptions{}
			newImg.DrawImage(destPic.Image, op)

			destPic.Image = newImg
			destPic.Width = newW
			destPic.Height = newH
			destImg = newImg // Update local reference
		}

		// Draw source to destination
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(dx), float64(dy))

		// Use standard SourceOver for sprite composition
		// opts.CompositeMode = ebiten.CompositeModeCopy

		// Clip the source image
		srcRect := image.Rect(x, y, x+w, y+h)
		subImg := srcPic.Image.SubImage(srcRect).(*ebiten.Image)

		destImg.DrawImage(subImg, opts)
	}
}

// MovePic is a backward-compatible wrapper for the global state
func MovePic(pic, x, y, w, h, dest, dx, dy int, args ...any) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.MovePic(pic, x, y, w, h, dest, dx, dy, args...)
}

// Synchronization Globals
// VM / Sequencer Globals
type OpCode = interpreter.OpCode

// Variable represents a reference to a variable name
type Variable = interpreter.Variable

type Sequencer struct {
	commands     []OpCode
	pc           int
	waitTicks    int
	active       bool
	ticksPerStep int
	vars         map[string]any
	parent       *Sequencer // Parent scope for variable lookup
	mode         int        // Added mode
	onComplete   func()     // Completion callback
}

var (
	mainSequencer *Sequencer
	vmLock        sync.Mutex

	// Legacy
	tickCount    int64
	ticksPerStep int = 12

	// Restored Legacy Globals
	tickCond *sync.Cond
	// tickLock is alias for vmLock? No, Wait uses tickLock.
	// Let's redefine it or alias it.
	// tickLock sync.Mutex // Re-introduce
	tickLock     sync.Mutex // Add it back
	midiSyncMode bool

	GlobalPPQ int = 480 // Default pulses per quarter note
)

var queuedCallback func()

func StartQueuedCallback() {
	if queuedCallback != nil {
		go func() { queuedCallback() }()
	}
}

var (
	targetTick int64 // Atomic
)

func NotifyTick(tick int) {
	// Update target tick (Audio Thread)
	// We do NOT execute VM here to avoid threading issues with Ebiten/GPU
	atomic.StoreInt64(&targetTick, int64(tick))
}

// SetVMVar sets a variable in the VM for use in mes() blocks
// This is needed because Go local variables are not accessible in the VM
func SetVMVar(name string, value any) {
	vmLock.Lock()
	defer vmLock.Unlock()

	if mainSequencer == nil {
		// Create a root sequencer to hold variables
		mainSequencer = &Sequencer{
			vars:   make(map[string]any),
			parent: nil,
		}
	}

	mainSequencer.vars[strings.ToLower(name)] = value
}

// Assign is a helper function for transpiled code to set variables
// that need to be accessible in mes() blocks
func Assign(name string, value any) any {
	SetVMVar(name, value)
	return value
}

func RegisterSequence(mode int, ops []OpCode, initialVars ...map[string]any) {
	fmt.Printf("RegisterSequence: %d (%d ops)\n", mode, len(ops))

	var wg *sync.WaitGroup
	// Only block for TIME mode (procedural execution like robot)
	// MIDI_TIME must be non-blocking to allow PlayMIDI to execute
	if mode != MidiTime {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	vmLock.Lock()

	// Determine sync mode
	if mode == MidiTime { // 1
		midiSyncMode = true
		fmt.Println("RegisterSequence: MIDI Sync Mode ON")
		// In MIDI mode, targetTick is driven ONLY by NotifyTick from audio player
	} else {
		midiSyncMode = false
		fmt.Println("RegisterSequence: MIDI Sync Mode OFF (Time Mode)")
		// Ensure targetTick is at least current tickCount to enable immediate execution
		if atomic.LoadInt64(&targetTick) < tickCount {
			atomic.StoreInt64(&targetTick, tickCount)
		}
	}

	var onCompleteFunc func()
	if wg != nil {
		onCompleteFunc = func() { wg.Done() }
	}

	// Initialize vars map
	vars := make(map[string]any)
	// Copy initial variables if provided
	if len(initialVars) > 0 {
		for k, v := range initialVars[0] {
			vars[strings.ToLower(k)] = v // Case-insensitive
		}
	}

	// Save current sequencer as parent
	parentSeq := mainSequencer

	mainSequencer = &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true, // Auto-start
		ticksPerStep: 12,   // Default
		vars:         vars,
		parent:       parentSeq, // Set parent scope
		mode:         mode,      // Set mode
		onComplete:   onCompleteFunc,
	}
	vmLock.Unlock()

	// Wait for sequence to complete (only for TIME mode)
	if wg != nil {
		wg.Wait()
	}
}

// SetVMVar sets a variable in the VM for use in mes() blocks
// Tick the VM (Called from Conductor/NotifyTick)
func UpdateVM(currentTick int) {
	// Update global debug tick
	tickCount = int64(currentTick)
	vmLock.Lock()
	defer vmLock.Unlock()

	if mainSequencer == nil || !mainSequencer.active {
		return
	}

	seq := mainSequencer

	// Handle Wait
	if seq.waitTicks > 0 {
		// fmt.Printf("VM: Waiting... %d ticks remaining\n", seq.waitTicks) // Too noisy?
		seq.waitTicks--
		return
	}

	// Execute Instructions
	// We execute until we hit a Wait or End
	for seq.pc < len(seq.commands) {
		op := seq.commands[seq.pc]

		// Measure command execution time
		cmdStart := time.Now()

		// Debug Log (enabled for debugging)
		fmt.Printf("VM: Executing [%d] %s (Tick %d)\n", seq.pc, op.Cmd.String(), tickCount)

		seq.pc++

		// Execute Op
		_, yield := ExecuteOp(op, seq)

		// Log if command took long time
		cmdElapsed := time.Since(cmdStart)
		if cmdElapsed > 3*time.Millisecond {
			fmt.Printf("PERF: [%d] %s took %v\n", seq.pc-1, op.Cmd.String(), cmdElapsed)
		}

		if yield {
			// If ExecuteOp returns true, it means we must wait (Yield)
			// fmt.Printf("VM: Yielding for %d ticks\n", seq.waitTicks)
			break
		}
	}

	if seq.pc >= len(seq.commands) {
		// End of sequence
		seq.active = false
		fmt.Println("VM: Sequence Finished")
		if seq.onComplete != nil {
			seq.onComplete()
			seq.onComplete = nil // Ensure only called once
		}
	}
}

func ResolveArg(arg any, seq *Sequencer) any {
	switch v := arg.(type) {
	case Variable:
		// Case-insensitive variable lookup (FILLY is case-insensitive)
		varName := strings.ToLower(string(v))

		// Search in current scope and parent scopes
		currentSeq := seq
		for currentSeq != nil {
			if val, ok := currentSeq.vars[varName]; ok {
				return val
			}
			currentSeq = currentSeq.parent
		}

		// Variable not found in any scope
		// fmt.Printf("VM Warning: Variable %s not set, using 0\n", v)
		return 0
	case OpCode:
		// Nested OpCode evaluation
		res, _ := ExecuteOp(v, seq)
		return res
	default:
		return v
	}
}

func ExecuteOp(op OpCode, seq *Sequencer) (any, bool) {
	// Returns (Result, Yield)

	// Resolve Arguments first (except for Assign where first arg is name)
	// Actually, resolvedArgs helper might be needed per command if we want lazy evaluation
	// But mostly eager is fine.
	// For Assign, Arg[0] is name (Variable or string), Arg[1] is value.

	switch op.Cmd {
	case interpreter.OpLiteral:
		// Literal value - return as-is
		if len(op.Args) > 0 {
			return op.Args[0], false
		}
		return nil, false

	case interpreter.OpAssign:
		if len(op.Args) >= 2 {
			varName := ""
			if s, ok := op.Args[0].(string); ok {
				varName = strings.ToLower(s) // Case-insensitive (FILLY is case-insensitive)
			} else if v, ok := op.Args[0].(Variable); ok {
				varName = strings.ToLower(string(v)) // Case-insensitive (FILLY is case-insensitive)
			}

			if varName != "" {
				val := ResolveArg(op.Args[1], seq)
				fmt.Printf("VM: Assign %s = %v\n", varName, val)
				seq.vars[varName] = val
			}
		}
		return nil, false

	case interpreter.OpWait:
		// Args[0] = step count
		steps := 1
		if len(op.Args) > 0 {
			if s, ok := ResolveArg(op.Args[0], seq).(int); ok {
				steps = s
			}
		}

		// Calculate total ticks to wait
		// seq.ticksPerStep is already set via "SetStep" or defaults
		totalTicks := steps * seq.ticksPerStep
		if totalTicks < 1 {
			totalTicks = 1
		}

		fmt.Printf("VM: Wait(%d steps) -> %d ticks\n", steps, totalTicks)

		// Set wait state in Sequencer
		seq.waitTicks = totalTicks

		// Yield execution
		return nil, true

	case interpreter.OpSetStep:
		if len(op.Args) > 0 {
			if count, ok := ResolveArg(op.Args[0], seq).(int); ok && count > 0 {
				if seq.mode == 0 { // TIME mode
					// User feedback: 100ms was too slow, requested 50ms.
					// 60FPS -> 50ms is 3 ticks.
					// Duration = count * 3 ticks.
					seq.ticksPerStep = count * 3
					fmt.Printf("VM: SetStep(%d) in TIME mode -> %d ticks (%.2fs)\n", count, seq.ticksPerStep, float64(seq.ticksPerStep)/60.0)
				} else {
					if GlobalPPQ <= 0 {
						GlobalPPQ = 480
					} // safety

					// CURRENT: 32分音符 × n → step(8) = 480 ticks (4分音符相当) ← 正解
					seq.ticksPerStep = (GlobalPPQ / 8) * count
				}

				if seq.ticksPerStep == 0 {
					seq.ticksPerStep = 1
				}
				// DEBUG: Show which formula is active
				fmt.Printf("VM: SetStep(%d) -> ticksPerStep=%d (PPQ=%d, Mode=%d)\n",
					count, seq.ticksPerStep, GlobalPPQ, seq.mode)
			}
		}
		return nil, false

	case interpreter.OpStep:
		// step() block: Args[0] = count, Args[1] = body ([]OpCode)
		if len(op.Args) < 2 {
			return nil, false
		}

		count, ok := ResolveArg(op.Args[0], seq).(int)
		if !ok {
			count = 1
		}

		body, ok := op.Args[1].([]OpCode)
		if !ok {
			return nil, false
		}

		fmt.Printf("VM: Step(%d) with %d operations\n", count, len(body))

		// Execute each operation in the body 'count' times
		for i := 0; i < count; i++ {
			fmt.Printf("VM: Step iteration %d/%d\n", i+1, count)
			for opIdx, subOp := range body {
				fmt.Printf("VM: Step[%d/%d] executing op: %v\n", opIdx+1, len(body), subOp.Cmd)
				_, yield := ExecuteOp(subOp, seq)
				if yield {
					// If any operation yields (like Wait), we should yield too
					// This prevents blocking the main thread
					fmt.Printf("VM: Step yielding at iteration %d, op %d\n", i+1, opIdx+1)
					return nil, true
				}
			}
		}

		fmt.Printf("VM: Step completed all %d iterations\n", count)
		return nil, false

	// Engine Commands Wrappers

	case interpreter.OpLoadPic:
		if len(op.Args) >= 1 {
			if path, ok := ResolveArg(op.Args[0], seq).(string); ok {
				id := LoadPic(path)
				fmt.Printf("VM: LoadPic(%s) -> %d\n", path, id)
				return id, false
			}
		}
		return 0, false

	case interpreter.OpCreatePic:
		// Resolving args manually to build slice
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}
		id := CreatePic(rArgs...)
		return id, false

	case interpreter.OpPutCast:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}
		id := PutCast(rArgs...)
		return id, false

	case interpreter.OpDelPic:
		if len(op.Args) >= 1 {
			if id, ok := ResolveArg(op.Args[0], seq).(int); ok {
				DelPic(id)
			}
		}
		return nil, false

	case interpreter.OpMovePic:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}

		if len(rArgs) == 4 {
			// Legacy syntax: MovePic(srcPic, destPic, dx, dy)
			// Use entire source image
			srcPicID := rArgs[0].(int)
			destPicID := rArgs[1].(int)
			dx := rArgs[2].(int)
			dy := rArgs[3].(int)

			// WARNING: Do NOT add renderMutex.Lock() here!
			// MovePic function already acquires the lock (line 840)
			// Get source picture dimensions
			var srcPic *Picture
			if globalEngine != nil {
				srcPic = globalEngine.pictures[srcPicID]
			} else {
				srcPic = pictures[srcPicID]
			}
			if srcPic != nil {
				// Call MovePic with full source dimensions
				MovePic(srcPicID, 0, 0, srcPic.Width, srcPic.Height, destPicID, dx, dy)
			}
		} else if len(rArgs) >= 8 {
			// Full syntax: MovePic(pic, x, y, w, h, dest, dx, dy, ...)
			MovePic(
				rArgs[0].(int), rArgs[1].(int), rArgs[2].(int),
				rArgs[3].(int), rArgs[4].(int), rArgs[5].(int),
				rArgs[6].(int), rArgs[7].(int),
				rArgs[8:]..., // optional args
			)
		}
		return nil, false

	case interpreter.OpMoveCast:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}
		MoveCast(rArgs...)
		return nil, false

	case interpreter.OpOpenWin:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}

		// DEBUG logging when debugLevel >= 2
		if debugLevel >= 2 {
			fmt.Printf("DEBUG: OpenWin called with %d args: %v\n", len(rArgs), rArgs)
		}

		// Ensure we have at least 8 arguments, using 0 for missing ones
		// Note: 0 for width/height means "use picture dimensions"
		for len(rArgs) < 8 {
			rArgs = append(rArgs, 0)
		}

		OpenWin(rArgs[0].(int), rArgs[1].(int), rArgs[2].(int),
			rArgs[3].(int), rArgs[4].(int), rArgs[5].(int),
			rArgs[6].(int), rArgs[7].(int))
		return nil, false

	case interpreter.OpCloseWin:
		if len(op.Args) >= 1 {
			if id, ok := ResolveArg(op.Args[0], seq).(int); ok {
				CloseWin(id)
			}
		}
		return nil, false

	case interpreter.OpCloseWinAll:
		CloseWinAll()
		return nil, false

	case interpreter.OpTextColor:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}
		if len(rArgs) >= 3 {
			TextColor(rArgs[0].(int), rArgs[1].(int), rArgs[2].(int))
		}
		return nil, false

	case interpreter.OpTextWrite:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}
		if len(rArgs) >= 4 {
			TextWrite(rArgs[0].(string), rArgs[1].(int), rArgs[2].(int), rArgs[3].(int))
		}
		return nil, false

	case interpreter.OpPlayWAVE:
		if len(op.Args) >= 1 {
			if path, ok := ResolveArg(op.Args[0], seq).(string); ok {
				PlayWAVE(path)
			}
		}
		return nil, false

	case interpreter.OpPlayMIDI:
		if len(op.Args) >= 1 {
			if path, ok := ResolveArg(op.Args[0], seq).(string); ok {
				PlayMidiFile(path)
			}
		}
		return nil, false

	case interpreter.OpExitTitle:
		ExitTitle()
		return nil, false

	case interpreter.OpMoveWin:
		rArgs := make([]any, len(op.Args))
		for i, a := range op.Args {
			rArgs[i] = ResolveArg(a, seq)
		}
		if len(rArgs) >= 2 {
			// MoveWin function already acquires renderMutex, don't lock here
			// MoveWin(winID, picID, ...)
			// Logic: MoveWin(id, pic, x, y, w, h, picX, picY)
			// But Fill script might use fewer args?
			// KUMA2 uses: Cmd: "MoveWin", Args: []any{0, 1}
			// This likely means MoveWin(win, pic). Other args derived or default?
			// Let's guess: MoveWin(winID, picID) updates picture only?
			// Or MoveWin(winID, picID, x, y) updates pos?
			// Existing function signature: func MoveWin(winID, pic, x, y, w, h, picX, picY int)
			// If we only have 2 args, we probably need to fetch current values for the rest?

			if len(rArgs) == 2 {
				// Special case for KUMA2: MoveWin(0, 1) -> Switch Picture?
				// We need access to current window state
				winID := rArgs[0].(int)
				picID := rArgs[1].(int)

				var win *Window
				var ok bool
				if globalEngine != nil {
					globalEngine.renderMutex.Lock()
					win, ok = globalEngine.windows[winID]
					globalEngine.renderMutex.Unlock()
				} else {
					renderMutex.Lock()
					win, ok = windows[winID]
					renderMutex.Unlock()
				}

				if ok {
					x := win.X - BorderThickness
					y := win.Y - TitleBarHeight - BorderThickness
					w := win.W
					h := win.H
					srcX := -win.SrcX
					srcY := -win.SrcY
					// Update Picture, keep others
					MoveWin(winID, picID, x, y, w, h, srcX, srcY)
				}
			} else {
				// Ensure 8 args
				for len(rArgs) < 8 {
					rArgs = append(rArgs, 0)
				}
				MoveWin(rArgs[0].(int), rArgs[1].(int), rArgs[2].(int),
					rArgs[3].(int), rArgs[4].(int), rArgs[5].(int),
					rArgs[6].(int), rArgs[7].(int))
			}
		}
		return nil, false

	// Control Flow Operations
	case interpreter.OpIf:
		// Args: [0] = condition (OpCode or value), [1] = consequence ([]OpCode), [2] = alternative ([]OpCode, optional)
		if len(op.Args) < 2 {
			return nil, false
		}

		// Evaluate condition
		condResult := ResolveArg(op.Args[0], seq)
		condition := false

		// Convert result to boolean
		switch v := condResult.(type) {
		case bool:
			condition = v
		case int:
			condition = v != 0
		case string:
			condition = v != ""
		}

		// Execute appropriate branch
		if condition {
			// Execute consequence
			if conseq, ok := op.Args[1].([]OpCode); ok {
				for _, subOp := range conseq {
					// Check for break/continue before executing
					if subOp.Cmd == interpreter.OpBreak || subOp.Cmd == interpreter.OpContinue {
						// Return a special marker to indicate break/continue
						return subOp.Cmd.String(), false
					}
					result, yield := ExecuteOp(subOp, seq)
					if yield {
						return nil, true
					}
					// Propagate break/continue from nested structures
					if result == "Break" || result == "Continue" {
						return result, false
					}
				}
			}
		} else if len(op.Args) > 2 {
			// Execute alternative
			if alt, ok := op.Args[2].([]OpCode); ok {
				for _, subOp := range alt {
					// Check for break/continue before executing
					if subOp.Cmd == interpreter.OpBreak || subOp.Cmd == interpreter.OpContinue {
						return subOp.Cmd.String(), false
					}
					result, yield := ExecuteOp(subOp, seq)
					if yield {
						return nil, true
					}
					// Propagate break/continue from nested structures
					if result == "Break" || result == "Continue" {
						return result, false
					}
				}
			}
		}
		return nil, false

	case interpreter.OpFor:
		// Args: [0] = init (OpCode or nil), [1] = condition (OpCode or value), [2] = post (OpCode or nil), [3] = body ([]OpCode)
		if len(op.Args) < 4 {
			return nil, false
		}

		// Execute init
		if op.Args[0] != nil {
			if initOp, ok := op.Args[0].(OpCode); ok {
				ExecuteOp(initOp, seq)
			}
		}

		// Loop
		for {
			// Check condition
			if op.Args[1] != nil {
				condResult := ResolveArg(op.Args[1], seq)
				condition := false
				switch v := condResult.(type) {
				case bool:
					condition = v
				case int:
					condition = v != 0
				case string:
					condition = v != ""
				}
				if !condition {
					break
				}
			}

			// Execute body
			if body, ok := op.Args[3].([]OpCode); ok {
				shouldBreak := false
				shouldContinue := false
				for _, subOp := range body {
					if subOp.Cmd == interpreter.OpBreak {
						shouldBreak = true
						break
					}
					if subOp.Cmd == interpreter.OpContinue {
						shouldContinue = true
						break
					}
					result, yield := ExecuteOp(subOp, seq)
					if yield {
						return nil, true
					}
					// Check if nested structure returned break/continue
					if result == "Break" {
						shouldBreak = true
						break
					}
					if result == "Continue" {
						shouldContinue = true
						break
					}
				}
				if shouldBreak {
					break
				}
				if shouldContinue {
					// Execute post and continue
					if op.Args[2] != nil {
						if postOp, ok := op.Args[2].(OpCode); ok {
							ExecuteOp(postOp, seq)
						}
					}
					continue
				}
			}

			// Execute post
			if op.Args[2] != nil {
				if postOp, ok := op.Args[2].(OpCode); ok {
					ExecuteOp(postOp, seq)
				}
			}
		}
		return nil, false

	case interpreter.OpWhile:
		// Args: [0] = condition (OpCode or value), [1] = body ([]OpCode)
		if len(op.Args) < 2 {
			return nil, false
		}

		// Loop
		for {
			// Check condition
			condResult := ResolveArg(op.Args[0], seq)
			condition := false
			switch v := condResult.(type) {
			case bool:
				condition = v
			case int:
				condition = v != 0
			case string:
				condition = v != ""
			}
			if !condition {
				break
			}

			// Execute body
			if body, ok := op.Args[1].([]OpCode); ok {
				shouldBreak := false
				for _, subOp := range body {
					if subOp.Cmd == interpreter.OpBreak {
						shouldBreak = true
						break
					}
					if subOp.Cmd == interpreter.OpContinue {
						break
					}
					result, yield := ExecuteOp(subOp, seq)
					if yield {
						return nil, true
					}
					// Check if nested structure returned break/continue
					if result == "Break" {
						shouldBreak = true
						break
					}
					if result == "Continue" {
						break
					}
				}
				if shouldBreak {
					break
				}
			}
		}
		return nil, false

	case interpreter.OpDoWhile:
		// Args: [0] = condition (OpCode or value), [1] = body ([]OpCode)
		if len(op.Args) < 2 {
			return nil, false
		}

		// Execute body at least once
		for {
			// Execute body
			if body, ok := op.Args[1].([]OpCode); ok {
				shouldBreak := false
				for _, subOp := range body {
					if subOp.Cmd == interpreter.OpBreak {
						shouldBreak = true
						break
					}
					if subOp.Cmd == interpreter.OpContinue {
						break
					}
					result, yield := ExecuteOp(subOp, seq)
					if yield {
						return nil, true
					}
					// Check if nested structure returned break/continue
					if result == "Break" {
						shouldBreak = true
						break
					}
					if result == "Continue" {
						break
					}
				}
				if shouldBreak {
					break
				}
			}

			// Check condition
			condResult := ResolveArg(op.Args[0], seq)
			condition := false
			switch v := condResult.(type) {
			case bool:
				condition = v
			case int:
				condition = v != 0
			case string:
				condition = v != ""
			}
			if !condition {
				break
			}
		}
		return nil, false

	case interpreter.OpSwitch:
		// Args: [0] = value, [1] = cases ([]any where each is []any{caseValue, []OpCode}), [2] = default ([]OpCode or nil)
		if len(op.Args) < 2 {
			return nil, false
		}

		switchValue := ResolveArg(op.Args[0], seq)
		cases, ok := op.Args[1].([]any)
		if !ok {
			return nil, false
		}

		// Try to match a case
		matched := false
		for _, c := range cases {
			caseData, ok := c.([]any)
			if !ok || len(caseData) < 2 {
				continue
			}

			caseValue := ResolveArg(caseData[0], seq)
			if switchValue == caseValue {
				matched = true
				if body, ok := caseData[1].([]OpCode); ok {
					shouldBreak := false
					for _, subOp := range body {
						if subOp.Cmd == interpreter.OpBreak {
							shouldBreak = true
							break
						}
						_, yield := ExecuteOp(subOp, seq)
						if yield {
							return nil, true
						}
					}
					if shouldBreak {
						break
					}
				}
				break
			}
		}

		// Execute default if no match
		if !matched && len(op.Args) > 2 && op.Args[2] != nil {
			if defaultBody, ok := op.Args[2].([]OpCode); ok {
				for _, subOp := range defaultBody {
					_, yield := ExecuteOp(subOp, seq)
					if yield {
						return nil, true
					}
				}
			}
		}
		return nil, false

	case interpreter.OpBreak:
		// Break is handled by the loop execution logic
		return nil, false

	case interpreter.OpContinue:
		// Continue is handled by the loop execution logic
		return nil, false

	// Expression evaluation operations
	case interpreter.OpInfix:
		// Args: [0] = operator, [1] = left, [2] = right
		if len(op.Args) < 3 {
			return nil, false
		}

		operator, _ := op.Args[0].(string)
		left := ResolveArg(op.Args[1], seq)
		right := ResolveArg(op.Args[2], seq)

		// Convert to int for comparison
		leftInt, leftIsInt := left.(int)
		rightInt, rightIsInt := right.(int)

		if leftIsInt && rightIsInt {
			switch operator {
			case "==":
				return leftInt == rightInt, false
			case "!=":
				return leftInt != rightInt, false
			case "<":
				return leftInt < rightInt, false
			case ">":
				return leftInt > rightInt, false
			case "<=":
				return leftInt <= rightInt, false
			case ">=":
				return leftInt >= rightInt, false
			case "+":
				return leftInt + rightInt, false
			case "-":
				return leftInt - rightInt, false
			case "*":
				return leftInt * rightInt, false
			case "/":
				if rightInt != 0 {
					return leftInt / rightInt, false
				}
				return 0, false
			case "%":
				if rightInt != 0 {
					return leftInt % rightInt, false
				}
				return 0, false
			case "&&":
				return (leftInt != 0) && (rightInt != 0), false
			case "||":
				return (leftInt != 0) || (rightInt != 0), false
			}
		}

		// String comparison
		leftStr, leftIsStr := left.(string)
		rightStr, rightIsStr := right.(string)
		if leftIsStr && rightIsStr {
			switch operator {
			case "==":
				return leftStr == rightStr, false
			case "!=":
				return leftStr != rightStr, false
			case "+":
				return leftStr + rightStr, false
			}
		}

		return nil, false

	case interpreter.OpPrefix:
		// Args: [0] = operator, [1] = operand
		if len(op.Args) < 2 {
			return nil, false
		}

		operator, _ := op.Args[0].(string)
		operand := ResolveArg(op.Args[1], seq)

		if operandInt, ok := operand.(int); ok {
			switch operator {
			case "-":
				return -operandInt, false
			case "!":
				return operandInt == 0, false
			}
		}

		return nil, false

	case interpreter.OpIndex:
		// Args: [0] = array, [1] = index
		if len(op.Args) < 2 {
			return nil, false
		}

		array := ResolveArg(op.Args[0], seq)
		index := ResolveArg(op.Args[1], seq)

		if arr, ok := array.([]int); ok {
			if idx, ok := index.(int); ok && idx >= 0 && idx < len(arr) {
				return arr[idx], false
			}
		}

		return 0, false

	case interpreter.OpCall:
		// Function call: Args[0] = function name (string), Args[1..n] = arguments
		if len(op.Args) < 1 {
			return nil, false
		}

		funcName, ok := op.Args[0].(string)
		if !ok {
			return nil, false
		}

		// Resolve arguments
		args := make([]any, 0, len(op.Args)-1)
		for i := 1; i < len(op.Args); i++ {
			args = append(args, ResolveArg(op.Args[i], seq))
		}

		// Call the engine function
		return CallEngineFunction(funcName, args...)

	case interpreter.OpRegisterSequence:
		// Register a new sequence (mes block)
		// Args[0] = mode (TIME or MIDI_TIME), Args[1] = body ([]OpCode)
		if len(op.Args) < 2 {
			return nil, false
		}

		mode := ResolveArg(op.Args[0], seq)
		modeInt, ok := mode.(int)
		if !ok {
			modeInt = 0 // Default to TIME mode
		}

		body, ok := op.Args[1].([]OpCode)
		if !ok {
			return nil, false
		}

		// Register the sequence in a goroutine to avoid blocking the main thread
		// This is critical because RegisterSequence may call wg.Wait() which would
		// block the Ebiten game loop
		go RegisterSequence(modeInt, body)
		return nil, false

	default:
		// Check user defined functions
		var fn reflect.Value
		var ok bool
		cmdStr := op.Cmd.String()
		if globalEngine != nil {
			fn, ok = globalEngine.userFuncs[cmdStr]
		}
		if !ok {
			fn, ok = userFuncs[cmdStr]
		}

		if ok {
			// Prepare arguments
			in := make([]reflect.Value, len(op.Args))
			for i, arg := range op.Args {
				resolved := ResolveArg(arg, seq)
				in[i] = reflect.ValueOf(resolved)
			}
			// Call asynchronously to prevent blocking the VM/UI thread
			go func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Printf("Recovered from panic in UserFunc %s: %v\n", cmdStr, r)
					}
				}()
				fn.Call(in)
			}()
			return nil, false // VM proceeds immediately
		}

		fmt.Printf("VM Error: Unknown OpCmd %s\n", op.Cmd.String())
		return nil, false
	}
}

// CallEngineFunction calls an engine function by name with the given arguments
func CallEngineFunction(funcName string, args ...any) (any, bool) {
	// Normalize function name to lowercase for case-insensitive matching
	funcName = strings.ToLower(funcName)

	// Map function names to engine functions
	switch funcName {
	case "loadpic":
		if len(args) >= 1 {
			if filename, ok := args[0].(string); ok {
				LoadPic(filename)
			}
		}
		return nil, false

	case "openwin":
		if len(args) >= 1 {
			if winID, ok := args[0].(int); ok {
				// OpenWin with default parameters
				// OpenWin(pic, x, y, w, h, picX, picY, col)
				// For simplified call OpenWin(winID), use picture dimensions (w=0, h=0)
				fmt.Printf("VM: Calling OpenWin(%d, 0, 0, 0, 0, 0, 0, 0)\n", winID)
				OpenWin(winID, 0, 0, 0, 0, 0, 0, 0)
			}
		}
		return nil, false

	case "closewin":
		if len(args) >= 1 {
			if winID, ok := args[0].(int); ok {
				CloseWin(winID)
			}
		}
		return nil, false

	case "movewin":
		if len(args) >= 2 {
			winID, ok1 := args[0].(int)
			picID, ok2 := args[1].(int)
			if ok1 && ok2 {
				// MoveWin with default parameters
				// MoveWin(winID, pic, x, y, w, h, picX, picY)
				// For simplified call MoveWin(winID, picID), use defaults
				MoveWin(winID, picID, 0, 0, 640, 480, 0, 0)
			}
		}
		return nil, false

	case "playwave":
		if len(args) >= 1 {
			if filename, ok := args[0].(string); ok {
				PlayWAVE(filename)
			}
		}
		return nil, false

	case "playmidi":
		if len(args) >= 1 {
			if filename, ok := args[0].(string); ok {
				PlayMIDI(filename)
			}
		}
		return nil, false

	case "del_all":
		// Delete all resources
		fmt.Printf("VM: del_all called\n")
		return nil, false

	case "del_me":
		// Delete current sequence
		fmt.Printf("VM: del_me called\n")
		return nil, false

	default:
		fmt.Printf("VM Warning: Unknown function %s\n", funcName)
		return nil, false
	}
}

var userFuncs = make(map[string]reflect.Value)

func RegisterUserFunc(name string, f any) {
	if globalEngine != nil {
		globalEngine.userFuncs[name] = reflect.ValueOf(f)
	}
	userFuncs[name] = reflect.ValueOf(f)
}

func StrPrint(format string, args ...any) string {
	// Convert FILLY format specifiers to Go format specifiers
	// %ld -> %d (decimal integer)
	// %lx -> %x (hexadecimal)
	// %s remains %s (string)
	goFormat := strings.ReplaceAll(format, "%ld", "%d")
	goFormat = strings.ReplaceAll(goFormat, "%lx", "%x")
	return fmt.Sprintf(goFormat, args...)
}

func StrLen(s string) int {
	return len(s)
}

func SubStr(s string, start, length int) string {
	if start >= len(s) {
		return ""
	}
	end := start + length
	if end > len(s) {
		end = len(s)
	}
	return s[start:end]
}

func StrFind(s string, sub string) int {
	// Return index
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

func StrCode(val int) string {
	// Convert character code to single-character string
	if val < 0 || val > 0x10FFFF {
		return ""
	}
	return string(rune(val))
}

// StrInput displays a prompt and returns user input as a string
// Note: This is a console-based implementation for cross-platform compatibility
func StrInput(prompt string) string {
	fmt.Print(prompt)
	var input string
	fmt.Scanln(&input)
	return input
}

// CharCode returns the character code of the first character in the string
func CharCode(s string) int {
	if len(s) == 0 {
		return 0
	}
	// Get the first rune (character) from the string
	r := []rune(s)[0]
	return int(r)
}

// StrUp converts all lowercase letters to uppercase
func StrUp(s string) string {
	return strings.ToUpper(s)
}

// StrLow converts all uppercase letters to lowercase
func StrLow(s string) string {
	return strings.ToLower(s)
}

// PutCast creates a new cast (EngineState method)
func (e *EngineState) PutCast(args ...any) int {
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()
	fmt.Printf("PutCast: args=%v\n", args)

	// Parse arguments - varies by usage
	// Common: PutCast(picID, destPic, x, y, ...)
	if len(args) < 4 {
		return 0
	}

	picID, _ := args[0].(int)
	destPic, _ := args[1].(int)
	x, _ := args[2].(int)
	y, _ := args[3].(int)

	// Get source picture
	srcPic, ok := e.pictures[picID]
	if !ok {
		return 0
	}

	// Get transparent color (only if explicitly specified)
	var transparentColor color.Color
	var hasTransparency bool = false

	if len(args) > 4 {
		if c, ok := args[4].(int); ok {
			r := uint8((c >> 16) & 0xff)
			g := uint8((c >> 8) & 0xff)
			b := uint8(c & 0xff)
			transparentColor = color.RGBA{r, g, b, 255}
			hasTransparency = true
			fmt.Printf("PutCast: Transparent color specified: #%06x\n", c)
		}
	}

	// If transparency is specified, create a transparency-processed image ONCE
	var castPicID int
	if hasTransparency {
		// Create transparency-processed image
		processedImg := convertTransparentColor(srcPic.Image, transparentColor)

		// Store as a new Picture
		castPicID = e.nextPicID
		e.nextPicID++
		e.pictures[castPicID] = &Picture{
			ID:     castPicID,
			Image:  processedImg,
			Width:  srcPic.Width,
			Height: srcPic.Height,
		}
		fmt.Printf("PutCast: Created transparency-processed picture ID=%d\n", castPicID)
	} else {
		// No transparency - use original picture
		castPicID = picID
	}

	// Create cast
	// Already holding renderMutex from line 1423
	castID := e.nextCastID
	e.nextCastID++

	// Track order
	e.castDrawOrder = append(e.castDrawOrder, castID)

	// Determine initial dimensions (can be overridden by args if implemented in PutCast later)
	width := srcPic.Width
	height := srcPic.Height
	srcX := 0
	srcY := 0

	// Check if PutCast has size arguments (indices 8,9) and src offsets (indices 10,11)
	if len(args) > 9 {
		if w, ok := args[8].(int); ok {
			width = w
		}
		if h, ok := args[9].(int); ok {
			height = h
		}
	}
	if len(args) > 11 {
		if sx, ok := args[10].(int); ok {
			srcX = sx
		}
		if sy, ok := args[11].(int); ok {
			srcY = sy
		}
	}

	e.casts[castID] = &Cast{
		ID:          castID,
		Picture:     castPicID, // Use transparency-processed picture ID if transparency was specified
		DestPicture: destPic,
		X:           x,
		Y:           y,
		W:           width,
		H:           height,
		SrcX:        srcX,
		SrcY:        srcY,
		Transparent: transparentColor, // Store for reference
		Visible:     true,
	}

	// Draw the cast immediately to the destination picture
	destPicture := e.pictures[destPic]
	if destPicture != nil && destPicture.Image != nil {
		castPic := e.pictures[castPicID]
		if castPic != nil && castPic.Image != nil {
			// Determine the region to draw
			var imgToDraw *ebiten.Image = castPic.Image

			// Apply clipping if width/height are specified
			if width > 0 && height > 0 && (srcX > 0 || srcY > 0 || width < castPic.Width || height < castPic.Height) {
				rMaxX := srcX + width
				rMaxY := srcY + height

				// Cap to image bounds
				if rMaxX > castPic.Width {
					rMaxX = castPic.Width
				}
				if rMaxY > castPic.Height {
					rMaxY = castPic.Height
				}

				if rMaxX > srcX && rMaxY > srcY {
					rect := image.Rect(srcX, srcY, rMaxX, rMaxY)
					imgToDraw = castPic.Image.SubImage(rect).(*ebiten.Image)
				}
			}

			// Draw to destination picture (transparency already processed in the image)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(x), float64(y))
			destPicture.Image.DrawImage(imgToDraw, opts)

			// DEBUG: Show Cast ID if debug level is 2 and font is available
			if debugLevel >= 2 && e.currentFont != nil {
				castLabel := fmt.Sprintf("C%d", castID)
				text.Draw(destPicture.Image, castLabel, e.currentFont, x+5, y+15, color.RGBA{255, 255, 0, 255})
			}

			fmt.Printf("  Created and drew cast ID=%d at %d,%d\n", castID, x, y)
		}
	}

	return castID
}

// PutCast is a backward-compatible wrapper for the global state
func PutCast(args ...any) int {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return -1
	}
	return globalEngine.PutCast(args...)
}

// MoveCast moves a cast and redraws the destination picture (EngineState method)
func (e *EngineState) MoveCast(args ...any) {
	fmt.Printf("MoveCast: args=%v\n", args)

	// Parse arguments
	if len(args) < 3 {
		return
	}

	castID, _ := args[0].(int)
	picID, _ := args[1].(int)
	x, _ := args[2].(int)
	y := 0
	if len(args) > 3 {
		y, _ = args[3].(int)
	}

	// Get cast
	cast, ok := e.casts[castID]
	if !ok {
		fmt.Printf("  ERROR: Cast ID=%d not found\n", castID)
		return
	}

	// Get source picture
	srcPic, ok := e.pictures[cast.Picture]
	if !ok {
		fmt.Printf("  ERROR: Source picture ID=%d not found\n", cast.Picture)
		return
	}

	// Get destination picture (from cast's stored destination)
	destPic := e.pictures[cast.DestPicture]
	if destPic == nil {
		fmt.Printf("  ERROR: Dest picture ID=%d not found\n", cast.DestPicture)
		return
	}

	// Update position FIRST
	cast.X = x
	cast.Y = y

	// NOTE: picID parameter in MoveCast is typically the ORIGINAL source picture ID,
	// not the processed picture ID. We should NOT update cast.Picture here because
	// it would overwrite the transparency-processed picture with the original.
	// The cast should continue using the picture ID assigned during PutCast.

	// Only update if picID is explicitly different and non-zero
	// (This handles animation frame changes where a NEW picture is provided)
	// For now, we ignore picID to preserve transparency processing

	// DEBUG: Analyze arguments for clipping info
	fmt.Printf("DEBUG: MoveCast CastID=%d PicID=%d Args=%v\n", castID, picID, args)
	fmt.Printf("DEBUG: Source Picture Size: %dx%d\n", srcPic.Width, srcPic.Height)
	// fmt.Printf("  About to Fill and redraw casts for dest=%d\n", cast.DestPicture)

	// Parse additional arguments for dimensions and source offset
	// Args indices based on log analysis:
	// 0: castID, 1: picID, 2: x, 3: y
	// 4: transparent? (ignored/unknown)
	// 5: width, 6: height
	// 7: srcX, 8: srcY

	if len(args) > 5 {
		if w, ok := args[5].(int); ok && w > 0 {
			cast.W = w
		}
	}
	if len(args) > 6 {
		if h, ok := args[6].(int); ok && h > 0 {
			cast.H = h
		}
	}
	// Important: Reset SrcX/SrcY to 0 if not provided?
	// The script seems to provide them explicitly when changing frames.
	// But if we change picture (picID > 0), we should verify if we need to reset.
	// For safety, if provided, update. If not provided... keep previous?
	// Based on MoveCast behavior, it usually passes all args.
	if len(args) > 7 {
		if sx, ok := args[7].(int); ok {
			cast.SrcX = sx
		}
	}
	if len(args) > 8 {
		if sy, ok := args[8].(int); ok {
			cast.SrcY = sy
		}
	}

	fmt.Printf("DEBUG: MoveCast Updated Cast ID=%d: Pos=(%d,%d) Size=%dx%d Src=(%d,%d)\n",
		castID, cast.X, cast.Y, cast.W, cast.H, cast.SrcX, cast.SrcY)

	// Double Buffering & Thread Safety:
	// Protect image manipulation with Mutex to prevent race conditions with Game.Draw
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	// Initialize/Resize BackBuffer
	if destPic.BackBuffer == nil || destPic.BackBuffer.Bounds().Dx() != destPic.Width || destPic.BackBuffer.Bounds().Dy() != destPic.Height {
		destPic.BackBuffer = ebiten.NewImage(destPic.Width, destPic.Height)
		// Ensure main Image is also valid/sized
		if destPic.Image == nil {
			destPic.Image = ebiten.NewImage(destPic.Width, destPic.Height)
		}
	}

	// We draw to BackBuffer
	targetImg := destPic.BackBuffer

	// Clear to transparent (not white) to preserve cast transparency
	targetImg.Clear()

	// Redraw all casts that belong to this destination picture onto the new image
	// Use explicit draw order (creation order)
	redrawCount := 0

	fmt.Printf("DEBUG MoveCast: Redrawing casts for dest picture %d\n", cast.DestPicture)
	for _, cID := range e.castDrawOrder {
		c, exists := e.casts[cID]
		if !exists {
			continue
		}

		if c.Visible && c.DestPicture == cast.DestPicture {
			castSrc := e.pictures[c.Picture]
			if castSrc != nil {
				fmt.Printf("  Cast %d: Pic=%d Pos=(%d,%d) Size=%dx%d SrcOffset=(%d,%d)\n",
					cID, c.Picture, c.X, c.Y, c.W, c.H, c.SrcX, c.SrcY)

				// Clip image if Cast has specific Width/Height
				var imgToDraw *ebiten.Image = castSrc.Image

				// Validate clipping rectangle
				if c.W > 0 && c.H > 0 {
					// Ensure we don't go out of bounds
					rMaxX := c.SrcX + c.W
					rMaxY := c.SrcY + c.H

					// Cap to image bounds
					if rMaxX > castSrc.Width {
						rMaxX = castSrc.Width
					}
					if rMaxY > castSrc.Height {
						rMaxY = castSrc.Height
					}

					// Only subimage if valid
					if rMaxX > c.SrcX && rMaxY > c.SrcY {
						rect := image.Rect(c.SrcX, c.SrcY, rMaxX, rMaxY)
						imgToDraw = castSrc.Image.SubImage(rect).(*ebiten.Image)
						fmt.Printf("    Clipped to rect: %v\n", rect)
					}
				}

				// NOTE: Transparency is already processed at Cast creation time (PutCast)
				// The Cast.Picture field references the transparency-processed image
				// DO NOT process transparency again here - it's inefficient and redundant

				// Draw to targetImg using native alpha blending
				opts := &ebiten.DrawImageOptions{}
				opts.GeoM.Translate(float64(c.X), float64(c.Y))
				targetImg.DrawImage(imgToDraw, opts)

				// DEBUG: Show Cast ID if debug level is 2 and font is available
				if debugLevel >= 2 && e.currentFont != nil {
					castLabel := fmt.Sprintf("C%d", cID)
					text.Draw(targetImg, castLabel, e.currentFont, c.X+5, c.Y+15, color.RGBA{255, 255, 0, 255})
				}

				redrawCount++
			}
		}
	}
	fmt.Printf("  Total casts redrawn: %d\n", redrawCount)

	// Atomic swap (effectively) of the image
	// The main loop will verify this pointer in the next frame
	// Reset BackBuffer pointer to the OLD Image (so we rotate generic buffers)
	// Wait, ebiten.Image semantics: if we just swap pointers, we are good.
	// dest.Image (Old) becomes BackBuffer (Next).
	// dest.BackBuffer (New Content) becomes Image (Current).

	temp := destPic.Image
	destPic.Image = destPic.BackBuffer
	destPic.BackBuffer = temp

	// fmt.Printf("  Redrawn %d casts on Pic %d\n", redrawCount, cast.DestPicture)

	// fmt.Printf("  Redrawn %d casts on Pic %d\n", redrawCount, cast.DestPicture)
	// fmt.Printf("  Redrawn %d casts on Pic %d\n", redrawCount, cast.DestPicture)

	// fmt.Printf("  Moved cast ID=%d to %d,%d\n", castID, x, y)
}

// MoveCast is a backward-compatible wrapper for the global state
func MoveCast(args ...any) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.MoveCast(args...)
}

// DelCast deletes a cast (EngineState method)
func (e *EngineState) DelCast(args ...any) {
	fmt.Printf("DelCast: args=%v\n", args)
	e.renderMutex.Lock()
	defer e.renderMutex.Unlock()

	if len(args) > 0 {
		if castID, ok := args[0].(int); ok {
			delete(e.casts, castID)

			// Remove from draw order
			for i, id := range e.castDrawOrder {
				if id == castID {
					e.castDrawOrder = append(e.castDrawOrder[:i], e.castDrawOrder[i+1:]...)
					break
				}
			}
		}
	}
}

// DelCast is a backward-compatible wrapper for the global state
func DelCast(args ...any) {
	if globalEngine == nil {
		fmt.Println("Error: globalEngine not initialized")
		return
	}
	globalEngine.DelCast(args...)
}

// GetPicture returns a picture by ID (for testing) - EngineState method
func (e *EngineState) GetPicture(id int) *Picture {
	return e.pictures[id]
}

// GetPicture returns a picture by ID (for testing)
func GetPicture(id int) *Picture {
	return pictures[id]
}

// GetWindow returns a window by ID (for testing) - EngineState method
func (e *EngineState) GetWindow(id int) *Window {
	return e.windows[id]
}

// GetCast returns a cast by ID (for testing) - EngineState method
func (e *EngineState) GetCast(id int) *Cast {
	return e.casts[id]
}

// drawWithColorKey draws src to dst with color key transparency
func drawWithColorKey(dst, src *ebiten.Image, x, y int, transparentColor color.Color) {
	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Create a new image with transparency applied
	tempImg := ebiten.NewImage(width, height)

	// Get RGBA values of transparent color
	tr, tg, tb, _ := transparentColor.RGBA()

	// Copy pixels, skipping transparent color
	for py := 0; py < height; py++ {
		for px := 0; px < width; px++ {
			// FIXED: Apply bounds offset for SubImages support
			c := src.At(px+bounds.Min.X, py+bounds.Min.Y)
			r, g, b, a := c.RGBA()

			// Skip if matches transparent color
			if r == tr && g == tg && b == tb {
				continue
			}

			// Draw pixel
			tempImg.Set(px, py, color.RGBA{
				R: uint8(r >> 8),
				G: uint8(g >> 8),
				B: uint8(b >> 8),
				A: uint8(a >> 8),
			})
		}
	}

	// Draw to destination
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(x), float64(y))
	dst.DrawImage(tempImg, opts)
}

// PicWidth returns picture width (EngineState method)
func (e *EngineState) PicWidth(pic int) int {
	if p, ok := e.pictures[pic]; ok {
		return p.Width
	}
	return 100 // Default fallback
}

// PicWidth returns picture width
func PicWidth(pic int) int {
	if globalEngine != nil {
		return globalEngine.PicWidth(pic)
	}
	if p, ok := pictures[pic]; ok {
		return p.Width
	}
	return 100 // Default fallback
}

// PicHeight returns picture height (EngineState method)
func (e *EngineState) PicHeight(pic int) int {
	if p, ok := e.pictures[pic]; ok {
		return p.Height
	}
	return 100 // Default fallback
}

// PicHeight returns picture height
func PicHeight(pic int) int {
	if globalEngine != nil {
		return globalEngine.PicHeight(pic)
	}
	if p, ok := pictures[pic]; ok {
		return p.Height
	}
	return 100 // Default fallback
}

func Random(max int) int {
	if max <= 0 {
		return 0
	}
	return int(time.Now().UnixNano()) % max
}

func PlayMIDI(args ...any) {
	fmt.Printf("PlayMIDI: %v\n", args)
	if len(args) > 0 {
		if path, ok := args[0].(string); ok {
			// Try to play MIDI file
			// If it fails (e.g., no SoundFont), just return without locking
			if soundFont == nil {
				fmt.Println("PlayMIDI: Skipping - No SoundFont loaded")
				return
			}

			PlayMidiFile(path)

			// Reset ticks to ensure synchronization with new MIDI track
			vmLock.Lock()
			tickCount = 0
			atomic.StoreInt64(&targetTick, 0)
			// targetTick will be updated by NotifyTick from MIDI player
			vmLock.Unlock()

			// VM Path: Activate Sequence
			vmLock.Lock()
			if mainSequencer != nil {
				mainSequencer.active = true
				fmt.Println("PlayMIDI: Sequence Activated")
			}
			vmLock.Unlock()

			// Legacy Path: Deferred Callback (if any)
			StartQueuedCallback()
		}
	}
}

// MoveSPic scales and copies a picture region to a destination
// Args: srcPicID, srcX, srcY, srcW, srcH, dstPicID, dstX, dstY, dstW, dstH, [transparentColor]
func MoveSPic(args ...any) {
	if len(args) < 10 {
		return
	}
	srcPicID, _ := args[0].(int)
	srcX, _ := args[1].(int)
	srcY, _ := args[2].(int)
	srcW, _ := args[3].(int)
	srcH, _ := args[4].(int)
	dstPicID, _ := args[5].(int)
	dstX, _ := args[6].(int)
	dstY, _ := args[7].(int)
	dstW, _ := args[8].(int)
	dstH, _ := args[9].(int)

	// Optional transparent color (RGB as separate args or as color.Color)
	var transparentColor color.Color
	hasTransparency := false
	if len(args) >= 13 {
		r, _ := args[10].(int)
		g, _ := args[11].(int)
		b, _ := args[12].(int)
		transparentColor = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
		hasTransparency = true
	} else if len(args) >= 11 {
		if c, ok := args[10].(color.Color); ok {
			transparentColor = c
			hasTransparency = true
		}
	}

	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		srcPic, ok1 := globalEngine.pictures[srcPicID]
		dstPic, ok2 := globalEngine.pictures[dstPicID]
		if !ok1 || !ok2 {
			return
		}

		// Create subimage from source
		rect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
		subImg := srcPic.Image.SubImage(rect).(*ebiten.Image)

		// Create a temporary image for the scaled result
		scaledImg := ebiten.NewImage(dstW, dstH)

		// Calculate scale factors
		scaleX := float64(dstW) / float64(srcW)
		scaleY := float64(dstH) / float64(srcH)

		// Apply scaling transformation
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(scaleX, scaleY)

		// If transparency is specified, apply color key
		if hasTransparency {
			// Draw with transparency by checking each pixel
			// For performance, we use Ebitengine's built-in filtering
			// The transparency will be handled by pre-processing the image
			opts.Filter = ebiten.FilterLinear // Use linear interpolation for scaling
			scaledImg.DrawImage(subImg, opts)

			// Apply transparency post-scaling
			// This is a simplified approach - for better performance,
			// transparency should be pre-processed in the source image
			bounds := scaledImg.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					c := scaledImg.At(x, y)
					r1, g1, b1, a1 := c.RGBA()
					r2, g2, b2, _ := transparentColor.RGBA()
					// Compare colors (with some tolerance for interpolation artifacts)
					if abs(int(r1>>8)-int(r2>>8)) < 5 &&
						abs(int(g1>>8)-int(g2>>8)) < 5 &&
						abs(int(b1>>8)-int(b2>>8)) < 5 {
						// Make transparent
						scaledImg.Set(x, y, color.RGBA{0, 0, 0, 0})
					} else {
						// Keep original color with alpha
						scaledImg.Set(x, y, color.RGBA{uint8(r1 >> 8), uint8(g1 >> 8), uint8(b1 >> 8), uint8(a1 >> 8)})
					}
				}
			}
		} else {
			opts.Filter = ebiten.FilterLinear // Use linear interpolation for scaling
			scaledImg.DrawImage(subImg, opts)
		}

		// Draw the scaled image to the destination
		dstOpts := &ebiten.DrawImageOptions{}
		dstOpts.GeoM.Translate(float64(dstX), float64(dstY))
		dstPic.Image.DrawImage(scaledImg, dstOpts)

	} else {
		// Fallback to legacy globals
		renderMutex.Lock()
		defer renderMutex.Unlock()

		srcPic, ok1 := pictures[srcPicID]
		dstPic, ok2 := pictures[dstPicID]
		if !ok1 || !ok2 {
			return
		}

		// Create subimage from source
		rect := image.Rect(srcX, srcY, srcX+srcW, srcY+srcH)
		subImg := srcPic.Image.SubImage(rect).(*ebiten.Image)

		// Create a temporary image for the scaled result
		scaledImg := ebiten.NewImage(dstW, dstH)

		// Calculate scale factors
		scaleX := float64(dstW) / float64(srcW)
		scaleY := float64(dstH) / float64(srcH)

		// Apply scaling transformation
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(scaleX, scaleY)

		// If transparency is specified, apply color key
		if hasTransparency {
			opts.Filter = ebiten.FilterLinear
			scaledImg.DrawImage(subImg, opts)

			// Apply transparency post-scaling
			bounds := scaledImg.Bounds()
			for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
				for x := bounds.Min.X; x < bounds.Max.X; x++ {
					c := scaledImg.At(x, y)
					r1, g1, b1, a1 := c.RGBA()
					r2, g2, b2, _ := transparentColor.RGBA()
					if abs(int(r1>>8)-int(r2>>8)) < 5 &&
						abs(int(g1>>8)-int(g2>>8)) < 5 &&
						abs(int(b1>>8)-int(b2>>8)) < 5 {
						scaledImg.Set(x, y, color.RGBA{0, 0, 0, 0})
					} else {
						scaledImg.Set(x, y, color.RGBA{uint8(r1 >> 8), uint8(g1 >> 8), uint8(b1 >> 8), uint8(a1 >> 8)})
					}
				}
			}
		} else {
			opts.Filter = ebiten.FilterLinear
			scaledImg.DrawImage(subImg, opts)
		}

		// Draw the scaled image to the destination
		dstOpts := &ebiten.DrawImageOptions{}
		dstOpts.GeoM.Translate(float64(dstX), float64(dstY))
		dstPic.Image.DrawImage(scaledImg, dstOpts)
	}
}

func ReversePic(args ...any) {
	if len(args) < 8 {
		return
	}
	srcPicID, _ := args[0].(int)
	srcX, _ := args[1].(int)
	srcY, _ := args[2].(int)
	width, _ := args[3].(int)
	height, _ := args[4].(int)
	dstPicID, _ := args[5].(int)
	dstX, _ := args[6].(int)
	dstY, _ := args[7].(int)

	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		srcPic, ok1 := globalEngine.pictures[srcPicID]
		dstPic, ok2 := globalEngine.pictures[dstPicID]
		if !ok1 || !ok2 {
			return
		}

		// Create subimage
		rect := image.Rect(srcX, srcY, srcX+width, srcY+height)
		subImg := srcPic.Image.SubImage(rect).(*ebiten.Image)

		// Flip horizontally
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(-1, 1)                            // Flip X
		opts.GeoM.Translate(float64(width), 0)            // Shift back (since flip moves it to -width)
		opts.GeoM.Translate(float64(dstX), float64(dstY)) // Move to dest

		// Draw
		dstPic.Image.DrawImage(subImg, opts)
	} else {
		// Fallback to legacy globals
		renderMutex.Lock()
		defer renderMutex.Unlock()

		srcPic, ok1 := pictures[srcPicID]
		dstPic, ok2 := pictures[dstPicID]
		if !ok1 || !ok2 {
			return
		}

		// Create subimage
		rect := image.Rect(srcX, srcY, srcX+width, srcY+height)
		subImg := srcPic.Image.SubImage(rect).(*ebiten.Image)

		// Flip horizontally
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Scale(-1, 1)                            // Flip X
		opts.GeoM.Translate(float64(width), 0)            // Shift back (since flip moves it to -width)
		opts.GeoM.Translate(float64(dstX), float64(dstY)) // Move to dest

		// Draw
		dstPic.Image.DrawImage(subImg, opts)
	}
}

// GetPicNo returns the Picture ID associated with a Window
func GetPicNo(winID int) int {
	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		if win, ok := globalEngine.windows[winID]; ok {
			return win.Picture
		}
		return -1 // Invalid window ID
	} else {
		// Fallback to legacy globals
		renderMutex.Lock()
		defer renderMutex.Unlock()

		if win, ok := windows[winID]; ok {
			return win.Picture
		}
		return -1 // Invalid window ID
	}
}

func GetMesNo(id int) int {
	return id
}

func DelMes(id int) {
	fmt.Printf("DelMes: %d\n", id)
}

func PostMes(args ...any) {
	fmt.Println("PostMes")
}

func Sc1Sub1(p []int)                                  { fmt.Println("Sc1Sub1") }
func Sc1Sub2(p []int)                                  { fmt.Println("Sc1Sub2") }
func Sc1Sub3(p []int)                                  { fmt.Println("Sc1Sub3") }
func Sc1Sub4(p []int)                                  { fmt.Println("Sc1Sub4") }
func CIText(stexts int, p int, x int, y int, time int) { fmt.Println("CIText Stub") }

// Procedural Execution State

var (
	procMode      int = 0 // 0: TIME, 1: MIDI_TIME
	procStep      int = 6 // Default 6 ticks (100ms) for compat, or initialized
	procWaitTicks int = 0
)

func EnterMes(mode int) {
	procMode = mode
	// Reset step defaults?
	// In TIME mode, SetStep(20) -> 1000ms. 50ms base.
	// In MIDI mode, SetStep(8) -> Quarter note.
	// Defaults:
	if mode == 0 { // TIME
		procStep = 6 * 3 // Default? Or 1? Let's wait for SetStep.
		// NOTE: Robot sample calls SetStep immediately after EnterMes usually.
	} else {
		// MIDI
		procStep = 24 // Default?
	}
}

func ExitMes() {
	// Reset/Cleanup if needed
}

func SetStep(n int) {
	if procMode == 0 { // TIME
		// 50ms base unit. 60FPS. 50ms = 3 ticks.
		// Wait(1) should wait n * 3 ticks.
		// storing ticks per unit in procStep
		procStep = n * 3
	} else { // MIDI
		// n * 32nd note.
		// GlobalPPQ = 480 (usually, or 96). 32nd note = PPQ / 8?
		// 4th note = 1 step of 8.
		// If step(8) = 4th note = PPQ.
		// Then step(1) = PPQ / 8.
		// Formula: Unit = (GlobalPPQ / 8) * n
		if GlobalPPQ == 0 {
			GlobalPPQ = 480
		} // Safer default
		procStep = (GlobalPPQ / 8) * n
	}
}

func Wait(n int) {
	// Block until duration passes

	if procMode == 0 { // TIME
		// duration in frame ticks
		ticksToWait := n * procStep
		startTick := tickCount // Capture current frame tick

		// Loop until specific ticks passed
		// We need to yield to avoid tight loop burn?
		// runtime.Gosched()? Or time.Sleep?
		// Since tickCount is updated by main thread, we just poll.
		// But we should sleep to save CPU. 1 frame = 16ms.
		for tickCount < startTick+int64(ticksToWait) {
			time.Sleep(1 * time.Millisecond)
		}
	} else { // MIDI
		// duration in MIDI ticks
		ticksToWait := n * procStep
		startMidTick := atomic.LoadInt64(&targetTick)

		// Wait until targetTick advances?
		// Wait, targetTick is the CURRENT MIDI time?
		// Yes, NotifyTick updates targetTick.
		// We want to wait until MIDI time advances by ticksToWait.

		target := startMidTick + int64(ticksToWait)
		for {
			current := atomic.LoadInt64(&targetTick)
			if current >= target {
				break
			}
			time.Sleep(1 * time.Millisecond)
		}
	}
}

func EndStep() {
	// Marker
}

func DelMe() {
	// End script goroutine
	runtime.Goexit()
}

func DelUs() {
	// Alias?
}

func DelAll() {
	// Alias?
}

func Maint() {
	// Maintenance?
}

// --- Drawing Functions ---

// Drawing state
var (
	currentLineSize   = 1
	currentPaintColor = color.RGBA{0, 0, 0, 255} // Black
	currentROP        = 0                        // COPYPEN (default)
)

// ROP (Raster Operation) codes
const (
	ROP_COPYPEN  = 0 // Copy source to destination
	ROP_XORPEN   = 1 // XOR source with destination
	ROP_MERGEPEN = 2 // OR source with destination
	ROP_NOTPEN   = 3 // NOT source
	ROP_MASKPEN  = 4 // AND source with destination
)

// SetLineSize sets the line width for subsequent drawing operations
func SetLineSize(size int) {
	if size < 1 {
		size = 1
	}
	currentLineSize = size
	fmt.Printf("SetLineSize: %d\n", size)
}

// SetPaintColor sets the drawing color for subsequent operations
func SetPaintColor(r, g, b int) {
	currentPaintColor = color.RGBA{uint8(r), uint8(g), uint8(b), 255}
	fmt.Printf("SetPaintColor: RGB(%d, %d, %d)\n", r, g, b)
}

// SetROP sets the raster operation mode for subsequent drawing
func SetROP(ropCode int) {
	currentROP = ropCode
	fmt.Printf("SetROP: %d\n", ropCode)
}

// DrawLine draws a line between two points on the specified Picture
func DrawLine(picID, x1, y1, x2, y2 int) {
	fmt.Printf("DrawLine: pic=%d from (%d,%d) to (%d,%d)\n", picID, x1, y1, x2, y2)

	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		pic, ok := globalEngine.pictures[picID]
		if !ok {
			fmt.Printf("  Picture ID=%d not found\n", picID)
			return
		}

		drawLineBresenham(pic.Image, x1, y1, x2, y2, currentPaintColor, currentLineSize)
	} else {
		renderMutex.Lock()
		defer renderMutex.Unlock()

		pic, ok := pictures[picID]
		if !ok {
			fmt.Printf("  Picture ID=%d not found\n", picID)
			return
		}

		drawLineBresenham(pic.Image, x1, y1, x2, y2, currentPaintColor, currentLineSize)
	}
}

// drawLineBresenham implements Bresenham's line algorithm with line width support
func drawLineBresenham(img *ebiten.Image, x1, y1, x2, y2 int, col color.Color, width int) {
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	sx := 1
	if x1 > x2 {
		sx = -1
	}
	sy := 1
	if y1 > y2 {
		sy = -1
	}
	err := dx - dy

	x, y := x1, y1
	for {
		// Draw pixel with width
		drawThickPixel(img, x, y, width, col)

		if x == x2 && y == y2 {
			break
		}

		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += sx
		}
		if e2 < dx {
			err += dx
			y += sy
		}
	}
}

// drawThickPixel draws a pixel with the specified width (creates a square)
func drawThickPixel(img *ebiten.Image, x, y, width int, col color.Color) {
	halfWidth := width / 2
	for dy := -halfWidth; dy <= halfWidth; dy++ {
		for dx := -halfWidth; dx <= halfWidth; dx++ {
			px, py := x+dx, y+dy
			if px >= 0 && py >= 0 && px < img.Bounds().Dx() && py < img.Bounds().Dy() {
				applyROP(img, px, py, col)
			}
		}
	}
}

// abs returns the absolute value of x
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// applyROP applies the current raster operation when drawing a pixel
func applyROP(img *ebiten.Image, x, y int, srcColor color.Color) {
	switch currentROP {
	case ROP_COPYPEN:
		// Simple copy
		img.Set(x, y, srcColor)

	case ROP_XORPEN:
		// XOR with destination
		dstColor := img.At(x, y)
		sr, sg, sb, sa := srcColor.RGBA()
		dr, dg, db, da := dstColor.RGBA()
		img.Set(x, y, color.RGBA{
			R: uint8((sr ^ dr) >> 8),
			G: uint8((sg ^ dg) >> 8),
			B: uint8((sb ^ db) >> 8),
			A: uint8((sa | da) >> 8),
		})

	case ROP_MERGEPEN:
		// OR with destination
		dstColor := img.At(x, y)
		sr, sg, sb, sa := srcColor.RGBA()
		dr, dg, db, da := dstColor.RGBA()
		img.Set(x, y, color.RGBA{
			R: uint8((sr | dr) >> 8),
			G: uint8((sg | dg) >> 8),
			B: uint8((sb | db) >> 8),
			A: uint8((sa | da) >> 8),
		})

	case ROP_NOTPEN:
		// NOT source
		sr, sg, sb, _ := srcColor.RGBA()
		img.Set(x, y, color.RGBA{
			R: uint8(^(sr >> 8)),
			G: uint8(^(sg >> 8)),
			B: uint8(^(sb >> 8)),
			A: 255,
		})

	case ROP_MASKPEN:
		// AND with destination
		dstColor := img.At(x, y)
		sr, sg, sb, sa := srcColor.RGBA()
		dr, dg, db, da := dstColor.RGBA()
		img.Set(x, y, color.RGBA{
			R: uint8((sr & dr) >> 8),
			G: uint8((sg & dg) >> 8),
			B: uint8((sb & db) >> 8),
			A: uint8((sa & da) >> 8),
		})

	default:
		// Default to COPYPEN
		img.Set(x, y, srcColor)
	}
}

// GetColor returns the RGB color value of the specified pixel
func GetColor(picID, x, y int) int {
	var col color.Color

	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		pic, ok := globalEngine.pictures[picID]
		if !ok {
			return 0
		}
		col = pic.Image.At(x, y)
	} else {
		renderMutex.Lock()
		defer renderMutex.Unlock()

		pic, ok := pictures[picID]
		if !ok {
			return 0
		}
		col = pic.Image.At(x, y)
	}

	// Convert to RGB integer (0xRRGGBB)
	r, g, b, _ := col.RGBA()
	return int((r>>8)<<16 | (g>>8)<<8 | (b >> 8))
}

// Fill modes
const (
	FILL_NONE  = 0 // No fill (outline only)
	FILL_HATCH = 1 // Hatch pattern fill
	FILL_SOLID = 2 // Solid fill
)

// DrawCircle draws a circle or ellipse with optional fill modes
// Args: picID, centerX, centerY, radiusX, radiusY, fillMode
func DrawCircle(args ...any) {
	if len(args) < 5 {
		fmt.Println("DrawCircle: insufficient arguments")
		return
	}

	picID := args[0].(int)
	cx := args[1].(int)
	cy := args[2].(int)
	rx := args[3].(int)
	ry := args[4].(int)
	fillMode := FILL_NONE
	if len(args) > 5 {
		fillMode = args[5].(int)
	}

	fmt.Printf("DrawCircle: pic=%d center=(%d,%d) radius=(%d,%d) fill=%d\n", picID, cx, cy, rx, ry, fillMode)

	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		pic, ok := globalEngine.pictures[picID]
		if !ok {
			fmt.Printf("  Picture ID=%d not found\n", picID)
			return
		}

		drawEllipse(pic.Image, cx, cy, rx, ry, currentPaintColor, fillMode)
	} else {
		renderMutex.Lock()
		defer renderMutex.Unlock()

		pic, ok := pictures[picID]
		if !ok {
			fmt.Printf("  Picture ID=%d not found\n", picID)
			return
		}

		drawEllipse(pic.Image, cx, cy, rx, ry, currentPaintColor, fillMode)
	}
}

// drawEllipse implements midpoint ellipse algorithm with fill support
func drawEllipse(img *ebiten.Image, cx, cy, rx, ry int, col color.Color, fillMode int) {
	if fillMode == FILL_SOLID {
		// Solid fill: draw filled ellipse
		drawFilledEllipse(img, cx, cy, rx, ry, col)
	} else if fillMode == FILL_HATCH {
		// Hatch fill: draw filled ellipse with hatch pattern
		drawHatchedEllipse(img, cx, cy, rx, ry, col)
	} else {
		// Outline only
		drawEllipseOutline(img, cx, cy, rx, ry, col)
	}
}

// drawEllipseOutline draws the outline of an ellipse using midpoint algorithm
func drawEllipseOutline(img *ebiten.Image, cx, cy, rx, ry int, col color.Color) {
	// Midpoint ellipse algorithm
	rx2 := rx * rx
	ry2 := ry * ry
	twoRx2 := 2 * rx2
	twoRy2 := 2 * ry2

	x := 0
	y := ry
	px := 0
	py := twoRx2 * y

	// Plot initial points
	plotEllipsePoints(img, cx, cy, x, y, col)

	// Region 1
	p := int(float64(ry2) - float64(rx2*ry) + (0.25 * float64(rx2)))
	for px < py {
		x++
		px += twoRy2
		if p < 0 {
			p += ry2 + px
		} else {
			y--
			py -= twoRx2
			p += ry2 + px - py
		}
		plotEllipsePoints(img, cx, cy, x, y, col)
	}

	// Region 2
	p = int(float64(ry2)*(float64(x)+0.5)*(float64(x)+0.5) + float64(rx2)*(float64(y)-1)*(float64(y)-1) - float64(rx2*ry2))
	for y > 0 {
		y--
		py -= twoRx2
		if p > 0 {
			p += rx2 - py
		} else {
			x++
			px += twoRy2
			p += rx2 - py + px
		}
		plotEllipsePoints(img, cx, cy, x, y, col)
	}
}

// plotEllipsePoints plots the 4 symmetric points of an ellipse
func plotEllipsePoints(img *ebiten.Image, cx, cy, x, y int, col color.Color) {
	bounds := img.Bounds()
	points := [][2]int{
		{cx + x, cy + y},
		{cx - x, cy + y},
		{cx + x, cy - y},
		{cx - x, cy - y},
	}

	for _, p := range points {
		if p[0] >= 0 && p[0] < bounds.Dx() && p[1] >= 0 && p[1] < bounds.Dy() {
			drawThickPixel(img, p[0], p[1], currentLineSize, col)
		}
	}
}

// drawFilledEllipse draws a solid filled ellipse
func drawFilledEllipse(img *ebiten.Image, cx, cy, rx, ry int, col color.Color) {
	bounds := img.Bounds()
	// Simple approach: iterate over bounding box and check if point is inside ellipse
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			// Check if point is inside ellipse: (x-cx)²/rx² + (y-cy)²/ry² <= 1
			dx := float64(x - cx)
			dy := float64(y - cy)
			if (dx*dx)/(float64(rx*rx))+(dy*dy)/(float64(ry*ry)) <= 1.0 {
				if x >= 0 && x < bounds.Dx() && y >= 0 && y < bounds.Dy() {
					applyROP(img, x, y, col)
				}
			}
		}
	}
}

// drawHatchedEllipse draws a hatch-filled ellipse
func drawHatchedEllipse(img *ebiten.Image, cx, cy, rx, ry int, col color.Color) {
	bounds := img.Bounds()
	// Hatch pattern: diagonal lines every 4 pixels
	for y := cy - ry; y <= cy+ry; y++ {
		for x := cx - rx; x <= cx+rx; x++ {
			// Check if point is inside ellipse
			dx := float64(x - cx)
			dy := float64(y - cy)
			if (dx*dx)/(float64(rx*rx))+(dy*dy)/(float64(ry*ry)) <= 1.0 {
				// Apply hatch pattern (diagonal lines)
				if (x+y)%4 == 0 {
					if x >= 0 && x < bounds.Dx() && y >= 0 && y < bounds.Dy() {
						applyROP(img, x, y, col)
					}
				}
			}
		}
	}
}

// DrawRect draws a rectangle with optional fill modes
// Args: picID, x, y, width, height, fillMode
func DrawRect(args ...any) {
	if len(args) < 5 {
		fmt.Println("DrawRect: insufficient arguments")
		return
	}

	picID := args[0].(int)
	x := args[1].(int)
	y := args[2].(int)
	width := args[3].(int)
	height := args[4].(int)
	fillMode := FILL_NONE
	if len(args) > 5 {
		fillMode = args[5].(int)
	}

	fmt.Printf("DrawRect: pic=%d pos=(%d,%d) size=(%dx%d) fill=%d\n", picID, x, y, width, height, fillMode)

	if globalEngine != nil {
		globalEngine.renderMutex.Lock()
		defer globalEngine.renderMutex.Unlock()

		pic, ok := globalEngine.pictures[picID]
		if !ok {
			fmt.Printf("  Picture ID=%d not found\n", picID)
			return
		}

		drawRectangle(pic.Image, x, y, width, height, currentPaintColor, fillMode)
	} else {
		renderMutex.Lock()
		defer renderMutex.Unlock()

		pic, ok := pictures[picID]
		if !ok {
			fmt.Printf("  Picture ID=%d not found\n", picID)
			return
		}

		drawRectangle(pic.Image, x, y, width, height, currentPaintColor, fillMode)
	}
}

// drawRectangle draws a rectangle with the specified fill mode
func drawRectangle(img *ebiten.Image, x, y, width, height int, col color.Color, fillMode int) {
	if fillMode == FILL_SOLID {
		// Solid fill
		drawFilledRect(img, x, y, width, height, col)
	} else if fillMode == FILL_HATCH {
		// Hatch fill
		drawHatchedRect(img, x, y, width, height, col)
	} else {
		// Outline only
		drawRectOutline(img, x, y, width, height, col)
	}
}

// drawRectOutline draws the outline of a rectangle
func drawRectOutline(img *ebiten.Image, x, y, width, height int, col color.Color) {
	// Top edge
	drawLineBresenham(img, x, y, x+width-1, y, col, currentLineSize)
	// Bottom edge
	drawLineBresenham(img, x, y+height-1, x+width-1, y+height-1, col, currentLineSize)
	// Left edge
	drawLineBresenham(img, x, y, x, y+height-1, col, currentLineSize)
	// Right edge
	drawLineBresenham(img, x+width-1, y, x+width-1, y+height-1, col, currentLineSize)
}

// drawFilledRect draws a solid filled rectangle
func drawFilledRect(img *ebiten.Image, x, y, width, height int, col color.Color) {
	bounds := img.Bounds()
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if px >= 0 && px < bounds.Dx() && py >= 0 && py < bounds.Dy() {
				applyROP(img, px, py, col)
			}
		}
	}
}

// drawHatchedRect draws a hatch-filled rectangle
func drawHatchedRect(img *ebiten.Image, x, y, width, height int, col color.Color) {
	bounds := img.Bounds()
	// Hatch pattern: diagonal lines every 4 pixels
	for py := y; py < y+height; py++ {
		for px := x; px < x+width; px++ {
			if (px+py)%4 == 0 {
				if px >= 0 && px < bounds.Dx() && py >= 0 && py < bounds.Dy() {
					applyROP(img, px, py, col)
				}
			}
		}
	}
}

// ============================================================================
// Array Operations (Requirement 31)
// ============================================================================

// ArraySize returns the number of elements in an array
// Requirement 31.1: WHEN ArraySize is called, THE Runtime SHALL return the number of elements in the array
func ArraySize(arr []int) int {
	size := len(arr)
	if debugLevel >= 2 {
		fmt.Printf("ArraySize: %d elements\n", size)
	}
	return size
}

// DelArrayAll removes all elements from an array
// Requirement 31.2: WHEN DelArrayAll is called, THE Runtime SHALL remove all elements from the array
// Returns a new empty slice (Go slices are passed by value, so we return the cleared slice)
func DelArrayAll(arr []int) []int {
	if debugLevel >= 2 {
		fmt.Printf("DelArrayAll: clearing array with %d elements\n", len(arr))
	}
	return []int{}
}

// DelArrayAt removes the element at the specified index
// Requirement 31.3: WHEN DelArrayAt is called, THE Runtime SHALL remove the element at the specified index
// Returns a new slice with the element removed
func DelArrayAt(arr []int, index int) []int {
	if debugLevel >= 2 {
		fmt.Printf("DelArrayAt: removing element at index %d from array with %d elements\n", index, len(arr))
	}

	// Validate index
	if index < 0 || index >= len(arr) {
		if debugLevel >= 1 {
			fmt.Printf("  Warning: index %d out of bounds (array size: %d)\n", index, len(arr))
		}
		return arr
	}

	// Create new slice without the element at index
	// Requirement 31.5: THE Runtime SHALL automatically resize arrays as needed during insertion and deletion
	result := make([]int, 0, len(arr)-1)
	result = append(result, arr[:index]...)
	result = append(result, arr[index+1:]...)

	if debugLevel >= 2 {
		fmt.Printf("  Array size after deletion: %d\n", len(result))
	}

	return result
}

// InsArrayAt inserts an element at the specified index
// Requirement 31.4: WHEN InsArrayAt is called, THE Runtime SHALL insert an element at the specified index
// Returns a new slice with the element inserted
func InsArrayAt(arr []int, index int, value int) []int {
	if debugLevel >= 2 {
		fmt.Printf("InsArrayAt: inserting value %d at index %d into array with %d elements\n", value, index, len(arr))
	}

	// Validate index (allow index == len(arr) for append)
	if index < 0 || index > len(arr) {
		if debugLevel >= 1 {
			fmt.Printf("  Warning: index %d out of bounds (array size: %d)\n", index, len(arr))
		}
		return arr
	}

	// Create new slice with space for the new element
	// Requirement 31.5: THE Runtime SHALL automatically resize arrays as needed during insertion and deletion
	result := make([]int, len(arr)+1)

	// Copy elements before insertion point
	copy(result[:index], arr[:index])

	// Insert new element
	result[index] = value

	// Copy elements after insertion point
	copy(result[index+1:], arr[index:])

	if debugLevel >= 2 {
		fmt.Printf("  Array size after insertion: %d\n", len(result))
	}

	return result
}

// ============================================================================
// Integer Bit Operations
// ============================================================================

// MakeLong combines two 16-bit values into a single 32-bit value
// Requirement 32.1: WHEN MakeLong is called, THE Runtime SHALL combine two 16-bit values into a single 32-bit value
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
// lowWord is placed in the lower 16 bits, hiWord in the upper 16 bits
func MakeLong(lowWord, hiWord int) int {
	// Mask to 16 bits to ensure we only use the lower 16 bits of each input
	low := lowWord & 0xFFFF
	hi := hiWord & 0xFFFF

	// Combine: shift hi to upper 16 bits and OR with low
	result := (hi << 16) | low

	if debugLevel >= 2 {
		fmt.Printf("MakeLong: lowWord=0x%04X, hiWord=0x%04X -> result=0x%08X (%d)\n", low, hi, result, result)
	}

	return result
}

// GetHiWord extracts the upper 16 bits of a 32-bit value
// Requirement 32.2: WHEN GetHiWord is called, THE Runtime SHALL extract the upper 16 bits of a 32-bit value
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
func GetHiWord(value int) int {
	// Shift right 16 bits and mask to 16 bits
	result := (value >> 16) & 0xFFFF

	if debugLevel >= 2 {
		fmt.Printf("GetHiWord: value=0x%08X (%d) -> hiWord=0x%04X (%d)\n", value, value, result, result)
	}

	return result
}

// GetLowWord extracts the lower 16 bits of a 32-bit value
// Requirement 32.3: WHEN GetLowWord is called, THE Runtime SHALL extract the lower 16 bits of a 32-bit value
// Requirement 32.4: THE Runtime SHALL preserve bit patterns during pack and unpack operations
func GetLowWord(value int) int {
	// Mask to lower 16 bits
	result := value & 0xFFFF

	if debugLevel >= 2 {
		fmt.Printf("GetLowWord: value=0x%08X (%d) -> lowWord=0x%04X (%d)\n", value, value, result, result)
	}

	return result
}

// ============================================================================
// File Operations - INI Files
// ============================================================================

// WriteIniInt writes an integer value to an INI file
// Requirement 27.1: WHEN WriteIniInt is called, THE Runtime SHALL write an integer value to the specified INI section and entry
// Requirement 27.5: THE Runtime SHALL create INI files if they do not exist
func WriteIniInt(filename, section, entry string, value int) {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("WriteIniInt: file=%s, section=%s, entry=%s, value=%d\n", filename, section, entry, value)
	}

	// Load or create INI file
	cfg, err := ini.Load(filename)
	if err != nil {
		// File doesn't exist, create new
		cfg = ini.Empty()
	}

	// Set the value
	cfg.Section(section).Key(entry).SetValue(fmt.Sprintf("%d", value))

	// Save the file
	err = cfg.SaveTo(filename)
	if err != nil {
		fmt.Printf("ERROR: WriteIniInt failed to save file %s: %v\n", filename, err)
	}
}

// GetIniInt reads an integer value from an INI file
// Requirement 27.2: WHEN GetIniInt is called, THE Runtime SHALL read an integer value from the specified INI section and entry
func GetIniInt(filename, section, entry string, defaultValue int) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("GetIniInt: file=%s, section=%s, entry=%s, default=%d\n", filename, section, entry, defaultValue)
	}

	// Load INI file
	cfg, err := ini.Load(filename)
	if err != nil {
		if debugLevel >= 1 {
			fmt.Printf("GetIniInt: file %s not found, returning default %d\n", filename, defaultValue)
		}
		return defaultValue
	}

	// Get the value
	value, err := cfg.Section(section).Key(entry).Int()
	if err != nil {
		if debugLevel >= 1 {
			fmt.Printf("GetIniInt: key not found, returning default %d\n", defaultValue)
		}
		return defaultValue
	}

	if debugLevel >= 2 {
		fmt.Printf("GetIniInt: returning %d\n", value)
	}

	return value
}

// WriteIniStr writes a string value to an INI file
// Requirement 27.3: WHEN WriteIniStr is called, THE Runtime SHALL write a string value to the specified INI section and entry
// Requirement 27.5: THE Runtime SHALL create INI files if they do not exist
func WriteIniStr(filename, section, entry, value string) {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("WriteIniStr: file=%s, section=%s, entry=%s, value=%s\n", filename, section, entry, value)
	}

	// Load or create INI file
	cfg, err := ini.Load(filename)
	if err != nil {
		// File doesn't exist, create new
		cfg = ini.Empty()
	}

	// Set the value
	cfg.Section(section).Key(entry).SetValue(value)

	// Save the file
	err = cfg.SaveTo(filename)
	if err != nil {
		fmt.Printf("ERROR: WriteIniStr failed to save file %s: %v\n", filename, err)
	}
}

// GetIniStr reads a string value from an INI file
// Requirement 27.4: WHEN GetIniStr is called, THE Runtime SHALL read a string value from the specified INI section and entry
func GetIniStr(filename, section, entry, defaultValue string) string {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("GetIniStr: file=%s, section=%s, entry=%s, default=%s\n", filename, section, entry, defaultValue)
	}

	// Load INI file
	cfg, err := ini.Load(filename)
	if err != nil {
		if debugLevel >= 1 {
			fmt.Printf("GetIniStr: file %s not found, returning default %s\n", filename, defaultValue)
		}
		return defaultValue
	}

	// Get the value
	value := cfg.Section(section).Key(entry).String()
	if value == "" {
		if debugLevel >= 1 {
			fmt.Printf("GetIniStr: key not found, returning default %s\n", defaultValue)
		}
		return defaultValue
	}

	if debugLevel >= 2 {
		fmt.Printf("GetIniStr: returning %s\n", value)
	}

	return value
}

// ============================================================================
// File Operations - File Management
// ============================================================================

// CopyFile copies a file from source to destination
// Requirement 28.1: WHEN CopyFile is called, THE Runtime SHALL copy a file from source to destination path
func CopyFile(src, dst string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("CopyFile: src=%s, dst=%s\n", src, dst)
	}

	// Read source file
	data, err := os.ReadFile(src)
	if err != nil {
		fmt.Printf("ERROR: CopyFile failed to read source %s: %v\n", src, err)
		return -1
	}

	// Write to destination
	err = os.WriteFile(dst, data, 0644)
	if err != nil {
		fmt.Printf("ERROR: CopyFile failed to write destination %s: %v\n", dst, err)
		return -1
	}

	return 0
}

// DelFile deletes a file
// Requirement 28.2: WHEN DelFile is called, THE Runtime SHALL delete the specified file
func DelFile(filename string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("DelFile: filename=%s\n", filename)
	}

	err := os.Remove(filename)
	if err != nil {
		fmt.Printf("ERROR: DelFile failed to delete %s: %v\n", filename, err)
		return -1
	}

	return 0
}

// IsExist checks if a file exists
// Requirement 28.3: WHEN IsExist is called, THE Runtime SHALL return whether the specified file exists
func IsExist(filename string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("IsExist: filename=%s\n", filename)
	}

	_, err := os.Stat(filename)
	if err == nil {
		return 1 // File exists
	}
	if os.IsNotExist(err) {
		return 0 // File does not exist
	}
	// Other error (permission, etc.)
	return 0
}

// MkDir creates a directory
// Requirement 28.4: WHEN MkDir is called, THE Runtime SHALL create the specified directory
func MkDir(dirname string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("MkDir: dirname=%s\n", dirname)
	}

	err := os.MkdirAll(dirname, 0755)
	if err != nil {
		fmt.Printf("ERROR: MkDir failed to create %s: %v\n", dirname, err)
		return -1
	}

	return 0
}

// RmDir removes a directory
// Requirement 28.5: WHEN RmDir is called, THE Runtime SHALL remove the specified directory
func RmDir(dirname string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("RmDir: dirname=%s\n", dirname)
	}

	err := os.Remove(dirname)
	if err != nil {
		fmt.Printf("ERROR: RmDir failed to remove %s: %v\n", dirname, err)
		return -1
	}

	return 0
}

// ChDir changes the current working directory
// Requirement 28.6: WHEN ChDir is called, THE Runtime SHALL change the current working directory
func ChDir(dirname string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("ChDir: dirname=%s\n", dirname)
	}

	err := os.Chdir(dirname)
	if err != nil {
		fmt.Printf("ERROR: ChDir failed to change to %s: %v\n", dirname, err)
		return -1
	}

	return 0
}

// GetCWD returns the current working directory
// Requirement 28.7: WHEN GetCWD is called, THE Runtime SHALL return the current working directory path
func GetCWD() string {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	cwd, err := os.Getwd()
	if err != nil {
		fmt.Printf("ERROR: GetCWD failed: %v\n", err)
		return ""
	}

	if debugLevel >= 2 {
		fmt.Printf("GetCWD: returning %s\n", cwd)
	}

	return cwd
}

// ============================================================================
// File Operations - Binary I/O
// ============================================================================

// OpenF opens a file and returns a file handle
// Requirement 29.1: WHEN OpenF is called, THE Runtime SHALL open a file and return a file handle
// mode: 0=read, 1=write, 2=read/write
func OpenF(filename string, mode int) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("OpenF: filename=%s, mode=%d\n", filename, mode)
	}

	var flag int
	switch mode {
	case 0: // Read
		flag = os.O_RDONLY
	case 1: // Write
		flag = os.O_WRONLY | os.O_CREATE | os.O_TRUNC
	case 2: // Read/Write
		flag = os.O_RDWR | os.O_CREATE
	default:
		fmt.Printf("ERROR: OpenF invalid mode %d\n", mode)
		return -1
	}

	file, err := os.OpenFile(filename, flag, 0644)
	if err != nil {
		fmt.Printf("ERROR: OpenF failed to open %s: %v\n", filename, err)
		return -1
	}

	// Assign file handle
	handle := globalEngine.nextFileHandle
	globalEngine.nextFileHandle++
	globalEngine.openFiles[handle] = file

	if debugLevel >= 2 {
		fmt.Printf("OpenF: assigned handle %d\n", handle)
	}

	return handle
}

// CloseF closes a file handle
// Requirement 29.2: WHEN CloseF is called, THE Runtime SHALL close the specified file handle
func CloseF(handle int) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 1 {
		fmt.Printf("CloseF: handle=%d\n", handle)
	}

	file, ok := globalEngine.openFiles[handle]
	if !ok {
		fmt.Printf("ERROR: CloseF invalid handle %d\n", handle)
		return -1
	}

	err := file.Close()
	if err != nil {
		fmt.Printf("ERROR: CloseF failed to close handle %d: %v\n", handle, err)
		return -1
	}

	delete(globalEngine.openFiles, handle)
	return 0
}

// SeekF moves the file pointer to the specified position
// Requirement 29.3: WHEN SeekF is called, THE Runtime SHALL move the file pointer to the specified position
// whence: 0=from start, 1=from current, 2=from end
func SeekF(handle int, offset int, whence int) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("SeekF: handle=%d, offset=%d, whence=%d\n", handle, offset, whence)
	}

	file, ok := globalEngine.openFiles[handle]
	if !ok {
		fmt.Printf("ERROR: SeekF invalid handle %d\n", handle)
		return -1
	}

	newPos, err := file.Seek(int64(offset), whence)
	if err != nil {
		fmt.Printf("ERROR: SeekF failed: %v\n", err)
		return -1
	}

	return int(newPos)
}

// ReadF reads 1-4 bytes from a file and returns as an integer
// Requirement 29.4: WHEN ReadF is called, THE Runtime SHALL read 1-4 bytes and return as an integer
// size: number of bytes to read (1-4)
func ReadF(handle int, size int) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("ReadF: handle=%d, size=%d\n", handle, size)
	}

	if size < 1 || size > 4 {
		fmt.Printf("ERROR: ReadF invalid size %d (must be 1-4)\n", size)
		return -1
	}

	file, ok := globalEngine.openFiles[handle]
	if !ok {
		fmt.Printf("ERROR: ReadF invalid handle %d\n", handle)
		return -1
	}

	buf := make([]byte, size)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		fmt.Printf("ERROR: ReadF failed: %v\n", err)
		return -1
	}

	if n < size {
		if debugLevel >= 1 {
			fmt.Printf("ReadF: read %d bytes, expected %d\n", n, size)
		}
	}

	// Convert bytes to integer (little-endian)
	result := 0
	for i := 0; i < n; i++ {
		result |= int(buf[i]) << (i * 8)
	}

	if debugLevel >= 2 {
		fmt.Printf("ReadF: returning %d (0x%X)\n", result, result)
	}

	return result
}

// WriteF writes an integer value as 1-4 bytes to a file
// Requirement 29.5: WHEN WriteF is called, THE Runtime SHALL write an integer value as 1-4 bytes
// size: number of bytes to write (1-4)
func WriteF(handle int, value int, size int) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("WriteF: handle=%d, value=%d, size=%d\n", handle, value, size)
	}

	if size < 1 || size > 4 {
		fmt.Printf("ERROR: WriteF invalid size %d (must be 1-4)\n", size)
		return -1
	}

	file, ok := globalEngine.openFiles[handle]
	if !ok {
		fmt.Printf("ERROR: WriteF invalid handle %d\n", handle)
		return -1
	}

	// Convert integer to bytes (little-endian)
	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = byte((value >> (i * 8)) & 0xFF)
	}

	n, err := file.Write(buf)
	if err != nil {
		fmt.Printf("ERROR: WriteF failed: %v\n", err)
		return -1
	}

	if n < size {
		fmt.Printf("ERROR: WriteF wrote %d bytes, expected %d\n", n, size)
		return -1
	}

	return 0
}

// StrReadF reads a null-terminated string from a file
// Requirement 29.6: WHEN StrReadF is called, THE Runtime SHALL read a null-terminated string from the file
func StrReadF(handle int) string {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("StrReadF: handle=%d\n", handle)
	}

	file, ok := globalEngine.openFiles[handle]
	if !ok {
		fmt.Printf("ERROR: StrReadF invalid handle %d\n", handle)
		return ""
	}

	var result []byte
	buf := make([]byte, 1)

	for {
		n, err := file.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			fmt.Printf("ERROR: StrReadF failed: %v\n", err)
			return ""
		}

		if n == 0 {
			break
		}

		if buf[0] == 0 {
			// Null terminator found
			break
		}

		result = append(result, buf[0])
	}

	str := string(result)
	if debugLevel >= 2 {
		fmt.Printf("StrReadF: returning '%s'\n", str)
	}

	return str
}

// StrWriteF writes a null-terminated string to a file
// Requirement 29.7: WHEN StrWriteF is called, THE Runtime SHALL write a null-terminated string to the file
func StrWriteF(handle int, str string) int {
	renderMutex.Lock()
	defer renderMutex.Unlock()

	if debugLevel >= 2 {
		fmt.Printf("StrWriteF: handle=%d, str='%s'\n", handle, str)
	}

	file, ok := globalEngine.openFiles[handle]
	if !ok {
		fmt.Printf("ERROR: StrWriteF invalid handle %d\n", handle)
		return -1
	}

	// Write string bytes
	_, err := file.WriteString(str)
	if err != nil {
		fmt.Printf("ERROR: StrWriteF failed to write string: %v\n", err)
		return -1
	}

	// Write null terminator
	_, err = file.Write([]byte{0})
	if err != nil {
		fmt.Printf("ERROR: StrWriteF failed to write null terminator: %v\n", err)
		return -1
	}

	return 0
}
