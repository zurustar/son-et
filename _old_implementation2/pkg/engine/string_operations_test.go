package engine

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestStrLen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int64
	}{
		{"empty string", "", 0},
		{"simple string", "hello", 5},
		{"unicode string", "こんにちは", 5},
		{"mixed string", "hello世界", 7},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, _, _, _ := newTestVM()
			seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

			op := interpreter.OpCode{
				Cmd: interpreter.OpCall,
				Args: []any{
					"strlen",
					tt.input,
				},
			}

			err := vm.executeCall(seq, op)
			if err != nil {
				t.Fatalf("StrLen failed: %v", err)
			}

			result := seq.GetVariable("__return__")
			if result != tt.expected {
				t.Errorf("StrLen(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSubStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		length   int
		expected string
	}{
		{"simple substring", "hello world", 0, 5, "hello"},
		{"middle substring", "hello world", 6, 5, "world"},
		{"unicode substring", "こんにちは", 0, 2, "こん"},
		{"out of bounds", "hello", 10, 5, ""},
		{"length exceeds", "hello", 2, 10, "llo"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, _, _, _ := newTestVM()
			seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

			op := interpreter.OpCode{
				Cmd: interpreter.OpCall,
				Args: []any{
					"substr",
					tt.input,
					int64(tt.start),
					int64(tt.length),
				},
			}

			err := vm.executeCall(seq, op)
			if err != nil {
				t.Fatalf("SubStr failed: %v", err)
			}

			result := seq.GetVariable("__return__")
			if result != tt.expected {
				t.Errorf("SubStr(%q, %d, %d) = %q, want %q", tt.input, tt.start, tt.length, result, tt.expected)
			}
		})
	}
}

func TestStrFind(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		search   string
		expected int64
	}{
		{"found at start", "hello world", "hello", 0},
		{"found in middle", "hello world", "world", 6},
		{"not found", "hello world", "xyz", -1},
		{"unicode search", "こんにちは世界", "世界", 5},
		{"empty search", "hello", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, _, _, _ := newTestVM()
			seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

			op := interpreter.OpCode{
				Cmd: interpreter.OpCall,
				Args: []any{
					"strfind",
					tt.input,
					tt.search,
				},
			}

			err := vm.executeCall(seq, op)
			if err != nil {
				t.Fatalf("StrFind failed: %v", err)
			}

			result := seq.GetVariable("__return__")
			if result != tt.expected {
				t.Errorf("StrFind(%q, %q) = %v, want %v", tt.input, tt.search, result, tt.expected)
			}
		})
	}
}

func TestStrPrint(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{"simple format", "hello %s", []any{"world"}, "hello world"},
		{"multiple args", "%d + %d = %d", []any{int64(1), int64(2), int64(3)}, "1 + 2 = 3"},
		{"no args", "hello", []any{}, "hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm, _, _, _ := newTestVM()
			seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

			args := []any{"strprint", tt.format}
			args = append(args, tt.args...)

			op := interpreter.OpCode{
				Cmd:  interpreter.OpCall,
				Args: args,
			}

			err := vm.executeCall(seq, op)
			if err != nil {
				t.Fatalf("StrPrint failed: %v", err)
			}

			result := seq.GetVariable("__return__")
			if result != tt.expected {
				t.Errorf("StrPrint(%q, %v) = %q, want %q", tt.format, tt.args, result, tt.expected)
			}
		})
	}
}

func TestStrUp(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"strup",
			"hello",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("StrUp failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	if result != "HELLO" {
		t.Errorf("StrUp(\"hello\") = %q, want \"HELLO\"", result)
	}
}

func TestStrLow(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"strlow",
			"HELLO",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("StrLow failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	if result != "hello" {
		t.Errorf("StrLow(\"HELLO\") = %q, want \"hello\"", result)
	}
}

func TestCharCode(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"charcode",
			"A",
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("CharCode failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	if result != int64(65) {
		t.Errorf("CharCode(\"A\") = %v, want 65", result)
	}
}

func TestStrCode(t *testing.T) {
	vm, _, _, _ := newTestVM()
	seq := NewSequencer([]interpreter.OpCode{}, TIME, nil)

	op := interpreter.OpCode{
		Cmd: interpreter.OpCall,
		Args: []any{
			"strcode",
			int64(65),
		},
	}

	err := vm.executeCall(seq, op)
	if err != nil {
		t.Fatalf("StrCode failed: %v", err)
	}

	result := seq.GetVariable("__return__")
	if result != "A" {
		t.Errorf("StrCode(65) = %q, want \"A\"", result)
	}
}
