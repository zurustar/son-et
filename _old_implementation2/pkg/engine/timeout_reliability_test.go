package engine

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestTimeoutTriggersForInfiniteLoop verifies that timeout triggers for infinite loops
func TestTimeoutTriggersForInfiniteLoop(t *testing.T) {
	// Create engine with 1 second timeout
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	eng := NewEngine(nil, assetLoader, imageDecoder)
	eng.SetHeadless(true)
	eng.SetTimeout(1 * time.Second)
	eng.SetDebugLevel(DebugLevelError)

	// Create sequence with very long wait (simulates infinite loop)
	opcodes := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpWait,
			Args: []any{int64(999999)}, // Wait for a very long time
		},
	}

	// Start engine
	eng.Start()

	// Register sequence
	seq := NewSequencer(opcodes, TIME, nil)
	eng.RegisterSequence(seq, 0)

	// Run engine loop with timeout
	startTime := time.Now()
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	maxDuration := 2 * time.Second // Allow some margin
	for {
		<-ticker.C

		if err := eng.Update(); err != nil {
			if err == ErrTerminated {
				break
			}
			t.Fatalf("Update error: %v", err)
		}

		if time.Since(startTime) > maxDuration {
			t.Fatal("Timeout did not trigger within expected duration")
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("Engine terminated after %v", elapsed)

	// Verify timeout triggered within reasonable time (1s timeout + 0.5s margin)
	if elapsed < 900*time.Millisecond || elapsed > 1500*time.Millisecond {
		t.Errorf("Timeout triggered at %v, expected ~1s", elapsed)
	}

	// Verify termination flag is set
	if !eng.IsTerminated() {
		t.Error("Engine should be terminated")
	}
}

// TestTimeoutTriggersForHungMesBlock verifies that timeout triggers for hung mes() blocks
func TestTimeoutTriggersForHungMesBlock(t *testing.T) {
	// Create engine with 1 second timeout
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	eng := NewEngine(nil, assetLoader, imageDecoder)
	eng.SetHeadless(true)
	eng.SetTimeout(1 * time.Second)
	eng.SetDebugLevel(DebugLevelError)

	// Create mes(TIME) block with infinite wait
	mesOpcodes := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpWait,
			Args: []any{int64(999999)}, // Wait for a very long time
		},
	}

	// Start engine
	eng.Start()

	// Register mes(TIME) block
	seq := NewSequencer(mesOpcodes, TIME, nil)
	eng.RegisterSequence(seq, 0)

	// Run engine loop with timeout
	startTime := time.Now()
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	maxDuration := 2 * time.Second // Allow some margin
	for {
		<-ticker.C

		if err := eng.Update(); err != nil {
			if err == ErrTerminated {
				break
			}
			t.Fatalf("Update error: %v", err)
		}

		if time.Since(startTime) > maxDuration {
			t.Fatal("Timeout did not trigger within expected duration")
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("Engine terminated after %v", elapsed)

	// Verify timeout triggered within reasonable time (1s timeout + 0.5s margin)
	if elapsed < 900*time.Millisecond || elapsed > 1500*time.Millisecond {
		t.Errorf("Timeout triggered at %v, expected ~1s", elapsed)
	}

	// Verify termination flag is set
	if !eng.IsTerminated() {
		t.Error("Engine should be terminated")
	}
}

// TestTimeoutTriggersForMIDIWait verifies that timeout triggers during MIDI wait operations
func TestTimeoutTriggersForMIDIWait(t *testing.T) {
	// Create engine with 1 second timeout
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	eng := NewEngine(nil, assetLoader, imageDecoder)
	eng.SetHeadless(true)
	eng.SetTimeout(1 * time.Second)
	eng.SetDebugLevel(DebugLevelError)

	// Create mes(MIDI_TIME) block with long wait
	mesOpcodes := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpWait,
			Args: []any{int64(999999)}, // Wait for a very long time
		},
	}

	// Start engine
	eng.Start()

	// Register mes(MIDI_TIME) block
	seq := NewSequencer(mesOpcodes, MIDI_TIME, nil)
	eng.RegisterSequence(seq, 0)

	// Run engine loop with timeout
	startTime := time.Now()
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	maxDuration := 2 * time.Second // Allow some margin
	for {
		<-ticker.C

		if err := eng.Update(); err != nil {
			if err == ErrTerminated {
				break
			}
			t.Fatalf("Update error: %v", err)
		}

		if time.Since(startTime) > maxDuration {
			t.Fatal("Timeout did not trigger within expected duration")
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("Engine terminated after %v", elapsed)

	// Verify timeout triggered within reasonable time (1s timeout + 0.5s margin)
	if elapsed < 900*time.Millisecond || elapsed > 1500*time.Millisecond {
		t.Errorf("Timeout triggered at %v, expected ~1s", elapsed)
	}

	// Verify termination flag is set
	if !eng.IsTerminated() {
		t.Error("Engine should be terminated")
	}
}

// TestTimeoutCompletesWithinDuration verifies timeout always completes within specified duration + margin
func TestTimeoutCompletesWithinDuration(t *testing.T) {
	testCases := []struct {
		name     string
		timeout  time.Duration
		maxDelay time.Duration
	}{
		{"1s timeout", 1 * time.Second, 500 * time.Millisecond},
		{"5s timeout", 5 * time.Second, 500 * time.Millisecond},
		{"10s timeout", 10 * time.Second, 500 * time.Millisecond},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create engine with specified timeout
			assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
			imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
			eng := NewEngine(nil, assetLoader, imageDecoder)
			eng.SetHeadless(true)
			eng.SetTimeout(tc.timeout)
			eng.SetDebugLevel(DebugLevelError)

			// Create sequence with very long wait (simulates infinite loop)
			opcodes := []interpreter.OpCode{
				{
					Cmd:  interpreter.OpWait,
					Args: []any{int64(999999)}, // Wait for a very long time
				},
			}

			// Start engine
			eng.Start()

			// Register sequence
			seq := NewSequencer(opcodes, TIME, nil)
			eng.RegisterSequence(seq, 0)

			// Run engine loop with timeout
			startTime := time.Now()
			ticker := time.NewTicker(time.Second / 60)
			defer ticker.Stop()

			maxDuration := tc.timeout + tc.maxDelay
			for {
				<-ticker.C

				if err := eng.Update(); err != nil {
					if err == ErrTerminated {
						break
					}
					t.Fatalf("Update error: %v", err)
				}

				if time.Since(startTime) > maxDuration {
					t.Fatalf("Timeout did not trigger within expected duration (%v)", maxDuration)
				}
			}

			elapsed := time.Since(startTime)
			t.Logf("Engine terminated after %v (timeout: %v)", elapsed, tc.timeout)

			// Verify timeout triggered within reasonable time
			minExpected := tc.timeout - 100*time.Millisecond // Allow 100ms early
			maxExpected := tc.timeout + tc.maxDelay
			if elapsed < minExpected || elapsed > maxExpected {
				t.Errorf("Timeout triggered at %v, expected between %v and %v", elapsed, minExpected, maxExpected)
			}
		})
	}
}

// TestTimeoutDoesNotTriggerForQuickCompletion verifies timeout doesn't trigger prematurely
func TestTimeoutDoesNotTriggerForQuickCompletion(t *testing.T) {
	// Create engine with 5 second timeout
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	eng := NewEngine(nil, assetLoader, imageDecoder)
	eng.SetHeadless(true)
	eng.SetTimeout(5 * time.Second)
	eng.SetDebugLevel(DebugLevelError)

	// Create simple sequence that completes quickly
	opcodes := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpWait,
			Args: []any{int64(10)}, // Wait for 10 ticks (< 1 second)
		},
	}

	// Start engine
	eng.Start()

	// Register sequence
	seq := NewSequencer(opcodes, TIME, nil)
	seq.SetNoLoop(true) // Don't loop
	eng.RegisterSequence(seq, 0)

	// Run engine loop
	startTime := time.Now()
	ticker := time.NewTicker(time.Second / 60)
	defer ticker.Stop()

	maxDuration := 2 * time.Second // Should complete well before timeout
	for {
		<-ticker.C

		if err := eng.Update(); err != nil {
			if err == ErrTerminated {
				break
			}
			t.Fatalf("Update error: %v", err)
		}

		if time.Since(startTime) > maxDuration {
			t.Fatal("Sequence did not complete within expected duration")
		}
	}

	elapsed := time.Since(startTime)
	t.Logf("Engine terminated after %v", elapsed)

	// Verify completion was due to sequence finishing, not timeout
	if elapsed > 1*time.Second {
		t.Errorf("Sequence took too long to complete: %v", elapsed)
	}

	// Verify termination flag is set (normal completion also sets it)
	if !eng.IsTerminated() {
		t.Error("Engine should be terminated")
	}
}

// TestContextCancellationPropagation verifies context cancellation propagates to goroutines
func TestContextCancellationPropagation(t *testing.T) {
	// Create engine with 1 second timeout
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	eng := NewEngine(nil, assetLoader, imageDecoder)
	eng.SetHeadless(true)
	eng.SetTimeout(1 * time.Second)
	eng.SetDebugLevel(DebugLevelError)

	// Start engine
	eng.Start()

	// Verify context is not cancelled initially
	select {
	case <-eng.GetContext().Done():
		t.Fatal("Context should not be cancelled initially")
	default:
	}

	// Wait for timeout to trigger
	time.Sleep(1200 * time.Millisecond)

	// Verify context is cancelled after timeout
	select {
	case <-eng.GetContext().Done():
		t.Log("Context cancelled as expected")
	default:
		t.Fatal("Context should be cancelled after timeout")
	}

	// Verify termination flag is set
	if !eng.IsTerminated() {
		t.Error("Engine should be terminated")
	}
}

// TestManualTerminationCancelsContext verifies manual termination cancels context
func TestManualTerminationCancelsContext(t *testing.T) {
	// Create engine without timeout
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}
	eng := NewEngine(nil, assetLoader, imageDecoder)
	eng.SetHeadless(true)
	eng.SetDebugLevel(DebugLevelError)

	// Start engine
	eng.Start()

	// Verify context is not cancelled initially
	select {
	case <-eng.GetContext().Done():
		t.Fatal("Context should not be cancelled initially")
	default:
	}

	// Manually terminate engine
	eng.Terminate()

	// Verify context is cancelled
	select {
	case <-eng.GetContext().Done():
		t.Log("Context cancelled as expected")
	default:
		t.Fatal("Context should be cancelled after manual termination")
	}

	// Verify termination flag is set
	if !eng.IsTerminated() {
		t.Error("Engine should be terminated")
	}
}
