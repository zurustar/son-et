package ast

import (
	"bytes"
	"strings"

	"github.com/zurustar/filly2exe/pkg/compiler/token"
)

type Node interface {
	TokenLiteral() string
	String() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

// Program is the root node
type Program struct {
	Statements []Statement
}

func (p *Program) TokenLiteral() string {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenLiteral()
	}
	return ""
}

func (p *Program) String() string {
	var out bytes.Buffer
	for _, s := range p.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// Identifier
type Identifier struct {
	Token token.Token // token.IDENT
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// ExpressionStatement
type ExpressionStatement struct {
	Token      token.Token // The first token of the expression
	Expression Expression
}

func (es *ExpressionStatement) statementNode()       {}
func (es *ExpressionStatement) TokenLiteral() string { return es.Token.Literal }
func (es *ExpressionStatement) String() string {
	if es.Expression != nil {
		return es.Expression.String()
	}
	return ""
}

// IntegerLiteral
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// StringLiteral
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return `"` + sl.Value + `"` }

// DefineStatement: #define NAME VALUE
type DefineStatement struct {
	Token token.Token // token used for '#' (we might use ILLEGAL or create one, or just store nothing?)
	// Actually lexer returns '#' as ILLEGAL? No, we might need to handle it.
	// Parser currently checks p.curToken.Literal == "#"

	Name  *Identifier
	Value Expression
}

func (ds *DefineStatement) statementNode()       {}
func (ds *DefineStatement) TokenLiteral() string { return ds.Token.Literal }
func (ds *DefineStatement) String() string {
	var out bytes.Buffer
	out.WriteString("#define ")
	out.WriteString(ds.Name.String())
	out.WriteString(" ")
	if ds.Value != nil {
		out.WriteString(ds.Value.String())
	}
	return out.String()
}

// AssignStatement: VAR = ... or VAR[IDX] = ...
type AssignStatement struct {
	Token token.Token // token.ASSIGN
	Name  *Identifier
	Index Expression // nil if simple assignment
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	var out bytes.Buffer
	out.WriteString(as.Name.String())
	if as.Index != nil {
		out.WriteString("[")
		out.WriteString(as.Index.String())
		out.WriteString("]")
	}
	out.WriteString(" = ")
	if as.Value != nil {
		out.WriteString(as.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// CallExpression: LoadPic(...)
type CallExpression struct {
	Token     token.Token // '('
	Function  *Identifier // Identifier
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	var out bytes.Buffer
	out.WriteString(ce.Function.String())
	out.WriteString("(")
	args := []string{}
	for _, a := range ce.Arguments {
		args = append(args, a.String())
	}
	out.WriteString(strings.Join(args, ", "))
	out.WriteString(")")
	return out.String()
}

// BlockStatement
type BlockStatement struct {
	Token      token.Token // '{'
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	var out bytes.Buffer
	for _, s := range bs.Statements {
		out.WriteString(s.String())
	}
	return out.String()
}

// MesBlockStatement: mes(time) { ... }
type MesBlockStatement struct {
	Token token.Token // token.MES
	Time  Expression  // Argument e.g. MIDI_TIME (Identifier)
	Body  *BlockStatement
}

func (mbs *MesBlockStatement) statementNode()       {}
func (mbs *MesBlockStatement) TokenLiteral() string { return mbs.Token.Literal }
func (mbs *MesBlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("mes(")
	out.WriteString(mbs.Time.String())
	out.WriteString(") ")
	out.WriteString(mbs.Body.String())
	return out.String()
}

// StepBlockStatement: step(8) { ... }
type StepBlockStatement struct {
	Token token.Token // token.STEP
	Count int64       // 8
	Body  *BlockStatement
}

func (sbs *StepBlockStatement) statementNode()       {}
func (sbs *StepBlockStatement) TokenLiteral() string { return sbs.Token.Literal }
func (sbs *StepBlockStatement) String() string {
	var out bytes.Buffer
	out.WriteString("step(")
	// out.WriteString(...) -> integer
	out.WriteString(") ")
	out.WriteString(sbs.Body.String())
	return out.String()
}

// WaitStatement: ,,,,
type WaitStatement struct {
	Token token.Token // token.COMMA
	Count int         // Number of commas
}

func (ws *WaitStatement) statementNode()       {}
func (ws *WaitStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WaitStatement) String() string {
	return strings.Repeat(",", ws.Count)
}

// LetStatement: int Name;
type LetStatement struct {
	Token token.Token // token.INT
	Name  *Identifier
	Value Expression
}

func (ls *LetStatement) statementNode()       {}
func (ls *LetStatement) TokenLiteral() string { return ls.Token.Literal }
func (ls *LetStatement) String() string {
	var out bytes.Buffer
	out.WriteString(ls.TokenLiteral() + " ")
	out.WriteString(ls.Name.String())
	if ls.Value != nil {
		out.WriteString(" = ")
		out.WriteString(ls.Value.String())
	}
	out.WriteString(";")
	return out.String()
}

// Parameter
type Parameter struct {
	Name    *Identifier
	Type    string     // "int", "string", etc.
	Default Expression // nil if required
}

func (p *Parameter) String() string {
	if p.Default != nil {
		return p.Name.String() + "=" + p.Default.String()
	}
	return p.Name.String()
}

// FunctionStatement: Name(Params) { ... }
type FunctionStatement struct {
	Token      token.Token // token.IDENT
	Name       *Identifier
	Parameters []*Parameter
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	var out bytes.Buffer
	out.WriteString(fs.Name.String())
	out.WriteString("(")
	params := []string{}
	for _, p := range fs.Parameters {
		params = append(params, p.String())
	}
	out.WriteString(strings.Join(params, ", "))
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// IfStatement
type IfStatement struct {
	Token       token.Token // token.IF
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement // else
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	var out bytes.Buffer
	out.WriteString("if")
	out.WriteString(is.Condition.String())
	out.WriteString(" ")
	out.WriteString(is.Consequence.String())
	if is.Alternative != nil {
		out.WriteString(" else ")
		out.WriteString(is.Alternative.String())
	}
	return out.String()
}

// ForStatement
type ForStatement struct {
	Token     token.Token // token.FOR
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	var out bytes.Buffer
	out.WriteString("for(")
	if fs.Init != nil {
		out.WriteString(fs.Init.String()) // includes ;
	} else {
		out.WriteString(";")
	}
	if fs.Condition != nil {
		out.WriteString(fs.Condition.String())
	}
	out.WriteString(";")
	if fs.Post != nil {
		out.WriteString(fs.Post.String())
	}
	out.WriteString(") ")
	out.WriteString(fs.Body.String())
	return out.String()
}

// PrefixExpression
type PrefixExpression struct {
	Token    token.Token // !, -
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(pe.Operator)
	out.WriteString(pe.Right.String())
	out.WriteString(")")
	return out.String()
}

// InfixExpression
type InfixExpression struct {
	Token    token.Token // +, -, *, /, ==, >, <
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString(" " + ie.Operator + " ")
	out.WriteString(ie.Right.String())
	out.WriteString(")")
	return out.String()
}

// IndexExpression
type IndexExpression struct {
	Token token.Token // [
	Left  Expression
	Index Expression // nil if []
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	var out bytes.Buffer
	out.WriteString("(")
	out.WriteString(ie.Left.String())
	out.WriteString("[")
	if ie.Index != nil {
		out.WriteString(ie.Index.String())
	}
	out.WriteString("])")
	return out.String()
}

// WhileStatement
type WhileStatement struct {
	Token     token.Token // token.WHILE
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("while(")
	out.WriteString(ws.Condition.String())
	out.WriteString(") ")
	out.WriteString(ws.Body.String())
	return out.String()
}

// DoWhileStatement
type DoWhileStatement struct {
	Token     token.Token // token.DO
	Body      *BlockStatement
	Condition Expression
}

func (dws *DoWhileStatement) statementNode()       {}
func (dws *DoWhileStatement) TokenLiteral() string { return dws.Token.Literal }
func (dws *DoWhileStatement) String() string {
	var out bytes.Buffer
	out.WriteString("do ")
	out.WriteString(dws.Body.String())
	out.WriteString(" while(")
	out.WriteString(dws.Condition.String())
	out.WriteString(")")
	return out.String()
}

// SwitchStatement
type SwitchStatement struct {
	Token   token.Token // token.SWITCH
	Value   Expression
	Cases   []*CaseClause
	Default *BlockStatement
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	var out bytes.Buffer
	out.WriteString("switch(")
	out.WriteString(ss.Value.String())
	out.WriteString(") {")
	for _, c := range ss.Cases {
		out.WriteString(c.String())
	}
	if ss.Default != nil {
		out.WriteString("default: ")
		out.WriteString(ss.Default.String())
	}
	out.WriteString("}")
	return out.String()
}

// CaseClause
type CaseClause struct {
	Token token.Token // token.CASE
	Value Expression
	Body  *BlockStatement
}

func (cc *CaseClause) String() string {
	var out bytes.Buffer
	out.WriteString("case ")
	out.WriteString(cc.Value.String())
	out.WriteString(": ")
	out.WriteString(cc.Body.String())
	return out.String()
}

// BreakStatement
type BreakStatement struct {
	Token token.Token // token.BREAK
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break" }

// ContinueStatement
type ContinueStatement struct {
	Token token.Token // token.CONTINUE
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue" }
