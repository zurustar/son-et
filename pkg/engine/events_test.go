package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestRegisterMesBlock_TIME(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register TIME mes() block
	handlerID := engine.RegisterMesBlock(EventTIME, opcodes, nil, 0)

	// Verify handler was registered
	if handlerID != 1 {
		t.Errorf("Expected handler ID 1, got %d", handlerID)
	}

	// Verify handler is in state
	handlers := engine.GetState().GetEventHandlers(EventTIME)
	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(handlers))
	}

	// Verify sequencer was created with TIME mode
	if handlers[0].Sequencer.GetMode() != TIME {
		t.Errorf("Expected TIME mode, got %d", handlers[0].Sequencer.GetMode())
	}
}

func TestRegisterMesBlock_MIDI_TIME(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register MIDI_TIME mes() block
	handlerID := engine.RegisterMesBlock(EventMIDI_TIME, opcodes, nil, 0)

	// Verify handler was registered
	if handlerID != 1 {
		t.Errorf("Expected handler ID 1, got %d", handlerID)
	}

	// Verify handler is in state
	handlers := engine.GetState().GetEventHandlers(EventMIDI_TIME)
	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler, got %d", len(handlers))
	}

	// Verify sequencer was created with MIDI_TIME mode
	if handlers[0].Sequencer.GetMode() != MIDI_TIME {
		t.Errorf("Expected MIDI_TIME mode, got %d", handlers[0].Sequencer.GetMode())
	}
}

func TestRegisterMultipleMesBlocks(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register multiple handlers for the same event
	id1 := engine.RegisterMesBlock(EventKEY, opcodes, nil, 0)
	id2 := engine.RegisterMesBlock(EventKEY, opcodes, nil, 0)
	id3 := engine.RegisterMesBlock(EventKEY, opcodes, nil, 0)

	// Verify IDs are unique
	if id1 != 1 || id2 != 2 || id3 != 3 {
		t.Errorf("Expected IDs 1, 2, 3, got %d, %d, %d", id1, id2, id3)
	}

	// Verify all handlers are registered
	handlers := engine.GetState().GetEventHandlers(EventKEY)
	if len(handlers) != 3 {
		t.Errorf("Expected 3 handlers, got %d", len(handlers))
	}
}

func TestRegisterUserEvent(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register USER event handlers with different user IDs
	engine.RegisterMesBlock(EventUSER, opcodes, nil, 100)
	engine.RegisterMesBlock(EventUSER, opcodes, nil, 200)
	engine.RegisterMesBlock(EventUSER, opcodes, nil, 100)

	// Verify handlers for user ID 100
	handlers := engine.GetState().GetUserEventHandlers(100)
	if len(handlers) != 2 {
		t.Errorf("Expected 2 handlers for user ID 100, got %d", len(handlers))
	}

	// Verify handlers for user ID 200
	handlers = engine.GetState().GetUserEventHandlers(200)
	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler for user ID 200, got %d", len(handlers))
	}
}

func TestTriggerEvent(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register event handlers
	engine.RegisterMesBlock(EventKEY, opcodes, nil, 0)
	engine.RegisterMesBlock(EventKEY, opcodes, nil, 0)

	// Trigger event with parameters
	data := NewEventData(10, 20, 30, 40)
	engine.TriggerEvent(EventKEY, data)

	// Verify sequencers were registered (2 handlers + 2 triggered = 4 total)
	// Note: TIME mode handlers also register sequences
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 2 {
		t.Errorf("Expected 2 sequencers after trigger, got %d", len(sequencers))
	}

	// Verify event parameters were set
	seq := sequencers[0]
	if seq.GetVariable("MesP1") != int64(10) {
		t.Errorf("Expected MesP1=10, got %v", seq.GetVariable("MesP1"))
	}
	if seq.GetVariable("MesP2") != int64(20) {
		t.Errorf("Expected MesP2=20, got %v", seq.GetVariable("MesP2"))
	}
	if seq.GetVariable("MesP3") != int64(30) {
		t.Errorf("Expected MesP3=30, got %v", seq.GetVariable("MesP3"))
	}
	if seq.GetVariable("MesP4") != int64(40) {
		t.Errorf("Expected MesP4=40, got %v", seq.GetVariable("MesP4"))
	}
}

func TestTriggerUserEvent(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register USER event handlers
	engine.RegisterMesBlock(EventUSER, opcodes, nil, 100)
	engine.RegisterMesBlock(EventUSER, opcodes, nil, 200)

	// Trigger user event 100
	data := NewEventData(1, 2, 3, 4)
	engine.TriggerUserEvent(100, data)

	// Verify only handler for user ID 100 was triggered
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 1 {
		t.Errorf("Expected 1 sequencer after trigger, got %d", len(sequencers))
	}

	// Verify event parameters
	seq := sequencers[0]
	if seq.GetVariable("MesP1") != int64(1) {
		t.Errorf("Expected MesP1=1, got %v", seq.GetVariable("MesP1"))
	}
}

func TestMultipleHandlersSameEvent(t *testing.T) {
	engine := NewEngine(nil, nil, nil)

	opcodes1 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(1)}},
	}
	opcodes2 := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("y"), int64(2)}},
	}

	// Register multiple handlers for CLICK event
	engine.RegisterMesBlock(EventCLICK, opcodes1, nil, 0)
	engine.RegisterMesBlock(EventCLICK, opcodes2, nil, 0)

	// Trigger event
	data := NewEventData(5, 6, 7, 8)
	engine.TriggerEvent(EventCLICK, data)

	// Verify both handlers were triggered
	sequencers := engine.GetState().GetSequencers()
	if len(sequencers) != 2 {
		t.Errorf("Expected 2 sequencers, got %d", len(sequencers))
	}

	// Verify both have event parameters
	for _, seq := range sequencers {
		if seq.GetVariable("MesP1") != int64(5) {
			t.Errorf("Expected MesP1=5, got %v", seq.GetVariable("MesP1"))
		}
	}
}

func TestEventTypeString(t *testing.T) {
	tests := []struct {
		eventType EventType
		expected  string
	}{
		{EventTIME, "TIME"},
		{EventMIDI_TIME, "MIDI_TIME"},
		{EventMIDI_END, "MIDI_END"},
		{EventKEY, "KEY"},
		{EventCLICK, "CLICK"},
		{EventRBDOWN, "RBDOWN"},
		{EventRBDBLCLK, "RBDBLCLK"},
		{EventUSER, "USER"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			result := tt.eventType.String()
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

func TestDeactivateEventHandler(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}
	seq := NewSequencer(opcodes, TIME, nil)

	// Register handler
	handlerID := state.RegisterEventHandler(EventKEY, seq, 0)

	// Verify handler is active
	handlers := state.GetEventHandlers(EventKEY)
	if len(handlers) != 1 {
		t.Errorf("Expected 1 active handler, got %d", len(handlers))
	}

	// Deactivate handler
	state.DeactivateEventHandler(handlerID)

	// Verify handler is no longer active
	handlers = state.GetEventHandlers(EventKEY)
	if len(handlers) != 0 {
		t.Errorf("Expected 0 active handlers, got %d", len(handlers))
	}
}

func TestCleanupInactiveEventHandlers(t *testing.T) {
	state := NewEngineState(nil, nil, nil)

	opcodes := []interpreter.OpCode{
		{Cmd: interpreter.OpAssign, Args: []any{interpreter.Variable("x"), int64(5)}},
	}

	// Register multiple handlers
	seq1 := NewSequencer(opcodes, TIME, nil)
	seq2 := NewSequencer(opcodes, TIME, nil)
	seq3 := NewSequencer(opcodes, TIME, nil)

	id1 := state.RegisterEventHandler(EventKEY, seq1, 0)
	state.RegisterEventHandler(EventKEY, seq2, 0)
	id3 := state.RegisterEventHandler(EventKEY, seq3, 0)

	// Deactivate some handlers
	state.DeactivateEventHandler(id1)
	state.DeactivateEventHandler(id3)

	// Cleanup
	state.CleanupInactiveEventHandlers()

	// Verify only active handler remains
	handlers := state.GetEventHandlers(EventKEY)
	if len(handlers) != 1 {
		t.Errorf("Expected 1 handler after cleanup, got %d", len(handlers))
	}
}
