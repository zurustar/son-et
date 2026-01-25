// Package lexer provides lexical analysis for FILLY scripts (.TFY files).
// It tokenizes source code into a sequence of tokens for the parser.
package lexer

import (
	"fmt"
)

// Lexer performs lexical analysis on FILLY source code.
type Lexer struct {
	input        string // source code
	position     int    // current position in input
	readPosition int    // next position to read
	ch           byte   // current character
	line         int    // current line number (1-indexed)
	column       int    // current column number (1-indexed)
}

// New creates a new Lexer for the given input source code.
func New(input string) *Lexer {
	l := &Lexer{
		input:  input,
		line:   1,
		column: 0,
	}
	l.readChar()
	return l
}

// readChar reads the next character and advances the position.
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++

	if l.ch == '\n' {
		l.line++
		l.column = 0
	} else {
		l.column++
	}
}

// peekChar returns the next character without advancing the position.
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// skipSingleLineComment skips a single-line comment starting with //.
// Requirement 2.9: Single-line comments (//) skip until next line.
func (l *Lexer) skipSingleLineComment() {
	// Skip the // characters
	l.readChar() // skip first /
	l.readChar() // skip second /

	// Skip until end of line or end of file
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
}

// skipMultiLineComment skips a multi-line comment starting with /*.
// Requirement 2.10: Multi-line comments (/* */) skip until closing */.
func (l *Lexer) skipMultiLineComment() {
	// Skip the /* characters
	l.readChar() // skip /
	l.readChar() // skip *

	// Skip until */ or end of file
	for {
		if l.ch == 0 {
			// End of file reached without closing comment
			break
		}
		if l.ch == '*' && l.peekChar() == '/' {
			// Found closing */
			l.readChar() // skip *
			l.readChar() // skip /
			break
		}
		l.readChar()
	}
}

// skipWhitespaceAndComments skips whitespace and comments.
// This combines whitespace skipping with comment skipping for cleaner token parsing.
// Requirement 2.9: Single-line comments (//) skip until next line.
// Requirement 2.10: Multi-line comments (/* */) skip until closing */.
// Requirement 2.11: Lexer skips whitespace without creating tokens.
func (l *Lexer) skipWhitespaceAndComments() {
	for {
		// Skip whitespace
		for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
			l.readChar()
		}

		// Check for comments
		if l.ch == '/' {
			if l.peekChar() == '/' {
				// Single-line comment
				l.skipSingleLineComment()
				continue
			} else if l.peekChar() == '*' {
				// Multi-line comment
				l.skipMultiLineComment()
				continue
			}
		}

		// No more whitespace or comments to skip
		break
	}
}

// isLetter returns true if the character is a letter or underscore.
// Used for identifying the start of identifiers.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_'
}

// isDigit returns true if the character is a digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isHexDigit returns true if the character is a hexadecimal digit.
func isHexDigit(ch byte) bool {
	return isDigit(ch) || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')
}

// readIdentifier reads an identifier or keyword from the input.
// Identifiers can contain letters, digits, and underscores, but must start with a letter or underscore.
// Requirement 2.2: Keywords are identified case-insensitively.
// Requirement 2.3: Identifiers are returned as IDENT tokens.
func (l *Lexer) readIdentifier() string {
	startPos := l.position
	for isLetter(l.ch) || isDigit(l.ch) {
		l.readChar()
	}
	return l.input[startPos:l.position]
}

// readNumber reads a numeric literal (decimal, hexadecimal, or floating point).
// Requirement 2.4: Integer literals (decimal or 0x-prefixed hexadecimal) create INT tokens.
// Requirement 2.5: Floating point literals create FLOAT tokens.
func (l *Lexer) readNumber() Token {
	startLine := l.line
	startColumn := l.column

	// Check for hexadecimal (0x or 0X prefix)
	if l.ch == '0' && (l.peekChar() == 'x' || l.peekChar() == 'X') {
		return l.readHexNumber(startLine, startColumn)
	}

	// Read decimal number (may include floating point)
	return l.readDecimalNumber(startLine, startColumn)
}

// readHexNumber reads a hexadecimal integer literal (0x or 0X prefix).
// Requirement 2.4: Hexadecimal integers with 0x/0X prefix create INT tokens.
func (l *Lexer) readHexNumber(startLine, startColumn int) Token {
	startPos := l.position

	// Skip '0'
	l.readChar()
	// Skip 'x' or 'X'
	l.readChar()

	// Read hex digits
	for isHexDigit(l.ch) {
		l.readChar()
	}

	literal := l.input[startPos:l.position]

	return Token{
		Type:    TOKEN_INT,
		Literal: literal,
		Line:    startLine,
		Column:  startColumn,
	}
}

// readDecimalNumber reads a decimal integer or floating point literal.
// Requirement 2.4: Decimal integers create INT tokens.
// Requirement 2.5: Floating point literals create FLOAT tokens.
func (l *Lexer) readDecimalNumber(startLine, startColumn int) Token {
	startPos := l.position
	tokenType := TOKEN_INT

	// Read integer part
	for isDigit(l.ch) {
		l.readChar()
	}

	// Check for decimal point (floating point number)
	if l.ch == '.' && isDigit(l.peekChar()) {
		tokenType = TOKEN_FLOAT
		// Skip '.'
		l.readChar()
		// Read fractional part
		for isDigit(l.ch) {
			l.readChar()
		}
	}

	literal := l.input[startPos:l.position]

	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    startLine,
		Column:  startColumn,
	}
}

// readDirective reads a preprocessor directive (#info, #include).
// Requirement 2.15: Preprocessor directives create DIRECTIVE tokens.
// Requirement 2.16: #info directive reads until end of line.
// Requirement 2.17: #include directive extracts the filename.
func (l *Lexer) readDirective() Token {
	startLine := l.line
	startColumn := l.column
	startPos := l.position

	// Skip the '#' character
	l.readChar()

	// Read the directive name
	directiveName := l.readIdentifier()

	// Skip whitespace after directive name
	for l.ch == ' ' || l.ch == '\t' {
		l.readChar()
	}

	// Read the rest of the line as the directive content
	contentStart := l.position
	for l.ch != '\n' && l.ch != '\r' && l.ch != 0 {
		l.readChar()
	}
	content := l.input[contentStart:l.position]

	// Combine directive name and content
	literal := l.input[startPos:l.position]

	// Determine token type based on directive name
	var tokenType TokenType
	switch directiveName {
	case "info":
		tokenType = TOKEN_INFO
	case "include":
		tokenType = TOKEN_INCLUDE
		// Extract filename from content (remove quotes if present)
		content = extractFilename(content)
		literal = content
	case "define":
		tokenType = TOKEN_DEFINE
	default:
		tokenType = TOKEN_DIRECTIVE
	}

	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    startLine,
		Column:  startColumn,
	}
}

// extractFilename extracts the filename from an #include directive content.
// It handles both "filename" and <filename> formats.
// It also handles trailing comments after the filename.
func extractFilename(content string) string {
	content = trimSpace(content)
	if len(content) < 2 {
		return content
	}

	// Handle "filename" format
	if content[0] == '"' {
		// Find the closing quote
		for i := 1; i < len(content); i++ {
			if content[i] == '"' {
				return content[1:i]
			}
		}
		// No closing quote found, return as-is
		return content
	}

	// Handle <filename> format
	if content[0] == '<' {
		// Find the closing bracket
		for i := 1; i < len(content); i++ {
			if content[i] == '>' {
				return content[1:i]
			}
		}
		// No closing bracket found, return as-is
		return content
	}

	return content
}

// trimSpace removes leading and trailing whitespace from a string.
func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t') {
		end--
	}
	return s[start:end]
}

// readString reads a string literal enclosed in double quotes.
// Requirement 2.6: String literals enclosed in double quotes create STRING tokens.
// Handles escape sequences: \n, \t, \\, \"
func (l *Lexer) readString() Token {
	startLine := l.line
	startColumn := l.column

	// Skip the opening double quote
	l.readChar()

	var result []byte
	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			// Handle escape sequences
			l.readChar()
			switch l.ch {
			case 'n':
				result = append(result, '\n')
			case 't':
				result = append(result, '\t')
			case '\\':
				result = append(result, '\\')
			case '"':
				result = append(result, '"')
			case 'r':
				result = append(result, '\r')
			case '0':
				result = append(result, 0)
			default:
				// Unknown escape sequence - keep the backslash and character
				result = append(result, '\\')
				result = append(result, l.ch)
			}
		} else {
			result = append(result, l.ch)
		}
		l.readChar()
	}

	// Skip the closing double quote (if present)
	if l.ch == '"' {
		l.readChar()
	}

	return Token{
		Type:    TOKEN_STRING,
		Literal: string(result),
		Line:    startLine,
		Column:  startColumn,
	}
}

// newToken creates a new token with the given type and literal.
func (l *Lexer) newToken(tokenType TokenType, literal string) Token {
	return Token{
		Type:    tokenType,
		Literal: literal,
		Line:    l.line,
		Column:  l.column,
	}
}

// NextToken returns the next token from the input.
// This method implements the core lexical analysis logic.
// Requirement 2.14: All tokens include line and column numbers for error reporting.
func (l *Lexer) NextToken() Token {
	var tok Token

	// Skip whitespace and comments (Requirements 2.9, 2.10, 2.11)
	l.skipWhitespaceAndComments()

	// Record position for token
	tok.Line = l.line
	tok.Column = l.column

	switch l.ch {
	case 0:
		// End of file (Requirement 2.13)
		tok.Type = TOKEN_EOF
		tok.Literal = ""

	// String literal (Requirement 2.6)
	case '"':
		return l.readString()

	// Arithmetic operators (Requirement 2.7)
	case '+':
		tok = l.newToken(TOKEN_PLUS, "+")
	case '-':
		tok = l.newToken(TOKEN_MINUS, "-")
	case '*':
		tok = l.newToken(TOKEN_ASTERISK, "*")
	case '/':
		// Note: Comments are already handled by skipWhitespaceAndComments,
		// so if we reach here, it's a division operator
		tok = l.newToken(TOKEN_SLASH, "/")
	case '%':
		tok = l.newToken(TOKEN_PERCENT, "%")

	// Assignment and comparison operators (Requirement 2.7)
	case '=':
		if l.peekChar() == '=' {
			// == (equality)
			tok.Type = TOKEN_EQ
			tok.Literal = "=="
			l.readChar() // consume first '='
		} else {
			// = (assignment)
			tok = l.newToken(TOKEN_ASSIGN, "=")
		}
	case '!':
		if l.peekChar() == '=' {
			// != (not equal)
			tok.Type = TOKEN_NEQ
			tok.Literal = "!="
			l.readChar() // consume '!'
		} else {
			// ! (logical not)
			tok = l.newToken(TOKEN_NOT, "!")
		}
	case '<':
		if l.peekChar() == '=' {
			// <= (less than or equal)
			tok.Type = TOKEN_LTE
			tok.Literal = "<="
			l.readChar() // consume '<'
		} else {
			// < (less than)
			tok = l.newToken(TOKEN_LT, "<")
		}
	case '>':
		if l.peekChar() == '=' {
			// >= (greater than or equal)
			tok.Type = TOKEN_GTE
			tok.Literal = ">="
			l.readChar() // consume '>'
		} else {
			// > (greater than)
			tok = l.newToken(TOKEN_GT, ">")
		}

	// Logical operators (Requirement 2.7)
	case '&':
		if l.peekChar() == '&' {
			// && (logical and)
			tok.Type = TOKEN_AND
			tok.Literal = "&&"
			l.readChar() // consume first '&'
		} else {
			// Single '&' is not a valid operator in FILLY
			tok = l.newToken(TOKEN_ILLEGAL, "&")
		}
	case '|':
		if l.peekChar() == '|' {
			// || (logical or)
			tok.Type = TOKEN_OR
			tok.Literal = "||"
			l.readChar() // consume first '|'
		} else {
			// Single '|' is not a valid operator in FILLY
			tok = l.newToken(TOKEN_ILLEGAL, "|")
		}

	// Delimiters (Requirement 2.8)
	case '(':
		tok = l.newToken(TOKEN_LPAREN, "(")
	case ')':
		tok = l.newToken(TOKEN_RPAREN, ")")
	case '{':
		tok = l.newToken(TOKEN_LBRACE, "{")
	case '}':
		tok = l.newToken(TOKEN_RBRACE, "}")
	case '[':
		tok = l.newToken(TOKEN_LBRACKET, "[")
	case ']':
		tok = l.newToken(TOKEN_RBRACKET, "]")
	case ',':
		tok = l.newToken(TOKEN_COMMA, ",")
	case ';':
		tok = l.newToken(TOKEN_SEMICOLON, ";")
	case ':':
		tok = l.newToken(TOKEN_COLON, ":")

	default:
		if l.ch == '#' {
			// Preprocessor directive (Requirement 2.15, 2.16, 2.17)
			return l.readDirective()
		} else if isLetter(l.ch) {
			// Read identifier or keyword
			tok.Literal = l.readIdentifier()
			tok.Type = LookupIdent(tok.Literal)
			// Don't call readChar() here because readIdentifier() already advanced past the identifier
			return tok
		} else if isDigit(l.ch) {
			// Read numeric literal (decimal, hexadecimal, or floating point)
			// Requirement 2.4: Integer literals create INT tokens
			// Requirement 2.5: Floating point literals create FLOAT tokens
			return l.readNumber()
		} else {
			// Unknown character - return ILLEGAL token (Requirement 2.12)
			tok.Type = TOKEN_ILLEGAL
			tok.Literal = string(l.ch)
			l.readChar()
			return tok
		}
	}

	l.readChar()
	return tok
}

// LexerError represents an error that occurred during lexical analysis.
// It includes location information for error reporting.
type LexerError struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface.
func (e *LexerError) Error() string {
	return fmt.Sprintf("lexer error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// NewLexerError creates a new LexerError with the given message and location.
func NewLexerError(message string, line, column int) *LexerError {
	return &LexerError{
		Message: message,
		Line:    line,
		Column:  column,
	}
}

// Tokenize returns all tokens from the input.
// It repeatedly calls NextToken until EOF is reached.
// Note: This method does not return errors for ILLEGAL tokens.
// Use TokenizeWithErrors to get detailed error information.
func (l *Lexer) Tokenize() ([]Token, error) {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TOKEN_EOF {
			break
		}
	}
	return tokens, nil
}

// TokenizeWithErrors returns all tokens and all errors from the input.
// Unlike Tokenize, this method collects all ILLEGAL token errors.
// Requirement 5.1: Lexer reports illegal characters with character, line, and column.
// Requirement 5.6: System collects all errors and returns them to caller.
func (l *Lexer) TokenizeWithErrors() ([]Token, []*LexerError) {
	var tokens []Token
	var errors []*LexerError

	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)

		// Collect errors for ILLEGAL tokens
		if tok.Type == TOKEN_ILLEGAL {
			errors = append(errors, NewLexerError(
				fmt.Sprintf("illegal character '%s'", tok.Literal),
				tok.Line,
				tok.Column,
			))
		}

		if tok.Type == TOKEN_EOF {
			break
		}
	}

	return tokens, errors
}
