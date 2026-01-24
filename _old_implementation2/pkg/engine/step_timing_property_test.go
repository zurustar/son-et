package engine

import (
	"math/rand"
	"testing"
	"testing/quick"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestStepTimingAccuracyProperty tests Property 3: Step Timing Accuracy
// **Feature: sample-game-fixes, Property 3: Step Timing Accuracy**
// **Validates: Requirements 2.2**
//
// Property: For any step scheduled for a specific tick or time in a mes() block,
// the sequencer should execute it at the correct moment according to the timing
// mode (TIME or MIDI_TIME).
func TestStepTimingAccuracyProperty(t *testing.T) {
	// Property: Steps execute at the correct tick based on ticksPerStep
	timingProperty := func(stepValue uint8, waitSteps uint8) bool {
		// Limit values to reasonable ranges
		stepVal := int(stepValue%16) + 1   // 1-16 ticks per step
		waitCount := int(waitSteps%10) + 1 // 1-10 steps to wait

		// Test TIME mode
		if !testTimingMode(t, TIME, stepVal, waitCount) {
			return false
		}

		// Test MIDI_TIME mode
		if !testTimingMode(t, MIDI_TIME, stepVal, waitCount) {
			return false
		}

		return true
	}

	if err := quick.Check(timingProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Step timing accuracy property failed: %v", err)
	}
}

// testTimingMode tests timing accuracy for a specific mode
func testTimingMode(t *testing.T, mode TimingMode, stepVal int, waitCount int) bool {
	engine := NewEngine(nil, nil, nil)
	engine.SetDebugLevel(DebugLevelError)

	// Create sequence: SetStep(stepVal), Wait(waitCount), Assign(x=1)
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpSetStep, Args: []any{int64(stepVal)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(waitCount)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}

	seq := NewSequencer(opcodes, mode, nil)
	seq.SetNoLoop(true)
	engine.RegisterSequence(seq, 0)

	// Calculate expected ticks
	var expectedTicks int
	if mode == TIME {
		// TIME mode: step(n) = n × 3 ticks
		expectedTicks = stepVal * 3 * waitCount
	} else {
		// MIDI_TIME mode: step(n) = n ticks
		expectedTicks = stepVal * waitCount
	}

	// Execute ticks and track when x is set
	actualTicks := 0
	maxTicks := expectedTicks + 100 // Safety margin

	for tick := 0; tick < maxTicks; tick++ {
		if mode == TIME {
			engine.UpdateVM()
		} else {
			engine.UpdateMIDISequences(1)
		}

		// Check if x has been set
		val := seq.GetVariable("x")
		if val == int64(1) {
			actualTicks = tick + 1
			break
		}
	}

	// Verify timing accuracy
	// Allow for 1 tick variance due to execution order
	// (SetStep and Wait execute in same tick, then wait starts)
	// Tick 1: SetStep executes
	// Tick 2: Wait executes, sets waitCount
	// Ticks 3 to 2+expectedTicks: waiting
	// Tick 3+expectedTicks: x=1 executes
	expectedExecutionTick := 2 + expectedTicks + 1

	if actualTicks < expectedExecutionTick-1 || actualTicks > expectedExecutionTick+1 {
		t.Logf("Timing mismatch for mode=%d, stepVal=%d, waitCount=%d: expected ~%d, got %d",
			mode, stepVal, waitCount, expectedExecutionTick, actualTicks)
		return false
	}

	return true
}

// TestStepTimingTIMEMode tests step timing specifically for TIME mode
// TIME mode: step(n) = n × 50ms = n × 3 ticks at 60 FPS
//
// Note: The engine executes multiple commands per tick until it hits a wait.
// So SetStep and Wait execute in the same tick, then waiting begins.
func TestStepTimingTIMEMode(t *testing.T) {
	testCases := []struct {
		name              string
		stepValue         int
		waitSteps         int
		expectedWaitTicks int
	}{
		{"step(1)_wait(1)", 1, 1, 3},  // 1 × 3 × 1 = 3 ticks
		{"step(2)_wait(1)", 2, 1, 6},  // 2 × 3 × 1 = 6 ticks
		{"step(1)_wait(2)", 1, 2, 6},  // 1 × 3 × 2 = 6 ticks
		{"step(3)_wait(2)", 3, 2, 18}, // 3 × 3 × 2 = 18 ticks
		{"step(8)_wait(1)", 8, 1, 24}, // 8 × 3 × 1 = 24 ticks (common in games)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine(nil, nil, nil)
			engine.SetDebugLevel(DebugLevelError)

			opcodes := []interpreter.OpCode{
				{Cmd: interpreter.OpSetStep, Args: []any{int64(tc.stepValue)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(tc.waitSteps)}},
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("done"), int64(1)}},
			}

			seq := NewSequencer(opcodes, TIME, nil)
			seq.SetNoLoop(true)
			engine.RegisterSequence(seq, 0)

			// Tick 1: Engine executes SetStep AND Wait in same tick (until wait is hit)
			engine.UpdateVM()
			if seq.GetPC() != 2 {
				t.Errorf("Expected PC=2 after first tick (SetStep+Wait), got %d", seq.GetPC())
			}
			if !seq.IsWaiting() {
				t.Error("Expected sequence to be waiting after first tick")
			}

			// The wait count should be the full expectedWaitTicks
			initialWaitCount := seq.GetWaitCount()
			if initialWaitCount != tc.expectedWaitTicks {
				t.Errorf("Expected initial waitCount=%d, got %d", tc.expectedWaitTicks, initialWaitCount)
			}

			// Execute wait ticks (each tick decrements the counter)
			for i := 0; i < tc.expectedWaitTicks; i++ {
				engine.UpdateVM()
			}

			// Verify wait completed
			if seq.IsWaiting() {
				t.Errorf("Expected sequence to finish waiting, but waitCount=%d", seq.GetWaitCount())
			}

			// Execute assignment
			engine.UpdateVM()
			if seq.GetVariable("done") != int64(1) {
				t.Error("Expected done=1 after wait completed")
			}
		})
	}
}

// TestStepTimingMIDITIMEMode tests step timing specifically for MIDI_TIME mode
// MIDI_TIME mode: step(n) = n ticks (MIDI ticks at 32nd note resolution)
func TestStepTimingMIDITIMEMode(t *testing.T) {
	testCases := []struct {
		name              string
		stepValue         int
		waitSteps         int
		expectedWaitTicks int
	}{
		{"step(1)_wait(1)", 1, 1, 1},  // 1 × 1 = 1 tick
		{"step(2)_wait(1)", 2, 1, 2},  // 2 × 1 = 2 ticks
		{"step(1)_wait(2)", 1, 2, 2},  // 1 × 2 = 2 ticks
		{"step(8)_wait(1)", 8, 1, 8},  // 8 × 1 = 8 ticks (common in y_saru)
		{"step(8)_wait(4)", 8, 4, 32}, // 8 × 4 = 32 ticks
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine(nil, nil, nil)
			engine.SetDebugLevel(DebugLevelError)

			opcodes := []interpreter.OpCode{
				{Cmd: interpreter.OpSetStep, Args: []any{int64(tc.stepValue)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(tc.waitSteps)}},
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("done"), int64(1)}},
			}

			seq := NewSequencer(opcodes, MIDI_TIME, nil)
			seq.SetNoLoop(true)
			engine.RegisterSequence(seq, 0)

			// Execute SetStep (tick 1)
			engine.UpdateMIDISequences(1)
			if seq.GetPC() != 1 {
				t.Errorf("Expected PC=1 after SetStep, got %d", seq.GetPC())
			}

			// Execute Wait (tick 2) - this sets the wait counter
			engine.UpdateMIDISequences(1)
			if seq.GetPC() != 2 {
				t.Errorf("Expected PC=2 after Wait, got %d", seq.GetPC())
			}
			if !seq.IsWaiting() {
				t.Error("Expected sequence to be waiting after Wait")
			}
			if seq.GetWaitCount() != tc.expectedWaitTicks {
				t.Errorf("Expected waitCount=%d, got %d", tc.expectedWaitTicks, seq.GetWaitCount())
			}

			// Execute wait ticks
			for i := 0; i < tc.expectedWaitTicks; i++ {
				engine.UpdateMIDISequences(1)
				if i < tc.expectedWaitTicks-1 && !seq.IsWaiting() {
					t.Errorf("Expected sequence to still be waiting at tick %d", i+3)
				}
			}

			// Verify wait completed
			if seq.IsWaiting() {
				t.Errorf("Expected sequence to finish waiting after %d ticks", tc.expectedWaitTicks)
			}

			// Execute assignment
			engine.UpdateMIDISequences(1)
			if seq.GetVariable("done") != int64(1) {
				t.Error("Expected done=1 after wait completed")
			}
		})
	}
}

// TestStepTimingRandomProperty tests step timing with random ticksPerStep values
// This provides broader coverage across the input space
func TestStepTimingRandomProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	// Run 30 iterations with random parameters
	for iteration := 0; iteration < 30; iteration++ {
		stepValue := rng.Intn(16) + 1   // 1-16
		waitSteps := rng.Intn(5) + 1    // 1-5
		mode := TimingMode(rng.Intn(2)) // TIME or MIDI_TIME

		t.Run("Random_Iteration", func(t *testing.T) {
			engine := NewEngine(nil, nil, nil)
			engine.SetDebugLevel(DebugLevelError)

			opcodes := []interpreter.OpCode{
				{Cmd: interpreter.OpSetStep, Args: []any{int64(stepValue)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(waitSteps)}},
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("marker"), int64(42)}},
			}

			seq := NewSequencer(opcodes, mode, nil)
			seq.SetNoLoop(true)
			engine.RegisterSequence(seq, 0)

			// Calculate expected wait ticks
			var expectedWaitTicks int
			if mode == TIME {
				expectedWaitTicks = stepValue * 3 * waitSteps
			} else {
				expectedWaitTicks = stepValue * waitSteps
			}

			// Execute until marker is set or max ticks reached
			maxTicks := expectedWaitTicks + 100
			markerSet := false

			for tick := 0; tick < maxTicks; tick++ {
				if mode == TIME {
					engine.UpdateVM()
				} else {
					engine.UpdateMIDISequences(1)
				}

				if seq.GetVariable("marker") == int64(42) {
					markerSet = true
					break
				}
			}

			if !markerSet {
				t.Errorf("Marker not set after %d ticks (mode=%d, step=%d, wait=%d)",
					maxTicks, mode, stepValue, waitSteps)
			}
		})
	}
}
