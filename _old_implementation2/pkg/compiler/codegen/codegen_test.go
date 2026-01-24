package codegen

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

func TestGenerateAssignment(t *testing.T) {
	input := `x = 5`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpAssign {
		t.Errorf("expected OpAssign, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 2 {
		t.Fatalf("expected 2 args, got %d", len(opcodes[0].Args))
	}

	varName, ok := opcodes[0].Args[0].(interpreter.Variable)
	if !ok {
		t.Errorf("expected Variable, got %T", opcodes[0].Args[0])
	}
	if string(varName) != "x" {
		t.Errorf("expected variable 'x', got '%s'", varName)
	}

	value, ok := opcodes[0].Args[1].(int64)
	if !ok {
		t.Errorf("expected int64, got %T", opcodes[0].Args[1])
	}
	if value != 5 {
		t.Errorf("expected value 5, got %d", value)
	}
}

func TestGenerateBinaryOp(t *testing.T) {
	input := `x = 5 + 3`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpAssign {
		t.Errorf("expected OpAssign, got %s", opcodes[0].Cmd)
	}

	// Check the value is a binary operation
	binOp, ok := opcodes[0].Args[1].(interpreter.OpCode)
	if !ok {
		t.Fatalf("expected OpCode for binary op, got %T", opcodes[0].Args[1])
	}

	if binOp.Cmd != interpreter.OpBinaryOp {
		t.Errorf("expected OpBinaryOp, got %s", binOp.Cmd)
	}

	if len(binOp.Args) != 3 {
		t.Fatalf("expected 3 args for binary op, got %d", len(binOp.Args))
	}

	operator, ok := binOp.Args[0].(string)
	if !ok || operator != "+" {
		t.Errorf("expected operator '+', got %v", binOp.Args[0])
	}
}

func TestGenerateIfStatement(t *testing.T) {
	input := `if (x > 5) { y = 10 }`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpIf {
		t.Errorf("expected OpIf, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 3 {
		t.Fatalf("expected 3 args (condition, consequence, alternative), got %d", len(opcodes[0].Args))
	}

	// Check condition is a binary op
	condition, ok := opcodes[0].Args[0].(interpreter.OpCode)
	if !ok {
		t.Errorf("expected OpCode for condition, got %T", opcodes[0].Args[0])
	}
	if condition.Cmd != interpreter.OpBinaryOp {
		t.Errorf("expected OpBinaryOp for condition, got %s", condition.Cmd)
	}

	// Check consequence is a slice of opcodes
	consequence, ok := opcodes[0].Args[1].([]interpreter.OpCode)
	if !ok {
		t.Errorf("expected []OpCode for consequence, got %T", opcodes[0].Args[1])
	}
	if len(consequence) != 1 {
		t.Errorf("expected 1 statement in consequence, got %d", len(consequence))
	}
}

func TestGenerateForLoop(t *testing.T) {
	input := `for (i = 0; i < 10; i = i + 1) { x = x + 1 }`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpFor {
		t.Errorf("expected OpFor, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 4 {
		t.Fatalf("expected 4 args (init, condition, post, body), got %d", len(opcodes[0].Args))
	}
}

func TestGenerateWhileLoop(t *testing.T) {
	input := `while (x < 10) { x = x + 1 }`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpWhile {
		t.Errorf("expected OpWhile, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 2 {
		t.Fatalf("expected 2 args (condition, body), got %d", len(opcodes[0].Args))
	}
}

func TestGenerateFunctionCall(t *testing.T) {
	input := `LoadPic(1, "test.bmp")`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpCall {
		t.Errorf("expected OpCall, got %s", opcodes[0].Cmd)
	}

	// The call expression itself is wrapped in OpCall
	callExpr, ok := opcodes[0].Args[0].(interpreter.OpCode)
	if !ok {
		t.Fatalf("expected OpCode for call expression, got %T", opcodes[0].Args[0])
	}

	if callExpr.Cmd != interpreter.OpCall {
		t.Errorf("expected OpCall for function, got %s", callExpr.Cmd)
	}

	// Check function name
	funcName, ok := callExpr.Args[0].(interpreter.Variable)
	if !ok {
		t.Errorf("expected Variable for function name, got %T", callExpr.Args[0])
	}
	if string(funcName) != "LoadPic" {
		t.Errorf("expected function name 'LoadPic', got '%s'", funcName)
	}

	// Check arguments
	if len(callExpr.Args) != 3 { // function name + 2 args
		t.Errorf("expected 3 items (name + 2 args), got %d", len(callExpr.Args))
	}
}

func TestGenerateArrayAccess(t *testing.T) {
	input := `x = arr[5]`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpAssign {
		t.Errorf("expected OpAssign, got %s", opcodes[0].Cmd)
	}

	// Check the value is an array access
	arrayAccess, ok := opcodes[0].Args[1].(interpreter.OpCode)
	if !ok {
		t.Fatalf("expected OpCode for array access, got %T", opcodes[0].Args[1])
	}

	if arrayAccess.Cmd != interpreter.OpArrayAccess {
		t.Errorf("expected OpArrayAccess, got %s", arrayAccess.Cmd)
	}

	if len(arrayAccess.Args) != 2 {
		t.Fatalf("expected 2 args (array, index), got %d", len(arrayAccess.Args))
	}
}

func TestGenerateArrayAssignment(t *testing.T) {
	input := `arr[5] = 10`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpArrayAssign {
		t.Errorf("expected OpArrayAssign, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 3 {
		t.Fatalf("expected 3 args (array, index, value), got %d", len(opcodes[0].Args))
	}
}

func TestGenerateMesStatement(t *testing.T) {
	input := `mes(TIME) { step(10) }`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpRegisterEventHandler {
		t.Errorf("expected OpRegisterEventHandler, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 2 {
		t.Fatalf("expected 2 args (event type, body), got %d", len(opcodes[0].Args))
	}

	eventType, ok := opcodes[0].Args[0].(string)
	if !ok {
		t.Errorf("expected string for event type, got %T", opcodes[0].Args[0])
	}
	if eventType != "TIME" {
		t.Errorf("expected event type 'TIME', got '%s'", eventType)
	}

	body, ok := opcodes[0].Args[1].([]interpreter.OpCode)
	if !ok {
		t.Errorf("expected []OpCode for body, got %T", opcodes[0].Args[1])
	}
	if len(body) != 1 {
		t.Errorf("expected 1 statement in body, got %d", len(body))
	}
}

func TestGenerateStepStatement(t *testing.T) {
	input := `step(10)`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	if opcodes[0].Cmd != interpreter.OpWait {
		t.Errorf("expected OpWait, got %s", opcodes[0].Cmd)
	}

	if len(opcodes[0].Args) != 1 {
		t.Fatalf("expected 1 arg (count), got %d", len(opcodes[0].Args))
	}

	count, ok := opcodes[0].Args[0].(int64)
	if !ok {
		t.Errorf("expected int64 for count, got %T", opcodes[0].Args[0])
	}
	if count != 10 {
		t.Errorf("expected count 10, got %d", count)
	}
}

func TestGenerateNestedExpressions(t *testing.T) {
	input := `x = (5 + 3) * 2`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 1 {
		t.Fatalf("expected 1 opcode, got %d", len(opcodes))
	}

	// The value should be a binary op (multiplication)
	multOp, ok := opcodes[0].Args[1].(interpreter.OpCode)
	if !ok {
		t.Fatalf("expected OpCode for multiplication, got %T", opcodes[0].Args[1])
	}

	if multOp.Cmd != interpreter.OpBinaryOp {
		t.Errorf("expected OpBinaryOp, got %s", multOp.Cmd)
	}

	// The left side should be another binary op (addition)
	addOp, ok := multOp.Args[1].(interpreter.OpCode)
	if !ok {
		t.Fatalf("expected OpCode for addition, got %T", multOp.Args[1])
	}

	if addOp.Cmd != interpreter.OpBinaryOp {
		t.Errorf("expected OpBinaryOp for addition, got %s", addOp.Cmd)
	}
}

func TestGenerateMultipleStatements(t *testing.T) {
	input := `x = 5
y = 10
z = x + y`

	opcodes := generateFromInput(t, input)

	if len(opcodes) != 3 {
		t.Fatalf("expected 3 opcodes, got %d", len(opcodes))
	}

	for i, op := range opcodes {
		if op.Cmd != interpreter.OpAssign {
			t.Errorf("statement %d: expected OpAssign, got %s", i, op.Cmd)
		}
	}
}

// Helper function
func generateFromInput(t *testing.T, input string) []interpreter.OpCode {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) > 0 {
		t.Fatalf("parser errors: %v", p.Errors())
	}

	g := New()
	opcodes := g.Generate(program)

	if len(g.Errors()) > 0 {
		t.Fatalf("generator errors: %v", g.Errors())
	}

	return opcodes
}
