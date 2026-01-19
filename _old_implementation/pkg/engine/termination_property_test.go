package engine

import (
	"testing"
	"testing/quick"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// Property 9: Termination Stops Execution
// Feature: user-input-handling, Property 9: Termination Stops Execution
// Validates: Requirements 6.1, 6.2
func TestProperty9_TerminationStopsExecution(t *testing.T) {
	// Property: When programTerminated is set, no new OpCodes execute
	t.Run("No new OpCodes execute after termination flag is set", func(t *testing.T) {
		property := func(waitTicks uint8, opsBeforeTermination uint8) bool {
			// Reset engine state
			ResetEngineForTest()

			// Constrain to reasonable ranges
			if waitTicks < 1 || waitTicks > 10 {
				return true // Skip invalid inputs
			}
			if opsBeforeTermination < 1 || opsBeforeTermination > 20 {
				return true // Skip invalid inputs
			}

			// Generate a sequence with Wait() operations and assignments
			// This allows us to track execution progress
			ops := make([]OpCode, 0)
			for i := uint8(0); i < opsBeforeTermination*2; i++ {
				// Alternate between Assign and Wait operations
				if i%2 == 0 {
					ops = append(ops, OpCode{
						Cmd: interpreter.OpAssign,
						Args: []any{
							interpreter.Variable("counter"),
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i / 2)}},
						},
					})
				} else {
					ops = append(ops, OpCode{
						Cmd:  interpreter.OpWait,
						Args: []any{int(waitTicks)},
					})
				}
			}

			// Register the sequence
			vmLock.Lock()
			mainSequencer = &Sequencer{
				commands:     ops,
				pc:           0,
				waitTicks:    0,
				active:       true,
				ticksPerStep: 12,
				vars:         make(map[string]any),
				mode:         Time,
			}
			sequencers = []*Sequencer{mainSequencer}
			vmLock.Unlock()

			// Execute a few operations normally
			programTerminated = false
			for i := uint8(0); i < opsBeforeTermination && i < 3; i++ {
				UpdateVM(int(i) + 1)
			}

			// Record PC before termination
			vmLock.Lock()
			pcBeforeTermination := mainSequencer.pc
			activeBeforeTermination := mainSequencer.active
			vmLock.Unlock()

			// Set termination flag mid-execution
			programTerminated = true

			// Try to execute more operations
			for i := uint8(0); i < 5; i++ {
				UpdateVM(int(opsBeforeTermination) + int(i) + 1)
			}

			// Verify no new OpCodes executed after termination
			vmLock.Lock()
			pcAfterTermination := mainSequencer.pc
			activeAfterTermination := mainSequencer.active
			vmLock.Unlock()

			// Property holds if:
			// 1. PC didn't advance after termination
			// 2. Sequence was marked inactive
			pcDidNotAdvance := pcAfterTermination == pcBeforeTermination
			sequenceInactive := !activeAfterTermination

			// If sequence was already inactive before termination, that's okay
			if !activeBeforeTermination {
				return true
			}

			return pcDidNotAdvance && sequenceInactive
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Termination stops execution across multiple sequences
	t.Run("Termination stops all concurrent sequences", func(t *testing.T) {
		property := func(seqCount uint8, waitTicks uint8) bool {
			// Reset engine state
			ResetEngineForTest()

			// Constrain to reasonable ranges
			if seqCount < 2 || seqCount > 5 {
				return true // Skip invalid inputs
			}
			if waitTicks < 1 || waitTicks > 5 {
				return true // Skip invalid inputs
			}

			// Create multiple sequences with Wait operations
			vmLock.Lock()
			sequencers = make([]*Sequencer, seqCount)
			for i := uint8(0); i < seqCount; i++ {
				ops := []OpCode{
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							interpreter.Variable("seq_counter"),
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i)}},
						},
					},
					{Cmd: interpreter.OpWait, Args: []any{int(waitTicks)}},
					{
						Cmd: interpreter.OpAssign,
						Args: []any{
							interpreter.Variable("seq_counter"),
							OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i) + 100}},
						},
					},
					{Cmd: interpreter.OpWait, Args: []any{int(waitTicks)}},
				}

				sequencers[i] = &Sequencer{
					commands:     ops,
					pc:           0,
					waitTicks:    0,
					active:       true,
					ticksPerStep: 12,
					vars:         make(map[string]any),
					mode:         Time,
				}
			}
			vmLock.Unlock()

			// Execute a few ticks normally
			programTerminated = false
			UpdateVM(1)
			UpdateVM(2)

			// Record PCs before termination
			vmLock.Lock()
			pcsBeforeTermination := make([]int, seqCount)
			for i := uint8(0); i < seqCount; i++ {
				pcsBeforeTermination[i] = sequencers[i].pc
			}
			vmLock.Unlock()

			// Set termination flag
			programTerminated = true

			// Try to execute more ticks
			UpdateVM(3)
			UpdateVM(4)
			UpdateVM(5)

			// Verify no sequences advanced after termination
			vmLock.Lock()
			allStopped := true
			for i := uint8(0); i < seqCount; i++ {
				pcAfter := sequencers[i].pc
				pcBefore := pcsBeforeTermination[i]

				// PC should not have advanced
				if pcAfter != pcBefore {
					allStopped = false
				}

				// Sequence should be marked inactive
				if sequencers[i].active {
					allStopped = false
				}
			}
			vmLock.Unlock()

			return allStopped
		}

		config := &quick.Config{MaxCount: 50}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	// Property: Termination prevents OpCode execution at any point
	t.Run("Termination effective at any execution point", func(t *testing.T) {
		property := func(executionPoint uint8, totalOps uint8) bool {
			// Reset engine state
			ResetEngineForTest()

			// Constrain to reasonable ranges
			if totalOps < 5 || totalOps > 20 {
				return true // Skip invalid inputs
			}
			if executionPoint >= totalOps {
				return true // Skip invalid inputs
			}

			// Generate a sequence with multiple operations
			ops := make([]OpCode, totalOps)
			for i := uint8(0); i < totalOps; i++ {
				ops[i] = OpCode{
					Cmd: interpreter.OpAssign,
					Args: []any{
						interpreter.Variable("progress"),
						OpCode{Cmd: interpreter.OpLiteral, Args: []any{int(i)}},
					},
				}
			}

			// Register the sequence
			vmLock.Lock()
			mainSequencer = &Sequencer{
				commands:     ops,
				pc:           0,
				waitTicks:    0,
				active:       true,
				ticksPerStep: 12,
				vars:         make(map[string]any),
				mode:         Time,
			}
			sequencers = []*Sequencer{mainSequencer}
			vmLock.Unlock()

			// Execute up to the termination point
			programTerminated = false
			for i := uint8(0); i < executionPoint; i++ {
				UpdateVM(int(i) + 1)
			}

			// Record state before termination
			vmLock.Lock()
			pcBeforeTermination := mainSequencer.pc
			vmLock.Unlock()

			// Set termination flag
			programTerminated = true

			// Try to execute remaining operations
			for i := executionPoint; i < totalOps+5; i++ {
				UpdateVM(int(i) + 1)
			}

			// Verify execution stopped
			vmLock.Lock()
			pcAfterTermination := mainSequencer.pc
			sequenceActive := mainSequencer.active
			vmLock.Unlock()

			// Property holds if PC didn't advance and sequence is inactive
			return pcAfterTermination == pcBeforeTermination && !sequenceActive
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}
