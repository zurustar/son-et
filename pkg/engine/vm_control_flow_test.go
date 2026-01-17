package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestVMIfStatement tests if-else execution in VM mode
func TestVMIfStatement(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: if (5 > 3) { x = 10 } else { x = 20 }
	ifOp := OpCode{
		Cmd: interpreter.OpIf,
		Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{">", 5, 3}},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"x", 10}},
			},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"x", 20}},
			},
		},
	}

	_, yield := ExecuteOp(ifOp, seq)
	if yield {
		t.Error("If statement should not yield")
	}

	if seq.vars["x"] != 10 {
		t.Errorf("Expected x=10, got x=%v", seq.vars["x"])
	}

	// Test: if (3 > 5) { x = 10 } else { x = 20 }
	seq.vars = make(map[string]any)
	ifOp2 := OpCode{
		Cmd: interpreter.OpIf,
		Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{">", 3, 5}},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"x", 10}},
			},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"x", 20}},
			},
		},
	}

	_, yield = ExecuteOp(ifOp2, seq)
	if yield {
		t.Error("If statement should not yield")
	}

	if seq.vars["x"] != 20 {
		t.Errorf("Expected x=20, got x=%v", seq.vars["x"])
	}
}

// TestVMForLoop tests for loop execution in VM mode
func TestVMForLoop(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: for (i = 0; i < 5; i = i + 1) { sum = sum + i }
	forOp := OpCode{
		Cmd: interpreter.OpFor,
		Args: []any{
			OpCode{Cmd: interpreter.OpAssign, Args: []any{"i", 0}},                                                                    // init
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", Variable("i"), 5}},                                                      // condition
			OpCode{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("i"), 1}}}}, // post
			[]OpCode{ // body
				{Cmd: interpreter.OpAssign, Args: []any{"sum", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("sum"), Variable("i")}}}},
			},
		},
	}

	seq.vars["sum"] = 0
	_, yield := ExecuteOp(forOp, seq)
	if yield {
		t.Error("For loop should not yield")
	}

	// sum should be 0+1+2+3+4 = 10
	if seq.vars["sum"] != 10 {
		t.Errorf("Expected sum=10, got sum=%v", seq.vars["sum"])
	}
}

// TestVMWhileLoop tests while loop execution in VM mode
func TestVMWhileLoop(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: while (i < 5) { sum = sum + i; i = i + 1 }
	seq.vars["i"] = 0
	seq.vars["sum"] = 0

	whileOp := OpCode{
		Cmd: interpreter.OpWhile,
		Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", Variable("i"), 5}}, // condition
			[]OpCode{ // body
				{Cmd: interpreter.OpAssign, Args: []any{"sum", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("sum"), Variable("i")}}}},
				{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("i"), 1}}}},
			},
		},
	}

	_, yield := ExecuteOp(whileOp, seq)
	if yield {
		t.Error("While loop should not yield")
	}

	// sum should be 0+1+2+3+4 = 10
	if seq.vars["sum"] != 10 {
		t.Errorf("Expected sum=10, got sum=%v", seq.vars["sum"])
	}
}

// TestVMSwitchStatement tests switch-case execution in VM mode
func TestVMSwitchStatement(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: switch (x) { case 1: y = 10; break; case 2: y = 20; break; default: y = 30 }
	seq.vars["x"] = 2

	switchOp := OpCode{
		Cmd: interpreter.OpSwitch,
		Args: []any{
			Variable("x"), // value
			[]any{ // cases
				[]any{1, []OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 10}}, {Cmd: interpreter.OpBreak, Args: []any{}}}},
				[]any{2, []OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 20}}, {Cmd: interpreter.OpBreak, Args: []any{}}}},
			},
			[]OpCode{ // default
				{Cmd: interpreter.OpAssign, Args: []any{"y", 30}},
			},
		},
	}

	_, yield := ExecuteOp(switchOp, seq)
	if yield {
		t.Error("Switch statement should not yield")
	}

	if seq.vars["y"] != 20 {
		t.Errorf("Expected y=20, got y=%v", seq.vars["y"])
	}

	// Test default case
	seq.vars["x"] = 99
	seq.vars["y"] = 0

	_, yield = ExecuteOp(switchOp, seq)
	if yield {
		t.Error("Switch statement should not yield")
	}

	if seq.vars["y"] != 30 {
		t.Errorf("Expected y=30 (default), got y=%v", seq.vars["y"])
	}
}

// TestVMBreakContinue tests break and continue in loops
func TestVMBreakContinue(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test break: for (i = 0; i < 10; i = i + 1) { if (i == 5) break; sum = sum + i }
	seq.vars["sum"] = 0

	forOp := OpCode{
		Cmd: interpreter.OpFor,
		Args: []any{
			OpCode{Cmd: interpreter.OpAssign, Args: []any{"i", 0}},                                                                    // init
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", Variable("i"), 10}},                                                     // condition
			OpCode{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("i"), 1}}}}, // post
			[]OpCode{ // body
				{Cmd: interpreter.OpIf, Args: []any{
					OpCode{Cmd: interpreter.OpInfix, Args: []any{"==", Variable("i"), 5}},
					[]OpCode{{Cmd: interpreter.OpBreak, Args: []any{}}},
				}},
				{Cmd: interpreter.OpAssign, Args: []any{"sum", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("sum"), Variable("i")}}}},
			},
		},
	}

	_, yield := ExecuteOp(forOp, seq)
	if yield {
		t.Error("For loop with break should not yield")
	}

	// sum should be 0+1+2+3+4 = 10 (stops at i=5)
	if seq.vars["sum"] != 10 {
		t.Errorf("Expected sum=10 (with break), got sum=%v", seq.vars["sum"])
	}
}

// TestVMNestedIfStatement tests nested if-else execution in VM mode (Requirement 20.3)
func TestVMNestedIfStatement(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: if (x > 5) { if (x > 10) { y = 1 } else { y = 2 } } else { y = 3 }
	seq.vars["x"] = 8

	nestedIfOp := OpCode{
		Cmd: interpreter.OpIf,
		Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{">", Variable("x"), 5}},
			[]OpCode{
				{Cmd: interpreter.OpIf, Args: []any{
					OpCode{Cmd: interpreter.OpInfix, Args: []any{">", Variable("x"), 10}},
					[]OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 1}}},
					[]OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 2}}},
				}},
			},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"y", 3}},
			},
		},
	}

	_, yield := ExecuteOp(nestedIfOp, seq)
	if yield {
		t.Error("Nested if statement should not yield")
	}

	// x=8: outer true (8>5), inner false (8<=10), so y=2
	if seq.vars["y"] != 2 {
		t.Errorf("Expected y=2, got y=%v", seq.vars["y"])
	}

	// Test with x=15
	seq.vars["x"] = 15
	seq.vars["y"] = 0
	_, yield = ExecuteOp(nestedIfOp, seq)
	if yield {
		t.Error("Nested if statement should not yield")
	}

	// x=15: outer true (15>5), inner true (15>10), so y=1
	if seq.vars["y"] != 1 {
		t.Errorf("Expected y=1, got y=%v", seq.vars["y"])
	}
}

// TestVMDoWhileLoop tests do-while loop execution in VM mode (Requirement 21.3)
func TestVMDoWhileLoop(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: do { sum = sum + i; i = i + 1 } while (i < 5)
	seq.vars["i"] = 0
	seq.vars["sum"] = 0

	doWhileOp := OpCode{
		Cmd: interpreter.OpDoWhile,
		Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", Variable("i"), 5}}, // condition
			[]OpCode{ // body
				{Cmd: interpreter.OpAssign, Args: []any{"sum", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("sum"), Variable("i")}}}},
				{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("i"), 1}}}},
			},
		},
	}

	_, yield := ExecuteOp(doWhileOp, seq)
	if yield {
		t.Error("Do-while loop should not yield")
	}

	// sum should be 0+1+2+3+4 = 10
	if seq.vars["sum"] != 10 {
		t.Errorf("Expected sum=10, got sum=%v", seq.vars["sum"])
	}

	// Test that do-while executes at least once even when condition is false
	seq.vars["i"] = 10
	seq.vars["count"] = 0

	doWhileOp2 := OpCode{
		Cmd: interpreter.OpDoWhile,
		Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", Variable("i"), 5}}, // false from start
			[]OpCode{ // body
				{Cmd: interpreter.OpAssign, Args: []any{"count", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("count"), 1}}}},
			},
		},
	}

	_, yield = ExecuteOp(doWhileOp2, seq)
	if yield {
		t.Error("Do-while loop should not yield")
	}

	// count should be 1 (executed once despite false condition)
	if seq.vars["count"] != 1 {
		t.Errorf("Expected count=1 (do-while executes at least once), got count=%v", seq.vars["count"])
	}
}

// TestVMSwitchMultipleCases tests switch with multiple cases and default (Requirements 22.2, 22.3)
func TestVMSwitchMultipleCases(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test: switch (x) { case 1: y = 10; break; case 2: y = 20; break; case 3: y = 30; break; default: y = 99 }
	switchOp := OpCode{
		Cmd: interpreter.OpSwitch,
		Args: []any{
			Variable("x"), // value
			[]any{ // cases
				[]any{1, []OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 10}}, {Cmd: interpreter.OpBreak, Args: []any{}}}},
				[]any{2, []OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 20}}, {Cmd: interpreter.OpBreak, Args: []any{}}}},
				[]any{3, []OpCode{{Cmd: interpreter.OpAssign, Args: []any{"y", 30}}, {Cmd: interpreter.OpBreak, Args: []any{}}}},
			},
			[]OpCode{ // default
				{Cmd: interpreter.OpAssign, Args: []any{"y", 99}},
			},
		},
	}

	// Test case 1
	seq.vars["x"] = 1
	_, yield := ExecuteOp(switchOp, seq)
	if yield {
		t.Error("Switch statement should not yield")
	}
	if seq.vars["y"] != 10 {
		t.Errorf("Expected y=10 for case 1, got y=%v", seq.vars["y"])
	}

	// Test case 3
	seq.vars["x"] = 3
	seq.vars["y"] = 0
	_, yield = ExecuteOp(switchOp, seq)
	if yield {
		t.Error("Switch statement should not yield")
	}
	if seq.vars["y"] != 30 {
		t.Errorf("Expected y=30 for case 3, got y=%v", seq.vars["y"])
	}

	// Test default case
	seq.vars["x"] = 999
	seq.vars["y"] = 0
	_, yield = ExecuteOp(switchOp, seq)
	if yield {
		t.Error("Switch statement should not yield")
	}
	if seq.vars["y"] != 99 {
		t.Errorf("Expected y=99 for default case, got y=%v", seq.vars["y"])
	}
}

// TestVMContinueInLoop tests continue statement in loops (Requirement 21.5)
func TestVMContinueInLoop(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	// Test continue: for (i = 0; i < 10; i = i + 1) { if (i % 2 == 0) continue; sum = sum + i }
	seq.vars["sum"] = 0

	forOp := OpCode{
		Cmd: interpreter.OpFor,
		Args: []any{
			OpCode{Cmd: interpreter.OpAssign, Args: []any{"i", 0}},                                                                    // init
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", Variable("i"), 10}},                                                     // condition
			OpCode{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("i"), 1}}}}, // post
			[]OpCode{ // body
				{Cmd: interpreter.OpIf, Args: []any{
					OpCode{Cmd: interpreter.OpInfix, Args: []any{"==", OpCode{Cmd: interpreter.OpInfix, Args: []any{"%", Variable("i"), 2}}, 0}},
					[]OpCode{{Cmd: interpreter.OpContinue, Args: []any{}}},
				}},
				{Cmd: interpreter.OpAssign, Args: []any{"sum", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", Variable("sum"), Variable("i")}}}},
			},
		},
	}

	_, yield := ExecuteOp(forOp, seq)
	if yield {
		t.Error("For loop with continue should not yield")
	}

	// sum should be 1+3+5+7+9 = 25 (only odd numbers)
	if seq.vars["sum"] != 25 {
		t.Errorf("Expected sum=25 (with continue skipping even numbers), got sum=%v", seq.vars["sum"])
	}
}

// TestVMInfixOperations tests infix expression evaluation
func TestVMInfixOperations(t *testing.T) {
	seq := &Sequencer{
		commands:     []OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         0,
	}

	tests := []struct {
		name     string
		op       string
		left     any
		right    any
		expected any
	}{
		{"Addition", "+", 5, 3, 8},
		{"Subtraction", "-", 5, 3, 2},
		{"Multiplication", "*", 5, 3, 15},
		{"Division", "/", 15, 3, 5},
		{"Modulo", "%", 17, 5, 2},
		{"Equal", "==", 5, 5, true},
		{"NotEqual", "!=", 5, 3, true},
		{"LessThan", "<", 3, 5, true},
		{"GreaterThan", ">", 5, 3, true},
		{"LessOrEqual", "<=", 3, 3, true},
		{"GreaterOrEqual", ">=", 5, 5, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			infixOp := OpCode{
				Cmd:  interpreter.OpInfix,
				Args: []any{tt.op, tt.left, tt.right},
			}

			result, _ := ExecuteOp(infixOp, seq)
			if result != tt.expected {
				t.Errorf("Expected %v %s %v = %v, got %v", tt.left, tt.op, tt.right, tt.expected, result)
			}
		})
	}
}
