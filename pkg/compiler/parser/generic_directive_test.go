package parser

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
)

// TestUnknownDirectiveDoesNotEatNextToken is a regression test for the bug where
// parseGenericDirective over-advanced, consuming the first token of the
// statement following an unknown directive (e.g. turning `int x;` into `x`).
// See docs/bug-hunt-findings.md finding C.
func TestUnknownDirectiveDoesNotEatNextToken(t *testing.T) {
	src := "#foo bar\nint x;\nint y;"
	p := New(lexer.New(src))
	prog, _ := p.ParseProgram()

	// Both declarations must survive as VarDeclaration nodes.
	var decls []*VarDeclaration
	for _, s := range prog.Statements {
		if vd, ok := s.(*VarDeclaration); ok {
			decls = append(decls, vd)
		}
	}
	if len(decls) != 2 {
		t.Fatalf("expected 2 VarDeclaration statements, got %d (statements=%#v)", len(decls), prog.Statements)
	}
	if decls[0].Names[0] != "x" || decls[1].Names[0] != "y" {
		t.Errorf("declarations corrupted: got %q and %q, want x and y", decls[0].Names[0], decls[1].Names[0])
	}
	if len(p.Errors()) != 0 {
		t.Errorf("unexpected parser errors: %v", p.Errors())
	}
}
