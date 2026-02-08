// Package compiler provides the compilation pipeline for FILLY scripts (.TFY files).
// This file defines the CompileError type for structured error reporting.
package compiler

import (
	"fmt"
	"strings"
)

// CompileError represents a structured compilation error with location information.
// It implements the error interface and provides detailed context about where
// the error occurred in the source code.
//
// Requirement 5.1: Lexer reports illegal characters with character, line, and column.
// Requirement 5.2: Parser reports syntax errors with expected/actual token types, line, and column.
// Requirement 5.3: Parser includes source code context (2 lines before and after error).
// Requirement 5.4: Parser includes pointer (^) indicating error column.
// Requirement 5.5: Compiler reports unknown AST node types in error messages.
// Requirement 5.6: System collects all errors and returns them to caller.
type CompileError struct {
	// Phase indicates which compilation phase generated the error.
	// Valid values: "lexer", "parser", "compiler"
	Phase string

	// Message is the human-readable error description.
	Message string

	// Line is the 1-indexed line number where the error occurred.
	Line int

	// Column is the 1-indexed column number where the error occurred.
	Column int

	// Context contains the source code around the error location.
	// This includes 2 lines before and after the error line,
	// with a pointer (^) indicating the error column.
	Context string
}

// Error implements the error interface.
// It returns a formatted error message including phase, location, message, and context.
func (e *CompileError) Error() string {
	if e.Context != "" {
		return fmt.Sprintf("%s error at line %d, column %d: %s\n%s",
			e.Phase, e.Line, e.Column, e.Message, e.Context)
	}
	return fmt.Sprintf("%s error at line %d, column %d: %s",
		e.Phase, e.Line, e.Column, e.Message)
}




// NewParserErrorWithContext creates a new CompileError for parser phase errors with source context.
//
// Parameters:
//   - message: The error description
//   - line: The 1-indexed line number
//   - column: The 1-indexed column number
//   - source: The full source code for generating context
//
// Returns:
//   - *CompileError: A new parser error with context
//
// Requirement 5.3: Parser includes source code context (2 lines before and after error).
// Requirement 5.4: Parser includes pointer (^) indicating error column.
func NewParserErrorWithContext(message string, line, column int, source string) *CompileError {
	return &CompileError{
		Phase:   "parser",
		Message: message,
		Line:    line,
		Column:  column,
		Context: GenerateErrorContext(source, line, column),
	}
}


// NewCompilerErrorWithContext creates a new CompileError for compiler phase errors with source context.
//
// Parameters:
//   - message: The error description
//   - line: The 1-indexed line number
//   - column: The 1-indexed column number
//   - source: The full source code for generating context
//
// Returns:
//   - *CompileError: A new compiler error with context
func NewCompilerErrorWithContext(message string, line, column int, source string) *CompileError {
	return &CompileError{
		Phase:   "compiler",
		Message: message,
		Line:    line,
		Column:  column,
		Context: GenerateErrorContext(source, line, column),
	}
}

// GenerateErrorContext generates source code context around an error location.
// It includes 2 lines before and 2 lines after the error line, with line numbers
// and a pointer (^) indicating the error column.
//
// Parameters:
//   - source: The full source code
//   - line: The 1-indexed line number of the error
//   - column: The 1-indexed column number of the error
//
// Returns:
//   - string: Formatted context string with line numbers and error pointer
//
// Requirement 5.3: Parser includes source code context (2 lines before and after error).
// Requirement 5.4: Parser includes pointer (^) indicating error column.
//
// Example output:
//
//	  2 | int x = 5;
//	  3 | int y = 10;
//	> 4 | int z = ;
//	    |         ^
//	  5 | int w = 20;
//	  6 | int v = 30;
func GenerateErrorContext(source string, line, column int) string {
	if source == "" || line <= 0 {
		return ""
	}

	lines := strings.Split(source, "\n")
	if line > len(lines) {
		return ""
	}

	// Calculate the range of lines to show (2 before and 2 after)
	start := line - 3 // 2 lines before (0-indexed: line-1-2 = line-3)
	if start < 0 {
		start = 0
	}
	end := line + 2 // 2 lines after (0-indexed: line-1+2+1 = line+2)
	if end > len(lines) {
		end = len(lines)
	}

	var buf strings.Builder

	// Calculate the width needed for line numbers
	maxLineNum := end
	lineNumWidth := len(fmt.Sprintf("%d", maxLineNum))

	for i := start; i < end; i++ {
		lineNum := i + 1 // Convert to 1-indexed
		lineContent := lines[i]

		if lineNum == line {
			// Error line - mark with >
			buf.WriteString(fmt.Sprintf("> %*d | %s\n", lineNumWidth, lineNum, lineContent))
			// Add pointer line
			// Calculate spaces: "> " + lineNumWidth + " | " + (column-1) spaces + "^"
			pointerIndent := 2 + lineNumWidth + 3 // "> " + lineNumWidth + " | "
			if column > 0 {
				buf.WriteString(fmt.Sprintf("%s%s^\n", strings.Repeat(" ", pointerIndent), strings.Repeat(" ", column-1)))
			} else {
				buf.WriteString(fmt.Sprintf("%s^\n", strings.Repeat(" ", pointerIndent)))
			}
		} else {
			// Context line
			buf.WriteString(fmt.Sprintf("  %*d | %s\n", lineNumWidth, lineNum, lineContent))
		}
	}

	return buf.String()
}



