package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestMidiEndDuringWait verifies that when MIDI playback ends during a wait operation,
// the VM resumes execution immediately instead of continuing to wait.
// This test validates Requirement 6.4: MIDI playback end during wait handling
func TestMidiEndDuringWait(t *testing.T) {
	// Save original sequencers
	originalSequencers := sequencers
	defer func() { sequencers = originalSequencers }()

	// Setup: Create a sequencer in MIDI_TIME mode with a wait operation
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{10}},    // Wait 10 steps
		{Cmd: interpreter.OpCall, Args: []interface{}{"nop"}}, // Next instruction after wait
	}

	seq := &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true,
		mode:         MidiTime, // MIDI_TIME mode
		ticksPerStep: 480,      // Standard PPQ
	}

	sequencers = []*Sequencer{seq}

	// Execute the Wait operation
	ExecuteOp(ops[0], seq)

	// Verify wait was set correctly (10 steps * 480 ticks/step - 1)
	expectedWait := 10*480 - 1
	if seq.waitTicks != expectedWait {
		t.Errorf("Expected waitTicks=%d, got %d", expectedWait, seq.waitTicks)
	}

	// Simulate MIDI playback ending
	midiFinished = true
	defer func() { midiFinished = false }() // Reset after test

	// Call UpdateVM - it should detect MIDI ended and resume execution
	UpdateVM(100)

	// Verify wait was cleared
	if seq.waitTicks != 0 {
		t.Errorf("Expected waitTicks to be cleared (0), got %d", seq.waitTicks)
	}

	// PC should still be at 0 because we haven't executed the next instruction yet
	// The next UpdateVM call will execute the instruction at PC=1
	if seq.pc != 0 {
		t.Errorf("Expected PC to remain at 0 (wait cleared but not executed yet), got %d", seq.pc)
	}

	// Call UpdateVM again - now it should execute the next instruction
	UpdateVM(101)

	// Now PC should have advanced
	if seq.pc != 1 {
		t.Errorf("Expected PC to advance to 1 after next UpdateVM, got %d", seq.pc)
	}
}

// TestMidiEndDuringWait_TimeMode verifies that MIDI end detection only applies
// to MIDI_TIME mode sequences, not TIME mode sequences
func TestMidiEndDuringWait_TimeMode(t *testing.T) {
	// Save original sequencers
	originalSequencers := sequencers
	defer func() { sequencers = originalSequencers }()

	// Setup: Create a sequencer in TIME mode with a wait operation
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{10}},    // Wait 10 steps
		{Cmd: interpreter.OpCall, Args: []interface{}{"nop"}}, // Next instruction after wait
	}

	seq := &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true,
		mode:         Time, // TIME mode (not MIDI_TIME)
		ticksPerStep: 12,   // Standard for TIME mode
	}

	sequencers = []*Sequencer{seq}

	// Execute the Wait operation
	ExecuteOp(ops[0], seq)

	// Verify wait was set correctly (10 steps * 12 ticks/step - 1)
	expectedWait := 10*12 - 1
	if seq.waitTicks != expectedWait {
		t.Errorf("Expected waitTicks=%d, got %d", expectedWait, seq.waitTicks)
	}

	initialWait := seq.waitTicks

	// Simulate MIDI playback ending
	midiFinished = true
	defer func() { midiFinished = false }() // Reset after test

	// Call UpdateVM - it should NOT clear wait for TIME mode
	UpdateVM(100)

	// Verify wait was decremented normally (not cleared)
	if seq.waitTicks != initialWait-1 {
		t.Errorf("Expected waitTicks to be decremented to %d, got %d", initialWait-1, seq.waitTicks)
	}

	// Verify PC did NOT advance
	if seq.pc != 0 {
		t.Errorf("Expected PC to remain at 0, got %d", seq.pc)
	}
}

// TestMidiEndDuringWait_Integration tests the full flow with multiple ticks
func TestMidiEndDuringWait_Integration(t *testing.T) {
	// Save original sequencers
	originalSequencers := sequencers
	defer func() { sequencers = originalSequencers }()

	// Setup: Create a sequencer in MIDI_TIME mode with a long wait
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{100}},   // Wait 100 steps (48000 ticks)
		{Cmd: interpreter.OpCall, Args: []interface{}{"nop"}}, // Next instruction after wait
	}

	seq := &Sequencer{
		commands:     ops,
		pc:           0,
		waitTicks:    0,
		active:       true,
		mode:         MidiTime,
		ticksPerStep: 480,
	}

	sequencers = []*Sequencer{seq}

	// Execute the Wait operation
	ExecuteOp(ops[0], seq)

	// Verify wait was set
	if seq.waitTicks <= 0 {
		t.Fatalf("Expected waitTicks > 0, got %d", seq.waitTicks)
	}

	initialWait := seq.waitTicks

	// Simulate several ticks passing (but not enough to complete the wait)
	for i := 0; i < 10; i++ {
		UpdateVM(i)
	}

	// Verify wait is being decremented
	if seq.waitTicks >= initialWait {
		t.Errorf("Expected waitTicks to decrease, got %d (initial: %d)", seq.waitTicks, initialWait)
	}

	// Verify PC has not advanced yet
	if seq.pc != 0 {
		t.Errorf("Expected PC to remain at 0 during wait, got %d", seq.pc)
	}

	// Now simulate MIDI ending
	midiFinished = true
	defer func() { midiFinished = false }() // Reset after test

	// Call UpdateVM - it should detect MIDI ended and resume execution
	UpdateVM(100)

	// Verify wait was cleared
	if seq.waitTicks != 0 {
		t.Errorf("Expected waitTicks to be cleared after MIDI end, got %d", seq.waitTicks)
	}

	// PC should still be at 0 because we haven't executed the next instruction yet
	if seq.pc != 0 {
		t.Errorf("Expected PC to remain at 0 (wait cleared but not executed yet), got %d", seq.pc)
	}

	// Call UpdateVM again - now it should execute the next instruction
	UpdateVM(101)

	// Now PC should have advanced
	if seq.pc != 1 {
		t.Errorf("Expected PC to advance to 1 after next UpdateVM, got %d", seq.pc)
	}
}

// TestMidiEndDuringWait_MultipleSequencers verifies that MIDI end affects all
// MIDI_TIME sequencers that are waiting
func TestMidiEndDuringWait_MultipleSequencers(t *testing.T) {
	// Save original sequencers
	originalSequencers := sequencers
	defer func() { sequencers = originalSequencers }()

	// Create multiple sequencers
	sequencers = []*Sequencer{
		// Sequencer 0: MIDI_TIME mode, waiting
		{
			commands: []OpCode{
				{Cmd: interpreter.OpWait, Args: []interface{}{10}},
			},
			pc:           0,
			waitTicks:    4799, // 10 steps * 480 - 1
			active:       true,
			mode:         MidiTime,
			ticksPerStep: 480,
		},
		// Sequencer 1: TIME mode, waiting (should not be affected)
		{
			commands: []OpCode{
				{Cmd: interpreter.OpWait, Args: []interface{}{10}},
			},
			pc:           0,
			waitTicks:    119, // 10 steps * 12 - 1
			active:       true,
			mode:         Time,
			ticksPerStep: 12,
		},
		// Sequencer 2: MIDI_TIME mode, not waiting
		{
			commands: []OpCode{
				{Cmd: interpreter.OpWait, Args: []interface{}{1}},
			},
			pc:           0,
			waitTicks:    0, // Not waiting
			active:       true,
			mode:         MidiTime,
			ticksPerStep: 480,
		},
	}

	// Simulate MIDI playback ending
	midiFinished = true
	defer func() { midiFinished = false }()

	// Call UpdateVM
	UpdateVM(100)

	// Verify Sequencer 0 (MIDI_TIME, waiting) had wait cleared
	if sequencers[0].waitTicks != 0 {
		t.Errorf("Sequencer 0: Expected waitTicks=0, got %d", sequencers[0].waitTicks)
	}
	// PC should still be at 0 (wait cleared but instruction not executed yet)
	if sequencers[0].pc != 0 {
		t.Errorf("Sequencer 0: Expected PC=0 (wait cleared), got %d", sequencers[0].pc)
	}

	// Verify Sequencer 1 (TIME mode, waiting) was decremented normally
	if sequencers[1].waitTicks != 118 { // 119 - 1
		t.Errorf("Sequencer 1: Expected waitTicks=118, got %d", sequencers[1].waitTicks)
	}
	if sequencers[1].pc != 0 {
		t.Errorf("Sequencer 1: Expected PC=0, got %d", sequencers[1].pc)
	}

	// Verify Sequencer 2 (MIDI_TIME, not waiting initially) executed Wait instruction
	// The Wait was executed and set waitTicks, PC was incremented to 1
	// But since len(commands)=1, it will loop back to PC=0 on next check
	if sequencers[2].waitTicks != 479 { // 1 step * 480 - 1
		t.Errorf("Sequencer 2: Expected waitTicks=479 (Wait executed), got %d", sequencers[2].waitTicks)
	}
	// PC loops back to 0 because it reached end of commands
	if sequencers[2].pc != 0 {
		t.Errorf("Sequencer 2: Expected PC=0 (looped back), got %d", sequencers[2].pc)
	}

	// Call UpdateVM again - now Sequencer 2's wait should be cleared by MIDI end
	UpdateVM(101)

	if sequencers[2].waitTicks != 0 {
		t.Errorf("Sequencer 2: Expected waitTicks=0 (cleared by MIDI end on second call), got %d", sequencers[2].waitTicks)
	}
}
