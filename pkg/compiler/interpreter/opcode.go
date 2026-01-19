package interpreter

// OpCmd represents an operation command type.
// Using int enum for type safety and performance.
type OpCmd int

// OpCode command types
const (
	// Assignment and expressions
	OpAssign OpCmd = iota
	OpLiteral
	OpVariable
	OpBinaryOp
	OpUnaryOp

	// Control flow
	OpIf
	OpFor
	OpWhile
	OpSwitch
	OpBreak
	OpContinue

	// Function calls
	OpCall

	// Sequence management
	OpRegisterSequence
	OpWait
	OpDelMe
	OpDelUs
	OpDelAll

	// mes() blocks
	OpRegisterEventHandler

	// Array operations
	OpArrayAccess
	OpArrayAssign
)

// String returns the string representation of an OpCmd for debugging
func (op OpCmd) String() string {
	switch op {
	case OpAssign:
		return "OpAssign"
	case OpLiteral:
		return "OpLiteral"
	case OpVariable:
		return "OpVariable"
	case OpBinaryOp:
		return "OpBinaryOp"
	case OpUnaryOp:
		return "OpUnaryOp"
	case OpIf:
		return "OpIf"
	case OpFor:
		return "OpFor"
	case OpWhile:
		return "OpWhile"
	case OpSwitch:
		return "OpSwitch"
	case OpBreak:
		return "OpBreak"
	case OpContinue:
		return "OpContinue"
	case OpCall:
		return "OpCall"
	case OpRegisterSequence:
		return "OpRegisterSequence"
	case OpWait:
		return "OpWait"
	case OpDelMe:
		return "OpDelMe"
	case OpDelUs:
		return "OpDelUs"
	case OpDelAll:
		return "OpDelAll"
	case OpRegisterEventHandler:
		return "OpRegisterEventHandler"
	case OpArrayAccess:
		return "OpArrayAccess"
	case OpArrayAssign:
		return "OpArrayAssign"
	default:
		return "Unknown"
	}
}

// OpCode represents a single operation in the VM.
// It uses a uniform structure for all operations.
type OpCode struct {
	Cmd  OpCmd // Command type (enum)
	Args []any // Arguments (can contain nested OpCodes)
}

// Variable wraps a string to distinguish variable references from string literals.
type Variable string

// String returns the variable name
func (v Variable) String() string {
	return string(v)
}
