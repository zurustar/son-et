package interpreter

import (
	"testing"
)

func TestOpCmd_String(t *testing.T) {
	tests := []struct {
		name string
		op   OpCmd
		want string
	}{
		{"OpAssign", OpAssign, "OpAssign"},
		{"OpLiteral", OpLiteral, "OpLiteral"},
		{"OpVariable", OpVariable, "OpVariable"},
		{"OpBinaryOp", OpBinaryOp, "OpBinaryOp"},
		{"OpIf", OpIf, "OpIf"},
		{"OpFor", OpFor, "OpFor"},
		{"OpCall", OpCall, "OpCall"},
		{"OpWait", OpWait, "OpWait"},
		{"OpRegisterSequence", OpRegisterSequence, "OpRegisterSequence"},
		{"OpArrayAccess", OpArrayAccess, "OpArrayAccess"},
		{"Unknown", OpCmd(999), "Unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.op.String(); got != tt.want {
				t.Errorf("OpCmd.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOpCode_Construction(t *testing.T) {
	t.Run("Simple OpCode", func(t *testing.T) {
		op := OpCode{
			Cmd:  OpLiteral,
			Args: []any{42},
		}

		if op.Cmd != OpLiteral {
			t.Errorf("Expected OpLiteral, got %v", op.Cmd)
		}
		if len(op.Args) != 1 {
			t.Errorf("Expected 1 arg, got %d", len(op.Args))
		}
		if op.Args[0] != 42 {
			t.Errorf("Expected arg 42, got %v", op.Args[0])
		}
	})

	t.Run("Nested OpCode", func(t *testing.T) {
		// Represents: x = 1 + 2
		op := OpCode{
			Cmd: OpAssign,
			Args: []any{
				Variable("x"),
				OpCode{
					Cmd: OpBinaryOp,
					Args: []any{
						"+",
						OpCode{Cmd: OpLiteral, Args: []any{1}},
						OpCode{Cmd: OpLiteral, Args: []any{2}},
					},
				},
			},
		}

		if op.Cmd != OpAssign {
			t.Errorf("Expected OpAssign, got %v", op.Cmd)
		}
		if len(op.Args) != 2 {
			t.Errorf("Expected 2 args, got %d", len(op.Args))
		}

		// Check variable
		if v, ok := op.Args[0].(Variable); !ok || v != "x" {
			t.Errorf("Expected Variable 'x', got %v", op.Args[0])
		}

		// Check nested OpCode
		if nested, ok := op.Args[1].(OpCode); !ok {
			t.Errorf("Expected nested OpCode, got %T", op.Args[1])
		} else if nested.Cmd != OpBinaryOp {
			t.Errorf("Expected OpBinaryOp, got %v", nested.Cmd)
		}
	})

	t.Run("OpCode with multiple nested levels", func(t *testing.T) {
		// Represents: if (x > 0) { y = x + 1 }
		op := OpCode{
			Cmd: OpIf,
			Args: []any{
				// Condition: x > 0
				OpCode{
					Cmd: OpBinaryOp,
					Args: []any{
						">",
						OpCode{Cmd: OpVariable, Args: []any{Variable("x")}},
						OpCode{Cmd: OpLiteral, Args: []any{0}},
					},
				},
				// Then block: y = x + 1
				[]OpCode{
					{
						Cmd: OpAssign,
						Args: []any{
							Variable("y"),
							OpCode{
								Cmd: OpBinaryOp,
								Args: []any{
									"+",
									OpCode{Cmd: OpVariable, Args: []any{Variable("x")}},
									OpCode{Cmd: OpLiteral, Args: []any{1}},
								},
							},
						},
					},
				},
			},
		}

		if op.Cmd != OpIf {
			t.Errorf("Expected OpIf, got %v", op.Cmd)
		}
		if len(op.Args) != 2 {
			t.Errorf("Expected 2 args (condition and then block), got %d", len(op.Args))
		}
	})
}

func TestVariable_String(t *testing.T) {
	v := Variable("testVar")
	if v.String() != "testVar" {
		t.Errorf("Expected 'testVar', got %v", v.String())
	}
}

func TestVariable_DistinguishFromLiteral(t *testing.T) {
	// Variables and string literals should be distinguishable
	varName := Variable("x")
	strLiteral := "x"

	// Type assertion should work for Variable
	if _, ok := any(varName).(Variable); !ok {
		t.Error("Variable type assertion failed")
	}

	// Type assertion should fail for string literal
	if _, ok := any(strLiteral).(Variable); ok {
		t.Error("String literal should not be a Variable type")
	}
}
