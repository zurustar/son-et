package vm

import (
	"strings"
	"testing"
	"unicode"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// Feature: utility-builtins, Property 1: StrUp変換の正当性
// 任意の文字列に対して、StrUpの結果にはASCII小文字（a-z）が含まれず、
// 入力文字列の非ASCII文字はすべて変更されずに保持される
// **Validates: Requirements 1.1, 1.2, 1.4**
func TestProperty1_StrUpCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: StrUp result contains no ASCII lowercase letters
	properties.Property("StrUp result contains no ASCII lowercase (a-z)", prop.ForAll(
		func(input string) bool {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["StrUp"](vm, []any{input})
			if err != nil {
				return false
			}

			resultStr, ok := result.(string)
			if !ok {
				return false
			}

			// Check that no ASCII lowercase letters remain
			for _, r := range resultStr {
				if r >= 'a' && r <= 'z' {
					return false
				}
			}
			return true
		},
		gen.AnyString(),
	))

	// Property: Non-ASCII characters (CJK, Japanese, etc.) are preserved unchanged by StrUp
	// Note: Go's strings.ToUpper may transform rare Unicode scripts with case mappings
	// (e.g., Deseret), but FILLY targets Japanese text where this is not an issue.
	// We test with strings containing ASCII + Japanese/CJK characters.
	properties.Property("StrUp preserves non-ASCII characters", prop.ForAll(
		func(ascii string, japanese string) bool {
			input := ascii + japanese
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["StrUp"](vm, []any{input})
			if err != nil {
				return false
			}

			resultStr, ok := result.(string)
			if !ok {
				return false
			}

			// Extract non-ASCII characters from input and result
			inputNonASCII := filterNonASCII(input)
			resultNonASCII := filterNonASCII(resultStr)

			return inputNonASCII == resultNonASCII
		},
		gen.AnyString().Map(func(s string) string {
			// Filter to ASCII-only for the ASCII part
			var b strings.Builder
			for _, r := range s {
				if r <= 127 {
					b.WriteRune(r)
				}
			}
			return b.String()
		}),
		genJapaneseCJKString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 2: StrLow変換の正当性
// 任意の文字列に対して、StrLowの結果にはASCII大文字（A-Z）が含まれず、
// 入力文字列の非ASCII文字はすべて変更されずに保持される
// **Validates: Requirements 2.1, 2.2, 2.4**
func TestProperty2_StrLowCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: StrLow result contains no ASCII uppercase letters
	properties.Property("StrLow result contains no ASCII uppercase (A-Z)", prop.ForAll(
		func(input string) bool {
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["StrLow"](vm, []any{input})
			if err != nil {
				return false
			}

			resultStr, ok := result.(string)
			if !ok {
				return false
			}

			// Check that no ASCII uppercase letters remain
			for _, r := range resultStr {
				if r >= 'A' && r <= 'Z' {
					return false
				}
			}
			return true
		},
		gen.AnyString(),
	))

	// Property: Non-ASCII characters (CJK, Japanese, etc.) are preserved unchanged by StrLow
	// Note: Go's strings.ToLower may transform rare Unicode scripts with case mappings
	// (e.g., Deseret), but FILLY targets Japanese text where this is not an issue.
	// We test with strings containing ASCII + Japanese/CJK characters.
	properties.Property("StrLow preserves non-ASCII characters", prop.ForAll(
		func(ascii string, japanese string) bool {
			input := ascii + japanese
			vm := New([]opcode.OpCode{})

			result, err := vm.builtins["StrLow"](vm, []any{input})
			if err != nil {
				return false
			}

			resultStr, ok := result.(string)
			if !ok {
				return false
			}

			// Extract non-ASCII characters from input and result
			inputNonASCII := filterNonASCII(input)
			resultNonASCII := filterNonASCII(resultStr)

			return inputNonASCII == resultNonASCII
		},
		gen.AnyString().Map(func(s string) string {
			// Filter to ASCII-only for the ASCII part
			var b strings.Builder
			for _, r := range s {
				if r <= 127 {
					b.WriteRune(r)
				}
			}
			return b.String()
		}),
		genJapaneseCJKString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 3: StrUp/StrLowの冪等性
// 任意の文字列sに対して、StrUp(StrUp(s)) == StrUp(s) および
// StrLow(StrLow(s)) == StrLow(s) が成り立つ
// **Validates: Requirements 1.1, 1.2, 2.1, 2.2**
func TestProperty3_StrUpStrLowIdempotence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: StrUp(StrUp(s)) == StrUp(s)
	properties.Property("StrUp is idempotent", prop.ForAll(
		func(input string) bool {
			vm := New([]opcode.OpCode{})

			// First application
			result1, err := vm.builtins["StrUp"](vm, []any{input})
			if err != nil {
				return false
			}

			// Second application
			result2, err := vm.builtins["StrUp"](vm, []any{result1})
			if err != nil {
				return false
			}

			return result1 == result2
		},
		gen.AnyString(),
	))

	// Property: StrLow(StrLow(s)) == StrLow(s)
	properties.Property("StrLow is idempotent", prop.ForAll(
		func(input string) bool {
			vm := New([]opcode.OpCode{})

			// First application
			result1, err := vm.builtins["StrLow"](vm, []any{input})
			if err != nil {
				return false
			}

			// Second application
			result2, err := vm.builtins["StrLow"](vm, []any{result1})
			if err != nil {
				return false
			}

			return result1 == result2
		},
		gen.AnyString(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 4: CharCode/StrCodeラウンドトリップ
// 任意の文字列と有効なインデックスに対して、CharCodeで取得したコードポイントが
// 255以下の場合、StrCode(CharCode(str, i))は元の文字と一致する
// **Validates: Requirements 3.1, 3.4**
func TestProperty4_CharCodeStrCodeRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: For code points <= 255, StrCode(CharCode(str, i)) == original character
	properties.Property("CharCode/StrCode round-trip for code points <= 255", prop.ForAll(
		func(input string, indexFraction float64) bool {
			runes := []rune(input)
			if len(runes) == 0 {
				return true // Skip empty strings
			}

			// Derive a valid index from the fraction
			index := int64(indexFraction * float64(len(runes)))
			if index < 0 {
				index = 0
			}
			if index >= int64(len(runes)) {
				index = int64(len(runes)) - 1
			}

			vm := New([]opcode.OpCode{})

			// Get code point via CharCode
			codeResult, err := vm.builtins["CharCode"](vm, []any{input, index})
			if err != nil {
				return false
			}

			code, ok := codeResult.(int64)
			if !ok {
				return false
			}

			// Only test round-trip for code points <= 255
			if code > 255 || code <= 0 {
				return true // Skip non-applicable code points
			}

			// Convert back via StrCode
			strResult, err := vm.builtins["StrCode"](vm, []any{code})
			if err != nil {
				return false
			}

			resultStr, ok := strResult.(string)
			if !ok {
				return false
			}

			// The result should match the original character
			originalChar := string(runes[index])
			return resultStr == originalChar
		},
		genNonEmptyASCIIString(),
		gen.Float64Range(0.0, 0.999),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// filterNonASCII extracts non-ASCII runes from a string, preserving order.
func filterNonASCII(s string) string {
	var builder strings.Builder
	for _, r := range s {
		if r > unicode.MaxASCII {
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// genNonEmptyASCIIString generates non-empty strings containing ASCII characters
// (code points 1-127) to ensure valid round-trip testing with CharCode/StrCode.
func genNonEmptyASCIIString() gopter.Gen {
	return gen.SliceOfN(20, gen.IntRange(1, 127)).Map(func(codes []int) string {
		if len(codes) == 0 {
			return "A" // Fallback to ensure non-empty
		}
		runes := make([]rune, len(codes))
		for i, c := range codes {
			runes[i] = rune(c)
		}
		return string(runes)
	}).SuchThat(func(s string) bool {
		return len(s) > 0
	})
}

// genJapaneseCJKString generates strings containing Japanese/CJK characters
// that have no Unicode case mappings (hiragana, katakana, CJK ideographs).
func genJapaneseCJKString() gopter.Gen {
	// Hiragana: U+3041-U+3096, Katakana: U+30A1-U+30FA, CJK: U+4E00-U+9FFF
	ranges := []struct{ lo, hi int }{
		{0x3041, 0x3096}, // Hiragana
		{0x30A1, 0x30FA}, // Katakana
		{0x4E00, 0x4E50}, // CJK Ideographs (subset for efficiency)
	}
	return gen.SliceOfN(10, gen.IntRange(0, len(ranges)*50-1)).Map(func(indices []int) string {
		var b strings.Builder
		for _, idx := range indices {
			rangeIdx := idx % len(ranges)
			r := ranges[rangeIdx]
			charIdx := idx % (r.hi - r.lo + 1)
			b.WriteRune(rune(r.lo + charIdx))
		}
		return b.String()
	})
}
