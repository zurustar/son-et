package engine

import (
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// TimingMode represents the execution timing mode
type TimingMode int

const (
	// TIME mode: Frame-based timing (60 FPS), blocking execution
	TIME TimingMode = iota
	// MIDI_TIME mode: MIDI-synchronized timing, non-blocking execution
	MIDI_TIME
)

// Sequencer represents an execution context for a sequence of OpCodes.
// It maintains execution state, timing information, and variable scope.
type Sequencer struct {
	// Execution state
	commands []interpreter.OpCode // OpCode sequence to execute
	pc       int                  // Program counter
	active   bool                 // Is sequence active?
	noLoop   bool                 // If true, don't loop back to beginning when complete

	// Timing state
	mode         TimingMode // TIME or MIDI_TIME
	waitCount    int        // Ticks remaining in current wait
	ticksPerStep int        // Ticks per step (set by SetStep, used by Wait)

	// Scope state
	vars   map[string]any // Variables in this scope (case-insensitive keys)
	parent *Sequencer     // Parent scope (for variable lookup)

	// Metadata
	id      int // Unique sequence ID
	groupID int // Group ID (for del_us)
}

// NewSequencer creates a new sequencer with the given commands and mode.
// parent can be nil for root-level sequences.
func NewSequencer(commands []interpreter.OpCode, mode TimingMode, parent *Sequencer) *Sequencer {
	// Set default ticksPerStep based on mode
	// TIME mode: 1 step = 50ms = 3 ticks at 60 FPS
	// MIDI_TIME mode: 1 step = 1 tick (MIDI ticks are already at 32nd note resolution)
	ticksPerStep := 3
	if mode == MIDI_TIME {
		ticksPerStep = 1
	}

	return &Sequencer{
		commands:     commands,
		pc:           0,
		active:       true,
		noLoop:       false, // By default, sequences loop
		mode:         mode,
		waitCount:    0,
		ticksPerStep: ticksPerStep,
		vars:         make(map[string]any),
		parent:       parent,
		id:           0, // Will be assigned by engine
		groupID:      0, // Will be assigned by engine
	}
}

// GetVariable retrieves a variable value by walking up the scope chain.
// Variable names are case-insensitive.
// Returns default values if not found: 0 for int, "" for string, []int{} for array.
func (s *Sequencer) GetVariable(name string) any {
	// Convert to lowercase for case-insensitive lookup
	key := strings.ToLower(name)

	// Check current scope
	if val, ok := s.vars[key]; ok {
		return val
	}

	// Walk up scope chain
	if s.parent != nil {
		return s.parent.GetVariable(name)
	}

	// Not found anywhere - return default value
	// We don't know the expected type, so return 0 as default
	// The caller should handle type conversion
	return 0
}

// SetVariable sets a variable value in the appropriate scope.
// If the variable exists in the scope chain, it updates that scope.
// Otherwise, it creates the variable in the current scope.
// Variable names are case-insensitive.
func (s *Sequencer) SetVariable(name string, value any) {
	// Convert to lowercase for case-insensitive lookup
	key := strings.ToLower(name)

	// Check if variable exists in current scope
	if _, ok := s.vars[key]; ok {
		s.vars[key] = value
		return
	}

	// Check if variable exists in parent scope
	if s.parent != nil && s.parent.HasVariable(name) {
		s.parent.SetVariable(name, value)
		return
	}

	// Variable doesn't exist anywhere - create in current scope
	s.vars[key] = value
}

// HasVariable checks if a variable exists in this scope or parent scopes.
// Variable names are case-insensitive.
func (s *Sequencer) HasVariable(name string) bool {
	key := strings.ToLower(name)

	if _, ok := s.vars[key]; ok {
		return true
	}

	if s.parent != nil {
		return s.parent.HasVariable(name)
	}

	return false
}

// GetArrayElement retrieves an array element with auto-expansion.
// If the array doesn't exist, creates it.
// If index >= len(array), expands array and zero-fills.
// Returns 0 if the element doesn't exist after expansion.
func (s *Sequencer) GetArrayElement(name string, index int) int {
	// Get or create array
	val := s.GetVariable(name)

	var arr []int
	switch v := val.(type) {
	case []int:
		arr = v
	case int:
		// Variable exists but is not an array - treat as empty array
		arr = []int{}
	default:
		// Variable doesn't exist - create empty array
		arr = []int{}
	}

	// Auto-expand if needed
	if index >= len(arr) {
		// Expand array to index+1 size, zero-fill
		newArr := make([]int, index+1)
		copy(newArr, arr)
		s.SetVariable(name, newArr)
		return 0 // Newly created element is 0
	}

	return arr[index]
}

// SetArrayElement sets an array element with auto-expansion.
// If the array doesn't exist, creates it.
// If index >= len(array), expands array and zero-fills.
func (s *Sequencer) SetArrayElement(name string, index int, value int) {
	// Get or create array
	val := s.GetVariable(name)

	var arr []int
	switch v := val.(type) {
	case []int:
		arr = v
	case int:
		// Variable exists but is not an array - replace with array
		arr = []int{}
	default:
		// Variable doesn't exist - create empty array
		arr = []int{}
	}

	// Auto-expand if needed
	if index >= len(arr) {
		// Expand array to index+1 size, zero-fill
		newArr := make([]int, index+1)
		copy(newArr, arr)
		arr = newArr
	}

	// Set value
	arr[index] = value

	// Store back (important for scope chain)
	s.SetVariable(name, arr)
}

// GetStringArrayElement retrieves a string array element with auto-expansion.
// If the array doesn't exist, creates it.
// If index >= len(array), expands array and empty-string-fills.
// Returns "" if the element doesn't exist after expansion.
func (s *Sequencer) GetStringArrayElement(name string, index int) string {
	// Get or create array
	val := s.GetVariable(name)

	var arr []string
	switch v := val.(type) {
	case []string:
		arr = v
	case string:
		// Variable exists but is not an array - treat as empty array
		arr = []string{}
	default:
		// Variable doesn't exist - create empty array
		arr = []string{}
	}

	// Auto-expand if needed
	if index >= len(arr) {
		// Expand array to index+1 size, empty-string-fill
		newArr := make([]string, index+1)
		copy(newArr, arr)
		s.SetVariable(name, newArr)
		return "" // Newly created element is ""
	}

	return arr[index]
}

// SetStringArrayElement sets a string array element with auto-expansion.
// If the array doesn't exist, creates it.
// If index >= len(array), expands array and empty-string-fills.
func (s *Sequencer) SetStringArrayElement(name string, index int, value string) {
	// Get or create array
	val := s.GetVariable(name)

	var arr []string
	switch v := val.(type) {
	case []string:
		arr = v
	case string:
		// Variable exists but is not an array - replace with array
		arr = []string{}
	default:
		// Variable doesn't exist - create empty array
		arr = []string{}
	}

	// Auto-expand if needed
	if index >= len(arr) {
		// Expand array to index+1 size, empty-string-fill
		newArr := make([]string, index+1)
		copy(newArr, arr)
		arr = newArr
	}

	// Set value
	arr[index] = value

	// Store back (important for scope chain)
	s.SetVariable(name, arr)
}

// IsActive returns whether the sequence is currently active
func (s *Sequencer) IsActive() bool {
	return s.active
}

// Deactivate marks the sequence as inactive
func (s *Sequencer) Deactivate() {
	s.active = false
}

// IsWaiting returns whether the sequence is currently waiting
func (s *Sequencer) IsWaiting() bool {
	return s.waitCount > 0
}

// DecrementWait decrements the wait counter
func (s *Sequencer) DecrementWait() {
	if s.waitCount > 0 {
		s.waitCount--
	}
}

// GetWaitCount returns the current wait counter
func (s *Sequencer) GetWaitCount() int {
	return s.waitCount
}

// SetWait sets the wait counter
func (s *Sequencer) SetWait(count int) {
	s.waitCount = count
}

// GetPC returns the current program counter
func (s *Sequencer) GetPC() int {
	return s.pc
}

// SetPC sets the program counter
func (s *Sequencer) SetPC(pc int) {
	s.pc = pc
}

// IncrementPC advances the program counter
func (s *Sequencer) IncrementPC() {
	s.pc++
}

// DecrementPC decrements the program counter
func (s *Sequencer) DecrementPC() {
	if s.pc > 0 {
		s.pc--
	}
}

// IsComplete returns whether the sequence has completed execution
func (s *Sequencer) IsComplete() bool {
	// If we're waiting, we're not complete yet
	if s.IsWaiting() {
		return false
	}
	return s.pc >= len(s.commands)
}

// GetCurrentCommand returns the current command to execute
func (s *Sequencer) GetCurrentCommand() *interpreter.OpCode {
	if s.IsComplete() {
		return nil
	}
	return &s.commands[s.pc]
}

// GetMode returns the timing mode
func (s *Sequencer) GetMode() TimingMode {
	return s.mode
}

// GetID returns the sequence ID
func (s *Sequencer) GetID() int {
	return s.id
}

// SetID sets the sequence ID
func (s *Sequencer) SetID(id int) {
	s.id = id
}

// GetGroupID returns the group ID
func (s *Sequencer) GetGroupID() int {
	return s.groupID
}

// SetGroupID sets the group ID
func (s *Sequencer) SetGroupID(groupID int) {
	s.groupID = groupID
}

// GetTicksPerStep returns the ticks per step
func (s *Sequencer) GetTicksPerStep() int {
	return s.ticksPerStep
}

// SetTicksPerStep sets the ticks per step
func (s *Sequencer) SetTicksPerStep(ticks int) {
	s.ticksPerStep = ticks
}

// SetNoLoop sets whether the sequence should loop
func (s *Sequencer) SetNoLoop(noLoop bool) {
	s.noLoop = noLoop
}

// ShouldLoop returns whether the sequence should loop back to the beginning
func (s *Sequencer) ShouldLoop() bool {
	return !s.noLoop
}
