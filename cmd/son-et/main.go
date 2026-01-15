package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/zurustar/filly2exe/pkg/compiler/ast"
	"github.com/zurustar/filly2exe/pkg/compiler/codegen"
	"github.com/zurustar/filly2exe/pkg/compiler/lexer"
	"github.com/zurustar/filly2exe/pkg/compiler/parser"
	"github.com/zurustar/filly2exe/pkg/compiler/token"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

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
			// Simple parse: quote to quote
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
				// Malformed include, just keep line?
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

// scanAssetReferences walks the AST and collects all asset filenames from LoadPic, PlayMIDI, PlayWAVE calls
func scanAssetReferences(program *ast.Program) []string {
	var assets []string
	seen := make(map[string]bool)

	var walker func(node ast.Node)
	walker = func(node ast.Node) {
		switch n := node.(type) {
		case *ast.Program:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.FunctionStatement:
			walker(n.Body)
		case *ast.ExpressionStatement:
			walker(n.Expression)
		case *ast.AssignStatement:
			walker(n.Value)
		case *ast.MesBlockStatement:
			walker(n.Body)
		case *ast.StepBlockStatement:
			walker(n.Body)
		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.IfStatement:
			walker(n.Condition)
			walker(n.Consequence)
			if n.Alternative != nil {
				walker(n.Alternative)
			}
		case *ast.ForStatement:
			if n.Init != nil {
				walker(n.Init)
			}
			if n.Condition != nil {
				walker(n.Condition)
			}
			if n.Post != nil {
				walker(n.Post)
			}
			walker(n.Body)
		case *ast.CallExpression:
			funcName := strings.ToLower(n.Function.Value)
			// Check for asset loading functions
			if funcName == "loadpic" || funcName == "playmidi" || funcName == "playwave" {
				if len(n.Arguments) > 0 {
					if strLit, ok := n.Arguments[0].(*ast.StringLiteral); ok {
						if !seen[strLit.Value] {
							assets = append(assets, strLit.Value)
							seen[strLit.Value] = true
						}
					}
				}
			}
			// Recursively check arguments
			for _, arg := range n.Arguments {
				walker(arg)
			}
		case *ast.InfixExpression:
			walker(n.Left)
			walker(n.Right)
		case *ast.PrefixExpression:
			walker(n.Right)
		case *ast.IndexExpression:
			walker(n.Left)
			if n.Index != nil {
				walker(n.Index)
			}
		}
	}

	walker(program)
	return assets
}

// findAssetFile performs case-insensitive file search in the given directory
// Returns the actual filesystem name (preserving case) for proper go:embed
func findAssetFile(dir, filename string) string {
	// Always do directory listing to get the actual filesystem name
	// This is necessary on case-insensitive filesystems (macOS, Windows)
	// to ensure go:embed uses the correct case
	entries, err := os.ReadDir(dir)
	if err != nil {
		return ""
	}

	lowerFilename := strings.ToLower(filename)
	for _, entry := range entries {
		if strings.ToLower(entry.Name()) == lowerFilename {
			// Return the actual filesystem name (with correct case)
			return filepath.Join(dir, entry.Name())
		}
	}

	return ""
}

func main() {
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		fmt.Println("Usage: son-et <input.fil>")
		os.Exit(1)
	}

	entryPath := args[0]
	// Using map to prevent cycles
	included := make(map[string]bool)

	fullCode, err := recursiveRead(entryPath, included)
	if err != nil {
		log.Fatalf("Error processing files: %v", err)
	}

	os.WriteFile("debug_merged.fil", fullCode, 0644)

	// 1. Lexing
	// DEBUG
	if false {
		d := lexer.New(string(fullCode))
		for {
			tok := d.NextToken()
			fmt.Printf("DEBUG TOKEN: %s (%q)\n", tok.Type, tok.Literal)
			if tok.Type == token.EOF {
				break
			}
		}
	}
	l := lexer.New(string(fullCode))

	// 2. Parsing
	p := parser.New(l)
	program := p.ParseProgram()

	// Debug: write merged content
	err = os.WriteFile("debug_merged.fil", fullCode, 0644)
	if err != nil {
		log.Printf("Warning: couldn't write debug file: %v", err)
	}

	if len(p.Errors()) != 0 {
		for _, msg := range p.Errors() {
			fmt.Printf("Parser Error: %s\n", msg)
		}
		os.Exit(1)
	}

	// 3. Scan for assets referenced in the code
	// First, scan the AST for LoadPic, PlayMIDI, PlayWAVE calls
	fileDir := filepath.Dir(entryPath)
	fmt.Fprintf(os.Stderr, "DEBUG: Scanning for asset references in code\n")

	referencedAssets := scanAssetReferences(program)
	fmt.Fprintf(os.Stderr, "DEBUG: Found %d asset references in code\n", len(referencedAssets))

	// Now find the actual files (case-insensitive matching)
	var assets []string
	seenAssets := make(map[string]bool)

	for _, assetRef := range referencedAssets {
		// Try to find the file with case-insensitive matching
		found := findAssetFile(fileDir, assetRef)
		if found != "" {
			rel, err := filepath.Rel(fileDir, found)
			if err != nil {
				fmt.Fprintf(os.Stderr, "DEBUG: Rel error for %s: %v\n", found, err)
				continue
			}
			if !seenAssets[rel] {
				assets = append(assets, rel)
				seenAssets[rel] = true
				fmt.Fprintf(os.Stderr, "DEBUG: Asset embedded: %s (referenced as: %s)\n", rel, assetRef)
			}
		} else {
			fmt.Fprintf(os.Stderr, "WARNING: Asset not found: %s\n", assetRef)
		}
	}

	// Also scan directory for common asset patterns (BMP, MID, WAV)
	// This catches dynamically generated filenames like StrPrint("ROBOT%03d.BMP", i)
	entries, err := os.ReadDir(fileDir)
	if err == nil {
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			ext := strings.ToLower(filepath.Ext(name))
			// Include common asset types
			if ext == ".bmp" || ext == ".mid" || ext == ".wav" {
				if !seenAssets[name] {
					assets = append(assets, name)
					seenAssets[name] = true
					fmt.Fprintf(os.Stderr, "DEBUG: Asset auto-detected: %s\n", name)
				}
			}
		}
	}

	fmt.Fprintf(os.Stderr, "DEBUG: Total assets to embed: %d\n", len(assets))

	// 4. Code Generation
	gen := codegen.New(assets)
	code := gen.Generate(program)

	// Output to file
	baseName := filepath.Base(entryPath)
	ext := filepath.Ext(baseName)
	paramName := strings.TrimSuffix(baseName, ext)
	outName := filepath.Join(filepath.Dir(entryPath), strings.ToLower(paramName)+"_game.go")

	err = os.WriteFile(outName, []byte(code), 0644)
	if err != nil {
		log.Fatalf("Failed to write output file: %v", err)
	}
	fmt.Fprintf(os.Stderr, "Generated %s\n", outName)

	// 4. Output (Optional: Print to stdout too?)
	// fmt.Println(code)
}
