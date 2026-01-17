package engine

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestMesTimeExecutesOnce tests that mes(TIME) block executes exactly once
func TestMesTimeExecutesOnce(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Create a sequence with a simple operation
	ops := []OpCode{
		{
			Cmd:  interpreter.OpAssign,
			Args: []any{Variable("counter"), 1},
		},
	}

	// Register sequence in TIME mode
	done := make(chan bool)
	go func() {
		RegisterSequence(Time, ops)
		done <- true
	}()

	// Simulate VM ticks to complete the sequence
	go func() {
		for i := 0; i < 10; i++ {
			time.Sleep(10 * time.Millisecond)
			UpdateVM(i)
		}
	}()

	// Wait for sequence to complete (with timeout)
	select {
	case <-done:
		// Success - sequence completed
	case <-time.After(2 * time.Second):
		t.Fatal("RegisterSequence did not complete within timeout")
	}

	// Verify sequence is no longer active
	vmLock.Lock()
	if mainSequencer != nil && mainSequencer.active {
		t.Error("Sequence should be inactive after completion")
	}
	vmLock.Unlock()
}

// TestRegisterSequenceBlocksInTimeMode tests that RegisterSequence blocks until sequence completes in TIME mode
func TestRegisterSequenceBlocksInTimeMode(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Create a sequence with a Wait operation
	ops := []OpCode{
		{
			Cmd:  interpreter.OpWait,
			Args: []any{1}, // Wait 1 step
		},
	}

	// Track when RegisterSequence returns
	completed := false
	go func() {
		RegisterSequence(Time, ops)
		completed = true
	}()

	// RegisterSequence should block, so completed should still be false
	time.Sleep(100 * time.Millisecond)
	if completed {
		t.Error("RegisterSequence should block in TIME mode, but it returned immediately")
	}

	// Simulate VM ticks to complete the sequence
	for i := 0; i < 20; i++ {
		UpdateVM(i)
		time.Sleep(10 * time.Millisecond)
	}

	// Now RegisterSequence should have completed
	time.Sleep(100 * time.Millisecond)
	if !completed {
		t.Error("RegisterSequence should have completed after sequence finished")
	}
}

// TestSubsequentCodeRunsAfterMesTime tests that code after mes(TIME) block executes
func TestSubsequentCodeRunsAfterMesTime(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Create a sequence
	ops := []OpCode{
		{
			Cmd:  interpreter.OpAssign,
			Args: []any{Variable("x"), 1},
		},
	}

	// Track completion
	completed := false
	go func() {
		RegisterSequence(Time, ops)
		// This code runs after mes(TIME) completes
		completed = true
	}()

	// Simulate VM ticks to complete the sequence
	for i := 0; i < 10; i++ {
		UpdateVM(i)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait a bit for the goroutine to complete
	time.Sleep(100 * time.Millisecond)

	// Verify subsequent code ran
	if !completed {
		t.Error("Subsequent code should run after mes(TIME) completes")
	}
}

// TestDelAllDelMeExecuteAfterMesTime tests that del_all/del_me execute after mes(TIME) completes
func TestDelAllDelMeExecuteAfterMesTime(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Initialize global engine for del_all/del_me
	globalEngine = NewTestEngine()

	// Create a sequence that includes del_all and del_me
	ops := []OpCode{
		{
			Cmd:  interpreter.OpAssign,
			Args: []any{Variable("x"), 1},
		},
		{
			Cmd:  interpreter.OpCall,
			Args: []any{"del_all"},
		},
		{
			Cmd:  interpreter.OpCall,
			Args: []any{"del_me"},
		},
	}

	// Register sequence
	go func() {
		RegisterSequence(Time, ops)
	}()

	// Simulate VM ticks
	terminated := false
	for i := 0; i < 20 && !terminated; i++ {
		UpdateVM(i)
		time.Sleep(10 * time.Millisecond)

		// Check if program terminated
		if programTerminated {
			terminated = true
			break
		}
	}

	// Verify del_me was executed (program terminated)
	if !terminated {
		t.Error("del_me should have terminated the program")
	}
}

// TestNoSequenceReregistrationAfterCompletion tests that sequence is not re-registered after completion
func TestNoSequenceReregistrationAfterCompletion(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Create a simple sequence
	ops := []OpCode{
		{
			Cmd:  interpreter.OpAssign,
			Args: []any{Variable("x"), 1},
		},
	}

	// Register sequence
	go func() {
		RegisterSequence(Time, ops)
	}()

	// Wait for sequence to be added
	time.Sleep(50 * time.Millisecond)

	// Get initial sequencer count
	vmLock.Lock()
	initialCount := len(sequencers)
	vmLock.Unlock()

	// Simulate VM ticks to complete the sequence
	for i := 0; i < 10; i++ {
		UpdateVM(i)
		time.Sleep(10 * time.Millisecond)
	}

	// Wait a bit more
	time.Sleep(100 * time.Millisecond)

	// Verify sequencer count hasn't increased
	vmLock.Lock()
	finalCount := len(sequencers)
	vmLock.Unlock()

	if finalCount > initialCount {
		t.Errorf("Sequence was re-registered: initial=%d, final=%d", initialCount, finalCount)
	}

	// Verify sequence is inactive
	vmLock.Lock()
	if len(sequencers) > 0 && sequencers[0] != nil && sequencers[0].active {
		t.Error("Sequence should be inactive after completion")
	}
	vmLock.Unlock()
}
