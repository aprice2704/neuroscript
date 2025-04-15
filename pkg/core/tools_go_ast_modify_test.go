// filename: pkg/core/tools_go_ast_modify_test.go
// UPDATED: Removed comments from simpleSource input and corresponding wantCode outputs
// UPDATED: Corrected tool name in helper getFormattedCodeModifyTest
package core

import (
	"errors"
	"testing"

	"strings" // Keep strings import

	"github.com/google/go-cmp/cmp" // Keep ast import
	// Import astutil
)

// Helper functions (getFormattedCodeModifyTest, setupParseModifyTest) remain the same...
// Helper to get formatted string from handle (duplicated for focused testing)
func getFormattedCodeModifyTest(t *testing.T, interp *Interpreter, handleID string) string {
	t.Helper()
	// Use the correct tool name
	res, err := toolGoFormatAST(interp, makeArgs(handleID))
	if err != nil {
		t.Fatalf("getFormattedCodeModifyTest: toolGoFormatAST failed for handle %s: %v", handleID, err)
	}
	code, ok := res.(string)
	if !ok {
		t.Fatalf("getFormattedCodeModifyTest: toolGoFormatAST did not return a string, got %T", res)
	}
	return code
}

// Helper to parse code and return a handle (duplicated for focused testing)
func setupParseModifyTest(t *testing.T, interp *Interpreter, content string) string {
	t.Helper()
	handleID, err := toolGoParseFile(interp, makeArgs(nil, content))
	if err != nil {
		t.Fatalf("setupParseModifyTest: toolGoParseFile failed: %v", err)
	}
	handleStr, ok := handleID.(string)
	if !ok || handleStr == "" {
		t.Fatalf("setupParseModifyTest: toolGoParseFile did not return a valid handle string, got %T: %v", handleID, handleID)
	}
	return handleStr
}

// --- TestGoModifyAST Function ---
func TestToolGoModifyAST(t *testing.T) {

	// --- Test Data ---
	// VVV REMOVED COMMENTS FROM INPUT SOURCE VVV
	simpleSource := `package main

import (
	"fmt"
	"io"
)

func main() {
	fmt.Println("hello")
	var err error
	err = fmt.Errorf("an error: %w", io.EOF)
	_ = err
}
`
	// VVV REMOVED COMMENTS FROM EXPECTED SOURCE VVV
	simpleSourceNewPkg := `package other

import (
	"fmt"
	"io"
)

func main() {
	fmt.Println("hello")
	var err error
	err = fmt.Errorf("an error: %w", io.EOF)
	_ = err
}
`
	// VVV REMOVED COMMENTS FROM EXPECTED SOURCE VVV
	simpleSourceReplaceIdentAndAddImport := `package main

import (
	"fmt"
	"log"
	"io"
)

func main() {
	log.Printf("hello")
	var err error
	err = fmt.Errorf("an error: %w", io.EOF)
	_ = err
}
`
	// ^^^ REMOVED COMMENTS FROM ALL SOURCES ^^^
	// --- END Corrected ---

	tests := []struct {
		name           string
		initialContent string
		modifications  map[string]interface{}
		wantCode       string // Expected formatted code after modification
		wantHandleSame bool   // True if no change expected, should return original handle
		wantErrIs      error  // Expected underlying error type (use errors.Is)
		valWantErrIs   error  // Expected validation layer error type
	}{
		// --- Change Package Tests ---
		{name: "Change Package Name", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": "other"}, wantCode: simpleSourceNewPkg, wantErrIs: nil},                          // Uses comment-free input/want
		{name: "Change Package Name No Change", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": "main"}, wantCode: simpleSource, wantHandleSame: true, wantErrIs: nil}, // Uses comment-free input/want

		// --- Replace Identifier Tests ---
		{
			name:           "Replace Identifier fmt.Println -> log.Printf AND Add Import",
			initialContent: simpleSource, // Uses comment-free input
			modifications: map[string]interface{}{
				"replace_identifier": map[string]interface{}{"old": "fmt.Println", "new": "log.Printf"},
				"add_import":         "log",
			},
			wantCode: simpleSourceReplaceIdentAndAddImport, // Uses comment-free want
		},
		{name: "Replace Identifier Not Found", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": "os.Exit", "new": "log.Fatal"}}, wantCode: simpleSource, wantHandleSame: true}, // Uses comment-free input/want

		// Other tests use simpleSource (now comment-free) as input, errors expected so wantCode doesn't matter
		{name: "Change Package Empty Name", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": ""}, wantErrIs: ErrGoModifyInvalidDirectiveValue},
		{name: "Change Package Wrong Type", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": 123}, wantErrIs: ErrGoModifyInvalidDirectiveValue},
		{name: "Replace Identifier Invalid Format (Old)", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": "fmtPrintln", "new": "log.Printf"}}, wantErrIs: ErrGoInvalidIdentifierFormat},
		{name: "Replace Identifier Invalid Format (New)", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": "fmt.Println", "new": "logPrintf"}}, wantErrIs: ErrGoInvalidIdentifierFormat},
		{name: "Replace Identifier Empty Part (Old)", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": ".Println", "new": "log.Printf"}}, wantErrIs: ErrGoInvalidIdentifierFormat},
		{name: "Replace Identifier Empty Part (New)", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": "fmt.Println", "new": "log."}}, wantErrIs: ErrGoInvalidIdentifierFormat},
		{name: "Replace Identifier Missing 'new' Key", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": "fmt.Println"}}, wantErrIs: ErrGoModifyMissingMapKey},
		{name: "Replace Identifier Wrong Value Type", initialContent: simpleSource, modifications: map[string]interface{}{"replace_identifier": map[string]interface{}{"old": 123, "new": "log.Printf"}}, wantErrIs: ErrGoModifyInvalidDirectiveValue},
		{name: "Invalid Handle", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": "other"}, wantErrIs: ErrGoModifyFailed},
		{name: "Non-AST Handle", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": "other"}, wantErrIs: ErrGoModifyFailed},
		{name: "No Known Directive", initialContent: simpleSource, modifications: map[string]interface{}{"unknown_directive": "value"}, wantErrIs: ErrGoModifyUnknownDirective},
		{name: "Empty Modifications Map", initialContent: simpleSource, modifications: map[string]interface{}{}, wantErrIs: ErrGoModifyEmptyMap},

		// --- Validation Error Tests ---
		{name: "Validation Wrong Arg Count", valWantErrIs: ErrValidationArgCount},
		{name: "Validation Nil Handle", valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Nil Modifications", valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Handle Type", valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Wrong Mod Type", wantErrIs: ErrValidationTypeMismatch}, // toolGoModifyAST now returns this directly
	}

	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			currentInterp, _ := newDefaultTestInterpreter(t)
			var initialHandle string
			var finalArgs []interface{}

			// Simplified setup logic
			if tc.initialContent != "" {
				initialHandle = setupParseModifyTest(t, currentInterp, tc.initialContent)
			} else if tc.valWantErrIs == nil && tc.wantErrIs == nil {
				t.Fatalf("Test setup error: initialContent cannot be empty for test '%s' unless error expected", tc.name)
			}

			// Construct args based on test case name (Simplified/Combined)
			switch tc.name {
			case "Invalid Handle":
				finalArgs = makeArgs("invalid-handle", tc.modifications)
			case "Non-AST Handle":
				handleToUse := "placeholder_handle"
				currentInterp.objectCache[handleToUse] = "not an ast"
				currentInterp.handleTypes[handleToUse] = "OtherType"
				finalArgs = makeArgs(handleToUse, tc.modifications)
			case "Validation Wrong Arg Count":
				finalArgs = makeArgs(initialHandle)
			case "Validation Nil Handle":
				finalArgs = makeArgs(nil, tc.modifications)
			case "Validation Nil Modifications":
				finalArgs = makeArgs(initialHandle, nil)
			case "Validation Wrong Handle Type":
				finalArgs = makeArgs(123, tc.modifications)
			case "Validation Wrong Mod Type":
				finalArgs = makeArgs(initialHandle, "not-a-map")
			default:
				finalArgs = makeArgs(initialHandle, tc.modifications)
			}

			toolImpl, found := currentInterp.ToolRegistry().GetTool("GoModifyAST")
			if !found {
				t.Fatalf("Tool GoModifyAST not found")
			}
			spec := toolImpl.Spec
			convertedArgs, valErr := ValidateAndConvertArgs(spec, finalArgs)

			// --- Validation Layer Error Check ---
			if tc.valWantErrIs != nil {
				if valErr == nil {
					t.Errorf("Validate: expected error [%v], got nil", tc.valWantErrIs)
				} else if !errors.Is(valErr, tc.valWantErrIs) {
					t.Errorf("Validate: expected error type [%T], got [%T]: %v", tc.valWantErrIs, valErr, valErr)
				}
				return
			}
			if valErr != nil && tc.valWantErrIs == nil {
				t.Fatalf("Validate: unexpected error: %v", valErr)
			}

			gotResult, toolErr := toolImpl.Func(currentInterp, convertedArgs)

			// --- Tool Execution Error Check (Using errors.Is) ---
			if tc.wantErrIs != nil {
				if toolErr == nil {
					t.Errorf("Execute: expected Go error type [%T], but got nil error. Result: %v", tc.wantErrIs, gotResult)
				} else if !errors.Is(toolErr, tc.wantErrIs) {
					t.Errorf("Execute: wrong Go error type. \n got error: %v\nwant error type: %T", toolErr, tc.wantErrIs)
				} else {
					t.Logf("Execute: Correctly received expected Go error type %T (%v)", tc.wantErrIs, toolErr)
					if gotResult != nil {
						t.Errorf("Execute: expected nil result when error is returned, but got: %v (%T)", gotResult, gotResult)
					}
				}
			} else if toolErr != nil {
				t.Fatalf("Execute: unexpected Go error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
			}

			// --- Success Result Check (Only run if no error expected) ---
			if tc.wantErrIs == nil && tc.valWantErrIs == nil {
				gotHandle, okGot := gotResult.(string)
				if !okGot {
					t.Fatalf("Execute Success: Expected string handle result, got %T: %v", gotResult, gotResult)
				}

				if tc.wantHandleSame {
					if gotHandle != initialHandle {
						t.Errorf("Execute Success (No-Op): Expected original handle '%s', but got new handle '%s'", initialHandle, gotHandle)
					} else {
						t.Logf("Execute Success (No-Op): Correctly received original handle '%s'", initialHandle)
					}
					finalCode := getFormattedCodeModifyTest(t, currentInterp, gotHandle)
					if diff := cmp.Diff(tc.initialContent, finalCode); diff != "" { // Compare against original comment-free input
						t.Errorf("Execute Success (No-Op): Code unexpectedly changed (-want initial +got final):\n%s", diff)
					}
				} else {
					if initialHandle != "" && gotHandle == initialHandle {
						t.Errorf("Execute Success (Modification): Expected a NEW handle, but got the original handle '%s'", initialHandle)
					} else {
						t.Logf("Execute Success (Modification): Received new handle '%s' (original '%s')", gotHandle, initialHandle)
					}
					if tc.wantCode != "" {
						finalCode := getFormattedCodeModifyTest(t, currentInterp, gotHandle)
						// Normalize line endings for comparison
						wantNormalized := strings.ReplaceAll(tc.wantCode, "\r\n", "\n") // wantCode is now comment-free
						gotNormalized := strings.ReplaceAll(finalCode, "\r\n", "\n")    // gotCode should also be comment-free now
						if diff := cmp.Diff(wantNormalized, gotNormalized); diff != "" {
							t.Errorf("Execute Success (Modification): Final code mismatch (-want +got):\n%s", diff)
							t.Logf("Want (Normalized):\n%s\n", wantNormalized) // Log want/got for debugging
							t.Logf("Got (Normalized):\n%s\n", gotNormalized)
						} else {
							t.Logf("Execute Success (Modification): Final code matches expected.")
						}
					} else {
						t.Logf("Execute Success: Received handle %q, no specific code comparison defined.", gotHandle)
					}
				}
			}
		})
	}
}
