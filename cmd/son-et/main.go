package main

import (
	"flag"
	"fmt"
	"image/color"
	"log"
	"os"
	"path/filepath"
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

var (
	headlessFlag = flag.Bool("headless", false, "Run in headless mode (no GUI)")
	timeoutFlag  = flag.String("timeout", "", "Execution timeout (e.g., 5s, 500ms, 2m)")
	debugFlag    = flag.Int("debug", 0, "Debug level (0=errors, 1=info, 2=debug)")
)

// Game implements ebiten.Game interface
type Game struct {
	engine *engine.Engine
}

func (g *Game) Update() error {
	return g.engine.Update()
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Fill virtual desktop with background color (teal)
	screen.Fill(color.RGBA{0x1F, 0x7E, 0x7F, 0xff})

	// TODO: Actual rendering will be implemented in Phase 4
	g.engine.Render()
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	// Return virtual desktop size
	return engine.VirtualDesktopWidth, engine.VirtualDesktopHeight
}

func main() {
	flag.Parse()

	// Check for HEADLESS environment variable
	if os.Getenv("HEADLESS") == "1" {
		*headlessFlag = true
	}

	args := flag.Args()
	if len(args) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <project_directory>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	projectDir := args[0]

	// Verify project directory exists
	if _, err := os.Stat(projectDir); os.IsNotExist(err) {
		log.Fatalf("Project directory does not exist: %s", projectDir)
	}

	// Find TFY file with main function
	tfyFile, err := findMainTFY(projectDir)
	if err != nil {
		log.Fatalf("Failed to find main TFY file: %v", err)
	}

	log.Printf("Loading project: %s", tfyFile)

	// Create asset loader (filesystem mode for now)
	assetLoader := engine.NewFilesystemAssetLoader(projectDir)

	// Create engine
	eng := engine.NewEngine(nil, assetLoader, nil)

	// Set debug level
	eng.SetDebugLevel(engine.DebugLevel(*debugFlag))

	// Set headless mode
	if *headlessFlag {
		eng.SetHeadless(true)
		log.Println("Running in headless mode")
	}

	// Set timeout
	if *timeoutFlag != "" {
		timeout, err := time.ParseDuration(*timeoutFlag)
		if err != nil {
			log.Fatalf("Invalid timeout format: %v", err)
		}
		eng.SetTimeout(timeout)
	}

	// Load and parse TFY file
	if err := loadAndExecute(eng, tfyFile, assetLoader); err != nil {
		log.Fatalf("Failed to execute script: %v", err)
	}

	// Start engine
	eng.Start()

	// Run game loop
	if *headlessFlag {
		runHeadless(eng)
	} else {
		runGUI(eng)
	}
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

	// Create main sequencer (TIME mode, no parent)
	mainSeq := engine.NewSequencer(opcodes, engine.TIME, nil)

	// Register main sequence with engine (group ID 0 = allocate new group)
	seqID := eng.RegisterSequence(mainSeq, 0)

	log.Printf("Registered main sequence %d with %d opcodes", seqID, len(opcodes))

	return nil
}

func runHeadless(eng *engine.Engine) {
	log.Println("Starting headless execution...")

	ticker := time.NewTicker(time.Second / targetFPS)
	defer ticker.Stop()

	for {
		<-ticker.C

		if err := eng.Update(); err != nil {
			log.Printf("Update error: %v", err)
			break
		}

		if eng.IsTerminated() {
			log.Println("Engine terminated")
			break
		}
	}

	eng.Shutdown()
}

func runGUI(eng *engine.Engine) {
	ebiten.SetWindowSize(engine.VirtualDesktopWidth, engine.VirtualDesktopHeight)
	ebiten.SetWindowTitle("son-et - FILLY Script Interpreter")
	ebiten.SetTPS(targetFPS)

	game := &Game{engine: eng}

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}

	eng.Shutdown()
}
