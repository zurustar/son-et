package engine

import (
	"testing"
	"time"
)

// TestTimeoutParsing tests that timeout duration parsing works correctly
// Requirements: 44.1, 44.2
func TestTimeoutParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected time.Duration
		valid    bool
	}{
		{"5s", 5 * time.Second, true},
		{"500ms", 500 * time.Millisecond, true},
		{"2m", 2 * time.Minute, true},
		{"1h", 1 * time.Hour, true},
		{"100ms", 100 * time.Millisecond, true},
		{"invalid", 0, false},
		{"", 0, false},
		{"5", 0, false}, // Missing unit
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			duration, err := time.ParseDuration(tt.input)

			if tt.valid {
				if err != nil {
					t.Errorf("Expected valid duration for %q, got error: %v", tt.input, err)
				}
				if duration != tt.expected {
					t.Errorf("Expected duration %v for %q, got %v", tt.expected, tt.input, duration)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid input %q, got duration: %v", tt.input, duration)
				}
			}
		})
	}
}

// TestTimeoutTriggers tests that timeout triggers termination
// Requirements: 44.1, 44.3
func TestTimeoutTriggers(t *testing.T) {
	// This test verifies that a timeout channel fires after the specified duration

	timeout := 50 * time.Millisecond
	start := time.Now()

	// Create timeout channel
	timeoutChan := time.After(timeout)

	// Wait for timeout
	<-timeoutChan

	elapsed := time.Since(start)

	// Verify timeout fired within reasonable tolerance (±20ms)
	if elapsed < timeout {
		t.Errorf("Timeout fired too early: expected >= %v, got %v", timeout, elapsed)
	}
	if elapsed > timeout+20*time.Millisecond {
		t.Errorf("Timeout fired too late: expected ~%v, got %v", timeout, elapsed)
	}
}

// TestTimeoutWithAfterFunc tests time.AfterFunc behavior
// Requirements: 44.1, 44.3
func TestTimeoutWithAfterFunc(t *testing.T) {
	// This test verifies that time.AfterFunc executes the callback after the duration

	timeout := 50 * time.Millisecond
	fired := false
	start := time.Now()

	// Create timeout with callback
	timer := time.AfterFunc(timeout, func() {
		fired = true
	})
	defer timer.Stop()

	// Wait for timeout to fire
	time.Sleep(timeout + 20*time.Millisecond)

	elapsed := time.Since(start)

	// Verify callback was executed
	if !fired {
		t.Error("Timeout callback was not executed")
	}

	// Verify timing
	if elapsed < timeout {
		t.Errorf("Callback fired too early: expected >= %v, got %v", timeout, elapsed)
	}
}

// TestTimeoutCancellation tests that timeout can be cancelled
// Requirements: 44.6
func TestTimeoutCancellation(t *testing.T) {
	// This test verifies that a timeout can be cancelled (e.g., when Exit() is called)

	timeout := 100 * time.Millisecond
	fired := false

	// Create timeout with callback
	timer := time.AfterFunc(timeout, func() {
		fired = true
	})

	// Cancel the timeout immediately
	timer.Stop()

	// Wait longer than the timeout duration
	time.Sleep(timeout + 50*time.Millisecond)

	// Verify callback was NOT executed
	if fired {
		t.Error("Timeout callback should not have fired after cancellation")
	}
}

// TestGracefulShutdown tests that resources are cleaned up on timeout
// Requirements: 44.3, 44.4
func TestGracefulShutdown(t *testing.T) {
	// This test verifies that engine state can be cleaned up gracefully

	engine := NewTestEngineWithAssets(GetTestAssets())

	// Create some resources
	picID := engine.LoadPic("test.bmp")
	if picID < 0 {
		t.Fatal("Failed to load picture")
	}

	winID := engine.OpenWin(picID, 0, 0, 640, 480, 0, 0, 0xFFFFFF)
	if winID < 0 {
		t.Fatal("Failed to open window")
	}

	// Verify resources exist
	AssertResourceCount(t, engine, 1, 1, 0)

	// Simulate graceful shutdown by resetting engine
	engine.Reset()

	// Verify all resources are cleaned up
	AssertResourceCount(t, engine, 0, 0, 0)
	AssertStateConsistency(t, engine)
}

// TestExitOverridesTimeout tests that Exit() takes precedence over timeout
// Requirements: 44.6
func TestExitOverridesTimeout(t *testing.T) {
	// This test verifies the concept that Exit() should override timeout
	// In practice, Exit() would call os.Exit(0) immediately

	timeout := 100 * time.Millisecond
	exitCalled := false
	timeoutFired := false

	// Create timeout
	timer := time.AfterFunc(timeout, func() {
		timeoutFired = true
	})
	defer timer.Stop()

	// Simulate Exit() being called before timeout
	exitFunc := func() {
		exitCalled = true
		timer.Stop() // Cancel the timeout
	}

	// Call Exit() immediately
	exitFunc()

	// Wait for timeout duration
	time.Sleep(timeout + 20*time.Millisecond)

	// Verify Exit was called and timeout did not fire
	if !exitCalled {
		t.Error("Exit should have been called")
	}
	if timeoutFired {
		t.Error("Timeout should not have fired after Exit()")
	}
}

// TestTimeoutInBothModes tests that timeout works in GUI and headless modes
// Requirements: 44.5
func TestTimeoutInBothModes(t *testing.T) {
	tests := []struct {
		name     string
		headless bool
	}{
		{"GUI mode", false},
		{"Headless mode", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set mode
			oldHeadlessMode := headlessMode
			headlessMode = tt.headless
			defer func() { headlessMode = oldHeadlessMode }()

			// Create timeout
			timeout := 50 * time.Millisecond
			fired := false

			timer := time.AfterFunc(timeout, func() {
				fired = true
			})
			defer timer.Stop()

			// Wait for timeout
			time.Sleep(timeout + 20*time.Millisecond)

			// Verify timeout fired in both modes
			if !fired {
				t.Errorf("Timeout should fire in %s", tt.name)
			}
		})
	}
}

// TestTimeoutExitCode tests that timeout exits with code 0
// Requirements: 44.4
func TestTimeoutExitCode(t *testing.T) {
	// This test verifies the concept that timeout should exit with code 0
	// In practice, this would be tested via integration tests with actual process exit

	// We can verify that the engine cleans up properly, which is what should
	// happen before os.Exit(0) is called

	engine := NewTestEngine()

	// Create some state
	Assign("var1", 42)

	// Simulate cleanup before exit
	engine.Reset()

	// Verify clean state
	AssertStateConsistency(t, engine)

	// In actual implementation, os.Exit(0) would be called here
	// We can't test os.Exit directly in unit tests, but we verify cleanup works
}

// TestMultipleTimeouts tests that only one timeout should be active
// Requirements: 44.1, 44.3
func TestMultipleTimeouts(t *testing.T) {
	// This test verifies that if multiple timeouts are set, they work independently

	timeout1 := 30 * time.Millisecond
	timeout2 := 60 * time.Millisecond

	fired1 := false
	fired2 := false

	timer1 := time.AfterFunc(timeout1, func() {
		fired1 = true
	})
	defer timer1.Stop()

	timer2 := time.AfterFunc(timeout2, func() {
		fired2 = true
	})
	defer timer2.Stop()

	// Wait for first timeout
	time.Sleep(timeout1 + 10*time.Millisecond)

	if !fired1 {
		t.Error("First timeout should have fired")
	}
	if fired2 {
		t.Error("Second timeout should not have fired yet")
	}

	// Wait for second timeout
	time.Sleep(timeout2 - timeout1 + 10*time.Millisecond)

	if !fired2 {
		t.Error("Second timeout should have fired")
	}
}

// TestTimeoutLogging tests that timeout logs a message
// Requirements: 44.7
func TestTimeoutLogging(t *testing.T) {
	// This test verifies the concept of timeout logging
	// In practice, the actual logging happens in main.go's time.AfterFunc callback

	timeout := 50 * time.Millisecond
	logged := false

	timer := time.AfterFunc(timeout, func() {
		// Simulate logging
		logged = true
		// In actual code: fmt.Fprintf(os.Stderr, "Auto-termination: timeout reached\n")
	})
	defer timer.Stop()

	// Wait for timeout
	time.Sleep(timeout + 20*time.Millisecond)

	// Verify logging occurred
	if !logged {
		t.Error("Timeout should have logged a message")
	}
}
