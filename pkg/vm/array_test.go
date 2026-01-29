package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestArrayDynamicAllocation tests dynamic array allocation.
// Requirement 19.1: When array is declared, system allocates storage for array.
func TestArrayDynamicAllocation(t *testing.T) {
	t.Run("creates array on first assignment", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// arr[0] = 42
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(0), int64(42)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, ok := vm.GetCurrentScope().Get("arr")
		if !ok {
			t.Fatal("expected array 'arr' to be created")
		}
		arr, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected *Array, got %T", val)
		}
		if arr.Len() < 1 {
			t.Errorf("expected array length >= 1, got %d", arr.Len())
		}
	})

	t.Run("creates array with correct initial size", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// arr[9] = 100 (should create array of size 10)
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(9), int64(100)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)
		if arr.Len() != 10 {
			t.Errorf("expected array length 10, got %d", arr.Len())
		}
	})
}

// TestArrayAutoExpansion tests automatic array expansion.
// Requirement 19.5: When array index exceeds array size, system automatically expands array.
// Requirement 19.6: System supports dynamic array resizing.
func TestArrayAutoExpansion(t *testing.T) {
	t.Run("expands array when index exceeds size", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Create initial array with 3 elements
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(1), int64(2), int64(3)}))

		// Assign to index 10 (should expand)
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(10), int64(999)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)
		if arr.Len() != 11 {
			t.Errorf("expected array length 11, got %d", arr.Len())
		}
		elem, _ := arr.Get(10)
		if elem != int64(999) {
			t.Errorf("expected arr[10] = 999, got %v", elem)
		}
	})

	t.Run("preserves existing elements during expansion", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Create initial array
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20), int64(30)}))

		// Expand array
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(5), int64(60)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)

		// Original elements should be preserved
		elem0, _ := arr.Get(0)
		elem1, _ := arr.Get(1)
		elem2, _ := arr.Get(2)
		if elem0 != int64(10) {
			t.Errorf("expected arr[0] = 10, got %v", elem0)
		}
		if elem1 != int64(20) {
			t.Errorf("expected arr[1] = 20, got %v", elem1)
		}
		if elem2 != int64(30) {
			t.Errorf("expected arr[2] = 30, got %v", elem2)
		}
	})
}

// TestArrayZeroInitialization tests zero initialization of new elements.
// Requirement 19.7: System initializes new array elements to zero.
func TestArrayZeroInitialization(t *testing.T) {
	t.Run("initializes new elements to zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Create array by assigning to index 5
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(5), int64(100)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)

		// Elements 0-4 should be zero
		for i := 0; i < 5; i++ {
			elem, _ := arr.Get(int64(i))
			if elem != int64(0) {
				t.Errorf("expected arr[%d] = 0, got %v", i, elem)
			}
		}
	})

	t.Run("initializes expanded elements to zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Create initial array
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(1), int64(2)}))

		// Expand array
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(5), int64(100)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)

		// New elements (2-4) should be zero
		for i := 2; i < 5; i++ {
			elem, _ := arr.Get(int64(i))
			if elem != int64(0) {
				t.Errorf("expected arr[%d] = 0, got %v", i, elem)
			}
		}
	})
}

// TestArrayPassByReference tests array pass-by-reference in function calls.
// Requirement 19.8: When array is passed to function, system passes it by reference.
func TestArrayPassByReference(t *testing.T) {
	t.Run("modifications in function affect original array", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create array in global scope using Array type
		vm.GetGlobalScope().Set("myArr", NewArrayFromSlice([]any{int64(1), int64(2), int64(3)}))

		// Define function that modifies array element
		// func modifyArray(arr[]) { arr[0] = 999 }
		vm.functions["modifyArray"] = &FunctionDef{
			Name: "modifyArray",
			Parameters: []FunctionParam{
				{Name: "arr", Type: "int", IsArray: true},
			},
			Body: []opcode.OpCode{
				{
					Cmd:  opcode.ArrayAssign,
					Args: []any{opcode.Variable("arr"), int64(0), int64(999)},
				},
			},
		}

		// Call function with array
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"modifyArray", opcode.Variable("myArr")},
		}

		_, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Original array should be modified
		val, _ := vm.GetGlobalScope().Get("myArr")
		arr := val.(*Array)
		elem, _ := arr.Get(0)
		if elem != int64(999) {
			t.Errorf("expected myArr[0] = 999 after function call, got %v", elem)
		}
	})

	t.Run("array expansion in function affects original", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create small array in global scope using Array type
		vm.GetGlobalScope().Set("smallArr", NewArrayFromSlice([]any{int64(1), int64(2)}))

		// Define function that expands array
		// func expandArray(arr[]) { arr[10] = 100 }
		vm.functions["expandArray"] = &FunctionDef{
			Name: "expandArray",
			Parameters: []FunctionParam{
				{Name: "arr", Type: "int", IsArray: true},
			},
			Body: []opcode.OpCode{
				{
					Cmd:  opcode.ArrayAssign,
					Args: []any{opcode.Variable("arr"), int64(10), int64(100)},
				},
			},
		}

		// Call function with array
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"expandArray", opcode.Variable("smallArr")},
		}

		_, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Original array should be expanded
		val, _ := vm.GetGlobalScope().Get("smallArr")
		arr := val.(*Array)
		if arr.Len() < 11 {
			t.Errorf("expected array length >= 11, got %d", arr.Len())
		}
		elem, _ := arr.Get(10)
		if elem != int64(100) {
			t.Errorf("expected smallArr[10] = 100, got %v", elem)
		}
	})

	t.Run("nested function calls preserve reference", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create array in global scope using Array type
		vm.GetGlobalScope().Set("nestedArr", NewArrayFromSlice([]any{int64(0), int64(0), int64(0)}))

		// Define inner function
		// func innerModify(arr[]) { arr[2] = 333 }
		vm.functions["innerModify"] = &FunctionDef{
			Name: "innerModify",
			Parameters: []FunctionParam{
				{Name: "arr", Type: "int", IsArray: true},
			},
			Body: []opcode.OpCode{
				{
					Cmd:  opcode.ArrayAssign,
					Args: []any{opcode.Variable("arr"), int64(2), int64(333)},
				},
			},
		}

		// Define outer function that calls inner
		// func outerModify(arr[]) { arr[1] = 222; innerModify(arr) }
		vm.functions["outerModify"] = &FunctionDef{
			Name: "outerModify",
			Parameters: []FunctionParam{
				{Name: "arr", Type: "int", IsArray: true},
			},
			Body: []opcode.OpCode{
				{
					Cmd:  opcode.ArrayAssign,
					Args: []any{opcode.Variable("arr"), int64(1), int64(222)},
				},
				{
					Cmd:  opcode.Call,
					Args: []any{"innerModify", opcode.Variable("arr")},
				},
			},
		}

		// Call outer function
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"outerModify", opcode.Variable("nestedArr")},
		}

		_, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Original array should have both modifications
		val, _ := vm.GetGlobalScope().Get("nestedArr")
		arr := val.(*Array)
		elem1, _ := arr.Get(1)
		elem2, _ := arr.Get(2)
		if elem1 != int64(222) {
			t.Errorf("expected nestedArr[1] = 222, got %v", elem1)
		}
		if elem2 != int64(333) {
			t.Errorf("expected nestedArr[2] = 333, got %v", elem2)
		}
	})
}

// TestArrayNegativeIndex tests negative index handling.
// Requirement 19.4: When array index is negative, system logs error and returns zero.
func TestArrayNegativeIndex(t *testing.T) {
	t.Run("negative index on access returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20), int64(30)}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAccess,
			Args: []any{opcode.Variable("arr"), int64(-1)},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for negative index, got %v", result)
		}
	})

	t.Run("negative index on assign returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20), int64(30)}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(-5), int64(999)},
		}

		result, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for negative index, got %v", result)
		}

		// Array should be unchanged
		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)
		if arr.Len() != 3 {
			t.Errorf("expected array length 3, got %d", arr.Len())
		}
	})
}

// TestArrayElementAccess tests array element access.
// Requirement 19.2: When array element is accessed, system returns value at specified index.
func TestArrayElementAccess(t *testing.T) {
	t.Run("accesses elements at various indices", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(100), int64(200), int64(300), int64(400), int64(500)}))

		testCases := []struct {
			index    int64
			expected int64
		}{
			{0, 100},
			{1, 200},
			{2, 300},
			{3, 400},
			{4, 500},
		}

		for _, tc := range testCases {
			opcode := opcode.OpCode{
				Cmd:  opcode.ArrayAccess,
				Args: []any{opcode.Variable("arr"), tc.index},
			}

			result, err := vm.executeArrayAccess(opcode)
			if err != nil {
				t.Fatalf("unexpected error at index %d: %v", tc.index, err)
			}
			if result != tc.expected {
				t.Errorf("arr[%d]: expected %d, got %v", tc.index, tc.expected, result)
			}
		}
	})

	t.Run("accesses string elements", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("names", NewArrayFromSlice([]any{"Alice", "Bob", "Charlie"}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAccess,
			Args: []any{opcode.Variable("names"), int64(1)},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Bob" {
			t.Errorf("expected 'Bob', got %v", result)
		}
	})
}

// TestArrayElementAssignment tests array element assignment.
// Requirement 19.3: When array element is assigned, system stores value at specified index.
func TestArrayElementAssignment(t *testing.T) {
	t.Run("assigns to existing element", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(1), int64(2), int64(3)}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(1), int64(999)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)
		elem, _ := arr.Get(1)
		if elem != int64(999) {
			t.Errorf("expected arr[1] = 999, got %v", elem)
		}
	})

	t.Run("assigns different types", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("mixed", NewArrayFromSlice([]any{nil, nil, nil}))

		// Assign integer
		vm.executeArrayAssign(opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("mixed"), int64(0), int64(42)},
		})

		// Assign string
		vm.executeArrayAssign(opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("mixed"), int64(1), "hello"},
		})

		// Assign float
		vm.executeArrayAssign(opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("mixed"), int64(2), float64(3.14)},
		})

		val, _ := vm.GetCurrentScope().Get("mixed")
		arr := val.(*Array)

		elem0, _ := arr.Get(0)
		elem1, _ := arr.Get(1)
		elem2, _ := arr.Get(2)

		if elem0 != int64(42) {
			t.Errorf("expected arr[0] = 42, got %v", elem0)
		}
		if elem1 != "hello" {
			t.Errorf("expected arr[1] = 'hello', got %v", elem1)
		}
		if elem2 != float64(3.14) {
			t.Errorf("expected arr[2] = 3.14, got %v", elem2)
		}
	})
}

// TestArrayWithExpressionIndex tests array access with expression indices.
func TestArrayWithExpressionIndex(t *testing.T) {
	t.Run("accesses with variable index", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20), int64(30)}))
		vm.GetCurrentScope().Set("i", int64(2))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAccess,
			Args: []any{opcode.Variable("arr"), opcode.Variable("i")},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(30) {
			t.Errorf("expected 30, got %v", result)
		}
	})

	t.Run("accesses with computed index", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(100), int64(200), int64(300), int64(400)}))

		// arr[1 + 2] = arr[3]
		opcode := opcode.OpCode{
			Cmd: opcode.ArrayAccess,
			Args: []any{
				opcode.Variable("arr"),
				opcode.OpCode{
					Cmd:  opcode.BinaryOp,
					Args: []any{"+", int64(1), int64(2)},
				},
			},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(400) {
			t.Errorf("expected 400, got %v", result)
		}
	})

	t.Run("assigns with computed index", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(0), int64(0), int64(0), int64(0)}))
		vm.GetCurrentScope().Set("offset", int64(2))

		// arr[offset + 1] = 999
		opcode := opcode.OpCode{
			Cmd: opcode.ArrayAssign,
			Args: []any{
				opcode.Variable("arr"),
				opcode.OpCode{
					Cmd:  opcode.BinaryOp,
					Args: []any{"+", opcode.Variable("offset"), int64(1)},
				},
				int64(999),
			},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)
		elem, _ := arr.Get(3)
		if elem != int64(999) {
			t.Errorf("expected arr[3] = 999, got %v", elem)
		}
	})
}

// TestArrayInLoops tests array usage in loops.
func TestArrayInLoops(t *testing.T) {
	t.Run("populates array in for loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// for (i = 0; i < 5; i = i + 1) { arr[i] = i * 10 }
		opcode := opcode.OpCode{
			Cmd: opcode.For,
			Args: []any{
				// init: i = 0
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
				},
				// condition: i < 5
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
				// post: i = i + 1
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
				// body: arr[i] = i * 10
				[]opcode.OpCode{
					{Cmd: opcode.ArrayAssign, Args: []any{
						opcode.Variable("arr"),
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"*", opcode.Variable("i"), int64(10)}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, ok := vm.GetCurrentScope().Get("arr")
		if !ok {
			t.Fatal("expected array 'arr' to be created")
		}
		arr := val.(*Array)

		expected := []int64{0, 10, 20, 30, 40}
		for i, exp := range expected {
			elem, _ := arr.Get(int64(i))
			if elem != exp {
				t.Errorf("arr[%d]: expected %d, got %v", i, exp, elem)
			}
		}
	})

	t.Run("sums array elements in while loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(1), int64(2), int64(3), int64(4), int64(5)}))
		vm.GetCurrentScope().Set("sum", int64(0))
		vm.GetCurrentScope().Set("i", int64(0))
		vm.GetCurrentScope().Set("len", int64(5))

		// while (i < len) { sum = sum + arr[i]; i = i + 1 }
		opcode := opcode.OpCode{
			Cmd: opcode.While,
			Args: []any{
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), opcode.Variable("len")}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("sum"),
						opcode.OpCode{
							Cmd: opcode.BinaryOp,
							Args: []any{
								"+",
								opcode.Variable("sum"),
								opcode.OpCode{
									Cmd:  opcode.ArrayAccess,
									Args: []any{opcode.Variable("arr"), opcode.Variable("i")},
								},
							},
						},
					}},
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sum, _ := vm.GetCurrentScope().Get("sum")
		if sum != int64(15) {
			t.Errorf("expected sum = 15, got %v", sum)
		}
	})
}
