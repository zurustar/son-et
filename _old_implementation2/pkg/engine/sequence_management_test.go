package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestRegisterSequence(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	// Create a sequence
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}
	seq := NewSequencer(opcodes, TIME, nil)

	// Register sequence
	seqID := engine.RegisterSequence(seq, 0)

	// Verify sequence was registered
	if seqID != 1 {
		t.Errorf("Expected sequence ID 1, got %d", seqID)
	}

	// Verify sequence is in state
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Errorf("Expected 1 sequencer, got %d", len(sequencers))
	}

	if sequencers[0].GetID() != seqID {
		t.Errorf("Expected sequencer ID %d, got %d", seqID, sequencers[0].GetID())
	}
}

func TestRegisterMultipleSequences(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	// Register multiple sequences
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, MIDI_TIME, nil)
	seq3 := NewSequencer(opcodes, TIME, nil)

	id1 := engine.RegisterSequence(seq1, 0)
	id2 := engine.RegisterSequence(seq2, 0)
	id3 := engine.RegisterSequence(seq3, 0)

	// Verify IDs are unique and sequential
	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Errorf("Expected IDs 1, 2, 3, got %d, %d, %d", id1, id2, id3)
	}

	// Verify all sequences are registered
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 3 {
		t.Errorf("Expected 3 sequencers, got %d", len(sequencers))
	}
}

func TestRegisterSequenceWithGroupID(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register sequences with specific group ID
	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, TIME, nil)

	groupID := 100
	engine.RegisterSequence(seq1, groupID)
	engine.RegisterSequence(seq2, groupID)

	// Verify both sequences have the same group ID
	sequencers := engine.GetState().GetSequencers()
	if sequencers[0].GetGroupID() != groupID {
		t.Errorf("Expected group ID %d, got %d", groupID, sequencers[0].GetGroupID())
	}
	if sequencers[1].GetGroupID() != groupID {
		t.Errorf("Expected group ID %d, got %d", groupID, sequencers[1].GetGroupID())
	}
}

func TestDeleteMe(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	seqID := engine.RegisterSequence(seq, 0)

	// Verify sequence is active
	if !seq.IsActive() {
		t.Error("Expected sequence to be active")
	}

	// Deactivate sequence
	engine.DeleteMe(seqID)

	// Verify sequence is inactive
	if seq.IsActive() {
		t.Error("Expected sequence to be inactive")
	}
}

func TestDeleteUs(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register sequences in the same group
	groupID := 100
	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, TIME, nil)
	seq3 := NewSequencer(opcodes, TIME, nil)

	engine.RegisterSequence(seq1, groupID)
	engine.RegisterSequence(seq2, groupID)
	engine.RegisterSequence(seq3, 200) // Different group

	// Deactivate group
	engine.DeleteUs(groupID)

	// Verify sequences in group are inactive
	if seq1.IsActive() {
		t.Error("Expected seq1 to be inactive")
	}
	if seq2.IsActive() {
		t.Error("Expected seq2 to be inactive")
	}

	// Verify sequence in different group is still active
	if !seq3.IsActive() {
		t.Error("Expected seq3 to be active")
	}
}

func TestDeleteAll(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register multiple sequences
	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, MIDI_TIME, nil)
	seq3 := NewSequencer(opcodes, TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)
	engine.RegisterSequence(seq3, 0)

	// Deactivate all sequences
	engine.DeleteAll()

	// Verify all sequences are inactive
	if seq1.IsActive() {
		t.Error("Expected seq1 to be inactive")
	}
	if seq2.IsActive() {
		t.Error("Expected seq2 to be inactive")
	}
	if seq3.IsActive() {
		t.Error("Expected seq3 to be inactive")
	}
}

func TestCleanupSequences(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register sequences
	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, TIME, nil)
	seq3 := NewSequencer(opcodes, TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)
	engine.RegisterSequence(seq3, 0)

	// Deactivate some sequences
	seq1.Deactivate()
	seq3.Deactivate()

	// Verify 3 sequences before cleanup
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 3 {
		t.Errorf("Expected 3 sequencers before cleanup, got %d", len(sequencers))
	}

	// Cleanup inactive sequences
	engine.CleanupSequences()

	// Verify only active sequence remains
	sequencers = engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Errorf("Expected 1 sequencer after cleanup, got %d", len(sequencers))
	}

	if sequencers[0] != seq2 {
		t.Error("Expected seq2 to remain after cleanup")
	}
}

func TestSequenceIsolation(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register multiple sequences
	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, TIME, nil)
	seq3 := NewSequencer(opcodes, TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)
	engine.RegisterSequence(seq3, 0)

	// Deactivate one sequence
	seq2.Deactivate()

	// Verify other sequences are unaffected
	if !seq1.IsActive() {
		t.Error("Expected seq1 to remain active")
	}
	if !seq3.IsActive() {
		t.Error("Expected seq3 to remain active")
	}
}

func TestGroupIDAllocation(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	// Allocate multiple group IDs
	id1 := state.AllocateGroupID()
	id2 := state.AllocateGroupID()
	id3 := state.AllocateGroupID()

	// Verify IDs are unique and sequential
	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Errorf("Expected group IDs 1, 2, 3, got %d, %d, %d", id1, id2, id3)
	}
}

func TestSequenceLifecycle(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Create and register sequence
	seq := NewSequencer(opcodes, TIME, nil)
	seqID := engine.RegisterSequence(seq, 0)

	// Verify initial state
	if !seq.IsActive() {
		t.Error("Expected sequence to be active after registration")
	}

	// Deactivate
	engine.DeleteMe(seqID)
	if seq.IsActive() {
		t.Error("Expected sequence to be inactive after DeleteMe")
	}

	// Cleanup
	engine.CleanupSequences()
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 0 {
		t.Errorf("Expected 0 sequencers after cleanup, got %d", len(sequencers))
	}
}
