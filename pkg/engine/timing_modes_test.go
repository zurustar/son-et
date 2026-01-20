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
	engine.UpdateVM()
	if seq.GetPC() != 1 {
		t.Errorf("Expected PC=1, got %d", seq.GetPC())
	}

	// Tick 2: Execute wait command, set waitCount=2 (2 steps × 1 tick/step)
	engine.UpdateVM()
	if seq.GetPC() != 2 {
		t.Errorf("Expected PC=2 after wait, got %d", seq.GetPC())
	}

	if !seq.IsWaiting() {
		t.Errorf("Expected sequence to be waiting")
	}

	// Ticks 3-4: Decrement wait (2 -> 1 -> 0)
	for i := 0; i < 2; i++ {
		engine.UpdateVM()
		if i < 1 && !seq.IsWaiting() {
			t.Errorf("Expected sequence to still be waiting at tick %d", i+3)
		}
	}

	if seq.IsWaiting() {
		t.Errorf("Expected sequence to finish waiting after 2 ticks")
	}

	// Tick 5: Execute next command (y=2)
	engine.UpdateVM()
	if seq.GetPC() != 3 {
		t.Errorf("Expected PC=3, got %d", seq.GetPC())
	}

	if seq.GetVariable("y") != int64(2) {
		t.Errorf("Expected y=2, got %v", seq.GetVariable("y"))
	}
}

// TestTimingMode_MixedModes verifies that TIME and MIDI_TIME sequences can coexist.
func TestTimingMode_MixedModes(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	// TIME mode sequence: wait(1) = 3 ticks
	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("a"), int64(10)}},
	}

	// MIDI_TIME mode sequence: wait(1) = 1 tick
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

	// Tick 2: seq1 waiting (3->2), seq2 waiting (1->0)
	engine.UpdateVM()

	if !seq1.IsWaiting() {
		t.Errorf("Expected seq1 to still be waiting")
	}

	if seq2.IsWaiting() {
		t.Errorf("Expected seq2 to finish waiting")
	}

	// Tick 3: seq1 waiting (2->1), seq2 executes
	engine.UpdateVM()

	if !seq1.IsWaiting() {
		t.Errorf("Expected seq1 to still be waiting")
	}

	if seq2.GetVariable("b") != int64(20) {
		t.Errorf("Expected b=20, got %v", seq2.GetVariable("b"))
	}

	// Tick 4: seq1 waiting (1->0), seq2 complete
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

	// Verify NO sequence was created (non-blocking)
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 0 {
		t.Errorf("Expected 0 sequencers for MIDI_TIME mode (non-blocking), got %d", len(sequencers))
	}

	// Sequence will be created when event is triggered
	engine.TriggerEvent(EventMIDI_TIME, nil)

	sequencers = engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Errorf("Expected 1 sequencer after trigger, got %d", len(sequencers))
	}
}

// TestGetStepSize verifies that stepSize is correctly set based on timing mode.
func TestGetStepSize(t *testing.T) {
	// TIME mode: 3 ticks per step (50ms at 60 FPS)
	seqTime := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	if seqTime.GetStepSize() != 3 {
		t.Errorf("Expected TIME mode stepSize=3, got %d", seqTime.GetStepSize())
	}

	// MIDI_TIME mode: 1 tick per step (32nd note)
	seqMidi := NewSequencer([]interpreter.OpCode{}, MIDI_TIME, nil)
	if seqMidi.GetStepSize() != 1 {
		t.Errorf("Expected MIDI_TIME mode stepSize=1, got %d", seqMidi.GetStepSize())
	}
}
