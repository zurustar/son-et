package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestStepBlockExecution(t *testing.T) {
	tests := []struct {
		name          string
		opcodes       []interpreter.OpCode
		expectedSteps int
		expectedX     int64
		description   string
	}{
		{
			name: "simple step block with 3 commands",
			opcodes: []interpreter.OpCode{
				// Flat sequence: cmd1, wait, cmd2, wait, cmd3, wait
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(5)}},
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(2)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(5)}},
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(3)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(5)}},
			},
			expectedSteps: 45, // 3 waits × 5 steps × 3 ticks/step
			expectedX:     3,
			description:   "Step block should execute commands with waits between them",
		},
		{
			name: "step block with empty steps (from commas)",
			opcodes: []interpreter.OpCode{
				// cmd1, wait, empty wait, cmd2, wait
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(2)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(2)}}, // Empty step
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(2)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(2)}},
			},
			expectedSteps: 18, // 3 waits × 2 steps × 3 ticks/step
			expectedX:     2,
			description:   "Empty steps (from commas) should just add waits",
		},
		{
			name: "step block ending early (end_step handled in codegen)",
			opcodes: []interpreter.OpCode{
				// Only first command, then sequence ends
				{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
				{Cmd: interpreter.OpWait, Args: []any{int64(5)}},
				// end_step in codegen stops generating more opcodes
			},
			expectedSteps: 15, // 1 wait × 5 steps × 3 ticks/step
			expectedX:     1,
			description:   "end_step in source stops codegen from adding more commands",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create engine with mock renderer
			logger := NewLogger(0) // No debug output
			mockRenderer := &MockRenderer{}
			mockAssetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
			mockImageDecoder := &MockImageDecoder{}
			state := NewEngineState(mockRenderer, mockAssetLoader, mockImageDecoder)
			engine := &Engine{
				state:  state,
				logger: logger,
			}
			vm := NewVM(state, engine, logger)

			// Create sequencer
			seq := NewSequencer(tt.opcodes, TIME, nil)

			// Execute opcodes until sequence completes or reaches expected steps
			stepCount := 0
			maxSteps := tt.expectedSteps + 10 // Safety limit

			for seq.IsActive() && stepCount < maxSteps {
				// Check if waiting
				if seq.IsWaiting() {
					seq.DecrementWait()
					stepCount++
					continue
				}

				// Execute current operation
				if !seq.IsComplete() {
					op := seq.GetCurrentCommand()
					err := vm.ExecuteOp(seq, *op)
					if err != nil {
						t.Fatalf("ExecuteOp failed: %v", err)
					}
					seq.IncrementPC()
				} else {
					// Sequence complete
					seq.Deactivate()
				}
			}

			if stepCount != tt.expectedSteps {
				t.Errorf("%s: expected %d steps, got %d", tt.description, tt.expectedSteps, stepCount)
			}

			if seq.IsActive() && stepCount >= maxSteps {
				t.Errorf("%s: sequence did not complete (infinite loop?)", tt.description)
			}

			// Check final value of x
			x := seq.GetVariable("x")
			if xInt, ok := x.(int64); !ok || xInt != tt.expectedX {
				t.Errorf("%s: expected x=%d, got %v", tt.description, tt.expectedX, x)
			}
		})
	}
}

func TestStepBlockWithEmptySteps(t *testing.T) {
	// Test step block with empty steps (from commas)
	// This matches the KUMA2 pattern: PlayWAVE(...);,,
	opcodes := []interpreter.OpCode{
		// Command 1
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}},
		// Empty step (from comma)
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}},
		// Command 2
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(2)}},
		{Cmd: interpreter.OpWait, Args: []any{int64(2)}},
	}

	logger := NewLogger(0)
	mockRenderer := &MockRenderer{}
	mockAssetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	mockImageDecoder := &MockImageDecoder{}
	state := NewEngineState(mockRenderer, mockAssetLoader, mockImageDecoder)
	engine := &Engine{
		state:  state,
		logger: logger,
	}
	vm := NewVM(state, engine, logger)

	seq := NewSequencer(opcodes, TIME, nil)

	// Execute
	stepCount := 0
	maxSteps := 30

	for seq.IsActive() && stepCount < maxSteps {
		if seq.IsWaiting() {
			seq.DecrementWait()
			stepCount++
			continue
		}

		if !seq.IsComplete() {
			op := seq.GetCurrentCommand()
			err := vm.ExecuteOp(seq, *op)
			if err != nil {
				t.Fatalf("ExecuteOp failed: %v", err)
			}
			seq.IncrementPC()
		} else {
			seq.Deactivate()
		}
	}

	// Should execute: cmd1, wait(2×3=6), empty wait(2×3=6), cmd2, wait(2×3=6) = 18 ticks total
	expectedSteps := 18
	if stepCount != expectedSteps {
		t.Errorf("Expected %d steps, got %d", expectedSteps, stepCount)
	}

	// Check that x was set to 2
	x := seq.GetVariable("x")
	if xInt, ok := x.(int64); !ok || xInt != 2 {
		t.Errorf("Expected x=2, got %v", x)
	}
}
