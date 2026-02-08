package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestRandomSingleArg tests Random(max) - returns random number from 0 to max-1
func TestRandomSingleArg(t *testing.T) {
	vm := New([]opcode.OpCode{})
	fn := vm.builtins["Random"]

	// Test multiple times to verify range
	for i := 0; i < 100; i++ {
		result, err := fn(vm, []any{int64(10)})
		if err != nil {
			t.Fatalf("Random(10) returned error: %v", err)
		}
		r, ok := result.(int64)
		if !ok {
			t.Fatalf("Random(10) returned non-int64: %T", result)
		}
		if r < 0 || r >= 10 {
			t.Errorf("Random(10) returned %d, expected 0-9", r)
		}
	}
}

// TestRandomTwoArgs tests Random(min, max) - returns random number from min to max-1
func TestRandomTwoArgs(t *testing.T) {
	vm := New([]opcode.OpCode{})
	fn := vm.builtins["Random"]

	// Test multiple times to verify range
	for i := 0; i < 100; i++ {
		result, err := fn(vm, []any{int64(5), int64(15)})
		if err != nil {
			t.Fatalf("Random(5, 15) returned error: %v", err)
		}
		r, ok := result.(int64)
		if !ok {
			t.Fatalf("Random(5, 15) returned non-int64: %T", result)
		}
		if r < 5 || r >= 15 {
			t.Errorf("Random(5, 15) returned %d, expected 5-14", r)
		}
	}
}

// TestRandomMaxLessThanMin tests that Random returns min when max <= min
func TestRandomMaxLessThanMin(t *testing.T) {
	vm := New([]opcode.OpCode{})
	fn := vm.builtins["Random"]

	// max == min
	result, err := fn(vm, []any{int64(5), int64(5)})
	if err != nil {
		t.Fatalf("Random(5, 5) returned error: %v", err)
	}
	if result != int64(5) {
		t.Errorf("Random(5, 5) returned %v, expected 5", result)
	}

	// max < min
	result, err = fn(vm, []any{int64(10), int64(5)})
	if err != nil {
		t.Fatalf("Random(10, 5) returned error: %v", err)
	}
	if result != int64(10) {
		t.Errorf("Random(10, 5) returned %v, expected 10", result)
	}
}

// TestRandomZeroMax tests Random(0) returns 0
func TestRandomZeroMax(t *testing.T) {
	vm := New([]opcode.OpCode{})
	fn := vm.builtins["Random"]

	result, err := fn(vm, []any{int64(0)})
	if err != nil {
		t.Fatalf("Random(0) returned error: %v", err)
	}
	if result != int64(0) {
		t.Errorf("Random(0) returned %v, expected 0", result)
	}
}

// TestRandomNoArgs tests that Random with no args returns error
func TestRandomNoArgs(t *testing.T) {
	vm := New([]opcode.OpCode{})
	fn := vm.builtins["Random"]

	_, err := fn(vm, []any{})
	if err == nil {
		t.Error("Random() should return error when called with no arguments")
	}
}
