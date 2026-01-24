package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestArraySize(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	seq.SetVariable("arr", []int64{1, 2, 3, 4, 5})

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"arraysize",
			"arr",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("ArraySize failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	if result != int64(5) {
		t.Errorf("ArraySize(arr) = %v, want 5", result)
	}
}

func TestDelArrayAll(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	seq.SetVariable("arr", []int64{1, 2, 3, 4, 5})

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"delarrayall",
			"arr",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("DelArrayAll failed: %v", err)
	}

	arr := seq.GetVariable("arr")
	if arrSlice, ok := arr.([]int64); !ok || len(arrSlice) != 0 {
		t.Errorf("DelArrayAll(arr) did not clear array, got %v", arr)
	}
}

func TestDelArrayAt(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	seq.SetVariable("arr", []int64{10, 20, 30, 40, 50})

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"delarrayat",
			"arr",
			int64(2),
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("DelArrayAt failed: %v", err)
	}

	arr := seq.GetVariable("arr")
	expected := []int64{10, 20, 40, 50}
	if arrSlice, ok := arr.([]int64); !ok {
		t.Errorf("DelArrayAt(arr, 2) result is not []int64, got %T", arr)
	} else if len(arrSlice) != len(expected) {
		t.Errorf("DelArrayAt(arr, 2) length = %d, want %d", len(arrSlice), len(expected))
	} else {
		for i, v := range expected {
			if arrSlice[i] != v {
				t.Errorf("DelArrayAt(arr, 2)[%d] = %d, want %d", i, arrSlice[i], v)
			}
		}
	}
}

func TestInsArrayAt(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	seq.SetVariable("arr", []int64{10, 20, 40, 50})

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"insarrayat",
			"arr",
			int64(2),
			int64(30),
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("InsArrayAt failed: %v", err)
	}

	arr := seq.GetVariable("arr")
	expected := []int64{10, 20, 30, 40, 50}
	if arrSlice, ok := arr.([]int64); !ok {
		t.Errorf("InsArrayAt(arr, 2, 30) result is not []int64, got %T", arr)
	} else if len(arrSlice) != len(expected) {
		t.Errorf("InsArrayAt(arr, 2, 30) length = %d, want %d", len(arrSlice), len(expected))
	} else {
		for i, v := range expected {
			if arrSlice[i] != v {
				t.Errorf("InsArrayAt(arr, 2, 30)[%d] = %d, want %d", i, arrSlice[i], v)
			}
		}
	}
}
