package interpreter

// OpCode represents a single VM instruction
// This matches the OpCode structure used by the engine VM
type OpCode struct {
	Cmd  string // Command name (e.g., "Assign", "Call", "If", "For")
	Args []any  // Command arguments (can contain nested OpCodes or Variables)
}

// Variable represents a reference to a variable name
// This is used to distinguish variable references from literal values
type Variable string
