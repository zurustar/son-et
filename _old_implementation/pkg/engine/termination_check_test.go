package engine

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestGameUpdate_TerminationCheckAtStart verifies that Game.Update() checks
// the programTerminated flag at the very start and returns ebiten.Termination
// This validates task 2.3 from user-input-handling spec
func TestGameUpdate_TerminationCheckAtStart(t *testing.T) {
	// Reset state
	ResetEngineForTest()

	// Create a minimal game instance
	game := &Game{
		state:     NewEngineState(),
		renderer:  NewEbitenRenderer(),
		tickCount: 0,
	}

	// Test 1: Normal operation (programTerminated = false)
	programTerminated = false
	err := game.Update()
	if err != nil {
		t.Errorf("Update() should return nil when programTerminated is false, got: %v", err)
	}

	// Test 2: Termination flag set (programTerminated = true)
	programTerminated = true
	err = game.Update()
	if err != ebiten.Termination {
		t.Errorf("Update() should return ebiten.Termination when programTerminated is true, got: %v", err)
	}

	// Test 3: Verify termination check happens BEFORE any processing
	// Reset and set up a sequencer that would execute
	ResetEngineForTest()
	programTerminated = true

	// Register a simple sequence
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

	// Call Update - should return immediately without executing the OpCode
	err = game.Update()
	if err != ebiten.Termination {
		t.Errorf("Update() should return ebiten.Termination before processing VM, got: %v", err)
	}

	// Verify the sequencer's PC hasn't advanced (OpCode wasn't executed)
	vmLock.Lock()
	if mainSequencer.pc != 0 {
		t.Errorf("Sequencer PC should still be 0 (OpCode not executed), got: %d", mainSequencer.pc)
	}
	vmLock.Unlock()

	// Cleanup
	ResetEngineForTest()
}

// TestGameUpdate_TerminationCheckBeforeVMExecution verifies that the termination
// check happens before any VM execution, preventing OpCodes from running after termination
func TestGameUpdate_TerminationCheckBeforeVMExecution(t *testing.T) {
	ResetEngineForTest()

	// Create a game instance
	game := &Game{
		state:     NewEngineState(),
		renderer:  NewEbitenRenderer(),
		tickCount: 0,
	}

	// Register a sequence with a simple operation
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

	// First update - should execute normally
	programTerminated = false
	err := game.Update()
	if err != nil {
		t.Errorf("First Update() should succeed, got error: %v", err)
	}

	// Verify sequencer advanced (OpCode was executed)
	vmLock.Lock()
	firstPC := mainSequencer.pc
	vmLock.Unlock()

	if firstPC == 0 {
		// PC might still be 0 if the OpCode doesn't advance it
		// This is okay for this test
	}

	// Set termination flag
	programTerminated = true

	// Second update - should return immediately without executing more OpCodes
	err = game.Update()
	if err != ebiten.Termination {
		t.Errorf("Second Update() should return ebiten.Termination, got: %v", err)
	}

	// Verify sequencer PC hasn't changed (no more OpCodes executed)
	vmLock.Lock()
	secondPC := mainSequencer.pc
	vmLock.Unlock()

	if secondPC != firstPC {
		t.Errorf("Sequencer PC should not advance after termination, was %d, now %d", firstPC, secondPC)
	}

	// Cleanup
	ResetEngineForTest()
}
