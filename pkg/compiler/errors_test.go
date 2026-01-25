package compiler

import (
	"errors"
	"strings"
	"testing"
)

// TestCompileError_Error tests the Error() method of CompileError.
func TestCompileError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *CompileError
		contains []string
	}{
		{
			name: "lexer error without context",
			err: &CompileError{
				Phase:   "lexer",
				Message: "illegal character '@'",
				Line:    5,
				Column:  10,
			},
			contains: []string{"lexer error", "line 5", "column 10", "illegal character '@'"},
		},
		{
			name: "parser error without context",
			err: &CompileError{
				Phase:   "parser",
				Message: "expected ';', got '}'",
				Line:    12,
				Column:  25,
			},
			contains: []string{"parser error", "line 12", "column 25", "expected ';', got '}'"},
		},
		{
			name: "compiler error without context",
			err: &CompileError{
				Phase:   "compiler",
				Message: "unknown AST node type: *parser.UnknownNode",
				Line:    0,
				Column:  0,
			},
			contains: []string{"compiler error", "line 0", "column 0", "unknown AST node type"},
		},
		{
			name: "error with context",
			err: &CompileError{
				Phase:   "parser",
				Message: "unexpected token",
				Line:    3,
				Column:  5,
				Context: "> 3 | int x = ;\n      ^",
			},
			contains: []string{"parser error", "line 3", "column 5", "unexpected token", "> 3 |"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errStr := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errStr, substr) {
					t.Errorf("Error() = %q, want to contain %q", errStr, substr)
				}
			}
		})
	}
}

// TestNewLexerError tests the NewLexerError helper function.
func TestNewLexerError(t *testing.T) {
	err := NewLexerError("illegal character '@'", 5, 10)

	if err.Phase != "lexer" {
		t.Errorf("Phase = %q, want %q", err.Phase, "lexer")
	}
	if err.Message != "illegal character '@'" {
		t.Errorf("Message = %q, want %q", err.Message, "illegal character '@'")
	}
	if err.Line != 5 {
		t.Errorf("Line = %d, want %d", err.Line, 5)
	}
	if err.Column != 10 {
		t.Errorf("Column = %d, want %d", err.Column, 10)
	}
	if err.Context != "" {
		t.Errorf("Context = %q, want empty", err.Context)
	}
}

// TestNewParserError tests the NewParserError helper function.
func TestNewParserError(t *testing.T) {
	err := NewParserError("expected ';'", 12, 25)

	if err.Phase != "parser" {
		t.Errorf("Phase = %q, want %q", err.Phase, "parser")
	}
	if err.Message != "expected ';'" {
		t.Errorf("Message = %q, want %q", err.Message, "expected ';'")
	}
	if err.Line != 12 {
		t.Errorf("Line = %d, want %d", err.Line, 12)
	}
	if err.Column != 25 {
		t.Errorf("Column = %d, want %d", err.Column, 25)
	}
}

// TestNewCompilerError tests the NewCompilerError helper function.
func TestNewCompilerError(t *testing.T) {
	err := NewCompilerError("unknown node type", 0, 0)

	if err.Phase != "compiler" {
		t.Errorf("Phase = %q, want %q", err.Phase, "compiler")
	}
	if err.Message != "unknown node type" {
		t.Errorf("Message = %q, want %q", err.Message, "unknown node type")
	}
	if err.Line != 0 {
		t.Errorf("Line = %d, want %d", err.Line, 0)
	}
	if err.Column != 0 {
		t.Errorf("Column = %d, want %d", err.Column, 0)
	}
}

// TestGenerateErrorContext tests the GenerateErrorContext function.
func TestGenerateErrorContext(t *testing.T) {
	source := `int a = 1;
int b = 2;
int c = 3;
int d = ;
int e = 5;
int f = 6;
int g = 7;`

	tests := []struct {
		name        string
		source      string
		line        int
		column      int
		contains    []string
		notContains []string
	}{
		{
			name:   "error in middle of file",
			source: source,
			line:   4,
			column: 9,
			contains: []string{
				"2 |", "int b = 2;", // 2 lines before
				"3 |", "int c = 3;", // 1 line before
				"> 4 |", "int d = ;", // error line with marker
				"^",                 // pointer
				"5 |", "int e = 5;", // 1 line after
				"6 |", "int f = 6;", // 2 lines after
			},
			notContains: []string{
				"1 |", // should not include line 1 (more than 2 lines before)
				"7 |", // should not include line 7 (more than 2 lines after)
			},
		},
		{
			name:   "error at beginning of file",
			source: source,
			line:   1,
			column: 5,
			contains: []string{
				"> 1 |", "int a = 1;", // error line
				"^",                 // pointer
				"2 |", "int b = 2;", // 1 line after
				"3 |", "int c = 3;", // 2 lines after
			},
			notContains: []string{
				"4 |", // should not include line 4 (more than 2 lines after)
			},
		},
		{
			name:   "error at end of file",
			source: source,
			line:   7,
			column: 5,
			contains: []string{
				"5 |", "int e = 5;", // 2 lines before
				"6 |", "int f = 6;", // 1 line before
				"> 7 |", "int g = 7;", // error line
				"^", // pointer
			},
			notContains: []string{
				"4 |", // should not include line 4 (more than 2 lines before)
			},
		},
		{
			name:     "empty source",
			source:   "",
			line:     1,
			column:   1,
			contains: []string{},
		},
		{
			name:     "invalid line number",
			source:   source,
			line:     0,
			column:   1,
			contains: []string{},
		},
		{
			name:     "line number exceeds source",
			source:   source,
			line:     100,
			column:   1,
			contains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := GenerateErrorContext(tt.source, tt.line, tt.column)

			for _, substr := range tt.contains {
				if !strings.Contains(context, substr) {
					t.Errorf("GenerateErrorContext() = %q, want to contain %q", context, substr)
				}
			}

			for _, substr := range tt.notContains {
				if strings.Contains(context, substr) {
					t.Errorf("GenerateErrorContext() = %q, should not contain %q", context, substr)
				}
			}
		})
	}
}

// TestNewLexerErrorWithContext tests the NewLexerErrorWithContext helper function.
func TestNewLexerErrorWithContext(t *testing.T) {
	source := `int x = 1;
int y = @;
int z = 3;`

	err := NewLexerErrorWithContext("illegal character '@'", 2, 9, source)

	if err.Phase != "lexer" {
		t.Errorf("Phase = %q, want %q", err.Phase, "lexer")
	}
	if err.Message != "illegal character '@'" {
		t.Errorf("Message = %q, want %q", err.Message, "illegal character '@'")
	}
	if err.Line != 2 {
		t.Errorf("Line = %d, want %d", err.Line, 2)
	}
	if err.Column != 9 {
		t.Errorf("Column = %d, want %d", err.Column, 9)
	}
	if err.Context == "" {
		t.Error("Context should not be empty")
	}
	if !strings.Contains(err.Context, "> 2 |") {
		t.Errorf("Context should contain error line marker, got %q", err.Context)
	}
	if !strings.Contains(err.Context, "^") {
		t.Errorf("Context should contain pointer, got %q", err.Context)
	}
}

// TestNewParserErrorWithContext tests the NewParserErrorWithContext helper function.
func TestNewParserErrorWithContext(t *testing.T) {
	source := `main() {
    int x = 5;
    int y = ;
    int z = 10;
}`

	err := NewParserErrorWithContext("expected expression", 3, 13, source)

	if err.Phase != "parser" {
		t.Errorf("Phase = %q, want %q", err.Phase, "parser")
	}
	if err.Context == "" {
		t.Error("Context should not be empty")
	}
	if !strings.Contains(err.Context, "> 3 |") {
		t.Errorf("Context should contain error line marker, got %q", err.Context)
	}
}

// TestNewCompilerErrorWithContext tests the NewCompilerErrorWithContext helper function.
func TestNewCompilerErrorWithContext(t *testing.T) {
	source := `main() {
    unknownStatement;
}`

	err := NewCompilerErrorWithContext("unknown statement type", 2, 5, source)

	if err.Phase != "compiler" {
		t.Errorf("Phase = %q, want %q", err.Phase, "compiler")
	}
	if err.Context == "" {
		t.Error("Context should not be empty")
	}
}

// TestIsCompileError tests the IsCompileError helper function.
func TestIsCompileError(t *testing.T) {
	tests := []struct {
		name      string
		err       error
		wantOk    bool
		wantPhase string
	}{
		{
			name:      "CompileError",
			err:       NewLexerError("test", 1, 1),
			wantOk:    true,
			wantPhase: "lexer",
		},
		{
			name:      "standard error",
			err:       errors.New("standard error"),
			wantOk:    false,
			wantPhase: "",
		},
		{
			name:      "nil error",
			err:       nil,
			wantOk:    false,
			wantPhase: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce, ok := IsCompileError(tt.err)
			if ok != tt.wantOk {
				t.Errorf("IsCompileError() ok = %v, want %v", ok, tt.wantOk)
			}
			if ok && ce.Phase != tt.wantPhase {
				t.Errorf("IsCompileError() Phase = %q, want %q", ce.Phase, tt.wantPhase)
			}
		})
	}
}

// TestWrapError tests the WrapError helper function.
func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error message")
	wrapped := WrapError(originalErr, "parser", 10, 20)

	if wrapped.Phase != "parser" {
		t.Errorf("Phase = %q, want %q", wrapped.Phase, "parser")
	}
	if wrapped.Message != "original error message" {
		t.Errorf("Message = %q, want %q", wrapped.Message, "original error message")
	}
	if wrapped.Line != 10 {
		t.Errorf("Line = %d, want %d", wrapped.Line, 10)
	}
	if wrapped.Column != 20 {
		t.Errorf("Column = %d, want %d", wrapped.Column, 20)
	}
}

// TestCollectErrors tests the CollectErrors helper function.
func TestCollectErrors(t *testing.T) {
	errs := []error{
		NewLexerError("lexer error", 1, 1),
		errors.New("standard error"),
		NewParserError("parser error", 2, 2),
	}

	collected := CollectErrors(errs, "compiler")

	if len(collected) != 3 {
		t.Fatalf("CollectErrors() returned %d errors, want 3", len(collected))
	}

	// First error should be the original lexer error
	if collected[0].Phase != "lexer" {
		t.Errorf("collected[0].Phase = %q, want %q", collected[0].Phase, "lexer")
	}

	// Second error should be wrapped with default phase
	if collected[1].Phase != "compiler" {
		t.Errorf("collected[1].Phase = %q, want %q", collected[1].Phase, "compiler")
	}
	if collected[1].Message != "standard error" {
		t.Errorf("collected[1].Message = %q, want %q", collected[1].Message, "standard error")
	}

	// Third error should be the original parser error
	if collected[2].Phase != "parser" {
		t.Errorf("collected[2].Phase = %q, want %q", collected[2].Phase, "parser")
	}
}

// TestCompileError_ImplementsError tests that CompileError implements the error interface.
func TestCompileError_ImplementsError(t *testing.T) {
	var err error = NewLexerError("test", 1, 1)
	if err == nil {
		t.Error("CompileError should implement error interface")
	}
}

// TestGenerateErrorContext_PointerPosition tests that the pointer is correctly positioned.
func TestGenerateErrorContext_PointerPosition(t *testing.T) {
	source := "int x = 5;"

	tests := []struct {
		name   string
		column int
	}{
		{"column 1", 1},
		{"column 5", 5},
		{"column 10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			context := GenerateErrorContext(source, 1, tt.column)

			// The context should contain the pointer
			if !strings.Contains(context, "^") {
				t.Errorf("Context should contain pointer '^', got %q", context)
			}

			// Count spaces before the pointer in the pointer line
			lines := strings.Split(context, "\n")
			var pointerLine string
			for _, line := range lines {
				if strings.Contains(line, "^") && !strings.Contains(line, "|") {
					pointerLine = line
					break
				}
			}

			if pointerLine == "" {
				t.Errorf("Could not find pointer line in context: %q", context)
			}
		})
	}
}
