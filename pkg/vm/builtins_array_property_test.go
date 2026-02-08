package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// Feature: utility-builtins, Property 5: ArraySizeの正確性
// 任意の整数値のスライスから作成した配列に対して、ArraySizeの結果はスライスの長さと一致する
// **Validates: Requirements 4.1, 4.3**
func TestProperty5_ArraySizeAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("ArraySize matches slice length", prop.ForAll(
		func(values []int64) bool {
			vm := New([]opcode.OpCode{})

			// Create array from slice of int64 values
			elements := make([]any, len(values))
			for i, v := range values {
				elements[i] = v
			}
			arr := NewArrayFromSlice(elements)

			result, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}

			resultInt, ok := result.(int64)
			if !ok {
				return false
			}

			return resultInt == int64(len(values))
		},
		gen.SliceOf(gen.Int64()),
	))

	properties.Property("ArraySize matches length after auto-expansion", prop.ForAll(
		func(initialSize int, expandIndex int64) bool {
			if initialSize < 0 {
				initialSize = 0
			}
			if initialSize > 100 {
				initialSize = 100
			}
			if expandIndex < int64(initialSize) || expandIndex > 500 {
				return true // skip invalid cases
			}

			vm := New([]opcode.OpCode{})

			arr := NewArray(initialSize)
			arr.Set(expandIndex, int64(42))

			result, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}

			resultInt, ok := result.(int64)
			if !ok {
				return false
			}

			expectedSize := int64(expandIndex) + 1
			return resultInt == expectedSize
		},
		gen.IntRange(0, 50),
		gen.Int64Range(50, 200),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 6: DelArrayAll後のサイズゼロと再利用
// 任意の配列に対して、DelArrayAll実行後にArraySizeは0を返し、その後要素を追加すると正常に格納される
// **Validates: Requirements 5.1, 5.2, 5.3**
func TestProperty6_DelArrayAllSizeZeroAndReuse(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("DelArrayAll sets size to zero and array is reusable", prop.ForAll(
		func(values []int64, newValue int64) bool {
			vm := New([]opcode.OpCode{})

			// Create array from slice
			elements := make([]any, len(values))
			for i, v := range values {
				elements[i] = v
			}
			arr := NewArrayFromSlice(elements)

			// Call DelArrayAll
			_, err := vm.builtins["DelArrayAll"](vm, []any{arr})
			if err != nil {
				return false
			}

			// ArraySize should be 0
			sizeResult, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}
			sizeInt, ok := sizeResult.(int64)
			if !ok {
				return false
			}
			if sizeInt != 0 {
				return false
			}

			// Reuse: add an element via Set (auto-expand)
			arr.Set(0, newValue)

			// Verify the element was stored correctly
			val, ok := arr.Get(0)
			if !ok {
				return false
			}
			if val != newValue {
				return false
			}

			// Verify size is now 1
			sizeResult2, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}
			sizeInt2, ok := sizeResult2.(int64)
			if !ok {
				return false
			}
			return sizeInt2 == 1
		},
		gen.SliceOf(gen.Int64()),
		gen.Int64(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 7: DelArrayAtのsplice動作
// 任意の非空配列と有効なインデックスに対して、DelArrayAt実行後に配列サイズが1減少し、
// 削除位置より前の要素は変更されず、削除位置以降の要素は1つ前にシフトされる
// **Validates: Requirements 6.1, 6.3**
func TestProperty7_DelArrayAtSpliceBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("DelArrayAt splice: size decreases by 1, elements shift correctly", prop.ForAll(
		func(values []int64, indexFraction float64) bool {
			if len(values) == 0 {
				return true // skip empty arrays
			}

			// Derive a valid index within [0, len-1]
			delIndex := int64(indexFraction * float64(len(values)))
			if delIndex < 0 {
				delIndex = 0
			}
			if delIndex >= int64(len(values)) {
				delIndex = int64(len(values)) - 1
			}

			vm := New([]opcode.OpCode{})

			// Create array
			elements := make([]any, len(values))
			for i, v := range values {
				elements[i] = v
			}
			arr := NewArrayFromSlice(elements)

			originalSize := int64(arr.Len())

			// Snapshot original elements
			originalElements := make([]int64, len(values))
			copy(originalElements, values)

			// Call DelArrayAt
			_, err := vm.builtins["DelArrayAt"](vm, []any{arr, delIndex})
			if err != nil {
				return false
			}

			// Size should decrease by 1
			newSize, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}
			newSizeInt, ok := newSize.(int64)
			if !ok {
				return false
			}
			if newSizeInt != originalSize-1 {
				return false
			}

			// Elements before delIndex should be unchanged
			for i := int64(0); i < delIndex; i++ {
				val, ok := arr.Get(i)
				if !ok {
					return false
				}
				if val != originalElements[i] {
					return false
				}
			}

			// Elements at delIndex and after should be shifted left by 1
			for i := delIndex; i < newSizeInt; i++ {
				val, ok := arr.Get(i)
				if !ok {
					return false
				}
				if val != originalElements[i+1] {
					return false
				}
			}

			return true
		},
		genNonEmptyInt64Slice(),
		gen.Float64Range(0.0, 0.999),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 8: InsArrayAtのsplice動作
// 任意の配列、有効なインデックス、値に対して、InsArrayAt実行後に配列サイズが1増加し、
// 挿入位置に指定した値が存在し、挿入位置より前の要素は変更されず、
// 挿入位置以降の元の要素は1つ後ろにシフトされる
// **Validates: Requirements 7.1, 7.3, 7.4**
func TestProperty8_InsArrayAtSpliceBehavior(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("InsArrayAt splice: size increases by 1, value at index, elements shift correctly", prop.ForAll(
		func(values []int64, indexFraction float64, insertValue int64) bool {
			// Derive a valid index within [0, len] (inclusive of len for append)
			insIndex := int64(indexFraction * float64(len(values)+1))
			if insIndex < 0 {
				insIndex = 0
			}
			if insIndex > int64(len(values)) {
				insIndex = int64(len(values))
			}

			vm := New([]opcode.OpCode{})

			// Create array
			elements := make([]any, len(values))
			for i, v := range values {
				elements[i] = v
			}
			arr := NewArrayFromSlice(elements)

			originalSize := int64(arr.Len())

			// Snapshot original elements
			originalElements := make([]int64, len(values))
			copy(originalElements, values)

			// Call InsArrayAt
			_, err := vm.builtins["InsArrayAt"](vm, []any{arr, insIndex, insertValue})
			if err != nil {
				return false
			}

			// Size should increase by 1
			newSize, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}
			newSizeInt, ok := newSize.(int64)
			if !ok {
				return false
			}
			if newSizeInt != originalSize+1 {
				return false
			}

			// Value at insIndex should be the inserted value
			val, ok := arr.Get(insIndex)
			if !ok {
				return false
			}
			if val != insertValue {
				return false
			}

			// Elements before insIndex should be unchanged
			for i := int64(0); i < insIndex; i++ {
				v, ok := arr.Get(i)
				if !ok {
					return false
				}
				if v != originalElements[i] {
					return false
				}
			}

			// Elements after insIndex should be the original elements shifted right by 1
			for i := insIndex; i < originalSize; i++ {
				v, ok := arr.Get(i + 1)
				if !ok {
					return false
				}
				if v != originalElements[i] {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.Int64()),
		gen.Float64Range(0.0, 0.999),
		gen.Int64(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: utility-builtins, Property 9: InsArrayAt/DelArrayAtラウンドトリップ
// 任意の配列、有効なインデックス、値に対して、InsArrayAtで挿入した後に
// 同じインデックスでDelArrayAtを実行すると、元の配列と同一の内容に戻る
// **Validates: Requirements 6.1, 7.1**
func TestProperty9_InsArrayAtDelArrayAtRoundTrip(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("InsArrayAt then DelArrayAt at same index restores original array", prop.ForAll(
		func(values []int64, indexFraction float64, insertValue int64) bool {
			// Derive a valid index within [0, len] for insert
			insIndex := int64(indexFraction * float64(len(values)+1))
			if insIndex < 0 {
				insIndex = 0
			}
			if insIndex > int64(len(values)) {
				insIndex = int64(len(values))
			}

			vm := New([]opcode.OpCode{})

			// Create array
			elements := make([]any, len(values))
			for i, v := range values {
				elements[i] = v
			}
			arr := NewArrayFromSlice(elements)

			// Snapshot original
			originalElements := make([]int64, len(values))
			copy(originalElements, values)
			originalSize := int64(len(values))

			// InsArrayAt
			_, err := vm.builtins["InsArrayAt"](vm, []any{arr, insIndex, insertValue})
			if err != nil {
				return false
			}

			// DelArrayAt at the same index
			_, err = vm.builtins["DelArrayAt"](vm, []any{arr, insIndex})
			if err != nil {
				return false
			}

			// Size should be restored
			sizeResult, err := vm.builtins["ArraySize"](vm, []any{arr})
			if err != nil {
				return false
			}
			sizeInt, ok := sizeResult.(int64)
			if !ok {
				return false
			}
			if sizeInt != originalSize {
				return false
			}

			// All elements should match original
			for i := int64(0); i < originalSize; i++ {
				val, ok := arr.Get(i)
				if !ok {
					return false
				}
				if val != originalElements[i] {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.Int64()),
		gen.Float64Range(0.0, 0.999),
		gen.Int64(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// genNonEmptyInt64Slice generates non-empty slices of int64 values.
func genNonEmptyInt64Slice() gopter.Gen {
	return gen.SliceOfN(50, gen.Int64()).SuchThat(func(s []int64) bool {
		return len(s) > 0
	})
}
