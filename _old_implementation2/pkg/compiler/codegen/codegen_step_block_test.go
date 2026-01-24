package codegen

import (
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/token"
)

func TestGenerateStepBlockStatement(t *testing.T) {
	tests := []struct {
		name        string
		stmt        *ast.StepStatement
		wantCmd     interpreter.OpCmd
		wantFlatSeq bool
		expectedOps int // Expected number of opcodes in flat sequence
	}{
		{
			name: "simple step",
			stmt: &ast.StepStatement{
				Token: token.Token{Type: token.STEP, Literal: "step"},
				Count: &ast.IntegerLiteral{Value: 10},
				Body:  nil,
			},
			wantCmd:     interpreter.OpWait,
			wantFlatSeq: false,
			expectedOps: 1,
		},
		{
			name: "step block with one command",
			stmt: &ast.StepStatement{
				Token: token.Token{Type: token.STEP, Literal: "step"},
				Count: &ast.IntegerLiteral{Value: 5},
				Body: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{Value: "DoSomething"},
							},
						},
					},
				},
			},
			wantCmd:     interpreter.OpCall,
			wantFlatSeq: true,
			expectedOps: 2, // Command + Wait
		},
		{
			name: "step block with multiple commands",
			stmt: &ast.StepStatement{
				Token: token.Token{Type: token.STEP, Literal: "step"},
				Count: &ast.IntegerLiteral{Value: 8},
				Body: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{Value: "Command1"},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{Value: "Command2"},
							},
						},
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{Value: "Command3"},
							},
						},
					},
				},
			},
			wantCmd:     interpreter.OpCall,
			wantFlatSeq: true,
			expectedOps: 6, // 3 commands + 3 waits
		},
		{
			name: "step block with empty steps (commas)",
			stmt: &ast.StepStatement{
				Token: token.Token{Type: token.STEP, Literal: "step"},
				Count: &ast.IntegerLiteral{Value: 10},
				Body: &ast.BlockStatement{
					Statements: []ast.Statement{
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{Value: "Command1"},
							},
						},
						// Empty statement (from comma)
						&ast.ExpressionStatement{
							Expression: nil,
						},
						&ast.ExpressionStatement{
							Expression: &ast.CallExpression{
								Function: &ast.Identifier{Value: "Command2"},
							},
						},
					},
				},
			},
			wantCmd:     interpreter.OpCall,
			wantFlatSeq: true,
			expectedOps: 5, // Command1 + Wait + EmptyWait + Command2 + Wait
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New()
			opcodes := g.generateStepStatement(tt.stmt)

			if len(opcodes) == 0 {
				t.Fatal("expected at least one opcode")
			}

			if opcodes[0].Cmd != tt.wantCmd {
				t.Errorf("expected first command %v, got %v", tt.wantCmd, opcodes[0].Cmd)
			}

			if tt.wantFlatSeq {
				// Verify it's a flat sequence (not a loop)
				if len(opcodes) != tt.expectedOps {
					t.Errorf("expected %d opcodes in flat sequence, got %d", tt.expectedOps, len(opcodes))
				}

				// Count wait operations
				waitCount := 0
				for _, op := range opcodes {
					if op.Cmd == interpreter.OpWait {
						waitCount++
					}
				}

				if waitCount == 0 {
					t.Error("expected at least one wait operation in sequence")
				}
			}
		})
	}
}

func TestStepBlockWithEndStep(t *testing.T) {
	// Test that end_step stops code generation
	stmt := &ast.StepStatement{
		Token: token.Token{Type: token.STEP, Literal: "step"},
		Count: &ast.IntegerLiteral{Value: 10},
		Body: &ast.BlockStatement{
			Statements: []ast.Statement{
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Function: &ast.Identifier{Value: "Command1"},
					},
				},
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Function: &ast.Identifier{Value: "end_step"},
					},
				},
				// This command should NOT be generated (after end_step)
				&ast.ExpressionStatement{
					Expression: &ast.CallExpression{
						Function: &ast.Identifier{Value: "Command2"},
					},
				},
			},
		},
	}

	g := New()
	opcodes := g.generateStepStatement(stmt)

	if len(opcodes) == 0 {
		t.Fatal("expected at least one opcode")
	}

	// Should generate: Command1 + Wait (end_step stops generation)
	if len(opcodes) != 2 {
		t.Errorf("expected 2 opcodes (Command1 + Wait), got %d", len(opcodes))
	}

	// First should be Command1
	if opcodes[0].Cmd != interpreter.OpCall {
		t.Errorf("expected OpCall, got %v", opcodes[0].Cmd)
	}

	// Second should be Wait
	if opcodes[1].Cmd != interpreter.OpWait {
		t.Errorf("expected OpWait, got %v", opcodes[1].Cmd)
	}

	// Verify Command2 was NOT generated
	for _, op := range opcodes {
		if op.Cmd == interpreter.OpCall && len(op.Args) > 0 {
			if funcName, ok := op.Args[0].(interpreter.Variable); ok && string(funcName) == "Command2" {
				t.Error("Command2 should not be generated after end_step")
			}
		}
	}
}
