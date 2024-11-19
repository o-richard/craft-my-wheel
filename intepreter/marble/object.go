package marble

import (
	"fmt"
	"strings"
)

const (
	ARRAY = "ARRAY"
)

type object interface {
	objectType() string
	String() string
}

type objInteger struct {
	value int64
}

func (o *objInteger) objectType() string { return INTEGER }
func (o *objInteger) String() string     { return fmt.Sprintf("%v", o.value) }

type objFloat struct {
	value float64
}

func (o *objFloat) objectType() string { return FLOAT }
func (o *objFloat) String() string     { return fmt.Sprintf("%v", o.value) }

type objBoolean struct {
	value bool
}

func (o *objBoolean) objectType() string { return "BOOLEAN" }
func (o *objBoolean) String() string     { return fmt.Sprintf("%v", o.value) }

type objString struct {
	value string
}

func (o *objString) objectType() string { return STRING }
func (o *objString) String() string     { return o.value }

type objArray struct {
	elements []object
}

func (o *objArray) objectType() string { return ARRAY }

func (o *objArray) String() string {
	var output strings.Builder
	elements := make([]string, len(o.elements))
	for i := range o.elements {
		elements[i] = o.elements[i].String()
	}
	_, _ = output.WriteString("[")
	_, _ = output.WriteString(strings.Join(elements, ", "))
	_, _ = output.WriteString("]")
	return output.String()
}

type objNull struct{}

func (o *objNull) objectType() string { return "NULL" }
func (o *objNull) String() string     { return "null" }

type objReturn struct {
	value object
}

func (o *objReturn) objectType() string { return RETURN }
func (o *objReturn) String() string     { return o.value.String() }

type objError struct {
	message string
}

func (o *objError) objectType() string { return "ERROR" }
func (o *objError) String() string     { return o.message }

type objFunction struct {
	parameters []*identifier
	body       *blockStatement
	env        *environment
}

func (o *objFunction) objectType() string { return FUNCTION }

func (o *objFunction) String() string {
	var output strings.Builder
	params := make([]string, len(o.parameters))
	for i := range o.parameters {
		params[i] = o.parameters[i].String()
	}
	_, _ = output.WriteString("func(")
	_, _ = output.WriteString(strings.Join(params, ", "))
	_, _ = output.WriteString(")")
	_, _ = output.WriteString(o.body.String())
	return output.String()
}

type objBuiltin struct {
	function func(token Token, args ...object) object
}

func (o *objBuiltin) objectType() string { return "BUILTIN" }
func (o *objBuiltin) String() string     { return "built-in function" }
