// Package compiler provides OpCode generation for FILLY scripts (.TFY files).
package compiler

import (
	"reflect"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
)

// TestCompileSimpleAssignment tests simple variable assignment (x = value).
func TestCompileSimpleAssignment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple integer assignment",
			input: "x = 5",
			expected: []OpCode{
				{
					Cmd:  OpAssign,
					Args: []any{Variable("x"), int64(5)},
				},
			},
		},
		{
			name:  "simple string assignment",
			input: `s = "hello"`,
			expected: []OpCode{
				{
					Cmd:  OpAssign,
					Args: []any{Variable("s"), "hello"},
				},
			},
		},
		{
			name:  "assignment with binary expression",
			input: "x = 5 + 3",
			expected: []OpCode{
				{
					Cmd: OpAssign,
					Args: []any{
						Variable("x"),
						OpCode{
							Cmd:  OpBinaryOp,
							Args: []any{"+", int64(5), int64(3)},
						},
					},
				},
			},
		},
		{
			name:  "assignment with variable reference",
			input: "y = x",
			expected: []OpCode{
				{
					Cmd:  OpAssign,
					Args: []any{Variable("y"), Variable("x")},
				},
			},
		},
		{
			name:  "assignment with complex expression",
			input: "result = a * b + c",
			expected: []OpCode{
				{
					Cmd: OpAssign,
					Args: []any{
						Variable("result"),
						OpCode{
							Cmd: OpBinaryOp,
							Args: []any{
								"+",
								OpCode{
									Cmd:  OpBinaryOp,
									Args: []any{"*", Variable("a"), Variable("b")},
								},
								Variable("c"),
							},
						},
					},
				},
			},
		},
		{
			name:  "assignment with unary expression",
			input: "x = -5",
			expected: []OpCode{
				{
					Cmd: OpAssign,
					Args: []any{
						Variable("x"),
						OpCode{
							Cmd:  OpUnaryOp,
							Args: []any{"-", int64(5)},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileArrayAssignment tests array element assignment (arr[i] = value).
func TestCompileArrayAssignment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "array assignment with integer index",
			input: "arr[0] = 10",
			expected: []OpCode{
				{
					Cmd:  OpArrayAssign,
					Args: []any{Variable("arr"), int64(0), int64(10)},
				},
			},
		},
		{
			name:  "array assignment with variable index",
			input: "arr[i] = value",
			expected: []OpCode{
				{
					Cmd:  OpArrayAssign,
					Args: []any{Variable("arr"), Variable("i"), Variable("value")},
				},
			},
		},
		{
			name:  "array assignment with expression index",
			input: "arr[i + 1] = x",
			expected: []OpCode{
				{
					Cmd: OpArrayAssign,
					Args: []any{
						Variable("arr"),
						OpCode{
							Cmd:  OpBinaryOp,
							Args: []any{"+", Variable("i"), int64(1)},
						},
						Variable("x"),
					},
				},
			},
		},
		{
			name:  "array assignment with expression value",
			input: "data[idx] = a + b",
			expected: []OpCode{
				{
					Cmd: OpArrayAssign,
					Args: []any{
						Variable("data"),
						Variable("idx"),
						OpCode{
							Cmd:  OpBinaryOp,
							Args: []any{"+", Variable("a"), Variable("b")},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileVarDeclaration tests that variable declarations don't generate OpCode.
func TestCompileVarDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:     "simple int declaration",
			input:    "int x;",
			expected: []OpCode{},
		},
		{
			name:     "multiple int declarations",
			input:    "int x, y, z;",
			expected: []OpCode{},
		},
		{
			name:     "array declaration",
			input:    "int arr[];",
			expected: []OpCode{},
		},
		{
			name:     "array declaration with size",
			input:    "int arr[10];",
			expected: []OpCode{},
		},
		{
			name:     "string declaration",
			input:    "str s;",
			expected: []OpCode{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if len(opcodes) != len(tt.expected) {
				t.Errorf("expected %d opcodes, got %d: %#v", len(tt.expected), len(opcodes), opcodes)
			}
		})
	}
}

// TestCompileMixedStatements tests compilation of mixed statements including assignments.
func TestCompileMixedStatements(t *testing.T) {
	input := `
		int x;
		x = 10;
		int arr[];
		arr[0] = x + 5;
	`

	expected := []OpCode{
		// int x; - no OpCode
		// x = 10;
		{
			Cmd:  OpAssign,
			Args: []any{Variable("x"), int64(10)},
		},
		// int arr[]; - no OpCode
		// arr[0] = x + 5;
		{
			Cmd: OpArrayAssign,
			Args: []any{
				Variable("arr"),
				int64(0),
				OpCode{
					Cmd:  OpBinaryOp,
					Args: []any{"+", Variable("x"), int64(5)},
				},
			},
		},
	}

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()
	if len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)
	if len(compileErrs) > 0 {
		t.Fatalf("compiler errors: %v", compileErrs)
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, expected)
	}
}

// TestCompileFunctionCall tests function call OpCode generation.
func TestCompileFunctionCall(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple function call with string argument",
			input: `LoadPic("image.bmp");`,
			expected: []OpCode{
				{
					Cmd:  OpCall,
					Args: []any{"LoadPic", "image.bmp"},
				},
			},
		},
		{
			name:  "function call with no arguments",
			input: `del_me();`,
			expected: []OpCode{
				{
					Cmd:  OpCall,
					Args: []any{"del_me"},
				},
			},
		},
		{
			name:  "function call with multiple arguments",
			input: `MovePic(src, 0, 0, 100, 100, dst, 0, 0);`,
			expected: []OpCode{
				{
					Cmd: OpCall,
					Args: []any{
						"MovePic",
						Variable("src"),
						int64(0), int64(0), int64(100), int64(100),
						Variable("dst"),
						int64(0), int64(0),
					},
				},
			},
		},
		{
			name:  "function call with expression argument",
			input: `SetValue(x + 1);`,
			expected: []OpCode{
				{
					Cmd: OpCall,
					Args: []any{
						"SetValue",
						OpCode{
							Cmd:  OpBinaryOp,
							Args: []any{"+", Variable("x"), int64(1)},
						},
					},
				},
			},
		},
		{
			name:  "function call with array access argument",
			input: `Process(arr[i]);`,
			expected: []OpCode{
				{
					Cmd: OpCall,
					Args: []any{
						"Process",
						OpCode{
							Cmd:  OpArrayAccess,
							Args: []any{Variable("arr"), Variable("i")},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileFunctionDefinition tests function definition OpCode generation.
func TestCompileFunctionDefinition(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple function definition",
			input: `myFunc() { x = 1; }`,
			expected: []OpCode{
				{
					Cmd: OpDefineFunction,
					Args: []any{
						"myFunc",
						[]any{},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("x"), int64(1)}},
						},
					},
				},
			},
		},
		{
			name:  "function with parameters",
			input: `add(int a, int b) { result = a + b; }`,
			expected: []OpCode{
				{
					Cmd: OpDefineFunction,
					Args: []any{
						"add",
						[]any{
							map[string]any{"name": "a", "type": "int", "isArray": false},
							map[string]any{"name": "b", "type": "int", "isArray": false},
						},
						[]OpCode{
							{
								Cmd: OpAssign,
								Args: []any{
									Variable("result"),
									OpCode{
										Cmd:  OpBinaryOp,
										Args: []any{"+", Variable("a"), Variable("b")},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "function with default parameter",
			input: `greet(int count=1) { x = count; }`,
			expected: []OpCode{
				{
					Cmd: OpDefineFunction,
					Args: []any{
						"greet",
						[]any{
							map[string]any{"name": "count", "type": "int", "isArray": false, "default": int64(1)},
						},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("x"), Variable("count")}},
						},
					},
				},
			},
		},
		{
			name:  "function with array parameter",
			input: `process(int arr[]) { x = 0; }`,
			expected: []OpCode{
				{
					Cmd: OpDefineFunction,
					Args: []any{
						"process",
						[]any{
							map[string]any{"name": "arr", "type": "int", "isArray": true},
						},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("x"), int64(0)}},
						},
					},
				},
			},
		},
		{
			name:  "function with function call in body",
			input: `wrapper() { innerFunc(); }`,
			expected: []OpCode{
				{
					Cmd: OpDefineFunction,
					Args: []any{
						"wrapper",
						[]any{},
						[]OpCode{
							{Cmd: OpCall, Args: []any{"innerFunc"}},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileMixedFunctionCallsAndAssignments tests compilation of mixed statements.
func TestCompileMixedFunctionCallsAndAssignments(t *testing.T) {
	input := `
		int x;
		x = 10;
		LoadPic("test.bmp");
		y = x + 5;
		Process(y);
	`

	expected := []OpCode{
		// int x; - no OpCode
		// x = 10;
		{
			Cmd:  OpAssign,
			Args: []any{Variable("x"), int64(10)},
		},
		// LoadPic("test.bmp");
		{
			Cmd:  OpCall,
			Args: []any{"LoadPic", "test.bmp"},
		},
		// y = x + 5;
		{
			Cmd: OpAssign,
			Args: []any{
				Variable("y"),
				OpCode{
					Cmd:  OpBinaryOp,
					Args: []any{"+", Variable("x"), int64(5)},
				},
			},
		},
		// Process(y);
		{
			Cmd:  OpCall,
			Args: []any{"Process", Variable("y")},
		},
	}

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()
	if len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)
	if len(compileErrs) > 0 {
		t.Fatalf("compiler errors: %v", compileErrs)
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, expected)
	}
}

// TestCompileIfStatement tests if statement OpCode generation.
func TestCompileIfStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple if statement",
			input: `if (x > 5) { y = 10; }`,
			expected: []OpCode{
				{
					Cmd: OpIf,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("x"), int64(5)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("y"), int64(10)}},
						},
						[]OpCode{},
					},
				},
			},
		},
		{
			name:  "if-else statement",
			input: `if (x > 5) { y = 10; } else { y = 0; }`,
			expected: []OpCode{
				{
					Cmd: OpIf,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("x"), int64(5)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("y"), int64(10)}},
						},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("y"), int64(0)}},
						},
					},
				},
			},
		},
		{
			name:  "if with equality condition",
			input: `if (x == 0) { result = 1; }`,
			expected: []OpCode{
				{
					Cmd: OpIf,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{"==", Variable("x"), int64(0)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("result"), int64(1)}},
						},
						[]OpCode{},
					},
				},
			},
		},
		{
			name:  "if with function call in body",
			input: `if (flag) { doSomething(); }`,
			expected: []OpCode{
				{
					Cmd: OpIf,
					Args: []any{
						Variable("flag"),
						[]OpCode{
							{Cmd: OpCall, Args: []any{"doSomething"}},
						},
						[]OpCode{},
					},
				},
			},
		},
		{
			name:  "if-else if-else chain",
			input: `if (x > 10) { y = 1; } else if (x > 5) { y = 2; } else { y = 3; }`,
			expected: []OpCode{
				{
					Cmd: OpIf,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("x"), int64(10)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("y"), int64(1)}},
						},
						[]OpCode{
							{
								Cmd: OpIf,
								Args: []any{
									OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("x"), int64(5)}},
									[]OpCode{
										{Cmd: OpAssign, Args: []any{Variable("y"), int64(2)}},
									},
									[]OpCode{
										{Cmd: OpAssign, Args: []any{Variable("y"), int64(3)}},
									},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "nested if statements",
			input: `if (a > 0) { if (b > 0) { c = 1; } }`,
			expected: []OpCode{
				{
					Cmd: OpIf,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("a"), int64(0)}},
						[]OpCode{
							{
								Cmd: OpIf,
								Args: []any{
									OpCode{Cmd: OpBinaryOp, Args: []any{">", Variable("b"), int64(0)}},
									[]OpCode{
										{Cmd: OpAssign, Args: []any{Variable("c"), int64(1)}},
									},
									[]OpCode{},
								},
							},
						},
						[]OpCode{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileForStatement tests for loop OpCode generation.
func TestCompileForStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple for loop",
			input: `for (i = 0; i < 10; i = i + 1) { x = i; }`,
			expected: []OpCode{
				{
					Cmd: OpFor,
					Args: []any{
						// init
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("i"), int64(0)}},
						},
						// condition
						OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("i"), int64(10)}},
						// post
						[]OpCode{
							{Cmd: OpAssign, Args: []any{
								Variable("i"),
								OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("i"), int64(1)}},
							}},
						},
						// body
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("x"), Variable("i")}},
						},
					},
				},
			},
		},
		{
			name:  "for loop with function call in body",
			input: `for (j = 0; j < 5; j = j + 1) { process(j); }`,
			expected: []OpCode{
				{
					Cmd: OpFor,
					Args: []any{
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("j"), int64(0)}},
						},
						OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("j"), int64(5)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{
								Variable("j"),
								OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("j"), int64(1)}},
							}},
						},
						[]OpCode{
							{Cmd: OpCall, Args: []any{"process", Variable("j")}},
						},
					},
				},
			},
		},
		{
			name:  "for loop with array access",
			input: `for (k = 0; k < n; k = k + 1) { arr[k] = k; }`,
			expected: []OpCode{
				{
					Cmd: OpFor,
					Args: []any{
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("k"), int64(0)}},
						},
						OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("k"), Variable("n")}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{
								Variable("k"),
								OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("k"), int64(1)}},
							}},
						},
						[]OpCode{
							{Cmd: OpArrayAssign, Args: []any{Variable("arr"), Variable("k"), Variable("k")}},
						},
					},
				},
			},
		},
		{
			name:  "for loop with break",
			input: `for (i = 0; i < 10; i = i + 1) { if (i == 5) { break; } }`,
			expected: []OpCode{
				{
					Cmd: OpFor,
					Args: []any{
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("i"), int64(0)}},
						},
						OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("i"), int64(10)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{
								Variable("i"),
								OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("i"), int64(1)}},
							}},
						},
						[]OpCode{
							{
								Cmd: OpIf,
								Args: []any{
									OpCode{Cmd: OpBinaryOp, Args: []any{"==", Variable("i"), int64(5)}},
									[]OpCode{
										{Cmd: OpBreak, Args: []any{}},
									},
									[]OpCode{},
								},
							},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileWhileStatement tests while loop OpCode generation.
func TestCompileWhileStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple while loop",
			input: `while (x < 10) { x = x + 1; }`,
			expected: []OpCode{
				{
					Cmd: OpWhile,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("x"), int64(10)}},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{
								Variable("x"),
								OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("x"), int64(1)}},
							}},
						},
					},
				},
			},
		},
		{
			name:  "while loop with variable condition",
			input: `while (running) { process(); }`,
			expected: []OpCode{
				{
					Cmd: OpWhile,
					Args: []any{
						Variable("running"),
						[]OpCode{
							{Cmd: OpCall, Args: []any{"process"}},
						},
					},
				},
			},
		},
		{
			name:  "while loop with break",
			input: `while (1) { if (done) { break; } }`,
			expected: []OpCode{
				{
					Cmd: OpWhile,
					Args: []any{
						int64(1),
						[]OpCode{
							{
								Cmd: OpIf,
								Args: []any{
									Variable("done"),
									[]OpCode{
										{Cmd: OpBreak, Args: []any{}},
									},
									[]OpCode{},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "while loop with continue",
			input: `while (i < 10) { if (i == 5) { continue; } x = i; }`,
			expected: []OpCode{
				{
					Cmd: OpWhile,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{"<", Variable("i"), int64(10)}},
						[]OpCode{
							{
								Cmd: OpIf,
								Args: []any{
									OpCode{Cmd: OpBinaryOp, Args: []any{"==", Variable("i"), int64(5)}},
									[]OpCode{
										{Cmd: OpContinue, Args: []any{}},
									},
									[]OpCode{},
								},
							},
							{Cmd: OpAssign, Args: []any{Variable("x"), Variable("i")}},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileSwitchStatement tests switch statement OpCode generation.
func TestCompileSwitchStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple switch statement",
			input: `switch (x) { case 1: y = 10; }`,
			expected: []OpCode{
				{
					Cmd: OpSwitch,
					Args: []any{
						Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("y"), int64(10)}},
								},
							},
						},
						[]OpCode{},
					},
				},
			},
		},
		{
			name:  "switch with multiple cases",
			input: `switch (x) { case 1: y = 10; case 2: y = 20; }`,
			expected: []OpCode{
				{
					Cmd: OpSwitch,
					Args: []any{
						Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("y"), int64(10)}},
								},
							},
							map[string]any{
								"value": int64(2),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("y"), int64(20)}},
								},
							},
						},
						[]OpCode{},
					},
				},
			},
		},
		{
			name:  "switch with default",
			input: `switch (x) { case 1: y = 10; default: y = 0; }`,
			expected: []OpCode{
				{
					Cmd: OpSwitch,
					Args: []any{
						Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("y"), int64(10)}},
								},
							},
						},
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("y"), int64(0)}},
						},
					},
				},
			},
		},
		{
			name:  "switch with expression value",
			input: `switch (a + b) { case 0: result = 1; }`,
			expected: []OpCode{
				{
					Cmd: OpSwitch,
					Args: []any{
						OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("a"), Variable("b")}},
						[]any{
							map[string]any{
								"value": int64(0),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("result"), int64(1)}},
								},
							},
						},
						[]OpCode{},
					},
				},
			},
		},
		{
			name:  "switch with break in case",
			input: `switch (x) { case 1: y = 10; break; case 2: y = 20; }`,
			expected: []OpCode{
				{
					Cmd: OpSwitch,
					Args: []any{
						Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("y"), int64(10)}},
									{Cmd: OpBreak, Args: []any{}},
								},
							},
							map[string]any{
								"value": int64(2),
								"body": []OpCode{
									{Cmd: OpAssign, Args: []any{Variable("y"), int64(20)}},
								},
							},
						},
						[]OpCode{},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileBreakContinue tests break and continue statement OpCode generation.
func TestCompileBreakContinue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "break statement",
			input: `break;`,
			expected: []OpCode{
				{Cmd: OpBreak, Args: []any{}},
			},
		},
		{
			name:  "continue statement",
			input: `continue;`,
			expected: []OpCode{
				{Cmd: OpContinue, Args: []any{}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileMesStatement tests mes (event handler) statement OpCode generation.
func TestCompileMesStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "simple mes statement with MIDI_TIME",
			input: `mes(MIDI_TIME) { x = 1; }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"MIDI_TIME",
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("x"), int64(1)}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with TIME event",
			input: `mes(TIME) { process(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"TIME",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"process"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with KEY event",
			input: `mes(KEY) { handleKey(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"KEY",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"handleKey"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with CLICK event",
			input: `mes(CLICK) { onClick(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"CLICK",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"onClick"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with MIDI_END event",
			input: `mes(MIDI_END) { cleanup(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"MIDI_END",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"cleanup"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with USER event",
			input: `mes(USER) { userHandler(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"USER",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"userHandler"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with empty body",
			input: `mes(MIDI_TIME) { }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"MIDI_TIME",
						[]OpCode(nil),
					},
				},
			},
		},
		{
			name:  "mes statement with multiple statements in body",
			input: `mes(MIDI_TIME) { x = 1; y = 2; process(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"MIDI_TIME",
						[]OpCode{
							{Cmd: OpAssign, Args: []any{Variable("x"), int64(1)}},
							{Cmd: OpAssign, Args: []any{Variable("y"), int64(2)}},
							{Cmd: OpCall, Args: []any{"process"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with if statement in body",
			input: `mes(KEY) { if (key == 27) { exit(); } }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"KEY",
						[]OpCode{
							{
								Cmd: OpIf,
								Args: []any{
									OpCode{Cmd: OpBinaryOp, Args: []any{"==", Variable("key"), int64(27)}},
									[]OpCode{
										{Cmd: OpCall, Args: []any{"exit"}},
									},
									[]OpCode{},
								},
							},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with RBDOWN event",
			input: `mes(RBDOWN) { rightClick(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"RBDOWN",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"rightClick"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with RBDBLCLK event",
			input: `mes(RBDBLCLK) { rightDoubleClick(); }`,
			expected: []OpCode{
				{
					Cmd: OpRegisterEventHandler,
					Args: []any{
						"RBDBLCLK",
						[]OpCode{
							{Cmd: OpCall, Args: []any{"rightDoubleClick"}},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileStepStatement tests step statement OpCode generation.
func TestCompileStepStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []OpCode
	}{
		{
			name:  "step with count and single function call",
			input: `step(10) { func1(); }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(10)}},
				{Cmd: OpCall, Args: []any{"func1"}},
			},
		},
		{
			name:  "step with count and function call followed by wait",
			input: `step(10) { func1();, }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(10)}},
				{Cmd: OpCall, Args: []any{"func1"}},
				{Cmd: OpWait, Args: []any{1}},
			},
		},
		{
			name:  "step with multiple function calls and waits",
			input: `step(10) { func1();, func2();,, }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(10)}},
				{Cmd: OpCall, Args: []any{"func1"}},
				{Cmd: OpWait, Args: []any{1}},
				{Cmd: OpCall, Args: []any{"func2"}},
				{Cmd: OpWait, Args: []any{2}},
			},
		},
		{
			// end_step is a marker that stops comma counting, not a function call
			// The parser skips end_step and continues parsing remaining statements
			name:  "step with end_step and del_me",
			input: `step(10) { func1();, func2();,, end_step; del_me; }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(10)}},
				{Cmd: OpCall, Args: []any{"func1"}},
				{Cmd: OpWait, Args: []any{1}},
				{Cmd: OpCall, Args: []any{"func2"}},
				{Cmd: OpWait, Args: []any{2}},
				// end_step is skipped by parser (it's a marker, not a command)
				{Cmd: OpCall, Args: []any{"del_me"}},
			},
		},
		{
			name:  "step without count",
			input: `step { func1();, }`,
			expected: []OpCode{
				{Cmd: OpCall, Args: []any{"func1"}},
				{Cmd: OpWait, Args: []any{1}},
			},
		},
		{
			name:  "step with variable count",
			input: `step(n) { process(); }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{Variable("n")}},
				{Cmd: OpCall, Args: []any{"process"}},
			},
		},
		{
			name:  "step with expression count",
			input: `step(x + 1) { doWork(); }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{OpCode{Cmd: OpBinaryOp, Args: []any{"+", Variable("x"), int64(1)}}}},
				{Cmd: OpCall, Args: []any{"doWork"}},
			},
		},
		{
			name:  "step with assignment in body",
			input: `step(5) { x = 10;, }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(5)}},
				{Cmd: OpAssign, Args: []any{Variable("x"), int64(10)}},
				{Cmd: OpWait, Args: []any{1}},
			},
		},
		{
			name:  "step with multiple consecutive waits",
			input: `step(8) { func1();,,, func2(); }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(8)}},
				{Cmd: OpCall, Args: []any{"func1"}},
				{Cmd: OpWait, Args: []any{3}},
				{Cmd: OpCall, Args: []any{"func2"}},
			},
		},
		{
			name:  "step with function call with arguments",
			input: `step(16) { MovePic(src, 0, 0);, }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(16)}},
				{Cmd: OpCall, Args: []any{"MovePic", Variable("src"), int64(0), int64(0)}},
				{Cmd: OpWait, Args: []any{1}},
			},
		},
		{
			name:  "step with empty body",
			input: `step(10) { }`,
			expected: []OpCode{
				{Cmd: OpSetStep, Args: []any{int64(10)}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := lexer.New(tt.input)
			p := parser.New(l)
			program, errs := p.ParseProgram()
			if len(errs) > 0 {
				t.Fatalf("parser errors: %v", errs)
			}

			c := New()
			opcodes, compileErrs := c.Compile(program)
			if len(compileErrs) > 0 {
				t.Fatalf("compiler errors: %v", compileErrs)
			}

			if !reflect.DeepEqual(opcodes, tt.expected) {
				t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, tt.expected)
			}
		})
	}
}

// TestCompileStepStatementInMes tests step statement inside mes block.
func TestCompileStepStatementInMes(t *testing.T) {
	input := `mes(MIDI_TIME) { step(10) { func1();, func2();,, del_me; } }`

	expected := []OpCode{
		{
			Cmd: OpRegisterEventHandler,
			Args: []any{
				"MIDI_TIME",
				[]OpCode{
					{Cmd: OpSetStep, Args: []any{int64(10)}},
					{Cmd: OpCall, Args: []any{"func1"}},
					{Cmd: OpWait, Args: []any{1}},
					{Cmd: OpCall, Args: []any{"func2"}},
					{Cmd: OpWait, Args: []any{2}},
					{Cmd: OpCall, Args: []any{"del_me"}},
				},
			},
		},
	}

	l := lexer.New(input)
	p := parser.New(l)
	program, errs := p.ParseProgram()
	if len(errs) > 0 {
		t.Fatalf("parser errors: %v", errs)
	}

	c := New()
	opcodes, compileErrs := c.Compile(program)
	if len(compileErrs) > 0 {
		t.Fatalf("compiler errors: %v", compileErrs)
	}

	if !reflect.DeepEqual(opcodes, expected) {
		t.Errorf("opcodes mismatch:\ngot:      %#v\nexpected: %#v", opcodes, expected)
	}
}
