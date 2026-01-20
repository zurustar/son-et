package token

// TokenType represents the type of a token.
type TokenType int

const (
	// Special tokens
	ILLEGAL TokenType = iota
	EOF
	COMMENT

	// Literals
	IDENT      // identifier
	INT_LIT    // integer literal
	FLOAT_LIT  // float literal
	STRING_LIT // string literal

	// Operators
	ASSIGN // =
	PLUS   // +
	MINUS  // -
	MULT   // *
	DIV    // /
	MOD    // %

	// Comparison
	EQ  // ==
	NEQ // !=
	LT  // <
	LTE // <=
	GT  // >
	GTE // >=

	// Logical
	AND // &&
	OR  // ||
	NOT // !

	// Delimiters
	LPAREN    // (
	RPAREN    // )
	LBRACE    // {
	RBRACE    // }
	LBRACKET  // [
	RBRACKET  // ]
	COMMA     // ,
	SEMICOLON // ;

	// Keywords
	IF
	ELSE
	FOR
	WHILE
	SWITCH
	CASE
	DEFAULT
	BREAK
	CONTINUE
	RETURN
	FUNCTION
	MES
	STEP
	TIME
	MIDI_TIME
	MIDI_END
	KEY
	CLICK
	RBDOWN
	RBDBLCLK
	USER

	// Type keywords
	INT
	STRING
	FLOAT_TYPE

	// Special keywords
	END_STEP
	DEL_ME
	DEL_US
	DEL_ALL
)

var keywords = map[string]TokenType{
	"if":        IF,
	"else":      ELSE,
	"for":       FOR,
	"while":     WHILE,
	"switch":    SWITCH,
	"case":      CASE,
	"default":   DEFAULT,
	"break":     BREAK,
	"continue":  CONTINUE,
	"return":    RETURN,
	"function":  FUNCTION,
	"mes":       MES,
	"step":      STEP,
	"time":      TIME,
	"midi_time": MIDI_TIME,
	"midi_end":  MIDI_END,
	"key":       KEY,
	"click":     CLICK,
	"rbdown":    RBDOWN,
	"rbdblclk":  RBDBLCLK,
	"user":      USER,
	"int":       INT,
	"string":    STRING,
	"float":     FLOAT_TYPE,
	"end_step":  END_STEP,
	"del_me":    DEL_ME,
	"del_us":    DEL_US,
	"del_all":   DEL_ALL,
}

// Token represents a lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

// LookupIdent checks if an identifier is a keyword (case-insensitive).
func LookupIdent(ident string) TokenType {
	// Convert to lowercase for case-insensitive matching
	lower := ""
	for _, ch := range ident {
		if ch >= 'A' && ch <= 'Z' {
			lower += string(ch + 32)
		} else {
			lower += string(ch)
		}
	}

	if tok, ok := keywords[lower]; ok {
		return tok
	}
	return IDENT
}

// String returns a string representation of the token type.
func (t TokenType) String() string {
	switch t {
	case ILLEGAL:
		return "ILLEGAL"
	case EOF:
		return "EOF"
	case COMMENT:
		return "COMMENT"
	case IDENT:
		return "IDENT"
	case INT_LIT:
		return "INT"
	case FLOAT_LIT:
		return "FLOAT"
	case STRING_LIT:
		return "STRING"
	case ASSIGN:
		return "="
	case PLUS:
		return "+"
	case MINUS:
		return "-"
	case MULT:
		return "*"
	case DIV:
		return "/"
	case MOD:
		return "%"
	case EQ:
		return "=="
	case NEQ:
		return "!="
	case LT:
		return "<"
	case LTE:
		return "<="
	case GT:
		return ">"
	case GTE:
		return ">="
	case AND:
		return "&&"
	case OR:
		return "||"
	case NOT:
		return "!"
	case LPAREN:
		return "("
	case RPAREN:
		return ")"
	case LBRACE:
		return "{"
	case RBRACE:
		return "}"
	case LBRACKET:
		return "["
	case RBRACKET:
		return "]"
	case COMMA:
		return ","
	case SEMICOLON:
		return ";"
	case IF:
		return "IF"
	case ELSE:
		return "ELSE"
	case FOR:
		return "FOR"
	case WHILE:
		return "WHILE"
	case SWITCH:
		return "SWITCH"
	case CASE:
		return "CASE"
	case DEFAULT:
		return "DEFAULT"
	case BREAK:
		return "BREAK"
	case CONTINUE:
		return "CONTINUE"
	case RETURN:
		return "RETURN"
	case FUNCTION:
		return "FUNCTION"
	case MES:
		return "MES"
	case STEP:
		return "STEP"
	case TIME:
		return "TIME"
	case MIDI_TIME:
		return "MIDI_TIME"
	case MIDI_END:
		return "MIDI_END"
	case KEY:
		return "KEY"
	case CLICK:
		return "CLICK"
	case RBDOWN:
		return "RBDOWN"
	case RBDBLCLK:
		return "RBDBLCLK"
	case USER:
		return "USER"
	case INT:
		return "INT_TYPE"
	case STRING:
		return "STRING_TYPE"
	case FLOAT_TYPE:
		return "FLOAT_TYPE"
	case END_STEP:
		return "END_STEP"
	case DEL_ME:
		return "DEL_ME"
	case DEL_US:
		return "DEL_US"
	case DEL_ALL:
		return "DEL_ALL"
	default:
		return "UNKNOWN"
	}
}
