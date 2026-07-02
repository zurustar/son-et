package lexer

import "testing"

// TestHexPrefixWithoutDigitsIsIllegal is a regression test for the phase-1 minor
// finding: "0x" with no following hex digits must be reported as ILLEGAL at the
// lexer stage rather than producing an INT token.
func TestHexPrefixWithoutDigitsIsIllegal(t *testing.T) {
	l := New("v = 0x")
	_, errs := l.TokenizeWithErrors()
	if len(errs) == 0 {
		t.Errorf("expected a lexer error for bare \"0x\", got none")
	}

	// Also verify the token type directly.
	l2 := New("0x")
	tok := l2.NextToken()
	if tok.Type != TOKEN_ILLEGAL {
		t.Errorf("expected TOKEN_ILLEGAL for \"0x\", got %v (literal %q)", tok.Type, tok.Literal)
	}

	// A valid hex literal must still lex as INT with no errors.
	l3 := New("0x1F")
	tok3 := l3.NextToken()
	if tok3.Type != TOKEN_INT || tok3.Literal != "0x1F" {
		t.Errorf("expected INT 0x1F, got %v %q", tok3.Type, tok3.Literal)
	}
	if len(l3.Errors()) != 0 {
		t.Errorf("valid hex literal should produce no errors, got %v", l3.Errors())
	}
}

// TestUnterminatedConstructsAreReported is a regression test for the phase-1
// minor finding: unterminated string/comment were silently accepted. The token
// output is unchanged, but Errors()/TokenizeWithErrors must now report them.
func TestUnterminatedConstructsAreReported(t *testing.T) {
	t.Run("unterminated string", func(t *testing.T) {
		l := New(`x = "abc`)
		_, errs := l.TokenizeWithErrors()
		if len(errs) == 0 {
			t.Errorf("expected an error for unterminated string, got none")
		}
	})

	t.Run("unterminated multi-line comment", func(t *testing.T) {
		l := New("a /* never closed")
		_, errs := l.TokenizeWithErrors()
		if len(errs) == 0 {
			t.Errorf("expected an error for unterminated comment, got none")
		}
	})

	t.Run("well-formed input has no lexer errors", func(t *testing.T) {
		l := New(`x = "abc" /* ok */ + 1`)
		_, errs := l.TokenizeWithErrors()
		if len(errs) != 0 {
			t.Errorf("expected no errors for valid input, got %v", errs)
		}
	})
}
