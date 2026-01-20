package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestUpdateVM_SingleSequence verifies that UpdateVM processes one tick for a single sequence.
func TestUpdateVM_SingleSequence(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("z"), int64(3)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	// Initially at pc=0
	if seq.GetPC() != 0 {
		t.Errorf("Expected initial PC=0, got %d", seq.GetPC())
	}

	// Execute one tick - should execute first command and advance PC
	err := engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1 after first tick, got %d", seq.GetPC())
	}

	if seq.GetVariable("x") != int64(1) {
		t.Errorf("Expected x=1, got %v", seq.GetVariable("x"))
	}

	// Execute second tick
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 after second tick, got %d", seq.GetPC())
	}

	if seq.GetVariable("y") != int64(2) {
		t.Errorf("Expected y=2, got %v", seq.GetVariable("y"))
	}

	// Execute third tick
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 3 {
		t.Errorf("Expected PC=3 after third tick, got %d", seq.GetPC())
	}

	if seq.GetVariable("z") != int64(3) {
		t.Errorf("Expected z=3, got %v", seq.GetVariable("z"))
	}

	// Sequence should be complete
	if !seq.IsComplete() {
		t.Errorf("Expected sequence to be complete")
	}
}

// TestUpdateVM_MultipleSequences verifies that UpdateVM processes all active sequences.
func TestUpdateVM_MultipleSequences(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("a"), int64(10)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("b"), int64(20)}},
	}

	opcodes2 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("c"), int64(30)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("d"), int64(40)}},
	}

	seq1 := NewSequencer(opcodes1, TIME, nil)
	seq2 := NewSequencer(opcodes2, TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)

	// Execute one tick - both sequences should advance
	err := engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq1.GetPC() != 1 {
		t.Errorf("Expected seq1 PC=1, got %d", seq1.GetPC())
	}

	if seq2.GetPC() != 1 {
		t.Errorf("Expected seq2 PC=1, got %d", seq2.GetPC())
	}

	if seq1.GetVariable("a") != int64(10) {
		t.Errorf("Expected a=10, got %v", seq1.GetVariable("a"))
	}

	if seq2.GetVariable("c") != int64(30) {
		t.Errorf("Expected c=30, got %v", seq2.GetVariable("c"))
	}
}

// TestUpdateVM_WaitCounter verifies that wait operations pause sequence execution.
func TestUpdateVM_WaitCounter(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(1)}}, // Wait 1 step = 3 ticks (TIME mode)
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	// Tick 1: Execute first command (x=1)
	err := engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1, got %d", seq.GetPC())
	}

	if seq.GetVariable("x") != int64(1) {
		t.Errorf("Expected x=1, got %v", seq.GetVariable("x"))
	}

	// Tick 2: Execute wait command, set waitCount=3 (1 step × 3 ticks/step)
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 after wait, got %d", seq.GetPC())
	}

	if !seq.IsWaiting() {
		t.Errorf("Expected sequence to be waiting")
	}

	// Tick 3: Decrement wait (3 -> 2), don't execute
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 during wait, got %d", seq.GetPC())
	}

	// Tick 4: Decrement wait (2 -> 1), don't execute
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 during wait, got %d", seq.GetPC())
	}

	// Tick 5: Decrement wait (1 -> 0), don't execute
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 during wait, got %d", seq.GetPC())
	}

	if seq.IsWaiting() {
		t.Errorf("Expected sequence to finish waiting")
	}

	// Tick 6: Execute next command (y=2)
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	if seq.GetPC() != 3 {
		t.Errorf("Expected PC=3 after wait, got %d", seq.GetPC())
	}

	if seq.GetVariable("y") != int64(2) {
		t.Errorf("Expected y=2, got %v", seq.GetVariable("y"))
	}
}

// TestUpdateVM_SequenceCompletion verifies that completed sequences are detected.
func TestUpdateVM_SequenceCompletion(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	// Execute the single command
	err := engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	// Sequence should be complete
	if !seq.IsComplete() {
		t.Errorf("Expected sequence to be complete")
	}

	// Should still be active (not automatically deactivated)
	if !seq.IsActive() {
		t.Errorf("Expected sequence to still be active")
	}

	// Next tick should not execute anything (PC beyond commands)
	err = engine.UpdateVM()
	if err != nil {
		t.Fatalf("UpdateVM failed: %v", err)
	}

	// PC should not advance beyond command count
	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1 (at end), got %d", seq.GetPC())
	}
}

// TestUpdateVM_IndependentExecution verifies that sequences execute independently.
func TestUpdateVM_IndependentExecution(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	// Sequence 1: No wait (TIME mode)
	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("a"), int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("b"), int64(2)}},
	}

	// Sequence 2: With wait (MIDI_TIME mode for simpler timing)
	opcodes2 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("c"), int64(3)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // 2 steps = 2 ticks (MIDI_TIME)
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("d"), int64(4)}},
	}

	seq1 := NewSequencer(opcodes1, TIME, nil)
	seq2 := NewSequencer(opcodes2, MIDI_TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)

	// Tick 1: Both execute first command
	engine.UpdateVM()

	if seq1.GetPC() != 1 || seq2.GetPC() != 1 {
		t.Errorf("Expected both at PC=1, got seq1=%d, seq2=%d", seq1.GetPC(), seq2.GetPC())
	}

	// Tick 2: seq1 executes second command, seq2 executes wait
	engine.UpdateVM()

	if seq1.GetPC() != 2 {
		t.Errorf("Expected seq1 PC=2, got %d", seq1.GetPC())
	}

	if seq2.GetPC() != 2 {
		t.Errorf("Expected seq2 PC=2, got %d", seq2.GetPC())
	}

	if !seq2.IsWaiting() {
		t.Errorf("Expected seq2 to be waiting")
	}

	// Tick 3: seq1 complete, seq2 waiting (2 -> 1)
	engine.UpdateVM()

	if seq1.GetPC() != 2 {
		t.Errorf("Expected seq1 PC=2 (complete), got %d", seq1.GetPC())
	}

	if seq2.GetPC() != 2 {
		t.Errorf("Expected seq2 PC=2 (waiting), got %d", seq2.GetPC())
	}

	// Tick 4: seq1 still complete, seq2 waiting (1 -> 0)
	engine.UpdateVM()

	if seq2.IsWaiting() {
		t.Errorf("Expected seq2 to finish waiting (waitCount should be 0)")
	}

	// Tick 5: seq1 still complete, seq2 executes third command
	engine.UpdateVM()

	if seq2.GetPC() != 3 {
		t.Errorf("Expected seq2 PC=3, got %d", seq2.GetPC())
	}

	if seq2.GetVariable("d") != int64(4) {
		t.Errorf("Expected d=4, got %v", seq2.GetVariable("d"))
	}
}

// TestUpdateVM_InactiveSequencesSkipped verifies that inactive sequences are not executed.
func TestUpdateVM_InactiveSequencesSkipped(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	seqID := engine.RegisterSequence(seq, 0)

	// Execute first command
	engine.UpdateVM()

	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1, got %d", seq.GetPC())
	}

	// Deactivate sequence
	engine.DeleteMe(seqID)

	if seq.IsActive() {
		t.Errorf("Expected sequence to be inactive")
	}

	// Execute another tick - should not advance PC
	engine.UpdateVM()

	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1 (unchanged), got %d", seq.GetPC())
	}

	// y should not be set
	if seq.GetVariable("y") != 0 {
		t.Errorf("Expected y=0 (not set), got %v", seq.GetVariable("y"))
	}
}
