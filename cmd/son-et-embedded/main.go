// Generic embedded build for FILLY projects
// Build with: go build -o <output> ./cmd/son-et-embedded
// Set PROJECT_DIR at build time using -ldflags

package main

import (
	"bufio"
	"bytes"
	"embed"
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

// These variables are set by embed_generated.go at build time
var (
	projectDir  string
	projectName string
	embeddedFS  embed.FS
)

func main() {
	if projectDir == "" {
		fmt.Fprintf(os.Stderr, "Error: This executable was not built correctly.\n")
		fmt.Fprintf(os.Stderr, "PROJECT_DIR must be set at build time.\n")
		fmt.Fprintf(os.Stderr, "\nPlease use the build_embedded.sh script to create embedded executables.\n")
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Running embedded project: %s\n", projectName)

	if err := executeEmbedded(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// executeEmbedded executes the embedded TFY project
func executeEmbedded() error {
	fmt.Fprintf(os.Stderr, "Executing embedded project: %s\n", projectName)

	// Step 1: Locate TFY files in embedded filesystem
	tfyFiles, err := findEmbeddedTFYFiles()
	if err != nil {
		return fmt.Errorf("failed to find embedded TFY files: %w", err)
	}

	if len(tfyFiles) == 0 {
		return fmt.Errorf("no TFY files found in embedded project: %s", projectDir)
	}

	fmt.Fprintf(os.Stderr, "Found %d embedded TFY file(s)\n", len(tfyFiles))

	// Step 2: Read and parse TFY files (with #include support)
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
	fullCode, err := recursiveReadEmbedded(entryFile, included)
	if err != nil {
		return fmt.Errorf("failed to read embedded TFY files: %w", err)
	}

	// Step 3: Parse the code
	l := lexer.New(string(fullCode))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Fprintf(os.Stderr, "\n=== PARSING ERRORS ===\n")
		fmt.Fprintf(os.Stderr, "File: %s\n\n", entryFile)
		for i, msg := range p.Errors() {
			fmt.Fprintf(os.Stderr, "Error %d: %s\n", i+1, msg)
		}
		fmt.Fprintf(os.Stderr, "\n")
		return fmt.Errorf("parse errors in %s", entryFile)
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

	// Step 5: Initialize engine with EmbeddedAssetLoader
	assetLoader := engine.NewEmbedFSAssetLoader(embeddedFS)
	imageDecoder := engine.NewBMPImageDecoder()

	// Initialize the engine
	engine.InitDirect(assetLoader, imageDecoder, func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "\n=== RUNTIME ERROR ===\n")
				fmt.Fprintf(os.Stderr, "The program encountered a runtime error:\n")
				fmt.Fprintf(os.Stderr, "%v\n\n", r)
				fmt.Fprintf(os.Stderr, "This may be due to:\n")
				fmt.Fprintf(os.Stderr, "  - Invalid function calls or arguments\n")
				fmt.Fprintf(os.Stderr, "  - Missing or corrupted embedded assets\n")
				fmt.Fprintf(os.Stderr, "  - Unsupported operations\n\n")
			}
		}()

		// Convert interpreter OpCode to engine OpCode format
		engineOps := convertToEngineOpCodes(script.Main.Body)

		// Register the main sequence
		engine.RegisterSequence(engine.Time, engineOps)
	})

	// Step 6: Start the engine
	fmt.Fprintf(os.Stderr, "Starting engine...\n")
	engine.Run()

	return nil
}

// findEmbeddedTFYFiles locates all .tfy files in the embedded filesystem
func findEmbeddedTFYFiles() ([]string, error) {
	var tfyFiles []string

	entries, err := embeddedFS.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded directory: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		ext := strings.ToLower(filepath.Ext(name))

		if ext == ".tfy" || ext == ".fil" {
			tfyFiles = append(tfyFiles, name)
		}
	}

	return tfyFiles, nil
}

// recursiveReadEmbedded reads a file and its includes recursively from embedded FS
func recursiveReadEmbedded(path string, included map[string]bool) ([]byte, error) {
	if included[path] {
		return []byte{}, nil
	}
	included[path] = true

	raw, err := embeddedFS.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read embedded file %s: %w", path, err)
	}

	reader := transform.NewReader(bytes.NewReader(raw), japanese.ShiftJIS.NewDecoder())
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode embedded file %s: %w", path, err)
	}

	var output bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(decoded))
	dir := filepath.Dir(path)

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "#include") {
			start := strings.Index(trimmed, "\"")
			end := strings.LastIndex(trimmed, "\"")
			if start != -1 && end != -1 && end > start {
				target := trimmed[start+1 : end]
				targetPath := filepath.Join(dir, target)

				inContent, err := recursiveReadEmbedded(targetPath, included)
				if err != nil {
					return nil, fmt.Errorf("error including %s from %s: %w", target, path, err)
				}
				output.Write(inContent)
				output.WriteString("\n")
			} else {
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
		if op.Cmd == "VarRef" {
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
			if v.Cmd == "VarRef" && len(v.Args) > 0 {
				if varName, ok := v.Args[0].(interpreter.Variable); ok {
					result[i] = string(varName)
					continue
				}
			}
			result[i] = engine.OpCode{
				Cmd:  v.Cmd,
				Args: convertOpCodeArgs(v.Args),
			}
		case []interpreter.OpCode:
			result[i] = convertToEngineOpCodes(v)
		case interpreter.Variable:
			result[i] = string(v)
		default:
			result[i] = arg
		}
	}
	return result
}
