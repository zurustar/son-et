package main

import (
	"embed"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/zurustar/son-et/pkg/compiler/codegen"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
	"github.com/zurustar/son-et/pkg/compiler/preprocessor"
	"github.com/zurustar/son-et/pkg/engine"
)

const (
	targetFPS = 60
)

// Embedded project support
// These variables are set by generated code when building with embedded project
// See scripts/build-embedded.sh for the build process
// When building without embedded project, these remain as zero values
var (
	embeddedFS      embed.FS
	embeddedProject string
)

var (
	headlessFlag = flag.Bool("headless", false, "Run in headless mode (no GUI)")
	timeoutFlag  = flag.String("timeout", "", "Execution timeout (e.g., 5s, 500ms, 2m)")
	debugFlag    = flag.Int("debug", 0, "Debug level (0=errors, 1=info, 2=debug)")
)

// Game implements ebiten.Game interface
type Game struct {
	engine   *engine.Engine
	renderer *engine.EbitenRenderer
}

func (g *Game) Update() error {
	err := g.engine.Update()
	if err != nil {
		// If engine is terminated, exit the game loop
		if err == engine.ErrTerminated || g.engine.IsTerminated() {
			return err // This will cause ebiten.RunGame to exit
		}
		// For other errors, log but continue
		log.Printf("Update error: %v", err)
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Use the renderer to draw the current frame
	g.renderer.RenderFrame(screen, g.engine.GetState())
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Return virtual desktop size
	return engine.VirtualDesktopWidth, engine.VirtualDesktopHeight
}

func main() {
	Run()
}

// reorderArgs reorders command-line arguments so that flags come before positional arguments.
// This allows flags to appear in any order relative to positional arguments.
// Example: "file.tfy --headless --timeout 5s" becomes "--headless --timeout 5s file.tfy"
func reorderArgs(args []string) []string {
	var flags []string
	var positional []string

	i := 0
	for i < len(args) {
		arg := args[i]
		if strings.HasPrefix(arg, "-") {
			// This is a flag
			flags = append(flags, arg)
			// Check if this flag takes a value (not a boolean flag)
			if arg == "--timeout" || arg == "-timeout" || arg == "--debug" || arg == "-debug" {
				// This flag takes a value, include the next argument
				if i+1 < len(args) {
					i++
					flags = append(flags, args[i])
				}
			}
		} else {
			// This is a positional argument
			positional = append(positional, arg)
		}
		i++
	}

	// Return flags first, then positional arguments
	return append(flags, positional...)
}

// Run is the main entry point, separated for testing and embedding.
func Run() {
	// Reorder arguments so flags come before positional arguments
	// This allows: ./son-et file.tfy --headless --timeout 5s
	reorderedArgs := reorderArgs(os.Args[1:])
	os.Args = append([]string{os.Args[0]}, reorderedArgs...)

	flag.Parse()

	log.Printf("Parsed flags: headless=%v, timeout=%s, debug=%d", *headlessFlag, *timeoutFlag, *debugFlag)
	log.Printf("Args: %v", flag.Args())

	// Force OpenGL backend (macOS 15.0 Metal compatibility issue)
	// This also prevents screen switching in headless mode
	os.Setenv("EBITEN_GRAPHICS_LIBRARY", "opengl")

	// Check for HEADLESS environment variable
	if os.Getenv("HEADLESS") == "1" {
		*headlessFlag = true
	}

	// Check if embeddedProject was set at build time
	if embeddedProject != "" {
		log.Printf("Running embedded project: %s", embeddedProject)
		runEmbeddedModeFromFS()
		return
	}

	// Check if running in embedded mode (no command-line arguments but has registered projects)
	args := flag.Args()
	if len(args) < 1 {
		// No embedded projects and no arguments - show usage
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <project_directory_or_tfy_file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	// Direct mode - load from filesystem
	projectPath := args[0]
	runDirectMode(projectPath)
}

func findMainTFY(projectDir string) (string, error) {
	var tfyFiles []string

	err := filepath.Walk(projectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := filepath.Ext(path)
			if ext == ".tfy" || ext == ".TFY" {
				tfyFiles = append(tfyFiles, path)
			}
		}
		return nil
	})

	if err != nil {
		return "", err
	}

	if len(tfyFiles) == 0 {
		return "", fmt.Errorf("no TFY files found in %s", projectDir)
	}

	// For now, just use the first TFY file
	// TODO: Search for main() function
	return tfyFiles[0], nil
}

// findEmbeddedTFY finds the TFY file in the embedded filesystem
func findEmbeddedTFY() (string, error) {
	// Files are embedded under embeddedProject directory (e.g., ".embed")
	// AssetLoader will handle the baseDir, so we just need the filename
	entries, err := embeddedFS.ReadDir(embeddedProject)
	if err != nil {
		return "", fmt.Errorf("failed to read embedded directory %s: %w", embeddedProject, err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := filepath.Ext(entry.Name())
		if ext == ".tfy" || ext == ".TFY" {
			// Return just the filename, AssetLoader will add baseDir
			return entry.Name(), nil
		}
	}

	return "", fmt.Errorf("no TFY files found in embedded filesystem")
}

// autoLoadSoundFont searches for and loads a SoundFont file from the project directory
func autoLoadSoundFont(eng *engine.Engine, projectDir string) error {
	// Get absolute path for project directory
	absProjectDir, err := filepath.Abs(projectDir)
	if err != nil {
		log.Printf("Warning: failed to get absolute path for %s: %v", projectDir, err)
		absProjectDir = projectDir
	}

	// Search locations: project directory, parent directory, and grandparent directory (repository root)
	// This handles cases like: repo/samples/kuma2 -> repo/samples -> repo
	parentDir := filepath.Dir(absProjectDir)
	grandparentDir := filepath.Dir(parentDir)
	searchLocations := []string{absProjectDir, parentDir, grandparentDir}

	// Search for any .sf2 file in all locations
	for _, dir := range searchLocations {
		var sf2Files []string

		// List files in directory
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue // Skip if directory cannot be read
		}

		// Find .sf2 files
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			ext := filepath.Ext(entry.Name())
			if ext == ".sf2" || ext == ".SF2" {
				sf2Files = append(sf2Files, entry.Name())
			}
		}

		// If found, use the first .sf2 file
		if len(sf2Files) > 0 {
			name := sf2Files[0]
			fullPath := filepath.Join(dir, name)
			log.Printf("Auto-loading SoundFont: %s", fullPath)

			// Use absolute path for loading
			if err := eng.LoadSoundFont(fullPath); err != nil {
				return fmt.Errorf("failed to load SoundFont %s: %w", fullPath, err)
			}
			log.Printf("Successfully loaded SoundFont: %s", fullPath)
			return nil
		}
	}

	// No SoundFont found - this is not an error, just a warning
	return fmt.Errorf("no SoundFont file (*.sf2) found in project directory or parent directories (MIDI playback will not work)")
}

func loadAndExecute(eng *engine.Engine, tfyFile string, assetLoader engine.AssetLoader) error {
	// Get base directory and filename
	baseDir := filepath.Dir(tfyFile)
	filename := filepath.Base(tfyFile)

	// Preprocess (handle #include, #info, encoding)
	prep := preprocessor.NewPreprocessor(baseDir, assetLoader)
	preprocessed, err := prep.Process(filename)
	if err != nil {
		return fmt.Errorf("preprocessing failed: %w", err)
	}

	// Lex
	lex := lexer.New(preprocessed)

	// Parse
	p := parser.New(lex)
	program := p.ParseProgram()
	if len(p.Errors()) > 0 {
		for _, msg := range p.Errors() {
			log.Printf("Parse error: %s", msg)
		}
		return fmt.Errorf("parsing failed with %d errors", len(p.Errors()))
	}

	log.Printf("Successfully parsed %d statements", len(program.Statements))

	// Generate OpCode
	gen := codegen.New()
	opcodes := gen.Generate(program)
	if len(gen.Errors()) > 0 {
		for _, msg := range gen.Errors() {
			log.Printf("Codegen error: %s", msg)
		}
		return fmt.Errorf("code generation failed with %d errors", len(gen.Errors()))
	}

	log.Printf("Generated %d opcodes", len(opcodes))

	// Execute top-level opcodes synchronously to register function definitions
	// This is done before starting the engine loop
	if err := eng.ExecuteTopLevel(opcodes); err != nil {
		return fmt.Errorf("failed to execute top-level code: %w", err)
	}

	// Now call main() function if it exists
	if err := eng.CallMainFunction(); err != nil {
		return fmt.Errorf("failed to call main(): %w", err)
	}

	return nil
}

func runHeadless(eng *engine.Engine) {
	log.Println("Starting headless execution...")

	ticker := time.NewTicker(time.Second / targetFPS)
	defer ticker.Stop()

	tickCount := 0
	for {
		<-ticker.C
		tickCount++

		if tickCount%60 == 0 {
			log.Printf("Headless tick: %d (%.1fs elapsed)", tickCount, float64(tickCount)/60.0)
		}

		// Check termination BEFORE update
		if eng.IsTerminated() {
			log.Println("Engine terminated (pre-check)")
			break
		}

		if err := eng.Update(); err != nil {
			if err == engine.ErrTerminated {
				log.Println("Engine terminated normally")
			} else {
				log.Printf("Update error: %v", err)
			}
			break
		}

		// Check termination AFTER update
		if eng.IsTerminated() {
			log.Println("Engine terminated (post-check)")
			break
		}
	}

	eng.Shutdown()
}

func runGUI(eng *engine.Engine, renderer *engine.EbitenRenderer) {
	ebiten.SetWindowSize(engine.VirtualDesktopWidth, engine.VirtualDesktopHeight)
	ebiten.SetWindowTitle("son-et - FILLY Script Interpreter")
	ebiten.SetTPS(targetFPS)

	game := &Game{
		engine:   eng,
		renderer: renderer,
	}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	eng.Shutdown()
}

// runEmbeddedModeFromFS runs a project from the embedded filesystem (embeddedFS).
// This is used when embeddedProject is set at build time.
func runEmbeddedModeFromFS() {
	log.Printf("Running embedded project from FS: %s", embeddedProject)

	// Create asset loader (embedded mode using embeddedFS with baseDir)
	assetLoader := engine.NewEmbedFSAssetLoaderWithBaseDir(embeddedFS, embeddedProject)

	// Create image decoder
	imageDecoder := engine.NewBMPImageDecoder()

	// Create renderer (only for GUI mode)
	var renderer engine.Renderer
	if !*headlessFlag {
		renderer = engine.NewEbitenRenderer()
	}

	// Create engine
	eng := engine.NewEngine(renderer, assetLoader, imageDecoder)

	// Set debug level
	eng.SetDebugLevel(engine.DebugLevel(*debugFlag))

	// Set logger on renderer if in GUI mode
	if !*headlessFlag && renderer != nil {
		if ebitenRenderer, ok := renderer.(*engine.EbitenRenderer); ok {
			ebitenRenderer.SetLogger(eng.GetLogger())
		}
	}

	// Set headless mode
	if *headlessFlag {
		eng.SetHeadless(true)
		log.Println("Running in headless mode")
	} else {
		log.Println("Running in GUI mode")
	}

	// Set timeout
	if *timeoutFlag != "" {
		timeout, err := time.ParseDuration(*timeoutFlag)
		if err != nil {
			log.Fatalf("Invalid timeout format: %v", err)
		}
		eng.SetTimeout(timeout)
	}

	// Find TFY file in embedded FS
	tfyFile, err := findEmbeddedTFY()
	if err != nil {
		log.Fatalf("Failed to find TFY file in embedded FS: %v", err)
	}

	log.Printf("Loading embedded TFY file: %s", tfyFile)

	// Load and execute TFY file from embedded FS
	if err := loadAndExecute(eng, tfyFile, assetLoader); err != nil {
		log.Fatalf("Failed to execute embedded script: %v", err)
	}

	log.Println("Embedded script loaded successfully")

	// Start engine
	eng.Start()

	log.Println("Engine started")

	// Run game loop
	if *headlessFlag {
		log.Println("Entering headless mode")
		runHeadless(eng)
	} else {
		// Cast renderer to EbitenRenderer for GUI mode
		ebitenRenderer, ok := renderer.(*engine.EbitenRenderer)
		if !ok {
			log.Fatal("Expected EbitenRenderer for GUI mode")
		}
		runGUI(eng, ebitenRenderer)
	}
}

// runDirectMode runs a project from the filesystem.
func runDirectMode(projectPath string) {
	// Check if projectPath is a file or directory
	info, err := os.Stat(projectPath)
	if os.IsNotExist(err) {
		log.Fatalf("Path does not exist: %s", projectPath)
	}

	var projectDir string
	var tfyFile string

	if info.IsDir() {
		// It's a directory - find TFY file
		projectDir = projectPath
		tfyFile, err = findMainTFY(projectDir)
		if err != nil {
			log.Fatalf("Failed to find main TFY file: %v", err)
		}
	} else {
		// It's a file - use its directory as project directory
		tfyFile = projectPath
		projectDir = filepath.Dir(projectPath)
	}

	log.Printf("Loading project: %s", tfyFile)

	// Create asset loader (filesystem mode)
	assetLoader := engine.NewFilesystemAssetLoader(projectDir)

	// Create image decoder
	imageDecoder := engine.NewBMPImageDecoder()

	// Create renderer (only for GUI mode)
	var renderer engine.Renderer
	if !*headlessFlag {
		renderer = engine.NewEbitenRenderer()
	}

	// Create engine
	eng := engine.NewEngine(renderer, assetLoader, imageDecoder)

	// Set debug level
	eng.SetDebugLevel(engine.DebugLevel(*debugFlag))

	// Set logger on renderer if in GUI mode
	if !*headlessFlag && renderer != nil {
		if ebitenRenderer, ok := renderer.(*engine.EbitenRenderer); ok {
			ebitenRenderer.SetLogger(eng.GetLogger())
		}
	}

	// Set headless mode
	if *headlessFlag {
		eng.SetHeadless(true)
		log.Println("Running in headless mode")
	} else {
		log.Println("Running in GUI mode")
	}

	// Set timeout
	if *timeoutFlag != "" {
		timeout, err := time.ParseDuration(*timeoutFlag)
		if err != nil {
			log.Fatalf("Invalid timeout format: %v", err)
		}
		eng.SetTimeout(timeout)
	}

	// Auto-load SoundFont if available
	if err := autoLoadSoundFont(eng, projectDir); err != nil {
		log.Printf("Warning: %v", err)
	}

	// Load and parse TFY file
	if err := loadAndExecute(eng, tfyFile, assetLoader); err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	log.Println("Script loaded successfully")

	// Start engine
	eng.Start()

	log.Println("Engine started")

	// Run game loop
	if *headlessFlag {
		log.Println("Entering headless mode")
		runHeadless(eng)
	} else {
		// Cast renderer to EbitenRenderer for GUI mode
		ebitenRenderer, ok := renderer.(*engine.EbitenRenderer)
		if !ok {
			log.Fatal("Expected EbitenRenderer for GUI mode")
		}
		runGUI(eng, ebitenRenderer)
	}
}
