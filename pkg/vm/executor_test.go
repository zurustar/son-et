package vm

import (
	"strings"
	"testing"

	"github.com/zurustar/son-et/pkg/opcode"
)

// TestExecuteAssign tests the OpAssign execution.
func TestExecuteAssign(t *testing.T) {
	t.Run("assigns integer value to variable", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x"), int64(42)},
		}

		result, err := vm.executeAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(42) {
			t.Errorf("expected result 42, got %v", result)
		}

		// Verify variable is set in scope
		val, ok := vm.GetCurrentScope().Get("x")
		if !ok {
			t.Error("expected variable 'x' to be set")
		}
		if val != int64(42) {
			t.Errorf("expected variable value 42, got %v", val)
		}
	})

	t.Run("assigns string value to variable", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("name"), "hello"},
		}

		_, err := vm.executeAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("name")
		if val != "hello" {
			t.Errorf("expected 'hello', got %v", val)
		}
	})

	t.Run("assigns result of binary operation", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// x = 10 + 5
		opcode := opcode.OpCode{
			Cmd: opcode.Assign,
			Args: []any{
				opcode.Variable("x"),
				opcode.OpCode{
					Cmd:  opcode.BinaryOp,
					Args: []any{"+", int64(10), int64(5)},
				},
			},
		}

		_, err := vm.executeAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("x")
		if val != int64(15) {
			t.Errorf("expected 15, got %v", val)
		}
	})

	t.Run("assigns value from another variable", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("y", int64(100))

		// x = y
		opcode := opcode.OpCode{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x"), opcode.Variable("y")},
		}

		_, err := vm.executeAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("x")
		if val != int64(100) {
			t.Errorf("expected 100, got %v", val)
		}
	})

	t.Run("returns error for insufficient arguments", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x")},
		}

		_, err := vm.executeAssign(opcode)
		if err == nil {
			t.Error("expected error for insufficient arguments")
		}
	})
}

// TestExecuteBinaryOp tests the OpBinaryOp execution.
func TestExecuteBinaryOp(t *testing.T) {
	t.Run("addition", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"+", int64(10), int64(5)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(15) {
			t.Errorf("expected 15, got %v", result)
		}
	})

	t.Run("subtraction", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"-", int64(10), int64(3)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(7) {
			t.Errorf("expected 7, got %v", result)
		}
	})

	t.Run("multiplication", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"*", int64(6), int64(7)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(42) {
			t.Errorf("expected 42, got %v", result)
		}
	})

	t.Run("division", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"/", int64(20), int64(4)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(5) {
			t.Errorf("expected 5, got %v", result)
		}
	})

	t.Run("division by zero returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"/", int64(10), int64(0)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for division by zero, got %v", result)
		}
	})

	t.Run("modulo", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"%", int64(17), int64(5)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(2) {
			t.Errorf("expected 2, got %v", result)
		}
	})

	t.Run("comparison equal", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"==", int64(5), int64(5)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})

	t.Run("comparison not equal", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"!=", int64(5), int64(3)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})

	t.Run("comparison less than", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"<", int64(3), int64(5)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})

	t.Run("comparison greater than", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{">", int64(10), int64(5)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})

	t.Run("logical AND", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"&&", int64(1), int64(1)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})

	t.Run("logical OR", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"||", int64(0), int64(1)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})

	t.Run("string concatenation", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"+", "hello", " world"},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello world" {
			t.Errorf("expected 'hello world', got %v", result)
		}
	})

	t.Run("float arithmetic", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"+", float64(1.5), float64(2.5)},
		}

		result, err := vm.executeBinaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != float64(4.0) {
			t.Errorf("expected 4.0, got %v", result)
		}
	})
}

// TestExecuteUnaryOp tests the OpUnaryOp execution.
func TestExecuteUnaryOp(t *testing.T) {
	t.Run("negation of integer", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.UnaryOp,
			Args: []any{"-", int64(42)},
		}

		result, err := vm.executeUnaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(-42) {
			t.Errorf("expected -42, got %v", result)
		}
	})

	t.Run("negation of float", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.UnaryOp,
			Args: []any{"-", float64(3.14)},
		}

		result, err := vm.executeUnaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != float64(-3.14) {
			t.Errorf("expected -3.14, got %v", result)
		}
	})

	t.Run("logical NOT of true", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.UnaryOp,
			Args: []any{"!", int64(1)},
		}

		result, err := vm.executeUnaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 (false), got %v", result)
		}
	})

	t.Run("logical NOT of false", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.UnaryOp,
			Args: []any{"!", int64(0)},
		}

		result, err := vm.executeUnaryOp(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(1) {
			t.Errorf("expected 1 (true), got %v", result)
		}
	})
}

// TestExecuteArrayAssign tests the OpArrayAssign execution.
func TestExecuteArrayAssign(t *testing.T) {
	t.Run("assigns value to new array", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(0), int64(42)},
		}

		_, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, ok := vm.GetCurrentScope().Get("arr")
		if !ok {
			t.Error("expected array 'arr' to be set")
		}
		arr, ok := val.(*Array)
		if !ok {
			t.Fatalf("expected *Array, got %T", val)
		}
		elem, _ := arr.Get(0)
		if elem != int64(42) {
			t.Errorf("expected arr[0] = 42, got %v", elem)
		}
	})

	t.Run("auto-expands array", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// First assign at index 0
		opcode1 := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(0), int64(1)},
		}
		vm.executeArrayAssign(opcode1)

		// Then assign at index 5 (should auto-expand)
		opcode2 := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(5), int64(100)},
		}
		_, err := vm.executeArrayAssign(opcode2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("arr")
		arr := val.(*Array)
		if arr.Len() < 6 {
			t.Errorf("expected array length >= 6, got %d", arr.Len())
		}
		elem5, _ := arr.Get(5)
		if elem5 != int64(100) {
			t.Errorf("expected arr[5] = 100, got %v", elem5)
		}
		// Check that intermediate elements are initialized to zero
		elem2, _ := arr.Get(2)
		if elem2 != int64(0) {
			t.Errorf("expected arr[2] = 0, got %v", elem2)
		}
	})

	t.Run("negative index returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAssign,
			Args: []any{opcode.Variable("arr"), int64(-1), int64(42)},
		}

		result, err := vm.executeArrayAssign(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for negative index, got %v", result)
		}
	})
}

// TestExecuteArrayAccess tests the OpArrayAccess execution.
func TestExecuteArrayAccess(t *testing.T) {
	t.Run("accesses array element", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20), int64(30)}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAccess,
			Args: []any{opcode.Variable("arr"), int64(1)},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(20) {
			t.Errorf("expected 20, got %v", result)
		}
	})

	t.Run("out of range returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20)}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAccess,
			Args: []any{opcode.Variable("arr"), int64(10)},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for out of range, got %v", result)
		}
	})

	t.Run("negative index returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("arr", NewArrayFromSlice([]any{int64(10), int64(20)}))

		opcode := opcode.OpCode{
			Cmd:  opcode.ArrayAccess,
			Args: []any{opcode.Variable("arr"), int64(-1)},
		}

		result, err := vm.executeArrayAccess(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for negative index, got %v", result)
		}
	})
}

// TestExecuteCall tests the OpCall execution.
func TestExecuteCall(t *testing.T) {
	t.Run("calls built-in function", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.RegisterBuiltinFunction("add", func(vm *VM, args []any) (any, error) {
			a, _ := toInt64(args[0])
			b, _ := toInt64(args[1])
			return a + b, nil
		})

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"add", int64(10), int64(5)},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(15) {
			t.Errorf("expected 15, got %v", result)
		}
	})

	t.Run("unknown function returns error", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"unknownFunc"},
		}

		_, err := vm.executeCall(opcode)
		if err == nil {
			t.Fatal("expected error for unknown function, got nil")
		}
		if !strings.Contains(err.Error(), "undefined function") {
			t.Errorf("expected 'undefined function' error, got %v", err)
		}
	})

	t.Run("calls user-defined function", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Register a user-defined function
		vm.functions["double"] = &FunctionDef{
			Name: "double",
			Parameters: []FunctionParam{
				{Name: "x", Type: "int", IsArray: false},
			},
			Body: []opcode.OpCode{
				{
					Cmd: opcode.Call,
					Args: []any{
						"return",
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"*", opcode.Variable("x"), int64(2)},
						},
					},
				},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"double", int64(21)},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(42) {
			t.Errorf("expected 42, got %v", result)
		}
	})

	t.Run("case-insensitive function lookup", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.RegisterBuiltinFunction("MyFunc", func(vm *VM, args []any) (any, error) {
			return int64(123), nil
		})

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"myfunc"},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(123) {
			t.Errorf("expected 123, got %v", result)
		}
	})
}

// TestEvaluateValue tests the evaluateValue helper function.
func TestEvaluateValue(t *testing.T) {
	t.Run("evaluates literal values", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Integer
		result, err := vm.evaluateValue(int64(42))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(42) {
			t.Errorf("expected 42, got %v", result)
		}

		// String
		result, err = vm.evaluateValue("hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "hello" {
			t.Errorf("expected 'hello', got %v", result)
		}

		// Float
		result, err = vm.evaluateValue(float64(3.14))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != float64(3.14) {
			t.Errorf("expected 3.14, got %v", result)
		}
	})

	t.Run("evaluates variable references", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("x", int64(100))

		result, err := vm.evaluateValue(opcode.Variable("x"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(100) {
			t.Errorf("expected 100, got %v", result)
		}
	})

	t.Run("undefined variable returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		result, err := vm.evaluateValue(opcode.Variable("undefined"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for undefined variable, got %v", result)
		}
	})

	t.Run("evaluates nested OpCode", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		opcode := opcode.OpCode{
			Cmd:  opcode.BinaryOp,
			Args: []any{"+", int64(10), int64(5)},
		}

		result, err := vm.evaluateValue(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(15) {
			t.Errorf("expected 15, got %v", result)
		}
	})

	t.Run("evaluates nil", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		result, err := vm.evaluateValue(nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != nil {
			t.Errorf("expected nil, got %v", result)
		}
	})
}

// TestHelperFunctions tests the helper conversion functions.
func TestHelperFunctions(t *testing.T) {
	t.Run("toInt64", func(t *testing.T) {
		tests := []struct {
			input    any
			expected int64
			ok       bool
		}{
			{int(42), 42, true},
			{int64(42), 42, true},
			{float64(42.9), 42, true},
			{"42", 42, true},
			{"not a number", 0, false},
		}

		for _, tt := range tests {
			result, ok := toInt64(tt.input)
			if ok != tt.ok {
				t.Errorf("toInt64(%v): expected ok=%v, got ok=%v", tt.input, tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("toInt64(%v): expected %d, got %d", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("toFloat64", func(t *testing.T) {
		tests := []struct {
			input    any
			expected float64
			ok       bool
		}{
			{int(42), 42.0, true},
			{int64(42), 42.0, true},
			{float64(3.14), 3.14, true},
			{"3.14", 3.14, true},
			{"not a number", 0, false},
		}

		for _, tt := range tests {
			result, ok := toFloat64(tt.input)
			if ok != tt.ok {
				t.Errorf("toFloat64(%v): expected ok=%v, got ok=%v", tt.input, tt.ok, ok)
			}
			if ok && result != tt.expected {
				t.Errorf("toFloat64(%v): expected %f, got %f", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("toString", func(t *testing.T) {
		tests := []struct {
			input    any
			expected string
		}{
			{"hello", "hello"},
			{int(42), "42"},
			{int64(42), "42"},
			{float64(3.14), "3.14"},
		}

		for _, tt := range tests {
			result := toString(tt.input)
			if result != tt.expected {
				t.Errorf("toString(%v): expected %s, got %s", tt.input, tt.expected, result)
			}
		}
	})

	t.Run("toBool", func(t *testing.T) {
		tests := []struct {
			input    any
			expected bool
		}{
			{true, true},
			{false, false},
			{int(0), false},
			{int(1), true},
			{int64(0), false},
			{int64(1), true},
			{float64(0), false},
			{float64(1), true},
			{"", false},
			{"hello", true},
			{nil, false},
		}

		for _, tt := range tests {
			result := toBool(tt.input)
			if result != tt.expected {
				t.Errorf("toBool(%v): expected %v, got %v", tt.input, tt.expected, result)
			}
		}
	})
}

// TestExecuteIf tests the OpIf execution.
func TestExecuteIf(t *testing.T) {
	t.Run("executes then block when condition is true", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// if (1) { x = 10 }
		opcode := opcode.OpCode{
			Cmd: opcode.If,
			Args: []any{
				int64(1), // condition (true)
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(10)}},
				},
				[]opcode.OpCode{}, // else block (empty)
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, ok := vm.GetCurrentScope().Get("x")
		if !ok {
			t.Error("expected variable 'x' to be set")
		}
		if val != int64(10) {
			t.Errorf("expected x = 10, got %v", val)
		}
	})

	t.Run("executes else block when condition is false", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// if (0) { x = 10 } else { x = 20 }
		opcode := opcode.OpCode{
			Cmd: opcode.If,
			Args: []any{
				int64(0), // condition (false)
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(10)}},
				},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(20)}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("x")
		if val != int64(20) {
			t.Errorf("expected x = 20, got %v", val)
		}
	})

	t.Run("handles nested if-else", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("a", int64(5))
		// if (a > 10) { x = 1 } else if (a > 3) { x = 2 } else { x = 3 }
		opcode := opcode.OpCode{
			Cmd: opcode.If,
			Args: []any{
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("a"), int64(10)}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
				},
				[]opcode.OpCode{
					{
						Cmd: opcode.If,
						Args: []any{
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("a"), int64(3)}},
							[]opcode.OpCode{
								{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(2)}},
							},
							[]opcode.OpCode{
								{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(3)}},
							},
						},
					},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("x")
		if val != int64(2) {
			t.Errorf("expected x = 2, got %v", val)
		}
	})

	t.Run("handles condition with comparison", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("y", int64(15))
		// if (y > 10) { x = 100 }
		opcode := opcode.OpCode{
			Cmd: opcode.If,
			Args: []any{
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("y"), int64(10)}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(100)}},
				},
				[]opcode.OpCode{},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("x")
		if val != int64(100) {
			t.Errorf("expected x = 100, got %v", val)
		}
	})
}

// TestExecuteFor tests the OpFor execution.
func TestExecuteFor(t *testing.T) {
	t.Run("executes simple for loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// sum = 0; for (i = 0; i < 5; i = i + 1) { sum = sum + i }
		vm.GetCurrentScope().Set("sum", int64(0))
		opcode := opcode.OpCode{
			Cmd: opcode.For,
			Args: []any{
				// init: i = 0
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
				},
				// condition: i < 5
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
				// post: i = i + 1
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
				// body: sum = sum + i
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("sum"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("sum"), opcode.Variable("i")}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// sum should be 0 + 1 + 2 + 3 + 4 = 10
		val, _ := vm.GetCurrentScope().Get("sum")
		if val != int64(10) {
			t.Errorf("expected sum = 10, got %v", val)
		}
	})

	t.Run("handles break in for loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("count", int64(0))
		// for (i = 0; i < 10; i = i + 1) { if (i == 3) { break } count = count + 1 }
		opcode := opcode.OpCode{
			Cmd: opcode.For,
			Args: []any{
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
				},
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(10)}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
				[]opcode.OpCode{
					{
						Cmd: opcode.If,
						Args: []any{
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(3)}},
							[]opcode.OpCode{{Cmd: opcode.Break, Args: []any{}}},
							[]opcode.OpCode{},
						},
					},
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("count"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("count"), int64(1)}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// count should be 3 (iterations 0, 1, 2)
		val, _ := vm.GetCurrentScope().Get("count")
		if val != int64(3) {
			t.Errorf("expected count = 3, got %v", val)
		}
	})

	t.Run("handles continue in for loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("sum", int64(0))
		// for (i = 0; i < 5; i = i + 1) { if (i == 2) { continue } sum = sum + i }
		opcode := opcode.OpCode{
			Cmd: opcode.For,
			Args: []any{
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
				},
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
				[]opcode.OpCode{
					{
						Cmd: opcode.If,
						Args: []any{
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(2)}},
							[]opcode.OpCode{{Cmd: opcode.Continue, Args: []any{}}},
							[]opcode.OpCode{},
						},
					},
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("sum"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("sum"), opcode.Variable("i")}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// sum should be 0 + 1 + 3 + 4 = 8 (skipping 2)
		val, _ := vm.GetCurrentScope().Get("sum")
		if val != int64(8) {
			t.Errorf("expected sum = 8, got %v", val)
		}
	})
}

// TestExecuteWhile tests the OpWhile execution.
func TestExecuteWhile(t *testing.T) {
	t.Run("executes simple while loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("i", int64(0))
		vm.GetCurrentScope().Set("sum", int64(0))
		// while (i < 5) { sum = sum + i; i = i + 1 }
		opcode := opcode.OpCode{
			Cmd: opcode.While,
			Args: []any{
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("sum"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("sum"), opcode.Variable("i")}},
					}},
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("sum")
		if val != int64(10) {
			t.Errorf("expected sum = 10, got %v", val)
		}
	})

	t.Run("handles break in while loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("i", int64(0))
		// while (1) { if (i == 5) { break } i = i + 1 }
		opcode := opcode.OpCode{
			Cmd: opcode.While,
			Args: []any{
				int64(1), // infinite loop
				[]opcode.OpCode{
					{
						Cmd: opcode.If,
						Args: []any{
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(5)}},
							[]opcode.OpCode{{Cmd: opcode.Break, Args: []any{}}},
							[]opcode.OpCode{},
						},
					},
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("i")
		if val != int64(5) {
			t.Errorf("expected i = 5, got %v", val)
		}
	})

	t.Run("handles continue in while loop", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("i", int64(0))
		vm.GetCurrentScope().Set("sum", int64(0))
		// while (i < 5) { i = i + 1; if (i == 3) { continue } sum = sum + i }
		opcode := opcode.OpCode{
			Cmd: opcode.While,
			Args: []any{
				opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(5)}},
				[]opcode.OpCode{
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("i"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
					}},
					{
						Cmd: opcode.If,
						Args: []any{
							opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(3)}},
							[]opcode.OpCode{{Cmd: opcode.Continue, Args: []any{}}},
							[]opcode.OpCode{},
						},
					},
					{Cmd: opcode.Assign, Args: []any{
						opcode.Variable("sum"),
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("sum"), opcode.Variable("i")}},
					}},
				},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// sum should be 1 + 2 + 4 + 5 = 12 (skipping 3)
		val, _ := vm.GetCurrentScope().Get("sum")
		if val != int64(12) {
			t.Errorf("expected sum = 12, got %v", val)
		}
	})
}

// TestExecuteSwitch tests the OpSwitch execution.
func TestExecuteSwitch(t *testing.T) {
	t.Run("executes matching case", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("x", int64(2))
		// switch (x) { case 1: y = 10; case 2: y = 20; default: y = 0 }
		opcode := opcode.OpCode{
			Cmd: opcode.Switch,
			Args: []any{
				opcode.Variable("x"),
				[]any{
					map[string]any{
						"value": int64(1),
						"body":  []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}}},
					},
					map[string]any{
						"value": int64(2),
						"body":  []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(20)}}},
					},
				},
				[]opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(0)}}},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("y")
		if val != int64(20) {
			t.Errorf("expected y = 20, got %v", val)
		}
	})

	t.Run("executes default when no case matches", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("x", int64(99))
		// switch (x) { case 1: y = 10; default: y = 0 }
		opcode := opcode.OpCode{
			Cmd: opcode.Switch,
			Args: []any{
				opcode.Variable("x"),
				[]any{
					map[string]any{
						"value": int64(1),
						"body":  []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}}},
					},
				},
				[]opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(0)}}},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("y")
		if val != int64(0) {
			t.Errorf("expected y = 0, got %v", val)
		}
	})

	t.Run("handles string case values", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("cmd", "start")
		// switch (cmd) { case "start": result = 1; case "stop": result = 2 }
		opcode := opcode.OpCode{
			Cmd: opcode.Switch,
			Args: []any{
				opcode.Variable("cmd"),
				[]any{
					map[string]any{
						"value": "start",
						"body":  []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("result"), int64(1)}}},
					},
					map[string]any{
						"value": "stop",
						"body":  []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("result"), int64(2)}}},
					},
				},
				[]opcode.OpCode{},
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		val, _ := vm.GetCurrentScope().Get("result")
		if val != int64(1) {
			t.Errorf("expected result = 1, got %v", val)
		}
	})

	t.Run("handles no matching case and no default", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetCurrentScope().Set("x", int64(99))
		// switch (x) { case 1: y = 10 }
		opcode := opcode.OpCode{
			Cmd: opcode.Switch,
			Args: []any{
				opcode.Variable("x"),
				[]any{
					map[string]any{
						"value": int64(1),
						"body":  []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}}},
					},
				},
				nil, // no default
			},
		}

		_, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// y should not be set
		_, ok := vm.GetCurrentScope().Get("y")
		if ok {
			t.Error("expected y to not be set")
		}
	})
}

// TestExecuteBreak tests the OpBreak execution.
func TestExecuteBreak(t *testing.T) {
	t.Run("returns break signal", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.Break,
			Args: []any{},
		}

		result, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := result.(*breakSignal); !ok {
			t.Errorf("expected breakSignal, got %T", result)
		}
	})
}

// TestExecuteContinue tests the OpContinue execution.
func TestExecuteContinue(t *testing.T) {
	t.Run("returns continue signal", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.Continue,
			Args: []any{},
		}

		result, err := vm.Execute(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if _, ok := result.(*continueSignal); !ok {
			t.Errorf("expected continueSignal, got %T", result)
		}
	})
}

// TestRecursiveFunctionCalls tests recursive function call support.
// Requirement 20.5: System supports recursive function calls.
func TestRecursiveFunctionCalls(t *testing.T) {
	t.Run("factorial function", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Define factorial function:
		// func factorial(n) {
		//     if (n <= 1) { return 1 }
		//     return n * factorial(n - 1)
		// }
		vm.functions["factorial"] = &FunctionDef{
			Name: "factorial",
			Parameters: []FunctionParam{
				{Name: "n", Type: "int", IsArray: false},
			},
			Body: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<=", opcode.Variable("n"), int64(1)}},
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"return", int64(1)}},
						},
						[]opcode.OpCode{},
					},
				},
				{
					Cmd: opcode.Call,
					Args: []any{
						"return",
						opcode.OpCode{
							Cmd: opcode.BinaryOp,
							Args: []any{
								"*",
								opcode.Variable("n"),
								opcode.OpCode{
									Cmd: opcode.Call,
									Args: []any{
										"factorial",
										opcode.OpCode{
											Cmd:  opcode.BinaryOp,
											Args: []any{"-", opcode.Variable("n"), int64(1)},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Call factorial(5) = 120
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"factorial", int64(5)},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(120) {
			t.Errorf("expected factorial(5) = 120, got %v", result)
		}
	})

	t.Run("fibonacci function", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Define fibonacci function:
		// func fib(n) {
		//     if (n <= 1) { return n }
		//     return fib(n - 1) + fib(n - 2)
		// }
		vm.functions["fib"] = &FunctionDef{
			Name: "fib",
			Parameters: []FunctionParam{
				{Name: "n", Type: "int", IsArray: false},
			},
			Body: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<=", opcode.Variable("n"), int64(1)}},
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"return", opcode.Variable("n")}},
						},
						[]opcode.OpCode{},
					},
				},
				{
					Cmd: opcode.Call,
					Args: []any{
						"return",
						opcode.OpCode{
							Cmd: opcode.BinaryOp,
							Args: []any{
								"+",
								opcode.OpCode{
									Cmd: opcode.Call,
									Args: []any{
										"fib",
										opcode.OpCode{
											Cmd:  opcode.BinaryOp,
											Args: []any{"-", opcode.Variable("n"), int64(1)},
										},
									},
								},
								opcode.OpCode{
									Cmd: opcode.Call,
									Args: []any{
										"fib",
										opcode.OpCode{
											Cmd:  opcode.BinaryOp,
											Args: []any{"-", opcode.Variable("n"), int64(2)},
										},
									},
								},
							},
						},
					},
				},
			},
		}

		// Call fib(10) = 55
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"fib", int64(10)},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(55) {
			t.Errorf("expected fib(10) = 55, got %v", result)
		}
	})

	t.Run("stack depth tracking during recursion", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		maxDepthReached := 0

		// Define a function that tracks max depth
		vm.functions["trackDepth"] = &FunctionDef{
			Name: "trackDepth",
			Parameters: []FunctionParam{
				{Name: "n", Type: "int", IsArray: false},
			},
			Body: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<=", opcode.Variable("n"), int64(0)}},
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"return", int64(0)}},
						},
						[]opcode.OpCode{},
					},
				},
				{
					Cmd: opcode.Call,
					Args: []any{
						"return",
						opcode.OpCode{
							Cmd: opcode.Call,
							Args: []any{
								"trackDepth",
								opcode.OpCode{
									Cmd:  opcode.BinaryOp,
									Args: []any{"-", opcode.Variable("n"), int64(1)},
								},
							},
						},
					},
				},
			},
		}

		// Register a builtin to track depth
		vm.RegisterBuiltinFunction("getDepth", func(vm *VM, args []any) (any, error) {
			depth := vm.GetStackDepth()
			if depth > maxDepthReached {
				maxDepthReached = depth
			}
			return int64(depth), nil
		})

		// Call trackDepth(10)
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"trackDepth", int64(10)},
		}

		_, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// After recursion completes, stack should be empty
		if vm.GetStackDepth() != 0 {
			t.Errorf("expected stack depth 0 after recursion, got %d", vm.GetStackDepth())
		}
	})
}

// TestFunctionReturnValue tests return value handling.
// Requirement 20.3: When function has return value, system passes it to caller.
// Requirement 20.4: When function has no return value, system returns zero.
func TestFunctionReturnValue(t *testing.T) {
	t.Run("function with explicit return value", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.functions["getValue"] = &FunctionDef{
			Name:       "getValue",
			Parameters: []FunctionParam{},
			Body: []opcode.OpCode{
				{Cmd: opcode.Call, Args: []any{"return", int64(42)}},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"getValue"},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(42) {
			t.Errorf("expected 42, got %v", result)
		}
	})

	t.Run("function without return statement returns zero", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.functions["noReturn"] = &FunctionDef{
			Name:       "noReturn",
			Parameters: []FunctionParam{},
			Body: []opcode.OpCode{
				// Just assign a variable, no return
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(100)}},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"noReturn"},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(0) {
			t.Errorf("expected 0 for function without return, got %v", result)
		}
	})

	t.Run("function with return in middle of body", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.functions["earlyReturn"] = &FunctionDef{
			Name:       "earlyReturn",
			Parameters: []FunctionParam{},
			Body: []opcode.OpCode{
				{Cmd: opcode.Call, Args: []any{"return", int64(10)}},
				// This should not be executed
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(999)}},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"earlyReturn"},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(10) {
			t.Errorf("expected 10, got %v", result)
		}

		// Verify x was not set (return should have stopped execution)
		_, ok := vm.GetGlobalScope().Get("x")
		if ok {
			t.Error("expected x to not be set after early return")
		}
	})
}

// TestStackOverflowDetection tests stack overflow detection.
// Requirement 20.6: System detects stack overflow and reports error.
// Requirement 20.7: System maintains maximum stack depth of 1000 frames.
// Requirement 20.8: When stack overflow occurs, system logs error and terminates execution.
func TestStackOverflowDetection(t *testing.T) {
	t.Run("detects stack overflow in recursive function", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Define an infinitely recursive function
		vm.functions["infinite"] = &FunctionDef{
			Name:       "infinite",
			Parameters: []FunctionParam{},
			Body: []opcode.OpCode{
				{Cmd: opcode.Call, Args: []any{"infinite"}},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"infinite"},
		}

		_, err := vm.executeCall(opcode)
		if err == nil {
			t.Error("expected stack overflow error")
		}
		if err != nil && err.Error() != "stack overflow: maximum depth 1000 exceeded" {
			t.Errorf("unexpected error message: %v", err)
		}
	})
}

// TestFunctionParameterBinding tests parameter binding to local scope.
// Requirement 9.7: When function parameters are passed, system binds them to local scope.
func TestFunctionParameterBinding(t *testing.T) {
	t.Run("binds parameters to local scope", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.functions["add"] = &FunctionDef{
			Name: "add",
			Parameters: []FunctionParam{
				{Name: "a", Type: "int", IsArray: false},
				{Name: "b", Type: "int", IsArray: false},
			},
			Body: []opcode.OpCode{
				{
					Cmd: opcode.Call,
					Args: []any{
						"return",
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", opcode.Variable("a"), opcode.Variable("b")},
						},
					},
				},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"add", int64(10), int64(20)},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(30) {
			t.Errorf("expected 30, got %v", result)
		}
	})

	t.Run("uses default parameter values", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.functions["greet"] = &FunctionDef{
			Name: "greet",
			Parameters: []FunctionParam{
				{Name: "name", Type: "string", IsArray: false, Default: "World", HasDefault: true},
			},
			Body: []opcode.OpCode{
				{Cmd: opcode.Call, Args: []any{"return", opcode.Variable("name")}},
			},
		}

		// Call without argument - should use default
		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"greet"},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "World" {
			t.Errorf("expected 'World', got %v", result)
		}
	})

	t.Run("parameter does not affect global variable", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		vm.GetGlobalScope().Set("x", int64(100))

		vm.functions["modifyX"] = &FunctionDef{
			Name: "modifyX",
			Parameters: []FunctionParam{
				{Name: "x", Type: "int", IsArray: false},
			},
			Body: []opcode.OpCode{
				// Modify local x
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(999)}},
				{Cmd: opcode.Call, Args: []any{"return", opcode.Variable("x")}},
			},
		}

		opcode := opcode.OpCode{
			Cmd:  opcode.Call,
			Args: []any{"modifyX", int64(50)},
		}

		result, err := vm.executeCall(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != int64(999) {
			t.Errorf("expected 999, got %v", result)
		}

		// Global x should be unchanged
		globalX, _ := vm.GetGlobalScope().Get("x")
		if globalX != int64(100) {
			t.Errorf("expected global x = 100, got %v", globalX)
		}
	})
}

// TestExecuteSetStep tests the OpSetStep execution.
// Requirement 6.1: When OpSetStep OpCode is executed, system initializes step counter with specified count.
func TestExecuteSetStep(t *testing.T) {
	t.Run("initializes VM step counter with literal int64", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{int64(10)},
		}

		_, err := vm.executeSetStep(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify step counter is set in VM (no handler executing)
		if vm.GetStepCounter() != 10 {
			t.Errorf("expected step counter 10, got %d", vm.GetStepCounter())
		}
	})

	t.Run("initializes VM step counter with variable", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Set variable n = 20
		vm.GetCurrentScope().Set("n", int64(20))

		opcode := opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{opcode.Variable("n")},
		}

		_, err := vm.executeSetStep(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify step counter is set in VM
		if vm.GetStepCounter() != 20 {
			t.Errorf("expected step counter 20, got %d", vm.GetStepCounter())
		}
	})

	t.Run("initializes VM step counter with expression", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		// Set variable x = 5
		vm.GetCurrentScope().Set("x", int64(5))

		// step(x + 3) should result in step counter = 8
		opcode := opcode.OpCode{
			Cmd: opcode.SetStep,
			Args: []any{opcode.OpCode{
				Cmd:  opcode.BinaryOp,
				Args: []any{"+", opcode.Variable("x"), int64(3)},
			}},
		}

		_, err := vm.executeSetStep(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify step counter is set in VM
		if vm.GetStepCounter() != 8 {
			t.Errorf("expected step counter 8, got %d", vm.GetStepCounter())
		}
	})

	t.Run("initializes handler step counter when handler is executing", func(t *testing.T) {
		vm := New([]opcode.OpCode{})

		// Create and set a current handler
		handler := NewEventHandler("test_handler", EventTIME, []opcode.OpCode{}, vm, nil)
		vm.SetCurrentHandler(handler)

		opcode := opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{int64(15)},
		}

		_, err := vm.executeSetStep(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify step counter is set in handler, not VM
		if handler.StepCounter != 15 {
			t.Errorf("expected handler step counter 15, got %d", handler.StepCounter)
		}
		// VM step counter should remain 0
		if vm.GetStepCounter() != 0 {
			t.Errorf("expected VM step counter 0, got %d", vm.GetStepCounter())
		}
	})

	t.Run("handles zero step count", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{int64(0)},
		}

		_, err := vm.executeSetStep(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if vm.GetStepCounter() != 0 {
			t.Errorf("expected step counter 0, got %d", vm.GetStepCounter())
		}
	})

	t.Run("handles float step count by truncating", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{float64(7.9)},
		}

		_, err := vm.executeSetStep(opcode)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Float should be truncated to int
		if vm.GetStepCounter() != 7 {
			t.Errorf("expected step counter 7, got %d", vm.GetStepCounter())
		}
	})

	t.Run("returns error when no arguments provided", func(t *testing.T) {
		vm := New([]opcode.OpCode{})
		opcode := opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{},
		}

		_, err := vm.executeSetStep(opcode)
		if err == nil {
			t.Error("expected error for missing arguments")
		}
	})
}
