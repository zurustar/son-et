package interpreter

// Function represents a user-defined function
type Function struct {
	Name       string            // Function name
	Parameters []Parameter       // Function parameters
	Body       []OpCode          // Function body as OpCode sequence
	Locals     map[string]string // Local variable name -> type (case-insensitive)
}

// Parameter represents a function parameter
type Parameter struct {
	Name    string // Parameter name
	Type    string // Parameter type ("int", "string", etc.)
	Default any    // Default value (nil if required)
}

// NewFunction creates a new Function instance
func NewFunction(name string) *Function {
	return &Function{
		Name:       name,
		Parameters: []Parameter{},
		Body:       []OpCode{},
		Locals:     make(map[string]string),
	}
}
