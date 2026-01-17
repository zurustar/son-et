package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
	"github.com/zurustar/son-et/pkg/engine"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func main() {
	// Parse command-line flags
	helpFlag := flag.Bool("help", false, "Display usage information")
	flag.Parse()

	// Display help if requested or no arguments provided
	if *helpFlag || flag.NArg() == 0 {
		displayHelp()
		os.Exit(0)
	}

	// Get directory argument
	directory := flag.Arg(0)

	// Validate directory exists
	info, err := os.Stat(directory)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Directory '%s' does not exist\n", directory)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "Error: Cannot access directory '%s': %v\n", directory, err)
		os.Exit(1)
	}

	if !info.IsDir() {
		fmt.Fprintf(os.Stderr, "Error: '%s' is not a directory\n", directory)
		os.Exit(1)
	}

	// Execute in direct mode
	if err := executeDirect(directory); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// displayHelp shows usage information
func displayHelp() {
	fmt.Println("son-et - FILLY Script Interpreter")
	fmt.Println()
	fmt.Println("USAGE:")
	fmt.Println("  son-et <directory>    Execute TFY project in the specified directory")
	fmt.Println("  son-et --help         Display this help message")
	fmt.Println()
	fmt.Println("DESCRIPTION:")
	fmt.Println("  son-et executes FILLY language scripts (.tfy files) directly from a")
	fmt.Println("  project directory. The interpreter will locate TFY files, convert them")
	fmt.Println("  to OpCode at runtime, and execute the project immediately.")
	fmt.Println()
	fmt.Println("EXAMPLES:")
	fmt.Println("  son-et samples/my_project   Run a sample project")
	fmt.Println("  son-et my_game              Run project in my_game directory")
	fmt.Println()
	fmt.Println("ENVIRONMENT:")
	fmt.Println("  DEBUG_LEVEL=0    Show only errors")
	fmt.Println("  DEBUG_LEVEL=1    Show important operations (default)")
	fmt.Println("  DEBUG_LEVEL=2    Show all debug information")
	fmt.Println()
}

// executeDirect executes a TFY project from a directory (direct mode)
func executeDirect(directory string) error {
	// Convert to absolute path for consistent behavior
	absDir, err := filepath.Abs(directory)
	if err != nil {
		return fmt.Errorf("failed to resolve directory path: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Executing project in: %s\n", absDir)

	// Step 1: Locate TFY files in directory
	tfyFiles, err := findTFYFiles(absDir)
	if err != nil {
		return fmt.Errorf("failed to find TFY files: %w", err)
	}

	if len(tfyFiles) == 0 {
		return fmt.Errorf("no TFY files found in directory: %s", absDir)
	}

	fmt.Fprintf(os.Stderr, "Found %d TFY file(s)\n", len(tfyFiles))

	// Step 2: Read and parse TFY files (with #include support)
	// Use the first TFY file as entry point (or look for main.tfy)
	entryFile := tfyFiles[0]
	for _, f := range tfyFiles {
		if strings.ToLower(filepath.Base(f)) == "main.tfy" {
			entryFile = f
			break
		}
	}

	fmt.Fprintf(os.Stderr, "Entry file: %s\n", filepath.Base(entryFile))

	// Read and merge files (handling #include directives)
	included := make(map[string]bool)
	fullCode, err := recursiveRead(entryFile, included)
	if err != nil {
		return fmt.Errorf("failed to read TFY files: %w", err)
	}

	// Step 3: Parse the code
	l := lexer.New(string(fullCode))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		// Format parsing errors with helpful context
		fmt.Fprintf(os.Stderr, "\n=== PARSING ERRORS ===\n")
		fmt.Fprintf(os.Stderr, "File: %s\n\n", entryFile)
		for i, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Error %d: %s\n", i+1, msg)
		}
		fmt.Fprintf(os.Stderr, "\n")
		return &ParseError{
			File:   entryFile,
			Errors: p.Errors(),
		}
	}

	fmt.Fprintf(os.Stderr, "Parsing successful\n")

	// Step 4: Convert to OpCode using interpreter
	interp := interpreter.NewInterpreter()
	script, err := interp.Interpret(program)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n=== INTERPRETATION ERROR ===\n")
		fmt.Fprintf(os.Stderr, "Failed to convert TFY script to OpCode:\n")
		fmt.Fprintf(os.Stderr, "%v\n\n", err)
		return fmt.Errorf("failed to interpret script: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Interpreted script: %d globals, %d functions, %d assets\n",
		len(script.Globals), len(script.Functions), len(script.Assets))

	// Step 5: Initialize engine with FilesystemAssetLoader
	assetLoader := engine.NewFilesystemAssetLoader(absDir)
	imageDecoder := engine.NewBMPImageDecoder()

	// Initialize the engine using the Init-like pattern
	// We need to set up the global engine and game state
	engine.InitDirect(assetLoader, imageDecoder, func() {
		// This function will be called to execute the script
		// Wrap in error handler to catch runtime errors
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "\n=== RUNTIME ERROR ===\n")
				fmt.Fprintf(os.Stderr, "The program encountered a runtime error:\n")
				fmt.Fprintf(os.Stderr, "%v\n\n", r)
				fmt.Fprintf(os.Stderr, "This may be due to:\n")
				fmt.Fprintf(os.Stderr, "  - Invalid function calls or arguments\n")
				fmt.Fprintf(os.Stderr, "  - Missing or corrupted assets\n")
				fmt.Fprintf(os.Stderr, "  - Unsupported operations\n\n")
			}
		}()

		// Convert interpreter OpCode to engine OpCode format
		engineOps := convertToEngineOpCodes(script.Main.Body)

		// Register the main sequence
		// Use TIME mode (0) for normal execution
		engine.RegisterSequence(engine.Time, engineOps)
	})

	// Step 6: Start the engine (this will block until the game exits)
	fmt.Fprintf(os.Stderr, "Starting engine...\n")
	engine.Run()

	return nil
}

// findTFYFiles locates all .tfy files in the given directory
func findTFYFiles(directory string) ([]string, error) {
	var tfyFiles []string

	entries, err := os.ReadDir(directory)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		// Look for .tfy or .fil files
		if ext == ".tfy" || ext == ".fil" {
			tfyFiles = append(tfyFiles, filepath.Join(directory, name))
		}
	}

	return tfyFiles, nil
}

// recursiveRead reads a file and its includes recursively
func recursiveRead(path string, included map[string]bool) ([]byte, error) {
	if included[path] {
		return []byte{}, nil // Already included
	}
	included[path] = true

	// Read file
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s: %w", path, err)
	}

	// Decode Shift_JIS -> UTF-8
	reader := transform.NewReader(bytes.NewReader(raw), japanese.ShiftJIS.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode %s: %w", path, err)
	}

	// Scan lines for #include
	var output bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(decoded))

	// Get directory of current file for relative includes
	dir := filepath.Dir(path)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#include") {
			// Parse: #include "FILENAME"
			start := strings.Index(trimmed, "\"")
			end := strings.LastIndex(trimmed, "\"")
			if start != -1 && end != -1 && end > start {
				target := trimmed[start+1 : end]
				targetPath := filepath.Join(dir, target)

				// Recursively read
				inContent, err := recursiveRead(targetPath, included)
				if err != nil {
					return nil, fmt.Errorf("error including %s from %s: %w", target, path, err)
				}
				output.Write(inContent)
				output.WriteString("\n")
			} else {
				// Malformed include, just keep line
				output.WriteString(line)
				output.WriteString("\n")
			}
		} else {
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	return output.Bytes(), nil
}

// convertToEngineOpCodes converts interpreter OpCodes to engine OpCodes
func convertToEngineOpCodes(interpOps []interpreter.OpCode) []engine.OpCode {
	engineOps := make([]engine.OpCode, 0, len(interpOps))
	for _, op := range interpOps {
		// Skip VarRef at top level - these should only appear as nested expressions
		if op.Cmd == interpreter.OpVarRef {
			continue
		}
		engineOps = append(engineOps, engine.OpCode{
			Cmd:  op.Cmd,
			Args: convertOpCodeArgs(op.Args),
		})
	}
	return engineOps
}

// convertOpCodeArgs recursively converts OpCode arguments
func convertOpCodeArgs(args []any) []any {
	result := make([]any, len(args))
	for i, arg := range args {
		switch v := arg.(type) {
		case interpreter.OpCode:
			// Special case: VarRef should be unwrapped to just the variable name
			if v.Cmd == interpreter.OpVarRef && len(v.Args) > 0 {
				if varName, ok := v.Args[0].(interpreter.Variable); ok {
					result[i] = engine.Variable(varName)
					continue
				}
			}
			// Nested OpCode - convert recursively
			result[i] = engine.OpCode{
				Cmd:  v.Cmd,
				Args: convertOpCodeArgs(v.Args),
			}
		case []interpreter.OpCode:
			// Slice of OpCodes - convert each
			result[i] = convertToEngineOpCodes(v)
		case interpreter.Variable:
			// Variable reference - convert to engine.Variable
			result[i] = engine.Variable(v)
		default:
			// Primitive value - keep as-is
			result[i] = arg
		}
	}
	return result
}

// ParseError represents a parsing error with file context
type ParseError struct {
	File   string
	Errors []string
}

func (e *ParseError) Error() string {
	var buf strings.Builder
	buf.WriteString(fmt.Sprintf("Parse errors in %s:\n", e.File))
	for _, msg := range e.Errors {
		buf.WriteString(fmt.Sprintf("  %s\n", msg))
	}
	return buf.String()
}

// AssetLoadError represents an asset loading error
type AssetLoadError struct {
	Filename string
	Err      error
}

func (e *AssetLoadError) Error() string {
	return fmt.Sprintf("Failed to load asset '%s': %v", e.Filename, e.Err)
}

// RuntimeError represents a runtime error with TFY line context
type RuntimeError struct {
	Message string
	Line    int
	File    string
}

func (e *RuntimeError) Error() string {
	if e.File != "" && e.Line > 0 {
		return fmt.Sprintf("Runtime error at %s:%d: %s", e.File, e.Line, e.Message)
	} else if e.Line > 0 {
		return fmt.Sprintf("Runtime error at line %d: %s", e.Line, e.Message)
	}
	return fmt.Sprintf("Runtime error: %s", e.Message)
}
