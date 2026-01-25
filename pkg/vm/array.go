// Package vm provides array support for the FILLY virtual machine.
package vm

import (
	"sync"
)

// Array represents a dynamic array in the FILLY VM.
// It wraps a slice to support pass-by-reference semantics.
// Requirement 19.8: When array is passed to function, system passes it by reference.
type Array struct {
	elements []any
	mu       sync.RWMutex
}

// NewArray creates a new Array with the specified initial size.
// All elements are initialized to zero.
// Requirement 19.1: When array is declared, system allocates storage for array.
// Requirement 19.7: System initializes new array elements to zero.
func NewArray(size int) *Array {
	elements := make([]any, size)
	for i := range elements {
		elements[i] = int64(0)
	}
	return &Array{elements: elements}
}

// NewArrayFromSlice creates a new Array from an existing slice.
func NewArrayFromSlice(slice []any) *Array {
	return &Array{elements: slice}
}

// Get retrieves the element at the specified index.
// Requirement 19.2: When array element is accessed, system returns value at specified index.
// Requirement 19.4: When array index is negative, system logs error and returns zero.
func (a *Array) Get(index int64) (any, bool) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if index < 0 || int(index) >= len(a.elements) {
		return int64(0), false
	}
	return a.elements[index], true
}

// Set sets the element at the specified index.
// If the index exceeds the current size, the array is automatically expanded.
// Requirement 19.3: When array element is assigned, system stores value at specified index.
// Requirement 19.5: When array index exceeds array size, system automatically expands array.
// Requirement 19.6: System supports dynamic array resizing.
// Requirement 19.7: System initializes new array elements to zero.
func (a *Array) Set(index int64, value any) {
	a.mu.Lock()
	defer a.mu.Unlock()

	if index < 0 {
		return
	}

	// Expand if necessary
	if int(index) >= len(a.elements) {
		newElements := make([]any, int(index)+1)
		copy(newElements, a.elements)
		// Initialize new elements to zero
		for i := len(a.elements); i < len(newElements); i++ {
			newElements[i] = int64(0)
		}
		a.elements = newElements
	}

	a.elements[index] = value
}

// Len returns the current length of the array.
func (a *Array) Len() int {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return len(a.elements)
}

// ToSlice returns a copy of the underlying slice.
func (a *Array) ToSlice() []any {
	a.mu.RLock()
	defer a.mu.RUnlock()
	result := make([]any, len(a.elements))
	copy(result, a.elements)
	return result
}
