package lexer

import (
	"testing"

	"github.com/zurustar/filly2exe/pkg/compiler/token"
)

func TestNextToken(t *testing.T) {
	input := `
	IMG01 = LoadPic("sample.bmp");
	MovePic(IMG01, 0, 0);,,
	mes(MIDI_TIME){
		step(8){
			,,,,
		}
	}
	`

	tests := []struct {
		expectedType    token.TokenType
		expectedLiteral string
	}{
		{token.IDENT, "IMG01"},
		{token.ASSIGN, "="},
		{token.IDENT, "LoadPic"},
		{token.LPAREN, "("},
		{token.STRING, "sample.bmp"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},

		{token.IDENT, "MovePic"},
		{token.LPAREN, "("},
		{token.IDENT, "IMG01"},
		{token.COMMA, ","},
		{token.NUMBER, "0"},
		{token.COMMA, ","},
		{token.NUMBER, "0"},
		{token.RPAREN, ")"},
		{token.SEMICOLON, ";"},
		{token.COMMA, ","},
		{token.COMMA, ","},

		{token.MES, "mes"},
		{token.LPAREN, "("},
		{token.IDENT, "MIDI_TIME"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},

		{token.STEP, "step"},
		{token.LPAREN, "("},
		{token.NUMBER, "8"},
		{token.RPAREN, ")"},
		{token.LBRACE, "{"},

		{token.COMMA, ","},
		{token.COMMA, ","},
		{token.COMMA, ","},
		{token.COMMA, ","},

		{token.RBRACE, "}"},
		{token.RBRACE, "}"},
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
