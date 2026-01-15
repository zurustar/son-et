package parser

import (
	"testing"

	"github.com/zurustar/filly2exe/pkg/compiler/ast"
	"github.com/zurustar/filly2exe/pkg/compiler/lexer"
)

func TestIfStatement(t *testing.T) {
	input := `
	if (x == 5) {
		y = 10;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T", program.Statements[0])
	}

	if stmt.Consequence == nil {
		t.Fatal("stmt.Consequence is nil")
	}

	if len(stmt.Consequence.Statements) != 1 {
		t.Errorf("consequence is not 1 statement. got=%d", len(stmt.Consequence.Statements))
	}
}

func TestIfElseStatement(t *testing.T) {
	input := `
	if (x > 5) {
		y = 10;
	} else {
		y = 20;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T", program.Statements[0])
	}

	if stmt.Alternative == nil {
		t.Fatal("stmt.Alternative is nil")
	}

	if len(stmt.Alternative.Statements) != 1 {
		t.Errorf("alternative is not 1 statement. got=%d", len(stmt.Alternative.Statements))
	}
}

func TestNestedIfStatement(t *testing.T) {
	input := `
	if (x > 5) {
		if (y < 10) {
			z = 1;
		}
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Consequence.Statements) != 1 {
		t.Fatalf("consequence does not contain 1 statement. got=%d", len(stmt.Consequence.Statements))
	}

	nestedIf, ok := stmt.Consequence.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("nested statement is not ast.IfStatement. got=%T", stmt.Consequence.Statements[0])
	}

	if nestedIf.Consequence == nil {
		t.Fatal("nested if consequence is nil")
	}
}

func TestComparisonOperators(t *testing.T) {
	tests := []struct {
		input    string
		operator string
	}{
		{"if (x == 5) { y = 1; }", "=="},
		{"if (x != 5) { y = 1; }", "!="},
		{"if (x < 5) { y = 1; }", "<"},
		{"if (x > 5) { y = 1; }", ">"},
		{"if (x <= 5) { y = 1; }", "<="},
		{"if (x >= 5) { y = 1; }", ">="},
	}

	for _, tt := range tests {
		l := lexer.New(tt.input)
		p := New(l)
		program := p.ParseProgram()
		checkParserErrors(t, p)

		if len(program.Statements) != 1 {
			t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
		}

		stmt, ok := program.Statements[0].(*ast.IfStatement)
		if !ok {
			t.Fatalf("program.Statements[0] is not ast.IfStatement. got=%T", program.Statements[0])
		}

		infix, ok := stmt.Condition.(*ast.InfixExpression)
		if !ok {
			t.Fatalf("condition is not ast.InfixExpression. got=%T", stmt.Condition)
		}

		if infix.Operator != tt.operator {
			t.Errorf("operator is not '%s'. got=%s", tt.operator, infix.Operator)
		}
	}
}

func checkParserErrors(t *testing.T, p *Parser) {
	errors := p.Errors()
	if len(errors) == 0 {
		return
	}

	t.Errorf("parser has %d errors", len(errors))
	for _, msg := range errors {
		t.Errorf("parser error: %q", msg)
	}
	t.FailNow()
}

func TestForLoop(t *testing.T) {
	input := `
	for (i = 0; i < 10; i = i + 1) {
		x = x + 1;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ForStatement. got=%T", program.Statements[0])
	}

	if stmt.Init == nil {
		t.Fatal("stmt.Init is nil")
	}

	if stmt.Condition == nil {
		t.Fatal("stmt.Condition is nil")
	}

	if stmt.Post == nil {
		t.Fatal("stmt.Post is nil")
	}

	if stmt.Body == nil {
		t.Fatal("stmt.Body is nil")
	}
}

func TestForLoopWithBreak(t *testing.T) {
	input := `
	for (i = 0; i < 10; i = i + 1) {
		if (i == 5) {
			break;
		}
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ForStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Body.Statements) != 1 {
		t.Fatalf("for body does not contain 1 statement. got=%d", len(stmt.Body.Statements))
	}

	ifStmt, ok := stmt.Body.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("body statement is not ast.IfStatement. got=%T", stmt.Body.Statements[0])
	}

	if len(ifStmt.Consequence.Statements) != 1 {
		t.Fatalf("if consequence does not contain 1 statement. got=%d", len(ifStmt.Consequence.Statements))
	}

	breakStmt, ok := ifStmt.Consequence.Statements[0].(*ast.BreakStatement)
	if !ok {
		t.Fatalf("statement is not ast.BreakStatement. got=%T", ifStmt.Consequence.Statements[0])
	}

	if breakStmt.TokenLiteral() != "break" {
		t.Errorf("break token literal is not 'break'. got=%s", breakStmt.TokenLiteral())
	}
}

func TestForLoopWithContinue(t *testing.T) {
	input := `
	for (i = 0; i < 10; i = i + 1) {
		if (i == 5) {
			continue;
		}
		x = x + 1;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.ForStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.ForStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Body.Statements) != 2 {
		t.Fatalf("for body does not contain 2 statements. got=%d", len(stmt.Body.Statements))
	}

	ifStmt, ok := stmt.Body.Statements[0].(*ast.IfStatement)
	if !ok {
		t.Fatalf("body statement is not ast.IfStatement. got=%T", stmt.Body.Statements[0])
	}

	continueStmt, ok := ifStmt.Consequence.Statements[0].(*ast.ContinueStatement)
	if !ok {
		t.Fatalf("statement is not ast.ContinueStatement. got=%T", ifStmt.Consequence.Statements[0])
	}

	if continueStmt.TokenLiteral() != "continue" {
		t.Errorf("continue token literal is not 'continue'. got=%s", continueStmt.TokenLiteral())
	}
}

func TestWhileLoop(t *testing.T) {
	input := `
	while (x < 10) {
		x = x + 1;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.WhileStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.WhileStatement. got=%T", program.Statements[0])
	}

	if stmt.Condition == nil {
		t.Fatal("stmt.Condition is nil")
	}

	if stmt.Body == nil {
		t.Fatal("stmt.Body is nil")
	}

	if len(stmt.Body.Statements) != 1 {
		t.Errorf("while body does not contain 1 statement. got=%d", len(stmt.Body.Statements))
	}
}

func TestDoWhileLoop(t *testing.T) {
	input := `
	do {
		x = x + 1;
	} while (x < 10);
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.DoWhileStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.DoWhileStatement. got=%T", program.Statements[0])
	}

	if stmt.Body == nil {
		t.Fatal("stmt.Body is nil")
	}

	if stmt.Condition == nil {
		t.Fatal("stmt.Condition is nil")
	}

	if len(stmt.Body.Statements) != 1 {
		t.Errorf("do-while body does not contain 1 statement. got=%d", len(stmt.Body.Statements))
	}
}

func TestDoWhileExecutesAtLeastOnce(t *testing.T) {
	// This test verifies the structure is correct for do-while
	// The actual execution behavior (at least once) is tested in integration tests
	input := `
	do {
		x = x + 1;
		y = y + 2;
	} while (x < 0);
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.DoWhileStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.DoWhileStatement. got=%T", program.Statements[0])
	}

	// Verify body has statements (will execute at least once)
	if len(stmt.Body.Statements) != 2 {
		t.Errorf("do-while body does not contain 2 statements. got=%d", len(stmt.Body.Statements))
	}
}

func TestSwitchStatement(t *testing.T) {
	input := `
	switch (x) {
		case 1:
			y = 10;
		case 2:
			y = 20;
		default:
			y = 0;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SwitchStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SwitchStatement. got=%T", program.Statements[0])
	}

	if stmt.Value == nil {
		t.Fatal("stmt.Value is nil")
	}

	if len(stmt.Cases) != 2 {
		t.Fatalf("switch does not have 2 cases. got=%d", len(stmt.Cases))
	}

	if stmt.Default == nil {
		t.Fatal("stmt.Default is nil")
	}
}

func TestSwitchWithMultipleCases(t *testing.T) {
	input := `
	switch (x) {
		case 1:
			y = 10;
		case 2:
			y = 20;
		case 3:
			y = 30;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SwitchStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SwitchStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Cases) != 3 {
		t.Fatalf("switch does not have 3 cases. got=%d", len(stmt.Cases))
	}
}

func TestSwitchWithDefaultOnly(t *testing.T) {
	input := `
	switch (x) {
		default:
			y = 0;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SwitchStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SwitchStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Cases) != 0 {
		t.Fatalf("switch should have 0 cases. got=%d", len(stmt.Cases))
	}

	if stmt.Default == nil {
		t.Fatal("stmt.Default is nil")
	}
}

func TestSwitchWithBreak(t *testing.T) {
	input := `
	switch (x) {
		case 1:
			y = 10;
			break;
		case 2:
			y = 20;
			break;
	}
	`

	l := lexer.New(input)
	p := New(l)
	program := p.ParseProgram()
	checkParserErrors(t, p)

	if len(program.Statements) != 1 {
		t.Fatalf("program.Statements does not contain 1 statement. got=%d", len(program.Statements))
	}

	stmt, ok := program.Statements[0].(*ast.SwitchStatement)
	if !ok {
		t.Fatalf("program.Statements[0] is not ast.SwitchStatement. got=%T", program.Statements[0])
	}

	if len(stmt.Cases) != 2 {
		t.Fatalf("switch does not have 2 cases. got=%d", len(stmt.Cases))
	}

	// Check first case has break
	firstCase := stmt.Cases[0]
	if len(firstCase.Body.Statements) != 2 {
		t.Fatalf("first case does not have 2 statements. got=%d", len(firstCase.Body.Statements))
	}

	breakStmt, ok := firstCase.Body.Statements[1].(*ast.BreakStatement)
	if !ok {
		t.Fatalf("second statement in first case is not ast.BreakStatement. got=%T", firstCase.Body.Statements[1])
	}

	if breakStmt.TokenLiteral() != "break" {
		t.Errorf("break token literal is not 'break'. got=%s", breakStmt.TokenLiteral())
	}
}
