package interpreter

// OpCmd represents an OpCode command type
type OpCmd int

// OpCode command constants
const (
	// Special/Internal commands
	OpLiteral OpCmd = iota // Literal value
	OpVarRef               // Variable reference

	// Assignment operations
	OpAssign      // Variable assignment
	OpAssignArray // Array element assignment

	// Control flow
	OpIf       // If statement
	OpFor      // For loop
	OpWhile    // While loop
	OpDoWhile  // Do-while loop
	OpSwitch   // Switch statement
	OpBreak    // Break statement
	OpContinue // Continue statement

	// Function operations
	OpCall // Function call

	// Sequencer operations
	OpRegisterSequence // Register a sequence (mes block)
	OpStep             // Step block
	OpWait             // Wait statement
	OpSetStep          // Set step count

	// Engine commands - Picture operations
	OpLoadPic   // Load picture
	OpCreatePic // Create picture
	OpDelPic    // Delete picture
	OpMovePic   // Move picture
	OpPutCast   // Put cast
	OpMoveCast  // Move cast

	// Engine commands - Window operations
	OpOpenWin     // Open window
	OpCloseWin    // Close window
	OpCloseWinAll // Close all windows
	OpMoveWin     // Move window

	// Engine commands - Text operations
	OpTextColor // Set text color
	OpTextWrite // Write text

	// Engine commands - Audio operations
	OpPlayWAVE // Play WAV file
	OpPlayMIDI // Play MIDI file

	// Engine commands - System operations
	OpExitTitle // Exit to title

	// Expression operations
	OpInfix  // Binary operation (e.g., +, -, *, /, ==, !=, <, >, etc.)
	OpPrefix // Unary operation (e.g., -, !)
	OpIndex  // Array index access
	OpArray  // Array declaration
)

// String returns the string representation of an OpCmd for debugging
func (op OpCmd) String() string {
	switch op {
	case OpLiteral:
		return "Literal"
	case OpVarRef:
		return "VarRef"
	case OpAssign:
		return "Assign"
	case OpAssignArray:
		return "AssignArray"
	case OpIf:
		return "If"
	case OpFor:
		return "For"
	case OpWhile:
		return "While"
	case OpDoWhile:
		return "DoWhile"
	case OpSwitch:
		return "Switch"
	case OpBreak:
		return "Break"
	case OpContinue:
		return "Continue"
	case OpCall:
		return "Call"
	case OpRegisterSequence:
		return "RegisterSequence"
	case OpStep:
		return "Step"
	case OpWait:
		return "Wait"
	case OpSetStep:
		return "SetStep"
	case OpLoadPic:
		return "LoadPic"
	case OpCreatePic:
		return "CreatePic"
	case OpDelPic:
		return "DelPic"
	case OpMovePic:
		return "MovePic"
	case OpPutCast:
		return "PutCast"
	case OpMoveCast:
		return "MoveCast"
	case OpOpenWin:
		return "OpenWin"
	case OpCloseWin:
		return "CloseWin"
	case OpCloseWinAll:
		return "CloseWinAll"
	case OpMoveWin:
		return "MoveWin"
	case OpTextColor:
		return "TextColor"
	case OpTextWrite:
		return "TextWrite"
	case OpPlayWAVE:
		return "PlayWAVE"
	case OpPlayMIDI:
		return "PlayMIDI"
	case OpExitTitle:
		return "ExitTitle"
	case OpInfix:
		return "Infix"
	case OpPrefix:
		return "Prefix"
	case OpIndex:
		return "Index"
	case OpArray:
		return "Array"
	default:
		return "Unknown"
	}
}

// OpCode represents a single VM instruction
// This matches the OpCode structure used by the engine VM
type OpCode struct {
	Cmd  OpCmd // Command type (enum)
	Args []any // Command arguments (can contain nested OpCodes or Variables)
}

// Variable represents a reference to a variable name
// This is used to distinguish variable references from literal values
type Variable string
