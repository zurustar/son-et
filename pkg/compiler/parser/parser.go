package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/token"
)

// Precedence levels for operators.
const (
	_ int = iota
	LOWEST
	OR          // ||
	AND         // &&
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.OR:       OR,
	token.AND:      AND,
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GT:       LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.MULT:     PRODUCT,
	token.DIV:      PRODUCT,
	token.MOD:      PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

// Parser parses FILLY source code into an AST.
type Parser struct {
	l      *lexer.Lexer
	errors []string
	source []string // Source code lines for error reporting

	curToken  token.Token
	peekToken token.Token

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// New creates a new Parser.
func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
		source: strings.Split(l.GetSource(), "\n"),
	}

	// Register prefix parse functions
	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT_LIT, p.parseIntegerLiteral)
	p.registerPrefix(token.FLOAT_LIT, p.parseFloatLiteral)
	p.registerPrefix(token.STRING_LIT, p.parseStringLiteral)
	p.registerPrefix(token.NOT, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)
	p.registerPrefix(token.LBRACKET, p.parseArrayLiteral)

	// Allow keywords to be used as identifiers in expressions
	// This is needed for cases like: PostMes(USER, ...) or int time=1
	p.registerPrefix(token.TIME, p.parseIdentifier)
	p.registerPrefix(token.MIDI_TIME, p.parseIdentifier)
	p.registerPrefix(token.MIDI_END, p.parseIdentifier)
	p.registerPrefix(token.KEY, p.parseIdentifier)
	p.registerPrefix(token.CLICK, p.parseIdentifier)
	p.registerPrefix(token.RBDOWN, p.parseIdentifier)
	p.registerPrefix(token.RBDBLCLK, p.parseIdentifier)
	p.registerPrefix(token.USER, p.parseIdentifier)
	p.registerPrefix(token.STEP, p.parseIdentifier)

	// Register infix parse functions
	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.MULT, p.parseInfixExpression)
	p.registerInfix(token.DIV, p.parseInfixExpression)
	p.registerInfix(token.MOD, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	// Read two tokens to initialize curToken and peekToken
	p.nextToken()
	p.nextToken()

	return p
}

// Errors returns the parser errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// ParseProgram parses the entire program.
func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		// Skip comments
		if p.curToken.Type == token.COMMENT {
			p.nextToken()
			continue
		}

		// Skip semicolons (statement terminators)
		if p.curToken.Type == token.SEMICOLON {
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.INT, token.STRING, token.FLOAT_TYPE:
		// Variable declaration
		return p.parseVarDeclaration()
	case token.FUNCTION:
		// Function declaration with 'function' keyword
		return p.parseFunctionDeclaration()
	case token.IF:
		return p.parseIfStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.BREAK:
		return p.parseBreakStatement()
	case token.CONTINUE:
		return p.parseContinueStatement()
	case token.RETURN:
		return p.parseReturnStatement()
	case token.MES:
		return p.parseMesStatement()
	case token.STEP:
		return p.parseStepStatement()
	case token.DEL_ME:
		return &ast.ExpressionStatement{
			Token:      p.curToken,
			Expression: &ast.CallExpression{Token: p.curToken, Function: &ast.Identifier{Token: p.curToken, Value: "del_me"}},
		}
	case token.DEL_US:
		return &ast.ExpressionStatement{
			Token:      p.curToken,
			Expression: &ast.CallExpression{Token: p.curToken, Function: &ast.Identifier{Token: p.curToken, Value: "del_us"}},
		}
	case token.DEL_ALL:
		return &ast.ExpressionStatement{
			Token:      p.curToken,
			Expression: &ast.CallExpression{Token: p.curToken, Function: &ast.Identifier{Token: p.curToken, Value: "del_all"}},
		}
	case token.END_STEP:
		return &ast.ExpressionStatement{
			Token:      p.curToken,
			Expression: &ast.CallExpression{Token: p.curToken, Function: &ast.Identifier{Token: p.curToken, Value: "end_step"}},
		}
	case token.IDENT:
		// Check if it's a function declaration: identifier followed by (
		if p.peekTokenIs(token.LPAREN) {
			// Could be function declaration or function call
			// Look ahead further to distinguish
			return p.parseFunctionOrCall()
		}
		// Check if it's an assignment
		if p.peekTokenIs(token.ASSIGN) {
			return p.parseAssignStatement()
		}
		// Check for array assignment (arr[i] = value)
		if p.peekTokenIs(token.LBRACKET) {
			// Look ahead to see if this is an assignment
			// We need to parse the expression and check if = follows
			expr := p.parseExpression(LOWEST)
			if p.peekTokenIs(token.ASSIGN) {
				stmt := &ast.AssignStatement{Token: p.curToken}
				stmt.Name = expr
				p.nextToken() // consume =
				p.nextToken()
				stmt.Value = p.parseExpression(LOWEST)
				return stmt
			}
			// Not an assignment, return as expression statement
			return &ast.ExpressionStatement{Token: p.curToken, Expression: expr}
		}
		return p.parseExpressionStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *Parser) parseAssignStatement() ast.Statement {
	stmt := &ast.AssignStatement{Token: p.curToken}

	// Parse left side (identifier or array access)
	name := p.parseExpression(LOWEST)
	stmt.Name = name

	if !p.expectPeek(token.ASSIGN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	return stmt
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)
	return stmt
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.EOF) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()
		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	// Detect base: 0x/0X for hex, otherwise decimal
	base := 10
	literal := p.curToken.Literal
	if len(literal) > 2 && literal[0] == '0' && (literal[1] == 'x' || literal[1] == 'X') {
		base = 16
		// For hex, strconv.ParseInt expects the string without the 0x prefix
		literal = literal[2:]
	}

	value, err := strconv.ParseInt(literal, base, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	lit := &ast.FloatLiteral{Token: p.curToken}

	value, err := strconv.ParseFloat(p.curToken.Literal, 64)
	if err != nil {
		msg := fmt.Sprintf("could not parse %q as float", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	array := &ast.ArrayLiteral{Token: p.curToken}
	array.Elements = p.parseExpressionList(token.RBRACKET)
	return array
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
	}

	p.nextToken()
	expression.Right = p.parseExpression(PREFIX)

	return expression
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Token:    p.curToken,
		Operator: p.curToken.Literal,
		Left:     left,
	}

	precedence := p.curPrecedence()
	p.nextToken()
	expression.Right = p.parseExpression(precedence)

	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.nextToken()

	exp := p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return exp
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function}
	exp.Arguments = p.parseExpressionList(token.RPAREN)
	return exp
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) parseExpressionList(end token.TokenType) []ast.Expression {
	list := []ast.Expression{}

	if p.peekTokenIs(end) {
		p.nextToken()
		return list
	}

	p.nextToken()
	list = append(list, p.parseExpression(LOWEST))

	for p.peekTokenIs(token.COMMA) {
		p.nextToken()
		p.nextToken()
		list = append(list, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(end) {
		return nil
	}

	return list
}

// Statement parsing functions
func (p *Parser) parseIfStatement() ast.Statement {
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Consequence = p.parseBlockStatement()

	if p.peekTokenIs(token.ELSE) {
		p.nextToken()

		// Check for "else if" - handle as nested if statement
		if p.peekTokenIs(token.IF) {
			p.nextToken()
			nestedIf := p.parseIfStatement()
			if nestedIf == nil {
				return nil
			}
			// Wrap the nested if in a block statement
			stmt.Alternative = &ast.BlockStatement{
				Token:      p.curToken,
				Statements: []ast.Statement{nestedIf},
			}
		} else if p.expectPeek(token.LBRACE) {
			stmt.Alternative = p.parseBlockStatement()
		} else {
			return nil
		}
	}

	return stmt
}

func (p *Parser) parseForStatement() ast.Statement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Init = p.parseStatement()

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.SEMICOLON) {
		return nil
	}

	p.nextToken()
	stmt.Post = p.parseStatement()

	// Allow optional trailing semicolon before closing paren
	// This supports both for(i=0; i<10; i=i+1) and for(i=0; i<10; i=i+1;)
	if p.peekTokenIs(token.SEMICOLON) {
		p.nextToken()
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseWhileStatement() ast.Statement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseSwitchStatement() ast.Statement {
	stmt := &ast.SwitchStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.Value = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curTokenIs(token.CASE) {
			caseClause := &ast.CaseClause{Token: p.curToken}

			p.nextToken()
			caseClause.Value = p.parseExpression(LOWEST)

			if !p.expectPeek(token.LBRACE) {
				return nil
			}

			caseClause.Body = p.parseBlockStatement()
			stmt.Cases = append(stmt.Cases, caseClause)
		} else if p.curTokenIs(token.DEFAULT) {
			if !p.expectPeek(token.LBRACE) {
				return nil
			}

			stmt.Default = p.parseBlockStatement()
		}

		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseBreakStatement() ast.Statement {
	return &ast.BreakStatement{Token: p.curToken}
}

func (p *Parser) parseContinueStatement() ast.Statement {
	return &ast.ContinueStatement{Token: p.curToken}
}

func (p *Parser) parseReturnStatement() ast.Statement {
	stmt := &ast.ReturnStatement{Token: p.curToken}

	p.nextToken()

	if !p.curTokenIs(token.EOF) && !p.curTokenIs(token.RBRACE) {
		stmt.ReturnValue = p.parseExpression(LOWEST)
	}

	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.Identifier {
	identifiers := []*ast.Identifier{}

	// Check if empty parameter list
	if p.peekTokenIs(token.RPAREN) {
		return identifiers
	}

	p.nextToken()

	// Handle first parameter
	if !p.curTokenIs(token.RPAREN) {
		// Check if this is a typed parameter (int x, str s, etc.)
		if p.curToken.Type == token.INT || p.curToken.Type == token.STRING || p.curToken.Type == token.FLOAT_TYPE {
			// Skip type, get parameter name
			p.nextToken()
			// Allow keywords as parameter names (e.g., int time, int step)
			if p.curToken.Type != token.IDENT && p.curToken.Type != token.TIME && p.curToken.Type != token.STEP {
				return identifiers
			}
		}

		ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
		identifiers = append(identifiers, ident)

		// Check for array parameter syntax: param[]
		if p.peekTokenIs(token.LBRACKET) {
			p.nextToken() // consume [
			if !p.expectPeek(token.RBRACKET) {
				return identifiers
			}
			// Array parameter - we don't need to store this info in the AST
			// since FILLY doesn't enforce type checking
		}

		// Check for default value (e.g., int time=1)
		if p.peekTokenIs(token.ASSIGN) {
			p.nextToken() // consume =
			p.nextToken() // consume default value
			// TODO: Store default value in AST if needed
		}

		// Handle remaining parameters
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next parameter

			// Check if this is a typed parameter
			if p.curToken.Type == token.INT || p.curToken.Type == token.STRING || p.curToken.Type == token.FLOAT_TYPE {
				// Skip type, get parameter name
				p.nextToken()
				// Allow keywords as parameter names
				if p.curToken.Type != token.IDENT && p.curToken.Type != token.TIME && p.curToken.Type != token.STEP {
					return identifiers
				}
			}

			ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
			identifiers = append(identifiers, ident)

			// Check for array parameter syntax: param[]
			if p.peekTokenIs(token.LBRACKET) {
				p.nextToken() // consume [
				if !p.expectPeek(token.RBRACKET) {
					return identifiers
				}
				// Array parameter - we don't need to store this info in the AST
				// since FILLY doesn't enforce type checking
			}

			// Check for default value
			if p.peekTokenIs(token.ASSIGN) {
				p.nextToken() // consume =
				p.nextToken() // consume default value
				// TODO: Store default value in AST if needed
			}
		}
	}

	return identifiers
}

func (p *Parser) parseMesStatement() ast.Statement {
	stmt := &ast.MesStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	p.nextToken()
	stmt.EventType = p.curToken.Type

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

func (p *Parser) parseStepStatement() ast.Statement {
	stmt := &ast.StepStatement{Token: p.curToken}

	// Check if there's a parenthesis (step(N)) or direct block (step{})
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken() // consume (

		p.nextToken()
		stmt.Count = p.parseExpression(LOWEST)

		if !p.expectPeek(token.RPAREN) {
			return nil
		}
	} else {
		// No parenthesis means step{} which defaults to step(1)
		stmt.Count = &ast.IntegerLiteral{
			Token: p.curToken,
			Value: 1,
		}
	}

	// Check if there's a block: step(N) { ... } or step { ... }
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		stmt.Body = p.parseStepBlock()
	}

	return stmt
}

// parseStepBlock parses a step block where each line represents one step
// and commas represent empty steps
func (p *Parser) parseStepBlock() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		// Skip comments
		if p.curToken.Type == token.COMMENT {
			p.nextToken()
			continue
		}

		// Handle leading commas (empty steps at the start)
		if p.curTokenIs(token.COMMA) {
			block.Statements = append(block.Statements, &ast.ExpressionStatement{
				Token:      p.curToken,
				Expression: nil, // Empty step
			})
			p.nextToken()
			continue
		}

		// Parse statement
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		// In step blocks, semicolons and commas are both statement terminators
		// Commas represent empty steps
		for p.peekTokenIs(token.SEMICOLON) || p.peekTokenIs(token.COMMA) {
			p.nextToken()
			// Multiple commas in a row represent multiple empty steps
			if p.curTokenIs(token.COMMA) {
				// Add empty statement for each comma
				block.Statements = append(block.Statements, &ast.ExpressionStatement{
					Token:      p.curToken,
					Expression: nil, // Empty step
				})
			}
		}

		p.nextToken()
	}

	return block
}

// parseVarDeclaration parses variable declarations like:
// int x;
// int x, y, z;
// int arr[];
// int arr[10];
func (p *Parser) parseVarDeclaration() ast.Statement {
	decl := &ast.VarDeclaration{Token: p.curToken}

	// Get type
	switch p.curToken.Type {
	case token.INT:
		decl.Type = "int"
	case token.STRING:
		decl.Type = "string"
	case token.FLOAT_TYPE:
		decl.Type = "float"
	}

	// Parse variable names
	if !p.expectPeek(token.IDENT) {
		return nil
	}

	for {
		varSpec := &ast.VarSpec{
			Name: &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
		}

		// Check for array declaration
		if p.peekTokenIs(token.LBRACKET) {
			p.nextToken() // consume [
			varSpec.IsArray = true

			// Check if there's a size
			if !p.peekTokenIs(token.RBRACKET) {
				p.nextToken()
				varSpec.Size = p.parseExpression(LOWEST)
			}

			if !p.expectPeek(token.RBRACKET) {
				return nil
			}
		}

		decl.Names = append(decl.Names, varSpec)

		// Check for more variables (comma-separated)
		if !p.peekTokenIs(token.COMMA) {
			break
		}
		p.nextToken() // consume comma
		if !p.expectPeek(token.IDENT) {
			return nil
		}
	}

	// Don't consume semicolon here - let ParseProgram handle it
	// This ensures consistent token advancement across all statement types
	return decl
}

// parseFunctionDeclaration parses a function declaration starting with 'function' keyword.
// Format: function name(params) { body }
func (p *Parser) parseFunctionDeclaration() ast.Statement {
	stmt := &ast.FunctionStatement{Token: p.curToken}

	if !p.expectPeek(token.IDENT) {
		return nil
	}

	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = p.parseFunctionParameters()

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if !p.expectPeek(token.LBRACE) {
		return nil
	}

	stmt.Body = p.parseBlockStatement()

	return stmt
}

// parseFunctionOrCall determines if this is a function declaration or call
func (p *Parser) parseFunctionOrCall() ast.Statement {
	// Save current position
	name := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Check if this looks like a typed parameter (int, str, float)
	// If so, it's definitely a function declaration
	if p.peekToken.Type == token.INT || p.peekToken.Type == token.STRING || p.peekToken.Type == token.FLOAT_TYPE {
		// This is a function declaration with typed parameters
		params := p.parseFunctionParameters()

		if !p.expectPeek(token.RPAREN) {
			return nil
		}

		if !p.expectPeek(token.LBRACE) {
			return nil
		}

		body := p.parseBlockStatement()

		return &ast.FunctionStatement{
			Token:      name.Token,
			Name:       name,
			Parameters: params,
			Body:       body,
		}
	}

	// Parse parameters/arguments manually to handle array syntax
	params := []*ast.Identifier{}

	// Check if empty parameter list
	if !p.peekTokenIs(token.RPAREN) {
		p.nextToken()

		// Parse first parameter/argument
		if p.curToken.Type == token.IDENT || p.curToken.Type == token.TIME || p.curToken.Type == token.STEP {
			ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

			// Check for array syntax: param[]
			if p.peekTokenIs(token.LBRACKET) {
				p.nextToken() // consume [
				if !p.expectPeek(token.RBRACKET) {
					return nil
				}
			}

			// Check if this is a function call (identifier followed by LPAREN)
			// If so, this must be a function call with complex arguments, not a declaration
			if p.peekTokenIs(token.LPAREN) {
				return p.parseFunctionCallFallback(name, params)
			}

			params = append(params, ident)

			// Check for default value (e.g., l=10)
			if p.peekTokenIs(token.ASSIGN) {
				p.nextToken() // consume =
				p.nextToken() // consume default value
			}

			// Parse remaining parameters
			for p.peekTokenIs(token.COMMA) {
				p.nextToken() // consume comma
				p.nextToken() // move to next parameter

				if p.curToken.Type == token.IDENT || p.curToken.Type == token.TIME || p.curToken.Type == token.STEP {
					ident := &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

					// Check for array syntax: param[]
					if p.peekTokenIs(token.LBRACKET) {
						p.nextToken() // consume [
						if !p.expectPeek(token.RBRACKET) {
							return nil
						}
					}

					// Check if this is a function call (identifier followed by LPAREN)
					if p.peekTokenIs(token.LPAREN) {
						return p.parseFunctionCallFallback(name, params)
					}

					params = append(params, ident)

					// Check for default value
					if p.peekTokenIs(token.ASSIGN) {
						p.nextToken() // consume =
						p.nextToken() // consume default value
					}
				} else {
					// Not a simple identifier - this must be a function call with complex arguments
					// Reparse as expression list
					p.errors = []string{} // Clear any errors from parameter parsing
					return p.parseFunctionCallFallback(name, params)
				}
			}
		} else {
			// Not a simple identifier - this must be a function call with complex arguments
			return p.parseFunctionCallFallback(name, params)
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	// If followed by {, it's a function declaration
	if p.peekTokenIs(token.LBRACE) {
		p.nextToken()
		body := p.parseBlockStatement()

		return &ast.FunctionStatement{
			Token:      name.Token,
			Name:       name,
			Parameters: params,
			Body:       body,
		}
	}

	// Otherwise, it's a function call
	return &ast.ExpressionStatement{
		Token: name.Token,
		Expression: &ast.CallExpression{
			Token:     name.Token,
			Function:  name,
			Arguments: convertIdentifiersToExpressions(params),
		},
	}
}

// parseFunctionCallFallback handles function calls with complex arguments
// We've already parsed some tokens, so we need to build the expression list from what we have
func (p *Parser) parseFunctionCallFallback(name *ast.Identifier, alreadyParsed []*ast.Identifier) ast.Statement {
	// Convert already parsed identifiers to expressions first
	args := convertIdentifiersToExpressions(alreadyParsed)

	// Parse current token as expression only if we're not at RPAREN
	if !p.curTokenIs(token.RPAREN) {
		expr := p.parseExpression(LOWEST)
		if expr != nil {
			args = append(args, expr)
		}

		// Parse remaining arguments
		for p.peekTokenIs(token.COMMA) {
			p.nextToken() // consume comma
			p.nextToken() // move to next argument
			expr := p.parseExpression(LOWEST)
			if expr != nil {
				args = append(args, expr)
			}
		}
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return &ast.ExpressionStatement{
		Token: name.Token,
		Expression: &ast.CallExpression{
			Token:     name.Token,
			Function:  name,
			Arguments: args,
		},
	}
}

// convertIdentifiersToExpressions converts a slice of Identifiers to Expressions
func convertIdentifiersToExpressions(idents []*ast.Identifier) []ast.Expression {
	exprs := make([]ast.Expression, len(idents))
	for i, ident := range idents {
		exprs[i] = ident
	}
	return exprs
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for !p.curTokenIs(token.RBRACE) && !p.curTokenIs(token.EOF) {
		if p.curToken.Type == token.COMMENT {
			p.nextToken()
			continue
		}

		// Handle leading commas (empty steps)
		if p.curTokenIs(token.COMMA) {
			block.Statements = append(block.Statements, &ast.ExpressionStatement{
				Token:      p.curToken,
				Expression: nil, // Empty step
			})
			p.nextToken()
			continue
		}

		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}

		// Skip optional semicolons
		for p.peekTokenIs(token.SEMICOLON) {
			p.nextToken()
		}

		// Handle trailing commas (empty steps after statements)
		for p.peekTokenIs(token.COMMA) {
			p.nextToken()
			// Add empty statement for each comma
			block.Statements = append(block.Statements, &ast.ExpressionStatement{
				Token:      p.curToken,
				Expression: nil, // Empty step
			})
		}

		p.nextToken()
	}

	return block
}

// Helper functions
func (p *Parser) curTokenIs(t token.TokenType) bool {
	return p.curToken.Type == t
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()

	// Skip comments in both curToken and peekToken
	for p.curToken.Type == token.COMMENT {
		p.curToken = p.peekToken
		p.peekToken = p.l.NextToken()
	}

	// Skip comments
	for p.peekToken.Type == token.COMMENT {
		p.peekToken = p.l.NextToken()
	}
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("expected next token to be %s, got %s instead at line %d, column %d",
		t, p.peekToken.Type, p.peekToken.Line, p.peekToken.Column)
	msg = p.addContext(msg, p.peekToken.Line, p.peekToken.Column)
	p.errors = append(p.errors, msg)
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("no prefix parse function for %s found at line %d, column %d",
		t, p.curToken.Line, p.curToken.Column)
	msg = p.addContext(msg, p.curToken.Line, p.curToken.Column)
	p.errors = append(p.errors, msg)
}

// addContext adds source code context to an error message
func (p *Parser) addContext(msg string, line, column int) string {
	if len(p.source) == 0 {
		return msg
	}

	// Add newline and context
	result := msg + "\n"

	// Show 2 lines before and after the error line
	start := line - 3
	if start < 1 {
		start = 1
	}
	end := line + 2
	if end > len(p.source) {
		end = len(p.source)
	}

	for i := start; i <= end; i++ {
		lineNum := i
		sourceLine := ""
		if lineNum > 0 && lineNum <= len(p.source) {
			sourceLine = p.source[lineNum-1]
		}

		if lineNum == line {
			// Highlight the error line
			result += fmt.Sprintf("  > %4d | %s\n", lineNum, sourceLine)
			// Add pointer to the column
			if column > 0 {
				pointer := strings.Repeat(" ", column+8) + "^"
				result += pointer + "\n"
			}
		} else {
			result += fmt.Sprintf("    %4d | %s\n", lineNum, sourceLine)
		}
	}

	return result
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}
