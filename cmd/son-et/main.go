package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

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

	// Create image decoder
	imageDecoder := engine.NewBMPImageDecoder()

	// Create engine
	eng := engine.NewEngine(nil, assetLoader, imageDecoder)

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
		log.Fatal("GUI mode not yet implemented - please use --headless flag")
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
