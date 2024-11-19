package marble

const (
	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"

	IDENTIFIER = "IDENTIFIER"
	STRING     = "STRING"
	FLOAT      = "FLOAT"
	INTEGER    = "INTEGER"

	ASSIGN   = "="
	ADD      = "+"
	SUBTRACT = "-"
	MULTIPLY = "*"
	DIVIDE   = "/"
	NEGATE   = "!"

	LT    = "<"
	GT    = ">"
	LTE   = "<="
	GTE   = ">="
	EQ    = "=="
	NOTEQ = "!="

	COMMA     = ","
	SEMICOLON = ";"

	LPAREN   = "("
	RPAREN   = ")"
	LBRACE   = "{"
	RBRACE   = "}"
	LBRACKET = "["
	RBRACKET = "]"

	FUNCTION = "FUNCTION"
	VARIABLE = "VARIABLE"
	TRUE     = "TRUE"
	FALSE    = "FALSE"
	IF       = "IF"
	ELSE     = "ELSE"
	RETURN   = "RETURN"
)

type TokenType string

type Token struct {
	Type       TokenType
	Literal    string
	LineNumber int
	ColNumber  int
}
