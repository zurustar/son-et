package engine

import (
	"testing"
	"testing/quick"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestForLoopSimple tests a simple for loop: for(i=0;i<5;i=i+1)
func TestForLoopSimple(t *testing.T) {
	_ = NewTestEngine()
	seq := &Sequencer{
		vars: make(map[string]any),
	}

	// Create OpCodes for: for(i=0; i<5; i=i+1) { count = count + 1 }
	initOps := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"i", 0}},
	}

	condOp := OpCode{
		Cmd: interpreter.OpInfix,
		Args: []any{
			"<",
			OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
			5,
		},
	}

	postOps := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{
			Cmd: interpreter.OpInfix,
			Args: []any{
				"+",
				OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
				1,
			},
		}}},
	}

	bodyOps := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"count", OpCode{
			Cmd: interpreter.OpInfix,
			Args: []any{
				"+",
				OpCode{Cmd: interpreter.OpVarRef, Args: []any{"count"}},
				1,
			},
		}}},
	}

	forOp := OpCode{
		Cmd:  interpreter.OpFor,
		Args: []any{initOps, condOp, postOps, bodyOps},
	}

	// Initialize count
	seq.vars["count"] = 0

	// Execute for loop
	ExecuteOp(forOp, seq)

	// Verify results
	if count, ok := seq.vars["count"].(int); !ok || count != 5 {
		t.Errorf("Expected count=5, got %v", seq.vars["count"])
	}

	if i, ok := seq.vars["i"].(int); !ok || i != 5 {
		t.Errorf("Expected i=5 after loop, got %v", seq.vars["i"])
	}
}

// TestForLoopLessOrEqual tests for loop with <= condition: for(i=0;i<=1;i=i+1)
func TestForLoopLessOrEqual(t *testing.T) {
	_ = NewTestEngine()
	seq := &Sequencer{
		vars: make(map[string]any),
	}

	// Create OpCodes for: for(i=0; i<=1; i=i+1) { count = count + 1 }
	initOps := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"i", 0}},
	}

	condOp := OpCode{
		Cmd: interpreter.OpInfix,
		Args: []any{
			"<=",
			OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
			1,
		},
	}

	postOps := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{
			Cmd: interpreter.OpInfix,
			Args: []any{
				"+",
				OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
				1,
			},
		}}},
	}

	bodyOps := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"count", OpCode{
			Cmd: interpreter.OpInfix,
			Args: []any{
				"+",
				OpCode{Cmd: interpreter.OpVarRef, Args: []any{"count"}},
				1,
			},
		}}},
	}

	forOp := OpCode{
		Cmd:  interpreter.OpFor,
		Args: []any{initOps, condOp, postOps, bodyOps},
	}

	// Initialize count
	seq.vars["count"] = 0

	// Execute for loop
	ExecuteOp(forOp, seq)

	// Verify results - should iterate twice (i=0 and i=1)
	if count, ok := seq.vars["count"].(int); !ok || count != 2 {
		t.Errorf("Expected count=2, got %v", seq.vars["count"])
	}

	if i, ok := seq.vars["i"].(int); !ok || i != 2 {
		t.Errorf("Expected i=2 after loop, got %v", seq.vars["i"])
	}
}

// Feature: core-engine, Property 26: For loop termination
// Validates: Requirements 21.1
func TestProperty26_ForLoopTermination(t *testing.T) {
	// Property: For loops terminate within expected iterations with correct final value
	property := func(start int8, bound int8, increment int8) bool {
		// Skip invalid cases
		if increment == 0 {
			return true // Skip: infinite loop
		}

		// Skip edge case where negating increment causes overflow
		// -128 is the minimum int8 value, and -(-128) overflows to -128
		if increment == -128 {
			return true // Skip: overflow edge case
		}

		// Determine if loop will execute based on increment direction
		var expectedIterations int
		var willExecute bool

		if increment > 0 {
			// For positive increment, loop executes if start < bound
			if start >= bound {
				willExecute = false
				expectedIterations = 0
			} else {
				willExecute = true
				// Calculate how many times we need to add increment to reach or exceed bound
				// Cast to int to avoid overflow
				diff := int(bound) - int(start)
				expectedIterations = (diff + int(increment) - 1) / int(increment)
			}
		} else {
			// For negative increment, loop executes if start > bound
			if start <= bound {
				willExecute = false
				expectedIterations = 0
			} else {
				willExecute = true
				// Calculate how many times we need to add (negative) increment to reach or go below bound
				// Cast to int to avoid overflow
				diff := int(start) - int(bound)
				expectedIterations = (diff + int(-increment) - 1) / int(-increment)
			}
		}

		// Limit iterations to prevent extremely long loops
		if expectedIterations > 1000 {
			return true // Skip: too many iterations
		}

		// Create test engine and sequencer
		_ = NewTestEngine()
		seq := &Sequencer{
			vars: make(map[string]any),
		}

		// Create OpCodes for: for(i=start; i<bound; i=i+increment) { count = count + 1 }
		initOps := []OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{"i", int(start)}},
		}

		// Choose comparison operator based on increment direction
		var compOp string
		if increment > 0 {
			compOp = "<"
		} else {
			compOp = ">"
		}

		condOp := OpCode{
			Cmd: interpreter.OpInfix,
			Args: []any{
				compOp,
				OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
				int(bound),
			},
		}

		postOps := []OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{
				Cmd: interpreter.OpInfix,
				Args: []any{
					"+",
					OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
					int(increment),
				},
			}}},
		}

		bodyOps := []OpCode{
			{Cmd: interpreter.OpAssign, Args: []any{"count", OpCode{
				Cmd: interpreter.OpInfix,
				Args: []any{
					"+",
					OpCode{Cmd: interpreter.OpVarRef, Args: []any{"count"}},
					1,
				},
			}}},
		}

		forOp := OpCode{
			Cmd:  interpreter.OpFor,
			Args: []any{initOps, condOp, postOps, bodyOps},
		}

		// Initialize count
		seq.vars["count"] = 0

		// Execute for loop
		ExecuteOp(forOp, seq)

		// Verify iteration count
		count, ok := seq.vars["count"].(int)
		if !ok {
			t.Logf("count is not an int: %T", seq.vars["count"])
			return false
		}

		if count != expectedIterations {
			t.Logf("Expected %d iterations, got %d (start=%d, bound=%d, increment=%d, willExecute=%v)",
				expectedIterations, count, start, bound, increment, willExecute)
			return false
		}

		// Verify final loop variable value
		i, ok := seq.vars["i"].(int)
		if !ok {
			t.Logf("i is not an int: %T", seq.vars["i"])
			return false
		}

		// Calculate expected final value
		var expectedFinalValue int
		if !willExecute {
			// Loop didn't execute, i should still be start value
			expectedFinalValue = int(start)
		} else {
			expectedFinalValue = int(start) + expectedIterations*int(increment)
		}

		// The final value should be the first value that fails the condition
		if willExecute {
			if increment > 0 {
				// For i < bound, final value should be >= bound
				if i < int(bound) {
					t.Logf("Expected final i >= %d, got %d (start=%d, increment=%d, iterations=%d)",
						bound, i, start, increment, expectedIterations)
					return false
				}
			} else {
				// For i > bound, final value should be <= bound
				if i > int(bound) {
					t.Logf("Expected final i <= %d, got %d (start=%d, increment=%d, iterations=%d)",
						bound, i, start, increment, expectedIterations)
					return false
				}
			}
		}

		// Verify the final value matches expected
		if i != expectedFinalValue {
			t.Logf("Expected final i=%d, got %d (start=%d, bound=%d, increment=%d, iterations=%d)",
				expectedFinalValue, i, start, bound, increment, expectedIterations)
			return false
		}

		return true
	}

	config := &quick.Config{
		MaxCount: 100, // Run 100 iterations as specified in design
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property violated: %v", err)
	}
}

// TestForLoopTerminationEdgeCases tests specific edge cases for loop termination
func TestForLoopTerminationEdgeCases(t *testing.T) {
	testCases := []struct {
		name             string
		start            int
		bound            int
		increment        int
		expectedIters    int
		expectedFinalVal int
	}{
		{
			name:             "Zero iterations (start >= bound)",
			start:            5,
			bound:            5,
			increment:        1,
			expectedIters:    0,
			expectedFinalVal: 5,
		},
		{
			name:             "Single iteration",
			start:            0,
			bound:            1,
			increment:        1,
			expectedIters:    1,
			expectedFinalVal: 1,
		},
		{
			name:             "Large increment",
			start:            0,
			bound:            100,
			increment:        10,
			expectedIters:    10,
			expectedFinalVal: 100,
		},
		{
			name:             "Negative increment",
			start:            10,
			bound:            0,
			increment:        -1,
			expectedIters:    10,
			expectedFinalVal: 0,
		},
		{
			name:             "Negative increment large step",
			start:            100,
			bound:            0,
			increment:        -10,
			expectedIters:    10,
			expectedFinalVal: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_ = NewTestEngine()
			seq := &Sequencer{
				vars: make(map[string]any),
			}

			// Create OpCodes for the loop
			initOps := []OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"i", tc.start}},
			}

			// Choose comparison operator based on increment direction
			var compOp string
			if tc.increment > 0 {
				compOp = "<"
			} else {
				compOp = ">"
			}

			condOp := OpCode{
				Cmd: interpreter.OpInfix,
				Args: []any{
					compOp,
					OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
					tc.bound,
				},
			}

			postOps := []OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"i", OpCode{
					Cmd: interpreter.OpInfix,
					Args: []any{
						"+",
						OpCode{Cmd: interpreter.OpVarRef, Args: []any{"i"}},
						tc.increment,
					},
				}}},
			}

			bodyOps := []OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"count", OpCode{
					Cmd: interpreter.OpInfix,
					Args: []any{
						"+",
						OpCode{Cmd: interpreter.OpVarRef, Args: []any{"count"}},
						1,
					},
				}}},
			}

			forOp := OpCode{
				Cmd:  interpreter.OpFor,
				Args: []any{initOps, condOp, postOps, bodyOps},
			}

			// Initialize count
			seq.vars["count"] = 0

			// Execute for loop
			ExecuteOp(forOp, seq)

			// Verify iteration count
			count, ok := seq.vars["count"].(int)
			if !ok {
				t.Fatalf("count is not an int: %T", seq.vars["count"])
			}

			if count != tc.expectedIters {
				t.Errorf("Expected %d iterations, got %d", tc.expectedIters, count)
			}

			// Verify final loop variable value
			i, ok := seq.vars["i"].(int)
			if !ok {
				t.Fatalf("i is not an int: %T", seq.vars["i"])
			}

			if i != tc.expectedFinalVal {
				t.Errorf("Expected final i=%d, got %d", tc.expectedFinalVal, i)
			}
		})
	}
}
