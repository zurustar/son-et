package vm

import (
	"strings"
	"testing"
)

func TestRuntimeError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *RuntimeError
		contains []string
	}{
		{
			name:     "basic error",
			err:      NewRuntimeError(ErrorDivisionByZero, "division by zero"),
			contains: []string{"DIVISION_BY_ZERO", "division by zero"},
		},
		{
			name:     "error with line",
			err:      NewRuntimeErrorWithLine(ErrorIndexOutOfRange, "index out of range", 42),
			contains: []string{"INDEX_OUT_OF_RANGE", "index out of range", "line 42"},
		},
		{
			name:     "error with context",
			err:      NewRuntimeErrorWithContext(ErrorFileNotFound, "file not found", "test.tfy", 10, "loading script"),
			contains: []string{"FILE_NOT_FOUND", "file not found", "test.tfy", "10"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, s := range tt.contains {
				if !strings.Contains(errStr, s) {
					t.Errorf("error string %q should contain %q", errStr, s)
				}
			}
		})
	}
}

func TestRuntimeError_IsFatal(t *testing.T) {
	tests := []struct {
		name    string
		errType ErrorType
		fatal   bool
	}{
		{"stack overflow is fatal", ErrorStackOverflow, true},
		{"out of memory is fatal", ErrorOutOfMemory, true},
		{"file not found is not fatal", ErrorFileNotFound, false},
		{"division by zero is not fatal", ErrorDivisionByZero, false},
		{"index out of range is not fatal", ErrorIndexOutOfRange, false},
		{"undefined variable is not fatal", ErrorUndefinedVar, false},
		{"undefined function is not fatal", ErrorUndefinedFunc, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := NewRuntimeError(tt.errType, "test")
			if err.IsFatal() != tt.fatal {
				t.Errorf("IsFatal() = %v, want %v", err.IsFatal(), tt.fatal)
			}
		})
	}
}

func TestErrorHelperFunctions(t *testing.T) {
	t.Run("NewFileNotFoundError", func(t *testing.T) {
		err := NewFileNotFoundError("test.wav")
		if err.Type != ErrorFileNotFound {
			t.Errorf("Type = %v, want %v", err.Type, ErrorFileNotFound)
		}
		if !strings.Contains(err.Message, "test.wav") {
			t.Errorf("Message should contain filename")
		}
	})

	t.Run("NewDivisionByZeroError", func(t *testing.T) {
		err := NewDivisionByZeroError()
		if err.Type != ErrorDivisionByZero {
			t.Errorf("Type = %v, want %v", err.Type, ErrorDivisionByZero)
		}
	})

	t.Run("NewIndexOutOfRangeError", func(t *testing.T) {
		err := NewIndexOutOfRangeError(10, 5)
		if err.Type != ErrorIndexOutOfRange {
			t.Errorf("Type = %v, want %v", err.Type, ErrorIndexOutOfRange)
		}
		if !strings.Contains(err.Message, "10") || !strings.Contains(err.Message, "5") {
			t.Errorf("Message should contain index and length")
		}
	})

	t.Run("NewUndefinedVariableError", func(t *testing.T) {
		err := NewUndefinedVariableError("myVar")
		if err.Type != ErrorUndefinedVar {
			t.Errorf("Type = %v, want %v", err.Type, ErrorUndefinedVar)
		}
		if !strings.Contains(err.Message, "myVar") {
			t.Errorf("Message should contain variable name")
		}
	})

	t.Run("NewUndefinedFunctionError", func(t *testing.T) {
		err := NewUndefinedFunctionError("myFunc")
		if err.Type != ErrorUndefinedFunc {
			t.Errorf("Type = %v, want %v", err.Type, ErrorUndefinedFunc)
		}
		if !strings.Contains(err.Message, "myFunc") {
			t.Errorf("Message should contain function name")
		}
	})

	t.Run("NewStackOverflowError", func(t *testing.T) {
		err := NewStackOverflowError(1001)
		if err.Type != ErrorStackOverflow {
			t.Errorf("Type = %v, want %v", err.Type, ErrorStackOverflow)
		}
		if !err.IsFatal() {
			t.Errorf("Stack overflow should be fatal")
		}
	})
}
