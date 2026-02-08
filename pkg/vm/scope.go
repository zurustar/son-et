// Package vm provides scope management for the FILLY virtual machine.
package vm

import (
	"sync"
)

// Scope represents a variable scope in the VM.
// It supports hierarchical scoping with parent scope lookup.
//
// Requirement 9.1: When variable is declared at top level, system stores it in global scope.
// Requirement 9.2: When variable is declared inside function, system stores it in local scope.
// Requirement 9.5: When variable is accessed, system searches local scope first, then global scope.
type Scope struct {
	variables map[string]any
	parent    *Scope
	mu        sync.RWMutex
}

// NewScope creates a new scope with an optional parent scope.
// Requirement 9.3: When function is called, system creates new local scope.
func NewScope(parent *Scope) *Scope {
	return &Scope{
		variables: make(map[string]any),
		parent:    parent,
	}
}

// Get retrieves a variable value by name.
// It first searches the current scope, then parent scopes.
// Requirement 9.5: When variable is accessed, system searches local scope first, then global scope.
//
// Parameters:
//   - name: The variable name to look up
//
// Returns:
//   - any: The variable value
//   - bool: true if the variable was found, false otherwise
func (s *Scope) Get(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// First, check current scope
	if value, ok := s.variables[name]; ok {
		return value, true
	}

	// Then, check parent scope
	if s.parent != nil {
		return s.parent.Get(name)
	}

	return nil, false
}

// Set sets a variable value in the appropriate scope.
// If the variable exists in any parent scope, it updates that scope.
// Otherwise, it creates the variable in the current scope.
// Requirement 9.6: When variable is assigned without prior declaration, system creates it in current scope.
//
// Parameters:
//   - name: The variable name
//   - value: The value to set
func (s *Scope) Set(name string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if variable exists in current scope
	if _, ok := s.variables[name]; ok {
		s.variables[name] = value
		return
	}

	// Check if variable exists in parent scope (without lock since parent.Set will lock)
	if s.parent != nil {
		s.mu.Unlock()
		if _, exists := s.parent.Get(name); exists {
			s.parent.Set(name, value)
			s.mu.Lock()
			return
		}
		s.mu.Lock()
	}

	// Create in current scope
	s.variables[name] = value
}

// GetLocal retrieves a variable value only from the current scope (not parent).
//
// Parameters:
//   - name: The variable name to look up
//
// Returns:
//   - any: The variable value
//   - bool: true if the variable was found in this scope, false otherwise
func (s *Scope) GetLocal(name string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.variables[name]
	return value, ok
}

// SetLocal sets a variable value only in the current scope.
// This is used for function parameters and local variable declarations.
// Requirement 9.7: When function parameters are passed, system binds them to local scope.
//
// Parameters:
//   - name: The variable name
//   - value: The value to set
func (s *Scope) SetLocal(name string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variables[name] = value
}

// Delete removes a variable from the current scope.
//
// Parameters:
//   - name: The variable name to delete
//
// Returns:
//   - bool: true if the variable was deleted, false if it didn't exist
func (s *Scope) Delete(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.variables[name]; ok {
		delete(s.variables, name)
		return true
	}
	return false
}

// Has checks if a variable exists in this scope or any parent scope.
//
// Parameters:
//   - name: The variable name to check
//
// Returns:
//   - bool: true if the variable exists, false otherwise
func (s *Scope) Has(name string) bool {
	_, ok := s.Get(name)
	return ok
}

// HasLocal checks if a variable exists only in the current scope.
//
// Parameters:
//   - name: The variable name to check
//
// Returns:
//   - bool: true if the variable exists in this scope, false otherwise
func (s *Scope) HasLocal(name string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.variables[name]
	return ok
}

// Parent returns the parent scope.
//
// Returns:
//   - *Scope: The parent scope, or nil if this is the root scope
func (s *Scope) Parent() *Scope {
	return s.parent
}

// Keys returns all variable names in the current scope (not including parent).
//
// Returns:
//   - []string: Slice of variable names
func (s *Scope) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.variables))
	for k := range s.variables {
		keys = append(keys, k)
	}
	return keys
}

// AllKeys returns all variable names in this scope and all parent scopes.
//
// Returns:
//   - []string: Slice of all variable names (may contain duplicates if shadowed)
func (s *Scope) AllKeys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.variables))
	for k := range s.variables {
		keys = append(keys, k)
	}

	if s.parent != nil {
		keys = append(keys, s.parent.AllKeys()...)
	}

	return keys
}

// Clear removes all variables from the current scope.
func (s *Scope) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.variables = make(map[string]any)
}

// Size returns the number of variables in the current scope (not including parent).
//
// Returns:
//   - int: Number of variables
func (s *Scope) Size() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.variables)
}
