// Package vm provides OpCode execution for the FILLY virtual machine.
package vm

import (
	"fmt"
	"math"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler"
)

// evaluateValue evaluates a value that may be a Variable, OpCode, or literal.
// It recursively resolves variables and executes nested OpCodes.
//
// Parameters:
//   - value: The value to evaluate (Variable, OpCode, or literal)
//
// Returns:
//   - any: The evaluated value
//   - error: Any error that occurred during evaluation
func (vm *VM) evaluateValue(value any) (any, error) {
	if value == nil {
		return nil, nil
	}

	switch v := value.(type) {
	case compiler.Variable:
		// Resolve variable from scope
		// Requirement 9.5: When variable is accessed, system searches local scope first, then global scope.
		resolved, ok := vm.GetCurrentScope().Get(string(v))
		if !ok {
			// Requirement 11.5: When variable is not found, system creates it with default value.
			vm.log.Warn("Variable not found, using default value 0", "name", string(v))
			return int64(0), nil
		}
		return resolved, nil

	case compiler.OpCode:
		// Execute nested OpCode
		return vm.Execute(v)

	default:
		// Return literal value as-is
		return v, nil
	}
}

// toInt64 converts a value to int64.
// Handles int, int64, float64, and string types.
func toInt64(v any) (int64, bool) {
	switch val := v.(type) {
	case int:
		return int64(val), true
	case int64:
		return val, true
	case float64:
		return int64(val), true
	case string:
		// Try to parse as number
		var i int64
		if _, err := fmt.Sscanf(val, "%d", &i); err == nil {
			return i, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toFloat64 converts a value to float64.
// Handles int, int64, float64, and string types.
func toFloat64(v any) (float64, bool) {
	switch val := v.(type) {
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	case float64:
		return val, true
	case string:
		var f float64
		if _, err := fmt.Sscanf(val, "%f", &f); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

// toString converts a value to string.
func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case int:
		return fmt.Sprintf("%d", val)
	case int64:
		return fmt.Sprintf("%d", val)
	case float64:
		return fmt.Sprintf("%g", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// toBool converts a value to bool.
// Non-zero numbers and non-empty strings are true.
func toBool(v any) bool {
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return v != nil
	}
}

// isNumeric checks if a value is numeric (int, int64, or float64).
func isNumeric(v any) bool {
	switch v.(type) {
	case int, int64, float64:
		return true
	default:
		return false
	}
}

// isFloat checks if a value is a float64.
func isFloat(v any) bool {
	_, ok := v.(float64)
	return ok
}

// executeAssign executes an OpAssign OpCode.
// OpAssign assigns a value to a variable.
// Args: [Variable(name), value]
//
// Requirement 8.2: When OpAssign is executed, system assigns value to specified variable.
func (vm *VM) executeAssign(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpAssign requires 2 arguments, got %d", len(opcode.Args))
	}

	// Get variable name
	varName, ok := opcode.Args[0].(compiler.Variable)
	if !ok {
		return nil, fmt.Errorf("OpAssign first argument must be Variable, got %T", opcode.Args[0])
	}

	// Evaluate the value
	value, err := vm.evaluateValue(opcode.Args[1])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate value: %w", err)
	}

	// Set the variable in the current scope
	// Requirement 9.6: When variable is assigned without prior declaration, system creates it in current scope.
	vm.GetCurrentScope().Set(string(varName), value)

	vm.log.Debug("Variable assigned", "name", string(varName), "value", value)
	return value, nil
}

// executeArrayAssign executes an OpArrayAssign OpCode.
// OpArrayAssign assigns a value to an array element.
// Args: [Variable(arrayName), index, value]
//
// Requirement 8.3: When OpArrayAssign is executed, system assigns value to specified array element.
// Requirement 19.5: When array index exceeds array size, system automatically expands array.
func (vm *VM) executeArrayAssign(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 3 {
		return nil, fmt.Errorf("OpArrayAssign requires 3 arguments, got %d", len(opcode.Args))
	}

	// Get array name
	arrayName, ok := opcode.Args[0].(compiler.Variable)
	if !ok {
		return nil, fmt.Errorf("OpArrayAssign first argument must be Variable, got %T", opcode.Args[0])
	}

	// Evaluate the index
	indexVal, err := vm.evaluateValue(opcode.Args[1])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate index: %w", err)
	}
	index, ok := toInt64(indexVal)
	if !ok {
		return nil, fmt.Errorf("array index must be numeric, got %T", indexVal)
	}

	// Requirement 19.4: When array index is negative, system logs error and returns zero.
	if index < 0 {
		vm.log.Error("Negative array index", "array", string(arrayName), "index", index)
		return int64(0), nil
	}

	// Evaluate the value
	value, err := vm.evaluateValue(opcode.Args[2])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate value: %w", err)
	}

	// Get or create the array
	scope := vm.GetCurrentScope()
	arrayVal, exists := scope.Get(string(arrayName))

	var arr *Array
	if exists {
		switch v := arrayVal.(type) {
		case *Array:
			// Already an Array type - use it directly
			// Requirement 19.8: When array is passed to function, system passes it by reference.
			arr = v
		case []any:
			// Legacy slice - convert to Array
			arr = NewArrayFromSlice(v)
			// Update the scope with the new Array type
			scope.Set(string(arrayName), arr)
		default:
			// Variable exists but is not an array - create new array
			arr = NewArray(int(index) + 1)
			scope.Set(string(arrayName), arr)
		}
	} else {
		// Create new array
		// Requirement 19.1: When array is declared, system allocates storage for array.
		arr = NewArray(int(index) + 1)
		scope.Set(string(arrayName), arr)
	}

	// Set the value (Array.Set handles expansion and zero initialization)
	// Requirement 19.3: When array element is assigned, system stores value at specified index.
	// Requirement 19.5: When array index exceeds array size, system automatically expands array.
	// Requirement 19.6: System supports dynamic array resizing.
	// Requirement 19.7: System initializes new array elements to zero.
	arr.Set(index, value)

	vm.log.Debug("Array element assigned", "array", string(arrayName), "index", index, "value", value)
	return value, nil
}

// executeCall executes an OpCall OpCode.
// OpCall invokes a function with arguments.
// Args: [functionName, arg1, arg2, ...]
//
// Requirement 8.4: When OpCall is executed, system calls specified function with arguments.
// Requirement 10.8: When unknown function is called, system logs error and continues execution.
func (vm *VM) executeCall(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 1 {
		return nil, fmt.Errorf("OpCall requires at least 1 argument (function name)")
	}

	// Get function name
	funcName, ok := opcode.Args[0].(string)
	if !ok {
		return nil, fmt.Errorf("OpCall first argument must be string, got %T", opcode.Args[0])
	}

	// Handle special "return" call
	if funcName == "return" {
		var returnValue any = int64(0)
		if len(opcode.Args) > 1 {
			val, err := vm.evaluateValue(opcode.Args[1])
			if err != nil {
				return nil, fmt.Errorf("failed to evaluate return value: %w", err)
			}
			returnValue = val
		}
		// Set return value in current stack frame
		if len(vm.callStack) > 0 {
			vm.callStack[len(vm.callStack)-1].ReturnValue = returnValue
		}
		// Return a special marker to indicate function return
		return &returnMarker{value: returnValue}, nil
	}

	// Evaluate arguments
	args := make([]any, 0, len(opcode.Args)-1)
	for i := 1; i < len(opcode.Args); i++ {
		val, err := vm.evaluateValue(opcode.Args[i])
		if err != nil {
			return nil, fmt.Errorf("failed to evaluate argument %d: %w", i, err)
		}
		args = append(args, val)
	}

	// Check for built-in function first
	if builtin, ok := vm.builtins[funcName]; ok {
		result, err := builtin(vm, args)
		if err != nil {
			vm.log.Error("Built-in function error", "function", funcName, "error", err)
			// Requirement 11.8: System continues execution after non-fatal errors.
			return int64(0), nil
		}
		return result, nil
	}

	// Check for case-insensitive built-in function match
	funcNameLower := strings.ToLower(funcName)
	for name, builtin := range vm.builtins {
		if strings.ToLower(name) == funcNameLower {
			result, err := builtin(vm, args)
			if err != nil {
				vm.log.Error("Built-in function error", "function", funcName, "error", err)
				return int64(0), nil
			}
			return result, nil
		}
	}

	// Check for user-defined function
	if userFunc, ok := vm.functions[funcName]; ok {
		return vm.callUserFunction(userFunc, args)
	}

	// Check for case-insensitive user function match
	for name, userFunc := range vm.functions {
		if strings.EqualFold(name, funcName) {
			return vm.callUserFunction(userFunc, args)
		}
	}

	// Requirement 10.8: When unknown function is called, system logs error and continues execution.
	// Requirement 11.6: When function is not found, system logs error and continues execution.
	vm.log.Warn("Unknown function called", "function", funcName)
	return int64(0), nil
}

// returnMarker is a special type to indicate a function return.
type returnMarker struct {
	value any
}

// callUserFunction calls a user-defined function.
// Requirement 20.1: When function is called, system pushes new stack frame.
// Requirement 20.2: When function returns, system pops stack frame.
// Requirement 9.3: When function is called, system creates new local scope.
// Requirement 9.4: When function returns, system destroys local scope.
func (vm *VM) callUserFunction(fn *FunctionDef, args []any) (any, error) {
	// Create new local scope
	localScope := NewScope(vm.globalScope)

	// Bind parameters to local scope
	// Requirement 9.7: When function parameters are passed, system binds them to local scope.
	for i, param := range fn.Parameters {
		var value any
		if i < len(args) {
			value = args[i]
		} else if param.HasDefault {
			value = param.Default
		} else {
			value = int64(0)
		}

		// Requirement 19.8: When array is passed to function, system passes it by reference.
		if param.IsArray {
			// Arrays are passed by reference (the slice itself)
			localScope.SetLocal(param.Name, value)
		} else {
			localScope.SetLocal(param.Name, value)
		}
	}

	// Push stack frame
	// Requirement 20.6: System detects stack overflow and reports error.
	if err := vm.PushStackFrame(fn.Name, localScope); err != nil {
		// Requirement 20.8: When stack overflow occurs, system logs error and terminates execution.
		return nil, err
	}

	// Execute function body
	var result any = int64(0)
	for _, op := range fn.Body {
		val, err := vm.Execute(op)
		if err != nil {
			// Check if this is a stack overflow error - these are fatal
			// Requirement 20.8: When stack overflow occurs, system logs error and terminates execution.
			if strings.Contains(err.Error(), "stack overflow") {
				vm.log.Error("Stack overflow in function", "function", fn.Name, "error", err)
				// Pop the current frame before returning the error
				vm.PopStackFrame()
				return nil, err
			}
			vm.log.Error("Error in function body", "function", fn.Name, "error", err)
		}

		// Check for return
		if ret, ok := val.(*returnMarker); ok {
			result = ret.value
			break
		}
	}

	// Pop stack frame
	// Requirement 20.2: When function returns, system pops stack frame.
	frame, err := vm.PopStackFrame()
	if err != nil {
		return nil, err
	}

	// Requirement 20.3: When function has return value, system passes it to caller.
	// Requirement 20.4: When function has no return value, system returns zero.
	if frame.ReturnValue != nil {
		result = frame.ReturnValue
	}

	return result, nil
}

// executeBinaryOp executes an OpBinaryOp OpCode.
// OpBinaryOp performs a binary operation (+, -, *, /, %, ==, !=, <, <=, >, >=, &&, ||).
// Args: [operator, leftOperand, rightOperand]
//
// Requirement 8.11: When OpBinaryOp is executed, system evaluates binary operation and returns result.
func (vm *VM) executeBinaryOp(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 3 {
		return nil, fmt.Errorf("OpBinaryOp requires 3 arguments, got %d", len(opcode.Args))
	}

	operator, ok := opcode.Args[0].(string)
	if !ok {
		return nil, fmt.Errorf("OpBinaryOp operator must be string, got %T", opcode.Args[0])
	}

	// Evaluate operands
	left, err := vm.evaluateValue(opcode.Args[1])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate left operand: %w", err)
	}

	right, err := vm.evaluateValue(opcode.Args[2])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate right operand: %w", err)
	}

	// Handle string concatenation
	if operator == "+" {
		_, leftIsString := left.(string)
		_, rightIsString := right.(string)
		if leftIsString || rightIsString {
			return toString(left) + toString(right), nil
		}
	}

	// Handle numeric operations
	switch operator {
	case "+", "-", "*", "/", "%":
		return vm.executeArithmeticOp(operator, left, right)
	case "==", "!=", "<", "<=", ">", ">=":
		return vm.executeComparisonOp(operator, left, right)
	case "&&", "||":
		return vm.executeLogicalOp(operator, left, right)
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", operator)
	}
}

// executeArithmeticOp executes arithmetic operations (+, -, *, /, %).
func (vm *VM) executeArithmeticOp(operator string, left, right any) (any, error) {
	// Determine if we should use float arithmetic
	useFloat := isFloat(left) || isFloat(right)

	if useFloat {
		leftF, ok := toFloat64(left)
		if !ok {
			return nil, fmt.Errorf("cannot convert left operand to float: %v", left)
		}
		rightF, ok := toFloat64(right)
		if !ok {
			return nil, fmt.Errorf("cannot convert right operand to float: %v", right)
		}

		switch operator {
		case "+":
			return leftF + rightF, nil
		case "-":
			return leftF - rightF, nil
		case "*":
			return leftF * rightF, nil
		case "/":
			// Requirement 11.3: When division by zero occurs, system logs error and returns zero.
			if rightF == 0 {
				vm.log.Error("Division by zero")
				return float64(0), nil
			}
			return leftF / rightF, nil
		case "%":
			// Requirement 11.3: When division by zero occurs, system logs error and returns zero.
			if rightF == 0 {
				vm.log.Error("Modulo by zero")
				return float64(0), nil
			}
			return math.Mod(leftF, rightF), nil
		}
	}

	// Integer arithmetic
	leftI, ok := toInt64(left)
	if !ok {
		return nil, fmt.Errorf("cannot convert left operand to int: %v", left)
	}
	rightI, ok := toInt64(right)
	if !ok {
		return nil, fmt.Errorf("cannot convert right operand to int: %v", right)
	}

	switch operator {
	case "+":
		return leftI + rightI, nil
	case "-":
		return leftI - rightI, nil
	case "*":
		return leftI * rightI, nil
	case "/":
		// Requirement 11.3: When division by zero occurs, system logs error and returns zero.
		if rightI == 0 {
			vm.log.Error("Division by zero")
			return int64(0), nil
		}
		return leftI / rightI, nil
	case "%":
		// Requirement 11.3: When division by zero occurs, system logs error and returns zero.
		if rightI == 0 {
			vm.log.Error("Modulo by zero")
			return int64(0), nil
		}
		return leftI % rightI, nil
	}

	return nil, fmt.Errorf("unknown arithmetic operator: %s", operator)
}

// executeComparisonOp executes comparison operations (==, !=, <, <=, >, >=).
func (vm *VM) executeComparisonOp(operator string, left, right any) (any, error) {
	// Handle string comparison
	leftStr, leftIsString := left.(string)
	rightStr, rightIsString := right.(string)
	if leftIsString && rightIsString {
		switch operator {
		case "==":
			return boolToInt(leftStr == rightStr), nil
		case "!=":
			return boolToInt(leftStr != rightStr), nil
		case "<":
			return boolToInt(leftStr < rightStr), nil
		case "<=":
			return boolToInt(leftStr <= rightStr), nil
		case ">":
			return boolToInt(leftStr > rightStr), nil
		case ">=":
			return boolToInt(leftStr >= rightStr), nil
		}
	}

	// Numeric comparison
	useFloat := isFloat(left) || isFloat(right)

	if useFloat {
		leftF, ok := toFloat64(left)
		if !ok {
			return nil, fmt.Errorf("cannot convert left operand to float: %v", left)
		}
		rightF, ok := toFloat64(right)
		if !ok {
			return nil, fmt.Errorf("cannot convert right operand to float: %v", right)
		}

		switch operator {
		case "==":
			return boolToInt(leftF == rightF), nil
		case "!=":
			return boolToInt(leftF != rightF), nil
		case "<":
			return boolToInt(leftF < rightF), nil
		case "<=":
			return boolToInt(leftF <= rightF), nil
		case ">":
			return boolToInt(leftF > rightF), nil
		case ">=":
			return boolToInt(leftF >= rightF), nil
		}
	}

	// Integer comparison
	leftI, ok := toInt64(left)
	if !ok {
		return nil, fmt.Errorf("cannot convert left operand to int: %v", left)
	}
	rightI, ok := toInt64(right)
	if !ok {
		return nil, fmt.Errorf("cannot convert right operand to int: %v", right)
	}

	switch operator {
	case "==":
		return boolToInt(leftI == rightI), nil
	case "!=":
		return boolToInt(leftI != rightI), nil
	case "<":
		return boolToInt(leftI < rightI), nil
	case "<=":
		return boolToInt(leftI <= rightI), nil
	case ">":
		return boolToInt(leftI > rightI), nil
	case ">=":
		return boolToInt(leftI >= rightI), nil
	}

	return nil, fmt.Errorf("unknown comparison operator: %s", operator)
}

// boolToInt converts a boolean to int64 (1 for true, 0 for false).
// FILLY uses integers for boolean values.
func boolToInt(b bool) int64 {
	if b {
		return 1
	}
	return 0
}

// executeLogicalOp executes logical operations (&&, ||).
func (vm *VM) executeLogicalOp(operator string, left, right any) (any, error) {
	leftBool := toBool(left)
	rightBool := toBool(right)

	switch operator {
	case "&&":
		return boolToInt(leftBool && rightBool), nil
	case "||":
		return boolToInt(leftBool || rightBool), nil
	}

	return nil, fmt.Errorf("unknown logical operator: %s", operator)
}

// executeUnaryOp executes an OpUnaryOp OpCode.
// OpUnaryOp performs a unary operation (-, !).
// Args: [operator, operand]
//
// Requirement 8.12: When OpUnaryOp is executed, system evaluates unary operation and returns result.
func (vm *VM) executeUnaryOp(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpUnaryOp requires 2 arguments, got %d", len(opcode.Args))
	}

	operator, ok := opcode.Args[0].(string)
	if !ok {
		return nil, fmt.Errorf("OpUnaryOp operator must be string, got %T", opcode.Args[0])
	}

	// Evaluate operand
	operand, err := vm.evaluateValue(opcode.Args[1])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate operand: %w", err)
	}

	switch operator {
	case "-":
		// Negation
		if f, ok := operand.(float64); ok {
			return -f, nil
		}
		if i, ok := toInt64(operand); ok {
			return -i, nil
		}
		return nil, fmt.Errorf("cannot negate non-numeric value: %v", operand)

	case "!":
		// Logical NOT
		return boolToInt(!toBool(operand)), nil

	default:
		return nil, fmt.Errorf("unknown unary operator: %s", operator)
	}
}

// executeArrayAccess executes an OpArrayAccess OpCode.
// OpArrayAccess accesses an array element by index.
// Args: [Variable(arrayName) or nested OpCode, index]
//
// Requirement 8.13: When OpArrayAccess is executed, system returns value at specified array index.
// Requirement 11.4: When array index is out of range, system logs error and returns zero.
func (vm *VM) executeArrayAccess(opcode compiler.OpCode) (any, error) {
	if len(opcode.Args) < 2 {
		return nil, fmt.Errorf("OpArrayAccess requires 2 arguments, got %d", len(opcode.Args))
	}

	// Evaluate the array (could be a Variable or nested expression)
	arrayVal, err := vm.evaluateValue(opcode.Args[0])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate array: %w", err)
	}

	// Evaluate the index
	indexVal, err := vm.evaluateValue(opcode.Args[1])
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate index: %w", err)
	}

	index, ok := toInt64(indexVal)
	if !ok {
		return nil, fmt.Errorf("array index must be numeric, got %T", indexVal)
	}

	// Requirement 19.4: When array index is negative, system logs error and returns zero.
	if index < 0 {
		vm.log.Error("Negative array index", "index", index)
		return int64(0), nil
	}

	// Handle Array type (new reference-based array)
	if arr, ok := arrayVal.(*Array); ok {
		val, found := arr.Get(index)
		if !found {
			// Requirement 11.4: When array index is out of range, system logs error and returns zero.
			vm.log.Error("Array index out of range", "index", index, "length", arr.Len())
		}
		return val, nil
	}

	// Handle legacy []any slice
	arr, ok := arrayVal.([]any)
	if !ok {
		// If it's not an array, treat it as a single-element array
		if index == 0 {
			return arrayVal, nil
		}
		// Requirement 11.4: When array index is out of range, system logs error and returns zero.
		vm.log.Error("Array index out of range", "index", index, "length", 1)
		return int64(0), nil
	}

	// Requirement 11.4: When array index is out of range, system logs error and returns zero.
	if int(index) >= len(arr) {
		vm.log.Error("Array index out of range", "index", index, "length", len(arr))
		return int64(0), nil
	}

	return arr[index], nil
}
