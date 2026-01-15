package codegen

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/ast"
	"github.com/zurustar/son-et/pkg/compiler/token"
)

type Generator struct {
	buf           bytes.Buffer
	userFuncs     map[string]bool
	globals       map[string]bool
	currentParams map[string]bool
	signatures    map[string]*ast.FunctionStatement
	assets        []string
	arrayUsage    map[string]bool
	vmVars        map[string]bool // Variables that need VM registration (used in mes() blocks)
}

func New(assets []string) *Generator {
	return &Generator{assets: assets}
}

func (g *Generator) Generate(program *ast.Program) string {
	// 0. Pre-scan for user defined functions and globals
	userFuncs := map[string]bool{}
	globalsMap := map[string]bool{}
	signatures := map[string]*ast.FunctionStatement{}

	for _, stmt := range program.Statements {
		if fn, ok := stmt.(*ast.FunctionStatement); ok {
			userFuncs[fn.Name.Value] = true
			signatures[fn.Name.Value] = fn
		}
		if let, ok := stmt.(*ast.LetStatement); ok {
			name := strings.ToLower(let.Name.Value)
			if strings.HasSuffix(name, "[]") {
				name = strings.TrimSuffix(name, "[]")
			}
			globalsMap[name] = true
		}
	}
	g.userFuncs = userFuncs
	g.globals = globalsMap
	g.signatures = signatures

	// 1. Scan for array usage (identifiers used with index)
	arrayUsage := g.scanArrayUsage(program)
	g.arrayUsage = arrayUsage

	// Basic preamble for the generated Go file
	g.buf.WriteString("package main\n\n")
	g.buf.WriteString("import (\n")
	g.buf.WriteString("\t\"embed\"\n")
	g.buf.WriteString("\t\"github.com/zurustar/son-et/pkg/engine\"\n")
	g.buf.WriteString(")\n\n")

	// Emit go:embed directives from discovered assets
	if len(g.assets) > 0 {
		g.buf.WriteString("//go:embed ")
		g.buf.WriteString(strings.Join(g.assets, " "))
		g.buf.WriteString("\n")
		g.buf.WriteString("var assets embed.FS\n\n")
	} else {
		g.buf.WriteString("// No assets found to embed\n")
		g.buf.WriteString("var assets embed.FS\n\n")
	}

	// Separate statements into Globals, Functions, and Main
	var globals []*ast.LetStatement
	var defines []*ast.DefineStatement
	var functions []*ast.FunctionStatement
	var mainFunc *ast.FunctionStatement

	for _, stmt := range program.Statements {
		switch s := stmt.(type) {
		case *ast.LetStatement:
			globals = append(globals, s)
		case *ast.DefineStatement:
			defines = append(defines, s)
		case *ast.FunctionStatement:
			if s.Name.Value == "main" {
				mainFunc = s
			} else {
				functions = append(functions, s)
			}
		}
	}

	// Emit Defines (Constants)
	g.buf.WriteString("// Constants\n")
	g.buf.WriteString("const (\n")
	// Implicit FILLY constants
	g.buf.WriteString("\ttime = 0\n")
	g.buf.WriteString("\tuser = 1\n")
	g.buf.WriteString("\tmidi_end = 0\n")
	// c1 is used as array, so not constant. Added to global vars below logic or map?

	for _, d := range defines {
		g.buf.WriteString(strings.ToLower(d.Name.Value)) // Defines in FILLY are usually UPPERCASE (TIME), Go const/var?
		// If we use constants, we must ensure case matching.
		// Codegen uses ToLower for identifiers?
		// Step 1130 genExpression:
		// if e.Value == "MIDI_TIME" { ... } else { strings.ToLower(e.Value) }
		// So we should Lower here too.
		g.buf.WriteString(" = ")
		g.genExpression(d.Value)
		g.buf.WriteString("\n")
	}
	g.buf.WriteString(")\n\n")

	// Emit Globals
	g.buf.WriteString("// Global Variables\n")
	for _, v := range globals {
		// var Name int  OR  var Name []int
		name := strings.ToLower(v.Name.Value)
		typeName := "int"
		if strings.HasSuffix(name, "[]") || arrayUsage[name] {
			name = strings.TrimSuffix(name, "[]")
			typeName = "[]int"
		}
		if v.Token.Type == token.STR {
			typeName = "string"
		}
		// Go variables are case sensitive, FILLY might be case insensitive?
		// Keeping original casing for now but exporting (Capitalized) might be safer for Engine text/template?
		// Whatever, keeping as is.
		if typeName == "[]int" {
			g.buf.WriteString(fmt.Sprintf("var %s = make([]int, 1000)\n", name))
		} else {
			g.buf.WriteString(fmt.Sprintf("var %s %s\n", name, typeName))
		}
	}
	// Inject missing globals
	g.buf.WriteString("var c1 []int\n")
	g.buf.WriteString("var c2 []int\n")
	g.buf.WriteString("var p2 []int\n")
	g.buf.WriteString("\n")

	// Emit Helper Functions
	for _, f := range functions {
		g.buf.WriteString(fmt.Sprintf("func %s(", f.Name.Value))

		// Reset and populate tracked params
		g.currentParams = map[string]bool{}

		params := []string{}
		for _, p := range f.Parameters {
			pName := p.Name.Value
			pType := p.Type
			if pType == "" {
				pType = "int"
			}
			if pType == "str" {
				pType = "string"
			} // normalize

			// Store param name in scope (lowercased)
			pNameLower := strings.ToLower(pName)
			if strings.HasSuffix(pNameLower, "[]") {
				pNameLower = strings.TrimSuffix(pNameLower, "[]")
			}
			g.currentParams[pNameLower] = true

			// Logic for param type generation (Go side name)
			if strings.HasSuffix(pName, "[]") {
				pName = strings.TrimSuffix(pName, "[]")
				// If Captured type doesn't have [], append it?
				// Parser sets Type="[]int" if brackets present.
				// If source: int x[], parser sets Name="x", Type="[]int".
				// If source: int x, parser sets Name="x", Type="int".
				// My parser logic (Step 2597):
				//   name += "[]"  <- Parser MODIFIES name.
				//   paramType = "[]int"
				// So p.Name.Value ALREADY contains [].
			}
			// CodeGen loop at lines 136 checks Suffix.
			// If Name has [], strip it for Go variable name.
			if strings.HasSuffix(pName, "[]") {
				pName = strings.TrimSuffix(pName, "[]")
			}

			params = append(params, fmt.Sprintf("%s %s", strings.ToLower(pName), pType))
		}
		g.buf.WriteString(strings.Join(params, ", "))

		g.buf.WriteString(") {\n")

		// Scan for locals and pre-declare
		locals := g.scanLocals(f.Body)
		keys := make([]string, 0, len(locals))
		for k := range locals {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Deterministic order

		for _, name := range keys {
			// Check if parameter or global
			if g.currentParams[name] || g.globals[name] {
				continue
			}
			if locals[name] == "[]int" {
				g.buf.WriteString(fmt.Sprintf("\tvar %s = make([]int, 1000)\n", name))
			} else {
				g.buf.WriteString(fmt.Sprintf("\tvar %s %s\n", name, locals[name]))
			}
			// Suppress unused variable error
			g.buf.WriteString(fmt.Sprintf("\t_ = %s\n", name))
		}

		g.genBlock(f.Body)
		g.buf.WriteString("}\n\n")
		g.currentParams = nil // Clear after func
	}

	// Emit Main Entry Point
	g.buf.WriteString("func main() {\n")
	// Register user functions
	for name := range g.userFuncs {
		// Use original name (case sensitive?)
		// In pre-scan we stored raw names.
		// Go function names are CaseSensitive. userFuncs keys are raw.
		g.buf.WriteString(fmt.Sprintf("\tengine.RegisterUserFunc(%q, %s)\n", name, name))
	}

	g.buf.WriteString("\tengine.Init(assets, func() {\n")

	if mainFunc != nil {
		// Scan for variables used in mes() blocks - these need VM registration
		g.vmVars = g.scanMesBlocksForVMVars(mainFunc.Body)

		// Scan for locals in main - but declare them as globals
		// This is because mes() blocks need to access main's variables
		locals := g.scanLocals(mainFunc.Body)
		keys := make([]string, 0, len(locals))
		for k := range locals {
			keys = append(keys, k)
		}
		sort.Strings(keys) // Deterministic order

		// Declare main's locals as globals (before the script function)
		g.buf.WriteString("// Main function variables (declared as globals for mes() block access)\n")
		for _, name := range keys {
			// Check if already declared as global
			if g.globals[name] {
				continue
			}
			if locals[name] == "[]int" {
				g.buf.WriteString(fmt.Sprintf("var %s = make([]int, 1000)\n", name))
			} else {
				g.buf.WriteString(fmt.Sprintf("var %s %s\n", name, locals[name]))
			}
			// Mark as unused to avoid compiler errors
			g.buf.WriteString(fmt.Sprintf("var _ = %s\n", name))
		}
		g.buf.WriteString("\n")

		g.genBlock(mainFunc.Body)
	}

	g.buf.WriteString("\t})\n")
	g.buf.WriteString("\tdefer engine.Close()\n")
	g.buf.WriteString("\tengine.Run()\n")
	g.buf.WriteString("}\n")

	return g.buf.String()
}

func (g *Generator) scanResources(program *ast.Program) []string {
	resources := []string{}
	seen := map[string]bool{}

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
			walker(n.Expression)
		case *ast.AssignStatement:
			walker(n.Value)
		case *ast.MesBlockStatement:
			walker(n.Body)
			walker(n.Time)
		case *ast.StepBlockStatement:
			walker(n.Body)
		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.CallExpression:
			funcName := n.Function.Value
			if funcName == "LoadPic" || funcName == "loadpic" || funcName == "PlayMIDI" || funcName == "playmidi" {
				if len(n.Arguments) > 0 {
					if strLit, ok := n.Arguments[0].(*ast.StringLiteral); ok {
						if !seen[strLit.Value] {
							resources = append(resources, strLit.Value)
							seen[strLit.Value] = true
						}
					}
				}
			}
			for _, arg := range n.Arguments {
				walker(arg)
			}
		}
	}

	walker(program)
	return resources
}

func (g *Generator) scanArrayUsage(program *ast.Program) map[string]bool {
	usage := map[string]bool{}
	var walker func(node ast.Node)
	walker = func(node ast.Node) {
		switch n := node.(type) {
		case *ast.Program:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.FunctionStatement:
			walker(n.Body)
		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.AssignStatement:
			if n.Index != nil {
				name := strings.ToLower(n.Name.Value)
				usage[name] = true
				walker(n.Index)
			}
			walker(n.Value)
		case *ast.IndexExpression:
			if ident, ok := n.Left.(*ast.Identifier); ok {
				name := strings.ToLower(ident.Value)
				usage[name] = true
			}
			walker(n.Left)
			walker(n.Index)
		case *ast.CallExpression:
			for _, arg := range n.Arguments {
				walker(arg)
			}
		case *ast.IfStatement:
			walker(n.Condition)
			walker(n.Consequence)
			if n.Alternative != nil {
				walker(n.Alternative)
			}
		case *ast.ForStatement:
			if n.Init != nil {
				walker(n.Init)
			} // Init might be AssignStatement
			if n.Condition != nil {
				walker(n.Condition)
			}
			if n.Post != nil {
				walker(n.Post)
			}
			walker(n.Body)
		case *ast.WhileStatement:
			walker(n.Condition)
			walker(n.Body)
		case *ast.DoWhileStatement:
			walker(n.Body)
			walker(n.Condition)
		case *ast.SwitchStatement:
			walker(n.Value)
			for _, c := range n.Cases {
				walker(c.Value)
				walker(c.Body)
			}
			if n.Default != nil {
				walker(n.Default)
			}
		case *ast.InfixExpression:
			walker(n.Left)
			walker(n.Right)
		case *ast.PrefixExpression:
			walker(n.Right)
		case *ast.ExpressionStatement:
			walker(n.Expression)
		case *ast.MesBlockStatement:
			walker(n.Body)
		case *ast.StepBlockStatement:
			walker(n.Body)
		}
	}
	walker(program)
	return usage
}

// collectVariablesInBlock collects all variable references in a block
// This is used to pass variables from parent scope to mes() blocks
// Enhanced to detect variables in nested expressions, function arguments, and array subscripts
func (g *Generator) collectVariablesInBlock(block *ast.BlockStatement) []string {
	vars := make(map[string]bool)

	var walker func(node ast.Node)
	walker = func(node ast.Node) {
		switch n := node.(type) {
		case *ast.Identifier:
			// Check if this is a variable reference (not a function name)
			varName := strings.ToLower(n.Value)
			// Skip known constants and functions
			if varName != "time" && varName != "midi_time" && varName != "user" &&
				varName != "midi_end" && !g.userFuncs[varName] {
				vars[varName] = true
			}
		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.ExpressionStatement:
			walker(n.Expression)
		case *ast.CallExpression:
			// Don't collect the function name itself, but DO collect arguments
			for _, arg := range n.Arguments {
				walker(arg)
			}
		case *ast.InfixExpression:
			// Collect variables from both sides of infix expressions (e.g., winW-320)
			walker(n.Left)
			walker(n.Right)
		case *ast.PrefixExpression:
			// Collect variables from prefix expressions (e.g., -x)
			walker(n.Right)
		case *ast.AssignStatement:
			// Collect variables from the value being assigned
			walker(n.Value)
			// Also collect from index if it's an array assignment
			if n.Index != nil {
				walker(n.Index)
			}
		case *ast.IfStatement:
			// Collect from condition and both branches
			walker(n.Condition)
			walker(n.Consequence)
			if n.Alternative != nil {
				walker(n.Alternative)
			}
		case *ast.ForStatement:
			// Collect from all parts of for loop
			if n.Init != nil {
				walker(n.Init)
			}
			if n.Condition != nil {
				walker(n.Condition)
			}
			if n.Post != nil {
				walker(n.Post)
			}
			walker(n.Body)
		case *ast.WhileStatement:
			// Collect from condition and body
			walker(n.Condition)
			walker(n.Body)
		case *ast.DoWhileStatement:
			// Collect from body and condition
			walker(n.Body)
			walker(n.Condition)
		case *ast.SwitchStatement:
			// Collect from switch value and all cases
			walker(n.Value)
			for _, c := range n.Cases {
				walker(c.Value)
				walker(c.Body)
			}
			if n.Default != nil {
				walker(n.Default)
			}
		case *ast.StepBlockStatement:
			walker(n.Body)
		case *ast.IndexExpression:
			// Collect from both the array and the index (e.g., arr[i])
			walker(n.Left)
			walker(n.Index)
		}
	}

	walker(block)

	// Convert to sorted slice for deterministic output
	result := make([]string, 0, len(vars))
	for v := range vars {
		result = append(result, v)
	}
	sort.Strings(result)
	return result
}

// scanMesBlocksForVMVars scans all mes() blocks in a function body
// and collects all variables that are referenced inside them.
// These variables need to be registered with the VM using engine.Assign()
func (g *Generator) scanMesBlocksForVMVars(block *ast.BlockStatement) map[string]bool {
	vmVars := make(map[string]bool)

	var walker func(node ast.Node)
	walker = func(node ast.Node) {
		switch n := node.(type) {
		case *ast.MesBlockStatement:
			// Collect all variables used in this mes() block
			vars := g.collectVariablesInBlock(n.Body)
			for _, v := range vars {
				vmVars[v] = true
			}
			// Don't recurse into the mes block body - we already collected its variables
		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.IfStatement:
			walker(n.Consequence)
			if n.Alternative != nil {
				walker(n.Alternative)
			}
		case *ast.ForStatement:
			walker(n.Body)
		case *ast.WhileStatement:
			walker(n.Body)
		case *ast.DoWhileStatement:
			walker(n.Body)
		case *ast.SwitchStatement:
			for _, c := range n.Cases {
				walker(c.Body)
			}
			if n.Default != nil {
				walker(n.Default)
			}
		case *ast.StepBlockStatement:
			walker(n.Body)
		}
	}

	walker(block)
	return vmVars
}

// inferType attempts to infer the type of an expression
func (g *Generator) inferType(expr ast.Expression) string {
	switch e := expr.(type) {
	case *ast.StringLiteral:
		return "string"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.CallExpression:
		// Check for known string-returning functions
		funcName := strings.ToLower(e.Function.Value)
		if funcName == "strcode" || funcName == "substr" || funcName == "strprint" ||
			funcName == "strinput" || funcName == "strup" || funcName == "strlow" ||
			funcName == "getinistr" || funcName == "strreadf" || funcName == "getcwd" {
			return "string"
		}
		// Default to int for most engine functions
		return "int"
	case *ast.InfixExpression:
		// If either operand is a string, result is string (for concatenation)
		leftType := g.inferType(e.Left)
		rightType := g.inferType(e.Right)
		if leftType == "string" || rightType == "string" {
			return "string"
		}
		return "int"
	case *ast.IndexExpression:
		// Array access returns int
		return "int"
	case *ast.Identifier:
		// Check if it's a known array
		varName := strings.ToLower(e.Value)
		if g.arrayUsage[varName] {
			return "[]int"
		}
		// Default to int
		return "int"
	default:
		return "int"
	}
}

func (g *Generator) scanLocals(block *ast.BlockStatement) map[string]string {
	locals := map[string]string{}
	var walker func(node ast.Node)
	walker = func(node ast.Node) {
		switch n := node.(type) {
		case *ast.AssignStatement:
			name := strings.ToLower(n.Name.Value)
			// infers type
			typ := "int"
			if lit, ok := n.Value.(*ast.StringLiteral); ok {
				_ = lit
				typ = "string"
			}
			// Check expression type?
			// If CallExpression (CIText), assumed void? No assignment.
			// If CallExpression (StrCode/SubStr), returns string?
			if call, ok := n.Value.(*ast.CallExpression); ok {
				// If known string func
				if call.Function.Value == "StrCode" || call.Function.Value == "SubStr" {
					typ = "string"
				}
			}
			// If Binary + and one is string?
			if _, exists := locals[name]; !exists {
				locals[name] = typ
			} else {
				// Upgrade int to string if mixed usage?
				if typ == "string" {
					locals[name] = "string"
				}
			}
		case *ast.LetStatement:
			name := strings.ToLower(n.Name.Value)
			typ := "int"
			if n.Token.Type == token.STR || n.Token.Literal == "str" {
				typ = "string"
			}
			if strings.HasSuffix(n.Name.Value, "[]") {
				name = strings.TrimSuffix(name, "[]")
				typ = "[]int"
			}
			locals[name] = typ
		case *ast.ForStatement:
			// Init stmt?
			if assign, ok := n.Init.(*ast.AssignStatement); ok {
				name := strings.ToLower(assign.Name.Value)
				if _, exists := locals[name]; !exists {
					locals[name] = "int"
				}
			}
			walker(n.Body)
		case *ast.WhileStatement:
			walker(n.Body)
		case *ast.DoWhileStatement:
			walker(n.Body)
		case *ast.SwitchStatement:
			for _, c := range n.Cases {
				walker(c.Body)
			}
			if n.Default != nil {
				walker(n.Default)
			}
		case *ast.BlockStatement:
			for _, s := range n.Statements {
				walker(s)
			}
		case *ast.IfStatement:
			walker(n.Consequence)
			if n.Alternative != nil {
				walker(n.Alternative)
			}
		case *ast.MesBlockStatement:
			walker(n.Body)
		case *ast.StepBlockStatement:
			walker(n.Body)
		case *ast.IndexExpression:
			// check usage
			if ident, ok := n.Left.(*ast.Identifier); ok {
				name := strings.ToLower(ident.Value)
				if _, exists := locals[name]; !exists {
					locals[name] = "[]int"
				} else {
					// Upgrade
					locals[name] = "[]int"
				}
			}
			walker(n.Left)
			walker(n.Index)
		case *ast.CallExpression:
			for _, arg := range n.Arguments {
				walker(arg)
			}
		case *ast.InfixExpression:
			walker(n.Left)
			walker(n.Right)
		case *ast.PrefixExpression:
			walker(n.Right)
		case *ast.ExpressionStatement:
			walker(n.Expression)
		}
	}
	walker(block)
	return locals
}

func (g *Generator) genStatement(stmt ast.Statement) {
	if stmt == nil {
		return
	}

	switch s := stmt.(type) {
	case *ast.LetStatement:
		// With scanLocals, local `let` statements only need to initialize if they are arrays.
		// Simple `let x` will be declared by scanLocals.
		name := strings.ToLower(s.Name.Value)
		if strings.HasSuffix(s.Name.Value, "[]") {
			name = strings.TrimSuffix(name, "[]")
			// Initialize slice to avoid panic on index
			g.buf.WriteString(fmt.Sprintf("\t%s = make([]int, 1000)\n", name))
		}
		// No need to declare `var x int` if it's already declared by scanLocals.
		// If it's a global, it's declared at the top.
		// If it's a parameter, it's in the function signature.

	case *ast.ExpressionStatement:
		g.genExpression(s.Expression)
		g.buf.WriteString("\n")
	case *ast.AssignStatement:
		g.buf.WriteString("\t")
		varName := strings.ToLower(s.Name.Value)
		g.buf.WriteString(varName)
		if s.Index != nil {
			g.buf.WriteString("[")
			g.genExpression(s.Index)
			g.buf.WriteString("]")
			g.buf.WriteString(" = ") // Accessing existing array
			g.genExpression(s.Value)
		} else {
			// Check if this variable needs VM registration
			if g.vmVars != nil && g.vmVars[varName] {
				// Generate engine.Assign() call for VM registration
				g.buf.WriteString(" = engine.Assign(")
				g.buf.WriteString(fmt.Sprintf("%q, ", s.Name.Value))
				g.genExpression(s.Value)
				g.buf.WriteString(")")

				// Add type assertion based on inferred type
				typ := g.inferType(s.Value)
				if typ == "string" {
					g.buf.WriteString(".(string)")
				} else if typ == "[]int" {
					g.buf.WriteString(".([]int)")
				} else {
					g.buf.WriteString(".(int)")
				}
			} else {
				// Normal assignment for variables not used in mes() blocks
				g.buf.WriteString(" = ")
				g.genExpression(s.Value)
			}
		}
		g.buf.WriteString("\n")

	case *ast.WaitStatement:
		// ,,,, -> engine.Wait(4)
		g.buf.WriteString(fmt.Sprintf("\tengine.Wait(%d)\n", s.Count))

	case *ast.IfStatement:
		g.buf.WriteString("\tif ")
		g.genExpression(s.Condition)
		g.buf.WriteString(" {\n")
		g.genBlock(s.Consequence)
		g.buf.WriteString("\t}")
		if s.Alternative != nil {
			g.buf.WriteString(" else {\n")
			g.genBlock(s.Alternative)
			g.buf.WriteString("\t}")
		}
		g.buf.WriteString("\n")

	case *ast.ForStatement:
		g.buf.WriteString("\tfor ")
		if s.Init != nil {
			// Force Assignment syntax, not :=
			// Because vars declared at top.
			if assign, ok := s.Init.(*ast.AssignStatement); ok {
				// Go: for i=0; ...
				// Filly: for(i=0; ...
				// Since vars are global or pre-declared, = is fine.
				g.buf.WriteString(strings.ToLower(assign.Name.Value))
				g.buf.WriteString(" = ")
				g.genExpression(assign.Value)
			} else {
				g.genStatement(s.Init)
			}
		}
		g.buf.WriteString("; ")
		if s.Condition != nil {
			g.genExpression(s.Condition)
		}
		g.buf.WriteString("; ")
		if s.Post != nil {
			if assign, ok := s.Post.(*ast.AssignStatement); ok {
				g.buf.WriteString(strings.ToLower(assign.Name.Value))
				g.buf.WriteString(" = ")
				g.genExpression(assign.Value)
			}
		}
		g.buf.WriteString(" {\n")
		g.genBlock(s.Body)
		g.buf.WriteString("\t}\n")

	case *ast.WhileStatement:
		g.buf.WriteString("\tfor ")
		g.genExpression(s.Condition)
		g.buf.WriteString(" {\n")
		g.genBlock(s.Body)
		g.buf.WriteString("\t}\n")

	case *ast.DoWhileStatement:
		// Go doesn't have do-while, so we use for with a break condition
		g.buf.WriteString("\tfor {\n")
		g.genBlock(s.Body)
		g.buf.WriteString("\t\tif !(")
		g.genExpression(s.Condition)
		g.buf.WriteString(") {\n")
		g.buf.WriteString("\t\t\tbreak\n")
		g.buf.WriteString("\t\t}\n")
		g.buf.WriteString("\t}\n")

	case *ast.SwitchStatement:
		g.buf.WriteString("\tswitch ")
		g.genExpression(s.Value)
		g.buf.WriteString(" {\n")
		for _, c := range s.Cases {
			g.buf.WriteString("\tcase ")
			g.genExpression(c.Value)
			g.buf.WriteString(":\n")
			g.genBlock(c.Body)
		}
		if s.Default != nil {
			g.buf.WriteString("\tdefault:\n")
			g.genBlock(s.Default)
		}
		g.buf.WriteString("\t}\n")

	case *ast.BreakStatement:
		g.buf.WriteString("\tbreak\n")

	case *ast.ContinueStatement:
		g.buf.WriteString("\tcontinue\n")

	case *ast.MesBlockStatement:
		// mes(TIME) { ... } -> engine.RegisterSequence(TIME, []engine.OpCode{ ... })
		// mes(MIDI_END) { ... } -> engine.RegisterMidiEndHandler(func() { ... })
		// mes(RBDOWN) { ... } -> engine.RegisterRBDownHandler(func() { ... })
		// mes(RBDBLCLK) { ... } -> engine.RegisterRBDblClkHandler(func() { ... })

		// Check if this is an event handler (MIDI_END, RBDOWN, RBDBLCLK)
		if ident, ok := s.Time.(*ast.Identifier); ok {
			timeValue := strings.ToLower(ident.Value)
			if timeValue == "midi_end" {
				// Generate event handler registration
				g.buf.WriteString("\tengine.RegisterMidiEndHandler(func() {\n")
				g.genBlock(s.Body)
				g.buf.WriteString("\t})\n")
				return
			} else if timeValue == "rbdown" {
				g.buf.WriteString("\tengine.RegisterRBDownHandler(func() {\n")
				g.genBlock(s.Body)
				g.buf.WriteString("\t})\n")
				return
			} else if timeValue == "rbdblclk" {
				g.buf.WriteString("\tengine.RegisterRBDblClkHandler(func() {\n")
				g.genBlock(s.Body)
				g.buf.WriteString("\t})\n")
				return
			}
		}

		// Regular mes() block for TIME or MIDI_TIME
		// Collect variables used in the mes block
		usedVars := g.collectVariablesInBlock(s.Body)

		// Generate code to pass variables to RegisterSequence
		g.buf.WriteString("\tengine.RegisterSequence(")
		g.genExpression(s.Time)
		g.buf.WriteString(", []engine.OpCode{\n")
		g.genOpCodes(s.Body)
		g.buf.WriteString("\t}, map[string]any{")

		// Add used variables to the map
		first := true
		for _, varName := range usedVars {
			if !first {
				g.buf.WriteString(", ")
			}
			first = false
			g.buf.WriteString(fmt.Sprintf("%q: %s", varName, varName))
		}
		g.buf.WriteString("})\n")

	case *ast.StepBlockStatement:
		// step(8) { ... }
		g.buf.WriteString(fmt.Sprintf("\tengine.SetStep(%d)\n", s.Count))
		g.genBlock(s.Body)
	}
}

func (g *Generator) genBlock(block *ast.BlockStatement) {
	if block == nil {
		return
	}
	for _, s := range block.Statements {
		g.genStatement(s)
	}
}

func (g *Generator) genExpression(expr ast.Expression) {
	if expr == nil {
		return
	}

	switch e := expr.(type) {
	case *ast.CallExpression:
		// MovePic -> engine.MovePic
		// Check global helper function map?
		// For now assume engine.FunctionName for builtins, and just FunctionName for user definitions.
		// How to distinguish?
		// Use TitleCase for engine? User funcs seem to be capitalized too in TFY (Scene1).
		// We should track defined functions.
		// For now, let's prefix everything with `engine.` EXCEPT if we know it's user defined.
		// But in this pass we don't know user defined functions list easily (unless we pass it).
		// Let's just output `e.Function.Value` and rely on `engine` package being dot-imported?
		// Or helper definitions being in `main`.
		// If I generate `func Scene1(...)` in main, calling `Scene1(...)` is fine.
		// If I call `MovePic`, and `MovePic` is not in main, it errors.
		// Unless I dot-import `pkg/engine`.
		// In `codegen.go` headers (Step 1130), I generated `import "github.com/.../engine"`.
		// I did NOT dot import.
		// So `engine.MovePic` is needed.
		// And `Scene1` is needed.
		// I'll leave the `engine.` prefix logic for later refinement or use a quick heuristic:
		// If implementation_plan said "all builtins in engine", I should prefix.
		funcName := e.Function.Value
		// Handle Default Arguments for User Functions
		args := e.Arguments
		if sig, ok := g.signatures[funcName]; ok {
			// Check count
			paramCount := len(sig.Parameters)
			argCount := len(args)
			if argCount < paramCount {
				// Append missing defaults
				// Note: This assumes positional missing args are at the end.
				// FILLY naming parameters?
				// Assuming positional.
				for i := argCount; i < paramCount; i++ {
					param := sig.Parameters[i]
					if param.Default != nil {
						args = append(args, param.Default)
					} else {
						// Error? Missing required arg.
						// Can't do much here, let Go compiler complain.
					}
				}
			}
		} else if strings.ToLower(funcName) == "openwin" {
			// Engine OpenWin expects 8 arguments, pad with 0s
			for len(args) < 8 {
				args = append(args, &ast.IntegerLiteral{Value: 0})
			}
		}

		if g.userFuncs[funcName] {
			g.buf.WriteString(funcName)
		} else {
			g.buf.WriteString("engine." + funcName)
		}
		g.buf.WriteString("(")
		for i, arg := range args {
			g.genExpression(arg)
			if i < len(args)-1 {
				g.buf.WriteString(", ")
			}
		}
		g.buf.WriteString(")")

	case *ast.Identifier:
		val := strings.ToLower(e.Value)
		switch val {
		case "midi_time":
			g.buf.WriteString("engine.MidiTime")
		case "end_step":
			g.buf.WriteString("engine.EndStep()")
		case "del_me":
			g.buf.WriteString("engine.DelMe()")
		case "del_us":
			g.buf.WriteString("engine.DelUs()")
		case "del_all": // Seen in TOKYOB1
			g.buf.WriteString("engine.DelAll()")
		case "maint": // Seen in TOKYOB1
			g.buf.WriteString("engine.Maint()")
		case "mesp1", "mesp2", "mesp3", "mesp4":
			g.buf.WriteString("engine.MesP" + strings.TrimPrefix(val, "mesp"))
		default:
			g.buf.WriteString(val)
		}

	case *ast.StringLiteral:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))

	case *ast.IntegerLiteral:
		g.buf.WriteString(fmt.Sprintf("%d", e.Value))

	case *ast.PrefixExpression:
		g.buf.WriteString(e.Operator)
		g.genExpression(e.Right)

	case *ast.InfixExpression:
		g.buf.WriteString("(")
		g.genExpression(e.Left)
		g.buf.WriteString(" " + e.Operator + " ")
		g.genExpression(e.Right)
		g.buf.WriteString(")")

	case *ast.IndexExpression:
		g.genExpression(e.Left)
		if e.Index != nil {
			g.buf.WriteString("[")
			g.genExpression(e.Index)
			g.buf.WriteString("]")
		}
	}
}

func (g *Generator) genOpArg(expr ast.Expression) {
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		g.buf.WriteString(fmt.Sprintf("%d", e.Value))
	case *ast.StringLiteral:
		g.buf.WriteString(fmt.Sprintf("%q", e.Value))
	case *ast.Identifier:
		name := e.Value
		if strings.ToLower(name) == "midi_time" {
			g.buf.WriteString("engine.MidiTime")
		} else if g.userFuncs[name] {
			g.buf.WriteString(name)
		} else {
			// Check if array
			lower := strings.ToLower(name)
			if g.arrayUsage[lower] {
				g.buf.WriteString(lower) // Convert array identifier to lowercase
			} else {
				g.buf.WriteString(fmt.Sprintf("engine.Variable(%q)", name))
			}
		}
	case *ast.CallExpression:
		// Nested Call -> engine.OpCode{Cmd: "...", Args: ...}
		g.buf.WriteString("engine.OpCode{Cmd: \"")
		g.buf.WriteString(e.Function.Value)
		g.buf.WriteString("\", Args: []any{")
		for i, arg := range e.Arguments {
			g.genOpArg(arg)
			if i < len(e.Arguments)-1 {
				g.buf.WriteString(", ")
			}
		}
		g.buf.WriteString("}}")
	case *ast.InfixExpression:
		// Generate OpCode for infix expression (for conditions)
		g.buf.WriteString("engine.OpCode{Cmd: \"Infix\", Args: []any{")
		g.buf.WriteString(fmt.Sprintf("%q, ", e.Operator))
		g.genOpArg(e.Left)
		g.buf.WriteString(", ")
		g.genOpArg(e.Right)
		g.buf.WriteString("}}")
	case *ast.PrefixExpression:
		// Generate OpCode for prefix expression
		g.buf.WriteString("engine.OpCode{Cmd: \"Prefix\", Args: []any{")
		g.buf.WriteString(fmt.Sprintf("%q, ", e.Operator))
		g.genOpArg(e.Right)
		g.buf.WriteString("}}")
	case *ast.IndexExpression:
		// Generate OpCode for array indexing
		g.buf.WriteString("engine.OpCode{Cmd: \"Index\", Args: []any{")
		g.genOpArg(e.Left)
		g.buf.WriteString(", ")
		g.genOpArg(e.Index)
		g.buf.WriteString("}}")
	default:
		g.buf.WriteString("nil")
	}
}

func (g *Generator) genOpCodes(block *ast.BlockStatement) {
	if block == nil {
		return
	}
	for _, s := range block.Statements {
		switch stmt := s.(type) {
		case *ast.ExpressionStatement:
			if call, ok := stmt.Expression.(*ast.CallExpression); ok {
				g.buf.WriteString("\t\t{Cmd: \"")
				funcName := call.Function.Value
				g.buf.WriteString(funcName)
				g.buf.WriteString("\", Args: []any{")
				for i, arg := range call.Arguments {
					g.genOpArg(arg)
					if i < len(call.Arguments)-1 {
						g.buf.WriteString(", ")
					}
				}
				g.buf.WriteString("}},\n")
			}
		case *ast.WaitStatement:
			g.buf.WriteString(fmt.Sprintf("\t\t{Cmd: \"Wait\", Args: []any{%d}},\n", stmt.Count))

		case *ast.StepBlockStatement:
			g.buf.WriteString(fmt.Sprintf("\t\t{Cmd: \"SetStep\", Args: []any{%d}},\n", stmt.Count))
			g.genOpCodes(stmt.Body)

		case *ast.AssignStatement:
			// {Cmd: "Assign", Args: []any{"Name", ValueOpArg}}
			g.buf.WriteString("\t\t{Cmd: \"Assign\", Args: []any{")
			// Name
			g.buf.WriteString(fmt.Sprintf("%q, ", stmt.Name.Value))
			// Value
			g.genOpArg(stmt.Value)
			g.buf.WriteString("}},\n")

		case *ast.IfStatement:
			// Generate if-else OpCode
			g.buf.WriteString("\t\t{Cmd: \"If\", Args: []any{")
			g.genOpArg(stmt.Condition)
			g.buf.WriteString(", []engine.OpCode{\n")
			g.genOpCodes(stmt.Consequence)
			g.buf.WriteString("\t\t}")
			if stmt.Alternative != nil {
				g.buf.WriteString(", []engine.OpCode{\n")
				g.genOpCodes(stmt.Alternative)
				g.buf.WriteString("\t\t}")
			}
			g.buf.WriteString("}},\n")

		case *ast.ForStatement:
			// Generate for loop OpCode
			g.buf.WriteString("\t\t{Cmd: \"For\", Args: []any{")
			// Init statement
			if stmt.Init != nil {
				if assign, ok := stmt.Init.(*ast.AssignStatement); ok {
					g.buf.WriteString("engine.OpCode{Cmd: \"Assign\", Args: []any{")
					g.buf.WriteString(fmt.Sprintf("%q, ", assign.Name.Value))
					g.genOpArg(assign.Value)
					g.buf.WriteString("}}, ")
				} else {
					g.buf.WriteString("nil, ")
				}
			} else {
				g.buf.WriteString("nil, ")
			}
			// Condition
			if stmt.Condition != nil {
				g.genOpArg(stmt.Condition)
			} else {
				g.buf.WriteString("nil")
			}
			g.buf.WriteString(", ")
			// Post statement
			if stmt.Post != nil {
				if assign, ok := stmt.Post.(*ast.AssignStatement); ok {
					g.buf.WriteString("engine.OpCode{Cmd: \"Assign\", Args: []any{")
					g.buf.WriteString(fmt.Sprintf("%q, ", assign.Name.Value))
					g.genOpArg(assign.Value)
					g.buf.WriteString("}}, ")
				} else {
					g.buf.WriteString("nil, ")
				}
			} else {
				g.buf.WriteString("nil, ")
			}
			// Body
			g.buf.WriteString("[]engine.OpCode{\n")
			g.genOpCodes(stmt.Body)
			g.buf.WriteString("\t\t}}},\n")

		case *ast.WhileStatement:
			// Generate while loop OpCode
			g.buf.WriteString("\t\t{Cmd: \"While\", Args: []any{")
			g.genOpArg(stmt.Condition)
			g.buf.WriteString(", []engine.OpCode{\n")
			g.genOpCodes(stmt.Body)
			g.buf.WriteString("\t\t}}},\n")

		case *ast.DoWhileStatement:
			// Generate do-while loop OpCode
			g.buf.WriteString("\t\t{Cmd: \"DoWhile\", Args: []any{")
			g.genOpArg(stmt.Condition)
			g.buf.WriteString(", []engine.OpCode{\n")
			g.genOpCodes(stmt.Body)
			g.buf.WriteString("\t\t}}},\n")

		case *ast.SwitchStatement:
			// Generate switch-case OpCode
			g.buf.WriteString("\t\t{Cmd: \"Switch\", Args: []any{")
			g.genOpArg(stmt.Value)
			g.buf.WriteString(", []any{\n")
			// Generate cases
			for _, c := range stmt.Cases {
				g.buf.WriteString("\t\t\t[]any{")
				g.genOpArg(c.Value)
				g.buf.WriteString(", []engine.OpCode{\n")
				g.genOpCodes(c.Body)
				g.buf.WriteString("\t\t\t}},\n")
			}
			g.buf.WriteString("\t\t}")
			// Default case
			if stmt.Default != nil {
				g.buf.WriteString(", []engine.OpCode{\n")
				g.genOpCodes(stmt.Default)
				g.buf.WriteString("\t\t}")
			} else {
				g.buf.WriteString(", nil")
			}
			g.buf.WriteString("}},\n")

		case *ast.BreakStatement:
			g.buf.WriteString("\t\t{Cmd: \"Break\", Args: []any{}},\n")

		case *ast.ContinueStatement:
			g.buf.WriteString("\t\t{Cmd: \"Continue\", Args: []any{}},\n")

		case *ast.MesBlockStatement:
			// Nested

		default:
			g.buf.WriteString(fmt.Sprintf("\t\t// Warning: Unsupported statement type in VM mode\n"))
		}
	}
}
