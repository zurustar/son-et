package engine

import (
	"math/rand"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestStepProgressionProperty tests Property 2: Sequencer Step Progression
// **Feature: sample-game-fixes, Property 2: Sequencer Step Progression**
// **Validates: Requirements 2.1, 2.3, 2.5**
//
// Property: For any mes() block containing step() sequences, the sequencer
// should execute each step in order without skipping or hanging on any valid
// step index (including edge cases like P14).
func TestStepProgressionProperty(t *testing.T) {
	// Test with various step counts from 0 to 100
	stepCounts := []int{0, 1, 5, 10, 14, 15, 20, 50, 100}

	for _, stepCount := range stepCounts {
		t.Run("StepCount_"+string(rune(stepCount+'0')), func(t *testing.T) {
			// Create a mes() block with stepCount steps
			// Each step will execute a simple operation (set a variable)
			commands := generateStepSequence(stepCount)

			// Create sequencer in MIDI_TIME mode (as used in y_saru)
			seq := NewSequencer(commands, MIDI_TIME, nil)
			seq.SetTicksPerStep(8) // step(8) as in y_saru

			// Create a minimal engine for execution
			engine := createMinimalEngine()
			vm := NewVM(engine.state, engine, engine.logger)

			// Execute all commands
			executedSteps := 0
			maxIterations := stepCount * 20 // Safety limit to prevent infinite loops

			for i := 0; i < maxIterations && !seq.IsComplete(); i++ {
				// Skip if waiting
				if seq.IsWaiting() {
					// Simulate tick progression
					seq.DecrementWait()
					continue
				}

				// Get current command
				cmd := seq.GetCurrentCommand()
				if cmd == nil {
					break
				}

				// Execute command
				err := vm.ExecuteOp(seq, *cmd)
				if err != nil {
					// Errors should not stop progression (resilient execution)
					t.Logf("Error at step %d: %v", executedSteps, err)
				}

				// Advance PC
				seq.IncrementPC()
				executedSteps++
			}

			// Verify sequencer completed
			if !seq.IsComplete() {
				t.Errorf("Sequencer did not complete after %d steps (expected %d)", executedSteps, stepCount)
			}

			// Verify all steps were executed
			// Each step sets a variable, so we should have stepCount variables set
			for i := 0; i < stepCount; i++ {
				varName := "step_" + string(rune(i+'0'))
				val := seq.GetVariable(varName)
				if val == 0 {
					t.Errorf("Step %d was not executed (variable %s not set)", i, varName)
				}
			}
		})
	}
}

// TestStepProgressionRandomProperty tests step progression with random step counts
// This provides broader coverage across the input space
func TestStepProgressionRandomProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(42))

	// Run 20 iterations with random step counts
	for iteration := 0; iteration < 20; iteration++ {
		stepCount := rng.Intn(100) + 1 // 1 to 100 steps

		t.Run("Random_Iteration_"+string(rune(iteration+'0')), func(t *testing.T) {
			// Create a mes() block with random step count
			commands := generateStepSequence(stepCount)

			// Create sequencer in MIDI_TIME mode
			seq := NewSequencer(commands, MIDI_TIME, nil)
			seq.SetTicksPerStep(rng.Intn(16) + 1) // Random ticks per step (1-16)

			// Create a minimal engine for execution
			engine := createMinimalEngine()
			vm := NewVM(engine.state, engine, engine.logger)

			// Execute all commands
			executedSteps := 0
			maxIterations := stepCount * 20

			for i := 0; i < maxIterations && !seq.IsComplete(); i++ {
				if seq.IsWaiting() {
					seq.DecrementWait()
					continue
				}

				cmd := seq.GetCurrentCommand()
				if cmd == nil {
					break
				}

				err := vm.ExecuteOp(seq, *cmd)
				if err != nil {
					t.Logf("Error at step %d: %v", executedSteps, err)
				}

				seq.IncrementPC()
				executedSteps++
			}

			// Verify completion
			if !seq.IsComplete() {
				t.Errorf("Sequencer did not complete with %d steps", stepCount)
			}
		})
	}
}

// generateStepSequence creates a sequence of OpCodes representing a mes() block
// with the specified number of steps. Each step sets a variable and waits.
func generateStepSequence(stepCount int) []interpreter.OpCode {
	commands := make([]interpreter.OpCode, 0, stepCount*3)

	for i := 0; i < stepCount; i++ {
		// Set a variable to mark this step as executed
		varName := "step_" + string(rune(i+'0'))
		commands = append(commands, interpreter.OpCode{
			Cmd:  interpreter.OpAssign,
			Args: []any{interpreter.Variable(varName), int64(i + 1)},
		})

		// Add a wait operation (1 step)
		commands = append(commands, interpreter.OpCode{
			Cmd:  interpreter.OpWait,
			Args: []any{int64(1)},
		})
	}

	return commands
}

// createMinimalEngine creates a minimal engine for testing
func createMinimalEngine() *Engine {
	// Create a mock renderer
	renderer := &MockRenderer{}

	// Create a mock asset loader
	assetLoader := &MockAssetLoader{}

	// Create a mock image decoder
	imageDecoder := &MockImageDecoder{}

	// Create engine
	engine := NewEngine(renderer, assetLoader, imageDecoder)
	engine.SetDebugLevel(DebugLevelError) // Minimize logging during tests

	return engine
}
