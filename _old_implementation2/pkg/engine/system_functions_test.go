package engine

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestRandom(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test Random(10) - should return 0-9
	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"random",
			int64(10),
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("Random failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	resultInt := result.(int64)
	if resultInt < 0 || resultInt >= 10 {
		t.Errorf("Random(10) = %d, want 0-9", resultInt)
	}
}

func TestRandomRange(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Test Random(5, 15) - should return 5-14
	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"random",
			int64(5),
			int64(15),
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("Random failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	resultInt := result.(int64)
	if resultInt < 5 || resultInt >= 15 {
		t.Errorf("Random(5, 15) = %d, want 5-14", resultInt)
	}
}

func TestRandomDistribution(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	// Generate 100 random numbers and check distribution
	counts := make(map[int64]int)
	for i := 0; i < 100; i++ {
		op := interpreter.OpCode{
			Cmd: interpreter.OpCall,
			Args: []any{
				"random",
				int64(10),
			},
		}

		err := vm.executeCall(seq, op)
		if err != nil {
			t.Fatalf("Random failed: %v", err)
		}

		result := seq.GetVariable("__return__")
		resultInt := result.(int64)
		counts[resultInt]++
	}

	// Check that we got at least some variety (not all the same number)
	if len(counts) < 5 {
		t.Errorf("Random distribution too narrow: got %d unique values, want at least 5", len(counts))
	}
}

func TestGetSysTime(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	before := time.Now().Unix()

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"getsystime",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("GetSysTime failed: %v", err)
	}

	after := time.Now().Unix()

	result := seq.GetVariable("__return__")
	resultInt := result.(int64)

	if resultInt < before || resultInt > after {
		t.Errorf("GetSysTime() = %d, want between %d and %d", resultInt, before, after)
	}
}

func TestWhatDay(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	expectedDay := time.Now().Day()

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"whatday",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("WhatDay failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	resultInt := int(result.(int64))

	if resultInt != expectedDay {
		t.Errorf("WhatDay() = %d, want %d", resultInt, expectedDay)
	}
}

func TestWhatTime(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	now := time.Now()

	tests := []struct {
		name     string
		mode     int
		expected int
	}{
		{"hour", 0, now.Hour()},
		{"minute", 1, now.Minute()},
		{"second", 2, now.Second()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := interpreter.OpCode{
				Cmd: interpreter.OpCall,
				Args: []any{
					"whattime",
					int64(tt.mode),
				},
			}

			err := vm.executeCall(seq, op)
			if err != nil {
				t.Fatalf("WhatTime(%d) failed: %v", tt.mode, err)
			}

			result := seq.GetVariable("__return__")
			resultInt := int(result.(int64))

			// Allow for small time differences (within 1 unit)
			if resultInt < tt.expected-1 || resultInt > tt.expected+1 {
				t.Errorf("WhatTime(%d) = %d, want approximately %d", tt.mode, resultInt, tt.expected)
			}
		})
	}
}

func TestWhatTimeInvalidMode(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"whattime",
			int64(99), // Invalid mode
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("WhatTime failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	resultInt := result.(int64)

	if resultInt != 0 {
		t.Errorf("WhatTime(99) = %d, want 0 (default for invalid mode)", resultInt)
	}
}
