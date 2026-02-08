package parser

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
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

// TestSampleFileROBOTParser tests the Parser with the sample file samples/robot/ROBOT.TFY.
// This test validates that the Parser can correctly parse a real FILLY script.
// Validates Requirements 3.1-3.16: Complete syntax analysis of FILLY scripts.
func TestSampleFileROBOTParser(t *testing.T) {
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
	p := New(l)

	// Parse the program
	program, errs := p.ParseProgram()

	// Log any errors (some errors are expected due to # directives)
	if len(errs) > 0 {
		t.Logf("Parser errors (%d total):", len(errs))
		for i, err := range errs {
			if i < 20 { // Limit output
				t.Logf("  %v", err)
			}
		}
		if len(errs) > 20 {
			t.Logf("  ... and %d more errors", len(errs)-20)
		}
	}

	// Verify that we got statements
	if len(program.Statements) == 0 {
		t.Fatal("Expected statements, got none")
	}

	t.Logf("Total statements: %d", len(program.Statements))

	// Count statement types
	stmtCounts := countStatementTypes(program.Statements)

	// Log statement statistics
	t.Logf("Statement type counts:")
	for stmtType, count := range stmtCounts {
		t.Logf("  %s: %d", stmtType, count)
	}

	// Verify key AST nodes are present
	verifyKeyASTNodesPresent(t, program)
}

// countStatementTypes counts the types of statements in the program.
func countStatementTypes(stmts []Statement) map[string]int {
	counts := make(map[string]int)

	var countStmt func(stmt Statement)
	countStmt = func(stmt Statement) {
		switch s := stmt.(type) {
		case *VarDeclaration:
			counts["VarDeclaration"]++
		case *FunctionStatement:
			counts["FunctionStatement"]++
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					countStmt(bodyStmt)
				}
			}
		case *AssignStatement:
			counts["AssignStatement"]++
		case *ExpressionStatement:
			counts["ExpressionStatement"]++
		case *IfStatement:
			counts["IfStatement"]++
			if s.Consequence != nil {
				for _, bodyStmt := range s.Consequence.Statements {
					countStmt(bodyStmt)
				}
			}
			if s.Alternative != nil {
				countStmt(s.Alternative)
			}
		case *ForStatement:
			counts["ForStatement"]++
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					countStmt(bodyStmt)
				}
			}
		case *WhileStatement:
			counts["WhileStatement"]++
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					countStmt(bodyStmt)
				}
			}
		case *SwitchStatement:
			counts["SwitchStatement"]++
		case *MesStatement:
			counts["MesStatement"]++
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					countStmt(bodyStmt)
				}
			}
		case *StepStatement:
			counts["StepStatement"]++
			if s.Body != nil {
				for _, cmd := range s.Body.Commands {
					if cmd.Statement != nil {
						countStmt(cmd.Statement)
					}
				}
			}
		case *BreakStatement:
			counts["BreakStatement"]++
		case *ContinueStatement:
			counts["ContinueStatement"]++
		case *ReturnStatement:
			counts["ReturnStatement"]++
		case *BlockStatement:
			counts["BlockStatement"]++
			for _, bodyStmt := range s.Statements {
				countStmt(bodyStmt)
			}
		default:
			counts["Unknown"]++
		}
	}

	for _, stmt := range stmts {
		countStmt(stmt)
	}

	return counts
}

// verifyKeyASTNodesPresent checks that key AST nodes from the sample file are correctly parsed.
func verifyKeyASTNodesPresent(t *testing.T, program *Program) {
	t.Helper()

	// Track what we find
	foundFunctions := make(map[string]bool)
	foundMesEvents := make(map[string]bool)
	foundVarDeclarations := false
	foundForLoops := false
	foundIfStatements := false
	foundStepStatements := false

	// Expected functions from ROBOT.TFY
	expectedFunctions := []string{
		"main",
		"start",
		"OP_walk",
		"BIRD",
		"OPENON",
		"OPENING",
	}

	// Expected mes event types
	expectedMesEvents := []string{
		"TIME",
		"MIDI_TIME",
		"MIDI_END",
	}

	// Walk through statements
	var walkStmt func(stmt Statement)
	walkStmt = func(stmt Statement) {
		switch s := stmt.(type) {
		case *VarDeclaration:
			foundVarDeclarations = true
		case *FunctionStatement:
			foundFunctions[s.Name] = true
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					walkStmt(bodyStmt)
				}
			}
		case *IfStatement:
			foundIfStatements = true
			if s.Consequence != nil {
				for _, bodyStmt := range s.Consequence.Statements {
					walkStmt(bodyStmt)
				}
			}
			if s.Alternative != nil {
				walkStmt(s.Alternative)
			}
		case *ForStatement:
			foundForLoops = true
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					walkStmt(bodyStmt)
				}
			}
		case *MesStatement:
			foundMesEvents[s.EventType] = true
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					walkStmt(bodyStmt)
				}
			}
		case *StepStatement:
			foundStepStatements = true
			if s.Body != nil {
				for _, cmd := range s.Body.Commands {
					if cmd.Statement != nil {
						walkStmt(cmd.Statement)
					}
				}
			}
		case *BlockStatement:
			for _, bodyStmt := range s.Statements {
				walkStmt(bodyStmt)
			}
		}
	}

	for _, stmt := range program.Statements {
		walkStmt(stmt)
	}

	// Check for expected functions
	for _, funcName := range expectedFunctions {
		if !foundFunctions[funcName] {
			t.Errorf("Expected function %q not found in AST", funcName)
		}
	}

	// Check for expected mes events
	for _, eventType := range expectedMesEvents {
		if !foundMesEvents[eventType] {
			t.Errorf("Expected mes event type %q not found in AST", eventType)
		}
	}

	// Check for key statement types
	if !foundVarDeclarations {
		t.Error("Expected VarDeclaration statements not found")
	}
	if !foundForLoops {
		t.Error("Expected ForStatement not found")
	}
	if !foundIfStatements {
		t.Error("Expected IfStatement not found")
	}
	if !foundStepStatements {
		t.Error("Expected StepStatement not found")
	}

	// Log what we found
	t.Logf("Found functions: %v", getKeys(foundFunctions))
	t.Logf("Found mes events: %v", getKeys(foundMesEvents))
}

// getKeys returns the keys of a map as a slice.
func getKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TestSampleFileMesBlockParsing tests parsing of mes() blocks from the sample file.
// Validates Requirement 3.11: mes(EVENT) blocks with event type and body.
// Validates Requirement 9.1: EVENT types recognition.
func TestSampleFileMesBlockParsing(t *testing.T) {
	input := `mes(TIME){step(20){,start();end_step;del_me;}}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	mesStmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if mesStmt.EventType != "TIME" {
		t.Errorf("expected event type 'TIME', got %q", mesStmt.EventType)
	}

	if mesStmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Body should contain a step statement
	if len(mesStmt.Body.Statements) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(mesStmt.Body.Statements))
	}

	stepStmt, ok := mesStmt.Body.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement in mes body, got %T", mesStmt.Body.Statements[0])
	}

	// Check step count
	if stepStmt.Count == nil {
		t.Fatal("expected step count, got nil")
	}

	countLit, ok := stepStmt.Count.(*IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for step count, got %T", stepStmt.Count)
	}

	if countLit.Value != 20 {
		t.Errorf("expected step count 20, got %d", countLit.Value)
	}
}

// TestSampleFileStepBlockParsing tests parsing of step blocks with commas.
// Validates Requirement 3.12: step() statements with count and body.
// Validates Requirement 9.2, 9.3: Commas in step blocks are wait instructions.
func TestSampleFileStepBlockParsing(t *testing.T) {
	input := `step{BIRD();,,,, OPENON();,,,, end_step; del_me;}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	stepStmt, ok := program.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", program.Statements[0])
	}

	// No count specified
	if stepStmt.Count != nil {
		t.Errorf("expected nil count, got %v", stepStmt.Count)
	}

	if stepStmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Log commands for debugging
	t.Logf("Step body has %d commands", len(stepStmt.Body.Commands))
	for i, cmd := range stepStmt.Body.Commands {
		if cmd.Statement != nil {
			t.Logf("  Command[%d]: Statement=%T, WaitCount=%d", i, cmd.Statement, cmd.WaitCount)
		} else {
			t.Logf("  Command[%d]: Statement=nil (wait-only), WaitCount=%d", i, cmd.WaitCount)
		}
	}

	// Verify we have commands with wait counts
	totalWaits := 0
	for _, cmd := range stepStmt.Body.Commands {
		totalWaits += cmd.WaitCount
	}

	if totalWaits == 0 {
		t.Error("expected wait counts from commas, got 0")
	}
}

// TestSampleFileForLoopParsing tests parsing of for loops from the sample file.
// Validates Requirement 3.8: For loops with init, condition, post, and body.
func TestSampleFileForLoopParsing(t *testing.T) {
	input := `for(i=0;i<=1;i=i+1){
    LPic[i]=LoadPic(StrPrint("ROBOT%03d.BMP",i));
  }`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	forStmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	// Check init
	if forStmt.Init == nil {
		t.Error("expected Init, got nil")
	}

	// Check condition
	if forStmt.Condition == nil {
		t.Error("expected Condition, got nil")
	}

	binExpr, ok := forStmt.Condition.(*BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression for condition, got %T", forStmt.Condition)
	}

	if binExpr.Operator != "<=" {
		t.Errorf("expected operator '<=', got %q", binExpr.Operator)
	}

	// Check post
	if forStmt.Post == nil {
		t.Error("expected Post, got nil")
	}

	// Check body
	if forStmt.Body == nil {
		t.Error("expected Body, got nil")
	}

	if len(forStmt.Body.Statements) != 1 {
		t.Errorf("expected 1 body statement, got %d", len(forStmt.Body.Statements))
	}
}

// TestSampleFileFunctionDefinitionParsing tests parsing of function definitions.
// Validates Requirement 3.4: Function definitions (name(params){body}).
// Validates Requirement 9.6, 9.7: Parameters with default values and array parameters.
func TestSampleFileFunctionDefinitionParsing(t *testing.T) {
	input := `OP_walk(c,p[],x,y,w,h,l=10){
    int d;
    d=0;
    if(l!=0){
      if(l<0){l=-l;d=1;}
    }
  }`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	funcStmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if funcStmt.Name != "OP_walk" {
		t.Errorf("expected function name 'OP_walk', got %q", funcStmt.Name)
	}

	// Check parameters
	if len(funcStmt.Parameters) != 7 {
		t.Fatalf("expected 7 parameters, got %d", len(funcStmt.Parameters))
	}

	// Check array parameter p[]
	if !funcStmt.Parameters[1].IsArray {
		t.Error("expected parameter 'p' to be an array")
	}

	// Check default value parameter l=10
	lastParam := funcStmt.Parameters[6]
	if lastParam.Name != "l" {
		t.Errorf("expected last parameter name 'l', got %q", lastParam.Name)
	}
	if lastParam.DefaultValue == nil {
		t.Error("expected default value for parameter 'l'")
	} else {
		intLit, ok := lastParam.DefaultValue.(*IntegerLiteral)
		if !ok {
			t.Errorf("expected IntegerLiteral for default value, got %T", lastParam.DefaultValue)
		} else if intLit.Value != 10 {
			t.Errorf("expected default value 10, got %d", intLit.Value)
		}
	}

	// Check body
	if funcStmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}
}

// TestSampleFileVarDeclarationParsing tests parsing of variable declarations.
// Validates Requirement 3.2: Variable declarations (int x, y[]; str s;).
// Validates Requirement 3.3: Array declarations (int arr[10]).
func TestSampleFileVarDeclarationParsing(t *testing.T) {
	input := `int LPic[],BasePic,FieldPic,BirdPic[],OPPic[],Dummy;`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	varDecl, ok := program.Statements[0].(*VarDeclaration)
	if !ok {
		t.Fatalf("expected VarDeclaration, got %T", program.Statements[0])
	}

	if varDecl.Type != "int" {
		t.Errorf("expected type 'int', got %q", varDecl.Type)
	}

	expectedNames := []string{"LPic", "BasePic", "FieldPic", "BirdPic", "OPPic", "Dummy"}
	expectedIsArray := []bool{true, false, false, true, true, false}

	if len(varDecl.Names) != len(expectedNames) {
		t.Fatalf("expected %d names, got %d", len(expectedNames), len(varDecl.Names))
	}

	for i, name := range expectedNames {
		if varDecl.Names[i] != name {
			t.Errorf("name[%d]: expected %q, got %q", i, name, varDecl.Names[i])
		}
		if varDecl.IsArray[i] != expectedIsArray[i] {
			t.Errorf("IsArray[%d]: expected %v, got %v", i, expectedIsArray[i], varDecl.IsArray[i])
		}
	}
}

// TestSampleFileNestedMesBlocks tests parsing of nested mes blocks.
// Validates Requirement 3.11: mes(EVENT) blocks with event type and body.
func TestSampleFileNestedMesBlocks(t *testing.T) {
	input := `mes(MIDI_TIME){step{
    mes(MIDI_TIME){step(8){
      OPENING();,
      end_step; del_me;
    }}end_step; del_me;
  }}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	mesStmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if mesStmt.EventType != "MIDI_TIME" {
		t.Errorf("expected event type 'MIDI_TIME', got %q", mesStmt.EventType)
	}

	// Count nested mes statements
	mesCount := 0
	var countMes func(stmt Statement)
	countMes = func(stmt Statement) {
		switch s := stmt.(type) {
		case *MesStatement:
			mesCount++
			if s.Body != nil {
				for _, bodyStmt := range s.Body.Statements {
					countMes(bodyStmt)
				}
			}
		case *StepStatement:
			if s.Body != nil {
				for _, cmd := range s.Body.Commands {
					if cmd.Statement != nil {
						countMes(cmd.Statement)
					}
				}
			}
		case *BlockStatement:
			for _, bodyStmt := range s.Statements {
				countMes(bodyStmt)
			}
		}
	}

	countMes(mesStmt)

	if mesCount != 2 {
		t.Errorf("expected 2 mes statements (including nested), got %d", mesCount)
	}
}

// TestSampleFileMainFunction tests parsing of the main() function pattern.
func TestSampleFileMainFunction(t *testing.T) {
	input := `main(){
  CapTitle("");
  WinW=WinInfo(0); WinH=WinInfo(1);
}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	funcStmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if funcStmt.Name != "main" {
		t.Errorf("expected function name 'main', got %q", funcStmt.Name)
	}

	if len(funcStmt.Parameters) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(funcStmt.Parameters))
	}

	if funcStmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Body should have statements
	if len(funcStmt.Body.Statements) < 2 {
		t.Errorf("expected at least 2 body statements, got %d", len(funcStmt.Body.Statements))
	}
}

// TestSampleFileHexColorLiterals tests parsing of hex color values.
// Validates Requirement 2.4: Hexadecimal integers with 0x prefix.
func TestSampleFileHexColorLiterals(t *testing.T) {
	input := `OpenWin(BirdPic[0],0,0,WinW,WinH,WinX,WinY,0xffffff);`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	callExpr, ok := exprStmt.Expression.(*CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", exprStmt.Expression)
	}

	if callExpr.Function != "OpenWin" {
		t.Errorf("expected function name 'OpenWin', got %q", callExpr.Function)
	}

	// Last argument should be hex color 0xffffff
	if len(callExpr.Arguments) < 8 {
		t.Fatalf("expected at least 8 arguments, got %d", len(callExpr.Arguments))
	}

	lastArg := callExpr.Arguments[7]
	intLit, ok := lastArg.(*IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for last argument, got %T", lastArg)
	}

	// 0xffffff = 16777215
	if intLit.Value != 16777215 {
		t.Errorf("expected hex value 16777215 (0xffffff), got %d", intLit.Value)
	}
}

// TestSampleFileArrayAssignment tests parsing of array assignments.
// Validates Requirement 3.6: Array assignment (arr[i] = value).
func TestSampleFileArrayAssignment(t *testing.T) {
	input := `LPic[i]=LoadPic(StrPrint("ROBOT%03d.BMP",i));`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	assignStmt, ok := program.Statements[0].(*AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", program.Statements[0])
	}

	// Name should be IndexExpression
	indexExpr, ok := assignStmt.Name.(*IndexExpression)
	if !ok {
		t.Fatalf("expected IndexExpression for Name, got %T", assignStmt.Name)
	}

	// Left should be identifier "LPic"
	ident, ok := indexExpr.Left.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier for Left, got %T", indexExpr.Left)
	}

	if ident.Value != "LPic" {
		t.Errorf("expected 'LPic', got %q", ident.Value)
	}

	// Value should be CallExpression
	callExpr, ok := assignStmt.Value.(*CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression for Value, got %T", assignStmt.Value)
	}

	if callExpr.Function != "LoadPic" {
		t.Errorf("expected function name 'LoadPic', got %q", callExpr.Function)
	}
}
