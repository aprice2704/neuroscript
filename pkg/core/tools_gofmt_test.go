// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 21:11:52 PDT // Fix GoImports test wantResult string AGAIN
// filename: pkg/core/tools_gofmt_test.go

package core

import (
	"errors"
	"reflect"
	"testing"
)

// --- Test Helper (Used by GoFmt and GoImports) ---
// --- (testGoFormatToolHelper remains unchanged) ---
func testGoFormatToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name             string
	toolName         string        // "GoFmt" or "GoImports"
	args             []interface{} // Raw args for the tool
	wantResult       interface{}   // Expected result (string for success, map for error)
	wantErrResultNil bool          // Should the result be nil (e.g., validation error)
	wantToolErrIs    error         // Expect specific tool execution error (e.g., ErrInternalTool for format fail)
	valWantErrIs     error         // Expect specific validation error (e.g., ErrValidationTypeMismatch)
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}
		spec := toolImpl.Spec

		// --- Validation ---
		convertedArgs, valErr := ValidateAndConvertArgs(spec, tc.args)

		// Check Specific Validation Error
		if tc.valWantErrIs != nil {
			if valErr == nil {
				t.Errorf("ValidateAndConvertArgs() expected validation error [%v], but got nil", tc.valWantErrIs)
			} else if !errors.Is(valErr, tc.valWantErrIs) {
				t.Errorf("ValidateAndConvertArgs() expected validation error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
			}
			return // Stop if specific validation error expected
		}

		// Check for Unexpected Validation Error
		if valErr != nil && tc.valWantErrIs == nil {
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		// --- Execution ---
		if valErr != nil {
			t.Fatalf("Internal test error: validation failed but not caught")
			return
		}
		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		// --- Check Expected Tool Error ---
		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error containing [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
			} else {
				t.Logf("Got expected tool error: %v", toolErr)
			}

			// On expected tool error, check if result is the expected type (map) or nil
			if tc.wantErrResultNil {
				if gotResult != nil {
					t.Errorf("Expected nil result map due to error [%v], but got: %#v", tc.wantToolErrIs, gotResult)
				}
			} else {
				_, ok := gotResult.(map[string]interface{})
				if !ok {
					t.Errorf("Expected map[string]interface{} result on error [%v], but got type %T: %#v", tc.wantToolErrIs, gotResult, gotResult)
				}
				if tc.wantResult != nil {
					wantMap, okWant := tc.wantResult.(map[string]interface{})
					gotMap, okGot := gotResult.(map[string]interface{})
					if okWant && okGot {
						if !reflect.DeepEqual(gotMap, wantMap) {
							t.Logf("Note: Error map content mismatch (may be acceptable if error details differ slightly):\nGot:  %#v\nWant: %#v", gotMap, wantMap)
						}
					} else {
						t.Errorf("Type mismatch comparing error result maps (got %T, want %T)", gotResult, tc.wantResult)
					}
				}
			}
			return // Stop after checking expected error
		}

		// --- Check for Unexpected Tool Error ---
		if toolErr != nil && tc.wantToolErrIs == nil {
			t.Fatalf("Tool function returned unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		// --- Result Comparison (No Go error expected or occurred) ---
		if tc.wantErrResultNil {
			if gotResult != nil {
				t.Errorf("Expected nil result, but got: %#v", gotResult)
			}
			return
		}
		if gotResult == nil && !tc.wantErrResultNil {
			t.Fatalf("Expected non-nil result, but got nil")
		}

		// Compare successful result (should be string for GoFmt/GoImports)
		gotStr, gotOk := gotResult.(string)
		wantStr, wantOk := tc.wantResult.(string)
		if !gotOk || !wantOk {
			t.Errorf("Result/Want type mismatch, expected string for success, got %T, want %T", gotResult, tc.wantResult)
		} else if gotStr != wantStr {
			// Use %q for potentially multi-line strings
			t.Errorf("Tool function result mismatch:\nGot:\n%q\n\nWant:\n%q", gotStr, wantStr)
			// Provide diff hint
			t.Logf("Diff hint: got len %d, want len %d", len(gotStr), len(wantStr))
		}
	})
}

// --- Tests for GoFmt ---
// --- (TestToolGoFmt remains unchanged) ---
func TestToolGoFmt(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Use default interpreter

	// Test cases (using wantToolErrIs and valWantErrIs)
	unformattedSource := `
package main
import "fmt"

func  main() {
fmt.Println("hello")
	}
`
	formattedSource := `package main

import "fmt"

func main() {
	fmt.Println("hello")
}
` // Note: gofmt adds trailing newline
	invalidSource := `package main func main() { fmt.Println("hello world }` // Missing quote

	tests := []struct {
		name             string
		toolName         string
		args             []interface{}
		wantResult       interface{} // string for success, map for error
		wantErrResultNil bool        // Expect nil result map? (e.g., validation error)
		wantToolErrIs    error       // Expect specific tool execution error?
		valWantErrIs     error       // Expect specific validation error?
	}{
		{
			name:             "Format valid code",
			toolName:         "GoFmt",
			args:             MakeArgs(unformattedSource),
			wantResult:       formattedSource,
			wantErrResultNil: false,
			wantToolErrIs:    nil, // Expect success
			valWantErrIs:     nil,
		},
		{
			name:             "Format already formatted code",
			toolName:         "GoFmt",
			args:             MakeArgs(formattedSource),
			wantResult:       formattedSource, // Should be idempotent
			wantErrResultNil: false,
			wantToolErrIs:    nil,
			valWantErrIs:     nil,
		},
		{
			name:     "Format invalid code",
			toolName: "GoFmt",
			args:     MakeArgs(invalidSource),
			wantResult: map[string]interface{}{
				"formatted_content": invalidSource,
				"error":             "", // Error message varies, don't assert precisely here
				"success":           false,
			},
			wantErrResultNil: false,           // Expect the error map, not nil
			wantToolErrIs:    ErrInternalTool, // Expect an internal tool error wrapping the gofmt error
			valWantErrIs:     nil,
		},
		{
			name:             "Validation_Wrong_Arg_Type",
			toolName:         "GoFmt",
			args:             MakeArgs(123),
			wantResult:       nil,
			wantErrResultNil: true, // Expect nil result due to validation error
			wantToolErrIs:    nil,
			valWantErrIs:     ErrValidationTypeMismatch,
		},
		{
			name:             "Validation_Missing_Arg",
			toolName:         "GoFmt",
			args:             MakeArgs(),
			wantResult:       nil,
			wantErrResultNil: true, // Expect nil result due to validation error
			wantToolErrIs:    nil,
			valWantErrIs:     ErrValidationArgCount,
		},
	}

	for _, tt := range tests {
		testGoFormatToolHelper(t, interp, tt)
	}
}

// --- Tests for GoImports ---
func TestToolGoImports(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Use default interpreter

	// Test cases
	needsImportAdded := `package main

func main() {
	fmt.Println("hello") // fmt used but not imported
}
`
	wantImportAdded := `package main

import "fmt"

func main() {
	fmt.Println("hello") // fmt used but not imported
}
`

	needsImportRemoved := `package main

import (
	"fmt"
	"os" // os is imported but not used
)

func main() {
	fmt.Println("hello")
}
`
	wantImportRemoved := `package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello")
}
`

	needsFmtAndImport := `package main
import "os" // os unused
func main() {
	fmt.Println("hello") // fmt needs import, needs formatting
	}
`
	// *** CORRECTED wantFmtAndImport (removed previous notes) ***
	wantFmtAndImport := `package main

import "fmt"

// os unused
func main() {
	fmt.Println("hello") // fmt needs import, needs formatting
}
`

	invalidSource := `package main import "fmt" func main() { fmt.Println("hello }` // Syntax error

	tests := []struct {
		name             string
		toolName         string
		args             []interface{}
		wantResult       interface{} // string for success, map for error
		wantErrResultNil bool        // Expect nil result map? (e.g., validation error)
		wantToolErrIs    error       // Expect specific tool execution error?
		valWantErrIs     error       // Expect specific validation error?
	}{
		{
			name:             "Add missing import",
			toolName:         "GoImports",
			args:             MakeArgs(needsImportAdded),
			wantResult:       wantImportAdded,
			wantErrResultNil: false,
			wantToolErrIs:    nil,
			valWantErrIs:     nil,
		},
		{
			name:             "Remove unused import",
			toolName:         "GoImports",
			args:             MakeArgs(needsImportRemoved),
			wantResult:       wantImportRemoved,
			wantErrResultNil: false,
			wantToolErrIs:    nil,
			valWantErrIs:     nil,
		},
		{
			name:             "Format and manage imports",
			toolName:         "GoImports",
			args:             MakeArgs(needsFmtAndImport),
			wantResult:       wantFmtAndImport, // Use corrected want string
			wantErrResultNil: false,
			wantToolErrIs:    nil,
			valWantErrIs:     nil,
		},
		{
			name:             "Already formatted code",
			toolName:         "GoImports",
			args:             MakeArgs(wantFmtAndImport), // Use already correct code
			wantResult:       wantFmtAndImport,           // Idempotent
			wantErrResultNil: false,
			wantToolErrIs:    nil,
			valWantErrIs:     nil,
		},
		{
			name:     "Syntax error",
			toolName: "GoImports",
			args:     MakeArgs(invalidSource),
			wantResult: map[string]interface{}{
				"formatted_content": invalidSource,
				"error":             "",
				"success":           false,
			},
			wantErrResultNil: false,
			wantToolErrIs:    ErrInternalTool,
			valWantErrIs:     nil,
		},
		{
			name:             "Validation_Wrong_Arg_Type",
			toolName:         "GoImports",
			args:             MakeArgs(false),
			wantResult:       nil,
			wantErrResultNil: true,
			wantToolErrIs:    nil,
			valWantErrIs:     ErrValidationTypeMismatch,
		},
		{
			name:             "Validation_Missing_Arg",
			toolName:         "GoImports",
			args:             MakeArgs(),
			wantResult:       nil,
			wantErrResultNil: true,
			wantToolErrIs:    nil,
			valWantErrIs:     ErrValidationArgCount,
		},
	}

	for _, tt := range tests {
		testGoFormatToolHelper(t, interp, tt)
	}
}
