// Package parser provides syntax analysis for FILLY scripts (.TFY files).
package parser

import (
	"github.com/zurustar/son-et/pkg/compiler/lexer"
)

// Node is the interface for all AST nodes.
type Node interface {
	TokenLiteral() string
}

// Statement is the interface for all statement nodes.
type Statement interface {
	Node
	statementNode()
}

// Expression is the interface for all expression nodes.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST.
type Program struct {
	Statements []Statement
}

// TokenLiteral returns the literal value of the first statement's token.
func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

// VarDeclaration represents a variable declaration statement.
// Example: int x, y[]; str s;
type VarDeclaration struct {
	Token   lexer.Token  // int or str token
	Type    string       // "int" or "str"
	Names   []string     // variable names
	IsArray []bool       // whether each variable is an array
	Sizes   []Expression // array sizes (if specified)
}

func (vd *VarDeclaration) statementNode()       {}
func (vd *VarDeclaration) TokenLiteral() string { return vd.Token.Literal }

// InfoDirective represents a #info preprocessor directive.
// Example: #info INAM "Title Name"
type InfoDirective struct {
	Token lexer.Token
	Key   string // INAM, ISBJ, IART, etc.
	Value string // directive value
}

func (id *InfoDirective) statementNode()       {}
func (id *InfoDirective) TokenLiteral() string { return id.Token.Literal }

// IncludeDirective represents a #include preprocessor directive.
// Example: #include "filename.tfy"
type IncludeDirective struct {
	Token    lexer.Token
	FileName string // included file name
}

func (id *IncludeDirective) statementNode()       {}
func (id *IncludeDirective) TokenLiteral() string { return id.Token.Literal }

// DefineDirective represents a #define preprocessor directive.
// Example: #define MAXLINE 24
type DefineDirective struct {
	Token lexer.Token
	Name  string // macro name
	Value string // macro value
}

func (dd *DefineDirective) statementNode()       {}
func (dd *DefineDirective) TokenLiteral() string { return dd.Token.Literal }

// FunctionStatement represents a function definition.
// Example: name(params){body}
type FunctionStatement struct {
	Token      lexer.Token
	Name       string
	Parameters []*Parameter
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }

// Parameter represents a function parameter.
// Example: int x, y[], str s, p[], c[], x=10, int time=1
type Parameter struct {
	Name         string
	Type         string     // "int", "str", or "" (no type specified)
	IsArray      bool       // whether this is an array parameter
	DefaultValue Expression // default value (optional)
}

// BlockStatement represents a block of statements enclosed in braces.
type BlockStatement struct {
	Token      lexer.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }

// AssignStatement represents an assignment statement.
// Example: x = expr or arr[i] = value
type AssignStatement struct {
	Token lexer.Token
	Name  Expression // Identifier or IndexExpression
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }

// ExpressionStatement represents a statement consisting of a single expression.
type ExpressionStatement struct {
	Token      lexer.Token
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }

// IfStatement represents an if statement.
// Example: if (condition) { consequence } else { alternative }
type IfStatement struct {
	Token       lexer.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative Statement // *BlockStatement or *IfStatement (else if)
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }

// ForStatement represents a for loop.
// Example: for(i=0; i<10; i=i+1) { body }
type ForStatement struct {
	Token     lexer.Token
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }

// WhileStatement represents a while loop.
// Example: while (condition) { body }
type WhileStatement struct {
	Token     lexer.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }

// SwitchStatement represents a switch statement.
// Example: switch (value) { case 1: ... default: ... }
type SwitchStatement struct {
	Token   lexer.Token
	Value   Expression
	Cases   []*CaseClause
	Default *BlockStatement
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }

// CaseClause represents a case clause in a switch statement.
type CaseClause struct {
	Token lexer.Token
	Value Expression
	Body  []Statement
}

func (cc *CaseClause) statementNode()       {}
func (cc *CaseClause) TokenLiteral() string { return cc.Token.Literal }

// MesStatement represents a mes (event handler) statement.
// Example: mes(MIDI_TIME) { body }
type MesStatement struct {
	Token     lexer.Token
	EventType string // TIME, MIDI_TIME, MIDI_END, KEY, CLICK, RBDOWN, RBDBLCLK, USER
	Body      *BlockStatement
}

func (ms *MesStatement) statementNode()       {}
func (ms *MesStatement) TokenLiteral() string { return ms.Token.Literal }

// StepStatement represents a step statement.
// Example: step(10) { commands } or step { commands }
type StepStatement struct {
	Token lexer.Token
	Count Expression // step(n)'s n, nil if omitted
	Body  *StepBody
}

func (ss *StepStatement) statementNode()       {}
func (ss *StepStatement) TokenLiteral() string { return ss.Token.Literal }

// StepBody represents the body of a step statement.
type StepBody struct {
	Commands []*StepCommand
}

// StepCommand represents a command within a step body.
type StepCommand struct {
	Statement Statement // statement to execute (nil for wait-only)
	WaitCount int       // number of consecutive commas (wait steps)
}

// BreakStatement represents a break statement.
type BreakStatement struct {
	Token lexer.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }

// ContinueStatement represents a continue statement.
type ContinueStatement struct {
	Token lexer.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       lexer.Token
	ReturnValue Expression // optional return value
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }

// LabelStatement represents a label for goto statements.
// Example: END:
type LabelStatement struct {
	Token lexer.Token
	Name  string
}

func (ls *LabelStatement) statementNode()       {}
func (ls *LabelStatement) TokenLiteral() string { return ls.Token.Literal }

// ============================================================================
// Expression Nodes
// ============================================================================

// Identifier represents an identifier expression.
type Identifier struct {
	Token lexer.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }

// IntegerLiteral represents an integer literal expression.
type IntegerLiteral struct {
	Token lexer.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }

// FloatLiteral represents a floating point literal expression.
type FloatLiteral struct {
	Token lexer.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }

// StringLiteral represents a string literal expression.
type StringLiteral struct {
	Token lexer.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }

// BinaryExpression represents a binary expression.
// Example: left + right, x == y
type BinaryExpression struct {
	Token    lexer.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (be *BinaryExpression) expressionNode()      {}
func (be *BinaryExpression) TokenLiteral() string { return be.Token.Literal }

// UnaryExpression represents a unary expression.
// Example: -x, !flag
type UnaryExpression struct {
	Token    lexer.Token
	Operator string
	Right    Expression
}

func (ue *UnaryExpression) expressionNode()      {}
func (ue *UnaryExpression) TokenLiteral() string { return ue.Token.Literal }

// CallExpression represents a function call expression.
// Example: func(arg1, arg2)
type CallExpression struct {
	Token     lexer.Token
	Function  string
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }

// IndexExpression represents an array index expression.
// Example: arr[i]
type IndexExpression struct {
	Token lexer.Token
	Left  Expression // array
	Index Expression // index
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }

// ArrayReference represents a reference to an entire array (used as function argument).
// Example: func(arr[]) - passes the entire array
type ArrayReference struct {
	Token lexer.Token
	Name  string // array name
}

func (ar *ArrayReference) expressionNode()      {}
func (ar *ArrayReference) TokenLiteral() string { return ar.Token.Literal }
