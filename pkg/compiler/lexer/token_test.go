package lexer

import "testing"

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  string
	}{
		// Special tokens
		{TOKEN_ILLEGAL, "ILLEGAL"},
		{TOKEN_EOF, "EOF"},
		{TOKEN_COMMENT, "COMMENT"},

		// Literals
		{TOKEN_IDENT, "IDENT"},
		{TOKEN_INT, "INT"},
		{TOKEN_FLOAT, "FLOAT"},
		{TOKEN_STRING, "STRING"},

		// Operators
		{TOKEN_PLUS, "+"},
		{TOKEN_MINUS, "-"},
		{TOKEN_ASTERISK, "*"},
		{TOKEN_SLASH, "/"},
		{TOKEN_PERCENT, "%"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_EQ, "=="},
		{TOKEN_NEQ, "!="},
		{TOKEN_LT, "<"},
		{TOKEN_GT, ">"},
		{TOKEN_LTE, "<="},
		{TOKEN_GTE, ">="},
		{TOKEN_AND, "&&"},
		{TOKEN_OR, "||"},
		{TOKEN_NOT, "!"},

		// Delimiters
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_LBRACKET, "["},
		{TOKEN_RBRACKET, "]"},
		{TOKEN_COMMA, ","},
		{TOKEN_SEMICOLON, ";"},

		// Keywords
		{TOKEN_INT_TYPE, "int"},
		{TOKEN_STR_TYPE, "str"},
		{TOKEN_IF, "if"},
		{TOKEN_ELSE, "else"},
		{TOKEN_FOR, "for"},
		{TOKEN_WHILE, "while"},
		{TOKEN_SWITCH, "switch"},
		{TOKEN_CASE, "case"},
		{TOKEN_DEFAULT, "default"},
		{TOKEN_BREAK, "break"},
		{TOKEN_CONTINUE, "continue"},
		{TOKEN_RETURN, "return"},
		{TOKEN_MES, "mes"},
		{TOKEN_STEP, "step"},
		{TOKEN_END_STEP, "end_step"},
		{TOKEN_DEL_ME, "del_me"},
		{TOKEN_DEL_US, "del_us"},
		{TOKEN_DEL_ALL, "del_all"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if got := tt.tokenType.String(); got != tt.expected {
				t.Errorf("TokenType.String() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestTokenTypeStringUnknown(t *testing.T) {
	// Test unknown token type returns "UNKNOWN"
	unknownType := TokenType(9999)
	if got := unknownType.String(); got != "UNKNOWN" {
		t.Errorf("Unknown TokenType.String() = %q, want %q", got, "UNKNOWN")
	}
}

func TestTokenTypeIsKeyword(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{TOKEN_INT_TYPE, true},
		{TOKEN_STR_TYPE, true},
		{TOKEN_IF, true},
		{TOKEN_MES, true},
		{TOKEN_STEP, true},
		{TOKEN_DEL_ALL, true},
		{TOKEN_IDENT, false},
		{TOKEN_INT, false},
		{TOKEN_PLUS, false},
		{TOKEN_LPAREN, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			if got := tt.tokenType.IsKeyword(); got != tt.expected {
				t.Errorf("TokenType.IsKeyword() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTokenTypeIsOperator(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{TOKEN_PLUS, true},
		{TOKEN_MINUS, true},
		{TOKEN_EQ, true},
		{TOKEN_AND, true},
		{TOKEN_NOT, true},
		{TOKEN_IDENT, false},
		{TOKEN_INT_TYPE, false},
		{TOKEN_LPAREN, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			if got := tt.tokenType.IsOperator(); got != tt.expected {
				t.Errorf("TokenType.IsOperator() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTokenTypeIsLiteral(t *testing.T) {
	tests := []struct {
		tokenType TokenType
		expected  bool
	}{
		{TOKEN_IDENT, true},
		{TOKEN_INT, true},
		{TOKEN_FLOAT, true},
		{TOKEN_STRING, true},
		{TOKEN_PLUS, false},
		{TOKEN_INT_TYPE, false},
		{TOKEN_LPAREN, false},
	}

	for _, tt := range tests {
		t.Run(tt.tokenType.String(), func(t *testing.T) {
			if got := tt.tokenType.IsLiteral(); got != tt.expected {
				t.Errorf("TokenType.IsLiteral() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestTokenStruct(t *testing.T) {
	tok := Token{
		Type:    TOKEN_IDENT,
		Literal: "myVar",
		Line:    10,
		Column:  5,
	}

	if tok.Type != TOKEN_IDENT {
		t.Errorf("Token.Type = %v, want %v", tok.Type, TOKEN_IDENT)
	}
	if tok.Literal != "myVar" {
		t.Errorf("Token.Literal = %q, want %q", tok.Literal, "myVar")
	}
	if tok.Line != 10 {
		t.Errorf("Token.Line = %d, want %d", tok.Line, 10)
	}
	if tok.Column != 5 {
		t.Errorf("Token.Column = %d, want %d", tok.Column, 5)
	}
}

func TestLookupIdent(t *testing.T) {
	tests := []struct {
		name     string
		ident    string
		expected TokenType
	}{
		// Lowercase keywords
		{"lowercase int", "int", TOKEN_INT_TYPE},
		{"lowercase str", "str", TOKEN_STR_TYPE},
		{"lowercase if", "if", TOKEN_IF},
		{"lowercase else", "else", TOKEN_ELSE},
		{"lowercase for", "for", TOKEN_FOR},
		{"lowercase while", "while", TOKEN_WHILE},
		{"lowercase switch", "switch", TOKEN_SWITCH},
		{"lowercase case", "case", TOKEN_CASE},
		{"lowercase default", "default", TOKEN_DEFAULT},
		{"lowercase break", "break", TOKEN_BREAK},
		{"lowercase continue", "continue", TOKEN_CONTINUE},
		{"lowercase return", "return", TOKEN_RETURN},
		{"lowercase mes", "mes", TOKEN_MES},
		{"lowercase step", "step", TOKEN_STEP},
		{"lowercase end_step", "end_step", TOKEN_END_STEP},
		{"lowercase del_me", "del_me", TOKEN_DEL_ME},
		{"lowercase del_us", "del_us", TOKEN_DEL_US},
		{"lowercase del_all", "del_all", TOKEN_DEL_ALL},

		// Uppercase keywords (case-insensitive)
		{"uppercase MES", "MES", TOKEN_MES},
		{"uppercase STEP", "STEP", TOKEN_STEP},
		{"uppercase IF", "IF", TOKEN_IF},
		{"uppercase FOR", "FOR", TOKEN_FOR},
		{"uppercase WHILE", "WHILE", TOKEN_WHILE},
		{"uppercase INT", "INT", TOKEN_INT_TYPE},
		{"uppercase STR", "STR", TOKEN_STR_TYPE},
		{"uppercase END_STEP", "END_STEP", TOKEN_END_STEP},
		{"uppercase DEL_ME", "DEL_ME", TOKEN_DEL_ME},

		// Mixed case keywords (case-insensitive)
		{"mixed case Mes", "Mes", TOKEN_MES},
		{"mixed case Step", "Step", TOKEN_STEP},
		{"mixed case If", "If", TOKEN_IF},
		{"mixed case For", "For", TOKEN_FOR},
		{"mixed case While", "While", TOKEN_WHILE},
		{"mixed case Int", "Int", TOKEN_INT_TYPE},
		{"mixed case Str", "Str", TOKEN_STR_TYPE},
		{"mixed case End_Step", "End_Step", TOKEN_END_STEP},
		{"mixed case Del_Me", "Del_Me", TOKEN_DEL_ME},
		{"mixed case mEs", "mEs", TOKEN_MES},
		{"mixed case sTeP", "sTeP", TOKEN_STEP},

		// Non-keywords should return TOKEN_IDENT
		{"identifier myVar", "myVar", TOKEN_IDENT},
		{"identifier x", "x", TOKEN_IDENT},
		{"identifier LoadPic", "LoadPic", TOKEN_IDENT},
		{"identifier MovePic", "MovePic", TOKEN_IDENT},
		{"identifier counter", "counter", TOKEN_IDENT},
		{"identifier _private", "_private", TOKEN_IDENT},
		{"identifier var123", "var123", TOKEN_IDENT},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LookupIdent(tt.ident); got != tt.expected {
				t.Errorf("LookupIdent(%q) = %v, want %v", tt.ident, got, tt.expected)
			}
		})
	}
}
