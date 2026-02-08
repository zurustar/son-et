// Package compiler provides the compilation pipeline for FILLY scripts (.TFY files).
// It transforms source code into OpCode through three phases:
// 1. Lexer: Tokenization
// 2. Parser: AST generation
// 3. Compiler: OpCode generation
//
// This package provides a unified API for compiling FILLY scripts:
// - Compile: Compiles source code string to OpCode
// - CompileFile: Compiles a file to OpCode (handles Shift-JIS encoding)
// - CompileWithOptions: Compiles with additional options
// - CompileScripts: Compiles multiple scripts loaded by script.Loader
// - CompileDirectory: Loads and compiles all scripts from a directory
// - FindMainScript: Finds the script containing the main function entry point
// - CompileWithEntryPoint: Compiles scripts starting from the main entry point
package compiler

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"strings"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/zurustar/son-et/pkg/compiler/compiler"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
	"github.com/zurustar/son-et/pkg/compiler/preprocessor"
	"github.com/zurustar/son-et/pkg/opcode"
	"github.com/zurustar/son-et/pkg/script"
)

// CompileOptions provides configuration options for compilation.
type CompileOptions struct {
	// Debug includes debug information in the output
	Debug bool
}

// Compile compiles source code to OpCode.
// It chains the lexer → parser → compiler pipeline.
//
// Parameters:
//   - source: UTF-8 encoded source code string
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - []error: Any compilation errors (empty if successful)
//
// Requirement 6.1: System provides a compilation pipeline chaining Lexer, Parser, Compiler.
// Requirement 6.2: Pipeline executes in Lexer → Parser → Compiler order.
// Requirement 6.3: If any phase fails, stop pipeline and return accumulated errors.
// Requirement 5.6: System collects all errors and returns them to caller.
// Requirement 10.2: CompileString function accepts script content as string.
func Compile(source string) ([]opcode.OpCode, []error) {
	// Phase 1: Lexical analysis
	l := lexer.New(source)

	// Phase 2: Syntax analysis
	p := parser.New(l)
	program, parseErrs := p.ParseProgram()

	// Requirement 6.3: If any phase fails, stop pipeline and return accumulated errors
	if len(parseErrs) > 0 {
		// Convert to CompileError with context if possible
		var compileErrors []error
		for _, err := range parseErrs {
			if pe, ok := err.(*parser.ParserError); ok {
				compileErrors = append(compileErrors, NewParserErrorWithContext(
					pe.Message, pe.Line, pe.Column, source))
			} else {
				compileErrors = append(compileErrors, err)
			}
		}
		return nil, compileErrors
	}

	// Phase 3: OpCode generation
	c := compiler.New()
	opcodes, compileErrs := c.Compile(program)

	// Requirement 6.3: Return all errors if compilation fails
	if len(compileErrs) > 0 {
		// Convert to CompileError with context if possible
		var compileErrors []error
		for _, err := range compileErrs {
			if ce, ok := err.(*compiler.CompilerError); ok {
				compileErrors = append(compileErrors, NewCompilerErrorWithContext(
					ce.Message, ce.Line, ce.Column, source))
			} else {
				compileErrors = append(compileErrors, err)
			}
		}
		return nil, compileErrors
	}

	// Requirement 6.4: Return generated OpCode sequence on success
	return opcodes, nil
}

// CompileFile compiles a file to OpCode.
// It reads the file, handles Shift-JIS to UTF-8 encoding conversion,
// and then compiles the content.
//
// Parameters:
//   - path: Path to the .TFY script file
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - []error: Any compilation errors (empty if successful)
//
// Requirement 1.1: Compiler reads script content from file path.
// Requirement 1.4: Returns descriptive error with file path if file cannot be read.
// Requirement 1.5: Correctly processes Shift-JIS encoded files.
// Requirement 10.1: Compile function accepts script path and returns OpCode or errors.
func CompileFile(path string) ([]opcode.OpCode, []error) {
	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to read file %s: %w", path, err)}
	}

	// Convert Shift-JIS to UTF-8
	content, err := convertShiftJISToUTF8(data)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to convert encoding for %s: %w", path, err)}
	}

	// Compile the content
	return Compile(content)
}

// CompileWithOptions compiles source code with additional options.
// This function provides more control over the compilation process.
//
// Parameters:
//   - source: UTF-8 encoded source code string
//   - opts: Compilation options
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - []error: Any compilation errors (empty if successful)
//
// Requirement 10.3: CompileWithOptions accepts compiler configuration options.
func CompileWithOptions(source string, opts CompileOptions) ([]opcode.OpCode, []error) {
	// Currently, the Debug option is reserved for future use.
	// The basic compilation pipeline is the same as Compile.
	// When Debug is true, additional debug information could be included
	// in the OpCode output (e.g., source line numbers, variable names).

	// For now, delegate to the standard Compile function
	opcodes, errs := Compile(source)

	if opts.Debug && len(errs) == 0 {
		// Future: Add debug information to opcodes
		// This could include source mapping, variable tracking, etc.
	}

	return opcodes, errs
}

// CompileFileWithOptions compiles a file with additional options.
// It combines file reading with option-based compilation.
//
// Parameters:
//   - path: Path to the .TFY script file
//   - opts: Compilation options
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - []error: Any compilation errors (empty if successful)
func CompileFileWithOptions(path string, opts CompileOptions) ([]opcode.OpCode, []error) {
	// Read file content
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to read file %s: %w", path, err)}
	}

	// Convert Shift-JIS to UTF-8
	content, err := convertShiftJISToUTF8(data)
	if err != nil {
		return nil, []error{fmt.Errorf("failed to convert encoding for %s: %w", path, err)}
	}

	// Compile with options
	return CompileWithOptions(content, opts)
}

// CompileResult represents the compilation result for a single script.
// It contains the file name, compiled OpCodes, and any errors that occurred.
type CompileResult struct {
	// FileName is the name of the script file
	FileName string
	// OpCodes is the compiled OpCode sequence (nil if compilation failed)
	OpCodes []opcode.OpCode
	// Errors contains any compilation errors (empty if successful)
	Errors []error
}

// CompileScripts compiles multiple scripts loaded by script.Loader.
// Each script is compiled independently, and results are returned for all scripts.
// This function does not stop on the first error; it compiles all scripts and
// collects all results.
//
// Parameters:
//   - scripts: Slice of Script structs from script.Loader (already UTF-8 converted)
//
// Returns:
//   - map[string][]opcode.OpCode: Map of file name to compiled OpCodes (only successful compilations)
//   - []error: All compilation errors from all scripts (empty if all successful)
//
// Requirement 6.5: System integrates with existing script loading functionality.
// Requirement 6.6: When processing multiple script files, system compiles each file independently.
func CompileScripts(scripts []script.Script) (map[string][]opcode.OpCode, []error) {
	results := make(map[string][]opcode.OpCode)
	var allErrors []error

	for _, s := range scripts {
		// Script.Content is already UTF-8 converted by script.Loader
		opcodes, errs := Compile(s.Content)

		if len(errs) > 0 {
			// Wrap errors with file name for context
			for _, err := range errs {
				allErrors = append(allErrors, fmt.Errorf("%s: %w", s.FileName, err))
			}
		} else {
			results[s.FileName] = opcodes
		}
	}

	return results, allErrors
}

// CompileScriptsWithResults compiles multiple scripts and returns detailed results.
// Unlike CompileScripts, this function returns a CompileResult for each script,
// allowing the caller to see both successful and failed compilations.
//
// Parameters:
//   - scripts: Slice of Script structs from script.Loader (already UTF-8 converted)
//
// Returns:
//   - []CompileResult: Compilation results for each script
//
// Requirement 6.5: System integrates with existing script loading functionality.
// Requirement 6.6: When processing multiple script files, system compiles each file independently.
func CompileScriptsWithResults(scripts []script.Script) []CompileResult {
	results := make([]CompileResult, 0, len(scripts))

	for _, s := range scripts {
		// Script.Content is already UTF-8 converted by script.Loader
		opcodes, errs := Compile(s.Content)

		result := CompileResult{
			FileName: s.FileName,
			OpCodes:  opcodes,
			Errors:   errs,
		}
		results = append(results, result)
	}

	return results
}

// CompileDirectory loads all .TFY scripts from a directory and compiles them.
// This is a convenience function that combines script.Loader with CompileScripts.
//
// Parameters:
//   - dirPath: Path to the directory containing .TFY script files
//
// Returns:
//   - map[string][]opcode.OpCode: Map of file name to compiled OpCodes (only successful compilations)
//   - []error: All errors (loading and compilation) (empty if all successful)
//
// Requirement 1.2: When multiple TFY files exist in directory, compiler loads all TFY files.
// Requirement 6.5: System integrates with existing script loading functionality.
// Requirement 6.6: When processing multiple script files, system compiles each file independently.
func CompileDirectory(dirPath string) (map[string][]opcode.OpCode, []error) {
	// Use script.Loader to find and load all .TFY files
	loader := script.NewLoader(dirPath)
	scripts, err := loader.LoadAllScripts()
	if err != nil {
		return nil, []error{fmt.Errorf("failed to load scripts from %s: %w", dirPath, err)}
	}

	// Compile all loaded scripts
	return CompileScripts(scripts)
}

// CompileDirectoryWithResults loads all .TFY scripts from a directory and compiles them,
// returning detailed results for each script.
//
// Parameters:
//   - dirPath: Path to the directory containing .TFY script files
//
// Returns:
//   - []CompileResult: Compilation results for each script
//   - error: Error if loading scripts failed (nil if loading succeeded)
//
// Requirement 1.2: When multiple TFY files exist in directory, compiler loads all TFY files.
// Requirement 6.5: System integrates with existing script loading functionality.
// Requirement 6.6: When processing multiple script files, system compiles each file independently.
func CompileDirectoryWithResults(dirPath string) ([]CompileResult, error) {
	// Use script.Loader to find and load all .TFY files
	loader := script.NewLoader(dirPath)
	scripts, err := loader.LoadAllScripts()
	if err != nil {
		return nil, fmt.Errorf("failed to load scripts from %s: %w", dirPath, err)
	}

	// Compile all loaded scripts
	return CompileScriptsWithResults(scripts), nil
}

// MainScriptInfo contains information about a script with a main function.
type MainScriptInfo struct {
	Script   *script.Script
	FileName string
}

// FindMainScript finds the script containing the main function entry point.
// It parses all scripts and looks for a function named "main" (case-insensitive).
//
// Parameters:
//   - scripts: Slice of Script structs from script.Loader (already UTF-8 converted)
//
// Returns:
//   - *MainScriptInfo: Information about the script containing main function
//   - error: Error if no main function found, or multiple main functions found
//
// Requirement 14.1: System scans all TFY files to identify the file containing main function.
// Requirement 14.2: When main function exists in multiple files, report error.
// Requirement 14.3: When main function is not found, report error.
func FindMainScript(scripts []script.Script) (*MainScriptInfo, error) {
	var mainScripts []MainScriptInfo

	for i := range scripts {
		s := &scripts[i]
		hasMain, err := containsMainFunction(s.Content)
		if err != nil {
			// Parse error - skip this file but log it
			continue
		}
		if hasMain {
			mainScripts = append(mainScripts, MainScriptInfo{
				Script:   s,
				FileName: s.FileName,
			})
		}
	}

	// Requirement 14.3: When main function is not found, report error
	if len(mainScripts) == 0 {
		return nil, fmt.Errorf("no main function found in any script file")
	}

	// Requirement 14.2: When main function exists in multiple files, report error
	if len(mainScripts) > 1 {
		names := make([]string, len(mainScripts))
		for i, info := range mainScripts {
			names[i] = info.FileName
		}
		return nil, fmt.Errorf("multiple main functions found in: %v", names)
	}

	return &mainScripts[0], nil
}

// containsMainFunction checks if the source code contains a main function definition.
// It uses the parser to accurately detect function definitions, avoiding false positives
// from comments or string literals.
//
// Parameters:
//   - content: UTF-8 encoded source code string
//
// Returns:
//   - bool: true if main function is found
//   - error: Parse error if any
func containsMainFunction(content string) (bool, error) {
	l := lexer.New(content)
	p := parser.New(l)
	program, _ := p.ParseProgram()

	// Even if there are parse errors, we can still check for main function
	// in the successfully parsed statements
	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*parser.FunctionStatement); ok {
			if strings.EqualFold(fn.Name, "main") {
				return true, nil
			}
		}
	}

	return false, nil
}

// CompileWithEntryPoint compiles scripts starting from the main entry point.
// It finds the main function, compiles all scripts, and returns the combined OpCodes.
//
// Parameters:
//   - scripts: Slice of Script structs from script.Loader (already UTF-8 converted)
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence from all scripts
//   - error: Error if compilation failed
//
// Requirement 13.1: Application calls compiler after loading scripts to generate OpCode.
// Requirement 14.4: When file containing main function is identified, start compilation from that file.
func CompileWithEntryPoint(scripts []script.Script) ([]opcode.OpCode, error) {
	// Find the main entry point
	mainInfo, err := FindMainScript(scripts)
	if err != nil {
		return nil, err
	}

	// Compile all scripts and collect OpCodes
	// The main script's OpCodes should be executed first
	var allOpCodes []opcode.OpCode

	// First, compile the main script
	mainOpCodes, errs := Compile(mainInfo.Script.Content)
	if len(errs) > 0 {
		return nil, fmt.Errorf("failed to compile main script %s: %v", mainInfo.FileName, errs[0])
	}
	allOpCodes = append(allOpCodes, mainOpCodes...)

	// Then compile other scripts (for function definitions, etc.)
	for i := range scripts {
		s := &scripts[i]
		if s.FileName == mainInfo.FileName {
			continue // Already compiled
		}

		opcodes, errs := Compile(s.Content)
		if len(errs) > 0 {
			return nil, fmt.Errorf("failed to compile script %s: %v", s.FileName, errs[0])
		}
		allOpCodes = append(allOpCodes, opcodes...)
	}

	return allOpCodes, nil
}

// CompileDirectoryWithEntryPoint loads all .TFY scripts from a directory,
// finds the main entry point, and compiles them.
//
// Parameters:
//   - dirPath: Path to the directory containing .TFY script files
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - error: Error if loading or compilation failed
//
// Requirement 13.1: Application calls compiler after loading scripts to generate OpCode.
// Requirement 14.4: When file containing main function is identified, start compilation from that file.
func CompileDirectoryWithEntryPoint(dirPath string) ([]opcode.OpCode, error) {
	// Use script.Loader to find and load all .TFY files
	loader := script.NewLoader(dirPath)
	scripts, err := loader.LoadAllScripts()
	if err != nil {
		return nil, fmt.Errorf("failed to load scripts from %s: %w", dirPath, err)
	}

	// Compile with entry point resolution
	return CompileWithEntryPoint(scripts)
}

// PreprocessResult contains the result of preprocessing.
type PreprocessResult = preprocessor.PreprocessResult

// CompileWithPreprocessor compiles a script using the preprocessor to expand #include directives.
// It starts from the entry file and recursively includes all dependencies.
//
// Parameters:
//   - dirPath: Path to the directory containing .TFY script files
//   - entryFile: The entry point file name (relative to dirPath)
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - *PreprocessResult: The preprocessing result (included files list)
//   - error: Error if preprocessing or compilation failed
//
// Requirement 16.1: Preprocessor starts processing from entry point file.
// Requirement 16.2: Preprocessor expands #include directives.
// Requirement 16.3: Preprocessor processes included files recursively.
// Requirement 16.6: Preprocessor outputs single combined source code.
func CompileWithPreprocessor(dirPath string, entryFile string) ([]opcode.OpCode, *PreprocessResult, error) {
	// Create preprocessor
	p := preprocessor.New(dirPath)

	// Preprocess the entry file
	result, err := p.PreprocessFile(entryFile)
	if err != nil {
		return nil, nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	// Compile the preprocessed source
	opcodes, errs := Compile(result.Source)
	if len(errs) > 0 {
		return nil, result, fmt.Errorf("compilation failed: %v", errs[0])
	}

	return opcodes, result, nil
}

// CompileWithPreprocessorFS compiles a script using the preprocessor with a custom file system.
// This is used for embedded file systems.
//
// Parameters:
//   - dirPath: Path to the directory containing .TFY script files
//   - entryFile: The entry point file name (relative to dirPath)
//   - fsys: The file system to use (can be embed.FS or os.DirFS)
//
// Returns:
//   - []opcode.OpCode: The compiled OpCode sequence
//   - *PreprocessResult: The preprocessing result (included files list)
//   - error: Error if preprocessing or compilation failed
func CompileWithPreprocessorFS(dirPath string, entryFile string, fsys fs.FS) ([]opcode.OpCode, *PreprocessResult, error) {
	// Create preprocessor with custom file system
	p := preprocessor.NewWithFS(dirPath, fsys)

	// Preprocess the entry file
	result, err := p.PreprocessFile(entryFile)
	if err != nil {
		return nil, nil, fmt.Errorf("preprocessing failed: %w", err)
	}

	// Compile the preprocessed source
	opcodes, errs := Compile(result.Source)
	if len(errs) > 0 {
		return nil, result, fmt.Errorf("compilation failed: %v", errs[0])
	}

	return opcodes, result, nil
}

// convertShiftJISToUTF8 converts Shift-JIS encoded data to UTF-8.
// This function handles the encoding conversion required for .TFY files
// which are typically encoded in Shift-JIS.
//
// Parameters:
//   - data: Raw bytes potentially in Shift-JIS encoding
//
// Returns:
//   - string: UTF-8 encoded string
//   - error: Conversion error if any
func convertShiftJISToUTF8(data []byte) (string, error) {
	// Create Shift-JIS decoder
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(strings.NewReader(string(data)), decoder)

	// Convert to UTF-8
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return "", fmt.Errorf("failed to decode Shift-JIS: %w", err)
	}

	return string(utf8Data), nil
}

// Re-export types from pkg/opcode for convenience
// This allows users to import only the main compiler package

// OpCode is re-exported from the opcode package
type OpCode = opcode.OpCode

// OpCmd is re-exported from the opcode package
type OpCmd = opcode.Cmd

// Variable is re-exported from the opcode package
type Variable = opcode.Variable

// Re-export OpCode command constants
const (
	OpAssign               = opcode.Assign
	OpArrayAssign          = opcode.ArrayAssign
	OpCall                 = opcode.Call
	OpBinaryOp             = opcode.BinaryOp
	OpUnaryOp              = opcode.UnaryOp
	OpArrayAccess          = opcode.ArrayAccess
	OpIf                   = opcode.If
	OpFor                  = opcode.For
	OpWhile                = opcode.While
	OpSwitch               = opcode.Switch
	OpBreak                = opcode.Break
	OpContinue             = opcode.Continue
	OpRegisterEventHandler = opcode.RegisterEventHandler
	OpWait                 = opcode.Wait
	OpSetStep              = opcode.SetStep
	OpDefineFunction       = opcode.DefineFunction
)

// Re-export error types from sub-packages for convenience

// LexerError is re-exported from the lexer sub-package
type LexerError = lexer.LexerError

// ParserError is re-exported from the parser sub-package
type ParserError = parser.ParserError

// CompilerError is re-exported from the compiler sub-package
type CompilerErrorType = compiler.CompilerError
