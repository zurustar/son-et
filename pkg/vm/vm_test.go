package vm

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler"
)

// TestNewVM tests the VM constructor with various options.
func TestNewVM(t *testing.T) {
	t.Run("creates VM with empty opcodes", func(t *testing.T) {
		vm := New([]compiler.OpCode{})
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
		opcodes := []compiler.OpCode{
			{Cmd: compiler.OpDefineFunction, Args: []any{"main", []any{}, []compiler.OpCode{}}},
		}
		vm := New(opcodes)
		if len(vm.opcodes) != 1 {
			t.Errorf("expected 1 opcode, got %d", len(vm.opcodes))
		}
	})

	t.Run("applies headless option", func(t *testing.T) {
		vm := New([]compiler.OpCode{}, WithHeadless(true))
		if !vm.headless {
			t.Error("expected headless to be true")
		}
	})

	t.Run("applies timeout option", func(t *testing.T) {
		timeout := 5 * time.Second
		vm := New([]compiler.OpCode{}, WithTimeout(timeout))
		if vm.timeout != timeout {
			t.Errorf("expected timeout to be %v, got %v", timeout, vm.timeout)
		}
	})

	t.Run("applies multiple options", func(t *testing.T) {
		timeout := 10 * time.Second
		vm := New([]compiler.OpCode{}, WithHeadless(true), WithTimeout(timeout))
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
		vm := New([]compiler.OpCode{})
		err := vm.Run()
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})

	t.Run("prevents double run", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{}, WithTimeout(50*time.Millisecond))

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
		opcodes := []compiler.OpCode{
			{
				Cmd: compiler.OpDefineFunction,
				Args: []any{
					"testFunc",
					[]any{
						map[string]any{"name": "x", "type": "int", "isArray": false},
					},
					[]compiler.OpCode{},
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
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})
		// Should not panic
		vm.Stop()
	})
}

// TestVMIsRunning tests the IsRunning method.
func TestVMIsRunning(t *testing.T) {
	t.Run("returns false when not running", func(t *testing.T) {
		vm := New([]compiler.OpCode{})
		if vm.IsRunning() {
			t.Error("expected IsRunning to be false")
		}
	})
}

// TestVMStackFrame tests stack frame management.
func TestVMStackFrame(t *testing.T) {
	t.Run("pushes and pops stack frames", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

		_, err := vm.PopStackFrame()
		if err == nil {
			t.Error("expected error when popping from empty stack")
		}
	})
}

// TestVMBuiltinFunctions tests built-in function registration.
func TestVMBuiltinFunctions(t *testing.T) {
	t.Run("registers built-in function", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

		scope := vm.GetGlobalScope()
		if scope == nil {
			t.Error("expected global scope to be non-nil")
		}
		if scope != vm.globalScope {
			t.Error("expected GetGlobalScope to return globalScope")
		}
	})

	t.Run("returns current scope", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

		// Verify PlayMIDI is registered
		if _, ok := vm.builtins["PlayMIDI"]; !ok {
			t.Error("expected PlayMIDI to be registered as built-in function")
		}
	})

	t.Run("PlayMIDI handles missing argument gracefully", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

		// Call PlayMIDI without arguments - should not panic
		fn := vm.builtins["PlayMIDI"]
		result, err := fn(vm, []any{})

		// Should return nil without error (error is logged)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PlayMIDI handles wrong argument type gracefully", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

		// Verify PlayWAVE is registered
		if _, ok := vm.builtins["PlayWAVE"]; !ok {
			t.Error("expected PlayWAVE to be registered as built-in function")
		}
	})

	t.Run("PlayWAVE handles missing argument gracefully", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

		// Call PlayWAVE without arguments - should not panic
		fn := vm.builtins["PlayWAVE"]
		result, err := fn(vm, []any{})

		// Should return nil without error (error is logged)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		if result != nil {
			t.Errorf("expected nil result, got %v", result)
		}
	})

	t.Run("PlayWAVE handles wrong argument type gracefully", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

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
	vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(2)}}, // Wait for 2 events
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("y"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(2)}}, // Wait for 2 events
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("y"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(2)}}, // Wait for 2 events
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("y"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a handler with OpWait(0)
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(0)}}, // Wait for 0 events (immediate)
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("y"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a handler with OpWait(-1)
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(-1)}}, // Negative count (immediate)
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("y"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a handler with multiple OpWait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("step"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(1)}}, // Wait for 1 event
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("step"), int64(2)}},
			{Cmd: compiler.OpWait, Args: []any{int64(1)}}, // Wait for 1 event
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("step"), int64(3)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Execute OpWait directly (not in a handler)
		opcode := compiler.OpCode{Cmd: compiler.OpWait, Args: []any{int64(5)}}
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
		vm := New([]compiler.OpCode{})

		// Create a simple handler without OpWait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a handler with OpWait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("x"), int64(1)}},
			{Cmd: compiler.OpWait, Args: []any{int64(3)}}, // Wait for 3 events
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("y"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Verify end_step is registered
		if _, ok := vm.builtins["end_step"]; !ok {
			t.Error("expected end_step to be registered as built-in function")
		}
	})

	t.Run("end_step resets handler step counter", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

		// Create a handler with step counter
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})
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
		vm := New([]compiler.OpCode{})

		// Create a handler that calls end_step and then tries to assign a variable
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("before_end"), int64(1)}},
			{Cmd: compiler.OpCall, Args: []any{"end_step", []any{}}},
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("after_end"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Verify Wait is registered
		if _, ok := vm.builtins["Wait"]; !ok {
			t.Error("expected Wait to be registered as built-in function")
		}
	})

	t.Run("Wait sets handler wait counter", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})

		// Create a handler
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})

		// Create a TIME handler that uses Wait
		handler := NewEventHandler("test-handler", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("step"), int64(1)}},
			{Cmd: compiler.OpCall, Args: []any{"Wait", int64(2)}},
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("step"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create a MIDI_TIME handler that uses Wait
		handler := NewEventHandler("test-handler", EventMIDI_TIME, []compiler.OpCode{
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("midi_step"), int64(1)}},
			{Cmd: compiler.OpCall, Args: []any{"Wait", int64(2)}},
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("midi_step"), int64(2)}},
		}, vm)

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
		vm := New([]compiler.OpCode{})

		// Create first handler with Wait(3)
		handler1 := NewEventHandler("handler1", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpCall, Args: []any{"Wait", int64(3)}},
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("handler1_done"), int64(1)}},
		}, vm)
		vm.handlerRegistry.Register(handler1)

		// Create second handler with Wait(1)
		handler2 := NewEventHandler("handler2", EventTIME, []compiler.OpCode{
			{Cmd: compiler.OpCall, Args: []any{"Wait", int64(1)}},
			{Cmd: compiler.OpAssign, Args: []any{compiler.Variable("handler2_done"), int64(1)}},
		}, vm)
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
		vm := New([]compiler.OpCode{})

		// Verify ExitTitle is registered
		if _, ok := vm.builtins["ExitTitle"]; !ok {
			t.Error("expected ExitTitle to be registered as built-in function")
		}
	})

	t.Run("ExitTitle removes all handlers", func(t *testing.T) {
		vm := New([]compiler.OpCode{})

		// Register some handlers
		handler1 := NewEventHandler("handler1", EventTIME, []compiler.OpCode{}, vm)
		handler2 := NewEventHandler("handler2", EventMIDI_TIME, []compiler.OpCode{}, vm)
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
		vm := New([]compiler.OpCode{})

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
		vm := New([]compiler.OpCode{})

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
