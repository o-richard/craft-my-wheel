package marble_test

import (
	"strings"
	"testing"

	eval "github.com/o-richard/intepreter/marble"
)

func TestEval(t *testing.T) {
	tests := []struct {
		name, input, output string
		success             bool
	}{
		{name: "var statement", input: "var foo = 1;", success: true},
		{name: "return statement", input: "var foo = 6 * 7; return foo + 2; 6;", output: "44", success: true},
		{name: "missing identifier", input: "var foo = 6; return bar; 7;", output: "identifier 'bar' not found"},
		{name: "negate prefix operator", input: "var foo = 6 * 7; !foo;", output: "false", success: true},
		{name: "minux prefix operator", input: "var foo = 6.7 + 8.3; var bar = -2; -foo;", output: "-15", success: true},
		{name: "invalid prefix operator", input: "var foo = -true", output: "unknown operator:"},
		{name: "comparison operators", input: `!((("foo" == "bar") != (" " + "bar")) == true)`, output: "false", success: true},
		{name: "string equality", input: `var foo = "bar"; foo == "bar"`, output: "true", success: true},
		{name: "equality operator", input: "[] == [];", output: "false", success: true},
		{name: "division by zero (integer)", input: "var foo = (7 * 8 + 8) / 0;", output: "invalid division by zero"},
		{name: "division by zero (float)", input: "var foo = (7 * 8 + 8.8) / 0.0;", output: "invalid division by zero"},
		{name: "nested statements", input: "var foo = 6 * 7; if (true) { if (true) { return 10; } } 6;", output: "10", success: true},
		{name: "if statements", input: "var bar = 3; var foo = if (true) { if (false) { return 10; } else { bar + 6; 5 } } foo + bar;", output: "8", success: true},
		{name: "wrong argument count (custom function)", input: "var add = func() {true};add(1, 2.0)", output: "wrong number of arguments"},
		{name: "invalid function", input: "true(1, 2.0)", output: "not a function"},
		{name: "array indexing", input: "func () {[1, 2.3, true, [false]]}()[3][-1]", output: "false", success: true},
		{name: "out of bounds array indexing", input: "[][0]", output: "out of bounds"},
		{name: "unsupported indexing", input: "true[false]", output: "unsupported index operation:"},
		{name: "built in functions", input: "var foo = push([], 1, 2.0, false, [true]); len(foo);", output: "4", success: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := eval.NewLexer([]byte(test.input))
			p := eval.NewParser(l)
			program := p.ParseProgram()
			actualErrors := p.Errors()
			if len(actualErrors) != 0 {
				t.Fatalf("unexpected errors: %v", actualErrors)
			}
			evaluated := eval.Eval(program, eval.NewEnvironment())
			var actuatlOutput string
			if evaluated != nil {
				actuatlOutput = evaluated.String()
			}
			if test.success && actuatlOutput != test.output {
				t.Fatalf("unexpected output, got=%v want=%v", actuatlOutput, test.output)
			}
			if !test.success && !strings.Contains(actuatlOutput, test.output) {
				t.Fatalf("unexpected output, got=%v want=%v", actuatlOutput, test.output)
			}
		})
	}
}
