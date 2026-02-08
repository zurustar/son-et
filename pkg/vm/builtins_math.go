package vm

import (
	"fmt"
	"math/rand/v2"
)

// registerMathBuiltins registers math-related built-in functions.
func (vm *VM) registerMathBuiltins() {
	// Random: Generate random number
	// Random(max) - returns random number from 0 to max-1
	// Random(min, max) - returns random number from min to max-1
	vm.RegisterBuiltinFunction("Random", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return int64(0), fmt.Errorf("Random requires at least 1 argument (max)")
		}

		var min, max int64
		if len(args) == 1 {
			min = 0
			if m, ok := toInt64(args[0]); ok {
				max = m
			} else {
				return int64(0), fmt.Errorf("Random: max must be integer")
			}
		} else {
			if m, ok := toInt64(args[0]); ok {
				min = m
			} else {
				return int64(0), fmt.Errorf("Random: min must be integer")
			}
			if m, ok := toInt64(args[1]); ok {
				max = m
			} else {
				return int64(0), fmt.Errorf("Random: max must be integer")
			}
		}

		if max <= min {
			return min, nil
		}

		// Generate random number in range [min, max)
		result := min + int64(rand.IntN(int(max-min)))
		return result, nil
	})

	// MakeLong: Combine two 16-bit values into a 32-bit value
	// MakeLong(low_word, high_word) = (high_word << 16) | (low_word & 0xFFFF)
	vm.RegisterBuiltinFunction("MakeLong", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return int64(0), fmt.Errorf("MakeLong requires 2 arguments (low_word, high_word)")
		}

		low, ok := toInt64(args[0])
		if !ok {
			return int64(0), fmt.Errorf("MakeLong: low_word must be integer")
		}

		high, ok := toInt64(args[1])
		if !ok {
			return int64(0), fmt.Errorf("MakeLong: high_word must be integer")
		}

		result := ((high & 0xFFFF) << 16) | (low & 0xFFFF)
		return result, nil
	})

	// GetHiWord: Extract the upper 16 bits from a 32-bit value
	// GetHiWord(value) = (value >> 16) & 0xFFFF
	vm.RegisterBuiltinFunction("GetHiWord", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return int64(0), fmt.Errorf("GetHiWord requires 1 argument (long_value)")
		}

		value, ok := toInt64(args[0])
		if !ok {
			return int64(0), fmt.Errorf("GetHiWord: argument must be integer")
		}

		result := (value >> 16) & 0xFFFF
		return result, nil
	})

	// GetLowWord: Extract the lower 16 bits from a 32-bit value
	// GetLowWord(value) = value & 0xFFFF
	vm.RegisterBuiltinFunction("GetLowWord", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return int64(0), fmt.Errorf("GetLowWord requires 1 argument (long_value)")
		}

		value, ok := toInt64(args[0])
		if !ok {
			return int64(0), fmt.Errorf("GetLowWord: argument must be integer")
		}

		result := value & 0xFFFF
		return result, nil
	})
}
