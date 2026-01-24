package codegen

import (
	"fmt"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/interpreter"
	"github.com/zurustar/son-et/pkg/compiler/token"
)

// Generator converts AST to OpCode sequences.
type Generator struct {
	errors []string
}

// New creates a new code generator.
func New() *Generator {
	return &Generator{
		errors: []string{},
	}
}

// Errors returns the generator errors.
func (g *Generator) Errors() []string {
	return g.errors
}

// Generate converts a program AST to OpCode sequences.
func (g *Generator) Generate(program *ast.Program) []interpreter.OpCode {
	var opcodes []interpreter.OpCode

	for _, stmt := range program.Statements {
		codes := g.generateStatement(stmt)
		opcodes = append(opcodes, codes...)
	}

	return opcodes
}

// generateStatement converts a statement to OpCodes.
func (g *Generator) generateStatement(stmt ast.Statement) []interpreter.OpCode {
	switch s := stmt.(type) {
	case *ast.VarDeclaration:
		return g.generateVarDeclaration(s)
	case *ast.AssignStatement:
		return g.generateAssignStatement(s)
	case *ast.ExpressionStatement:
		return g.generateExpressionStatement(s)
	case *ast.IfStatement:
		return g.generateIfStatement(s)
	case *ast.ForStatement:
		return g.generateForStatement(s)
	case *ast.WhileStatement:
		return g.generateWhileStatement(s)
	case *ast.SwitchStatement:
		return g.generateSwitchStatement(s)
	case *ast.BreakStatement:
		return []interpreter.OpCode{{Cmd: interpreter.OpBreak}}
	case *ast.ContinueStatement:
		return []interpreter.OpCode{{Cmd: interpreter.OpContinue}}
	case *ast.ReturnStatement:
		return g.generateReturnStatement(s)
	case *ast.FunctionStatement:
		return g.generateFunctionStatement(s)
	case *ast.MesStatement:
		return g.generateMesStatement(s)
	case *ast.StepStatement:
		return g.generateStepStatement(s)
	case *ast.BlockStatement:
		return g.generateBlockStatement(s)
	default:
		g.errors = append(g.errors, fmt.Sprintf("unknown statement type: %T", stmt))
		return nil
	}
}

// generateAssignStatement converts an assignment to OpCode.
func (g *Generator) generateAssignStatement(stmt *ast.AssignStatement) []interpreter.OpCode {
	// Check if it's array assignment
	if indexExpr, ok := stmt.Name.(*ast.IndexExpression); ok {
		// arr[index] = value
		arrayName := g.generateExpression(indexExpr.Left)
		index := g.generateExpression(indexExpr.Index)
		value := g.generateExpression(stmt.Value)

		return []interpreter.OpCode{{
			Cmd:  interpreter.OpArrayAssign,
			Args: []any{arrayName, index, value},
		}}
	}

	// Regular variable assignment
	ident, ok := stmt.Name.(*ast.Identifier)
	if !ok {
		g.errors = append(g.errors, fmt.Sprintf("invalid assignment target: %T", stmt.Name))
		return nil
	}

	value := g.generateExpression(stmt.Value)

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpAssign,
		Args: []any{interpreter.Variable(ident.Value), value},
	}}
}

// generateVarDeclaration converts a variable declaration to OpCode.
// For now, variable declarations are handled implicitly by the VM
// (variables are created on first assignment), so we just return empty opcodes.
func (g *Generator) generateVarDeclaration(stmt *ast.VarDeclaration) []interpreter.OpCode {
	// Variable declarations don't generate opcodes in the current VM design
	// Variables are created dynamically on first assignment
	// Arrays with explicit sizes could be initialized here in the future
	return []interpreter.OpCode{}
}

// generateExpressionStatement converts an expression statement to OpCode.
func (g *Generator) generateExpressionStatement(stmt *ast.ExpressionStatement) []interpreter.OpCode {
	// For expression statements, check if it's already a function call
	if callExpr, ok := stmt.Expression.(*ast.CallExpression); ok {
		// It's a function call - generate it directly as OpCall
		funcName := g.generateExpression(callExpr.Function)
		args := make([]any, len(callExpr.Arguments))
		for i, arg := range callExpr.Arguments {
			args[i] = g.generateExpression(arg)
		}
		return []interpreter.OpCode{{
			Cmd:  interpreter.OpCall,
			Args: append([]any{funcName}, args...),
		}}
	}

	// For other expressions, just evaluate them (shouldn't normally happen)
	expr := g.generateExpression(stmt.Expression)
	return []interpreter.OpCode{{
		Cmd:  interpreter.OpCall,
		Args: []any{expr},
	}}
}

// generateExpression converts an expression to OpCode.
func (g *Generator) generateExpression(expr ast.Expression) any {
	switch e := expr.(type) {
	case *ast.Identifier:
		return interpreter.Variable(e.Value)
	case *ast.IntegerLiteral:
		return e.Value
	case *ast.FloatLiteral:
		return e.Value
	case *ast.StringLiteral:
		return e.Value
	case *ast.ArrayLiteral:
		elements := make([]any, len(e.Elements))
		for i, el := range e.Elements {
			elements[i] = g.generateExpression(el)
		}
		return elements
	case *ast.IndexExpression:
		// arr[index]
		array := g.generateExpression(e.Left)
		index := g.generateExpression(e.Index)
		return interpreter.OpCode{
			Cmd:  interpreter.OpArrayAccess,
			Args: []any{array, index},
		}
	case *ast.PrefixExpression:
		right := g.generateExpression(e.Right)
		return interpreter.OpCode{
			Cmd:  interpreter.OpUnaryOp,
			Args: []any{e.Operator, right},
		}
	case *ast.InfixExpression:
		left := g.generateExpression(e.Left)
		right := g.generateExpression(e.Right)
		return interpreter.OpCode{
			Cmd:  interpreter.OpBinaryOp,
			Args: []any{e.Operator, left, right},
		}
	case *ast.CallExpression:
		// Function call
		funcName := g.generateExpression(e.Function)
		args := make([]any, len(e.Arguments))
		for i, arg := range e.Arguments {
			args[i] = g.generateExpression(arg)
		}
		return interpreter.OpCode{
			Cmd:  interpreter.OpCall,
			Args: append([]any{funcName}, args...),
		}
	default:
		g.errors = append(g.errors, fmt.Sprintf("unknown expression type: %T", expr))
		return nil
	}
}

// generateIfStatement converts an if statement to OpCode.
func (g *Generator) generateIfStatement(stmt *ast.IfStatement) []interpreter.OpCode {
	condition := g.generateExpression(stmt.Condition)
	consequence := g.generateBlockStatement(stmt.Consequence)

	var alternative []interpreter.OpCode
	if stmt.Alternative != nil {
		alternative = g.generateBlockStatement(stmt.Alternative)
	}

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpIf,
		Args: []any{condition, consequence, alternative},
	}}
}

// generateForStatement converts a for loop to OpCode.
func (g *Generator) generateForStatement(stmt *ast.ForStatement) []interpreter.OpCode {
	init := g.generateStatement(stmt.Init)
	condition := g.generateExpression(stmt.Condition)
	post := g.generateStatement(stmt.Post)
	body := g.generateBlockStatement(stmt.Body)

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpFor,
		Args: []any{init, condition, post, body},
	}}
}

// generateWhileStatement converts a while loop to OpCode.
func (g *Generator) generateWhileStatement(stmt *ast.WhileStatement) []interpreter.OpCode {
	condition := g.generateExpression(stmt.Condition)
	body := g.generateBlockStatement(stmt.Body)

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpWhile,
		Args: []any{condition, body},
	}}
}

// generateSwitchStatement converts a switch statement to OpCode.
func (g *Generator) generateSwitchStatement(stmt *ast.SwitchStatement) []interpreter.OpCode {
	value := g.generateExpression(stmt.Value)

	cases := make([]any, len(stmt.Cases))
	for i, c := range stmt.Cases {
		caseValue := g.generateExpression(c.Value)
		caseBody := g.generateBlockStatement(c.Body)
		cases[i] = []any{caseValue, caseBody}
	}

	var defaultCase []interpreter.OpCode
	if stmt.Default != nil {
		defaultCase = g.generateBlockStatement(stmt.Default)
	}

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpSwitch,
		Args: []any{value, cases, defaultCase},
	}}
}

// generateReturnStatement converts a return statement to OpCode.
func (g *Generator) generateReturnStatement(stmt *ast.ReturnStatement) []interpreter.OpCode {
	if stmt.ReturnValue != nil {
		value := g.generateExpression(stmt.ReturnValue)
		return []interpreter.OpCode{{
			Cmd:  interpreter.OpCall,
			Args: []any{"return", value},
		}}
	}
	return []interpreter.OpCode{{
		Cmd:  interpreter.OpCall,
		Args: []any{"return"},
	}}
}

// generateFunctionStatement converts a function definition to OpCode.
func (g *Generator) generateFunctionStatement(stmt *ast.FunctionStatement) []interpreter.OpCode {
	params := make([]any, len(stmt.Parameters))
	for i, p := range stmt.Parameters {
		params[i] = p.Value
	}

	body := g.generateBlockStatement(stmt.Body)

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpCall,
		Args: []any{"define_function", stmt.Name.Value, params, body},
	}}
}

// generateMesStatement converts a mes() block to OpCode.
func (g *Generator) generateMesStatement(stmt *ast.MesStatement) []interpreter.OpCode {
	eventType := g.eventTypeToString(stmt.EventType)

	// Use regular block generation for mes() blocks
	// Commas in mes() blocks are handled only within step() blocks
	body := g.generateBlockStatement(stmt.Body)

	return []interpreter.OpCode{{
		Cmd:  interpreter.OpRegisterEventHandler,
		Args: []any{eventType, body},
	}}
}

// generateStepStatement converts a step() call to OpCode.
// Simple form: step(n); → OpWait(n)
// Block form: step(n) { cmd1;, cmd2;,, cmd3; } → Flat sequence with waits
//
// Timing: step(n) = n × 50ms = n × 0.05 seconds
// The VM will convert this to ticks based on the current frame rate
func (g *Generator) generateStepStatement(stmt *ast.StepStatement) []interpreter.OpCode {
	count := g.generateExpression(stmt.Count)

	// Simple form: step(n);
	if stmt.Body == nil {
		// Pass step count directly to VM - it will convert to ticks
		// step(n) = n × 50ms
		return []interpreter.OpCode{{
			Cmd:  interpreter.OpWait,
			Args: []any{count},
		}}
	}

	// Block form: step(n) { ... }
	// Generate flat sequence with SetStep first, then commands with Wait(comma_count)
	// The comma syntax (;,) means "execute and wait 1 step"
	// Multiple commas (;,,) mean "wait multiple steps" (Wait(2))
	var opcodes []interpreter.OpCode

	// First, generate SetStep to set the step duration
	setStepOp := interpreter.OpCode{
		Cmd:  interpreter.OpSetStep,
		Args: []any{count},
	}
	opcodes = append(opcodes, setStepOp)

	i := 0
	for i < len(stmt.Body.Statements) {
		cmd := stmt.Body.Statements[i]

		// Check for empty statement (from commas) - count consecutive commas
		if exprStmt, ok := cmd.(*ast.ExpressionStatement); ok && exprStmt.Expression == nil {
			// Count consecutive empty statements (commas)
			waitCount := 0
			for i < len(stmt.Body.Statements) {
				if exprStmt, ok := stmt.Body.Statements[i].(*ast.ExpressionStatement); ok && exprStmt.Expression == nil {
					waitCount++
					i++
				} else {
					break
				}
			}

			// Generate Wait with the count
			opcodes = append(opcodes, interpreter.OpCode{
				Cmd:  interpreter.OpWait,
				Args: []any{int64(waitCount)},
			})
			continue
		}

		// Check for end_step - it's just a marker, no opcode needed
		if callExpr, ok := cmd.(*ast.ExpressionStatement); ok {
			if call, ok := callExpr.Expression.(*ast.CallExpression); ok {
				if ident, ok := call.Function.(*ast.Identifier); ok && ident.Value == "end_step" {
					// end_step just marks the end, no opcode needed
					break
				}
			}
		}

		// Generate command OpCode
		cmdOps := g.generateStatement(cmd)
		opcodes = append(opcodes, cmdOps...)

		i++
	}

	return opcodes
}

// multiplyExpression multiplies an expression by a constant.
// If the expression is a literal int, returns the product as a literal.
// Otherwise, returns a binary operation OpCode.
func (g *Generator) multiplyExpression(expr any, multiplier int) any {
	// If expr is already a literal int, multiply it directly
	if val, ok := expr.(int64); ok {
		return val * int64(multiplier)
	}

	// Otherwise, return the expression as-is and let runtime handle it
	// For now, we assume step() always has a literal argument
	// TODO: Support dynamic expressions in step()
	return expr
}

// generateBlockStatement converts a block statement to OpCodes.
func (g *Generator) generateBlockStatement(stmt *ast.BlockStatement) []interpreter.OpCode {
	var opcodes []interpreter.OpCode

	for _, s := range stmt.Statements {
		// Skip empty statements (from commas outside of step() blocks)
		if exprStmt, ok := s.(*ast.ExpressionStatement); ok && exprStmt.Expression == nil {
			continue
		}

		codes := g.generateStatement(s)
		opcodes = append(opcodes, codes...)
	}

	return opcodes
}

// eventTypeToString converts a token type to event type string.
func (g *Generator) eventTypeToString(tokenType token.TokenType) string {
	switch tokenType {
	case token.TIME:
		return "TIME"
	case token.MIDI_TIME:
		return "MIDI_TIME"
	case token.MIDI_END:
		return "MIDI_END"
	case token.KEY:
		return "KEY"
	case token.CLICK:
		return "CLICK"
	case token.RBDOWN:
		return "RBDOWN"
	case token.RBDBLCLK:
		return "RBDBLCLK"
	case token.USER:
		return "USER"
	default:
		return "UNKNOWN"
	}
}
