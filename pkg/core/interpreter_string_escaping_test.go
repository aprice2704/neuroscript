// NeuroScript Version: 0.4.0
// File version: 0.1.2 // Corrected surrogate representation for interpreter tests, removed AST-build error tests.
// Purpose: Integration tests for string literal unescaping by the interpreter.
// filename: pkg/core/interpreter_string_escaping_test.go
package core

import (
	"testing"
	// testing_helpers.go (containing createTestStep, runExecuteStepsTest, executeStepsTestCase)
	// and ast.go (Position, StringLiteralNode, etc.) are assumed to be in this package.
)

// TestInterpretStringEscaping verifies that the interpreter correctly
// handles strings that have been unescaped by the AST builder.
func TestInterpretStringEscaping(t *testing.T) {
	pos := &Position{Line: 1, Column: 1, File: "escape_integration_test"}

	testCases := []executeStepsTestCase{
		// Basic Escapes (Value field contains the Go string after unescaping)
		{
			name: "Interpret Backspace",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "text\bback"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "text\bback"},
			expectedResult: "text\bback",
		},
		{
			name: "Interpret Tab",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "col1\tcol2"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "col1\tcol2"},
			expectedResult: "col1\tcol2",
		},
		{
			name: "Interpret Newline",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "first\nsecond"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "first\nsecond"},
			expectedResult: "first\nsecond",
		},
		{
			name: "Interpret Formfeed",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "page1\fpage2"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "page1\fpage2"},
			expectedResult: "page1\fpage2",
		},
		{
			name: "Interpret Carriage Return",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "over\rwrite"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "over\rwrite"},
			expectedResult: "over\rwrite",
		},
		{
			name: "Interpret Vertical Tab",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "v\vtab"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "v\vtab"},
			expectedResult: "v\vtab",
		},
		{
			name: "Interpret Tilde",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "approx~equal"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "approx~equal"},
			expectedResult: "approx~equal",
		},
		{
			name: "Interpret Backtick",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "code `block`"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "code `block`"},
			expectedResult: "code `block`",
		},

		// Quote Escapes
		{
			name: "Interpret Double Quote",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: `a "quoted" string`}, nil),
			},
			expectedVars:   map[string]interface{}{"val": `a "quoted" string`},
			expectedResult: `a "quoted" string`,
		},
		{
			name: "Interpret Single Quote",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: `it's great`}, nil),
			},
			expectedVars:   map[string]interface{}{"val": `it's great`},
			expectedResult: `it's great`,
		},
		{
			name: "Interpret Backslash",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: `a path C:\folder`}, nil),
			},
			expectedVars:   map[string]interface{}{"val": `a path C:\folder`},
			expectedResult: `a path C:\folder`,
		},

		// Unicode Escapes
		{
			name: "Interpret Unicode BMP",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "currency: €"}, nil), // Euro sign
			},
			expectedVars:   map[string]interface{}{"val": "currency: €"},
			expectedResult: "currency: €",
		},
		{
			name: "Interpret Unicode Surrogate Pair",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "face: 😀"}, nil), // Grinning face emoji
			},
			expectedVars:   map[string]interface{}{"val": "face: 😀"},
			expectedResult: "face: 😀",
		},
		{
			name: "Interpret Unicode Unpaired High Surrogate",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "unpaired: " + string(rune(0xD83D)) + " after"}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "unpaired: " + string(rune(0xD83D)) + " after"},
			expectedResult: "unpaired: " + string(rune(0xD83D)) + " after",
		},
		{
			name: "Interpret Unicode High Surrogate at EOS",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: "eos: " + string(rune(0xD83D))}, nil),
			},
			expectedVars:   map[string]interface{}{"val": "eos: " + string(rune(0xD83D))},
			expectedResult: "eos: " + string(rune(0xD83D)),
		},
		{
			name: "Interpret Unicode High Surrogate followed by non-low surrogate unicode",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: string(rune(0xD83D)) + "A"}, nil), // High surrogate followed by 'A'
			},
			expectedVars:   map[string]interface{}{"val": string(rune(0xD83D)) + "A"},
			expectedResult: string(rune(0xD83D)) + "A",
		},
		{
			name: "Interpret Unicode High Surrogate followed by non-unicode escape",
			inputSteps: []Step{
				createTestStep("set", "val", &StringLiteralNode{Pos: pos, Value: string(rune(0xD83D)) + "\n"}, nil), // High surrogate followed by newline
			},
			expectedVars:   map[string]interface{}{"val": string(rune(0xD83D)) + "\n"},
			expectedResult: string(rune(0xD83D)) + "\n",
		},

		// Note: Error cases for string unescaping (e.g., "string ends with bare backslash",
		// "incomplete unicode escape", "invalid hex in unicode escape", "unknown escape sequence")
		// are errors that occur during the AST building phase when UnescapeNeuroScriptString is called.
		// To test these, one would typically test the ASTBuilder.Build method directly with
		// NeuroScript source code strings designed to cause these specific unescaping errors,
		// and then assert that ASTBuilder.Build returns an error.
		// These types of errors are not tested in this file, as this file focuses on
		// the interpreter's behavior with already constructed (and correctly unescaped) AST nodes.
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			runExecuteStepsTest(t, tc)
		})
	}
}
