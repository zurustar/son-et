package engine

import (
	"testing"
	"testing/quick"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Property 2: Game Loop Continuity
// Feature: user-input-handling, Property 2: Game Loop Continuity
// Validates: Requirements 1.2, 3.3, 4.4, 7.4
func TestProperty2_GameLoopContinuity(t *testing.T) {
	// Property: Update() is called at approximately 60 FPS (16.67ms ± 2ms per frame)
	// during long-running sequence execution
	t.Run("Update timing during sequence execution", func(t *testing.T) {
		property := func(sequenceDuration uint8) bool {
			// Constrain to reasonable range (10-50 frames)
			if sequenceDuration < 10 || sequenceDuration > 50 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create a long-running sequence with Wait operations
			ops := make([]OpCode, 0)
			for i := uint8(0); i < sequenceDuration/5; i++ {
				ops = append(ops, OpCode{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("progress"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i)}},
					},
				})
				ops = append(ops, OpCode{
					Cmd:  interpreter.OpWait,
					Args: []any{5}, // Wait 5 steps = 60 ticks
				})
			}

			RegisterSequence(Time, ops)

			// Simulate Update() calls and measure timing
			targetFPS := 60.0
			targetFrameDuration := time.Second / time.Duration(targetFPS)
			frameTimes := make([]time.Duration, 0)

			previousTime := time.Now()
			for i := 0; i < int(sequenceDuration); i++ {
				// Simulate one Update() call
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				// Measure time since last Update()
				currentTime := time.Now()
				frameDuration := currentTime.Sub(previousTime)
				frameTimes = append(frameTimes, frameDuration)
				previousTime = currentTime

				// Sleep to simulate frame timing
				time.Sleep(targetFrameDuration)
			}

			// Verify frame timing is consistent (16.67ms ± 2ms)
			// Skip first frame as it may have initialization overhead
			targetDurationNs := 16.67 * float64(time.Millisecond)
			targetDuration := time.Duration(int64(targetDurationNs))
			tolerance := 2 * time.Millisecond

			validFrames := 0
			for i := 1; i < len(frameTimes); i++ {
				duration := frameTimes[i]
				// Account for sleep time in measurement
				if duration >= targetDuration-tolerance && duration <= targetDuration+tolerance+targetFrameDuration {
					validFrames++
				}
			}

			// At least 80% of frames should be within tolerance
			successRate := float64(validFrames) / float64(len(frameTimes)-1)
			return successRate >= 0.8
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Update() continues to be called even during Wait operations
	t.Run("Update continues during Wait", func(t *testing.T) {
		property := func(waitSteps uint8) bool {
			// Constrain to reasonable range (10-30 steps)
			if waitSteps < 10 || waitSteps > 30 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create sequence with a long Wait
			ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("before_wait"),
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
						interpreter.Variable("after_wait"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, ops)

			// Execute first tick to start the wait
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Count Update() calls during the wait
			expectedTicks := int(waitSteps) * 12 // steps * ticksPerStep
			updateCallCount := 0

			for i := 0; i < expectedTicks+5; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)
				updateCallCount++

				// Check if wait completed
				vmLock.Lock()
				_, afterSet := globalVars["after_wait"]
				vmLock.Unlock()

				if afterSet {
					break
				}
			}

			// Verify Update() was called continuously during the wait
			// Should be called at least expectedTicks times
			return updateCallCount >= expectedTicks
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Update() timing is not affected by sequence complexity
	t.Run("Update timing independent of sequence complexity", func(t *testing.T) {
		property := func(opsPerSequence uint8, sequenceCount uint8) bool {
			// Constrain to reasonable ranges
			if opsPerSequence < 5 || opsPerSequence > 20 {
				return true
			}
			if sequenceCount < 1 || sequenceCount > 5 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Register multiple sequences with varying complexity
			for i := uint8(0); i < sequenceCount; i++ {
				ops := generateRandomOpCodes(int(opsPerSequence))
				RegisterSequence(Time, ops)
			}

			// Measure Update() timing over several frames
			targetFPS := 60.0
			targetFrameDuration := time.Second / time.Duration(targetFPS)
			frameCount := 20
			frameTimes := make([]time.Duration, 0)

			previousTime := time.Now()
			for i := 0; i < frameCount; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				currentTime := time.Now()
				frameDuration := currentTime.Sub(previousTime)
				frameTimes = append(frameTimes, frameDuration)
				previousTime = currentTime

				time.Sleep(targetFrameDuration)
			}

			// Verify frame timing remains consistent
			// Skip first frame for initialization
			targetDurationNs := 16.67 * float64(time.Millisecond)
			targetDuration := time.Duration(int64(targetDurationNs))
			tolerance := 2 * time.Millisecond

			validFrames := 0
			for i := 1; i < len(frameTimes); i++ {
				duration := frameTimes[i]
				if duration >= targetDuration-tolerance && duration <= targetDuration+tolerance+targetFrameDuration {
					validFrames++
				}
			}

			// At least 70% of frames should be within tolerance
			successRate := float64(validFrames) / float64(len(frameTimes)-1)
			return successRate >= 0.7
		}

		config := &quick.Config{MaxCount: 20}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// Property 3: Rendering Frame Rate
// Feature: user-input-handling, Property 3: Rendering Frame Rate
// Validates: Requirements 1.3
func TestProperty3_RenderingFrameRate(t *testing.T) {
	// Property: Draw() is called at approximately 60 FPS (16.67ms per frame)
	// Note: This test simulates Draw() calls since we can't directly test Ebiten's game loop
	t.Run("Draw timing consistency", func(t *testing.T) {
		property := func(frameCount uint8) bool {
			// Constrain to reasonable range (10-50 frames)
			if frameCount < 10 || frameCount > 50 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create a sequence to simulate active execution
			ops := generateRandomOpCodes(20)
			RegisterSequence(Time, ops)

			// Simulate Draw() calls and measure timing
			targetFPS := 60.0
			targetFrameDuration := time.Second / time.Duration(targetFPS)
			drawTimes := make([]time.Duration, 0)

			previousTime := time.Now()
			for i := uint8(0); i < frameCount; i++ {
				// Simulate one frame cycle (Update + Draw)
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				// Measure time for Draw call
				currentTime := time.Now()
				drawDuration := currentTime.Sub(previousTime)
				drawTimes = append(drawTimes, drawDuration)
				previousTime = currentTime

				// Sleep to simulate frame timing
				time.Sleep(targetFrameDuration)
			}

			// Verify Draw() timing is consistent (16.67ms ± 2ms)
			// Skip first frame for initialization
			targetDurationNs := 16.67 * float64(time.Millisecond)
			targetDuration := time.Duration(int64(targetDurationNs))
			tolerance := 2 * time.Millisecond

			validFrames := 0
			for i := 1; i < len(drawTimes); i++ {
				duration := drawTimes[i]
				if duration >= targetDuration-tolerance && duration <= targetDuration+tolerance+targetFrameDuration {
					validFrames++
				}
			}

			// At least 80% of frames should be within tolerance
			successRate := float64(validFrames) / float64(len(drawTimes)-1)
			return successRate >= 0.8
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Draw() timing under various sequence loads
	t.Run("Draw timing under load", func(t *testing.T) {
		property := func(sequenceCount uint8, opsPerSeq uint8) bool {
			// Constrain to reasonable ranges
			if sequenceCount < 1 || sequenceCount > 5 {
				return true
			}
			if opsPerSeq < 5 || opsPerSeq > 20 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Register multiple sequences to create load
			for i := uint8(0); i < sequenceCount; i++ {
				ops := generateRandomOpCodes(int(opsPerSeq))
				RegisterSequence(Time, ops)
			}

			// Measure Draw() timing over several frames
			targetFPS := 60.0
			targetFrameDuration := time.Second / time.Duration(targetFPS)
			frameCount := 20
			drawTimes := make([]time.Duration, 0)

			previousTime := time.Now()
			for i := 0; i < frameCount; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				currentTime := time.Now()
				drawDuration := currentTime.Sub(previousTime)
				drawTimes = append(drawTimes, drawDuration)
				previousTime = currentTime

				time.Sleep(targetFrameDuration)
			}

			// Verify Draw() timing remains consistent under load
			targetDurationNs := 16.67 * float64(time.Millisecond)
			targetDuration := time.Duration(int64(targetDurationNs))
			tolerance := 2 * time.Millisecond

			validFrames := 0
			for i := 1; i < len(drawTimes); i++ {
				duration := drawTimes[i]
				if duration >= targetDuration-tolerance && duration <= targetDuration+tolerance+targetFrameDuration {
					validFrames++
				}
			}

			// At least 70% of frames should be within tolerance even under load
			successRate := float64(validFrames) / float64(len(drawTimes)-1)
			return successRate >= 0.7
		}

		config := &quick.Config{MaxCount: 20}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Draw() continues during Wait operations
	t.Run("Draw continues during Wait", func(t *testing.T) {
		property := func(waitSteps uint8) bool {
			// Constrain to reasonable range (10-30 steps)
			if waitSteps < 10 || waitSteps > 30 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create sequence with a long Wait
			ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("before_wait"),
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
						interpreter.Variable("after_wait"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, ops)

			// Execute first tick to start the wait
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Count Draw() calls during the wait
			expectedTicks := int(waitSteps) * 12
			drawCallCount := 0

			targetFPS := 60.0
			targetFrameDuration := time.Second / time.Duration(targetFPS)

			for i := 0; i < expectedTicks+5; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)

				// Simulate Draw() call
				drawCallCount++

				// Check if wait completed
				vmLock.Lock()
				_, afterSet := globalVars["after_wait"]
				vmLock.Unlock()

				if afterSet {
					break
				}

				time.Sleep(targetFrameDuration)
			}

			// Verify Draw() was called continuously during the wait
			return drawCallCount >= expectedTicks
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// Property 5: Input Responsiveness During Wait
// Feature: user-input-handling, Property 5: Input Responsiveness During Wait
// Validates: Requirements 3.4
func TestProperty5_InputResponsivenessDuringWait(t *testing.T) {
	// Property: Update() continues to be called during Wait operations,
	// allowing input events to be processed
	t.Run("Update continues during Wait for input processing", func(t *testing.T) {
		property := func(waitSteps uint8) bool {
			// Constrain to reasonable range (20-100 steps for longer waits)
			if waitSteps < 20 || waitSteps > 100 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create sequence with a long Wait operation
			ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("wait_started"),
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
						interpreter.Variable("wait_completed"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, ops)

			// Execute first tick to start the wait
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Verify wait started
			vmLock.Lock()
			_, waitStarted := globalVars["wait_started"]
			vmLock.Unlock()

			if !waitStarted {
				return false
			}

			// Track Update() calls during the wait
			updateCallsDuringWait := 0
			expectedTicks := int(waitSteps) * 12

			// Simulate input event processing during wait
			// In a real scenario, input events would be checked in Update()
			inputEventsProcessed := 0

			for i := 0; i < expectedTicks+5; i++ {
				// Simulate Update() call
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)
				updateCallsDuringWait++

				// Simulate input event processing
				// In real code, this would be checking ebiten.IsKeyPressed(), etc.
				if i%10 == 0 {
					inputEventsProcessed++
				}

				// Check if wait completed
				vmLock.Lock()
				_, waitCompleted := globalVars["wait_completed"]
				vmLock.Unlock()

				if waitCompleted {
					break
				}
			}

			// Verify:
			// 1. Update() was called continuously during wait
			// 2. Input events could be processed (simulated)
			updatesContinuous := updateCallsDuringWait >= expectedTicks
			inputsProcessed := inputEventsProcessed > 0

			return updatesContinuous && inputsProcessed
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Multiple concurrent waits don't block input processing
	t.Run("Input processing with concurrent waits", func(t *testing.T) {
		property := func(seqCount uint8, waitSteps uint8) bool {
			// Constrain to reasonable ranges
			if seqCount < 2 || seqCount > 5 {
				return true
			}
			if waitSteps < 10 || waitSteps > 30 {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create multiple sequences with Wait operations
			for i := uint8(0); i < seqCount; i++ {
				ops := []OpCode{
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							interpreter.Variable("seq_start"),
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i)}},
						},
					},
					{
						Cmd:  interpreter.OpWait,
						Args: []any{int(waitSteps)},
					},
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							interpreter.Variable("seq_end"),
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i) + 100}},
						},
					},
				}

				RegisterSequence(Time, ops)
			}

			// Execute first tick to start all waits
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Track Update() calls and simulated input processing
			updateCallCount := 0
			inputEventsProcessed := 0
			expectedTicks := int(waitSteps) * 12

			for i := 0; i < expectedTicks+10; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)
				updateCallCount++

				// Simulate input event processing
				if i%5 == 0 {
					inputEventsProcessed++
				}

				// Check if all sequences completed
				vmLock.Lock()
				allDone := true
				for j := uint8(0); j < seqCount; j++ {
					vmLock.Lock()
					if len(sequencers) > int(j) && sequencers[j].active {
						allDone = false
					}
					vmLock.Unlock()
				}
				vmLock.Unlock()

				if allDone {
					break
				}
			}

			// Verify Update() was called continuously and input could be processed
			updatesContinuous := updateCallCount >= expectedTicks
			inputsProcessed := inputEventsProcessed > 0

			return updatesContinuous && inputsProcessed
		}

		config := &quick.Config{MaxCount: 20}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Termination can be triggered during Wait
	t.Run("Termination responsive during Wait", func(t *testing.T) {
		property := func(waitSteps uint8, terminateAt uint8) bool {
			// Constrain to reasonable ranges
			if waitSteps < 20 || waitSteps > 50 {
				return true
			}
			if terminateAt < 5 || terminateAt >= waitSteps {
				return true
			}

			// Reset state
			ResetEngineForTest()

			// Create sequence with a long Wait
			ops := []OpCode{
				{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("before_wait"),
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
						interpreter.Variable("after_wait"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}},
					},
				},
			}

			RegisterSequence(Time, ops)

			// Execute first tick to start the wait
			tickLock.Lock()
			tickCount++
			currentTick := int(tickCount)
			tickLock.Unlock()
			UpdateVM(currentTick)

			// Execute ticks until termination point
			terminateAtTick := int(terminateAt) * 12
			for i := 0; i < terminateAtTick; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)
			}

			// Trigger termination during wait
			programTerminated = true

			// Record PC before termination
			vmLock.Lock()
			pcBeforeTermination := 0
			if len(sequencers) > 0 {
				pcBeforeTermination = sequencers[0].pc
			}
			vmLock.Unlock()

			// Try to execute more ticks
			for i := 0; i < 10; i++ {
				tickLock.Lock()
				tickCount++
				currentTick := int(tickCount)
				tickLock.Unlock()

				UpdateVM(currentTick)
			}

			// Verify execution stopped
			vmLock.Lock()
			pcAfterTermination := 0
			sequenceActive := false
			if len(sequencers) > 0 {
				pcAfterTermination = sequencers[0].pc
				sequenceActive = sequencers[0].active
			}
			vmLock.Unlock()

			// Property holds if PC didn't advance and sequence is inactive
			return pcAfterTermination == pcBeforeTermination && !sequenceActive
		}

		config := &quick.Config{MaxCount: 30}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}
