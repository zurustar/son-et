// Package vm provides the event system for the FILLY virtual machine.
// The event system implements an event-driven execution model with support for:
// - Event types (TIME, MIDI_TIME, MIDI_END, mouse events)
// - Event queue with chronological ordering
// - Event handlers registered via mes() syntax
// - Event dispatching to registered handlers
package vm

import (
	"sort"
	"sync"
	"time"

	"github.com/zurustar/son-et/pkg/opcode"
)

// EventType represents the type of an event.
// Requirement 1.7: System supports event types: TIME, MIDI_TIME, MIDI_END, LBDOWN, RBDOWN, RBDBLCLK
type EventType string

const (
	// EventTIME is generated periodically by the timer system.
	// Requirement 3.1: System generates TIME events periodically.
	EventTIME EventType = "TIME"

	// EventMIDI_TIME is generated on each MIDI tick during MIDI playback.
	// Requirement 4.3: When MIDI is playing, system generates MIDI_TIME events synchronized to MIDI tempo.
	EventMIDI_TIME EventType = "MIDI_TIME"

	// EventMIDI_END is generated when MIDI playback completes.
	// Requirement 4.5: When MIDI playback completes, system generates MIDI_END event.
	EventMIDI_END EventType = "MIDI_END"

	// EventLBDOWN is generated when the left mouse button is pressed.
	// Requirement 7.1: When left mouse button is pressed, system generates LBDOWN event.
	EventLBDOWN EventType = "LBDOWN"

	// EventRBDOWN is generated when the right mouse button is pressed.
	// Requirement 7.2: When right mouse button is pressed, system generates RBDOWN event.
	EventRBDOWN EventType = "RBDOWN"

	// EventRBDBLCLK is generated when the right mouse button is double-clicked.
	// Requirement 7.3: When right mouse button is double-clicked, system generates RBDBLCLK event.
	EventRBDBLCLK EventType = "RBDBLCLK"
)

// Event represents an event in the event system.
// Events are stored in the event queue and dispatched to registered handlers.
//
// Requirement 1.1: System provides event queue that stores events in chronological order.
// Requirement 1.2: When event is added to queue, system assigns timestamp.
type Event struct {
	// Type is the event type (TIME, MIDI_TIME, etc.)
	Type EventType

	// Timestamp is when the event was created/queued.
	// Requirement 1.2: When event is added to queue, system assigns timestamp.
	Timestamp time.Time

	// Params contains event-specific parameters.
	// For mouse events: MesP1 (window ID), MesP2 (X coordinate), MesP3 (Y coordinate)
	// For MIDI_TIME: Tick, Tempo
	// Requirement 1.6: When handler is executing, system provides access to event-specific parameters.
	Params map[string]any
}

// NewEvent creates a new event with the given type.
// The timestamp is automatically set to the current time.
//
// Requirement 1.2: When event is added to queue, system assigns timestamp.
func NewEvent(eventType EventType) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Params:    make(map[string]any),
	}
}

// NewEventWithParams creates a new event with the given type and parameters.
// The timestamp is automatically set to the current time.
func NewEventWithParams(eventType EventType, params map[string]any) *Event {
	return &Event{
		Type:      eventType,
		Timestamp: time.Now(),
		Params:    params,
	}
}

// GetParam retrieves a parameter value by name.
// Returns the value and true if found, or nil and false if not found.
func (e *Event) GetParam(name string) (any, bool) {
	if e.Params == nil {
		return nil, false
	}
	val, ok := e.Params[name]
	return val, ok
}

// SetParam sets a parameter value.
func (e *Event) SetParam(name string, value any) {
	if e.Params == nil {
		e.Params = make(map[string]any)
	}
	e.Params[name] = value
}

// DefaultQueueSize is the default maximum size of the event queue.
// Requirement 14.7: System limits queue size to prevent event queue overflow.
const DefaultQueueSize = 1000

// EventQueue is a thread-safe queue for storing events in chronological order.
//
// Requirement 1.1: System provides event queue that stores events in chronological order.
// Requirement 14.7: System limits queue size to prevent event queue overflow.
// Requirement 14.8: When queue is full, system discards oldest events.
type EventQueue struct {
	events  []*Event
	maxSize int
	mu      sync.Mutex
}

// NewEventQueue creates a new event queue with the default maximum size.
func NewEventQueue() *EventQueue {
	return &EventQueue{
		events:  make([]*Event, 0, DefaultQueueSize),
		maxSize: DefaultQueueSize,
	}
}

// NewEventQueueWithSize creates a new event queue with a custom maximum size.
func NewEventQueueWithSize(maxSize int) *EventQueue {
	if maxSize <= 0 {
		maxSize = DefaultQueueSize
	}
	return &EventQueue{
		events:  make([]*Event, 0, maxSize),
		maxSize: maxSize,
	}
}

// Push adds an event to the queue.
// If the event has no timestamp, one is assigned.
// The queue is kept sorted by timestamp (ascending).
//
// Requirement 1.1: System provides event queue that stores events in chronological order.
// Requirement 1.2: When event is added to queue, system assigns timestamp.
// Requirement 14.7: System limits queue size to prevent event queue overflow.
// Requirement 14.8: When queue is full, system discards oldest events.
func (eq *EventQueue) Push(event *Event) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	// Assign timestamp if not set
	// Requirement 1.2: When event is added to queue, system assigns timestamp.
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Check if queue is full
	// Requirement 14.8: When queue is full, system discards oldest events.
	if len(eq.events) >= eq.maxSize {
		// Remove the oldest event (first element after sorting)
		eq.events = eq.events[1:]
	}

	// Add the event
	eq.events = append(eq.events, event)

	// Sort by timestamp (ascending)
	// Requirement 1.1: System provides event queue that stores events in chronological order.
	sort.Slice(eq.events, func(i, j int) bool {
		return eq.events[i].Timestamp.Before(eq.events[j].Timestamp)
	})
}

// Pop removes and returns the oldest event from the queue.
// Returns nil and false if the queue is empty.
//
// Requirement 1.3: When event loop processes events, system dispatches events in chronological order.
func (eq *EventQueue) Pop() (*Event, bool) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if len(eq.events) == 0 {
		return nil, false
	}

	// Get the oldest event (first element)
	event := eq.events[0]
	eq.events = eq.events[1:]

	return event, true
}

// Peek returns the oldest event without removing it.
// Returns nil and false if the queue is empty.
func (eq *EventQueue) Peek() (*Event, bool) {
	eq.mu.Lock()
	defer eq.mu.Unlock()

	if len(eq.events) == 0 {
		return nil, false
	}

	return eq.events[0], true
}

// Len returns the number of events in the queue.
func (eq *EventQueue) Len() int {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	return len(eq.events)
}

// Clear removes all events from the queue.
func (eq *EventQueue) Clear() {
	eq.mu.Lock()
	defer eq.mu.Unlock()
	eq.events = eq.events[:0]
}

// EventHandler represents a handler for a specific event type.
// Handlers are registered via mes() syntax and executed when matching events occur.
//
// Requirement 1.4: When event is dispatched, system calls all handlers registered for that event type.
// Requirement 1.5: When multiple handlers are registered for same event type, system executes them in registration order.
type EventHandler struct {
	// ID is a unique identifier for this handler.
	ID string

	// EventType is the type of event this handler responds to.
	EventType EventType

	// OpCodes is the compiled code to execute when the event occurs.
	OpCodes []opcode.OpCode

	// Active indicates whether this handler is currently active.
	// Handlers can be deactivated by del_me or del_us.
	Active bool

	// StepCounter holds the step value from step(n).
	// This represents the number of TIME events to wait per comma.
	// Since TIME events are generated every 50ms, step(n) means each comma waits n × 50ms.
	// Requirement 6.1: When OpSetStep is executed, system initializes step counter.
	StepCounter int

	// WaitCounter tracks how many events to wait before continuing.
	// Requirement 6.2: When OpWait is executed, system pauses execution until next event.
	WaitCounter int

	// CurrentPC is the current program counter within the handler's OpCodes.
	CurrentPC int

	// VM is a reference to the VM for executing OpCodes.
	VM *VM

	// CurrentEvent holds the event being processed (for MesP1, MesP2, MesP3 access).
	// Requirement 1.6: When handler is executing, system provides access to event-specific parameters.
	CurrentEvent *Event

	// MarkedForDeletion indicates the handler should be removed after execution.
	// This is set by del_me/del_us.
	MarkedForDeletion bool

	// ParentScope is the scope in which the handler was registered.
	// This allows the handler to access variables from the enclosing scope (like C blocks).
	ParentScope *Scope
}

// NewEventHandler creates a new event handler.
// parentScope is the scope in which the handler was registered, allowing access to enclosing variables.
func NewEventHandler(id string, eventType EventType, opcodes []opcode.OpCode, vm *VM, parentScope *Scope) *EventHandler {
	return &EventHandler{
		ID:          id,
		EventType:   eventType,
		OpCodes:     opcodes,
		Active:      true,
		StepCounter: 0,
		WaitCounter: 0,
		CurrentPC:   0,
		VM:          vm,
		ParentScope: parentScope,
	}
}

// Execute executes the handler's OpCodes for the given event.
// Returns an error if execution fails.
// The handler supports pausing (via OpWait) and resuming from where it left off.
//
// Requirement 1.4: When event is dispatched, system calls all handlers registered for that event type.
// Requirement 1.6: When handler is executing, system provides access to event-specific parameters.
// Requirement 6.2: When OpWait is executed, system pauses execution until next event.
// Requirement 6.3: When event occurs during step execution, system proceeds to next step.
func (eh *EventHandler) Execute(event *Event) error {
	if !eh.Active {
		return nil
	}

	// If the handler is waiting, decrement the wait counter
	// Requirement 6.3: When event occurs during step execution, system proceeds to next step.
	if eh.WaitCounter > 0 {
		eh.WaitCounter--
		// ログは削除（頻繁すぎるため）

		// If still waiting, don't execute any OpCodes
		if eh.WaitCounter > 0 {
			return nil
		}
		// WaitCounter reached 0, resume execution from CurrentPC
		eh.VM.log.Debug("Handler resuming execution", "handler", eh.ID, "pc", eh.CurrentPC)
	}

	// Store the current event for parameter access
	eh.CurrentEvent = event

	// Set this handler as the current handler in the VM (for del_me support)
	previousHandler := eh.VM.currentHandler
	eh.VM.currentHandler = eh

	// Save the current local scope and set the parent scope for this handler
	// This allows the handler to access variables from the enclosing scope (like C blocks)
	previousLocalScope := eh.VM.localScope
	if eh.ParentScope != nil {
		eh.VM.localScope = eh.ParentScope
	}

	// Set event parameters in the VM's scope for access via MesP1, MesP2, MesP3
	// Requirement 1.6: When handler is executing, system provides access to event-specific parameters.
	if event.Params != nil {
		scope := eh.VM.GetCurrentScope()
		if p1, ok := event.Params["MesP1"]; ok {
			scope.Set("MesP1", p1)
		}
		if p2, ok := event.Params["MesP2"]; ok {
			scope.Set("MesP2", p2)
		}
		if p3, ok := event.Params["MesP3"]; ok {
			scope.Set("MesP3", p3)
		}
	}

	// Execute the handler's OpCodes starting from CurrentPC
	for eh.CurrentPC < len(eh.OpCodes) {
		if !eh.Active {
			break
		}

		opcode := eh.OpCodes[eh.CurrentPC]
		result, err := eh.VM.Execute(opcode)
		if err != nil {
			// Log error but continue execution
			eh.VM.log.Error("Handler execution error", "handler", eh.ID, "error", err)
		}

		eh.CurrentPC++

		// Check if we need to wait (pause execution)
		// Requirement 6.2: When OpWait is executed, system pauses execution until next event.
		if _, isWait := result.(*waitMarker); isWait {
			eh.VM.log.Debug("Handler pausing execution", "handler", eh.ID, "pc", eh.CurrentPC, "waitCounter", eh.WaitCounter)
			// Restore the previous handler and local scope, then return
			eh.VM.currentHandler = previousHandler
			eh.VM.localScope = previousLocalScope
			return nil
		}
	}

	// Handler completed all OpCodes, reset for next event
	// Requirement 6.8: When all steps are completed, system automatically terminates step block.
	eh.CurrentPC = 0
	eh.VM.log.Debug("Handler execution completed, resetting PC", "handler", eh.ID)

	// Restore the previous handler and local scope
	eh.VM.currentHandler = previousHandler
	eh.VM.localScope = previousLocalScope

	return nil
}

// Remove marks the handler for removal.
// Requirement 2.9: When del_me is called, system removes currently executing handler.
// Requirement 2.11: When del_us is called, system removes currently executing handler (same as del_me).
func (eh *EventHandler) Remove() {
	eh.Active = false
	eh.MarkedForDeletion = true
}

// HandlerRegistry manages registered event handlers.
//
// Requirement 1.5: When multiple handlers are registered for same event type, system executes them in registration order.
// Requirement 2.1: When OpRegisterEventHandler is executed, system registers handler for specified event type.
// Requirement 2.10: When del_all is called, system removes all registered handlers.
type HandlerRegistry struct {
	// handlers maps event types to their registered handlers.
	// Handlers are stored in registration order.
	handlers map[EventType][]*EventHandler

	// handlersByID maps handler IDs to handlers for quick lookup.
	handlersByID map[string]*EventHandler

	// nextID is used to generate unique handler IDs.
	nextID int

	mu sync.RWMutex
}

// NewHandlerRegistry creates a new handler registry.
func NewHandlerRegistry() *HandlerRegistry {
	return &HandlerRegistry{
		handlers:     make(map[EventType][]*EventHandler),
		handlersByID: make(map[string]*EventHandler),
		nextID:       1,
	}
}

// Register registers a new event handler and returns its ID.
// Handlers are stored in registration order.
//
// Requirement 2.1: When OpRegisterEventHandler is executed, system registers handler for specified event type.
// Requirement 1.5: When multiple handlers are registered for same event type, system executes them in registration order.
func (hr *HandlerRegistry) Register(handler *EventHandler) string {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Generate unique ID if not set
	if handler.ID == "" {
		handler.ID = generateHandlerID(hr.nextID)
		hr.nextID++
	}

	// Add to handlers list (maintains registration order)
	hr.handlers[handler.EventType] = append(hr.handlers[handler.EventType], handler)

	// Add to ID map for quick lookup
	hr.handlersByID[handler.ID] = handler

	return handler.ID
}

// generateHandlerID generates a unique handler ID.
func generateHandlerID(n int) string {
	return "handler_" + string(rune('0'+n%10)) + string(rune('0'+(n/10)%10)) + string(rune('0'+(n/100)%10))
}

// Unregister removes a handler by ID.
// Requirement 2.9: When del_me is called, system removes currently executing handler.
func (hr *HandlerRegistry) Unregister(id string) bool {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	handler, ok := hr.handlersByID[id]
	if !ok {
		return false
	}

	// Remove from ID map
	delete(hr.handlersByID, id)

	// Remove from handlers list
	eventType := handler.EventType
	handlers := hr.handlers[eventType]
	for i, h := range handlers {
		if h.ID == id {
			hr.handlers[eventType] = append(handlers[:i], handlers[i+1:]...)
			break
		}
	}

	return true
}

// UnregisterAll removes all registered handlers.
// Requirement 2.10: When del_all is called, system removes all registered handlers.
func (hr *HandlerRegistry) UnregisterAll() {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	hr.handlers = make(map[EventType][]*EventHandler)
	hr.handlersByID = make(map[string]*EventHandler)
}

// GetHandlers returns all handlers for a given event type in registration order.
// Requirement 1.5: When multiple handlers are registered for same event type, system executes them in registration order.
func (hr *HandlerRegistry) GetHandlers(eventType EventType) []*EventHandler {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	handlers := hr.handlers[eventType]
	// Return a copy to avoid race conditions
	result := make([]*EventHandler, len(handlers))
	copy(result, handlers)
	return result
}

// GetHandler returns a handler by ID.
func (hr *HandlerRegistry) GetHandler(id string) (*EventHandler, bool) {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	handler, ok := hr.handlersByID[id]
	return handler, ok
}

// GetAllHandlers returns all registered handlers.
func (hr *HandlerRegistry) GetAllHandlers() []*EventHandler {
	hr.mu.RLock()
	defer hr.mu.RUnlock()

	var result []*EventHandler
	for _, handlers := range hr.handlers {
		result = append(result, handlers...)
	}
	return result
}

// CleanupMarkedHandlers removes all handlers marked for deletion.
func (hr *HandlerRegistry) CleanupMarkedHandlers() {
	hr.mu.Lock()
	defer hr.mu.Unlock()

	// Find and remove marked handlers
	for eventType, handlers := range hr.handlers {
		var remaining []*EventHandler
		for _, h := range handlers {
			if h.MarkedForDeletion {
				delete(hr.handlersByID, h.ID)
			} else {
				remaining = append(remaining, h)
			}
		}
		hr.handlers[eventType] = remaining
	}
}

// Count returns the total number of registered handlers.
func (hr *HandlerRegistry) Count() int {
	hr.mu.RLock()
	defer hr.mu.RUnlock()
	return len(hr.handlersByID)
}

// EventDispatcher dispatches events to registered handlers.
//
// Requirement 1.3: When event loop processes events, system dispatches events in chronological order.
// Requirement 1.4: When event is dispatched, system calls all handlers registered for that event type.
type EventDispatcher struct {
	queue    *EventQueue
	registry *HandlerRegistry
	vm       *VM
}

// NewEventDispatcher creates a new event dispatcher.
func NewEventDispatcher(queue *EventQueue, registry *HandlerRegistry, vm *VM) *EventDispatcher {
	return &EventDispatcher{
		queue:    queue,
		registry: registry,
		vm:       vm,
	}
}

// Dispatch dispatches an event to all registered handlers.
// Handlers are called in registration order.
//
// Requirement 1.4: When event is dispatched, system calls all handlers registered for that event type.
// Requirement 1.5: When multiple handlers are registered for same event type, system executes them in registration order.
func (ed *EventDispatcher) Dispatch(event *Event) error {
	// ログは削除（頻繁すぎるため）

	// Get all handlers for this event type
	handlers := ed.registry.GetHandlers(event.Type)

	// Execute handlers in registration order
	// Requirement 1.5: When multiple handlers are registered for same event type, system executes them in registration order.
	for _, handler := range handlers {
		if handler.Active {
			if err := handler.Execute(event); err != nil {
				if ed.vm != nil {
					ed.vm.log.Error("Handler execution failed", "handler", handler.ID, "error", err)
				}
			}
		}
	}

	// Cleanup handlers marked for deletion
	ed.registry.CleanupMarkedHandlers()

	return nil
}

// ProcessQueue processes all events in the queue.
// Events are processed in chronological order.
//
// Requirement 1.3: When event loop processes events, system dispatches events in chronological order.
// Requirement 14.3: When events are available, system processes them in order.
func (ed *EventDispatcher) ProcessQueue() error {
	for {
		event, ok := ed.queue.Pop()
		if !ok {
			// Queue is empty
			break
		}

		if err := ed.Dispatch(event); err != nil {
			if ed.vm != nil {
				ed.vm.log.Error("Event dispatch failed", "type", event.Type, "error", err)
			}
		}
	}

	return nil
}

// ProcessOne processes a single event from the queue.
// Returns false if the queue is empty.
func (ed *EventDispatcher) ProcessOne() (bool, error) {
	event, ok := ed.queue.Pop()
	if !ok {
		return false, nil
	}

	err := ed.Dispatch(event)
	return true, err
}

// QueueEvent adds an event to the queue.
func (ed *EventDispatcher) QueueEvent(event *Event) {
	ed.queue.Push(event)
}

// GetQueue returns the event queue.
func (ed *EventDispatcher) GetQueue() *EventQueue {
	return ed.queue
}

// GetRegistry returns the handler registry.
func (ed *EventDispatcher) GetRegistry() *HandlerRegistry {
	return ed.registry
}
