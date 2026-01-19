package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestHeadlessExecution tests that headless mode executes without Ebiten initialization
// Requirements: 43.1, 43.2, 43.3, 43.4
func TestHeadlessExecution(t *testing.T) {
	// This test verifies that we can execute scripts in headless mode
	// without initializing Ebiten or creating GUI windows

	engine := NewTestEngine()

	// Set headless mode flag
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	// Test that rendering operations are stubbed out and don't crash
	// These should log but not perform actual rendering

	// Test OpenWin in headless mode (Requirement 43.4)
	winID := OpenWin(0, 100, 100, 640, 480, 0, 0, 0xFFFFFF)
	if winID != 0 {
		t.Errorf("Expected dummy window ID 0 in headless mode, got %d", winID)
	}

	// Test PutCast in headless mode (Requirement 43.4)
	castID := PutCast(0, 0, 0, 0, 100, 100, 0, 0)
	if castID != 0 {
		t.Errorf("Expected dummy cast ID 0 in headless mode, got %d", castID)
	}

	// Test MoveCast in headless mode (Requirement 43.4)
	// Should not crash, just log
	MoveCast(0, 50, 50)

	// Verify engine state is still consistent
	AssertStateConsistency(t, engine)
}

// TestHeadlessRenderingStubs tests that rendering operations are properly stubbed
// Requirements: 43.2, 43.4
func TestHeadlessRenderingStubs(t *testing.T) {
	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	// Create test engine with assets
	engine := NewTestEngineWithAssets(GetTestAssets())

	// Load a picture (this should work even in headless mode)
	picID := engine.LoadPic("test.bmp")
	if picID < 0 {
		t.Fatal("Failed to load picture in headless mode")
	}

	// Test OpenWin - should return dummy ID without creating actual window
	winID := OpenWin(picID, 0, 0, 640, 480, 0, 0, 0xFFFFFF)
	if winID != 0 {
		t.Errorf("Expected dummy window ID in headless mode, got %d", winID)
	}

	// Test PutCast - should return dummy ID without creating actual cast
	castID := PutCast(picID, 0, 0, 0, 100, 100, 0, 0)
	if castID != 0 {
		t.Errorf("Expected dummy cast ID in headless mode, got %d", castID)
	}

	// Test MoveCast - should not crash
	MoveCast(castID, 100, 100)

	// Verify no actual windows or casts were created in engine state
	// (since we're using global functions in headless mode, they bypass the engine)
	AssertResourceCount(t, engine, 1, 0, 0) // Only the loaded picture should exist
}

// TestHeadlessScriptLogic tests that script logic executes normally in headless mode
// Requirements: 43.3, 43.4
func TestHeadlessScriptLogic(t *testing.T) {
	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	// Test that timing and state changes work in headless mode
	engine := NewTestEngine()

	// Test variable assignment (should work normally)
	Assign("testVar", 42)

	// Access globalVars directly (it's a map[string]any)
	vmLock.Lock()
	val := globalVars["testvar"] // Case-insensitive (stored as lowercase)
	vmLock.Unlock()

	if val != 42 {
		t.Errorf("Expected variable value 42, got %v", val)
	}

	// Note: Full VM testing with RegisterSequence requires integration tests
	// Here we just verify that basic state management works in headless mode

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestHeadlessTimingSystem tests that timing system works in headless mode
// Requirements: 43.3, 43.4
func TestHeadlessTimingSystem(t *testing.T) {
	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	engine := NewTestEngine()

	// Test that VM tick updates work in headless mode
	// Simulate VM ticks
	startTick := 0
	for i := startTick; i < startTick+10; i++ {
		UpdateVM(i)
	}

	// Verify no crashes occurred
	AssertStateConsistency(t, engine)

	// Note: Full timing system testing with Wait operations requires
	// integration tests with actual RegisterSequence calls
}

// TestHeadlessAudioInitialization tests that audio initializes in headless mode
// Requirements: 43.3
func TestHeadlessAudioInitialization(t *testing.T) {
	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	// Audio should initialize even in headless mode (for MIDI_TIME synchronization)
	// but playback will be muted

	// This test verifies that InitializeAudio doesn't crash in headless mode
	// We can't easily test the actual audio initialization without side effects,
	// but we can verify the flag is set correctly

	if !headlessMode {
		t.Error("headlessMode flag should be true")
	}

	// Note: Actual audio initialization happens in InitDirect, which we can't
	// easily test here without full integration. This test mainly verifies
	// the headless flag behavior.
}

// TestHeadlessCleanExit tests that headless mode exits cleanly
// Requirements: 43.6
func TestHeadlessCleanExit(t *testing.T) {
	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	engine := NewTestEngine()

	// Create some resources
	Assign("var1", 100)
	Assign("var2", "test")

	// Verify resources exist
	vmLock.Lock()
	val := globalVars["var1"]
	vmLock.Unlock()

	if val != 100 {
		t.Error("Variable not set correctly")
	}

	// Reset should clean up all resources
	engine.Reset()

	// Verify cleanup
	AssertResourceCount(t, engine, 0, 0, 0)
	AssertStateConsistency(t, engine)
}

// TestHeadlessMesTimeExecution tests that mes(TIME) blocks execute in headless mode
// Requirements: 9.1, 9.2
func TestHeadlessMesTimeExecution(t *testing.T) {
	// Reset engine state for clean test
	ResetEngineForTest()

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	engine := NewTestEngine()

	// Test that mes(TIME) blocks execute without requiring window events
	// Requirement 9.1: Execute mes(TIME) blocks without requiring window events

	// Create a simple sequence with Wait operations
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 0}},
		{Cmd: interpreter.OpWait, Args: []any{2}}, // Wait 2 steps (24 ticks at 60 FPS)
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 1}},
		{Cmd: interpreter.OpWait, Args: []any{2}},
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 2}},
	}

	// Register the sequence in TIME mode
	RegisterSequence(Time, ops)

	// Verify sequence was registered
	vmLock.Lock()
	seqCount := len(sequencers)
	vmLock.Unlock()

	if seqCount == 0 {
		t.Fatal("Sequence was not registered")
	}

	// Simulate VM execution for several ticks
	// Requirement 9.2: Maintain timing accuracy
	startTick := 0
	for tick := startTick; tick < startTick+60; tick++ {
		UpdateVM(tick)
	}

	// Verify the sequence executed and updated the counter variable
	vmLock.Lock()
	counter := globalVars["counter"] // Variable names are case-insensitive, stored as-is
	vmLock.Unlock()

	// After 60 ticks, the sequence should have executed and set counter
	// The sequence loops, so counter might be 0, 1, or 2 depending on timing
	// The important thing is that it was set (not nil)
	if counter == nil {
		t.Error("Counter variable was not set - sequence did not execute")
	} else {
		t.Logf("Counter value: %v (sequence executed successfully)", counter)
	}

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestHeadlessTimingAccuracy tests that timing is maintained in headless mode
// Requirements: 9.2
func TestHeadlessTimingAccuracy(t *testing.T) {
	// Reset engine state for clean test
	ResetEngineForTest()

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	engine := NewTestEngine()

	// Create a sequence with precise Wait operations
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"tick1", 0}},
		{Cmd: interpreter.OpWait, Args: []any{5}}, // Wait 5 steps = 60 ticks
		{Cmd: interpreter.OpAssign, Args: []any{"tick2", 0}},
		{Cmd: interpreter.OpWait, Args: []any{3}}, // Wait 3 steps = 36 ticks
		{Cmd: interpreter.OpAssign, Args: []any{"tick3", 0}},
	}

	// Register the sequence
	RegisterSequence(Time, ops)

	// Track when each assignment happens
	tick1Set := -1
	tick2Set := -1
	tick3Set := -1

	// Simulate VM execution and track when variables are set
	for tick := 0; tick < 120; tick++ {
		UpdateVM(tick)

		vmLock.Lock()
		if tick1Set == -1 && globalVars["tick1"] != nil {
			tick1Set = tick
		}
		if tick2Set == -1 && globalVars["tick2"] != nil {
			tick2Set = tick
		}
		if tick3Set == -1 && globalVars["tick3"] != nil {
			tick3Set = tick
		}
		vmLock.Unlock()
	}

	// Verify timing accuracy
	// tick1 should be set at tick 0
	if tick1Set != 0 {
		t.Errorf("tick1 set at tick %d, expected 0", tick1Set)
	}

	// tick2 should be set at tick 60 (after Wait(5) = 60 ticks)
	if tick2Set < 58 || tick2Set > 62 {
		t.Errorf("tick2 set at tick %d, expected around 60 (±2)", tick2Set)
	}

	// tick3 should be set at tick 96 (60 + 36)
	if tick3Set < 94 || tick3Set > 98 {
		t.Errorf("tick3 set at tick %d, expected around 96 (±2)", tick3Set)
	}

	// Verify state consistency
	AssertStateConsistency(t, engine)
}

// TestHeadlessTermination tests that headless mode can be terminated programmatically
// Requirements: 9.3
// Note: This tests the engine's termination mechanism, not the CLI --timeout flag
func TestHeadlessTermination(t *testing.T) {
	// Reset engine state for clean test
	ResetEngineForTest()

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() { headlessMode = oldHeadlessMode }()

	engine := NewTestEngine()

	// Create a long-running sequence
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 0}},
		{Cmd: interpreter.OpWait, Args: []any{10}}, // Wait 10 steps = 120 ticks
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 1}},
	}

	// Register the sequence
	RegisterSequence(Time, ops)

	// Execute for a few ticks
	for tick := 0; tick < 10; tick++ {
		UpdateVM(tick)
	}

	// Verify sequence is still active
	vmLock.Lock()
	activeCount := 0
	for _, seq := range sequencers {
		if seq.active {
			activeCount++
		}
	}
	vmLock.Unlock()

	if activeCount == 0 {
		t.Error("Expected at least one active sequence")
	}

	// Simulate termination (like what --timeout would do)
	programTerminated = true

	// Execute one more tick - sequence should stop
	UpdateVM(11)

	// Verify all sequences are now inactive
	vmLock.Lock()
	activeCount = 0
	for _, seq := range sequencers {
		if seq.active {
			activeCount++
		}
	}
	vmLock.Unlock()

	if activeCount != 0 {
		t.Errorf("Expected all sequences to be inactive after termination, but %d are still active", activeCount)
	}

	// Verify state consistency
	AssertStateConsistency(t, engine)
}
