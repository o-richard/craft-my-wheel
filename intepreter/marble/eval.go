package marble

import "fmt"

var (
	objectNull  = &objNull{}
	objectTrue  = &objBoolean{value: true}
	objectFalse = &objBoolean{value: false}
)

var (
	builtins = map[string]*objBuiltin{
		"len": {
			function: func(token Token, args ...object) object {
				if len(args) != 1 {
					return &objError{message: fmt.Sprintf("line %v col %v: wrong number of arguments", token.LineNumber, token.ColNumber)}
				}
				switch arg := args[0].(type) {
				case *objArray:
					return &objInteger{value: int64(len(arg.elements))}
				case *objString:
					return &objInteger{value: int64(len(arg.value))}
				}
				return &objError{message: fmt.Sprintf("line %v col %v: invalid argument type: %v", token.LineNumber, token.ColNumber, args[0].objectType())}
			},
		},
		"push": {
			function: func(token Token, args ...object) object {
				if maxArgs := 2; len(args) < maxArgs {
					return &objError{message: fmt.Sprintf("line %v col %v: wrong number of arguments", token.LineNumber, token.ColNumber)}
				}
				switch arg := args[0].(type) {
				case *objArray:
					slice := append(make([]object, 0, len(arg.elements)+len(args)-1), arg.elements...)
					return &objArray{elements: append(slice, args[1:]...)}
				}
				return &objError{message: fmt.Sprintf("line %v col %v: invalid argument type: %v", token.LineNumber, token.ColNumber, args[0].objectType())}
			},
		},
		"print": {
			function: func(token Token, args ...object) object {
				for i := range args {
					fmt.Println(args[i].String())
				}
				return objectNull
			},
		},
	}
)

func Eval(node node, env *environment) object {
	switch node := node.(type) {
	case *program:
		return evalProgram(node, env)
	case *varStatement:
		value := Eval(node.value, env)
		if _, ok := value.(*objError); ok {
			return value
		}
		env.set(node.name.token.Literal, value)
	case *returnStatement:
		value := Eval(node.value, env)
		if _, ok := value.(*objError); ok {
			return value
		}
		return &objReturn{value: value}
	case *expressionStatement:
		return Eval(node.value, env)
	case *blockStatement:
		return evalBlockStatement(node, env)
	case *identifier:
		return evalIdentifier(node.token, env)
	case *integerLiteral:
		return &objInteger{value: node.value}
	case *floatLiteral:
		return &objFloat{value: node.value}
	case *booleanLiteral:
		return evalBoolean(node.value)
	case *stringLiteral:
		return &objString{value: node.token.Literal}
	case *arrayLiteral:
		elements, ok := evalExpressions(node.elements, env)
		if !ok {
			return elements[0]
		}
		return &objArray{elements: elements}
	case *prefixExpression:
		right := Eval(node.right, env)
		if _, ok := right.(*objError); ok {
			return right
		}
		return evalPrefixExpression(node.operator, right)
	case *infixExpression:
		left := Eval(node.left, env)
		if _, ok := left.(*objError); ok {
			return left
		}
		right := Eval(node.right, env)
		if _, ok := right.(*objError); ok {
			return right
		}
		return evalInfixExpression(node.operator, left, right)
	case *ifExpression:
		return evalIfExpression(node, env)
	case *functionExpression:
		return &objFunction{body: node.body, parameters: node.parameters, env: env}
	case *callExpression:
		function := Eval(node.function, env)
		if _, ok := function.(*objError); ok {
			return function
		}
		args, ok := evalExpressions(node.arguments, env)
		if !ok {
			return args[0]
		}
		return applyFunction(node.token, function, args)
	case *indexExpression:
		left := Eval(node.left, env)
		if _, ok := left.(*objError); ok {
			return left
		}
		index := Eval(node.index, env)
		if _, ok := index.(*objError); ok {
			return index
		}
		return evalIndexExpression(node.token, left, index)
	}
	return nil
}

func evalProgram(p *program, env *environment) object {
	var result object
	for i := range p.statements {
		result = Eval(p.statements[i], env)

		switch result := result.(type) {
		case *objReturn:
			return result.value
		case *objError:
			return result
		}
	}
	return result
}

func evalBlockStatement(b *blockStatement, env *environment) object {
	var result object
	for i := range b.statements {
		result = Eval(b.statements[i], env)

		switch result := result.(type) {
		case *objReturn:
			return result
		case *objError:
			return result
		}
	}
	return result
}

func evalIdentifier(token Token, env *environment) object {
	value, ok := env.get(token.Literal)
	if ok {
		return value
	}
	value, ok = builtins[token.Literal]
	if ok {
		return value
	}
	return &objError{message: fmt.Sprintf("line %v col %v: identifier '%v' not found", token.LineNumber, token.ColNumber, token.Literal)}
}

func evalBoolean(b bool) *objBoolean {
	if b {
		return objectTrue
	}
	return objectFalse
}

func evalPrefixExpression(operator Token, right object) object {
	switch operator.Literal {
	case "!":
		if right == objectFalse || right == objectNull {
			return objectTrue
		}
		return objectFalse
	case "-":
		switch right := right.(type) {
		case *objInteger:
			return &objInteger{value: -right.value}
		case *objFloat:
			return &objFloat{value: -right.value}
		}
	}
	return &objError{message: fmt.Sprintf("line %v col %v: unknown operator: %v%v", operator.LineNumber, operator.ColNumber, operator.Literal, right.objectType())}
}

func evalInfixExpression(operator Token, left, right object) object {
	switch {
	case left.objectType() == INTEGER && right.objectType() == INTEGER:
		return evalIntegerInfixExpression(operator, left, right)
	case (left.objectType() == FLOAT || left.objectType() == INTEGER) && (right.objectType() == FLOAT || right.objectType() == INTEGER):
		return evalFloatInfixExpression(operator, left, right)
	case left.objectType() == STRING && right.objectType() == STRING:
		return evalStringInfixExpression(operator, left, right)
	case operator.Literal == "==":
		return evalBoolean(left == right)
	case operator.Literal == "!=":
		return evalBoolean(left != right)
	}
	return &objError{message: fmt.Sprintf("line %v col %v: unknown operator: %v %v %v", operator.LineNumber, operator.ColNumber, left.objectType(), operator.Literal, right.objectType())}
}

func evalIntegerInfixExpression(operator Token, left, right object) object {
	leftValue := left.(*objInteger).value
	rightValue := right.(*objInteger).value

	switch operator.Literal {
	case "+":
		return &objInteger{value: leftValue + rightValue}
	case "-":
		return &objInteger{value: leftValue - rightValue}
	case "*":
		return &objInteger{value: leftValue * rightValue}
	case "/":
		if rightValue == 0 {
			return &objError{message: fmt.Sprintf("line %v col %v: invalid division by zero", operator.LineNumber, operator.ColNumber)}
		}
		return &objInteger{value: leftValue / rightValue}
	case "<":
		return evalBoolean(leftValue < rightValue)
	case ">":
		return evalBoolean(leftValue > rightValue)
	case ">=":
		return evalBoolean(leftValue >= rightValue)
	case "<=":
		return evalBoolean(leftValue <= rightValue)
	case "==":
		return evalBoolean(leftValue == rightValue)
	case "!=":
		return evalBoolean(leftValue != rightValue)
	}
	return &objError{message: fmt.Sprintf("line %v col %v: unknown operator: %v %v %v", operator.LineNumber, operator.ColNumber, left.objectType(), operator.Literal, right.objectType())}
}

func evalFloatInfixExpression(operator Token, left, right object) object {
	var leftValue float64
	switch left := left.(type) {
	case *objInteger:
		leftValue = float64(left.value)
	case *objFloat:
		leftValue = left.value
	}
	var rightValue float64
	switch right := right.(type) {
	case *objInteger:
		rightValue = float64(right.value)
	case *objFloat:
		rightValue = right.value
	}

	switch operator.Literal {
	case "+":
		return &objFloat{value: leftValue + rightValue}
	case "-":
		return &objFloat{value: leftValue - rightValue}
	case "*":
		return &objFloat{value: leftValue * rightValue}
	case "/":
		if rightValue == 0 {
			return &objError{message: fmt.Sprintf("line %v col %v: invalid division by zero", operator.LineNumber, operator.ColNumber)}
		}
		return &objFloat{value: leftValue / rightValue}
	case "<":
		return evalBoolean(leftValue < rightValue)
	case ">":
		return evalBoolean(leftValue > rightValue)
	case ">=":
		return evalBoolean(leftValue >= rightValue)
	case "<=":
		return evalBoolean(leftValue <= rightValue)
	case "==":
		return evalBoolean(leftValue == rightValue)
	case "!=":
		return evalBoolean(leftValue != rightValue)
	}
	return &objError{message: fmt.Sprintf("line %v col %v: unknown operator: %v %v %v", operator.LineNumber, operator.ColNumber, left.objectType(), operator.Literal, right.objectType())}
}

func evalStringInfixExpression(operator Token, left, right object) object {
	leftValue := left.(*objString).value
	rightValue := right.(*objString).value

	switch operator.Literal {
	case "+":
		return &objString{value: leftValue + rightValue}
	case "==":
		return evalBoolean(leftValue == rightValue)
	case "!=":
		return evalBoolean(leftValue != rightValue)
	}
	return &objError{message: fmt.Sprintf("line %v col %v: unknown operator: %v %v %v", operator.LineNumber, operator.ColNumber, left.objectType(), operator.Literal, right.objectType())}
}

func evalIfExpression(e *ifExpression, env *environment) object {
	condition := Eval(e.condition, env)
	if _, ok := condition.(*objError); ok {
		return condition
	}
	if condition != objectNull && condition != objectFalse {
		return Eval(e.consequence, env)
	} else if e.alternative != nil {
		return Eval(e.alternative, env)
	}
	return objectNull
}

func evalExpressions(expressions []expression, env *environment) ([]object, bool) {
	result := make([]object, 0, len(expressions))
	for i := range expressions {
		evaluated := Eval(expressions[i], env)
		if _, ok := evaluated.(*objError); ok {
			return []object{evaluated}, false
		}
		result = append(result, evaluated)
	}
	return result, true
}

func applyFunction(token Token, o object, args []object) object {
	switch function := o.(type) {
	case *objFunction:
		if len(args) != len(function.parameters) {
			return &objError{message: fmt.Sprintf("line %v col %v: wrong number of arguments", token.LineNumber, token.ColNumber)}
		}
		env := newEnclosedEnvironment(function.env)
		for i := range function.parameters {
			env.set(function.parameters[i].token.Literal, args[i])
		}
		evaluated := Eval(function.body, env)
		if returnValue, ok := evaluated.(*objReturn); ok {
			return returnValue.value
		}
		return evaluated
	case *objBuiltin:
		return function.function(token, args...)
	}
	return &objError{message: fmt.Sprintf("line %v col %v: '%v' is not a function", token.LineNumber, token.ColNumber, o.objectType())}
}

func evalIndexExpression(token Token, left, right object) object {
	if left.objectType() == ARRAY && right.objectType() == INTEGER {
		return evalArrayIndexExpression(token, left, right)
	}
	return &objError{message: fmt.Sprintf("line %v col %v: unsupported index operation: %v", token.LineNumber, token.ColNumber, left.objectType())}
}

func evalArrayIndexExpression(token Token, left, right object) object {
	elements := left.(*objArray).elements
	index := right.(*objInteger).value
	count := int64(len(elements))
	if index < 0 {
		index = count + index
	}
	if index < 0 || index >= count {
		return &objError{message: fmt.Sprintf("line %v col %v: index '%v' is out of bounds", token.LineNumber, token.ColNumber, index)}
	}
	return elements[index]
}
