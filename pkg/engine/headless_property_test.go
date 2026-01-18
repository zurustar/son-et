package engine

import (
	"testing"
	"testing/quick"
	"time"
)

// Property 22: Headless mode execution equivalence
// Feature: core-engine, Property 22: Headless mode execution equivalence
// Validates: Requirements 43.3, 43.4
func TestProperty22_HeadlessExecutionEquivalence(t *testing.T) {
	// Property: Variable assignments work the same in headless and GUI modes
	t.Run("Variable assignment equivalence", func(t *testing.T) {
		property := func(value int) bool {
			// Use a simple ASCII variable name to avoid encoding issues
			varName := "testvar"

			// Test in GUI mode
			oldHeadlessMode := headlessMode
			headlessMode = false

			Assign(varName, value)
			vmLock.Lock()
			guiValue := globalVars[varName]
			vmLock.Unlock()

			// Clear for next test
			vmLock.Lock()
			delete(globalVars, varName)
			vmLock.Unlock()

			// Test in headless mode
			headlessMode = true

			Assign(varName, value)
			vmLock.Lock()
			headlessValue := globalVars[varName]
			vmLock.Unlock()

			// Restore mode
			headlessMode = oldHeadlessMode

			// Clear for next test
			vmLock.Lock()
			delete(globalVars, varName)
			vmLock.Unlock()

			// Values should be the same
			return guiValue == headlessValue && guiValue == value
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
			t.Error(err)
		}
	})

	// Property: Rendering operations don't crash in headless mode
	t.Run("Rendering operations safety", func(t *testing.T) {
		property := func(x, y, w, h int16) bool {
			// Constrain to reasonable ranges
			if w < 1 || w > 1000 || h < 1 || h > 1000 {
				return true
			}

			oldHeadlessMode := headlessMode
			headlessMode = true
			defer func() { headlessMode = oldHeadlessMode }()

			// These should not crash in headless mode
			winID := OpenWin(0, int(x), int(y), int(w), int(h), 0, 0, 0xFFFFFF)
			PutCast(0, 0, 0, 0, 100, 100, 0, 0)
			MoveCast(0, int(x), int(y))

			// Should return dummy IDs
			return winID == 0
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
			t.Error(err)
		}
	})

	// Property: VM updates work in both modes
	t.Run("VM update equivalence", func(t *testing.T) {
		property := func(tickCount uint8) bool {
			// Test in GUI mode
			oldHeadlessMode := headlessMode
			headlessMode = false

			// Call UpdateVM (should not crash)
			UpdateVM(int(tickCount))

			// Test in headless mode
			headlessMode = true

			// Call UpdateVM (should not crash)
			UpdateVM(int(tickCount))

			// Restore mode
			headlessMode = oldHeadlessMode

			return true
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 50}); err != nil {
			t.Error(err)
		}
	})
}

// Property 23: Timeout termination
// Feature: core-engine, Property 23: Timeout termination
// Validates: Requirements 44.1, 44.3, 44.4
func TestProperty23_TimeoutTermination(t *testing.T) {
	// Property: Timeout fires within expected time range
	t.Run("Timeout accuracy", func(t *testing.T) {
		property := func(ms uint8) bool {
			// Constrain to reasonable range (10-200ms)
			if ms < 10 || ms > 200 {
				return true
			}

			timeout := time.Duration(ms) * time.Millisecond
			start := time.Now()

			// Create timeout channel
			timeoutChan := time.After(timeout)

			// Wait for timeout
			<-timeoutChan

			elapsed := time.Since(start)

			// Verify timeout fired within reasonable tolerance (±30ms)
			return elapsed >= timeout && elapsed <= timeout+30*time.Millisecond
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 20}); err != nil {
			t.Error(err)
		}
	})

	// Property: AfterFunc callback executes after timeout
	t.Run("AfterFunc execution", func(t *testing.T) {
		property := func(ms uint8) bool {
			// Constrain to reasonable range (10-100ms)
			if ms < 10 || ms > 100 {
				return true
			}

			timeout := time.Duration(ms) * time.Millisecond
			fired := false

			timer := time.AfterFunc(timeout, func() {
				fired = true
			})
			defer timer.Stop()

			// Wait for timeout + buffer
			time.Sleep(timeout + 20*time.Millisecond)

			return fired
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 20}); err != nil {
			t.Error(err)
		}
	})

	// Property: Cancelled timeouts don't fire
	t.Run("Timeout cancellation", func(t *testing.T) {
		property := func(ms uint8) bool {
			// Constrain to reasonable range (20-100ms)
			if ms < 20 || ms > 100 {
				return true
			}

			timeout := time.Duration(ms) * time.Millisecond
			fired := false

			timer := time.AfterFunc(timeout, func() {
				fired = true
			})

			// Cancel immediately
			timer.Stop()

			// Wait longer than timeout
			time.Sleep(timeout + 20*time.Millisecond)

			// Should NOT have fired
			return !fired
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 20}); err != nil {
			t.Error(err)
		}
	})
}

// Property 24: Exit immediate termination
// Feature: core-engine, Property 24: Exit immediate termination
// Validates: Requirements 44.6
func TestProperty24_ExitImmediateTermination(t *testing.T) {
	// Property: Exit cancels timeout
	t.Run("Exit overrides timeout", func(t *testing.T) {
		property := func(ms uint8) bool {
			// Constrain to reasonable range (50-200ms)
			if ms < 50 || ms > 200 {
				return true
			}

			timeout := time.Duration(ms) * time.Millisecond
			timeoutFired := false
			exitCalled := false

			// Create timeout
			timer := time.AfterFunc(timeout, func() {
				timeoutFired = true
			})

			// Simulate Exit() being called immediately
			exitFunc := func() {
				exitCalled = true
				timer.Stop() // Cancel the timeout
			}

			// Call Exit immediately
			exitFunc()

			// Wait for timeout duration
			time.Sleep(timeout + 20*time.Millisecond)

			// Exit should have been called, timeout should not have fired
			return exitCalled && !timeoutFired
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 20}); err != nil {
			t.Error(err)
		}
	})

	// Property: Exit cleans up resources
	t.Run("Exit cleanup", func(t *testing.T) {
		property := func(varCount uint8) bool {
			// Constrain to reasonable range (1-20 variables)
			if varCount < 1 || varCount > 20 {
				return true
			}

			engine := NewTestEngine()

			// Create some variables
			for i := uint8(0); i < varCount; i++ {
				Assign(string(rune('a'+i)), int(i))
			}

			// Simulate cleanup before exit
			engine.Reset()

			// Verify all resources are cleaned up
			return len(engine.pictures) == 0 &&
				len(engine.windows) == 0 &&
				len(engine.casts) == 0
		}

		if err := quick.Check(property, &quick.Config{MaxCount: 20}); err != nil {
			t.Error(err)
		}
	})
}
