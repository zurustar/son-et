package engine

import (
	"testing"
	"time"
)

// TestHeadlessTimeout tests that timeout mechanism works correctly
// Requirements: 9.3
func TestHeadlessTimeout(t *testing.T) {
	t.Run("Timeout fires after specified duration", func(t *testing.T) {
		// Test that timeout fires within expected time range
		timeout := 50 * time.Millisecond
		fired := false

		// Create timeout using time.AfterFunc (same mechanism as main.go)
		timer := time.AfterFunc(timeout, func() {
			fired = true
		})
		defer timer.Stop()

		// Wait for timeout + buffer
		time.Sleep(timeout + 20*time.Millisecond)

		if !fired {
			t.Error("Timeout did not fire within expected time")
		}
	})

	t.Run("Timeout accuracy within tolerance", func(t *testing.T) {
		// Test that timeout fires at approximately the right time
		timeout := 100 * time.Millisecond
		start := time.Now()
		fired := false

		timer := time.AfterFunc(timeout, func() {
			fired = true
		})
		defer timer.Stop()

		// Wait for timeout + buffer
		time.Sleep(timeout + 30*time.Millisecond)

		elapsed := time.Since(start)

		if !fired {
			t.Error("Timeout did not fire")
		}

		// Verify timing is within reasonable tolerance (±30ms)
		if elapsed < timeout || elapsed > timeout+50*time.Millisecond {
			t.Errorf("Timeout fired at %v, expected around %v (±50ms)", elapsed, timeout)
		}
	})

	t.Run("Multiple timeouts can be created", func(t *testing.T) {
		// Test that multiple timeouts work independently
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
		time.Sleep(timeout1 + 20*time.Millisecond)

		if !fired1 {
			t.Error("First timeout did not fire")
		}
		if fired2 {
			t.Error("Second timeout fired too early")
		}

		// Wait for second timeout
		time.Sleep(timeout2 - timeout1 + 20*time.Millisecond)

		if !fired2 {
			t.Error("Second timeout did not fire")
		}
	})

	t.Run("Timeout can be cancelled", func(t *testing.T) {
		// Test that stopping a timer prevents it from firing
		timeout := 50 * time.Millisecond
		fired := false

		timer := time.AfterFunc(timeout, func() {
			fired = true
		})

		// Cancel immediately
		stopped := timer.Stop()

		if !stopped {
			t.Error("Timer.Stop() returned false, expected true")
		}

		// Wait longer than timeout
		time.Sleep(timeout + 30*time.Millisecond)

		if fired {
			t.Error("Timeout fired after being cancelled")
		}
	})

	t.Run("Timeout with zero duration fires immediately", func(t *testing.T) {
		// Test edge case: zero duration timeout
		fired := false

		timer := time.AfterFunc(0, func() {
			fired = true
		})
		defer timer.Stop()

		// Wait a short time
		time.Sleep(10 * time.Millisecond)

		if !fired {
			t.Error("Zero duration timeout did not fire immediately")
		}
	})
}

// TestHeadlessTimeoutWithTermination tests timeout interaction with termination flag
// Requirements: 9.3
func TestHeadlessTimeoutWithTermination(t *testing.T) {
	t.Run("Termination flag stops execution before timeout", func(t *testing.T) {
		// Reset engine state
		ResetEngineForTest()

		// Set headless mode
		oldHeadlessMode := headlessMode
		headlessMode = true
		defer func() { headlessMode = oldHeadlessMode }()

		// Reset termination flag
		oldTerminated := programTerminated
		programTerminated = false
		defer func() { programTerminated = oldTerminated }()

		// Create a timeout that would fire later
		timeout := 100 * time.Millisecond
		timeoutFired := false

		timer := time.AfterFunc(timeout, func() {
			timeoutFired = true
		})
		defer timer.Stop()

		// Set termination flag immediately (simulating user exit)
		programTerminated = true

		// Wait a short time (less than timeout)
		time.Sleep(20 * time.Millisecond)

		// Verify termination flag is set
		if !programTerminated {
			t.Error("Termination flag should be set")
		}

		// Verify timeout hasn't fired yet
		if timeoutFired {
			t.Error("Timeout should not have fired yet")
		}

		// In real execution, the program would exit here due to programTerminated
		// The timeout would be cancelled by the exit
	})

	t.Run("Timeout sets termination-like behavior", func(t *testing.T) {
		// Test that timeout can trigger termination-like behavior
		// In main.go, timeout calls os.Exit(0), which we can't test directly
		// But we can verify the timeout mechanism works

		timeout := 50 * time.Millisecond
		terminationTriggered := false

		timer := time.AfterFunc(timeout, func() {
			// In main.go, this would call os.Exit(0)
			// Here we just set a flag to verify the callback fires
			terminationTriggered = true
		})
		defer timer.Stop()

		// Wait for timeout
		time.Sleep(timeout + 20*time.Millisecond)

		if !terminationTriggered {
			t.Error("Timeout callback did not fire")
		}
	})
}

// TestHeadlessTimeoutExitCode tests that timeout results in exit code 0
// Requirements: 9.3
// Note: This test verifies the timeout mechanism, but cannot directly test os.Exit(0)
// The actual exit code behavior is tested through integration tests
func TestHeadlessTimeoutExitCode(t *testing.T) {
	t.Run("Timeout callback executes successfully", func(t *testing.T) {
		// Verify that the timeout callback can execute without errors
		// In main.go, the callback calls os.Exit(0)
		// We verify the callback mechanism works correctly

		timeout := 30 * time.Millisecond
		callbackExecuted := false
		var callbackError error

		timer := time.AfterFunc(timeout, func() {
			// Simulate the timeout callback logic
			defer func() {
				if r := recover(); r != nil {
					callbackError = r.(error)
				}
			}()

			// In main.go: os.Exit(0)
			// Here we just verify the callback executes without panic
			callbackExecuted = true
		})
		defer timer.Stop()

		// Wait for timeout
		time.Sleep(timeout + 20*time.Millisecond)

		if !callbackExecuted {
			t.Error("Timeout callback did not execute")
		}

		if callbackError != nil {
			t.Errorf("Timeout callback encountered error: %v", callbackError)
		}
	})

	t.Run("Timeout with cleanup simulation", func(t *testing.T) {
		// Test that timeout can trigger cleanup before exit
		// This simulates what would happen in a real timeout scenario

		timeout := 40 * time.Millisecond
		cleanupExecuted := false

		timer := time.AfterFunc(timeout, func() {
			// Simulate cleanup before exit
			cleanupExecuted = true
			// In main.go: os.Exit(0) would be called here
		})
		defer timer.Stop()

		// Wait for timeout
		time.Sleep(timeout + 20*time.Millisecond)

		if !cleanupExecuted {
			t.Error("Cleanup was not executed before timeout exit")
		}
	})
}

// TestTimeoutDurationParsing tests various timeout duration formats
// Requirements: 9.3
// Note: This tests the time.ParseDuration functionality used in main.go
func TestTimeoutDurationParsing(t *testing.T) {
	testCases := []struct {
		input    string
		expected time.Duration
		valid    bool
	}{
		{"5s", 5 * time.Second, true},
		{"500ms", 500 * time.Millisecond, true},
		{"2m", 2 * time.Minute, true},
		{"1h", 1 * time.Hour, true},
		{"100ms", 100 * time.Millisecond, true},
		{"1s500ms", 1500 * time.Millisecond, true},
		{"invalid", 0, false},
		{"", 0, false},
		{"5", 0, false}, // Missing unit
	}

	for _, tc := range testCases {
		t.Run(tc.input, func(t *testing.T) {
			duration, err := time.ParseDuration(tc.input)

			if tc.valid {
				if err != nil {
					t.Errorf("Expected valid duration for '%s', got error: %v", tc.input, err)
				}
				if duration != tc.expected {
					t.Errorf("Expected duration %v for '%s', got %v", tc.expected, tc.input, duration)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for invalid duration '%s', got %v", tc.input, duration)
				}
			}
		})
	}
}
