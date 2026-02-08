package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestMakeLong tests the MakeLong builtin function.
// Requirements: 8.1, 8.2
func TestMakeLong(t *testing.T) {
	tests := []struct {
		name     string
		low      int64
		high     int64
		expected int64
	}{
		{
			name:     "basic combination low=2 high=1",
			low:      2,
			high:     1,
			expected: 0x00010002, // 65538
		},
		{
			name:     "both zero",
			low:      0,
			high:     0,
			expected: 0,
		},
		{
			name:     "max 16-bit values",
			low:      0xFFFF,
			high:     0xFFFF,
			expected: int64(0xFFFFFFFF), // (0xFFFF << 16) | 0xFFFF
		},
		{
			name:     "16-bit overflow uses only lower 16 bits",
			low:      0x1FFFF, // exceeds 16-bit, lower 16 bits = 0xFFFF
			high:     0x10001, // exceeds 16-bit, lower 16 bits = 0x0001
			expected: 0x0001FFFF, // Req 8.2: each arg uses only lower 16 bits
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["MakeLong"](vm, []any{tt.low, tt.high})
			if err != nil {
				t.Fatalf("MakeLong returned error: %v", err)
			}

			resultInt, ok := result.(int64)
			if !ok {
				t.Fatalf("MakeLong returned non-int64: %T", result)
			}

			if resultInt != tt.expected {
				t.Errorf("MakeLong(%d, %d) = 0x%X (%d), want 0x%X (%d)",
					tt.low, tt.high, resultInt, resultInt, tt.expected, tt.expected)
			}
		})
	}
}

// TestGetHiWord tests the GetHiWord builtin function.
// Requirements: 9.1, 9.2
func TestGetHiWord(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "basic decomposition 0x00010002",
			input:    0x00010002,
			expected: 1,
		},
		{
			name:     "zero returns 0",
			input:    0,
			expected: 0,
		},
		{
			name:     "known value 0xFFFFFFFF",
			input:    0xFFFFFFFF,
			expected: 0xFFFF, // 65535
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["GetHiWord"](vm, []any{tt.input})
			if err != nil {
				t.Fatalf("GetHiWord returned error: %v", err)
			}

			resultInt, ok := result.(int64)
			if !ok {
				t.Fatalf("GetHiWord returned non-int64: %T", result)
			}

			if resultInt != tt.expected {
				t.Errorf("GetHiWord(0x%X) = %d (0x%X), want %d (0x%X)",
					tt.input, resultInt, resultInt, tt.expected, tt.expected)
			}
		})
	}
}

// TestGetLowWord tests the GetLowWord builtin function.
// Requirements: 10.1, 10.2
func TestGetLowWord(t *testing.T) {
	tests := []struct {
		name     string
		input    int64
		expected int64
	}{
		{
			name:     "basic decomposition 0x00010002",
			input:    0x00010002,
			expected: 2,
		},
		{
			name:     "zero returns 0",
			input:    0,
			expected: 0,
		},
		{
			name:     "known value 0xFFFFFFFF",
			input:    0xFFFFFFFF,
			expected: 0xFFFF, // 65535
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["GetLowWord"](vm, []any{tt.input})
			if err != nil {
				t.Fatalf("GetLowWord returned error: %v", err)
			}

			resultInt, ok := result.(int64)
			if !ok {
				t.Fatalf("GetLowWord returned non-int64: %T", result)
			}

			if resultInt != tt.expected {
				t.Errorf("GetLowWord(0x%X) = %d (0x%X), want %d (0x%X)",
					tt.input, resultInt, resultInt, tt.expected, tt.expected)
			}
		})
	}
}
