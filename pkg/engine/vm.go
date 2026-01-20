package engine

import (
	"fmt"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// VM executes OpCode sequences within the engine.
type VM struct {
	state  *EngineState
	engine *Engine
	logger *Logger
}

// NewVM creates a new VM with the given state and logger.
func NewVM(state *EngineState, engine *Engine, logger *Logger) *VM {
	return &VM{
		state:  state,
		engine: engine,
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

	case interpreter.OpRegisterEventHandler:
		// mes() blocks are registered as event handlers
		// This is a stub for now - full implementation in Phase 3.3
		vm.logger.LogDebug("RegisterEventHandler (not yet implemented)")
		return nil

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

// executeCall handles function calls
func (vm *VM) executeCall(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) == 0 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpCall requires at least 1 argument (function name)")
	}

	// Get function name
	var funcName string
	switch fn := op.Args[0].(type) {
	case string:
		funcName = fn
	case interpreter.Variable:
		funcName = string(fn)
	default:
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpCall first argument must be string or Variable, got %T", op.Args[0])
	}

	// Handle special built-in functions
	switch funcName {
	case "define_function":
		return vm.executeDefineFunction(seq, op)
	case "return":
		// TODO: Implement return statement handling
		vm.logger.LogDebug("Return statement (not yet implemented)")
		return nil
	default:
		// Try to call user-defined function
		return vm.executeUserFunction(seq, funcName, op.Args[1:])
	}
}

// executeDefineFunction handles function definition registration
func (vm *VM) executeDefineFunction(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 4 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function requires 4 arguments (name, params, body), got %d", len(op.Args))
	}

	// Get function name
	funcName, ok := op.Args[1].(string)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function name must be string, got %T", op.Args[1])
	}

	// Get parameters
	params, ok := op.Args[2].([]any)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function params must be []any, got %T", op.Args[2])
	}

	// Convert parameters to strings
	paramNames := make([]string, len(params))
	for i, p := range params {
		paramNames[i] = fmt.Sprintf("%v", p)
	}

	// Get function body
	body, ok := op.Args[3].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function body must be []OpCode, got %T", op.Args[3])
	}

	// Register function in engine state
	vm.state.RegisterFunction(funcName, paramNames, body)
	vm.logger.LogDebug("Defined function: %s with %d parameters", funcName, len(paramNames))

	return nil
}

// executeUserFunction handles user-defined function calls
func (vm *VM) executeUserFunction(seq *Sequencer, funcName string, args []any) error {
	// Look up function definition
	funcDef, ok := vm.state.GetFunction(funcName)
	if !ok {
		// Not a user-defined function - try built-in functions
		return vm.executeBuiltinFunction(seq, funcName, args)
	}

	// Evaluate arguments
	evaluatedArgs := make([]any, len(args))
	for i, arg := range args {
		val, err := vm.evaluateValue(seq, arg)
		if err != nil {
			return err
		}
		evaluatedArgs[i] = val
	}

	// Create new sequencer for function execution with current sequencer as parent
	funcSeq := NewSequencer(funcDef.Body, seq.GetMode(), seq)

	// Bind parameters to arguments
	for i, paramName := range funcDef.Parameters {
		if i < len(evaluatedArgs) {
			funcSeq.SetVariable(paramName, evaluatedArgs[i])
		} else {
			// Parameter not provided, use default value (0)
			funcSeq.SetVariable(paramName, 0)
		}
	}

	// Execute function body synchronously
	vm.logger.LogDebug("Calling user function: %s with %d arguments", funcName, len(evaluatedArgs))
	return vm.executeBlock(funcSeq, funcDef.Body)
}

// executeBuiltinFunction handles built-in function calls
func (vm *VM) executeBuiltinFunction(seq *Sequencer, funcName string, args []any) error {
	// Evaluate arguments
	evaluatedArgs := make([]any, len(args))
	for i, arg := range args {
		vm.logger.LogDebug("  arg[%d] BEFORE eval: %v (type: %T)", i, arg, arg)
		val, err := vm.evaluateValue(seq, arg)
		if err != nil {
			return err
		}
		evaluatedArgs[i] = val
		vm.logger.LogDebug("  arg[%d] AFTER eval: %v (type: %T)", i, val, val)
	}

	vm.logger.LogDebug("Call: %s (built-in function)", funcName)

	// Handle built-in functions
	switch strings.ToLower(funcName) {
	case "loadpic":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("LoadPic", fmt.Sprintf("%v", evaluatedArgs), "LoadPic requires 1 argument (filename)")
		}
		filename := fmt.Sprintf("%v", evaluatedArgs[0])
		picID := vm.engine.LoadPic(filename)
		// Store result in a special return variable (for now, just ignore)
		_ = picID
		return nil

	case "createpic":
		if len(evaluatedArgs) < 2 {
			return NewRuntimeError("CreatePic", fmt.Sprintf("%v", evaluatedArgs), "CreatePic requires 2 arguments (width, height)")
		}
		width := int(vm.toInt(evaluatedArgs[0]))
		height := int(vm.toInt(evaluatedArgs[1]))
		picID := vm.engine.CreatePic(width, height)
		_ = picID
		return nil

	case "delpic":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("DelPic", fmt.Sprintf("%v", evaluatedArgs), "DelPic requires 1 argument (picID)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.DelPic(picID)
		return nil

	case "picwidth":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PicWidth", fmt.Sprintf("%v", evaluatedArgs), "PicWidth requires 1 argument (picID)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		width := vm.engine.PicWidth(picID)
		_ = width
		return nil

	case "picheight":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PicHeight", fmt.Sprintf("%v", evaluatedArgs), "PicHeight requires 1 argument (picID)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		height := vm.engine.PicHeight(picID)
		_ = height
		return nil

	case "movepic":
		if len(evaluatedArgs) < 8 {
			return NewRuntimeError("MovePic", fmt.Sprintf("%v", evaluatedArgs), "MovePic requires 8 arguments (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)")
		}
		srcID := int(vm.toInt(evaluatedArgs[0]))
		srcX := int(vm.toInt(evaluatedArgs[1]))
		srcY := int(vm.toInt(evaluatedArgs[2]))
		srcW := int(vm.toInt(evaluatedArgs[3]))
		srcH := int(vm.toInt(evaluatedArgs[4]))
		dstID := int(vm.toInt(evaluatedArgs[5]))
		dstX := int(vm.toInt(evaluatedArgs[6]))
		dstY := int(vm.toInt(evaluatedArgs[7]))
		vm.engine.MovePic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)
		return nil

	case "movespic":
		if len(evaluatedArgs) < 10 {
			return NewRuntimeError("MoveSPic", fmt.Sprintf("%v", evaluatedArgs), "MoveSPic requires 10 arguments (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)")
		}
		srcID := int(vm.toInt(evaluatedArgs[0]))
		srcX := int(vm.toInt(evaluatedArgs[1]))
		srcY := int(vm.toInt(evaluatedArgs[2]))
		srcW := int(vm.toInt(evaluatedArgs[3]))
		srcH := int(vm.toInt(evaluatedArgs[4]))
		dstID := int(vm.toInt(evaluatedArgs[5]))
		dstX := int(vm.toInt(evaluatedArgs[6]))
		dstY := int(vm.toInt(evaluatedArgs[7]))
		dstW := int(vm.toInt(evaluatedArgs[8]))
		dstH := int(vm.toInt(evaluatedArgs[9]))
		vm.engine.MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)
		return nil

	case "reversepic":
		if len(evaluatedArgs) < 8 {
			return NewRuntimeError("ReversePic", fmt.Sprintf("%v", evaluatedArgs), "ReversePic requires 8 arguments (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)")
		}
		srcID := int(vm.toInt(evaluatedArgs[0]))
		srcX := int(vm.toInt(evaluatedArgs[1]))
		srcY := int(vm.toInt(evaluatedArgs[2]))
		srcW := int(vm.toInt(evaluatedArgs[3]))
		srcH := int(vm.toInt(evaluatedArgs[4]))
		dstID := int(vm.toInt(evaluatedArgs[5]))
		dstX := int(vm.toInt(evaluatedArgs[6]))
		dstY := int(vm.toInt(evaluatedArgs[7]))
		vm.engine.ReversePic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)
		return nil

	case "openwin":
		// OpenWin can be called with 1, 5, or 8 arguments
		// OpenWin(pic) - 1 arg: x=0, y=0, w=0, h=0, picX=0, picY=0, col=0
		// OpenWin(pic, x, y, w, h) - 5 args: picX=0, picY=0, col=0
		// OpenWin(pic, x, y, w, h, picX, picY, col) - 8 args (full)
		// Pad missing arguments with 0
		for len(evaluatedArgs) < 8 {
			evaluatedArgs = append(evaluatedArgs, int64(0))
		}

		picID := int(vm.toInt(evaluatedArgs[0]))
		x := int(vm.toInt(evaluatedArgs[1]))
		y := int(vm.toInt(evaluatedArgs[2]))
		width := int(vm.toInt(evaluatedArgs[3]))
		height := int(vm.toInt(evaluatedArgs[4]))
		picX := int(vm.toInt(evaluatedArgs[5]))
		picY := int(vm.toInt(evaluatedArgs[6]))
		color := int(vm.toInt(evaluatedArgs[7]))
		winID := vm.engine.OpenWin(picID, x, y, width, height, picX, picY, color)
		_ = winID
		return nil

	case "movewin":
		if len(evaluatedArgs) < 7 {
			return NewRuntimeError("MoveWin", fmt.Sprintf("%v", evaluatedArgs), "MoveWin requires 7 arguments (winID, x, y, width, height, picX, picY)")
		}
		winID := int(vm.toInt(evaluatedArgs[0]))
		x := int(vm.toInt(evaluatedArgs[1]))
		y := int(vm.toInt(evaluatedArgs[2]))
		width := int(vm.toInt(evaluatedArgs[3]))
		height := int(vm.toInt(evaluatedArgs[4]))
		picX := int(vm.toInt(evaluatedArgs[5]))
		picY := int(vm.toInt(evaluatedArgs[6]))
		vm.engine.MoveWin(winID, x, y, width, height, picX, picY)
		return nil

	case "closewin":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("CloseWin", fmt.Sprintf("%v", evaluatedArgs), "CloseWin requires 1 argument (winID)")
		}
		winID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.CloseWin(winID)
		return nil

	case "closewinall":
		vm.engine.CloseWinAll()
		return nil

	case "captitle":
		if len(evaluatedArgs) < 2 {
			return NewRuntimeError("CapTitle", fmt.Sprintf("%v", evaluatedArgs), "CapTitle requires 2 arguments (winID, caption)")
		}
		winID := int(vm.toInt(evaluatedArgs[0]))
		caption := fmt.Sprintf("%v", evaluatedArgs[1])
		vm.engine.CapTitle(winID, caption)
		return nil

	case "getpicno":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("GetPicNo", fmt.Sprintf("%v", evaluatedArgs), "GetPicNo requires 1 argument (winID)")
		}
		winID := int(vm.toInt(evaluatedArgs[0]))
		picID := vm.engine.GetPicNo(winID)
		_ = picID
		return nil

	default:
		// Unknown built-in function - just log and ignore
		vm.logger.LogDebug("Unknown built-in function: %s", funcName)
		return nil
	}
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
