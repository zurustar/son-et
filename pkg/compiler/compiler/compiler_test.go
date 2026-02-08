// Package compiler provides OpCode generation for FILLY scripts (.TFY files).
package compiler

import (
	"reflect"
	"testing"

	"github.com/zurustar/son-et/pkg/compiler/lexer"
	"github.com/zurustar/son-et/pkg/compiler/parser"
	"github.com/zurustar/son-et/pkg/opcode"
)

// TestCompileSimpleAssignment tests simple variable assignment (x = value).
func TestCompileSimpleAssignment(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []opcode.OpCode
	}{
		{
			name:  "simple integer assignment",
			input: "x = 5",
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("x"), int64(5)},
				},
			},
		},
		{
			name:  "simple string assignment",
			input: `s = "hello"`,
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("s"), "hello"},
				},
			},
		},
		{
			name:  "assignment with binary expression",
			input: "x = 5 + 3",
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Assign,
					Args: []any{
						opcode.Variable("x"),
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", int64(5), int64(3)},
						},
					},
				},
			},
		},
		{
			name:  "assignment with variable reference",
			input: "y = x",
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.Assign,
					Args: []any{opcode.Variable("y"), opcode.Variable("x")},
				},
			},
		},
		{
			name:  "assignment with complex expression",
			input: "result = a * b + c",
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Assign,
					Args: []any{
						opcode.Variable("result"),
						opcode.OpCode{
							Cmd: opcode.BinaryOp,
							Args: []any{
								"+",
								opcode.OpCode{
									Cmd:  opcode.BinaryOp,
									Args: []any{"*", opcode.Variable("a"), opcode.Variable("b")},
								},
								opcode.Variable("c"),
							},
						},
					},
				},
			},
		},
		{
			name:  "assignment with unary expression",
			input: "x = -5",
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Assign,
					Args: []any{
						opcode.Variable("x"),
						opcode.OpCode{
							Cmd:  opcode.UnaryOp,
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
		expected []opcode.OpCode
	}{
		{
			name:  "array assignment with integer index",
			input: "arr[0] = 10",
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.ArrayAssign,
					Args: []any{opcode.Variable("arr"), int64(0), int64(10)},
				},
			},
		},
		{
			name:  "array assignment with variable index",
			input: "arr[i] = value",
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.ArrayAssign,
					Args: []any{opcode.Variable("arr"), opcode.Variable("i"), opcode.Variable("value")},
				},
			},
		},
		{
			name:  "array assignment with expression index",
			input: "arr[i + 1] = x",
			expected: []opcode.OpCode{
				{
					Cmd: opcode.ArrayAssign,
					Args: []any{
						opcode.Variable("arr"),
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", opcode.Variable("i"), int64(1)},
						},
						opcode.Variable("x"),
					},
				},
			},
		},
		{
			name:  "array assignment with expression value",
			input: "data[idx] = a + b",
			expected: []opcode.OpCode{
				{
					Cmd: opcode.ArrayAssign,
					Args: []any{
						opcode.Variable("data"),
						opcode.Variable("idx"),
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", opcode.Variable("a"), opcode.Variable("b")},
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

// TestCompileVarDeclaration tests that variable declarations generate initialization OpCodes.
func TestCompileVarDeclaration(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []opcode.OpCode
	}{
		{
			name:  "simple int declaration",
			input: "int x;",
			expected: []opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(0)}},
			},
		},
		{
			name:  "multiple int declarations",
			input: "int x, y, z;",
			expected: []opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(0)}},
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(0)}},
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("z"), int64(0)}},
			},
		},
		{
			name:  "array declaration",
			input: "int arr[];",
			expected: []opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("arr"), []any{}}},
			},
		},
		{
			name:  "array declaration with size",
			input: "int arr[10];",
			expected: []opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("arr"), []any{}}},
			},
		},
		{
			name:  "string declaration",
			input: "str s;",
			expected: []opcode.OpCode{
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("s"), ""}},
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

// TestCompileMixedStatements tests compilation of mixed statements including assignments.
func TestCompileMixedStatements(t *testing.T) {
	input := `
		int x;
		x = 10;
		int arr[];
		arr[0] = x + 5;
	`

	expected := []opcode.OpCode{
		// int x; - initializes x to 0
		{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x"), int64(0)},
		},
		// x = 10;
		{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x"), int64(10)},
		},
		// int arr[]; - initializes arr to empty array
		{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("arr"), []any{}},
		},
		// arr[0] = x + 5;
		{
			Cmd: opcode.ArrayAssign,
			Args: []any{
				opcode.Variable("arr"),
				int64(0),
				opcode.OpCode{
					Cmd:  opcode.BinaryOp,
					Args: []any{"+", opcode.Variable("x"), int64(5)},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple function call with string argument",
			input: `LoadPic("image.bmp");`,
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"LoadPic", "image.bmp"},
				},
			},
		},
		{
			name:  "function call with no arguments",
			input: `del_me();`,
			expected: []opcode.OpCode{
				{
					Cmd:  opcode.Call,
					Args: []any{"del_me"},
				},
			},
		},
		{
			name:  "function call with multiple arguments",
			input: `MovePic(src, 0, 0, 100, 100, dst, 0, 0);`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Call,
					Args: []any{
						"MovePic",
						opcode.Variable("src"),
						int64(0), int64(0), int64(100), int64(100),
						opcode.Variable("dst"),
						int64(0), int64(0),
					},
				},
			},
		},
		{
			name:  "function call with expression argument",
			input: `SetValue(x + 1);`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Call,
					Args: []any{
						"SetValue",
						opcode.OpCode{
							Cmd:  opcode.BinaryOp,
							Args: []any{"+", opcode.Variable("x"), int64(1)},
						},
					},
				},
			},
		},
		{
			name:  "function call with array access argument",
			input: `Process(arr[i]);`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Call,
					Args: []any{
						"Process",
						opcode.OpCode{
							Cmd:  opcode.ArrayAccess,
							Args: []any{opcode.Variable("arr"), opcode.Variable("i")},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple function definition",
			input: `myFunc() { x = 1; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.DefineFunction,
					Args: []any{
						"myFunc",
						[]any{},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
						},
					},
				},
			},
		},
		{
			name:  "function with parameters",
			input: `add(int a, int b) { result = a + b; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.DefineFunction,
					Args: []any{
						"add",
						[]any{
							map[string]any{"name": "a", "type": "int", "isArray": false},
							map[string]any{"name": "b", "type": "int", "isArray": false},
						},
						[]opcode.OpCode{
							{
								Cmd: opcode.Assign,
								Args: []any{
									opcode.Variable("result"),
									opcode.OpCode{
										Cmd:  opcode.BinaryOp,
										Args: []any{"+", opcode.Variable("a"), opcode.Variable("b")},
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
			expected: []opcode.OpCode{
				{
					Cmd: opcode.DefineFunction,
					Args: []any{
						"greet",
						[]any{
							map[string]any{"name": "count", "type": "int", "isArray": false, "default": int64(1)},
						},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), opcode.Variable("count")}},
						},
					},
				},
			},
		},
		{
			name:  "function with array parameter",
			input: `process(int arr[]) { x = 0; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.DefineFunction,
					Args: []any{
						"process",
						[]any{
							map[string]any{"name": "arr", "type": "int", "isArray": true},
						},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(0)}},
						},
					},
				},
			},
		},
		{
			name:  "function with function call in body",
			input: `wrapper() { innerFunc(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.DefineFunction,
					Args: []any{
						"wrapper",
						[]any{},
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"innerFunc"}},
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

	expected := []opcode.OpCode{
		// int x; - initializes x to 0
		{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x"), int64(0)},
		},
		// x = 10;
		{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable("x"), int64(10)},
		},
		// LoadPic("test.bmp");
		{
			Cmd:  opcode.Call,
			Args: []any{"LoadPic", "test.bmp"},
		},
		// y = x + 5;
		{
			Cmd: opcode.Assign,
			Args: []any{
				opcode.Variable("y"),
				opcode.OpCode{
					Cmd:  opcode.BinaryOp,
					Args: []any{"+", opcode.Variable("x"), int64(5)},
				},
			},
		},
		// Process(y);
		{
			Cmd:  opcode.Call,
			Args: []any{"Process", opcode.Variable("y")},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple if statement",
			input: `if (x > 5) { y = 10; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("x"), int64(5)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}},
						},
						[]opcode.OpCode{},
					},
				},
			},
		},
		{
			name:  "if-else statement",
			input: `if (x > 5) { y = 10; } else { y = 0; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("x"), int64(5)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}},
						},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(0)}},
						},
					},
				},
			},
		},
		{
			name:  "if with equality condition",
			input: `if (x == 0) { result = 1; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("x"), int64(0)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("result"), int64(1)}},
						},
						[]opcode.OpCode{},
					},
				},
			},
		},
		{
			name:  "if with function call in body",
			input: `if (flag) { doSomething(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.Variable("flag"),
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"doSomething"}},
						},
						[]opcode.OpCode{},
					},
				},
			},
		},
		{
			name:  "if-else if-else chain",
			input: `if (x > 10) { y = 1; } else if (x > 5) { y = 2; } else { y = 3; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("x"), int64(10)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(1)}},
						},
						[]opcode.OpCode{
							{
								Cmd: opcode.If,
								Args: []any{
									opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("x"), int64(5)}},
									[]opcode.OpCode{
										{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
									},
									[]opcode.OpCode{
										{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(3)}},
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
			expected: []opcode.OpCode{
				{
					Cmd: opcode.If,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("a"), int64(0)}},
						[]opcode.OpCode{
							{
								Cmd: opcode.If,
								Args: []any{
									opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("b"), int64(0)}},
									[]opcode.OpCode{
										{Cmd: opcode.Assign, Args: []any{opcode.Variable("c"), int64(1)}},
									},
									[]opcode.OpCode{},
								},
							},
						},
						[]opcode.OpCode{},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple for loop",
			input: `for (i = 0; i < 10; i = i + 1) { x = i; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.For,
					Args: []any{
						// init
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
						},
						// condition
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(10)}},
						// post
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{
								opcode.Variable("i"),
								opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
							}},
						},
						// body
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), opcode.Variable("i")}},
						},
					},
				},
			},
		},
		{
			name:  "for loop with function call in body",
			input: `for (j = 0; j < 5; j = j + 1) { process(j); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.For,
					Args: []any{
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("j"), int64(0)}},
						},
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("j"), int64(5)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{
								opcode.Variable("j"),
								opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("j"), int64(1)}},
							}},
						},
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"process", opcode.Variable("j")}},
						},
					},
				},
			},
		},
		{
			name:  "for loop with array access",
			input: `for (k = 0; k < n; k = k + 1) { arr[k] = k; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.For,
					Args: []any{
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("k"), int64(0)}},
						},
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("k"), opcode.Variable("n")}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{
								opcode.Variable("k"),
								opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("k"), int64(1)}},
							}},
						},
						[]opcode.OpCode{
							{Cmd: opcode.ArrayAssign, Args: []any{opcode.Variable("arr"), opcode.Variable("k"), opcode.Variable("k")}},
						},
					},
				},
			},
		},
		{
			name:  "for loop with break",
			input: `for (i = 0; i < 10; i = i + 1) { if (i == 5) { break; } }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.For,
					Args: []any{
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), int64(0)}},
						},
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(10)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{
								opcode.Variable("i"),
								opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), int64(1)}},
							}},
						},
						[]opcode.OpCode{
							{
								Cmd: opcode.If,
								Args: []any{
									opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(5)}},
									[]opcode.OpCode{
										{Cmd: opcode.Break, Args: []any{}},
									},
									[]opcode.OpCode{},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple while loop",
			input: `while (x < 10) { x = x + 1; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.While,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("x"), int64(10)}},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{
								opcode.Variable("x"),
								opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("x"), int64(1)}},
							}},
						},
					},
				},
			},
		},
		{
			name:  "while loop with variable condition",
			input: `while (running) { process(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.While,
					Args: []any{
						opcode.Variable("running"),
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"process"}},
						},
					},
				},
			},
		},
		{
			name:  "while loop with break",
			input: `while (1) { if (done) { break; } }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.While,
					Args: []any{
						int64(1),
						[]opcode.OpCode{
							{
								Cmd: opcode.If,
								Args: []any{
									opcode.Variable("done"),
									[]opcode.OpCode{
										{Cmd: opcode.Break, Args: []any{}},
									},
									[]opcode.OpCode{},
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
			expected: []opcode.OpCode{
				{
					Cmd: opcode.While,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), int64(10)}},
						[]opcode.OpCode{
							{
								Cmd: opcode.If,
								Args: []any{
									opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("i"), int64(5)}},
									[]opcode.OpCode{
										{Cmd: opcode.Continue, Args: []any{}},
									},
									[]opcode.OpCode{},
								},
							},
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), opcode.Variable("i")}},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple switch statement",
			input: `switch (x) { case 1: y = 10; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Switch,
					Args: []any{
						opcode.Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}},
								},
							},
						},
						[]opcode.OpCode{},
					},
				},
			},
		},
		{
			name:  "switch with multiple cases",
			input: `switch (x) { case 1: y = 10; case 2: y = 20; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Switch,
					Args: []any{
						opcode.Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}},
								},
							},
							map[string]any{
								"value": int64(2),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(20)}},
								},
							},
						},
						[]opcode.OpCode{},
					},
				},
			},
		},
		{
			name:  "switch with default",
			input: `switch (x) { case 1: y = 10; default: y = 0; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Switch,
					Args: []any{
						opcode.Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}},
								},
							},
						},
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(0)}},
						},
					},
				},
			},
		},
		{
			name:  "switch with expression value",
			input: `switch (a + b) { case 0: result = 1; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Switch,
					Args: []any{
						opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("a"), opcode.Variable("b")}},
						[]any{
							map[string]any{
								"value": int64(0),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("result"), int64(1)}},
								},
							},
						},
						[]opcode.OpCode{},
					},
				},
			},
		},
		{
			name:  "switch with break in case",
			input: `switch (x) { case 1: y = 10; break; case 2: y = 20; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.Switch,
					Args: []any{
						opcode.Variable("x"),
						[]any{
							map[string]any{
								"value": int64(1),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(10)}},
									{Cmd: opcode.Break, Args: []any{}},
								},
							},
							map[string]any{
								"value": int64(2),
								"body": []opcode.OpCode{
									{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(20)}},
								},
							},
						},
						[]opcode.OpCode{},
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
		expected []opcode.OpCode
	}{
		{
			name:  "break statement",
			input: `break;`,
			expected: []opcode.OpCode{
				{Cmd: opcode.Break, Args: []any{}},
			},
		},
		{
			name:  "continue statement",
			input: `continue;`,
			expected: []opcode.OpCode{
				{Cmd: opcode.Continue, Args: []any{}},
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
		expected []opcode.OpCode
	}{
		{
			name:  "simple mes statement with MIDI_TIME",
			input: `mes(MIDI_TIME) { x = 1; }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"MIDI_TIME",
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with TIME event",
			input: `mes(TIME) { process(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"TIME",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"process"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with KEY event",
			input: `mes(KEY) { handleKey(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"KEY",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"handleKey"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with CLICK event",
			input: `mes(CLICK) { onClick(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"CLICK",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"onClick"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with MIDI_END event",
			input: `mes(MIDI_END) { cleanup(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"MIDI_END",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"cleanup"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with USER event",
			input: `mes(USER) { userHandler(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"USER",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"userHandler"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with empty body",
			input: `mes(MIDI_TIME) { }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"MIDI_TIME",
						[]opcode.OpCode(nil),
					},
				},
			},
		},
		{
			name:  "mes statement with multiple statements in body",
			input: `mes(MIDI_TIME) { x = 1; y = 2; process(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"MIDI_TIME",
						[]opcode.OpCode{
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(1)}},
							{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), int64(2)}},
							{Cmd: opcode.Call, Args: []any{"process"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with if statement in body",
			input: `mes(KEY) { if (key == 27) { exit(); } }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"KEY",
						[]opcode.OpCode{
							{
								Cmd: opcode.If,
								Args: []any{
									opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"==", opcode.Variable("key"), int64(27)}},
									[]opcode.OpCode{
										{Cmd: opcode.Call, Args: []any{"exit"}},
									},
									[]opcode.OpCode{},
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
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"RBDOWN",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"rightClick"}},
						},
					},
				},
			},
		},
		{
			name:  "mes statement with RBDBLCLK event",
			input: `mes(RBDBLCLK) { rightDoubleClick(); }`,
			expected: []opcode.OpCode{
				{
					Cmd: opcode.RegisterEventHandler,
					Args: []any{
						"RBDBLCLK",
						[]opcode.OpCode{
							{Cmd: opcode.Call, Args: []any{"rightDoubleClick"}},
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
		expected []opcode.OpCode
	}{
		{
			name:  "step with count and single function call",
			input: `step(10) { func1(); }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(10)}},
				{Cmd: opcode.Call, Args: []any{"func1"}},
			},
		},
		{
			name:  "step with count and function call followed by wait",
			input: `step(10) { func1();, }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(10)}},
				{Cmd: opcode.Call, Args: []any{"func1"}},
				{Cmd: opcode.Wait, Args: []any{1}},
			},
		},
		{
			name:  "step with multiple function calls and waits",
			input: `step(10) { func1();, func2();,, }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(10)}},
				{Cmd: opcode.Call, Args: []any{"func1"}},
				{Cmd: opcode.Wait, Args: []any{1}},
				{Cmd: opcode.Call, Args: []any{"func2"}},
				{Cmd: opcode.Wait, Args: []any{2}},
			},
		},
		{
			// end_step is a marker that stops comma counting, not a function call
			// The parser skips end_step and continues parsing remaining statements
			name:  "step with end_step and del_me",
			input: `step(10) { func1();, func2();,, end_step; del_me; }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(10)}},
				{Cmd: opcode.Call, Args: []any{"func1"}},
				{Cmd: opcode.Wait, Args: []any{1}},
				{Cmd: opcode.Call, Args: []any{"func2"}},
				{Cmd: opcode.Wait, Args: []any{2}},
				// end_step is skipped by parser (it's a marker, not a command)
				{Cmd: opcode.Call, Args: []any{"del_me"}},
			},
		},
		{
			name:  "step without count",
			input: `step { func1();, }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.Call, Args: []any{"func1"}},
				{Cmd: opcode.Wait, Args: []any{1}},
			},
		},
		{
			name:  "step with variable count",
			input: `step(n) { process(); }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{opcode.Variable("n")}},
				{Cmd: opcode.Call, Args: []any{"process"}},
			},
		},
		{
			name:  "step with expression count",
			input: `step(x + 1) { doWork(); }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("x"), int64(1)}}}},
				{Cmd: opcode.Call, Args: []any{"doWork"}},
			},
		},
		{
			name:  "step with assignment in body",
			input: `step(5) { x = 10;, }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(5)}},
				{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), int64(10)}},
				{Cmd: opcode.Wait, Args: []any{1}},
			},
		},
		{
			name:  "step with multiple consecutive waits",
			input: `step(8) { func1();,,, func2(); }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(8)}},
				{Cmd: opcode.Call, Args: []any{"func1"}},
				{Cmd: opcode.Wait, Args: []any{3}},
				{Cmd: opcode.Call, Args: []any{"func2"}},
			},
		},
		{
			name:  "step with function call with arguments",
			input: `step(16) { MovePic(src, 0, 0);, }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(16)}},
				{Cmd: opcode.Call, Args: []any{"MovePic", opcode.Variable("src"), int64(0), int64(0)}},
				{Cmd: opcode.Wait, Args: []any{1}},
			},
		},
		{
			name:  "step with empty body",
			input: `step(10) { }`,
			expected: []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{int64(10)}},
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

	expected := []opcode.OpCode{
		{
			Cmd: opcode.RegisterEventHandler,
			Args: []any{
				"MIDI_TIME",
				[]opcode.OpCode{
					{Cmd: opcode.SetStep, Args: []any{int64(10)}},
					{Cmd: opcode.Call, Args: []any{"func1"}},
					{Cmd: opcode.Wait, Args: []any{1}},
					{Cmd: opcode.Call, Args: []any{"func2"}},
					{Cmd: opcode.Wait, Args: []any{2}},
					{Cmd: opcode.Call, Args: []any{"del_me"}},
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
