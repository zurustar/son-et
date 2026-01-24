package engine

import (
	"math/rand"
	"testing"
	"testing/quick"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestStepCompletionDetectionProperty tests Property 4: Step Block Completion Detection
// **Feature: sample-game-fixes, Property 4: Step Block Completion Detection**
// **Validates: Requirements 2.4**
//
// Property: For any mes() block, when all steps are complete, the sequencer
// should mark the block as finished and allow the engine to proceed with
// termination checks.
func TestStepCompletionDetectionProperty(t *testing.T) {
	// Property: Sequencer marks block as finished when all steps complete
	completionProperty := func(stepCount uint8) bool {
		// Limit step count to reasonable range (1-50)
		steps := int(stepCount%50) + 1

		// Create engine
		eng := NewEngine(nil, nil, nil)
		eng.SetHeadless(true)
		eng.SetTimeout(5 * time.Second)
		eng.SetDebugLevel(DebugLevelError)

		// Create sequence with steps that complete
		opcodes := generateCompletingSequence(steps)

		// Create sequencer that doesn't loop (so it can complete)
		seq := NewSequencer(opcodes, TIME, nil)
		seq.SetNoLoop(true)
		eng.RegisterSequence(seq, 0)

		// Start engine
		eng.Start()

		// Run until completion or timeout
		maxTicks := steps * 20 // Safety margin
		completed := false

		for tick := 0; tick < maxTicks; tick++ {
			err := eng.Update()
			if err == ErrTerminated {
				completed = true
				break
			}
		}

		// Verify sequence completed
		if !seq.IsComplete() {
			return false
		}

		// Verify engine terminated (all sequences complete)
		return completed
	}

	if err := quick.Check(completionProperty, &quick.Config{MaxCount: 30}); err != nil {
		t.Errorf("Step completion detection property failed: %v", err)
	}
}

// generateCompletingSequence creates a sequence that will complete after executing
func generateCompletingSequence(stepCount int) []interpreter.OpCode {
	opcodes := make([]interpreter.OpCode, 0, stepCount*2)

	for i := 0; i < stepCount; i++ {
		// Set a variable to mark progress
		opcodes = append(opcodes, interpreter.OpCode{
			Cmd:  interpreter.OpAssign,
			Args: []any{interpreter.Variable("progress"), int64(i + 1)},
		})
	}

	return opcodes
}

// TestStepBlockCompletionWithWaits tests completion detection with wait operations
func TestStepBlockCompletionWithWaits(t *testing.T) {
	testCases := []struct {
		name      string
		stepCount int
		waitSteps int
	}{
		{"1_step_1_wait", 1, 1},
		{"3_steps_1_wait", 3, 1},
		{"5_steps_2_waits", 5, 2},
		{"10_steps_3_waits", 10, 3},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewEngine(nil, nil, nil)
			engine.SetHeadless(true)
			engine.SetTimeout(5 * time.Second)
			engine.SetDebugLevel(DebugLevelError)

			// Create sequence with steps and waits
			opcodes := generateSequenceWithWaits(tc.stepCount, tc.waitSteps)

			seq := NewSequencer(opcodes, TIME, nil)
			seq.SetNoLoop(true)
			engine.RegisterSequence(seq, 0)

			engine.Start()

			// Run until completion
			maxTicks := tc.stepCount * tc.waitSteps * 10
			completed := false

			for tick := 0; tick < maxTicks; tick++ {
				err := engine.Update()
				if err == ErrTerminated {
					completed = true
					break
				}
			}

			// Verify completion
			if !seq.IsComplete() {
				t.Errorf("Sequence should be complete after all steps")
			}
			if !completed {
				t.Errorf("Engine should terminate when sequence completes")
			}

			// Verify all steps executed
			progress := seq.GetVariable("progress")
			if progress != int64(tc.stepCount) {
				t.Errorf("Expected progress=%d, got %v", tc.stepCount, progress)
			}
		})
	}
}

// generateSequenceWithWaits creates a sequence with assignments and waits
func generateSequenceWithWaits(stepCount, waitSteps int) []interpreter.OpCode {
	opcodes := make([]interpreter.OpCode, 0, stepCount*2)

	for i := 0; i < stepCount; i++ {
		// Set progress variable
		opcodes = append(opcodes, interpreter.OpCode{
			Cmd:  interpreter.OpAssign,
			Args: []any{interpreter.Variable("progress"), int64(i + 1)},
		})

		// Add wait every few steps
		if (i+1)%waitSteps == 0 && i < stepCount-1 {
			opcodes = append(opcodes, interpreter.OpCode{
				Cmd:  interpreter.OpWait,
				Args: []any{int64(1)},
			})
		}
	}

	return opcodes
}

// TestMesBlockCompletionTriggerTermination tests that mes() block completion
// triggers engine termination checks
func TestMesBlockCompletionTriggerTermination(t *testing.T) {
	engine := NewEngine(nil, nil, nil)
	engine.SetHeadless(true)
	engine.SetTimeout(5 * time.Second)
	engine.SetDebugLevel(DebugLevelError)

	// Create a simple sequence that completes quickly
	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("z"), int64(3)}},
	}

	seq := NewSequencer(opcodes, TIME, nil)
	seq.SetNoLoop(true)
	engine.RegisterSequence(seq, 0)

	engine.Start()

	// Run a few ticks
	terminated := false
	for tick := 0; tick < 10; tick++ {
		err := engine.Update()
		if err == ErrTerminated {
			terminated = true
			break
		}
	}

	// Verify termination occurred
	if !terminated {
		t.Error("Engine should terminate when all sequences complete")
	}
	if !engine.IsTerminated() {
		t.Error("Engine.IsTerminated() should return true")
	}
}

// TestMultipleMesBlocksCompletion tests completion detection with multiple sequences
func TestMultipleMesBlocksCompletion(t *testing.T) {
	engine := NewEngine(nil, nil, nil)
	engine.SetHeadless(true)
	engine.SetTimeout(5 * time.Second)
	engine.SetDebugLevel(DebugLevelError)

	// Create two sequences with different lengths
	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("a"), int64(1)}},
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("b"), int64(2)}},
	}

	opcodes2 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("c"), int64(3)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // Wait 2 steps = 6 ticks
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("d"), int64(4)}},
	}

	seq1 := NewSequencer(opcodes1, TIME, nil)
	seq1.SetNoLoop(true)
	seq2 := NewSequencer(opcodes2, TIME, nil)
	seq2.SetNoLoop(true)

	engine.RegisterSequence(seq1, 0)
	engine.RegisterSequence(seq2, 0)

	engine.Start()

	// Run until both complete
	terminated := false
	for tick := 0; tick < 20; tick++ {
		err := engine.Update()
		if err == ErrTerminated {
			terminated = true
			break
		}
	}

	// Verify both sequences completed
	if !seq1.IsComplete() {
		t.Error("Sequence 1 should be complete")
	}
	if !seq2.IsComplete() {
		t.Error("Sequence 2 should be complete")
	}
	if !terminated {
		t.Error("Engine should terminate when all sequences complete")
	}
}

// TestStepCompletionRandomProperty tests completion with random step counts
func TestStepCompletionRandomProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	// Run 20 iterations with random parameters
	for iteration := 0; iteration < 20; iteration++ {
		stepCount := rng.Intn(20) + 1   // 1-20 steps
		mode := TimingMode(rng.Intn(2)) // TIME or MIDI_TIME

		t.Run("Random_Iteration", func(t *testing.T) {
			engine := NewEngine(nil, nil, nil)
			engine.SetHeadless(true)
			engine.SetTimeout(5 * time.Second)
			engine.SetDebugLevel(DebugLevelError)

			// Create sequence
			opcodes := generateCompletingSequence(stepCount)

			seq := NewSequencer(opcodes, mode, nil)
			seq.SetNoLoop(true)
			engine.RegisterSequence(seq, 0)

			engine.Start()

			// Run until completion
			maxTicks := stepCount * 10
			completed := false

			for tick := 0; tick < maxTicks; tick++ {
				err := engine.Update()
				if err == ErrTerminated {
					completed = true
					break
				}
			}

			// Verify completion
			if !seq.IsComplete() {
				t.Errorf("Sequence should be complete (mode=%d, steps=%d)", mode, stepCount)
			}
			if !completed {
				t.Errorf("Engine should terminate (mode=%d, steps=%d)", mode, stepCount)
			}
		})
	}
}

// TestSequencerIsCompleteMethod tests the IsComplete() method directly
func TestSequencerIsCompleteMethod(t *testing.T) {
	t.Run("Empty sequence is complete", func(t *testing.T) {
		seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
		if !seq.IsComplete() {
			t.Error("Empty sequence should be complete")
		}
	})

	t.Run("Sequence with commands not complete initially", func(t *testing.T) {
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		if seq.IsComplete() {
			t.Error("Sequence with commands should not be complete initially")
		}
	})

	t.Run("Sequence complete when PC reaches end", func(t *testing.T) {
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		seq.IncrementPC() // Move past the only command
		if !seq.IsComplete() {
			t.Error("Sequence should be complete when PC >= len(commands)")
		}
	})

	t.Run("Waiting sequence not complete", func(t *testing.T) {
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		seq.IncrementPC() // Move past the command
		seq.SetWait(5)    // But set a wait
		if seq.IsComplete() {
			t.Error("Waiting sequence should not be complete even if PC at end")
		}
	})
}

// TestAllSequencesCompleteMethod tests the AllSequencesComplete() engine method
func TestAllSequencesCompleteMethod(t *testing.T) {
	t.Run("No sequences means complete", func(t *testing.T) {
		engine := NewEngine(nil, nil, nil)
		if !engine.AllSequencesComplete() {
			t.Error("Engine with no sequences should report all complete")
		}
	})

	t.Run("Active incomplete sequence means not complete", func(t *testing.T) {
		engine := NewEngine(nil, nil, nil)
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		seq.SetNoLoop(true)
		engine.RegisterSequence(seq, 0)

		if engine.AllSequencesComplete() {
			t.Error("Engine with incomplete sequence should not report all complete")
		}
	})

	t.Run("All sequences complete means complete", func(t *testing.T) {
		engine := NewEngine(nil, nil, nil)
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		seq.SetNoLoop(true)
		engine.RegisterSequence(seq, 0)

		// Complete the sequence
		seq.IncrementPC()

		if !engine.AllSequencesComplete() {
			t.Error("Engine with all sequences complete should report all complete")
		}
	})

	t.Run("Inactive sequence doesn't block completion", func(t *testing.T) {
		engine := NewEngine(nil, nil, nil)
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		seq.SetNoLoop(true)
		engine.RegisterSequence(seq, 0)

		// Deactivate the sequence
		seq.Deactivate()

		if !engine.AllSequencesComplete() {
			t.Error("Engine with only inactive sequences should report all complete")
		}
	})
}
