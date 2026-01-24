package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TestGetMesNo tests querying sequence IDs.
func TestGetMesNo(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a sequence
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seqID := engine.RegisterSequence(seq, 0)

	// Query the sequence ID
	result := engine.GetMesNo(seqID)
	if result != seqID {
		t.Errorf("Expected GetMesNo to return %d, got %d", seqID, result)
	}

	// Query non-existent sequence
	result = engine.GetMesNo(999)
	if result != 0 {
		t.Errorf("Expected GetMesNo to return 0 for non-existent sequence, got %d", result)
	}
}

// TestDelMes tests terminating a specific sequence.
func TestDelMes(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create two sequences
	seq1 := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seqID1 := engine.RegisterSequence(seq1, 0)

	seq2 := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	_ = engine.RegisterSequence(seq2, 0)

	// Verify both are active
	if !seq1.IsActive() {
		t.Error("Sequence 1 should be active")
	}
	if !seq2.IsActive() {
		t.Error("Sequence 2 should be active")
	}

	// Delete sequence 1
	engine.DelMes(seqID1)

	// Verify sequence 1 is inactive, sequence 2 is still active
	if seq1.IsActive() {
		t.Error("Sequence 1 should be inactive after DelMes")
	}
	if !seq2.IsActive() {
		t.Error("Sequence 2 should still be active")
	}
}

// TestFreezeMes tests pausing a sequence.
func TestFreezeMes(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a sequence
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seqID := engine.RegisterSequence(seq, 0)

	// Verify it's active
	if !seq.IsActive() {
		t.Error("Sequence should be active")
	}

	// Freeze the sequence
	engine.FreezeMes(seqID)

	// Verify it's inactive (paused)
	// Note: Current implementation uses deactivation for freezing
	if seq.IsActive() {
		t.Error("Sequence should be inactive after FreezeMes")
	}
}

// TestActivateMes tests resuming a paused sequence.
func TestActivateMes(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a sequence
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	seqID := engine.RegisterSequence(seq, 0)

	// Freeze it
	engine.FreezeMes(seqID)

	// Try to activate it
	// Note: Current implementation is a no-op
	engine.ActivateMes(seqID)

	// In a full implementation, the sequence would be active again
	// For now, we just verify the function doesn't crash
}

// TestPostMes tests sending custom messages.
func TestPostMes(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Register a USER event handler
	userID := 100
	handlerCalled := false
	commands := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpCall,
			Args: []any{"test_function"},
		},
	}

	// Register the handler
	engine.state.RegisterEventHandler(EventUSER, commands, TIME, nil, userID)

	// Send a message to trigger the handler
	engine.PostMes(userID, 1, 2, 3, 4)

	// Verify a new sequence was created for the handler
	sequencers := engine.state.GetSequencers()
	if len(sequencers) == 0 {
		t.Error("Expected PostMes to create a new sequence for the handler")
	}

	// Verify event parameters were set
	if len(sequencers) > 0 {
		seq := sequencers[0]
		mesP1 := seq.GetVariable("MesP1")
		if mesP1 != int64(1) {
			t.Errorf("Expected MesP1=1, got %v", mesP1)
		}
		mesP2 := seq.GetVariable("MesP2")
		if mesP2 != int64(2) {
			t.Errorf("Expected MesP2=2, got %v", mesP2)
		}
		mesP3 := seq.GetVariable("MesP3")
		if mesP3 != int64(3) {
			t.Errorf("Expected MesP3=3, got %v", mesP3)
		}
		mesP4 := seq.GetVariable("MesP4")
		if mesP4 != int64(4) {
			t.Errorf("Expected MesP4=4, got %v", mesP4)
		}
	}

	_ = handlerCalled // Suppress unused variable warning
}

// TestPostMesEventTypes tests PostMes with different event types.
func TestPostMesEventTypes(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	testCases := []struct {
		name          string
		messageType   int
		expectedEvent EventType
	}{
		{"TIME", 0, EventTIME},
		{"MIDI_TIME", 1, EventMIDI_TIME},
		{"MIDI_END", 2, EventMIDI_END},
		{"KEY", 3, EventKEY},
		{"CLICK", 4, EventCLICK},
		{"RBDOWN", 5, EventRBDOWN},
		{"RBDBLCLK", 6, EventRBDBLCLK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Register a handler for this event type
			commands := []interpreter.OpCode{
				{
					Cmd:  interpreter.OpCall,
					Args: []any{"test_function"},
				},
			}
			engine.state.RegisterEventHandler(tc.expectedEvent, commands, TIME, nil, 0)

			// Get initial sequence count
			initialCount := len(engine.state.GetSequencers())

			// Send message
			engine.PostMes(tc.messageType, 10, 20, 30, 40)

			// Verify a new sequence was created
			newCount := len(engine.state.GetSequencers())
			if newCount <= initialCount {
				t.Errorf("Expected PostMes to create a new sequence for %s event", tc.name)
			}
		})
	}
}

// TestPostMesMultipleHandlers tests that PostMes triggers all matching handlers.
func TestPostMesMultipleHandlers(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	userID := 200

	// Register multiple handlers for the same USER event
	commands1 := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpCall,
			Args: []any{"handler1"},
		},
	}
	commands2 := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpCall,
			Args: []any{"handler2"},
		},
	}

	engine.state.RegisterEventHandler(EventUSER, commands1, TIME, nil, userID)
	engine.state.RegisterEventHandler(EventUSER, commands2, TIME, nil, userID)

	// Get initial sequence count
	initialCount := len(engine.state.GetSequencers())

	// Send message
	engine.PostMes(userID, 1, 2, 3, 4)

	// Verify two new sequences were created (one for each handler)
	newCount := len(engine.state.GetSequencers())
	expectedCount := initialCount + 2
	if newCount != expectedCount {
		t.Errorf("Expected %d sequences after PostMes, got %d", expectedCount, newCount)
	}
}

// TestMessageSystemIntegration tests the complete message system workflow.
func TestMessageSystemIntegration(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a main sequence
	mainSeq := NewSequencer([]interpreter.OpCode{}, TIME, nil)
	mainSeqID := engine.RegisterSequence(mainSeq, 0)

	// Verify it exists
	if engine.GetMesNo(mainSeqID) != mainSeqID {
		t.Error("Main sequence should exist")
	}

	// Register a USER event handler
	userID := 300
	handlerCommands := []interpreter.OpCode{
		{
			Cmd:  interpreter.OpCall,
			Args: []any{"event_handler"},
		},
	}
	engine.state.RegisterEventHandler(EventUSER, handlerCommands, TIME, nil, userID)

	// Send a message to trigger the handler
	engine.PostMes(userID, 5, 10, 15, 20)

	// Verify handler was triggered (new sequence created)
	sequencers := engine.state.GetSequencers()
	if len(sequencers) < 2 {
		t.Error("Expected at least 2 sequences (main + handler)")
	}

	// Freeze the main sequence
	engine.FreezeMes(mainSeqID)
	if mainSeq.IsActive() {
		t.Error("Main sequence should be frozen")
	}

	// Delete a sequence
	if len(sequencers) >= 2 {
		handlerSeqID := sequencers[1].GetID()
		engine.DelMes(handlerSeqID)
		if sequencers[1].IsActive() {
			t.Error("Handler sequence should be deleted")
		}
	}
}
