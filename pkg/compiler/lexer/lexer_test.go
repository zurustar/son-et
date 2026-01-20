package lexer

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/token"
)

func TestLexer_BasicTokens(t *testing.T) {
	input := `= + - * / %`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.ASSIGN, "="},
		{token.PLUS, "+"},
		{token.MINUS, "-"},
		{token.MULT, "*"},
		{token.DIV, "/"},
		{token.MOD, "%"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_Comparison(t *testing.T) {
	input := `== != < <= > >=`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.EQ, "=="},
		{token.NEQ, "!="},
		{token.LT, "<"},
		{token.LTE, "<="},
		{token.GT, ">"},
		{token.GTE, ">="},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_Logical(t *testing.T) {
	input := `&& || !`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.AND, "&&"},
		{token.OR, "||"},
		{token.NOT, "!"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}
	}
}

func TestLexer_Delimiters(t *testing.T) {
	input := `( ) { } [ ] , ;`

	tests := []token.TokenType{
		token.LPAREN,
		token.RPAREN,
		token.LBRACE,
		token.RBRACE,
		token.LBRACKET,
		token.RBRACKET,
		token.COMMA,
		token.SEMICOLON,
		token.EOF,
	}

	l := New(input)

	for i, expectedType := range tests {
		tok := l.NextToken()

		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexer_Keywords(t *testing.T) {
	input := `if else for while switch case default break continue return function mes step`

	tests := []token.TokenType{
		token.IF,
		token.ELSE,
		token.FOR,
		token.WHILE,
		token.SWITCH,
		token.CASE,
		token.DEFAULT,
		token.BREAK,
		token.CONTINUE,
		token.RETURN,
		token.FUNCTION,
		token.MES,
		token.STEP,
		token.EOF,
	}

	l := New(input)

	for i, expectedType := range tests {
		tok := l.NextToken()

		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexer_CaseInsensitiveKeywords(t *testing.T) {
	input := `IF If iF FUNCTION Function MES Mes`

	tests := []token.TokenType{
		token.IF,
		token.IF,
		token.IF,
		token.FUNCTION,
		token.FUNCTION,
		token.MES,
		token.MES,
		token.EOF,
	}

	l := New(input)

	for i, expectedType := range tests {
		tok := l.NextToken()

		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexer_Identifiers(t *testing.T) {
	input := `x myVar test123 _private`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.IDENT, "myVar"},
		{token.IDENT, "test123"},
		{token.IDENT, "_private"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_Numbers(t *testing.T) {
	input := `123 456 3.14 0.5`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.INT_LIT, "123"},
		{token.INT_LIT, "456"},
		{token.FLOAT_LIT, "3.14"},
		{token.FLOAT_LIT, "0.5"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_Strings(t *testing.T) {
	input := `"hello" "world" "test 123"`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.STRING, "hello"},
		{token.STRING, "world"},
		{token.STRING, "test 123"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_Comments(t *testing.T) {
	input := `x = 5 // this is a comment
y = 10`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "x"},
		{token.ASSIGN, "="},
		{token.INT, "5"},
		{token.COMMENT, "// this is a comment"},
		{token.IDENT, "y"},
		{token.ASSIGN, "="},
		{token.INT, "10"},
		{token.EOF, ""},
	}

	l := New(input)

	for i, tt := range tests {
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, tt.expectedType, tok.Type)
		}

		if tok.Literal != tt.expectedLiteral {
			t.Fatalf("tests[%d] - literal wrong. expected=%q, got=%q",
				i, tt.expectedLiteral, tok.Literal)
		}
	}
}

func TestLexer_LineAndColumn(t *testing.T) {
	input := `x = 5
y = 10`

	l := New(input)

	// x at line 1, column 1
	tok := l.NextToken()
	if tok.Line != 1 || tok.Column != 1 {
		t.Errorf("Expected x at line 1, column 1, got line %d, column %d", tok.Line, tok.Column)
	}

	// = at line 1, column 3
	tok = l.NextToken()
	if tok.Line != 1 || tok.Column != 3 {
		t.Errorf("Expected = at line 1, column 3, got line %d, column %d", tok.Line, tok.Column)
	}

	// 5 at line 1, column 5
	tok = l.NextToken()
	if tok.Line != 1 || tok.Column != 5 {
		t.Errorf("Expected 5 at line 1, column 5, got line %d, column %d", tok.Line, tok.Column)
	}

	// y at line 2, column 1
	tok = l.NextToken()
	if tok.Line != 2 || tok.Column != 1 {
		t.Errorf("Expected y at line 2, column 1, got line %d, column %d", tok.Line, tok.Column)
	}
}

func TestLexer_ComplexExpression(t *testing.T) {
	input := `if (x > 5 && y < 10) {
    z = x + y * 2
}`

	l := New(input)

	expectedTypes := []token.TokenType{
		token.IF,
		token.LPAREN,
		token.IDENT,
		token.GT,
		token.INT,
		token.AND,
		token.IDENT,
		token.LT,
		token.INT,
		token.RPAREN,
		token.LBRACE,
		token.IDENT,
		token.ASSIGN,
		token.IDENT,
		token.PLUS,
		token.IDENT,
		token.MULT,
		token.INT,
		token.RBRACE,
		token.EOF,
	}

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexer_ArraySyntax(t *testing.T) {
	input := `arr[0] = 5
x = arr[i + 1]`

	l := New(input)

	expectedTypes := []token.TokenType{
		token.IDENT,    // arr
		token.LBRACKET, // [
		token.INT,      // 0
		token.RBRACKET, // ]
		token.ASSIGN,   // =
		token.INT,      // 5
		token.IDENT,    // x
		token.ASSIGN,   // =
		token.IDENT,    // arr
		token.LBRACKET, // [
		token.IDENT,    // i
		token.PLUS,     // +
		token.INT,      // 1
		token.RBRACKET, // ]
		token.EOF,
	}

	for i, expectedType := range expectedTypes {
		tok := l.NextToken()
		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}

func TestLexer_MesEventTypes(t *testing.T) {
	input := `TIME MIDI_TIME MIDI_END KEY CLICK RBDOWN RBDBLCLK USER`

	tests := []token.TokenType{
		token.TIME,
		token.MIDI_TIME,
		token.MIDI_END,
		token.KEY,
		token.CLICK,
		token.RBDOWN,
		token.RBDBLCLK,
		token.USER,
		token.EOF,
	}

	l := New(input)

	for i, expectedType := range tests {
		tok := l.NextToken()

		if tok.Type != expectedType {
			t.Fatalf("tests[%d] - tokentype wrong. expected=%q, got=%q",
				i, expectedType, tok.Type)
		}
	}
}
