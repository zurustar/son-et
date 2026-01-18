package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestUpdateVM_TerminationCheck verifies that UpdateVM checks programTerminated
// and marks all active sequences as inactive when termination is requested
// This validates task 3.1 from user-input-handling spec
func TestUpdateVM_TerminationCheck(t *testing.T) {
	ResetEngineForTest()

	// Create a sequence
	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
	}

	vmLock.Lock()
	seq := &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}
	sequencers = []*Sequencer{seq}
	vmLock.Unlock()

	// First update - should work normally
	programTerminated = false
	UpdateVM(1)

	vmLock.Lock()
	if !seq.active {
		t.Error("Sequence should still be active after normal update")
	}
	vmLock.Unlock()

	// Set termination flag
	programTerminated = true

	// Second update - should mark sequence as inactive
	UpdateVM(2)

	vmLock.Lock()
	if seq.active {
		t.Error("Sequence should be inactive after termination")
	}
	vmLock.Unlock()

	// Cleanup
	ResetEngineForTest()
}

// TestUpdateVM_TerminationWithMultipleSequences verifies that UpdateVM
// marks ALL active sequences as inactive when termination is requested
func TestUpdateVM_TerminationWithMultipleSequences(t *testing.T) {
	ResetEngineForTest()

	// Create multiple sequences
	ops1 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{1}},
	}
	ops2 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{2}},
	}
	ops3 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{3}},
	}

	vmLock.Lock()
	seq1 := &Sequencer{
		commands:     ops1,
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}
	seq2 := &Sequencer{
		commands:     ops2,
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}
	seq3 := &Sequencer{
		commands:     ops3,
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}
	sequencers = []*Sequencer{seq1, seq2, seq3}
	vmLock.Unlock()

	// Set termination flag
	programTerminated = true

	// Update - should mark all sequences as inactive
	UpdateVM(1)

	vmLock.Lock()
	if seq1.active {
		t.Error("Sequence 1 should be inactive after termination")
	}
	if seq2.active {
		t.Error("Sequence 2 should be inactive after termination")
	}
	if seq3.active {
		t.Error("Sequence 3 should be inactive after termination")
	}
	vmLock.Unlock()

	// Cleanup
	ResetEngineForTest()
}

// TestUpdateVM_TerminationEarlyReturn verifies that UpdateVM returns
// immediately when programTerminated is set, without executing any OpCodes
func TestUpdateVM_TerminationEarlyReturn(t *testing.T) {
	ResetEngineForTest()

	// Create a sequence with multiple operations
	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"test_var", 1}},
		{Cmd: interpreter.OpAssign, Args: []any{"test_var", 2}},
		{Cmd: interpreter.OpAssign, Args: []any{"test_var", 3}},
	}

	vmLock.Lock()
	seq := &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true,
		ticksPerStep: 12,
		vars:         make(map[string]any),
		mode:         Time,
	}
	sequencers = []*Sequencer{seq}
	vmLock.Unlock()

	// Set termination flag BEFORE any execution
	programTerminated = true

	// Update - should return immediately without executing any OpCodes
	UpdateVM(1)

	vmLock.Lock()
	pc := seq.pc
	vmLock.Unlock()

	// PC should still be 0 (no OpCodes executed)
	if pc != 0 {
		t.Errorf("Sequencer PC should be 0 (no OpCodes executed), got: %d", pc)
	}

	// Cleanup
	ResetEngineForTest()
}
