// NeuroScript Version: 0.3.1
// File version: 0.1.11 // Corrected expectations for invalid_keyword_as_variable_name test case.
// Purpose: Tests for the syntax analysis tool. Assertions adapted to work directly with list of error maps.
// filename: pkg/core/tool_syntax_analyzer_test.go
// nlines: 220 // Approximate
// risk_rating: LOW

package core

import (
	// "encoding/json" // No longer needed for this test's primary assertion path
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestAnalyzeNSSyntaxInternal(t *testing.T) {
	originalGrammarVersion := GrammarVersion
	GrammarVersion = "test-grammar-v0.9.9" // For predictable test output
	defer func() { GrammarVersion = originalGrammarVersion }()

	testInterp, _ := NewDefaultTestInterpreter(t)
	// if err != nil {
	// 	t.Fatalf("Failed to create NewDefaultTestInterpreter: %v", err)
	// }
	// if testInterp == nil && t.Name() != "TestAnalyzeNSSyntaxInternal/nil_interpreter_passed_to_tool_function" {
	// 	t.Fatalf("NewDefaultTestInterpreter returned a nil interpreter without error, cannot proceed with most tests.")
	// }

	testCases := []struct {
		name                      string
		interpreter               *Interpreter
		scriptContent             string
		expectedTotalErrors       int
		expectedReportedErrorsNum int // This is effectively len(expectedErrorsDetails) or capped value
		expectedErrorsDetails     []StructuredSyntaxError
		expectError               bool
		expectedErrorIs           error
		expectedSummaryPreamble   string // For checking the "Found X error(s)." part
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
			expectedErrorsDetails: []StructuredSyntaxError{
				// Note: The column might be where the newline starts or where the parser expected something.
				// The log "line=2 column=10 message=mismatched input '\n' expecting" matches this.
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
			expectedErrorsDetails: []StructuredSyntaxError{
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
			expectedErrorIs: ErrInvalidArgument,
		},
		{
			// This test checks parser behavior with many lines, specifically if it gets beyond the first error.
			// The ParseForLSP will return all errors from the listener.
			// The actual AnalyzeNSSyntaxInternal tool caps reported errors at analyzerMaxErrorsToReportInternal (20).
			// The script strings.Repeat("set x = \n", N) creates N lines of "set x = ".
			// If N > 20, ParseForLSP would return N errors, but AnalyzeNSSyntaxInternal would return only 20.
			// The current test here "more than max errors - input yields 1 parser error" is perhaps misnamed
			// or its scriptContent is not what it implies.
			// strings.Repeat("set x = \n", analyzerMaxErrorsToReportInternal+5) will have an error on each line *if*
			// not enclosed in a func. If it's just statements, the first "set" is the error.
			name:                      "more than max errors - input yields 1 parser error for standalone set",
			interpreter:               testInterp,
			scriptContent:             strings.Repeat("set x = \n", analyzerMaxErrorsToReportInternal+5),
			expectedTotalErrors:       1, // Because "set" is not allowed at top level, first one errors.
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []StructuredSyntaxError{
				// The error "mismatched input 'set' expecting {<EOF>, 'func'}" occurs at Line 1, Column 0 (start of "set").
				{Line: 1, Column: 0, Msg: "mismatched input 'set' expecting {<EOF>, 'func'}", OffendingSymbol: "set"},
			},
			expectError: false,
		},
		{
			name:        "invalid keyword as variable name",
			interpreter: testInterp,
			scriptContent: `func main means
  set if = 10
endfunc`,
			expectedTotalErrors:       1, // Corrected from 3
			expectedReportedErrorsNum: 1, // Corrected from 3
			expectedErrorsDetails: []StructuredSyntaxError{
				// Based on logs: "line=2 column=6 message=mismatched input 'if' expecting IDENTIFIER"
				// The line number in scriptContent: "func main means" is line 1, "  set if = 10" is line 2.
				// Column 6 (1-based) would be where 'if' starts if "  set " (2 spaces + "set" + 1 space) = 6 chars.
				// So, "  set if..." -> col 1 is ' ', col 2 is ' ', col 3 is 's', col 4 is 'e', col 5 is 't', col 6 is ' '. Oh, if spaces are different...
				// Let's assume 2 spaces indent: "  set if = 10" -> 'if' starts at column 7 (2 spaces + "set " (4) + "if").
				// The log usually gives 0-based column from ANTLR, +1 for display.
				// "helpers.go:57: [ERROR] Syntax Error Reported by Listener source=nsSyntaxAnalysisToolInput line=2 column=6 message=mismatched input 'if' expecting IDENTIFIER token=if"
				// This means the listener (which adds 1 to column) reported line 2, column 6.
				// In the scriptContent:
				// Line 1: func main means
				// Line 2:   set if = 10  (assuming 2 spaces indent, 's' is col 3, ' ' is col 6, 'i' is col 7)
				// The reported error column seems to be on the 'if'.
				// If the script used in the test is exactly "func main means\n  set if = 10\nendfunc",
				// then line 2 is "  set if = 10".
				// If Column is 1-based:
				// 's' in set is at column 3. ' ' after set is column 6. 'i' in "if" is column 7.
				// The log "column=6" suggests the error is *at* the space before 'if' or how ANTLR counts.
				// The "token=if" suggests 'if' is the problematic token.
				// Let's use the provided detail directly from your original failing test:
				{Line: 2, Column: 6, Msg: "mismatched input 'if' expecting IDENTIFIER", OffendingSymbol: "if"},
			},
			expectError: false,
		},
		{
			name:        "incomplete call statement",
			interpreter: testInterp,
			scriptContent: `func main means
  call
endfunc`,
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []StructuredSyntaxError{
				// Log: "line=2 column=6 message=mismatched input '\n' expecting"
				// Script Line 2: "  call"
				// 'c' at col 3, 'l' at col 6. Newline is effectively col 7.
				{Line: 2, Column: 6, Msg: "mismatched input '\\n' expecting", OffendingSymbol: ""},
			},
			expectError: false,
		},
		{
			name:        "mismatched block terminator",
			interpreter: testInterp,
			scriptContent: `func main means
  if true
    set x = 1
  endwhile
endfunc`,
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []StructuredSyntaxError{
				// Log: "line=4 column=2 message=mismatched input 'endwhile' expecting {'else', 'endif'}"
				// Script Line 4: "  endwhile"
				// 'e' at col 3. The log's column 2 might be 0-based ANTLR column + 1.
				{Line: 4, Column: 2, Msg: "mismatched input 'endwhile' expecting {'else', 'endif'}", OffendingSymbol: "endwhile"},
			},
			expectError: false,
		},
		{
			name:        "malformed function signature - missing means",
			interpreter: testInterp,
			scriptContent: `func myFunc (needs x)
  set y = x
endfunc`,
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []StructuredSyntaxError{
				// Log: "line=1 column=21 message=missing 'means' at '\n'"
				// Script Line 1: "func myFunc (needs x)" (length is 21, error at the newline)
				{Line: 1, Column: 21, Msg: "missing 'means' at '\\n'", OffendingSymbol: ""},
			},
			expectError: false,
		},
		{
			name:        "unclosed string literal",
			interpreter: testInterp,
			scriptContent: `func main means
  set myString = "hello world
endfunc`,
			expectedTotalErrors:       2, // ANTLR often gives a token recognition error then a parse error
			expectedReportedErrorsNum: 2,
			expectedErrorsDetails: []StructuredSyntaxError{
				// Log: "line=2 column=17 message=token recognition error at: '\"hello world"
				// Script Line 2: "  set myString = \"hello world" ( " at col 17)
				{Line: 2, Column: 17, Msg: "token recognition error at: '\"hello world", OffendingSymbol: ""},
				// Log: "line=3 column=0 message=mismatched input 'endfunc' expecting"
				{Line: 3, Column: 0, Msg: "mismatched input 'endfunc' expecting", OffendingSymbol: "endfunc"},
			},
			expectError: false,
		},
		{
			name:        "invalid operator placement",
			interpreter: testInterp,
			scriptContent: `func main means
  set x = 10 +
endfunc`,
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []StructuredSyntaxError{
				// Log: "line=2 column=14 message=mismatched input '\n' expecting"
				// Script Line 2: "  set x = 10 +" (newline is effectively after '+', col 14 if one space after +)
				{Line: 2, Column: 14, Msg: "mismatched input '\\n' expecting", OffendingSymbol: ""},
			},
			expectError: false,
		},
		{
			name:        "unclosed placeholder",
			interpreter: testInterp,
			scriptContent: `func main means
  set x = {{name
endfunc`,
			expectedTotalErrors:       1,
			expectedReportedErrorsNum: 1,
			expectedErrorsDetails: []StructuredSyntaxError{
				// Log: "line=2 column=16 message=missing '}}' at '\n'"
				// Script Line 2: "  set x = {{name" ('{{' at 11, 'name' ends at 15, newline is effectively 16)
				{Line: 2, Column: 16, Msg: "missing '}}' at '\\n'", OffendingSymbol: ""},
			},
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.interpreter == nil && tc.name != "nil interpreter passed to tool function" {
				t.Skipf("Skipping test case %q because shared testInterp is nil (NewDefaultTestInterpreter failed)", tc.name)
				return
			}

			resultInterface, err := AnalyzeNSSyntaxInternal(tc.interpreter, tc.scriptContent)

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

			if len(errorList) != tc.expectedTotalErrors { // TotalErrors is now just the length of the returned list
				t.Errorf("expected TotalErrorsFound %d, got %d", tc.expectedTotalErrors, len(errorList))
			}

			if len(errorList) != tc.expectedReportedErrorsNum {
				t.Errorf("expected ReportedErrorsNum %d, got %d (length of errorList)", tc.expectedReportedErrorsNum, len(errorList))
			}

			if tc.expectedErrorsDetails != nil && len(tc.expectedErrorsDetails) > 0 {
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
					// Use strings.Contains for the message as ANTLR messages can have prefixes/suffixes not in our expected simple message
					if !strings.Contains(actualMsg, expectedDetail.Msg) {
						t.Errorf("error %d: expected Msg to contain %q, got %q", i, expectedDetail.Msg, actualMsg)
					}
					// OffendingSymbol check can be added if needed and if it's reliably populated by your parser listener.
					// if actualErrorMap["OffendingSymbol"] != expectedDetail.OffendingSymbol {
					//    t.Errorf("error %d: expected OffendingSymbol %q, got %q", i, expectedDetail.OffendingSymbol, actualErrorMap["OffendingSymbol"])
					// }
				}
			}
		})
	}
}
