package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestTimingMode_TIME_StepDuration verifies that TIME mode uses 3 ticks per step (50ms).
func TestTimingMode_TIME_StepDuration(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // wait(2) = 2 steps = 6 ticks
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	// Tick 1: Execute first command (x=1)
	engine.UpdateVM()
	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1, got %d", seq.GetPC())
	}

	// Tick 2: Execute wait command, set waitCount=6 (2 steps × 3 ticks/step)
	engine.UpdateVM()
	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 after wait, got %d", seq.GetPC())
	}

	if !seq.IsWaiting() {
		t.Errorf("Expected sequence to be waiting")
	}

	// Ticks 3-8: Decrement wait (6 -> 5 -> 4 -> 3 -> 2 -> 1 -> 0)
	for i := 0; i < 6; i++ {
		engine.UpdateVM()
		if i < 5 && !seq.IsWaiting() {
			t.Errorf("Expected sequence to still be waiting at tick %d", i+3)
		}
	}

	if seq.IsWaiting() {
		t.Errorf("Expected sequence to finish waiting after 6 ticks")
	}

	// Tick 9: Execute next command (y=2)
	engine.UpdateVM()
	if seq.GetPC() != 3 {
		t.Errorf("Expected PC=3, got %d", seq.GetPC())
	}

	if seq.GetVariable("y") != int64(2) {
		t.Errorf("Expected y=2, got %v", seq.GetVariable("y"))
	}
}

// TestTimingMode_MIDI_TIME_StepDuration verifies that MIDI_TIME mode uses 1 tick per step.
// Note: MIDI_TIME sequences are updated by UpdateMIDISequences, not UpdateVM.
func TestTimingMode_MIDI_TIME_StepDuration(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // wait(2) = 2 steps = 2 ticks
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
	}

	seq := NewSequencer(opcodes, MIDI_TIME, nil)
	engine.RegisterSequence(seq, 0)

	// Tick 1: Execute first command (x=1)
	engine.UpdateMIDISequences(1)
	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1, got %d", seq.GetPC())
	}

	// Tick 2: Execute wait command, set waitCount=2 (2 steps × 1 tick/step)
	engine.UpdateMIDISequences(1)
	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 after wait, got %d", seq.GetPC())
	}

	if !seq.IsWaiting() {
		t.Errorf("Expected sequence to be waiting")
	}

	// Ticks 3-4: Decrement wait (2 -> 1 -> 0)
	for i := 0; i < 2; i++ {
		engine.UpdateMIDISequences(1)
		if i < 1 && !seq.IsWaiting() {
			t.Errorf("Expected sequence to still be waiting at tick %d", i+3)
		}
	}

	if seq.IsWaiting() {
		t.Errorf("Expected sequence to finish waiting after 2 ticks")
	}

	// Tick 5: Execute next command (y=2)
	engine.UpdateMIDISequences(1)
	if seq.GetPC() != 3 {
		t.Errorf("Expected PC=3, got %d", seq.GetPC())
	}

	if seq.GetVariable("y") != int64(2) {
		t.Errorf("Expected y=2, got %v", seq.GetVariable("y"))
	}
}

// TestTimingMode_MixedModes verifies that TIME and MIDI_TIME sequences can coexist.
// Note: MIDI_TIME sequences are updated by UpdateMIDISequences, not UpdateVM.
func TestTimingMode_MixedModes(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	// TIME mode sequence: wait(1) = 3 ticks
	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("a"), int64(10)}},
	}

	// MIDI_TIME mode sequence: wait(1) = 1 tick (updated by MIDI ticks)
	opcodes2 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("b"), int64(20)}},
	}

	seq1 := NewSequencer(opcodes1, TIME, nil)
	seq2 := NewSequencer(opcodes2, MIDI_TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)

	// Tick 1: Both execute wait
	engine.UpdateVM()

	if !seq1.IsWaiting() || !seq2.IsWaiting() {
		t.Errorf("Expected both sequences to be waiting")
	}

	// Tick 2: seq1 waiting (3->2), seq2 NOT updated (MIDI_TIME needs UpdateMIDISequences)
	engine.UpdateVM()

	if !seq1.IsWaiting() {
		t.Errorf("Expected seq1 to still be waiting")
	}

	// MIDI_TIME sequences are updated by UpdateMIDISequences, not UpdateVM
	// So seq2 should still be waiting
	if !seq2.IsWaiting() {
		t.Errorf("Expected seq2 to still be waiting (MIDI_TIME not updated by UpdateVM)")
	}

	// Simulate MIDI tick to update seq2
	engine.UpdateMIDISequences(1)

	if seq2.IsWaiting() {
		t.Errorf("Expected seq2 to finish waiting after MIDI tick")
	}

	// Another MIDI tick to execute seq2's assignment
	engine.UpdateMIDISequences(1)

	if seq2.GetVariable("b") != int64(20) {
		t.Errorf("Expected b=20, got %v", seq2.GetVariable("b"))
	}

	// Continue TIME mode sequence
	// Tick 3: seq1 waiting (2->1)
	engine.UpdateVM()
	// Tick 4: seq1 waiting (1->0)
	engine.UpdateVM()

	if seq1.IsWaiting() {
		t.Errorf("Expected seq1 to finish waiting")
	}

	// Tick 5: seq1 executes
	engine.UpdateVM()

	if seq1.GetVariable("a") != int64(10) {
		t.Errorf("Expected a=10, got %v", seq1.GetVariable("a"))
	}
}

// TestTimingMode_BlockingBehavior_TIME verifies that TIME mode blocks during mes() registration.
// Note: Full blocking behavior requires main loop integration and will be verified in Task 7.3.8.
func TestTimingMode_BlockingBehavior_TIME(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}

	// Register TIME mes() block
	handlerID := engine.RegisterMesBlock(EventTIME, opcodes, nil, 0)

	// Verify handler was registered
	if handlerID == 0 {
		t.Errorf("Expected non-zero handler ID")
	}

	// Verify a sequence was created for immediate execution
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Errorf("Expected 1 sequencer for TIME mode, got %d", len(sequencers))
	}

	// Note: Full blocking behavior (RegisterMesBlock waits until sequence completes)
	// requires main loop integration and will be verified in Task 7.3.8
}

// TestTimingMode_NonBlockingBehavior_MIDI_TIME verifies that MIDI_TIME mode doesn't block.
// Note: MIDI_TIME mode creates a sequence immediately but doesn't block the caller.
func TestTimingMode_NonBlockingBehavior_MIDI_TIME(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}

	// Register MIDI_TIME mes() block
	handlerID := engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)

	// Verify handler was registered
	if handlerID == 0 {
		t.Errorf("Expected non-zero handler ID")
	}

	// MIDI_TIME mode creates a sequence immediately (non-blocking)
	// The sequence runs asynchronously, driven by MIDI ticks
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Errorf("Expected 1 sequencer for MIDI_TIME mode, got %d", len(sequencers))
	}

	// Verify the sequence is in MIDI_TIME mode
	if sequencers[0].GetMode() != MIDI_TIME {
		t.Errorf("Expected MIDI_TIME mode, got %d", sequencers[0].GetMode())
	}
}

// TestGetTicksPerStep verifies that ticksPerStep is correctly set based on timing mode.
func TestGetTicksPerStep(t *testing.T) {
	// TIME mode: 3 ticks per step (50ms at 60 FPS)
	seqTime := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	if seqTime.GetTicksPerStep() != 3 {
		t.Errorf("Expected TIME mode ticksPerStep=3, got %d", seqTime.GetTicksPerStep())
	}

	// MIDI_TIME mode: 1 tick per step (MIDI ticks are at 32nd note resolution)
	seqMidi := NewSequencer([]interpreter.OpCode{}, MIDI_TIME, nil)
	if seqMidi.GetTicksPerStep() != 1 {
		t.Errorf("Expected MIDI_TIME mode ticksPerStep=1, got %d", seqMidi.GetTicksPerStep())
	}
}

// TestTimingMode_TIME_MultipleSequences verifies that multiple TIME mode sequences can run concurrently.
func TestTimingMode_TIME_MultipleSequences(t *testing.T) {
	engine := NewEngine(nil, nil, nil)
	engine.Start()

	// Sequence 1: wait(1) = 3 ticks, then set a=10
	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("a"), int64(10)}},
	}

	// Sequence 2: wait(2) = 6 ticks, then set b=20
	opcodes2 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("b"), int64(20)}},
	}

	seq1 := NewSequencer(opcodes1, TIME, nil)
	seq2 := NewSequencer(opcodes2, TIME, nil)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)

	// Tick 1: Both execute wait
	engine.Update()
	if !seq1.IsWaiting() || !seq2.IsWaiting() {
		t.Errorf("Expected both sequences to be waiting after tick 1")
	}

	// Ticks 2-4: seq1 waiting (3->2->1->0), seq2 waiting (6->5->4->3)
	for i := 0; i < 3; i++ {
		engine.Update()
	}

	if seq1.IsWaiting() {
		t.Errorf("Expected seq1 to finish waiting after 4 ticks")
	}
	if !seq2.IsWaiting() {
		t.Errorf("Expected seq2 to still be waiting after 4 ticks")
	}

	// Tick 5: seq1 executes a=10, seq2 waiting (3->2)
	engine.Update()
	if seq1.GetVariable("a") != int64(10) {
		t.Errorf("Expected a=10, got %v", seq1.GetVariable("a"))
	}

	// Ticks 6-7: seq2 waiting (2->1->0)
	for i := 0; i < 2; i++ {
		engine.Update()
	}

	if seq2.IsWaiting() {
		t.Errorf("Expected seq2 to finish waiting after 7 ticks")
	}

	// Tick 8: seq2 executes b=20
	engine.Update()
	if seq2.GetVariable("b") != int64(20) {
		t.Errorf("Expected b=20, got %v", seq2.GetVariable("b"))
	}
}

// TestTimingMode_TIME_AdvancesEachFrame verifies that TIME mode sequences advance each frame.
func TestTimingMode_TIME_AdvancesEachFrame(t *testing.T) {
	engine := NewEngine(nil, nil, nil)
	engine.Start()

	// Sequence with multiple commands
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("z"), int64(3)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	// Each Update should execute one command
	engine.Update()
	if seq.GetVariable("x") != int64(1) {
		t.Errorf("Expected x=1 after tick 1, got %v", seq.GetVariable("x"))
	}
	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1 after tick 1, got %d", seq.GetPC())
	}

	engine.Update()
	if seq.GetVariable("y") != int64(2) {
		t.Errorf("Expected y=2 after tick 2, got %v", seq.GetVariable("y"))
	}
	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 after tick 2, got %d", seq.GetPC())
	}

	engine.Update()
	if seq.GetVariable("z") != int64(3) {
		t.Errorf("Expected z=3 after tick 3, got %v", seq.GetVariable("z"))
	}
	if seq.GetPC() != 3 {
		t.Errorf("Expected PC=3 after tick 3, got %d", seq.GetPC())
	}

	if !seq.IsComplete() {
		t.Errorf("Expected sequence to be complete after 3 ticks")
	}
}
