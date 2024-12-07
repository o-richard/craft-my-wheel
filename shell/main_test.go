package main

import (
	"fmt"
	"testing"
)

func TestParseCommandArg(t *testing.T) {
	tests := []struct {
		input, expected string
	}{
		{input: ``},
		{input: `asd   ghj  'ssd   ddfff" ss'`, expected: `asd ghj ssd   ddfff" ss`},
		{input: `\" "gh fh" abc  def "ssdd's"`, expected: `" gh fh abc def ssdd's`},
		{input: `'ac\c\cd\n\"' \" \\`, expected: `ac\c\cd\n\" " \`},
		{input: `"ab  c" asd`, expected: `ab  c asd`},
		{input: `"hello\"insidequotes"script\"`, expected: `hello"insidequotesscript"`},
		{input: `"mixed\"quote'shell'\\"`, expected: `mixed"quote'shell'\`},
		{input: `"\\"`, expected: `\`},
		{input: `example\ \ \ \ \ \ world`, expected: "example      world"},
	}
	for i, test := range tests {
		t.Run(fmt.Sprintf("Test %v", i), func(t *testing.T) {
			_, actualOutput := parseArg([]byte(test.input))
			if actualOutput != test.expected {
				t.Errorf("input: `%v`, got = %v, want = %v", test.input, actualOutput, test.expected)
			}
		})
	}
}
