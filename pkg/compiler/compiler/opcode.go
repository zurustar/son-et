// Package compiler provides OpCode generation for FILLY scripts (.TFY files).
package compiler

// OpCmd represents an OpCode command type.
// Each OpCmd corresponds to a specific operation that the VM can execute.
type OpCmd string

// OpCode command types for all supported operations.
// These constants define the instruction set for the FILLY virtual machine.
const (
	// OpAssign assigns a value to a variable.
	// Args: [Variable(name), value]
	OpAssign OpCmd = "Assign"

	// OpArrayAssign assigns a value to an array element.
	// Args: [Variable(arrayName), index, value]
	OpArrayAssign OpCmd = "ArrayAssign"

	// OpCall invokes a function with arguments.
	// Args: [functionName, arg1, arg2, ...]
	OpCall OpCmd = "Call"

	// OpBinaryOp performs a binary operation (+, -, *, /, %, ==, !=, <, <=, >, >=, &&, ||).
	// Args: [operator, leftOperand, rightOperand]
	OpBinaryOp OpCmd = "BinaryOp"

	// OpUnaryOp performs a unary operation (-, !).
	// Args: [operator, operand]
	OpUnaryOp OpCmd = "UnaryOp"

	// OpArrayAccess accesses an array element by index.
	// Args: [Variable(arrayName), index]
	OpArrayAccess OpCmd = "ArrayAccess"

	// OpIf executes conditional branching.
	// Args: [condition, thenBlock []OpCode, elseBlock []OpCode]
	OpIf OpCmd = "If"

	// OpFor executes a for loop.
	// Args: [initBlock []OpCode, condition, postBlock []OpCode, bodyBlock []OpCode]
	OpFor OpCmd = "For"

	// OpWhile executes a while loop.
	// Args: [condition, bodyBlock []OpCode]
	OpWhile OpCmd = "While"

	// OpSwitch executes a switch statement.
	// Args: [value, cases []CaseClause, defaultBlock []OpCode]
	OpSwitch OpCmd = "Switch"

	// OpBreak breaks out of the current loop.
	// Args: []
	OpBreak OpCmd = "Break"

	// OpContinue continues to the next iteration of the current loop.
	// Args: []
	OpContinue OpCmd = "Continue"

	// OpRegisterEventHandler registers an event handler for mes() blocks.
	// Args: [eventType string, bodyBlock []OpCode]
	OpRegisterEventHandler OpCmd = "RegisterEventHandler"

	// OpWait waits for a specified number of steps in step() blocks.
	// Args: [stepCount int]
	OpWait OpCmd = "Wait"

	// OpSetStep sets the step duration for subsequent commands in step() blocks.
	// Args: [stepDuration int]
	OpSetStep OpCmd = "SetStep"

	// OpDefineFunction defines a user-defined function.
	// Args: [functionName string, parameters []map[string]any, bodyBlock []OpCode]
	// Each parameter map contains: name, type, isArray, and optionally default
	OpDefineFunction OpCmd = "DefineFunction"
)

// OpCode represents a single OpCode instruction.
// It consists of a command type (Cmd) and a slice of arguments (Args).
// The Args can contain various types including:
// - Primitive values (int, string, float)
// - Variable references (Variable type)
// - Nested OpCode structures for complex expressions
// - Slices of OpCode for block statements
type OpCode struct {
	Cmd  OpCmd
	Args []any
}

// Variable represents a variable reference in OpCode arguments.
// This type distinguishes variable references from literal string values.
// When the VM encounters a Variable in Args, it should resolve the variable
// by name from the current scope.
type Variable string
