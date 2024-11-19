package marble

import (
	"fmt"
	"strconv"
)

const (
	_ = iota
	lowest
	equals          // ==, !=
	less_greater    // <, >, >=, <=
	add_subtract    // +, -
	multiply_divide // *, /
	prefix          // -x, !x
	call            // myFunction(x)
	index           // array[index]
)

type parser struct {
	l *lexer

	issues []string

	next    Token
	current Token
}

func NewParser(l *lexer) *parser {
	p := &parser{l: l}

	p.nextToken()
	p.nextToken()
	return p
}

func (p *parser) Errors() []string {
	return p.issues
}

func (p *parser) ParseProgram() *program {
	program := &program{}

	for p.current.Type != EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			program.statements = append(program.statements, stmt)
		}
		p.nextToken()
	}

	return program
}

func (p *parser) nextToken() {
	p.current = p.next
	p.next = p.l.NextToken()
}

func (p *parser) expectToken(t TokenType) bool {
	if p.next.Type == t {
		p.nextToken()
		return true
	}
	p.issues = append(p.issues, fmt.Sprintf("line %v column %v: expected next token to be %v, got %v instead", p.next.LineNumber, p.next.ColNumber, t, p.next.Type))
	return false
}

func (p *parser) parseStatement() statement {
	switch p.current.Type {
	case VARIABLE:
		if stmt := p.parseVarStatement(); stmt != nil {
			return stmt
		}
		return nil
	case RETURN:
		return p.parseReturnStatement()
	default:
		return p.parseExpressionStatement()
	}
}

func (p *parser) parseVarStatement() *varStatement {
	stmt := &varStatement{token: p.current}
	if !p.expectToken(IDENTIFIER) {
		return nil
	}
	stmt.name = &identifier{token: p.current}
	if !p.expectToken(ASSIGN) {
		return nil
	}
	p.nextToken()
	stmt.value = p.parseExpression(lowest)
	if p.next.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *parser) parseReturnStatement() *returnStatement {
	stmt := &returnStatement{token: p.current}
	p.nextToken()
	stmt.value = p.parseExpression(lowest)
	if p.next.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func (p *parser) parseExpressionStatement() *expressionStatement {
	stmt := &expressionStatement{token: p.current}
	stmt.value = p.parseExpression(lowest)
	if p.next.Type == SEMICOLON {
		p.nextToken()
	}
	return stmt
}

func tokenPrecedence(t TokenType) int {
	switch t {
	case EQ, NOTEQ:
		return equals
	case LT, GT, LTE, GTE:
		return less_greater
	case ADD, SUBTRACT:
		return add_subtract
	case MULTIPLY, DIVIDE:
		return multiply_divide
	case LPAREN:
		return call
	case LBRACKET:
		return index
	default:
		return lowest
	}
}

func (p *parser) parseExpression(precedence int) expression {
	var left expression
	switch p.current.Type {
	case IDENTIFIER:
		left = &identifier{token: p.current}
	case INTEGER:
		left = p.parseIntegerLiteral()
	case FLOAT:
		left = p.parseFloatLiteral()
	case TRUE, FALSE:
		left = &booleanLiteral{token: p.current, value: p.current.Type == TRUE}
	case STRING:
		left = &stringLiteral{token: p.current}
	case LBRACKET:
		left = p.parseArrayLiteral()
	case SUBTRACT, NEGATE:
		left = p.parsePrefixExpression()
	case LPAREN:
		left = p.parseGroupedExpression()
	case IF:
		left = p.parseIfExpression()
	case FUNCTION:
		left = p.parseFunctionExpression()
	default:
		p.issues = append(p.issues, fmt.Sprintf("missing prefix parse function for %v", p.current.Type))
		return nil
	}

	for p.next.Type != SEMICOLON && precedence < tokenPrecedence(p.next.Type) {
		switch p.next.Type {
		case ADD, SUBTRACT, MULTIPLY, DIVIDE, EQ, NOTEQ, LT, LTE, GT, GTE:
			p.nextToken()
			left = p.parseInfixExpression(left)
		case LPAREN:
			p.nextToken()
			left = p.parseCallExpression(left)
		case LBRACKET:
			p.nextToken()
			left = p.parseIndexExpression(left)
		default:
			return left
		}
	}
	return left
}

func (p *parser) parseIntegerLiteral() *integerLiteral {
	value, err := strconv.ParseInt(p.current.Literal, 0, 64)
	if err != nil {
		p.issues = append(p.issues, fmt.Sprintf("line %v column %v: could not parse %v as integer", p.current.LineNumber, p.current.ColNumber, p.current.Literal))
		return nil
	}
	return &integerLiteral{token: p.current, value: value}
}

func (p *parser) parseFloatLiteral() *floatLiteral {
	value, err := strconv.ParseFloat(p.current.Literal, 64)
	if err != nil {
		p.issues = append(p.issues, fmt.Sprintf("line %v column %v: could not parse %v as float", p.current.LineNumber, p.current.ColNumber, p.current.Literal))
		return nil
	}
	return &floatLiteral{token: p.current, value: value}
}

func (p *parser) parseArrayLiteral() *arrayLiteral {
	e := &arrayLiteral{token: p.current}
	e.elements = p.parseExpressionList(RBRACKET)
	return e
}

func (p *parser) parseExpressionList(end TokenType) []expression {
	expressions := make([]expression, 0)
	if p.next.Type == end {
		p.nextToken()
		return expressions
	}
	p.nextToken()
	expressions = append(expressions, p.parseExpression(lowest))
	for p.next.Type == COMMA {
		p.nextToken()
		p.nextToken()
		expressions = append(expressions, p.parseExpression(lowest))
	}
	if !p.expectToken(end) {
		return nil
	}
	return expressions
}

func (p *parser) parsePrefixExpression() *prefixExpression {
	e := &prefixExpression{operator: p.current}
	p.nextToken()
	e.right = p.parseExpression(prefix)
	return e
}

func (p *parser) parseInfixExpression(left expression) *infixExpression {
	e := &infixExpression{operator: p.current, left: left}
	precedence := tokenPrecedence(p.current.Type)
	p.nextToken()
	e.right = p.parseExpression(precedence)
	return e
}

func (p *parser) parseGroupedExpression() expression {
	p.nextToken()
	e := p.parseExpression(lowest)
	if !p.expectToken(RPAREN) {
		return nil
	}
	return e
}

func (p *parser) parseBlockStatement() *blockStatement {
	b := &blockStatement{token: p.current}
	p.nextToken()
	for p.current.Type != RBRACE && p.current.Type != EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			b.statements = append(b.statements, stmt)
		}
		p.nextToken()
	}
	return b
}

func (p *parser) parseIfExpression() *ifExpression {
	e := &ifExpression{token: p.current}
	if !p.expectToken(LPAREN) {
		return nil
	}
	p.nextToken()
	e.condition = p.parseExpression(lowest)
	if !p.expectToken(RPAREN) {
		return nil
	}
	if !p.expectToken(LBRACE) {
		return nil
	}
	e.consequence = p.parseBlockStatement()
	if p.next.Type == ELSE {
		p.nextToken()
		if !p.expectToken(LBRACE) {
			return nil
		}
		e.alternative = p.parseBlockStatement()
	}
	return e
}

func (p *parser) parseFunctionExpression() *functionExpression {
	e := &functionExpression{token: p.current}
	if !p.expectToken(LPAREN) {
		return nil
	}
	e.parameters = p.parseFunctionParameters()
	if !p.expectToken(LBRACE) {
		return nil
	}
	e.body = p.parseBlockStatement()
	return e
}

func (p *parser) parseFunctionParameters() []*identifier {
	identifiers := make([]*identifier, 0)
	if p.next.Type == RPAREN {
		p.nextToken()
		return identifiers
	}
	p.nextToken()
	identifiers = append(identifiers, &identifier{token: p.current})
	for p.next.Type == COMMA {
		p.nextToken()
		p.nextToken()
		identifiers = append(identifiers, &identifier{token: p.current})
	}
	if !p.expectToken(RPAREN) {
		return nil
	}
	return identifiers
}

func (p *parser) parseCallExpression(function expression) *callExpression {
	e := &callExpression{token: p.current, function: function}
	e.arguments = p.parseExpressionList(RPAREN)
	return e
}

func (p *parser) parseIndexExpression(left expression) *indexExpression {
	e := &indexExpression{token: p.current, left: left}
	p.nextToken()
	e.index = p.parseExpression(lowest)
	if !p.expectToken(RBRACKET) {
		return nil
	}
	return e
}
