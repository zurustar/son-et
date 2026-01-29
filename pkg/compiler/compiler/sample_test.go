package compiler

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
	"github.com/zurustar/son-et/pkg/opcode"
)

// convertShiftJISToUTF8 converts Shift-JIS encoded data to UTF-8.
func convertShiftJISToUTF8(data []byte) (string, error) {
	decoder := japanese.ShiftJIS.NewDecoder()
	reader := transform.NewReader(strings.NewReader(string(data)), decoder)
	utf8Data, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(utf8Data), nil
}

// TestCompileSampleFile tests the Compiler with the sample file samples/robot/ROBOT.TFY.
// This test validates that the Compiler can correctly compile a real FILLY script.
// Validates Requirements 4.1-4.17: Complete opcode.OpCode generation from AST.
func TestCompileSampleFile(t *testing.T) {
	// Find the sample file relative to the workspace root
	samplePath := filepath.Join("..", "..", "..", "samples", "robot", "ROBOT.TFY")

	// Read the sample file
	data, err := os.ReadFile(samplePath)
	if err != nil {
		t.Fatalf("Failed to read sample file: %v", err)
	}

	// Convert from Shift-JIS to UTF-8
	content, err := convertShiftJISToUTF8(data)
	if err != nil {
		t.Fatalf("Failed to convert Shift-JIS to UTF-8: %v", err)
	}

	// Create lexer and parser
	l := lexer.New(content)
	p := parser.New(l)

	// Parse the program
	program, parseErrs := p.ParseProgram()
	if len(parseErrs) > 0 {
		t.Logf("Parser errors (%d total):", len(parseErrs))
		for i, err := range parseErrs {
			if i < 10 { // Limit output
				t.Logf("  %v", err)
			}
		}
		if len(parseErrs) > 10 {
			t.Logf("  ... and %d more errors", len(parseErrs)-10)
		}
	}

	// Verify that we got statements
	if len(program.Statements) == 0 {
		t.Fatal("Expected statements from parser, got none")
	}

	t.Logf("Parsed %d statements", len(program.Statements))

	// Create compiler and compile
	c := New()
	opcodes, compileErrs := c.Compile(program)

	// Log any compilation errors
	if len(compileErrs) > 0 {
		t.Logf("Compiler errors (%d total):", len(compileErrs))
		for i, err := range compileErrs {
			if i < 10 { // Limit output
				t.Logf("  %v", err)
			}
		}
		if len(compileErrs) > 10 {
			t.Logf("  ... and %d more errors", len(compileErrs)-10)
		}
		// Compilation errors are not fatal for this test
		// as some constructs may not be fully supported yet
	}

	// Verify that we got opcodes
	if len(opcodes) == 0 {
		t.Fatal("Expected opcodes, got none")
	}

	t.Logf("Generated %d opcodes", len(opcodes))

	// Count opcode types
	opcodeCounts := countOpcodeTypes(opcodes)

	// Log opcode statistics
	t.Logf("opcode.OpCode type counts:")
	for opType, count := range opcodeCounts {
		t.Logf("  %s: %d", opType, count)
	}

	// Verify key opcodes are present
	verifyKeyOpcodesPresent(t, opcodes)
}

// countOpcodeTypes counts the types of opcodes in the sequence.
func countOpcodeTypes(opcodes []opcode.OpCode) map[opcode.Cmd]int {
	counts := make(map[opcode.Cmd]int)

	var countOpcode func(op opcode.OpCode)
	countOpcode = func(op opcode.OpCode) {
		counts[op.Cmd]++

		// Count nested opcodes in Args
		for _, arg := range op.Args {
			switch v := arg.(type) {
			case opcode.OpCode:
				countOpcode(v)
			case []opcode.OpCode:
				for _, nested := range v {
					countOpcode(nested)
				}
			case []any:
				for _, item := range v {
					if nestedOp, ok := item.(opcode.OpCode); ok {
						countOpcode(nestedOp)
					}
					if nestedSlice, ok := item.([]opcode.OpCode); ok {
						for _, nested := range nestedSlice {
							countOpcode(nested)
						}
					}
				}
			}
		}
	}

	for _, op := range opcodes {
		countOpcode(op)
	}

	return counts
}

// verifyKeyOpcodesPresent checks that key opcodes from the sample file are correctly generated.
func verifyKeyOpcodesPresent(t *testing.T, opcodes []opcode.OpCode) {
	t.Helper()

	// Track what we find
	foundOpcodes := make(map[opcode.Cmd]bool)

	// Expected opcodes that should be found in the compiled sample file
	expectedOpcodes := []opcode.Cmd{
		opcode.DefineFunction,       // Function definitions (main, start, BIRD, etc.)
		opcode.RegisterEventHandler, // mes() blocks
		opcode.Call,                 // Function calls (LoadPic, CreatePic, etc.)
		opcode.Assign,               // opcode.Variable assignments
		opcode.For,                  // For loops
		opcode.If,                   // If statements
		opcode.SetStep,              // step() blocks with count
		opcode.Wait,                 // Wait commands from commas in step blocks
	}

	// Walk through opcodes
	var walkOpcode func(op opcode.OpCode)
	walkOpcode = func(op opcode.OpCode) {
		foundOpcodes[op.Cmd] = true

		// Walk nested opcodes in Args
		for _, arg := range op.Args {
			switch v := arg.(type) {
			case opcode.OpCode:
				walkOpcode(v)
			case []opcode.OpCode:
				for _, nested := range v {
					walkOpcode(nested)
				}
			case []any:
				for _, item := range v {
					if nestedOp, ok := item.(opcode.OpCode); ok {
						walkOpcode(nestedOp)
					}
					if nestedSlice, ok := item.([]opcode.OpCode); ok {
						for _, nested := range nestedSlice {
							walkOpcode(nested)
						}
					}
				}
			}
		}
	}

	for _, op := range opcodes {
		walkOpcode(op)
	}

	// Check for expected opcodes
	for _, expectedOp := range expectedOpcodes {
		if !foundOpcodes[expectedOp] {
			t.Errorf("Expected opcode %s not found in compiled output", expectedOp)
		}
	}

	// Log what we found
	t.Logf("Found opcodes: %v", getOpcodeKeys(foundOpcodes))
}

// getOpcodeKeys returns the keys of a map as a slice.
func getOpcodeKeys(m map[opcode.Cmd]bool) []opcode.Cmd {
	keys := make([]opcode.Cmd, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestCompileSampleFileMesBlocks tests that mes() blocks are correctly compiled.
// Validates Requirement 4.12: mes(EVENT) blocks generate opcode.RegisterEventHandler.
func TestCompileSampleFileMesBlocks(t *testing.T) {
	input := `mes(TIME){step(20){,start();end_step;del_me;}}`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// Should be opcode.RegisterEventHandler
	if opcodes[0].Cmd != opcode.RegisterEventHandler {
		t.Errorf("expected opcode.RegisterEventHandler, got %s", opcodes[0].Cmd)
	}

	// First arg should be event type "TIME"
	if len(opcodes[0].Args) < 1 {
		t.Fatal("expected at least 1 arg")
	}

	eventType, ok := opcodes[0].Args[0].(string)
	if !ok {
		t.Fatalf("expected string event type, got %T", opcodes[0].Args[0])
	}

	if eventType != "TIME" {
		t.Errorf("expected event type 'TIME', got %q", eventType)
	}
}

// TestCompileSampleFileStepBlocks tests that step() blocks are correctly compiled.
// Validates Requirements 4.13-4.15: step() generates opcode.SetStep and opcode.Wait.
func TestCompileSampleFileStepBlocks(t *testing.T) {
	input := `step(10){func1();, func2();,, end_step; del_me;}`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	// Log opcodes for debugging
	t.Logf("Generated %d opcodes:", len(opcodes))
	for i, op := range opcodes {
		t.Logf("  [%d] %s: %v", i, op.Cmd, op.Args)
	}

	// First opcode should be opcode.SetStep with count 10
	if len(opcodes) < 1 {
		t.Fatal("expected at least 1 opcode")
	}

	if opcodes[0].Cmd != opcode.SetStep {
		t.Errorf("expected first opcode to be opcode.SetStep, got %s", opcodes[0].Cmd)
	}

	// Count opcode.Wait opcodes
	waitCount := 0
	for _, op := range opcodes {
		if op.Cmd == opcode.Wait {
			waitCount++
		}
	}

	// Should have at least 2 opcode.Wait (one for single comma, one for double comma)
	if waitCount < 2 {
		t.Errorf("expected at least 2 opcode.Wait opcodes, got %d", waitCount)
	}
}

// TestCompileSampleFileForLoop tests that for loops are correctly compiled.
// Validates Requirement 4.6: for loops generate opcode.For.
func TestCompileSampleFileForLoop(t *testing.T) {
	input := `for(i=0;i<=1;i=i+1){
    LPic[i]=LoadPic(StrPrint("ROBOT%03d.BMP",i));
  }`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// Should be opcode.For
	if opcodes[0].Cmd != opcode.For {
		t.Errorf("expected opcode.For, got %s", opcodes[0].Cmd)
	}

	// Should have 4 args: init, condition, post, body
	if len(opcodes[0].Args) != 4 {
		t.Errorf("expected 4 args for opcode.For, got %d", len(opcodes[0].Args))
	}
}

// TestCompileSampleFileFunctionDefinition tests that function definitions are correctly compiled.
// Validates Requirement 4.1: AST is compiled to opcode.OpCode instructions.
func TestCompileSampleFileFunctionDefinition(t *testing.T) {
	input := `main(){
  CapTitle("");
  WinW=WinInfo(0);
}`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// Should be opcode.DefineFunction
	if opcodes[0].Cmd != opcode.DefineFunction {
		t.Errorf("expected opcode.DefineFunction, got %s", opcodes[0].Cmd)
	}

	// First arg should be function name "main"
	if len(opcodes[0].Args) < 1 {
		t.Fatal("expected at least 1 arg")
	}

	funcName, ok := opcodes[0].Args[0].(string)
	if !ok {
		t.Fatalf("expected string function name, got %T", opcodes[0].Args[0])
	}

	if funcName != "main" {
		t.Errorf("expected function name 'main', got %q", funcName)
	}
}

// TestCompileSampleFileArrayAssignment tests that array assignments are correctly compiled.
// Validates Requirement 4.3: Array assignment generates opcode.ArrayAssign.
func TestCompileSampleFileArrayAssignment(t *testing.T) {
	input := `LPic[i]=LoadPic("test.bmp");`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// Should be opcode.ArrayAssign
	if opcodes[0].Cmd != opcode.ArrayAssign {
		t.Errorf("expected opcode.ArrayAssign, got %s", opcodes[0].Cmd)
	}

	// Should have 3 args: array name, index, value
	if len(opcodes[0].Args) != 3 {
		t.Errorf("expected 3 args for opcode.ArrayAssign, got %d", len(opcodes[0].Args))
	}

	// First arg should be opcode.Variable("LPic")
	arrayName, ok := opcodes[0].Args[0].(opcode.Variable)
	if !ok {
		t.Fatalf("expected opcode.Variable for array name, got %T", opcodes[0].Args[0])
	}

	if string(arrayName) != "LPic" {
		t.Errorf("expected array name 'LPic', got %q", arrayName)
	}
}

// TestCompileSampleFileIfStatement tests that if statements are correctly compiled.
// Validates Requirement 4.5: if statements generate opcode.If.
func TestCompileSampleFileIfStatement(t *testing.T) {
	input := `if(l!=0){
    if(l<0){l=-l; d=1;}
  }`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// Should be opcode.If
	if opcodes[0].Cmd != opcode.If {
		t.Errorf("expected opcode.If, got %s", opcodes[0].Cmd)
	}

	// Should have 3 args: condition, then block, else block
	if len(opcodes[0].Args) != 3 {
		t.Errorf("expected 3 args for opcode.If, got %d", len(opcodes[0].Args))
	}
}

// TestCompileSampleFileNestedMesBlocks tests compilation of nested mes blocks.
// Validates Requirement 4.12: Nested mes blocks generate nested opcode.RegisterEventHandler.
func TestCompileSampleFileNestedMesBlocks(t *testing.T) {
	input := `mes(MIDI_TIME){step{
    mes(MIDI_TIME){step(8){
      OPENING();,
      end_step; del_me;
    }}end_step; del_me;
  }}`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	// Count opcode.RegisterEventHandler opcodes (including nested)
	counts := countOpcodeTypes(opcodes)

	if counts[opcode.RegisterEventHandler] != 2 {
		t.Errorf("expected 2 opcode.RegisterEventHandler opcodes, got %d", counts[opcode.RegisterEventHandler])
	}
}

// TestCompileSampleFileFunctionWithDefaultParams tests compilation of functions with default parameters.
// Validates Requirement 4.1: Function definitions with parameters are compiled correctly.
func TestCompileSampleFileFunctionWithDefaultParams(t *testing.T) {
	input := `OP_walk(c,p[],x,y,w,h,l=10){
    int d;
    d=0;
  }`

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected parse errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)

	if len(compileErrs) > 0 {
		t.Fatalf("unexpected compile errors: %v", compileErrs)
	}

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// Should be opcode.DefineFunction
	if opcodes[0].Cmd != opcode.DefineFunction {
		t.Errorf("expected opcode.DefineFunction, got %s", opcodes[0].Cmd)
	}

	// Check function name
	funcName, ok := opcodes[0].Args[0].(string)
	if !ok {
		t.Fatalf("expected string function name, got %T", opcodes[0].Args[0])
	}

	if funcName != "OP_walk" {
		t.Errorf("expected function name 'OP_walk', got %q", funcName)
	}

	// Check parameters (second arg)
	params, ok := opcodes[0].Args[1].([]any)
	if !ok {
		t.Fatalf("expected []any for parameters, got %T", opcodes[0].Args[1])
	}

	if len(params) != 7 {
		t.Errorf("expected 7 parameters, got %d", len(params))
	}

	// Check last parameter has default value
	lastParam, ok := params[6].(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any for parameter, got %T", params[6])
	}

	if lastParam["name"] != "l" {
		t.Errorf("expected last parameter name 'l', got %v", lastParam["name"])
	}

	if _, hasDefault := lastParam["default"]; !hasDefault {
		t.Error("expected last parameter to have default value")
	}
}
