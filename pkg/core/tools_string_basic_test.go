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
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple", toolName: "StringLength", args: makeArgs("hello"), wantResult: int64(5)},
		{name: "Empty", toolName: "StringLength", args: makeArgs(""), wantResult: int64(0)},
		{name: "UTF8", toolName: "StringLength", args: makeArgs("你好"), wantResult: int64(2)}, // 2 runes
		{name: "Validation Wrong Type", toolName: "StringLength", args: makeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Validation Nil", toolName: "StringLength", args: makeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Count", toolName: "StringLength", args: makeArgs("a", "b"), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, struct {
			name          string
			toolName      string
			args          []interface{}
			wantResult    interface{}
			wantToolErrIs error
			valWantErrIs  error
		}{tt.name, tt.toolName, tt.args, tt.wantResult, tt.wantToolErrIs, tt.valWantErrIs}) // Pass tt.toolName directly
	}
}

func TestToolSubstring(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple Substring", toolName: "Substring", args: makeArgs("abcdef", int64(1), int64(4)), wantResult: "bcd"},
		{name: "Substring From Start", toolName: "Substring", args: makeArgs("abcdef", int64(0), int64(3)), wantResult: "abc"},
		{name: "Substring To End", toolName: "Substring", args: makeArgs("abcdef", int64(3), int64(6)), wantResult: "def"},
		{name: "Substring Full String", toolName: "Substring", args: makeArgs("abcdef", int64(0), int64(6)), wantResult: "abcdef"},
		{name: "Substring Empty Start=End", toolName: "Substring", args: makeArgs("abcdef", int64(2), int64(2)), wantResult: ""},
		{name: "Substring Empty Start>End", toolName: "Substring", args: makeArgs("abcdef", int64(4), int64(1)), wantResult: ""},
		{name: "Substring Clamp High End", toolName: "Substring", args: makeArgs("abcdef", int64(3), int64(10)), wantResult: "def"},
		{name: "Substring Clamp Low Start", toolName: "Substring", args: makeArgs("abcdef", int64(-2), int64(3)), wantResult: "abc"},
		{name: "Substring Clamp Both", toolName: "Substring", args: makeArgs("abcdef", int64(-1), int64(10)), wantResult: "abcdef"},
		{name: "Substring Empty String", toolName: "Substring", args: makeArgs("", int64(0), int64(0)), wantResult: ""},
		{name: "Substring UTF8", toolName: "Substring", args: makeArgs("你好世界", int64(1), int64(3)), wantResult: "好世"}, // Indices are rune-based
		// Validation errors
		{name: "Non-string Input (Validation)", toolName: "Substring", args: makeArgs(123, int64(0), int64(1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Non-int Start (Validation)", toolName: "Substring", args: makeArgs("abc", "b", int64(1)), valWantErrIs: ErrValidationTypeMismatch}, // Coercion fails
		{name: "Non-int End (Validation)", toolName: "Substring", args: makeArgs("abc", int64(0), "c"), valWantErrIs: ErrValidationTypeMismatch},   // Coercion fails
		{name: "Nil Input", toolName: "Substring", args: makeArgs(nil, int64(0), int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil Start", toolName: "Substring", args: makeArgs("abc", nil, int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil End", toolName: "Substring", args: makeArgs("abc", int64(0), nil), valWantErrIs: ErrValidationRequiredArgNil},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, struct {
			name          string
			toolName      string
			args          []interface{}
			wantResult    interface{}
			wantToolErrIs error
			valWantErrIs  error
		}{tt.name, tt.toolName, tt.args, tt.wantResult, tt.wantToolErrIs, tt.valWantErrIs}) // Pass tt.toolName directly
	}
}

func TestToolToUpperLower(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string // ToUpper or ToLower
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		// ToUpper
		{name: "ToUpper Simple", toolName: "ToUpper", args: makeArgs("hello"), wantResult: "HELLO"},
		{name: "ToUpper Mixed", toolName: "ToUpper", args: makeArgs("Hello World"), wantResult: "HELLO WORLD"},
		{name: "ToUpper Already Upper", toolName: "ToUpper", args: makeArgs("UPPER"), wantResult: "UPPER"},
		{name: "ToUpper Empty", toolName: "ToUpper", args: makeArgs(""), wantResult: ""},
		{name: "ToUpper Numbers/Symbols", toolName: "ToUpper", args: makeArgs("123!@#"), wantResult: "123!@#"},
		{name: "ToUpper Validation Nil", toolName: "ToUpper", args: makeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "ToUpper Validation Wrong Type", toolName: "ToUpper", args: makeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
		// ToLower
		{name: "ToLower Simple", toolName: "ToLower", args: makeArgs("HELLO"), wantResult: "hello"},
		{name: "ToLower Mixed", toolName: "ToLower", args: makeArgs("Hello World"), wantResult: "hello world"},
		{name: "ToLower Already Lower", toolName: "ToLower", args: makeArgs("lower"), wantResult: "lower"},
		{name: "ToLower Empty", toolName: "ToLower", args: makeArgs(""), wantResult: ""},
		{name: "ToLower Numbers/Symbols", toolName: "ToLower", args: makeArgs("123!@#"), wantResult: "123!@#"},
		{name: "ToLower Validation Nil", toolName: "ToLower", args: makeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "ToLower Validation Wrong Type", toolName: "ToLower", args: makeArgs(true), valWantErrIs: ErrValidationTypeMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, struct {
			name          string
			toolName      string
			args          []interface{}
			wantResult    interface{}
			wantToolErrIs error
			valWantErrIs  error
		}{tt.name, tt.toolName, tt.args, tt.wantResult, tt.wantToolErrIs, tt.valWantErrIs}) // Pass tt.toolName directly
	}
}

func TestToolTrimSpace(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Trim Both", toolName: "TrimSpace", args: makeArgs("  hello  "), wantResult: "hello"},
		{name: "Trim Leading", toolName: "TrimSpace", args: makeArgs("\t hello"), wantResult: "hello"},
		{name: "Trim Trailing", toolName: "TrimSpace", args: makeArgs("hello \n "), wantResult: "hello"},
		{name: "Trim None", toolName: "TrimSpace", args: makeArgs("hello"), wantResult: "hello"},
		{name: "Trim Only Space", toolName: "TrimSpace", args: makeArgs("   "), wantResult: ""},
		{name: "Trim Empty", toolName: "TrimSpace", args: makeArgs(""), wantResult: ""},
		{name: "Trim Internal Space", toolName: "TrimSpace", args: makeArgs(" hello world "), wantResult: "hello world"},
		{name: "Validation Nil", toolName: "TrimSpace", args: makeArgs(nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Type", toolName: "TrimSpace", args: makeArgs(123), valWantErrIs: ErrValidationTypeMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, struct {
			name          string
			toolName      string
			args          []interface{}
			wantResult    interface{}
			wantToolErrIs error
			valWantErrIs  error
		}{tt.name, tt.toolName, tt.args, tt.wantResult, tt.wantToolErrIs, tt.valWantErrIs}) // Pass tt.toolName directly
	}
}

func TestToolReplaceAll(t *testing.T) {
	interp, _ := newDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Simple Replace", toolName: "ReplaceAll", args: makeArgs("hello world", "l", "X"), wantResult: "heXXo worXd"},
		{name: "Replace Multiple", toolName: "ReplaceAll", args: makeArgs("banana", "a", "o"), wantResult: "bonono"},
		// *** UPDATED Expected Result for Replace Empty Old ***
		{name: "Replace Empty Old", toolName: "ReplaceAll", args: makeArgs("abc", "", "X"), wantResult: "XaXbXcX"}, // Corrected expectation
		{name: "Replace Empty New", toolName: "ReplaceAll", args: makeArgs("abc", "b", ""), wantResult: "ac"},
		{name: "Replace Not Found", toolName: "ReplaceAll", args: makeArgs("abc", "z", "X"), wantResult: "abc"},
		{name: "Replace In Empty", toolName: "ReplaceAll", args: makeArgs("", "a", "X"), wantResult: ""},
		{name: "Non-string Input", toolName: "ReplaceAll", args: makeArgs(123, "a", "b"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Non-string Old", toolName: "ReplaceAll", args: makeArgs("abc", 1, "b"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Non-string New", toolName: "ReplaceAll", args: makeArgs("abc", "a", 2), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Nil Input", toolName: "ReplaceAll", args: makeArgs(nil, "a", "b"), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil Old", toolName: "ReplaceAll", args: makeArgs("abc", nil, "b"), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil New", toolName: "ReplaceAll", args: makeArgs("abc", "a", nil), valWantErrIs: ErrValidationRequiredArgNil},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, struct {
			name          string
			toolName      string
			args          []interface{}
			wantResult    interface{}
			wantToolErrIs error
			valWantErrIs  error
		}{tt.name, tt.toolName, tt.args, tt.wantResult, tt.wantToolErrIs, tt.valWantErrIs}) // Pass tt.toolName directly
	}
}
