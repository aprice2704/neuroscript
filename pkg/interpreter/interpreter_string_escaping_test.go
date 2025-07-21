// NeuroScript Version: 0.5.2
// File version: 7.1.0
// Purpose: Removed unused 'pos' variable to fix compiler error.
// filename: pkg/interpreter/interpreter_string_escaping_test.go
// nlines: 85
// risk_rating: LOW

package interpreter

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestInterpretStringEscaping verifies that the interpreter correctly
// handles strings that have been unescaped by the AST builder.
func TestInterpretStringEscaping(t *testing.T) {
	// FIX: Use the local test case struct from interpreter_suite_test.go
	testCases := []localExecuteStepsTestCase{
		{
			name: "Interpret Backspace",
			inputSteps: []ast.Step{
				// FIX: Use the local helper from interpreter_suite_test.go
				createTestStep("set", "val", &ast.StringLiteralNode{Value: "text\bback"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "text\bback"}},
			expectedResult: lang.StringValue{Value: "text\bback"},
		},
		{
			name: "Interpret Tab",
			inputSteps: []ast.Step{
				createTestStep("set", "val", &ast.StringLiteralNode{Value: "col1\tcol2"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "col1\tcol2"}},
			expectedResult: lang.StringValue{Value: "col1\tcol2"},
		},
		{
			name: "Interpret Newline",
			inputSteps: []ast.Step{
				createTestStep("set", "val", &ast.StringLiteralNode{Value: "first\nsecond"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "first\nsecond"}},
			expectedResult: lang.StringValue{Value: "first\nsecond"},
		},
		{
			name: "Interpret Double Quote",
			inputSteps: []ast.Step{
				createTestStep("set", "val", &ast.StringLiteralNode{Value: `a "quoted" string`}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: `a "quoted" string`}},
			expectedResult: lang.StringValue{Value: `a "quoted" string`},
		},
		{
			name: "Interpret Backslash",
			inputSteps: []ast.Step{
				createTestStep("set", "val", &ast.StringLiteralNode{Value: `a path C:\folder`}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: `a path C:\folder`}},
			expectedResult: lang.StringValue{Value: `a path C:\folder`},
		},
		{
			name: "Interpret Unicode BMP",
			inputSteps: []ast.Step{
				createTestStep("set", "val", &ast.StringLiteralNode{Value: "currency: €"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "currency: €"}},
			expectedResult: lang.StringValue{Value: "currency: €"},
		},
		{
			name: "Interpret Unicode Surrogate Pair",
			inputSteps: []ast.Step{
				createTestStep("set", "val", &ast.StringLiteralNode{Value: "face: 😀"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "face: 😀"}},
			expectedResult: lang.StringValue{Value: "face: 😀"},
		},
	}

	for _, tc := range testCases {
		// FIX: Use the local test runner from interpreter_suite_test.go
		runLocalExecuteStepsTest(t, tc)
	}
}
