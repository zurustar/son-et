// Package audio provides audio-related components for the FILLY virtual machine.
// This includes the Timer for generating TIME events, MIDI player, and WAV player.
package audio

import (
	"sync"
	"time"

	"github.com/zurustar/son-et/pkg/vm"
)

// DefaultTimerInterval is the default interval for TIME event generation.
// Requirement 3.3: System provides default timer interval of 50 milliseconds.
const DefaultTimerInterval = 50 * time.Millisecond

// Timer generates periodic TIME events for the event system.
// It runs in a separate goroutine and pushes TIME events to the event queue
// at regular intervals.
//
// Requirement 3.1: System generates TIME events periodically.
// Requirement 3.2: When timer interval is set, system uses that interval for TIME event generation.
// Requirement 3.3: System provides default timer interval of 50 milliseconds.
// Requirement 3.4: When TIME event is generated, system adds it to event queue.
// Requirement 3.5: When multiple TIME handlers are registered, system calls all of them for each TIME event.
// Requirement 3.6: System maintains accurate timing even when handler execution takes time.
type Timer struct {
	// interval is the duration between TIME events.
	// Requirement 3.2: When timer interval is set, system uses that interval for TIME event generation.
	interval time.Duration

	// eventQueue is the event queue to push TIME events to.
	// Requirement 3.4: When TIME event is generated, system adds it to event queue.
	eventQueue *vm.EventQueue

	// ticker is the underlying time.Ticker for periodic events.
	ticker *time.Ticker

	// running indicates whether the timer is currently running.
	running bool

	// stopCh is used to signal the timer goroutine to stop.
	stopCh chan struct{}

	// doneCh is used to signal that the timer goroutine has stopped.
	doneCh chan struct{}

	// mu protects the timer state.
	mu sync.Mutex
}

// NewTimer creates a new Timer with the specified interval and event queue.
// If interval is 0 or negative, the default interval (50ms) is used.
//
// Requirement 3.3: System provides default timer interval of 50 milliseconds.
//
// Parameters:
//   - interval: The duration between TIME events (use 0 for default 50ms)
//   - eventQueue: The event queue to push TIME events to
//
// Returns:
//   - *Timer: The initialized Timer instance
func NewTimer(interval time.Duration, eventQueue *vm.EventQueue) *Timer {
	if interval <= 0 {
		interval = DefaultTimerInterval
	}

	return &Timer{
		interval:   interval,
		eventQueue: eventQueue,
		ticker:     nil,
		running:    false,
		stopCh:     nil,
	}
}

// Start starts the timer, generating TIME events at the configured interval.
// If the timer is already running, this method does nothing.
//
// Requirement 3.1: System generates TIME events periodically.
// Requirement 3.6: System maintains accurate timing even when handler execution takes time.
//
// The timer runs in a separate goroutine to ensure accurate timing regardless
// of how long event handlers take to execute.
func (t *Timer) Start() {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.running {
		return
	}

	t.running = true
	t.stopCh = make(chan struct{})
	t.doneCh = make(chan struct{})
	t.ticker = time.NewTicker(t.interval)

	// Start the timer goroutine
	// Requirement 3.6: System maintains accurate timing even when handler execution takes time.
	go t.run()
}

// run is the main timer loop that generates TIME events.
// It runs in a separate goroutine and generates events at each tick.
func (t *Timer) run() {
	defer close(t.doneCh)

	for {
		select {
		case <-t.stopCh:
			return
		case _, ok := <-t.ticker.C:
			if !ok {
				return
			}
			// Generate TIME event
			// Requirement 3.1: System generates TIME events periodically.
			// Requirement 3.4: When TIME event is generated, system adds it to event queue.
			t.generateTimeEvent()
		}
	}
}

// generateTimeEvent creates and pushes a TIME event to the event queue.
//
// Requirement 3.4: When TIME event is generated, system adds it to event queue.
func (t *Timer) generateTimeEvent() {
	event := vm.NewEvent(vm.EventTIME)
	t.eventQueue.Push(event)
}

// Stop stops the timer, ceasing TIME event generation.
// If the timer is not running, this method does nothing.
func (t *Timer) Stop() {
	t.mu.Lock()

	if !t.running {
		t.mu.Unlock()
		return
	}

	t.running = false

	// Signal the goroutine to stop first
	if t.stopCh != nil {
		close(t.stopCh)
	}

	// Get doneCh before unlocking
	doneCh := t.doneCh

	t.mu.Unlock()

	// Wait for the goroutine to finish (outside the lock to avoid deadlock)
	if doneCh != nil {
		<-doneCh
	}

	// Now safe to clean up ticker
	t.mu.Lock()
	if t.ticker != nil {
		t.ticker.Stop()
		t.ticker = nil
	}
	t.stopCh = nil
	t.doneCh = nil
	t.mu.Unlock()
}

// IsRunning returns whether the timer is currently running.
func (t *Timer) IsRunning() bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.running
}

// GetInterval returns the current timer interval.
func (t *Timer) GetInterval() time.Duration {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.interval
}

// SetInterval sets a new timer interval.
// If the timer is running, it will be restarted with the new interval.
//
// Requirement 3.2: When timer interval is set, system uses that interval for TIME event generation.
//
// Parameters:
//   - interval: The new duration between TIME events (use 0 for default 50ms)
func (t *Timer) SetInterval(interval time.Duration) {
	if interval <= 0 {
		interval = DefaultTimerInterval
	}

	t.mu.Lock()
	wasRunning := t.running
	t.mu.Unlock()

	// If running, stop and restart with new interval
	if wasRunning {
		t.Stop()
		t.mu.Lock()
		t.interval = interval
		t.mu.Unlock()
		t.Start()
	} else {
		t.mu.Lock()
		t.interval = interval
		t.mu.Unlock()
	}
}
