// NeuroScript Version: 0.3.1
// File version: 0.1.5
// Correct Substring test expectation for negative length.
// nlines: 215
// risk_rating: LOW
// filename: pkg/core/tools_string_basic_test.go
package core

import (
	"errors" // Import errors
	"reflect"

	// Keep strings
	// "strings" // No longer needed for error checking
	"testing"
)

// Define a local test helper specific to basic string tools or adapt the general one.
// Adapting the general one used in list tests for consistency.
func testStringToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name          string
	toolName      string
	args          []interface{}
	wantResult    interface{} // Expected result *if* no error
	wantToolErrIs error       // Specific Go error expected *from the tool function*
	valWantErrIs  error       // Specific Go error expected *from validation*
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) { // Add t.Run for subtests
		// Use tc.toolName directly as provided by the test case
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
		if !found {
			// Use tc.toolName in the error message
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
			if !reflect.DeepEqual(gotResult, tc.wantResult) {
				t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	})
}

// --- Test Functions ---

func TestToolStringLength(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple", toolName: "Length", args: MakeArgs("hello"), wantResult: int64(5)},
		{name: "Empty", toolName: "Length", args: MakeArgs(""), wantResult: int64(0)},
		{name: "UTF8", toolName: "Length", args: MakeArgs("你好"), wantResult: int64(2)}, // 2 runes
		{name: "Validation Wrong Type", toolName: "Length", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Nil", toolName: "Length", args: MakeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Count", toolName: "Length", args: MakeArgs("a", "b"), valWantErrIs: ErrValidationArgCount},
		{name: "Validation Missing Arg", toolName: "Length", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolSubstring(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple_Substring", toolName: "Substring", args: MakeArgs("abcdef", int64(1), int64(3)), wantResult: "bcd"}, // start=1, len=3 -> bcd
		{name: "Substring_From_Start", toolName: "Substring", args: MakeArgs("abcdef", int64(0), int64(3)), wantResult: "abc"},
		{name: "Substring_To_End", toolName: "Substring", args: MakeArgs("abcdef", int64(3), int64(3)), wantResult: "def"}, // start=3, len=3 -> def
		{name: "Substring_Full_String", toolName: "Substring", args: MakeArgs("abcdef", int64(0), int64(6)), wantResult: "abcdef"},
		{name: "Substring_Empty_Len_0", toolName: "Substring", args: MakeArgs("abcdef", int64(2), int64(0)), wantResult: ""}, // start=2, len=0 -> ""
		// Corrected: Negative length is an error
		{name: "Substring_Negative_Length", toolName: "Substring", args: MakeArgs("abcdef", int64(4), int64(-1)), wantToolErrIs: ErrListIndexOutOfBounds},
		{name: "Substring_Clamp_High_Length", toolName: "Substring", args: MakeArgs("abcdef", int64(3), int64(10)), wantResult: "def"}, // Length clamps
		// Corrected: Negative start index should error
		{name: "Substring_Negative_Start", toolName: "Substring", args: MakeArgs("abcdef", int64(-2), int64(3)), wantToolErrIs: ErrListIndexOutOfBounds},
		{name: "Substring_Empty_String", toolName: "Substring", args: MakeArgs("", int64(0), int64(0)), wantResult: ""},
		{name: "Substring_UTF8", toolName: "Substring", args: MakeArgs("你好世界", int64(1), int64(2)), wantResult: "好世"}, // Indices are rune-based, start=1, len=2 -> 好世
		// Validation errors
		{name: "Validation_Non-string_Input", toolName: "Substring", args: MakeArgs(123, int64(0), int64(1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-int_Start", toolName: "Substring", args: MakeArgs("abc", "b", int64(1)), valWantErrIs: ErrValidationTypeMismatch},  // Coercion fails
		{name: "Validation_Non-int_Length", toolName: "Substring", args: MakeArgs("abc", int64(0), "c"), valWantErrIs: ErrValidationTypeMismatch}, // Coercion fails
		{name: "Validation_Nil_Input", toolName: "Substring", args: MakeArgs(nil, int64(0), int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_Start", toolName: "Substring", args: MakeArgs("abc", nil, int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_Length", toolName: "Substring", args: MakeArgs("abc", int64(0), nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Missing_Length", toolName: "Substring", args: MakeArgs("abc", int64(0)), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolToUpperLower(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string // ToUpper or ToLower
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		// ToUpper
		{name: "ToUpper Simple", toolName: "ToUpper", args: MakeArgs("hello"), wantResult: "HELLO"},
		{name: "ToUpper Mixed", toolName: "ToUpper", args: MakeArgs("Hello World"), wantResult: "HELLO WORLD"},
		{name: "ToUpper Already Upper", toolName: "ToUpper", args: MakeArgs("UPPER"), wantResult: "UPPER"},
		{name: "ToUpper Empty", toolName: "ToUpper", args: MakeArgs(""), wantResult: ""},
		{name: "ToUpper Numbers/Symbols", toolName: "ToUpper", args: MakeArgs("123!@#"), wantResult: "123!@#"},
		{name: "ToUpper Validation Nil", toolName: "ToUpper", args: MakeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "ToUpper Validation Wrong Type", toolName: "ToUpper", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "ToUpper Validation Missing Arg", toolName: "ToUpper", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
		// ToLower
		{name: "ToLower Simple", toolName: "ToLower", args: MakeArgs("HELLO"), wantResult: "hello"},
		{name: "ToLower Mixed", toolName: "ToLower", args: MakeArgs("Hello World"), wantResult: "hello world"},
		{name: "ToLower Already Lower", toolName: "ToLower", args: MakeArgs("lower"), wantResult: "lower"},
		{name: "ToLower Empty", toolName: "ToLower", args: MakeArgs(""), wantResult: ""},
		{name: "ToLower Numbers/Symbols", toolName: "ToLower", args: MakeArgs("123!@#"), wantResult: "123!@#"},
		{name: "ToLower Validation Nil", toolName: "ToLower", args: MakeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "ToLower Validation Wrong Type", toolName: "ToLower", args: MakeArgs(true), valWantErrIs: ErrValidationTypeMismatch},
		{name: "ToLower Validation Missing Arg", toolName: "ToLower", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolTrimSpace(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Trim Both", toolName: "TrimSpace", args: MakeArgs("  hello  "), wantResult: "hello"},
		{name: "Trim Leading", toolName: "TrimSpace", args: MakeArgs("\t hello"), wantResult: "hello"},
		{name: "Trim Trailing", toolName: "TrimSpace", args: MakeArgs("hello \n "), wantResult: "hello"},
		{name: "Trim None", toolName: "TrimSpace", args: MakeArgs("hello"), wantResult: "hello"},
		{name: "Trim Only Space", toolName: "TrimSpace", args: MakeArgs("   "), wantResult: ""},
		{name: "Trim Empty", toolName: "TrimSpace", args: MakeArgs(""), wantResult: ""},
		{name: "Trim Internal Space", toolName: "TrimSpace", args: MakeArgs(" hello world "), wantResult: "hello world"},
		{name: "Validation Nil", toolName: "TrimSpace", args: MakeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Type", toolName: "TrimSpace", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Missing Arg", toolName: "TrimSpace", args: MakeArgs(), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolReplaceAll(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		// Added count=-1 to all test cases
		{name: "Simple_Replace", toolName: "Replace", args: MakeArgs("hello world", "l", "X", int64(-1)), wantResult: "heXXo worXd"},
		{name: "Replace_Multiple", toolName: "Replace", args: MakeArgs("banana", "a", "o", int64(-1)), wantResult: "bonono"},
		{name: "Replace_Empty_Old", toolName: "Replace", args: MakeArgs("abc", "", "X", int64(-1)), wantResult: "XaXbXcX"},
		{name: "Replace_Empty_New", toolName: "Replace", args: MakeArgs("abc", "b", "", int64(-1)), wantResult: "ac"},
		{name: "Replace_Not_Found", toolName: "Replace", args: MakeArgs("abc", "z", "X", int64(-1)), wantResult: "abc"},
		{name: "Replace_In_Empty", toolName: "Replace", args: MakeArgs("", "a", "X", int64(-1)), wantResult: ""},
		{name: "Replace_With_Count_1", toolName: "Replace", args: MakeArgs("hello world", "l", "X", int64(1)), wantResult: "heXlo world"},
		{name: "Replace_With_Count_2", toolName: "Replace", args: MakeArgs("hello world", "l", "X", int64(2)), wantResult: "heXXo world"},
		// Validation errors
		{name: "Validation_Non-string_Input", toolName: "Replace", args: MakeArgs(123, "a", "b", int64(-1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-string_Old", toolName: "Replace", args: MakeArgs("abc", 1, "b", int64(-1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-string_New", toolName: "Replace", args: MakeArgs("abc", "a", 2, int64(-1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Non-int_Count", toolName: "Replace", args: MakeArgs("abc", "a", "b", "c"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation_Nil_Input", toolName: "Replace", args: MakeArgs(nil, "a", "b", int64(-1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_Old", toolName: "Replace", args: MakeArgs("abc", nil, "b", int64(-1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_New", toolName: "Replace", args: MakeArgs("abc", "a", nil, int64(-1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Nil_Count", toolName: "Replace", args: MakeArgs("abc", "a", "b", nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation_Missing_Count", toolName: "Replace", args: MakeArgs("abc", "a", "b"), valWantErrIs: ErrValidationRequiredArgMissing},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}
