// Package lexer provides lexical analysis for FILLY scripts (.TFY files).
package lexer

import "strings"

// TokenType represents the type of a token.
type TokenType int

// Token types
const (
	// Special tokens
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_COMMENT

	// Literals
	TOKEN_IDENT  // identifier
	TOKEN_INT    // integer literal
	TOKEN_FLOAT  // floating point literal
	TOKEN_STRING // string literal

	// Preprocessor directives
	TOKEN_DIRECTIVE // generic directive
	TOKEN_INFO      // #info
	TOKEN_INCLUDE   // #include
	TOKEN_DEFINE    // #define

	// Operators
	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_ASTERISK // *
	TOKEN_SLASH    // /
	TOKEN_PERCENT  // %
	TOKEN_ASSIGN   // =
	TOKEN_EQ       // ==
	TOKEN_NEQ      // !=
	TOKEN_LT       // <
	TOKEN_GT       // >
	TOKEN_LTE      // <=
	TOKEN_GTE      // >=
	TOKEN_AND      // &&
	TOKEN_OR       // ||
	TOKEN_NOT      // !

	// Delimiters
	TOKEN_LPAREN    // (
	TOKEN_RPAREN    // )
	TOKEN_LBRACE    // {
	TOKEN_RBRACE    // }
	TOKEN_LBRACKET  // [
	TOKEN_RBRACKET  // ]
	TOKEN_COMMA     // ,
	TOKEN_SEMICOLON // ;
	TOKEN_COLON     // : (for case labels)

	// Keywords
	TOKEN_INT_TYPE  // int
	TOKEN_STR_TYPE  // str
	TOKEN_REAL_TYPE // real (floating point)
	TOKEN_IF        // if
	TOKEN_ELSE      // else
	TOKEN_FOR       // for
	TOKEN_WHILE     // while
	TOKEN_SWITCH    // switch
	TOKEN_CASE      // case
	TOKEN_DEFAULT   // default
	TOKEN_BREAK     // break
	TOKEN_CONTINUE  // continue
	TOKEN_RETURN    // return
	TOKEN_MES       // mes
	TOKEN_STEP      // step
	TOKEN_END_STEP  // end_step
	TOKEN_DEL_ME    // del_me
	TOKEN_DEL_US    // del_us
	TOKEN_DEL_ALL   // del_all
)

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// tokenTypeNames maps TokenType to its string representation.
var tokenTypeNames = map[TokenType]string{
	// Special tokens
	TOKEN_ILLEGAL: "ILLEGAL",
	TOKEN_EOF:     "EOF",
	TOKEN_COMMENT: "COMMENT",

	// Literals
	TOKEN_IDENT:  "IDENT",
	TOKEN_INT:    "INT",
	TOKEN_FLOAT:  "FLOAT",
	TOKEN_STRING: "STRING",

	// Preprocessor directives
	TOKEN_DIRECTIVE: "DIRECTIVE",
	TOKEN_INFO:      "INFO",
	TOKEN_INCLUDE:   "INCLUDE",
	TOKEN_DEFINE:    "DEFINE",

	// Operators
	TOKEN_PLUS:     "+",
	TOKEN_MINUS:    "-",
	TOKEN_ASTERISK: "*",
	TOKEN_SLASH:    "/",
	TOKEN_PERCENT:  "%",
	TOKEN_ASSIGN:   "=",
	TOKEN_EQ:       "==",
	TOKEN_NEQ:      "!=",
	TOKEN_LT:       "<",
	TOKEN_GT:       ">",
	TOKEN_LTE:      "<=",
	TOKEN_GTE:      ">=",
	TOKEN_AND:      "&&",
	TOKEN_OR:       "||",
	TOKEN_NOT:      "!",

	// Delimiters
	TOKEN_LPAREN:    "(",
	TOKEN_RPAREN:    ")",
	TOKEN_LBRACE:    "{",
	TOKEN_RBRACE:    "}",
	TOKEN_LBRACKET:  "[",
	TOKEN_RBRACKET:  "]",
	TOKEN_COMMA:     ",",
	TOKEN_SEMICOLON: ";",
	TOKEN_COLON:     ":",

	// Keywords
	TOKEN_INT_TYPE:  "int",
	TOKEN_STR_TYPE:  "str",
	TOKEN_REAL_TYPE: "real",
	TOKEN_IF:        "if",
	TOKEN_ELSE:      "else",
	TOKEN_FOR:       "for",
	TOKEN_WHILE:     "while",
	TOKEN_SWITCH:    "switch",
	TOKEN_CASE:      "case",
	TOKEN_DEFAULT:   "default",
	TOKEN_BREAK:     "break",
	TOKEN_CONTINUE:  "continue",
	TOKEN_RETURN:    "return",
	TOKEN_MES:       "mes",
	TOKEN_STEP:      "step",
	TOKEN_END_STEP:  "end_step",
	TOKEN_DEL_ME:    "del_me",
	TOKEN_DEL_US:    "del_us",
	TOKEN_DEL_ALL:   "del_all",
}

// String returns a string representation of the token type.
func (t TokenType) String() string {
	if name, ok := tokenTypeNames[t]; ok {
		return name
	}
	return "UNKNOWN"
}

// IsKeyword returns true if the token type is a keyword.
func (t TokenType) IsKeyword() bool {
	return t >= TOKEN_INT_TYPE && t <= TOKEN_DEL_ALL
}

// IsOperator returns true if the token type is an operator.
func (t TokenType) IsOperator() bool {
	return t >= TOKEN_PLUS && t <= TOKEN_NOT
}

// IsLiteral returns true if the token type is a literal.
func (t TokenType) IsLiteral() bool {
	return t >= TOKEN_IDENT && t <= TOKEN_STRING
}

// keywords maps keyword strings (lowercase) to their TokenType.
// All keywords are stored in lowercase for case-insensitive matching.
var keywords = map[string]TokenType{
	"int":      TOKEN_INT_TYPE,
	"str":      TOKEN_STR_TYPE,
	"real":     TOKEN_REAL_TYPE,
	"if":       TOKEN_IF,
	"else":     TOKEN_ELSE,
	"for":      TOKEN_FOR,
	"while":    TOKEN_WHILE,
	"switch":   TOKEN_SWITCH,
	"case":     TOKEN_CASE,
	"default":  TOKEN_DEFAULT,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"return":   TOKEN_RETURN,
	"mes":      TOKEN_MES,
	"step":     TOKEN_STEP,
	"end_step": TOKEN_END_STEP,
	"del_me":   TOKEN_DEL_ME,
	"del_us":   TOKEN_DEL_US,
	"del_all":  TOKEN_DEL_ALL,
}

// LookupIdent checks if the given identifier is a keyword.
// The lookup is case-insensitive (MES, mes, Mes all map to TOKEN_MES).
// If the identifier is a keyword, it returns the corresponding TokenType.
// Otherwise, it returns TOKEN_IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := keywords[strings.ToLower(ident)]; ok {
		return tok
	}
	return TOKEN_IDENT
}
