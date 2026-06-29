package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestReturnInsideForPropagates is a regression test for the bug where a
// `return` inside a for/while loop body was swallowed: the loop kept iterating
// and the return value was overwritten by later iterations.
// See docs/bug-hunt-findings.md finding A.
func TestReturnInsideForPropagates(t *testing.T) {
	vm := New(nil)

	// body: if (i == 2) { return 99 }; hits[i] = i
	body := []opcode.OpCode{
		{Cmd: opcode.If, Args: []any{
			opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(2)}},
			[]opcode.OpCode{{Cmd: opcode.Call, Args: []any{"return", int64(99)}}},
			[]opcode.OpCode{},
		}},
		{Cmd: opcode.ArrayAssign, Args: []any{opcode.Variable("hits"), opcode.Variable("i"), opcode.Variable("i")}},
	}
	forOp := opcode.OpCode{Cmd: opcode.For, Args: []any{
		[]opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}}},
		opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
		[]opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"),
			opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}}}}},
		body,
	}}

	result, err := vm.Execute(forOp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rm, ok := result.(*returnMarker)
	if !ok {
		t.Fatalf("return inside for was not propagated: got %T = %v", result, result)
	}
	if rv, _ := toInt64(rm.value); rv != 99 {
		t.Errorf("return value = %v, want 99", rm.value)
	}

	// hits should only contain 0 and 1; iterations at i>=2 must not have run.
	if arrVal, ok := vm.GetGlobalScope().Get("hits"); ok {
		if arr, ok := arrVal.(*Array); ok {
			if arr.Len() > 2 {
				t.Errorf("loop kept running after return: hits=%v (len %d), want at most [0 1]", arr.ToSlice(), arr.Len())
			}
		}
	}
}

// TestReturnInsideWhilePropagates is the while-loop counterpart.
func TestReturnInsideWhilePropagates(t *testing.T) {
	vm := New(nil)
	vm.GetGlobalScope().Set("i", int64(0))

	whileOp := opcode.OpCode{Cmd: opcode.While, Args: []any{
		opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
		[]opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"),
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}}}},
			{Cmd: opcode.Call, Args: []any{"return", int64(7)}},
		},
	}}

	result, err := vm.Execute(whileOp)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result.(*returnMarker); !ok {
		t.Fatalf("return inside while was not propagated: got %T = %v", result, result)
	}
	// The loop must break out after the first iteration (i == 1), not run to 5.
	if iv, _ := vm.GetGlobalScope().Get("i"); func() bool { n, _ := toInt64(iv); return n != 1 }() {
		t.Errorf("while kept running after return: i=%v, want 1", iv)
	}
}
