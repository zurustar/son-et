// Package compiler provides the compilation pipeline for FILLY scripts (.TFY files).
package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/compiler"
	"github.com/zurustar/son-et/pkg/script"
)

// TestCompile tests the Compile function with various source code inputs.
func TestCompile(t *testing.T) {
	tests := []struct {
		name        string
		source      string
		wantOpcodes bool // true if we expect opcodes, false if we expect errors
		wantErrLen  int  // expected number of errors (0 for success)
	}{
		{
			name:        "empty source",
			source:      "",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "simple variable assignment",
			source:      "x = 5;",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "variable declaration and assignment",
			source:      "int x; x = 10;",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "function call",
			source:      "LoadPic(\"test.bmp\");",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "if statement",
			source:      "if (x > 5) { y = 10; }",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "for loop",
			source:      "for (i = 0; i < 10; i = i + 1) { x = i; }",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "while loop",
			source:      "while (x < 10) { x = x + 1; }",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "function definition",
			source:      "myFunc(int x) { return x + 1; }",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "mes event handler",
			source:      "mes(MIDI_TIME) { step(10) { func1();, } }",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "binary expression",
			source:      "x = 5 + 3 * 2;",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "array access",
			source:      "x = arr[0];",
			wantOpcodes: true,
			wantErrLen:  0,
		},
		{
			name:        "array assignment",
			source:      "arr[0] = 5;",
			wantOpcodes: true,
			wantErrLen:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opcodes, errs := Compile(tt.source)

			if tt.wantErrLen > 0 {
				if len(errs) != tt.wantErrLen {
					t.Errorf("Compile() error count = %d, want %d", len(errs), tt.wantErrLen)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("Compile() unexpected errors: %v", errs)
				}
			}

			if tt.wantOpcodes && len(errs) == 0 {
				// For non-empty source, we should have some opcodes
				// (except for empty source or declarations only)
				if tt.source != "" && tt.source != "int x;" {
					// Most test cases should produce opcodes
					_ = opcodes // opcodes are valid
				}
			}
		})
	}
}

// TestCompileSimpleAssignment tests that a simple assignment produces correct OpCode.
func TestCompileSimpleAssignment(t *testing.T) {
	source := "x = 5;"
	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Fatalf("Compile() unexpected errors: %v", errs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("Compile() expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != compiler.OpAssign {
		t.Errorf("Compile() expected OpAssign, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 2 {
		t.Fatalf("Compile() expected 2 args, got %d", len(opcodes[0].Args))
	}

	// Check variable name
	varName, ok := opcodes[0].Args[0].(compiler.Variable)
	if !ok {
		t.Errorf("Compile() expected Variable type for first arg, got %T", opcodes[0].Args[0])
	}
	if string(varName) != "x" {
		t.Errorf("Compile() expected variable name 'x', got '%s'", varName)
	}

	// Check value
	value, ok := opcodes[0].Args[1].(int64)
	if !ok {
		t.Errorf("Compile() expected int64 type for second arg, got %T", opcodes[0].Args[1])
	}
	if value != 5 {
		t.Errorf("Compile() expected value 5, got %d", value)
	}
}

// TestCompileFunctionCall tests that a function call produces correct OpCode.
func TestCompileFunctionCall(t *testing.T) {
	source := `LoadPic("test.bmp");`
	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Fatalf("Compile() unexpected errors: %v", errs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("Compile() expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != compiler.OpCall {
		t.Errorf("Compile() expected OpCall, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) < 2 {
		t.Fatalf("Compile() expected at least 2 args, got %d", len(opcodes[0].Args))
	}

	// Check function name
	funcName, ok := opcodes[0].Args[0].(string)
	if !ok {
		t.Errorf("Compile() expected string type for function name, got %T", opcodes[0].Args[0])
	}
	if funcName != "LoadPic" {
		t.Errorf("Compile() expected function name 'LoadPic', got '%s'", funcName)
	}

	// Check argument
	arg, ok := opcodes[0].Args[1].(string)
	if !ok {
		t.Errorf("Compile() expected string type for argument, got %T", opcodes[0].Args[1])
	}
	if arg != "test.bmp" {
		t.Errorf("Compile() expected argument 'test.bmp', got '%s'", arg)
	}
}

// TestCompileIfStatement tests that an if statement produces correct OpCode.
func TestCompileIfStatement(t *testing.T) {
	source := "if (x > 5) { y = 10; }"
	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Fatalf("Compile() unexpected errors: %v", errs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("Compile() expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != compiler.OpIf {
		t.Errorf("Compile() expected OpIf, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 3 {
		t.Fatalf("Compile() expected 3 args (condition, then, else), got %d", len(opcodes[0].Args))
	}
}

// TestCompileMesStatement tests that a mes statement produces correct OpCode.
func TestCompileMesStatement(t *testing.T) {
	source := "mes(MIDI_TIME) { x = 1; }"
	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Fatalf("Compile() unexpected errors: %v", errs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("Compile() expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != compiler.OpRegisterEventHandler {
		t.Errorf("Compile() expected OpRegisterEventHandler, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 2 {
		t.Fatalf("Compile() expected 2 args (event type, body), got %d", len(opcodes[0].Args))
	}

	// Check event type
	eventType, ok := opcodes[0].Args[0].(string)
	if !ok {
		t.Errorf("Compile() expected string type for event type, got %T", opcodes[0].Args[0])
	}
	if eventType != "MIDI_TIME" {
		t.Errorf("Compile() expected event type 'MIDI_TIME', got '%s'", eventType)
	}
}

// TestCompileStepStatement tests that a step statement produces correct OpCode.
func TestCompileStepStatement(t *testing.T) {
	source := "step(10) { func1();, func2();,, }"
	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Fatalf("Compile() unexpected errors: %v", errs)
	}

	// Expected: OpSetStep(10), OpCall(func1), OpWait(1), OpCall(func2), OpWait(2)
	if len(opcodes) < 3 {
		t.Fatalf("Compile() expected at least 3 opcodes, got %d", len(opcodes))
	}

	// First should be OpSetStep
	if opcodes[0].Cmd != compiler.OpSetStep {
		t.Errorf("Compile() expected OpSetStep, got %s", opcodes[0].Cmd)
	}
}

// TestCompileWithOptions tests the CompileWithOptions function.
func TestCompileWithOptions(t *testing.T) {
	source := "x = 5;"

	// Test with Debug = false
	opcodes, errs := CompileWithOptions(source, CompileOptions{Debug: false})
	if len(errs) > 0 {
		t.Errorf("CompileWithOptions() unexpected errors: %v", errs)
	}
	if len(opcodes) != 1 {
		t.Errorf("CompileWithOptions() expected 1 opcode, got %d", len(opcodes))
	}

	// Test with Debug = true
	opcodes, errs = CompileWithOptions(source, CompileOptions{Debug: true})
	if len(errs) > 0 {
		t.Errorf("CompileWithOptions() unexpected errors: %v", errs)
	}
	if len(opcodes) != 1 {
		t.Errorf("CompileWithOptions() expected 1 opcode, got %d", len(opcodes))
	}
}

// TestCompileFile tests the CompileFile function.
func TestCompileFile(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tfy")

	// Write UTF-8 content (simulating already converted content)
	content := "x = 5;\n"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test CompileFile
	opcodes, errs := CompileFile(tmpFile)
	if len(errs) > 0 {
		t.Errorf("CompileFile() unexpected errors: %v", errs)
	}
	if len(opcodes) != 1 {
		t.Errorf("CompileFile() expected 1 opcode, got %d", len(opcodes))
	}
}

// TestCompileFileNotFound tests CompileFile with a non-existent file.
func TestCompileFileNotFound(t *testing.T) {
	_, errs := CompileFile("/nonexistent/path/test.tfy")

	if len(errs) == 0 {
		t.Error("CompileFile() expected error for non-existent file")
	}
}

// TestCompileFileWithOptions tests the CompileFileWithOptions function.
func TestCompileFileWithOptions(t *testing.T) {
	// Create a temporary file with test content
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.tfy")

	content := "x = 5;\n"
	err := os.WriteFile(tmpFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}

	// Test CompileFileWithOptions
	opcodes, errs := CompileFileWithOptions(tmpFile, CompileOptions{Debug: true})
	if len(errs) > 0 {
		t.Errorf("CompileFileWithOptions() unexpected errors: %v", errs)
	}
	if len(opcodes) != 1 {
		t.Errorf("CompileFileWithOptions() expected 1 opcode, got %d", len(opcodes))
	}
}

// TestConvertShiftJISToUTF8 tests the Shift-JIS to UTF-8 conversion.
func TestConvertShiftJISToUTF8(t *testing.T) {
	// Test with ASCII content (should pass through unchanged)
	asciiContent := []byte("hello world")
	result, err := convertShiftJISToUTF8(asciiContent)
	if err != nil {
		t.Errorf("convertShiftJISToUTF8() unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("convertShiftJISToUTF8() expected 'hello world', got '%s'", result)
	}

	// Test with Shift-JIS encoded Japanese text
	// "テスト" in Shift-JIS is: 0x83 0x65 0x83 0x58 0x83 0x67
	shiftJISContent := []byte{0x83, 0x65, 0x83, 0x58, 0x83, 0x67}
	result, err = convertShiftJISToUTF8(shiftJISContent)
	if err != nil {
		t.Errorf("convertShiftJISToUTF8() unexpected error: %v", err)
	}
	if result != "テスト" {
		t.Errorf("convertShiftJISToUTF8() expected 'テスト', got '%s'", result)
	}
}

// TestReExportedTypes tests that re-exported types work correctly.
func TestReExportedTypes(t *testing.T) {
	// Test that OpCode type alias works
	var opcode OpCode
	opcode.Cmd = OpAssign
	opcode.Args = []any{Variable("x"), 5}

	if opcode.Cmd != compiler.OpAssign {
		t.Errorf("OpCode type alias not working correctly")
	}

	// Test that Variable type alias works
	var v Variable = "test"
	if string(v) != "test" {
		t.Errorf("Variable type alias not working correctly")
	}

	// Test that OpCmd constants are correctly re-exported
	if OpAssign != compiler.OpAssign {
		t.Errorf("OpAssign constant not correctly re-exported")
	}
	if OpCall != compiler.OpCall {
		t.Errorf("OpCall constant not correctly re-exported")
	}
	if OpIf != compiler.OpIf {
		t.Errorf("OpIf constant not correctly re-exported")
	}
}

// TestCompilePipeline tests the full compilation pipeline.
func TestCompilePipeline(t *testing.T) {
	// Test a more complex program that exercises the full pipeline
	source := `
		int x, y;
		x = 10;
		y = 20;
		
		myFunc(int a, int b) {
			return a + b;
		}
		
		if (x > 5) {
			y = myFunc(x, y);
		}
		
		for (i = 0; i < 10; i = i + 1) {
			x = x + 1;
		}
		
		mes(MIDI_TIME) {
			step(16) {
				LoadPic("test.bmp");,
			}
		}
	`

	opcodes, errs := Compile(source)

	if len(errs) > 0 {
		t.Fatalf("Compile() unexpected errors: %v", errs)
	}

	// We should have multiple opcodes for this program
	if len(opcodes) < 5 {
		t.Errorf("Compile() expected at least 5 opcodes for complex program, got %d", len(opcodes))
	}

	// Verify we have the expected opcode types
	hasAssign := false
	hasDefineFunction := false
	hasIf := false
	hasFor := false
	hasRegisterEventHandler := false

	for _, op := range opcodes {
		switch op.Cmd {
		case compiler.OpAssign:
			hasAssign = true
		case compiler.OpDefineFunction:
			hasDefineFunction = true
		case compiler.OpIf:
			hasIf = true
		case compiler.OpFor:
			hasFor = true
		case compiler.OpRegisterEventHandler:
			hasRegisterEventHandler = true
		}
	}

	if !hasAssign {
		t.Error("Compile() expected OpAssign in output")
	}
	if !hasDefineFunction {
		t.Error("Compile() expected OpDefineFunction in output")
	}
	if !hasIf {
		t.Error("Compile() expected OpIf in output")
	}
	if !hasFor {
		t.Error("Compile() expected OpFor in output")
	}
	if !hasRegisterEventHandler {
		t.Error("Compile() expected OpRegisterEventHandler in output")
	}
}

// TestCompileScripts tests the CompileScripts function with multiple scripts.
func TestCompileScripts(t *testing.T) {
	// Create test scripts (simulating what script.Loader would return)
	scripts := []script.Script{
		{
			FileName: "test1.tfy",
			Content:  "x = 5;",
			Size:     6,
		},
		{
			FileName: "test2.tfy",
			Content:  "y = 10;",
			Size:     7,
		},
		{
			FileName: "test3.tfy",
			Content:  "z = x + y;",
			Size:     10,
		},
	}

	results, errs := CompileScripts(scripts)

	if len(errs) > 0 {
		t.Errorf("CompileScripts() unexpected errors: %v", errs)
	}

	if len(results) != 3 {
		t.Errorf("CompileScripts() expected 3 results, got %d", len(results))
	}

	// Check that each script was compiled
	for _, s := range scripts {
		if _, ok := results[s.FileName]; !ok {
			t.Errorf("CompileScripts() missing result for %s", s.FileName)
		}
	}

	// Verify test1.tfy has correct OpCode
	if opcodes, ok := results["test1.tfy"]; ok {
		if len(opcodes) != 1 {
			t.Errorf("CompileScripts() expected 1 opcode for test1.tfy, got %d", len(opcodes))
		}
		if opcodes[0].Cmd != compiler.OpAssign {
			t.Errorf("CompileScripts() expected OpAssign for test1.tfy, got %s", opcodes[0].Cmd)
		}
	}
}

// TestCompileScriptsWithErrors tests CompileScripts with scripts that have errors.
func TestCompileScriptsWithErrors(t *testing.T) {
	scripts := []script.Script{
		{
			FileName: "valid.tfy",
			Content:  "x = 5;",
			Size:     6,
		},
		{
			FileName: "invalid.tfy",
			Content:  "x = ;", // Invalid syntax
			Size:     5,
		},
	}

	results, errs := CompileScripts(scripts)

	// Should have at least one error from invalid.tfy
	if len(errs) == 0 {
		t.Error("CompileScripts() expected errors for invalid script")
	}

	// Valid script should still be compiled
	if _, ok := results["valid.tfy"]; !ok {
		t.Error("CompileScripts() should have compiled valid.tfy despite invalid.tfy error")
	}

	// Invalid script should not be in results
	if _, ok := results["invalid.tfy"]; ok {
		t.Error("CompileScripts() should not have result for invalid.tfy")
	}

	// Check that error contains file name
	foundFileNameInError := false
	for _, err := range errs {
		if strings.Contains(err.Error(), "invalid.tfy") {
			foundFileNameInError = true
			break
		}
	}
	if !foundFileNameInError {
		t.Error("CompileScripts() error should contain file name")
	}
}

// TestCompileScriptsWithResults tests the CompileScriptsWithResults function.
func TestCompileScriptsWithResults(t *testing.T) {
	scripts := []script.Script{
		{
			FileName: "test1.tfy",
			Content:  "x = 5;",
			Size:     6,
		},
		{
			FileName: "test2.tfy",
			Content:  "y = ;", // Invalid syntax
			Size:     5,
		},
	}

	results := CompileScriptsWithResults(scripts)

	if len(results) != 2 {
		t.Fatalf("CompileScriptsWithResults() expected 2 results, got %d", len(results))
	}

	// Check first result (valid)
	if results[0].FileName != "test1.tfy" {
		t.Errorf("CompileScriptsWithResults() expected first result to be test1.tfy, got %s", results[0].FileName)
	}
	if len(results[0].Errors) > 0 {
		t.Errorf("CompileScriptsWithResults() unexpected errors for test1.tfy: %v", results[0].Errors)
	}
	if len(results[0].OpCodes) != 1 {
		t.Errorf("CompileScriptsWithResults() expected 1 opcode for test1.tfy, got %d", len(results[0].OpCodes))
	}

	// Check second result (invalid)
	if results[1].FileName != "test2.tfy" {
		t.Errorf("CompileScriptsWithResults() expected second result to be test2.tfy, got %s", results[1].FileName)
	}
	if len(results[1].Errors) == 0 {
		t.Error("CompileScriptsWithResults() expected errors for test2.tfy")
	}
	if results[1].OpCodes != nil {
		t.Error("CompileScriptsWithResults() expected nil OpCodes for test2.tfy")
	}
}

// TestCompileScriptsEmpty tests CompileScripts with empty input.
func TestCompileScriptsEmpty(t *testing.T) {
	scripts := []script.Script{}

	results, errs := CompileScripts(scripts)

	if len(errs) > 0 {
		t.Errorf("CompileScripts() unexpected errors for empty input: %v", errs)
	}

	if len(results) != 0 {
		t.Errorf("CompileScripts() expected 0 results for empty input, got %d", len(results))
	}
}

// TestCompileDirectory tests the CompileDirectory function.
func TestCompileDirectory(t *testing.T) {
	// Create a temporary directory with test scripts
	tmpDir := t.TempDir()

	// Create test script files
	script1 := filepath.Join(tmpDir, "test1.tfy")
	script2 := filepath.Join(tmpDir, "test2.TFY") // Test case-insensitive extension

	err := os.WriteFile(script1, []byte("x = 5;"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = os.WriteFile(script2, []byte("y = 10;"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Test CompileDirectory
	results, errs := CompileDirectory(tmpDir)

	if len(errs) > 0 {
		t.Errorf("CompileDirectory() unexpected errors: %v", errs)
	}

	if len(results) != 2 {
		t.Errorf("CompileDirectory() expected 2 results, got %d", len(results))
	}

	// Check that both scripts were compiled
	if _, ok := results["test1.tfy"]; !ok {
		t.Error("CompileDirectory() missing result for test1.tfy")
	}
	if _, ok := results["test2.TFY"]; !ok {
		t.Error("CompileDirectory() missing result for test2.TFY")
	}
}

// TestCompileDirectoryNotFound tests CompileDirectory with non-existent directory.
func TestCompileDirectoryNotFound(t *testing.T) {
	_, errs := CompileDirectory("/nonexistent/directory/path")

	if len(errs) == 0 {
		t.Error("CompileDirectory() expected error for non-existent directory")
	}
}

// TestCompileDirectoryWithResults tests the CompileDirectoryWithResults function.
func TestCompileDirectoryWithResults(t *testing.T) {
	// Create a temporary directory with test scripts
	tmpDir := t.TempDir()

	// Create test script files
	script1 := filepath.Join(tmpDir, "valid.tfy")
	script2 := filepath.Join(tmpDir, "invalid.tfy")

	err := os.WriteFile(script1, []byte("x = 5;"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	err = os.WriteFile(script2, []byte("x = ;"), 0644) // Invalid syntax
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Test CompileDirectoryWithResults
	results, loadErr := CompileDirectoryWithResults(tmpDir)

	if loadErr != nil {
		t.Fatalf("CompileDirectoryWithResults() unexpected load error: %v", loadErr)
	}

	if len(results) != 2 {
		t.Fatalf("CompileDirectoryWithResults() expected 2 results, got %d", len(results))
	}

	// Find results by file name
	var validResult, invalidResult *CompileResult
	for i := range results {
		if results[i].FileName == "valid.tfy" {
			validResult = &results[i]
		} else if results[i].FileName == "invalid.tfy" {
			invalidResult = &results[i]
		}
	}

	if validResult == nil {
		t.Fatal("CompileDirectoryWithResults() missing result for valid.tfy")
	}
	if invalidResult == nil {
		t.Fatal("CompileDirectoryWithResults() missing result for invalid.tfy")
	}

	// Check valid result
	if len(validResult.Errors) > 0 {
		t.Errorf("CompileDirectoryWithResults() unexpected errors for valid.tfy: %v", validResult.Errors)
	}
	if len(validResult.OpCodes) != 1 {
		t.Errorf("CompileDirectoryWithResults() expected 1 opcode for valid.tfy, got %d", len(validResult.OpCodes))
	}

	// Check invalid result
	if len(invalidResult.Errors) == 0 {
		t.Error("CompileDirectoryWithResults() expected errors for invalid.tfy")
	}
}

// TestCompileDirectoryWithResultsNotFound tests CompileDirectoryWithResults with non-existent directory.
func TestCompileDirectoryWithResultsNotFound(t *testing.T) {
	_, err := CompileDirectoryWithResults("/nonexistent/directory/path")

	if err == nil {
		t.Error("CompileDirectoryWithResults() expected error for non-existent directory")
	}
}

// TestCompileResultType tests the CompileResult type.
func TestCompileResultType(t *testing.T) {
	result := CompileResult{
		FileName: "test.tfy",
		OpCodes:  []compiler.OpCode{{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), 5}}},
		Errors:   nil,
	}

	if result.FileName != "test.tfy" {
		t.Errorf("CompileResult.FileName expected 'test.tfy', got '%s'", result.FileName)
	}

	if len(result.OpCodes) != 1 {
		t.Errorf("CompileResult.OpCodes expected 1 opcode, got %d", len(result.OpCodes))
	}

	if result.Errors != nil {
		t.Errorf("CompileResult.Errors expected nil, got %v", result.Errors)
	}
}
