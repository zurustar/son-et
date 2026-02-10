package vm

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
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
			opcodes := make([]opcode.OpCode, len(values))
			for i := range values {
				varName := opcode.Variable("x" + string(rune('0'+i%10)) + string(rune('0'+i/10)))
				opcodes[i] = opcode.OpCode{
					Cmd:  opcode.Assign,
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
			opcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("x"), initialValue},
				},
				{
					Cmd: opcode.Assign,
					Args: []any{
						opcode.Variable("y"),
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", opcode.Variable("x"), int64(1)},
						},
					},
				},
				{
					Cmd: opcode.Assign,
					Args: []any{
						opcode.Variable("z"),
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", opcode.Variable("y"), int64(1)},
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.Assign,
				Args: []any{opcode.Variable(varName), value},
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.Assign,
				Args: []any{opcode.Variable(varName), value},
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.Assign,
				Args: []any{opcode.Variable(varName), value},
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

			vm := New([]opcode.OpCode{})
			// Set source variable
			vm.GetCurrentScope().Set(srcVar, value)

			// Assign dst = src
			opcode := opcode.OpCode{
				Cmd:  opcode.Assign,
				Args: []any{opcode.Variable(dstVar), opcode.Variable(srcVar)},
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
			vm := New([]opcode.OpCode{})

			// Assign var = a + b
			opcode := opcode.OpCode{
				Cmd: opcode.Assign,
				Args: []any{
					opcode.Variable(varName),
					opcode.OpCode{
						Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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

			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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

			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.BinaryOp,
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

			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})

			// Create and set a current handler
			handler := NewEventHandler("test_handler", EventTIME, []opcode.OpCode{}, vm, nil)
			vm.SetCurrentHandler(handler)

			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})
			// Set variable with the step count value
			vm.GetCurrentScope().Set(varName, stepCount)

			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
				Args: []any{opcode.Variable(varName)},
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

			vm := New([]opcode.OpCode{})

			// step(a + b) should result in step counter = a + b
			opcode := opcode.OpCode{
				Cmd: opcode.SetStep,
				Args: []any{opcode.OpCode{
					Cmd:  opcode.BinaryOp,
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
			vm := New([]opcode.OpCode{})

			// First set a non-zero value
			vm.SetStepCounter(100)

			// Then set to zero
			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})
			opcode := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})

			// Set first value
			opcode1 := opcode.OpCode{
				Cmd:  opcode.SetStep,
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
			opcode2 := opcode.OpCode{
				Cmd:  opcode.SetStep,
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

			vm := New([]opcode.OpCode{})

			// Create a handler with OpWait
			// The handler will wait for waitCount events before continuing
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(waitCount)},
				},
				// After wait, assign a variable to indicate completion
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create a handler with OpWait
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(waitCount)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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
			vm := New([]opcode.OpCode{})

			// Create a handler with OpWait(0)
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(0)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create a handler with negative OpWait
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(negativeCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create first handler
			handler1Opcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(waitCount1)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("handler1_completed"), int64(1)},
				},
			}
			handler1 := NewEventHandler("handler1", EventTIME, handler1Opcodes, vm, nil)
			vm.GetHandlerRegistry().Register(handler1)

			// Create second handler
			handler2Opcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(waitCount2)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("handler2_completed"), int64(1)},
				},
			}
			handler2 := NewEventHandler("handler2", EventTIME, handler2Opcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create a handler with OpSetStep
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.SetStep,
					Args: []any{stepCount},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create a handler with two OpWait instructions
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(wait1)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("step1_completed"), int64(1)},
				},
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(wait2)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("step2_completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create TIME handler
			timeHandlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(waitCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("time_completed"), int64(1)},
				},
			}
			timeHandler := NewEventHandler("time_handler", EventTIME, timeHandlerOpcodes, vm, nil)
			vm.GetHandlerRegistry().Register(timeHandler)

			// Create MIDI_TIME handler
			midiHandlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(waitCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("midi_completed"), int64(1)},
				},
			}
			midiHandler := NewEventHandler("midi_handler", EventMIDI_TIME, midiHandlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create a handler that simulates consecutive commas
			// In the compiler, n consecutive commas generate OpWait with Args[0] = n
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(commaCount)}, // n consecutive commas = wait for n events
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_commas"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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
			vm := New([]opcode.OpCode{})

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(1)}, // Single comma
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Handler with two consecutive comma sequences
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(commas1)}, // First sequence of commas
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_first_commas"), int64(1)},
				},
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(commas2)}, // Second sequence of commas
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_second_commas"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{int64(commaCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("midi_completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventMIDI_TIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Set the comma count in a variable
			vm.GetGlobalScope().Set("comma_count", commaCount)

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Wait,
					Args: []any{opcode.Variable("comma_count")}, // Wait count from variable
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create a handler that uses Wait(n)
			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("before_wait"), int64(1)},
				},
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(waitCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_wait"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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
			vm := New([]opcode.OpCode{})

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("before_wait"), int64(1)},
				},
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(0)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_wait"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(negativeCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(wait1)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_first_wait"), int64(1)},
				},
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(wait2)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("after_second_wait"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Set the wait count in a variable
			vm.GetGlobalScope().Set("wait_count", waitCount)

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", opcode.Variable("wait_count")},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventTIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			handlerOpcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(waitCount)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("midi_completed"), int64(1)},
				},
			}

			handler := NewEventHandler("test_handler", EventMIDI_TIME, handlerOpcodes, vm, nil)
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

			vm := New([]opcode.OpCode{})

			// Create first handler
			handler1Opcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(wait1)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("handler1_completed"), int64(1)},
				},
			}
			handler1 := NewEventHandler("handler1", EventTIME, handler1Opcodes, vm, nil)
			vm.GetHandlerRegistry().Register(handler1)

			// Create second handler
			handler2Opcodes := []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"Wait", int64(wait2)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("handler2_completed"), int64(1)},
				},
			}
			handler2 := NewEventHandler("handler2", EventTIME, handler2Opcodes, vm, nil)
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

// TestPropertySwitch1_NoFallthrough tests that only the matched case body is executed
// and no fallthrough occurs, regardless of break presence.
// **Validates: Requirements 1.1, 1.3**
// Feature: switch-edge-cases, Property 1: フォールスルーなし
func TestPropertySwitch1_NoFallthrough(t *testing.T) {
	const minSuccessfulTests = 100

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = minSuccessfulTests

	properties := gopter.NewProperties(parameters)

	// Property: For any switch value and list of case values, only the matched case's
	// marker variable is set. All other case marker variables remain unset.
	properties.Property("only matched case body executes, no fallthrough", prop.ForAll(
		func(switchVal int64, caseVals []int64) bool {
			// Need at least 2 cases to test fallthrough behavior
			if len(caseVals) < 2 {
				return true
			}
			// Limit number of cases to a reasonable size
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Deduplicate case values to avoid ambiguous matching
			seen := make(map[int64]bool)
			uniqueCaseVals := make([]int64, 0, len(caseVals))
			for _, v := range caseVals {
				if !seen[v] {
					seen[v] = true
					uniqueCaseVals = append(uniqueCaseVals, v)
				}
			}
			if len(uniqueCaseVals) < 2 {
				return true
			}
			caseVals = uniqueCaseVals

			// Build case clauses: each case body assigns marker_i = 1
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				markerVar := opcode.Variable(fmt.Sprintf("marker_%d", i))
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{markerVar, int64(1)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			// Build switch OpCode
			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			// Execute
			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Determine which case should have matched
			matchedIndex := -1
			for i, cv := range caseVals {
				if cv == switchVal {
					matchedIndex = i
					break
				}
			}

			// Verify: only the matched case's marker is set
			for i := range caseVals {
				markerName := fmt.Sprintf("marker_%d", i)
				val, exists := vm.GetGlobalScope().Get(markerName)
				if i == matchedIndex {
					// Matched case: marker must be set to 1
					if !exists || val != int64(1) {
						return false
					}
				} else {
					// Non-matched case: marker must NOT be set
					if exists {
						return false
					}
				}
			}

			return true
		},
		gen.Int64Range(-50, 50),
		gen.SliceOfN(8, gen.Int64Range(-50, 50)),
	))

	// Property: break inside a matched case body does not cause fallthrough.
	// The case body with break should still only execute that one case.
	properties.Property("break in case body does not cause fallthrough", prop.ForAll(
		func(switchVal int64, otherVals []int64) bool {
			// Need at least 1 other value
			if len(otherVals) < 1 {
				return true
			}
			const maxOther = 5
			if len(otherVals) > maxOther {
				otherVals = otherVals[:maxOther]
			}

			// Ensure switchVal is not in otherVals, then prepend it
			// so we guarantee a match at index 0
			filteredOther := make([]int64, 0, len(otherVals))
			for _, v := range otherVals {
				if v != switchVal {
					filteredOther = append(filteredOther, v)
				}
			}
			if len(filteredOther) == 0 {
				return true
			}

			// Build all case values: switchVal first, then others
			allCaseVals := append([]int64{switchVal}, filteredOther...)

			// Build case clauses: matched case has break after assignment
			cases := make([]any, len(allCaseVals))
			for i, cv := range allCaseVals {
				markerVar := opcode.Variable(fmt.Sprintf("marker_%d", i))
				var caseBody []opcode.OpCode
				if i == 0 {
					// Matched case: assign marker, then break, then another assign (should be skipped)
					caseBody = []opcode.OpCode{
						{
							Cmd:  opcode.Assign,
							Args: []any{markerVar, int64(1)},
						},
						{
							Cmd:  opcode.Break,
							Args: []any{},
						},
						{
							Cmd: opcode.Assign,
							Args: []any{
								opcode.Variable(fmt.Sprintf("after_break_%d", i)),
								int64(1),
							},
						},
					}
				} else {
					caseBody = []opcode.OpCode{
						{
							Cmd:  opcode.Assign,
							Args: []any{markerVar, int64(1)},
						},
					}
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Matched case (index 0) marker should be set
			val, exists := vm.GetGlobalScope().Get("marker_0")
			if !exists || val != int64(1) {
				return false
			}

			// Code after break in matched case should NOT have executed
			_, afterBreakExists := vm.GetGlobalScope().Get("after_break_0")
			if afterBreakExists {
				return false
			}

			// No other case markers should be set (no fallthrough)
			for i := 1; i < len(allCaseVals); i++ {
				markerName := fmt.Sprintf("marker_%d", i)
				_, exists := vm.GetGlobalScope().Get(markerName)
				if exists {
					return false
				}
			}

			return true
		},
		gen.Int64Range(-50, 50),
		gen.SliceOfN(5, gen.Int64Range(-50, 50)),
	))

	// Property: when no case matches, no case body executes
	properties.Property("no case body executes when switch value matches nothing", prop.ForAll(
		func(switchVal int64, caseVals []int64) bool {
			if len(caseVals) == 0 {
				return true
			}
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Ensure switchVal does NOT match any case value
			for i := range caseVals {
				if caseVals[i] == switchVal {
					caseVals[i] = switchVal + 100
				}
			}

			// Build case clauses
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				markerVar := opcode.Variable(fmt.Sprintf("marker_%d", i))
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{markerVar, int64(1)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// No markers should be set
			for i := range caseVals {
				markerName := fmt.Sprintf("marker_%d", i)
				_, exists := vm.GetGlobalScope().Get(markerName)
				if exists {
					return false
				}
			}

			return true
		},
		gen.Int64Range(-50, 50),
		gen.SliceOfN(8, gen.Int64Range(-50, 50)),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: switch-edge-cases, Property 2: ループ内breakの正確性
// **Validates: Requirements 1.2**
func TestPropertySwitch2_BreakInLoopAccuracy(t *testing.T) {
	const minSuccessfulTests = 100

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = minSuccessfulTests

	properties := gopter.NewProperties(parameters)

	// Property: For any loop count and switch value, a break inside a switch case
	// within a for loop only exits the switch, not the outer loop.
	// The loop counter should always reach the expected iteration count.
	properties.Property("break in switch case does not terminate outer for loop", prop.ForAll(
		func(loopCount int64, switchVal int64, caseVals []int64) bool {
			// Constrain loop count to a reasonable positive range
			if loopCount < 1 {
				return true
			}
			const maxLoopCount = 20
			if loopCount > maxLoopCount {
				loopCount = maxLoopCount
			}
			// Need at least 1 case value
			if len(caseVals) < 1 {
				return true
			}
			const maxCases = 5
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Deduplicate case values
			seen := make(map[int64]bool)
			uniqueCaseVals := make([]int64, 0, len(caseVals))
			for _, v := range caseVals {
				if !seen[v] {
					seen[v] = true
					uniqueCaseVals = append(uniqueCaseVals, v)
				}
			}
			caseVals = uniqueCaseVals

			// Build case clauses: each case body has a break statement
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				cases[i] = map[string]any{
					"value": cv,
					"body": []opcode.OpCode{
						{Cmd: opcode.Break, Args: []any{}},
					},
				}
			}

			// Build: for (i = 0; i < loopCount; i = i + 1) {
			//          switch (switchVal) { case cv1: break; case cv2: break; ... }
			//          count = count + 1
			//        }
			forOp := opcode.OpCode{
				Cmd: opcode.For,
				Args: []any{
					// init: i = 0
					[]opcode.OpCode{
						{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
					},
					// condition: i < loopCount
					opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), loopCount}},
					// post: i = i + 1
					[]opcode.OpCode{
						{Cmd: opcode.Assign, Args: []any{
							opcode.Variable("i"),
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
						}},
					},
					// body: switch + count increment
					[]opcode.OpCode{
						{
							Cmd:  opcode.Switch,
							Args: []any{switchVal, cases},
						},
						{Cmd: opcode.Assign, Args: []any{
							opcode.Variable("count"),
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("count"), int64(1)}},
						}},
					},
				},
			}

			// Execute with count initialized to 0
			vm := New([]opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("count"), int64(0)}},
				forOp,
			})
			err := vm.Run()
			if err != nil {
				return false
			}

			// The loop should have completed all iterations regardless of break in switch
			val, exists := vm.GetGlobalScope().Get("count")
			if !exists {
				return false
			}
			return val == loopCount
		},
		gen.Int64Range(1, 20),
		gen.Int64Range(-50, 50),
		gen.SliceOfN(5, gen.Int64Range(-50, 50)),
	))

	// Property: break in switch default block within a for loop does not terminate the loop.
	properties.Property("break in switch default does not terminate outer for loop", prop.ForAll(
		func(loopCount int64) bool {
			if loopCount < 1 {
				return true
			}
			const maxLoopCount = 20
			if loopCount > maxLoopCount {
				loopCount = maxLoopCount
			}

			// Build: for (i = 0; i < loopCount; i = i + 1) {
			//          switch (i) { case -999: ; default: break }
			//          count = count + 1
			//        }
			// case -999 will never match any valid loop index, so default always executes
			forOp := opcode.OpCode{
				Cmd: opcode.For,
				Args: []any{
					// init: i = 0
					[]opcode.OpCode{
						{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
					},
					// condition: i < loopCount
					opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), loopCount}},
					// post: i = i + 1
					[]opcode.OpCode{
						{Cmd: opcode.Assign, Args: []any{
							opcode.Variable("i"),
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
						}},
					},
					// body: switch(i) { case -999: ; default: break } count = count + 1
					[]opcode.OpCode{
						{
							Cmd: opcode.Switch,
							Args: []any{
								opcode.Variable("i"),
								[]any{
									map[string]any{
										"value": int64(-999),
										"body":  []opcode.OpCode{},
									},
								},
								// default block with break
								[]opcode.OpCode{
									{Cmd: opcode.Break, Args: []any{}},
								},
							},
						},
						{Cmd: opcode.Assign, Args: []any{
							opcode.Variable("count"),
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("count"), int64(1)}},
						}},
					},
				},
			}

			vm := New([]opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("count"), int64(0)}},
				forOp,
			})
			err := vm.Run()
			if err != nil {
				return false
			}

			val, exists := vm.GetGlobalScope().Get("count")
			if !exists {
				return false
			}
			return val == loopCount
		},
		gen.Int64Range(1, 20),
	))

	// Property: code after break in switch case body is skipped, but loop continues.
	// This verifies both that break exits the switch AND that the loop is unaffected.
	properties.Property("code after break in switch is skipped but loop iteration count is correct", prop.ForAll(
		func(loopCount int64, matchIndex int64) bool {
			if loopCount < 2 {
				return true
			}
			const maxLoopCount = 15
			if loopCount > maxLoopCount {
				loopCount = maxLoopCount
			}
			// matchIndex is the loop iteration where the switch case matches
			if matchIndex < 0 || matchIndex >= loopCount {
				matchIndex = matchIndex % loopCount
				if matchIndex < 0 {
					matchIndex += loopCount
				}
			}

			// Build: for (i = 0; i < loopCount; i = i + 1) {
			//          switch (i) {
			//            case matchIndex:
			//              before_break = before_break + 1
			//              break
			//              after_break = after_break + 1  // should be skipped
			//            default:
			//              // empty
			//          }
			//          count = count + 1
			//        }
			forOp := opcode.OpCode{
				Cmd: opcode.For,
				Args: []any{
					// init: i = 0
					[]opcode.OpCode{
						{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
					},
					// condition: i < loopCount
					opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), loopCount}},
					// post: i = i + 1
					[]opcode.OpCode{
						{Cmd: opcode.Assign, Args: []any{
							opcode.Variable("i"),
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
						}},
					},
					// body
					[]opcode.OpCode{
						{
							Cmd: opcode.Switch,
							Args: []any{
								opcode.Variable("i"),
								[]any{
									map[string]any{
										"value": matchIndex,
										"body": []opcode.OpCode{
											{Cmd: opcode.Assign, Args: []any{
												opcode.Variable("before_break"),
												opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("before_break"), int64(1)}},
											}},
											{Cmd: opcode.Break, Args: []any{}},
											{Cmd: opcode.Assign, Args: []any{
												opcode.Variable("after_break"),
												opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("after_break"), int64(1)}},
											}},
										},
									},
								},
								// empty default
								[]opcode.OpCode{},
							},
						},
						{Cmd: opcode.Assign, Args: []any{
							opcode.Variable("count"),
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("count"), int64(1)}},
						}},
					},
				},
			}

			vm := New([]opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("count"), int64(0)}},
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("before_break"), int64(0)}},
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("after_break"), int64(0)}},
				forOp,
			})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Loop should complete all iterations
			countVal, _ := vm.GetGlobalScope().Get("count")
			if countVal != loopCount {
				return false
			}

			// before_break should be 1 (executed once when i == matchIndex)
			beforeVal, _ := vm.GetGlobalScope().Get("before_break")
			if beforeVal != int64(1) {
				return false
			}

			// after_break should remain 0 (skipped due to break)
			afterVal, _ := vm.GetGlobalScope().Get("after_break")
			if afterVal != int64(0) {
				return false
			}

			return true
		},
		gen.Int64Range(2, 15),
		gen.Int64Range(0, 14),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestPropertySwitch3_CaseValueMatchingAccuracy tests that the correct case is selected
// for integer and string switch values.
// **Validates: Requirements 2.1, 2.2**
// Feature: switch-edge-cases, Property 3: case値マッチングの正確性
func TestPropertySwitch3_CaseValueMatchingAccuracy(t *testing.T) {
	const minSuccessfulTests = 100

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = minSuccessfulTests

	properties := gopter.NewProperties(parameters)

	// Property: For any integer switch value, when a case value equals the switch value,
	// that case's body executes and sets the result variable correctly.
	properties.Property("integer switch value selects the correct case", prop.ForAll(
		func(switchVal int64, caseVals []int64) bool {
			// Need at least 2 cases to meaningfully test selection
			if len(caseVals) < 2 {
				return true
			}
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Deduplicate case values to avoid ambiguous matching
			seen := make(map[int64]bool)
			uniqueCaseVals := make([]int64, 0, len(caseVals))
			for _, v := range caseVals {
				if !seen[v] {
					seen[v] = true
					uniqueCaseVals = append(uniqueCaseVals, v)
				}
			}
			if len(uniqueCaseVals) < 2 {
				return true
			}
			caseVals = uniqueCaseVals

			// Build case clauses: each case body assigns result = caseIndex
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{opcode.Variable("result"), int64(i)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Determine which case should have matched (first match wins)
			expectedIndex := int64(-1)
			for i, cv := range caseVals {
				if cv == switchVal {
					expectedIndex = int64(i)
					break
				}
			}

			val, exists := vm.GetGlobalScope().Get("result")
			if expectedIndex == -1 {
				// No case should match; result should not be set
				return !exists
			}
			// Matched case should have set result to its index
			if !exists {
				return false
			}
			return val == expectedIndex
		},
		gen.Int64Range(-50, 50),
		gen.SliceOfN(8, gen.Int64Range(-50, 50)),
	))

	// Property: For any string switch value, when a case value equals the switch value,
	// that case's body executes and sets the result variable correctly.
	properties.Property("string switch value selects the correct case", prop.ForAll(
		func(switchVal string, caseVals []string) bool {
			// Need at least 2 cases
			if len(caseVals) < 2 {
				return true
			}
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Deduplicate case values
			seen := make(map[string]bool)
			uniqueCaseVals := make([]string, 0, len(caseVals))
			for _, v := range caseVals {
				if !seen[v] {
					seen[v] = true
					uniqueCaseVals = append(uniqueCaseVals, v)
				}
			}
			if len(uniqueCaseVals) < 2 {
				return true
			}
			caseVals = uniqueCaseVals

			// Build case clauses: each case body assigns result = caseIndex
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{opcode.Variable("result"), int64(i)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Determine which case should have matched (first match wins)
			expectedIndex := int64(-1)
			for i, cv := range caseVals {
				if cv == switchVal {
					expectedIndex = int64(i)
					break
				}
			}

			val, exists := vm.GetGlobalScope().Get("result")
			if expectedIndex == -1 {
				return !exists
			}
			if !exists {
				return false
			}
			return val == expectedIndex
		},
		gen.AlphaString(),
		gen.SliceOfN(8, gen.AlphaString()),
	))

	// Property: Among multiple cases, the first matching case wins.
	// When the switch value appears at multiple positions (after dedup it won't,
	// but we test by ensuring the first case with the matching value is selected).
	properties.Property("first matching case wins among multiple integer cases", prop.ForAll(
		func(switchVal int64, numCasesBefore int64, numCasesAfter int64) bool {
			// Constrain to reasonable sizes
			if numCasesBefore < 0 {
				numCasesBefore = 0
			}
			if numCasesBefore > 5 {
				numCasesBefore = 5
			}
			if numCasesAfter < 1 {
				numCasesAfter = 1
			}
			if numCasesAfter > 5 {
				numCasesAfter = 5
			}

			// Build cases: non-matching cases before, then the matching case, then non-matching after
			totalCases := int(numCasesBefore) + 1 + int(numCasesAfter)
			cases := make([]any, totalCases)
			matchIdx := int(numCasesBefore)

			for i := 0; i < totalCases; i++ {
				var caseVal int64
				if i == matchIdx {
					caseVal = switchVal
				} else {
					// Use a value guaranteed to not equal switchVal
					caseVal = switchVal + int64(i) + 1
				}
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{opcode.Variable("result"), int64(i)},
					},
				}
				cases[i] = map[string]any{
					"value": caseVal,
					"body":  caseBody,
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// The matching case at matchIdx should have been selected
			val, exists := vm.GetGlobalScope().Get("result")
			if !exists {
				return false
			}
			return val == int64(matchIdx)
		},
		gen.Int64Range(-100, 100),
		gen.Int64Range(0, 5),
		gen.Int64Range(1, 5),
	))

	// Property: First matching case wins for string values.
	properties.Property("first matching case wins among multiple string cases", prop.ForAll(
		func(switchVal string, numCasesBefore int64, numCasesAfter int64) bool {
			if numCasesBefore < 0 {
				numCasesBefore = 0
			}
			if numCasesBefore > 5 {
				numCasesBefore = 5
			}
			if numCasesAfter < 1 {
				numCasesAfter = 1
			}
			if numCasesAfter > 5 {
				numCasesAfter = 5
			}

			totalCases := int(numCasesBefore) + 1 + int(numCasesAfter)
			cases := make([]any, totalCases)
			matchIdx := int(numCasesBefore)

			for i := 0; i < totalCases; i++ {
				var caseVal string
				if i == matchIdx {
					caseVal = switchVal
				} else {
					// Use a value guaranteed to not equal switchVal
					caseVal = fmt.Sprintf("%s_nonmatch_%d", switchVal, i)
				}
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{opcode.Variable("result"), int64(i)},
					},
				}
				cases[i] = map[string]any{
					"value": caseVal,
					"body":  caseBody,
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			val, exists := vm.GetGlobalScope().Get("result")
			if !exists {
				return false
			}
			return val == int64(matchIdx)
		},
		gen.AlphaString(),
		gen.Int64Range(0, 5),
		gen.Int64Range(1, 5),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}


// TestPropertySwitch4_DefaultFallbackAccuracy tests that when no case matches,
// the default block executes if present, and nothing executes if absent.
// **Validates: Requirements 2.3, 2.4**
// Feature: switch-edge-cases, Property 4: defaultフォールバックの正確性
func TestPropertySwitch4_DefaultFallbackAccuracy(t *testing.T) {
	const minSuccessfulTests = 100

	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = minSuccessfulTests

	properties := gopter.NewProperties(parameters)

	// Property: When switch value does not match any case and a default block exists,
	// the default block executes.
	properties.Property("default block executes when no case matches", prop.ForAll(
		func(switchVal int64, caseVals []int64, defaultResult int64) bool {
			// Need at least 1 case to meaningfully test
			if len(caseVals) < 1 {
				return true
			}
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Ensure switchVal does NOT match any case value
			for i := range caseVals {
				if caseVals[i] == switchVal {
					caseVals[i] = switchVal + int64(i) + 100
				}
			}

			// Build case clauses: each case body assigns result = caseIndex
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{opcode.Variable("result"), int64(i)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			// Build default block: assigns default_executed = 1 and result = defaultResult
			defaultBlock := []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("default_executed"), int64(1)},
				},
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("result"), defaultResult},
				},
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases, defaultBlock},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Default block should have executed
			defVal, defExists := vm.GetGlobalScope().Get("default_executed")
			if !defExists || defVal != int64(1) {
				return false
			}

			// Result should be the default result value
			resVal, resExists := vm.GetGlobalScope().Get("result")
			if !resExists {
				return false
			}
			return resVal == defaultResult
		},
		gen.Int64Range(-50, 50),
		gen.SliceOfN(5, gen.Int64Range(100, 200)),
		gen.Int64Range(-1000, 1000),
	))

	// Property: When switch value does not match any case and no default block exists,
	// nothing executes and switch exits cleanly.
	properties.Property("no execution when no case matches and no default", prop.ForAll(
		func(switchVal int64, caseVals []int64) bool {
			// Need at least 1 case
			if len(caseVals) < 1 {
				return true
			}
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Ensure switchVal does NOT match any case value
			for i := range caseVals {
				if caseVals[i] == switchVal {
					caseVals[i] = switchVal + int64(i) + 100
				}
			}

			// Build case clauses: each case body assigns marker_i = 1
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				markerVar := opcode.Variable(fmt.Sprintf("marker_%d", i))
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{markerVar, int64(1)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			// No default block — only 2 args
			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// No case markers should be set
			for i := range caseVals {
				markerName := fmt.Sprintf("marker_%d", i)
				_, exists := vm.GetGlobalScope().Get(markerName)
				if exists {
					return false
				}
			}

			// No default_executed marker should exist
			_, defExists := vm.GetGlobalScope().Get("default_executed")
			if defExists {
				return false
			}

			return true
		},
		gen.Int64Range(-50, 50),
		gen.SliceOfN(5, gen.Int64Range(100, 200)),
	))

	// Property: When switch value does not match any string case and a default block exists,
	// the default block executes.
	properties.Property("default block executes for unmatched string switch value", prop.ForAll(
		func(caseVals []string, defaultMarker int64) bool {
			if len(caseVals) < 1 {
				return true
			}
			const maxCases = 10
			if len(caseVals) > maxCases {
				caseVals = caseVals[:maxCases]
			}

			// Use a switch value guaranteed to not match any case
			switchVal := "__NOMATCH__"
			for _, cv := range caseVals {
				if cv == switchVal {
					switchVal = switchVal + "_x"
				}
			}

			// Build case clauses
			cases := make([]any, len(caseVals))
			for i, cv := range caseVals {
				caseBody := []opcode.OpCode{
					{
						Cmd:  opcode.Assign,
						Args: []any{opcode.Variable(fmt.Sprintf("case_%d", i)), int64(1)},
					},
				}
				cases[i] = map[string]any{
					"value": cv,
					"body":  caseBody,
				}
			}

			// Default block assigns default_executed = defaultMarker
			defaultBlock := []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("default_executed"), defaultMarker},
				},
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases, defaultBlock},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Default should have executed
			defVal, defExists := vm.GetGlobalScope().Get("default_executed")
			if !defExists || defVal != defaultMarker {
				return false
			}

			// No case body should have executed
			for i := range caseVals {
				caseName := fmt.Sprintf("case_%d", i)
				_, exists := vm.GetGlobalScope().Get(caseName)
				if exists {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(5, gen.AlphaString()),
		gen.Int64Range(1, 1000),
	))

	// Property: Default block with multiple statements executes all statements.
	properties.Property("default block executes all statements", prop.ForAll(
		func(switchVal int64, numStatements int) bool {
			if numStatements < 1 {
				numStatements = 1
			}
			const maxStatements = 10
			if numStatements > maxStatements {
				numStatements = maxStatements
			}

			// Build a single case that won't match
			cases := []any{
				map[string]any{
					"value": switchVal + 1, // guaranteed non-match
					"body": []opcode.OpCode{
						{
							Cmd:  opcode.Assign,
							Args: []any{opcode.Variable("case_executed"), int64(1)},
						},
					},
				},
			}

			// Build default block with numStatements assignments
			defaultBlock := make([]opcode.OpCode, numStatements)
			for i := 0; i < numStatements; i++ {
				defaultBlock[i] = opcode.OpCode{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable(fmt.Sprintf("default_stmt_%d", i)), int64(i)},
				}
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases, defaultBlock},
			}

			vm := New([]opcode.OpCode{switchOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Case should NOT have executed
			_, caseExists := vm.GetGlobalScope().Get("case_executed")
			if caseExists {
				return false
			}

			// All default statements should have executed
			for i := 0; i < numStatements; i++ {
				stmtName := fmt.Sprintf("default_stmt_%d", i)
				val, exists := vm.GetGlobalScope().Get(stmtName)
				if !exists || val != int64(i) {
					return false
				}
			}

			return true
		},
		gen.Int64Range(-50, 50),
		gen.IntRange(1, 10),
	))

	// Property: Code after switch continues executing regardless of default execution.
	properties.Property("code after switch executes after default fallback", prop.ForAll(
		func(switchVal int64, afterVal int64) bool {
			// Build a single case that won't match
			cases := []any{
				map[string]any{
					"value": switchVal + 1,
					"body": []opcode.OpCode{
						{
							Cmd:  opcode.Assign,
							Args: []any{opcode.Variable("case_executed"), int64(1)},
						},
					},
				},
			}

			// Default block
			defaultBlock := []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("default_executed"), int64(1)},
				},
			}

			switchOp := opcode.OpCode{
				Cmd:  opcode.Switch,
				Args: []any{switchVal, cases, defaultBlock},
			}

			// Code after switch
			afterOp := opcode.OpCode{
				Cmd:  opcode.Assign,
				Args: []any{opcode.Variable("after_switch"), afterVal},
			}

			vm := New([]opcode.OpCode{switchOp, afterOp})
			err := vm.Run()
			if err != nil {
				return false
			}

			// Default should have executed
			defVal, defExists := vm.GetGlobalScope().Get("default_executed")
			if !defExists || defVal != int64(1) {
				return false
			}

			// Code after switch should have executed
			afterSwitchVal, afterExists := vm.GetGlobalScope().Get("after_switch")
			if !afterExists {
				return false
			}
			return afterSwitchVal == afterVal
		},
		gen.Int64Range(-50, 50),
		gen.Int64Range(-1000, 1000),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
