// NeuroScript Version: 0.7.0
// File version: 2
// Purpose: Expands test coverage for string literals to include raw strings and various quote combinations.
// filename: pkg/parser/string_quotes_test.go
// nlines: 80
// risk_rating: LOW

package parser

import (
	"fmt"
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
			// Using fmt.Sprintf to avoid complex Go string literal escaping in the test setup itself.
			script := fmt.Sprintf("func t() means\n set x = %s\nendfunc", tc.scriptLiteral)
			expr := parseExpression(t, script)
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
