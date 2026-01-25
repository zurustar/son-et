package audio

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/vm"
)

// TestNewTimer tests the Timer constructor.
func TestNewTimer(t *testing.T) {
	eventQueue := vm.NewEventQueue()

	t.Run("with custom interval", func(t *testing.T) {
		interval := 100 * time.Millisecond
		timer := NewTimer(interval, eventQueue)

		if timer == nil {
			t.Fatal("NewTimer returned nil")
		}
		if timer.GetInterval() != interval {
			t.Errorf("expected interval %v, got %v", interval, timer.GetInterval())
		}
		if timer.IsRunning() {
			t.Error("timer should not be running after creation")
		}
	})

	t.Run("with zero interval uses default", func(t *testing.T) {
		timer := NewTimer(0, eventQueue)

		if timer.GetInterval() != DefaultTimerInterval {
			t.Errorf("expected default interval %v, got %v", DefaultTimerInterval, timer.GetInterval())
		}
	})

	t.Run("with negative interval uses default", func(t *testing.T) {
		timer := NewTimer(-10*time.Millisecond, eventQueue)

		if timer.GetInterval() != DefaultTimerInterval {
			t.Errorf("expected default interval %v, got %v", DefaultTimerInterval, timer.GetInterval())
		}
	})
}

// TestTimerStartStop tests starting and stopping the timer.
func TestTimerStartStop(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	timer := NewTimer(10*time.Millisecond, eventQueue)

	t.Run("start sets running to true", func(t *testing.T) {
		timer.Start()
		defer timer.Stop()

		if !timer.IsRunning() {
			t.Error("timer should be running after Start()")
		}
	})

	t.Run("stop sets running to false", func(t *testing.T) {
		timer.Start()
		timer.Stop()

		if timer.IsRunning() {
			t.Error("timer should not be running after Stop()")
		}
	})

	t.Run("double start is safe", func(t *testing.T) {
		timer.Start()
		timer.Start() // Should not panic or cause issues
		timer.Stop()

		if timer.IsRunning() {
			t.Error("timer should not be running after Stop()")
		}
	})

	t.Run("double stop is safe", func(t *testing.T) {
		timer.Start()
		timer.Stop()
		timer.Stop() // Should not panic

		if timer.IsRunning() {
			t.Error("timer should not be running after Stop()")
		}
	})

	t.Run("stop without start is safe", func(t *testing.T) {
		newTimer := NewTimer(10*time.Millisecond, eventQueue)
		newTimer.Stop() // Should not panic

		if newTimer.IsRunning() {
			t.Error("timer should not be running")
		}
	})
}

// TestTimerGeneratesEvents tests that the timer generates TIME events.
// Requirement 3.1: System generates TIME events periodically.
// Requirement 3.4: When TIME event is generated, system adds it to event queue.
func TestTimerGeneratesEvents(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	interval := 20 * time.Millisecond
	timer := NewTimer(interval, eventQueue)

	// Start the timer
	timer.Start()

	// Wait for some events to be generated
	time.Sleep(75 * time.Millisecond)

	// Stop the timer
	timer.Stop()

	// Check that events were generated
	eventCount := eventQueue.Len()
	if eventCount < 2 {
		t.Errorf("expected at least 2 events, got %d", eventCount)
	}

	// Verify event types
	for eventQueue.Len() > 0 {
		event, ok := eventQueue.Pop()
		if !ok {
			break
		}
		if event.Type != vm.EventTIME {
			t.Errorf("expected event type TIME, got %v", event.Type)
		}
	}
}

// TestTimerEventTimestamp tests that TIME events have timestamps.
// Requirement 1.2: When event is added to queue, system assigns timestamp.
func TestTimerEventTimestamp(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	timer := NewTimer(10*time.Millisecond, eventQueue)

	beforeStart := time.Now()
	timer.Start()

	// Wait for an event
	time.Sleep(25 * time.Millisecond)

	timer.Stop()
	afterStop := time.Now()

	// Get an event and check its timestamp
	event, ok := eventQueue.Pop()
	if !ok {
		t.Fatal("expected at least one event")
	}

	if event.Timestamp.Before(beforeStart) {
		t.Error("event timestamp should be after timer start")
	}
	if event.Timestamp.After(afterStop) {
		t.Error("event timestamp should be before timer stop")
	}
}

// TestTimerSetInterval tests changing the timer interval.
// Requirement 3.2: When timer interval is set, system uses that interval for TIME event generation.
func TestTimerSetInterval(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	timer := NewTimer(100*time.Millisecond, eventQueue)

	t.Run("set interval when not running", func(t *testing.T) {
		newInterval := 200 * time.Millisecond
		timer.SetInterval(newInterval)

		if timer.GetInterval() != newInterval {
			t.Errorf("expected interval %v, got %v", newInterval, timer.GetInterval())
		}
	})

	t.Run("set interval when running", func(t *testing.T) {
		timer.Start()
		defer timer.Stop()

		newInterval := 50 * time.Millisecond
		timer.SetInterval(newInterval)

		if timer.GetInterval() != newInterval {
			t.Errorf("expected interval %v, got %v", newInterval, timer.GetInterval())
		}
	})

	t.Run("set zero interval uses default", func(t *testing.T) {
		timer.SetInterval(0)

		if timer.GetInterval() != DefaultTimerInterval {
			t.Errorf("expected default interval %v, got %v", DefaultTimerInterval, timer.GetInterval())
		}
	})
}

// TestDefaultTimerInterval tests the default timer interval constant.
// Requirement 3.3: System provides default timer interval of 50 milliseconds.
func TestDefaultTimerInterval(t *testing.T) {
	expected := 50 * time.Millisecond
	if DefaultTimerInterval != expected {
		t.Errorf("expected default interval %v, got %v", expected, DefaultTimerInterval)
	}
}

// TestTimerAccurateTiming tests that the timer maintains accurate timing.
// Requirement 3.6: System maintains accurate timing even when handler execution takes time.
func TestTimerAccurateTiming(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	interval := 20 * time.Millisecond
	timer := NewTimer(interval, eventQueue)

	timer.Start()

	// Wait for multiple events
	time.Sleep(110 * time.Millisecond)

	timer.Stop()

	// We should have approximately 5 events (110ms / 20ms = 5.5)
	// Allow some tolerance for timing variations
	eventCount := eventQueue.Len()
	if eventCount < 4 || eventCount > 7 {
		t.Errorf("expected approximately 5 events, got %d", eventCount)
	}
}

// TestTimerConcurrentAccess tests that the timer is safe for concurrent access.
func TestTimerConcurrentAccess(t *testing.T) {
	eventQueue := vm.NewEventQueue()
	timer := NewTimer(10*time.Millisecond, eventQueue)

	// Start multiple goroutines that access the timer
	done := make(chan bool)

	go func() {
		for i := 0; i < 10; i++ {
			timer.Start()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			timer.Stop()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			_ = timer.IsRunning()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	go func() {
		for i := 0; i < 10; i++ {
			_ = timer.GetInterval()
			time.Sleep(5 * time.Millisecond)
		}
		done <- true
	}()

	// Wait for all goroutines to complete
	for i := 0; i < 4; i++ {
		<-done
	}

	// Clean up
	timer.Stop()
}
