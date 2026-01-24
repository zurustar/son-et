package engine

import (
	"testing"
	"testing/quick"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestTerminationProperty verifies that the engine terminates correctly
// when all sequences complete or timeout is reached.
// **Validates: Requirements 1.1, 1.3, 1.4, 1.5**
func TestTerminationProperty(t *testing.T) {
	// Property: Engine with finite sequences terminates within expected time
	terminationProperty := func(numOps uint8) bool {
		// Limit ops to reasonable range (1-50)
		ops := int(numOps%50) + 1

		// Create engine
		eng := NewEngine(nil, nil, nil)
		eng.SetHeadless(true)
		eng.SetTimeout(1 * time.Second)
		eng.SetDebugLevel(DebugLevelError)

		// Create simple sequence with finite ops (no loops)
		// Use OpLiteral as a no-op (just evaluates a literal)
		opcodes := make([]interpreter.OpCode, ops)
		for i := 0; i < ops; i++ {
			opcodes[i] = interpreter.OpCode{
				Cmd:  interpreter.OpLiteral,
				Args: []any{0},
			}
		}

		// Create sequencer that doesn't loop
		seq := NewSequencer(opcodes, TIME, nil)
		seq.SetNoLoop(true)
		eng.RegisterSequence(seq, 0)

		// Start engine
		eng.Start()

		// Run until termination or timeout
		startTime := time.Now()
		maxDuration := 2 * time.Second
		terminated := false

		for time.Since(startTime) < maxDuration {
			err := eng.Update()
			if err == ErrTerminated || eng.IsTerminated() {
				terminated = true
				break
			}
		}

		return terminated
	}

	if err := quick.Check(terminationProperty, &quick.Config{MaxCount: 50}); err != nil {
		t.Errorf("Termination property failed: %v", err)
	}
}

// TestTimeoutTerminationProperty verifies timeout causes termination
// **Validates: Requirements 1.3, 1.4**
func TestTimeoutTerminationProperty(t *testing.T) {
	// Property: Engine with timeout terminates even with infinite loop
	timeoutProperty := func(timeoutMs uint8) bool {
		// Limit timeout to 100-300ms for fast testing
		timeout := time.Duration(100+int(timeoutMs)%200) * time.Millisecond

		// Create engine
		eng := NewEngine(nil, nil, nil)
		eng.SetHeadless(true)
		eng.SetTimeout(timeout)
		eng.SetDebugLevel(DebugLevelError)

		// Create infinite sequence (loops forever)
		opcodes := []interpreter.OpCode{
			{Cmd: interpreter.OpLiteral, Args: []any{0}},
		}
		seq := NewSequencer(opcodes, TIME, nil)
		// Default: loops back to beginning
		eng.RegisterSequence(seq, 0)

		// Start engine
		eng.Start()

		// Run until termination
		startTime := time.Now()
		maxDuration := timeout + 1*time.Second
		terminated := false

		for time.Since(startTime) < maxDuration {
			err := eng.Update()
			if err == ErrTerminated || eng.IsTerminated() {
				terminated = true
				break
			}
		}

		// Should terminate
		return terminated
	}

	if err := quick.Check(timeoutProperty, &quick.Config{MaxCount: 10}); err != nil {
		t.Errorf("Timeout termination property failed: %v", err)
	}
}
