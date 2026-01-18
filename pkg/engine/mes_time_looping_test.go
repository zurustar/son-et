package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestMesTimeLooping tests that mes(TIME) blocks loop back to the beginning
// when they reach the end of their command list
func TestMesTimeLooping(t *testing.T) {
	// Setup - reset global state
	sequencers = nil
	programTerminated = false

	// Create a simple mes(TIME) block with a single Assign command
	// This will be executed once per tick, and should loop back
	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"testVar", 1}},
	}

	// Register the sequence
	RegisterSequence(Time, ops)

	// Verify sequence was registered
	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequence to be registered, got %d", len(sequencers))
	}

	// Run for 10 ticks - the sequence should loop 10 times
	for tick := 1; tick <= 10; tick++ {
		UpdateVM(tick)

		// After each tick, verify the sequence is still active and pc is reset to 0
		if !sequencers[0].active {
			t.Errorf("Expected sequence to be active at tick %d", tick)
		}

		// After executing the single command, pc should loop back to 0
		if sequencers[0].pc != 0 {
			t.Errorf("Expected pc to be 0 after looping at tick %d, got %d", tick, sequencers[0].pc)
		}
	}

	// Verify that the sequence is still active (hasn't terminated)
	if !sequencers[0].active {
		t.Error("Expected mes(TIME) sequence to still be active after looping")
	}
}

// TestMesMidiTimeLooping tests that mes(MIDI_TIME) blocks also loop
func TestMesMidiTimeLooping(t *testing.T) {
	// Setup - reset global state
	sequencers = nil
	programTerminated = false

	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"testVar", 1}},
	}

	// Register as MIDI_TIME mode
	RegisterSequence(MidiTime, ops)

	// Verify sequence was registered
	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequence to be registered, got %d", len(sequencers))
	}

	// Run for 10 ticks - the sequence should loop 10 times
	for tick := 1; tick <= 10; tick++ {
		UpdateVM(tick)

		// After each tick, verify the sequence is still active and pc is reset to 0
		if !sequencers[0].active {
			t.Errorf("Expected sequence to be active at tick %d", tick)
		}

		// After executing the single command, pc should loop back to 0
		if sequencers[0].pc != 0 {
			t.Errorf("Expected pc to be 0 after looping at tick %d, got %d", tick, sequencers[0].pc)
		}
	}

	// Verify that the sequence is still active
	if !sequencers[0].active {
		t.Error("Expected mes(MIDI_TIME) sequence to still be active after looping")
	}
}
