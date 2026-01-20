package engine

import (
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// EventType represents the type of mes() event.
type EventType int

const (
	// EventTIME is a frame-driven event (blocking execution)
	EventTIME EventType = iota
	// EventMIDI_TIME is a MIDI-driven event (non-blocking execution)
	EventMIDI_TIME
	// EventMIDI_END is triggered when MIDI playback completes
	EventMIDI_END
	// EventKEY is triggered on keyboard input
	EventKEY
	// EventCLICK is triggered on mouse click
	EventCLICK
	// EventRBDOWN is triggered on right button down
	EventRBDOWN
	// EventRBDBLCLK is triggered on right button double-click
	EventRBDBLCLK
	// EventUSER is a custom user-defined event
	EventUSER
)

// String returns the string representation of an EventType
func (e EventType) String() string {
	switch e {
	case EventTIME:
		return "TIME"
	case EventMIDI_TIME:
		return "MIDI_TIME"
	case EventMIDI_END:
		return "MIDI_END"
	case EventKEY:
		return "KEY"
	case EventCLICK:
		return "CLICK"
	case EventRBDOWN:
		return "RBDOWN"
	case EventRBDBLCLK:
		return "RBDBLCLK"
	case EventUSER:
		return "USER"
	default:
		return "Unknown"
	}
}

// EventHandler represents a registered event handler (mes() block).
// It stores the OpCode template and creates new Sequencer instances when triggered.
type EventHandler struct {
	ID        int                  // Unique handler ID
	EventType EventType            // Type of event to handle
	Commands  []interpreter.OpCode // OpCode template for creating sequencers
	Mode      TimingMode           // Timing mode (TIME or MIDI_TIME)
	Parent    *Sequencer           // Parent scope for variable lookup
	Active    bool                 // Is this handler active?
	UserID    int                  // For USER events, the custom message ID
}

// EventData holds parameters passed to event handlers.
type EventData struct {
	MesP1 int // Parameter 1
	MesP2 int // Parameter 2
	MesP3 int // Parameter 3
	MesP4 int // Parameter 4
}

// NewEventData creates a new EventData with the given parameters.
func NewEventData(p1, p2, p3, p4 int) *EventData {
	return &EventData{
		MesP1: p1,
		MesP2: p2,
		MesP3: p3,
		MesP4: p4,
	}
}
