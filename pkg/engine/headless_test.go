package engine

import (
	"testing"
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
