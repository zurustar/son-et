package engine

import (
	"bytes"
	"io"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogger_DebugLevels(t *testing.T) {
	// Capture stdout
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(DebugLevelError)

	// Level 0: Only errors
	logger.LogError("error message")
	logger.LogInfo("info message")
	logger.LogDebug("debug message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "error message") {
		t.Error("Expected error message to be logged")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should not be logged at level 0")
	}
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged at level 0")
	}
}

func TestLogger_InfoLevel(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(DebugLevelInfo)

	logger.LogError("error message")
	logger.LogInfo("info message")
	logger.LogDebug("debug message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "error message") {
		t.Error("Expected error message to be logged")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Expected info message to be logged at level 1")
	}
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should not be logged at level 1")
	}
}

func TestLogger_DebugLevel(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(DebugLevelDebug)

	logger.LogError("error message")
	logger.LogInfo("info message")
	logger.LogDebug("debug message")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "error message") {
		t.Error("Expected error message to be logged")
	}
	if !strings.Contains(output, "info message") {
		t.Error("Expected info message to be logged at level 2")
	}
	if !strings.Contains(output, "debug message") {
		t.Error("Expected debug message to be logged at level 2")
	}
}

func TestLogger_TimestampFormat(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(DebugLevelError)
	logger.LogError("test")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	// Check timestamp format [HH:MM:SS.mmm]
	if !strings.Contains(output, "[") || !strings.Contains(output, "]") {
		t.Error("Expected timestamp in brackets")
	}

	// Extract timestamp
	start := strings.Index(output, "[")
	end := strings.Index(output, "]")
	if start == -1 || end == -1 {
		t.Fatal("Could not find timestamp brackets")
	}

	timestamp := output[start+1 : end]
	parts := strings.Split(timestamp, ":")
	if len(parts) != 3 {
		t.Errorf("Expected HH:MM:SS.mmm format, got %s", timestamp)
	}

	// Check milliseconds
	if !strings.Contains(parts[2], ".") {
		t.Error("Expected milliseconds in timestamp")
	}
}

func TestLogger_SetLevel(t *testing.T) {
	logger := NewLogger(DebugLevelError)

	if logger.GetLevel() != DebugLevelError {
		t.Errorf("Expected level 0, got %d", logger.GetLevel())
	}

	logger.SetLevel(DebugLevelDebug)
	if logger.GetLevel() != DebugLevelDebug {
		t.Errorf("Expected level 2, got %d", logger.GetLevel())
	}
}

func TestLogger_ThreadSafety(t *testing.T) {
	logger := NewLogger(DebugLevelDebug)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			logger.LogInfo("message %d", n)
			logger.SetLevel(DebugLevelInfo)
			_ = logger.GetLevel()
		}(i)
	}

	wg.Wait()
}

func TestFormatTimestamp(t *testing.T) {
	testTime := time.Date(2024, 1, 1, 14, 30, 45, 123456789, time.UTC)
	result := formatTimestamp(testTime)

	expected := "14:30:45.123"
	if result != expected {
		t.Errorf("Expected %s, got %s", expected, result)
	}
}

func TestLogger_MessageFormatting(t *testing.T) {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := NewLogger(DebugLevelError)
	logger.LogError("test %d %s", 42, "hello")

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "test 42 hello") {
		t.Errorf("Expected formatted message, got %s", output)
	}
}

func ExampleLogger() {
	logger := NewLogger(DebugLevelInfo)

	logger.LogError("This is an error")
	logger.LogInfo("This is info")
	logger.LogDebug("This won't be shown")

	logger.SetLevel(DebugLevelDebug)
	logger.LogDebug("Now this will be shown")
}
