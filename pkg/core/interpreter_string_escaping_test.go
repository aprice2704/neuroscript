// NeuroScript Version: 0.4.0
// File version: 0.1.4
// Purpose: Updated string escaping tests to expect core.Value types.
// filename: pkg/core/interpreter_string_escaping_test.go
// nlines: 175
// risk_rating: LOW

package core

import (
	"testing"
)

// TestInterpretStringEscaping verifies that the interpreter correctly
// handles strings that have been unescaped by the AST builder.
func TestInterpretStringEscaping(t *testing.T) {
	pos := &Position{Line: 1, Column: 1, File: "escape_integration_test"}

	testCases := []executeStepsTestCase{
		{
			name: "Interpret Backspace",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "text\bback"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "text\bback"}},
			expectedResult: StringValue{Value: "text\bback"},
		},
		{
			name: "Interpret Tab",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "col1\tcol2"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "col1\tcol2"}},
			expectedResult: StringValue{Value: "col1\tcol2"},
		},
		{
			name: "Interpret Newline",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "first\nsecond"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "first\nsecond"}},
			expectedResult: StringValue{Value: "first\nsecond"},
		},
		{
			name: "Interpret Formfeed",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "page1\fpage2"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "page1\fpage2"}},
			expectedResult: StringValue{Value: "page1\fpage2"},
		},
		{
			name: "Interpret Carriage Return",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "over\rwrite"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "over\rwrite"}},
			expectedResult: StringValue{Value: "over\rwrite"},
		},
		{
			name: "Interpret Vertical Tab",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "v\vtab"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "v\vtab"}},
			expectedResult: StringValue{Value: "v\vtab"},
		},
		{
			name: "Interpret Tilde",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "approx~equal"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "approx~equal"}},
			expectedResult: StringValue{Value: "approx~equal"},
		},
		{
			name: "Interpret Backtick",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "code `block`"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "code `block`"}},
			expectedResult: StringValue{Value: "code `block`"},
		},
		{
			name: "Interpret Double Quote",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: `a "quoted" string`}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: `a "quoted" string`}},
			expectedResult: StringValue{Value: `a "quoted" string`},
		},
		{
			name: "Interpret Single Quote",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "it's great"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "it's great"}},
			expectedResult: StringValue{Value: "it's great"},
		},
		{
			name: "Interpret Backslash",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: `a path C:\folder`}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: `a path C:\folder`}},
			expectedResult: StringValue{Value: `a path C:\folder`},
		},
		{
			name: "Interpret Unicode BMP",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "currency: €"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "currency: €"}},
			expectedResult: StringValue{Value: "currency: €"},
		},
		{
			name: "Interpret Unicode Surrogate Pair",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "face: 😀"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "face: 😀"}},
			expectedResult: StringValue{Value: "face: 😀"},
		},
		{
			name: "Interpret Unicode Unpaired High Surrogate",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "unpaired: " + string(rune(0xFFFD)) + " after"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "unpaired: " + string(rune(0xFFFD)) + " after"}},
			expectedResult: StringValue{Value: "unpaired: " + string(rune(0xFFFD)) + " after"},
		},
		{
			name: "Interpret Unicode High Surrogate at EOS",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "eos: " + string(rune(0xFFFD))}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: "eos: " + string(rune(0xFFFD))}},
			expectedResult: StringValue{Value: "eos: " + string(rune(0xFFFD))},
		},
		{
			name: "Interpret Unicode High Surrogate followed by non-low surrogate unicode",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: string(rune(0xFFFD)) + "A"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: string(rune(0xFFFD)) + "A"}},
			expectedResult: StringValue{Value: string(rune(0xFFFD)) + "A"},
		},
		{
			name: "Interpret Unicode High Surrogate followed by non-unicode escape",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: string(rune(0xFFFD)) + "\n"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": StringValue{Value: string(rune(0xFFFD)) + "\n"}},
			expectedResult: StringValue{Value: string(rune(0xFFFD)) + "\n"},
		},
	}

	for _, tc := range testCases {
		runExecuteStepsTest(t, tc)
	}
}
