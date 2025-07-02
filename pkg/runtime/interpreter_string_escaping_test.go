// NeuroScript Version: 0.4.0
// File version: 0.2.0
// Purpose: Corrected expectedVars map literals to use map[string]Value, ensuring type safety in tests.
// filename: pkg/runtime/interpreter_string_escaping_test.go
// nlines: 175
// risk_rating: LOW

package runtime

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// TestInterpretStringEscaping verifies that the interpreter correctly
// handles strings that have been unescaped by the AST builder.
func TestInterpretStringEscaping(t *testing.T) {
	pos := &lang.Position{Line: 1, Column: 1, File: "escape_integration_test"}

	testCases := []testutil.executeStepsTestCase{
		{
			name: "Interpret Backspace",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "text\bback"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "text\bback"}},
			expectedResult: lang.StringValue{Value: "text\bback"},
		},
		{
			name: "Interpret Tab",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "col1\tcol2"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "col1\tcol2"}},
			expectedResult: lang.StringValue{Value: "col1\tcol2"},
		},
		{
			name: "Interpret Newline",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "first\nsecond"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "first\nsecond"}},
			expectedResult: lang.StringValue{Value: "first\nsecond"},
		},
		{
			name: "Interpret Formfeed",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "page1\fpage2"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "page1\fpage2"}},
			expectedResult: lang.StringValue{Value: "page1\fpage2"},
		},
		{
			name: "Interpret Carriage Return",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "over\rwrite"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "over\rwrite"}},
			expectedResult: lang.StringValue{Value: "over\rwrite"},
		},
		{
			name: "Interpret Vertical Tab",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "v\vtab"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "v\vtab"}},
			expectedResult: lang.StringValue{Value: "v\vtab"},
		},
		{
			name: "Interpret Tilde",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "approx~equal"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "approx~equal"}},
			expectedResult: lang.StringValue{Value: "approx~equal"},
		},
		{
			name: "Interpret Backtick",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "code `block`"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "code `block`"}},
			expectedResult: lang.StringValue{Value: "code `block`"},
		},
		{
			name: "Interpret Double Quote",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: `a "quoted" string`}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: `a "quoted" string`}},
			expectedResult: lang.StringValue{Value: `a "quoted" string`},
		},
		{
			name: "Interpret Single Quote",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "it's great"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "it's great"}},
			expectedResult: lang.StringValue{Value: "it's great"},
		},
		{
			name: "Interpret Backslash",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: `a path C:\folder`}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: `a path C:\folder`}},
			expectedResult: lang.StringValue{Value: `a path C:\folder`},
		},
		{
			name: "Interpret Unicode BMP",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "currency: €"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "currency: €"}},
			expectedResult: lang.StringValue{Value: "currency: €"},
		},
		{
			name: "Interpret Unicode Surrogate Pair",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "face: 😀"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "face: 😀"}},
			expectedResult: lang.StringValue{Value: "face: 😀"},
		},
		{
			name: "Interpret Unicode Unpaired High Surrogate",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "unpaired: " + string(rune(0xFFFD)) + " after"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "unpaired: " + string(rune(0xFFFD)) + " after"}},
			expectedResult: lang.StringValue{Value: "unpaired: " + string(rune(0xFFFD)) + " after"},
		},
		{
			name: "Interpret Unicode High Surrogate at EOS",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: "eos: " + string(rune(0xFFFD))}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: "eos: " + string(rune(0xFFFD))}},
			expectedResult: lang.StringValue{Value: "eos: " + string(rune(0xFFFD))},
		},
		{
			name: "Interpret Unicode High Surrogate followed by non-low surrogate unicode",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: string(rune(0xFFFD)) + "A"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: string(rune(0xFFFD)) + "A"}},
			expectedResult: lang.StringValue{Value: string(rune(0xFFFD)) + "A"},
		},
		{
			name: "Interpret Unicode High Surrogate followed by non-unicode escape",
			inputSteps: []ast.Step{
				testutil.createTestStep("set", "val", &ast.StringLiteralNode{Position: pos, Value: string(rune(0xFFFD)) + "\n"}, nil),
			},
			expectedVars:   map[string]lang.Value{"val": lang.StringValue{Value: string(rune(0xFFFD)) + "\n"}},
			expectedResult: lang.StringValue{Value: string(rune(0xFFFD)) + "\n"},
		},
	}

	for _, tc := range testCases {
		testutil.runExecuteStepsTest(t, tc)
	}
}
