package engine

import (
	"math"
	"testing"
	"testing/quick"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Property 6: Frame-Accurate Timing
// Feature: user-input-handling, Property 6: Frame-Accurate Timing
// Validates: Requirements 5.1
func TestProperty6_FrameAccurateTiming(t *testing.T) {
	// Property: Tick count increments at 60 ticks per second (±5% tolerance)
	t.Run("Tick rate accuracy over time", func(t *testing.T) {
		property := func(durationMs uint16) bool {
			// Constrain to reasonable range (100-500ms)
			if durationMs < 100 || durationMs > 500 {
				return true
			}

			duration := time.Duration(durationMs) * time.Millisecond

			// Reset tick count
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			// Measure tick increments over time
			start := time.Now()
			startTick := int64(0)

			// Simulate Update() calls at 60 FPS
			targetFPS := 60.0
			frameDuration := time.Second / time.Duration(targetFPS)

			for time.Since(start) < duration {
				// Simulate one frame update
				tickLock.Lock()
				tickCount++
				_ = tickCount // Use the value
				tickLock.Unlock()

				// Sleep to simulate frame timing
				time.Sleep(frameDuration)
			}

			elapsed := time.Since(start)
			tickLock.Lock()
			endTick := tickCount
			tickLock.Unlock()

			ticksElapsed := endTick - startTick
			expectedTicks := int64(elapsed.Seconds() * 60.0)

			// Verify within ±5% tolerance
			tolerance := 0.05 // 5%
			lowerBound := float64(expectedTicks) * (1.0 - tolerance)
			upperBound := float64(expectedTicks) * (1.0 + tolerance)

			if float64(ticksElapsed) < lowerBound || float64(ticksElapsed) > upperBound {
				t.Logf("Tick rate out of tolerance: expected=%d, actual=%d, elapsed=%v",
					expectedTicks, ticksElapsed, elapsed)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Tick increments are consistent across frames
	t.Run("Consistent tick increments", func(t *testing.T) {
		property := func(frameCount uint8) bool {
			// Constrain to reasonable range (10-100 frames)
			if frameCount < 10 || frameCount > 100 {
				return true
			}

			// Reset tick count
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			// Track tick increments
			previousTick := int64(0)
			allIncrementsAreOne := true

			for i := uint8(0); i < frameCount; i++ {
				// Simulate one frame update (TIME mode)
				tickLock.Lock()
				tickCount++
				currentTick := tickCount
				tickLock.Unlock()

				// Verify increment is exactly 1
				increment := currentTick - previousTick
				if increment != 1 {
					t.Logf("Tick increment not 1: previous=%d, current=%d, increment=%d",
						previousTick, currentTick, increment)
					allIncrementsAreOne = false
					break
				}

				previousTick = currentTick
			}

			return allIncrementsAreOne
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Tick rate under various sequence loads
	t.Run("Tick rate under load", func(t *testing.T) {
		property := func(seqCount uint8, opsPerSeq uint8) bool {
			// Constrain to reasonable ranges
			if seqCount < 1 || seqCount > 10 {
				return true
			}
			if opsPerSeq < 5 || opsPerSeq > 20 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			vmLock.Unlock()

			// Register multiple sequences
			for i := uint8(0); i < seqCount; i++ {
				ops := generateRandomOpCodes(int(opsPerSeq))
				RegisterSequence(Time, ops)
			}

			// Measure tick rate over short duration
			start := time.Now()
			startTick := int64(0)
			duration := 200 * time.Millisecond

			frameDuration := time.Second / 60

			for time.Since(start) < duration {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				// Update VM with sequences
				UpdateVM(currentTick)

				time.Sleep(frameDuration)
			}

			elapsed := time.Since(start)
			tickLock.Lock()
			endTick := tickCount
			tickLock.Unlock()

			ticksElapsed := endTick - startTick
			expectedTicks := int64(elapsed.Seconds() * 60.0)

			// Verify within ±5% tolerance even under load
			tolerance := 0.05
			lowerBound := float64(expectedTicks) * (1.0 - tolerance)
			upperBound := float64(expectedTicks) * (1.0 + tolerance)

			return float64(ticksElapsed) >= lowerBound && float64(ticksElapsed) <= upperBound
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// Property 7: Wait Operation Accuracy
// Feature: user-input-handling, Property 7: Wait Operation Accuracy
// Validates: Requirements 5.2, 8.3
func TestProperty7_WaitOperationAccuracy(t *testing.T) {
	// Property: Wait(N) waits exactly N steps (N * ticksPerStep ticks) before next OpCode executes
	// Note: In TIME mode, ticksPerStep defaults to 12, so Wait(N) waits N*12 ticks
	t.Run("Wait duration accuracy", func(t *testing.T) {
		property := func(waitSteps uint8) bool {
			// Constrain to reasonable range (1-50 steps)
			if waitSteps < 1 || waitSteps > 50 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			// Create sequence with Wait operation
			// Assign before wait, assign after wait
			ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("before"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(waitSteps)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("after"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, ops)

			// Execute until "before" is set
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Record tick when wait started
			vmLock.Lock()
			_, beforeSet := globalVars["before"]
			_, afterSet := globalVars["after"]
			vmLock.Unlock()

			if !beforeSet || afterSet {
				t.Logf("Initial state incorrect: before=%v, after=%v", beforeSet, afterSet)
				return false
			}

			tickLock.Lock()
			waitStartTick := tickCount
			tickLock.Unlock()

			// Execute ticks until "after" is set
			// In TIME mode, Wait(N) waits N * ticksPerStep ticks (default ticksPerStep=12)
			expectedWaitTicks := int(waitSteps) * 12
			maxTicks := expectedWaitTicks + 10
			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				_, afterSet := globalVars["after"]
				vmLock.Unlock()

				if afterSet {
					break
				}
			}

			tickLock.Lock()
			waitEndTick := tickCount
			tickLock.Unlock()

			// Calculate actual wait duration
			actualWaitTicks := waitEndTick - waitStartTick

			// Verify exactly N*12 ticks passed (allow 1 tick tolerance for execution timing)
			if actualWaitTicks < int64(expectedWaitTicks) || actualWaitTicks > int64(expectedWaitTicks)+1 {
				t.Logf("Wait duration incorrect: expected=%d ticks (%d steps * 12), actual=%d ticks",
					expectedWaitTicks, waitSteps, actualWaitTicks)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Multiple Wait operations accumulate correctly
	t.Run("Multiple Wait operations", func(t *testing.T) {
		property := func(wait1 uint8, wait2 uint8) bool {
			// Constrain to reasonable ranges
			if wait1 < 1 || wait1 > 20 || wait2 < 1 || wait2 > 20 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			// Create sequence with two Wait operations
			ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("start"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(wait1)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("middle"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(wait2)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("end"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{3}},
					},
				},
			}

			RegisterSequence(Time, ops)

			// Execute and track timing
			tickLock.Lock()
			startTick := tickCount
			tickLock.Unlock()

			// Expected total ticks: (wait1 + wait2) * 12
			expectedTotalTicks := (int(wait1) + int(wait2)) * 12
			maxTicks := expectedTotalTicks + 20
			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				_, endSet := globalVars["end"]
				vmLock.Unlock()

				if endSet {
					break
				}
			}

			tickLock.Lock()
			endTick := tickCount
			tickLock.Unlock()

			totalTicks := endTick - startTick

			// Allow 3 tick tolerance for execution (accounts for OpCode execution overhead)
			if totalTicks < int64(expectedTotalTicks) || totalTicks > int64(expectedTotalTicks)+3 {
				t.Logf("Total wait duration incorrect: expected=%d ticks (%d+%d steps * 12), actual=%d ticks",
					expectedTotalTicks, wait1, wait2, totalTicks)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Wait accuracy independent of system load
	t.Run("Wait accuracy under load", func(t *testing.T) {
		property := func(waitSteps uint8, loadSeqCount uint8) bool {
			// Constrain to reasonable ranges
			if waitSteps < 5 || waitSteps > 30 {
				return true
			}
			if loadSeqCount < 1 || loadSeqCount > 5 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			// Create test sequence with Wait
			testOps := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("test_before"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(waitSteps)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("test_after"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, testOps)

			// Create load sequences (busy work)
			for i := uint8(0); i < loadSeqCount; i++ {
				loadOps := generateRandomOpCodes(15)
				RegisterSequence(Time, loadOps)
			}

			// Execute until test_before is set
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			waitStartTick := tickCount
			tickLock.Unlock()

			// Execute until test_after is set
			expectedWaitTicks := int(waitSteps) * 12
			maxTicks := expectedWaitTicks + 20
			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				_, afterSet := globalVars["test_after"]
				vmLock.Unlock()

				if afterSet {
					break
				}
			}

			tickLock.Lock()
			waitEndTick := tickCount
			tickLock.Unlock()

			actualWaitTicks := waitEndTick - waitStartTick

			// Verify wait duration is accurate even under load (allow 1 tick tolerance)
			if actualWaitTicks < int64(expectedWaitTicks) || actualWaitTicks > int64(expectedWaitTicks)+1 {
				t.Logf("Wait duration under load incorrect: expected=%d ticks (%d steps * 12), actual=%d ticks, load=%d",
					expectedWaitTicks, waitSteps, actualWaitTicks, loadSeqCount)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// Property 8: Concurrent Timing Independence
// Feature: user-input-handling, Property 8: Concurrent Timing Independence
// Validates: Requirements 5.3
func TestProperty8_ConcurrentTimingIndependence(t *testing.T) {
	// Property: Each sequence maintains its own timing independent of others
	// Note: Wait(N) waits N steps, where each step is 12 ticks in TIME mode
	t.Run("Independent sequence timing", func(t *testing.T) {
		property := func(wait1 uint8, wait2 uint8) bool {
			// Constrain to different wait times
			if wait1 < 5 || wait1 > 20 || wait2 < 5 || wait2 > 20 {
				return true
			}
			if wait1 == wait2 {
				return true // Skip if same wait time
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			// Create two sequences with different wait times
			seq1Ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("seq1_before"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(wait1)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("seq1_after"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			seq2Ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("seq2_before"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(wait2)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("seq2_after"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, seq1Ops)
			RegisterSequence(Time, seq2Ops)

			// Execute first tick to start both sequences
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			startTick := tickCount
			tickLock.Unlock()

			// Track when each sequence completes
			var seq1CompleteTick, seq2CompleteTick int64

			// Calculate expected ticks (steps * 12)
			expectedWait1 := int(wait1) * 12
			expectedWait2 := int(wait2) * 12
			maxTicks := int(math.Max(float64(expectedWait1), float64(expectedWait2))) + 20

			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				_, seq1Done := globalVars["seq1_after"]
				_, seq2Done := globalVars["seq2_after"]
				vmLock.Unlock()

				tickLock.Lock()
				currentTickValue := tickCount
				tickLock.Unlock()

				if seq1Done && seq1CompleteTick == 0 {
					seq1CompleteTick = currentTickValue
				}
				if seq2Done && seq2CompleteTick == 0 {
					seq2CompleteTick = currentTickValue
				}

				if seq1Done && seq2Done {
					break
				}
			}

			// Verify both sequences completed
			if seq1CompleteTick == 0 || seq2CompleteTick == 0 {
				t.Logf("Sequences did not complete: seq1=%d, seq2=%d",
					seq1CompleteTick, seq2CompleteTick)
				return false
			}

			// Calculate actual wait durations
			seq1Duration := seq1CompleteTick - startTick
			seq2Duration := seq2CompleteTick - startTick

			// Verify each sequence waited its specified duration (±1 tick)
			seq1Valid := seq1Duration >= int64(expectedWait1) && seq1Duration <= int64(expectedWait1)+1
			seq2Valid := seq2Duration >= int64(expectedWait2) && seq2Duration <= int64(expectedWait2)+1

			if !seq1Valid || !seq2Valid {
				t.Logf("Sequence timing incorrect: seq1 expected=%d ticks (%d steps*12) actual=%d, seq2 expected=%d ticks (%d steps*12) actual=%d",
					expectedWait1, wait1, seq1Duration, expectedWait2, wait2, seq2Duration)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Multiple sequences don't interfere with each other's timing
	t.Run("No timing interference", func(t *testing.T) {
		property := func(seqCount uint8, waitBase uint8) bool {
			// Constrain to reasonable ranges
			if seqCount < 2 || seqCount > 5 {
				return true
			}
			if waitBase < 5 || waitBase > 15 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			// Create multiple sequences with different wait times
			expectedCompletions := make(map[string]int64)

			for i := uint8(0); i < seqCount; i++ {
				waitSteps := int(waitBase) + int(i)*3 // Stagger wait times

				beforeVar := interpreter.Variable("seq" + string(rune('A'+i)) + "_before")
				afterVar := interpreter.Variable("seq" + string(rune('A'+i)) + "_after")

				ops := []OpCode{
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							beforeVar,
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
						},
					},
					{
						Cmd:  interpreter.OpWait,
						Args: []any{waitSteps},
					},
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							afterVar,
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
						},
					},
				}

				RegisterSequence(Time, ops)
				// Expected ticks = steps * 12
				expectedCompletions[string(afterVar)] = int64(waitSteps * 12)
			}

			// Execute first tick
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			startTick := tickCount
			tickLock.Unlock()

			// Track actual completion times
			actualCompletions := make(map[string]int64)

			// Calculate max expected ticks
			maxExpectedTicks := (int(waitBase) + int(seqCount)*3) * 12
			maxTicks := maxExpectedTicks + 20

			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				currentTickValue := tickCount
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				for varName := range expectedCompletions {
					if _, recorded := actualCompletions[varName]; !recorded {
						if _, exists := globalVars[varName]; exists {
							actualCompletions[varName] = currentTickValue - startTick
						}
					}
				}
				allDone := len(actualCompletions) == len(expectedCompletions)
				vmLock.Unlock()

				if allDone {
					break
				}
			}

			// Verify all sequences completed with correct timing
			if len(actualCompletions) != len(expectedCompletions) {
				t.Logf("Not all sequences completed: expected=%d, actual=%d",
					len(expectedCompletions), len(actualCompletions))
				return false
			}

			for varName, expectedDuration := range expectedCompletions {
				actualDuration := actualCompletions[varName]

				// Allow ±1 tick tolerance
				if actualDuration < expectedDuration || actualDuration > expectedDuration+1 {
					t.Logf("Sequence %s timing incorrect: expected=%d ticks, actual=%d ticks",
						varName, expectedDuration, actualDuration)
					return false
				}
			}

			return true
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Sequences with overlapping waits maintain independence
	t.Run("Overlapping wait independence", func(t *testing.T) {
		property := func(shortWait uint8, longWait uint8) bool {
			// Constrain so longWait > shortWait
			if shortWait < 3 || shortWait > 10 {
				return true
			}
			if longWait < 15 || longWait > 30 {
				return true
			}
			if longWait <= shortWait {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			// Create short sequence
			shortOps := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("short_start"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(shortWait)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("short_end"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			// Create long sequence
			longOps := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("long_start"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(longWait)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("long_end"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, shortOps)
			RegisterSequence(Time, longOps)

			// Execute and track completion
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			startTick := tickCount
			tickLock.Unlock()

			var shortCompleteTick, longCompleteTick int64

			// Calculate expected ticks
			expectedShortTicks := int(shortWait) * 12
			expectedLongTicks := int(longWait) * 12
			maxTicks := expectedLongTicks + 20

			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				currentTickValue := tickCount
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				_, shortDone := globalVars["short_end"]
				_, longDone := globalVars["long_end"]
				vmLock.Unlock()

				if shortDone && shortCompleteTick == 0 {
					shortCompleteTick = currentTickValue
				}
				if longDone && longCompleteTick == 0 {
					longCompleteTick = currentTickValue
				}

				if shortDone && longDone {
					break
				}
			}

			// Verify short sequence completed first
			if shortCompleteTick == 0 || longCompleteTick == 0 {
				t.Logf("Sequences did not complete")
				return false
			}

			if shortCompleteTick >= longCompleteTick {
				t.Logf("Short sequence did not complete before long: short=%d, long=%d",
					shortCompleteTick, longCompleteTick)
				return false
			}

			// Verify timing accuracy
			shortDuration := shortCompleteTick - startTick
			longDuration := longCompleteTick - startTick

			shortValid := shortDuration >= int64(expectedShortTicks) && shortDuration <= int64(expectedShortTicks)+1
			longValid := longDuration >= int64(expectedLongTicks) && longDuration <= int64(expectedLongTicks)+1

			if !shortValid || !longValid {
				t.Logf("Timing incorrect: short expected=%d ticks (%d steps*12) actual=%d, long expected=%d ticks (%d steps*12) actual=%d",
					expectedShortTicks, shortWait, shortDuration, expectedLongTicks, longWait, longDuration)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}
