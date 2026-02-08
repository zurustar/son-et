package vm

import "fmt"

// registerArrayBuiltins registers array-related built-in functions.
func (vm *VM) registerArrayBuiltins() {
	// ArraySize: Get the number of elements in an array
	// Requirement 4.1: Return the current element count as an integer
	// Requirement 4.2: Return 0 for empty arrays
	// Requirement 4.3: Return accurate count after auto-expansion
	vm.RegisterBuiltinFunction("ArraySize", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return int64(0), nil
		}

		arr, ok := args[0].(*Array)
		if !ok {
			return int64(0), nil
		}

		result := int64(arr.Len())
		v.log.Debug("ArraySize called", "size", result)
		return result, nil
	})

	// DelArrayAll: Delete all elements from an array
	// Requirement 5.1: Delete all elements and set array size to 0
	// Requirement 5.2: No error on empty array
	// Requirement 5.3: Array can be reused after clearing
	vm.RegisterBuiltinFunction("DelArrayAll", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("DelArrayAll requires 1 argument (array), got %d", len(args))
		}

		arr, ok := args[0].(*Array)
		if !ok {
			return nil, fmt.Errorf("DelArrayAll argument must be an array, got %T", args[0])
		}

		arr.Clear()
		v.log.Debug("DelArrayAll called")
		return int64(0), nil
	})

	// DelArrayAt: Delete element at specified index (splice operation)
	// Requirement 6.1: Delete element at index, shift subsequent elements forward, decrease size by 1
	// Requirement 6.2: Return error for out-of-range index
	// Requirement 6.3: Delete last element without shift
	vm.RegisterBuiltinFunction("DelArrayAt", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return nil, fmt.Errorf("DelArrayAt requires 2 arguments (array, index), got %d", len(args))
		}

		arr, ok := args[0].(*Array)
		if !ok {
			return nil, fmt.Errorf("DelArrayAt first argument must be an array, got %T", args[0])
		}

		index, ok := toInt64(args[1])
		if !ok {
			return nil, fmt.Errorf("DelArrayAt index must be integer, got %T", args[1])
		}

		if err := arr.DeleteAt(index); err != nil {
			return nil, err
		}

		v.log.Debug("DelArrayAt called", "index", index)
		return int64(0), nil
	})

	// InsArrayAt: Insert element at specified index (splice operation)
	// Requirement 7.1: Insert value at index, shift existing elements backward, increase size by 1
	// Requirement 7.2: Return error for out-of-range index
	// Requirement 7.3: Insert at head shifts all existing elements
	// Requirement 7.4: Insert at end (index=size) appends to array
	vm.RegisterBuiltinFunction("InsArrayAt", func(v *VM, args []any) (any, error) {
		if len(args) < 3 {
			return nil, fmt.Errorf("InsArrayAt requires 3 arguments (array, index, value), got %d", len(args))
		}

		arr, ok := args[0].(*Array)
		if !ok {
			return nil, fmt.Errorf("InsArrayAt first argument must be an array, got %T", args[0])
		}

		index, ok := toInt64(args[1])
		if !ok {
			return nil, fmt.Errorf("InsArrayAt index must be integer, got %T", args[1])
		}

		value := args[2]

		if err := arr.InsertAt(index, value); err != nil {
			return nil, err
		}

		v.log.Debug("InsArrayAt called", "index", index, "value", value)
		return int64(0), nil
	})
}
