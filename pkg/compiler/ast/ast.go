package ast

import "github.com/zurustar/son-et/pkg/compiler/token"

// Node is the base interface for all AST nodes.
type Node interface {
	TokenLiteral() string
	String() string
}

// Statement represents a statement node.
type Statement interface {
	Node
	statementNode()
}

// Expression represents an expression node.
type Expression interface {
	Node
	expressionNode()
}

// Program is the root node of the AST.
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
	result := ""
	for _, s := range p.Statements {
		result += s.String()
	}
	return result
}

// Identifier represents an identifier.
type Identifier struct {
	Token token.Token
	Value string
}

func (i *Identifier) expressionNode()      {}
func (i *Identifier) TokenLiteral() string { return i.Token.Literal }
func (i *Identifier) String() string       { return i.Value }

// IntegerLiteral represents an integer literal.
type IntegerLiteral struct {
	Token token.Token
	Value int64
}

func (il *IntegerLiteral) expressionNode()      {}
func (il *IntegerLiteral) TokenLiteral() string { return il.Token.Literal }
func (il *IntegerLiteral) String() string       { return il.Token.Literal }

// FloatLiteral represents a float literal.
type FloatLiteral struct {
	Token token.Token
	Value float64
}

func (fl *FloatLiteral) expressionNode()      {}
func (fl *FloatLiteral) TokenLiteral() string { return fl.Token.Literal }
func (fl *FloatLiteral) String() string       { return fl.Token.Literal }

// StringLiteral represents a string literal.
type StringLiteral struct {
	Token token.Token
	Value string
}

func (sl *StringLiteral) expressionNode()      {}
func (sl *StringLiteral) TokenLiteral() string { return sl.Token.Literal }
func (sl *StringLiteral) String() string       { return sl.Token.Literal }

// ArrayLiteral represents an array literal.
type ArrayLiteral struct {
	Token    token.Token
	Elements []Expression
}

func (al *ArrayLiteral) expressionNode()      {}
func (al *ArrayLiteral) TokenLiteral() string { return al.Token.Literal }
func (al *ArrayLiteral) String() string {
	result := "["
	for i, el := range al.Elements {
		if i > 0 {
			result += ", "
		}
		result += el.String()
	}
	result += "]"
	return result
}

// IndexExpression represents array indexing (arr[index]).
type IndexExpression struct {
	Token token.Token
	Left  Expression
	Index Expression
}

func (ie *IndexExpression) expressionNode()      {}
func (ie *IndexExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *IndexExpression) String() string {
	return ie.Left.String() + "[" + ie.Index.String() + "]"
}

// PrefixExpression represents a prefix expression (!x, -x).
type PrefixExpression struct {
	Token    token.Token
	Operator string
	Right    Expression
}

func (pe *PrefixExpression) expressionNode()      {}
func (pe *PrefixExpression) TokenLiteral() string { return pe.Token.Literal }
func (pe *PrefixExpression) String() string {
	return "(" + pe.Operator + pe.Right.String() + ")"
}

// InfixExpression represents an infix expression (x + y).
type InfixExpression struct {
	Token    token.Token
	Left     Expression
	Operator string
	Right    Expression
}

func (ie *InfixExpression) expressionNode()      {}
func (ie *InfixExpression) TokenLiteral() string { return ie.Token.Literal }
func (ie *InfixExpression) String() string {
	return "(" + ie.Left.String() + " " + ie.Operator + " " + ie.Right.String() + ")"
}

// CallExpression represents a function call.
type CallExpression struct {
	Token     token.Token
	Function  Expression
	Arguments []Expression
}

func (ce *CallExpression) expressionNode()      {}
func (ce *CallExpression) TokenLiteral() string { return ce.Token.Literal }
func (ce *CallExpression) String() string {
	result := ce.Function.String() + "("
	for i, arg := range ce.Arguments {
		if i > 0 {
			result += ", "
		}
		result += arg.String()
	}
	result += ")"
	return result
}

// AssignStatement represents an assignment statement.
type AssignStatement struct {
	Token token.Token
	Name  Expression // Can be Identifier or IndexExpression
	Value Expression
}

func (as *AssignStatement) statementNode()       {}
func (as *AssignStatement) TokenLiteral() string { return as.Token.Literal }
func (as *AssignStatement) String() string {
	return as.Name.String() + " = " + as.Value.String()
}

// ExpressionStatement represents an expression statement.
type ExpressionStatement struct {
	Token      token.Token
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

// BlockStatement represents a block of statements.
type BlockStatement struct {
	Token      token.Token
	Statements []Statement
}

func (bs *BlockStatement) statementNode()       {}
func (bs *BlockStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BlockStatement) String() string {
	result := "{ "
	for _, s := range bs.Statements {
		result += s.String() + " "
	}
	result += "}"
	return result
}

// IfStatement represents an if statement.
type IfStatement struct {
	Token       token.Token
	Condition   Expression
	Consequence *BlockStatement
	Alternative *BlockStatement
}

func (is *IfStatement) statementNode()       {}
func (is *IfStatement) TokenLiteral() string { return is.Token.Literal }
func (is *IfStatement) String() string {
	result := "if " + is.Condition.String() + " " + is.Consequence.String()
	if is.Alternative != nil {
		result += " else " + is.Alternative.String()
	}
	return result
}

// ForStatement represents a for loop.
type ForStatement struct {
	Token     token.Token
	Init      Statement
	Condition Expression
	Post      Statement
	Body      *BlockStatement
}

func (fs *ForStatement) statementNode()       {}
func (fs *ForStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *ForStatement) String() string {
	return "for (" + fs.Init.String() + "; " + fs.Condition.String() + "; " + fs.Post.String() + ") " + fs.Body.String()
}

// WhileStatement represents a while loop.
type WhileStatement struct {
	Token     token.Token
	Condition Expression
	Body      *BlockStatement
}

func (ws *WhileStatement) statementNode()       {}
func (ws *WhileStatement) TokenLiteral() string { return ws.Token.Literal }
func (ws *WhileStatement) String() string {
	return "while " + ws.Condition.String() + " " + ws.Body.String()
}

// SwitchStatement represents a switch statement.
type SwitchStatement struct {
	Token   token.Token
	Value   Expression
	Cases   []*CaseClause
	Default *BlockStatement
}

func (ss *SwitchStatement) statementNode()       {}
func (ss *SwitchStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *SwitchStatement) String() string {
	result := "switch " + ss.Value.String() + " { "
	for _, c := range ss.Cases {
		result += c.String() + " "
	}
	if ss.Default != nil {
		result += "default: " + ss.Default.String()
	}
	result += " }"
	return result
}

// CaseClause represents a case clause in a switch statement.
type CaseClause struct {
	Token token.Token
	Value Expression
	Body  *BlockStatement
}

func (cc *CaseClause) String() string {
	return "case " + cc.Value.String() + ": " + cc.Body.String()
}

// BreakStatement represents a break statement.
type BreakStatement struct {
	Token token.Token
}

func (bs *BreakStatement) statementNode()       {}
func (bs *BreakStatement) TokenLiteral() string { return bs.Token.Literal }
func (bs *BreakStatement) String() string       { return "break" }

// ContinueStatement represents a continue statement.
type ContinueStatement struct {
	Token token.Token
}

func (cs *ContinueStatement) statementNode()       {}
func (cs *ContinueStatement) TokenLiteral() string { return cs.Token.Literal }
func (cs *ContinueStatement) String() string       { return "continue" }

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Token       token.Token
	ReturnValue Expression
}

func (rs *ReturnStatement) statementNode()       {}
func (rs *ReturnStatement) TokenLiteral() string { return rs.Token.Literal }
func (rs *ReturnStatement) String() string {
	result := "return"
	if rs.ReturnValue != nil {
		result += " " + rs.ReturnValue.String()
	}
	return result
}

// FunctionStatement represents a function definition.
type FunctionStatement struct {
	Token      token.Token
	Name       *Identifier
	Parameters []*Identifier
	Body       *BlockStatement
}

func (fs *FunctionStatement) statementNode()       {}
func (fs *FunctionStatement) TokenLiteral() string { return fs.Token.Literal }
func (fs *FunctionStatement) String() string {
	result := "function " + fs.Name.String() + "("
	for i, p := range fs.Parameters {
		if i > 0 {
			result += ", "
		}
		result += p.String()
	}
	result += ") " + fs.Body.String()
	return result
}

// MesStatement represents a mes() block.
type MesStatement struct {
	Token     token.Token
	EventType token.TokenType
	Body      *BlockStatement
}

func (ms *MesStatement) statementNode()       {}
func (ms *MesStatement) TokenLiteral() string { return ms.Token.Literal }
func (ms *MesStatement) String() string {
	return "mes(" + ms.EventType.String() + ") " + ms.Body.String()
}

// StepStatement represents a step() call.
type StepStatement struct {
	Token token.Token
	Count Expression
}

func (ss *StepStatement) statementNode()       {}
func (ss *StepStatement) TokenLiteral() string { return ss.Token.Literal }
func (ss *StepStatement) String() string {
	return "step(" + ss.Count.String() + ")"
}
