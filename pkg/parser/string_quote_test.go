// NeuroScript Version: 0.7.0
// File version: 3
// Purpose: Corrected test case logic to no longer double-wrap expressions in a function block, which was causing parser failures.
// filename: pkg/parser/string_quote_test.go
// nlines: 80
// risk_rating: LOW

package parser

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestStringLiteralParsing(t *testing.T) {
	testCases := []struct {
		name          string
		scriptLiteral string
		expectedValue string
	}{
		{
			name:          "simple double quotes",
			scriptLiteral: `"hello"`,
			expectedValue: "hello",
		},
		{
			name:          "simple single quotes",
			scriptLiteral: `'world'`,
			expectedValue: "world",
		},
		{
			name:          "double quotes inside single",
			scriptLiteral: `'a "b" c'`,
			expectedValue: `a "b" c`,
		},
		{
			name:          "single quotes inside double",
			scriptLiteral: `"d 'e' f"`,
			expectedValue: `d 'e' f`,
		},
		{
			name:          "json in single quotes",
			scriptLiteral: `'{"key":"value"}'`,
			expectedValue: `{"key":"value"}`,
		},
		{
			name:          "escaped double quote",
			scriptLiteral: `"a \" b"`,
			expectedValue: `a " b`,
		},
		{
			name:          "escaped single quote",
			scriptLiteral: `'c \' d'`,
			expectedValue: `c ' d`,
		},
		{
			name:          "simple raw string",
			scriptLiteral: "```raw string```",
			expectedValue: "raw string",
		},
		{
			name:          "raw string with quotes",
			scriptLiteral: "```'hello' and \"world\"```",
			expectedValue: `'hello' and "world"`,
		},
		{
			name:          "empty double-quoted string",
			scriptLiteral: `""`,
			expectedValue: "",
		},
		{
			name:          "empty single-quoted string",
			scriptLiteral: `''`,
			expectedValue: "",
		},
		{
			name:          "empty raw string",
			scriptLiteral: "``````",
			expectedValue: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// FIX: The parseExpression helper already wraps the expression in a valid
			// function and set statement. The script variable should ONLY be the literal itself.
			// The previous code was creating a nested function, causing the syntax error.
			expr := parseExpression(t, tc.scriptLiteral)
			strNode, ok := expr.(*ast.StringLiteralNode)
			if !ok {
				t.Fatalf("Expected a StringLiteralNode, but got %T", expr)
			}
			if strNode.Value != tc.expectedValue {
				t.Errorf("Expected value '%s', but got '%s'", tc.expectedValue, strNode.Value)
			}
		})
	}
}
