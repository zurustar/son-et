package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/compiler"
)

// Feature: execution-engine, Property 19: エラー後の実行継続
// *任意の*致命的でないエラー（ゼロ除算、範囲外アクセス等）について、エラー発生後も実行が継続される
// **Validates: Requirements 11.8**
func TestProperty19_ExecutionContinuesAfterNonFatalError(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	// Property: Division by zero returns zero and execution continues
	// Requirement 11.3: When division by zero occurs, system logs error and returns zero.
	// Requirement 11.8: System continues execution after non-fatal errors.
	properties.Property("division by zero returns zero and execution continues", prop.ForAll(
		func(dividend int64) bool {
			vm := New(nil)

			// Create division by zero operation
			divOp := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"/", dividend, int64(0)},
			}

			result, err := vm.Execute(divOp)
			if err != nil {
				return false // Should not return error
			}

			// Result should be zero
			if i, ok := toInt64(result); ok {
				return i == 0
			}
			if f, ok := toFloat64(result); ok {
				return f == 0
			}
			return false
		},
		gen.Int64(),
	))

	// Property: Modulo by zero returns zero and execution continues
	properties.Property("modulo by zero returns zero and execution continues", prop.ForAll(
		func(dividend int64) bool {
			vm := New(nil)

			// Create modulo by zero operation
			modOp := compiler.OpCode{
				Cmd:  compiler.OpBinaryOp,
				Args: []any{"%", dividend, int64(0)},
			}

			result, err := vm.Execute(modOp)
			if err != nil {
				return false // Should not return error
			}

			// Result should be zero
			if i, ok := toInt64(result); ok {
				return i == 0
			}
			if f, ok := toFloat64(result); ok {
				return f == 0
			}
			return false
		},
		gen.Int64(),
	))

	// Property: Negative array index returns zero and execution continues
	// Requirement 11.4: When array index is out of range, system logs error and returns zero.
	// Requirement 19.4: When array index is negative, system logs error and returns zero.
	properties.Property("negative array index returns zero and execution continues", prop.ForAll(
		func(negIndex int64) bool {
			if negIndex >= 0 {
				negIndex = -negIndex - 1 // Ensure negative
			}

			vm := New(nil)

			// Set up an array
			arr := NewArray(5)
			for i := 0; i < 5; i++ {
				arr.Set(int64(i), int64(i*10))
			}
			vm.globalScope.Set("testArray", arr)

			// Create array access with negative index
			accessOp := compiler.OpCode{
				Cmd:  compiler.OpArrayAccess,
				Args: []any{compiler.Variable("testArray"), negIndex},
			}

			result, err := vm.Execute(accessOp)
			if err != nil {
				return false // Should not return error
			}

			// Result should be zero
			if i, ok := toInt64(result); ok {
				return i == 0
			}
			return false
		},
		gen.Int64Range(-1000, -1),
	))

	// Property: Out of range array index returns zero and execution continues
	properties.Property("out of range array index returns zero and execution continues", prop.ForAll(
		func(arraySize int, extraIndex int) bool {
			if arraySize < 1 {
				arraySize = 1
			}
			if arraySize > 100 {
				arraySize = 100
			}
			if extraIndex < 1 {
				extraIndex = 1
			}

			vm := New(nil)

			// Set up an array with specific size
			arr := NewArray(arraySize)
			for i := 0; i < arraySize; i++ {
				arr.Set(int64(i), int64(i*10))
			}
			vm.globalScope.Set("testArray", arr)

			// Access beyond array bounds
			outOfRangeIndex := int64(arraySize + extraIndex)
			accessOp := compiler.OpCode{
				Cmd:  compiler.OpArrayAccess,
				Args: []any{compiler.Variable("testArray"), outOfRangeIndex},
			}

			result, err := vm.Execute(accessOp)
			if err != nil {
				return false // Should not return error
			}

			// Result should be zero
			if i, ok := toInt64(result); ok {
				return i == 0
			}
			return false
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 100),
	))

	// Property: Undefined variable returns default value (zero) and execution continues
	// Requirement 11.5: When variable is not found, system creates it with default value.
	properties.Property("undefined variable returns default value and execution continues", prop.ForAll(
		func(varName string) bool {
			if varName == "" {
				varName = "testVar"
			}
			// Sanitize variable name
			varName = sanitizeVarName(varName)

			vm := New(nil)

			// Try to access undefined variable
			result, err := vm.evaluateValue(compiler.Variable(varName))
			if err != nil {
				return false // Should not return error
			}

			// Result should be zero (default value)
			if i, ok := toInt64(result); ok {
				return i == 0
			}
			return false
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	// Property: Unknown function call returns error
	// 未定義関数が呼ばれた場合はエラーで終了する
	properties.Property("unknown function call returns error", prop.ForAll(
		func(funcName string) bool {
			if funcName == "" {
				funcName = "unknownFunc"
			}
			// Sanitize function name
			funcName = sanitizeVarName(funcName)

			vm := New(nil)

			// Call unknown function
			callOp := compiler.OpCode{
				Cmd:  compiler.OpCall,
				Args: []any{funcName},
			}

			_, err := vm.Execute(callOp)
			// Should return error for undefined function
			return err != nil
		},
		gen.AlphaString().SuchThat(func(s string) bool { return len(s) > 0 && len(s) < 50 }),
	))

	// Property: Execution continues after multiple non-fatal errors
	properties.Property("execution continues after multiple non-fatal errors", prop.ForAll(
		func(errorCount int) bool {
			if errorCount < 1 {
				errorCount = 1
			}
			if errorCount > 10 {
				errorCount = 10
			}

			vm := New(nil)

			// Execute multiple operations that cause non-fatal errors
			for i := 0; i < errorCount; i++ {
				// Division by zero
				divOp := compiler.OpCode{
					Cmd:  compiler.OpBinaryOp,
					Args: []any{"/", int64(i + 1), int64(0)},
				}
				_, err := vm.Execute(divOp)
				if err != nil {
					return false
				}
			}

			// After all errors, VM should still be able to execute normal operations
			assignOp := compiler.OpCode{
				Cmd:  compiler.OpAssign,
				Args: []any{compiler.Variable("result"), int64(42)},
			}
			result, err := vm.Execute(assignOp)
			if err != nil {
				return false
			}

			// Verify the assignment worked
			if i, ok := toInt64(result); ok {
				return i == 42
			}
			return false
		},
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// sanitizeVarName ensures the variable name is valid
func sanitizeVarName(name string) string {
	if len(name) == 0 {
		return "var"
	}
	// Keep only alphanumeric characters
	result := make([]byte, 0, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9' && len(result) > 0) {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		return "var"
	}
	return string(result)
}
