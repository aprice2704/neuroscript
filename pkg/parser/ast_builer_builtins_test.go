// filename: pkg/parser/ast_builder_builtins_test.go
// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Adds dedicated test coverage for built-in functions like len(), log(), and trigonometric functions.
// nlines: 60
// risk_rating: LOW

package parser

import (
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestBuiltinFunctionParsing(t *testing.T) {
	testCases := []string{
		"len(my_list)",
		"ln(x)",
		"log(x)",
		"sin(x)",
		"cos(x)",
		"tan(x)",
		"asin(x)",
		"acos(x)",
		"atan(x)",
	}

	for _, tc := range testCases {
		// The test name is the function call itself, e.g., "len(my_list)"
		t.Run(tc, func(t *testing.T) {
			expr := parseExpression(t, tc)

			callNode, ok := expr.(*ast.CallableExprNode)
			if !ok {
				t.Fatalf("Expected a CallableExprNode, got %T", expr)
			}

			// Extract the function name (e.g., "len" from "len(my_list)")
			expectedFuncName := tc[:strings.Index(tc, "(")]
			if callNode.Target.Name != expectedFuncName {
				t.Errorf("Expected target name to be '%s', but got '%s'", expectedFuncName, callNode.Target.Name)
			}

			if callNode.Target.IsTool {
				t.Error("Expected IsTool to be false for a built-in function")
			}
		})
	}
}
