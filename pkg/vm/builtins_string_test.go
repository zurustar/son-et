package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestStrUp tests the StrUp builtin function.
// Requirements: 1.1, 1.2, 1.3, 1.4
func TestStrUp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic lowercase to uppercase",
			input:    "hello",
			expected: "HELLO",
		},
		{
			name:     "already uppercase",
			input:    "HELLO",
			expected: "HELLO",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "HELLO WORLD",
		},
		{
			name:     "Japanese mixed with ASCII",
			input:    "Hello世界",
			expected: "HELLO世界",
		},
		{
			name:     "Japanese only",
			input:    "こんにちは",
			expected: "こんにちは",
		},
		{
			name:     "numbers and symbols unchanged",
			input:    "abc123!@#",
			expected: "ABC123!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["StrUp"](vm, []any{tt.input})
			if err != nil {
				t.Fatalf("StrUp returned error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("StrUp returned non-string: %T", result)
			}

			if resultStr != tt.expected {
				t.Errorf("StrUp(%q) = %q, want %q", tt.input, resultStr, tt.expected)
			}
		})
	}
}

// TestStrLow tests the StrLow builtin function.
// Requirements: 2.1, 2.2, 2.3, 2.4
func TestStrLow(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "basic uppercase to lowercase",
			input:    "HELLO",
			expected: "hello",
		},
		{
			name:     "already lowercase",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "mixed case",
			input:    "Hello World",
			expected: "hello world",
		},
		{
			name:     "Japanese mixed with ASCII",
			input:    "Hello世界",
			expected: "hello世界",
		},
		{
			name:     "Japanese only",
			input:    "こんにちは",
			expected: "こんにちは",
		},
		{
			name:     "numbers and symbols unchanged",
			input:    "ABC123!@#",
			expected: "abc123!@#",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["StrLow"](vm, []any{tt.input})
			if err != nil {
				t.Fatalf("StrLow returned error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("StrLow returned non-string: %T", result)
			}

			if resultStr != tt.expected {
				t.Errorf("StrLow(%q) = %q, want %q", tt.input, resultStr, tt.expected)
			}
		})
	}
}

// TestCharCode tests the CharCode builtin function.
// Requirements: 3.1, 3.2, 3.3, 3.4
func TestCharCode(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		index    int64
		expected int64
	}{
		{
			name:     "ASCII character A",
			str:      "ABC",
			index:    0,
			expected: 65, // 'A'
		},
		{
			name:     "ASCII character at index 1",
			str:      "ABC",
			index:    1,
			expected: 66, // 'B'
		},
		{
			name:     "Japanese character",
			str:      "あいう",
			index:    0,
			expected: 0x3042, // 'あ' = U+3042
		},
		{
			name:     "Japanese character at index 1",
			str:      "あいう",
			index:    1,
			expected: 0x3044, // 'い' = U+3044
		},
		{
			name:     "empty string",
			str:      "",
			index:    0,
			expected: 0,
		},
		{
			name:     "negative index",
			str:      "ABC",
			index:    -1,
			expected: 0,
		},
		{
			name:     "index equals length",
			str:      "ABC",
			index:    3,
			expected: 0,
		},
		{
			name:     "index beyond length",
			str:      "ABC",
			index:    100,
			expected: 0,
		},
		{
			name:     "mixed ASCII and Japanese",
			str:      "Hello世界",
			index:    5,
			expected: 0x4E16, // '世' = U+4E16
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["CharCode"](vm, []any{tt.str, tt.index})
			if err != nil {
				t.Fatalf("CharCode returned error: %v", err)
			}

			resultInt, ok := result.(int64)
			if !ok {
				t.Fatalf("CharCode returned non-int64: %T", result)
			}

			if resultInt != tt.expected {
				t.Errorf("CharCode(%q, %d) = %d (0x%X), want %d (0x%X)", tt.str, tt.index, resultInt, resultInt, tt.expected, tt.expected)
			}
		})
	}
}
