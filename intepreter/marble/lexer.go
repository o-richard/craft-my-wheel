package marble

type lexer struct {
	input []byte

	currentByte byte

	currentIndex int
	nextIndex    int

	currentLineNumber int
	currentColNumber  int
}

func NewLexer(input []byte) *lexer {
	l := &lexer{input: make([]byte, len(input))}
	copy(l.input, input)
	l.readByte()
	return l
}

func (l *lexer) NextToken() Token {
	l.skipWhitespaceAndComments()

	var tok Token
	switch l.currentByte {
	case '=':
		tok = l.readOperator(l.currentByte, ASSIGN, EQ)
	case '+':
		tok = l.newToken(ADD, "+")
	case '-':
		tok = l.newToken(SUBTRACT, "-")
	case '*':
		tok = l.newToken(MULTIPLY, "*")
	case '/':
		tok = l.newToken(DIVIDE, "/")
	case '!':
		tok = l.readOperator(l.currentByte, NEGATE, NOTEQ)
	case '<':
		tok = l.readOperator(l.currentByte, LT, LTE)
	case '>':
		tok = l.readOperator(l.currentByte, GT, GTE)
	case ',':
		tok = l.newToken(COMMA, ",")
	case ';':
		tok = l.newToken(SEMICOLON, ";")
	case '(':
		tok = l.newToken(LPAREN, "(")
	case ')':
		tok = l.newToken(RPAREN, ")")
	case '{':
		tok = l.newToken(LBRACE, "{")
	case '}':
		tok = l.newToken(RBRACE, "}")
	case '[':
		tok = l.newToken(LBRACKET, "[")
	case ']':
		tok = l.newToken(RBRACKET, "]")
	case '"':
		tok = l.readString()
	case 0:
		tok = l.newToken(EOF, "")
	default:
		if validIdentifierByte(l.currentByte) {
			return l.readIdentifier()
		}
		if validNumberDigit(l.currentByte) {
			return l.readNumber()
		}
		tok = l.newToken(ILLEGAL, string(l.currentByte))
	}

	l.readByte()
	return tok
}

func (l *lexer) readByte() {
	if l.nextIndex >= len(l.input) {
		l.currentByte = 0
	} else {
		l.currentByte = l.input[l.nextIndex]
	}
	l.currentIndex = l.nextIndex
	l.nextIndex++

	if l.currentByte == '\n' {
		l.currentLineNumber++
		l.currentColNumber = -1
	}
	l.currentColNumber++
}

func (l *lexer) skipWhitespaceAndComments() {
	for {
		var modified bool

		for l.currentByte == ' ' || l.currentByte == '\t' || l.currentByte == '\n' || l.currentByte == '\r' {
			modified = true
			l.readByte()
		}
		if l.currentByte == '/' && l.peekNextByte() == '/' {
			modified = true
			for {
				l.readByte()
				if l.currentByte == 0 || l.currentByte == '\n' {
					break
				}
			}
		}

		if !modified {
			break
		}
	}
}

func (l *lexer) peekNextByte() byte {
	if l.nextIndex >= len(l.input) {
		return 0
	}
	return l.input[l.nextIndex]
}

func (l *lexer) newToken(tt TokenType, literal string) Token {
	return Token{Type: tt, Literal: literal, LineNumber: l.currentLineNumber + 1, ColNumber: l.currentColNumber}
}

func (l *lexer) readOperator(previous byte, single, combined TokenType) Token {
	if l.peekNextByte() == '=' {
		l.readByte()
		return Token{Type: combined, Literal: string([]byte{previous, '='}), LineNumber: l.currentLineNumber + 1, ColNumber: l.currentColNumber - 1}
	}
	return l.newToken(single, string(previous))
}

func (l *lexer) readString() Token {
	currentIndex := l.currentIndex + 1
	lineNumber := l.currentLineNumber + 1
	colNumber := l.currentColNumber

	for {
		l.readByte()
		if l.currentByte == 0 || l.currentByte == '"' {
			break
		}
	}
	return Token{Type: STRING, Literal: string(l.input[currentIndex:l.currentIndex]), LineNumber: lineNumber, ColNumber: colNumber}
}

func validIdentifierByte(char byte) bool {
	return (('a' <= char && char <= 'z') || ('A' <= char && char <= 'Z') || char == '_')
}

func (l *lexer) readIdentifier() Token {
	currentIndex := l.currentIndex
	lineNumber := l.currentLineNumber + 1
	colNumber := l.currentColNumber

	for validIdentifierByte(l.currentByte) {
		l.readByte()
	}
	literal := string(l.input[currentIndex:l.currentIndex])

	var tokentype TokenType
	switch literal {
	case "func":
		tokentype = FUNCTION
	case "var":
		tokentype = VARIABLE
	case "true":
		tokentype = TRUE
	case "false":
		tokentype = FALSE
	case "if":
		tokentype = IF
	case "else":
		tokentype = ELSE
	case "return":
		tokentype = RETURN
	default:
		tokentype = IDENTIFIER
	}
	return Token{Type: tokentype, Literal: literal, LineNumber: lineNumber, ColNumber: colNumber}
}

func validNumberDigit(char byte) bool {
	return '0' <= char && char <= '9'
}

func (l *lexer) readNumber() Token {
	currentIndex := l.currentIndex
	lineNumber := l.currentLineNumber + 1
	colNumber := l.currentColNumber

	var periodPresence bool
	for validNumberDigit(l.currentByte) {
		l.readByte()

		if l.currentByte == '.' && !periodPresence && validNumberDigit(l.peekNextByte()) {
			periodPresence = true
			l.readByte()
		}
	}

	var tokentype TokenType
	if periodPresence {
		tokentype = FLOAT
	} else {
		tokentype = INTEGER
	}
	return Token{Type: tokentype, Literal: string(l.input[currentIndex:l.currentIndex]), LineNumber: lineNumber, ColNumber: colNumber}
}
