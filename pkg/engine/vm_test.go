package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Helper function to create a test VM with engine
func newTestVM() (*VM, *Engine, *EngineState, *Logger) {
	state := NewEngineState(nil, nil, nil)
	logger := NewLogger(DebugLevelError)
	engine := NewEngine(nil, nil, nil)
	engine.state = state
	engine.logger = logger
	vm := NewVM(state, engine, logger)
	return vm, engine, state, logger
}

func TestExecuteOpAssign(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test: x = 5
	op := interpreter.OpCode{
		Cmd: interpreter.OpAssign,
		Args: []any{
			interpreter.Variable("x"),
			int64(5),
		},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify variable was set
	val := seq.GetVariable("x")
	if val != int64(5) {
		t.Errorf("Expected x=5, got %v", val)
	}
}

func TestExecuteOpAssignExpression(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test: x = 5 + 3
	op := interpreter.OpCode{
		Cmd: interpreter.OpAssign,
		Args: []any{
			interpreter.Variable("x"),
			interpreter.OpCode{
				Cmd: interpreter.OpBinaryOp,
				Args: []any{
					"+",
					int64(5),
					int64(3),
				},
			},
		},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify variable was set
	val := seq.GetVariable("x")
	if val != int64(8) {
		t.Errorf("Expected x=8, got %v", val)
	}
}

func TestExecuteOpIf(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seq.SetVariable("x", int64(10))

	// Test: if (x > 5) { y = 1 } else { y = 2 }
	op := interpreter.OpCode{
		Cmd: interpreter.OpIf,
		Args: []any{
			interpreter.OpCode{
				Cmd: interpreter.OpBinaryOp,
				Args: []any{
					">",
					interpreter.Variable("x"),
					int64(5),
				},
			},
			[]interpreter.OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("y"),
						int64(1),
					},
				},
			},
			[]interpreter.OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("y"),
						int64(2),
					},
				},
			},
		},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify y was set to 1 (then branch)
	val := seq.GetVariable("y")
	if val != int64(1) {
		t.Errorf("Expected y=1, got %v", val)
	}
}

func TestExecuteOpIfElse(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seq.SetVariable("x", int64(3))

	// Test: if (x > 5) { y = 1 } else { y = 2 }
	op := interpreter.OpCode{
		Cmd: interpreter.OpIf,
		Args: []any{
			interpreter.OpCode{
				Cmd: interpreter.OpBinaryOp,
				Args: []any{
					">",
					interpreter.Variable("x"),
					int64(5),
				},
			},
			[]interpreter.OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("y"),
						int64(1),
					},
				},
			},
			[]interpreter.OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("y"),
						int64(2),
					},
				},
			},
		},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify y was set to 2 (else branch)
	val := seq.GetVariable("y")
	if val != int64(2) {
		t.Errorf("Expected y=2, got %v", val)
	}
}

func TestExecuteOpFor(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test: for (i = 0; i < 5; i = i + 1) { sum = sum + i }
	op := interpreter.OpCode{
		Cmd: interpreter.OpFor,
		Args: []any{
			// Init: i = 0
			interpreter.OpCode{
				Cmd: interpreter.OpAssign,
				Args: []any{
					interpreter.Variable("i"),
					int64(0),
				},
			},
			// Condition: i < 5
			interpreter.OpCode{
				Cmd: interpreter.OpBinaryOp,
				Args: []any{
					"<",
					interpreter.Variable("i"),
					int64(5),
				},
			},
			// Increment: i = i + 1
			interpreter.OpCode{
				Cmd: interpreter.OpAssign,
				Args: []any{
					interpreter.Variable("i"),
					interpreter.OpCode{
						Cmd: interpreter.OpBinaryOp,
						Args: []any{
							"+",
							interpreter.Variable("i"),
							int64(1),
						},
					},
				},
			},
			// Body: sum = sum + i
			[]interpreter.OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("sum"),
						interpreter.OpCode{
							Cmd: interpreter.OpBinaryOp,
							Args: []any{
								"+",
								interpreter.Variable("sum"),
								interpreter.Variable("i"),
							},
						},
					},
				},
			},
		},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify sum = 0 + 1 + 2 + 3 + 4 = 10
	val := seq.GetVariable("sum")
	if val != int64(10) {
		t.Errorf("Expected sum=10, got %v", val)
	}
}

func TestExecuteOpWhile(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seq.SetVariable("i", int64(0))

	// Test: while (i < 5) { i = i + 1 }
	op := interpreter.OpCode{
		Cmd: interpreter.OpWhile,
		Args: []any{
			// Condition: i < 5
			interpreter.OpCode{
				Cmd: interpreter.OpBinaryOp,
				Args: []any{
					"<",
					interpreter.Variable("i"),
					int64(5),
				},
			},
			// Body: i = i + 1
			[]interpreter.OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("i"),
						interpreter.OpCode{
							Cmd: interpreter.OpBinaryOp,
							Args: []any{
								"+",
								interpreter.Variable("i"),
								int64(1),
							},
						},
					},
				},
			},
		},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify i = 5
	val := seq.GetVariable("i")
	if val != int64(5) {
		t.Errorf("Expected i=5, got %v", val)
	}
}

func TestExecuteOpWait(t *testing.T) {
	vm, _, _, _ := newTestVM()

	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test: wait(10)
	op := interpreter.OpCode{
		Cmd:  interpreter.OpWait,
		Args: []any{int64(10)},
	}

	err := vm.ExecuteOp(seq, op)
	if err != nil {
		t.Fatalf("ExecuteOp failed: %v", err)
	}

	// Verify wait counter was set
	if !seq.IsWaiting() {
		t.Error("Expected sequence to be waiting")
	}
}

func TestEvaluateBinaryOp(t *testing.T) {
	vm, _, _, _ := newTestVM()

	tests := []struct {
		name     string
		op       string
		left     any
		right    any
		expected any
	}{
		{"add", "+", int64(5), int64(3), int64(8)},
		{"subtract", "-", int64(5), int64(3), int64(2)},
		{"multiply", "*", int64(5), int64(3), int64(15)},
		{"divide", "/", int64(15), int64(3), int64(5)},
		{"modulo", "%", int64(17), int64(5), int64(2)},
		{"equal", "==", int64(5), int64(5), true},
		{"not equal", "!=", int64(5), int64(3), true},
		{"less than", "<", int64(3), int64(5), true},
		{"greater than", ">", int64(5), int64(3), true},
		{"less or equal", "<=", int64(5), int64(5), true},
		{"greater or equal", ">=", int64(5), int64(5), true},
		{"logical and", "&&", true, true, true},
		{"logical or", "||", false, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := vm.applyBinaryOp(tt.op, tt.left, tt.right)
			if err != nil {
				t.Fatalf("applyBinaryOp failed: %v", err)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToBool(t *testing.T) {
	vm, _, _, _ := newTestVM()

	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"int zero", int(0), false},
		{"int non-zero", int(5), true},
		{"int64 zero", int64(0), false},
		{"int64 non-zero", int64(5), true},
		{"float64 zero", float64(0.0), false},
		{"float64 non-zero", float64(3.14), true},
		{"string empty", "", false},
		{"string non-empty", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.toBool(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestToInt(t *testing.T) {
	vm, _, _, _ := newTestVM()

	tests := []struct {
		name     string
		value    any
		expected int64
	}{
		{"int", int(5), int64(5)},
		{"int64", int64(10), int64(10)},
		{"float64", float64(3.14), int64(3)},
		{"bool true", true, int64(1)},
		{"bool false", false, int64(0)},
		{"string", "hello", int64(0)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := vm.toInt(tt.value)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
