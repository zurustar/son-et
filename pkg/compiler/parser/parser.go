// Package parser provides syntax analysis for FILLY scripts (.TFY files).
// It transforms a sequence of tokens into an Abstract Syntax Tree (AST).
package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
)

// Operator precedence levels
const (
	_ int = iota
	LOWEST
	OR          // ||
	AND         // &&
	EQUALS      // ==, !=
	LESSGREATER // <, >, <=, >=
	SUM         // +, -
	PRODUCT     // *, /, %
	PREFIX      // -x, !x
	CALL        // func(x)
	INDEX       // array[index]
)

// precedences maps token types to their precedence levels
var precedences = map[lexer.TokenType]int{
	lexer.TOKEN_OR:       OR,
	lexer.TOKEN_AND:      AND,
	lexer.TOKEN_ASSIGN:   EQUALS, // = used as comparison in conditions (C-style)
	lexer.TOKEN_EQ:       EQUALS,
	lexer.TOKEN_NEQ:      EQUALS,
	lexer.TOKEN_LT:       LESSGREATER,
	lexer.TOKEN_GT:       LESSGREATER,
	lexer.TOKEN_LTE:      LESSGREATER,
	lexer.TOKEN_GTE:      LESSGREATER,
	lexer.TOKEN_PLUS:     SUM,
	lexer.TOKEN_MINUS:    SUM,
	lexer.TOKEN_ASTERISK: PRODUCT,
	lexer.TOKEN_SLASH:    PRODUCT,
	lexer.TOKEN_PERCENT:  PRODUCT,
	lexer.TOKEN_LPAREN:   CALL,
	lexer.TOKEN_LBRACKET: INDEX,
}

// Parse function types for Pratt parser
type (
	prefixParseFn func() Expression
	infixParseFn  func(Expression) Expression
)

// ParserError represents an error that occurred during parsing.
// It includes location information and source context for error reporting.
type ParserError struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface.
func (e *ParserError) Error() string {
	return fmt.Sprintf("parser error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

// NewParserError creates a new ParserError with the given message and location.
func NewParserError(message string, line, column int) *ParserError {
	return &ParserError{
		Message: message,
		Line:    line,
		Column:  column,
	}
}

// Parser performs syntax analysis on FILLY tokens.
type Parser struct {
	lexer  *lexer.Lexer
	tokens []lexer.Token
	pos    int
	errors []*ParserError
	source string // original source code for error context

	// Pratt parser function maps
	prefixParseFns map[lexer.TokenType]prefixParseFn
	infixParseFns  map[lexer.TokenType]infixParseFn
}

// New creates a new Parser for the given Lexer.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		lexer:          l,
		errors:         []*ParserError{},
		prefixParseFns: make(map[lexer.TokenType]prefixParseFn),
		infixParseFns:  make(map[lexer.TokenType]infixParseFn),
	}

	// Tokenize all input
	tokens, _ := l.Tokenize()
	p.tokens = tokens

	// Register prefix parse functions
	p.registerPrefix(lexer.TOKEN_IDENT, p.parseIdentifier)
	p.registerPrefix(lexer.TOKEN_INT, p.parseIntegerLiteral)
	p.registerPrefix(lexer.TOKEN_FLOAT, p.parseFloatLiteral)
	p.registerPrefix(lexer.TOKEN_STRING, p.parseStringLiteral)
	p.registerPrefix(lexer.TOKEN_MINUS, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_NOT, p.parsePrefixExpression)
	p.registerPrefix(lexer.TOKEN_LPAREN, p.parseGroupedExpression)

	// Register infix parse functions
	p.registerInfix(lexer.TOKEN_PLUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_MINUS, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_ASTERISK, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_SLASH, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_PERCENT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_ASSIGN, p.parseInfixExpression) // = as comparison (C-style)
	p.registerInfix(lexer.TOKEN_EQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_NEQ, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GT, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LTE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_GTE, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_AND, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_OR, p.parseInfixExpression)
	p.registerInfix(lexer.TOKEN_LPAREN, p.parseCallExpression)
	p.registerInfix(lexer.TOKEN_LBRACKET, p.parseIndexExpression)

	return p
}

// registerPrefix registers a prefix parse function for a token type.
func (p *Parser) registerPrefix(tokenType lexer.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

// registerInfix registers an infix parse function for a token type.
func (p *Parser) registerInfix(tokenType lexer.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

// Errors returns the list of parsing errors.
func (p *Parser) Errors() []*ParserError {
	return p.errors
}

// curToken returns the current token.
func (p *Parser) curToken() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos]
}

// peekToken returns the next token without advancing.
func (p *Parser) peekToken() lexer.Token {
	if p.pos+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.pos+1]
}

// nextToken advances to the next token.
func (p *Parser) nextToken() {
	p.pos++
}

// curTokenIs checks if the current token is of the given type.
func (p *Parser) curTokenIs(t lexer.TokenType) bool {
	return p.curToken().Type == t
}

// peekTokenIs checks if the next token is of the given type.
func (p *Parser) peekTokenIs(t lexer.TokenType) bool {
	return p.peekToken().Type == t
}

// expectPeek advances if the next token is of the expected type.
// Returns true if successful, false otherwise (and adds an error).
func (p *Parser) expectPeek(t lexer.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

// curPrecedence returns the precedence of the current token.
func (p *Parser) curPrecedence() int {
	if prec, ok := precedences[p.curToken().Type]; ok {
		return prec
	}
	return LOWEST
}

// peekPrecedence returns the precedence of the next token.
func (p *Parser) peekPrecedence() int {
	if prec, ok := precedences[p.peekToken().Type]; ok {
		return prec
	}
	return LOWEST
}

// ============================================================================
// Error handling methods
// ============================================================================

// addError adds an error message to the parser's error list with location information.
// Requirement 5.2: Parser reports syntax errors with expected/actual token types, line, and column.
func (p *Parser) addError(msg string, line, column int) {
	p.errors = append(p.errors, NewParserError(msg, line, column))
}

// addErrorAtCurrent adds an error at the current token's location.
func (p *Parser) addErrorAtCurrent(msg string) {
	tok := p.curToken()
	p.addError(msg, tok.Line, tok.Column)
}

// peekError adds an error for unexpected token type.
// Requirement 5.2: Parser reports syntax errors with expected/actual token types, line, and column.
func (p *Parser) peekError(t lexer.TokenType) {
	tok := p.peekToken()
	msg := fmt.Sprintf("expected %s, got %s", t.String(), tok.Type.String())
	p.addError(msg, tok.Line, tok.Column)
}

// noPrefixParseFnError adds an error for missing prefix parse function.
func (p *Parser) noPrefixParseFnError(t lexer.TokenType) {
	tok := p.curToken()
	msg := fmt.Sprintf("no prefix parse function for %s found", t.String())
	p.addError(msg, tok.Line, tok.Column)
}

// ============================================================================
// Main parsing methods
// ============================================================================

// ParseProgram parses the entire program and returns the AST.
// Requirement 3.1: Parser builds an AST representing program structure.
// Requirement 5.6: System collects all errors and returns them to caller.
func (p *Parser) ParseProgram() (*Program, []error) {
	program := &Program{
		Statements: []Statement{},
	}

	for !p.curTokenIs(lexer.TOKEN_EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	// Convert ParserError to error interface
	var errs []error
	for _, e := range p.errors {
		errs = append(errs, e)
	}

	return program, errs
}

// parseStatement parses a single statement.
func (p *Parser) parseStatement() Statement {
	// Skip empty statements (just semicolons)
	for p.curTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
		if p.curTokenIs(lexer.TOKEN_EOF) || p.curTokenIs(lexer.TOKEN_RBRACE) {
			return nil
		}
	}

	switch p.curToken().Type {
	case lexer.TOKEN_INFO:
		return p.parseInfoDirective()
	case lexer.TOKEN_INCLUDE:
		return p.parseIncludeDirective()
	case lexer.TOKEN_DEFINE:
		return p.parseDefineDirective()
	case lexer.TOKEN_DIRECTIVE:
		// Unknown directive - skip it
		return p.parseGenericDirective()
	case lexer.TOKEN_INT_TYPE, lexer.TOKEN_STR_TYPE, lexer.TOKEN_REAL_TYPE:
		return p.parseVarDeclaration()
	case lexer.TOKEN_IF:
		return p.parseIfStatement()
	case lexer.TOKEN_FOR:
		return p.parseForStatement()
	case lexer.TOKEN_WHILE:
		return p.parseWhileStatement()
	case lexer.TOKEN_SWITCH:
		return p.parseSwitchStatement()
	case lexer.TOKEN_MES:
		return p.parseMesStatement()
	case lexer.TOKEN_STEP:
		return p.parseStepStatement()
	case lexer.TOKEN_BREAK:
		return p.parseBreakStatement()
	case lexer.TOKEN_CONTINUE:
		return p.parseContinueStatement()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStatement()
	case lexer.TOKEN_LBRACE:
		return p.parseBlockStatement()
	case lexer.TOKEN_IDENT:
		return p.parseIdentifierStatement()
	// Requirement 9.5: del_me, del_us, del_all are treated as function calls
	case lexer.TOKEN_DEL_ME, lexer.TOKEN_DEL_US, lexer.TOKEN_DEL_ALL:
		return p.parseSpecialCommandStatement()
	// Requirement 9.4: end_step is treated as step block end marker (also can be standalone)
	case lexer.TOKEN_END_STEP:
		return p.parseSpecialCommandStatement()
	default:
		return p.parseExpressionStatement()
	}
}

// parseIdentifierStatement parses statements starting with an identifier.
// This could be a function definition, assignment, label, or expression statement.
// Requirement 9.9: Function definitions without 'function' keyword.
func (p *Parser) parseIdentifierStatement() Statement {
	// Check if this is a label: NAME:
	if p.peekTokenIs(lexer.TOKEN_COLON) {
		return p.parseLabelStatement()
	}

	// Check if this is a function definition: name(params){body}
	if p.isFunctionDefinition() {
		return p.parseFunctionDefinition()
	}

	// Otherwise, parse as expression (which may become assignment)
	return p.parseExpressionOrAssignment()
}

// isFunctionDefinition checks if the current position is a function definition.
// Function definition: identifier followed by ( params ) {
func (p *Parser) isFunctionDefinition() bool {
	if !p.curTokenIs(lexer.TOKEN_IDENT) {
		return false
	}
	if !p.peekTokenIs(lexer.TOKEN_LPAREN) {
		return false
	}

	// Save current position
	savedPos := p.pos

	// Skip identifier and (
	p.nextToken() // now at (
	p.nextToken() // now inside params or at )

	// Skip to matching )
	parenDepth := 1
	for parenDepth > 0 && !p.curTokenIs(lexer.TOKEN_EOF) {
		if p.curTokenIs(lexer.TOKEN_LPAREN) {
			parenDepth++
		} else if p.curTokenIs(lexer.TOKEN_RPAREN) {
			parenDepth--
		}
		if parenDepth > 0 {
			p.nextToken()
		}
	}

	// Check if next token is {
	isFuncDef := p.peekTokenIs(lexer.TOKEN_LBRACE)

	// Restore position
	p.pos = savedPos

	return isFuncDef
}

// parseExpressionOrAssignment parses an expression that may be an assignment.
func (p *Parser) parseExpressionOrAssignment() Statement {
	// Check if this is a simple assignment: identifier = value
	// We need to handle this before parseExpression because = is also registered
	// as an infix operator for C-style comparisons in conditions.
	if p.curTokenIs(lexer.TOKEN_IDENT) && p.peekTokenIs(lexer.TOKEN_ASSIGN) {
		ident := &Identifier{
			Token: p.curToken(),
			Value: p.curToken().Literal,
		}
		p.nextToken() // move to =
		tok := p.curToken()
		p.nextToken() // move past =

		value := p.parseExpression(LOWEST)

		stmt := &AssignStatement{
			Token: tok,
			Name:  ident,
			Value: value,
		}

		// Skip optional semicolon
		if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}

		return stmt
	}

	// Check if this is an array assignment: arr[index] = value
	// We need to look ahead to find the = after the array access
	if p.curTokenIs(lexer.TOKEN_IDENT) && p.peekTokenIs(lexer.TOKEN_LBRACKET) {
		// Save position to check for assignment
		savedPos := p.pos

		// Parse the array access part manually to check for assignment
		p.nextToken() // move to [

		// Skip to matching ]
		bracketDepth := 1
		p.nextToken() // move past [
		for bracketDepth > 0 && !p.curTokenIs(lexer.TOKEN_EOF) {
			if p.curTokenIs(lexer.TOKEN_LBRACKET) {
				bracketDepth++
			} else if p.curTokenIs(lexer.TOKEN_RBRACKET) {
				bracketDepth--
			}
			if bracketDepth > 0 {
				p.nextToken()
			}
		}

		// Check if next token is =
		isArrayAssignment := p.peekTokenIs(lexer.TOKEN_ASSIGN)

		// Restore position
		p.pos = savedPos

		if isArrayAssignment {
			// Parse the array access expression without = being treated as infix
			arrayExpr := p.parseArrayAccessOnly()

			p.nextToken() // move to =
			tok := p.curToken()
			p.nextToken() // move past =

			value := p.parseExpression(LOWEST)

			stmt := &AssignStatement{
				Token: tok,
				Name:  arrayExpr,
				Value: value,
			}

			// Skip optional semicolon
			if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
				p.nextToken()
			}

			return stmt
		}
	}

	// Parse as general expression
	expr := p.parseExpression(LOWEST)

	// Check for assignment (for other expression types)
	if p.peekTokenIs(lexer.TOKEN_ASSIGN) {
		p.nextToken() // move to =
		tok := p.curToken()
		p.nextToken() // move past =

		value := p.parseExpression(LOWEST)

		stmt := &AssignStatement{
			Token: tok,
			Name:  expr,
			Value: value,
		}

		// Skip optional semicolon
		if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}

		return stmt
	}

	// Otherwise it's an expression statement
	stmt := &ExpressionStatement{
		Token:      p.curToken(),
		Expression: expr,
	}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseArrayAccessOnly parses an array access expression (arr[index]) without
// treating = as an infix operator. This is used for array assignment targets.
func (p *Parser) parseArrayAccessOnly() Expression {
	ident := &Identifier{
		Token: p.curToken(),
		Value: p.curToken().Literal,
	}

	if !p.peekTokenIs(lexer.TOKEN_LBRACKET) {
		return ident
	}

	p.nextToken() // move to [
	p.nextToken() // move past [

	index := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return &IndexExpression{
		Token: ident.Token,
		Left:  ident,
		Index: index,
	}
}

// parseBlockStatement parses a block of statements enclosed in braces.
func (p *Parser) parseBlockStatement() *BlockStatement {
	block := &BlockStatement{
		Token:      p.curToken(),
		Statements: []Statement{},
	}

	p.nextToken() // skip {

	for !p.curTokenIs(lexer.TOKEN_RBRACE) && !p.curTokenIs(lexer.TOKEN_EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

// ============================================================================
// Expression parsing (Pratt parser)
// ============================================================================

// parseExpression parses an expression with the given precedence.
func (p *Parser) parseExpression(precedence int) Expression {
	prefix := p.prefixParseFns[p.curToken().Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken().Type)
		return nil
	}

	leftExp := prefix()

	for !p.peekTokenIs(lexer.TOKEN_SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken().Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

// parseIdentifier parses an identifier expression.
func (p *Parser) parseIdentifier() Expression {
	return &Identifier{
		Token: p.curToken(),
		Value: p.curToken().Literal,
	}
}

// parseIntegerLiteral parses an integer literal expression.
// Requirement 2.4: Integer literals (decimal or 0x-prefixed hexadecimal).
func (p *Parser) parseIntegerLiteral() Expression {
	lit := &IntegerLiteral{Token: p.curToken()}

	literal := p.curToken().Literal
	var value int64
	var err error

	// Check for hexadecimal
	if strings.HasPrefix(strings.ToLower(literal), "0x") {
		value, err = strconv.ParseInt(literal[2:], 16, 64)
	} else {
		value, err = strconv.ParseInt(literal, 10, 64)
	}

	if err != nil {
		tok := p.curToken()
		msg := fmt.Sprintf("could not parse %q as integer", literal)
		p.addError(msg, tok.Line, tok.Column)
		return nil
	}

	lit.Value = value
	return lit
}

// parseFloatLiteral parses a floating point literal expression.
// Requirement 2.5: Floating point literals.
func (p *Parser) parseFloatLiteral() Expression {
	lit := &FloatLiteral{Token: p.curToken()}

	value, err := strconv.ParseFloat(p.curToken().Literal, 64)
	if err != nil {
		tok := p.curToken()
		msg := fmt.Sprintf("could not parse %q as float", tok.Literal)
		p.addError(msg, tok.Line, tok.Column)
		return nil
	}

	lit.Value = value
	return lit
}

// parseStringLiteral parses a string literal expression.
// Requirement 2.6: String literals enclosed in double quotes.
func (p *Parser) parseStringLiteral() Expression {
	return &StringLiteral{
		Token: p.curToken(),
		Value: p.curToken().Literal,
	}
}

// parsePrefixExpression parses a prefix expression (-x, !x).
func (p *Parser) parsePrefixExpression() Expression {
	expression := &UnaryExpression{
		Token:    p.curToken(),
		Operator: p.curToken().Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

// parseGroupedExpression parses a grouped expression (parentheses).
func (p *Parser) parseGroupedExpression() Expression {
	p.nextToken() // skip (

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return exp
}

// parseInfixExpression parses an infix expression (a + b, x == y).
// Requirement 3.14: Expressions respect operator precedence.
func (p *Parser) parseInfixExpression(left Expression) Expression {
	expression := &BinaryExpression{
		Token:    p.curToken(),
		Operator: p.curToken().Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

// parseCallExpression parses a function call expression.
// Requirement 3.13: Function calls with function name and arguments.
func (p *Parser) parseCallExpression(function Expression) Expression {
	// Get function name from identifier
	ident, ok := function.(*Identifier)
	if !ok {
		tok := p.curToken()
		msg := "expected identifier for function call"
		p.addError(msg, tok.Line, tok.Column)
		return nil
	}

	exp := &CallExpression{
		Token:    p.curToken(),
		Function: ident.Value,
	}

	exp.Arguments = p.parseExpressionList(lexer.TOKEN_RPAREN)

	return exp
}

// parseIndexExpression parses an array index expression.
// Requirement 3.6: Array access with IndexExpression.
func (p *Parser) parseIndexExpression(left Expression) Expression {
	exp := &IndexExpression{
		Token: p.curToken(),
		Left:  left,
	}

	p.nextToken() // skip [
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RBRACKET) {
		return nil
	}

	return exp
}

// parseExpressionList parses a comma-separated list of expressions.
// Requirement 3.19: Array references (arr[]) in function call arguments.
func (p *Parser) parseExpressionList(end lexer.TokenType) []Expression {
	list := []Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpressionOrArrayRef())

	for p.peekTokenIs(lexer.TOKEN_COMMA) {
		p.nextToken() // skip to comma
		p.nextToken() // skip comma
		list = append(list, p.parseExpressionOrArrayRef())
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// parseExpressionOrArrayRef parses an expression that may be an array reference (arr[]).
// Requirement 3.19: Array references (arr[]) in function call arguments.
func (p *Parser) parseExpressionOrArrayRef() Expression {
	// Check for array reference pattern: identifier followed by []
	if p.curTokenIs(lexer.TOKEN_IDENT) && p.peekTokenIs(lexer.TOKEN_LBRACKET) {
		// Look ahead to see if it's arr[] (empty brackets = array reference)
		// or arr[expr] (index expression)
		savedPos := p.pos
		name := p.curToken().Literal
		tok := p.curToken()

		p.nextToken() // move to [
		if p.peekTokenIs(lexer.TOKEN_RBRACKET) {
			// This is an array reference: arr[]
			p.nextToken() // move to ]
			return &ArrayReference{
				Token: tok,
				Name:  name,
			}
		}

		// Restore position and parse as normal expression
		p.pos = savedPos
	}

	return p.parseExpression(LOWEST)
}

// ============================================================================
// Statement parsing (placeholders for subsequent tasks)
// ============================================================================

// parseVarDeclaration parses a variable declaration.
// Requirement 3.2: Variable declarations (int x, y[]; str s;) create VarDeclaration nodes
// Requirement 3.3: Array declarations (int arr[10]) with array flag and size expression
//
// Supported patterns:
//   - int x;                           - single variable
//   - str s;                           - string variable
//   - int x, y, z;                     - multiple variables
//   - int arr[];                       - array without size
//   - int arr[10];                     - array with size
//   - int LPic[],BasePic,FieldPic[];   - mixed arrays and scalars
//
// Also handles function definitions with return type:
//   - int funcName(params) { body }    - function with return type
func (p *Parser) parseVarDeclaration() Statement {
	returnType := p.curToken().Literal

	// Skip to first identifier
	if !p.expectPeek(lexer.TOKEN_IDENT) {
		return nil
	}

	// Check if this is a function definition (identifier followed by '(')
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		// This is a function definition with return type
		return p.parseFunctionDefinitionWithReturnType(returnType)
	}

	// Regular variable declaration
	stmt := &VarDeclaration{
		Token:   p.curToken(),
		Type:    returnType,
		Names:   []string{},
		IsArray: []bool{},
		Sizes:   []Expression{},
	}

	// Parse variable names
	for {
		name := p.curToken().Literal
		stmt.Names = append(stmt.Names, name)

		// Check for array syntax []
		isArray := false
		var size Expression
		if p.peekTokenIs(lexer.TOKEN_LBRACKET) {
			p.nextToken() // skip to [
			isArray = true
			if !p.peekTokenIs(lexer.TOKEN_RBRACKET) {
				p.nextToken()
				size = p.parseExpression(LOWEST)
			}
			if !p.expectPeek(lexer.TOKEN_RBRACKET) {
				return nil
			}
		}
		stmt.IsArray = append(stmt.IsArray, isArray)
		stmt.Sizes = append(stmt.Sizes, size)

		// Check for more variables
		if !p.peekTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken() // skip to comma
		if !p.expectPeek(lexer.TOKEN_IDENT) {
			return nil
		}
	}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseFunctionDefinitionWithReturnType parses a function definition with a return type.
// Example: int funcName(int x, str y) { body }
func (p *Parser) parseFunctionDefinitionWithReturnType(returnType string) Statement {
	stmt := &FunctionStatement{
		Token: p.curToken(),
		Name:  p.curToken().Literal,
	}
	// Note: returnType is currently not stored in FunctionStatement
	// It could be added later if needed

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseParameters()

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseFunctionDefinition parses a function definition.
// Requirement 3.4: Function definitions (name(params){body}) create FunctionStatement nodes.
// Requirement 9.9: Function definitions without 'function' keyword.
//
// Supported patterns:
//   - main() { ... }                    - no parameters
//   - myFunc(x) { ... }                 - single untyped parameter
//   - myFunc(int x) { ... }             - single typed parameter
//   - myFunc(int x, y) { ... }          - mixed typed and untyped
//   - myFunc(arr[]) { ... }             - array parameter
//   - myFunc(int arr[]) { ... }         - typed array parameter
//   - myFunc(x=10) { ... }              - parameter with default value
//   - myFunc(int time=1) { ... }        - typed parameter with default value
//   - OP_walk(c, p[], x, y, w, h, l=10) - complex mixed parameters
func (p *Parser) parseFunctionDefinition() Statement {
	stmt := &FunctionStatement{
		Token: p.curToken(),
		Name:  p.curToken().Literal,
	}

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseParameters()

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseParameters parses function parameters.
// Requirement 9.6: Parameters with default values (int time=1).
// Requirement 9.7: Array parameters (int arr[]).
// Requirement 9.10: Keywords as identifiers in expression context.
//
// Supported parameter patterns:
//   - x                    - untyped parameter
//   - int x                - typed parameter
//   - arr[]                - untyped array parameter
//   - int arr[]            - typed array parameter
//   - x=10                 - parameter with default value
//   - int time=1           - typed parameter with default value
//   - int x, y, z          - multiple parameters (type applies to first only)
//   - c, p[], x, y, l=10   - mixed parameters
func (p *Parser) parseParameters() []*Parameter {
	params := []*Parameter{}

	if p.peekTokenIs(lexer.TOKEN_RPAREN) {
		p.nextToken()
		return params
	}

	p.nextToken()

	for {
		param := p.parseParameter()
		if param != nil {
			params = append(params, param)
		}

		if !p.peekTokenIs(lexer.TOKEN_COMMA) {
			break
		}
		p.nextToken() // skip to comma
		p.nextToken() // skip comma
	}

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	return params
}

// parseParameter parses a single function parameter.
// Handles type prefix (int/str), array syntax ([]), and default values (=expr).
func (p *Parser) parseParameter() *Parameter {
	param := &Parameter{}

	// Check for type prefix (int or str)
	if p.curTokenIs(lexer.TOKEN_INT_TYPE) || p.curTokenIs(lexer.TOKEN_STR_TYPE) {
		param.Type = p.curToken().Literal
		p.nextToken()
	}

	// Parameter name
	if !p.curTokenIs(lexer.TOKEN_IDENT) {
		return nil
	}
	param.Name = p.curToken().Literal

	// Check for array syntax []
	if p.peekTokenIs(lexer.TOKEN_LBRACKET) {
		p.nextToken() // skip to [
		param.IsArray = true
		if !p.expectPeek(lexer.TOKEN_RBRACKET) {
			return nil
		}
	}

	// Check for default value
	if p.peekTokenIs(lexer.TOKEN_ASSIGN) {
		p.nextToken() // skip to =
		p.nextToken() // skip =
		param.DefaultValue = p.parseExpression(LOWEST)
	}

	return param
}

// parseIfStatement parses an if statement.
// Requirement 3.7: If statements with condition, consequence, and optional alternative.
// Supports both braced blocks and single statements without braces.
func (p *Parser) parseIfStatement() Statement {
	stmt := &IfStatement{Token: p.curToken()}

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Check if consequence has braces or is a single statement
	if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		stmt.Consequence = p.parseBlockStatement()
	} else {
		// Single statement without braces: if(cond) stmt;
		p.nextToken()
		singleStmt := p.parseStatement()
		if singleStmt != nil {
			stmt.Consequence = &BlockStatement{
				Token:      p.curToken(),
				Statements: []Statement{singleStmt},
			}
		}
		// Skip semicolon if present
		if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}
	}

	// Check for else
	if p.peekTokenIs(lexer.TOKEN_ELSE) {
		p.nextToken() // move to else

		// else の後は単なる文（statement）
		// それが if 文なら else if になるし、ブロックならブロック
		p.nextToken() // move past else
		altStmt := p.parseStatement()
		if altStmt != nil {
			// If it's already a block or if statement, use it directly
			if block, ok := altStmt.(*BlockStatement); ok {
				stmt.Alternative = block
			} else if ifStmt, ok := altStmt.(*IfStatement); ok {
				stmt.Alternative = ifStmt
			} else {
				// Wrap single statement in a block
				stmt.Alternative = &BlockStatement{
					Token:      p.curToken(),
					Statements: []Statement{altStmt},
				}
			}
		}
	}

	return stmt
}

// parseForStatement parses a for loop.
// Requirement 3.8: For loops with init, condition, post, and body.
// Supports both braced blocks and single statements without braces.
func (p *Parser) parseForStatement() Statement {
	stmt := &ForStatement{Token: p.curToken()}

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// Parse init
	p.nextToken()
	if !p.curTokenIs(lexer.TOKEN_SEMICOLON) {
		stmt.Init = p.parseStatement()
	}

	// Ensure we're at semicolon after init
	if !p.curTokenIs(lexer.TOKEN_SEMICOLON) {
		if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
			return nil
		}
	}

	// Parse condition
	p.nextToken()
	if !p.curTokenIs(lexer.TOKEN_SEMICOLON) {
		stmt.Condition = p.parseExpression(LOWEST)
	}

	if !p.expectPeek(lexer.TOKEN_SEMICOLON) {
		return nil
	}

	// Parse post
	p.nextToken()
	if !p.curTokenIs(lexer.TOKEN_RPAREN) {
		stmt.Post = p.parseStatement()
	}

	// Ensure we're at )
	if !p.curTokenIs(lexer.TOKEN_RPAREN) {
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	}

	// Check if body has braces or is a single statement
	if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		stmt.Body = p.parseBlockStatement()
	} else {
		// Single statement without braces: for(...) stmt;
		p.nextToken()
		singleStmt := p.parseStatement()
		if singleStmt != nil {
			stmt.Body = &BlockStatement{
				Token:      p.curToken(),
				Statements: []Statement{singleStmt},
			}
		}
		// Skip semicolon if present
		if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}
	}

	return stmt
}

// parseWhileStatement parses a while loop.
// Requirement 3.9: While loops with condition and body.
// Supports both braced blocks and single statements without braces.
func (p *Parser) parseWhileStatement() Statement {
	stmt := &WhileStatement{Token: p.curToken()}

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	// Check if body has braces or is a single statement
	if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		stmt.Body = p.parseBlockStatement()
	} else {
		// Single statement without braces: while(cond) stmt;
		p.nextToken()
		singleStmt := p.parseStatement()
		if singleStmt != nil {
			stmt.Body = &BlockStatement{
				Token:      p.curToken(),
				Statements: []Statement{singleStmt},
			}
		}
		// Skip semicolon if present
		if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
		}
	}

	return stmt
}

// parseSwitchStatement parses a switch statement.
// Requirement 3.10: Switch statements with value, case clauses, and optional default.
// TODO: Full implementation in task 3.6
func (p *Parser) parseSwitchStatement() Statement {
	stmt := &SwitchStatement{Token: p.curToken()}

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	p.nextToken() // skip {

	// Parse case clauses
	for !p.curTokenIs(lexer.TOKEN_RBRACE) && !p.curTokenIs(lexer.TOKEN_EOF) {
		if p.curTokenIs(lexer.TOKEN_CASE) {
			caseClause := p.parseCaseClause()
			if caseClause != nil {
				stmt.Cases = append(stmt.Cases, caseClause)
			}
		} else if p.curTokenIs(lexer.TOKEN_DEFAULT) {
			p.nextToken() // skip default
			if p.curTokenIs(lexer.TOKEN_COLON) {
				p.nextToken() // skip :
			}
			stmt.Default = &BlockStatement{
				Token:      p.curToken(),
				Statements: []Statement{},
			}
			for !p.curTokenIs(lexer.TOKEN_CASE) && !p.curTokenIs(lexer.TOKEN_RBRACE) && !p.curTokenIs(lexer.TOKEN_EOF) {
				s := p.parseStatement()
				if s != nil {
					stmt.Default.Statements = append(stmt.Default.Statements, s)
				}
				p.nextToken()
			}
		} else {
			p.nextToken()
		}
	}

	return stmt
}

// parseCaseClause parses a case clause in a switch statement.
func (p *Parser) parseCaseClause() *CaseClause {
	clause := &CaseClause{Token: p.curToken()}

	p.nextToken() // skip case
	clause.Value = p.parseExpression(LOWEST)

	if p.peekTokenIs(lexer.TOKEN_COLON) {
		p.nextToken() // skip to :
	}
	p.nextToken() // skip :

	// Parse statements until next case, default, or }
	for !p.curTokenIs(lexer.TOKEN_CASE) && !p.curTokenIs(lexer.TOKEN_DEFAULT) &&
		!p.curTokenIs(lexer.TOKEN_RBRACE) && !p.curTokenIs(lexer.TOKEN_EOF) {
		stmt := p.parseStatement()
		if stmt != nil {
			clause.Body = append(clause.Body, stmt)
		}
		p.nextToken()
	}

	return clause
}

// parseMesStatement parses a mes (event handler) statement.
// Requirement 3.11: mes(EVENT) blocks with event type and body.
// Requirement 9.1: EVENT types: TIME, MIDI_TIME, MIDI_END, KEY, CLICK, RBDOWN, RBDBLCLK, USER
func (p *Parser) parseMesStatement() Statement {
	stmt := &MesStatement{Token: p.curToken()}

	if !p.expectPeek(lexer.TOKEN_LPAREN) {
		return nil
	}

	// Parse event type (identifier)
	if !p.expectPeek(lexer.TOKEN_IDENT) {
		return nil
	}
	stmt.EventType = p.curToken().Literal

	if !p.expectPeek(lexer.TOKEN_RPAREN) {
		return nil
	}

	if !p.expectPeek(lexer.TOKEN_LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	// Skip optional trailing semicolon: mes(EVENT){};
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseStepStatement parses a step statement.
// Requirement 3.12, 9.2, 9.3: step() statements with count and body.
// TODO: Full implementation in task 3.8
func (p *Parser) parseStepStatement() Statement {
	stmt := &StepStatement{Token: p.curToken()}

	// Check for step count: step(n) or step
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken() // skip to (
		if !p.peekTokenIs(lexer.TOKEN_RPAREN) {
			p.nextToken()
			stmt.Count = p.parseExpression(LOWEST)
		}
		if !p.expectPeek(lexer.TOKEN_RPAREN) {
			return nil
		}
	}

	// Check for body
	if p.peekTokenIs(lexer.TOKEN_LBRACE) {
		p.nextToken()
		stmt.Body = p.parseStepBody()
	}

	return stmt
}

// parseStepBody parses the body of a step statement.
// Commas are interpreted as wait commands.
// Requirement 9.2, 9.3: Commas in step blocks are wait instructions.
// Requirement 9.4: end_step keyword marks step block end (but statements after it are still parsed)
//
// In FILLY, the step body can contain:
// - Statements followed by commas (wait instructions)
// - Consecutive commas (multiple wait steps)
// - end_step keyword (marks end of timed sequence, but statements after it are still executed)
// - del_me, del_us, del_all (cleanup commands, often after end_step)
//
// Example from ROBOT.TFY: step(20){,start();end_step;del_me;}
// - Leading comma: 1 wait step
// - start(): function call
// - end_step: marks end of timed sequence
// - del_me: cleanup command (still part of step body)
func (p *Parser) parseStepBody() *StepBody {
	body := &StepBody{Commands: []*StepCommand{}}

	p.nextToken() // skip {

	endStepSeen := false

	for !p.curTokenIs(lexer.TOKEN_RBRACE) && !p.curTokenIs(lexer.TOKEN_EOF) {

		// Handle end_step keyword
		if p.curTokenIs(lexer.TOKEN_END_STEP) {
			endStepSeen = true
			p.nextToken()
			// Skip optional semicolon after end_step
			if p.curTokenIs(lexer.TOKEN_SEMICOLON) {
				p.nextToken()
			}
			continue
		}

		// Skip empty statements (just semicolons)
		if p.curTokenIs(lexer.TOKEN_SEMICOLON) {
			p.nextToken()
			continue
		}

		// Count leading commas (wait-only) - only before end_step
		waitCount := 0
		if !endStepSeen {
			for p.curTokenIs(lexer.TOKEN_COMMA) {
				waitCount++
				p.nextToken()
			}

			if waitCount > 0 {
				cmd := &StepCommand{
					Statement: nil,
					WaitCount: waitCount,
				}
				body.Commands = append(body.Commands, cmd)
			}
		}

		// Check for end of block or end_step after commas
		if p.curTokenIs(lexer.TOKEN_RBRACE) || p.curTokenIs(lexer.TOKEN_EOF) ||
			p.curTokenIs(lexer.TOKEN_END_STEP) {
			continue
		}

		// Parse statement
		stmt := p.parseStatement()
		if stmt == nil {
			// parseStatement returned nil (e.g., hit } after skipping semicolons)
			continue
		}

		// Count trailing commas - only before end_step
		waitCount = 0
		if !endStepSeen {
			for p.peekTokenIs(lexer.TOKEN_COMMA) {
				p.nextToken()
				waitCount++
			}
		}

		cmd := &StepCommand{
			Statement: stmt,
			WaitCount: waitCount,
		}
		body.Commands = append(body.Commands, cmd)

		p.nextToken()
	}

	return body
}

// parseBreakStatement parses a break statement.
func (p *Parser) parseBreakStatement() Statement {
	stmt := &BreakStatement{Token: p.curToken()}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseContinueStatement parses a continue statement.
func (p *Parser) parseContinueStatement() Statement {
	stmt := &ContinueStatement{Token: p.curToken()}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseReturnStatement parses a return statement.
func (p *Parser) parseReturnStatement() Statement {
	stmt := &ReturnStatement{Token: p.curToken()}

	// Check for return value
	if !p.peekTokenIs(lexer.TOKEN_SEMICOLON) && !p.peekTokenIs(lexer.TOKEN_RBRACE) {
		p.nextToken()
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseExpressionStatement parses an expression statement.
func (p *Parser) parseExpressionStatement() Statement {
	stmt := &ExpressionStatement{
		Token:      p.curToken(),
		Expression: p.parseExpression(LOWEST),
	}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseSpecialCommandStatement parses special command keywords as function calls.
// Requirement 9.5: del_me, del_us, del_all are treated as function calls.
// Requirement 9.4: end_step is treated as step block end marker.
// These keywords can be used without parentheses: del_me; or del_me();
func (p *Parser) parseSpecialCommandStatement() Statement {
	tok := p.curToken()
	funcName := tok.Literal

	// Create a CallExpression for the special command
	callExpr := &CallExpression{
		Token:     tok,
		Function:  funcName,
		Arguments: []Expression{},
	}

	// Check for optional parentheses: del_me() or del_me
	if p.peekTokenIs(lexer.TOKEN_LPAREN) {
		p.nextToken() // skip to (
		// Parse arguments if any (usually none for these commands)
		callExpr.Arguments = p.parseExpressionList(lexer.TOKEN_RPAREN)
	}

	stmt := &ExpressionStatement{
		Token:      tok,
		Expression: callExpr,
	}

	// Skip optional semicolon
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}

// parseInfoDirective parses a #info preprocessor directive.
// Requirement 3.17: #info directives create InfoDirective nodes.
// Example: #info INAM "Title Name"
func (p *Parser) parseInfoDirective() Statement {
	tok := p.curToken()
	literal := tok.Literal

	// Parse key and value from the literal
	// Format: #info KEY "value" or #info KEY value
	key := ""
	value := ""

	// Skip "#info " prefix and parse the rest
	parts := strings.SplitN(literal, " ", 2)
	if len(parts) >= 1 {
		key = parts[0]
	}
	if len(parts) >= 2 {
		value = strings.TrimSpace(parts[1])
		// Remove quotes if present
		if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
			value = value[1 : len(value)-1]
		}
	}

	return &InfoDirective{
		Token: tok,
		Key:   key,
		Value: value,
	}
}

// parseIncludeDirective parses a #include preprocessor directive.
// Requirement 3.18: #include directives create IncludeDirective nodes.
// Example: #include "filename.tfy"
func (p *Parser) parseIncludeDirective() Statement {
	tok := p.curToken()

	return &IncludeDirective{
		Token:    tok,
		FileName: tok.Literal, // Lexer already extracted the filename
	}
}

// parseDefineDirective parses a #define preprocessor directive.
// Example: #define MAXLINE 24
func (p *Parser) parseDefineDirective() Statement {
	tok := p.curToken()

	// Parse the literal to extract name and value
	// Format: "#define NAME value"
	literal := tok.Literal
	name := ""
	value := ""

	// Skip "#define " prefix
	if len(literal) > 8 {
		rest := literal[8:] // Skip "#define "
		// Find the first space to separate name and value
		for i, ch := range rest {
			if ch == ' ' || ch == '\t' {
				name = rest[:i]
				value = rest[i+1:]
				break
			}
		}
		if name == "" {
			name = rest
		}
	}

	return &DefineDirective{
		Token: tok,
		Name:  name,
		Value: value,
	}
}

// parseGenericDirective parses an unknown preprocessor directive.
// It simply skips the directive and returns nil.
func (p *Parser) parseGenericDirective() Statement {
	// Skip the directive token
	p.nextToken()
	return nil
}

// parseLabelStatement parses a label statement (for goto).
// Example: END:
func (p *Parser) parseLabelStatement() Statement {
	stmt := &LabelStatement{
		Token: p.curToken(),
		Name:  p.curToken().Literal,
	}

	p.nextToken() // skip to :
	// Skip optional semicolon after label
	if p.peekTokenIs(lexer.TOKEN_SEMICOLON) {
		p.nextToken()
	}

	return stmt
}
