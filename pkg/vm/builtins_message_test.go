package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// helperRegisterHandler creates and registers an EventHandler with the given VM and event type.
// Returns the registered handler. The handler's numeric ID corresponds to the registration order (1-based).
func helperRegisterHandler(t *testing.T, vm *VM, eventType EventType) *EventHandler {
	t.Helper()
	handler := NewEventHandler("", eventType, []opcode.OpCode{}, vm, nil)
	vm.handlerRegistry.Register(handler)
	return handler
}

// TestFreezeMes tests the FreezeMes builtin function.
// Requirements: 1.1, 1.4, 1.5
func TestFreezeMes(t *testing.T) {
	t.Run("normal: sets Active to false", func(t *testing.T) {
		// Requirement 1.1: FreezeMes sets Active flag to false.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)

		if !handler.Active {
			t.Fatal("handler should be active initially")
		}

		result, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("FreezeMes returned error: %v", err)
		}
		if result != nil {
			t.Errorf("FreezeMes should return nil, got %v", result)
		}
		if handler.Active {
			t.Error("handler should be inactive after FreezeMes")
		}
	})

	t.Run("preserves CurrentPC and WaitCounter", func(t *testing.T) {
		// Requirement 1.3: FreezeMes preserves CurrentPC and WaitCounter.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)
		handler.CurrentPC = 42
		handler.WaitCounter = 7

		_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("FreezeMes returned error: %v", err)
		}

		if handler.CurrentPC != 42 {
			t.Errorf("CurrentPC should be preserved, got %d, want 42", handler.CurrentPC)
		}
		if handler.WaitCounter != 7 {
			t.Errorf("WaitCounter should be preserved, got %d, want 7", handler.WaitCounter)
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		// Requirement 1.5: Missing arguments logs warning only.
		vm := New([]opcode.OpCode{})

		result, err := vm.builtins["FreezeMes"](vm, []any{})
		if err != nil {
			t.Fatalf("FreezeMes should not return error on missing args, got: %v", err)
		}
		if result != nil {
			t.Errorf("FreezeMes should return nil on missing args, got %v", result)
		}
	})

	t.Run("non-existent handler number", func(t *testing.T) {
		// Requirement 1.4: Non-existent handler number logs warning only.
		vm := New([]opcode.OpCode{})

		result, err := vm.builtins["FreezeMes"](vm, []any{int64(999)})
		if err != nil {
			t.Fatalf("FreezeMes should not return error for non-existent handler, got: %v", err)
		}
		if result != nil {
			t.Errorf("FreezeMes should return nil for non-existent handler, got %v", result)
		}
	})

	t.Run("idempotent: freeze already frozen handler", func(t *testing.T) {
		// Freezing an already frozen handler should not cause errors.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)

		// First freeze
		_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("first FreezeMes returned error: %v", err)
		}
		if handler.Active {
			t.Fatal("handler should be inactive after first FreezeMes")
		}

		// Second freeze (idempotent)
		result, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("second FreezeMes returned error: %v", err)
		}
		if result != nil {
			t.Errorf("second FreezeMes should return nil, got %v", result)
		}
		if handler.Active {
			t.Error("handler should still be inactive after second FreezeMes")
		}
	})

	t.Run("does not mark handler for deletion", func(t *testing.T) {
		// FreezeMes should NOT call Remove() — handler stays in registry.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)

		_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("FreezeMes returned error: %v", err)
		}

		if handler.MarkedForDeletion {
			t.Error("FreezeMes should not mark handler for deletion")
		}

		// Handler should still be findable in the registry
		_, exists := vm.handlerRegistry.GetHandlerByNumber(1)
		if !exists {
			t.Error("frozen handler should still exist in registry")
		}
	})
}

// TestActivateMes tests the ActivateMes builtin function.
// Requirements: 2.1, 2.3, 2.4
func TestActivateMes(t *testing.T) {
	t.Run("normal: sets Active to true", func(t *testing.T) {
		// Requirement 2.1: ActivateMes sets Active flag to true.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)

		// First freeze the handler
		handler.Active = false

		result, err := vm.builtins["ActivateMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("ActivateMes returned error: %v", err)
		}
		if result != nil {
			t.Errorf("ActivateMes should return nil, got %v", result)
		}
		if !handler.Active {
			t.Error("handler should be active after ActivateMes")
		}
	})

	t.Run("preserves CurrentPC and WaitCounter on reactivation", func(t *testing.T) {
		// Requirement 2.2: Reactivated handler resumes from previous CurrentPC and WaitCounter.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)
		handler.CurrentPC = 10
		handler.WaitCounter = 3

		// Freeze
		_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("FreezeMes returned error: %v", err)
		}

		// Activate
		_, err = vm.builtins["ActivateMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("ActivateMes returned error: %v", err)
		}

		if !handler.Active {
			t.Error("handler should be active after ActivateMes")
		}
		if handler.CurrentPC != 10 {
			t.Errorf("CurrentPC should be preserved after round-trip, got %d, want 10", handler.CurrentPC)
		}
		if handler.WaitCounter != 3 {
			t.Errorf("WaitCounter should be preserved after round-trip, got %d, want 3", handler.WaitCounter)
		}
	})

	t.Run("missing arguments", func(t *testing.T) {
		// Requirement 2.4: Missing arguments logs warning only.
		vm := New([]opcode.OpCode{})

		result, err := vm.builtins["ActivateMes"](vm, []any{})
		if err != nil {
			t.Fatalf("ActivateMes should not return error on missing args, got: %v", err)
		}
		if result != nil {
			t.Errorf("ActivateMes should return nil on missing args, got %v", result)
		}
	})

	t.Run("non-existent handler number", func(t *testing.T) {
		// Requirement 2.3: Non-existent handler number logs warning only.
		vm := New([]opcode.OpCode{})

		result, err := vm.builtins["ActivateMes"](vm, []any{int64(999)})
		if err != nil {
			t.Fatalf("ActivateMes should not return error for non-existent handler, got: %v", err)
		}
		if result != nil {
			t.Errorf("ActivateMes should return nil for non-existent handler, got %v", result)
		}
	})

	t.Run("idempotent: activate already active handler", func(t *testing.T) {
		// Activating an already active handler should not cause errors.
		vm := New([]opcode.OpCode{})
		handler := helperRegisterHandler(t, vm, EventTIME)

		// Handler is already active by default
		if !handler.Active {
			t.Fatal("handler should be active initially")
		}

		result, err := vm.builtins["ActivateMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("ActivateMes on already active handler returned error: %v", err)
		}
		if result != nil {
			t.Errorf("ActivateMes should return nil, got %v", result)
		}
		if !handler.Active {
			t.Error("handler should still be active")
		}
	})

	t.Run("multiple handlers: freeze and activate specific one", func(t *testing.T) {
		// Verify that FreezeMes/ActivateMes targets only the specified handler.
		vm := New([]opcode.OpCode{})
		handler1 := helperRegisterHandler(t, vm, EventTIME)  // handler number 1
		handler2 := helperRegisterHandler(t, vm, EventCLICK) // handler number 2

		// Freeze handler 1 only
		_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("FreezeMes returned error: %v", err)
		}

		if handler1.Active {
			t.Error("handler1 should be inactive after FreezeMes(1)")
		}
		if !handler2.Active {
			t.Error("handler2 should still be active (not targeted)")
		}

		// Activate handler 1
		_, err = vm.builtins["ActivateMes"](vm, []any{int64(1)})
		if err != nil {
			t.Fatalf("ActivateMes returned error: %v", err)
		}

		if !handler1.Active {
			t.Error("handler1 should be active after ActivateMes(1)")
		}
		if !handler2.Active {
			t.Error("handler2 should still be active")
		}
	})
}
