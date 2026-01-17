package engine

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestMultipleMesBlocks tests that multiple mes(TIME) blocks can run concurrently
func TestMultipleMesBlocks(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Initialize engine
	e := NewEngineState()
	globalEngine = e

	// Initialize global vars
	globalVars = make(map[string]any)
	globalVars["counter1"] = 0
	globalVars["counter2"] = 0

	// Create first mes(TIME) block with Wait() calls
	seq1 := []OpCode{
		{Cmd: interpreter.OpSetStep, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"counter1", OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}}}},
		{Cmd: interpreter.OpWait, Args: []any{OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}}}},
		{Cmd: interpreter.OpAssign, Args: []any{"counter1", OpCode{Cmd: interpreter.OpLiteral, Args: []any{2}}}},
	}

	// Create second mes(TIME) block without Wait() calls (executes once and completes)
	seq2 := []OpCode{
		{Cmd: interpreter.OpIf, Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"<", OpCode{Cmd: interpreter.OpVarRef, Args: []any{Variable("counter2")}}, OpCode{Cmd: interpreter.OpLiteral, Args: []any{5}}}},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"counter2", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", OpCode{Cmd: interpreter.OpVarRef, Args: []any{Variable("counter2")}}, OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}}}}}},
			},
		}},
	}

	// Register both sequences
	go RegisterSequence(Time, seq1)
	go RegisterSequence(Time, seq2)

	// Wait for sequences to be registered
	time.Sleep(100 * time.Millisecond)

	// Verify both sequences are active initially
	vmLock.Lock()
	if len(sequencers) != 2 {
		t.Fatalf("Expected 2 sequences, got %d", len(sequencers))
	}
	if !sequencers[0].active {
		t.Error("Sequence 0 should be active initially")
	}
	if !sequencers[1].active {
		t.Error("Sequence 1 should be active initially")
	}
	vmLock.Unlock()

	// Run VM for several ticks
	for tick := 1; tick <= 20; tick++ {
		UpdateVM(tick)
		time.Sleep(16 * time.Millisecond) // ~60 FPS
	}

	// Verify counter1 was updated by seq1
	if globalVars["counter1"].(int) < 1 {
		t.Errorf("counter1 should be at least 1, got %v", globalVars["counter1"])
	}

	// Verify counter2 was incremented by seq2 (executes once)
	if globalVars["counter2"].(int) != 1 {
		t.Errorf("counter2 should be 1 (executed once), got %v", globalVars["counter2"])
	}

	// Verify seq2 completed (no Wait() calls, so it executes once and finishes)
	vmLock.Lock()
	if sequencers[1].active {
		t.Error("Sequence 1 should have completed (no Wait() calls)")
	}
	vmLock.Unlock()
}

// TestVariableSharingBetweenMesBlocks tests that variables are shared between mes() blocks
func TestVariableSharingBetweenMesBlocks(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Initialize engine
	e := NewEngineState()
	globalEngine = e

	// Initialize global vars
	globalVars = make(map[string]any)
	globalVars["shared_var"] = 0

	// Create first mes(TIME) block that sets shared_var
	seq1 := []OpCode{
		{Cmd: interpreter.OpSetStep, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"shared_var", OpCode{Cmd: interpreter.OpLiteral, Args: []any{42}}}},
		{Cmd: interpreter.OpWait, Args: []any{OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}}}},
	}

	// Register first sequence
	go RegisterSequence(Time, seq1)

	// Wait for sequence to be registered and execute
	time.Sleep(100 * time.Millisecond)

	// Run VM for several ticks to let seq1 set shared_var
	for tick := 1; tick <= 5; tick++ {
		UpdateVM(tick)
		time.Sleep(16 * time.Millisecond)
	}

	// Verify shared_var was set by seq1
	if globalVars["shared_var"].(int) != 42 {
		t.Errorf("shared_var should be 42, got %v", globalVars["shared_var"])
	}

	// Now create second mes(TIME) block that reads shared_var
	seq2 := []OpCode{
		{Cmd: interpreter.OpIf, Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"==", OpCode{Cmd: interpreter.OpVarRef, Args: []any{Variable("shared_var")}}, OpCode{Cmd: interpreter.OpLiteral, Args: []any{42}}}},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"result", OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}}}},
			},
		}},
	}

	// Register second sequence
	go RegisterSequence(Time, seq2)

	// Wait for sequence to be registered
	time.Sleep(100 * time.Millisecond)

	// Run VM for several more ticks to let seq2 execute
	for tick := 6; tick <= 10; tick++ {
		UpdateVM(tick)
		time.Sleep(16 * time.Millisecond)
	}

	// Verify result was set by seq2 (proving it read shared_var)
	if result, ok := globalVars["result"]; !ok || result.(int) != 1 {
		t.Errorf("result should be 1, got %v", result)
	}
}

// TestMesBlockWithIfStatements tests mes() blocks with if statements
func TestMesBlockWithIfStatements(t *testing.T) {
	// Reset engine state
	ResetEngineForTest()

	// Initialize engine
	e := NewEngineState()
	globalEngine = e

	// Initialize global vars
	globalVars = make(map[string]any)
	globalVars["flag"] = 0
	globalVars["counter"] = 0

	// Create mes(TIME) block with if statement (executes once and completes)
	seq := []OpCode{
		{Cmd: interpreter.OpIf, Args: []any{
			OpCode{Cmd: interpreter.OpInfix, Args: []any{"==", OpCode{Cmd: interpreter.OpVarRef, Args: []any{Variable("flag")}}, OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}}}},
			[]OpCode{
				{Cmd: interpreter.OpAssign, Args: []any{"counter", OpCode{Cmd: interpreter.OpInfix, Args: []any{"+", OpCode{Cmd: interpreter.OpVarRef, Args: []any{Variable("counter")}}, OpCode{Cmd: interpreter.OpLiteral, Args: []any{1}}}}}},
			},
		}},
	}

	// Register sequence
	go RegisterSequence(Time, seq)

	// Wait for sequence to be registered
	time.Sleep(100 * time.Millisecond)

	// Run VM for 5 ticks with flag=0
	for tick := 1; tick <= 5; tick++ {
		UpdateVM(tick)
		time.Sleep(16 * time.Millisecond)
	}

	// Verify counter is still 0 (flag was 0, so if body didn't execute)
	if globalVars["counter"].(int) != 0 {
		t.Errorf("counter should be 0, got %v", globalVars["counter"])
	}

	// Verify sequence completed (no Wait() calls)
	vmLock.Lock()
	if len(sequencers) > 0 && sequencers[0].active {
		t.Error("Sequence should have completed")
	}
	vmLock.Unlock()

	// Register the sequence again with flag=1
	globalVars["flag"] = 1
	go RegisterSequence(Time, seq)

	// Wait for sequence to be registered
	time.Sleep(100 * time.Millisecond)

	// Run VM for 5 more ticks with flag=1
	for tick := 6; tick <= 10; tick++ {
		UpdateVM(tick)
		time.Sleep(16 * time.Millisecond)
	}

	// Verify counter was incremented once (flag is 1, so if body executed)
	if globalVars["counter"].(int) != 1 {
		t.Errorf("counter should be 1, got %v", globalVars["counter"])
	}
}

// TestMesBlockWithForLoop tests mes() blocks with for loops
// Note: For loops in continuous mes() blocks will execute on every frame,
// so we need to be careful about infinite loops
func TestMesBlockWithForLoop(t *testing.T) {
	t.Skip("For loops in continuous mes() blocks need special handling - skipping for now")
}
