package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestESCKeyHandling_ProgramTerminatedFlag verifies that when the ESC key
// sets the programTerminated flag, the Update() method returns ebiten.Termination.
// This validates task 2.4 from user-input-handling spec (Requirements 3.1, 3.2)
//
// NOTE: Direct testing of ebiten.IsKeyPressed(ebiten.KeyEscape) is not possible
// in unit tests as it requires actual keyboard input. This test verifies the
// behavior after the flag is set, which is the testable part of ESC key handling.
func TestESCKeyHandling_ProgramTerminatedFlag(t *testing.T) {
	// Reset state
	ResetEngineForTest()

	// Create a minimal game instance
	game := &Game{
		state:     NewEngineState(),
		renderer:  NewEbitenRenderer(),
		tickCount: 0,
	}

	// Test 1: Normal operation - programTerminated = false
	// Update() should return nil (no error)
	programTerminated = false
	err := game.Update()
	if err != nil {
		t.Errorf("Update() should return nil when programTerminated is false, got: %v", err)
	}

	// Test 2: After ESC key press - programTerminated = true
	// Simulate the effect of ESC key press by setting the flag
	// (In real execution, this would be set by: ebiten.IsKeyPressed(ebiten.KeyEscape))
	programTerminated = true
	err = game.Update()
	if err != ebiten.Termination {
		t.Errorf("Update() should return ebiten.Termination when programTerminated is true, got: %v", err)
	}

	// Cleanup
	ResetEngineForTest()
}

// TestESCKeyHandling_TerminationBeforeVMExecution verifies that the termination
// check happens before VM execution, preventing OpCodes from running after
// the ESC key is pressed (or programTerminated flag is set).
// This validates task 2.4 from user-input-handling spec (Requirements 3.1, 3.2)
func TestESCKeyHandling_TerminationBeforeVMExecution(t *testing.T) {
	ResetEngineForTest()

	// Create a game instance
	game := &Game{
		state:     NewEngineState(),
		renderer:  NewEbitenRenderer(),
		tickCount: 0,
	}

	// Register a sequence with multiple operations
	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
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

	// First update - should execute normally
	programTerminated = false
	err := game.Update()
	if err != nil {
		t.Errorf("First Update() should succeed, got error: %v", err)
	}

	// Record the PC after first update
	vmLock.Lock()
	pcAfterFirstUpdate := mainSequencer.pc
	vmLock.Unlock()

	// Simulate ESC key press by setting termination flag
	// (In real execution: ebiten.IsKeyPressed(ebiten.KeyEscape) returns true)
	programTerminated = true

	// Second update - should return immediately without executing more OpCodes
	err = game.Update()
	if err != ebiten.Termination {
		t.Errorf("Second Update() should return ebiten.Termination after ESC press, got: %v", err)
	}

	// Verify sequencer PC hasn't changed (no more OpCodes executed)
	vmLock.Lock()
	pcAfterTermination := mainSequencer.pc
	vmLock.Unlock()

	if pcAfterTermination != pcAfterFirstUpdate {
		t.Errorf("Sequencer PC should not advance after termination, was %d, now %d",
			pcAfterFirstUpdate, pcAfterTermination)
	}

	// Cleanup
	ResetEngineForTest()
}

// TestESCKeyHandling_ImmediateTermination verifies that when programTerminated
// is set, Update() returns ebiten.Termination immediately without any processing.
// This validates task 2.4 from user-input-handling spec (Requirements 3.1, 3.2)
func TestESCKeyHandling_ImmediateTermination(t *testing.T) {
	ResetEngineForTest()

	// Create a game instance
	game := &Game{
		state:     NewEngineState(),
		renderer:  NewEbitenRenderer(),
		tickCount: 0,
	}

	// Register a sequence that would normally execute
	ops := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{1}},
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

	// Set termination flag BEFORE any execution
	// (Simulates ESC key being pressed before first update)
	programTerminated = true

	// Call Update - should return immediately without executing any OpCode
	err := game.Update()
	if err != ebiten.Termination {
		t.Errorf("Update() should return ebiten.Termination immediately when programTerminated is true, got: %v", err)
	}

	// Verify the sequencer's PC is still 0 (no OpCodes executed)
	vmLock.Lock()
	if mainSequencer.pc != 0 {
		t.Errorf("Sequencer PC should still be 0 (no OpCodes executed), got: %d", mainSequencer.pc)
	}
	// Verify the sequencer is still active (termination check happens in Update, not UpdateVM)
	if !mainSequencer.active {
		t.Error("Sequencer should still be active (Update returned before calling UpdateVM)")
	}
	vmLock.Unlock()

	// Cleanup
	ResetEngineForTest()
}

// TestESCKeyHandling_MultipleSequences verifies that termination stops
// execution for all active sequences, not just the main sequencer.
// This validates task 2.4 from user-input-handling spec (Requirements 3.1, 3.2)
func TestESCKeyHandling_MultipleSequences(t *testing.T) {
	ResetEngineForTest()

	// Create a game instance
	game := &Game{
		state:     NewEngineState(),
		renderer:  NewEbitenRenderer(),
		tickCount: 0,
	}

	// Register multiple sequences
	ops1 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
	}
	ops2 := []interpreter.OpCode{
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
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
	sequencers = []*Sequencer{seq1, seq2}
	vmLock.Unlock()

	// First update - should execute normally for both sequences
	programTerminated = false
	err := game.Update()
	if err != nil {
		t.Errorf("First Update() should succeed, got error: %v", err)
	}

	// Record PCs after first update
	vmLock.Lock()
	pc1AfterFirst := seq1.pc
	pc2AfterFirst := seq2.pc
	vmLock.Unlock()

	// Simulate ESC key press
	programTerminated = true

	// Second update - should return immediately without executing more OpCodes
	err = game.Update()
	if err != ebiten.Termination {
		t.Errorf("Second Update() should return ebiten.Termination after ESC press, got: %v", err)
	}

	// Verify both sequencers' PCs haven't changed
	vmLock.Lock()
	pc1AfterTermination := seq1.pc
	pc2AfterTermination := seq2.pc
	vmLock.Unlock()

	if pc1AfterTermination != pc1AfterFirst {
		t.Errorf("Sequence 1 PC should not advance after termination, was %d, now %d",
			pc1AfterFirst, pc1AfterTermination)
	}
	if pc2AfterTermination != pc2AfterFirst {
		t.Errorf("Sequence 2 PC should not advance after termination, was %d, now %d",
			pc2AfterFirst, pc2AfterTermination)
	}

	// Cleanup
	ResetEngineForTest()
}
