package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// Feature: utility-builtins, Property 10: MakeLong/GetHiWord/GetLowWordラウンドトリップ
// 任意の2つの16ビット値（0〜65535）に対して、MakeLongで結合した結果から
// GetHiWordとGetLowWordで分解すると、元の値が復元される。
// すなわち GetLowWord(MakeLong(low, high)) == low かつ
// GetHiWord(MakeLong(low, high)) == high
// **Validates: Requirements 8.1, 9.1, 10.1**
func TestProperty10_MakeLongGetHiWordGetLowWordRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: GetLowWord(MakeLong(low, high)) == low
	properties.Property("GetLowWord(MakeLong(low, high)) == low", prop.ForAll(
		func(low, high int64) bool {
			vm := New([]opcode.OpCode{})

			// Combine low and high into a 32-bit value
			makeLongResult, err := vm.builtins["MakeLong"](vm, []any{low, high})
			if err != nil {
				return false
			}

			// Extract the low word
			getLowResult, err := vm.builtins["GetLowWord"](vm, []any{makeLongResult})
			if err != nil {
				return false
			}

			resultInt, ok := getLowResult.(int64)
			if !ok {
				return false
			}

			return resultInt == low
		},
		gen.Int64Range(0, 65535),
		gen.Int64Range(0, 65535),
	))

	// Property: GetHiWord(MakeLong(low, high)) == high
	properties.Property("GetHiWord(MakeLong(low, high)) == high", prop.ForAll(
		func(low, high int64) bool {
			vm := New([]opcode.OpCode{})

			// Combine low and high into a 32-bit value
			makeLongResult, err := vm.builtins["MakeLong"](vm, []any{low, high})
			if err != nil {
				return false
			}

			// Extract the high word
			getHiResult, err := vm.builtins["GetHiWord"](vm, []any{makeLongResult})
			if err != nil {
				return false
			}

			resultInt, ok := getHiResult.(int64)
			if !ok {
				return false
			}

			return resultInt == high
		},
		gen.Int64Range(0, 65535),
		gen.Int64Range(0, 65535),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
