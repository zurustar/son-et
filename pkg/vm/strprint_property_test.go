package vm

import (
	"fmt"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// Property-based tests for StrPrint built-in function.
// These tests verify the correctness properties defined in the design document.

// TestProperty1_StrPrintFormatConversionCorrectness tests that StrPrint correctly
// converts FILLY format specifiers (%ld, %lx) to Go format specifiers (%d, %x)
// and produces the same result as fmt.Sprintf.
//
// **Validates: Requirements 1.1, 1.2, 1.3, 1.4, 1.5**
// Feature: missing-builtin-functions, Property 1: StrPrint フォーマット変換の正確性
func TestProperty1_StrPrintFormatConversionCorrectness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: %ld format produces same result as fmt.Sprintf with %d
	// Requirement 1.2: System supports %ld format specifier for decimal integers, converting to Go's %d.
	properties.Property("%ld format produces same result as fmt.Sprintf with %d", prop.ForAll(
		func(n int64) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%ld", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%d", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: %lx format produces same result as fmt.Sprintf with %x
	// Requirement 1.3: System supports %lx format specifier for hexadecimal, converting to Go's %x.
	properties.Property("%lx format produces same result as fmt.Sprintf with %x", prop.ForAll(
		func(n int64) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%lx", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%x", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: %03ld format produces same result as fmt.Sprintf with %03d
	// Requirement 1.5: System supports width and padding specifiers like %03d.
	properties.Property("%03ld format produces same result as fmt.Sprintf with %03d", prop.ForAll(
		func(n int64) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%03ld", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%03d", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: %05ld format produces same result as fmt.Sprintf with %05d
	// Requirement 1.5: System supports width and padding specifiers like %03d.
	properties.Property("%05ld format produces same result as fmt.Sprintf with %05d", prop.ForAll(
		func(n int64) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%05ld", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%05d", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: %08lx format produces same result as fmt.Sprintf with %08x
	// Requirement 1.5: System supports width and padding specifiers like %03d.
	properties.Property("%08lx format produces same result as fmt.Sprintf with %08x", prop.ForAll(
		func(n int64) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%08lx", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%08x", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: %s format produces same result as fmt.Sprintf with %s
	// Requirement 1.4: System supports %s format specifier for strings.
	properties.Property("%s format produces same result as fmt.Sprintf with %s", prop.ForAll(
		func(s string) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%s", s})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%s", s)
			return result == expected
		},
		gen.AnyString(),
	))

	// Property: Multiple format specifiers produce correct result
	// Requirement 1.1: When StrPrint is called with format string and arguments, system returns formatted string.
	properties.Property("multiple format specifiers produce correct result", prop.ForAll(
		func(s string, n1 int64, n2 int64) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%s: %ld (0x%lx)", s, n1, n2})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%s: %d (0x%x)", s, n1, n2)
			return result == expected
		},
		gen.AnyString(),
		gen.Int64(),
		gen.Int64(),
	))

	// Property: ROBOT sample use case - filename generation
	// Requirement 1.1, 1.5: StrPrint("ROBOT%03d.BMP", i) produces correct filename
	properties.Property("ROBOT filename generation produces correct result", prop.ForAll(
		func(n int64) bool {
			// Limit to reasonable range for filename numbers
			if n < 0 {
				n = -n
			}
			if n > 999 {
				n = n % 1000
			}

			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"ROBOT%03d.BMP", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("ROBOT%03d.BMP", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: Negative numbers are handled correctly with %ld
	// Requirement 1.2: System supports %ld format specifier for decimal integers.
	properties.Property("negative numbers are handled correctly with %ld", prop.ForAll(
		func(n int64) bool {
			// Ensure negative number
			if n >= 0 {
				n = -n - 1
			}

			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			result, err := fn(vm, []any{"%ld", n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf("%d", n)
			return result == expected
		},
		gen.Int64(),
	))

	// Property: Zero is handled correctly
	// Requirement 1.2, 1.3: System supports %ld and %lx format specifiers.
	properties.Property("zero is handled correctly with %ld and %lx", prop.ForAll(
		func(_ bool) bool {
			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			// Test %ld with zero
			result1, err := fn(vm, []any{"%ld", int64(0)})
			if err != nil || result1 != "0" {
				return false
			}

			// Test %lx with zero
			result2, err := fn(vm, []any{"%lx", int64(0)})
			if err != nil || result2 != "0" {
				return false
			}

			// Test %03ld with zero
			result3, err := fn(vm, []any{"%03ld", int64(0)})
			if err != nil || result3 != "000" {
				return false
			}

			return true
		},
		gen.Bool(),
	))

	// Property: Width specifier with various widths produces correct padding
	// Requirement 1.5: System supports width and padding specifiers.
	properties.Property("width specifier produces correct padding", prop.ForAll(
		func(n int64, width int) bool {
			// Limit width to reasonable range
			if width < 1 {
				width = 1
			}
			if width > 20 {
				width = 20
			}

			vm := New([]opcode.OpCode{})
			fn := vm.builtins["StrPrint"]

			// Create format string with width
			format := fmt.Sprintf("%%0%dld", width)
			goFormat := fmt.Sprintf("%%0%dd", width)

			result, err := fn(vm, []any{format, n})
			if err != nil {
				return false
			}

			expected := fmt.Sprintf(goFormat, n)
			return result == expected
		},
		gen.Int64(),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
