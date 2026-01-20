package engine

import (
	"fmt"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// VM executes OpCode sequences within the engine.
type VM struct {
	state  *EngineState
	logger *Logger
}

// NewVM creates a new VM with the given state and logger.
func NewVM(state *EngineState, logger *Logger) *VM {
	return &VM{
		state:  state,
		logger: logger,
	}
}

// ExecuteOp executes a single OpCode operation.
// This is the central dispatch point for all OpCode execution.
// Returns an error if the operation fails.
func (vm *VM) ExecuteOp(seq *Sequencer, op interpreter.OpCode) error {
	vm.logger.LogDebug("ExecuteOp: %s (args: %d)", op.Cmd.String(), len(op.Args))

	switch op.Cmd {
	case interpreter.OpAssign:
		return vm.executeAssign(seq, op)

	case interpreter.OpCall:
		return vm.executeCall(seq, op)

	case interpreter.OpIf:
		return vm.executeIf(seq, op)

	case interpreter.OpFor:
		return vm.executeFor(seq, op)

	case interpreter.OpWhile:
		return vm.executeWhile(seq, op)

	case interpreter.OpWait:
		return vm.executeWait(seq, op)

	case interpreter.OpBinaryOp:
		// Binary operations are evaluated as part of expressions
		// They should not appear at the statement level
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpBinaryOp cannot be executed as statement")

	case interpreter.OpRegisterSequence:
		// Sequence registration is handled by the engine, not the VM
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpRegisterSequence should be handled by engine")

	default:
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "Unknown OpCode: %s", op.Cmd.String())
	}
}

// executeAssign handles variable assignment: x = value
func (vm *VM) executeAssign(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpAssign requires 2 arguments, got %d", len(op.Args))
	}

	// First argument must be a Variable
	varName, ok := op.Args[0].(interpreter.Variable)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpAssign first argument must be Variable, got %T", op.Args[0])
	}

	// Second argument is the value (can be literal or expression)
	value, err := vm.evaluateValue(seq, op.Args[1])
	if err != nil {
		return err
	}

	// Set the variable
	seq.SetVariable(string(varName), value)
	vm.logger.LogDebug("Assign: %s = %v", varName, value)

	return nil
}

// executeCall handles function calls (stub for now)
// seq parameter will be used when full function call implementation is added
func (vm *VM) executeCall(seq *Sequencer, op interpreter.OpCode) error {
	_ = seq // Will be used in full implementation

	if len(op.Args) == 0 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpCall requires at least 1 argument (function name)")
	}

	// For now, just log the call
	// Full implementation will come in later phases
	vm.logger.LogDebug("Call: %v (stub)", op.Args[0])

	return nil
}

// executeIf handles if statements: if (condition) { thenBlock } else { elseBlock }
func (vm *VM) executeIf(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) < 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpIf requires at least 2 arguments, got %d", len(op.Args))
	}

	// Evaluate condition
	condition, err := vm.evaluateValue(seq, op.Args[0])
	if err != nil {
		return err
	}

	// Convert condition to boolean
	condBool := vm.toBool(condition)
	vm.logger.LogDebug("If: condition = %v", condBool)

	// Get then block
	thenBlock, ok := op.Args[1].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpIf second argument must be []OpCode, got %T", op.Args[1])
	}

	// Get else block (optional)
	var elseBlock []interpreter.OpCode
	if len(op.Args) >= 3 {
		elseBlock, ok = op.Args[2].([]interpreter.OpCode)
		if !ok {
			return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpIf third argument must be []OpCode, got %T", op.Args[2])
		}
	}

	// Execute appropriate block
	if condBool {
		return vm.executeBlock(seq, thenBlock)
	} else if len(elseBlock) > 0 {
		return vm.executeBlock(seq, elseBlock)
	}

	return nil
}

// executeFor handles for loops: for (init; condition; increment) { body }
func (vm *VM) executeFor(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 4 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpFor requires 4 arguments, got %d", len(op.Args))
	}

	// Execute init
	if op.Args[0] != nil {
		if initOp, ok := op.Args[0].(interpreter.OpCode); ok {
			if err := vm.ExecuteOp(seq, initOp); err != nil {
				return err
			}
		}
	}

	// Get condition, increment, and body
	condition := op.Args[1]
	increment := op.Args[2]
	body, ok := op.Args[3].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpFor fourth argument must be []OpCode, got %T", op.Args[3])
	}

	// Execute loop
	for {
		// Evaluate condition
		if condition != nil {
			condValue, err := vm.evaluateValue(seq, condition)
			if err != nil {
				return err
			}
			if !vm.toBool(condValue) {
				break
			}
		}

		// Execute body
		if err := vm.executeBlock(seq, body); err != nil {
			return err
		}

		// Execute increment
		if increment != nil {
			if incOp, ok := increment.(interpreter.OpCode); ok {
				if err := vm.ExecuteOp(seq, incOp); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// executeWhile handles while loops: while (condition) { body }
func (vm *VM) executeWhile(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpWhile requires 2 arguments, got %d", len(op.Args))
	}

	condition := op.Args[0]
	body, ok := op.Args[1].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpWhile second argument must be []OpCode, got %T", op.Args[1])
	}

	// Execute loop
	for {
		// Evaluate condition
		condValue, err := vm.evaluateValue(seq, condition)
		if err != nil {
			return err
		}

		if !vm.toBool(condValue) {
			break
		}

		// Execute body
		if err := vm.executeBlock(seq, body); err != nil {
			return err
		}
	}

	return nil
}

// executeWait handles wait operations: wait(n)
// The wait duration depends on the sequence's timing mode:
// - TIME mode: n steps × 3 ticks/step = 3n ticks (50ms per step at 60 FPS)
// - MIDI_TIME mode: n steps × 1 tick/step = n ticks (32nd note per step)
func (vm *VM) executeWait(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 1 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpWait requires 1 argument, got %d", len(op.Args))
	}

	// Evaluate wait count (number of steps)
	waitValue, err := vm.evaluateValue(seq, op.Args[0])
	if err != nil {
		return err
	}

	// Convert to int (number of steps)
	steps := vm.toInt(waitValue)

	// Calculate actual tick count based on stepSize
	ticks := int(steps) * seq.GetStepSize()

	vm.logger.LogDebug("Wait: %d steps × %d ticks/step = %d ticks", steps, seq.GetStepSize(), ticks)

	// Set wait counter
	seq.SetWait(ticks)

	return nil
}

// executeBlock executes a block of OpCodes sequentially
func (vm *VM) executeBlock(seq *Sequencer, block []interpreter.OpCode) error {
	for _, op := range block {
		if err := vm.ExecuteOp(seq, op); err != nil {
			return err
		}
	}
	return nil
}

// evaluateValue evaluates a value (literal, variable, or expression)
func (vm *VM) evaluateValue(seq *Sequencer, value any) (any, error) {
	switch v := value.(type) {
	case int, int64, float64, string, bool:
		// Literal value
		return v, nil

	case interpreter.Variable:
		// Variable reference
		return seq.GetVariable(string(v)), nil

	case interpreter.OpCode:
		// Nested expression
		return vm.evaluateExpression(seq, v)

	default:
		return nil, fmt.Errorf("cannot evaluate value of type %T", value)
	}
}

// evaluateExpression evaluates an expression OpCode
func (vm *VM) evaluateExpression(seq *Sequencer, op interpreter.OpCode) (any, error) {
	switch op.Cmd {
	case interpreter.OpBinaryOp:
		return vm.evaluateBinaryOp(seq, op)

	case interpreter.OpCall:
		// Function call expression (stub for now)
		return 0, nil

	default:
		return nil, fmt.Errorf("cannot evaluate expression: %s", op.Cmd.String())
	}
}

// evaluateBinaryOp evaluates a binary operation
func (vm *VM) evaluateBinaryOp(seq *Sequencer, op interpreter.OpCode) (any, error) {
	if len(op.Args) != 3 {
		return nil, fmt.Errorf("OpBinaryOp requires 3 arguments, got %d", len(op.Args))
	}

	// Get operator
	operator, ok := op.Args[0].(string)
	if !ok {
		return nil, fmt.Errorf("OpBinaryOp first argument must be string, got %T", op.Args[0])
	}

	// Evaluate left and right operands
	left, err := vm.evaluateValue(seq, op.Args[1])
	if err != nil {
		return nil, err
	}

	right, err := vm.evaluateValue(seq, op.Args[2])
	if err != nil {
		return nil, err
	}

	// Perform operation
	return vm.applyBinaryOp(operator, left, right)
}

// applyBinaryOp applies a binary operator to two values
func (vm *VM) applyBinaryOp(op string, left, right any) (any, error) {
	// Convert to int64 for arithmetic
	leftInt := vm.toInt(left)
	rightInt := vm.toInt(right)

	switch op {
	case "+":
		return leftInt + rightInt, nil
	case "-":
		return leftInt - rightInt, nil
	case "*":
		return leftInt * rightInt, nil
	case "/":
		if rightInt == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return leftInt / rightInt, nil
	case "%":
		if rightInt == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return leftInt % rightInt, nil
	case "==":
		return leftInt == rightInt, nil
	case "!=":
		return leftInt != rightInt, nil
	case "<":
		return leftInt < rightInt, nil
	case ">":
		return leftInt > rightInt, nil
	case "<=":
		return leftInt <= rightInt, nil
	case ">=":
		return leftInt >= rightInt, nil
	case "&&":
		return vm.toBool(left) && vm.toBool(right), nil
	case "||":
		return vm.toBool(left) || vm.toBool(right), nil
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", op)
	}
}

// toBool converts a value to boolean
func (vm *VM) toBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	default:
		return false
	}
}

// toInt converts a value to int64
func (vm *VM) toInt(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		// Try to parse string as int
		// For now, just return 0
		return 0
	default:
		return 0
	}
}
