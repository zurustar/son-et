package token

type TokenType string

type Token struct {
	Type    TokenType
	Literal string
	Line    int
}

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	// Identifiers + Literals
	IDENT  = "IDENT"  // MovePic, IMG01
	NUMBER = "NUMBER" // 123, 0xff
	STRING = "STRING" // "abc"

	// Operators and Delimiters
	ASSIGN    = "="
	PLUS      = "+"
	MINUS     = "-"
	ASTERISK  = "*"
	SLASH     = "/"
	COMMA     = ","
	SEMICOLON = ";"
	COLON     = ":"
	LPAREN    = "("
	RPAREN    = ")"
	LBRACE    = "{"
	RBRACE    = "}"
	LBRACKET  = "["
	RBRACKET  = "]"

	BANG   = "!"
	EQ     = "=="
	NOT_EQ = "!="
	LT     = "<"
	GT     = ">"
	LTE    = "<="
	GTE    = ">="
	AND    = "&&"
	OR     = "||"

	// Keywords
	MES      = "MES"
	STEP     = "STEP"
	INT      = "INT"
	STR      = "STR"
	IF       = "IF"
	ELSE     = "ELSE"
	FOR      = "FOR"
	WHILE    = "WHILE"
	DO       = "DO"
	SWITCH   = "SWITCH"
	CASE     = "CASE"
	DEFAULT  = "DEFAULT"
	BREAK    = "BREAK"
	CONTINUE = "CONTINUE"
)

func LookupIdent(ident string) TokenType {
	switch ident {
	case "mes":
		return MES
	case "step":
		return STEP
	case "int":
		return INT
	case "str":
		return STR
	case "if":
		return IF
	case "else":
		return ELSE
	case "for":
		return FOR
	case "while":
		return WHILE
	case "do":
		return DO
	case "switch":
		return SWITCH
	case "case":
		return CASE
	case "default":
		return DEFAULT
	case "break":
		return BREAK
	case "continue":
		return CONTINUE
	}
	return IDENT
}
