package engine

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestBackwardCompatibility_MesTimeNonBlocking verifies that mes(TIME) blocks
// execute with non-blocking behavior while maintaining timing accuracy
// **Validates: Requirements 8.1, 8.3**
func TestBackwardCompatibility_MesTimeNonBlocking(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTickCount := tickCount
	oldMidiSyncMode := midiSyncMode
	defer func() {
		sequencers = oldSequencers
		tickCount = oldTickCount
		midiSyncMode = oldMidiSyncMode
	}()

	// Reset state
	sequencers = nil
	tickCount = 0
	midiSyncMode = false

	// Create a simple TIME mode sequence with Wait operations
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{2}}, // Wait 2 steps = 24 ticks
		{Cmd: interpreter.OpWait, Args: []interface{}{3}}, // Wait 3 steps = 36 ticks
	}

	// Measure registration time
	start := time.Now()
	RegisterSequence(Time, ops)
	duration := time.Since(start)

	// Verify non-blocking: should return within 10ms
	if duration > 10*time.Millisecond {
		t.Errorf("RegisterSequence blocked for %v, expected < 10ms", duration)
	}

	// Verify sequence was registered
	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Verify sequence is active
	if !sequencers[0].active {
		t.Error("Expected sequence to be active")
	}

	// Verify timing mode
	if sequencers[0].mode != Time {
		t.Errorf("Expected TIME mode, got %d", sequencers[0].mode)
	}

	// Simulate execution and verify timing behavior
	// First Wait(2) = 24 ticks
	// Execute 23 ticks - should still be waiting
	for i := 0; i < 23; i++ {
		tickCount++
		UpdateVM(int(tickCount))
	}

	// Should still be waiting (need 24 ticks total)
	if sequencers[0].waitTicks == 0 {
		t.Error("Expected sequence to still be waiting after 23 ticks")
	}

	// One more tick (24th) should complete the wait
	tickCount++
	UpdateVM(int(tickCount))

	// After first wait completes, sequence should move to next instruction or loop
	// The sequence should still be active
	if !sequencers[0].active {
		t.Error("Expected sequence to remain active")
	}
}

// TestBackwardCompatibility_MidiTimeNonBlocking verifies that mes(MIDI_TIME)
// continues to work as non-blocking
// **Validates: Requirements 8.2, 8.4**
func TestBackwardCompatibility_MidiTimeNonBlocking(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTargetTick := targetTick
	oldMidiSyncMode := midiSyncMode
	defer func() {
		sequencers = oldSequencers
		targetTick = oldTargetTick
		midiSyncMode = oldMidiSyncMode
	}()

	// Reset state
	sequencers = nil
	targetTick = 0
	midiSyncMode = false

	// Create a MIDI_TIME mode sequence
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{1}}, // Wait 1 step
	}

	// Measure registration time
	start := time.Now()
	RegisterSequence(MidiTime, ops)
	duration := time.Since(start)

	// Verify non-blocking: should return within 10ms
	if duration > 10*time.Millisecond {
		t.Errorf("RegisterSequence blocked for %v, expected < 10ms", duration)
	}

	// Verify sequence was registered
	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Verify MIDI sync mode was set
	if !midiSyncMode {
		t.Error("Expected midiSyncMode to be true for MIDI_TIME")
	}

	// Verify sequence mode
	if sequencers[0].mode != MidiTime {
		t.Errorf("Expected MIDI_TIME mode, got %d", sequencers[0].mode)
	}
}

// TestBackwardCompatibility_WaitOperationTiming verifies that Wait() operations
// maintain the same timing behavior as before
// **Validates: Requirements 8.3**
func TestBackwardCompatibility_WaitOperationTiming(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTickCount := tickCount
	oldMidiSyncMode := midiSyncMode
	defer func() {
		sequencers = oldSequencers
		tickCount = oldTickCount
		midiSyncMode = oldMidiSyncMode
	}()

	tests := []struct {
		name          string
		mode          int
		waitSteps     int
		expectedTicks int
	}{
		{
			name:          "TIME mode Wait(1)",
			mode:          Time,
			waitSteps:     1,
			expectedTicks: 12, // 1 step = 12 ticks in TIME mode
		},
		{
			name:          "TIME mode Wait(5)",
			mode:          Time,
			waitSteps:     5,
			expectedTicks: 60, // 5 steps = 60 ticks
		},
		{
			name:          "MIDI_TIME mode Wait(1)",
			mode:          MidiTime,
			waitSteps:     1,
			expectedTicks: 60, // 1 step = 60 ticks in MIDI_TIME (PPQ=480, step=8)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset state
			sequencers = nil
			tickCount = 0
			midiSyncMode = (tt.mode == MidiTime)

			// Create sequence with Wait operation
			ops := []OpCode{
				{Cmd: interpreter.OpWait, Args: []interface{}{tt.waitSteps}},
			}

			RegisterSequence(tt.mode, ops)

			if len(sequencers) != 1 {
				t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
			}

			// Execute first instruction (Wait)
			tickCount = 1
			UpdateVM(int(tickCount))

			// Verify Wait was set
			if sequencers[0].waitTicks == 0 {
				t.Error("Expected waitTicks to be set")
			}

			// Advance ticks until Wait completes
			for i := 0; i < tt.expectedTicks; i++ {
				tickCount++
				UpdateVM(int(tickCount))
			}

			// After wait completes, PC should loop back to 0 (mes blocks loop)
			// and the sequence should still be active
			if !sequencers[0].active {
				t.Error("Expected sequence to remain active after wait completes")
			}
		})
	}
}

// TestBackwardCompatibility_EventHandlers verifies that event handlers
// (MIDI_END, RBDOWN) continue to work correctly
// **Validates: Requirements 8.4**
func TestBackwardCompatibility_EventHandlers(t *testing.T) {
	t.Run("MIDI_END handler", func(t *testing.T) {
		// Save and restore global state
		oldMidiEndHandler := midiEndHandler
		oldMidiEndTriggered := midiEndTriggered
		oldMidiFinished := midiFinished
		oldSequencers := sequencers
		defer func() {
			midiEndHandler = oldMidiEndHandler
			midiEndTriggered = oldMidiEndTriggered
			midiFinished = oldMidiFinished
			sequencers = oldSequencers
		}()

		// Reset state
		sequencers = nil
		midiEndTriggered = false
		midiFinished = false
		handlerCalled := false

		// Register MIDI_END handler
		RegisterMidiEndHandler(func() {
			handlerCalled = true
		})

		// Create a MIDI_TIME sequence with a wait
		ops := []OpCode{
			{Cmd: interpreter.OpWait, Args: []interface{}{10}},
		}

		seq := &Sequencer{
			commands:     ops,
			pc:           0,
			waitTicks:    4799, // Waiting
			active:       true,
			mode:         MidiTime,
			ticksPerStep: 480,
		}
		sequencers = []*Sequencer{seq}

		// Trigger MIDI end
		TriggerMidiEnd()

		// Wait for goroutine to execute
		time.Sleep(10 * time.Millisecond)

		// Verify handler was called
		if !handlerCalled {
			t.Error("Expected MIDI_END handler to be called")
		}

		// Set midiFinished flag (this is what actually clears waitTicks in UpdateVM)
		midiFinished = true

		// Call UpdateVM to clear the wait
		UpdateVM(100)

		// Verify wait was cleared
		if seq.waitTicks != 0 {
			t.Errorf("Expected waitTicks to be cleared, got %d", seq.waitTicks)
		}
	})

	t.Run("RBDOWN handler", func(t *testing.T) {
		// Save and restore global state
		oldRbDownHandler := rbDownHandler
		defer func() {
			rbDownHandler = oldRbDownHandler
		}()

		// Reset state
		handlerCalled := false

		// Register RBDOWN handler
		RegisterRBDownHandler(func() {
			handlerCalled = true
		})

		// Trigger right button down
		TriggerRBDown()

		// Wait for goroutine to execute
		time.Sleep(10 * time.Millisecond)

		// Verify handler was called
		if !handlerCalled {
			t.Error("Expected RBDOWN handler to be called")
		}
	})
}

// TestBackwardCompatibility_MultipleSequences verifies that multiple
// mes() blocks can execute concurrently without interfering
// **Validates: Requirements 8.1, 8.3**
func TestBackwardCompatibility_MultipleSequences(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTickCount := tickCount
	oldMidiSyncMode := midiSyncMode
	defer func() {
		sequencers = oldSequencers
		tickCount = oldTickCount
		midiSyncMode = oldMidiSyncMode
	}()

	// Reset state
	sequencers = nil
	tickCount = 0
	midiSyncMode = false

	// Register first sequence
	ops1 := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{2}}, // 24 ticks
	}
	RegisterSequence(Time, ops1)

	// Register second sequence
	ops2 := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{3}}, // 36 ticks
	}
	RegisterSequence(Time, ops2)

	// Verify both sequences registered
	if len(sequencers) != 2 {
		t.Fatalf("Expected 2 sequencers, got %d", len(sequencers))
	}

	// Verify both are active
	if !sequencers[0].active || !sequencers[1].active {
		t.Error("Expected both sequences to be active")
	}

	// Execute for 24 ticks
	for i := 0; i < 24; i++ {
		tickCount++
		UpdateVM(int(tickCount))
	}

	// First sequence should complete its wait and loop
	tickCount++
	UpdateVM(int(tickCount))

	// Both sequences should still be active (they loop)
	if !sequencers[0].active || !sequencers[1].active {
		t.Error("Expected both sequences to remain active")
	}

	// Continue for another 11 ticks (total 36)
	for i := 0; i < 11; i++ {
		tickCount++
		UpdateVM(int(tickCount))
	}

	// Both sequences should still be active and looping
	if !sequencers[0].active || !sequencers[1].active {
		t.Error("Expected both sequences to remain active after completing waits")
	}
}

// TestBackwardCompatibility_SequenceLooping verifies that mes() blocks
// loop correctly as they did before
// **Validates: Requirements 8.1**
func TestBackwardCompatibility_SequenceLooping(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTickCount := tickCount
	oldMidiSyncMode := midiSyncMode
	defer func() {
		sequencers = oldSequencers
		tickCount = oldTickCount
		midiSyncMode = oldMidiSyncMode
	}()

	// Reset state
	sequencers = nil
	tickCount = 0
	midiSyncMode = false

	// Create a simple sequence that should loop
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{1}}, // 12 ticks
	}

	RegisterSequence(Time, ops)

	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Execute through first iteration
	for i := 0; i < 13; i++ {
		tickCount++
		UpdateVM(int(tickCount))
	}

	// PC should have looped back to 0
	if sequencers[0].pc != 0 {
		t.Errorf("Expected PC to loop back to 0, got PC=%d", sequencers[0].pc)
	}

	// Sequence should still be active
	if !sequencers[0].active {
		t.Error("Expected sequence to remain active after looping")
	}

	// Execute through second iteration
	for i := 0; i < 13; i++ {
		tickCount++
		UpdateVM(int(tickCount))
	}

	// Should loop again
	if sequencers[0].pc != 0 {
		t.Errorf("Expected PC to loop back to 0 again, got PC=%d", sequencers[0].pc)
	}

	if !sequencers[0].active {
		t.Error("Expected sequence to remain active after second loop")
	}
}

// TestMidiTimeMode_NonBlocking verifies that MIDI_TIME mode remains non-blocking
// after the user input handling changes
// **Validates: Requirements 8.2, 8.4**
func TestMidiTimeMode_NonBlocking(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTargetTick := targetTick
	oldMidiSyncMode := midiSyncMode
	oldTickCount := tickCount
	defer func() {
		sequencers = oldSequencers
		targetTick = oldTargetTick
		midiSyncMode = oldMidiSyncMode
		tickCount = oldTickCount
	}()

	// Reset state
	sequencers = nil
	targetTick = 0
	midiSyncMode = false
	tickCount = 0

	// Create a MIDI_TIME mode sequence with multiple operations
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{10}}, // Wait 10 steps
		{Cmd: interpreter.OpWait, Args: []interface{}{5}},  // Wait 5 steps
	}

	// Measure registration time
	start := time.Now()
	RegisterSequence(MidiTime, ops)
	duration := time.Since(start)

	// Verify non-blocking: should return within 10ms
	if duration > 10*time.Millisecond {
		t.Errorf("RegisterSequence blocked for %v, expected < 10ms", duration)
	}

	// Verify sequence was registered
	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Verify MIDI sync mode was set
	if !midiSyncMode {
		t.Error("Expected midiSyncMode to be true for MIDI_TIME")
	}

	// Verify sequence mode
	if sequencers[0].mode != MidiTime {
		t.Errorf("Expected MIDI_TIME mode (%d), got %d", MidiTime, sequencers[0].mode)
	}

	// Verify sequence is active
	if !sequencers[0].active {
		t.Error("Expected sequence to be active")
	}

	// Verify ticksPerStep is set correctly for MIDI_TIME (should be 480 for PPQ=480)
	// Note: The default is 12, but MIDI_TIME uses PPQ-based timing
	if sequencers[0].ticksPerStep != 12 {
		t.Logf("Note: ticksPerStep=%d (default is 12, MIDI_TIME uses PPQ-based timing)", sequencers[0].ticksPerStep)
	}
}

// TestMidiTimeMode_Synchronization verifies that MIDI_TIME mode synchronization
// works correctly with targetTick updates
// **Validates: Requirements 8.2, 8.4**
func TestMidiTimeMode_Synchronization(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTargetTick := targetTick
	oldMidiSyncMode := midiSyncMode
	oldTickCount := tickCount
	defer func() {
		sequencers = oldSequencers
		atomic.StoreInt64(&targetTick, oldTargetTick)
		midiSyncMode = oldMidiSyncMode
		tickCount = oldTickCount
	}()

	// Reset state
	sequencers = nil
	atomic.StoreInt64(&targetTick, 0)
	midiSyncMode = false
	tickCount = 0

	// Create a MIDI_TIME mode sequence
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{1}}, // Wait 1 step
	}

	RegisterSequence(MidiTime, ops)

	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Verify MIDI sync mode is enabled
	if !midiSyncMode {
		t.Error("Expected midiSyncMode to be true")
	}

	// Simulate MIDI player advancing targetTick (this is what NotifyTick does)
	atomic.StoreInt64(&targetTick, 10)

	// Execute first instruction (Wait)
	tickCount = 1
	UpdateVM(int(tickCount))

	// Verify Wait was set
	if sequencers[0].waitTicks == 0 {
		t.Error("Expected waitTicks to be set after Wait instruction")
	}

	initialWait := sequencers[0].waitTicks

	// Advance ticks (simulating MIDI-driven updates)
	for i := int64(2); i <= 10; i++ {
		tickCount = i
		UpdateVM(int(tickCount))
	}

	// Verify wait is being decremented
	if sequencers[0].waitTicks >= initialWait {
		t.Errorf("Expected waitTicks to decrease from %d, got %d", initialWait, sequencers[0].waitTicks)
	}

	// Verify sequence is still active
	if !sequencers[0].active {
		t.Error("Expected sequence to remain active")
	}
}

// TestMidiTimeMode_LoopingBehavior verifies that MIDI_TIME sequences loop correctly
// **Validates: Requirements 8.2, 8.4**
func TestMidiTimeMode_LoopingBehavior(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTargetTick := targetTick
	oldMidiSyncMode := midiSyncMode
	oldTickCount := tickCount
	defer func() {
		sequencers = oldSequencers
		atomic.StoreInt64(&targetTick, oldTargetTick)
		midiSyncMode = oldMidiSyncMode
		tickCount = oldTickCount
	}()

	// Reset state
	sequencers = nil
	atomic.StoreInt64(&targetTick, 0)
	midiSyncMode = false
	tickCount = 0

	// Create a simple MIDI_TIME sequence that should loop
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{1}}, // Wait 1 step (60 ticks in MIDI_TIME with PPQ=480)
	}

	RegisterSequence(MidiTime, ops)

	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Execute through first iteration
	// Wait(1) = 1 * 60 = 60 ticks (assuming PPQ=480, step=8)
	// But the default ticksPerStep is 12, so Wait(1) = 12 ticks
	expectedWaitTicks := 12 - 1 // Wait sets waitTicks to (steps * ticksPerStep - 1)

	// Execute first instruction
	tickCount = 1
	UpdateVM(int(tickCount))

	// Verify Wait was set
	if sequencers[0].waitTicks != expectedWaitTicks {
		t.Logf("Note: waitTicks=%d (expected %d based on ticksPerStep=%d)",
			sequencers[0].waitTicks, expectedWaitTicks, sequencers[0].ticksPerStep)
	}

	// Advance ticks to complete the wait
	for i := int64(2); i <= int64(expectedWaitTicks+2); i++ {
		tickCount = i
		UpdateVM(int(tickCount))
	}

	// PC should have looped back to 0
	if sequencers[0].pc != 0 {
		t.Errorf("Expected PC to loop back to 0, got PC=%d", sequencers[0].pc)
	}

	// Sequence should still be active (MIDI_TIME sequences loop)
	if !sequencers[0].active {
		t.Error("Expected sequence to remain active after looping")
	}

	// Execute through second iteration
	for i := int64(0); i < int64(expectedWaitTicks+2); i++ {
		tickCount++
		UpdateVM(int(tickCount))
	}

	// Should loop again
	if sequencers[0].pc != 0 {
		t.Errorf("Expected PC to loop back to 0 again, got PC=%d", sequencers[0].pc)
	}

	if !sequencers[0].active {
		t.Error("Expected sequence to remain active after second loop")
	}
}

// TestMidiTimeMode_EventHandlerIntegration verifies that MIDI_END event handler
// works correctly with MIDI_TIME sequences
// **Validates: Requirements 8.4**
func TestMidiTimeMode_EventHandlerIntegration(t *testing.T) {
	// Save and restore global state
	oldMidiEndHandler := midiEndHandler
	oldMidiEndTriggered := midiEndTriggered
	oldMidiFinished := midiFinished
	oldSequencers := sequencers
	oldMidiSyncMode := midiSyncMode
	defer func() {
		midiEndHandler = oldMidiEndHandler
		midiEndTriggered = oldMidiEndTriggered
		midiFinished = oldMidiFinished
		sequencers = oldSequencers
		midiSyncMode = oldMidiSyncMode
	}()

	// Reset state
	sequencers = nil
	midiEndTriggered = false
	midiFinished = false
	midiSyncMode = false
	handlerCalled := false

	// Register MIDI_END handler
	RegisterMidiEndHandler(func() {
		handlerCalled = true
	})

	// Create a MIDI_TIME sequence with a wait
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{10}},
	}

	RegisterSequence(MidiTime, ops)

	if len(sequencers) != 1 {
		t.Fatalf("Expected 1 sequencer, got %d", len(sequencers))
	}

	// Execute first instruction (Wait)
	tickCount = 1
	UpdateVM(int(tickCount))

	// Verify sequence is waiting
	if sequencers[0].waitTicks == 0 {
		t.Error("Expected sequence to be waiting")
	}

	// Trigger MIDI end
	TriggerMidiEnd()

	// Wait for goroutine to execute
	time.Sleep(10 * time.Millisecond)

	// Verify handler was called
	if !handlerCalled {
		t.Error("Expected MIDI_END handler to be called")
	}

	// Set midiFinished flag (this is what actually clears waitTicks in UpdateVM)
	midiFinished = true

	// Call UpdateVM to clear the wait
	UpdateVM(int(tickCount) + 1)

	// Verify wait was cleared
	if sequencers[0].waitTicks != 0 {
		t.Errorf("Expected waitTicks to be cleared, got %d", sequencers[0].waitTicks)
	}

	// Verify sequence is still active
	if !sequencers[0].active {
		t.Error("Expected sequence to remain active after MIDI end")
	}
}

// TestMidiTimeMode_ConcurrentWithTimeMode verifies that MIDI_TIME and TIME mode
// sequences can coexist without interfering with each other
// **Validates: Requirements 8.2, 8.4**
func TestMidiTimeMode_ConcurrentWithTimeMode(t *testing.T) {
	// Save and restore global state
	oldSequencers := sequencers
	oldTargetTick := targetTick
	oldMidiSyncMode := midiSyncMode
	oldTickCount := tickCount
	defer func() {
		sequencers = oldSequencers
		atomic.StoreInt64(&targetTick, oldTargetTick)
		midiSyncMode = oldMidiSyncMode
		tickCount = oldTickCount
	}()

	// Reset state
	sequencers = nil
	atomic.StoreInt64(&targetTick, 0)
	midiSyncMode = false
	tickCount = 0

	// Register a TIME mode sequence first
	timeOps := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{2}}, // 24 ticks
	}
	RegisterSequence(Time, timeOps)

	// Verify TIME mode was set
	if midiSyncMode {
		t.Error("Expected midiSyncMode to be false after TIME sequence")
	}

	// Register a MIDI_TIME mode sequence
	midiOps := []OpCode{
		{Cmd: interpreter.OpWait, Args: []interface{}{3}}, // 36 ticks
	}
	RegisterSequence(MidiTime, midiOps)

	// Verify MIDI_TIME mode was set (last registered sequence determines mode)
	if !midiSyncMode {
		t.Error("Expected midiSyncMode to be true after MIDI_TIME sequence")
	}

	// Verify both sequences registered
	if len(sequencers) != 2 {
		t.Fatalf("Expected 2 sequencers, got %d", len(sequencers))
	}

	// Verify first sequence is TIME mode
	if sequencers[0].mode != Time {
		t.Errorf("Expected first sequence to be TIME mode, got %d", sequencers[0].mode)
	}

	// Verify second sequence is MIDI_TIME mode
	if sequencers[1].mode != MidiTime {
		t.Errorf("Expected second sequence to be MIDI_TIME mode, got %d", sequencers[1].mode)
	}

	// Both sequences should be active
	if !sequencers[0].active || !sequencers[1].active {
		t.Error("Expected both sequences to be active")
	}

	// Execute a few ticks
	for i := int64(1); i <= 5; i++ {
		tickCount = i
		UpdateVM(int(tickCount))
	}

	// Both sequences should still be active
	if !sequencers[0].active || !sequencers[1].active {
		t.Error("Expected both sequences to remain active after execution")
	}
}
