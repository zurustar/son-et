package engine

import (
	"sync"
	"testing"
	"testing/quick"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Property 4: Concurrent Sequence Execution
// Feature: user-input-handling, Property 4: Concurrent Sequence Execution
// Validates: Requirements 1.4
func TestProperty4_ConcurrentSequenceExecution(t *testing.T) {
	// Property: Multiple mes(TIME) blocks execute in parallel without blocking
	t.Run("Multiple sequences execute concurrently", func(t *testing.T) {
		property := func(seqCount uint8) bool {
			// Constrain to reasonable range (2-10 sequences)
			if seqCount < 2 || seqCount > 10 {
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

			// Track registration times
			registrationTimes := make([]time.Duration, seqCount)
			var registrationMutex sync.Mutex

			// Register multiple sequences concurrently
			var wg sync.WaitGroup
			start := time.Now()

			for i := uint8(0); i < seqCount; i++ {
				wg.Add(1)
				go func(idx uint8) {
					defer wg.Done()

					// Use lowercase letters for variable names
					seqLetter := string(rune('a' + idx))

					// Create sequence with some operations
					ops := []OpCode{
						{
							Cmd: interpreter.OpAssign,
							Args: []any{
								interpreter.Variable("seq" + seqLetter + "_start"),
								OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
							},
						},
						{
							Cmd:  interpreter.OpWait,
							Args: []any{5},
						},
						{
							Cmd: interpreter.OpAssign,
							Args: []any{
								interpreter.Variable("seq" + seqLetter + "_end"),
								OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
							},
						},
					}

					regStart := time.Now()
					RegisterSequence(Time, ops)
					regDuration := time.Since(regStart)

					registrationMutex.Lock()
					registrationTimes[idx] = regDuration
					registrationMutex.Unlock()
				}(i)
			}

			// Wait for all registrations to complete
			wg.Wait()
			totalRegistrationTime := time.Since(start)

			// Verify all registrations were non-blocking (< 10ms each)
			for i, regTime := range registrationTimes {
				if regTime >= 10*time.Millisecond {
					t.Logf("Sequence %d registration took too long: %v", i, regTime)
					return false
				}
			}

			// Verify total registration time is reasonable
			// If sequences were blocking, total time would be seqCount * 10ms
			// With non-blocking, total time should be < 50ms regardless of count
			if totalRegistrationTime >= 50*time.Millisecond {
				t.Logf("Total registration time too long: %v for %d sequences", totalRegistrationTime, seqCount)
				return false
			}

			// Verify all sequences are registered and active
			vmLock.Lock()
			activeCount := 0
			for _, seq := range sequencers {
				if seq != nil && seq.active {
					activeCount++
				}
			}
			vmLock.Unlock()

			if activeCount != int(seqCount) {
				t.Logf("Not all sequences are active: expected=%d, actual=%d", seqCount, activeCount)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Concurrent sequences execute in parallel without blocking each other
	t.Run("Sequences execute in parallel", func(t *testing.T) {
		property := func(seqCount uint8) bool {
			// Constrain to reasonable range (2-8 sequences)
			if seqCount < 2 || seqCount > 8 {
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

			// Register multiple sequences with different wait times
			for i := uint8(0); i < seqCount; i++ {
				waitSteps := 5 + int(i)*2 // Stagger wait times

				// Use lowercase letters for variable names
				seqLetter := string(rune('a' + i))

				ops := []OpCode{
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							interpreter.Variable("seq" + seqLetter + "_start"),
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
							interpreter.Variable("seq" + seqLetter + "_end"),
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
						},
					},
				}

				RegisterSequence(Time, ops)
			}

			// Execute VM for enough ticks to complete all sequences
			// First tick to start all sequences
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Calculate max wait time needed
			maxWaitSteps := 5 + int(seqCount-1)*2
			maxTicks := maxWaitSteps*12 + 20

			// Execute remaining ticks
			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)
			}

			// Verify all sequences completed
			vmLock.Lock()
			completedCount := 0
			for i := uint8(0); i < seqCount; i++ {
				seqLetter := string(rune('a' + i))
				endVar := "seq" + seqLetter + "_end"
				if _, exists := globalVars[endVar]; exists {
					completedCount++
				}
			}
			vmLock.Unlock()

			if completedCount != int(seqCount) {
				t.Logf("Not all sequences completed: expected=%d, actual=%d", seqCount, completedCount)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: No blocking between concurrent sequences
	t.Run("No blocking between sequences", func(t *testing.T) {
		property := func(fastWait uint8, slowWait uint8) bool {
			// Constrain to ensure fast < slow
			if fastWait < 3 || fastWait > 10 {
				return true
			}
			if slowWait < 15 || slowWait > 30 {
				return true
			}
			if slowWait <= fastWait {
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

			// Register fast sequence
			fastOps := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("fast_start"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(fastWait)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("fast_end"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			// Register slow sequence
			slowOps := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("slow_start"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}},
					},
				},
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(slowWait)},
				},
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("slow_end"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, fastOps)
			RegisterSequence(Time, slowOps)

			// Execute and track when fast sequence completes
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			startTick := tickCount
			tickLock.Unlock()

			var fastCompleteTick int64
			expectedFastTicks := int(fastWait) * 12
			maxTicks := int(slowWait)*12 + 20

			for i := 0; i < maxTicks; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				currentTickValue := tickCount
				tickLock.Unlock()

				UpdateVM(currentTick)

				vmLock.Lock()
				_, fastDone := globalVars["fast_end"]
				_, slowDone := globalVars["slow_end"]
				vmLock.Unlock()

				if fastDone && fastCompleteTick == 0 {
					fastCompleteTick = currentTickValue
				}

				if fastDone && slowDone {
					break
				}
			}

			// Verify fast sequence completed at expected time
			// (not blocked by slow sequence)
			if fastCompleteTick == 0 {
				t.Logf("Fast sequence did not complete")
				return false
			}

			fastDuration := fastCompleteTick - startTick

			// Allow ±1 tick tolerance
			if fastDuration < int64(expectedFastTicks) || fastDuration > int64(expectedFastTicks)+1 {
				t.Logf("Fast sequence blocked: expected=%d ticks (%d steps*12), actual=%d ticks",
					expectedFastTicks, fastWait, fastDuration)
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

// Property 10: Update Advances VM
// Feature: user-input-handling, Property 10: Update Advances VM
// Validates: Requirements 7.1
func TestProperty10_UpdateAdvancesVM(t *testing.T) {
	// Property: Each Game.Update() call increments tick count by exactly 1 in TIME mode
	t.Run("Update increments tick by 1", func(t *testing.T) {
		property := func(updateCount uint8) bool {
			// Constrain to reasonable range (5-100 updates)
			if updateCount < 5 || updateCount > 100 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			midiSyncMode = false // Ensure TIME mode
			vmLock.Unlock()

			// Create a simple Game instance
			game := &Game{
				tickCount: 0,
			}

			// Track tick values
			previousTick := int64(0)
			allIncrementsAreOne := true

			for i := uint8(0); i < updateCount; i++ {
				// Call Update (simulates game loop)
				_ = game.Update()

				// Check tick count
				tickLock.Lock()
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

	// Property: Tick count advances consistently across multiple sequences
	t.Run("Tick advances with active sequences", func(t *testing.T) {
		property := func(seqCount uint8, updateCount uint8) bool {
			// Constrain to reasonable ranges
			if seqCount < 1 || seqCount > 5 {
				return true
			}
			if updateCount < 10 || updateCount > 50 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			midiSyncMode = false // Ensure TIME mode
			vmLock.Unlock()

			// Register multiple sequences
			for i := uint8(0); i < seqCount; i++ {
				ops := generateRandomOpCodes(10)
				RegisterSequence(Time, ops)
			}

			// Create a simple Game instance
			game := &Game{
				tickCount: 0,
			}

			// Track tick increments
			previousTick := int64(0)
			allIncrementsAreOne := true

			for i := uint8(0); i < updateCount; i++ {
				// Call Update
				_ = game.Update()

				// Check tick count
				tickLock.Lock()
				currentTick := tickCount
				tickLock.Unlock()

				// Verify increment is exactly 1
				increment := currentTick - previousTick
				if increment != 1 {
					t.Logf("Tick increment not 1 with %d sequences: previous=%d, current=%d, increment=%d",
						seqCount, previousTick, currentTick, increment)
					allIncrementsAreOne = false
					break
				}

				previousTick = currentTick
			}

			return allIncrementsAreOne
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Tick count is independent of sequence complexity
	t.Run("Tick advances independent of complexity", func(t *testing.T) {
		property := func(opsPerSeq uint8) bool {
			// Constrain to reasonable range (5-50 operations)
			if opsPerSeq < 5 || opsPerSeq > 50 {
				return true
			}

			// Reset state
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			midiSyncMode = false // Ensure TIME mode
			vmLock.Unlock()

			// Register sequence with varying complexity
			ops := generateRandomOpCodes(int(opsPerSeq))
			RegisterSequence(Time, ops)

			// Create a simple Game instance
			game := &Game{
				tickCount: 0,
			}

			// Execute 20 updates
			previousTick := int64(0)
			allIncrementsAreOne := true

			for i := 0; i < 20; i++ {
				// Call Update
				_ = game.Update()

				// Check tick count
				tickLock.Lock()
				currentTick := tickCount
				tickLock.Unlock()

				// Verify increment is exactly 1
				increment := currentTick - previousTick
				if increment != 1 {
					t.Logf("Tick increment not 1 with %d ops: previous=%d, current=%d, increment=%d",
						opsPerSeq, previousTick, currentTick, increment)
					allIncrementsAreOne = false
					break
				}

				previousTick = currentTick
			}

			return allIncrementsAreOne
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// Property 11: UpdateVM Non-blocking
// Feature: user-input-handling, Property 11: UpdateVM Non-blocking
// Validates: Requirements 7.3
func TestProperty11_UpdateVMNonBlocking(t *testing.T) {
	// Property: UpdateVM execution time is < 5ms for typical sequences
	t.Run("UpdateVM completes quickly", func(t *testing.T) {
		property := func(seqCount uint8, opsPerSeq uint8) bool {
			// Constrain to reasonable ranges
			if seqCount < 1 || seqCount > 10 {
				return true
			}
			if opsPerSeq < 5 || opsPerSeq > 30 {
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

			// Register multiple sequences
			for i := uint8(0); i < seqCount; i++ {
				ops := generateRandomOpCodes(int(opsPerSeq))
				RegisterSequence(Time, ops)
			}

			// Measure UpdateVM execution time
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()

			start := time.Now()
			UpdateVM(currentTick)
			elapsed := time.Since(start)

			// Verify execution time is < 5ms
			if elapsed >= 5*time.Millisecond {
				t.Logf("UpdateVM took too long: %v with %d sequences of %d ops each",
					elapsed, seqCount, opsPerSeq)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: UpdateVM execution time scales linearly with sequence count
	t.Run("UpdateVM scales linearly", func(t *testing.T) {
		property := func(seqCount uint8) bool {
			// Constrain to reasonable range (1-15 sequences)
			if seqCount < 1 || seqCount > 15 {
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

			// Register sequences
			for i := uint8(0); i < seqCount; i++ {
				ops := generateRandomOpCodes(10)
				RegisterSequence(Time, ops)
			}

			// Measure UpdateVM execution time
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()

			start := time.Now()
			UpdateVM(currentTick)
			elapsed := time.Since(start)

			// Verify execution time is reasonable (< 1ms per sequence)
			maxExpected := time.Duration(seqCount) * time.Millisecond
			if elapsed >= maxExpected {
				t.Logf("UpdateVM scaling issue: %v for %d sequences (expected < %v)",
					elapsed, seqCount, maxExpected)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: UpdateVM with Wait operations is non-blocking
	t.Run("UpdateVM non-blocking with Wait", func(t *testing.T) {
		property := func(waitSteps uint8) bool {
			// Constrain to reasonable range (5-50 steps)
			if waitSteps < 5 || waitSteps > 50 {
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

			// Execute first tick (starts wait)
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Measure UpdateVM execution time during wait
			tickLock.Lock()
			tickCount++
			currentTick = int(tickCount)
			tickLock.Unlock()

			start := time.Now()
			UpdateVM(currentTick)
			elapsed := time.Since(start)

			// Verify UpdateVM is non-blocking even during wait
			// Should complete in < 1ms since it just decrements waitTicks
			if elapsed >= time.Millisecond {
				t.Logf("UpdateVM blocked during Wait(%d): %v", waitSteps, elapsed)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: UpdateVM execution time independent of wait duration
	t.Run("UpdateVM time independent of wait duration", func(t *testing.T) {
		property := func(shortWait uint8, longWait uint8) bool {
			// Constrain to ensure different wait times
			if shortWait < 5 || shortWait > 15 {
				return true
			}
			if longWait < 30 || longWait > 60 {
				return true
			}

			// Test short wait
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			shortOps := []OpCode{
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(shortWait)},
				},
			}
			RegisterSequence(Time, shortOps)

			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			tickCount++
			currentTick = int(tickCount)
			tickLock.Unlock()

			startShort := time.Now()
			UpdateVM(currentTick)
			shortElapsed := time.Since(startShort)

			// Test long wait
			tickLock.Lock()
			tickCount = 0
			tickLock.Unlock()

			vmLock.Lock()
			sequencers = nil
			globalVars = make(map[string]any)
			vmLock.Unlock()

			longOps := []OpCode{
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int(longWait)},
				},
			}
			RegisterSequence(Time, longOps)

			tickLock.Lock()
			tickCount++
			currentTick = int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			tickLock.Lock()
			tickCount++
			currentTick = int(tickCount)
			tickLock.Unlock()

			startLong := time.Now()
			UpdateVM(currentTick)
			longElapsed := time.Since(startLong)

			// Verify both execution times are similar (within 2x)
			// Wait duration should not affect UpdateVM execution time
			ratio := float64(longElapsed) / float64(shortElapsed)
			if ratio > 2.0 || ratio < 0.5 {
				t.Logf("UpdateVM time varies with wait duration: short=%v (wait=%d), long=%v (wait=%d), ratio=%.2f",
					shortElapsed, shortWait, longElapsed, longWait, ratio)
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
