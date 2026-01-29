package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// Property-based tests for stack frame management.
// These tests verify the correctness properties defined in the design document.

// TestProperty23_StackFrameRoundTrip tests that stack frames are correctly
// pushed when a function is called and popped when it returns.
// **Validates: Requirements 20.1, 20.2**
// Feature: execution-engine, Property 23: スタックフレームのラウンドトリップ
func TestProperty23_StackFrameRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Stack depth increases by 1 when pushing a frame
	properties.Property("stack depth increases by 1 when pushing a frame", prop.ForAll(
		func(functionName string, initialDepth int) bool {
			// Limit initial depth to avoid stack overflow
			if initialDepth < 0 {
				initialDepth = 0
			}
			if initialDepth > 100 {
				initialDepth = 100
			}

			vm := New([]opcode.OpCode{})

			// Push initial frames to reach initialDepth
			for i := 0; i < initialDepth; i++ {
				scope := NewScope(vm.globalScope)
				if err := vm.PushStackFrame("setup", scope); err != nil {
					return false
				}
			}

			// Verify initial depth
			if vm.GetStackDepth() != initialDepth {
				return false
			}

			// Push a new frame (simulating function call)
			// Requirement 20.1: When function is called, system pushes new stack frame.
			scope := NewScope(vm.globalScope)
			if err := vm.PushStackFrame(functionName, scope); err != nil {
				return false
			}

			// Stack depth should be initialDepth + 1
			return vm.GetStackDepth() == initialDepth+1
		},
		gen.Identifier(),
		gen.IntRange(0, 100),
	))

	// Property: Stack depth decreases by 1 when popping a frame
	properties.Property("stack depth decreases by 1 when popping a frame", prop.ForAll(
		func(functionName string, initialDepth int) bool {
			// Ensure at least 1 frame to pop
			if initialDepth < 1 {
				initialDepth = 1
			}
			if initialDepth > 100 {
				initialDepth = 100
			}

			vm := New([]opcode.OpCode{})

			// Push initial frames
			for i := 0; i < initialDepth; i++ {
				scope := NewScope(vm.globalScope)
				if err := vm.PushStackFrame(functionName, scope); err != nil {
					return false
				}
			}

			// Verify initial depth
			if vm.GetStackDepth() != initialDepth {
				return false
			}

			// Pop a frame (simulating function return)
			// Requirement 20.2: When function returns, system pops stack frame.
			_, err := vm.PopStackFrame()
			if err != nil {
				return false
			}

			// Stack depth should be initialDepth - 1
			return vm.GetStackDepth() == initialDepth-1
		},
		gen.Identifier(),
		gen.IntRange(1, 100),
	))

	// Property: Push then pop returns to original depth (round-trip)
	properties.Property("push then pop returns to original depth", prop.ForAll(
		func(functionName string, initialDepth int) bool {
			// Limit initial depth
			if initialDepth < 0 {
				initialDepth = 0
			}
			if initialDepth > 100 {
				initialDepth = 100
			}

			vm := New([]opcode.OpCode{})

			// Push initial frames
			for i := 0; i < initialDepth; i++ {
				scope := NewScope(vm.globalScope)
				if err := vm.PushStackFrame("setup", scope); err != nil {
					return false
				}
			}

			// Record initial depth
			depthBefore := vm.GetStackDepth()

			// Push a frame (function call)
			scope := NewScope(vm.globalScope)
			if err := vm.PushStackFrame(functionName, scope); err != nil {
				return false
			}

			// Verify depth during call is n+1
			depthDuring := vm.GetStackDepth()
			if depthDuring != depthBefore+1 {
				return false
			}

			// Pop the frame (function return)
			_, err := vm.PopStackFrame()
			if err != nil {
				return false
			}

			// Verify depth after return is n
			depthAfter := vm.GetStackDepth()
			return depthAfter == depthBefore
		},
		gen.Identifier(),
		gen.IntRange(0, 100),
	))

	// Property: Multiple push/pop operations maintain correct depth
	properties.Property("multiple push/pop operations maintain correct depth", prop.ForAll(
		func(operations []bool) bool {
			// Limit operations
			if len(operations) > 50 {
				operations = operations[:50]
			}

			vm := New([]opcode.OpCode{})
			expectedDepth := 0

			for _, isPush := range operations {
				if isPush {
					// Push operation
					scope := NewScope(vm.globalScope)
					if err := vm.PushStackFrame("func", scope); err != nil {
						// Stack overflow is expected if we exceed MaxStackDepth
						if expectedDepth >= MaxStackDepth {
							continue
						}
						return false
					}
					expectedDepth++
				} else {
					// Pop operation
					if expectedDepth > 0 {
						_, err := vm.PopStackFrame()
						if err != nil {
							return false
						}
						expectedDepth--
					}
					// If expectedDepth is 0, skip pop (can't pop from empty stack)
				}

				// Verify depth matches expected
				if vm.GetStackDepth() != expectedDepth {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(30, gen.Bool()),
	))

	// Property: Popped frame contains correct function name
	properties.Property("popped frame contains correct function name", prop.ForAll(
		func(functionName string) bool {
			vm := New([]opcode.OpCode{})

			// Push a frame with specific function name
			scope := NewScope(vm.globalScope)
			if err := vm.PushStackFrame(functionName, scope); err != nil {
				return false
			}

			// Pop and verify function name
			frame, err := vm.PopStackFrame()
			if err != nil {
				return false
			}

			return frame.FunctionName == functionName
		},
		gen.Identifier(),
	))

	// Property: Nested function calls maintain LIFO order
	properties.Property("nested function calls maintain LIFO order", prop.ForAll(
		func(functionNames []string) bool {
			// Limit depth
			if len(functionNames) == 0 {
				return true
			}
			if len(functionNames) > 50 {
				functionNames = functionNames[:50]
			}

			vm := New([]opcode.OpCode{})

			// Push all frames
			for _, name := range functionNames {
				scope := NewScope(vm.globalScope)
				if err := vm.PushStackFrame(name, scope); err != nil {
					return false
				}
			}

			// Pop all frames and verify LIFO order
			for i := len(functionNames) - 1; i >= 0; i-- {
				frame, err := vm.PopStackFrame()
				if err != nil {
					return false
				}
				if frame.FunctionName != functionNames[i] {
					return false
				}
			}

			return vm.GetStackDepth() == 0
		},
		gen.SliceOfN(20, gen.Identifier()),
	))

	// Property: Local scope is correctly associated with stack frame
	properties.Property("local scope is correctly associated with stack frame", prop.ForAll(
		func(functionName string, varName string, value int) bool {
			vm := New([]opcode.OpCode{})

			// Create local scope with a variable
			localScope := NewScope(vm.globalScope)
			localScope.SetLocal(varName, value)

			// Push frame with this scope
			if err := vm.PushStackFrame(functionName, localScope); err != nil {
				return false
			}

			// Verify current scope is the local scope
			if vm.GetCurrentScope() != localScope {
				return false
			}

			// Verify variable is accessible
			val, exists := vm.GetCurrentScope().Get(varName)
			if !exists || val != value {
				return false
			}

			// Pop frame
			frame, err := vm.PopStackFrame()
			if err != nil {
				return false
			}

			// Verify frame has correct local scope
			if frame.LocalScope != localScope {
				return false
			}

			// After pop, current scope should be global
			return vm.GetCurrentScope() == vm.globalScope
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Int(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty23_StackFrameRoundTripWithUserFunction tests stack frame behavior
// with actual user-defined function calls.
// **Validates: Requirements 20.1, 20.2**
// Feature: execution-engine, Property 23: スタックフレームのラウンドトリップ
func TestProperty23_StackFrameRoundTripWithUserFunction(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: User function call pushes frame and return pops it
	properties.Property("user function call maintains stack depth invariant", prop.ForAll(
		func(funcName string, paramValue int64) bool {
			// Ensure valid function name
			if funcName == "" || funcName == "return" {
				funcName = "testFunc"
			}

			// Create a simple function that returns its parameter
			funcDef := &FunctionDef{
				Name: funcName,
				Parameters: []FunctionParam{
					{Name: "x", Type: "int", IsArray: false},
				},
				Body: []opcode.OpCode{
					{
						Cmd:  opcode.Call,
						Args: []any{"return", opcode.Variable("x")},
					},
				},
			}

			vm := New([]opcode.OpCode{})
			vm.functions[funcName] = funcDef

			// Record initial stack depth
			depthBefore := vm.GetStackDepth()

			// Call the function
			result, err := vm.callUserFunction(funcDef, []any{paramValue})
			if err != nil {
				return false
			}

			// Verify return value
			if result != paramValue {
				return false
			}

			// Verify stack depth is restored
			depthAfter := vm.GetStackDepth()
			return depthAfter == depthBefore
		},
		gen.Identifier(),
		gen.Int64(),
	))

	// Property: Recursive function calls maintain correct stack depth
	properties.Property("recursive function calls maintain correct stack depth", prop.ForAll(
		func(recursionDepth int) bool {
			// Limit recursion depth
			if recursionDepth < 1 {
				recursionDepth = 1
			}
			if recursionDepth > 50 {
				recursionDepth = 50
			}

			vm := New([]opcode.OpCode{})

			// Record initial depth
			depthBefore := vm.GetStackDepth()

			// Simulate recursive calls by pushing frames
			for i := 0; i < recursionDepth; i++ {
				scope := NewScope(vm.globalScope)
				if err := vm.PushStackFrame("recursive", scope); err != nil {
					return false
				}
			}

			// Verify depth during recursion
			if vm.GetStackDepth() != depthBefore+recursionDepth {
				return false
			}

			// Simulate returns by popping frames
			for i := 0; i < recursionDepth; i++ {
				_, err := vm.PopStackFrame()
				if err != nil {
					return false
				}
			}

			// Verify depth is restored
			return vm.GetStackDepth() == depthBefore
		},
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
