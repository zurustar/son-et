package vm

import (
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

func TestSubStr(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		start    int64
		length   int64
		expected string
	}{
		{
			name:     "basic substring",
			str:      "Hello, World!",
			start:    0,
			length:   5,
			expected: "Hello",
		},
		{
			name:     "middle substring",
			str:      "Hello, World!",
			start:    7,
			length:   5,
			expected: "World",
		},
		{
			name:     "length exceeds string",
			str:      "Hello",
			start:    2,
			length:   100,
			expected: "llo",
		},
		{
			name:     "start at end",
			str:      "Hello",
			start:    5,
			length:   1,
			expected: "",
		},
		{
			name:     "start beyond end",
			str:      "Hello",
			start:    10,
			length:   1,
			expected: "",
		},
		{
			name:     "zero length",
			str:      "Hello",
			start:    0,
			length:   0,
			expected: "",
		},
		{
			name:     "negative start treated as 0",
			str:      "Hello",
			start:    -1,
			length:   3,
			expected: "Hel",
		},
		{
			name:     "Japanese characters",
			str:      "こんにちは世界",
			start:    0,
			length:   5,
			expected: "こんにちは",
		},
		{
			name:     "Japanese middle",
			str:      "こんにちは世界",
			start:    5,
			length:   2,
			expected: "世界",
		},
		{
			name:     "mixed ASCII and Japanese",
			str:      "Hello世界",
			start:    5,
			length:   2,
			expected: "世界",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			// Call SubStr builtin
			result, err := vm.builtins["SubStr"](vm, []any{tt.str, tt.start, tt.length})
			if err != nil {
				t.Fatalf("SubStr returned error: %v", err)
			}

			resultStr, ok := result.(string)
			if !ok {
				t.Fatalf("SubStr returned non-string: %T", result)
			}

			if resultStr != tt.expected {
				t.Errorf("SubStr(%q, %d, %d) = %q, want %q", tt.str, tt.start, tt.length, resultStr, tt.expected)
			}
		})
	}
}

func TestStrFind(t *testing.T) {
	tests := []struct {
		name     string
		str      string
		search   string
		expected int64
	}{
		{
			name:     "find at start",
			str:      "Hello, World!",
			search:   "Hello",
			expected: 0,
		},
		{
			name:     "find in middle",
			str:      "Hello, World!",
			search:   "World",
			expected: 7,
		},
		{
			name:     "find single char",
			str:      "Hello, World!",
			search:   ",",
			expected: 5,
		},
		{
			name:     "not found",
			str:      "Hello, World!",
			search:   "xyz",
			expected: -1,
		},
		{
			name:     "empty search string",
			str:      "Hello",
			search:   "",
			expected: 0,
		},
		{
			name:     "search longer than string",
			str:      "Hi",
			search:   "Hello",
			expected: -1,
		},
		{
			name:     "Japanese characters",
			str:      "こんにちは世界",
			search:   "世界",
			expected: 5,
		},
		{
			name:     "Japanese not found",
			str:      "こんにちは世界",
			search:   "さようなら",
			expected: -1,
		},
		{
			name:     "find comma in Japanese text",
			str:      "テスト,データ",
			search:   ",",
			expected: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vm := New([]opcode.OpCode{})

			// Call StrFind builtin
			result, err := vm.builtins["StrFind"](vm, []any{tt.str, tt.search})
			if err != nil {
				t.Fatalf("StrFind returned error: %v", err)
			}

			resultInt, ok := result.(int64)
			if !ok {
				t.Fatalf("StrFind returned non-int64: %T", result)
			}

			if resultInt != tt.expected {
				t.Errorf("StrFind(%q, %q) = %d, want %d", tt.str, tt.search, resultInt, tt.expected)
			}
		})
	}
}

// TestSubStrStrFindIntegration tests the combination of SubStr and StrFind
// as used in the robot sample (CITEXT.TFY)
func TestSubStrStrFindIntegration(t *testing.T) {
	vm := New([]opcode.OpCode{})

	// Simulate the robot sample pattern:
	// ln = StrFind(sTexts, ",")
	// wTexts = wTexts + SubStr(sTexts, 0, ln)
	// sTexts = SubStr(sTexts, ln+1, StrLen(sTexts)-ln-1)

	sTexts := "Hello,World,Test"
	wTexts := ""

	// First iteration
	ln, _ := vm.builtins["StrFind"](vm, []any{sTexts, ","})
	lnInt := ln.(int64)

	if lnInt != 5 {
		t.Errorf("First StrFind = %d, want 5", lnInt)
	}

	part, _ := vm.builtins["SubStr"](vm, []any{sTexts, int64(0), lnInt})
	wTexts = wTexts + part.(string)

	if wTexts != "Hello" {
		t.Errorf("After first SubStr, wTexts = %q, want %q", wTexts, "Hello")
	}

	// Get remaining string
	strLen, _ := vm.builtins["StrLen"](vm, []any{sTexts})
	remaining, _ := vm.builtins["SubStr"](vm, []any{sTexts, lnInt + 1, strLen.(int64) - lnInt - 1})
	sTexts = remaining.(string)

	if sTexts != "World,Test" {
		t.Errorf("After first iteration, sTexts = %q, want %q", sTexts, "World,Test")
	}
}
