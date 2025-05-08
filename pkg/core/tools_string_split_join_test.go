// filename: pkg/core/tools_string_split_join_test.go
package core

import (
	"errors" // Import errors
	"reflect"

	// "strings" // No longer needed
	"testing"
)

// Adapt the general test helper logic (used in list tests) for string split/join tools
func testStringSplitJoinToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{} // Expected result *if* no error
	wantToolErrIs error       // Specific Go error expected *from the tool function*
	valWantErrIs  error       // Specific Go error expected *from validation*
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
			// Regardless of match details, if specific error was expected, stop.
			return
		}

		// Check for Unexpected Validation Error
		if valErr != nil && tc.valWantErrIs == nil { // Check if validation error occurred when none was expected
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		// --- Execution (Only if validation passed and wasn't expected to fail) ---
		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		// Check Specific Tool Error
		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
			}
			// If specific tool error was expected, don't check result
			return
		}

		// Check for Unexpected Tool Error
		if toolErr != nil && tc.wantToolErrIs == nil {
			t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		// --- Result Comparison (only if no errors occurred or were expected via wantToolErrIs) ---
		if tc.wantToolErrIs == nil { // Only compare results if no specific tool error was expected
			// Special case for []string results from SplitString
			if tc.toolName == "TOOL.SplitString" || tc.toolName == "TOOL.SplitWords" {
				wantSlice, wantOk := tc.wantResult.([]string)
				// *** FIXED ASSIGNMENT: Capture third return value (error) with blank identifier ***
				gotSlice, gotOk, _ := ConvertToSliceOfString(gotResult) // Use exported helper, ignore error return
				if !wantOk {
					t.Fatalf("WantResult for %s test is not []string", tc.toolName)
				}
				if !gotOk {
					// Log the actual error from ConvertToSliceOfString if it occurred unexpectedly
					_, _, convErr := ConvertToSliceOfString(gotResult)
					t.Errorf("GotResult for %s test is not convertible to []string, got %T. Conversion error: %v", tc.toolName, gotResult, convErr)
				} else if !reflect.DeepEqual(gotSlice, wantSlice) {
					t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
						gotSlice, gotSlice, wantSlice, wantSlice)
				}
			} else if !reflect.DeepEqual(gotResult, tc.wantResult) { // Default comparison
				t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	})
}

func TestToolSplitString(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Ignore sandboxDir
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{} // Expected: []string
		wantToolErrIs error
		valWantErrIs  error
	}{
		// *** FIXED toolName prefix ***
		{name: "Simple Split", toolName: "Split", args: MakeArgs("a,b,c", ","), wantResult: []string{"a", "b", "c"}},
		{name: "Split With Spaces", toolName: "Split", args: MakeArgs(" a , b , c ", ","), wantResult: []string{" a ", " b ", " c "}},
		{name: "Multi-char Delimiter", toolName: "Split", args: MakeArgs("one<>two<>three", "<>"), wantResult: []string{"one", "two", "three"}},
		{name: "Leading Delimiter", toolName: "Split", args: MakeArgs(",a,b", ","), wantResult: []string{"", "a", "b"}},
		{name: "Trailing Delimiter", toolName: "Split", args: MakeArgs("a,b,", ","), wantResult: []string{"a", "b", ""}},
		{name: "Only Delimiter", toolName: "Split", args: MakeArgs(",", ","), wantResult: []string{"", ""}},
		{name: "Empty String", toolName: "Split", args: MakeArgs("", ","), wantResult: []string{""}},
		{name: "Empty Delimiter", toolName: "Split", args: MakeArgs("abc", ""), wantResult: []string{"a", "b", "c"}}, // Splits between UTF-8 chars
		{name: "No Delimiter Found", toolName: "Split", args: MakeArgs("abc", ","), wantResult: []string{"abc"}},
		{name: "Non-string Input", toolName: "Split", args: MakeArgs(123, ","), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Non-string Delimiter", toolName: "Split", args: MakeArgs("abc", 1), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Nil Input", toolName: "Split", args: MakeArgs(nil, ","), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil Delimiter", toolName: "Split", args: MakeArgs("abc", nil), valWantErrIs: ErrValidationRequiredArgNil},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolSplitWords(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t) // Ignore sandboxDir
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{} // Expected: []string
		wantToolErrIs error
		valWantErrIs  error
	}{
		// *** FIXED toolName prefix ***
		{name: "Simple Words", toolName: "SplitWords", args: MakeArgs("hello world"), wantResult: []string{"hello", "world"}},
		{name: "Multiple Spaces", toolName: "SplitWords", args: MakeArgs("  hello \t world  \n next"), wantResult: []string{"hello", "world", "next"}},
		{name: "Leading/Trailing Space", toolName: "SplitWords", args: MakeArgs(" hello "), wantResult: []string{"hello"}},
		{name: "Punctuation", toolName: "SplitWords", args: MakeArgs("hello, world!"), wantResult: []string{"hello,", "world!"}},
		{name: "Empty String", toolName: "SplitWords", args: MakeArgs(""), wantResult: []string{}},
		{name: "Only Whitespace", toolName: "SplitWords", args: MakeArgs(" \t \n "), wantResult: []string{}},
		{name: "Non-string Input", toolName: "SplitWords", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Nil Input", toolName: "SplitWords", args: MakeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolJoinStrings(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)   // Ignore sandboxDir
	stringSlice := []interface{}{"a", "b", "c"} // Use []interface{} for input arg
	mixedSlice := []interface{}{"a", int64(1), true}
	numSlice := []interface{}{int64(1), float64(2.5), int64(3)}

	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		// *** FIXED toolName prefix ***
		{name: "Join Simple", toolName: "Join", args: MakeArgs(stringSlice, ","), wantResult: "a,b,c"},
		{name: "Join Empty Sep", toolName: "Join", args: MakeArgs(stringSlice, ""), wantResult: "abc"},
		{name: "Join Single Elem", toolName: "Join", args: MakeArgs([]interface{}{"a"}, ","), wantResult: "a"},
		{name: "Join Empty Slice", toolName: "Join", args: MakeArgs([]interface{}{}, ","), wantResult: ""},
		{name: "Join Mixed Types", toolName: "Join", args: MakeArgs(mixedSlice, "-"), wantResult: "a-1-true"}, // Converts elements to string
		{name: "Join Numeric Types", toolName: "Join", args: MakeArgs(numSlice, " "), wantResult: "1 2.5 3"},
		{name: "Non-slice First Arg (Validation Err)", toolName: "Join", args: MakeArgs("abc", ","), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Non-string Separator (Validation Err)", toolName: "Join", args: MakeArgs(stringSlice, 123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Nil Slice", toolName: "Join", args: MakeArgs(nil, ","), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil Separator", toolName: "Join", args: MakeArgs(stringSlice, nil), valWantErrIs: ErrValidationRequiredArgNil},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}
