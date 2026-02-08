package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestBuiltinArraySize tests the ArraySize builtin function.
// Requirements: 4.1, 4.2, 4.3
func TestBuiltinArraySize(t *testing.T) {
	t.Run("empty array returns 0", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArray(0)

		result, err := vm.builtins["ArraySize"](vm, []any{arr})
		if err != nil {
			t.Fatalf("ArraySize returned error: %v", err)
		}

		resultInt, ok := result.(int64)
		if !ok {
			t.Fatalf("ArraySize returned non-int64: %T", result)
		}

		if resultInt != 0 {
			t.Errorf("ArraySize(empty) = %d, want 0", resultInt)
		}
	})

	t.Run("non-empty array returns correct count", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20), int64(30)})

		result, err := vm.builtins["ArraySize"](vm, []any{arr})
		if err != nil {
			t.Fatalf("ArraySize returned error: %v", err)
		}

		resultInt, ok := result.(int64)
		if !ok {
			t.Fatalf("ArraySize returned non-int64: %T", result)
		}

		if resultInt != 3 {
			t.Errorf("ArraySize([10,20,30]) = %d, want 3", resultInt)
		}
	})

	t.Run("auto-expanded array returns correct count", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArray(2)
		// Auto-expand by setting index 9 (expands to size 10)
		arr.Set(9, int64(99))

		result, err := vm.builtins["ArraySize"](vm, []any{arr})
		if err != nil {
			t.Fatalf("ArraySize returned error: %v", err)
		}

		resultInt, ok := result.(int64)
		if !ok {
			t.Fatalf("ArraySize returned non-int64: %T", result)
		}

		if resultInt != 10 {
			t.Errorf("ArraySize(auto-expanded to 10) = %d, want 10", resultInt)
		}
	})
}

// TestBuiltinDelArrayAll tests the DelArrayAll builtin function.
// Requirements: 5.1, 5.2, 5.3
func TestBuiltinDelArrayAll(t *testing.T) {
	t.Run("non-empty array becomes size 0", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(1), int64(2), int64(3)})

		_, err := vm.builtins["DelArrayAll"](vm, []any{arr})
		if err != nil {
			t.Fatalf("DelArrayAll returned error: %v", err)
		}

		if arr.Len() != 0 {
			t.Errorf("after DelArrayAll, array size = %d, want 0", arr.Len())
		}
	})

	t.Run("empty array no error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArray(0)

		_, err := vm.builtins["DelArrayAll"](vm, []any{arr})
		if err != nil {
			t.Fatalf("DelArrayAll on empty array returned error: %v", err)
		}

		if arr.Len() != 0 {
			t.Errorf("after DelArrayAll on empty, array size = %d, want 0", arr.Len())
		}
	})

	t.Run("array reusable after clear", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(1), int64(2), int64(3)})

		_, err := vm.builtins["DelArrayAll"](vm, []any{arr})
		if err != nil {
			t.Fatalf("DelArrayAll returned error: %v", err)
		}

		// After clearing, add elements via Set (auto-expand)
		arr.Set(0, int64(42))
		if arr.Len() != 1 {
			t.Errorf("after re-adding element, array size = %d, want 1", arr.Len())
		}

		val, ok := arr.Get(0)
		if !ok {
			t.Fatal("Get(0) returned not ok after re-adding")
		}
		if val != int64(42) {
			t.Errorf("Get(0) = %v, want 42", val)
		}
	})
}

// TestBuiltinDelArrayAt tests the DelArrayAt builtin function.
// Requirements: 6.1, 6.2, 6.3
func TestBuiltinDelArrayAt(t *testing.T) {
	t.Run("valid index deletes and shifts", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20), int64(30), int64(40)})

		// Delete element at index 1 (value 20)
		_, err := vm.builtins["DelArrayAt"](vm, []any{arr, int64(1)})
		if err != nil {
			t.Fatalf("DelArrayAt returned error: %v", err)
		}

		if arr.Len() != 3 {
			t.Errorf("after DelArrayAt, array size = %d, want 3", arr.Len())
		}

		// Verify elements shifted: [10, 30, 40]
		expected := []int64{10, 30, 40}
		for i, exp := range expected {
			val, ok := arr.Get(int64(i))
			if !ok {
				t.Fatalf("Get(%d) returned not ok", i)
			}
			if val != exp {
				t.Errorf("Get(%d) = %v, want %d", i, val, exp)
			}
		}
	})

	t.Run("delete last element", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20), int64(30)})

		// Delete last element at index 2
		_, err := vm.builtins["DelArrayAt"](vm, []any{arr, int64(2)})
		if err != nil {
			t.Fatalf("DelArrayAt returned error: %v", err)
		}

		if arr.Len() != 2 {
			t.Errorf("after DelArrayAt last, array size = %d, want 2", arr.Len())
		}

		// Verify remaining: [10, 20]
		expected := []int64{10, 20}
		for i, exp := range expected {
			val, ok := arr.Get(int64(i))
			if !ok {
				t.Fatalf("Get(%d) returned not ok", i)
			}
			if val != exp {
				t.Errorf("Get(%d) = %v, want %d", i, val, exp)
			}
		}
	})

	t.Run("negative index returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20)})

		_, err := vm.builtins["DelArrayAt"](vm, []any{arr, int64(-1)})
		if err == nil {
			t.Fatal("DelArrayAt with negative index should return error")
		}
	})

	t.Run("index >= size returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20)})

		_, err := vm.builtins["DelArrayAt"](vm, []any{arr, int64(2)})
		if err == nil {
			t.Fatal("DelArrayAt with index >= size should return error")
		}
	})
}

// TestBuiltinInsArrayAt tests the InsArrayAt builtin function.
// Requirements: 7.1, 7.2, 7.3, 7.4
func TestBuiltinInsArrayAt(t *testing.T) {
	t.Run("valid index inserts and shifts", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20), int64(30)})

		// Insert 99 at index 1
		_, err := vm.builtins["InsArrayAt"](vm, []any{arr, int64(1), int64(99)})
		if err != nil {
			t.Fatalf("InsArrayAt returned error: %v", err)
		}

		if arr.Len() != 4 {
			t.Errorf("after InsArrayAt, array size = %d, want 4", arr.Len())
		}

		// Verify: [10, 99, 20, 30]
		expected := []int64{10, 99, 20, 30}
		for i, exp := range expected {
			val, ok := arr.Get(int64(i))
			if !ok {
				t.Fatalf("Get(%d) returned not ok", i)
			}
			if val != exp {
				t.Errorf("Get(%d) = %v, want %d", i, val, exp)
			}
		}
	})

	t.Run("insert at head (index 0)", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20), int64(30)})

		// Insert 99 at index 0
		_, err := vm.builtins["InsArrayAt"](vm, []any{arr, int64(0), int64(99)})
		if err != nil {
			t.Fatalf("InsArrayAt returned error: %v", err)
		}

		if arr.Len() != 4 {
			t.Errorf("after InsArrayAt at head, array size = %d, want 4", arr.Len())
		}

		// Verify: [99, 10, 20, 30]
		expected := []int64{99, 10, 20, 30}
		for i, exp := range expected {
			val, ok := arr.Get(int64(i))
			if !ok {
				t.Fatalf("Get(%d) returned not ok", i)
			}
			if val != exp {
				t.Errorf("Get(%d) = %v, want %d", i, val, exp)
			}
		}
	})

	t.Run("insert at tail (index == size)", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20), int64(30)})

		// Insert 99 at index 3 (== size, appends)
		_, err := vm.builtins["InsArrayAt"](vm, []any{arr, int64(3), int64(99)})
		if err != nil {
			t.Fatalf("InsArrayAt returned error: %v", err)
		}

		if arr.Len() != 4 {
			t.Errorf("after InsArrayAt at tail, array size = %d, want 4", arr.Len())
		}

		// Verify: [10, 20, 30, 99]
		expected := []int64{10, 20, 30, 99}
		for i, exp := range expected {
			val, ok := arr.Get(int64(i))
			if !ok {
				t.Fatalf("Get(%d) returned not ok", i)
			}
			if val != exp {
				t.Errorf("Get(%d) = %v, want %d", i, val, exp)
			}
		}
	})

	t.Run("negative index returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20)})

		_, err := vm.builtins["InsArrayAt"](vm, []any{arr, int64(-1), int64(99)})
		if err == nil {
			t.Fatal("InsArrayAt with negative index should return error")
		}
	})

	t.Run("index > size returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		arr := NewArrayFromSlice([]any{int64(10), int64(20)})

		// index 3 > size 2
		_, err := vm.builtins["InsArrayAt"](vm, []any{arr, int64(3), int64(99)})
		if err == nil {
			t.Fatal("InsArrayAt with index > size should return error")
		}
	})
}
