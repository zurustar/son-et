package interpreter

import (
	"fmt"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/ast"
)

// Interpreter converts TFY scripts (AST) to OpCode sequences
type Interpreter struct {
	assets    []string             // Discovered asset files
	globals   map[string]bool      // Global variable names (case-insensitive)
	userFuncs map[string]*Function // User-defined functions
}

// NewInterpreter creates a new Interpreter instance
func NewInterpreter() *Interpreter {
	return &Interpreter{
		assets:    []string{},
		globals:   make(map[string]bool),
		userFuncs: make(map[string]*Function),
	}
}

// Interpret converts a TFY script (AST) to OpCode sequences
func (i *Interpreter) Interpret(program *ast.Program) (*Script, error) {
	script := NewScript()

	// First pass: collect global variables and function definitions
	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			// Global variable declaration
			varName := normalizeVarName(s.Name.Value)
			i.globals[varName] = true
			script.Globals[varName] = "int" // Default type

		case *ast.FunctionStatement:
			// User-defined function
			fn, err := i.interpretFunction(s)
			if err != nil {
				return nil, fmt.Errorf("error interpreting function %s: %w", s.Name.Value, err)
			}
			funcName := normalizeVarName(s.Name.Value)
			i.userFuncs[funcName] = fn
			script.Functions[funcName] = fn
		}
	}

	// Second pass: convert main function body to OpCode
	// Check if there's a user-defined main() function
	if mainUserFunc, ok := script.Functions["main"]; ok {
		// Use the user-defined main function
		script.Main = mainUserFunc
	} else {
		// No main() function - collect top-level statements
		mainFunc := NewFunction("main")
		for _, stmt := range program.Statements {
			// Skip function definitions and global declarations (already processed)
			if _, ok := stmt.(*ast.FunctionStatement); ok {
				continue
			}
			if _, ok := stmt.(*ast.LetStatement); ok {
				continue
			}

			// Convert statement to OpCode
			ops, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, fmt.Errorf("error interpreting statement: %w", err)
			}
			mainFunc.Body = append(mainFunc.Body, ops...)
		}
		script.Main = mainFunc
	}

	// Scan for assets
	script.Assets = i.scanAssets(program)

	return script, nil
}

// interpretFunction converts a function AST to OpCode
func (i *Interpreter) interpretFunction(fn *ast.FunctionStatement) (*Function, error) {
	funcName := normalizeVarName(fn.Name.Value)
	function := NewFunction(funcName)

	// Convert parameters
	for _, param := range fn.Parameters {
		p := Parameter{
			Name: normalizeVarName(param.Name.Value),
			Type: param.Type,
		}
		// Convert default value if present
		if param.Default != nil {
			val, err := i.evaluateConstant(param.Default)
			if err != nil {
				return nil, fmt.Errorf("error evaluating parameter default: %w", err)
			}
			p.Default = val
		}
		function.Parameters = append(function.Parameters, p)
		// Track as local variable
		function.Locals[p.Name] = p.Type
	}

	// Scan function body for local variable declarations
	i.scanLocalVariables(fn.Body, function)

	// Convert function body
	for _, stmt := range fn.Body.Statements {
		ops, err := i.interpretStatement(stmt)
		if err != nil {
			return nil, fmt.Errorf("error in function %s: %w", funcName, err)
		}
		function.Body = append(function.Body, ops...)
	}

	return function, nil
}

// scanLocalVariables scans a block for local variable declarations
func (i *Interpreter) scanLocalVariables(block *ast.BlockStatement, function *Function) {
	for _, stmt := range block.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			// Local variable declaration
			varName := normalizeVarName(s.Name.Value)
			function.Locals[varName] = "int" // Default type

		case *ast.BlockStatement:
			// Recursively scan nested blocks
			i.scanLocalVariables(s, function)

		case *ast.IfStatement:
			if s.Consequence != nil {
				i.scanLocalVariables(s.Consequence, function)
			}
			if s.Alternative != nil {
				i.scanLocalVariables(s.Alternative, function)
			}

		case *ast.ForStatement:
			if s.Body != nil {
				i.scanLocalVariables(s.Body, function)
			}

		case *ast.WhileStatement:
			if s.Body != nil {
				i.scanLocalVariables(s.Body, function)
			}

		case *ast.DoWhileStatement:
			if s.Body != nil {
				i.scanLocalVariables(s.Body, function)
			}

		case *ast.MesBlockStatement:
			if s.Body != nil {
				i.scanLocalVariables(s.Body, function)
			}

		case *ast.StepBlockStatement:
			if s.Body != nil {
				i.scanLocalVariables(s.Body, function)
			}

		case *ast.SwitchStatement:
			for _, c := range s.Cases {
				if c.Body != nil {
					i.scanLocalVariables(c.Body, function)
				}
			}
			if s.Default != nil {
				i.scanLocalVariables(s.Default, function)
			}
		}
	}
}

// interpretStatement converts a statement AST to OpCode
func (i *Interpreter) interpretStatement(stmt ast.Statement) ([]OpCode, error) {
	switch s := stmt.(type) {
	case *ast.ExpressionStatement:
		// Expression statement (e.g., function call)
		op, err := i.interpretExpression(s.Expression)
		if err != nil {
			return nil, err
		}
		return []OpCode{op}, nil

	case *ast.AssignStatement:
		// Variable assignment: VAR = VALUE or VAR[IDX] = VALUE
		varName := Variable(normalizeVarName(s.Name.Value))
		valueOp, err := i.interpretExpression(s.Value)
		if err != nil {
			return nil, err
		}

		if s.Index != nil {
			// Array assignment: VAR[IDX] = VALUE
			indexOp, err := i.interpretExpression(s.Index)
			if err != nil {
				return nil, err
			}
			return []OpCode{{
				Cmd:  OpAssignArray,
				Args: []any{varName, indexOp, valueOp},
			}}, nil
		}

		// Simple assignment: VAR = VALUE
		return []OpCode{{
			Cmd:  OpAssign,
			Args: []any{varName, valueOp},
		}}, nil

	case *ast.LetStatement:
		// Variable declaration: int VAR = VALUE
		varName := Variable(normalizeVarName(s.Name.Value))
		if s.Value != nil {
			valueOp, err := i.interpretExpression(s.Value)
			if err != nil {
				return nil, err
			}
			return []OpCode{{
				Cmd:  OpAssign,
				Args: []any{varName, valueOp},
			}}, nil
		}
		// Declaration without initialization
		return []OpCode{{
			Cmd:  OpAssign,
			Args: []any{varName, 0},
		}}, nil

	case *ast.IfStatement:
		// If statement: if (COND) { ... } else { ... }
		condOp, err := i.interpretExpression(s.Condition)
		if err != nil {
			return nil, err
		}

		thenOps := []OpCode{}
		for _, stmt := range s.Consequence.Statements {
			ops, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, err
			}
			thenOps = append(thenOps, ops...)
		}

		elseOps := []OpCode{}
		if s.Alternative != nil {
			for _, stmt := range s.Alternative.Statements {
				ops, err := i.interpretStatement(stmt)
				if err != nil {
					return nil, err
				}
				elseOps = append(elseOps, ops...)
			}
		}

		return []OpCode{{
			Cmd:  OpIf,
			Args: []any{condOp, thenOps, elseOps},
		}}, nil

	case *ast.ForStatement:
		// For loop: for (INIT; COND; POST) { ... }
		var initOps []OpCode
		if s.Init != nil {
			ops, err := i.interpretStatement(s.Init)
			if err != nil {
				return nil, err
			}
			initOps = ops
		}

		var condOp OpCode
		if s.Condition != nil {
			op, err := i.interpretExpression(s.Condition)
			if err != nil {
				return nil, err
			}
			condOp = op
		}

		var postOps []OpCode
		if s.Post != nil {
			ops, err := i.interpretStatement(s.Post)
			if err != nil {
				return nil, err
			}
			postOps = ops
		}

		bodyOps := []OpCode{}
		for _, stmt := range s.Body.Statements {
			ops, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, err
			}
			bodyOps = append(bodyOps, ops...)
		}

		return []OpCode{{
			Cmd:  OpFor,
			Args: []any{initOps, condOp, postOps, bodyOps},
		}}, nil

	case *ast.WhileStatement:
		// While loop: while (COND) { ... }
		condOp, err := i.interpretExpression(s.Condition)
		if err != nil {
			return nil, err
		}

		bodyOps := []OpCode{}
		for _, stmt := range s.Body.Statements {
			ops, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, err
			}
			bodyOps = append(bodyOps, ops...)
		}

		return []OpCode{{
			Cmd:  OpWhile,
			Args: []any{condOp, bodyOps},
		}}, nil

	case *ast.MesBlockStatement:
		// mes() block: mes(TIME) { ... }
		modeOp, err := i.interpretExpression(s.Time)
		if err != nil {
			return nil, err
		}

		bodyOps := []OpCode{}
		for _, stmt := range s.Body.Statements {
			ops, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, err
			}
			bodyOps = append(bodyOps, ops...)
		}

		return []OpCode{{
			Cmd:  OpRegisterSequence,
			Args: []any{modeOp, bodyOps},
		}}, nil

	case *ast.StepBlockStatement:
		// step(n) block: step(5) { ... }
		// This block executes its body once, but sets the step resolution
		// The body is executed immediately (not repeated)

		// Interpret the body statements
		bodyOps := []OpCode{}
		for _, stmt := range s.Body.Statements {
			ops, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, err
			}
			bodyOps = append(bodyOps, ops...)
		}

		// Return a single OpStep operation with count and body
		return []OpCode{{
			Cmd:  OpStep,
			Args: []any{int(s.Count), bodyOps},
		}}, nil

	case *ast.WaitStatement:
		// Wait statement: ,,,,
		return []OpCode{{
			Cmd:  OpWait,
			Args: []any{s.Count},
		}}, nil

	case *ast.BreakStatement:
		return []OpCode{{Cmd: OpBreak, Args: []any{}}}, nil

	case *ast.ContinueStatement:
		return []OpCode{{Cmd: OpContinue, Args: []any{}}}, nil

	case *ast.BlockStatement:
		// Block statement (used in control flow)
		ops := []OpCode{}
		for _, stmt := range s.Statements {
			stmtOps, err := i.interpretStatement(stmt)
			if err != nil {
				return nil, err
			}
			ops = append(ops, stmtOps...)
		}
		return ops, nil

	default:
		return nil, fmt.Errorf("unsupported statement type: %T", stmt)
	}
}

// interpretExpression converts an expression AST to OpCode
func (i *Interpreter) interpretExpression(expr ast.Expression) (OpCode, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		// Integer literal: return as-is
		return OpCode{Cmd: OpLiteral, Args: []any{int(e.Value)}}, nil

	case *ast.StringLiteral:
		// String literal: return as-is
		return OpCode{Cmd: OpLiteral, Args: []any{e.Value}}, nil

	case *ast.Identifier:
		// Check for special constants first
		varName := normalizeVarName(e.Value)
		switch varName {
		case "midi_time":
			// MIDI_TIME constant = 1
			return OpCode{Cmd: OpLiteral, Args: []any{1}}, nil
		case "time":
			// TIME constant = 0
			return OpCode{Cmd: OpLiteral, Args: []any{0}}, nil
		case "end_step":
			// END_STEP constant = -1
			return OpCode{Cmd: OpLiteral, Args: []any{-1}}, nil
		default:
			// Variable reference - return as Variable type directly
			// Return a special marker OpCode that will be unwrapped to Variable
			return OpCode{Cmd: OpVarRef, Args: []any{Variable(varName)}}, nil
		}

	case *ast.CallExpression:
		// Function call: FUNC(ARG1, ARG2, ...)
		funcName := normalizeVarName(e.Function.Value)

		// Special handling for Wait() - convert to OpWait
		if funcName == "wait" {
			if len(e.Arguments) >= 1 {
				argOp, err := i.interpretExpression(e.Arguments[0])
				if err != nil {
					return OpCode{}, err
				}
				return OpCode{Cmd: OpWait, Args: []any{argOp}}, nil
			}
			return OpCode{}, fmt.Errorf("Wait() requires 1 argument")
		}

		args := []any{funcName}

		for _, arg := range e.Arguments {
			argOp, err := i.interpretExpression(arg)
			if err != nil {
				return OpCode{}, err
			}
			args = append(args, argOp)
		}

		return OpCode{Cmd: OpCall, Args: args}, nil

	case *ast.InfixExpression:
		// Binary operation: LEFT OP RIGHT
		leftOp, err := i.interpretExpression(e.Left)
		if err != nil {
			return OpCode{}, err
		}

		rightOp, err := i.interpretExpression(e.Right)
		if err != nil {
			return OpCode{}, err
		}

		return OpCode{
			Cmd:  OpInfix,
			Args: []any{e.Operator, leftOp, rightOp},
		}, nil

	case *ast.PrefixExpression:
		// Unary operation: OP RIGHT
		rightOp, err := i.interpretExpression(e.Right)
		if err != nil {
			return OpCode{}, err
		}

		return OpCode{
			Cmd:  OpPrefix,
			Args: []any{e.Operator, rightOp},
		}, nil

	case *ast.IndexExpression:
		// Array access: VAR[IDX]
		leftOp, err := i.interpretExpression(e.Left)
		if err != nil {
			return OpCode{}, err
		}

		if e.Index != nil {
			indexOp, err := i.interpretExpression(e.Index)
			if err != nil {
				return OpCode{}, err
			}
			return OpCode{
				Cmd:  OpIndex,
				Args: []any{leftOp, indexOp},
			}, nil
		}

		// Array declaration: VAR[]
		return OpCode{
			Cmd:  OpArray,
			Args: []any{leftOp},
		}, nil

	default:
		return OpCode{}, fmt.Errorf("unsupported expression type: %T", expr)
	}
}

// evaluateConstant evaluates a constant expression at compile time
// This is used for parameter default values
func (i *Interpreter) evaluateConstant(expr ast.Expression) (any, error) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return int(e.Value), nil
	case *ast.StringLiteral:
		return e.Value, nil
	default:
		return nil, fmt.Errorf("unsupported constant expression: %T", expr)
	}
}

// scanAssets discovers asset references in the AST
func (i *Interpreter) scanAssets(program *ast.Program) []string {
	assets := []string{}
	seen := make(map[string]bool)

	var walker func(node ast.Node)
	walker = func(node ast.Node) {
		switch n := node.(type) {
		case *ast.Program:
			for _, s := range n.Statements {
				walker(s)
			}

		case *ast.FunctionStatement:
			walker(n.Body)

		case *ast.ExpressionStatement:
			if n.Expression != nil {
				walker(n.Expression)
			}

		case *ast.AssignStatement:
			if n.Value != nil {
				walker(n.Value)
			}
			if n.Index != nil {
				walker(n.Index)
			}

		case *ast.LetStatement:
			if n.Value != nil {
				walker(n.Value)
			}

		case *ast.MesBlockStatement:
			if n.Body != nil {
				walker(n.Body)
			}

		case *ast.StepBlockStatement:
			if n.Body != nil {
				walker(n.Body)
			}

		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}

		case *ast.IfStatement:
			if n.Condition != nil {
				walker(n.Condition)
			}
			if n.Consequence != nil {
				walker(n.Consequence)
			}
			if n.Alternative != nil {
				walker(n.Alternative)
			}

		case *ast.ForStatement:
			if n.Init != nil {
				walker(n.Init)
			}
			if n.Condition != nil {
				walker(n.Condition)
			}
			if n.Post != nil {
				walker(n.Post)
			}
			if n.Body != nil {
				walker(n.Body)
			}

		case *ast.WhileStatement:
			if n.Condition != nil {
				walker(n.Condition)
			}
			if n.Body != nil {
				walker(n.Body)
			}

		case *ast.DoWhileStatement:
			if n.Body != nil {
				walker(n.Body)
			}
			if n.Condition != nil {
				walker(n.Condition)
			}

		case *ast.SwitchStatement:
			if n.Value != nil {
				walker(n.Value)
			}
			for _, c := range n.Cases {
				if c.Value != nil {
					walker(c.Value)
				}
				if c.Body != nil {
					walker(c.Body)
				}
			}
			if n.Default != nil {
				walker(n.Default)
			}

		case *ast.CallExpression:
			// Check for asset loading functions
			funcName := normalizeVarName(n.Function.Value)
			if funcName == "loadpic" || funcName == "playmidi" || funcName == "playwave" {
				// Extract string literal filename from first argument
				if len(n.Arguments) > 0 {
					if strLit, ok := n.Arguments[0].(*ast.StringLiteral); ok {
						filename := strLit.Value
						if !seen[filename] {
							assets = append(assets, filename)
							seen[filename] = true
						}
					}
				}
			}

			// Recursively check arguments
			for _, arg := range n.Arguments {
				walker(arg)
			}

		case *ast.InfixExpression:
			if n.Left != nil {
				walker(n.Left)
			}
			if n.Right != nil {
				walker(n.Right)
			}

		case *ast.PrefixExpression:
			if n.Right != nil {
				walker(n.Right)
			}

		case *ast.IndexExpression:
			if n.Left != nil {
				walker(n.Left)
			}
			if n.Index != nil {
				walker(n.Index)
			}
		}
	}

	walker(program)
	return assets
}

// normalizeVarName converts variable names to lowercase for case-insensitive storage
func normalizeVarName(name string) string {
	return strings.ToLower(name)
}
