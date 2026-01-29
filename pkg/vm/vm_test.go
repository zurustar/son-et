package vm

import (
	"fmt"
	"image/color"
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/opcode"
	"github.com/zurustar/son-et/pkg/graphics"
)

// TestNewVM tests the VM constructor with various options.
func TestNewVM(t *testing.T) {
	t.Run("creates VM with empty opcodes", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		if vm == nil {
			t.Fatal("expected VM to be created")
		}
		if vm.pc != 0 {
			t.Errorf("expected pc to be 0, got %d", vm.pc)
		}
		if vm.globalScope == nil {
			t.Error("expected globalScope to be initialized")
		}
		if vm.running {
			t.Error("expected running to be false")
		}
	})

	t.Run("creates VM with opcodes", func(t *testing.T) {
		opcodes := []opcode.OpCode{
			{Cmd: opcode.DefineFunction, Args: []any{"main", []any{}, []opcode.OpCode{}}},
		}
		vm := New(opcodes)
		if len(vm.opcodes) != 1 {
			t.Errorf("expected 1 opcode, got %d", len(vm.opcodes))
		}
	})

	t.Run("applies headless option", func(t *testing.T) {
		vm := New([]opcode.OpCode{}, WithHeadless(true))
		if !vm.headless {
			t.Error("expected headless to be true")
		}
	})

	t.Run("applies timeout option", func(t *testing.T) {
		timeout := 5 * time.Second
		vm := New([]opcode.OpCode{}, WithTimeout(timeout))
		if vm.timeout != timeout {
			t.Errorf("expected timeout to be %v, got %v", timeout, vm.timeout)
		}
	})

	t.Run("applies multiple options", func(t *testing.T) {
		timeout := 10 * time.Second
		vm := New([]opcode.OpCode{}, WithHeadless(true), WithTimeout(timeout))
		if !vm.headless {
			t.Error("expected headless to be true")
		}
		if vm.timeout != timeout {
			t.Errorf("expected timeout to be %v, got %v", timeout, vm.timeout)
		}
	})
}

// TestVMRun tests the VM Run method.
func TestVMRun(t *testing.T) {
	t.Run("runs empty opcode sequence", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		err := vm.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("prevents double run", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Start first run in goroutine
		done := make(chan error)
		go func() {
			done <- vm.Run()
		}()

		// Wait a bit for the first run to start
		time.Sleep(10 * time.Millisecond)

		// The first run should complete quickly since there are no opcodes
		err := <-done
		if err != nil {
			t.Errorf("expected no error from first run, got %v", err)
		}
	})

	t.Run("respects timeout", func(t *testing.T) {
		// Create a VM with a very short timeout
		vm := New([]opcode.OpCode{}, WithTimeout(50*time.Millisecond))

		start := time.Now()
		err := vm.Run()
		elapsed := time.Since(start)

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should complete quickly since there are no opcodes
		if elapsed > 100*time.Millisecond {
			t.Errorf("expected quick completion, took %v", elapsed)
		}
	})

	t.Run("collects function definitions", func(t *testing.T) {
		opcodes := []opcode.OpCode{
			{
				Cmd: opcode.DefineFunction,
				Args: []any{
					"testFunc",
					[]any{
						map[string]any{"name": "x", "type": "int", "isArray": false},
					},
					[]opcode.OpCode{},
				},
			},
		}
		vm := New(opcodes)
		err := vm.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if _, ok := vm.functions["testFunc"]; !ok {
			t.Error("expected function 'testFunc' to be registered")
		}
	})
}

// TestVMStop tests the VM Stop method.
func TestVMStop(t *testing.T) {
	t.Run("stops running VM", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Start VM in goroutine
		done := make(chan error)
		go func() {
			done <- vm.Run()
		}()

		// Wait a bit then stop
		time.Sleep(10 * time.Millisecond)
		vm.Stop()

		// Should complete without error
		err := <-done
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("stop on non-running VM is safe", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Should not panic
		vm.Stop()
	})
}

// TestVMIsRunning tests the IsRunning method.
func TestVMIsRunning(t *testing.T) {
	t.Run("returns false when not running", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		if vm.IsRunning() {
			t.Error("expected IsRunning to be false")
		}
	})
}

// TestVMStackFrame tests stack frame management.
func TestVMStackFrame(t *testing.T) {
	t.Run("pushes and pops stack frames", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Initial depth should be 0
		if vm.GetStackDepth() != 0 {
			t.Errorf("expected stack depth 0, got %d", vm.GetStackDepth())
		}

		// Push a frame
		scope := NewScope(vm.globalScope)
		err := vm.PushStackFrame("func1", scope)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if vm.GetStackDepth() != 1 {
			t.Errorf("expected stack depth 1, got %d", vm.GetStackDepth())
		}

		// Push another frame
		scope2 := NewScope(scope)
		err = vm.PushStackFrame("func2", scope2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if vm.GetStackDepth() != 2 {
			t.Errorf("expected stack depth 2, got %d", vm.GetStackDepth())
		}

		// Pop a frame
		frame, err := vm.PopStackFrame()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if frame.FunctionName != "func2" {
			t.Errorf("expected function name 'func2', got '%s'", frame.FunctionName)
		}
		if vm.GetStackDepth() != 1 {
			t.Errorf("expected stack depth 1, got %d", vm.GetStackDepth())
		}

		// Pop last frame
		frame, err = vm.PopStackFrame()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if frame.FunctionName != "func1" {
			t.Errorf("expected function name 'func1', got '%s'", frame.FunctionName)
		}
		if vm.GetStackDepth() != 0 {
			t.Errorf("expected stack depth 0, got %d", vm.GetStackDepth())
		}
	})

	t.Run("detects stack overflow", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Push MaxStackDepth frames
		for i := 0; i < MaxStackDepth; i++ {
			scope := NewScope(vm.globalScope)
			err := vm.PushStackFrame("func", scope)
			if err != nil {
				t.Fatalf("unexpected error at depth %d: %v", i, err)
			}
		}

		// Next push should fail
		scope := NewScope(vm.globalScope)
		err := vm.PushStackFrame("overflow", scope)
		if err == nil {
			t.Error("expected stack overflow error")
		}
	})

	t.Run("pop from empty stack returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		_, err := vm.PopStackFrame()
		if err == nil {
			t.Error("expected error when popping from empty stack")
		}
	})
}

// TestVMBuiltinFunctions tests built-in function registration.
func TestVMBuiltinFunctions(t *testing.T) {
	t.Run("registers built-in function", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		vm.RegisterBuiltinFunction("testFunc", func(vm *VM, args []any) (any, error) {
			return 42, nil
		})

		// Verify function is registered
		if _, ok := vm.builtins["testFunc"]; !ok {
			t.Error("expected built-in function to be registered")
		}
	})
}

// TestVMGetScope tests scope access methods.
func TestVMGetScope(t *testing.T) {
	t.Run("returns global scope", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		scope := vm.GetGlobalScope()
		if scope == nil {
			t.Error("expected global scope to be non-nil")
		}
		if scope != vm.globalScope {
			t.Error("expected GetGlobalScope to return globalScope")
		}
	})

	t.Run("returns current scope", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Without local scope, should return global
		scope := vm.GetCurrentScope()
		if scope != vm.globalScope {
			t.Error("expected current scope to be global scope")
		}

		// With local scope, should return local
		localScope := NewScope(vm.globalScope)
		vm.PushStackFrame("test", localScope)

		scope = vm.GetCurrentScope()
		if scope != localScope {
			t.Error("expected current scope to be local scope")
		}
	})
}

// TestVMBuiltinPlayMIDI tests the PlayMIDI built-in function registration.
// Requirement 10.1: When PlayMIDI is called, system calls MIDI playback function.
func TestVMBuiltinPlayMIDI(t *testing.T) {
	t.Run("PlayMIDI is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify PlayMIDI is registered
		if _, ok := vm.builtins["PlayMIDI"]; !ok {
			t.Error("expected PlayMIDI to be registered as built-in function")
		}
	})

	t.Run("PlayMIDI handles missing argument with error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call PlayMIDI without arguments - should return error
		fn := vm.builtins["PlayMIDI"]
		result, err := fn(vm, []any{})

		// Should return error for missing argument
		if err == nil {
			t.Errorf("expected error for missing argument, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PlayMIDI handles wrong argument type gracefully", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call PlayMIDI with wrong type - should not panic
		fn := vm.builtins["PlayMIDI"]
		result, err := fn(vm, []any{123}) // int instead of string

		// Should return nil without error (error is logged)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PlayMIDI handles missing audio system gracefully", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call PlayMIDI without audio system - should not panic
		fn := vm.builtins["PlayMIDI"]
		result, err := fn(vm, []any{"test.mid"})

		// Should return nil without error (error is logged)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

// TestVMBuiltinPlayWAVE tests the PlayWAVE built-in function registration.
// Requirement 10.2: When PlayWAVE is called, system calls WAV playback function.
func TestVMBuiltinPlayWAVE(t *testing.T) {
	t.Run("PlayWAVE is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify PlayWAVE is registered
		if _, ok := vm.builtins["PlayWAVE"]; !ok {
			t.Error("expected PlayWAVE to be registered as built-in function")
		}
	})

	t.Run("PlayWAVE handles missing argument with error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call PlayWAVE without arguments - should return error
		fn := vm.builtins["PlayWAVE"]
		result, err := fn(vm, []any{})

		// Should return error for missing argument
		if err == nil {
			t.Errorf("expected error for missing argument, got nil")
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PlayWAVE handles wrong argument type gracefully", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call PlayWAVE with wrong type - should not panic
		fn := vm.builtins["PlayWAVE"]
		result, err := fn(vm, []any{123}) // int instead of string

		// Should return nil without error (error is logged)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PlayWAVE handles missing audio system gracefully", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call PlayWAVE without audio system - should not panic
		fn := vm.builtins["PlayWAVE"]
		result, err := fn(vm, []any{"test.wav"})

		// Should return nil without error (error is logged)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

// TestVMDefaultBuiltins tests that all default built-in functions are registered.
func TestVMDefaultBuiltins(t *testing.T) {
	vm := New([]opcode.OpCode{})

	expectedBuiltins := []string{
		"del_me",
		"del_us",
		"del_all",
		"PlayMIDI",
		"PlayWAVE",
	}

	for _, name := range expectedBuiltins {
		t.Run("has "+name+" built-in", func(t *testing.T) {
			if _, ok := vm.builtins[name]; !ok {
				t.Errorf("expected %s to be registered as built-in function", name)
			}
		})
	}
}

// TestOpWait tests the OpWait implementation.
// Requirement 6.2: When OpWait OpCode is executed, system pauses execution until next event.
// Requirement 6.3: When event occurs during step execution, system proceeds to next step.
func TestOpWait(t *testing.T) {
	t.Run("OpWait sets handler wait counter", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(2)}}, // Wait for 2 events
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Execute the handler with a TIME event
		event := NewEvent(EventTIME)
		err := handler.Execute(event)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Check that x was assigned
		x, _ := vm.globalScope.Get("x")
		if x != int64(1) {
			t.Errorf("expected x to be 1, got %v", x)
		}

		// Check that y was NOT assigned (handler should have paused)
		y, _ := vm.globalScope.Get("y")
		if y != nil && y != int64(0) {
			t.Errorf("expected y to be nil or 0, got %v", y)
		}

		// Check that wait counter is set
		if handler.WaitCounter != 2 {
			t.Errorf("expected wait counter to be 2, got %d", handler.WaitCounter)
		}

		// Check that CurrentPC is at the correct position (after OpWait)
		if handler.CurrentPC != 2 {
			t.Errorf("expected CurrentPC to be 2, got %d", handler.CurrentPC)
		}
	})

	t.Run("OpWait decrements wait counter on event", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(2)}}, // Wait for 2 events
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// First event: execute until OpWait
		event1 := NewEvent(EventTIME)
		handler.Execute(event1)

		// Second event: decrement wait counter (still waiting)
		event2 := NewEvent(EventTIME)
		handler.Execute(event2)

		// Wait counter should be 1
		if handler.WaitCounter != 1 {
			t.Errorf("expected wait counter to be 1, got %d", handler.WaitCounter)
		}

		// y should still not be assigned
		y, _ := vm.globalScope.Get("y")
		if y != nil && y != int64(0) {
			t.Errorf("expected y to be nil or 0, got %v", y)
		}
	})

	t.Run("OpWait resumes execution when wait counter reaches 0", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(2)}}, // Wait for 2 events
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// First event: execute until OpWait
		event1 := NewEvent(EventTIME)
		handler.Execute(event1)

		// Second event: decrement wait counter
		event2 := NewEvent(EventTIME)
		handler.Execute(event2)

		// Third event: wait counter reaches 0, resume execution
		event3 := NewEvent(EventTIME)
		handler.Execute(event3)

		// y should now be assigned
		y, _ := vm.globalScope.Get("y")
		if y != int64(2) {
			t.Errorf("expected y to be 2, got %v", y)
		}

		// Wait counter should be 0
		if handler.WaitCounter != 0 {
			t.Errorf("expected wait counter to be 0, got %d", handler.WaitCounter)
		}

		// CurrentPC should be reset to 0 (handler completed)
		if handler.CurrentPC != 0 {
			t.Errorf("expected CurrentPC to be 0, got %d", handler.CurrentPC)
		}
	})

	t.Run("OpWait with count 0 continues immediately", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with OpWait(0)
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(0)}}, // Wait for 0 events (immediate)
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Execute the handler
		event := NewEvent(EventTIME)
		handler.Execute(event)

		// Both x and y should be assigned (no waiting)
		x, _ := vm.globalScope.Get("x")
		if x != int64(1) {
			t.Errorf("expected x to be 1, got %v", x)
		}

		y, _ := vm.globalScope.Get("y")
		if y != int64(2) {
			t.Errorf("expected y to be 2, got %v", y)
		}
	})

	t.Run("OpWait with negative count continues immediately", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with OpWait(-1)
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(-1)}}, // Negative count (immediate)
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Execute the handler
		event := NewEvent(EventTIME)
		handler.Execute(event)

		// Both x and y should be assigned (no waiting)
		x, _ := vm.globalScope.Get("x")
		if x != int64(1) {
			t.Errorf("expected x to be 1, got %v", x)
		}

		y, _ := vm.globalScope.Get("y")
		if y != int64(2) {
			t.Errorf("expected y to be 2, got %v", y)
		}
	})

	t.Run("multiple OpWait in sequence", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with multiple OpWait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(1)}}, // Wait for 1 event
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(2)}},
			{Cmd: opcode.Wait, Args: []any{int64(1)}}, // Wait for 1 event
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(3)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Event 1: execute until first OpWait
		event1 := NewEvent(EventTIME)
		handler.Execute(event1)

		step, _ := vm.globalScope.Get("step")
		if step != int64(1) {
			t.Errorf("after event 1: expected step to be 1, got %v", step)
		}

		// Event 2: resume, execute until second OpWait
		event2 := NewEvent(EventTIME)
		handler.Execute(event2)

		step, _ = vm.globalScope.Get("step")
		if step != int64(2) {
			t.Errorf("after event 2: expected step to be 2, got %v", step)
		}

		// Event 3: resume, complete handler
		event3 := NewEvent(EventTIME)
		handler.Execute(event3)

		step, _ = vm.globalScope.Get("step")
		if step != int64(3) {
			t.Errorf("after event 3: expected step to be 3, got %v", step)
		}
	})

	t.Run("OpWait outside handler is ignored", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Execute OpWait directly (not in a handler)
		opcode := opcode.OpCode{Cmd: opcode.Wait, Args: []any{int64(5)}}
		result, err := vm.Execute(opcode)

		// Should not return an error
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return nil (not a wait marker)
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

// TestEventHandlerPauseResume tests the pause and resume functionality of EventHandler.
// Requirement 6.2: When OpWait is executed, system pauses execution until next event.
// Requirement 6.3: When event occurs during step execution, system proceeds to next step.
func TestEventHandlerPauseResume(t *testing.T) {
	t.Run("handler resets PC after completion", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a simple handler without OpWait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Execute the handler
		event := NewEvent(EventTIME)
		handler.Execute(event)

		// PC should be reset to 0
		if handler.CurrentPC != 0 {
			t.Errorf("expected CurrentPC to be 0, got %d", handler.CurrentPC)
		}

		// Execute again - should work the same
		handler.Execute(event)

		x, _ := vm.globalScope.Get("x")
		if x != int64(1) {
			t.Errorf("expected x to be 1, got %v", x)
		}
	})

	t.Run("handler preserves PC during wait", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
			{Cmd: opcode.Wait, Args: []any{int64(3)}}, // Wait for 3 events
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Execute until OpWait
		event := NewEvent(EventTIME)
		handler.Execute(event)

		// PC should be at position 2 (after OpWait)
		if handler.CurrentPC != 2 {
			t.Errorf("expected CurrentPC to be 2, got %d", handler.CurrentPC)
		}

		// Execute again (still waiting)
		handler.Execute(event)

		// PC should still be at position 2
		if handler.CurrentPC != 2 {
			t.Errorf("expected CurrentPC to still be 2, got %d", handler.CurrentPC)
		}
	})
}

// TestVMBuiltinEndStep tests the end_step built-in function.
// Requirement 6.7: When end_step is called, system terminates step block execution.
// Requirement 10.6: When end_step is called, system terminates current step block.
func TestVMBuiltinEndStep(t *testing.T) {
	t.Run("end_step is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify end_step is registered
		if _, ok := vm.builtins["end_step"]; !ok {
			t.Error("expected end_step to be registered as built-in function")
		}
	})

	t.Run("end_step resets handler step counter", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler with step counter
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{}, vm, nil)
		handler.StepCounter = 10
		handler.WaitCounter = 5
		handler.CurrentPC = 3

		vm.SetCurrentHandler(handler)

		// Call end_step
		fn := vm.builtins["end_step"]
		result, err := fn(vm, []any{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Check that counters are reset
		if handler.StepCounter != 0 {
			t.Errorf("expected StepCounter to be 0, got %d", handler.StepCounter)
		}
		if handler.WaitCounter != 0 {
			t.Errorf("expected WaitCounter to be 0, got %d", handler.WaitCounter)
		}
		// CurrentPC should be set to end of handler
		if handler.CurrentPC != 0 {
			t.Errorf("expected CurrentPC to be 0 (end of empty handler), got %d", handler.CurrentPC)
		}
	})

	t.Run("end_step resets VM step counter when no handler", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.SetStepCounter(10)

		// Call end_step without a handler
		fn := vm.builtins["end_step"]
		result, err := fn(vm, []any{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Check that VM step counter is reset
		if vm.GetStepCounter() != 0 {
			t.Errorf("expected step counter to be 0, got %d", vm.GetStepCounter())
		}
	})

	t.Run("end_step stops handler execution", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler that calls end_step and then tries to assign a variable
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("before_end"), int64(1)}},
			{Cmd: opcode.Call, Args: []any{"end_step", []any{}}},
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("after_end"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Execute the handler
		event := NewEvent(EventTIME)
		handler.Execute(event)

		// before_end should be assigned
		before, _ := vm.globalScope.Get("before_end")
		if before != int64(1) {
			t.Errorf("expected before_end to be 1, got %v", before)
		}

		// after_end should NOT be assigned (end_step stops execution)
		after, exists := vm.globalScope.Get("after_end")
		if exists && after != nil && after != int64(0) {
			t.Errorf("expected after_end to not be set, got %v", after)
		}
	})
}

// TestVMBuiltinWait tests the Wait built-in function.
// Requirement 17.1: When Wait(n) is called, system pauses execution for n events.
// Requirement 17.2: When Wait(n) is called in mes(TIME) handler, system waits for n TIME events.
// Requirement 17.3: When Wait(n) is called in mes(MIDI_TIME) handler, system waits for n MIDI_TIME events.
// Requirement 17.4: When Wait(0) is called, system continues execution immediately.
// Requirement 17.5: When Wait(n) is called with n<0, system treats it as Wait(0).
// Requirement 17.6: System maintains separate wait counter for each handler.
func TestVMBuiltinWait(t *testing.T) {
	t.Run("Wait is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify Wait is registered
		if _, ok := vm.builtins["Wait"]; !ok {
			t.Error("expected Wait to be registered as built-in function")
		}
	})

	t.Run("Wait sets handler wait counter", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{}, vm, nil)
		vm.SetCurrentHandler(handler)

		// Call Wait(5)
		fn := vm.builtins["Wait"]
		result, err := fn(vm, []any{int64(5)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return a wait marker
		if _, ok := result.(*waitMarker); !ok {
			t.Errorf("expected waitMarker, got %T", result)
		}

		// Check that wait counter is set
		if handler.WaitCounter != 5 {
			t.Errorf("expected WaitCounter to be 5, got %d", handler.WaitCounter)
		}
	})

	t.Run("Wait(0) continues immediately", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{}, vm, nil)
		vm.SetCurrentHandler(handler)

		// Call Wait(0)
		fn := vm.builtins["Wait"]
		result, err := fn(vm, []any{int64(0)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return nil (no wait)
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Wait counter should not be set
		if handler.WaitCounter != 0 {
			t.Errorf("expected WaitCounter to be 0, got %d", handler.WaitCounter)
		}
	})

	t.Run("Wait with negative count continues immediately", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{}, vm, nil)
		vm.SetCurrentHandler(handler)

		// Call Wait(-5)
		fn := vm.builtins["Wait"]
		result, err := fn(vm, []any{int64(-5)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return nil (no wait)
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Wait counter should not be set
		if handler.WaitCounter != 0 {
			t.Errorf("expected WaitCounter to be 0, got %d", handler.WaitCounter)
		}
	})

	t.Run("Wait outside handler is ignored", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Call Wait without a handler
		fn := vm.builtins["Wait"]
		result, err := fn(vm, []any{int64(5)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return nil
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("Wait with float argument", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{}, vm, nil)
		vm.SetCurrentHandler(handler)

		// Call Wait(3.7) - should truncate to 3
		fn := vm.builtins["Wait"]
		result, err := fn(vm, []any{float64(3.7)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return a wait marker
		if _, ok := result.(*waitMarker); !ok {
			t.Errorf("expected waitMarker, got %T", result)
		}

		// Check that wait counter is set to truncated value
		if handler.WaitCounter != 3 {
			t.Errorf("expected WaitCounter to be 3, got %d", handler.WaitCounter)
		}
	})

	t.Run("Wait with no arguments defaults to 1", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{}, vm, nil)
		vm.SetCurrentHandler(handler)

		// Call Wait() with no arguments
		fn := vm.builtins["Wait"]
		result, err := fn(vm, []any{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		// Should return a wait marker
		if _, ok := result.(*waitMarker); !ok {
			t.Errorf("expected waitMarker, got %T", result)
		}

		// Check that wait counter defaults to 1
		if handler.WaitCounter != 1 {
			t.Errorf("expected WaitCounter to be 1, got %d", handler.WaitCounter)
		}
	})

	t.Run("Wait in TIME handler waits for TIME events", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a TIME handler that uses Wait
		handler := NewEventHandler("test-handler", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(1)}},
			{Cmd: opcode.Call, Args: []any{"Wait", int64(2)}},
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("step"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Event 1: execute until Wait
		event1 := NewEvent(EventTIME)
		handler.Execute(event1)

		step, _ := vm.globalScope.Get("step")
		if step != int64(1) {
			t.Errorf("after event 1: expected step to be 1, got %v", step)
		}

		// Event 2: decrement wait counter
		event2 := NewEvent(EventTIME)
		handler.Execute(event2)

		step, _ = vm.globalScope.Get("step")
		if step != int64(1) {
			t.Errorf("after event 2: expected step to still be 1, got %v", step)
		}

		// Event 3: wait counter reaches 0, resume
		event3 := NewEvent(EventTIME)
		handler.Execute(event3)

		step, _ = vm.globalScope.Get("step")
		if step != int64(2) {
			t.Errorf("after event 3: expected step to be 2, got %v", step)
		}
	})

	t.Run("Wait in MIDI_TIME handler waits for MIDI_TIME events", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create a MIDI_TIME handler that uses Wait
		handler := NewEventHandler("test-handler", EventMIDI_TIME, []opcode.OpCode{
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("midi_step"), int64(1)}},
			{Cmd: opcode.Call, Args: []any{"Wait", int64(2)}},
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("midi_step"), int64(2)}},
		}, vm, nil)

		vm.handlerRegistry.Register(handler)

		// Event 1: execute until Wait
		event1 := NewEvent(EventMIDI_TIME)
		handler.Execute(event1)

		step, _ := vm.globalScope.Get("midi_step")
		if step != int64(1) {
			t.Errorf("after event 1: expected midi_step to be 1, got %v", step)
		}

		// Event 2: decrement wait counter
		event2 := NewEvent(EventMIDI_TIME)
		handler.Execute(event2)

		// Event 3: wait counter reaches 0, resume
		event3 := NewEvent(EventMIDI_TIME)
		handler.Execute(event3)

		step, _ = vm.globalScope.Get("midi_step")
		if step != int64(2) {
			t.Errorf("after event 3: expected midi_step to be 2, got %v", step)
		}
	})

	t.Run("multiple handlers have separate wait counters", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create first handler with Wait(3)
		handler1 := NewEventHandler("handler1", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Call, Args: []any{"Wait", int64(3)}},
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("handler1_done"), int64(1)}},
		}, vm, nil)
		vm.handlerRegistry.Register(handler1)

		// Create second handler with Wait(1)
		handler2 := NewEventHandler("handler2", EventTIME, []opcode.OpCode{
			{Cmd: opcode.Call, Args: []any{"Wait", int64(1)}},
			{Cmd: opcode.Assign, Args: []any{opcode.Variable("handler2_done"), int64(1)}},
		}, vm, nil)
		vm.handlerRegistry.Register(handler2)

		// Event 1: both handlers start waiting
		event1 := NewEvent(EventTIME)
		vm.eventDispatcher.Dispatch(event1)

		// handler1 should have WaitCounter = 3
		if handler1.WaitCounter != 3 {
			t.Errorf("expected handler1 WaitCounter to be 3, got %d", handler1.WaitCounter)
		}
		// handler2 should have WaitCounter = 1
		if handler2.WaitCounter != 1 {
			t.Errorf("expected handler2 WaitCounter to be 1, got %d", handler2.WaitCounter)
		}

		// Event 2: handler2 completes, handler1 still waiting
		event2 := NewEvent(EventTIME)
		vm.eventDispatcher.Dispatch(event2)

		// handler2 should be done
		h2done, _ := vm.globalScope.Get("handler2_done")
		if h2done != int64(1) {
			t.Errorf("expected handler2_done to be 1, got %v", h2done)
		}

		// handler1 should still be waiting
		h1done, _ := vm.globalScope.Get("handler1_done")
		if h1done != nil && h1done != int64(0) {
			t.Errorf("expected handler1_done to not be set, got %v", h1done)
		}

		// Events 3 and 4: handler1 completes
		event3 := NewEvent(EventTIME)
		vm.eventDispatcher.Dispatch(event3)
		event4 := NewEvent(EventTIME)
		vm.eventDispatcher.Dispatch(event4)

		// handler1 should now be done
		h1done, _ = vm.globalScope.Get("handler1_done")
		if h1done != int64(1) {
			t.Errorf("expected handler1_done to be 1, got %v", h1done)
		}
	})
}

// TestVMBuiltinExitTitle tests the ExitTitle built-in function.
// Requirement 10.7: When ExitTitle is called, system terminates program.
// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
// Requirement 15.4: When ExitTitle is called, system terminates event loop.
// Requirement 15.7: System provides graceful shutdown mechanism.
func TestVMBuiltinExitTitle(t *testing.T) {
	t.Run("ExitTitle is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify ExitTitle is registered
		if _, ok := vm.builtins["ExitTitle"]; !ok {
			t.Error("expected ExitTitle to be registered as built-in function")
		}
	})

	t.Run("ExitTitle removes all handlers", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Register some handlers
		handler1 := NewEventHandler("handler1", EventTIME, []opcode.OpCode{}, vm, nil)
		handler2 := NewEventHandler("handler2", EventMIDI_TIME, []opcode.OpCode{}, vm, nil)
		vm.handlerRegistry.Register(handler1)
		vm.handlerRegistry.Register(handler2)

		// Verify handlers are registered
		if vm.handlerRegistry.Count() != 2 {
			t.Errorf("expected 2 handlers, got %d", vm.handlerRegistry.Count())
		}

		// Call ExitTitle
		fn := vm.builtins["ExitTitle"]
		result, err := fn(vm, []any{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// All handlers should be removed
		if vm.handlerRegistry.Count() != 0 {
			t.Errorf("expected 0 handlers after ExitTitle, got %d", vm.handlerRegistry.Count())
		}
	})

	t.Run("ExitTitle stops VM", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Start VM in a goroutine
		done := make(chan error)
		go func() {
			done <- vm.Run()
		}()

		// Wait a bit for VM to start
		time.Sleep(10 * time.Millisecond)

		// Call ExitTitle
		fn := vm.builtins["ExitTitle"]
		fn(vm, []any{})

		// VM should stop
		select {
		case err := <-done:
			if err != nil {
				t.Errorf("expected no error, got %v", err)
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("VM did not stop after ExitTitle")
		}
	})

	t.Run("ExitTitle handles nil audio system gracefully", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Ensure audio system is nil
		vm.audioSystem = nil

		// Call ExitTitle - should not panic
		fn := vm.builtins["ExitTitle"]
		result, err := fn(vm, []any{})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})
}

// TestVMBuiltinStrPrint tests the StrPrint built-in function.
// Validates: Requirements 1.1-1.8
func TestVMBuiltinStrPrint(t *testing.T) {
	t.Run("StrPrint is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify StrPrint is registered
		if _, ok := vm.builtins["StrPrint"]; !ok {
			t.Error("expected StrPrint to be registered as built-in function")
		}
	})

	// Requirement 1.1: Basic format string with arguments
	t.Run("basic format string with arguments", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Hello %s", "World"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Hello World" {
			t.Errorf("expected 'Hello World', got %v", result)
		}
	})

	// Requirement 1.2: %ld format specifier for decimal integers
	t.Run("%ld format specifier for decimal integers", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Number: %ld", int64(42)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Number: 42" {
			t.Errorf("expected 'Number: 42', got %v", result)
		}
	})

	t.Run("%ld with negative number", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Value: %ld", int64(-123)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Value: -123" {
			t.Errorf("expected 'Value: -123', got %v", result)
		}
	})

	// Requirement 1.3: %lx format specifier for hexadecimal
	t.Run("%lx format specifier for hexadecimal", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Hex: %lx", int64(255)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Hex: ff" {
			t.Errorf("expected 'Hex: ff', got %v", result)
		}
	})

	t.Run("%lx with zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Hex: %lx", int64(0)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Hex: 0" {
			t.Errorf("expected 'Hex: 0', got %v", result)
		}
	})

	// Requirement 1.4: %s format specifier for strings
	t.Run("%s format specifier for strings", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Name: %s", "Alice"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Name: Alice" {
			t.Errorf("expected 'Name: Alice', got %v", result)
		}
	})

	t.Run("%s with empty string", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Value: [%s]", ""})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Value: []" {
			t.Errorf("expected 'Value: []', got %v", result)
		}
	})

	// Requirement 1.5: Width and padding specifiers
	t.Run("%03d width and padding specifier", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"ROBOT%03d.BMP", int64(1)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "ROBOT001.BMP" {
			t.Errorf("expected 'ROBOT001.BMP', got %v", result)
		}
	})

	t.Run("%05ld width and padding with %ld", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"ID: %05ld", int64(42)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "ID: 00042" {
			t.Errorf("expected 'ID: 00042', got %v", result)
		}
	})

	t.Run("%08lx width and padding with %lx", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Addr: %08lx", int64(4096)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Addr: 00001000" {
			t.Errorf("expected 'Addr: 00001000', got %v", result)
		}
	})

	// Requirement 1.6: Escape sequences
	t.Run("\\n escape sequence", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Line1\\nLine2"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Line1\nLine2" {
			t.Errorf("expected 'Line1\\nLine2', got %v", result)
		}
	})

	t.Run("\\t escape sequence", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Col1\\tCol2"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Col1\tCol2" {
			t.Errorf("expected 'Col1\\tCol2', got %v", result)
		}
	})

	t.Run("\\r escape sequence", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Start\\rEnd"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Start\rEnd" {
			t.Errorf("expected 'Start\\rEnd', got %v", result)
		}
	})

	t.Run("multiple escape sequences", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"A\\nB\\tC\\rD"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "A\nB\tC\rD" {
			t.Errorf("expected 'A\\nB\\tC\\rD', got %v", result)
		}
	})

	// Requirement 1.7: Fewer arguments than format specifiers
	t.Run("fewer arguments than format specifiers", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		// Should not crash, Go's fmt.Sprintf handles this gracefully
		result, err := fn(vm, []any{"Value: %s %d"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// Result should contain MISSING markers from Go's fmt.Sprintf
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	t.Run("one argument missing", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"%s and %s", "first"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// Should contain the first argument and a MISSING marker for the second
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	// Requirement 1.8: More arguments than format specifiers
	t.Run("more arguments than format specifiers", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Value: %s", "used", "ignored1", "ignored2"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		// Extra arguments should be appended by Go's fmt.Sprintf
		// The result will contain the extra arguments
		if result == "" {
			t.Error("expected non-empty result")
		}
	})

	// Edge cases
	t.Run("empty format string", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{""})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got %v", result)
		}
	})

	t.Run("no format specifiers", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"Hello World"})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Hello World" {
			t.Errorf("expected 'Hello World', got %v", result)
		}
	})

	t.Run("no arguments", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string, got %v", result)
		}
	})

	t.Run("non-string format argument", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{123})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "" {
			t.Errorf("expected empty string for non-string format, got %v", result)
		}
	})

	// Multiple format specifiers
	t.Run("multiple format specifiers", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		result, err := fn(vm, []any{"%s: %ld (0x%lx)", "Value", int64(255), int64(255)})
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != "Value: 255 (0xff)" {
			t.Errorf("expected 'Value: 255 (0xff)', got %v", result)
		}
	})

	// Real-world use case from ROBOT sample
	t.Run("ROBOT sample use case", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		fn := vm.builtins["StrPrint"]

		// Test the actual use case from ROBOT.TFY: StrPrint("ROBOT%03d.BMP", i)
		for i := int64(0); i <= 1; i++ {
			result, err := fn(vm, []any{"ROBOT%03d.BMP", i})
			if err != nil {
				t.Errorf("expected no error for i=%d, got %v", i, err)
			}
			expected := "ROBOT00" + string('0'+byte(i)) + ".BMP"
			if result != expected {
				t.Errorf("for i=%d: expected '%s', got %v", i, expected, result)
			}
		}
	})
}

// mockGraphicsSystem is a mock implementation of GraphicsSystemInterface for testing.
// This mock is used to test CreatePic and CapTitle built-in functions without requiring a real graphics system.
type mockGraphicsSystem struct {
	pictures       map[int]*mockPicture
	nextPicID      int
	createPicErr   error // Error to return from CreatePic
	windows        map[int]*mockWindow
	nextWinID      int
	capTitleAllCnt int    // Count of CapTitleAll calls
	lastCapTitle   string // Last title set by CapTitleAll
}

type mockPicture struct {
	id     int
	width  int
	height int
}

type mockWindow struct {
	id      int
	picID   int
	caption string
}

func newMockGraphicsSystem() *mockGraphicsSystem {
	return &mockGraphicsSystem{
		pictures:  make(map[int]*mockPicture),
		nextPicID: 0,
		windows:   make(map[int]*mockWindow),
		nextWinID: 0,
	}
}

func (m *mockGraphicsSystem) LoadPic(filename string) (int, error) {
	return -1, fmt.Errorf("not implemented")
}

func (m *mockGraphicsSystem) CreatePic(width, height int) (int, error) {
	if m.createPicErr != nil {
		return -1, m.createPicErr
	}
	if width <= 0 || height <= 0 {
		return -1, fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}
	id := m.nextPicID
	m.nextPicID++
	m.pictures[id] = &mockPicture{id: id, width: width, height: height}
	return id, nil
}

func (m *mockGraphicsSystem) CreatePicFrom(srcID int) (int, error) {
	if m.createPicErr != nil {
		return -1, m.createPicErr
	}
	src, ok := m.pictures[srcID]
	if !ok {
		return -1, fmt.Errorf("source picture %d not found", srcID)
	}
	id := m.nextPicID
	m.nextPicID++
	m.pictures[id] = &mockPicture{id: id, width: src.width, height: src.height}
	return id, nil
}

func (m *mockGraphicsSystem) CreatePicWithSize(srcID, width, height int) (int, error) {
	if m.createPicErr != nil {
		return -1, m.createPicErr
	}
	// Check if source picture exists (要件 2.3)
	if _, ok := m.pictures[srcID]; !ok {
		return -1, fmt.Errorf("source picture %d not found", srcID)
	}
	// Validate dimensions (要件 2.4)
	if width <= 0 || height <= 0 {
		return -1, fmt.Errorf("invalid dimensions: width=%d, height=%d", width, height)
	}
	id := m.nextPicID
	m.nextPicID++
	// Create empty picture with specified size (要件 2.1, 2.2)
	m.pictures[id] = &mockPicture{id: id, width: width, height: height}
	return id, nil
}

func (m *mockGraphicsSystem) DelPic(id int) error {
	if _, ok := m.pictures[id]; !ok {
		return fmt.Errorf("picture %d not found", id)
	}
	delete(m.pictures, id)
	return nil
}

func (m *mockGraphicsSystem) PicWidth(id int) int {
	if pic, ok := m.pictures[id]; ok {
		return pic.width
	}
	return 0
}

func (m *mockGraphicsSystem) PicHeight(id int) int {
	if pic, ok := m.pictures[id]; ok {
		return pic.height
	}
	return 0
}

func (m *mockGraphicsSystem) MovePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode int) error {
	return nil
}

func (m *mockGraphicsSystem) MovePicWithSpeed(srcID, srcX, srcY, width, height, dstID, dstX, dstY, mode, speed int) error {
	return nil
}

func (m *mockGraphicsSystem) MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH int) error {
	return nil
}

func (m *mockGraphicsSystem) TransPic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int, transColor any) error {
	return nil
}

func (m *mockGraphicsSystem) ReversePic(srcID, srcX, srcY, width, height, dstID, dstX, dstY int) error {
	return nil
}

func (m *mockGraphicsSystem) OpenWin(picID int, opts ...any) (int, error) {
	id := m.nextWinID
	m.nextWinID++
	m.windows[id] = &mockWindow{id: id, picID: picID, caption: ""}
	return id, nil
}

func (m *mockGraphicsSystem) MoveWin(id int, opts ...any) error {
	return nil
}

func (m *mockGraphicsSystem) CloseWin(id int) error {
	delete(m.windows, id)
	return nil
}

func (m *mockGraphicsSystem) CloseWinAll() {
	m.windows = make(map[int]*mockWindow)
}

func (m *mockGraphicsSystem) CapTitle(id int, title string) error {
	if win, ok := m.windows[id]; ok {
		win.caption = title
	}
	// 存在しないウィンドウIDでもエラーを返さない (要件 3.4)
	return nil
}

func (m *mockGraphicsSystem) CapTitleAll(title string) {
	m.capTitleAllCnt++
	m.lastCapTitle = title
	for _, win := range m.windows {
		win.caption = title
	}
}

func (m *mockGraphicsSystem) GetPicNo(id int) (int, error) {
	return 0, nil
}

func (m *mockGraphicsSystem) GetWinByPicID(picID int) (int, error) {
	return 0, nil
}

func (m *mockGraphicsSystem) PutCast(winID, picID, x, y, srcX, srcY, w, h int) (int, error) {
	return 0, nil
}

func (m *mockGraphicsSystem) PutCastWithTransColor(winID, picID, x, y, srcX, srcY, w, h int, transColor color.Color) (int, error) {
	return 0, nil
}

func (m *mockGraphicsSystem) MoveCast(id int, opts ...any) error {
	return nil
}

func (m *mockGraphicsSystem) MoveCastWithOptions(id int, opts ...graphics.CastOption) error {
	return nil
}

func (m *mockGraphicsSystem) DelCast(id int) error {
	return nil
}

func (m *mockGraphicsSystem) TextWrite(picID, x, y int, text string) error {
	return nil
}

func (m *mockGraphicsSystem) SetFont(name string, size int, opts ...any) error {
	return nil
}

func (m *mockGraphicsSystem) SetTextColor(c any) error {
	return nil
}

func (m *mockGraphicsSystem) SetBgColor(c any) error {
	return nil
}

func (m *mockGraphicsSystem) SetBackMode(mode int) error {
	return nil
}

func (m *mockGraphicsSystem) DrawLine(picID, x1, y1, x2, y2 int) error {
	return nil
}

func (m *mockGraphicsSystem) DrawRect(picID, x1, y1, x2, y2, fillMode int) error {
	return nil
}

func (m *mockGraphicsSystem) FillRect(picID, x1, y1, x2, y2 int, c any) error {
	return nil
}

func (m *mockGraphicsSystem) DrawCircle(picID, x, y, radius, fillMode int) error {
	return nil
}

func (m *mockGraphicsSystem) SetLineSize(size int) {}

func (m *mockGraphicsSystem) SetPaintColor(c any) error {
	return nil
}

func (m *mockGraphicsSystem) GetColor(picID, x, y int) (int, error) {
	return 0, nil
}

func (m *mockGraphicsSystem) GetVirtualWidth() int {
	return 800
}

func (m *mockGraphicsSystem) GetVirtualHeight() int {
	return 600
}

// TestVMBuiltinCreatePic tests the CreatePic built-in function.
// Validates: Requirements 2.1-2.5
func TestVMBuiltinCreatePic(t *testing.T) {
	t.Run("CreatePic is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify CreatePic is registered
		if _, ok := vm.builtins["CreatePic"]; !ok {
			t.Error("expected CreatePic to be registered as built-in function")
		}
	})

	// Test 1: CreatePic with 3 arguments creates picture with correct size (要件 2.1)
	t.Run("CreatePic with 3 arguments creates picture with correct size", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID), int64(200), int64(150)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		newID, ok := result.(int)
		if !ok {
			t.Fatalf("expected int result, got %T", result)
		}

		// Verify new picture has correct dimensions
		if mockGS.PicWidth(newID) != 200 {
			t.Errorf("expected width 200, got %d", mockGS.PicWidth(newID))
		}
		if mockGS.PicHeight(newID) != 150 {
			t.Errorf("expected height 150, got %d", mockGS.PicHeight(newID))
		}
	})

	// Test 2: CreatePic with 3 arguments creates empty picture (要件 2.2)
	// Note: The mock doesn't actually copy content, so we verify the new picture
	// has different dimensions from the source to confirm it's not a copy
	t.Run("CreatePic with 3 arguments creates empty picture (not copying source)", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture with specific size
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID), int64(50), int64(75)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		newID, ok := result.(int)
		if !ok {
			t.Fatalf("expected int result, got %T", result)
		}

		// Verify new picture has different dimensions from source
		// (confirming it's not a copy of the source)
		if mockGS.PicWidth(newID) == mockGS.PicWidth(srcID) && mockGS.PicHeight(newID) == mockGS.PicHeight(srcID) {
			t.Error("new picture should have different dimensions from source (not a copy)")
		}

		// Verify new picture has the specified dimensions
		if mockGS.PicWidth(newID) != 50 {
			t.Errorf("expected width 50, got %d", mockGS.PicWidth(newID))
		}
		if mockGS.PicHeight(newID) != 75 {
			t.Errorf("expected height 75, got %d", mockGS.PicHeight(newID))
		}
	})

	// Test 3: CreatePic with non-existent source ID returns error (要件 2.3)
	t.Run("CreatePic with non-existent source ID returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(999), int64(100), int64(100)})

		// Should return error
		if err == nil {
			t.Error("expected error for non-existent source ID, got nil")
		}

		// Result should be -1
		if result != -1 {
			t.Errorf("expected result -1, got %v", result)
		}
	})

	// Test 4: CreatePic with zero width returns error (要件 2.4)
	t.Run("CreatePic with zero width returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID), int64(0), int64(100)})

		// Should return error
		if err == nil {
			t.Error("expected error for zero width, got nil")
		}

		// Result should be -1
		if result != -1 {
			t.Errorf("expected result -1, got %v", result)
		}
	})

	// Test 4b: CreatePic with zero height returns error (要件 2.4)
	t.Run("CreatePic with zero height returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID), int64(100), int64(0)})

		// Should return error
		if err == nil {
			t.Error("expected error for zero height, got nil")
		}

		// Result should be -1
		if result != -1 {
			t.Errorf("expected result -1, got %v", result)
		}
	})

	// Test 4c: CreatePic with negative width returns error (要件 2.4)
	t.Run("CreatePic with negative width returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID), int64(-10), int64(100)})

		// Should return error
		if err == nil {
			t.Error("expected error for negative width, got nil")
		}

		// Result should be -1
		if result != -1 {
			t.Errorf("expected result -1, got %v", result)
		}
	})

	// Test 4d: CreatePic with negative height returns error (要件 2.4)
	t.Run("CreatePic with negative height returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID), int64(100), int64(-10)})

		// Should return error
		if err == nil {
			t.Error("expected error for negative height, got nil")
		}

		// Result should be -1
		if result != -1 {
			t.Errorf("expected result -1, got %v", result)
		}
	})

	// Test 5: Backward compatibility - 1-argument pattern still works (要件 2.5)
	t.Run("backward compatibility: 1-argument pattern (CreatePicFrom)", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(srcID)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		newID, ok := result.(int)
		if !ok {
			t.Fatalf("expected int result, got %T", result)
		}

		// Verify new picture has same dimensions as source (copy behavior)
		if mockGS.PicWidth(newID) != 100 {
			t.Errorf("expected width 100, got %d", mockGS.PicWidth(newID))
		}
		if mockGS.PicHeight(newID) != 100 {
			t.Errorf("expected height 100, got %d", mockGS.PicHeight(newID))
		}
	})

	// Test 6: Backward compatibility - 2-argument pattern still works (要件 2.5)
	t.Run("backward compatibility: 2-argument pattern (CreatePic width, height)", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(200), int64(150)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		newID, ok := result.(int)
		if !ok {
			t.Fatalf("expected int result, got %T", result)
		}

		// Verify new picture has specified dimensions
		if mockGS.PicWidth(newID) != 200 {
			t.Errorf("expected width 200, got %d", mockGS.PicWidth(newID))
		}
		if mockGS.PicHeight(newID) != 150 {
			t.Errorf("expected height 150, got %d", mockGS.PicHeight(newID))
		}
	})

	// Edge case: CreatePic without graphics system
	t.Run("CreatePic without graphics system returns -1", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Don't set graphics system

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{int64(100), int64(100)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		if result != -1 {
			t.Errorf("expected result -1 when no graphics system, got %v", result)
		}
	})

	// Edge case: CreatePic with no arguments
	t.Run("CreatePic with no arguments returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		fn := vm.builtins["CreatePic"]
		_, err := fn(vm, []any{})

		if err == nil {
			t.Error("expected error for no arguments, got nil")
		}
	})

	// Edge case: CreatePic with non-integer arguments
	t.Run("CreatePic with non-integer arguments returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{"not", "integers"})

		if err == nil {
			t.Error("expected error for non-integer arguments, got nil")
		}

		if result != -1 {
			t.Errorf("expected result -1, got %v", result)
		}
	})

	// Test with float arguments (should be converted to int)
	t.Run("CreatePic with float arguments", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{float64(100), float64(200)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		newID, ok := result.(int)
		if !ok {
			t.Fatalf("expected int result, got %T", result)
		}

		// Verify dimensions (float should be truncated to int)
		if mockGS.PicWidth(newID) != 100 {
			t.Errorf("expected width 100, got %d", mockGS.PicWidth(newID))
		}
		if mockGS.PicHeight(newID) != 200 {
			t.Errorf("expected height 200, got %d", mockGS.PicHeight(newID))
		}
	})

	// Test 3-argument pattern with float arguments
	t.Run("CreatePic 3-argument pattern with float arguments", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create source picture first
		srcID, _ := mockGS.CreatePic(100, 100)

		fn := vm.builtins["CreatePic"]
		result, err := fn(vm, []any{float64(srcID), float64(150), float64(200)})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		newID, ok := result.(int)
		if !ok {
			t.Fatalf("expected int result, got %T", result)
		}

		// Verify dimensions
		if mockGS.PicWidth(newID) != 150 {
			t.Errorf("expected width 150, got %d", mockGS.PicWidth(newID))
		}
		if mockGS.PicHeight(newID) != 200 {
			t.Errorf("expected height 200, got %d", mockGS.PicHeight(newID))
		}
	})
}

// TestVMBuiltinCapTitle tests the CapTitle built-in function.
// Validates: Requirements 3.1-3.5
func TestVMBuiltinCapTitle(t *testing.T) {
	t.Run("CapTitle is registered as built-in", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Verify CapTitle is registered
		if _, ok := vm.builtins["CapTitle"]; !ok {
			t.Error("expected CapTitle to be registered as built-in function")
		}
	})

	// Test case 1: CapTitle with 1 argument sets caption for ALL windows (requirement 3.1)
	t.Run("CapTitle with 1 argument sets caption for ALL windows", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create multiple windows
		mockGS.OpenWin(0)
		mockGS.OpenWin(1)
		mockGS.OpenWin(2)

		fn := vm.builtins["CapTitle"]
		result, err := fn(vm, []any{"Test Title"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify all windows have the same caption
		for id, win := range mockGS.windows {
			if win.caption != "Test Title" {
				t.Errorf("window %d: expected caption 'Test Title', got '%s'", id, win.caption)
			}
		}

		// Verify CapTitleAll was called
		if mockGS.capTitleAllCnt != 1 {
			t.Errorf("expected CapTitleAll to be called once, got %d", mockGS.capTitleAllCnt)
		}
	})

	// Test case 2: CapTitle with 1 argument when no windows exist (requirement 3.2)
	t.Run("CapTitle with 1 argument when no windows exist", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// No windows created

		fn := vm.builtins["CapTitle"]
		result, err := fn(vm, []any{"Test Title"})

		// Should not return error (requirement 3.2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify CapTitleAll was still called
		if mockGS.capTitleAllCnt != 1 {
			t.Errorf("expected CapTitleAll to be called once, got %d", mockGS.capTitleAllCnt)
		}
	})

	// Test case 3: CapTitle with 2 arguments sets caption for specific window only (requirement 3.3)
	t.Run("CapTitle with 2 arguments sets caption for specific window only", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create multiple windows
		mockGS.OpenWin(0) // winID 0
		mockGS.OpenWin(1) // winID 1
		mockGS.OpenWin(2) // winID 2

		// Set initial captions
		mockGS.windows[0].caption = "Original 0"
		mockGS.windows[1].caption = "Original 1"
		mockGS.windows[2].caption = "Original 2"

		fn := vm.builtins["CapTitle"]
		result, err := fn(vm, []any{int64(1), "New Title"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify only window 1 has the new caption
		if mockGS.windows[0].caption != "Original 0" {
			t.Errorf("window 0: expected caption 'Original 0', got '%s'", mockGS.windows[0].caption)
		}
		if mockGS.windows[1].caption != "New Title" {
			t.Errorf("window 1: expected caption 'New Title', got '%s'", mockGS.windows[1].caption)
		}
		if mockGS.windows[2].caption != "Original 2" {
			t.Errorf("window 2: expected caption 'Original 2', got '%s'", mockGS.windows[2].caption)
		}

		// Verify CapTitleAll was NOT called
		if mockGS.capTitleAllCnt != 0 {
			t.Errorf("expected CapTitleAll to not be called, got %d", mockGS.capTitleAllCnt)
		}
	})

	// Test case 4: CapTitle with non-existent window ID does not error (requirement 3.4)
	t.Run("CapTitle with non-existent window ID does not error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create one window
		mockGS.OpenWin(0) // winID 0
		mockGS.windows[0].caption = "Original"

		fn := vm.builtins["CapTitle"]
		// Call with non-existent window ID 999
		result, err := fn(vm, []any{int64(999), "New Title"})

		// Should not return error (requirement 3.4)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify existing window caption is unchanged
		if mockGS.windows[0].caption != "Original" {
			t.Errorf("window 0: expected caption 'Original', got '%s'", mockGS.windows[0].caption)
		}
	})

	// Test case 5: CapTitle with empty string clears caption (requirement 3.5)
	t.Run("CapTitle with empty string clears caption", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create window with initial caption
		mockGS.OpenWin(0)
		mockGS.windows[0].caption = "Initial Title"

		fn := vm.builtins["CapTitle"]
		// Call with empty string (1 argument pattern)
		result, err := fn(vm, []any{""})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify caption is cleared
		if mockGS.windows[0].caption != "" {
			t.Errorf("window 0: expected empty caption, got '%s'", mockGS.windows[0].caption)
		}
	})

	// Test case 5b: CapTitle with empty string clears caption (2 argument pattern)
	t.Run("CapTitle with empty string clears caption (2 argument pattern)", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create window with initial caption
		mockGS.OpenWin(0)
		mockGS.windows[0].caption = "Initial Title"

		fn := vm.builtins["CapTitle"]
		// Call with empty string (2 argument pattern)
		result, err := fn(vm, []any{int64(0), ""})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify caption is cleared
		if mockGS.windows[0].caption != "" {
			t.Errorf("window 0: expected empty caption, got '%s'", mockGS.windows[0].caption)
		}
	})

	// Test case 6: CapTitle without graphics system returns nil (edge case)
	t.Run("CapTitle without graphics system returns nil", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// No graphics system set

		fn := vm.builtins["CapTitle"]
		result, err := fn(vm, []any{"Test Title"})

		// Should not return error
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	// Test case 7: CapTitle with no arguments returns error (edge case)
	t.Run("CapTitle with no arguments returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		fn := vm.builtins["CapTitle"]
		_, err := fn(vm, []any{})

		// Should return error
		if err == nil {
			t.Error("expected error for missing arguments")
		}
	})

	// Additional test: CapTitle with float window ID
	t.Run("CapTitle with float window ID", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		mockGS := newMockGraphicsSystem()
		vm.SetGraphicsSystem(mockGS)

		// Create windows
		mockGS.OpenWin(0) // winID 0
		mockGS.OpenWin(1) // winID 1
		mockGS.windows[0].caption = "Original 0"
		mockGS.windows[1].caption = "Original 1"

		fn := vm.builtins["CapTitle"]
		// Call with float window ID (should be truncated to int)
		result, err := fn(vm, []any{float64(1.9), "Float Title"})

		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}

		// Verify window 1 (truncated from 1.9) has the new caption
		if mockGS.windows[1].caption != "Float Title" {
			t.Errorf("window 1: expected caption 'Float Title', got '%s'", mockGS.windows[1].caption)
		}
		// Verify window 0 is unchanged
		if mockGS.windows[0].caption != "Original 0" {
			t.Errorf("window 0: expected caption 'Original 0', got '%s'", mockGS.windows[0].caption)
		}
	})
}
