// NeuroScript Version: 0.3.1
// File version: 0.1.3
// Correct Join validation expectations. Fix result comparison for Split/SplitWords.
// nlines: 150
// risk_rating: MEDIUM
// filename: pkg/core/tools_string_split_join_test.go
package core

import (
	"errors" // Import errors
	"reflect"

	// "strings" // No longer needed
	"testing"
)

// Assume NewDefaultTestInterpreter and MakeArgs are defined in testing_helpers.go or similar

// testStringSplitJoinToolHelper encapsulates the logic for validating and executing string split/join tools.
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
			return // Stop if validation failed as expected
		}
		if valErr != nil && tc.valWantErrIs == nil {
			t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
		}

		// --- Execution (Only if Validation Passed) ---
		gotResult, toolErr := toolImpl.Func(interp, convertedArgs)

		// Check Specific Tool Error
		if tc.wantToolErrIs != nil {
			if toolErr == nil {
				t.Errorf("Tool function expected error [%v], but got nil. Result: %v (%T)", tc.wantToolErrIs, gotResult, gotResult)
			} else if !errors.Is(toolErr, tc.wantToolErrIs) {
				t.Errorf("Tool function expected error type [%v], but got type [%T] with value: %v", tc.wantToolErrIs, toolErr, toolErr)
			}
			return // Stop if tool error occurred as expected
		}
		if toolErr != nil && tc.wantToolErrIs == nil {
			t.Fatalf("Tool function unexpected error: %v. Result: %v (%T)", toolErr, gotResult, gotResult)
		}

		// --- Result Comparison (Only if No Errors Expected/Occurred) ---
		// Corrected: Handle []string return type specifically for Split/SplitWords
		switch tc.toolName {
		case "Split", "SplitWords":
			wantSlice, wantOk := tc.wantResult.([]string)
			gotSlice, gotOk := gotResult.([]string) // Tool now returns []string
			if !wantOk {
				t.Fatalf("WantResult for %s test is not []string", tc.toolName)
			}
			if !gotOk {
				t.Errorf("GotResult for %s test is not []string, got %T", tc.toolName, gotResult)
			} else if !reflect.DeepEqual(gotSlice, wantSlice) {
				t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotSlice, gotSlice, wantSlice, wantSlice)
			}
		default: // Default comparison for Join (returns string)
			if !reflect.DeepEqual(gotResult, tc.wantResult) {
				t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	})
}

func TestToolSplitString(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{} // Expected: []string
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple_Split", toolName: "Split", args: MakeArgs("a,b,c", ","), wantResult: []string{"a", "b", "c"}},
		{name: "Split_With_Spaces", toolName: "Split", args: MakeArgs(" a , b , c ", ","), wantResult: []string{" a ", " b ", " c "}},
		{name: "Multi-char_Delimiter", toolName: "Split", args: MakeArgs("one<>two<>three", "<>"), wantResult: []string{"one", "two", "three"}},
		{name: "Leading_Delimiter", toolName: "Split", args: MakeArgs(",a,b", ","), wantResult: []string{"", "a", "b"}},
		{name: "Trailing_Delimiter", toolName: "Split", args: MakeArgs("a,b,", ","), wantResult: []string{"a", "b", ""}},
		{name: "Only_Delimiter", toolName: "Split", args: MakeArgs(",", ","), wantResult: []string{"", ""}},
		{name: "Empty_String", toolName: "Split", args: MakeArgs("", ","), wantResult: []string{""}},
		{name: "Empty_Delimiter", toolName: "Split", args: MakeArgs("abc", ""), wantResult: []string{"a", "b", "c"}}, // Splits between UTF-8 chars
		{name: "No_Delimiter_Found", toolName: "Split", args: MakeArgs("abc", ","), wantResult: []string{"abc"}},
		{name: "Validation_Non-string_Input", toolName: "Split", args: MakeArgs(123, ","), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-string_Delimiter", toolName: "Split", args: MakeArgs("abc", 1), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Nil_Input", toolName: "Split", args: MakeArgs(nil, ","), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_Delimiter", toolName: "Split", args: MakeArgs("abc", nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Missing_Delimiter", toolName: "Split", args: MakeArgs("abc"), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolSplitWords(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{} // Expected: []string
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple_Words", toolName: "SplitWords", args: MakeArgs("hello world"), wantResult: []string{"hello", "world"}},
		{name: "Multiple_Spaces", toolName: "SplitWords", args: MakeArgs("  hello \t world  \n next"), wantResult: []string{"hello", "world", "next"}},
		{name: "Leading/Trailing_Space", toolName: "SplitWords", args: MakeArgs(" hello "), wantResult: []string{"hello"}},
		{name: "Punctuation", toolName: "SplitWords", args: MakeArgs("hello, world!"), wantResult: []string{"hello,", "world!"}},
		{name: "Empty_String", toolName: "SplitWords", args: MakeArgs(""), wantResult: []string{}},
		{name: "Only_Whitespace", toolName: "SplitWords", args: MakeArgs(" \t \n "), wantResult: []string{}},
		{name: "Validation_Non-string_Input", toolName: "SplitWords", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Nil_Input", toolName: "SplitWords", args: MakeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Missing_Input", toolName: "SplitWords", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolJoinStrings(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	// Use []string for wantResult and for valid args
	stringSlice := []string{"a", "b", "c"}
	mixedSlice := []interface{}{"a", int64(1), true}            // For validation test
	numSlice := []interface{}{int64(1), float64(2.5), int64(3)} // For validation test

	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Join_Simple", toolName: "Join", args: MakeArgs(stringSlice, ","), wantResult: "a,b,c"},
		{name: "Join_Empty_Sep", toolName: "Join", args: MakeArgs(stringSlice, ""), wantResult: "abc"},
		{name: "Join_Single_Elem", toolName: "Join", args: MakeArgs([]string{"a"}, ","), wantResult: "a"},
		{name: "Join_Empty_Slice", toolName: "Join", args: MakeArgs([]string{}, ","), wantResult: ""},
		// Corrected: Expect validation error for non-string slice elements
		{name: "Join_Mixed_Types", toolName: "Join", args: MakeArgs(mixedSlice, "-"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Join_Numeric_Types", toolName: "Join", args: MakeArgs(numSlice, " "), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-slice_First_Arg", toolName: "Join", args: MakeArgs("abc", ","), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-string_Separator", toolName: "Join", args: MakeArgs(stringSlice, 123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Nil_Slice", toolName: "Join", args: MakeArgs(nil, ","), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_Separator", toolName: "Join", args: MakeArgs(stringSlice, nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Missing_Separator", toolName: "Join", args: MakeArgs(stringSlice), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}
