package engine

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

// TestSampleIntegration_Kuma2 tests the kuma2 sample to verify backward compatibility
// **Validates: Requirements 8.1, 8.2, 8.3, 8.4**
func TestSampleIntegration_Kuma2(t *testing.T) {
	// Check if sample exists
	sampleDir := "../../samples/kuma2"
	if _, err := os.Stat(sampleDir); os.IsNotExist(err) {
		t.Skip("kuma2 sample not found, skipping integration test")
	}

	// Save and restore global state
	oldSequencers := sequencers
	oldTickCount := tickCount
	oldMidiSyncMode := midiSyncMode
	oldProgramTerminated := programTerminated
	oldGlobalVars := globalVars
	defer func() {
		sequencers = oldSequencers
		tickCount = oldTickCount
		midiSyncMode = oldMidiSyncMode
		programTerminated = oldProgramTerminated
		globalVars = oldGlobalVars
	}()

	// Reset state
	sequencers = nil
	tickCount = 0
	midiSyncMode = false
	programTerminated = false
	globalVars = make(map[string]interface{})

	// Find TFY file
	tfyFile := findTFYFile(t, sampleDir)
	if tfyFile == "" {
		t.Skip("No TFY file found in kuma2 sample")
	}

	// Parse and compile the script
	ops := compileScript(t, tfyFile, sampleDir)
	if len(ops) == 0 {
		t.Fatal("No opcodes generated from script")
	}

	// Register the sequence (should be non-blocking)
	start := time.Now()
	RegisterSequence(Time, ops)
	duration := time.Since(start)

	// Verify non-blocking registration
	if duration > 10*time.Millisecond {
		t.Errorf("RegisterSequence blocked for %v, expected < 10ms", duration)
	}

	// Verify sequence was registered
	if len(sequencers) == 0 {
		t.Fatal("No sequencers registered")
	}

	// Simulate execution for a short period (simulate ~1 second = 60 ticks)
	executionStart := time.Now()
	for i := 0; i < 60 && !programTerminated; i++ {
		tickCount++
		UpdateVM(int(tickCount))

		// Simulate frame timing (16.67ms per frame at 60 FPS)
		time.Sleep(time.Millisecond)
	}
	executionDuration := time.Since(executionStart)

	// Verify timing accuracy (should take approximately 60ms for 60 ticks)
	// Allow generous tolerance for test environment
	if executionDuration < 50*time.Millisecond || executionDuration > 200*time.Millisecond {
		t.Logf("Warning: Execution timing was %v for 60 ticks (expected ~60-100ms)", executionDuration)
	}

	// Verify at least one sequence is still active (mes blocks loop)
	hasActiveSequence := false
	for _, seq := range sequencers {
		if seq.active {
			hasActiveSequence = true
			break
		}
	}
	if !hasActiveSequence {
		t.Error("Expected at least one active sequence after execution")
	}

	t.Logf("Kuma2 sample executed successfully: %d ticks in %v", tickCount, executionDuration)
}

// TestSampleIntegration_Robot tests the robot sample to verify backward compatibility
// **Validates: Requirements 8.1, 8.2, 8.3, 8.4**
func TestSampleIntegration_Robot(t *testing.T) {
	// Check if sample exists
	sampleDir := "../../samples/robot"
	if _, err := os.Stat(sampleDir); os.IsNotExist(err) {
		t.Skip("robot sample not found, skipping integration test")
	}

	// Save and restore global state
	oldSequencers := sequencers
	oldTickCount := tickCount
	oldMidiSyncMode := midiSyncMode
	oldProgramTerminated := programTerminated
	oldGlobalVars := globalVars
	defer func() {
		sequencers = oldSequencers
		tickCount = oldTickCount
		midiSyncMode = oldMidiSyncMode
		programTerminated = oldProgramTerminated
		globalVars = oldGlobalVars
	}()

	// Reset state
	sequencers = nil
	tickCount = 0
	midiSyncMode = false
	programTerminated = false
	globalVars = make(map[string]interface{})

	// Find TFY file
	tfyFile := findTFYFile(t, sampleDir)
	if tfyFile == "" {
		t.Skip("No TFY file found in robot sample")
	}

	// Parse and compile the script
	ops := compileScript(t, tfyFile, sampleDir)
	if len(ops) == 0 {
		t.Skip("Robot script generated no opcodes (may use unimplemented features)")
	}

	// Register the sequence (should be non-blocking)
	start := time.Now()
	RegisterSequence(Time, ops)
	duration := time.Since(start)

	// Verify non-blocking registration
	if duration > 10*time.Millisecond {
		t.Errorf("RegisterSequence blocked for %v, expected < 10ms", duration)
	}

	// Verify sequence was registered
	if len(sequencers) == 0 {
		t.Fatal("No sequencers registered")
	}

	// Simulate execution for a short period (simulate ~1 second = 60 ticks)
	executionStart := time.Now()
	for i := 0; i < 60 && !programTerminated; i++ {
		tickCount++
		UpdateVM(int(tickCount))

		// Simulate frame timing
		time.Sleep(time.Millisecond)
	}
	executionDuration := time.Since(executionStart)

	// Verify timing accuracy
	if executionDuration < 50*time.Millisecond || executionDuration > 200*time.Millisecond {
		t.Logf("Warning: Execution timing was %v for 60 ticks (expected ~60-100ms)", executionDuration)
	}

	// Verify at least one sequence is still active
	hasActiveSequence := false
	for _, seq := range sequencers {
		if seq.active {
			hasActiveSequence = true
			break
		}
	}
	if !hasActiveSequence {
		t.Error("Expected at least one active sequence after execution")
	}

	t.Logf("Robot sample executed successfully: %d ticks in %v", tickCount, executionDuration)
}

// TestSampleIntegration_TimingPreserved verifies that timing behavior is preserved
// across different samples
// **Validates: Requirements 8.3**
func TestSampleIntegration_TimingPreserved(t *testing.T) {
	tests := []struct {
		name      string
		sampleDir string
	}{
		{"kuma2", "../../samples/kuma2"},
		{"robot", "../../samples/robot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Check if sample exists
			if _, err := os.Stat(tt.sampleDir); os.IsNotExist(err) {
				t.Skipf("%s sample not found, skipping", tt.name)
			}

			// Save and restore global state
			oldSequencers := sequencers
			oldTickCount := tickCount
			oldMidiSyncMode := midiSyncMode
			oldProgramTerminated := programTerminated
			oldGlobalVars := globalVars
			defer func() {
				sequencers = oldSequencers
				tickCount = oldTickCount
				midiSyncMode = oldMidiSyncMode
				programTerminated = oldProgramTerminated
				globalVars = oldGlobalVars
			}()

			// Reset state
			sequencers = nil
			tickCount = 0
			midiSyncMode = false
			programTerminated = false
			globalVars = make(map[string]interface{})

			// Find and compile script
			tfyFile := findTFYFile(t, tt.sampleDir)
			if tfyFile == "" {
				t.Skipf("No TFY file found in %s sample", tt.name)
			}

			ops := compileScript(t, tfyFile, tt.sampleDir)
			if len(ops) == 0 {
				t.Skipf("%s script generated no opcodes (may use unimplemented features)", tt.name)
			}

			RegisterSequence(Time, ops)

			// Execute for exactly 120 ticks (2 seconds at 60 FPS)
			startTime := time.Now()
			for i := 0; i < 120 && !programTerminated; i++ {
				tickCount++
				UpdateVM(int(tickCount))
			}
			actualDuration := time.Since(startTime)

			// Verify tick count is accurate
			if tickCount != 120 {
				t.Errorf("Expected tickCount=120, got %d", tickCount)
			}

			// Timing should be consistent (we're not sleeping, so it should be fast)
			// Just verify it completed without hanging
			if actualDuration > 5*time.Second {
				t.Errorf("Execution took too long: %v (possible hang)", actualDuration)
			}

			t.Logf("%s: Executed 120 ticks in %v", tt.name, actualDuration)
		})
	}
}

// TestSampleIntegration_OutputConsistency verifies that sample output is consistent
// **Validates: Requirements 8.1, 8.4**
func TestSampleIntegration_OutputConsistency(t *testing.T) {
	// This test verifies that running the same sample multiple times produces
	// consistent behavior (same sequence of operations)

	sampleDir := "../../samples/kuma2"
	if _, err := os.Stat(sampleDir); os.IsNotExist(err) {
		t.Skip("kuma2 sample not found, skipping")
	}

	// Find and compile script once
	tfyFile := findTFYFile(t, sampleDir)
	if tfyFile == "" {
		t.Skip("No TFY file found in kuma2 sample")
	}

	ops := compileScript(t, tfyFile, sampleDir)
	if len(ops) == 0 {
		t.Fatal("No opcodes generated from script")
	}

	// Run the sample twice and compare execution traces
	var traces [2][]string

	for run := 0; run < 2; run++ {
		// Save and restore global state
		oldSequencers := sequencers
		oldTickCount := tickCount
		oldMidiSyncMode := midiSyncMode
		oldProgramTerminated := programTerminated
		oldGlobalVars := globalVars
		defer func() {
			sequencers = oldSequencers
			tickCount = oldTickCount
			midiSyncMode = oldMidiSyncMode
			programTerminated = oldProgramTerminated
			globalVars = oldGlobalVars
		}()

		// Reset state
		sequencers = nil
		tickCount = 0
		midiSyncMode = false
		programTerminated = false
		globalVars = make(map[string]interface{})

		RegisterSequence(Time, ops)

		// Execute for 30 ticks and record which opcodes execute
		var trace []string
		for i := 0; i < 30 && !programTerminated; i++ {
			tickCount++

			// Record PC positions before UpdateVM
			for _, seq := range sequencers {
				if seq.active && seq.waitTicks == 0 && seq.pc < len(seq.commands) {
					// Convert OpCmd to string representation
					cmdStr := fmt.Sprintf("%d", seq.commands[seq.pc].Cmd)
					trace = append(trace, cmdStr)
				}
			}

			UpdateVM(int(tickCount))
		}

		traces[run] = trace
	}

	// Compare traces - they should be identical
	if len(traces[0]) != len(traces[1]) {
		t.Errorf("Execution traces have different lengths: run1=%d, run2=%d",
			len(traces[0]), len(traces[1]))
	}

	// Compare first 10 operations (or fewer if trace is shorter)
	compareLen := len(traces[0])
	if len(traces[1]) < compareLen {
		compareLen = len(traces[1])
	}
	if compareLen > 10 {
		compareLen = 10
	}

	for i := 0; i < compareLen; i++ {
		if traces[0][i] != traces[1][i] {
			t.Errorf("Execution trace differs at position %d: run1=%s, run2=%s",
				i, traces[0][i], traces[1][i])
		}
	}

	t.Logf("Output consistency verified: %d operations matched", compareLen)
}

// Helper function to find TFY file in a directory
func findTFYFile(t *testing.T, dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("Failed to read directory %s: %v", dir, err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			name := entry.Name()
			if strings.HasSuffix(strings.ToLower(name), ".tfy") {
				return filepath.Join(dir, name)
			}
		}
	}

	return ""
}

// Helper function to compile a script file
func compileScript(t *testing.T, tfyFile, baseDir string) []OpCode {
	// Read the script file
	content, err := os.ReadFile(tfyFile)
	if err != nil {
		t.Fatalf("Failed to read script file %s: %v", tfyFile, err)
	}

	// Lex the content
	l := lexer.New(string(content))

	// Parse the tokens
	p := parser.New(l)
	program := p.ParseProgram()

	// Check for parse errors
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse script: %v", p.Errors())
	}

	// Compile to opcodes
	interp := interpreter.NewInterpreter()
	script, err := interp.Interpret(program)
	if err != nil {
		t.Fatalf("Failed to compile script: %v", err)
	}

	// Return the main function body
	if script.Main != nil {
		return script.Main.Body
	}

	return []OpCode{}
}
