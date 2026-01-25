package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/compiler"
)

// Property-based tests for Array operations.
// These tests verify the correctness properties defined in the design document.

// TestProperty21_ArrayAutoExpansion tests that arrays automatically expand
// when an element is assigned at an index beyond the current size.
// **Validates: Requirements 19.5**
// Feature: execution-engine, Property 21: 配列の自動拡張
func TestProperty21_ArrayAutoExpansion(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After assignment to index i, array size is at least i+1
	properties.Property("array size is at least index+1 after assignment", prop.ForAll(
		func(index int64, value int64) bool {
			// Limit index to reasonable range to avoid memory issues
			if index < 0 || index > 10000 {
				return true // Skip invalid indices
			}

			arr := NewArray(0) // Start with empty array
			arr.Set(index, value)

			// Array size should be at least index+1
			return arr.Len() >= int(index)+1
		},
		gen.Int64Range(0, 1000),
		gen.Int64(),
	))

	// Property: Assignment to index beyond current size expands array
	properties.Property("assignment beyond current size expands array", prop.ForAll(
		func(initialSize int, targetIndex int64) bool {
			// Ensure valid parameters
			if initialSize < 0 {
				initialSize = 0
			}
			if initialSize > 100 {
				initialSize = 100
			}
			if targetIndex < 0 || targetIndex > 1000 {
				return true // Skip invalid indices
			}

			arr := NewArray(initialSize)
			originalSize := arr.Len()

			// Assign to target index
			arr.Set(targetIndex, int64(999))

			// If target index was beyond original size, array should have expanded
			if int(targetIndex) >= originalSize {
				return arr.Len() >= int(targetIndex)+1
			}
			// If target index was within original size, size should be unchanged
			return arr.Len() == originalSize
		},
		gen.IntRange(0, 100),
		gen.Int64Range(0, 200),
	))

	// Property: Existing elements are preserved during expansion
	properties.Property("existing elements are preserved during expansion", prop.ForAll(
		func(initialValues []int64, targetIndex int64) bool {
			// Limit sizes
			if len(initialValues) > 50 {
				initialValues = initialValues[:50]
			}
			if targetIndex < 0 || targetIndex > 1000 {
				return true
			}

			// Create array with initial values
			elements := make([]any, len(initialValues))
			for i, v := range initialValues {
				elements[i] = v
			}
			arr := NewArrayFromSlice(elements)

			// Record original values
			originalValues := make([]int64, len(initialValues))
			for i := range initialValues {
				val, _ := arr.Get(int64(i))
				if v, ok := val.(int64); ok {
					originalValues[i] = v
				}
			}

			// Expand array by assigning to larger index
			expandIndex := int64(len(initialValues)) + targetIndex
			if expandIndex > 1000 {
				expandIndex = 1000
			}
			arr.Set(expandIndex, int64(12345))

			// Verify original values are preserved
			for i, expected := range originalValues {
				val, _ := arr.Get(int64(i))
				if v, ok := val.(int64); ok {
					if v != expected {
						return false
					}
				} else {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(20, gen.Int64()),
		gen.Int64Range(1, 100),
	))

	// Property: New elements are initialized to zero during expansion
	properties.Property("new elements are initialized to zero during expansion", prop.ForAll(
		func(initialSize int, targetIndex int64) bool {
			// Ensure valid parameters
			if initialSize < 0 {
				initialSize = 0
			}
			if initialSize > 50 {
				initialSize = 50
			}
			if targetIndex < int64(initialSize) {
				targetIndex = int64(initialSize) + 10
			}
			if targetIndex > 200 {
				targetIndex = 200
			}

			arr := NewArray(initialSize)

			// Assign to target index (causes expansion)
			arr.Set(targetIndex, int64(999))

			// Check that elements between initialSize and targetIndex are zero
			for i := initialSize; i < int(targetIndex); i++ {
				val, _ := arr.Get(int64(i))
				if val != int64(0) {
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 50),
		gen.Int64Range(51, 150),
	))

	// Property: Array expansion via VM OpArrayAssign works correctly
	properties.Property("VM OpArrayAssign expands array correctly", prop.ForAll(
		func(initialSize int, targetIndex int64, value int64) bool {
			// Ensure valid parameters
			if initialSize < 0 {
				initialSize = 0
			}
			if initialSize > 50 {
				initialSize = 50
			}
			if targetIndex < 0 || targetIndex > 200 {
				return true
			}

			vm := New([]compiler.OpCode{})

			// Create initial array
			if initialSize > 0 {
				elements := make([]any, initialSize)
				for i := range elements {
					elements[i] = int64(i)
				}
				vm.GetCurrentScope().Set("arr", NewArrayFromSlice(elements))
			}

			// Execute OpArrayAssign
			opcode := compiler.OpCode{
				Cmd:  compiler.OpArrayAssign,
				Args: []any{compiler.Variable("arr"), targetIndex, value},
			}

			_, err := vm.executeArrayAssign(opcode)
			if err != nil {
				return false
			}

			// Verify array was created/expanded
			arrVal, ok := vm.GetCurrentScope().Get("arr")
			if !ok {
				return false
			}
			arr, ok := arrVal.(*Array)
			if !ok {
				return false
			}

			// Array size should be at least targetIndex+1
			if arr.Len() < int(targetIndex)+1 {
				return false
			}

			// Value should be set correctly
			elem, _ := arr.Get(targetIndex)
			return elem == value
		},
		gen.IntRange(0, 50),
		gen.Int64Range(0, 100),
		gen.Int64(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty22_ArrayPassByReference tests that arrays are passed by reference
// to functions, so modifications in the function affect the original array.
// **Validates: Requirements 19.8**
// Feature: execution-engine, Property 22: 配列の参照渡し
func TestProperty22_ArrayPassByReference(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: Modifications in function affect original array
	properties.Property("modifications in function affect original array", prop.ForAll(
		func(initialValues []int64, modifyIndex int64, newValue int64) bool {
			// Ensure valid parameters
			if len(initialValues) == 0 {
				initialValues = []int64{0}
			}
			if len(initialValues) > 50 {
				initialValues = initialValues[:50]
			}
			if modifyIndex < 0 {
				modifyIndex = 0
			}
			if modifyIndex >= int64(len(initialValues)) {
				modifyIndex = int64(len(initialValues)) - 1
			}

			vm := New([]compiler.OpCode{})

			// Create array in global scope
			elements := make([]any, len(initialValues))
			for i, v := range initialValues {
				elements[i] = v
			}
			vm.GetGlobalScope().Set("testArr", NewArrayFromSlice(elements))

			// Define function that modifies array element
			vm.functions["modifyElement"] = &FunctionDef{
				Name: "modifyElement",
				Parameters: []FunctionParam{
					{Name: "arr", Type: "int", IsArray: true},
					{Name: "idx", Type: "int", IsArray: false},
					{Name: "val", Type: "int", IsArray: false},
				},
				Body: []compiler.OpCode{
					{
						Cmd:  compiler.OpArrayAssign,
						Args: []any{compiler.Variable("arr"), compiler.Variable("idx"), compiler.Variable("val")},
					},
				},
			}

			// Call function with array
			opcode := compiler.OpCode{
				Cmd:  compiler.OpCall,
				Args: []any{"modifyElement", compiler.Variable("testArr"), modifyIndex, newValue},
			}

			_, err := vm.executeCall(opcode)
			if err != nil {
				return false
			}

			// Original array should be modified
			arrVal, _ := vm.GetGlobalScope().Get("testArr")
			arr := arrVal.(*Array)
			elem, _ := arr.Get(modifyIndex)

			return elem == newValue
		},
		gen.SliceOfN(10, gen.Int64()),
		gen.Int64Range(0, 9),
		gen.Int64(),
	))

	// Property: Array expansion in function affects original array
	properties.Property("array expansion in function affects original array", prop.ForAll(
		func(initialSize int, expandIndex int64, value int64) bool {
			// Ensure valid parameters
			if initialSize < 1 {
				initialSize = 1
			}
			if initialSize > 20 {
				initialSize = 20
			}
			if expandIndex < int64(initialSize) {
				expandIndex = int64(initialSize) + 5
			}
			if expandIndex > 100 {
				expandIndex = 100
			}

			vm := New([]compiler.OpCode{})

			// Create small array in global scope
			elements := make([]any, initialSize)
			for i := range elements {
				elements[i] = int64(i)
			}
			vm.GetGlobalScope().Set("smallArr", NewArrayFromSlice(elements))

			// Define function that expands array
			vm.functions["expandArray"] = &FunctionDef{
				Name: "expandArray",
				Parameters: []FunctionParam{
					{Name: "arr", Type: "int", IsArray: true},
					{Name: "idx", Type: "int", IsArray: false},
					{Name: "val", Type: "int", IsArray: false},
				},
				Body: []compiler.OpCode{
					{
						Cmd:  compiler.OpArrayAssign,
						Args: []any{compiler.Variable("arr"), compiler.Variable("idx"), compiler.Variable("val")},
					},
				},
			}

			// Call function with array
			opcode := compiler.OpCode{
				Cmd:  compiler.OpCall,
				Args: []any{"expandArray", compiler.Variable("smallArr"), expandIndex, value},
			}

			_, err := vm.executeCall(opcode)
			if err != nil {
				return false
			}

			// Original array should be expanded
			arrVal, _ := vm.GetGlobalScope().Get("smallArr")
			arr := arrVal.(*Array)

			// Array should have expanded
			if arr.Len() < int(expandIndex)+1 {
				return false
			}

			// Value should be set
			elem, _ := arr.Get(expandIndex)
			return elem == value
		},
		gen.IntRange(1, 20),
		gen.Int64Range(21, 50),
		gen.Int64(),
	))

	// Property: Nested function calls preserve reference
	properties.Property("nested function calls preserve array reference", prop.ForAll(
		func(value1 int64, value2 int64) bool {
			vm := New([]compiler.OpCode{})

			// Create array in global scope
			vm.GetGlobalScope().Set("nestedArr", NewArrayFromSlice([]any{int64(0), int64(0), int64(0)}))

			// Define inner function
			vm.functions["innerModify"] = &FunctionDef{
				Name: "innerModify",
				Parameters: []FunctionParam{
					{Name: "arr", Type: "int", IsArray: true},
					{Name: "val", Type: "int", IsArray: false},
				},
				Body: []compiler.OpCode{
					{
						Cmd:  compiler.OpArrayAssign,
						Args: []any{compiler.Variable("arr"), int64(2), compiler.Variable("val")},
					},
				},
			}

			// Define outer function that calls inner
			vm.functions["outerModify"] = &FunctionDef{
				Name: "outerModify",
				Parameters: []FunctionParam{
					{Name: "arr", Type: "int", IsArray: true},
					{Name: "val1", Type: "int", IsArray: false},
					{Name: "val2", Type: "int", IsArray: false},
				},
				Body: []compiler.OpCode{
					{
						Cmd:  compiler.OpArrayAssign,
						Args: []any{compiler.Variable("arr"), int64(1), compiler.Variable("val1")},
					},
					{
						Cmd:  compiler.OpCall,
						Args: []any{"innerModify", compiler.Variable("arr"), compiler.Variable("val2")},
					},
				},
			}

			// Call outer function
			opcode := compiler.OpCode{
				Cmd:  compiler.OpCall,
				Args: []any{"outerModify", compiler.Variable("nestedArr"), value1, value2},
			}

			_, err := vm.executeCall(opcode)
			if err != nil {
				return false
			}

			// Original array should have both modifications
			arrVal, _ := vm.GetGlobalScope().Get("nestedArr")
			arr := arrVal.(*Array)

			elem1, _ := arr.Get(1)
			elem2, _ := arr.Get(2)

			return elem1 == value1 && elem2 == value2
		},
		gen.Int64(),
		gen.Int64(),
	))

	// Property: Multiple functions modifying same array all affect original
	properties.Property("multiple functions modifying same array all affect original", prop.ForAll(
		func(values []int64) bool {
			// Limit values
			if len(values) == 0 {
				return true
			}
			if len(values) > 10 {
				values = values[:10]
			}

			vm := New([]compiler.OpCode{})

			// Create array in global scope
			elements := make([]any, len(values))
			for i := range elements {
				elements[i] = int64(0)
			}
			vm.GetGlobalScope().Set("multiArr", NewArrayFromSlice(elements))

			// Define function that sets a specific index
			vm.functions["setElement"] = &FunctionDef{
				Name: "setElement",
				Parameters: []FunctionParam{
					{Name: "arr", Type: "int", IsArray: true},
					{Name: "idx", Type: "int", IsArray: false},
					{Name: "val", Type: "int", IsArray: false},
				},
				Body: []compiler.OpCode{
					{
						Cmd:  compiler.OpArrayAssign,
						Args: []any{compiler.Variable("arr"), compiler.Variable("idx"), compiler.Variable("val")},
					},
				},
			}

			// Call function multiple times with different indices
			for i, val := range values {
				opcode := compiler.OpCode{
					Cmd:  compiler.OpCall,
					Args: []any{"setElement", compiler.Variable("multiArr"), int64(i), val},
				}

				_, err := vm.executeCall(opcode)
				if err != nil {
					return false
				}
			}

			// Verify all values were set
			arrVal, _ := vm.GetGlobalScope().Get("multiArr")
			arr := arrVal.(*Array)

			for i, expected := range values {
				elem, _ := arr.Get(int64(i))
				if elem != expected {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(10, gen.Int64()),
	))

	// Property: Array reference is maintained across recursive calls
	properties.Property("array reference is maintained across recursive calls", prop.ForAll(
		func(depth int, value int64) bool {
			// Limit recursion depth
			if depth < 1 {
				depth = 1
			}
			if depth > 20 {
				depth = 20
			}

			vm := New([]compiler.OpCode{})

			// Create array in global scope
			elements := make([]any, depth)
			for i := range elements {
				elements[i] = int64(0)
			}
			vm.GetGlobalScope().Set("recursiveArr", NewArrayFromSlice(elements))

			// Define recursive function that modifies array at each level
			// recursiveModify(arr, depth, val) {
			//   if (depth > 0) {
			//     arr[depth-1] = val
			//     recursiveModify(arr, depth-1, val)
			//   }
			// }
			vm.functions["recursiveModify"] = &FunctionDef{
				Name: "recursiveModify",
				Parameters: []FunctionParam{
					{Name: "arr", Type: "int", IsArray: true},
					{Name: "d", Type: "int", IsArray: false},
					{Name: "val", Type: "int", IsArray: false},
				},
				Body: []compiler.OpCode{
					{
						Cmd: compiler.OpIf,
						Args: []any{
							// condition: d > 0
							compiler.OpCode{
								Cmd:  compiler.OpBinaryOp,
								Args: []any{">", compiler.Variable("d"), int64(0)},
							},
							// then block
							[]compiler.OpCode{
								// arr[d-1] = val
								{
									Cmd: compiler.OpArrayAssign,
									Args: []any{
										compiler.Variable("arr"),
										compiler.OpCode{
											Cmd:  compiler.OpBinaryOp,
											Args: []any{"-", compiler.Variable("d"), int64(1)},
										},
										compiler.Variable("val"),
									},
								},
								// recursiveModify(arr, d-1, val)
								{
									Cmd: compiler.OpCall,
									Args: []any{
										"recursiveModify",
										compiler.Variable("arr"),
										compiler.OpCode{
											Cmd:  compiler.OpBinaryOp,
											Args: []any{"-", compiler.Variable("d"), int64(1)},
										},
										compiler.Variable("val"),
									},
								},
							},
							// else block (empty)
							[]compiler.OpCode{},
						},
					},
				},
			}

			// Call recursive function
			opcode := compiler.OpCode{
				Cmd:  compiler.OpCall,
				Args: []any{"recursiveModify", compiler.Variable("recursiveArr"), int64(depth), value},
			}

			_, err := vm.executeCall(opcode)
			if err != nil {
				return false
			}

			// All elements should be set to value
			arrVal, _ := vm.GetGlobalScope().Get("recursiveArr")
			arr := arrVal.(*Array)

			for i := 0; i < depth; i++ {
				elem, _ := arr.Get(int64(i))
				if elem != value {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 15),
		gen.Int64(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
