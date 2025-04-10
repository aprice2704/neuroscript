// filename: pkg/core/tools_math_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

func TestToolAdd(t *testing.T) {
	// Use newDefaultTestInterpreter which registers core tools
	dummyInterp := newDefaultTestInterpreter()

	tests := []struct {
		name           string
		args           []interface{} // Raw arguments for the tool call
		wantResult     interface{}   // Expected result (should always be float64 now)
		wantErr        bool          // Expect an error from the tool function itself
		valWantErr     bool          // Expect an error from ValidateAndConvertArgs?
		valErrContains string        // Substring for expected validation error
	}{
		// --- Valid Cases (Expect float64 results) ---
		{name: "Int + Int", args: makeArgs(int64(5), int64(3)), wantResult: float64(8.0), wantErr: false, valWantErr: false},
		{name: "Float + Float", args: makeArgs(float64(2.5), float64(1.5)), wantResult: float64(4.0), wantErr: false, valWantErr: false},
		{name: "Int + Float", args: makeArgs(int64(5), float64(1.5)), wantResult: float64(6.5), wantErr: false, valWantErr: false},
		{name: "Float + Int", args: makeArgs(float64(2.5), int64(3)), wantResult: float64(5.5), wantErr: false, valWantErr: false},
		{name: "StringInt + Int", args: makeArgs("10", int64(3)), wantResult: float64(13.0), wantErr: false, valWantErr: false}, // Validation coerces to float
		{name: "Int + StringInt", args: makeArgs(int64(3), "10"), wantResult: float64(13.0), wantErr: false, valWantErr: false},
		{name: "StringFloat + Float", args: makeArgs("1.5", float64(2.5)), wantResult: float64(4.0), wantErr: false, valWantErr: false},
		{name: "Float + StringFloat", args: makeArgs(float64(2.5), "1.5"), wantResult: float64(4.0), wantErr: false, valWantErr: false},
		{name: "StringInt + StringFloat", args: makeArgs("10", "2.5"), wantResult: float64(12.5), wantErr: false, valWantErr: false},
		{name: "Negative Int + Int", args: makeArgs(int64(-5), int64(3)), wantResult: float64(-2.0), wantErr: false, valWantErr: false},
		{name: "Int + Zero", args: makeArgs(int64(7), int64(0)), wantResult: float64(7.0), wantErr: false, valWantErr: false},
		{name: "Float + Zero", args: makeArgs(float64(7.5), float64(0.0)), wantResult: float64(7.5), wantErr: false, valWantErr: false},

		// --- Validation Error Cases ---
		{name: "Wrong Arg Count (1)", args: makeArgs(int64(1)), wantResult: nil, wantErr: false, valWantErr: true, valErrContains: "expected exactly 2 arguments"},
		{name: "Wrong Arg Count (3)", args: makeArgs(int64(1), int64(2), int64(3)), wantResult: nil, wantErr: false, valWantErr: true, valErrContains: "expected exactly 2 arguments"},
		{name: "Non-Numeric String Arg1", args: makeArgs("abc", int64(5)), wantResult: nil, wantErr: false, valWantErr: true, valErrContains: "type validation failed for argument 'num1' of tool 'TOOL.Add': value abc (string) cannot be converted to float (float64)"}, // Kept detailed error
		{name: "Non-Numeric String Arg2", args: makeArgs(int64(5), "def"), wantResult: nil, wantErr: false, valWantErr: true, valErrContains: "type validation failed for argument 'num2' of tool 'TOOL.Add': value def (string) cannot be converted to float (float64)"}, // Kept detailed error
		// *** UPDATED Expected Error Strings ***
		{name: "Boolean Arg1", args: makeArgs(true, int64(5)), wantResult: nil, wantErr: false, valWantErr: true, valErrContains: "type validation failed for argument 'num1' of tool 'TOOL.Add': value true (bool) cannot be converted to float (float64)"},
		{name: "Slice Arg2", args: makeArgs(int64(5), []string{"a"}), wantResult: nil, wantErr: false, valWantErr: true, valErrContains: "type validation failed for argument 'num2' of tool 'TOOL.Add': value [a] ([]string) cannot be converted to float (float64)"},
	}

	// ToolSpec now requires ArgTypeFloat
	spec := ToolSpec{Name: "TOOL.Add", Args: []ArgSpec{
		{Name: "num1", Type: ArgTypeFloat, Required: true},
		{Name: "num2", Type: ArgTypeFloat, Required: true},
	}, ReturnType: ArgTypeFloat} // Return type also updated

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// --- Validation Simulation ---
			convertedArgs, valErr := ValidateAndConvertArgs(spec, tt.args)

			// Check Validation Error Expectation
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error presence mismatch. Got error: %v, wantErr: %v", valErr, tt.valWantErr)
				// Add details about the args for debugging
				t.Logf("Args passed to ValidateAndConvertArgs: %#v", tt.args)
				return
			}

			// If validation error was expected, check the content
			if tt.valWantErr {
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected an error but got nil")
				} else if tt.valErrContains != "" && !strings.Contains(valErr.Error(), tt.valErrContains) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.valErrContains, valErr)
				}
				return // Stop test here if validation failed as expected
			}

			// If validation passed unexpectedly, fail
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			// --- Execution (Only if validation passed and was expected to pass) ---
			gotResult, toolErr := toolAdd(dummyInterp, convertedArgs) // Call the tool function directly

			if (toolErr != nil) != tt.wantErr {
				// Now toolAdd *could* error if conversion inside it failed somehow, but it shouldn't
				t.Fatalf("toolAdd() execution error = %v, wantErr %v", toolErr, tt.wantErr)
			}

			// --- Result Comparison ---
			// Use reflect.DeepEqual for robust comparison, especially with floats
			if !reflect.DeepEqual(gotResult, tt.wantResult) {
				t.Errorf("toolAdd() result mismatch:\ngot:  %#v (%T)\nwant: %#v (%T)",
					gotResult, gotResult, tt.wantResult, tt.wantResult)
			}
		})
	}
}
