// filename: pkg/core/tools_gofmt_test.go
package core

import (
	"errors" // Import errors
	// "strings" // Remove unused import
	"testing"
)

// Define a local test helper specific to GoFmt or adapt the general one
// For simplicity, let's adapt the general one here.
// testGoFmtToolHelper definition (copied from list test, adapt if needed)
func testGoFmtToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{}
	wantToolErrIs error // Expect specific tool error (or nil for success)
	valWantErrIs  error // Expect specific validation error
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) { // Add t.Run for subtests
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
				t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
			} else if !errors.Is(valErr, tc.valWantErrIs) {
				t.Errorf("ValidateAndConvertArgs() expected error type [%v], but got type [%T] with value: %v", tc.valWantErrIs, valErr, valErr)
			}
			return // Stop if specific validation error expected
		}

		// Check for Unexpected Validation Error
		if valErr != nil && tc.valWantErrIs == nil {
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		// --- Execution ---
		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		// Check Specific Tool Error
		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
			}
			// If an error was expected, don't compare result values directly.
			// For GoFmt errors, the result might be the original source or an error message,
			// but we prioritize checking the error type here.
			return
		}

		// Check for Unexpected Tool Error
		if toolErr != nil && tc.wantToolErrIs == nil {
			t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		// --- Result Comparison ---
		if tc.wantToolErrIs == nil { // Only compare if no specific tool error expected
			// GoFmt returns string result or error
			gotStr, gotOk := gotResult.(string)
			wantStr, wantOk := tc.wantResult.(string)
			if !gotOk || !wantOk {
				t.Errorf("Result/Want type mismatch, expected string, got %T, want %T", gotResult, tc.wantResult)
			} else if gotStr != wantStr {
				t.Errorf("Tool function result mismatch:\nGot:\n%s\n\nWant:\n%s", gotStr, wantStr)
			}
		}
	})
}

func TestToolGoFmt(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)

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
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{
			name:       "Format valid code",
			toolName:   "GoFmt",
			args:       MakeArgs(unformattedSource),
			wantResult: formattedSource,
		},
		{
			name:       "Format already formatted code",
			toolName:   "GoFmt",
			args:       MakeArgs(formattedSource),
			wantResult: formattedSource, // Should be idempotent
		},
		{
			name:          "Format invalid code",
			toolName:      "GoFmt",
			args:          MakeArgs(invalidSource),
			wantResult:    invalidSource,   // Should return original source on format error
			wantToolErrIs: ErrInternalTool, // Expect an internal tool error wrapping the gofmt error
		},
		{
			name:         "Validation_Wrong_Arg_Type",
			toolName:     "GoFmt",
			args:         MakeArgs(123),
			valWantErrIs: ErrValidationTypeMismatch,
		},
		{
			name:         "Validation_Missing_Arg",
			toolName:     "GoFmt",
			args:         MakeArgs(),
			valWantErrIs: ErrValidationArgCount,
		},
	}

	for _, tt := range tests {
		testGoFmtToolHelper(t, interp, tt) // Use the adapted helper
	}
}
