package interpreter

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

// Helper function to parse TFY code
func parseCode(t *testing.T, code string) *ast.Program {
	l := lexer.New(code)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("Parser errors: %v", p.Errors())
	}

	return program
}

// Test simple script conversion to OpCode
func TestSimpleScriptConversion(t *testing.T) {
	code := `
int x;
x = 10;
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check globals
	if _, ok := script.Globals["x"]; !ok {
		t.Errorf("Expected global variable 'x' to be tracked")
	}

	// Check main function body
	if len(script.Main.Body) == 0 {
		t.Errorf("Expected main function body to have OpCodes")
	}

	// Check assignment OpCode
	if script.Main.Body[0].Cmd != "Assign" {
		t.Errorf("Expected first OpCode to be Assign, got %s", script.Main.Body[0].Cmd)
	}
}

// Test function definition conversion
func TestFunctionDefinitionConversion(t *testing.T) {
	code := `
getValue() {
	int result;
	result = 42;
	return result;
}
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check function was registered
	if _, ok := script.Functions["getvalue"]; !ok {
		t.Errorf("Expected function 'getvalue' to be registered")
	}

	fn := script.Functions["getvalue"]

	// Check local variables
	if _, ok := fn.Locals["result"]; !ok {
		t.Errorf("Expected local variable 'result' to be tracked")
	}

	// Check function body
	if len(fn.Body) == 0 {
		t.Errorf("Expected function body to have OpCodes")
	}
}

// Test asset discovery
func TestAssetDiscovery(t *testing.T) {
	code := `
main() {
	LoadPic("image1.bmp");
	PlayMIDI("music.mid");
	PlayWAVE("sound.wav");
}
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check assets
	expectedAssets := map[string]bool{
		"image1.bmp": false,
		"music.mid":  false,
		"sound.wav":  false,
	}

	for _, asset := range script.Assets {
		if _, ok := expectedAssets[asset]; ok {
			expectedAssets[asset] = true
		}
	}

	for asset, found := range expectedAssets {
		if !found {
			t.Errorf("Expected asset '%s' to be discovered", asset)
		}
	}
}

// Test variable scope tracking
func TestVariableScopeTracking(t *testing.T) {
	code := `
int globalVar;

myFunc(int param) {
	int localVar;
	localVar = param + 1;
}

main() {
	globalVar = 10;
	myFunc(globalVar);
}
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check global variable
	if _, ok := script.Globals["globalvar"]; !ok {
		t.Errorf("Expected global variable 'globalvar' to be tracked")
	}

	// Check function parameter and local variable
	fn := script.Functions["myfunc"]
	if fn == nil {
		t.Fatalf("Expected function 'myfunc' to exist")
	}

	if len(fn.Parameters) != 1 {
		t.Errorf("Expected 1 parameter, got %d", len(fn.Parameters))
	}

	if fn.Parameters[0].Name != "param" {
		t.Errorf("Expected parameter name 'param', got '%s'", fn.Parameters[0].Name)
	}

	if _, ok := fn.Locals["localvar"]; !ok {
		t.Errorf("Expected local variable 'localvar' to be tracked")
	}
}

// Test case-insensitive variable names
func TestCaseInsensitiveVariables(t *testing.T) {
	code := `
int MyVar;
MyVar = 10;
myvar = 20;
MYVAR = 30;
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// All should map to the same variable (case-insensitive)
	if _, ok := script.Globals["myvar"]; !ok {
		t.Errorf("Expected global variable 'myvar' (normalized) to be tracked")
	}

	// Check that all assignments reference the same variable
	for _, op := range script.Main.Body {
		if op.Cmd == "Assign" {
			if len(op.Args) > 0 {
				if varRef, ok := op.Args[0].(Variable); ok {
					if string(varRef) != "myvar" {
						t.Errorf("Expected variable reference to be 'myvar', got '%s'", varRef)
					}
				}
			}
		}
	}
}

// Test expression conversion
func TestExpressionConversion(t *testing.T) {
	code := `
int x;
x = 10 + 20 * 3;
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check that the expression was converted to OpCode
	if len(script.Main.Body) == 0 {
		t.Fatalf("Expected main function body to have OpCodes")
	}

	assignOp := script.Main.Body[0]
	if assignOp.Cmd != "Assign" {
		t.Errorf("Expected Assign OpCode, got %s", assignOp.Cmd)
	}

	// The value should be an OpCode representing the expression
	if len(assignOp.Args) < 2 {
		t.Fatalf("Expected Assign to have 2 arguments")
	}

	// Check that the expression is an OpCode (not a literal)
	if exprOp, ok := assignOp.Args[1].(OpCode); ok {
		if exprOp.Cmd != "+" {
			t.Errorf("Expected expression root to be '+', got '%s'", exprOp.Cmd)
		}
	} else {
		t.Errorf("Expected expression to be OpCode, got %T", assignOp.Args[1])
	}
}

// Test control flow conversion
func TestControlFlowConversion(t *testing.T) {
	code := `
int x;
if (x > 10) {
	x = 20;
} else {
	x = 5;
}
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check that if statement was converted
	if len(script.Main.Body) == 0 {
		t.Fatalf("Expected main function body to have OpCodes")
	}

	ifOp := script.Main.Body[0]
	if ifOp.Cmd != "If" {
		t.Errorf("Expected If OpCode, got %s", ifOp.Cmd)
	}

	// Check that if has 3 arguments: condition, then, else
	if len(ifOp.Args) != 3 {
		t.Errorf("Expected If to have 3 arguments, got %d", len(ifOp.Args))
	}
}

// Test mes() block conversion
func TestMesBlockConversion(t *testing.T) {
	code := `
mes(TIME) {
	Wait(10);
}
`
	program := parseCode(t, code)
	interp := NewInterpreter()

	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Interpret error: %v", err)
	}

	// Check that mes block was converted to RegisterSequence
	if len(script.Main.Body) == 0 {
		t.Fatalf("Expected main function body to have OpCodes")
	}

	mesOp := script.Main.Body[0]
	if mesOp.Cmd != "RegisterSequence" {
		t.Errorf("Expected RegisterSequence OpCode, got %s", mesOp.Cmd)
	}

	// Check that mes has 2 arguments: mode, body
	if len(mesOp.Args) != 2 {
		t.Errorf("Expected RegisterSequence to have 2 arguments, got %d", len(mesOp.Args))
	}
}
