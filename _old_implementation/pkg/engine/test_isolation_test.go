package engine

import (
	"sync/atomic"
	"testing"
)

// TestResetEngineForTest_CleansUpAllState tests that ResetEngineForTest properly cleans up all state
func TestResetEngineForTest_CleansUpAllState(t *testing.T) {
	// Set up some state
	mainSequencer = &Sequencer{
		commands: []OpCode{},
		pc:       5,
		active:   true,
	}
	sequencers = []*Sequencer{mainSequencer}
	globalVars = map[string]any{"test": 123}
	tickCount = 100
	ticksPerStep = 24
	midiSyncMode = true
	GlobalPPQ = 960
	atomic.StoreInt64(&targetTick, 200)
	programTerminated = true

	// Reset
	ResetEngineForTest()

	// Verify all state is cleaned up
	if mainSequencer != nil {
		t.Error("mainSequencer should be nil after reset")
	}
	if sequencers != nil {
		t.Error("sequencers should be nil after reset")
	}
	if len(globalVars) != 0 {
		t.Errorf("globalVars should be empty, got %d entries", len(globalVars))
	}
	if tickCount != 0 {
		t.Errorf("tickCount should be 0, got %d", tickCount)
	}
	if ticksPerStep != 12 {
		t.Errorf("ticksPerStep should be 12, got %d", ticksPerStep)
	}
	if midiSyncMode {
		t.Error("midiSyncMode should be false after reset")
	}
	if GlobalPPQ != 480 {
		t.Errorf("GlobalPPQ should be 480, got %d", GlobalPPQ)
	}
	if atomic.LoadInt64(&targetTick) != 0 {
		t.Errorf("targetTick should be 0, got %d", atomic.LoadInt64(&targetTick))
	}
	if programTerminated {
		t.Error("programTerminated should be false after reset")
	}
}

// TestResetEngineForTest_AllowsMultipleTests tests that multiple tests can run in sequence without interference
func TestResetEngineForTest_AllowsMultipleTests(t *testing.T) {
	// Test 1: Set some state
	ResetEngineForTest()
	globalVars["test1"] = 100
	tickCount = 50

	// Test 2: Reset and verify clean state
	ResetEngineForTest()
	if val, ok := globalVars["test1"]; ok {
		t.Errorf("globalVars should not contain test1 after reset, got %v", val)
	}
	if tickCount != 0 {
		t.Errorf("tickCount should be 0 after reset, got %d", tickCount)
	}

	// Test 3: Set different state
	globalVars["test2"] = 200
	tickCount = 75

	// Test 4: Reset and verify clean state again
	ResetEngineForTest()
	if val, ok := globalVars["test2"]; ok {
		t.Errorf("globalVars should not contain test2 after reset, got %v", val)
	}
	if tickCount != 0 {
		t.Errorf("tickCount should be 0 after reset, got %d", tickCount)
	}
}

// TestResetEngineForTest_NoGlobalStateLeakage tests that there is no global state leakage between tests
func TestResetEngineForTest_NoGlobalStateLeakage(t *testing.T) {
	// This test should run independently and not see state from other tests
	ResetEngineForTest()

	// Verify clean state
	if mainSequencer != nil {
		t.Error("mainSequencer should be nil at test start")
	}
	if sequencers != nil {
		t.Error("sequencers should be nil at test start")
	}
	if len(globalVars) != 0 {
		t.Errorf("globalVars should be empty at test start, got %d entries", len(globalVars))
	}
	if tickCount != 0 {
		t.Errorf("tickCount should be 0 at test start, got %d", tickCount)
	}
	if atomic.LoadInt64(&targetTick) != 0 {
		t.Errorf("targetTick should be 0 at test start, got %d", atomic.LoadInt64(&targetTick))
	}
}
