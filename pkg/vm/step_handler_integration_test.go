// Package vm provides integration tests for the step handler removal feature.
//
// タスク 4: 統合テストの作成
// - samples/sabo2が1回だけ実行されることを確認するテスト
// - クリックハンドラが複数回動作することを確認するテスト
//
// Requirements: 3.1, 3.2
package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestIntegration_StepHandlerRemovedAfterCompletion tests that a handler with step() block
// is removed after completion and does not re-execute on subsequent events.
//
// This simulates the behavior of samples/sabo2's TIME handler which has a step() block.
// The handler should execute once and then be removed from the registry.
//
// Validates: Requirement 3.1 (mes(LBDOWN) handler without step() executes on every event)
// Note: This test validates the inverse - handlers WITH step() are removed after completion.
func TestIntegration_StepHandlerRemovedAfterCompletion(t *testing.T) {
	vm := New([]opcode.OpCode{})

	// Track execution count
	executionCount := 0

	// Create a handler with step() block (like sabo2's TIME handler)
	// The handler has OpSetStep which indicates it has a step() block
	handlerOpcodes := []opcode.OpCode{
		// step(1) - sets up step block
		{Cmd: opcode.SetStep, Args: []any{int64(1)}},
		// Assign to track execution
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("executed"), int64(1)}},
	}

	handler := NewEventHandler("step-handler", EventTIME, handlerOpcodes, vm, nil)
	// HasStepBlock should be set by executeRegisterEventHandler, but we set it manually here
	// to simulate the behavior after the fix
	handler.HasStepBlock = containsOpSetStep(handlerOpcodes)

	vm.handlerRegistry.Register(handler)

	// Verify HasStepBlock is true
	if !handler.HasStepBlock {
		t.Fatal("Expected HasStepBlock to be true for handler with OpSetStep")
	}

	// First event: execute the handler
	event1 := NewEvent(EventTIME)
	err := handler.Execute(event1)
	if err != nil {
		t.Fatalf("First execution failed: %v", err)
	}

	// Check that the handler executed
	executed, _ := vm.globalScope.Get("executed")
	if executed != int64(1) {
		t.Errorf("Expected executed to be 1, got %v", executed)
	}
	executionCount++

	// Handler should be marked for deletion after completion
	if !handler.MarkedForDeletion {
		t.Error("Expected handler to be marked for deletion after completion")
	}

	// Handler should be inactive
	if handler.Active {
		t.Error("Expected handler to be inactive after completion")
	}

	// Cleanup marked handlers (simulates what EventDispatcher does)
	vm.handlerRegistry.CleanupMarkedHandlers()

	// Verify handler is removed from registry
	handlers := vm.handlerRegistry.GetHandlers(EventTIME)
	if len(handlers) != 0 {
		t.Errorf("Expected 0 handlers after cleanup, got %d", len(handlers))
	}

	// Second event: should not execute anything (handler is removed)
	vm.globalScope.Set("executed", int64(0)) // Reset
	event2 := NewEvent(EventTIME)

	// Try to execute the handler again (should do nothing since it's inactive)
	err = handler.Execute(event2)
	if err != nil {
		t.Fatalf("Second execution failed: %v", err)
	}

	// executed should still be 0 (handler didn't run)
	executed, _ = vm.globalScope.Get("executed")
	if executed != int64(0) {
		t.Errorf("Expected executed to be 0 after handler removal, got %v", executed)
	}
}

// TestIntegration_ClickHandlerContinuesOnMultipleEvents tests that a handler without step() block
// continues to work on multiple events.
//
// This simulates the behavior of mes(LBDOWN) or mes(CLICK) handlers which don't have step() blocks.
// The handler should execute on every event and reset its PC after each execution.
//
// Validates: Requirement 3.1 (mes(LBDOWN) handler without step() executes on every LBDOWN event)
// Validates: Requirement 3.2 (mes(CLICK) handler without step() executes on every CLICK event)
func TestIntegration_ClickHandlerContinuesOnMultipleEvents(t *testing.T) {
	testCases := []struct {
		name      string
		eventType EventType
	}{
		{"LBDOWN handler", EventLBDOWN},
		{"CLICK handler", EventCLICK},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			// Create a handler WITHOUT step() block (like a click handler)
			// No OpSetStep in the opcodes
			handlerOpcodes := []opcode.OpCode{
				// Increment counter on each execution
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("click_count"), int64(0)}},
			}

			handler := NewEventHandler("click-handler", tc.eventType, handlerOpcodes, vm, nil)
			// HasStepBlock should be false since there's no OpSetStep
			handler.HasStepBlock = containsOpSetStep(handlerOpcodes)

			vm.handlerRegistry.Register(handler)

			// Verify HasStepBlock is false
			if handler.HasStepBlock {
				t.Fatal("Expected HasStepBlock to be false for handler without OpSetStep")
			}

			// Execute the handler multiple times
			for i := 1; i <= 5; i++ {
				event := NewEvent(tc.eventType)
				err := handler.Execute(event)
				if err != nil {
					t.Fatalf("Execution %d failed: %v", i, err)
				}

				// Handler should NOT be marked for deletion
				if handler.MarkedForDeletion {
					t.Errorf("Execution %d: Handler should not be marked for deletion", i)
				}

				// Handler should still be active
				if !handler.Active {
					t.Errorf("Execution %d: Handler should still be active", i)
				}

				// PC should be reset to 0 after each execution
				if handler.CurrentPC != 0 {
					t.Errorf("Execution %d: Expected CurrentPC to be 0, got %d", i, handler.CurrentPC)
				}
			}

			// Verify handler is still in registry
			handlers := vm.handlerRegistry.GetHandlers(tc.eventType)
			if len(handlers) != 1 {
				t.Errorf("Expected 1 handler in registry, got %d", len(handlers))
			}
		})
	}
}

// TestIntegration_StepHandlerWithWaitRemovedAfterCompletion tests that a handler with step() block
// and Wait() calls is removed after all steps complete.
//
// This simulates a more realistic scenario like sabo2 where the handler has multiple steps
// with Wait() calls between them.
//
// Validates: Requirement 3.1, 3.2 (handlers with step() are removed, handlers without step() continue)
func TestIntegration_StepHandlerWithWaitRemovedAfterCompletion(t *testing.T) {
	vm := New([]opcode.OpCode{})

	// Create a handler with step() block and Wait() calls
	// This simulates: step(1) { action1; Wait(2); action2; }
	handlerOpcodes := []opcode.OpCode{
		// step(1) - sets up step block
		{Cmd: opcode.SetStep, Args: []any{int64(1)}},
		// First action
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(1)}},
		// Wait for 2 events
		{Cmd: opcode.Wait, Args: []any{int64(2)}},
		// Second action
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(2)}},
	}

	handler := NewEventHandler("step-wait-handler", EventTIME, handlerOpcodes, vm, nil)
	handler.HasStepBlock = containsOpSetStep(handlerOpcodes)

	vm.handlerRegistry.Register(handler)

	// Verify HasStepBlock is true
	if !handler.HasStepBlock {
		t.Fatal("Expected HasStepBlock to be true")
	}

	// Event 1: Execute until Wait
	event1 := NewEvent(EventTIME)
	handler.Execute(event1)

	step, _ := vm.globalScope.Get("step")
	if step != int64(1) {
		t.Errorf("After event 1: expected step to be 1, got %v", step)
	}

	// Handler should still be active (waiting)
	if !handler.Active {
		t.Error("Handler should still be active while waiting")
	}

	// Event 2: Decrement wait counter
	event2 := NewEvent(EventTIME)
	handler.Execute(event2)

	// Still waiting
	if handler.WaitCounter != 1 {
		t.Errorf("Expected WaitCounter to be 1, got %d", handler.WaitCounter)
	}

	// Event 3: Resume and complete
	event3 := NewEvent(EventTIME)
	handler.Execute(event3)

	step, _ = vm.globalScope.Get("step")
	if step != int64(2) {
		t.Errorf("After event 3: expected step to be 2, got %v", step)
	}

	// Handler should be marked for deletion after completion
	if !handler.MarkedForDeletion {
		t.Error("Expected handler to be marked for deletion after completion")
	}

	// Handler should be inactive
	if handler.Active {
		t.Error("Expected handler to be inactive after completion")
	}
}

// TestIntegration_MixedHandlers tests that step handlers and click handlers can coexist
// and behave correctly.
//
// Validates: Requirements 3.1, 3.2
func TestIntegration_MixedHandlers(t *testing.T) {
	vm := New([]opcode.OpCode{})

	// Create a step handler (like sabo2's TIME handler)
	stepHandlerOpcodes := []opcode.OpCode{
		{Cmd: opcode.SetStep, Args: []any{int64(1)}},
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("time_executed"), int64(1)}},
	}
	stepHandler := NewEventHandler("step-handler", EventTIME, stepHandlerOpcodes, vm, nil)
	stepHandler.HasStepBlock = containsOpSetStep(stepHandlerOpcodes)
	vm.handlerRegistry.Register(stepHandler)

	// Create a click handler (without step)
	clickHandlerOpcodes := []opcode.OpCode{
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("click_executed"), int64(1)}},
	}
	clickHandler := NewEventHandler("click-handler", EventLBDOWN, clickHandlerOpcodes, vm, nil)
	clickHandler.HasStepBlock = containsOpSetStep(clickHandlerOpcodes)
	vm.handlerRegistry.Register(clickHandler)

	// Verify initial state
	if !stepHandler.HasStepBlock {
		t.Error("Step handler should have HasStepBlock=true")
	}
	if clickHandler.HasStepBlock {
		t.Error("Click handler should have HasStepBlock=false")
	}

	// Execute TIME event - step handler should complete and be removed
	timeEvent := NewEvent(EventTIME)
	stepHandler.Execute(timeEvent)
	vm.handlerRegistry.CleanupMarkedHandlers()

	// Step handler should be removed
	timeHandlers := vm.handlerRegistry.GetHandlers(EventTIME)
	if len(timeHandlers) != 0 {
		t.Errorf("Expected 0 TIME handlers after step completion, got %d", len(timeHandlers))
	}

	// Execute LBDOWN events multiple times - click handler should continue working
	for i := 0; i < 3; i++ {
		clickEvent := NewEvent(EventLBDOWN)
		clickHandler.Execute(clickEvent)
		vm.handlerRegistry.CleanupMarkedHandlers()

		// Click handler should still be in registry
		clickHandlers := vm.handlerRegistry.GetHandlers(EventLBDOWN)
		if len(clickHandlers) != 1 {
			t.Errorf("Iteration %d: Expected 1 LBDOWN handler, got %d", i, len(clickHandlers))
		}
	}
}

// TestIntegration_Sabo2LikeScenario tests a scenario similar to samples/sabo2.
//
// sabo2 has a mes(TIME) handler with step(2) that plays MIDI and animates windows.
// After the step block completes, the handler should be removed and not re-execute.
//
// Validates: Requirements 3.1, 3.2
func TestIntegration_Sabo2LikeScenario(t *testing.T) {
	vm := New([]opcode.OpCode{})

	// Simulate sabo2's TIME handler structure:
	// mes(TIME) {
	//     step(2) {
	//         PlayMIDI("SAMPLE.MID");
	//         OpenWin(0);
	//         ,,,,,  // Wait for events
	//         MoveWin(0, 1);
	//         ...
	//         end_step;
	//     }
	// }
	//
	// Simplified version for testing:
	handlerOpcodes := []opcode.OpCode{
		// step(2) - each comma waits for 2 TIME events
		{Cmd: opcode.SetStep, Args: []any{int64(2)}},
		// First action (like PlayMIDI)
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("action"), int64(1)}},
		// Wait for 2 events (like one comma with step(2))
		{Cmd: opcode.Wait, Args: []any{int64(2)}},
		// Second action (like MoveWin)
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("action"), int64(2)}},
		// Wait for 2 more events
		{Cmd: opcode.Wait, Args: []any{int64(2)}},
		// Final action
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("action"), int64(3)}},
	}

	handler := NewEventHandler("sabo2-like-handler", EventTIME, handlerOpcodes, vm, nil)
	handler.HasStepBlock = containsOpSetStep(handlerOpcodes)
	vm.handlerRegistry.Register(handler)

	// Track how many times the handler starts from the beginning
	startCount := 0

	// Execute events until handler completes
	for i := 0; i < 10; i++ {
		// Check if handler is starting from beginning
		if handler.CurrentPC == 0 && handler.Active {
			startCount++
		}

		event := NewEvent(EventTIME)
		handler.Execute(event)

		// If handler is marked for deletion, cleanup and stop
		if handler.MarkedForDeletion {
			vm.handlerRegistry.CleanupMarkedHandlers()
			break
		}
	}

	// Handler should have started only once
	if startCount != 1 {
		t.Errorf("Expected handler to start 1 time, but it started %d times", startCount)
	}

	// Handler should be removed from registry
	handlers := vm.handlerRegistry.GetHandlers(EventTIME)
	if len(handlers) != 0 {
		t.Errorf("Expected 0 handlers after completion, got %d", len(handlers))
	}

	// Final action should have been executed
	action, _ := vm.globalScope.Get("action")
	if action != int64(3) {
		t.Errorf("Expected final action to be 3, got %v", action)
	}
}

// TestIntegration_EventDispatcherWithStepHandler tests that EventDispatcher correctly
// handles step handlers and removes them after completion.
//
// Validates: Requirements 3.1, 3.2
func TestIntegration_EventDispatcherWithStepHandler(t *testing.T) {
	vm := New([]opcode.OpCode{})
	queue := NewEventQueue()
	dispatcher := NewEventDispatcher(queue, vm.handlerRegistry, vm)

	// Create a step handler
	stepHandlerOpcodes := []opcode.OpCode{
		{Cmd: opcode.SetStep, Args: []any{int64(1)}},
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("dispatched"), int64(1)}},
	}
	stepHandler := NewEventHandler("step-handler", EventTIME, stepHandlerOpcodes, vm, nil)
	stepHandler.HasStepBlock = containsOpSetStep(stepHandlerOpcodes)
	vm.handlerRegistry.Register(stepHandler)

	// Dispatch first TIME event
	event1 := NewEvent(EventTIME)
	err := dispatcher.Dispatch(event1)
	if err != nil {
		t.Fatalf("First dispatch failed: %v", err)
	}

	// Handler should be removed after dispatch (cleanup happens in Dispatch)
	handlers := vm.handlerRegistry.GetHandlers(EventTIME)
	if len(handlers) != 0 {
		t.Errorf("Expected 0 handlers after first dispatch, got %d", len(handlers))
	}

	// Dispatch second TIME event - should not execute anything
	vm.globalScope.Set("dispatched", int64(0))
	event2 := NewEvent(EventTIME)
	err = dispatcher.Dispatch(event2)
	if err != nil {
		t.Fatalf("Second dispatch failed: %v", err)
	}

	// dispatched should still be 0 (no handler to execute)
	dispatched, _ := vm.globalScope.Get("dispatched")
	if dispatched != int64(0) {
		t.Errorf("Expected dispatched to be 0 after handler removal, got %v", dispatched)
	}
}

// TestIntegration_EventDispatcherWithClickHandler tests that EventDispatcher correctly
// handles click handlers and keeps them active after each execution.
//
// Validates: Requirements 3.1, 3.2
func TestIntegration_EventDispatcherWithClickHandler(t *testing.T) {
	vm := New([]opcode.OpCode{})
	queue := NewEventQueue()
	dispatcher := NewEventDispatcher(queue, vm.handlerRegistry, vm)

	// Create a click handler (without step)
	clickHandlerOpcodes := []opcode.OpCode{
		{Cmd: opcode.Assign, Args: []any{opcode.Variable("click_count"), int64(0)}},
	}
	clickHandler := NewEventHandler("click-handler", EventLBDOWN, clickHandlerOpcodes, vm, nil)
	clickHandler.HasStepBlock = containsOpSetStep(clickHandlerOpcodes)
	vm.handlerRegistry.Register(clickHandler)

	// Dispatch multiple LBDOWN events
	for i := 0; i < 5; i++ {
		event := NewEvent(EventLBDOWN)
		err := dispatcher.Dispatch(event)
		if err != nil {
			t.Fatalf("Dispatch %d failed: %v", i+1, err)
		}

		// Handler should still be in registry
		handlers := vm.handlerRegistry.GetHandlers(EventLBDOWN)
		if len(handlers) != 1 {
			t.Errorf("Dispatch %d: Expected 1 handler, got %d", i+1, len(handlers))
		}
	}
}
