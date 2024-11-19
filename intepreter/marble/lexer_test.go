package marble_test

import (
	"fmt"
	"testing"

	lexer "github.com/o-richard/intepreter/marble"
)

func TestNextToken(t *testing.T) {
	input := `var five = 5; // integer variable
var five_point_five = 5.5; // float variable
var foo_bar = "foo bar"; // string variable

var add = func(x, y) {
	var z = x + y;
	return (z - 2) * 2 / 2;
};
var result = add(five, five_point_five); // blank comment

// blank comment
var negate = !true;
5 <= 5;
5 >= 5;
5 != 5;
5 == 5;

5 < 10 > 5;
if (5 < 10) {
	return true;
} else {
	return false;
}
[1, 2, 3];
55;`
	expected := []lexer.Token{
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 1, ColNumber: 1},
		{Type: lexer.IDENTIFIER, Literal: "five", LineNumber: 1, ColNumber: 5},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 1, ColNumber: 10},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 1, ColNumber: 12},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 1, ColNumber: 13},
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 2, ColNumber: 1},
		{Type: lexer.IDENTIFIER, Literal: "five_point_five", LineNumber: 2, ColNumber: 5},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 2, ColNumber: 21},
		{Type: lexer.FLOAT, Literal: "5.5", LineNumber: 2, ColNumber: 23},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 2, ColNumber: 26},
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 3, ColNumber: 1},
		{Type: lexer.IDENTIFIER, Literal: "foo_bar", LineNumber: 3, ColNumber: 5},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 3, ColNumber: 13},
		{Type: lexer.STRING, Literal: "foo bar", LineNumber: 3, ColNumber: 15},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 3, ColNumber: 24},
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 5, ColNumber: 1},
		{Type: lexer.IDENTIFIER, Literal: "add", LineNumber: 5, ColNumber: 5},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 5, ColNumber: 9},
		{Type: lexer.FUNCTION, Literal: "func", LineNumber: 5, ColNumber: 11},
		{Type: lexer.LPAREN, Literal: "(", LineNumber: 5, ColNumber: 15},
		{Type: lexer.IDENTIFIER, Literal: "x", LineNumber: 5, ColNumber: 16},
		{Type: lexer.COMMA, Literal: ",", LineNumber: 5, ColNumber: 17},
		{Type: lexer.IDENTIFIER, Literal: "y", LineNumber: 5, ColNumber: 19},
		{Type: lexer.RPAREN, Literal: ")", LineNumber: 5, ColNumber: 20},
		{Type: lexer.LBRACE, Literal: "{", LineNumber: 5, ColNumber: 22},
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 6, ColNumber: 2},
		{Type: lexer.IDENTIFIER, Literal: "z", LineNumber: 6, ColNumber: 6},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 6, ColNumber: 8},
		{Type: lexer.IDENTIFIER, Literal: "x", LineNumber: 6, ColNumber: 10},
		{Type: lexer.ADD, Literal: "+", LineNumber: 6, ColNumber: 12},
		{Type: lexer.IDENTIFIER, Literal: "y", LineNumber: 6, ColNumber: 14},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 6, ColNumber: 15},
		{Type: lexer.RETURN, Literal: "return", LineNumber: 7, ColNumber: 2},
		{Type: lexer.LPAREN, Literal: "(", LineNumber: 7, ColNumber: 9},
		{Type: lexer.IDENTIFIER, Literal: "z", LineNumber: 7, ColNumber: 10},
		{Type: lexer.SUBTRACT, Literal: "-", LineNumber: 7, ColNumber: 12},
		{Type: lexer.INTEGER, Literal: "2", LineNumber: 7, ColNumber: 14},
		{Type: lexer.RPAREN, Literal: ")", LineNumber: 7, ColNumber: 15},
		{Type: lexer.MULTIPLY, Literal: "*", LineNumber: 7, ColNumber: 17},
		{Type: lexer.INTEGER, Literal: "2", LineNumber: 7, ColNumber: 19},
		{Type: lexer.DIVIDE, Literal: "/", LineNumber: 7, ColNumber: 21},
		{Type: lexer.INTEGER, Literal: "2", LineNumber: 7, ColNumber: 23},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 7, ColNumber: 24},
		{Type: lexer.RBRACE, Literal: "}", LineNumber: 8, ColNumber: 1},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 8, ColNumber: 2},
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 9, ColNumber: 1},
		{Type: lexer.IDENTIFIER, Literal: "result", LineNumber: 9, ColNumber: 5},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 9, ColNumber: 12},
		{Type: lexer.IDENTIFIER, Literal: "add", LineNumber: 9, ColNumber: 14},
		{Type: lexer.LPAREN, Literal: "(", LineNumber: 9, ColNumber: 17},
		{Type: lexer.IDENTIFIER, Literal: "five", LineNumber: 9, ColNumber: 18},
		{Type: lexer.COMMA, Literal: ",", LineNumber: 9, ColNumber: 22},
		{Type: lexer.IDENTIFIER, Literal: "five_point_five", LineNumber: 9, ColNumber: 24},
		{Type: lexer.RPAREN, Literal: ")", LineNumber: 9, ColNumber: 39},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 9, ColNumber: 40},
		{Type: lexer.VARIABLE, Literal: "var", LineNumber: 12, ColNumber: 1},
		{Type: lexer.IDENTIFIER, Literal: "negate", LineNumber: 12, ColNumber: 5},
		{Type: lexer.ASSIGN, Literal: "=", LineNumber: 12, ColNumber: 12},
		{Type: lexer.NEGATE, Literal: "!", LineNumber: 12, ColNumber: 14},
		{Type: lexer.TRUE, Literal: "true", LineNumber: 12, ColNumber: 15},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 12, ColNumber: 19},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 13, ColNumber: 1},
		{Type: lexer.LTE, Literal: "<=", LineNumber: 13, ColNumber: 3},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 13, ColNumber: 6},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 13, ColNumber: 7},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 14, ColNumber: 1},
		{Type: lexer.GTE, Literal: ">=", LineNumber: 14, ColNumber: 3},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 14, ColNumber: 6},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 14, ColNumber: 7},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 15, ColNumber: 1},
		{Type: lexer.NOTEQ, Literal: "!=", LineNumber: 15, ColNumber: 3},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 15, ColNumber: 6},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 15, ColNumber: 7},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 16, ColNumber: 1},
		{Type: lexer.EQ, Literal: "==", LineNumber: 16, ColNumber: 3},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 16, ColNumber: 6},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 16, ColNumber: 7},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 18, ColNumber: 1},
		{Type: lexer.LT, Literal: "<", LineNumber: 18, ColNumber: 3},
		{Type: lexer.INTEGER, Literal: "10", LineNumber: 18, ColNumber: 5},
		{Type: lexer.GT, Literal: ">", LineNumber: 18, ColNumber: 8},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 18, ColNumber: 10},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 18, ColNumber: 11},
		{Type: lexer.IF, Literal: "if", LineNumber: 19, ColNumber: 1},
		{Type: lexer.LPAREN, Literal: "(", LineNumber: 19, ColNumber: 4},
		{Type: lexer.INTEGER, Literal: "5", LineNumber: 19, ColNumber: 5},
		{Type: lexer.LT, Literal: "<", LineNumber: 19, ColNumber: 7},
		{Type: lexer.INTEGER, Literal: "10", LineNumber: 19, ColNumber: 9},
		{Type: lexer.RPAREN, Literal: ")", LineNumber: 19, ColNumber: 11},
		{Type: lexer.LBRACE, Literal: "{", LineNumber: 19, ColNumber: 13},
		{Type: lexer.RETURN, Literal: "return", LineNumber: 20, ColNumber: 2},
		{Type: lexer.TRUE, Literal: "true", LineNumber: 20, ColNumber: 9},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 20, ColNumber: 13},
		{Type: lexer.RBRACE, Literal: "}", LineNumber: 21, ColNumber: 1},
		{Type: lexer.ELSE, Literal: "else", LineNumber: 21, ColNumber: 3},
		{Type: lexer.LBRACE, Literal: "{", LineNumber: 21, ColNumber: 8},
		{Type: lexer.RETURN, Literal: "return", LineNumber: 22, ColNumber: 2},
		{Type: lexer.FALSE, Literal: "false", LineNumber: 22, ColNumber: 9},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 22, ColNumber: 14},
		{Type: lexer.RBRACE, Literal: "}", LineNumber: 23, ColNumber: 1},
		{Type: lexer.LBRACKET, Literal: "[", LineNumber: 24, ColNumber: 1},
		{Type: lexer.INTEGER, Literal: "1", LineNumber: 24, ColNumber: 2},
		{Type: lexer.COMMA, Literal: ",", LineNumber: 24, ColNumber: 3},
		{Type: lexer.INTEGER, Literal: "2", LineNumber: 24, ColNumber: 5},
		{Type: lexer.COMMA, Literal: ",", LineNumber: 24, ColNumber: 6},
		{Type: lexer.INTEGER, Literal: "3", LineNumber: 24, ColNumber: 8},
		{Type: lexer.RBRACKET, Literal: "]", LineNumber: 24, ColNumber: 9},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 24, ColNumber: 10},
		{Type: lexer.INTEGER, Literal: "55", LineNumber: 25, ColNumber: 1},
		{Type: lexer.SEMICOLON, Literal: ";", LineNumber: 25, ColNumber: 3},
		{Type: lexer.EOF, Literal: "", LineNumber: 25, ColNumber: 4},
	}

	l := lexer.NewLexer([]byte(input))
	for i, expectedToken := range expected {
		if i == len(expected)-2 {
			fmt.Printf("here")
		}
		actualToken := l.NextToken()

		var invalid bool
		if actualToken.Type != expectedToken.Type {
			invalid = true
			t.Errorf("test %d:invalid token type. got=%v, want=%v", i, actualToken.Type, expectedToken.Type)
		}
		if actualToken.Literal != expectedToken.Literal {
			invalid = true
			t.Errorf("test %d:invalid token literal. got=%v, want=%v", i, actualToken.Literal, expectedToken.Literal)
		}
		if actualToken.LineNumber != expectedToken.LineNumber {
			invalid = true
			t.Errorf("test %d:invalid line number. got=%v, want=%v", i, actualToken.LineNumber, expectedToken.LineNumber)
		}
		if actualToken.ColNumber != expectedToken.ColNumber {
			invalid = true
			t.Errorf("test %d: invalid column number. got=%v, want=%v", i, actualToken.ColNumber, expectedToken.ColNumber)
		}
		if invalid {
			t.FailNow()
		}
	}
}
