// NeuroScript Version: 0.3.1
// File version: 0.1.15
// Purpose: Implements a more robust parser error check that is less brittle to grammar changes.
// filename: pkg/tool/syntax/tools_syntax_analyzer_test.go
// nlines: 220
// risk_rating: LOW

package syntax_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/tool/syntax"
)

func TestAnalyzeNSSyntaxInternal(t *testing.T) {
	originalGrammarVersion := lang.GrammarVersion
	lang.GrammarVersion = "test-grammar-v0.9.9" // For predictable test output
	defer func() { lang.GrammarVersion = originalGrammarVersion }()

	testInterp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}

	testCases := []struct {
		name                      string
		interpreter               tool.Runtime
		scriptContent             string
		expectedTotalErrors       int
		expectedReportedErrorsNum int
		expectedErrorsDetails     []parser.StructuredSyntaxError
		expectError               bool
		expectedErrorIs           error
		expectedSummaryPreamble   string
	}{
		{
			name:                      "valid script - no errors",
			interpreter:               testInterp,
			scriptContent:             "func main means\n  set x = 10\n  emit x\nendfunc",
			expectedTotalErrors:       0,
			expectedReportedErrorsNum: 0,
			expectError:               false,
		},
		{
			name:                      "script with one syntax error - incomplete set",
			interpreter:               testInterp,
			scriptContent:             "func main means\n  set x = \nendfunc",
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []parser.StructuredSyntaxError{
				{Line: 2, Column: 10, Msg: "mismatched input '\\n' expecting", OffendingSymbol: ""},
			},
			expectError: false,
		},
		{
			name:                      "script with multiple syntax errors - incomplete set and call",
			interpreter:               testInterp,
			scriptContent:             "func main means\n  set x = \n  call \nendfunc",
			expectedTotalErrors:       2,
			expectedReportedErrorsNum: 2,
			expectedErrorsDetails: []parser.StructuredSyntaxError{
				{Line: 2, Column: 10, Msg: "mismatched input '\\n' expecting", OffendingSymbol: ""},
				{Line: 3, Column: 7, Msg: "mismatched input '\\n' expecting", OffendingSymbol: ""},
			},
			expectError: false,
		},
		{
			name:                      "empty script",
			interpreter:               testInterp,
			scriptContent:             "",
			expectedTotalErrors:       0,
			expectedReportedErrorsNum: 0,
			expectError:               false,
		},
		{
			name:            "nil interpreter passed to tool function",
			interpreter:     nil,
			scriptContent:   "set x = 1",
			expectError:     true,
			expectedErrorIs: lang.ErrInvalidArgument,
		},
		{
			name:                      "more than max errors - input yields 1 parser error for standalone set",
			interpreter:               testInterp,
			scriptContent:             "set x = 1", // Simplified from repeat, the error is the same.
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []parser.StructuredSyntaxError{
				// This check is now more robust. It verifies the core error ("mismatched input 'set'")
				// without being brittle about the list of all possible expected tokens, which can
				// change frequently with grammar updates.
				{Line: 1, Column: 0, Msg: "mismatched input 'set' expecting", OffendingSymbol: "set"},
			},
			expectError: false,
		},
		{
			name:        "invalid keyword as variable name",
			interpreter: testInterp,
			scriptContent: `func main means
  set if = 10
endfunc`,
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []parser.StructuredSyntaxError{
				{Line: 2, Column: 6, Msg: "mismatched input 'if' expecting IDENTIFIER", OffendingSymbol: "if"},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.interpreter == nil && tc.name != "nil interpreter passed to tool function" {
				t.Skipf("Skipping test case %q because shared testInterp is nil (NewTestInterpreter failed)", tc.name)
				return
			}

			resultInterface, err := syntax.AnalyzeNSSyntaxInternal(tc.interpreter, tc.scriptContent)

			if tc.expectError {
				if err == nil {
					t.Errorf("expected an error, but got nil")
					return
				}
				if tc.expectedErrorIs != nil && !errors.Is(err, tc.expectedErrorIs) {
					t.Errorf("expected error to be %v, but got %v", tc.expectedErrorIs, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("AnalyzeNSSyntaxInternal() returned an unexpected error: %v", err)
			}

			errorList, ok := resultInterface.([]map[string]interface{})
			if !ok {
				t.Fatalf("AnalyzeNSSyntaxInternal did not return a []map[string]interface{}. Got type: %T, value: %v", resultInterface, resultInterface)
			}

			if len(errorList) != tc.expectedTotalErrors {
				t.Errorf("expected TotalErrorsFound %d, got %d", tc.expectedTotalErrors, len(errorList))
			}

			if len(errorList) != tc.expectedReportedErrorsNum {
				t.Errorf("expected ReportedErrorsNum %d, got %d (length of errorList)", tc.expectedReportedErrorsNum, len(errorList))
			}

			if len(tc.expectedErrorsDetails) > 0 {
				if len(errorList) < len(tc.expectedErrorsDetails) {
					t.Errorf("expected at least %d reported errors for detail checking, got %d", len(tc.expectedErrorsDetails), len(errorList))
				}
				for i, expectedDetail := range tc.expectedErrorsDetails {
					if i >= len(errorList) {
						break
					}
					actualErrorMap := errorList[i]

					var actualLine, actualColumn int
					if lineVal, ok := actualErrorMap["Line"].(float64); ok {
						actualLine = int(lineVal)
					} else if lineVal, ok := actualErrorMap["Line"].(int); ok {
						actualLine = lineVal
					}

					if colVal, ok := actualErrorMap["Column"].(float64); ok {
						actualColumn = int(colVal)
					} else if colVal, ok := actualErrorMap["Column"].(int); ok {
						actualColumn = colVal
					}

					actualMsg := fmt.Sprintf("%v", actualErrorMap["Msg"])

					if actualLine != expectedDetail.Line {
						t.Errorf("error %d: expected Line %d, got %d", i, expectedDetail.Line, actualLine)
					}
					if actualColumn != expectedDetail.Column {
						t.Errorf("error %d: expected Column %d, got %d", i, expectedDetail.Column, actualColumn)
					}
					if !strings.Contains(actualMsg, expectedDetail.Msg) {
						t.Errorf("error %d: expected Msg to contain %q, got %q", i, expectedDetail.Msg, actualMsg)
					}
				}
			}
		})
	}
}
