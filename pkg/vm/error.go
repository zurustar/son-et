// Package vm provides error handling for the FILLY virtual machine.
package vm

import (
	"fmt"
)

// ErrorType represents the type of runtime error.
type ErrorType string

const (
	// Fatal errors - execution must stop
	ErrorStackOverflow ErrorType = "STACK_OVERFLOW"
	ErrorOutOfMemory   ErrorType = "OUT_OF_MEMORY"

	// Non-fatal errors - execution continues
	ErrorFileNotFound     ErrorType = "FILE_NOT_FOUND"
	ErrorDivisionByZero   ErrorType = "DIVISION_BY_ZERO"
	ErrorIndexOutOfRange  ErrorType = "INDEX_OUT_OF_RANGE"
	ErrorUndefinedVar     ErrorType = "UNDEFINED_VARIABLE"
	ErrorUndefinedFunc    ErrorType = "UNDEFINED_FUNCTION"
	ErrorInvalidOperation ErrorType = "INVALID_OPERATION"
)

// RuntimeError represents a runtime error in the VM.
// Requirement 11.1: When runtime error occurs, system logs error with context information.
// Requirement 11.7: System provides error messages including line number when available.
type RuntimeError struct {
	Type    ErrorType
	Message string
	Line    int    // Line number if available, -1 otherwise
	File    string // File name if available
	Context string // Additional context information
}

// Error implements the error interface.
func (e *RuntimeError) Error() string {
	if e.Line >= 0 && e.File != "" {
		return fmt.Sprintf("[%s] %s at %s:%d", e.Type, e.Message, e.File, e.Line)
	}
	if e.Line >= 0 {
		return fmt.Sprintf("[%s] %s at line %d", e.Type, e.Message, e.Line)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// IsFatal returns true if the error is fatal and execution should stop.
// Requirement 11.8: System continues execution after non-fatal errors.
func (e *RuntimeError) IsFatal() bool {
	switch e.Type {
	case ErrorStackOverflow, ErrorOutOfMemory, ErrorUndefinedFunc:
		return true
	default:
		return false
	}
}

// NewRuntimeError creates a new RuntimeError.
func NewRuntimeError(errType ErrorType, message string) *RuntimeError {
	return &RuntimeError{
		Type:    errType,
		Message: message,
		Line:    -1,
	}
}

// NewRuntimeErrorWithLine creates a new RuntimeError with line information.
func NewRuntimeErrorWithLine(errType ErrorType, message string, line int) *RuntimeError {
	return &RuntimeError{
		Type:    errType,
		Message: message,
		Line:    line,
	}
}

// NewRuntimeErrorWithContext creates a new RuntimeError with full context.
func NewRuntimeErrorWithContext(errType ErrorType, message, file string, line int, context string) *RuntimeError {
	return &RuntimeError{
		Type:    errType,
		Message: message,
		File:    file,
		Line:    line,
		Context: context,
	}
}

// Error helper functions for common error types

// NewFileNotFoundError creates a file not found error.
// Requirement 11.2: When file is not found, system logs error and continues execution.
func NewFileNotFoundError(filename string) *RuntimeError {
	return NewRuntimeError(ErrorFileNotFound, fmt.Sprintf("file not found: %s", filename))
}

// NewDivisionByZeroError creates a division by zero error.
// Requirement 11.3: When division by zero occurs, system logs error and returns zero.
func NewDivisionByZeroError() *RuntimeError {
	return NewRuntimeError(ErrorDivisionByZero, "division by zero")
}

// NewIndexOutOfRangeError creates an index out of range error.
// Requirement 11.4: When array index is out of range, system logs error and returns zero.
func NewIndexOutOfRangeError(index int64, length int) *RuntimeError {
	return NewRuntimeError(ErrorIndexOutOfRange, fmt.Sprintf("index %d out of range (length %d)", index, length))
}

// NewUndefinedVariableError creates an undefined variable error.
// Requirement 11.5: When variable is not found, system creates it with default value.
func NewUndefinedVariableError(name string) *RuntimeError {
	return NewRuntimeError(ErrorUndefinedVar, fmt.Sprintf("undefined variable: %s", name))
}

// NewUndefinedFunctionError creates an undefined function error.
// Requirement 11.6: When function is not found, system logs error and continues execution.
func NewUndefinedFunctionError(name string) *RuntimeError {
	return NewRuntimeError(ErrorUndefinedFunc, fmt.Sprintf("undefined function: %s", name))
}

// NewStackOverflowError creates a stack overflow error.
// Requirement 20.8: When stack overflow occurs, system logs error and terminates execution.
func NewStackOverflowError(depth int) *RuntimeError {
	return NewRuntimeError(ErrorStackOverflow, fmt.Sprintf("stack overflow: depth %d exceeds maximum %d", depth, MaxStackDepth))
}
