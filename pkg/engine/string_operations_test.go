package engine

import (
	"fmt"
	"math/rand"
	"strings"
	"testing"
	"testing/quick"
	"unicode"
)

// TestStrLen tests the StrLen function
// Requirement 13.1: WHEN StrLen is called, THE Runtime SHALL return the length of the string
func TestStrLen(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"empty string", "", 0},
		{"single char", "a", 1},
		{"multiple chars", "hello", 5},
		{"unicode chars", "こんにちは", 15}, // UTF-8 byte count, not character count
		{"mixed", "hello世界", 11},       // UTF-8 byte count
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrLen(tt.input)
			if result != tt.expected {
				t.Errorf("StrLen(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestSubStr tests the SubStr function
// Requirement 13.2: WHEN SubStr is called, THE Runtime SHALL extract a substring
func TestSubStr(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		start    int
		length   int
		expected string
	}{
		{"normal case", "hello", 1, 3, "ell"},
		{"start at 0", "hello", 0, 3, "hel"},
		{"to end", "hello", 2, 10, "llo"},
		{"start beyond length", "hello", 10, 5, ""},
		{"empty string", "", 0, 5, ""},
		{"zero length", "hello", 2, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SubStr(tt.input, tt.start, tt.length)
			if result != tt.expected {
				t.Errorf("SubStr(%q, %d, %d) = %q, want %q", tt.input, tt.start, tt.length, result, tt.expected)
			}
		})
	}
}

// TestStrFind tests the StrFind function
// Requirement 13.3: WHEN StrFind is called, THE Runtime SHALL return the index of the first occurrence
func TestStrFind(t *testing.T) {
	tests := []struct {
		name     string
		haystack string
		needle   string
		expected int
	}{
		{"found at start", "hello", "he", 0},
		{"found in middle", "hello", "ll", 2},
		{"found at end", "hello", "lo", 3},
		{"not found", "hello", "xyz", -1},
		{"empty needle", "hello", "", 0},
		{"empty haystack", "", "x", -1},
		{"both empty", "", "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrFind(tt.haystack, tt.needle)
			if result != tt.expected {
				t.Errorf("StrFind(%q, %q) = %d, want %d", tt.haystack, tt.needle, result, tt.expected)
			}
		})
	}
}

// TestCharCode tests the CharCode function
// Requirement 30.2: WHEN CharCode is called, THE Runtime SHALL return the character code
func TestCharCode(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"ASCII letter", "A", 65},
		{"ASCII digit", "0", 48},
		{"lowercase", "a", 97},
		{"space", " ", 32},
		{"unicode", "あ", 12354},
		{"empty string", "", 0},
		{"multiple chars", "ABC", 65}, // Should return first char
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CharCode(tt.input)
			if result != tt.expected {
				t.Errorf("CharCode(%q) = %d, want %d", tt.input, result, tt.expected)
			}
		})
	}
}

// TestStrCode tests the StrCode function
// Requirement 30.5: WHEN StrCode is called, THE Runtime SHALL convert a character code to a string
func TestStrCode(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected string
	}{
		{"ASCII letter", 65, "A"},
		{"ASCII digit", 48, "0"},
		{"lowercase", 97, "a"},
		{"space", 32, " "},
		{"unicode", 12354, "あ"},
		{"negative", -1, ""},
		{"too large", 0x110000, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrCode(tt.input)
			if result != tt.expected {
				t.Errorf("StrCode(%d) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestStrUp tests the StrUp function
// Requirement 30.3: WHEN StrUp is called, THE Runtime SHALL convert to uppercase
func TestStrUp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"lowercase", "hello", "HELLO"},
		{"mixed case", "HeLLo", "HELLO"},
		{"already upper", "HELLO", "HELLO"},
		{"with numbers", "hello123", "HELLO123"},
		{"empty", "", ""},
		{"unicode", "café", "CAFÉ"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrUp(tt.input)
			if result != tt.expected {
				t.Errorf("StrUp(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// TestStrLow tests the StrLow function
// Requirement 30.4: WHEN StrLow is called, THE Runtime SHALL convert to lowercase
func TestStrLow(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"uppercase", "HELLO", "hello"},
		{"mixed case", "HeLLo", "hello"},
		{"already lower", "hello", "hello"},
		{"with numbers", "HELLO123", "hello123"},
		{"empty", "", ""},
		{"unicode", "CAFÉ", "café"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrLow(tt.input)
			if result != tt.expected {
				t.Errorf("StrLow(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// Property 16: String operation correctness
// Feature: core-engine, Property 16: String operation correctness
// Validates: Requirements 13.1, 13.2, 13.3
func TestProperty16_StringOperationCorrectness(t *testing.T) {
	// Property: StrLen returns the actual length
	t.Run("StrLen correctness", func(t *testing.T) {
		property := func(s string) bool {
			return StrLen(s) == len(s)
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: SubStr with valid bounds returns correct substring
	t.Run("SubStr correctness", func(t *testing.T) {
		property := func(s string) bool {
			if len(s) == 0 {
				return true
			}
			// Generate valid start and length
			start := rand.Intn(len(s))
			maxLen := len(s) - start
			if maxLen == 0 {
				return true
			}
			length := rand.Intn(maxLen) + 1

			result := SubStr(s, start, length)
			expected := s[start : start+length]
			return result == expected
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrFind returns correct index or -1
	t.Run("StrFind correctness", func(t *testing.T) {
		property := func(s string, sub string) bool {
			result := StrFind(s, sub)
			stdResult := strings.Index(s, sub)
			return result == stdResult
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: CharCode and StrCode are inverses for valid codes
	t.Run("CharCode/StrCode round trip", func(t *testing.T) {
		property := func(r rune) bool {
			if r < 0 || r > 0x10FFFF {
				return true // Skip invalid runes
			}
			s := string(r)
			code := CharCode(s)
			result := StrCode(code)
			return result == s
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrUp converts all lowercase to uppercase
	t.Run("StrUp correctness", func(t *testing.T) {
		property := func(s string) bool {
			result := StrUp(s)
			expected := strings.ToUpper(s)

			// Check that it matches standard library
			if result != expected {
				return false
			}

			// Check that result has no lowercase letters that have uppercase equivalents
			for _, r := range result {
				if unicode.IsLower(r) && unicode.ToUpper(r) != r {
					return false
				}
			}
			return true
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrLow converts all uppercase to lowercase
	t.Run("StrLow correctness", func(t *testing.T) {
		property := func(s string) bool {
			result := StrLow(s)
			expected := strings.ToLower(s)

			// Check that it matches standard library
			if result != expected {
				return false
			}

			// Check that result has no uppercase letters that have lowercase equivalents
			for _, r := range result {
				if unicode.IsUpper(r) && unicode.ToLower(r) != r {
					return false
				}
			}
			return true
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrUp and StrLow are inverses for ASCII
	t.Run("StrUp/StrLow round trip", func(t *testing.T) {
		property := func(s string) bool {
			// Only test with ASCII letters for reliable round-trip
			asciiOnly := true
			for _, r := range s {
				if r > 127 {
					asciiOnly = false
					break
				}
			}
			if !asciiOnly {
				return true
			}

			upper := StrUp(s)
			lower := StrLow(upper)
			return lower == StrLow(s)
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}

// TestStrPrint tests the StrPrint function
// Requirement 13.4, 13.5: WHEN StrPrint is called, THE Runtime SHALL format according to format specifiers
func TestStrPrint(t *testing.T) {
	tests := []struct {
		name     string
		format   string
		args     []any
		expected string
	}{
		{"string format", "Hello %s", []any{"World"}, "Hello World"},
		{"decimal format", "Value: %ld", []any{42}, "Value: 42"},
		{"hex format", "Hex: %lx", []any{255}, "Hex: ff"},
		{"multiple args", "%s: %ld", []any{"Count", 10}, "Count: 10"},
		{"no args", "Hello", []any{}, "Hello"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StrPrint(tt.format, tt.args...)
			if result != tt.expected {
				t.Errorf("StrPrint(%q, %v) = %q, want %q", tt.format, tt.args, result, tt.expected)
			}
		})
	}
}

// Property 17: String formatting correctness
// Feature: core-engine, Property 17: String formatting correctness
// Validates: Requirements 13.4, 13.5
func TestProperty17_StringFormattingCorrectness(t *testing.T) {
	// Property: StrPrint with %s formats strings correctly
	t.Run("String formatting", func(t *testing.T) {
		property := func(s string) bool {
			format := "Value: %s"
			result := StrPrint(format, s)
			expected := "Value: " + s
			return result == expected
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrPrint with %ld formats integers correctly
	t.Run("Decimal formatting", func(t *testing.T) {
		property := func(n int) bool {
			format := "Number: %ld"
			result := StrPrint(format, n)
			expected := fmt.Sprintf("Number: %d", n)
			return result == expected
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrPrint with %lx formats hex correctly
	t.Run("Hex formatting", func(t *testing.T) {
		property := func(n uint) bool {
			// Use uint to avoid negative numbers in hex
			format := "Hex: %lx"
			result := StrPrint(format, n)
			expected := fmt.Sprintf("Hex: %x", n)
			return result == expected
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})

	// Property: StrPrint with multiple format specifiers
	t.Run("Multiple format specifiers", func(t *testing.T) {
		property := func(s string, n int) bool {
			format := "%s: %ld"
			result := StrPrint(format, s, n)
			expected := fmt.Sprintf("%s: %d", s, n)
			return result == expected
		}
		if err := quick.Check(property, &quick.Config{MaxCount: 100}); err != nil {
			t.Error(err)
		}
	})
}
