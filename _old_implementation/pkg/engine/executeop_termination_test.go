package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestExecuteOp_TerminationCheck verifies that ExecuteOp checks the
// programTerminated flag before executing any OpCode and returns
// ebiten.Termination when the flag is set.
// This validates task 3.2 from user-input-handling spec
func TestExecuteOp_TerminationCheck(t *testing.T) {
	ResetEngineForTest()

	// Create a test sequencer
	seq := &Sequencer{
		commands:     []interpreter.OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}

	// Test various OpCodes to ensure termination check works for all
	testCases := []struct {
		name   string
		opCode OpCode
	}{
		{
			name: "OpWait",
			opCode: OpCode{
				Cmd:  interpreter.OpWait,
				Args: []any{1},
			},
		},
		{
			name: "OpAssign",
			opCode: OpCode{
				Cmd:  interpreter.OpAssign,
				Args: []any{"testvar", 42},
			},
		},
		{
			name: "OpLiteral",
			opCode: OpCode{
				Cmd:  interpreter.OpLiteral,
				Args: []any{100},
			},
		},
		{
			name: "OpVarRef",
			opCode: OpCode{
				Cmd:  interpreter.OpVarRef,
				Args: []any{"testvar"},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset termination flag
			programTerminated = false

			// Test 1: Normal execution (programTerminated = false)
			result, yield := ExecuteOp(tc.opCode, seq)
			if result == ebiten.Termination {
				t.Errorf("%s: ExecuteOp should not return ebiten.Termination when programTerminated is false", tc.name)
			}

			// Test 2: With termination flag set (programTerminated = true)
			programTerminated = true
			result, yield = ExecuteOp(tc.opCode, seq)

			// Verify ExecuteOp returns ebiten.Termination
			if result != ebiten.Termination {
				t.Errorf("%s: ExecuteOp should return ebiten.Termination when programTerminated is true, got: %v", tc.name, result)
			}

			// Verify yield is false (not yielding, just terminating)
			if yield {
				t.Errorf("%s: ExecuteOp should return yield=false when terminating, got: %v", tc.name, yield)
			}
		})
	}

	// Cleanup
	ResetEngineForTest()
}

// TestExecuteOp_TerminationPropagation verifies that when ExecuteOp returns
// ebiten.Termination, it propagates correctly through UpdateVM
func TestExecuteOp_TerminationPropagation(t *testing.T) {
	ResetEngineForTest()

	// Register a sequence with multiple operations
	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 0}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 2}},
	}

	vmLock.Lock()
	mainSequencer = &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}
	sequencers = []*Sequencer{mainSequencer}
	vmLock.Unlock()

	// Execute first operation (should succeed)
	programTerminated = false
	UpdateVM(1)

	// Verify first operation executed
	// Note: PC advances to 2 because both Assign and Wait execute in the same tick
	// (Wait yields but PC has already advanced)
	vmLock.Lock()
	firstPC := mainSequencer.pc
	if firstPC < 1 {
		t.Errorf("At least one operation should have executed, PC should be >= 1, got: %d", firstPC)
	}
	vmLock.Unlock()

	// Set termination flag before next operation
	programTerminated = true

	// Try to execute next operation (should be blocked by termination check)
	UpdateVM(2)

	// Verify sequence was marked inactive
	vmLock.Lock()
	if mainSequencer.active {
		t.Error("Sequence should be marked inactive after termination")
	}
	vmLock.Unlock()

	// Verify counter variable is still 0 (second assignment didn't execute)
	if globalVars != nil {
		if counter, ok := globalVars["counter"]; ok {
			if counterInt, ok := counter.(int); ok && counterInt != 0 {
				t.Errorf("Counter should still be 0 (second assignment blocked), got: %d", counterInt)
			}
		}
	}

	// Cleanup
	ResetEngineForTest()
}

// TestExecuteOp_TerminationBeforeExpensiveOperation verifies that
// termination check happens before expensive operations execute
func TestExecuteOp_TerminationBeforeExpensiveOperation(t *testing.T) {
	ResetEngineForTest()

	// Create a sequencer
	seq := &Sequencer{
		commands:     []interpreter.OpCode{},
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}

	// Set termination flag BEFORE calling ExecuteOp
	programTerminated = true

	// Try to execute an operation that would normally do work
	// (e.g., LoadPic, CreatePic, etc.)
	op := OpCode{
		Cmd:  interpreter.OpLoadPic,
		Args: []any{"test.bmp"},
	}

	result, yield := ExecuteOp(op, seq)

	// Verify ExecuteOp returned immediately with Termination
	if result != ebiten.Termination {
		t.Errorf("ExecuteOp should return ebiten.Termination before executing LoadPic, got: %v", result)
	}

	if yield {
		t.Errorf("ExecuteOp should return yield=false when terminating, got: %v", yield)
	}

	// Verify no side effects occurred (no picture was loaded)
	// This is implicit - if LoadPic had executed, it would have tried to load a file
	// and likely failed or created a picture entry

	// Cleanup
	ResetEngineForTest()
}
