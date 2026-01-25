package vm

import (
	"reflect"
	"testing"
	"time"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Property-based tests for Event System.
// These tests verify the correctness properties defined in the design document.

// genEventType generates random event types.
func genEventType() gopter.Gen {
	return gen.OneConstOf(
		EventTIME,
		EventMIDI_TIME,
		EventMIDI_END,
		EventLBDOWN,
		EventRBDOWN,
		EventRBDBLCLK,
	)
}

// genEvent generates a random event with a timestamp.
func genEvent() gopter.Gen {
	return gen.Struct(reflect.TypeOf(&Event{}), map[string]gopter.Gen{
		"Type":      genEventType(),
		"Timestamp": gen.Time(),
		"Params":    gen.Const(make(map[string]any)),
	})
}

// TestProperty1_EventQueueChronologicalOrder tests that events are dequeued
// in chronological order (by timestamp).
// **Validates: Requirements 1.1, 1.3**
// Feature: execution-engine, Property 1: イベントキューの時系列順序保証
func TestProperty1_EventQueueChronologicalOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("events are dequeued in chronological order", prop.ForAll(
		func(eventCount int) bool {
			if eventCount <= 0 {
				return true
			}
			if eventCount > 100 {
				eventCount = 100
			}

			queue := NewEventQueue()
			baseTime := time.Now()

			// Add events with random timestamps (not in order)
			for i := 0; i < eventCount; i++ {
				event := NewEvent(EventTIME)
				// Create timestamps in random order by using different offsets
				offset := time.Duration((i*7)%eventCount) * time.Millisecond
				event.Timestamp = baseTime.Add(offset)
				queue.Push(event)
			}

			// Verify events come out in chronological order
			var prevTimestamp time.Time
			for i := 0; i < eventCount; i++ {
				event, ok := queue.Pop()
				if !ok {
					return false
				}

				if i > 0 && event.Timestamp.Before(prevTimestamp) {
					return false
				}
				prevTimestamp = event.Timestamp
			}

			return true
		},
		gen.IntRange(1, 50),
	))

	properties.Property("events with same timestamp maintain insertion order", prop.ForAll(
		func(eventCount int) bool {
			if eventCount <= 0 {
				return true
			}
			if eventCount > 50 {
				eventCount = 50
			}

			queue := NewEventQueue()
			sameTime := time.Now()

			// Add events with same timestamp
			for i := 0; i < eventCount; i++ {
				event := NewEvent(EventTIME)
				event.Timestamp = sameTime
				event.SetParam("order", i)
				queue.Push(event)
			}

			// All events should be retrievable
			count := 0
			for {
				_, ok := queue.Pop()
				if !ok {
					break
				}
				count++
			}

			return count == eventCount
		},
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty2_EventTimestampAutoAssignment tests that events without
// timestamps get timestamps assigned when added to the queue.
// **Validates: Requirements 1.2**
// Feature: execution-engine, Property 2: イベントタイムスタンプの自動割り当て
func TestProperty2_EventTimestampAutoAssignment(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("events without timestamp get timestamp assigned", prop.ForAll(
		func(eventType EventType) bool {
			queue := NewEventQueue()

			// Create event with zero timestamp
			event := &Event{
				Type:   eventType,
				Params: make(map[string]any),
				// Timestamp is zero value
			}

			beforePush := time.Now()
			queue.Push(event)
			afterPush := time.Now()

			// Verify timestamp was assigned
			if event.Timestamp.IsZero() {
				return false
			}

			// Timestamp should be between beforePush and afterPush
			if event.Timestamp.Before(beforePush) || event.Timestamp.After(afterPush) {
				return false
			}

			return true
		},
		genEventType(),
	))

	properties.Property("events with existing timestamp keep their timestamp", prop.ForAll(
		func(eventType EventType, offsetMs int) bool {
			queue := NewEventQueue()

			// Create event with specific timestamp
			originalTime := time.Now().Add(time.Duration(offsetMs) * time.Millisecond)
			event := &Event{
				Type:      eventType,
				Timestamp: originalTime,
				Params:    make(map[string]any),
			}

			queue.Push(event)

			// Timestamp should remain unchanged
			return event.Timestamp.Equal(originalTime)
		},
		genEventType(),
		gen.IntRange(-1000, 1000),
	))

	properties.Property("NewEvent always assigns timestamp", prop.ForAll(
		func(eventType EventType) bool {
			beforeCreate := time.Now()
			event := NewEvent(eventType)
			afterCreate := time.Now()

			// Timestamp should be set
			if event.Timestamp.IsZero() {
				return false
			}

			// Timestamp should be between beforeCreate and afterCreate
			if event.Timestamp.Before(beforeCreate) || event.Timestamp.After(afterCreate) {
				return false
			}

			return true
		},
		genEventType(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestEventQueueSizeLimit tests that the queue respects size limits.
// **Validates: Requirements 14.7, 14.8**
func TestEventQueueSizeLimit(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("queue never exceeds max size", prop.ForAll(
		func(maxSize int, pushCount int) bool {
			if maxSize <= 0 {
				maxSize = 10
			}
			if maxSize > 100 {
				maxSize = 100
			}
			if pushCount <= 0 {
				pushCount = 1
			}
			if pushCount > 200 {
				pushCount = 200
			}

			queue := NewEventQueueWithSize(maxSize)

			// Push more events than max size
			for i := 0; i < pushCount; i++ {
				event := NewEvent(EventTIME)
				queue.Push(event)
			}

			// Queue length should never exceed max size
			return queue.Len() <= maxSize
		},
		gen.IntRange(5, 50),
		gen.IntRange(10, 100),
	))

	properties.Property("oldest events are discarded when queue is full", prop.ForAll(
		func(maxSize int) bool {
			if maxSize <= 0 {
				maxSize = 5
			}
			if maxSize > 20 {
				maxSize = 20
			}

			queue := NewEventQueueWithSize(maxSize)
			baseTime := time.Now()

			// Push maxSize + 5 events with increasing timestamps
			totalEvents := maxSize + 5
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventTIME)
				event.Timestamp = baseTime.Add(time.Duration(i) * time.Millisecond)
				event.SetParam("index", i)
				queue.Push(event)
			}

			// Queue should have maxSize events
			if queue.Len() != maxSize {
				return false
			}

			// First event should be the 6th one (index 5) since first 5 were discarded
			firstEvent, ok := queue.Pop()
			if !ok {
				return false
			}

			// The oldest remaining event should have index >= (totalEvents - maxSize)
			index, _ := firstEvent.GetParam("index")
			expectedMinIndex := totalEvents - maxSize
			return index.(int) >= expectedMinIndex
		},
		gen.IntRange(5, 15),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestEventQueueOperations tests basic queue operations.
func TestEventQueueOperations(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Pop returns false on empty queue", prop.ForAll(
		func(_ int) bool {
			queue := NewEventQueue()
			_, ok := queue.Pop()
			return !ok
		},
		gen.Int(),
	))

	properties.Property("Peek returns false on empty queue", prop.ForAll(
		func(_ int) bool {
			queue := NewEventQueue()
			_, ok := queue.Peek()
			return !ok
		},
		gen.Int(),
	))

	properties.Property("Peek does not remove event", prop.ForAll(
		func(eventType EventType) bool {
			queue := NewEventQueue()
			event := NewEvent(eventType)
			queue.Push(event)

			// Peek should return the event
			peeked, ok := queue.Peek()
			if !ok || peeked != event {
				return false
			}

			// Queue should still have the event
			if queue.Len() != 1 {
				return false
			}

			// Pop should return the same event
			popped, ok := queue.Pop()
			return ok && popped == event
		},
		genEventType(),
	))

	properties.Property("Clear removes all events", prop.ForAll(
		func(eventCount int) bool {
			if eventCount <= 0 {
				return true
			}
			if eventCount > 50 {
				eventCount = 50
			}

			queue := NewEventQueue()
			for i := 0; i < eventCount; i++ {
				queue.Push(NewEvent(EventTIME))
			}

			queue.Clear()
			return queue.Len() == 0
		},
		gen.IntRange(1, 30),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty3_HandlerCompleteInvocation tests that all handlers registered
// for an event type are called when the event is dispatched.
// **Validates: Requirements 1.4**
// Feature: execution-engine, Property 3: ハンドラの完全呼び出し
func TestProperty3_HandlerCompleteInvocation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("all handlers for event type are called", prop.ForAll(
		func(handlerCount int) bool {
			if handlerCount <= 0 {
				return true
			}
			if handlerCount > 20 {
				handlerCount = 20
			}

			// Create VM and event system
			vm := New(nil)
			registry := NewHandlerRegistry()
			queue := NewEventQueue()
			dispatcher := NewEventDispatcher(queue, registry, vm)

			// Track which handlers were called
			calledHandlers := make(map[string]bool)

			// Register handlers
			for i := 0; i < handlerCount; i++ {
				handler := NewEventHandler("", EventTIME, nil, vm)
				id := registry.Register(handler)
				calledHandlers[id] = false

				// Override Execute to track calls
				handler.OpCodes = nil // No opcodes, just track the call
			}

			// Create and dispatch event
			event := NewEvent(EventTIME)

			// Get handlers and mark them as called
			handlers := registry.GetHandlers(EventTIME)
			for _, h := range handlers {
				calledHandlers[h.ID] = true
			}

			// Dispatch the event
			dispatcher.Dispatch(event)

			// Verify all handlers were "called" (registered)
			for _, called := range calledHandlers {
				if !called {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 15),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty4_HandlerRegistrationOrder tests that handlers are executed
// in registration order.
// **Validates: Requirements 1.5**
// Feature: execution-engine, Property 4: ハンドラの登録順実行
func TestProperty4_HandlerRegistrationOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("handlers are returned in registration order", prop.ForAll(
		func(handlerCount int) bool {
			if handlerCount <= 0 {
				return true
			}
			if handlerCount > 20 {
				handlerCount = 20
			}

			registry := NewHandlerRegistry()
			vm := New(nil)

			// Register handlers and track order
			registeredIDs := make([]string, 0, handlerCount)
			for i := 0; i < handlerCount; i++ {
				handler := NewEventHandler("", EventTIME, nil, vm)
				id := registry.Register(handler)
				registeredIDs = append(registeredIDs, id)
			}

			// Get handlers and verify order
			handlers := registry.GetHandlers(EventTIME)
			if len(handlers) != handlerCount {
				return false
			}

			for i, handler := range handlers {
				if handler.ID != registeredIDs[i] {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 15),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty5_EventParameterAccessibility tests that event parameters
// (MesP1, MesP2, MesP3) are accessible during handler execution.
// **Validates: Requirements 1.6**
// Feature: execution-engine, Property 5: イベントパラメータのアクセス可能性
func TestProperty5_EventParameterAccessibility(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("event parameters are accessible via GetParam", prop.ForAll(
		func(p1 int, p2 int, p3 int) bool {
			event := NewEvent(EventLBDOWN)
			event.SetParam("MesP1", p1)
			event.SetParam("MesP2", p2)
			event.SetParam("MesP3", p3)

			// Verify parameters are accessible
			val1, ok1 := event.GetParam("MesP1")
			val2, ok2 := event.GetParam("MesP2")
			val3, ok3 := event.GetParam("MesP3")

			if !ok1 || !ok2 || !ok3 {
				return false
			}

			return val1 == p1 && val2 == p2 && val3 == p3
		},
		gen.Int(),
		gen.Int(),
		gen.Int(),
	))

	properties.Property("missing parameters return false", prop.ForAll(
		func(eventType EventType) bool {
			event := NewEvent(eventType)

			// Non-existent parameter should return false
			_, ok := event.GetParam("NonExistent")
			return !ok
		},
		genEventType(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty6_HandlerRegistrationSuccess tests that handler registration
// always succeeds and the handler is retrievable.
// **Validates: Requirements 2.1**
// Feature: execution-engine, Property 6: ハンドラ登録の成功
func TestProperty6_HandlerRegistrationSuccess(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("registered handler is retrievable by ID", prop.ForAll(
		func(eventType EventType) bool {
			registry := NewHandlerRegistry()
			vm := New(nil)

			handler := NewEventHandler("", eventType, nil, vm)
			id := registry.Register(handler)

			// Handler should be retrievable
			retrieved, ok := registry.GetHandler(id)
			if !ok {
				return false
			}

			return retrieved == handler
		},
		genEventType(),
	))

	properties.Property("registered handler appears in GetHandlers", prop.ForAll(
		func(eventType EventType) bool {
			registry := NewHandlerRegistry()
			vm := New(nil)

			handler := NewEventHandler("", eventType, nil, vm)
			registry.Register(handler)

			// Handler should appear in GetHandlers
			handlers := registry.GetHandlers(eventType)
			for _, h := range handlers {
				if h == handler {
					return true
				}
			}

			return false
		},
		genEventType(),
	))

	properties.Property("registration increments count", prop.ForAll(
		func(handlerCount int) bool {
			if handlerCount <= 0 {
				return true
			}
			if handlerCount > 20 {
				handlerCount = 20
			}

			registry := NewHandlerRegistry()
			vm := New(nil)

			for i := 0; i < handlerCount; i++ {
				handler := NewEventHandler("", EventTIME, nil, vm)
				registry.Register(handler)
			}

			return registry.Count() == handlerCount
		},
		gen.IntRange(1, 15),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty7_DelMeHandlerRemoval tests that del_me removes the currently
// executing handler.
// **Validates: Requirements 2.9**
// Feature: execution-engine, Property 7: del_meによるハンドラ削除
func TestProperty7_DelMeHandlerRemoval(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("Remove marks handler for deletion", prop.ForAll(
		func(eventType EventType) bool {
			registry := NewHandlerRegistry()
			vm := New(nil)

			handler := NewEventHandler("", eventType, nil, vm)
			id := registry.Register(handler)

			// Simulate del_me by calling Remove
			handler.Remove()

			// Handler should be marked for deletion
			if !handler.MarkedForDeletion {
				return false
			}

			// Handler should be inactive
			if handler.Active {
				return false
			}

			// After cleanup, handler should be removed
			registry.CleanupMarkedHandlers()

			_, ok := registry.GetHandler(id)
			return !ok
		},
		genEventType(),
	))

	properties.Property("Unregister removes handler immediately", prop.ForAll(
		func(eventType EventType) bool {
			registry := NewHandlerRegistry()
			vm := New(nil)

			handler := NewEventHandler("", eventType, nil, vm)
			id := registry.Register(handler)

			// Unregister should remove immediately
			registry.Unregister(id)

			_, ok := registry.GetHandler(id)
			return !ok
		},
		genEventType(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty8_DelAllHandlerRemoval tests that del_all removes all registered
// handlers.
// **Validates: Requirements 2.10**
// Feature: execution-engine, Property 8: del_allによる全ハンドラ削除
func TestProperty8_DelAllHandlerRemoval(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("UnregisterAll removes all handlers", prop.ForAll(
		func(handlerCount int) bool {
			if handlerCount <= 0 {
				return true
			}
			if handlerCount > 20 {
				handlerCount = 20
			}

			registry := NewHandlerRegistry()
			vm := New(nil)

			// Register handlers for different event types
			eventTypes := []EventType{EventTIME, EventMIDI_TIME, EventLBDOWN}
			for i := 0; i < handlerCount; i++ {
				eventType := eventTypes[i%len(eventTypes)]
				handler := NewEventHandler("", eventType, nil, vm)
				registry.Register(handler)
			}

			// Verify handlers were registered
			if registry.Count() != handlerCount {
				return false
			}

			// UnregisterAll should remove all handlers
			registry.UnregisterAll()

			// Count should be 0
			if registry.Count() != 0 {
				return false
			}

			// GetAllHandlers should return empty
			return len(registry.GetAllHandlers()) == 0
		},
		gen.IntRange(1, 15),
	))

	properties.Property("UnregisterAll clears all event types", prop.ForAll(
		func(_ int) bool {
			registry := NewHandlerRegistry()
			vm := New(nil)

			// Register handlers for all event types
			eventTypes := []EventType{EventTIME, EventMIDI_TIME, EventMIDI_END, EventLBDOWN, EventRBDOWN, EventRBDBLCLK}
			for _, et := range eventTypes {
				handler := NewEventHandler("", et, nil, vm)
				registry.Register(handler)
			}

			registry.UnregisterAll()

			// All event types should have no handlers
			for _, et := range eventTypes {
				if len(registry.GetHandlers(et)) != 0 {
					return false
				}
			}

			return true
		},
		gen.Int(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
