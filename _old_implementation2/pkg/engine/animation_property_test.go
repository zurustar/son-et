package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestMesBlockLooping tests that mes() blocks loop continuously until terminated
// **Feature: sample-game-fixes, Property 5: Animation Frame Completeness**
// **Validates: Requirements 3.1, 3.2, 3.5**
func TestMesBlockLooping(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	commands := []interpreter.OpCode{
		{
			Cmd: interpreter.OpAssign,
			Args: []any{
				interpreter.Variable("loopCount"),
				interpreter.OpCode{
					Cmd: interpreter.OpBinaryOp,
					Args: []any{
						"+",
						interpreter.Variable("loopCount"),
						int64(1),
					},
				},
			},
		},
	}

	seq := NewSequencer(commands, TIME, nil)
	engine.RegisterSequence(seq, 0)

	ticksToRun := 10
	for tick := 0; tick < ticksToRun; tick++ {
		err := engine.UpdateVM()
		if err != nil {
			t.Fatalf("UpdateVM failed at tick %d: %v", tick, err)
		}
	}

	loopCount := seq.GetVariable("loopCount")
	actualCount, ok := loopCount.(int64)
	if !ok {
		if intVal, ok := loopCount.(int); ok {
			actualCount = int64(intVal)
		} else {
			actualCount = 0
		}
	}

	if actualCount < int64(ticksToRun) {
		t.Errorf("Expected mes() block to loop at least %d times, got %d", ticksToRun, actualCount)
	}

	if !seq.IsActive() {
		t.Errorf("mes() block should still be active after looping")
	}

	t.Logf("mes() block successfully looped %d times", actualCount)
}
