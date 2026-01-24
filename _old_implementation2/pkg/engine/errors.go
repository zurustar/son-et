package engine

import (
	"fmt"
)

// ErrorType represents the type of error.
type ErrorType int

const (
	// ErrorTypeParser represents a parsing error
	ErrorTypeParser ErrorType = iota
	// ErrorTypeRuntime represents a runtime execution error
	ErrorTypeRuntime
	// ErrorTypeAsset represents an asset loading error
	ErrorTypeAsset
)

// EngineError represents an error with context information.
type EngineError struct {
	Type     ErrorType
	Message  string
	Line     int    // For parsing errors
	Column   int    // For parsing errors
	OpCode   string // For runtime errors
	Args     string // For runtime errors
	Filename string // For asset errors
	Cause    error  // Underlying error
}

// Error implements the error interface.
func (e *EngineError) Error() string {
	switch e.Type {
	case ErrorTypeParser:
		return fmt.Sprintf("Parse error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	case ErrorTypeRuntime:
		if e.Args != "" {
			return fmt.Sprintf("Runtime error in %s(%s): %s", e.OpCode, e.Args, e.Message)
		}
		return fmt.Sprintf("Runtime error in %s: %s", e.OpCode, e.Message)
	case ErrorTypeAsset:
		if e.Cause != nil {
			return fmt.Sprintf("Asset error loading '%s': %s (cause: %v)", e.Filename, e.Message, e.Cause)
		}
		return fmt.Sprintf("Asset error loading '%s': %s", e.Filename, e.Message)
	default:
		return e.Message
	}
}

// Unwrap returns the underlying error.
func (e *EngineError) Unwrap() error {
	return e.Cause
}

// NewParseError creates a new parsing error.
func NewParseError(line, column int, message string, args ...interface{}) *EngineError {
	return &EngineError{
		Type:    ErrorTypeParser,
		Message: fmt.Sprintf(message, args...),
		Line:    line,
		Column:  column,
	}
}

// NewRuntimeError creates a new runtime error.
func NewRuntimeError(opCode string, args string, message string, msgArgs ...interface{}) *EngineError {
	return &EngineError{
		Type:    ErrorTypeRuntime,
		Message: fmt.Sprintf(message, msgArgs...),
		OpCode:  opCode,
		Args:    args,
	}
}

// NewAssetError creates a new asset loading error.
func NewAssetError(filename string, message string, cause error) *EngineError {
	return &EngineError{
		Type:     ErrorTypeAsset,
		Message:  message,
		Filename: filename,
		Cause:    cause,
	}
}

// ReportError logs an error with appropriate level and context.
func (e *Engine) ReportError(err error) {
	if engineErr, ok := err.(*EngineError); ok {
		e.logger.LogError("%s", engineErr.Error())
	} else {
		e.logger.LogError("Error: %v", err)
	}
}

// ReportParseError reports a parsing error.
func (e *Engine) ReportParseError(line, column int, message string, args ...interface{}) {
	err := NewParseError(line, column, message, args...)
	e.ReportError(err)
}

// ReportRuntimeError reports a runtime error.
func (e *Engine) ReportRuntimeError(opCode string, args string, message string, msgArgs ...interface{}) {
	err := NewRuntimeError(opCode, args, message, msgArgs...)
	e.ReportError(err)
}

// ReportAssetError reports an asset loading error.
func (e *Engine) ReportAssetError(filename string, message string, cause error) {
	err := NewAssetError(filename, message, cause)
	e.ReportError(err)
}

// EndStepSignal is a special signal (not an error) used to break out of step blocks.
// It's returned by end_step to exit the step loop.
type EndStepSignal struct{}

func (e *EndStepSignal) Error() string {
	return "end_step"
}

// IsEndStepSignal checks if an error is an EndStepSignal.
func IsEndStepSignal(err error) bool {
	_, ok := err.(*EndStepSignal)
	return ok
}
