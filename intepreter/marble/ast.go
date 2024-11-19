package marble

import (
	"fmt"
	"strings"
)

type node interface {
	node()
	String() string
}

type statement interface {
	node
	statementNode()
}

type expression interface {
	node
	expressionNode()
}

type program struct {
	statements []statement
}

func (p *program) node() {}

func (p *program) String() string {
	var output strings.Builder
	for i := range p.statements {
		_, _ = output.WriteString(p.statements[i].String())
	}
	return output.String()
}

type varStatement struct {
	token Token // VARIABLE token
	name  *identifier
	value expression
}

func (s *varStatement) node()          {}
func (s *varStatement) statementNode() {}

func (s *varStatement) String() string {
	var output strings.Builder
	_, _ = output.WriteString(s.token.Literal)
	_, _ = output.WriteString(" ")
	_, _ = output.WriteString(s.name.String())
	_, _ = output.WriteString(" = ")
	_, _ = output.WriteString(s.value.String())
	_, _ = output.WriteString(";")
	return output.String()
}

type returnStatement struct {
	token Token // RETURN token
	value expression
}

func (s *returnStatement) node()          {}
func (s *returnStatement) statementNode() {}

func (s *returnStatement) String() string {
	var output strings.Builder
	_, _ = output.WriteString(s.token.Literal)
	_, _ = output.WriteString(" ")
	_, _ = output.WriteString(s.value.String())
	_, _ = output.WriteString(";")
	return output.String()
}

type expressionStatement struct {
	token Token // first token of the expression
	value expression
}

func (s *expressionStatement) node()          {}
func (s *expressionStatement) statementNode() {}

func (s *expressionStatement) String() string {
	var output strings.Builder
	_, _ = output.WriteString(s.value.String())
	_, _ = output.WriteString(";")
	return output.String()
}

type blockStatement struct {
	token      Token // LBRACE token
	statements []statement
}

func (s *blockStatement) node()          {}
func (s *blockStatement) statementNode() {}

func (s *blockStatement) String() string {
	var output strings.Builder
	_, _ = output.WriteString("{")
	for i := range s.statements {
		_, _ = output.WriteString(s.statements[i].String())
	}
	_, _ = output.WriteString("}")
	return output.String()
}

type identifier struct {
	token Token // IDENTIFIER token
}

func (e *identifier) node()           {}
func (e *identifier) expressionNode() {}

func (e *identifier) String() string { return e.token.Literal }

type integerLiteral struct {
	token Token // INTEGER token
	value int64
}

func (e *integerLiteral) node()           {}
func (e *integerLiteral) expressionNode() {}

func (e *integerLiteral) String() string {
	if e == nil {
		return ""
	}
	return e.token.Literal
}

type floatLiteral struct {
	token Token // FLOAT token
	value float64
}

func (e *floatLiteral) node()           {}
func (e *floatLiteral) expressionNode() {}

func (e *floatLiteral) String() string {
	if e == nil {
		return ""
	}
	return e.token.Literal
}

type booleanLiteral struct {
	token Token // TRUE or FALSE token
	value bool
}

func (e *booleanLiteral) node()           {}
func (e *booleanLiteral) expressionNode() {}
func (e *booleanLiteral) String() string  { return e.token.Literal }

type stringLiteral struct {
	token Token // STRING token
}

func (e *stringLiteral) node()           {}
func (e *stringLiteral) expressionNode() {}
func (e *stringLiteral) String() string  { return fmt.Sprintf(`"%v"`, e.token.Literal) }

type arrayLiteral struct {
	token    Token // LBRACKET token
	elements []expression
}

func (e *arrayLiteral) node()           {}
func (e *arrayLiteral) expressionNode() {}

func (e *arrayLiteral) String() string {
	var output strings.Builder
	elements := make([]string, len(e.elements))
	for i := range e.elements {
		elements[i] = e.elements[i].String()
	}
	_, _ = output.WriteString("[")
	_, _ = output.WriteString(strings.Join(elements, ", "))
	_, _ = output.WriteString("]")
	return output.String()
}

type prefixExpression struct {
	operator Token // SUBTRACT or NEGATE token
	right    expression
}

func (e *prefixExpression) node()           {}
func (e *prefixExpression) expressionNode() {}

func (e *prefixExpression) String() string {
	var output strings.Builder
	_, _ = output.WriteString("(")
	_, _ = output.WriteString(e.operator.Literal)
	_, _ = output.WriteString(e.right.String())
	_, _ = output.WriteString(")")
	return output.String()
}

type infixExpression struct {
	operator Token
	left     expression
	right    expression
}

func (e *infixExpression) node()           {}
func (e *infixExpression) expressionNode() {}

func (e *infixExpression) String() string {
	var output strings.Builder
	_, _ = output.WriteString("(")
	_, _ = output.WriteString(e.left.String())
	_, _ = output.WriteString(" ")
	_, _ = output.WriteString(e.operator.Literal)
	_, _ = output.WriteString(" ")
	_, _ = output.WriteString(e.right.String())
	_, _ = output.WriteString(")")
	return output.String()
}

type ifExpression struct {
	token       Token // IF token
	condition   expression
	consequence *blockStatement
	alternative *blockStatement
}

func (e *ifExpression) node()           {}
func (e *ifExpression) expressionNode() {}

func (e *ifExpression) String() string {
	if e == nil {
		return ""
	}

	var output strings.Builder
	_, _ = output.WriteString(e.token.Literal)
	_, _ = output.WriteString(" (")
	_, _ = output.WriteString(e.condition.String())
	_, _ = output.WriteString(") ")
	_, _ = output.WriteString(e.consequence.String())
	if e.alternative != nil {
		_, _ = output.WriteString(" else ")
		_, _ = output.WriteString(e.alternative.String())
	}
	return output.String()
}

type functionExpression struct {
	token      Token // FUNCTION token
	parameters []*identifier
	body       *blockStatement
}

func (e *functionExpression) node()           {}
func (e *functionExpression) expressionNode() {}

func (e *functionExpression) String() string {
	if e == nil {
		return ""
	}

	var output strings.Builder
	params := make([]string, len(e.parameters))
	for i := range e.parameters {
		params[i] = e.parameters[i].String()
	}
	_, _ = output.WriteString(e.token.Literal)
	_, _ = output.WriteString("(")
	_, _ = output.WriteString(strings.Join(params, ", "))
	_, _ = output.WriteString(")")
	_, _ = output.WriteString(e.body.String())
	return output.String()
}

type callExpression struct {
	token     Token // LPAREN token
	function  expression
	arguments []expression
}

func (e *callExpression) node()           {}
func (e *callExpression) expressionNode() {}

func (e *callExpression) String() string {
	var output strings.Builder
	args := make([]string, len(e.arguments))
	for i := range e.arguments {
		args[i] = e.arguments[i].String()
	}
	_, _ = output.WriteString(e.function.String())
	_, _ = output.WriteString("(")
	_, _ = output.WriteString(strings.Join(args, ", "))
	_, _ = output.WriteString(")")
	return output.String()
}

type indexExpression struct {
	token Token // LBRACKET token
	left  expression
	index expression
}

func (e *indexExpression) node()           {}
func (e *indexExpression) expressionNode() {}

func (e *indexExpression) String() string {
	if e == nil {
		return ""
	}

	var output strings.Builder
	_, _ = output.WriteString("(")
	_, _ = output.WriteString(e.left.String())
	_, _ = output.WriteString("[")
	_, _ = output.WriteString(e.index.String())
	_, _ = output.WriteString("])")
	return output.String()
}
