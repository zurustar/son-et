package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/token"
)

const (
	_ int = iota
	LOWEST
	LOGICAL_OR  // ||
	LOGICAL_AND // &&
	EQUALS      // ==
	LESSGREATER // > or <
	SUM         // +
	PRODUCT     // *
	PREFIX      // -X or !X
	CALL        // myFunction(X)
	INDEX       // array[index]
)

var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NOT_EQ:   EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.AND:      LOGICAL_AND,
	token.OR:       LOGICAL_OR,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	token.LPAREN:   CALL,
	token.LBRACKET: INDEX,
}

type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

type Parser struct {
	l *lexer.Lexer

	curToken  token.Token
	peekToken token.Token

	errors []string

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		errors: []string{},
	}

	p.prefixParseFns = make(map[token.TokenType]prefixParseFn)
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.NUMBER, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	p.infixParseFns = make(map[token.TokenType]infixParseFn)
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NOT_EQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)
	p.registerInfix(token.AND, p.parseInfixExpression)
	p.registerInfix(token.OR, p.parseInfixExpression)
	p.registerInfix(token.LPAREN, p.parseCallExpression)
	p.registerInfix(token.LBRACKET, p.parseIndexExpression)

	p.nextToken()
	p.nextToken()

	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Errors() []string {
	return p.errors
}

func (p *Parser) ParseProgram() *ast.Program {
	program := &ast.Program{}
	program.Statements = []ast.Statement{}

	for p.curToken.Type != token.EOF {
		// 1. Skip # directives (or parse #define)
		if p.curToken.Literal == "#" {
			if p.peekToken.Literal == "define" { // Handle #define
				stmt := p.parseDefineStatement()
				if stmt != nil {
					program.Statements = append(program.Statements, stmt)
				}
				p.nextToken()
				continue
			}
			// Skip other directives (#info, #include remains?)
			p.skipLine()
			continue
		}

		// 2. Global Variable Declaration
		if p.curToken.Type == token.INT || p.curToken.Type == token.STR {
			vars := p.parseGlobalVars()
			program.Statements = append(program.Statements, vars...)
			continue
		}

		// 3. Function Definition (start with IDENT)
		// Check if it's a function definition: Name(...) { ... }
		// We use infinite lookahead to distinguish from CallStatement: Name(...);
		if p.curToken.Type == token.IDENT && p.peekToken.Type == token.LPAREN {
			if p.isFunctionDefinition() {
				fn := p.parseFunction()
				if fn != nil {
					program.Statements = append(program.Statements, fn)
				}
				// After parsing a function, we continue the loop to parse next item
				continue
			}
			// If not a function definition, it falls through to parseStatement below
		}

		// 4. Statements (Top-level statements)
		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *Parser) parseGlobalVars() []ast.Statement {
	// int Name, Name[], ...; or str Name, Name[], ...;
	stmts := []ast.Statement{}

	// p.curToken is INT or STR
	typeToken := p.curToken // Save the type token
	p.nextToken()           // Skip type, curToken is now the first identifier

	for p.curToken.Type == token.IDENT {
		name := p.curToken.Literal
		isArray := false

		// Look ahead for []
		if p.peekToken.Type == token.LBRACKET {
			p.nextToken() // curToken = [
			if !p.expectPeek(token.RBRACKET) {
				return nil
			}
			isArray = true
			// curToken is now ]
		}

		// Create variable declaration
		stmt := &ast.LetStatement{
			Token: typeToken,
			Name:  &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: name}, Value: name},
		}
		if isArray {
			stmt.Name.Value = name + "[]"
		}

		stmts = append(stmts, stmt)

		// Check what comes next using Peek to avoid over-consuming
		if p.peekToken.Type == token.COMMA {
			p.nextToken() // curToken is now COMMA
			p.nextToken() // curToken is now next IDENT
			continue
		}

		if p.peekToken.Type == token.SEMICOLON {
			p.nextToken() // curToken is SEMICOLON
			// Done with variable declaration
			break
		}

		// If neither comma nor semicolon, we assume the list ends here.
		// curToken is currently the last part of this declaration (IDENT or RBRACKET).
		// We do NOT consume the next token so that ParseProgram can consume it as the start of the next statement.
		break
	}
	return stmts
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.curToken.Type {
	case token.SEMICOLON:
		return nil
	case token.INT, token.STR:
		// Local variable declaration: int x,y[],z; or str s;
		// Parse using parseGlobalVars logic but treat as local
		// For now, return the first LetStatement from the list
		stmts := p.parseGlobalVars()
		if len(stmts) > 0 {
			// HACK: only return first statement
			// Multiple declarations on one line are lost for now
			// TODO: Support multiple declarations properly
			return stmts[0]
		}
		return nil
	case token.IF:
		return p.parseIfStatement()
	case token.FOR:
		return p.parseForStatement()
	case token.WHILE:
		return p.parseWhileStatement()
	case token.DO:
		return p.parseDoWhileStatement()
	case token.SWITCH:
		return p.parseSwitchStatement()
	case token.BREAK:
		stmt := &ast.BreakStatement{Token: p.curToken}
		if p.peekToken.Type == token.SEMICOLON {
			p.nextToken()
		}
		return stmt
	case token.CONTINUE:
		stmt := &ast.ContinueStatement{Token: p.curToken}
		if p.peekToken.Type == token.SEMICOLON {
			p.nextToken()
		}
		return stmt
	case token.MES:
		return p.parseMesBlockStatement()
	case token.STEP:
		return p.parseStepBlockStatement()
	case token.IDENT:
		// Lookahead to decide Assignment vs Call vs Label
		if p.peekToken.Type == token.COLON {
			// Label statement: LABEL:
			// Just skip the label for now (labels are not used in code generation)
			p.nextToken() // skip COLON
			return nil    // Return nil to skip this statement
		}
		if p.peekToken.Type == token.ASSIGN {
			return p.parseAssignStatement()
		}
		if p.peekToken.Type == token.LBRACKET {
			// Array Assignment or Index Expression?
			// If assignment: ident[idx] = val
			// If expression: ident[idx]
			// We can peek further? No, peekToken is LBRACKET.
			// We can use a heuristic or assume AssignStatement logic handles the distinction?
			// parseAssignStatement currently checks for LBRACKET.
			// But if it is NOT an assignment, parseAssignStatement might fail?
			// Let's modify parseAssignStatement to handle fallback? No, simpler:
			// Treat as AssignStatement attempt.
			return p.parseAssignStatement()
		}
		// Default to ExpressionStatement (Call or variable usage statement)
		return p.parseExpressionStatement()

	case token.COMMA:
		return p.parseWaitStatement()
	default:
		return nil
	}
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	stmt := &ast.IfStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	stmt.Consequence = p.parseStatementOrBlock()

	if p.peekToken.Type == token.ELSE {
		p.nextToken()
		stmt.Alternative = p.parseStatementOrBlock()
	}

	return stmt
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	stmt := &ast.ForStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	// Init
	p.nextToken()
	if p.curToken.Type != token.SEMICOLON {
		// Parsing AssignStatement or similar?
		// FILLY for loop: for(i=0; i<10; i=i+1)
		// Init is assignment.
		// My parseAssignStatement starts with IDENT.
		if p.curToken.Type == token.IDENT {
			stmt.Init = p.parseAssignStatement()
		}
		// Expect Semicolon
		if p.peekToken.Type == token.SEMICOLON {
			p.nextToken()
			// Consume semicolon? Loop below expects next is Condition.
		}
	}
	p.nextToken()

	// Condition
	if p.curToken.Type != token.SEMICOLON {
		stmt.Condition = p.parseExpression(LOWEST)
		p.nextToken() // Skip condition end
	}
	// Expect Semicolon
	if p.curToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	// Post
	if p.curToken.Type != token.RPAREN {
		if p.curToken.Type == token.IDENT {
			stmt.Post = p.parseAssignStatement()
		}
		p.nextToken()
	}

	stmt.Body = p.parseStatementOrBlock()

	return stmt
}

// parseStatementOrBlock parses either a block statement {...} or a single statement
func (p *Parser) parseStatementOrBlock() *ast.BlockStatement {
	if p.peekToken.Type == token.LBRACE {
		p.nextToken() // consume token before {
		return p.parseBlockStatement()
	}

	// Single statement - wrap it in a BlockStatement
	p.nextToken() // move to the statement
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	stmt := p.parseStatement()
	if stmt != nil {
		block.Statements = append(block.Statements, stmt)
	}

	return block
}

func (p *Parser) parseFunction() *ast.FunctionStatement {
	// Name(p1, p2=val) { ... }
	stmt := &ast.FunctionStatement{
		Token: p.curToken,
		Name:  &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal},
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}

	stmt.Parameters = []*ast.Parameter{}

	if p.peekToken.Type != token.RPAREN {
		p.nextToken()

		// Parse first param
		param := p.parseParameter()
		stmt.Parameters = append(stmt.Parameters, param)

		for p.peekToken.Type == token.COMMA {
			p.nextToken()
			p.nextToken()
			param := p.parseParameter()
			stmt.Parameters = append(stmt.Parameters, param)
		}
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

func (p *Parser) parseParameter() *ast.Parameter {
	// Format: [Type] Name [Array] [= Default]
	// Type can be INT or IDENT (e.g. str).

	// Format: [Type] Name [Array] [= Default]
	// Type can be INT or IDENT (e.g. str).

	// Check if we have Type Name
	// Cases:
	// 1. int x
	// 2. str x
	// 3. x (implicitly typed?) -> Not observed in CIText, but maybe elsewhere.
	// 4. int x[]

	var name string
	var paramType string = "int" // Default to int

	if p.curToken.Type == token.INT {
		paramType = "int"
		p.nextToken() // Skip 'int'
	} else if p.curToken.Type == token.STR {
		paramType = "string"
		p.nextToken() // Skip 'str'
	} else if p.curToken.Type == token.IDENT && p.peekToken.Type == token.IDENT {
		val := p.curToken.Literal
		if val == "str" {
			val = "string"
		}
		paramType = val
		p.nextToken()
	}

	if p.curToken.Type != token.IDENT {
		return nil
	}
	name = p.curToken.Literal

	if p.peekToken.Type == token.LBRACKET {
		// Array
		p.nextToken() // curToken = [
		if p.peekToken.Type == token.RBRACKET {
			p.nextToken() // curToken = ]
			name += "[]"
			paramType = "[]int" // Assume int array for mixed/array params
		}
	}

	param := &ast.Parameter{
		Name: &ast.Identifier{Token: token.Token{Type: token.IDENT, Literal: name}, Value: name},
		Type: paramType,
	}

	if p.peekToken.Type == token.ASSIGN {
		p.nextToken()
		p.nextToken()
		param.Default = p.parseExpression(LOWEST)
	}

	return param
}

func (p *Parser) skipLine() {
	// Skip tokens until we find a token on a new line?
	// The Lexer stores line numbers. We can loop until p.curToken.Line > parsingLine
	currentLine := p.curToken.Line
	for p.curToken.Type != token.EOF && p.curToken.Line == currentLine {
		p.nextToken()
	}
}

func (p *Parser) parseAssignStatement() *ast.AssignStatement {
	stmt := &ast.AssignStatement{Token: p.curToken}
	stmt.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // move to = or [

	if p.curToken.Type == token.LBRACKET {
		p.nextToken() // consume [
		stmt.Index = p.parseExpression(LOWEST)

		if !p.expectPeek(token.RBRACKET) {
			return nil
		}

		if !p.expectPeek(token.ASSIGN) {
			return nil
		}
		p.nextToken() // move past = to expression start
	} else if p.curToken.Type == token.ASSIGN {
		p.nextToken() // move to Value start
	}

	stmt.Value = p.parseExpression(LOWEST)

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseExpressionStatement() *ast.ExpressionStatement {
	stmt := &ast.ExpressionStatement{Token: p.curToken}
	stmt.Expression = p.parseExpression(LOWEST)

	// SPECIAL CASE: If the expression is just an identifier (no parentheses),
	// and it's one of the special commands that can be called without parentheses,
	// convert it to a function call with no arguments.
	// This handles: del_all, del_me, del_us, end_step, etc.
	if ident, ok := stmt.Expression.(*ast.Identifier); ok {
		identLower := strings.ToLower(ident.Value)
		specialCommands := map[string]bool{
			"del_all":  true,
			"del_me":   true,
			"del_us":   true,
			"end_step": true,
		}
		if specialCommands[identLower] {
			// Convert to function call with no arguments
			stmt.Expression = &ast.CallExpression{
				Token:     ident.Token,
				Function:  ident,
				Arguments: []ast.Expression{},
			}
		}
	}

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseWaitStatement() *ast.WaitStatement {
	stmt := &ast.WaitStatement{Token: p.curToken, Count: 0}

	// Count how many consecutive commas
	for p.curToken.Type == token.COMMA {
		stmt.Count++
		// Warning: This consumes commas. But parseStatement loop calls nextToken().
		// We need to be careful not to consume the NEXT statement's start.

		// If peek is comma, advance.
		// If peek is not comma, stop.
		if p.peekToken.Type == token.COMMA {
			p.nextToken()
		} else {
			break
		}
	}
	return stmt
}

func (p *Parser) parseMesBlockStatement() *ast.MesBlockStatement {
	stmt := &ast.MesBlockStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Time = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	stmt.Body = p.parseStatementOrBlock()

	return stmt
}

func (p *Parser) parseStepBlockStatement() *ast.StepBlockStatement {
	stmt := &ast.StepBlockStatement{Token: p.curToken}

	// Optional argument: step(10), step(), or step { ... }
	if p.peekToken.Type == token.LPAREN {
		p.nextToken() // consume (
		p.nextToken() // move to next token

		// Check if empty parens: step()
		if p.curToken.Type == token.RPAREN {
			stmt.Count = 1 // Default
		} else {
			// Parse the number
			val, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
			if err != nil {
				p.errors = append(p.errors, fmt.Sprintf("msg: %q", err))
				return nil
			}
			stmt.Count = val

			if !p.expectPeek(token.RPAREN) {
				return nil
			}
		}
	} else {
		stmt.Count = 1 // Default
	}

	stmt.Body = p.parseStatementOrBlock()

	return stmt
}

func (p *Parser) parseBlockStatement() *ast.BlockStatement {
	block := &ast.BlockStatement{Token: p.curToken}
	block.Statements = []ast.Statement{}

	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.nextToken()
	}

	return block
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.noPrefixParseFnError(p.curToken.Type)
		return nil
	}
	leftExp := prefix()

	for !p.peekTokenIs(token.SEMICOLON) && precedence < p.peekPrecedence() {
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}

		p.nextToken()

		leftExp = infix(leftExp)
	}

	return leftExp
}

func (p *Parser) peekTokenIs(t token.TokenType) bool {
	return p.peekToken.Type == t
}

func (p *Parser) noPrefixParseFnError(t token.TokenType) {
	msg := fmt.Sprintf("Parser Error: line %d: no prefix parse function for %s found", p.curToken.Line, t)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{Token: p.curToken}

	var value int64
	var err error

	if strings.HasPrefix(p.curToken.Literal, "0x") || strings.HasPrefix(p.curToken.Literal, "0X") {
		value, err = strconv.ParseInt(p.curToken.Literal, 0, 64)
	} else {
		// Treat as decimal even if leading zero
		value, err = strconv.ParseInt(p.curToken.Literal, 10, 64)
	}

	if err != nil {
		msg := fmt.Sprintf("Parser Error: could not parse %q as integer", p.curToken.Literal)
		p.errors = append(p.errors, msg)
		return nil
	}

	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Token: p.curToken, Value: p.curToken.Literal}
}

func (p *Parser) parseCallExpression(function ast.Expression) ast.Expression {
	exp := &ast.CallExpression{Token: p.curToken, Function: function.(*ast.Identifier)} // Assume Identifier for now, or Cast
	// Actually, Filly calls are IDENT(...).
	// parseCallExpression is registered for LPAREN.
	// Infix: Left is Identifier. Token is LPAREN.
	// so function is Left.
	if ident, ok := function.(*ast.Identifier); ok {
		exp.Function = ident
	} else {
		// Could be complex expression? (func())()
		// Filly unlikely supports this.
	}

	exp.Arguments = p.parseCallArguments()
	return exp
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

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	exp := &ast.IndexExpression{Token: p.curToken, Left: left}

	p.nextToken()
	if p.curToken.Type == token.RBRACKET {
		// Empty index []
		exp.Index = nil
		return exp
	}

	exp.Index = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RBRACKET) {
		return nil
	}

	return exp
}

func (p *Parser) curPrecedence() int {
	if p, ok := precedences[p.curToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peekToken.Type]; ok {
		return p
	}
	return LOWEST
}

func (p *Parser) parseCallArguments() []ast.Expression {
	args := []ast.Expression{}

	if p.peekToken.Type == token.RPAREN {
		p.nextToken()
		return args
	}

	p.nextToken()
	args = append(args, p.parseExpression(LOWEST))

	for p.peekToken.Type == token.COMMA {
		p.nextToken()
		p.nextToken()
		args = append(args, p.parseExpression(LOWEST))
	}

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	return args
}

func (p *Parser) parseDefineStatement() *ast.DefineStatement {
	// Current token is "#"
	ds := &ast.DefineStatement{Token: p.curToken}

	p.nextToken() // move to "define"
	// Expect "define"
	if p.curToken.Literal != "define" {
		return nil // Should not happen based on check
	}

	p.nextToken() // move to Name
	if p.curToken.Type != token.IDENT {
		return nil
	}
	ds.Name = &ast.Identifier{Token: p.curToken, Value: p.curToken.Literal}

	p.nextToken() // move to Value

	// Value can be Expression.
	// e.g. #define MAX 100
	ds.Value = p.parseExpression(LOWEST)

	// Directives usually don't end with semicolon, they correspond to line?
	// ParseExpression might consume until it sees something matching?
	// Precedence LOWEST stops at semicolon or something?
	// We expect line end?
	// After parseExpression, we should be at end of line or so.
	// But parseExpression parses `100` and stops.
	// That's fine.

	return ds
}

func (p *Parser) isFunctionDefinition() bool {
	// Snapshot the lexer state to look ahead arbitrarily
	snapshot := p.l.Snapshot()
	defer p.l.Reset(snapshot)

	// Since p.peekToken is already LPAREN, we start reading AFTER it.
	// Lexer is positioned after peekToken.
	// However, we need to be careful. The lexer's current ch is the char AFTER the token just read (peekToken).
	// So calling NextToken() will read the token AFTER peekToken (LPAREN).

	parenCount := 1 // We consider LPAREN as started

	for {
		tok := p.l.NextToken()
		if tok.Type == token.EOF {
			return false
		}
		if tok.Type == token.LPAREN {
			parenCount++
		}
		if tok.Type == token.RPAREN {
			parenCount--
			if parenCount == 0 {
				// Found matching closing parenthesis
				// check next token
				nextToken := p.l.NextToken()
				return nextToken.Type == token.LBRACE
			}
		}
	}
}

func (p *Parser) expectPeek(t token.TokenType) bool {
	if p.peekToken.Type == t {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.TokenType) {
	msg := fmt.Sprintf("Parser Error: line %d: expected next token to be %s, got %s instead", p.curToken.Line, t, p.peekToken.Type)
	p.errors = append(p.errors, msg)
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	stmt := &ast.WhileStatement{Token: p.curToken}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	stmt.Body = p.parseStatementOrBlock()

	return stmt
}

func (p *Parser) parseDoWhileStatement() *ast.DoWhileStatement {
	stmt := &ast.DoWhileStatement{Token: p.curToken}

	stmt.Body = p.parseStatementOrBlock()

	if !p.expectPeek(token.WHILE) {
		return nil
	}

	if !p.expectPeek(token.LPAREN) {
		return nil
	}
	p.nextToken()
	stmt.Condition = p.parseExpression(LOWEST)

	if !p.expectPeek(token.RPAREN) {
		return nil
	}

	if p.peekToken.Type == token.SEMICOLON {
		p.nextToken()
	}

	return stmt
}

func (p *Parser) parseSwitchStatement() *ast.SwitchStatement {
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

	stmt.Cases = []*ast.CaseClause{}

	p.nextToken()

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.CASE {
			caseClause := &ast.CaseClause{Token: p.curToken}
			p.nextToken()
			caseClause.Value = p.parseExpression(LOWEST)

			if !p.expectPeek(token.COLON) {
				return nil
			}

			// Parse statements until we hit another case, default, or }
			caseClause.Body = &ast.BlockStatement{Token: p.curToken}
			caseClause.Body.Statements = []ast.Statement{}

			p.nextToken()

			for p.curToken.Type != token.CASE && p.curToken.Type != token.DEFAULT && p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
				s := p.parseStatement()
				if s != nil {
					caseClause.Body.Statements = append(caseClause.Body.Statements, s)
				}
				p.nextToken()
			}

			stmt.Cases = append(stmt.Cases, caseClause)
			continue
		}

		if p.curToken.Type == token.DEFAULT {
			if !p.expectPeek(token.COLON) {
				return nil
			}

			stmt.Default = &ast.BlockStatement{Token: p.curToken}
			stmt.Default.Statements = []ast.Statement{}

			p.nextToken()

			for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
				s := p.parseStatement()
				if s != nil {
					stmt.Default.Statements = append(stmt.Default.Statements, s)
				}
				p.nextToken()
			}
			break
		}

		p.nextToken()
	}

	return stmt
}
