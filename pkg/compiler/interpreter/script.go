package interpreter

// Script represents a compiled TFY script
type Script struct {
	Globals   map[string]string    // Variable name -> type (case-insensitive)
	Functions map[string]*Function // Function name -> Function (case-insensitive)
	Main      *Function            // Main function OpCode
	Assets    []string             // Asset file names
}

// NewScript creates a new Script instance
func NewScript() *Script {
	return &Script{
		Globals:   make(map[string]string),
		Functions: make(map[string]*Function),
		Assets:    []string{},
	}
}
