package lexer

import (
	"testing"
)

// TestLexerBasicStructure tests the basic Lexer structure and initialization.
func TestLexerBasicStructure(t *testing.T) {
	input := "hello"
	l := New(input)

	if l.input != input {
		t.Errorf("Lexer input wrong. expected=%q, got=%q", input, l.input)
	}
	if l.line != 1 {
		t.Errorf("Lexer line wrong. expected=1, got=%d", l.line)
	}
	if l.column != 1 {
		t.Errorf("Lexer column wrong. expected=1, got=%d", l.column)
	}
	if l.ch != 'h' {
		t.Errorf("Lexer ch wrong. expected='h', got=%q", l.ch)
	}
}

// TestSkipWhitespace tests that whitespace is properly skipped.
// Validates Requirement 2.11: Lexer skips whitespace without creating tokens.
func TestSkipWhitespace(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"  hello", "hello"},
		{"\t\thello", "hello"},
		{"\n\nhello", "hello"},
		{"\r\nhello", "hello"},
		{"  \t\n\r  hello", "hello"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
		if tok.Type != TOKEN_IDENT {
			t.Errorf("input=%q: expected type=TOKEN_IDENT, got=%v", tt.input, tok.Type)
		}
	}
}

// TestReadIdentifier tests identifier reading.
// Validates Requirement 2.3: Identifiers are returned as IDENT tokens.
func TestReadIdentifier(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"hello", "hello"},
		{"_private", "_private"},
		{"var123", "var123"},
		{"IMG01", "IMG01"},
		{"LoadPic", "LoadPic"},
		{"MIDI_TIME", "MIDI_TIME"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
		if tok.Type != TOKEN_IDENT {
			t.Errorf("input=%q: expected type=TOKEN_IDENT, got=%v", tt.input, tok.Type)
		}
	}
}

// TestKeywordRecognition tests that keywords are properly recognized.
// Validates Requirement 2.2: Keywords are identified case-insensitively.
func TestKeywordRecognition(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
	}{
		// Lowercase keywords
		{"int", TOKEN_INT_TYPE},
		{"str", TOKEN_STR_TYPE},
		{"if", TOKEN_IF},
		{"else", TOKEN_ELSE},
		{"for", TOKEN_FOR},
		{"while", TOKEN_WHILE},
		{"switch", TOKEN_SWITCH},
		{"case", TOKEN_CASE},
		{"default", TOKEN_DEFAULT},
		{"break", TOKEN_BREAK},
		{"continue", TOKEN_CONTINUE},
		{"return", TOKEN_RETURN},
		{"mes", TOKEN_MES},
		{"step", TOKEN_STEP},
		{"end_step", TOKEN_END_STEP},
		{"del_me", TOKEN_DEL_ME},
		{"del_us", TOKEN_DEL_US},
		{"del_all", TOKEN_DEL_ALL},

		// Uppercase keywords (case-insensitive)
		{"INT", TOKEN_INT_TYPE},
		{"STR", TOKEN_STR_TYPE},
		{"IF", TOKEN_IF},
		{"ELSE", TOKEN_ELSE},
		{"FOR", TOKEN_FOR},
		{"WHILE", TOKEN_WHILE},
		{"MES", TOKEN_MES},
		{"STEP", TOKEN_STEP},

		// Mixed case keywords
		{"Int", TOKEN_INT_TYPE},
		{"Mes", TOKEN_MES},
		{"Step", TOKEN_STEP},
		{"End_Step", TOKEN_END_STEP},
		{"Del_Me", TOKEN_DEL_ME},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("input=%q: expected type=%v, got=%v", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.input {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.input, tok.Literal)
		}
	}
}

// TestEOFToken tests that EOF token is returned at end of input.
// Validates Requirement 2.13: EOF token is created at end of input.
func TestEOFToken(t *testing.T) {
	l := New("")
	tok := l.NextToken()

	if tok.Type != TOKEN_EOF {
		t.Errorf("expected TOKEN_EOF, got=%v", tok.Type)
	}
}

// TestIllegalToken tests that illegal characters return ILLEGAL tokens.
// Validates Requirement 2.12: Illegal characters create ILLEGAL tokens.
func TestIllegalToken(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"@", "@"},
		// Note: "#" is now a directive prefix, not illegal
		{"$", "$"},
		{"~", "~"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != TOKEN_ILLEGAL {
			t.Errorf("input=%q: expected TOKEN_ILLEGAL, got=%v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
	}
}

// TestLineAndColumnTracking tests that line and column numbers are tracked correctly.
// Validates Requirement 2.14: All tokens include line and column numbers.
func TestLineAndColumnTracking(t *testing.T) {
	input := "hello\nworld"
	l := New(input)

	// First token: "hello" at line 1, column 1
	tok1 := l.NextToken()
	if tok1.Line != 1 {
		t.Errorf("token1 line: expected=1, got=%d", tok1.Line)
	}
	if tok1.Column != 1 {
		t.Errorf("token1 column: expected=1, got=%d", tok1.Column)
	}

	// Second token: "world" at line 2, column 1
	tok2 := l.NextToken()
	if tok2.Line != 2 {
		t.Errorf("token2 line: expected=2, got=%d", tok2.Line)
	}
	if tok2.Column != 1 {
		t.Errorf("token2 column: expected=1, got=%d", tok2.Column)
	}
}

// TestMultipleIdentifiers tests tokenizing multiple identifiers.
func TestMultipleIdentifiers(t *testing.T) {
	input := "hello world foo bar"
	l := New(input)

	expected := []string{"hello", "world", "foo", "bar"}

	for _, exp := range expected {
		tok := l.NextToken()
		if tok.Type != TOKEN_IDENT {
			t.Errorf("expected TOKEN_IDENT, got=%v", tok.Type)
		}
		if tok.Literal != exp {
			t.Errorf("expected literal=%q, got=%q", exp, tok.Literal)
		}
	}

	// Should get EOF at the end
	tok := l.NextToken()
	if tok.Type != TOKEN_EOF {
		t.Errorf("expected TOKEN_EOF, got=%v", tok.Type)
	}
}

// TestTokenize tests the Tokenize method.
func TestTokenize(t *testing.T) {
	input := "int x str y"
	l := New(input)

	tokens, err := l.Tokenize()
	if err != nil {
		t.Fatalf("Tokenize returned error: %v", err)
	}

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_INT_TYPE, "int"},
		{TOKEN_IDENT, "x"},
		{TOKEN_STR_TYPE, "str"},
		{TOKEN_IDENT, "y"},
		{TOKEN_EOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.tokenType {
			t.Errorf("tokens[%d] type: expected=%v, got=%v", i, exp.tokenType, tokens[i].Type)
		}
		if tokens[i].Literal != exp.literal {
			t.Errorf("tokens[%d] literal: expected=%q, got=%q", i, exp.literal, tokens[i].Literal)
		}
	}
}

// TestDecimalIntegerLiterals tests parsing of decimal integer literals.
// Validates Requirement 2.4: Integer literals (decimal) create INT tokens.
func TestDecimalIntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0", "0"},
		{"1", "1"},
		{"42", "42"},
		{"123", "123"},
		{"999999", "999999"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != TOKEN_INT {
			t.Errorf("input=%q: expected type=TOKEN_INT, got=%v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
	}
}

// TestHexadecimalIntegerLiterals tests parsing of hexadecimal integer literals.
// Validates Requirement 2.4: Integer literals (0x-prefixed hexadecimal) create INT tokens.
func TestHexadecimalIntegerLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0x0", "0x0"},
		{"0xFF", "0xFF"},
		{"0xff", "0xff"},
		{"0x1A", "0x1A"},
		{"0x1a", "0x1a"},
		{"0X0", "0X0"},
		{"0XFF", "0XFF"},
		{"0Xff", "0Xff"},
		{"0xABCDEF", "0xABCDEF"},
		{"0x123456", "0x123456"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != TOKEN_INT {
			t.Errorf("input=%q: expected type=TOKEN_INT, got=%v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
	}
}

// TestFloatingPointLiterals tests parsing of floating point literals.
// Validates Requirement 2.5: Floating point literals create FLOAT tokens.
func TestFloatingPointLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"0.0", "0.0"},
		{"0.5", "0.5"},
		{"3.14", "3.14"},
		{"123.456", "123.456"},
		{"0.123", "0.123"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != TOKEN_FLOAT {
			t.Errorf("input=%q: expected type=TOKEN_FLOAT, got=%v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
	}
}

// TestNumericLiteralsWithContext tests numeric literals in context with other tokens.
func TestNumericLiteralsWithContext(t *testing.T) {
	input := "int x 123 0xFF 3.14"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_INT_TYPE, "int"},
		{TOKEN_IDENT, "x"},
		{TOKEN_INT, "123"},
		{TOKEN_INT, "0xFF"},
		{TOKEN_FLOAT, "3.14"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestNumericLiteralLineAndColumn tests that numeric literals have correct line and column.
// Validates Requirement 2.14: All tokens include line and column numbers.
func TestNumericLiteralLineAndColumn(t *testing.T) {
	input := "123\n0xFF\n3.14"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
		line      int
		column    int
	}{
		{TOKEN_INT, "123", 1, 1},
		{TOKEN_INT, "0xFF", 2, 1},
		{TOKEN_FLOAT, "3.14", 3, 1},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
		if tok.Line != exp.line {
			t.Errorf("token[%d] line: expected=%d, got=%d", i, exp.line, tok.Line)
		}
		if tok.Column != exp.column {
			t.Errorf("token[%d] column: expected=%d, got=%d", i, exp.column, tok.Column)
		}
	}
}

// TestIntegerFollowedByDot tests that integer followed by dot (not float) is handled correctly.
// For example, "123." should be parsed as INT "123" followed by something else.
func TestIntegerFollowedByDot(t *testing.T) {
	// "123." where dot is not followed by digit should be INT "123"
	// This is important for method calls like "obj.method"
	input := "123.abc"
	l := New(input)

	tok1 := l.NextToken()
	if tok1.Type != TOKEN_INT {
		t.Errorf("expected TOKEN_INT, got=%v", tok1.Type)
	}
	if tok1.Literal != "123" {
		t.Errorf("expected literal='123', got=%q", tok1.Literal)
	}
}

// TestStringLiterals tests parsing of string literals.
// Validates Requirement 2.6: String literals enclosed in double quotes create STRING tokens.
func TestStringLiterals(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello"`, "hello"},
		{`"world"`, "world"},
		{`"image.bmp"`, "image.bmp"},
		{`""`, ""},
		{`"hello world"`, "hello world"},
		{`"123"`, "123"},
		{`"test_file.txt"`, "test_file.txt"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != TOKEN_STRING {
			t.Errorf("input=%q: expected type=TOKEN_STRING, got=%v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
	}
}

// TestStringLiteralEscapeSequences tests parsing of escape sequences in string literals.
// Validates Requirement 2.6: String literals handle escape sequences.
func TestStringLiteralEscapeSequences(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello\nworld"`, "hello\nworld"},
		{`"tab\there"`, "tab\there"},
		{`"back\\slash"`, "back\\slash"},
		{`"quote\"here"`, "quote\"here"},
		{`"carriage\rreturn"`, "carriage\rreturn"},
		{`"null\0char"`, "null\x00char"},
		{`"\n\t\\\""`, "\n\t\\\""},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != TOKEN_STRING {
			t.Errorf("input=%q: expected type=TOKEN_STRING, got=%v", tt.input, tok.Type)
		}
		if tok.Literal != tt.expected {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expected, tok.Literal)
		}
	}
}

// TestStringLiteralLineAndColumn tests that string literals have correct line and column.
// Validates Requirement 2.14: All tokens include line and column numbers.
func TestStringLiteralLineAndColumn(t *testing.T) {
	input := `"hello"
"world"`
	l := New(input)

	// First string at line 1, column 1
	tok1 := l.NextToken()
	if tok1.Type != TOKEN_STRING {
		t.Errorf("token1 type: expected=TOKEN_STRING, got=%v", tok1.Type)
	}
	if tok1.Literal != "hello" {
		t.Errorf("token1 literal: expected='hello', got=%q", tok1.Literal)
	}
	if tok1.Line != 1 {
		t.Errorf("token1 line: expected=1, got=%d", tok1.Line)
	}
	if tok1.Column != 1 {
		t.Errorf("token1 column: expected=1, got=%d", tok1.Column)
	}

	// Second string at line 2, column 1
	tok2 := l.NextToken()
	if tok2.Type != TOKEN_STRING {
		t.Errorf("token2 type: expected=TOKEN_STRING, got=%v", tok2.Type)
	}
	if tok2.Literal != "world" {
		t.Errorf("token2 literal: expected='world', got=%q", tok2.Literal)
	}
	if tok2.Line != 2 {
		t.Errorf("token2 line: expected=2, got=%d", tok2.Line)
	}
	if tok2.Column != 1 {
		t.Errorf("token2 column: expected=1, got=%d", tok2.Column)
	}
}

// TestStringLiteralWithContext tests string literals in context with other tokens.
func TestStringLiteralWithContext(t *testing.T) {
	input := `LoadPic("image.bmp")`
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "LoadPic"},
		{TOKEN_LPAREN, "("},
		{TOKEN_STRING, "image.bmp"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestUnterminatedStringLiteral tests handling of unterminated string literals.
func TestUnterminatedStringLiteral(t *testing.T) {
	// Unterminated string should still return a STRING token with content up to EOF
	input := `"hello`
	l := New(input)
	tok := l.NextToken()

	if tok.Type != TOKEN_STRING {
		t.Errorf("expected type=TOKEN_STRING, got=%v", tok.Type)
	}
	if tok.Literal != "hello" {
		t.Errorf("expected literal='hello', got=%q", tok.Literal)
	}
}

// TestArithmeticOperators tests parsing of arithmetic operators.
// Validates Requirement 2.7: Operators create appropriate operator tokens.
func TestArithmeticOperators(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"+", TOKEN_PLUS, "+"},
		{"-", TOKEN_MINUS, "-"},
		{"*", TOKEN_ASTERISK, "*"},
		{"/", TOKEN_SLASH, "/"},
		{"%", TOKEN_PERCENT, "%"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("input=%q: expected type=%v, got=%v", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

// TestAssignmentOperator tests parsing of assignment operator.
// Validates Requirement 2.7: Operators create appropriate operator tokens.
func TestAssignmentOperator(t *testing.T) {
	l := New("=")
	tok := l.NextToken()

	if tok.Type != TOKEN_ASSIGN {
		t.Errorf("expected type=TOKEN_ASSIGN, got=%v", tok.Type)
	}
	if tok.Literal != "=" {
		t.Errorf("expected literal='=', got=%q", tok.Literal)
	}
}

// TestComparisonOperators tests parsing of comparison operators.
// Validates Requirement 2.7: Operators create appropriate operator tokens.
func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"==", TOKEN_EQ, "=="},
		{"!=", TOKEN_NEQ, "!="},
		{"<", TOKEN_LT, "<"},
		{">", TOKEN_GT, ">"},
		{"<=", TOKEN_LTE, "<="},
		{">=", TOKEN_GTE, ">="},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("input=%q: expected type=%v, got=%v", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

// TestLogicalOperators tests parsing of logical operators.
// Validates Requirement 2.7: Operators create appropriate operator tokens.
func TestLogicalOperators(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"&&", TOKEN_AND, "&&"},
		{"||", TOKEN_OR, "||"},
		{"!", TOKEN_NOT, "!"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("input=%q: expected type=%v, got=%v", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

// TestDelimiters tests parsing of delimiter tokens.
// Validates Requirement 2.8: Delimiters create appropriate delimiter tokens.
func TestDelimiters(t *testing.T) {
	tests := []struct {
		input        string
		expectedType TokenType
		expectedLit  string
	}{
		{"(", TOKEN_LPAREN, "("},
		{")", TOKEN_RPAREN, ")"},
		{"{", TOKEN_LBRACE, "{"},
		{"}", TOKEN_RBRACE, "}"},
		{"[", TOKEN_LBRACKET, "["},
		{"]", TOKEN_RBRACKET, "]"},
		{",", TOKEN_COMMA, ","},
		{";", TOKEN_SEMICOLON, ";"},
		{":", TOKEN_COLON, ":"},
	}

	for _, tt := range tests {
		l := New(tt.input)
		tok := l.NextToken()

		if tok.Type != tt.expectedType {
			t.Errorf("input=%q: expected type=%v, got=%v", tt.input, tt.expectedType, tok.Type)
		}
		if tok.Literal != tt.expectedLit {
			t.Errorf("input=%q: expected literal=%q, got=%q", tt.input, tt.expectedLit, tok.Literal)
		}
	}
}

// TestMultiCharOperatorDistinction tests that multi-character operators are correctly
// distinguished from their single-character counterparts.
func TestMultiCharOperatorDistinction(t *testing.T) {
	// Test "= =" (two separate assigns) vs "==" (equality)
	input := "= == != < <= > >= && || !"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_ASSIGN, "="},
		{TOKEN_EQ, "=="},
		{TOKEN_NEQ, "!="},
		{TOKEN_LT, "<"},
		{TOKEN_LTE, "<="},
		{TOKEN_GT, ">"},
		{TOKEN_GTE, ">="},
		{TOKEN_AND, "&&"},
		{TOKEN_OR, "||"},
		{TOKEN_NOT, "!"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestOperatorsInExpression tests operators in a typical expression context.
func TestOperatorsInExpression(t *testing.T) {
	input := "x + y * 2 - z / 3"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_PLUS, "+"},
		{TOKEN_IDENT, "y"},
		{TOKEN_ASTERISK, "*"},
		{TOKEN_INT, "2"},
		{TOKEN_MINUS, "-"},
		{TOKEN_IDENT, "z"},
		{TOKEN_SLASH, "/"},
		{TOKEN_INT, "3"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestComparisonInCondition tests comparison operators in a condition context.
func TestComparisonInCondition(t *testing.T) {
	input := "x >= 0 && x <= 100"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_GTE, ">="},
		{TOKEN_INT, "0"},
		{TOKEN_AND, "&&"},
		{TOKEN_IDENT, "x"},
		{TOKEN_LTE, "<="},
		{TOKEN_INT, "100"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestFunctionCallWithDelimiters tests delimiters in a function call context.
func TestFunctionCallWithDelimiters(t *testing.T) {
	input := "func(a, b, c)"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "func"},
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "a"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "b"},
		{TOKEN_COMMA, ","},
		{TOKEN_IDENT, "c"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestArrayAccessWithDelimiters tests delimiters in array access context.
func TestArrayAccessWithDelimiters(t *testing.T) {
	input := "arr[i] = value;"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "arr"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_IDENT, "i"},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_IDENT, "value"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestBlockWithBraces tests braces in a block context.
func TestBlockWithBraces(t *testing.T) {
	input := "if (x > 0) { y = 1; }"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IF, "if"},
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "x"},
		{TOKEN_GT, ">"},
		{TOKEN_INT, "0"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_IDENT, "y"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "1"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestSwitchCaseWithColon tests colon in switch-case context.
func TestSwitchCaseWithColon(t *testing.T) {
	input := "case 1: x = 10;"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_CASE, "case"},
		{TOKEN_INT, "1"},
		{TOKEN_COLON, ":"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_SEMICOLON, ";"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestInvalidSingleAmpersand tests that single '&' is treated as illegal.
func TestInvalidSingleAmpersand(t *testing.T) {
	l := New("&")
	tok := l.NextToken()

	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("expected type=TOKEN_ILLEGAL, got=%v", tok.Type)
	}
	if tok.Literal != "&" {
		t.Errorf("expected literal='&', got=%q", tok.Literal)
	}
}

// TestInvalidSinglePipe tests that single '|' is treated as illegal.
func TestInvalidSinglePipe(t *testing.T) {
	l := New("|")
	tok := l.NextToken()

	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("expected type=TOKEN_ILLEGAL, got=%v", tok.Type)
	}
	if tok.Literal != "|" {
		t.Errorf("expected literal='|', got=%q", tok.Literal)
	}
}

// TestOperatorLineAndColumn tests that operators have correct line and column.
// Validates Requirement 2.14: All tokens include line and column numbers.
func TestOperatorLineAndColumn(t *testing.T) {
	input := "x + y\n== !="
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
		line      int
		column    int
	}{
		{TOKEN_IDENT, "x", 1, 1},
		{TOKEN_PLUS, "+", 1, 3},
		{TOKEN_IDENT, "y", 1, 5},
		{TOKEN_EQ, "==", 2, 1},
		{TOKEN_NEQ, "!=", 2, 4},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
		if tok.Line != exp.line {
			t.Errorf("token[%d] line: expected=%d, got=%d", i, exp.line, tok.Line)
		}
		if tok.Column != exp.column {
			t.Errorf("token[%d] column: expected=%d, got=%d", i, exp.column, tok.Column)
		}
	}
}

// TestComplexExpression tests a complex expression with multiple operators and delimiters.
func TestComplexExpression(t *testing.T) {
	input := "(a + b) * (c - d) / e % f"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "a"},
		{TOKEN_PLUS, "+"},
		{TOKEN_IDENT, "b"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_ASTERISK, "*"},
		{TOKEN_LPAREN, "("},
		{TOKEN_IDENT, "c"},
		{TOKEN_MINUS, "-"},
		{TOKEN_IDENT, "d"},
		{TOKEN_RPAREN, ")"},
		{TOKEN_SLASH, "/"},
		{TOKEN_IDENT, "e"},
		{TOKEN_PERCENT, "%"},
		{TOKEN_IDENT, "f"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestNotEqualVsNotFollowedByEqual tests distinction between != and ! followed by =.
func TestNotEqualVsNotFollowedByEqual(t *testing.T) {
	// "!=" should be TOKEN_NEQ
	l1 := New("!=")
	tok1 := l1.NextToken()
	if tok1.Type != TOKEN_NEQ {
		t.Errorf("'!=' expected type=TOKEN_NEQ, got=%v", tok1.Type)
	}

	// "! =" should be TOKEN_NOT followed by TOKEN_ASSIGN
	l2 := New("! =")
	tok2a := l2.NextToken()
	tok2b := l2.NextToken()
	if tok2a.Type != TOKEN_NOT {
		t.Errorf("'! =' first token expected type=TOKEN_NOT, got=%v", tok2a.Type)
	}
	if tok2b.Type != TOKEN_ASSIGN {
		t.Errorf("'! =' second token expected type=TOKEN_ASSIGN, got=%v", tok2b.Type)
	}
}

// TestSingleLineComment tests that single-line comments are skipped.
// Validates Requirement 2.9: Single-line comments (//) skip until next line.
func TestSingleLineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			tokenType TokenType
			literal   string
		}
	}{
		{
			name:  "comment at end of line",
			input: "x = 5 // this is a comment",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_ASSIGN, "="},
				{TOKEN_INT, "5"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "comment on its own line",
			input: "// comment\nx = 5",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_ASSIGN, "="},
				{TOKEN_INT, "5"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "multiple single-line comments",
			input: "// first comment\nx = 5\n// second comment\ny = 10",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_ASSIGN, "="},
				{TOKEN_INT, "5"},
				{TOKEN_IDENT, "y"},
				{TOKEN_ASSIGN, "="},
				{TOKEN_INT, "10"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "comment only",
			input: "// just a comment",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "empty comment",
			input: "//\nx",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.tokenType {
					t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
				}
				if tok.Literal != exp.literal {
					t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
				}
			}
		})
	}
}

// TestMultiLineComment tests that multi-line comments are skipped.
// Validates Requirement 2.10: Multi-line comments (/* */) skip until closing */.
func TestMultiLineComment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []struct {
			tokenType TokenType
			literal   string
		}
	}{
		{
			name:  "simple multi-line comment",
			input: "x /* comment */ y",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_IDENT, "y"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "multi-line comment spanning lines",
			input: "x /* this is\na multi-line\ncomment */ y",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_IDENT, "y"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "multi-line comment at start",
			input: "/* comment */ x = 5",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_ASSIGN, "="},
				{TOKEN_INT, "5"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "multi-line comment at end",
			input: "x = 5 /* comment */",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_ASSIGN, "="},
				{TOKEN_INT, "5"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "empty multi-line comment",
			input: "x /**/ y",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_IDENT, "y"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "multi-line comment with asterisks",
			input: "x /* ** comment ** */ y",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_IDENT, "x"},
				{TOKEN_IDENT, "y"},
				{TOKEN_EOF, ""},
			},
		},
		{
			name:  "comment only",
			input: "/* just a comment */",
			expected: []struct {
				tokenType TokenType
				literal   string
			}{
				{TOKEN_EOF, ""},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			for i, exp := range tt.expected {
				tok := l.NextToken()
				if tok.Type != exp.tokenType {
					t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
				}
				if tok.Literal != exp.literal {
					t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
				}
			}
		})
	}
}

// TestUnterminatedMultiLineComment tests handling of unterminated multi-line comments.
func TestUnterminatedMultiLineComment(t *testing.T) {
	// Unterminated multi-line comment should skip to EOF
	input := "x /* unterminated comment"
	l := New(input)

	tok1 := l.NextToken()
	if tok1.Type != TOKEN_IDENT {
		t.Errorf("expected TOKEN_IDENT, got=%v", tok1.Type)
	}
	if tok1.Literal != "x" {
		t.Errorf("expected literal='x', got=%q", tok1.Literal)
	}

	tok2 := l.NextToken()
	if tok2.Type != TOKEN_EOF {
		t.Errorf("expected TOKEN_EOF, got=%v", tok2.Type)
	}
}

// TestMixedComments tests mixing single-line and multi-line comments.
func TestMixedComments(t *testing.T) {
	input := `x = 5 // single line comment
/* multi
line
comment */
y = 10 // another single line
/* another multi */ z = 15`

	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "5"},
		{TOKEN_IDENT, "y"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_IDENT, "z"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "15"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestDivisionVsComment tests that division operator is not confused with comments.
func TestDivisionVsComment(t *testing.T) {
	input := "a / b // comment\nc /* comment */ / d"
	l := New(input)

	expected := []struct {
		tokenType TokenType
		literal   string
	}{
		{TOKEN_IDENT, "a"},
		{TOKEN_SLASH, "/"},
		{TOKEN_IDENT, "b"},
		{TOKEN_IDENT, "c"},
		{TOKEN_SLASH, "/"},
		{TOKEN_IDENT, "d"},
		{TOKEN_EOF, ""},
	}

	for i, exp := range expected {
		tok := l.NextToken()
		if tok.Type != exp.tokenType {
			t.Errorf("token[%d] type: expected=%v, got=%v", i, exp.tokenType, tok.Type)
		}
		if tok.Literal != exp.literal {
			t.Errorf("token[%d] literal: expected=%q, got=%q", i, exp.literal, tok.Literal)
		}
	}
}

// TestCommentLineTracking tests that line numbers are correctly tracked after comments.
// Validates Requirement 2.14: All tokens include line and column numbers.
func TestCommentLineTracking(t *testing.T) {
	input := `// comment on line 1
x
/* multi
line
comment */
y`

	l := New(input)

	// x should be on line 2
	tok1 := l.NextToken()
	if tok1.Literal != "x" {
		t.Errorf("expected literal='x', got=%q", tok1.Literal)
	}
	if tok1.Line != 2 {
		t.Errorf("x line: expected=2, got=%d", tok1.Line)
	}

	// y should be on line 6
	tok2 := l.NextToken()
	if tok2.Literal != "y" {
		t.Errorf("expected literal='y', got=%q", tok2.Literal)
	}
	if tok2.Line != 6 {
		t.Errorf("y line: expected=6, got=%d", tok2.Line)
	}
}
