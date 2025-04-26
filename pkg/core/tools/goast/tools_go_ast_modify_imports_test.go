// filename: pkg/core/tools_go_ast_modify_imports_test.go
// UPDATED: Expect specific errors using errors.Is, removed wantResultMsg
package goast

import (
	"errors" // Now needed for errors.Is
	// "strings" // No longer needed for error checks
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/go-cmp/cmp"
)

// Assume testFsToolHelper, core.NewDefaultTestInterpreter, core.MakeArgs, etc. are available from testing_helpers_test.go
// Assume golangASTTypeTag, CachedAst are defined in tools_go_ast.go

// Helper to get formatted string from handle (duplicated for focused testing)
func getFormattedCodeModifyImportsTest(t *testing.T, interp *core.Interpreter, handleID string) string {
	t.Helper()
	res, err := toolGoFormatAST(interp, core.MakeArgs(handleID))
	if err != nil {
		t.Fatalf("getFormattedCodeModifyImportsTest: toolGoFormatAST failed for handle %s: %v", handleID, err)
	}
	code, ok := res.(string)
	if !ok {
		t.Fatalf("getFormattedCodeModifyImportsTest: toolGoFormatAST did not return a string, got %T", res)
	}
	return code
}

// Helper to parse code and return a handle (duplicated for focused testing)
func setupParseModifyImportsTest(t *testing.T, interp *core.Interpreter, content string) string {
	t.Helper()
	handleID, err := toolGoParseFile(interp, core.MakeArgs(nil, content))
	if err != nil {
		t.Fatalf("setupParseModifyImportsTest: toolGoParseFile failed: %v", err)
	}
	handleStr, ok := handleID.(string)
	if !ok || handleStr == "" {
		t.Fatalf("setupParseModifyImportsTest: toolGoParseFile did not return a valid handle string, got %T: %v", handleID, handleID)
	}
	return handleStr
}

// --- TestGoModifyASTImports Function ---
func TestToolGoModifyASTImports(t *testing.T) {

	// --- Test Data ---
	simpleSource := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
`
	simpleSourceAddImport_Want := `package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello")
}
`
	simpleSourceNoFmt_Want := `package main

func main() {
	fmt.Println("hello")
}
`
	simpleSourceReplaceFmt_Want := `package main

import "new/fmt/path"

func main() {
	fmt.Println("hello")
}
`
	simpleSourceReplaceFmtWithAlias_Want := `package main

import x "new/fmt/path"

func main() { x.Println("hi") }
`
	simpleSourceNewPkgAddImport_Want := `package other

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("hello")
}
`

	tests := []struct {
		name           string
		initialContent string
		modifications  map[string]interface{}
		wantCode       string // Expected formatted code after modification
		wantHandleSame bool   // True if no change expected, should return original handle
		wantErrIs      error  // Expected underlying error type (use errors.Is)
	}{
		// --- Add Import Tests ---
		{name: "Add Import", initialContent: simpleSource, modifications: map[string]interface{}{"add_import": "os"}, wantCode: simpleSourceAddImport_Want},
		{name: "Add Existing Import", initialContent: simpleSource, modifications: map[string]interface{}{"add_import": "fmt"}, wantCode: simpleSource, wantHandleSame: true},
		{name: "Add Import Empty Path", initialContent: simpleSource, modifications: map[string]interface{}{"add_import": ""}, wantErrIs: core.ErrGoModifyInvalidDirectiveValue},

		// --- Remove Import Tests ---
		{name: "Remove Import", initialContent: simpleSource, modifications: map[string]interface{}{"remove_import": "fmt"}, wantCode: simpleSourceNoFmt_Want},
		{name: "Remove Non-Existent Import", initialContent: simpleSource, modifications: map[string]interface{}{"remove_import": "nosuchpkg"}, wantCode: simpleSource, wantHandleSame: true},
		{name: "Remove Import Empty Path", initialContent: simpleSource, modifications: map[string]interface{}{"remove_import": ""}, wantErrIs: core.ErrGoModifyInvalidDirectiveValue},

		// --- Replace Import Tests ---
		{name: "Replace Import", initialContent: simpleSource, modifications: map[string]interface{}{"replace_import": map[string]interface{}{"old_path": "fmt", "new_path": "new/fmt/path"}}, wantCode: simpleSourceReplaceFmt_Want},
		{name: "Replace Import With Alias", initialContent: `package main; import x "fmt"; func main(){ x.Println("hi") }`, modifications: map[string]interface{}{"replace_import": map[string]interface{}{"old_path": "fmt", "new_path": "new/fmt/path"}}, wantCode: simpleSourceReplaceFmtWithAlias_Want},
		{name: "Replace Non-Existent Import", initialContent: simpleSource, modifications: map[string]interface{}{"replace_import": map[string]interface{}{"old_path": "nosuchpkg", "new_path": "new/nosuchpkg"}}, wantCode: simpleSource, wantHandleSame: true},
		{name: "Replace Import Missing Keys", initialContent: simpleSource, modifications: map[string]interface{}{"replace_import": map[string]interface{}{"old_path": "fmt"}}, wantErrIs: core.ErrGoModifyMissingMapKey},
		{name: "Replace Import Empty Path (Old)", initialContent: simpleSource, modifications: map[string]interface{}{"replace_import": map[string]interface{}{"old_path": "", "new_path": "new/path"}}, wantErrIs: core.ErrGoModifyInvalidDirectiveValue},
		{name: "Replace Import Empty Path (New)", initialContent: simpleSource, modifications: map[string]interface{}{"replace_import": map[string]interface{}{"old_path": "fmt", "new_path": ""}}, wantErrIs: core.ErrGoModifyInvalidDirectiveValue},

		// --- Combined Test ---
		{name: "Change Package and Add Import", initialContent: simpleSource, modifications: map[string]interface{}{"change_package": "other", "add_import": "os"}, wantCode: simpleSourceNewPkgAddImport_Want},
	}

	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			currentInterp, _ := core.NewDefaultTestInterpreter(t)
			initialHandle := setupParseModifyImportsTest(t, currentInterp, tc.initialContent)
			finalArgs := core.MakeArgs(initialHandle, tc.modifications)

			toolImpl, found := currentInterp.ToolRegistry().GetTool("GoModifyAST")
			if !found {
				t.Fatalf("Tool GoModifyAST not found")
			}

			gotResult, toolErr := toolImpl.Func(currentInterp, finalArgs)

			// --- UPDATEDcore.Error Checking Logic ---
			if tc.wantErrIs != nil {
				// Case 1: Expecting a specific Go error type from the tool
				if toolErr == nil {
					t.Errorf("Execute: expected Go error type [%T], but got nil error. Result: %v", tc.wantErrIs, gotResult)
				} else if !errors.Is(toolErr, tc.wantErrIs) {
					t.Errorf("Execute: wrong Go error type. \n got error: %v\nwant error type: %T", toolErr, tc.wantErrIs)
				} else {
					t.Logf("Execute: Correctly received expected Go error type %T (%v)", tc.wantErrIs, toolErr)
					// Check that result is nil when error is expected
					if gotResult != nil {
						t.Errorf("Execute: expected nil result when error is returned, but got: %v (%T)", gotResult, gotResult)
					}
				}
			} else if toolErr != nil {
				// Case 2: Unexpected Go error from the tool
				t.Fatalf("Execute: unexpected Go error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
			}
			// --- END UPDATEDcore.Error Checking ---

			// --- Success Result Check (Only run if no error expected) ---
			if tc.wantErrIs == nil {
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
					finalCode := getFormattedCodeModifyImportsTest(t, currentInterp, gotHandle)
					if diff := cmp.Diff(tc.initialContent, finalCode); diff != "" {
						t.Errorf("Execute Success (No-Op): Code unexpectedly changed (-want initial +got final):\n%s", diff)
					}
				} else {
					if initialHandle != "" && gotHandle == initialHandle {
						t.Errorf("Execute Success (Modification): Expected a NEW handle, but got the original handle '%s'", initialHandle)
					} else {
						t.Logf("Execute Success (Modification): Received new handle '%s' (original '%s')", gotHandle, initialHandle)
					}
					if tc.wantCode != "" {
						finalCode := getFormattedCodeModifyImportsTest(t, currentInterp, gotHandle)
						if diff := cmp.Diff(tc.wantCode, finalCode); diff != "" {
							t.Errorf("Execute Success (Modification): Final code mismatch (-want +got):\n%s", diff)
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
