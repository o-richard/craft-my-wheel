package marble_test

import (
	"strings"
	"testing"

	parser "github.com/o-richard/intepreter/marble"
)

func TestParseValidProgram(t *testing.T) {
	tests := []struct {
		name, input, output string
	}{
		{name: "arithmetic expression", input: "5 / 5.5 * 5 + -5 - x", output: "((((5 / 5.5) * 5) + (-5)) - x);"},
		{name: "string expression", input: `"foo" + " " + "bar"`, output: `(("foo" + " ") + "bar");`},
		{name: "boolean comparison expression", input: "(true == false) != !false;", output: "((true == false) != (!false));"},
		{name: "arithmetic comparison expression", input: "(1 - 5) < 6 == 7 > 10 <= (45 >= 22)", output: "(((1 - 5) < 6) == ((7 > 10) <= (45 >= 22)));"},
		{name: "array expression", input: `[2, 5.6, "string", [true, false], func(){x + y}, []];`, output: `[2, 5.6, "string", [true, false], func(){(x + y);}, []];`},
		{name: "if expression", input: "if (true) { 8 + 9 * 10; } else { false; }", output: "if (true) {(8 + (9 * 10));} else {false;};"},
		{name: "function call expression", input: "func (x) { } (a+b)", output: "func(x){}((a + b));"},
		{name: "array index expression", input: "array[6-7]*67", output: "((array[(6 - 7)]) * 67);"},
		{name: "var statement", input: `var foo = [9, 9.9, "bar", [true, false], 9 + 9.9];`, output: `var foo = [9, 9.9, "bar", [true, false], (9 + 9.9)];`},
		{name: "return statement", input: "var foo = 2.3; foo; 1; var y = if (true) {true}; var bar = 6.9; return foo;", output: "var foo = 2.3;foo;1;var y = if (true) {true;};var bar = 6.9;return foo;"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := parser.NewLexer([]byte(test.input))
			p := parser.NewParser(l)
			program := p.ParseProgram()
			actualErrors := p.Errors()
			if len(actualErrors) != 0 {
				t.Fatalf("unexpected errors: %v", actualErrors)
			}
			actuatlOutput := program.String()
			if actuatlOutput != test.output {
				t.Fatalf("unexpected output, got=%v want=%v", actuatlOutput, test.output)
			}
		})
	}
}

func TestParseInvalidProgram(t *testing.T) {
	tests := []struct {
		name, input, issue string
	}{
		{name: "missing identifier (var statement)", input: "var", issue: "expected next token to be "},
		{name: "missing assign (var statement)", input: "var x", issue: "expected next token to be "},
		{name: "invalid prefix (expression statement)", input: "x true = 6;", issue: "missing prefix parse function for "},
		{name: "invalid integer (integer)", input: "92233720368547758079223372036854775807;", issue: "could not parse "},
		{name: "missing right bracket (array)", input: "[1, 2, 3, 4;", issue: "expected next token to be "},
		{name: "missing right parenthesis (grouped expression)", input: "(1 + 2 * 3 / 4", issue: "expected next token to be "},
		{name: "missing left parenthesis (if expression)", input: "if", issue: "expected next token to be "},
		{name: "missing right parenthesis (if expression)", input: "if (true", issue: "expected next token to be "},
		{name: "missing left curly brace (if expression)", input: "if (true)", issue: "expected next token to be "},
		{name: "missing left curly brace (if-else expression)", input: "if (true) {} else", issue: "expected next token to be "},
		{name: "missing left parenthesis (function expression)", input: "func", issue: "expected next token to be "},
		{name: "missing right parenthesis (function expression)", input: "func(", issue: "expected next token to be "},
		{name: "missing left curly brace (function expression)", input: "func()", issue: "expected next token to be "},
		{name: "missing right bracket (array index expression)", input: "array[0", issue: "expected next token to be "},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			l := parser.NewLexer([]byte(test.input))
			p := parser.NewParser(l)
			_ = p.ParseProgram()
			actualErrors := p.Errors()
			if len(actualErrors) == 0 {
				t.Fatalf("expected an error containing the issue: %v, got=%v", test.issue, actualErrors)
			}
			if !strings.Contains(actualErrors[0], test.issue) {
				t.Fatalf("expected an error containing the issue: %v, got=%v", test.issue, actualErrors[0])
			}
		})
	}
}
