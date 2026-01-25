// Package parser provides syntax analysis for FILLY scripts (.TFY files).
package parser

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
)

// TestParserNew tests the Parser constructor.
func TestParserNew(t *testing.T) {
	input := "int x;"
	l := lexer.New(input)
	p := New(l)

	if p == nil {
		t.Fatal("New() returned nil")
	}

	if p.lexer != l {
		t.Error("Parser lexer not set correctly")
	}

	if len(p.tokens) == 0 {
		t.Error("Parser tokens not tokenized")
	}

	if p.prefixParseFns == nil {
		t.Error("prefixParseFns not initialized")
	}

	if p.infixParseFns == nil {
		t.Error("infixParseFns not initialized")
	}
}

// TestParserTokenNavigation tests curToken, peekToken, nextToken methods.
func TestParserTokenNavigation(t *testing.T) {
	input := "x + y"
	l := lexer.New(input)
	p := New(l)

	// First token should be 'x'
	if p.curToken().Type != lexer.TOKEN_IDENT {
		t.Errorf("curToken() expected IDENT, got %s", p.curToken().Type)
	}
	if p.curToken().Literal != "x" {
		t.Errorf("curToken().Literal expected 'x', got %s", p.curToken().Literal)
	}

	// Peek should be '+'
	if p.peekToken().Type != lexer.TOKEN_PLUS {
		t.Errorf("peekToken() expected PLUS, got %s", p.peekToken().Type)
	}
}

// TestParseIntegerLiteral tests parsing integer literals.
func TestParseIntegerLiteral(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"42", 42},
		{"0", 0},
		{"123456", 123456},
		{"0x10", 16},
		{"0xFF", 255},
		{"0xABCD", 43981},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program, errs := p.ParseProgram()

		if len(errs) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, errs)
			continue
		}

		if len(program.Statements) != 1 {
			t.Errorf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
			continue
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Errorf("input %q: expected ExpressionStatement, got %T", tt.input, program.Statements[0])
			continue
		}

		lit, ok := stmt.Expression.(*IntegerLiteral)
		if !ok {
			t.Errorf("input %q: expected IntegerLiteral, got %T", tt.input, stmt.Expression)
			continue
		}

		if lit.Value != tt.expected {
			t.Errorf("input %q: expected %d, got %d", tt.input, tt.expected, lit.Value)
		}
	}
}

// TestParseStringLiteral tests parsing string literals.
func TestParseStringLiteral(t *testing.T) {
	input := `"hello world"`
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	lit, ok := stmt.Expression.(*StringLiteral)
	if !ok {
		t.Fatalf("expected StringLiteral, got %T", stmt.Expression)
	}

	if lit.Value != "hello world" {
		t.Errorf("expected 'hello world', got %q", lit.Value)
	}
}

// TestParseBinaryExpression tests parsing binary expressions.
func TestParseBinaryExpression(t *testing.T) {
	tests := []struct {
		input    string
		left     int64
		operator string
		right    int64
	}{
		{"5 + 5", 5, "+", 5},
		{"5 - 5", 5, "-", 5},
		{"5 * 5", 5, "*", 5},
		{"5 / 5", 5, "/", 5},
		{"5 % 5", 5, "%", 5},
		{"5 > 5", 5, ">", 5},
		{"5 < 5", 5, "<", 5},
		{"5 == 5", 5, "==", 5},
		{"5 != 5", 5, "!=", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program, errs := p.ParseProgram()

		if len(errs) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, errs)
			continue
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Errorf("input %q: expected ExpressionStatement, got %T", tt.input, program.Statements[0])
			continue
		}

		exp, ok := stmt.Expression.(*BinaryExpression)
		if !ok {
			t.Errorf("input %q: expected BinaryExpression, got %T", tt.input, stmt.Expression)
			continue
		}

		if exp.Operator != tt.operator {
			t.Errorf("input %q: expected operator %q, got %q", tt.input, tt.operator, exp.Operator)
		}
	}
}

// TestParseUnaryExpression tests parsing unary expressions.
func TestParseUnaryExpression(t *testing.T) {
	tests := []struct {
		input    string
		operator string
		value    int64
	}{
		{"-5", "-", 5},
		{"!5", "!", 5},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program, errs := p.ParseProgram()

		if len(errs) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, errs)
			continue
		}

		stmt, ok := program.Statements[0].(*ExpressionStatement)
		if !ok {
			t.Errorf("input %q: expected ExpressionStatement, got %T", tt.input, program.Statements[0])
			continue
		}

		exp, ok := stmt.Expression.(*UnaryExpression)
		if !ok {
			t.Errorf("input %q: expected UnaryExpression, got %T", tt.input, stmt.Expression)
			continue
		}

		if exp.Operator != tt.operator {
			t.Errorf("input %q: expected operator %q, got %q", tt.input, tt.operator, exp.Operator)
		}
	}
}

// TestOperatorPrecedence tests operator precedence parsing.
func TestOperatorPrecedence(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"1 + 2 * 3", "((1) + ((2) * (3)))"},
		{"1 * 2 + 3", "(((1) * (2)) + (3))"},
		{"1 + 2 + 3", "(((1) + (2)) + (3))"},
		{"-1 * 2", "((-(1)) * (2))"},
		{"1 < 2 == 3 > 4", "(((1) < (2)) == ((3) > (4)))"},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program, errs := p.ParseProgram()

		if len(errs) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, errs)
			continue
		}

		if len(program.Statements) != 1 {
			t.Errorf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
		}
	}
}

// TestParseVarDeclaration tests parsing variable declarations.
// Requirement 3.2: Variable declarations (int x, y[]; str s;) create VarDeclaration nodes
func TestParseVarDeclaration(t *testing.T) {
	tests := []struct {
		input         string
		expectedType  string
		expectedNames []string
	}{
		{"int x;", "int", []string{"x"}},
		{"str s;", "str", []string{"s"}},
		{"int x, y;", "int", []string{"x", "y"}},
		{"int arr[];", "int", []string{"arr"}},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program, errs := p.ParseProgram()

		if len(errs) > 0 {
			t.Errorf("input %q: unexpected errors: %v", tt.input, errs)
			continue
		}

		if len(program.Statements) != 1 {
			t.Errorf("input %q: expected 1 statement, got %d", tt.input, len(program.Statements))
			continue
		}

		stmt, ok := program.Statements[0].(*VarDeclaration)
		if !ok {
			t.Errorf("input %q: expected VarDeclaration, got %T", tt.input, program.Statements[0])
			continue
		}

		if stmt.Type != tt.expectedType {
			t.Errorf("input %q: expected type %q, got %q", tt.input, tt.expectedType, stmt.Type)
		}

		if len(stmt.Names) != len(tt.expectedNames) {
			t.Errorf("input %q: expected %d names, got %d", tt.input, len(tt.expectedNames), len(stmt.Names))
		}
	}
}

// TestParseVarDeclarationComprehensive tests comprehensive variable declaration patterns.
// Requirement 3.2: Variable declarations (int x, y[]; str s;) create VarDeclaration nodes
// Requirement 3.3: Array declarations (int arr[10]) with array flag and size expression
func TestParseVarDeclarationComprehensive(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		expectedType    string
		expectedNames   []string
		expectedIsArray []bool
		expectedSizes   []int64 // -1 means no size specified, >= 0 means size value
	}{
		// Basic single variable declarations
		{
			name:            "single int variable",
			input:           "int x;",
			expectedType:    "int",
			expectedNames:   []string{"x"},
			expectedIsArray: []bool{false},
			expectedSizes:   []int64{-1},
		},
		{
			name:            "single str variable",
			input:           "str s;",
			expectedType:    "str",
			expectedNames:   []string{"s"},
			expectedIsArray: []bool{false},
			expectedSizes:   []int64{-1},
		},
		// Multiple variable declarations
		{
			name:            "multiple int variables",
			input:           "int x, y, z;",
			expectedType:    "int",
			expectedNames:   []string{"x", "y", "z"},
			expectedIsArray: []bool{false, false, false},
			expectedSizes:   []int64{-1, -1, -1},
		},
		{
			name:            "four int variables (like WinW,WinH,WinX,WinY)",
			input:           "int WinW,WinH,WinX,WinY;",
			expectedType:    "int",
			expectedNames:   []string{"WinW", "WinH", "WinX", "WinY"},
			expectedIsArray: []bool{false, false, false, false},
			expectedSizes:   []int64{-1, -1, -1, -1},
		},
		// Array declarations without size
		{
			name:            "single array without size",
			input:           "int arr[];",
			expectedType:    "int",
			expectedNames:   []string{"arr"},
			expectedIsArray: []bool{true},
			expectedSizes:   []int64{-1},
		},
		{
			name:            "multiple arrays without size",
			input:           "int p1[],p2[],c1[],c2[];",
			expectedType:    "int",
			expectedNames:   []string{"p1", "p2", "c1", "c2"},
			expectedIsArray: []bool{true, true, true, true},
			expectedSizes:   []int64{-1, -1, -1, -1},
		},
		// Array declarations with size
		{
			name:            "array with size",
			input:           "int arr[10];",
			expectedType:    "int",
			expectedNames:   []string{"arr"},
			expectedIsArray: []bool{true},
			expectedSizes:   []int64{10},
		},
		{
			name:            "array with larger size",
			input:           "int buffer[256];",
			expectedType:    "int",
			expectedNames:   []string{"buffer"},
			expectedIsArray: []bool{true},
			expectedSizes:   []int64{256},
		},
		// Mixed arrays and scalars (from ROBOT.TFY sample)
		{
			name:            "mixed arrays and scalars (ROBOT.TFY pattern)",
			input:           "int LPic[],BasePic,FieldPic,BirdPic[],OPPic[],Dummy;",
			expectedType:    "int",
			expectedNames:   []string{"LPic", "BasePic", "FieldPic", "BirdPic", "OPPic", "Dummy"},
			expectedIsArray: []bool{true, false, false, true, true, false},
			expectedSizes:   []int64{-1, -1, -1, -1, -1, -1},
		},
		// Without semicolon (FILLY allows optional semicolons)
		{
			name:            "without semicolon",
			input:           "int x",
			expectedType:    "int",
			expectedNames:   []string{"x"},
			expectedIsArray: []bool{false},
			expectedSizes:   []int64{-1},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*VarDeclaration)
			if !ok {
				t.Fatalf("expected VarDeclaration, got %T", program.Statements[0])
			}

			// Check type
			if stmt.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, stmt.Type)
			}

			// Check names count
			if len(stmt.Names) != len(tt.expectedNames) {
				t.Fatalf("expected %d names, got %d: %v", len(tt.expectedNames), len(stmt.Names), stmt.Names)
			}

			// Check each name
			for i, expectedName := range tt.expectedNames {
				if stmt.Names[i] != expectedName {
					t.Errorf("name[%d]: expected %q, got %q", i, expectedName, stmt.Names[i])
				}
			}

			// Check IsArray flags
			if len(stmt.IsArray) != len(tt.expectedIsArray) {
				t.Fatalf("expected %d IsArray flags, got %d", len(tt.expectedIsArray), len(stmt.IsArray))
			}

			for i, expectedIsArray := range tt.expectedIsArray {
				if stmt.IsArray[i] != expectedIsArray {
					t.Errorf("IsArray[%d]: expected %v, got %v", i, expectedIsArray, stmt.IsArray[i])
				}
			}

			// Check Sizes
			if len(stmt.Sizes) != len(tt.expectedSizes) {
				t.Fatalf("expected %d Sizes, got %d", len(tt.expectedSizes), len(stmt.Sizes))
			}

			for i, expectedSize := range tt.expectedSizes {
				if expectedSize == -1 {
					// No size specified
					if stmt.Sizes[i] != nil {
						t.Errorf("Sizes[%d]: expected nil, got %v", i, stmt.Sizes[i])
					}
				} else {
					// Size specified
					if stmt.Sizes[i] == nil {
						t.Errorf("Sizes[%d]: expected %d, got nil", i, expectedSize)
						continue
					}
					intLit, ok := stmt.Sizes[i].(*IntegerLiteral)
					if !ok {
						t.Errorf("Sizes[%d]: expected IntegerLiteral, got %T", i, stmt.Sizes[i])
						continue
					}
					if intLit.Value != expectedSize {
						t.Errorf("Sizes[%d]: expected %d, got %d", i, expectedSize, intLit.Value)
					}
				}
			}
		})
	}
}

// TestParseVarDeclarationWithExpression tests array declarations with expression sizes.
// Requirement 3.3: Array declarations (int arr[10]) with array flag and size expression
func TestParseVarDeclarationWithExpression(t *testing.T) {
	// Test array with expression as size
	input := "int arr[10 + 5];"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*VarDeclaration)
	if !ok {
		t.Fatalf("expected VarDeclaration, got %T", program.Statements[0])
	}

	if len(stmt.Sizes) != 1 || stmt.Sizes[0] == nil {
		t.Fatal("expected size expression")
	}

	// Size should be a BinaryExpression
	binExpr, ok := stmt.Sizes[0].(*BinaryExpression)
	if !ok {
		t.Fatalf("expected BinaryExpression for size, got %T", stmt.Sizes[0])
	}

	if binExpr.Operator != "+" {
		t.Errorf("expected operator '+', got %q", binExpr.Operator)
	}
}

// TestParseVarDeclarationCaseInsensitive tests case-insensitive type keywords.
// Requirement 9.8: Keywords are case-insensitive
func TestParseVarDeclarationCaseInsensitive(t *testing.T) {
	tests := []struct {
		input        string
		expectedType string
	}{
		{"int x;", "int"},
		{"INT x;", "INT"},
		{"Int x;", "Int"},
		{"str s;", "str"},
		{"STR s;", "STR"},
		{"Str s;", "Str"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*VarDeclaration)
			if !ok {
				t.Fatalf("expected VarDeclaration, got %T", program.Statements[0])
			}

			// Type should preserve original case
			if stmt.Type != tt.expectedType {
				t.Errorf("expected type %q, got %q", tt.expectedType, stmt.Type)
			}
		})
	}
}

// TestParseCallExpression tests parsing function call expressions.
func TestParseCallExpression(t *testing.T) {
	input := "add(1, 2 * 3, 4 + 5)"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*CallExpression)
	if !ok {
		t.Fatalf("expected CallExpression, got %T", stmt.Expression)
	}

	if exp.Function != "add" {
		t.Errorf("expected function name 'add', got %q", exp.Function)
	}

	if len(exp.Arguments) != 3 {
		t.Errorf("expected 3 arguments, got %d", len(exp.Arguments))
	}
}

// TestParseIndexExpression tests parsing array index expressions.
func TestParseIndexExpression(t *testing.T) {
	input := "arr[1 + 2]"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
	}

	exp, ok := stmt.Expression.(*IndexExpression)
	if !ok {
		t.Fatalf("expected IndexExpression, got %T", stmt.Expression)
	}

	ident, ok := exp.Left.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier for Left, got %T", exp.Left)
	}

	if ident.Value != "arr" {
		t.Errorf("expected 'arr', got %q", ident.Value)
	}
}

// TestParseAssignStatement tests parsing assignment statements.
func TestParseAssignStatement(t *testing.T) {
	input := "x = 5"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*AssignStatement)
	if !ok {
		t.Fatalf("expected AssignStatement, got %T", program.Statements[0])
	}

	ident, ok := stmt.Name.(*Identifier)
	if !ok {
		t.Fatalf("expected Identifier for Name, got %T", stmt.Name)
	}

	if ident.Value != "x" {
		t.Errorf("expected 'x', got %q", ident.Value)
	}
}

// TestParseIfStatement tests parsing if statements.
func TestParseIfStatement(t *testing.T) {
	input := "if (x > 5) { y = 10; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*IfStatement)
	if !ok {
		t.Fatalf("expected IfStatement, got %T", program.Statements[0])
	}

	if stmt.Condition == nil {
		t.Error("expected Condition, got nil")
	}

	if stmt.Consequence == nil {
		t.Error("expected Consequence, got nil")
	}
}

// TestParseForStatement tests parsing for statements.
func TestParseForStatement(t *testing.T) {
	input := "for (i = 0; i < 10; i = i + 1) { x = i; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*ForStatement)
	if !ok {
		t.Fatalf("expected ForStatement, got %T", program.Statements[0])
	}

	if stmt.Init == nil {
		t.Error("expected Init, got nil")
	}

	if stmt.Condition == nil {
		t.Error("expected Condition, got nil")
	}

	if stmt.Body == nil {
		t.Error("expected Body, got nil")
	}
}

// TestParseWhileStatement tests parsing while statements.
func TestParseWhileStatement(t *testing.T) {
	input := "while (x > 0) { x = x - 1; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*WhileStatement)
	if !ok {
		t.Fatalf("expected WhileStatement, got %T", program.Statements[0])
	}

	if stmt.Condition == nil {
		t.Error("expected Condition, got nil")
	}

	if stmt.Body == nil {
		t.Error("expected Body, got nil")
	}
}

// TestParseFunctionDefinition tests parsing function definitions.
// Requirement 3.4: Function definitions (name(params){body}) create FunctionStatement nodes
// Requirement 9.9: Function definitions without 'function' keyword
func TestParseFunctionDefinition(t *testing.T) {
	input := "myFunc(int x, y) { return x + y; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if stmt.Name != "myFunc" {
		t.Errorf("expected function name 'myFunc', got %q", stmt.Name)
	}

	if len(stmt.Parameters) != 2 {
		t.Errorf("expected 2 parameters, got %d", len(stmt.Parameters))
	}

	if stmt.Body == nil {
		t.Error("expected Body, got nil")
	}
}

// TestParseFunctionDefinitionComprehensive tests comprehensive function definition patterns.
// Requirement 3.4: Function definitions (name(params){body}) create FunctionStatement nodes
// Requirement 9.6: Parameters with default values (int time=1)
// Requirement 9.7: Array parameters (int arr[])
// Requirement 9.9: Function definitions without 'function' keyword
// Requirement 9.10: Keywords as identifiers in expression context
func TestParseFunctionDefinitionComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedName   string
		expectedParams []struct {
			name         string
			paramType    string
			isArray      bool
			hasDefault   bool
			defaultValue int64 // only for integer defaults
		}
		expectedBodyStmts int
	}{
		{
			name:         "no parameters - main()",
			input:        "main() { x = 1; }",
			expectedName: "main",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{},
			expectedBodyStmts: 1,
		},
		{
			name:         "no parameters - start()",
			input:        "start() { y = 2; }",
			expectedName: "start",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{},
			expectedBodyStmts: 1,
		},
		{
			name:         "single typed parameter",
			input:        "myFunc(int x) { return x; }",
			expectedName: "myFunc",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "x", paramType: "int", isArray: false, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "single untyped parameter",
			input:        "myFunc(x) { return x; }",
			expectedName: "myFunc",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "x", paramType: "", isArray: false, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "multiple typed parameters",
			input:        "add(int x, int y) { return x + y; }",
			expectedName: "add",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "x", paramType: "int", isArray: false, hasDefault: false},
				{name: "y", paramType: "int", isArray: false, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "mixed typed and untyped parameters",
			input:        "myFunc(int x, y) { return x + y; }",
			expectedName: "myFunc",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "x", paramType: "int", isArray: false, hasDefault: false},
				{name: "y", paramType: "", isArray: false, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "array parameter without type",
			input:        "process(arr[]) { return arr[0]; }",
			expectedName: "process",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "arr", paramType: "", isArray: true, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "array parameter with type (Requirement 9.7)",
			input:        "process(int arr[]) { return arr[0]; }",
			expectedName: "process",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "arr", paramType: "int", isArray: true, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "parameter with default value (Requirement 9.6)",
			input:        "myFunc(x=10) { return x; }",
			expectedName: "myFunc",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "x", paramType: "", isArray: false, hasDefault: true, defaultValue: 10},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "typed parameter with default value (Requirement 9.6, 9.10)",
			input:        "myFunc(int time=1) { return time; }",
			expectedName: "myFunc",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "time", paramType: "int", isArray: false, hasDefault: true, defaultValue: 1},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "complex parameters from ROBOT.TFY - OP_walk",
			input:        "OP_walk(c, p[], x, y, w, h, l=10) { int d; }",
			expectedName: "OP_walk",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "c", paramType: "", isArray: false, hasDefault: false},
				{name: "p", paramType: "", isArray: true, hasDefault: false},
				{name: "x", paramType: "", isArray: false, hasDefault: false},
				{name: "y", paramType: "", isArray: false, hasDefault: false},
				{name: "w", paramType: "", isArray: false, hasDefault: false},
				{name: "h", paramType: "", isArray: false, hasDefault: false},
				{name: "l", paramType: "", isArray: false, hasDefault: true, defaultValue: 10},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "multiple array parameters",
			input:        "process(int p1[], p2[], c1[], c2[]) { return 0; }",
			expectedName: "process",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "p1", paramType: "int", isArray: true, hasDefault: false},
				{name: "p2", paramType: "", isArray: true, hasDefault: false},
				{name: "c1", paramType: "", isArray: true, hasDefault: false},
				{name: "c2", paramType: "", isArray: true, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "string type parameter",
			input:        "greet(str name) { return name; }",
			expectedName: "greet",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{
				{name: "name", paramType: "str", isArray: false, hasDefault: false},
			},
			expectedBodyStmts: 1,
		},
		{
			name:         "empty body",
			input:        "noop() { }",
			expectedName: "noop",
			expectedParams: []struct {
				name, paramType     string
				isArray, hasDefault bool
				defaultValue        int64
			}{},
			expectedBodyStmts: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*FunctionStatement)
			if !ok {
				t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
			}

			// Check function name
			if stmt.Name != tt.expectedName {
				t.Errorf("expected function name %q, got %q", tt.expectedName, stmt.Name)
			}

			// Check parameter count
			if len(stmt.Parameters) != len(tt.expectedParams) {
				t.Fatalf("expected %d parameters, got %d", len(tt.expectedParams), len(stmt.Parameters))
			}

			// Check each parameter
			for i, expected := range tt.expectedParams {
				param := stmt.Parameters[i]

				if param.Name != expected.name {
					t.Errorf("param[%d].Name: expected %q, got %q", i, expected.name, param.Name)
				}

				if param.Type != expected.paramType {
					t.Errorf("param[%d].Type: expected %q, got %q", i, expected.paramType, param.Type)
				}

				if param.IsArray != expected.isArray {
					t.Errorf("param[%d].IsArray: expected %v, got %v", i, expected.isArray, param.IsArray)
				}

				if expected.hasDefault {
					if param.DefaultValue == nil {
						t.Errorf("param[%d].DefaultValue: expected value, got nil", i)
					} else {
						intLit, ok := param.DefaultValue.(*IntegerLiteral)
						if !ok {
							t.Errorf("param[%d].DefaultValue: expected IntegerLiteral, got %T", i, param.DefaultValue)
						} else if intLit.Value != expected.defaultValue {
							t.Errorf("param[%d].DefaultValue: expected %d, got %d", i, expected.defaultValue, intLit.Value)
						}
					}
				} else {
					if param.DefaultValue != nil {
						t.Errorf("param[%d].DefaultValue: expected nil, got %v", i, param.DefaultValue)
					}
				}
			}

			// Check body
			if stmt.Body == nil {
				t.Fatal("expected Body, got nil")
			}

			if len(stmt.Body.Statements) != tt.expectedBodyStmts {
				t.Errorf("expected %d body statements, got %d", tt.expectedBodyStmts, len(stmt.Body.Statements))
			}
		})
	}
}

// TestParseFunctionDefinitionVsCallExpression tests that function definitions
// are correctly distinguished from function calls.
func TestParseFunctionDefinitionVsCallExpression(t *testing.T) {
	// Function definition: identifier followed by ( params ) {
	defInput := "myFunc(x) { return x; }"
	l := lexer.New(defInput)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("function definition: unexpected errors: %v", errs)
	}

	_, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Errorf("expected FunctionStatement for definition, got %T", program.Statements[0])
	}

	// Function call: identifier followed by ( args ) ;
	callInput := "myFunc(x);"
	l = lexer.New(callInput)
	p = New(l)
	program, errs = p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("function call: unexpected errors: %v", errs)
	}

	exprStmt, ok := program.Statements[0].(*ExpressionStatement)
	if !ok {
		t.Fatalf("expected ExpressionStatement for call, got %T", program.Statements[0])
	}

	_, ok = exprStmt.Expression.(*CallExpression)
	if !ok {
		t.Errorf("expected CallExpression, got %T", exprStmt.Expression)
	}
}

// TestParseFunctionDefinitionWithNestedBlocks tests function definitions with nested control structures.
func TestParseFunctionDefinitionWithNestedBlocks(t *testing.T) {
	input := `OP_walk(c, p[], x, y, w, h, l=10) {
		int d;
		d = 0;
		if (l != 0) {
			if (l < 0) {
				l = -l;
				d = 1;
			}
		}
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if stmt.Name != "OP_walk" {
		t.Errorf("expected function name 'OP_walk', got %q", stmt.Name)
	}

	if len(stmt.Parameters) != 7 {
		t.Errorf("expected 7 parameters, got %d", len(stmt.Parameters))
	}

	// Check that body contains statements
	if stmt.Body == nil || len(stmt.Body.Statements) == 0 {
		t.Error("expected non-empty body")
	}
}

// TestParseFunctionDefinitionFromSample tests function definitions from ROBOT.TFY sample.
func TestParseFunctionDefinitionFromSample(t *testing.T) {
	// Test main() function pattern
	mainInput := `main(){
		CapTitle("");
		WinW=WinInfo(0);
	}`

	l := lexer.New(mainInput)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("main(): unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if stmt.Name != "main" {
		t.Errorf("expected 'main', got %q", stmt.Name)
	}

	if len(stmt.Parameters) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(stmt.Parameters))
	}

	// Test BIRDON() function pattern
	birdonInput := `BIRDON(){
		SetFont(640,"test",128);
	}`

	l = lexer.New(birdonInput)
	p = New(l)
	program, errs = p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("BIRDON(): unexpected errors: %v", errs)
	}

	stmt, ok = program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if stmt.Name != "BIRDON" {
		t.Errorf("expected 'BIRDON', got %q", stmt.Name)
	}
}

// TestParseMesStatement tests parsing mes statements.
func TestParseMesStatement(t *testing.T) {
	input := "mes(MIDI_TIME) { step { func(); } }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if stmt.EventType != "MIDI_TIME" {
		t.Errorf("expected event type 'MIDI_TIME', got %q", stmt.EventType)
	}

	if stmt.Body == nil {
		t.Error("expected Body, got nil")
	}
}

// TestParseMesStatementComprehensive tests parsing mes statements with all event types.
// Requirement 3.11: mes(EVENT) blocks with event type and body
// Requirement 9.1: EVENT types: TIME, MIDI_TIME, MIDI_END, KEY, CLICK, RBDOWN, RBDBLCLK, USER
func TestParseMesStatementComprehensive(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedEventType string
		expectedBodyStmts int
	}{
		// All valid event types from Requirement 9.1
		{
			name:              "TIME event",
			input:             "mes(TIME) { x = 1; }",
			expectedEventType: "TIME",
			expectedBodyStmts: 1,
		},
		{
			name:              "MIDI_TIME event",
			input:             "mes(MIDI_TIME) { x = 1; }",
			expectedEventType: "MIDI_TIME",
			expectedBodyStmts: 1,
		},
		{
			name:              "MIDI_END event",
			input:             "mes(MIDI_END) { x = 1; }",
			expectedEventType: "MIDI_END",
			expectedBodyStmts: 1,
		},
		{
			name:              "KEY event",
			input:             "mes(KEY) { x = 1; }",
			expectedEventType: "KEY",
			expectedBodyStmts: 1,
		},
		{
			name:              "CLICK event",
			input:             "mes(CLICK) { x = 1; }",
			expectedEventType: "CLICK",
			expectedBodyStmts: 1,
		},
		{
			name:              "RBDOWN event",
			input:             "mes(RBDOWN) { x = 1; }",
			expectedEventType: "RBDOWN",
			expectedBodyStmts: 1,
		},
		{
			name:              "RBDBLCLK event",
			input:             "mes(RBDBLCLK) { x = 1; }",
			expectedEventType: "RBDBLCLK",
			expectedBodyStmts: 1,
		},
		{
			name:              "USER event",
			input:             "mes(USER) { x = 1; }",
			expectedEventType: "USER",
			expectedBodyStmts: 1,
		},
		// Case variations (event types are identifiers, case-sensitive)
		{
			name:              "lowercase time event",
			input:             "mes(time) { x = 1; }",
			expectedEventType: "time",
			expectedBodyStmts: 1,
		},
		{
			name:              "mixed case Time event",
			input:             "mes(Time) { x = 1; }",
			expectedEventType: "Time",
			expectedBodyStmts: 1,
		},
		// Empty body
		{
			name:              "empty body",
			input:             "mes(TIME) { }",
			expectedEventType: "TIME",
			expectedBodyStmts: 0,
		},
		// Multiple statements in body
		{
			name:              "multiple statements in body",
			input:             "mes(TIME) { x = 1; y = 2; z = 3; }",
			expectedEventType: "TIME",
			expectedBodyStmts: 3,
		},
		// Nested step block (common pattern from ROBOT.TFY)
		{
			name:              "with step block",
			input:             "mes(TIME) { step(20) { func(); } }",
			expectedEventType: "TIME",
			expectedBodyStmts: 1,
		},
		// Complex body with control structures
		{
			name:              "with if statement",
			input:             "mes(KEY) { if (x > 0) { y = 1; } }",
			expectedEventType: "KEY",
			expectedBodyStmts: 1,
		},
		// Real pattern from ROBOT.TFY: mes(TIME){step(20){,start();end_step;del_me;}}
		{
			name:              "ROBOT.TFY main pattern",
			input:             "mes(TIME) { step(20) { start(); } }",
			expectedEventType: "TIME",
			expectedBodyStmts: 1,
		},
		// Nested mes blocks (from ROBOT.TFY start() function)
		{
			name:              "nested mes blocks",
			input:             "mes(MIDI_TIME) { mes(MIDI_TIME) { x = 1; } }",
			expectedEventType: "MIDI_TIME",
			expectedBodyStmts: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*MesStatement)
			if !ok {
				t.Fatalf("expected MesStatement, got %T", program.Statements[0])
			}

			// Check event type
			if stmt.EventType != tt.expectedEventType {
				t.Errorf("expected event type %q, got %q", tt.expectedEventType, stmt.EventType)
			}

			// Check body exists
			if stmt.Body == nil {
				t.Fatal("expected Body, got nil")
			}

			// Check body statement count
			if len(stmt.Body.Statements) != tt.expectedBodyStmts {
				t.Errorf("expected %d body statements, got %d", tt.expectedBodyStmts, len(stmt.Body.Statements))
			}
		})
	}
}

// TestParseMesStatementCaseInsensitiveKeyword tests that the 'mes' keyword is case-insensitive.
// Requirement 9.8: Keywords are case-insensitive (MES, mes, Mes all work)
func TestParseMesStatementCaseInsensitiveKeyword(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase mes", "mes(TIME) { x = 1; }"},
		{"uppercase MES", "MES(TIME) { x = 1; }"},
		{"mixed case Mes", "Mes(TIME) { x = 1; }"},
		{"mixed case mEs", "mEs(TIME) { x = 1; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*MesStatement)
			if !ok {
				t.Fatalf("expected MesStatement, got %T", program.Statements[0])
			}

			if stmt.EventType != "TIME" {
				t.Errorf("expected event type 'TIME', got %q", stmt.EventType)
			}
		})
	}
}

// TestParseMesStatementWithDelMe tests mes statements with del_me (common pattern).
// This pattern is used in ROBOT.TFY to delete the event handler after execution.
func TestParseMesStatementWithDelMe(t *testing.T) {
	input := `mes(TIME) {
		step(20) {
			start();
		}
		del_me;
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if stmt.EventType != "TIME" {
		t.Errorf("expected event type 'TIME', got %q", stmt.EventType)
	}

	// Body should have 2 statements: step block and del_me
	if len(stmt.Body.Statements) != 2 {
		t.Errorf("expected 2 body statements, got %d", len(stmt.Body.Statements))
	}
}

// TestParseMesStatementMIDIEndPattern tests the MIDI_END event pattern from ROBOT.TFY.
// mes(MIDI_END) is triggered when MIDI playback ends.
func TestParseMesStatementMIDIEndPattern(t *testing.T) {
	input := `mes(MIDI_END) {
		mes(TIME) {
			step(5) {
				message();
			}
		}
		del_me;
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if stmt.EventType != "MIDI_END" {
		t.Errorf("expected event type 'MIDI_END', got %q", stmt.EventType)
	}

	// Body should have 2 statements: nested mes(TIME) and del_me
	if len(stmt.Body.Statements) != 2 {
		t.Errorf("expected 2 body statements, got %d", len(stmt.Body.Statements))
	}

	// First statement should be a nested MesStatement
	nestedMes, ok := stmt.Body.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected nested MesStatement, got %T", stmt.Body.Statements[0])
	}

	if nestedMes.EventType != "TIME" {
		t.Errorf("expected nested event type 'TIME', got %q", nestedMes.EventType)
	}
}

// TestParseMesStatementInFunction tests mes statements inside function definitions.
// This is the common pattern in ROBOT.TFY where mes blocks are defined inside functions.
func TestParseMesStatementInFunction(t *testing.T) {
	input := `start() {
		mes(MIDI_TIME) {
			step {
				BIRD();
			}
		}
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	funcStmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if funcStmt.Name != "start" {
		t.Errorf("expected function name 'start', got %q", funcStmt.Name)
	}

	// Function body should have 1 statement (mes block)
	if len(funcStmt.Body.Statements) != 1 {
		t.Fatalf("expected 1 body statement, got %d", len(funcStmt.Body.Statements))
	}

	mesStmt, ok := funcStmt.Body.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", funcStmt.Body.Statements[0])
	}

	if mesStmt.EventType != "MIDI_TIME" {
		t.Errorf("expected event type 'MIDI_TIME', got %q", mesStmt.EventType)
	}
}

// TestParseMesStatementMultipleInFunction tests multiple mes statements in a function.
// From ROBOT.TFY start() function which has multiple mes blocks.
func TestParseMesStatementMultipleInFunction(t *testing.T) {
	input := `start() {
		mes(MIDI_TIME) { x = 1; }
		mes(TIME) { y = 2; }
		mes(MIDI_END) { z = 3; }
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	funcStmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	// Function body should have 3 mes statements
	if len(funcStmt.Body.Statements) != 3 {
		t.Fatalf("expected 3 body statements, got %d", len(funcStmt.Body.Statements))
	}

	expectedEventTypes := []string{"MIDI_TIME", "TIME", "MIDI_END"}
	for i, expectedType := range expectedEventTypes {
		mesStmt, ok := funcStmt.Body.Statements[i].(*MesStatement)
		if !ok {
			t.Fatalf("statement %d: expected MesStatement, got %T", i, funcStmt.Body.Statements[i])
		}
		if mesStmt.EventType != expectedType {
			t.Errorf("statement %d: expected event type %q, got %q", i, expectedType, mesStmt.EventType)
		}
	}
}

// TestParseStepStatement tests parsing step statements.
func TestParseStepStatement(t *testing.T) {
	input := "step(10) { func1();, func2();,, }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", program.Statements[0])
	}

	if stmt.Count == nil {
		t.Error("expected Count, got nil")
	}

	if stmt.Body == nil {
		t.Error("expected Body, got nil")
	}
}

// TestParseFunctionDefinitionWithLocalVarDeclarations tests function definitions
// with local variable declarations inside the body (from ROBOT.TFY start() pattern).
func TestParseFunctionDefinitionWithLocalVarDeclarations(t *testing.T) {
	input := `start(){
		int p1[],p2[],c1[],c2[];
		x = 1;
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if stmt.Name != "start" {
		t.Errorf("expected 'start', got %q", stmt.Name)
	}

	if len(stmt.Parameters) != 0 {
		t.Errorf("expected 0 parameters, got %d", len(stmt.Parameters))
	}

	// Check body has 2 statements: var declaration and assignment
	if stmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	if len(stmt.Body.Statements) != 2 {
		t.Fatalf("expected 2 body statements, got %d", len(stmt.Body.Statements))
	}

	// First statement should be VarDeclaration
	varDecl, ok := stmt.Body.Statements[0].(*VarDeclaration)
	if !ok {
		t.Fatalf("expected VarDeclaration, got %T", stmt.Body.Statements[0])
	}

	if len(varDecl.Names) != 4 {
		t.Errorf("expected 4 variable names, got %d", len(varDecl.Names))
	}

	// All should be arrays
	for i, isArray := range varDecl.IsArray {
		if !isArray {
			t.Errorf("expected varDecl.IsArray[%d] to be true", i)
		}
	}
}

// TestParseFunctionDefinitionEdgeCases tests edge cases for function definitions.
func TestParseFunctionDefinitionEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectError  bool
		expectedName string
	}{
		{
			name:         "function with underscore in name",
			input:        "my_func() { }",
			expectError:  false,
			expectedName: "my_func",
		},
		{
			name:         "function with numbers in name",
			input:        "func123() { }",
			expectError:  false,
			expectedName: "func123",
		},
		{
			name:         "function with uppercase name",
			input:        "BIRDON() { }",
			expectError:  false,
			expectedName: "BIRDON",
		},
		{
			name:         "function with mixed case name",
			input:        "OP_walk() { }",
			expectError:  false,
			expectedName: "OP_walk",
		},
		{
			name:         "function with single character name",
			input:        "f() { }",
			expectError:  false,
			expectedName: "f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if tt.expectError {
				if len(errs) == 0 {
					t.Error("expected errors, got none")
				}
				return
			}

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*FunctionStatement)
			if !ok {
				t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
			}

			if stmt.Name != tt.expectedName {
				t.Errorf("expected function name %q, got %q", tt.expectedName, stmt.Name)
			}
		})
	}
}

// TestParseParameterWithExpressionDefault tests parameters with expression default values.
func TestParseParameterWithExpressionDefault(t *testing.T) {
	// Test with negative default value
	input := "myFunc(x=-5) { return x; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if len(stmt.Parameters) != 1 {
		t.Fatalf("expected 1 parameter, got %d", len(stmt.Parameters))
	}

	param := stmt.Parameters[0]
	if param.Name != "x" {
		t.Errorf("expected parameter name 'x', got %q", param.Name)
	}

	if param.DefaultValue == nil {
		t.Fatal("expected default value, got nil")
	}

	// Default value should be a UnaryExpression (-5)
	unary, ok := param.DefaultValue.(*UnaryExpression)
	if !ok {
		t.Fatalf("expected UnaryExpression, got %T", param.DefaultValue)
	}

	if unary.Operator != "-" {
		t.Errorf("expected operator '-', got %q", unary.Operator)
	}
}

// ============================================================================
// Expression Parsing Tests (Task 3.5)
// Tests for operator precedence, function calls, array access, and grouped expressions
// Validates Requirements 3.13 and 3.14
// ============================================================================

// TestOperatorPrecedenceComprehensive tests all operator precedence levels from design doc.
// Requirement 3.14: Expressions respect operator precedence.
//
// Precedence levels (from lowest to highest):
//
//	LOWEST
//	OR          (||)
//	AND         (&&)
//	EQUALS      (==, !=)
//	LESSGREATER (<, >, <=, >=)
//	SUM         (+, -)
//	PRODUCT     (*, /, %)
//	PREFIX      (-x, !x)
//	CALL        (func(x))
//	INDEX       (array[index])
func TestOperatorPrecedenceComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string // Expected AST structure as string
	}{
		// Basic arithmetic precedence
		{
			name:     "multiplication before addition",
			input:    "1 + 2 * 3",
			expected: "(1 + (2 * 3))",
		},
		{
			name:     "division before subtraction",
			input:    "6 - 4 / 2",
			expected: "(6 - (4 / 2))",
		},
		{
			name:     "modulo same as multiplication",
			input:    "10 % 3 * 2",
			expected: "((10 % 3) * 2)",
		},
		// Comparison operators
		{
			name:     "comparison lower than arithmetic",
			input:    "1 + 2 < 3 + 4",
			expected: "((1 + 2) < (3 + 4))",
		},
		{
			name:     "equality lower than comparison",
			input:    "1 < 2 == 3 > 4",
			expected: "((1 < 2) == (3 > 4))",
		},
		{
			name:     "not equal same as equal",
			input:    "a == b != c",
			expected: "((a == b) != c)",
		},
		// Logical operators
		{
			name:     "AND lower than equality",
			input:    "a == b && c == d",
			expected: "((a == b) && (c == d))",
		},
		{
			name:     "OR lower than AND",
			input:    "a && b || c && d",
			expected: "((a && b) || (c && d))",
		},
		{
			name:     "complex logical expression",
			input:    "a || b && c || d",
			expected: "((a || (b && c)) || d)",
		},
		// Prefix operators
		{
			name:     "prefix minus high precedence",
			input:    "-1 * 2",
			expected: "((-1) * 2)",
		},
		{
			name:     "prefix not high precedence",
			input:    "!a && b",
			expected: "((!a) && b)",
		},
		{
			name:     "double prefix",
			input:    "--5",
			expected: "(-(-5))",
		},
		{
			name:     "not not",
			input:    "!!true",
			expected: "(!(!true))",
		},
		// Left-to-right associativity
		{
			name:     "left associativity addition",
			input:    "1 + 2 + 3",
			expected: "((1 + 2) + 3)",
		},
		{
			name:     "left associativity subtraction",
			input:    "10 - 5 - 2",
			expected: "((10 - 5) - 2)",
		},
		{
			name:     "left associativity multiplication",
			input:    "2 * 3 * 4",
			expected: "((2 * 3) * 4)",
		},
		// Mixed precedence
		{
			name:     "complex mixed expression",
			input:    "1 + 2 * 3 - 4 / 2",
			expected: "((1 + (2 * 3)) - (4 / 2))",
		},
		{
			name:     "comparison with arithmetic",
			input:    "a + b > c - d",
			expected: "((a + b) > (c - d))",
		},
		{
			name:     "logical with comparison and arithmetic",
			input:    "a + 1 > b && c - 1 < d",
			expected: "(((a + 1) > b) && ((c - 1) < d))",
		},
		// All comparison operators
		{
			name:     "less than",
			input:    "a < b",
			expected: "(a < b)",
		},
		{
			name:     "greater than",
			input:    "a > b",
			expected: "(a > b)",
		},
		{
			name:     "less than or equal",
			input:    "a <= b",
			expected: "(a <= b)",
		},
		{
			name:     "greater than or equal",
			input:    "a >= b",
			expected: "(a >= b)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			actual := expressionToString(stmt.Expression)
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestGroupedExpressions tests parentheses affecting precedence.
// Requirement 3.14: Expressions respect operator precedence (including grouping).
func TestGroupedExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple grouping",
			input:    "(1 + 2) * 3",
			expected: "((1 + 2) * 3)",
		},
		{
			name:     "nested grouping",
			input:    "((1 + 2) * 3)",
			expected: "((1 + 2) * 3)",
		},
		{
			name:     "grouping on right",
			input:    "1 * (2 + 3)",
			expected: "(1 * (2 + 3))",
		},
		{
			name:     "multiple groups",
			input:    "(1 + 2) * (3 + 4)",
			expected: "((1 + 2) * (3 + 4))",
		},
		{
			name:     "deeply nested",
			input:    "((1 + 2) * (3 + 4)) / 5",
			expected: "(((1 + 2) * (3 + 4)) / 5)",
		},
		{
			name:     "grouping with logical operators",
			input:    "(a || b) && (c || d)",
			expected: "((a || b) && (c || d))",
		},
		{
			name:     "grouping overrides precedence",
			input:    "(a + b) * (c - d)",
			expected: "((a + b) * (c - d))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			actual := expressionToString(stmt.Expression)
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestCallExpressionComprehensive tests function call parsing.
// Requirement 3.13: Function calls with function name and arguments.
func TestCallExpressionComprehensive(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		expectedFunc string
		expectedArgs int
	}{
		{
			name:         "no arguments",
			input:        "func()",
			expectedFunc: "func",
			expectedArgs: 0,
		},
		{
			name:         "single integer argument",
			input:        "func(1)",
			expectedFunc: "func",
			expectedArgs: 1,
		},
		{
			name:         "single string argument",
			input:        `func("hello")`,
			expectedFunc: "func",
			expectedArgs: 1,
		},
		{
			name:         "single identifier argument",
			input:        "func(x)",
			expectedFunc: "func",
			expectedArgs: 1,
		},
		{
			name:         "multiple arguments",
			input:        "func(1, 2, 3)",
			expectedFunc: "func",
			expectedArgs: 3,
		},
		{
			name:         "expression arguments",
			input:        "func(1 + 2, 3 * 4)",
			expectedFunc: "func",
			expectedArgs: 2,
		},
		{
			name:         "nested function calls",
			input:        "outer(inner(x))",
			expectedFunc: "outer",
			expectedArgs: 1,
		},
		{
			name:         "mixed argument types",
			input:        `func(1, "str", x, arr[0])`,
			expectedFunc: "func",
			expectedArgs: 4,
		},
		{
			name:         "function call with array access argument",
			input:        "func(arr[i])",
			expectedFunc: "func",
			expectedArgs: 1,
		},
		{
			name:         "function call with complex expression",
			input:        "func(a + b * c)",
			expectedFunc: "func",
			expectedArgs: 1,
		},
		// Real-world examples from FILLY
		{
			name:         "LoadPic with string",
			input:        `LoadPic("image.bmp")`,
			expectedFunc: "LoadPic",
			expectedArgs: 1,
		},
		{
			name:         "MovePic with multiple args",
			input:        "MovePic(src, 0, 0, 100, 100, dst, 0, 0)",
			expectedFunc: "MovePic",
			expectedArgs: 8,
		},
		{
			name:         "SetFont with mixed args",
			input:        `SetFont(640, "test", 128)`,
			expectedFunc: "SetFont",
			expectedArgs: 3,
		},
		{
			name:         "WinInfo with single arg",
			input:        "WinInfo(0)",
			expectedFunc: "WinInfo",
			expectedArgs: 1,
		},
		{
			name:         "CapTitle with empty string",
			input:        `CapTitle("")`,
			expectedFunc: "CapTitle",
			expectedArgs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			call, ok := stmt.Expression.(*CallExpression)
			if !ok {
				t.Fatalf("expected CallExpression, got %T", stmt.Expression)
			}

			if call.Function != tt.expectedFunc {
				t.Errorf("expected function %q, got %q", tt.expectedFunc, call.Function)
			}

			if len(call.Arguments) != tt.expectedArgs {
				t.Errorf("expected %d arguments, got %d", tt.expectedArgs, len(call.Arguments))
			}
		})
	}
}

// TestIndexExpressionComprehensive tests array access parsing.
// Requirement 3.6: Array access with IndexExpression.
func TestIndexExpressionComprehensive(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedArray string
		indexType     string // "int", "ident", "expr"
	}{
		{
			name:          "integer index",
			input:         "arr[0]",
			expectedArray: "arr",
			indexType:     "int",
		},
		{
			name:          "identifier index",
			input:         "arr[i]",
			expectedArray: "arr",
			indexType:     "ident",
		},
		{
			name:          "expression index",
			input:         "arr[i + 1]",
			expectedArray: "arr",
			indexType:     "expr",
		},
		{
			name:          "complex expression index",
			input:         "arr[i * 2 + j]",
			expectedArray: "arr",
			indexType:     "expr",
		},
		{
			name:          "function call as index",
			input:         "arr[getIndex()]",
			expectedArray: "arr",
			indexType:     "expr",
		},
		{
			name:          "nested array access",
			input:         "matrix[i][j]",
			expectedArray: "matrix",
			indexType:     "ident",
		},
		{
			name:          "array access with negative index expression",
			input:         "arr[len - 1]",
			expectedArray: "arr",
			indexType:     "expr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			// For nested array access, the outer expression is IndexExpression
			idx, ok := stmt.Expression.(*IndexExpression)
			if !ok {
				t.Fatalf("expected IndexExpression, got %T", stmt.Expression)
			}

			// Check the array name (may be nested)
			var arrayName string
			switch left := idx.Left.(type) {
			case *Identifier:
				arrayName = left.Value
			case *IndexExpression:
				// Nested array access - get the innermost identifier
				inner, ok := left.Left.(*Identifier)
				if ok {
					arrayName = inner.Value
				}
			}

			if arrayName != tt.expectedArray {
				t.Errorf("expected array %q, got %q", tt.expectedArray, arrayName)
			}

			// Verify index is not nil
			if idx.Index == nil {
				t.Error("expected index expression, got nil")
			}
		})
	}
}

// TestCallAndIndexPrecedence tests that CALL and INDEX have highest precedence.
// Requirement 3.14: Expressions respect operator precedence.
func TestCallAndIndexPrecedence(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "call higher than arithmetic",
			input:    "1 + func(2)",
			expected: "(1 + func(2))",
		},
		{
			name:     "index higher than arithmetic",
			input:    "1 + arr[2]",
			expected: "(1 + arr[2])",
		},
		{
			name:     "call in multiplication",
			input:    "func(2) * 3",
			expected: "(func(2) * 3)",
		},
		{
			name:     "index in multiplication",
			input:    "arr[0] * 3",
			expected: "(arr[0] * 3)",
		},
		{
			name:     "call and index combined",
			input:    "func(arr[i])",
			expected: "func(arr[i])",
		},
		{
			name:     "arithmetic with call and index",
			input:    "arr[i] + func(x)",
			expected: "(arr[i] + func(x))",
		},
		{
			name:     "comparison with call",
			input:    "func(x) > func(y)",
			expected: "(func(x) > func(y))",
		},
		{
			name:     "logical with index",
			input:    "arr[0] && arr[1]",
			expected: "(arr[0] && arr[1])",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			actual := expressionToString(stmt.Expression)
			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestComplexExpressions tests complex expressions combining multiple features.
// Requirement 3.13, 3.14: Function calls and operator precedence.
func TestComplexExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "function call with arithmetic argument",
			input:    "func(a + b * c)",
			expected: "func((a + (b * c)))",
		},
		{
			name:     "array access with arithmetic index",
			input:    "arr[i + j * 2]",
			expected: "arr[(i + (j * 2))]",
		},
		{
			name:     "nested function calls with arithmetic",
			input:    "outer(inner(x) + 1)",
			expected: "outer((inner(x) + 1))",
		},
		{
			name:     "comparison with function calls",
			input:    "getX() < getY()",
			expected: "(getX() < getY())",
		},
		{
			name:     "logical expression with array access",
			input:    "arr[0] > 0 && arr[1] < 10",
			expected: "((arr[0] > 0) && (arr[1] < 10))",
		},
		{
			name:     "assignment target with index",
			input:    "arr[i] = x + 1",
			expected: "arr[i] = (x + 1)", // This is an assignment, not expression
		},
		{
			name:     "complex FILLY-like expression",
			input:    "WinW - x * 2 + offset",
			expected: "((WinW - (x * 2)) + offset)",
		},
		{
			name:     "multiple function calls in expression",
			input:    "getWidth() + getHeight() * 2",
			expected: "(getWidth() + (getHeight() * 2))",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			var actual string
			switch stmt := program.Statements[0].(type) {
			case *ExpressionStatement:
				actual = expressionToString(stmt.Expression)
			case *AssignStatement:
				actual = expressionToString(stmt.Name) + " = " + expressionToString(stmt.Value)
			default:
				t.Fatalf("unexpected statement type: %T", program.Statements[0])
			}

			if actual != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestFloatLiteralParsing tests parsing of floating point literals.
// Requirement 2.5: Floating point literals.
func TestFloatLiteralParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"3.14", 3.14},
		{"0.5", 0.5},
		{"100.0", 100.0},
		{"0.001", 0.001},
		{"123.456", 123.456},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*ExpressionStatement)
			if !ok {
				t.Fatalf("expected ExpressionStatement, got %T", program.Statements[0])
			}

			lit, ok := stmt.Expression.(*FloatLiteral)
			if !ok {
				t.Fatalf("expected FloatLiteral, got %T", stmt.Expression)
			}

			if lit.Value != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, lit.Value)
			}
		})
	}
}

// TestExpressionInAssignment tests expressions on the right side of assignments.
// Requirement 3.5: Assignment statements with target and value.
func TestExpressionInAssignment(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedName  string
		expectedValue string
	}{
		{
			name:          "simple assignment",
			input:         "x = 5",
			expectedName:  "x",
			expectedValue: "5",
		},
		{
			name:          "assignment with arithmetic",
			input:         "x = 1 + 2",
			expectedName:  "x",
			expectedValue: "(1 + 2)",
		},
		{
			name:          "assignment with function call",
			input:         "x = func()",
			expectedName:  "x",
			expectedValue: "func()",
		},
		{
			name:          "assignment with array access",
			input:         "x = arr[0]",
			expectedName:  "x",
			expectedValue: "arr[0]",
		},
		{
			name:          "assignment with complex expression",
			input:         "x = a + b * c",
			expectedName:  "x",
			expectedValue: "(a + (b * c))",
		},
		{
			name:          "array assignment",
			input:         "arr[i] = x + 1",
			expectedName:  "arr[i]",
			expectedValue: "(x + 1)",
		},
		{
			name:          "assignment from WinInfo (FILLY pattern)",
			input:         "WinW = WinInfo(0)",
			expectedName:  "WinW",
			expectedValue: "WinInfo(0)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*AssignStatement)
			if !ok {
				t.Fatalf("expected AssignStatement, got %T", program.Statements[0])
			}

			actualName := expressionToString(stmt.Name)
			if actualName != tt.expectedName {
				t.Errorf("expected name %q, got %q", tt.expectedName, actualName)
			}

			actualValue := expressionToString(stmt.Value)
			if actualValue != tt.expectedValue {
				t.Errorf("expected value %q, got %q", tt.expectedValue, actualValue)
			}
		})
	}
}

// expressionToString converts an Expression to a string representation for testing.
// This helper function creates a canonical string representation of the AST.
func expressionToString(exp Expression) string {
	if exp == nil {
		return "nil"
	}

	switch e := exp.(type) {
	case *IntegerLiteral:
		return e.Token.Literal
	case *FloatLiteral:
		return e.Token.Literal
	case *StringLiteral:
		return `"` + e.Value + `"`
	case *Identifier:
		return e.Value
	case *UnaryExpression:
		return "(" + e.Operator + expressionToString(e.Right) + ")"
	case *BinaryExpression:
		return "(" + expressionToString(e.Left) + " " + e.Operator + " " + expressionToString(e.Right) + ")"
	case *CallExpression:
		args := ""
		for i, arg := range e.Arguments {
			if i > 0 {
				args += ", "
			}
			args += expressionToString(arg)
		}
		return e.Function + "(" + args + ")"
	case *IndexExpression:
		return expressionToString(e.Left) + "[" + expressionToString(e.Index) + "]"
	default:
		return "unknown"
	}
}

// TestExpressionParsingFromSamplePatterns tests expression patterns from ROBOT.TFY sample.
func TestExpressionParsingFromSamplePatterns(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "WinInfo call",
			input: "WinW=WinInfo(0);",
		},
		{
			name:  "CapTitle with empty string",
			input: `CapTitle("");`,
		},
		{
			name:  "SetFont with multiple args",
			input: `SetFont(640,"test",128);`,
		},
		{
			name:  "LoadPic with string",
			input: `LoadPic("image.bmp");`,
		},
		{
			name:  "arithmetic expression",
			input: "x = WinW - 100;",
		},
		{
			name:  "comparison in condition",
			input: "if (x > 0) { y = 1; }",
		},
		{
			name:  "logical AND",
			input: "if (x > 0 && y < 10) { z = 1; }",
		},
		{
			name:  "logical OR",
			input: "if (a || b) { c = 1; }",
		},
		{
			name:  "array access in expression",
			input: "x = arr[i] + 1;",
		},
		{
			name:  "function call with expression arg",
			input: "func(x + y);",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) == 0 {
				t.Fatal("expected at least 1 statement")
			}
		})
	}
}

// ============================================================================
// Control Structure Tests (Task 3.6)
// Tests for if, for, while, and switch statements
// Validates Requirements 3.7, 3.8, 3.9, 3.10
// ============================================================================

// TestParseIfStatementComprehensive tests comprehensive if statement patterns.
// Requirement 3.7: If statements with condition, consequence, and optional alternative.
func TestParseIfStatementComprehensive(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasAlternative bool
		isElseIf       bool
	}{
		{
			name:           "simple if without else",
			input:          "if (x > 5) { y = 10; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with else",
			input:          "if (x > 5) { y = 10; } else { y = 0; }",
			hasAlternative: true,
			isElseIf:       false,
		},
		{
			name:           "if with else if",
			input:          "if (x > 5) { y = 10; } else if (x > 0) { y = 5; }",
			hasAlternative: true,
			isElseIf:       true,
		},
		{
			name:           "if with else if and else",
			input:          "if (x > 5) { y = 10; } else if (x > 0) { y = 5; } else { y = 0; }",
			hasAlternative: true,
			isElseIf:       true,
		},
		{
			name:           "if with complex condition",
			input:          "if (x > 5 && y < 10) { z = 1; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with OR condition",
			input:          "if (a || b) { c = 1; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with equality check",
			input:          "if (x == 0) { y = 1; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with inequality check",
			input:          "if (x != 0) { y = 1; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with function call in condition",
			input:          "if (isValid()) { process(); }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with array access in condition",
			input:          "if (arr[i] > 0) { sum = sum + arr[i]; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "nested if statements",
			input:          "if (x > 0) { if (y > 0) { z = 1; } }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "if with multiple statements in body",
			input:          "if (x > 0) { a = 1; b = 2; c = 3; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		// FILLY-specific patterns from ROBOT.TFY
		{
			name:           "FILLY pattern: l != 0 check",
			input:          "if (l != 0) { d = 1; }",
			hasAlternative: false,
			isElseIf:       false,
		},
		{
			name:           "FILLY pattern: l < 0 with negation",
			input:          "if (l < 0) { l = -l; d = 1; }",
			hasAlternative: false,
			isElseIf:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*IfStatement)
			if !ok {
				t.Fatalf("expected IfStatement, got %T", program.Statements[0])
			}

			if stmt.Condition == nil {
				t.Error("expected Condition, got nil")
			}

			if stmt.Consequence == nil {
				t.Error("expected Consequence, got nil")
			}

			if tt.hasAlternative {
				if stmt.Alternative == nil {
					t.Error("expected Alternative, got nil")
				}
				if tt.isElseIf {
					_, ok := stmt.Alternative.(*IfStatement)
					if !ok {
						t.Errorf("expected else if (IfStatement), got %T", stmt.Alternative)
					}
				} else {
					_, ok := stmt.Alternative.(*BlockStatement)
					if !ok {
						t.Errorf("expected else block (BlockStatement), got %T", stmt.Alternative)
					}
				}
			} else {
				if stmt.Alternative != nil {
					t.Errorf("expected no Alternative, got %T", stmt.Alternative)
				}
			}
		})
	}
}

// TestParseForStatementComprehensive tests comprehensive for loop patterns.
// Requirement 3.8: For loops with init, condition, post, and body.
func TestParseForStatementComprehensive(t *testing.T) {
	tests := []struct {
		name         string
		input        string
		hasInit      bool
		hasCondition bool
		hasPost      bool
	}{
		{
			name:         "standard for loop",
			input:        "for (i = 0; i < 10; i = i + 1) { x = i; }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
		{
			name:         "for loop without init",
			input:        "for (; i < 10; i = i + 1) { x = i; }",
			hasInit:      false,
			hasCondition: true,
			hasPost:      true,
		},
		{
			name:         "for loop without post",
			input:        "for (i = 0; i < 10;) { i = i + 1; }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      false,
		},
		{
			name:         "for loop with only condition",
			input:        "for (; i < 10;) { i = i + 1; }",
			hasInit:      false,
			hasCondition: true,
			hasPost:      false,
		},
		{
			name:         "for loop with complex condition",
			input:        "for (i = 0; i < n && arr[i] != 0; i = i + 1) { sum = sum + arr[i]; }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
		{
			name:         "for loop with decrement",
			input:        "for (i = 10; i > 0; i = i - 1) { x = i; }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
		{
			name:         "for loop with function call in condition",
			input:        "for (i = 0; i < getLength(); i = i + 1) { process(i); }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
		{
			name:         "nested for loops",
			input:        "for (i = 0; i < 10; i = i + 1) { for (j = 0; j < 10; j = j + 1) { x = i + j; } }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
		{
			name:         "for loop with multiple statements in body",
			input:        "for (i = 0; i < 10; i = i + 1) { a = i; b = i * 2; c = i * 3; }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
		// FILLY-specific patterns
		{
			name:         "FILLY pattern: array iteration",
			input:        "for (i = 0; i < len; i = i + 1) { arr[i] = 0; }",
			hasInit:      true,
			hasCondition: true,
			hasPost:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ForStatement)
			if !ok {
				t.Fatalf("expected ForStatement, got %T", program.Statements[0])
			}

			if tt.hasInit {
				if stmt.Init == nil {
					t.Error("expected Init, got nil")
				}
			} else {
				if stmt.Init != nil {
					t.Errorf("expected no Init, got %T", stmt.Init)
				}
			}

			if tt.hasCondition {
				if stmt.Condition == nil {
					t.Error("expected Condition, got nil")
				}
			} else {
				if stmt.Condition != nil {
					t.Errorf("expected no Condition, got %v", stmt.Condition)
				}
			}

			if tt.hasPost {
				if stmt.Post == nil {
					t.Error("expected Post, got nil")
				}
			} else {
				if stmt.Post != nil {
					t.Errorf("expected no Post, got %T", stmt.Post)
				}
			}

			if stmt.Body == nil {
				t.Error("expected Body, got nil")
			}
		})
	}
}

// TestParseWhileStatementComprehensive tests comprehensive while loop patterns.
// Requirement 3.9: While loops with condition and body.
func TestParseWhileStatementComprehensive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "simple while loop",
			input: "while (x > 0) { x = x - 1; }",
		},
		{
			name:  "while with equality condition",
			input: "while (x != 0) { x = x - 1; }",
		},
		{
			name:  "while with complex condition",
			input: "while (x > 0 && y < 10) { x = x - 1; y = y + 1; }",
		},
		{
			name:  "while with OR condition",
			input: "while (a || b) { process(); }",
		},
		{
			name:  "while with function call in condition",
			input: "while (hasMore()) { process(); }",
		},
		{
			name:  "while with array access in condition",
			input: "while (arr[i] != 0) { i = i + 1; }",
		},
		{
			name:  "nested while loops",
			input: "while (x > 0) { while (y > 0) { y = y - 1; } x = x - 1; }",
		},
		{
			name:  "while with multiple statements in body",
			input: "while (running) { update(); render(); wait(); }",
		},
		{
			name:  "while true pattern (identifier as condition)",
			input: "while (running) { process(); }",
		},
		// FILLY-specific patterns
		{
			name:  "FILLY pattern: counter decrement",
			input: "while (count > 0) { doSomething(); count = count - 1; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*WhileStatement)
			if !ok {
				t.Fatalf("expected WhileStatement, got %T", program.Statements[0])
			}

			if stmt.Condition == nil {
				t.Error("expected Condition, got nil")
			}

			if stmt.Body == nil {
				t.Error("expected Body, got nil")
			}
		})
	}
}

// TestParseSwitchStatement tests parsing switch statements.
// Requirement 3.10: Switch statements with value, case clauses, and optional default.
func TestParseSwitchStatement(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedCases int
		hasDefault    bool
	}{
		{
			name:          "simple switch with one case",
			input:         "switch (x) { case 1: y = 1; }",
			expectedCases: 1,
			hasDefault:    false,
		},
		{
			name:          "switch with multiple cases",
			input:         "switch (x) { case 1: y = 1; case 2: y = 2; case 3: y = 3; }",
			expectedCases: 3,
			hasDefault:    false,
		},
		{
			name:          "switch with default",
			input:         "switch (x) { case 1: y = 1; default: y = 0; }",
			expectedCases: 1,
			hasDefault:    true,
		},
		{
			name:          "switch with multiple cases and default",
			input:         "switch (x) { case 1: y = 1; case 2: y = 2; default: y = 0; }",
			expectedCases: 2,
			hasDefault:    true,
		},
		{
			name:          "switch with expression value",
			input:         "switch (x + 1) { case 1: y = 1; }",
			expectedCases: 1,
			hasDefault:    false,
		},
		{
			name:          "switch with function call value",
			input:         "switch (getValue()) { case 1: y = 1; case 2: y = 2; }",
			expectedCases: 2,
			hasDefault:    false,
		},
		{
			name:          "switch with break statements",
			input:         "switch (x) { case 1: y = 1; break; case 2: y = 2; break; }",
			expectedCases: 2,
			hasDefault:    false,
		},
		{
			name:          "switch with multiple statements per case",
			input:         "switch (x) { case 1: a = 1; b = 2; c = 3; }",
			expectedCases: 1,
			hasDefault:    false,
		},
		{
			name:          "switch with only default",
			input:         "switch (x) { default: y = 0; }",
			expectedCases: 0,
			hasDefault:    true,
		},
		// FILLY-specific patterns
		{
			name:          "FILLY pattern: direction switch",
			input:         "switch (dir) { case 0: x = x + 1; case 1: y = y + 1; case 2: x = x - 1; case 3: y = y - 1; }",
			expectedCases: 4,
			hasDefault:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*SwitchStatement)
			if !ok {
				t.Fatalf("expected SwitchStatement, got %T", program.Statements[0])
			}

			if stmt.Value == nil {
				t.Error("expected Value, got nil")
			}

			if len(stmt.Cases) != tt.expectedCases {
				t.Errorf("expected %d cases, got %d", tt.expectedCases, len(stmt.Cases))
			}

			if tt.hasDefault {
				if stmt.Default == nil {
					t.Error("expected Default, got nil")
				}
			} else {
				if stmt.Default != nil {
					t.Errorf("expected no Default, got %T", stmt.Default)
				}
			}
		})
	}
}

// TestParseSwitchStatementCaseValues tests that case values are parsed correctly.
// Requirement 3.10: Switch statements with value, case clauses, and optional default.
func TestParseSwitchStatementCaseValues(t *testing.T) {
	input := "switch (x) { case 1: a = 1; case 2: b = 2; case 10: c = 10; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*SwitchStatement)
	if !ok {
		t.Fatalf("expected SwitchStatement, got %T", program.Statements[0])
	}

	expectedValues := []int64{1, 2, 10}
	for i, caseClause := range stmt.Cases {
		intLit, ok := caseClause.Value.(*IntegerLiteral)
		if !ok {
			t.Errorf("case[%d]: expected IntegerLiteral, got %T", i, caseClause.Value)
			continue
		}
		if intLit.Value != expectedValues[i] {
			t.Errorf("case[%d]: expected value %d, got %d", i, expectedValues[i], intLit.Value)
		}
	}
}

// TestParseBreakStatement tests parsing break statements.
// Requirement 4.9: Break statements generate OpBreak instruction.
func TestParseBreakStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "break in switch",
			input: "switch (x) { case 1: y = 1; break; }",
		},
		{
			name:  "break in for loop",
			input: "for (i = 0; i < 10; i = i + 1) { if (i == 5) { break; } }",
		},
		{
			name:  "break in while loop",
			input: "while (x > 0) { if (done) { break; } x = x - 1; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) == 0 {
				t.Fatal("expected at least 1 statement")
			}
		})
	}
}

// TestParseContinueStatement tests parsing continue statements.
// Requirement 4.10: Continue statements generate OpContinue instruction.
func TestParseContinueStatement(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "continue in for loop",
			input: "for (i = 0; i < 10; i = i + 1) { if (i == 5) { continue; } process(i); }",
		},
		{
			name:  "continue in while loop",
			input: "while (x > 0) { x = x - 1; if (skip) { continue; } process(); }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) == 0 {
				t.Fatal("expected at least 1 statement")
			}
		})
	}
}

// TestParseReturnStatement tests parsing return statements.
// Requirement 4.11: Return statements generate OpCall("return") instruction.
func TestParseReturnStatement(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasReturnValue bool
	}{
		{
			name:           "return without value",
			input:          "return;",
			hasReturnValue: false,
		},
		{
			name:           "return with integer",
			input:          "return 42;",
			hasReturnValue: true,
		},
		{
			name:           "return with expression",
			input:          "return x + y;",
			hasReturnValue: true,
		},
		{
			name:           "return with function call",
			input:          "return getValue();",
			hasReturnValue: true,
		},
		{
			name:           "return with identifier",
			input:          "return result;",
			hasReturnValue: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*ReturnStatement)
			if !ok {
				t.Fatalf("expected ReturnStatement, got %T", program.Statements[0])
			}

			if tt.hasReturnValue {
				if stmt.ReturnValue == nil {
					t.Error("expected ReturnValue, got nil")
				}
			} else {
				if stmt.ReturnValue != nil {
					t.Errorf("expected no ReturnValue, got %T", stmt.ReturnValue)
				}
			}
		})
	}
}

// TestControlStructuresInFunction tests control structures inside function bodies.
// This validates that control structures work correctly in real-world contexts.
func TestControlStructuresInFunction(t *testing.T) {
	input := `OP_walk(c, p[], x, y, w, h, l=10) {
		int d;
		d = 0;
		if (l != 0) {
			if (l < 0) {
				l = -l;
				d = 1;
			}
			for (i = 0; i < l; i = i + 1) {
				x = x + 1;
			}
			while (d > 0) {
				d = d - 1;
			}
		}
		return d;
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	funcStmt, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}

	if funcStmt.Name != "OP_walk" {
		t.Errorf("expected function name 'OP_walk', got %q", funcStmt.Name)
	}

	// Check that body contains expected statements
	if funcStmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Body should have: VarDeclaration, AssignStatement, IfStatement, ReturnStatement
	if len(funcStmt.Body.Statements) < 4 {
		t.Errorf("expected at least 4 body statements, got %d", len(funcStmt.Body.Statements))
	}
}

// TestNestedControlStructures tests deeply nested control structures.
func TestNestedControlStructures(t *testing.T) {
	input := `test() {
		if (a > 0) {
			for (i = 0; i < 10; i = i + 1) {
				while (j > 0) {
					if (k == 0) {
						break;
					}
					j = j - 1;
				}
			}
		}
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(program.Statements) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(program.Statements))
	}

	_, ok := program.Statements[0].(*FunctionStatement)
	if !ok {
		t.Fatalf("expected FunctionStatement, got %T", program.Statements[0])
	}
}

// ============================================================================
// Step Statement Tests (Task 3.8)
// ============================================================================

// TestParseStepStatementComprehensive tests comprehensive step statement patterns.
// Requirement 3.12: step() statements with count and optional body
// Requirement 9.2: Commas in step blocks are interpreted as wait instructions
// Requirement 9.3: Consecutive commas are counted as multiple wait steps
// Requirement 9.4: end_step keyword marks step block end
func TestParseStepStatementComprehensive(t *testing.T) {
	tests := []struct {
		name             string
		input            string
		expectedCount    int64 // -1 means no count specified
		expectedCommands int   // number of StepCommands in body
	}{
		{
			name:             "step with count and simple body",
			input:            "step(10) { func1(); }",
			expectedCount:    10,
			expectedCommands: 1,
		},
		{
			name:             "step without count",
			input:            "step { func1(); }",
			expectedCount:    -1,
			expectedCommands: 1,
		},
		{
			name:             "step with count and comma after statement",
			input:            "step(10) { func1();, }",
			expectedCount:    10,
			expectedCommands: 1,
		},
		{
			name:             "step with multiple statements and commas",
			input:            "step(10) { func1();, func2();, }",
			expectedCount:    10,
			expectedCommands: 2,
		},
		{
			name:             "step with consecutive commas (multiple waits)",
			input:            "step(10) { func1();,, }",
			expectedCount:    10,
			expectedCommands: 1,
		},
		{
			name:             "step with end_step",
			input:            "step(20) { func1(); end_step; }",
			expectedCount:    20,
			expectedCommands: 1,
		},
		{
			name:             "step with end_step and del_me",
			input:            "step(20) { func1(); end_step; del_me; }",
			expectedCount:    20,
			expectedCommands: 2, // func1 and del_me (end_step is marker)
		},
		{
			name:             "step with leading comma (wait-only)",
			input:            "step(10) { , func1(); }",
			expectedCount:    10,
			expectedCommands: 2, // wait-only command and func1
		},
		{
			name:             "step with multiple leading commas",
			input:            "step(10) { ,,, func1(); }",
			expectedCount:    10,
			expectedCommands: 2, // wait-only command (3 commas) and func1
		},
		{
			name:             "step with empty parentheses",
			input:            "step() { func1(); }",
			expectedCount:    -1,
			expectedCommands: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*StepStatement)
			if !ok {
				t.Fatalf("expected StepStatement, got %T", program.Statements[0])
			}

			// Check count
			if tt.expectedCount == -1 {
				if stmt.Count != nil {
					t.Errorf("expected no Count, got %v", stmt.Count)
				}
			} else {
				if stmt.Count == nil {
					t.Errorf("expected Count %d, got nil", tt.expectedCount)
				} else {
					intLit, ok := stmt.Count.(*IntegerLiteral)
					if !ok {
						t.Errorf("expected IntegerLiteral for Count, got %T", stmt.Count)
					} else if intLit.Value != tt.expectedCount {
						t.Errorf("expected Count %d, got %d", tt.expectedCount, intLit.Value)
					}
				}
			}

			// Check body
			if stmt.Body == nil {
				t.Fatal("expected Body, got nil")
			}

			if len(stmt.Body.Commands) != tt.expectedCommands {
				t.Errorf("expected %d commands, got %d", tt.expectedCommands, len(stmt.Body.Commands))
			}
		})
	}
}

// TestParseStepStatementWaitCounts tests that wait counts are correctly parsed.
// Requirement 9.2: Commas in step blocks are interpreted as wait instructions
// Requirement 9.3: Consecutive commas are counted as multiple wait steps
func TestParseStepStatementWaitCounts(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedWaits []int // wait count for each command
	}{
		{
			name:          "single comma after statement",
			input:         "step(10) { func1();, }",
			expectedWaits: []int{1},
		},
		{
			name:          "two commas after statement",
			input:         "step(10) { func1();,, }",
			expectedWaits: []int{2},
		},
		{
			name:          "four commas after statement",
			input:         "step(10) { func1();,,,, }",
			expectedWaits: []int{4},
		},
		{
			name:          "multiple statements with different wait counts",
			input:         "step(10) { func1();, func2();,, func3();,,, }",
			expectedWaits: []int{1, 2, 3},
		},
		{
			name:          "statement without trailing comma",
			input:         "step(10) { func1(); }",
			expectedWaits: []int{0},
		},
		{
			name:          "leading commas (wait-only)",
			input:         "step(10) { ,,, func1(); }",
			expectedWaits: []int{3, 0}, // first is wait-only with 3 commas, second is func1 with 0
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*StepStatement)
			if !ok {
				t.Fatalf("expected StepStatement, got %T", program.Statements[0])
			}

			if stmt.Body == nil {
				t.Fatal("expected Body, got nil")
			}

			if len(stmt.Body.Commands) != len(tt.expectedWaits) {
				t.Fatalf("expected %d commands, got %d", len(tt.expectedWaits), len(stmt.Body.Commands))
			}

			for i, expectedWait := range tt.expectedWaits {
				if stmt.Body.Commands[i].WaitCount != expectedWait {
					t.Errorf("command[%d].WaitCount: expected %d, got %d",
						i, expectedWait, stmt.Body.Commands[i].WaitCount)
				}
			}
		})
	}
}

// TestParseStepStatementFromROBOTSample tests step patterns from ROBOT.TFY sample.
// These are real-world patterns that must work correctly.
func TestParseStepStatementFromROBOTSample(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			name:  "main pattern: step(20){,start();end_step;del_me;}",
			input: "step(20) { , start(); end_step; del_me; }",
		},
		{
			name:  "BIRD pattern with multiple commas",
			input: "step { ,,,, ,,,, ,,,, MoveWin(); end_step; del_me; }",
		},
		{
			name:  "OPENING pattern with comma after statement",
			input: "step { , OPWin=OpenWin(); end_step; del_me; }",
		},
		{
			name:  "OP_walk pattern with conditional end_step",
			input: "step { MoveCast(); x=x-16;, MoveCast(); x=x-16; if(count>l){end_step; del_me;}, }",
		},
		{
			name:  "step without count in mes block",
			input: "step { BIRD();,,,, OPENON();,,,, ,,,, ,, }",
		},
		{
			name:  "step(8) with complex body",
			input: "step(8) { OPENING();, BIRDOFF();, ,, ,, ,, }",
		},
		{
			name:  "step(10) with many commas",
			input: "step(10) { MINI_F();,,,,, ,,,,, ,,,,, ,,,,, ,,,,, }",
		},
		{
			name:  "step(5) with message and CloseWinAll",
			input: "step(5) { ,,,,, ,,,,, message();,,,,, CloseWinAll(); end_step; del_all; del_me; }",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			if len(program.Statements) != 1 {
				t.Fatalf("expected 1 statement, got %d", len(program.Statements))
			}

			stmt, ok := program.Statements[0].(*StepStatement)
			if !ok {
				t.Fatalf("expected StepStatement, got %T", program.Statements[0])
			}

			if stmt.Body == nil {
				t.Error("expected Body, got nil")
			}
		})
	}
}

// TestParseStepStatementEndStep tests end_step handling.
// Requirement 9.4: end_step keyword marks step block end
func TestParseStepStatementEndStep(t *testing.T) {
	// Test that end_step properly terminates step body parsing
	input := "step(20) { func1();, func2(); end_step; del_me; }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", program.Statements[0])
	}

	if stmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Commands should include: func1 (with wait), func2, del_me
	// end_step is a marker and should not create a separate command
	if len(stmt.Body.Commands) < 2 {
		t.Errorf("expected at least 2 commands, got %d", len(stmt.Body.Commands))
	}

	// First command should be func1 with wait count 1
	if stmt.Body.Commands[0].WaitCount != 1 {
		t.Errorf("first command WaitCount: expected 1, got %d", stmt.Body.Commands[0].WaitCount)
	}

	// Check that func1 is a CallExpression
	if stmt.Body.Commands[0].Statement != nil {
		exprStmt, ok := stmt.Body.Commands[0].Statement.(*ExpressionStatement)
		if ok {
			callExpr, ok := exprStmt.Expression.(*CallExpression)
			if ok && callExpr.Function != "func1" {
				t.Errorf("expected func1, got %s", callExpr.Function)
			}
		}
	}
}

// TestParseStepStatementNestedInMes tests step statements nested inside mes blocks.
// This is a common pattern in ROBOT.TFY.
func TestParseStepStatementNestedInMes(t *testing.T) {
	input := `mes(TIME) {
		step(20) {
			, start();
			end_step;
			del_me;
		}
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	mesStmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if mesStmt.EventType != "TIME" {
		t.Errorf("expected EventType 'TIME', got %q", mesStmt.EventType)
	}

	if mesStmt.Body == nil || len(mesStmt.Body.Statements) == 0 {
		t.Fatal("expected mes body with statements")
	}

	// First statement in mes body should be StepStatement
	stepStmt, ok := mesStmt.Body.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement in mes body, got %T", mesStmt.Body.Statements[0])
	}

	// Check step count
	if stepStmt.Count == nil {
		t.Error("expected step Count, got nil")
	} else {
		intLit, ok := stepStmt.Count.(*IntegerLiteral)
		if !ok {
			t.Errorf("expected IntegerLiteral for Count, got %T", stepStmt.Count)
		} else if intLit.Value != 20 {
			t.Errorf("expected Count 20, got %d", intLit.Value)
		}
	}

	if stepStmt.Body == nil {
		t.Error("expected step Body, got nil")
	}
}

// TestParseStepStatementNestedMesInStep tests nested mes blocks inside step.
// Pattern from ROBOT.TFY: mes(MIDI_TIME){step{...mes(MIDI_TIME){step(8){...}}}}
func TestParseStepStatementNestedMesInStep(t *testing.T) {
	input := `mes(MIDI_TIME) {
		step {
			BIRD();,,,,
			mes(MIDI_TIME) {
				step(8) {
					OPENING();,
					end_step;
					del_me;
				}
			}
			end_step;
			del_me;
		}
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	mesStmt, ok := program.Statements[0].(*MesStatement)
	if !ok {
		t.Fatalf("expected MesStatement, got %T", program.Statements[0])
	}

	if mesStmt.EventType != "MIDI_TIME" {
		t.Errorf("expected EventType 'MIDI_TIME', got %q", mesStmt.EventType)
	}

	// Check that the outer step contains nested mes
	if mesStmt.Body == nil || len(mesStmt.Body.Statements) == 0 {
		t.Fatal("expected mes body with statements")
	}

	stepStmt, ok := mesStmt.Body.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", mesStmt.Body.Statements[0])
	}

	// Step without count
	if stepStmt.Count != nil {
		t.Error("expected no Count for outer step")
	}

	if stepStmt.Body == nil {
		t.Fatal("expected step Body, got nil")
	}

	// Should have multiple commands including nested mes
	if len(stepStmt.Body.Commands) < 2 {
		t.Errorf("expected at least 2 commands in step body, got %d", len(stepStmt.Body.Commands))
	}
}

// TestParseStepStatementWaitOnlyCommands tests wait-only commands (leading commas).
// Requirement 9.3: Consecutive commas are counted as multiple wait steps
func TestParseStepStatementWaitOnlyCommands(t *testing.T) {
	input := "step(10) { ,,,,, func1(); }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", program.Statements[0])
	}

	if stmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Should have 2 commands: wait-only (5 commas) and func1
	if len(stmt.Body.Commands) != 2 {
		t.Fatalf("expected 2 commands, got %d", len(stmt.Body.Commands))
	}

	// First command should be wait-only with 5 waits
	firstCmd := stmt.Body.Commands[0]
	if firstCmd.Statement != nil {
		t.Error("expected first command to be wait-only (nil Statement)")
	}
	if firstCmd.WaitCount != 5 {
		t.Errorf("expected first command WaitCount 5, got %d", firstCmd.WaitCount)
	}

	// Second command should be func1 with 0 waits
	secondCmd := stmt.Body.Commands[1]
	if secondCmd.Statement == nil {
		t.Error("expected second command to have Statement")
	}
	if secondCmd.WaitCount != 0 {
		t.Errorf("expected second command WaitCount 0, got %d", secondCmd.WaitCount)
	}
}

// TestParseStepStatementComplexPattern tests complex patterns from ROBOT.TFY.
func TestParseStepStatementComplexPattern(t *testing.T) {
	// Pattern from ROBOT.TFY start() function
	input := `step(10) {
		MINI_F();,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,,
		Credit();,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,, ,,,,,
		BIRDON();,,,,,
		PlayMIDI("JMK021.MID");
		end_step;
		del_me;
	}`

	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", program.Statements[0])
	}

	// Check count is 10
	if stmt.Count == nil {
		t.Fatal("expected Count, got nil")
	}
	intLit, ok := stmt.Count.(*IntegerLiteral)
	if !ok {
		t.Fatalf("expected IntegerLiteral for Count, got %T", stmt.Count)
	}
	if intLit.Value != 10 {
		t.Errorf("expected Count 10, got %d", intLit.Value)
	}

	if stmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Should have multiple commands
	if len(stmt.Body.Commands) < 4 {
		t.Errorf("expected at least 4 commands, got %d", len(stmt.Body.Commands))
	}
}

// TestParseStepStatementWithAssignment tests step with assignment statements.
func TestParseStepStatementWithAssignment(t *testing.T) {
	input := "step { MoveCast(); x=x-16;, MoveCast(); x=x-16;, }"
	l := lexer.New(input)
	p := New(l)
	program, errs := p.ParseProgram()

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	stmt, ok := program.Statements[0].(*StepStatement)
	if !ok {
		t.Fatalf("expected StepStatement, got %T", program.Statements[0])
	}

	if stmt.Body == nil {
		t.Fatal("expected Body, got nil")
	}

	// Should have commands for MoveCast and assignments
	if len(stmt.Body.Commands) < 2 {
		t.Errorf("expected at least 2 commands, got %d", len(stmt.Body.Commands))
	}
}

// TestParseStepStatementCaseInsensitive tests case-insensitive step keyword.
// Requirement 9.8: Keywords are case-insensitive
func TestParseStepStatementCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase step", "step(10) { func(); }"},
		{"uppercase STEP", "STEP(10) { func(); }"},
		{"mixed case Step", "Step(10) { func(); }"},
		{"mixed case sTeP", "sTeP(10) { func(); }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			_, ok := program.Statements[0].(*StepStatement)
			if !ok {
				t.Fatalf("expected StepStatement, got %T", program.Statements[0])
			}
		})
	}
}

// TestParseStepStatementEndStepCaseInsensitive tests case-insensitive end_step keyword.
// Requirement 9.8: Keywords are case-insensitive
func TestParseStepStatementEndStepCaseInsensitive(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"lowercase end_step", "step(10) { func(); end_step; }"},
		{"uppercase END_STEP", "step(10) { func(); END_STEP; }"},
		{"mixed case End_Step", "step(10) { func(); End_Step; }"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := New(l)
			program, errs := p.ParseProgram()

			if len(errs) > 0 {
				t.Fatalf("unexpected errors: %v", errs)
			}

			stmt, ok := program.Statements[0].(*StepStatement)
			if !ok {
				t.Fatalf("expected StepStatement, got %T", program.Statements[0])
			}

			if stmt.Body == nil {
				t.Error("expected Body, got nil")
			}
		})
	}
}
