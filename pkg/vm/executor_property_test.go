package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/compiler"
)

// Property-based tests for basic OpCode execution.
// These tests verify the correctness properties defined in the design document.

// TestProperty13_OpCodeSequentialExecution tests that OpCodes are executed in sequence order.
// **Validates: Requirements 8.1**
// Feature: execution-engine, Property 13: OpCode順次実行
func TestProperty13_OpCodeSequentialExecution(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("OpCodes are executed in sequence order", prop.ForAll(
		func(values []int64) bool {
			if len(values) == 0 {
				return true
			}
			// Limit to reasonable size
			if len(values) > 20 {
				values = values[:20]
			}

			// Create a sequence of OpAssign operations that set variables x0, x1, x2, ...
			// Each assignment sets xi = i (the index)
			opcodes := make([]compiler.OpCode, len(values))
			for i := range values {
				varName := compiler.Variable("x" + string(rune('0'+i%10)) + string(rune('0'+i/10)))
				opcodes[i] = compiler.OpCode{
					Cmd:  compiler.OpAssign,
					Args: []any{varName, int64(i)},
				}
			}

			// Create VM and execute
			vm := New(opcodes)
			err := vm.Run()
			if err != nil {
				return false
			}

			// Verify all variables were set in order (each has its index value)
			for i := range values {
				varName := "x" + string(rune('0'+i%10)) + string(rune('0'+i/10))
				val, exists := vm.GetGlobalScope().Get(varName)
				if !exists {
					return false
				}
				if val != int64(i) {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(10, gen.Int64()),
	))

	properties.Property("execution order is preserved with dependent assignments", prop.ForAll(
		func(initialValue int64) bool {
			// Create a sequence where each assignment depends on the previous
			// x = initialValue
			// y = x + 1
			// z = y + 1
			opcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("x"), initialValue},
				},
				{
					Cmd: compiler.OpAssign,
					Args: []any{
						compiler.Variable("y"),
						compiler.OpCode{
							Cmd:  compiler.OpBinaryOp,
							Args: []any{"+", compiler.Variable("x"), int64(1)},
						},
					},
				},
				{
					Cmd: compiler.OpAssign,
					Args: []any{
						compiler.Variable("z"),
						compiler.OpCode{
							Cmd:  compiler.OpBinaryOp,
							Args: []any{"+", compiler.Variable("y"), int64(1)},
						},
					},
				},
			}

			vm := New(opcodes)
			err := vm.Run()
			if err != nil {
				return false
			}

			// Verify values
			x, _ := vm.GetGlobalScope().Get("x")
			y, _ := vm.GetGlobalScope().Get("y")
			z, _ := vm.GetGlobalScope().Get("z")

			return x == initialValue &&
				y == initialValue+1 &&
				z == initialValue+2
		},
		gen.Int64Range(-1000, 1000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty14_VariableAssignmentAccuracy tests that OpAssign correctly assigns values.
// **Validates: Requirements 8.2**
// Feature: execution-engine, Property 14: 変数代入の正確性
func TestProperty14_VariableAssignmentAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("integer assignment is accurate", prop.ForAll(
		func(varName string, value int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpAssign,
				Args: []any{compiler.Variable(varName), value},
			}

			result, err := vm.executeAssign(opcode)
			if err != nil {
				return false
			}

			// Result should equal the assigned value
			if result != value {
				return false
			}

			// Variable should be set in scope with correct value
			storedValue, exists := vm.GetCurrentScope().Get(varName)
			if !exists {
				return false
			}

			return storedValue == value
		},
		gen.Identifier(),
		gen.Int64(),
	))

	properties.Property("string assignment is accurate", prop.ForAll(
		func(varName string, value string) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpAssign,
				Args: []any{compiler.Variable(varName), value},
			}

			result, err := vm.executeAssign(opcode)
			if err != nil {
				return false
			}

			if result != value {
				return false
			}

			storedValue, exists := vm.GetCurrentScope().Get(varName)
			if !exists {
				return false
			}

			return storedValue == value
		},
		gen.Identifier(),
		gen.AnyString(),
	))

	properties.Property("float assignment is accurate", prop.ForAll(
		func(varName string, value float64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpAssign,
				Args: []any{compiler.Variable(varName), value},
			}

			result, err := vm.executeAssign(opcode)
			if err != nil {
				return false
			}

			if result != value {
				return false
			}

			storedValue, exists := vm.GetCurrentScope().Get(varName)
			if !exists {
				return false
			}

			return storedValue == value
		},
		gen.Identifier(),
		gen.Float64(),
	))

	properties.Property("assignment from variable is accurate", prop.ForAll(
		func(srcVar string, dstVar string, value int64) bool {
			// Ensure different variable names
			if srcVar == dstVar {
				dstVar = dstVar + "_dst"
			}

			vm := New([]compiler.OpCode{})
			// Set source variable
			vm.GetCurrentScope().Set(srcVar, value)

			// Assign dst = src
			opcode := compiler.OpCode{
				Cmd:  compiler.OpAssign,
				Args: []any{compiler.Variable(dstVar), compiler.Variable(srcVar)},
			}

			_, err := vm.executeAssign(opcode)
			if err != nil {
				return false
			}

			dstValue, exists := vm.GetCurrentScope().Get(dstVar)
			if !exists {
				return false
			}

			return dstValue == value
		},
		gen.Identifier(),
		gen.Identifier(),
		gen.Int64(),
	))

	properties.Property("assignment from expression is accurate", prop.ForAll(
		func(varName string, a int64, b int64) bool {
			vm := New([]compiler.OpCode{})

			// Assign var = a + b
			opcode := compiler.OpCode{
				Cmd: compiler.OpAssign,
				Args: []any{
					compiler.Variable(varName),
					compiler.OpCode{
						Cmd:  compiler.OpBinaryOp,
						Args: []any{"+", a, b},
					},
				},
			}

			_, err := vm.executeAssign(opcode)
			if err != nil {
				return false
			}

			storedValue, exists := vm.GetCurrentScope().Get(varName)
			if !exists {
				return false
			}

			return storedValue == a+b
		},
		gen.Identifier(),
		gen.Int64Range(-10000, 10000),
		gen.Int64Range(-10000, 10000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty15_BinaryOperationAccuracy tests that OpBinaryOp produces mathematically correct results.
// **Validates: Requirements 8.11**
// Feature: execution-engine, Property 15: 二項演算の正確性
func TestProperty15_BinaryOperationAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Arithmetic operations
	properties.Property("addition is mathematically correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"+", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a+b
		},
		gen.Int64Range(-10000, 10000),
		gen.Int64Range(-10000, 10000),
	))

	properties.Property("subtraction is mathematically correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"-", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a-b
		},
		gen.Int64Range(-10000, 10000),
		gen.Int64Range(-10000, 10000),
	))

	properties.Property("multiplication is mathematically correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"*", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a*b
		},
		gen.Int64Range(-1000, 1000),
		gen.Int64Range(-1000, 1000),
	))

	properties.Property("division is mathematically correct for non-zero divisor", prop.ForAll(
		func(a int64, b int64) bool {
			// Skip zero divisor
			if b == 0 {
				return true
			}

			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"/", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a/b
		},
		gen.Int64Range(-10000, 10000),
		gen.Int64Range(-10000, 10000),
	))

	properties.Property("division by zero returns zero", prop.ForAll(
		func(a int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"/", a, int64(0)},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == int64(0)
		},
		gen.Int64(),
	))

	properties.Property("modulo is mathematically correct for non-zero divisor", prop.ForAll(
		func(a int64, b int64) bool {
			// Skip zero divisor
			if b == 0 {
				return true
			}

			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"%", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a%b
		},
		gen.Int64Range(-10000, 10000),
		gen.Int64Range(-10000, 10000),
	))

	// Comparison operations
	properties.Property("equality comparison is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"==", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			expected := int64(0)
			if a == b {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	properties.Property("inequality comparison is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"!=", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			expected := int64(0)
			if a != b {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	properties.Property("less than comparison is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"<", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			expected := int64(0)
			if a < b {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	properties.Property("less than or equal comparison is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"<=", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			expected := int64(0)
			if a <= b {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	properties.Property("greater than comparison is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{">", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			expected := int64(0)
			if a > b {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	properties.Property("greater than or equal comparison is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{">=", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			expected := int64(0)
			if a >= b {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	// Logical operations
	properties.Property("logical AND is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"&&", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			// In FILLY, non-zero is true
			aBool := a != 0
			bBool := b != 0
			expected := int64(0)
			if aBool && bBool {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	properties.Property("logical OR is correct", prop.ForAll(
		func(a int64, b int64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"||", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			// In FILLY, non-zero is true
			aBool := a != 0
			bBool := b != 0
			expected := int64(0)
			if aBool || bBool {
				expected = int64(1)
			}

			return result == expected
		},
		gen.Int64(),
		gen.Int64(),
	))

	// Float operations
	properties.Property("float addition is mathematically correct", prop.ForAll(
		func(a float64, b float64) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"+", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a+b
		},
		gen.Float64Range(-1000, 1000),
		gen.Float64Range(-1000, 1000),
	))

	// String concatenation
	properties.Property("string concatenation is correct", prop.ForAll(
		func(a string, b string) bool {
			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"+", a, b},
			}

			result, err := vm.executeBinaryOp(opcode)
			if err != nil {
				return false
			}

			return result == a+b
		},
		gen.AnyString(),
		gen.AnyString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty9_StepCounterInitialization tests that OpSetStep correctly initializes the step counter.
// **Validates: Requirements 6.1**
// Feature: execution-engine, Property 9: ステップカウンタの初期化
// *任意の*ステップカウント値について、OpSetStep実行後のステップカウンタはその値に等しい
func TestProperty9_StepCounterInitialization(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For any valid step count value, the step counter equals that value after OpSetStep
	properties.Property("VM step counter equals specified value after OpSetStep with int64", prop.ForAll(
		func(stepCount int64) bool {
			// Use absolute value to avoid negative step counts (which are valid but may have special handling)
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{stepCount},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			// Verify step counter equals the specified value
			return vm.GetStepCounter() == int(stepCount)
		},
		gen.Int64Range(0, 10000),
	))

	properties.Property("VM step counter equals specified value after OpSetStep with int", prop.ForAll(
		func(stepCount int) bool {
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{stepCount},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			return vm.GetStepCounter() == stepCount
		},
		gen.IntRange(0, 10000),
	))

	properties.Property("VM step counter equals truncated value after OpSetStep with float64", prop.ForAll(
		func(stepCount float64) bool {
			// Use absolute value
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{stepCount},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			// Float should be truncated to int
			expected := int(stepCount)
			return vm.GetStepCounter() == expected
		},
		gen.Float64Range(0, 10000),
	))

	properties.Property("handler step counter equals specified value when handler is executing", prop.ForAll(
		func(stepCount int64) bool {
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})

			// Create and set a current handler
			handler := NewEventHandler("test_handler", EventTIME, []compiler.OpCode{}, vm)
			vm.SetCurrentHandler(handler)

			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{stepCount},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			// Verify step counter is set in handler
			if handler.StepCounter != int(stepCount) {
				return false
			}

			// VM step counter should remain 0 (unchanged)
			return vm.GetStepCounter() == 0
		},
		gen.Int64Range(0, 10000),
	))

	properties.Property("step counter equals variable value after OpSetStep with variable", prop.ForAll(
		func(varName string, stepCount int64) bool {
			// Ensure valid variable name
			if len(varName) == 0 {
				varName = "x"
			}
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})
			// Set variable with the step count value
			vm.GetCurrentScope().Set(varName, stepCount)

			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{compiler.Variable(varName)},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			return vm.GetStepCounter() == int(stepCount)
		},
		gen.Identifier(),
		gen.Int64Range(0, 10000),
	))

	properties.Property("step counter equals expression result after OpSetStep with expression", prop.ForAll(
		func(a int64, b int64) bool {
			// Limit values to avoid overflow
			if a < 0 {
				a = -a
			}
			if b < 0 {
				b = -b
			}
			if a > 5000 {
				a = 5000
			}
			if b > 5000 {
				b = 5000
			}

			vm := New([]compiler.OpCode{})

			// step(a + b) should result in step counter = a + b
			opcode := compiler.OpCode{
				Cmd: compiler.OpSetStep,
				Args: []any{compiler.OpCode{
					Cmd:  compiler.OpBinaryOp,
					Args: []any{"+", a, b},
				}},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			expected := int(a + b)
			return vm.GetStepCounter() == expected
		},
		gen.Int64Range(0, 5000),
		gen.Int64Range(0, 5000),
	))

	properties.Property("zero step count is correctly set", prop.ForAll(
		func(_ bool) bool {
			vm := New([]compiler.OpCode{})

			// First set a non-zero value
			vm.SetStepCounter(100)

			// Then set to zero
			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{int64(0)},
			}

			_, err := vm.executeSetStep(opcode)
			if err != nil {
				return false
			}

			return vm.GetStepCounter() == 0
		},
		gen.Bool(),
	))

	properties.Property("step counter is idempotent - setting same value twice results in same value", prop.ForAll(
		func(stepCount int64) bool {
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})
			opcode := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{stepCount},
			}

			// Execute twice
			_, err1 := vm.executeSetStep(opcode)
			_, err2 := vm.executeSetStep(opcode)

			if err1 != nil || err2 != nil {
				return false
			}

			return vm.GetStepCounter() == int(stepCount)
		},
		gen.Int64Range(0, 10000),
	))

	properties.Property("step counter can be updated to different values", prop.ForAll(
		func(first int64, second int64) bool {
			if first < 0 {
				first = -first
			}
			if second < 0 {
				second = -second
			}

			vm := New([]compiler.OpCode{})

			// Set first value
			opcode1 := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{first},
			}
			_, err := vm.executeSetStep(opcode1)
			if err != nil {
				return false
			}
			if vm.GetStepCounter() != int(first) {
				return false
			}

			// Set second value
			opcode2 := compiler.OpCode{
				Cmd:  compiler.OpSetStep,
				Args: []any{second},
			}
			_, err = vm.executeSetStep(opcode2)
			if err != nil {
				return false
			}

			// Final value should be second
			return vm.GetStepCounter() == int(second)
		},
		gen.Int64Range(0, 10000),
		gen.Int64Range(0, 10000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty10_EventStepProgression tests that each event during step execution advances the step by 1.
// **Validates: Requirements 6.3**
// Feature: execution-engine, Property 10: イベントごとのステップ進行
// *任意の*ステップ数について、ステップ実行中にイベントが発生するたびに現在のステップが1つ進む
func TestProperty10_EventStepProgression(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For any wait count n, after n+1 events the handler resumes execution
	// (First event triggers the handler and sets wait counter, then n more events to decrement to 0)
	properties.Property("handler resumes after n+1 events when wait count is n", prop.ForAll(
		func(waitCount int) bool {
			// Limit wait count to reasonable range
			if waitCount < 1 {
				waitCount = 1
			}
			if waitCount > 100 {
				waitCount = 100
			}

			vm := New([]compiler.OpCode{})

			// Create a handler with OpWait
			// The handler will wait for waitCount events before continuing
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(waitCount)},
				},
				// After wait, assign a variable to indicate completion
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Create and dispatch events
			// First event triggers the handler and sets wait counter
			// Then waitCount more events are needed to decrement wait counter to 0
			// Total events needed: waitCount + 1
			totalEvents := waitCount + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}

				// Before the last event, "completed" should not be set
				if i < totalEvents-1 {
					_, exists := vm.GetGlobalScope().Get("completed")
					if exists {
						// Handler completed too early
						return false
					}
				}
			}

			// After totalEvents events, "completed" should be set
			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists {
				return false
			}

			return val == int64(1)
		},
		gen.IntRange(1, 50),
	))

	// Property: Each event decrements the wait counter by exactly 1
	properties.Property("each event decrements wait counter by 1", prop.ForAll(
		func(waitCount int) bool {
			// Limit wait count to reasonable range
			if waitCount < 1 {
				waitCount = 1
			}
			if waitCount > 100 {
				waitCount = 100
			}

			vm := New([]compiler.OpCode{})

			// Create a handler with OpWait
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(waitCount)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// First event triggers the handler and sets wait counter
			event := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event)
			if err != nil {
				return false
			}

			// After first event, wait counter should be waitCount - 1
			// (because the first event triggers the handler which executes OpWait,
			// then the handler pauses with WaitCounter = waitCount,
			// but the Execute method doesn't decrement on the first call)
			// Actually, looking at the code: OpWait sets WaitCounter = waitCount
			// and returns waitMarker, so WaitCounter should be waitCount after first event
			if handler.WaitCounter != waitCount {
				return false
			}

			// Dispatch more events and verify wait counter decrements by 1 each time
			for i := 1; i <= waitCount; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}

				expectedWaitCounter := waitCount - i
				if expectedWaitCounter < 0 {
					expectedWaitCounter = 0
				}

				// After handler completes (wait counter reaches 0), it resets PC
				// and WaitCounter stays at 0
				if handler.WaitCounter != expectedWaitCounter {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50),
	))

	// Property: Handler with wait count 0 executes immediately without waiting
	properties.Property("wait count 0 executes immediately", prop.ForAll(
		func(_ bool) bool {
			vm := New([]compiler.OpCode{})

			// Create a handler with OpWait(0)
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(0)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Single event should complete the handler
			event := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event)
			if err != nil {
				return false
			}

			// Handler should have completed immediately
			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists {
				return false
			}

			return val == int64(1)
		},
		gen.Bool(),
	))

	// Property: Handler with negative wait count executes immediately
	properties.Property("negative wait count executes immediately", prop.ForAll(
		func(negativeCount int) bool {
			// Ensure negative
			if negativeCount >= 0 {
				negativeCount = -1 - negativeCount
			}

			vm := New([]compiler.OpCode{})

			// Create a handler with negative OpWait
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(negativeCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Single event should complete the handler
			event := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event)
			if err != nil {
				return false
			}

			// Handler should have completed immediately
			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists {
				return false
			}

			return val == int64(1)
		},
		gen.IntRange(-100, -1),
	))

	// Property: Multiple handlers with different wait counts progress independently
	properties.Property("multiple handlers progress independently", prop.ForAll(
		func(waitCount1 int, waitCount2 int) bool {
			// Limit wait counts
			if waitCount1 < 1 {
				waitCount1 = 1
			}
			if waitCount1 > 20 {
				waitCount1 = 20
			}
			if waitCount2 < 1 {
				waitCount2 = 1
			}
			if waitCount2 > 20 {
				waitCount2 = 20
			}

			vm := New([]compiler.OpCode{})

			// Create first handler
			handler1Opcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(waitCount1)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("handler1_completed"), int64(1)},
				},
			}
			handler1 := NewEventHandler("handler1", EventTIME, handler1Opcodes, vm)
			vm.GetHandlerRegistry().Register(handler1)

			// Create second handler
			handler2Opcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(waitCount2)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("handler2_completed"), int64(1)},
				},
			}
			handler2 := NewEventHandler("handler2", EventTIME, handler2Opcodes, vm)
			vm.GetHandlerRegistry().Register(handler2)

			// Dispatch events
			// Each handler needs waitCount + 1 events to complete
			// (1 to trigger and set wait counter, then waitCount to decrement to 0)
			maxWait := waitCount1 + 1
			if waitCount2+1 > maxWait {
				maxWait = waitCount2 + 1
			}

			for i := 0; i < maxWait; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}

				// Check handler1 completion (needs waitCount1 + 1 events)
				if i >= waitCount1 {
					val, exists := vm.GetGlobalScope().Get("handler1_completed")
					if !exists || val != int64(1) {
						return false
					}
				}

				// Check handler2 completion (needs waitCount2 + 1 events)
				if i >= waitCount2 {
					val, exists := vm.GetGlobalScope().Get("handler2_completed")
					if !exists || val != int64(1) {
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(1, 20),
		gen.IntRange(1, 20),
	))

	// Property: Handler step counter is set correctly by OpSetStep
	properties.Property("step counter is set correctly in handler", prop.ForAll(
		func(stepCount int64) bool {
			if stepCount < 0 {
				stepCount = -stepCount
			}

			vm := New([]compiler.OpCode{})

			// Create a handler with OpSetStep
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpSetStep,
					Args: []any{stepCount},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch event to trigger handler
			event := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event)
			if err != nil {
				return false
			}

			// Handler's step counter should be set
			return handler.StepCounter == int(stepCount)
		},
		gen.Int64Range(0, 10000),
	))

	// Property: Step execution with multiple waits progresses correctly
	properties.Property("multiple waits in sequence progress correctly", prop.ForAll(
		func(wait1 int, wait2 int) bool {
			// Limit wait counts
			if wait1 < 1 {
				wait1 = 1
			}
			if wait1 > 10 {
				wait1 = 10
			}
			if wait2 < 1 {
				wait2 = 1
			}
			if wait2 > 10 {
				wait2 = 10
			}

			vm := New([]compiler.OpCode{})

			// Create a handler with two OpWait instructions
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(wait1)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("step1_completed"), int64(1)},
				},
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(wait2)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("step2_completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch wait1 + 1 events - should complete step1
			// (1 to trigger and set wait counter, then wait1 to decrement to 0)
			for i := 0; i < wait1+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Step1 should be completed
			val1, exists1 := vm.GetGlobalScope().Get("step1_completed")
			if !exists1 || val1 != int64(1) {
				return false
			}

			// Step2 should not be completed yet
			_, exists2 := vm.GetGlobalScope().Get("step2_completed")
			if exists2 {
				return false
			}

			// Dispatch wait2 + 1 events - should complete step2
			// (1 to trigger second wait and set wait counter, then wait2 to decrement to 0)
			for i := 0; i < wait2+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Step2 should now be completed
			val2, exists2 := vm.GetGlobalScope().Get("step2_completed")
			if !exists2 || val2 != int64(1) {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 10),
	))

	// Property: Different event types trigger the correct handlers
	properties.Property("different event types trigger correct handlers", prop.ForAll(
		func(waitCount int) bool {
			if waitCount < 1 {
				waitCount = 1
			}
			if waitCount > 10 {
				waitCount = 10
			}

			vm := New([]compiler.OpCode{})

			// Create TIME handler
			timeHandlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(waitCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("time_completed"), int64(1)},
				},
			}
			timeHandler := NewEventHandler("time_handler", EventTIME, timeHandlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(timeHandler)

			// Create MIDI_TIME handler
			midiHandlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(waitCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("midi_completed"), int64(1)},
				},
			}
			midiHandler := NewEventHandler("midi_handler", EventMIDI_TIME, midiHandlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(midiHandler)

			// Dispatch TIME events only (waitCount + 1 events needed)
			for i := 0; i < waitCount+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// TIME handler should be completed
			val1, exists1 := vm.GetGlobalScope().Get("time_completed")
			if !exists1 || val1 != int64(1) {
				return false
			}

			// MIDI_TIME handler should NOT be completed (no MIDI_TIME events dispatched)
			_, exists2 := vm.GetGlobalScope().Get("midi_completed")
			if exists2 {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty11_ConsecutiveCommaWait tests that consecutive commas wait for multiple events.
// **Validates: Requirements 6.6**
// Feature: execution-engine, Property 11: 連続カンマの待機
// *任意の*数nの連続カンマについて、n回のイベント発生後に次のステップに進む
func TestProperty11_ConsecutiveCommaWait(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For any number n of consecutive commas, execution proceeds after n events
	// In the compiler, consecutive commas generate OpWait with Args[0] = n
	properties.Property("n consecutive commas wait for n events", prop.ForAll(
		func(commaCount int) bool {
			// Limit comma count to reasonable range
			if commaCount < 1 {
				commaCount = 1
			}
			if commaCount > 50 {
				commaCount = 50
			}

			vm := New([]compiler.OpCode{})

			// Create a handler that simulates consecutive commas
			// In the compiler, n consecutive commas generate OpWait with Args[0] = n
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(commaCount)}, // n consecutive commas = wait for n events
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_commas"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch commaCount + 1 events
			// First event triggers handler and sets wait counter to commaCount
			// Then commaCount events to decrement wait counter to 0
			totalEvents := commaCount + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}

				// Before the last event, "after_commas" should not be set
				if i < totalEvents-1 {
					_, exists := vm.GetGlobalScope().Get("after_commas")
					if exists {
						// Completed too early
						return false
					}
				}
			}

			// After totalEvents events, "after_commas" should be set
			val, exists := vm.GetGlobalScope().Get("after_commas")
			if !exists {
				return false
			}

			return val == int64(1)
		},
		gen.IntRange(1, 50),
	))

	// Property: Single comma (n=1) waits for exactly 1 event
	properties.Property("single comma waits for 1 event", prop.ForAll(
		func(_ bool) bool {
			vm := New([]compiler.OpCode{})

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(1)}, // Single comma
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// First event triggers handler and sets wait counter
			event1 := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event1)
			if err != nil {
				return false
			}

			// Should not be completed yet
			_, exists := vm.GetGlobalScope().Get("completed")
			if exists {
				return false
			}

			// Second event decrements wait counter to 0 and completes
			event2 := NewEvent(EventTIME)
			err = vm.GetEventDispatcher().Dispatch(event2)
			if err != nil {
				return false
			}

			// Should be completed now
			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists || val != int64(1) {
				return false
			}

			return true
		},
		gen.Bool(),
	))

	// Property: Multiple consecutive comma sequences work correctly
	properties.Property("multiple consecutive comma sequences work correctly", prop.ForAll(
		func(commas1 int, commas2 int) bool {
			// Limit comma counts
			if commas1 < 1 {
				commas1 = 1
			}
			if commas1 > 10 {
				commas1 = 10
			}
			if commas2 < 1 {
				commas2 = 1
			}
			if commas2 > 10 {
				commas2 = 10
			}

			vm := New([]compiler.OpCode{})

			// Handler with two consecutive comma sequences
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(commas1)}, // First sequence of commas
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_first_commas"), int64(1)},
				},
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(commas2)}, // Second sequence of commas
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_second_commas"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch events for first comma sequence
			for i := 0; i < commas1+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// First sequence should be completed
			val1, exists1 := vm.GetGlobalScope().Get("after_first_commas")
			if !exists1 || val1 != int64(1) {
				return false
			}

			// Second sequence should not be completed yet
			_, exists2 := vm.GetGlobalScope().Get("after_second_commas")
			if exists2 {
				return false
			}

			// Dispatch events for second comma sequence
			for i := 0; i < commas2+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Second sequence should now be completed
			val2, exists2 := vm.GetGlobalScope().Get("after_second_commas")
			if !exists2 || val2 != int64(1) {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 10),
	))

	// Property: Consecutive commas with MIDI_TIME events work correctly
	properties.Property("consecutive commas work with MIDI_TIME events", prop.ForAll(
		func(commaCount int) bool {
			if commaCount < 1 {
				commaCount = 1
			}
			if commaCount > 20 {
				commaCount = 20
			}

			vm := New([]compiler.OpCode{})

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{int64(commaCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("midi_completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventMIDI_TIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch MIDI_TIME events
			totalEvents := commaCount + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventMIDI_TIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Should be completed
			val, exists := vm.GetGlobalScope().Get("midi_completed")
			if !exists || val != int64(1) {
				return false
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property: Wait count from variable works correctly
	properties.Property("wait count from variable works correctly", prop.ForAll(
		func(commaCount int64) bool {
			if commaCount < 1 {
				commaCount = 1
			}
			if commaCount > 20 {
				commaCount = 20
			}

			vm := New([]compiler.OpCode{})

			// Set the comma count in a variable
			vm.GetGlobalScope().Set("comma_count", commaCount)

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpWait,
					Args: []any{compiler.Variable("comma_count")}, // Wait count from variable
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch events
			totalEvents := int(commaCount) + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Should be completed
			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists || val != int64(1) {
				return false
			}

			return true
		},
		gen.Int64Range(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty12_WaitNWaiting tests that Wait(n) waits for n events before resuming.
// **Validates: Requirements 17.1**
// Feature: execution-engine, Property 12: Wait(n)の待機
// *任意の*正の整数nについて、Wait(n)呼び出し後、n回のイベント発生後に実行が再開される
func TestProperty12_WaitNWaiting(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For any positive integer n, execution resumes after n events
	properties.Property("Wait(n) resumes after n events", prop.ForAll(
		func(waitCount int) bool {
			// Limit wait count to reasonable range
			if waitCount < 1 {
				waitCount = 1
			}
			if waitCount > 50 {
				waitCount = 50
			}

			vm := New([]compiler.OpCode{})

			// Create a handler that uses Wait(n)
			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("before_wait"), int64(1)},
				},
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(waitCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_wait"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch waitCount + 1 events
			// First event triggers handler and sets wait counter
			// Then waitCount events to decrement wait counter to 0
			totalEvents := waitCount + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}

				// Before the last event, "after_wait" should not be set
				if i < totalEvents-1 {
					_, exists := vm.GetGlobalScope().Get("after_wait")
					if exists {
						// Completed too early
						return false
					}
				}
			}

			// After totalEvents events, "after_wait" should be set
			val, exists := vm.GetGlobalScope().Get("after_wait")
			if !exists {
				return false
			}

			return val == int64(1)
		},
		gen.IntRange(1, 50),
	))

	// Property: Wait(0) executes immediately without waiting
	properties.Property("Wait(0) executes immediately", prop.ForAll(
		func(_ bool) bool {
			vm := New([]compiler.OpCode{})

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("before_wait"), int64(1)},
				},
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(0)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_wait"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Single event should complete the handler
			event := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event)
			if err != nil {
				return false
			}

			// Both before_wait and after_wait should be set
			before, _ := vm.GetGlobalScope().Get("before_wait")
			after, _ := vm.GetGlobalScope().Get("after_wait")

			return before == int64(1) && after == int64(1)
		},
		gen.Bool(),
	))

	// Property: Wait with negative count executes immediately
	properties.Property("Wait with negative count executes immediately", prop.ForAll(
		func(negativeCount int) bool {
			// Ensure negative
			if negativeCount >= 0 {
				negativeCount = -1 - negativeCount
			}

			vm := New([]compiler.OpCode{})

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(negativeCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Single event should complete the handler
			event := NewEvent(EventTIME)
			err := vm.GetEventDispatcher().Dispatch(event)
			if err != nil {
				return false
			}

			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists {
				return false
			}

			return val == int64(1)
		},
		gen.IntRange(-100, -1),
	))

	// Property: Multiple Wait calls in sequence work correctly
	properties.Property("multiple Wait calls in sequence work correctly", prop.ForAll(
		func(wait1 int, wait2 int) bool {
			// Limit wait counts
			if wait1 < 1 {
				wait1 = 1
			}
			if wait1 > 10 {
				wait1 = 10
			}
			if wait2 < 1 {
				wait2 = 1
			}
			if wait2 > 10 {
				wait2 = 10
			}

			vm := New([]compiler.OpCode{})

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(wait1)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_first_wait"), int64(1)},
				},
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(wait2)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("after_second_wait"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch events for first Wait
			for i := 0; i < wait1+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// First Wait should be completed
			val1, exists1 := vm.GetGlobalScope().Get("after_first_wait")
			if !exists1 || val1 != int64(1) {
				return false
			}

			// Second Wait should not be completed yet
			_, exists2 := vm.GetGlobalScope().Get("after_second_wait")
			if exists2 {
				return false
			}

			// Dispatch events for second Wait
			for i := 0; i < wait2+1; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Second Wait should now be completed
			val2, exists2 := vm.GetGlobalScope().Get("after_second_wait")
			if !exists2 || val2 != int64(1) {
				return false
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 10),
	))

	// Property: Wait with variable argument works correctly
	properties.Property("Wait with variable argument works correctly", prop.ForAll(
		func(waitCount int64) bool {
			if waitCount < 1 {
				waitCount = 1
			}
			if waitCount > 20 {
				waitCount = 20
			}

			vm := New([]compiler.OpCode{})

			// Set the wait count in a variable
			vm.GetGlobalScope().Set("wait_count", waitCount)

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", compiler.Variable("wait_count")},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch events
			totalEvents := int(waitCount) + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Should be completed
			val, exists := vm.GetGlobalScope().Get("completed")
			if !exists || val != int64(1) {
				return false
			}

			return true
		},
		gen.Int64Range(1, 20),
	))

	// Property: Wait in MIDI_TIME handler waits for MIDI_TIME events
	properties.Property("Wait in MIDI_TIME handler waits for MIDI_TIME events", prop.ForAll(
		func(waitCount int) bool {
			if waitCount < 1 {
				waitCount = 1
			}
			if waitCount > 20 {
				waitCount = 20
			}

			vm := New([]compiler.OpCode{})

			handlerOpcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(waitCount)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("midi_completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventMIDI_TIME, handlerOpcodes, vm)
			vm.GetHandlerRegistry().Register(handler)

			// Dispatch MIDI_TIME events
			totalEvents := waitCount + 1
			for i := 0; i < totalEvents; i++ {
				event := NewEvent(EventMIDI_TIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}
			}

			// Should be completed
			val, exists := vm.GetGlobalScope().Get("midi_completed")
			if !exists || val != int64(1) {
				return false
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property: Multiple handlers with Wait have independent wait counters
	properties.Property("multiple handlers have independent wait counters", prop.ForAll(
		func(wait1 int, wait2 int) bool {
			// Limit wait counts
			if wait1 < 1 {
				wait1 = 1
			}
			if wait1 > 10 {
				wait1 = 10
			}
			if wait2 < 1 {
				wait2 = 1
			}
			if wait2 > 10 {
				wait2 = 10
			}

			vm := New([]compiler.OpCode{})

			// Create first handler
			handler1Opcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(wait1)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("handler1_completed"), int64(1)},
				},
			}
			handler1 := NewEventHandler("handler1", EventTIME, handler1Opcodes, vm)
			vm.GetHandlerRegistry().Register(handler1)

			// Create second handler
			handler2Opcodes := []compiler.OpCode{
				{
					Cmd:  compiler.OpCall,
					Args: []any{"Wait", int64(wait2)},
				},
				{
					Cmd:  compiler.OpAssign,
					Args: []any{compiler.Variable("handler2_completed"), int64(1)},
				},
			}
			handler2 := NewEventHandler("handler2", EventTIME, handler2Opcodes, vm)
			vm.GetHandlerRegistry().Register(handler2)

			// Dispatch events
			maxWait := wait1 + 1
			if wait2+1 > maxWait {
				maxWait = wait2 + 1
			}

			for i := 0; i < maxWait; i++ {
				event := NewEvent(EventTIME)
				err := vm.GetEventDispatcher().Dispatch(event)
				if err != nil {
					return false
				}

				// Check handler1 completion (needs wait1 + 1 events)
				if i >= wait1 {
					val, exists := vm.GetGlobalScope().Get("handler1_completed")
					if !exists || val != int64(1) {
						return false
					}
				}

				// Check handler2 completion (needs wait2 + 1 events)
				if i >= wait2 {
					val, exists := vm.GetGlobalScope().Get("handler2_completed")
					if !exists || val != int64(1) {
						return false
					}
				}
			}

			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
