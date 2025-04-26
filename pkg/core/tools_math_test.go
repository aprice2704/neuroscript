// filename: pkg/core/tools_math_test.go
package core

import (
	"errors" // Import errors
	"math"
	"reflect"

	// "strings" // No longer needed
	"testing"
)

// Adapt the general test helper logic (used in list tests) for math tools
func testMathToolHelper(t *testing.T, interp *Interpreter, tc struct {
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
			// Special handling for float results due to potential precision issues
			wantFloat, wantIsFloat := tc.wantResult.(float64)
			gotFloat, gotIsFloat := toFloat64(gotResult) // Use coercion helper

			if wantIsFloat && gotIsFloat {
				// Compare floats with a tolerance
				tolerance := 1e-9
				if math.Abs(wantFloat-gotFloat) > tolerance {
					t.Errorf("Tool function float result mismatch:\n  Got:  %v (%T)\n  Want: %v (%T)",
						gotResult, gotResult, tc.wantResult, tc.wantResult)
				}
			} else if !reflect.DeepEqual(gotResult, tc.wantResult) { // Default comparison for non-floats
				t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
					gotResult, gotResult, tc.wantResult, tc.wantResult)
			}
		}
	})
}

func TestToolAdd(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Add Integers", toolName: "Add", args: MakeArgs(int64(5), int64(3)), wantResult: float64(8)},
		{name: "Add Floats", toolName: "Add", args: MakeArgs(float64(2.5), float64(1.5)), wantResult: float64(4.0)},
		{name: "Add Mixed Int/Float", toolName: "Add", args: MakeArgs(int64(5), float64(2.5)), wantResult: float64(7.5)},
		{name: "Add Mixed Float/Int", toolName: "Add", args: MakeArgs(float64(1.5), int64(3)), wantResult: float64(4.5)},
		{name: "Add Zero", toolName: "Add", args: MakeArgs(int64(10), int64(0)), wantResult: float64(10)},
		{name: "Add Negative", toolName: "Add", args: MakeArgs(int64(5), int64(-2)), wantResult: float64(3)},
		{name: "Add Negative Floats", toolName: "Add", args: MakeArgs(float64(-1.5), float64(-0.5)), wantResult: float64(-2.0)},
		{name: "Add String Number Coercion", toolName: "Add", args: MakeArgs("10", "20.5"), wantResult: float64(30.5)}, // Validation should coerce
		// Validation Errors
		{name: "Non-Numeric String Arg1", toolName: "Add", args: MakeArgs("abc", int64(1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Non-Numeric String Arg2", toolName: "Add", args: MakeArgs(int64(1), "def"), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Boolean Arg1", toolName: "Add", args: MakeArgs(true, int64(1)), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Slice Arg2", toolName: "Add", args: MakeArgs(int64(1), []string{"a"}), valWantErrIs: ErrValidationTypeMismatch},
		{name: "Nil Arg1", toolName: "Add", args: MakeArgs(nil, int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Nil Arg2", toolName: "Add", args: MakeArgs(int64(1), nil), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Wrong Arg Count (1)", toolName: "Add", args: MakeArgs(int64(1)), valWantErrIs: ErrValidationArgCount},
		{name: "Wrong Arg Count (0)", toolName: "Add", args: MakeArgs(), valWantErrIs: ErrValidationArgCount},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}

// Add tests for other math tools (Subtract, Multiply, Divide, Modulo) similarly...
// Example for Subtract:
func TestToolSubtract(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Subtract Integers", toolName: "Subtract", args: MakeArgs(int64(5), int64(3)), wantResult: float64(2)},
		{name: "Subtract Floats", toolName: "Subtract", args: MakeArgs(float64(2.5), float64(1.5)), wantResult: float64(1.0)},
		{name: "Subtract String Number Coercion", toolName: "Subtract", args: MakeArgs("10", "5.5"), wantResult: float64(4.5)},
		{name: "Validation Nil Arg1", toolName: "Subtract", args: MakeArgs(nil, int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Type Arg2", toolName: "Subtract", args: MakeArgs(int64(1), "abc"), valWantErrIs: ErrValidationTypeMismatch},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}

// Example for Divide:
func TestToolDivide(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name          string
		toolName      string
		args          []interface{}
		wantResult    interface{}
		wantToolErrIs error
		valWantErrIs  error
	}{
		{name: "Divide Integers", toolName: "Divide", args: MakeArgs(int64(10), int64(2)), wantResult: float64(5.0)},
		{name: "Divide Floats", toolName: "Divide", args: MakeArgs(float64(5.0), float64(2.0)), wantResult: float64(2.5)},
		{name: "Divide by Zero", toolName: "Divide", args: MakeArgs(int64(10), int64(0)), wantToolErrIs: ErrInternalTool}, // Division by zero is an execution error
		{name: "Divide Zero by Number", toolName: "Divide", args: MakeArgs(int64(0), int64(5)), wantResult: float64(0.0)},
		{name: "Validation Nil Arg1", toolName: "Divide", args: MakeArgs(nil, int64(1)), valWantErrIs: ErrValidationRequiredArgNil},
		{name: "Validation Wrong Type Arg2", toolName: "Divide", args: MakeArgs(int64(1), false), valWantErrIs: ErrValidationTypeMismatch},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}

// Add tests for Multiply and Modulo following the same pattern...
