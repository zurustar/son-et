// Package opcode defines the instruction set for the FILLY virtual machine.
// This package is the foundation that both the compiler and VM depend on.
// The compiler generates OpCode sequences, and the VM executes them.
package opcode

// Cmd represents an OpCode command type.
// Each Cmd corresponds to a specific operation that the VM can execute.
type Cmd string

// OpCode command types for all supported operations.
// These constants define the instruction set for the FILLY virtual machine.
const (
	// Assign assigns a value to a variable.
	// Args: [Variable(name), value]
	Assign Cmd = "Assign"

	// ArrayAssign assigns a value to an array element.
	// Args: [Variable(arrayName), index, value]
	ArrayAssign Cmd = "ArrayAssign"

	// Call invokes a function with arguments.
	// Args: [functionName, arg1, arg2, ...]
	Call Cmd = "Call"

	// BinaryOp performs a binary operation (+, -, *, /, %, ==, !=, <, <=, >, >=, &&, ||).
	// Args: [operator, leftOperand, rightOperand]
	BinaryOp Cmd = "BinaryOp"

	// UnaryOp performs a unary operation (-, !).
	// Args: [operator, operand]
	UnaryOp Cmd = "UnaryOp"

	// ArrayAccess accesses an array element by index.
	// Args: [Variable(arrayName), index]
	ArrayAccess Cmd = "ArrayAccess"

	// If executes conditional branching.
	// Args: [condition, thenBlock []OpCode, elseBlock []OpCode]
	If Cmd = "If"

	// For executes a for loop.
	// Args: [initBlock []OpCode, condition, postBlock []OpCode, bodyBlock []OpCode]
	For Cmd = "For"

	// While executes a while loop.
	// Args: [condition, bodyBlock []OpCode]
	While Cmd = "While"

	// Switch executes a switch statement.
	// Args: [value, cases []CaseClause, defaultBlock []OpCode]
	Switch Cmd = "Switch"

	// Break breaks out of the current loop.
	// Args: []
	Break Cmd = "Break"

	// Continue continues to the next iteration of the current loop.
	// Args: []
	Continue Cmd = "Continue"

	// RegisterEventHandler registers an event handler for mes() blocks.
	// Args: [eventType string, bodyBlock []OpCode]
	RegisterEventHandler Cmd = "RegisterEventHandler"

	// Wait waits for a specified number of steps in step() blocks.
	// Args: [stepCount int]
	Wait Cmd = "Wait"

	// SetStep sets the step duration for subsequent commands in step() blocks.
	// Args: [stepDuration int]
	SetStep Cmd = "SetStep"

	// DefineFunction defines a user-defined function.
	// Args: [functionName string, parameters []map[string]any, bodyBlock []OpCode]
	// Each parameter map contains: name, type, isArray, and optionally default
	DefineFunction Cmd = "DefineFunction"
)

// OpCode represents a single instruction for the VM.
// It consists of a command type (Cmd) and a slice of arguments (Args).
// The Args can contain various types including:
// - Primitive values (int, string, float)
// - Variable references (Variable type)
// - Nested OpCode structures for complex expressions
// - Slices of OpCode for block statements
type OpCode struct {
	Cmd  Cmd
	Args []any
}

// Variable represents a variable reference in OpCode arguments.
// This type distinguishes variable references from literal string values.
// When the VM encounters a Variable in Args, it should resolve the variable
// by name from the current scope.
type Variable string
