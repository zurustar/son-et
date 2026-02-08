// Package compiler provides OpCode generation for FILLY scripts (.TFY files).
// It transforms an AST into a sequence of OpCode instructions.
package compiler

import (
	"fmt"

	"github.com/zurustar/son-et/pkg/compiler/parser"
	"github.com/zurustar/son-et/pkg/opcode"
)

// CompilerError represents an error that occurred during compilation.
// It includes location information when available from AST nodes.
type CompilerError struct {
	Message string
	Line    int
	Column  int
}

// Error implements the error interface.
func (e *CompilerError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("compiler error at line %d, column %d: %s", e.Line, e.Column, e.Message)
	}
	return fmt.Sprintf("compiler error: %s", e.Message)
}

// NewCompilerError creates a new CompilerError with the given message and location.
func NewCompilerError(message string, line, column int) *CompilerError {
	return &CompilerError{
		Message: message,
		Line:    line,
		Column:  column,
	}
}

// Compiler generates OpCode from an AST.
type Compiler struct {
	errors []*CompilerError
}

// New creates a new Compiler.
func New() *Compiler {
	return &Compiler{
		errors: []*CompilerError{},
	}
}

// Compile compiles the given AST program into OpCode instructions.
// It iterates through all statements in the program and generates OpCode for each.
// Returns the generated OpCode sequence and any compilation errors.
// Requirement 5.5: Compiler reports unknown AST node types in error messages.
// Requirement 5.6: System collects all errors and returns them to caller.
func (c *Compiler) Compile(program *parser.Program) ([]opcode.OpCode, []error) {
	if program == nil {
		return nil, []error{NewCompilerError("program is nil", 0, 0)}
	}

	var opcodes []opcode.OpCode

	// Iterate through all statements in the program
	for _, stmt := range program.Statements {
		stmtOpcodes := c.compileStatement(stmt)
		opcodes = append(opcodes, stmtOpcodes...)
	}

	// Convert CompilerError to error interface
	var errs []error
	for _, e := range c.errors {
		errs = append(errs, e)
	}

	return opcodes, errs
}

// Errors returns the list of compilation errors.
func (c *Compiler) Errors() []*CompilerError {
	return c.errors
}

// addError adds an error message to the compiler's error list with location information.
// Requirement 5.5: Compiler reports unknown AST node types in error messages.
func (c *Compiler) addError(line, column int, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	c.errors = append(c.errors, NewCompilerError(msg, line, column))
}

// addErrorNoLocation adds an error message without location information.
func (c *Compiler) addErrorNoLocation(format string, args ...any) {
	c.addError(0, 0, format, args...)
}

// getStatementLocation extracts line and column from a statement's token.
func getStatementLocation(stmt parser.Statement) (int, int) {
	switch s := stmt.(type) {
	case *parser.VarDeclaration:
		return s.Token.Line, s.Token.Column
	case *parser.FunctionStatement:
		return s.Token.Line, s.Token.Column
	case *parser.BlockStatement:
		return s.Token.Line, s.Token.Column
	case *parser.AssignStatement:
		return s.Token.Line, s.Token.Column
	case *parser.ExpressionStatement:
		return s.Token.Line, s.Token.Column
	case *parser.IfStatement:
		return s.Token.Line, s.Token.Column
	case *parser.ForStatement:
		return s.Token.Line, s.Token.Column
	case *parser.WhileStatement:
		return s.Token.Line, s.Token.Column
	case *parser.SwitchStatement:
		return s.Token.Line, s.Token.Column
	case *parser.MesStatement:
		return s.Token.Line, s.Token.Column
	case *parser.StepStatement:
		return s.Token.Line, s.Token.Column
	case *parser.BreakStatement:
		return s.Token.Line, s.Token.Column
	case *parser.ContinueStatement:
		return s.Token.Line, s.Token.Column
	case *parser.ReturnStatement:
		return s.Token.Line, s.Token.Column
	default:
		return 0, 0
	}
}

// getExpressionLocation extracts line and column from an expression's token.
func getExpressionLocation(expr parser.Expression) (int, int) {
	switch e := expr.(type) {
	case *parser.Identifier:
		return e.Token.Line, e.Token.Column
	case *parser.IntegerLiteral:
		return e.Token.Line, e.Token.Column
	case *parser.FloatLiteral:
		return e.Token.Line, e.Token.Column
	case *parser.StringLiteral:
		return e.Token.Line, e.Token.Column
	case *parser.BinaryExpression:
		return e.Token.Line, e.Token.Column
	case *parser.UnaryExpression:
		return e.Token.Line, e.Token.Column
	case *parser.CallExpression:
		return e.Token.Line, e.Token.Column
	case *parser.IndexExpression:
		return e.Token.Line, e.Token.Column
	case *parser.ArrayReference:
		return e.Token.Line, e.Token.Column
	default:
		return 0, 0
	}
}

// compileStatement compiles a single statement into OpCode instructions.
// It dispatches to the appropriate compile method based on the statement type.
// Requirement 5.5: Compiler reports unknown AST node types in error messages.
func (c *Compiler) compileStatement(stmt parser.Statement) []opcode.OpCode {
	switch s := stmt.(type) {
	case *parser.VarDeclaration:
		return c.compileVarDeclaration(s)
	case *parser.FunctionStatement:
		return c.compileFunctionStatement(s)
	case *parser.BlockStatement:
		return c.compileBlockStatement(s)
	case *parser.AssignStatement:
		return c.compileAssignStatement(s)
	case *parser.ExpressionStatement:
		return c.compileExpressionStatement(s)
	case *parser.IfStatement:
		return c.compileIfStatement(s)
	case *parser.ForStatement:
		return c.compileForStatement(s)
	case *parser.WhileStatement:
		return c.compileWhileStatement(s)
	case *parser.SwitchStatement:
		return c.compileSwitchStatement(s)
	case *parser.MesStatement:
		return c.compileMesStatement(s)
	case *parser.StepStatement:
		return c.compileStepStatement(s)
	case *parser.BreakStatement:
		return c.compileBreakStatement(s)
	case *parser.ContinueStatement:
		return c.compileContinueStatement(s)
	case *parser.ReturnStatement:
		return c.compileReturnStatement(s)
	case *parser.InfoDirective:
		// #info directives are metadata and don't generate OpCode
		return []opcode.OpCode{}
	case *parser.IncludeDirective:
		// #include directives are handled during preprocessing, not compilation
		return []opcode.OpCode{}
	case *parser.DefineDirective:
		// #define directives are handled during preprocessing, not compilation
		return []opcode.OpCode{}
	case *parser.LabelStatement:
		// Labels are used for goto, which is not yet implemented
		// For now, just skip them
		return []opcode.OpCode{}
	default:
		line, col := getStatementLocation(stmt)
		c.addError(line, col, "unknown statement type: %T", stmt)
		return []opcode.OpCode{}
	}
}

// compileExpression compiles an expression and returns its value representation.
// The returned value can be:
// - Primitive values (int64, float64, string)
// - Variable references (Variable type)
// - OpCode for complex expressions (BinaryOp, UnaryOp, etc.)
// Requirement 5.5: Compiler reports unknown AST node types in error messages.
func (c *Compiler) compileExpression(expr parser.Expression) any {
	if expr == nil {
		return nil
	}

	switch e := expr.(type) {
	case *parser.Identifier:
		return c.compileIdentifier(e)
	case *parser.IntegerLiteral:
		return c.compileIntegerLiteral(e)
	case *parser.FloatLiteral:
		return c.compileFloatLiteral(e)
	case *parser.StringLiteral:
		return c.compileStringLiteral(e)
	case *parser.BinaryExpression:
		return c.compileBinaryExpression(e)
	case *parser.UnaryExpression:
		return c.compileUnaryExpression(e)
	case *parser.CallExpression:
		return c.compileCallExpression(e)
	case *parser.IndexExpression:
		return c.compileIndexExpression(e)
	case *parser.ArrayReference:
		return c.compileArrayReference(e)
	default:
		line, col := getExpressionLocation(expr)
		c.addError(line, col, "unknown expression type: %T", expr)
		return nil
	}
}

// ============================================================================
// Statement Compilation Methods
// ============================================================================

// compileVarDeclaration compiles a variable declaration statement.
// For global variables (outside functions), this generates initialization OpCodes.
// For local variables, they are created dynamically when first assigned.
func (c *Compiler) compileVarDeclaration(vd *parser.VarDeclaration) []opcode.OpCode {
	var opcodes []opcode.OpCode

	// Generate initialization OpCodes for each declared variable
	for i, name := range vd.Names {
		var defaultValue any
		if vd.Type == "str" || vd.Type == "string" {
			defaultValue = ""
		} else {
			defaultValue = int64(0)
		}

		// For arrays, initialize with an empty array
		if vd.IsArray[i] {
			// Create an empty array
			// If size is specified, we could pre-allocate, but FILLY arrays grow dynamically
			defaultValue = []any{}
		}

		// Generate assignment OpCode to initialize the variable
		opcodes = append(opcodes, opcode.OpCode{
			Cmd:  opcode.Assign,
			Args: []any{opcode.Variable(name), defaultValue},
		})
	}

	return opcodes
}

// compileFunctionStatement compiles a function definition.
// It generates an OpDefineFunction with the function name, parameters, and compiled body.
func (c *Compiler) compileFunctionStatement(fs *parser.FunctionStatement) []opcode.OpCode {
	// Compile the function body
	var bodyOpcodes []opcode.OpCode
	if fs.Body != nil {
		bodyOpcodes = c.compileBlockStatement(fs.Body)
	}

	// Build parameter list with names, types, and default values
	params := make([]any, 0, len(fs.Parameters))
	for _, param := range fs.Parameters {
		paramInfo := map[string]any{
			"name":    param.Name,
			"type":    param.Type,
			"isArray": param.IsArray,
		}
		if param.DefaultValue != nil {
			paramInfo["default"] = c.compileExpression(param.DefaultValue)
		}
		params = append(params, paramInfo)
	}

	// Generate OpDefineFunction with function name, parameters, and body
	return []opcode.OpCode{
		{
			Cmd: opcode.DefineFunction,
			Args: []any{
				fs.Name,
				params,
				bodyOpcodes,
			},
		},
	}
}

// compileBlockStatement compiles a block of statements.
func (c *Compiler) compileBlockStatement(bs *parser.BlockStatement) []opcode.OpCode {
	var opcodes []opcode.OpCode
	for _, stmt := range bs.Statements {
		stmtOpcodes := c.compileStatement(stmt)
		opcodes = append(opcodes, stmtOpcodes...)
	}
	return opcodes
}

// compileAssignStatement compiles an assignment statement.
// Handles both simple variable assignment (x = value) and array element assignment (arr[i] = value).
// For simple assignment: generates OpAssign with opcode.Variable(name) and compiled value.
// For array assignment: generates OpArrayAssign with array name, index, and value.
func (c *Compiler) compileAssignStatement(as *parser.AssignStatement) []opcode.OpCode {
	value := c.compileExpression(as.Value)

	switch target := as.Name.(type) {
	case *parser.Identifier:
		// Simple variable assignment: x = value
		// opcode.OpCode{Cmd: opcode.Assign, Args: []any{opcode.Variable("x"), value}}
		return []opcode.OpCode{
			{
				Cmd:  opcode.Assign,
				Args: []any{opcode.Variable(target.Value), value},
			},
		}

	case *parser.IndexExpression:
		// Array element assignment: arr[i] = value
		// opcode.OpCode{Cmd: opcode.ArrayAssign, Args: []any{opcode.Variable("arr"), index, value}}
		// The Left of IndexExpression should be an Identifier for the array name
		var arrayName opcode.Variable
		switch left := target.Left.(type) {
		case *parser.Identifier:
			arrayName = opcode.Variable(left.Value)
		default:
			// For nested array access or other complex expressions,
			// compile the left side as an expression
			arrayName = opcode.Variable("")
			line, col := getExpressionLocation(target.Left)
			c.addError(line, col, "array assignment target must be an identifier, got %T", target.Left)
			return []opcode.OpCode{}
		}

		index := c.compileExpression(target.Index)
		return []opcode.OpCode{
			{
				Cmd:  opcode.ArrayAssign,
				Args: []any{arrayName, index, value},
			},
		}

	default:
		line, col := as.Token.Line, as.Token.Column
		c.addError(line, col, "invalid assignment target: %T", as.Name)
		return []opcode.OpCode{}
	}
}

// compileExpressionStatement compiles an expression statement.
// If the expression is a CallExpression, it generates an OpCall.
// Otherwise, it compiles the expression (which may have side effects).
func (c *Compiler) compileExpressionStatement(es *parser.ExpressionStatement) []opcode.OpCode {
	if es.Expression == nil {
		return []opcode.OpCode{}
	}

	// Check if the expression is a function call
	if ce, ok := es.Expression.(*parser.CallExpression); ok {
		// Generate OpCall for function calls
		args := []any{ce.Function}
		for _, arg := range ce.Arguments {
			args = append(args, c.compileExpression(arg))
		}
		return []opcode.OpCode{
			{Cmd: opcode.Call, Args: args},
		}
	}

	// For other expressions, compile them (they may have side effects)
	// The result is discarded but the expression is still evaluated
	result := c.compileExpression(es.Expression)
	if result == nil {
		return []opcode.OpCode{}
	}

	// If the result is an OpCode (e.g., from a complex expression), return it
	if op, ok := result.(opcode.OpCode); ok {
		return []opcode.OpCode{op}
	}

	// For simple values (literals, variables), no OpCode is generated
	// as they have no side effects
	return []opcode.OpCode{}
}

// compileIfStatement compiles an if statement.
// Generates OpIf with condition, then block, and optional else block.
// For if-else if chains, the else block contains another OpIf.
//
// Example: if (x > 5) { y = 10 } else { y = 0 }
// opcode.OpCode{
//
//	Cmd: opcode.If,
//	Args: []any{
//	    // condition
//	    opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{">", opcode.Variable("x"), 5}},
//	    // then block
//	    []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), 10}}},
//	    // else block
//	    []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("y"), 0}}},
//	},
//
// }
func (c *Compiler) compileIfStatement(is *parser.IfStatement) []opcode.OpCode {
	// Compile the condition
	condition := c.compileExpression(is.Condition)

	// Compile the then block (consequence)
	thenBlock := []opcode.OpCode{}
	if is.Consequence != nil {
		thenBlock = c.compileBlockStatement(is.Consequence)
	}

	// Compile the else block (alternative)
	elseBlock := []opcode.OpCode{}
	if is.Alternative != nil {
		switch alt := is.Alternative.(type) {
		case *parser.BlockStatement:
			// Simple else block
			elseBlock = c.compileBlockStatement(alt)
		case *parser.IfStatement:
			// else if chain - compile as nested if statement
			elseBlock = c.compileIfStatement(alt)
		default:
			line, col := getStatementLocation(is.Alternative)
			c.addError(line, col, "unknown alternative type in if statement: %T", is.Alternative)
		}
	}

	return []opcode.OpCode{
		{
			Cmd: opcode.If,
			Args: []any{
				condition,
				thenBlock,
				elseBlock,
			},
		},
	}
}

// compileForStatement compiles a for loop.
// Generates OpFor with init, condition, post, and body blocks.
//
// Example: for(i=0; i<10; i=i+1) { ... }
// opcode.OpCode{
//
//	Cmd: opcode.For,
//	Args: []any{
//	    // init
//	    []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), 0}}},
//	    // condition
//	    opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"<", opcode.Variable("i"), 10}},
//	    // post
//	    []opcode.OpCode{{Cmd: opcode.Assign, Args: []any{opcode.Variable("i"), opcode.OpCode{Cmd: opcode.BinaryOp, Args: []any{"+", opcode.Variable("i"), 1}}}}},
//	    // body
//	    []opcode.OpCode{...},
//	},
//
// }
func (c *Compiler) compileForStatement(fs *parser.ForStatement) []opcode.OpCode {
	// Compile the init statement
	initBlock := []opcode.OpCode{}
	if fs.Init != nil {
		initBlock = c.compileStatement(fs.Init)
	}

	// Compile the condition
	var condition any
	if fs.Condition != nil {
		condition = c.compileExpression(fs.Condition)
	}

	// Compile the post statement
	postBlock := []opcode.OpCode{}
	if fs.Post != nil {
		postBlock = c.compileStatement(fs.Post)
	}

	// Compile the body
	bodyBlock := []opcode.OpCode{}
	if fs.Body != nil {
		bodyBlock = c.compileBlockStatement(fs.Body)
	}

	return []opcode.OpCode{
		{
			Cmd: opcode.For,
			Args: []any{
				initBlock,
				condition,
				postBlock,
				bodyBlock,
			},
		},
	}
}

// compileWhileStatement compiles a while loop.
// Generates OpWhile with condition and body blocks.
//
// Example: while (condition) { body }
// opcode.OpCode{
//
//	Cmd: opcode.While,
//	Args: []any{
//	    // condition
//	    opcode.OpCode{...},
//	    // body
//	    []opcode.OpCode{...},
//	},
//
// }
func (c *Compiler) compileWhileStatement(ws *parser.WhileStatement) []opcode.OpCode {
	// Compile the condition
	var condition any
	if ws.Condition != nil {
		condition = c.compileExpression(ws.Condition)
	}

	// Compile the body
	bodyBlock := []opcode.OpCode{}
	if ws.Body != nil {
		bodyBlock = c.compileBlockStatement(ws.Body)
	}

	return []opcode.OpCode{
		{
			Cmd: opcode.While,
			Args: []any{
				condition,
				bodyBlock,
			},
		},
	}
}

// compileSwitchStatement compiles a switch statement.
// Generates OpSwitch with value, cases array, and optional default block.
//
// Example: switch (value) { case 1: ... default: ... }
// opcode.OpCode{
//
//	Cmd: opcode.Switch,
//	Args: []any{
//	    // value
//	    opcode.Variable("x"),
//	    // cases (array of case clauses)
//	    []any{
//	        map[string]any{"value": 1, "body": []opcode.OpCode{...}},
//	        map[string]any{"value": 2, "body": []opcode.OpCode{...}},
//	    },
//	    // default block
//	    []opcode.OpCode{...},
//	},
//
// }
func (c *Compiler) compileSwitchStatement(ss *parser.SwitchStatement) []opcode.OpCode {
	// Compile the switch value
	value := c.compileExpression(ss.Value)

	// Compile the case clauses
	cases := make([]any, 0, len(ss.Cases))
	for _, caseClause := range ss.Cases {
		// Compile the case value
		caseValue := c.compileExpression(caseClause.Value)

		// Compile the case body
		caseBody := []opcode.OpCode{}
		for _, stmt := range caseClause.Body {
			stmtOpcodes := c.compileStatement(stmt)
			caseBody = append(caseBody, stmtOpcodes...)
		}

		cases = append(cases, map[string]any{
			"value": caseValue,
			"body":  caseBody,
		})
	}

	// Compile the default block
	defaultBlock := []opcode.OpCode{}
	if ss.Default != nil {
		defaultBlock = c.compileBlockStatement(ss.Default)
	}

	return []opcode.OpCode{
		{
			Cmd: opcode.Switch,
			Args: []any{
				value,
				cases,
				defaultBlock,
			},
		},
	}
}

// compileMesStatement compiles a mes (event handler) statement.
// Generates OpRegisterEventHandler with event type and compiled body.
//
// Example: mes(MIDI_TIME) { step { ... } }
// opcode.OpCode{
//
//	Cmd: opcode.RegisterEventHandler,
//	Args: []any{
//	    "MIDI_TIME",
//	    []opcode.OpCode{...}, // body OpCode
//	},
//
// }
func (c *Compiler) compileMesStatement(ms *parser.MesStatement) []opcode.OpCode {
	// Compile the body block
	bodyOpcodes := []opcode.OpCode{}
	if ms.Body != nil {
		bodyOpcodes = c.compileBlockStatement(ms.Body)
	}

	// Generate OpRegisterEventHandler with event type and body
	return []opcode.OpCode{
		{
			Cmd: opcode.RegisterEventHandler,
			Args: []any{
				ms.EventType,
				bodyOpcodes,
			},
		},
	}
}

// compileStepStatement compiles a step statement.
// Step statements have special syntax where commas represent wait commands.
//
// Example: step(10) { func1();, func2();,, end_step; del_me; }
// Generates:
//
//	[]opcode.OpCode{
//	    {Cmd: opcode.SetStep, Args: []any{10}},
//	    {Cmd: opcode.Call, Args: []any{"func1"}},
//	    {Cmd: opcode.Wait, Args: []any{1}},  // 1 comma
//	    {Cmd: opcode.Call, Args: []any{"func2"}},
//	    {Cmd: opcode.Wait, Args: []any{2}},  // 2 commas
//	    {Cmd: opcode.Call, Args: []any{"del_me"}},
//	}
func (c *Compiler) compileStepStatement(ss *parser.StepStatement) []opcode.OpCode {
	var opcodes []opcode.OpCode

	// If Count is specified, generate OpSetStep first
	if ss.Count != nil {
		countValue := c.compileExpression(ss.Count)
		opcodes = append(opcodes, opcode.OpCode{
			Cmd:  opcode.SetStep,
			Args: []any{countValue},
		})
	}

	// If there's no body, return just the OpSetStep (if any)
	if ss.Body == nil {
		return opcodes
	}

	// Process each StepCommand in the body
	for _, cmd := range ss.Body.Commands {
		// If Statement is not nil, compile it
		if cmd.Statement != nil {
			stmtOpcodes := c.compileStatement(cmd.Statement)
			opcodes = append(opcodes, stmtOpcodes...)
		}

		// If WaitCount > 0, generate OpWait with the count
		if cmd.WaitCount > 0 {
			opcodes = append(opcodes, opcode.OpCode{
				Cmd:  opcode.Wait,
				Args: []any{cmd.WaitCount},
			})
		}
	}

	return opcodes
}

// compileBreakStatement compiles a break statement.
func (c *Compiler) compileBreakStatement(bs *parser.BreakStatement) []opcode.OpCode {
	return []opcode.OpCode{
		{Cmd: opcode.Break, Args: []any{}},
	}
}

// compileContinueStatement compiles a continue statement.
func (c *Compiler) compileContinueStatement(cs *parser.ContinueStatement) []opcode.OpCode {
	return []opcode.OpCode{
		{Cmd: opcode.Continue, Args: []any{}},
	}
}

// compileReturnStatement compiles a return statement.
func (c *Compiler) compileReturnStatement(rs *parser.ReturnStatement) []opcode.OpCode {
	args := []any{"return"}
	if rs.ReturnValue != nil {
		args = append(args, c.compileExpression(rs.ReturnValue))
	}
	return []opcode.OpCode{
		{Cmd: opcode.Call, Args: args},
	}
}

// ============================================================================
// Expression Compilation Methods
// ============================================================================

// compileIdentifier compiles an identifier expression.
// Returns a Variable reference for the identifier.
func (c *Compiler) compileIdentifier(id *parser.Identifier) any {
	return opcode.Variable(id.Value)
}

// compileIntegerLiteral compiles an integer literal expression.
// Returns the integer value directly.
func (c *Compiler) compileIntegerLiteral(il *parser.IntegerLiteral) any {
	return il.Value
}

// compileFloatLiteral compiles a float literal expression.
// Returns the float value directly.
func (c *Compiler) compileFloatLiteral(fl *parser.FloatLiteral) any {
	return fl.Value
}

// compileStringLiteral compiles a string literal expression.
// Returns the string value directly.
func (c *Compiler) compileStringLiteral(sl *parser.StringLiteral) any {
	return sl.Value
}

// compileBinaryExpression compiles a binary expression.
// Returns an OpCode with OpBinaryOp command.
func (c *Compiler) compileBinaryExpression(be *parser.BinaryExpression) any {
	left := c.compileExpression(be.Left)
	right := c.compileExpression(be.Right)
	return opcode.OpCode{
		Cmd:  opcode.BinaryOp,
		Args: []any{be.Operator, left, right},
	}
}

// compileUnaryExpression compiles a unary expression.
// Returns an OpCode with OpUnaryOp command.
func (c *Compiler) compileUnaryExpression(ue *parser.UnaryExpression) any {
	operand := c.compileExpression(ue.Right)
	return opcode.OpCode{
		Cmd:  opcode.UnaryOp,
		Args: []any{ue.Operator, operand},
	}
}

// compileCallExpression compiles a function call expression.
// Returns an OpCode with OpCall command.
func (c *Compiler) compileCallExpression(ce *parser.CallExpression) any {
	args := []any{ce.Function}
	for _, arg := range ce.Arguments {
		args = append(args, c.compileExpression(arg))
	}
	return opcode.OpCode{
		Cmd:  opcode.Call,
		Args: args,
	}
}

// compileIndexExpression compiles an array index expression.
// Returns an OpCode with OpArrayAccess command.
func (c *Compiler) compileIndexExpression(ie *parser.IndexExpression) any {
	array := c.compileExpression(ie.Left)
	index := c.compileExpression(ie.Index)
	return opcode.OpCode{
		Cmd:  opcode.ArrayAccess,
		Args: []any{array, index},
	}
}

// compileArrayReference compiles an array reference expression (arr[]).
// This is used when passing an entire array as a function argument.
// Returns a Variable reference for the array.
func (c *Compiler) compileArrayReference(ar *parser.ArrayReference) any {
	return opcode.Variable(ar.Name)
}
