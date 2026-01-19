package engine

import (
	"testing"
	"testing/quick"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Property 1: RegisterSequence Non-blocking
// Feature: user-input-handling, Property 1: RegisterSequence Non-blocking
// Validates: Requirements 1.1
func TestProperty1_RegisterSequenceNonBlocking(t *testing.T) {
	// Property: RegisterSequence returns within 10ms for TIME mode
	t.Run("RegisterSequence returns within 10ms for TIME mode", func(t *testing.T) {
		property := func(opCount uint8) bool {
			// Constrain to reasonable range (1-50 operations)
			if opCount < 1 || opCount > 50 {
				return true
			}

			// Generate random OpCode sequence
			ops := generateRandomOpCodes(int(opCount))

			// Measure time from call to return
			start := time.Now()
			RegisterSequence(Time, ops)
			elapsed := time.Since(start)

			// Verify returns within 10ms
			return elapsed < 10*time.Millisecond
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: RegisterSequence returns immediately regardless of sequence complexity
	t.Run("RegisterSequence timing independent of sequence length", func(t *testing.T) {
		property := func(opCount uint8) bool {
			// Test with varying sequence lengths (1-100 operations)
			if opCount < 1 {
				opCount = 1
			}

			// Generate random OpCode sequence
			ops := generateRandomOpCodes(int(opCount))

			// Measure time from call to return
			start := time.Now()
			RegisterSequence(Time, ops)
			elapsed := time.Since(start)

			// Verify returns within 10ms regardless of length
			return elapsed < 10*time.Millisecond
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: Multiple RegisterSequence calls are all non-blocking
	t.Run("Multiple RegisterSequence calls are non-blocking", func(t *testing.T) {
		property := func(seqCount uint8) bool {
			// Constrain to reasonable range (1-10 sequences)
			if seqCount < 1 || seqCount > 10 {
				return true
			}

			// Measure total time for multiple registrations
			start := time.Now()
			for i := uint8(0); i < seqCount; i++ {
				ops := generateRandomOpCodes(10)
				RegisterSequence(Time, ops)
			}
			elapsed := time.Since(start)

			// Total time should be < 10ms * seqCount (with generous buffer)
			// Each call should be fast, so total should be reasonable
			maxExpected := time.Duration(seqCount) * 10 * time.Millisecond
			return elapsed < maxExpected
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
			t.Error(err)
		}
	})
}

// generateRandomOpCodes creates a random sequence of valid OpCodes for testing
func generateRandomOpCodes(count int) []OpCode {
	ops := make([]OpCode, count)

	// Create a variety of safe OpCodes that won't cause side effects during testing
	opTypes := []interpreter.OpCmd{
		interpreter.OpAssign,
		interpreter.OpWait,
		interpreter.OpLiteral,
		interpreter.OpVarRef,
	}

	for i := 0; i < count; i++ {
		// Pick a random OpCode type
		opType := opTypes[i%len(opTypes)]

		switch opType {
		case interpreter.OpAssign:
			// Assign operation: variable name and value
			ops[i] = OpCode{
				Cmd: opType,
				Args: []any{
					interpreter.Variable("testvar"),
					OpCode{Cmd: interpreter.OpLiteral, Args: []any{i}},
				},
			}
		case interpreter.OpWait:
			// Wait operation: number of ticks
			ops[i] = OpCode{
				Cmd:  opType,
				Args: []any{1}, // Wait 1 tick
			}
		case interpreter.OpLiteral:
			// Literal value
			ops[i] = OpCode{
				Cmd:  opType,
				Args: []any{i},
			}
		case interpreter.OpVarRef:
			// Variable reference
			ops[i] = OpCode{
				Cmd:  opType,
				Args: []any{interpreter.Variable("testvar")},
			}
		}
	}

	return ops
}
