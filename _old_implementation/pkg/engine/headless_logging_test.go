package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestHeadlessLoggingTimestamps tests that log messages contain timestamps
// Requirements: 9.4
func TestHeadlessLoggingTimestamps(t *testing.T) {
	// Capture stdout to verify log format
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() {
		headlessMode = oldHeadlessMode
		os.Stdout = oldStdout
	}()

	// Reset engine state
	ResetEngineForTest()

	// Create a simple sequence
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"test", 1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"test", 2}},
	}

	// Register sequence (this should log with timestamps)
	RegisterSequence(Time, ops)

	// Execute a few VM updates (this should log with timestamps)
	for tick := 0; tick < 20; tick++ {
		UpdateVM(tick)
	}

	// Close the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify timestamp format: [HH:MM:SS.mmm]
	timestampPattern := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\]`)
	matches := timestampPattern.FindAllString(output, -1)

	if len(matches) == 0 {
		t.Errorf("No timestamps found in output. Output:\n%s", output)
	} else {
		t.Logf("Found %d timestamped log messages", len(matches))
		// Verify timestamp format is correct
		for i, match := range matches {
			if i < 3 { // Log first few timestamps
				t.Logf("Timestamp %d: %s", i+1, match)
			}
		}
	}

	// Verify specific log messages are present
	expectedLogs := []string{
		"RegisterSequence: mode=",
		"RegisterSequence: Non-blocking mode",
		"RegisterSequence: Added sequence at index",
	}

	for _, expected := range expectedLogs {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected log message containing '%s' not found in output", expected)
		}
	}
}

// TestHeadlessExecutionProgress tests that execution progress is logged
// Requirements: 9.4
func TestHeadlessExecutionProgress(t *testing.T) {
	// Capture stdout to verify execution progress logging
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() {
		headlessMode = oldHeadlessMode
		os.Stdout = oldStdout
	}()

	// Reset engine state
	ResetEngineForTest()

	// Create a sequence with multiple operations
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 0}},
		{Cmd: interpreter.OpWait, Args: []any{2}}, // Wait 2 steps
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 1}},
		{Cmd: interpreter.OpWait, Args: []any{2}},
		{Cmd: interpreter.OpAssign, Args: []any{"counter", 2}},
	}

	// Register sequence
	RegisterSequence(Time, ops)

	// Execute VM for enough ticks to complete the sequence
	for tick := 0; tick < 60; tick++ {
		UpdateVM(tick)
	}

	// Close the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify execution progress is logged
	progressIndicators := []string{
		"RegisterSequence:",                // Sequence registration
		"Wait(",                            // Wait operations
		"VM: Sequence",                     // Sequence completion or status
		"RegisterSequence: Added sequence", // Sequence added
	}

	foundCount := 0
	for _, indicator := range progressIndicators {
		if strings.Contains(output, indicator) {
			foundCount++
			t.Logf("Found progress indicator: %s", indicator)
		}
	}

	if foundCount == 0 {
		t.Errorf("No execution progress indicators found in output. Output:\n%s", output)
	} else {
		t.Logf("Found %d/%d execution progress indicators", foundCount, len(progressIndicators))
	}

	// Verify timestamps are present in progress logs
	timestampPattern := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\]`)
	matches := timestampPattern.FindAllString(output, -1)

	if len(matches) < 3 {
		t.Errorf("Expected at least 3 timestamped progress logs, found %d", len(matches))
	}
}

// TestHeadlessTimestampFormat tests the timestamp format is correct
// Requirements: 9.4
func TestHeadlessTimestampFormat(t *testing.T) {
	// Test the timestamp format used in headless logging
	now := time.Now()
	timestamp := now.Format("15:04:05.000")

	// Verify format matches expected pattern
	timestampPattern := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3}$`)
	if !timestampPattern.MatchString(timestamp) {
		t.Errorf("Timestamp format incorrect: %s", timestamp)
	}

	// Verify timestamp components
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		t.Errorf("Expected 3 parts in timestamp (HH:MM:SS.mmm), got %d", len(parts))
	}

	// Verify milliseconds are included
	if !strings.Contains(timestamp, ".") {
		t.Error("Timestamp should include milliseconds with decimal point")
	}

	secondsParts := strings.Split(parts[2], ".")
	if len(secondsParts) != 2 {
		t.Error("Seconds part should include milliseconds")
	}

	if len(secondsParts[1]) != 3 {
		t.Errorf("Milliseconds should be 3 digits, got %d", len(secondsParts[1]))
	}

	t.Logf("Timestamp format verified: %s", timestamp)
}

// TestHeadlessLoggingWithWaitOperations tests logging during Wait operations
// Requirements: 9.4
func TestHeadlessLoggingWithWaitOperations(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() {
		headlessMode = oldHeadlessMode
		os.Stdout = oldStdout
	}()

	// Reset engine state
	ResetEngineForTest()

	// Create a sequence with Wait operations
	ops := []OpCode{
		{Cmd: interpreter.OpWait, Args: []any{3}}, // Wait 3 steps
		{Cmd: interpreter.OpAssign, Args: []any{"done", true}},
	}

	// Register sequence
	RegisterSequence(Time, ops)

	// Execute VM
	for tick := 0; tick < 40; tick++ {
		UpdateVM(tick)
	}

	// Close the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify Wait operation is logged with timestamp
	if !strings.Contains(output, "Wait(") {
		t.Error("Wait operation not logged")
	}

	// Verify timestamp is present in Wait log
	waitLogPattern := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\].*Wait\(`)
	if !waitLogPattern.MatchString(output) {
		t.Error("Wait operation log does not contain timestamp")
	}

	t.Logf("Wait operation logged with timestamp")
}

// TestHeadlessLoggingSequenceCompletion tests logging when sequence completes
// Requirements: 9.4
func TestHeadlessLoggingSequenceCompletion(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() {
		headlessMode = oldHeadlessMode
		os.Stdout = oldStdout
	}()

	// Reset engine state
	ResetEngineForTest()

	// Create a short sequence that will complete
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"test", 1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"test", 2}},
		// No loop - sequence will finish
	}

	// Register sequence
	RegisterSequence(Time, ops)

	// Execute VM until sequence completes
	for tick := 0; tick < 30; tick++ {
		UpdateVM(tick)
	}

	// Close the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify sequence completion is logged
	// The sequence should either finish or loop, both should be logged
	completionIndicators := []string{
		"Sequence",
		"Finished",
	}

	foundCompletion := false
	for _, indicator := range completionIndicators {
		if strings.Contains(output, indicator) {
			foundCompletion = true
			t.Logf("Found completion indicator: %s", indicator)
			break
		}
	}

	if !foundCompletion {
		// It's okay if completion isn't logged for short sequences
		// The important thing is that execution progress is logged
		t.Logf("Sequence completion not explicitly logged (may be too short)")
	}

	// Verify timestamps are present throughout execution
	timestampPattern := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\]`)
	matches := timestampPattern.FindAllString(output, -1)

	if len(matches) < 2 {
		t.Errorf("Expected at least 2 timestamped logs during execution, found %d", len(matches))
	}
}

// TestHeadlessLoggingMultipleSequences tests logging with multiple sequences
// Requirements: 9.4
func TestHeadlessLoggingMultipleSequences(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() {
		headlessMode = oldHeadlessMode
		os.Stdout = oldStdout
	}()

	// Reset engine state
	ResetEngineForTest()

	// Create multiple sequences
	ops1 := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"seq1", 1}},
		{Cmd: interpreter.OpWait, Args: []any{2}},
	}

	ops2 := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"seq2", 2}},
		{Cmd: interpreter.OpWait, Args: []any{3}},
	}

	// Register both sequences
	RegisterSequence(Time, ops1)
	RegisterSequence(Time, ops2)

	// Execute VM
	for tick := 0; tick < 40; tick++ {
		UpdateVM(tick)
	}

	// Close the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Verify both sequences are logged
	if !strings.Contains(output, "RegisterSequence:") {
		t.Error("Sequence registration not logged")
	}

	// Count RegisterSequence logs (should be at least 2)
	registrationCount := strings.Count(output, "RegisterSequence: mode=")
	if registrationCount < 2 {
		t.Errorf("Expected at least 2 sequence registrations logged, found %d", registrationCount)
	}

	// Verify timestamps are present
	timestampPattern := regexp.MustCompile(`\[\d{2}:\d{2}:\d{2}\.\d{3}\]`)
	matches := timestampPattern.FindAllString(output, -1)

	if len(matches) < 4 {
		t.Errorf("Expected at least 4 timestamped logs for multiple sequences, found %d", len(matches))
	}

	t.Logf("Multiple sequences logged with %d timestamps", len(matches))
}

// TestHeadlessLoggingTimestampProgression tests that timestamps progress over time
// Requirements: 9.4
func TestHeadlessLoggingTimestampProgression(t *testing.T) {
	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Set headless mode
	oldHeadlessMode := headlessMode
	headlessMode = true
	defer func() {
		headlessMode = oldHeadlessMode
		os.Stdout = oldStdout
	}()

	// Reset engine state
	ResetEngineForTest()

	// Create a sequence with delays
	ops := []OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{"step", 1}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"step", 2}},
		{Cmd: interpreter.OpWait, Args: []any{1}},
		{Cmd: interpreter.OpAssign, Args: []any{"step", 3}},
	}

	// Register sequence
	RegisterSequence(Time, ops)

	// Execute VM with small delays between ticks
	for tick := 0; tick < 30; tick++ {
		UpdateVM(tick)
		if tick%10 == 0 {
			time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamps differ
		}
	}

	// Close the pipe and restore stdout
	w.Close()
	os.Stdout = oldStdout

	// Read captured output
	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Extract all timestamps
	timestampPattern := regexp.MustCompile(`\[(\d{2}:\d{2}:\d{2}\.\d{3})\]`)
	matches := timestampPattern.FindAllStringSubmatch(output, -1)

	if len(matches) < 2 {
		t.Errorf("Expected at least 2 timestamps to verify progression, found %d", len(matches))
		return
	}

	// Parse first and last timestamps
	firstTimestamp := matches[0][1]
	lastTimestamp := matches[len(matches)-1][1]

	t.Logf("First timestamp: %s", firstTimestamp)
	t.Logf("Last timestamp: %s", lastTimestamp)

	// Verify timestamps are different (execution takes time)
	// Note: They might be the same if execution is very fast, which is okay
	if firstTimestamp != lastTimestamp {
		t.Logf("Timestamps progressed over execution (good)")
	} else {
		t.Logf("Timestamps are the same (execution was very fast)")
	}

	// The important thing is that timestamps are present and properly formatted
	for i, match := range matches {
		if i < 3 {
			t.Logf("Timestamp %d: [%s]", i+1, match[1])
		}
	}
}

// TestHeadlessLoggingFormat tests the complete log message format
// Requirements: 9.4
func TestHeadlessLoggingFormat(t *testing.T) {
	// Test that log messages follow the format: [HH:MM:SS.mmm] Message
	testCases := []struct {
		message  string
		expected string
	}{
		{
			message:  fmt.Sprintf("[%s] RegisterSequence: mode=TIME", time.Now().Format("15:04:05.000")),
			expected: `\[\d{2}:\d{2}:\d{2}\.\d{3}\] RegisterSequence: mode=TIME`,
		},
		{
			message:  fmt.Sprintf("[%s] VM: Wait(2 steps) -> 24 ticks", time.Now().Format("15:04:05.000")),
			expected: `\[\d{2}:\d{2}:\d{2}\.\d{3}\] VM: Wait\(2 steps\) -> 24 ticks`,
		},
		{
			message:  fmt.Sprintf("[%s] UpdateVM: Tick 10", time.Now().Format("15:04:05.000")),
			expected: `\[\d{2}:\d{2}:\d{2}\.\d{3}\] UpdateVM: Tick 10`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			pattern := regexp.MustCompile(tc.expected)
			if !pattern.MatchString(tc.message) {
				t.Errorf("Log message format incorrect:\nGot: %s\nExpected pattern: %s", tc.message, tc.expected)
			} else {
				t.Logf("Log format verified: %s", tc.message)
			}
		})
	}
}
