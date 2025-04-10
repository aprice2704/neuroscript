// filename: pkg/core/tools_string_utils_test.go
package core

import (
	// Added reflect back
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

// TestToolLineCountString
func TestToolLineCountString(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name           string
		inputArg       interface{} // Changed to interface{} for type test
		wantResult     int64
		wantErr        bool // Combined error check
		valErrContains string
	}{
		{name: "Raw String One Line", inputArg: "Hello", wantResult: 1, wantErr: false, valErrContains: ""},
		{name: "Raw String Multi Line", inputArg: "Hello\nWorld\nTest", wantResult: 3, wantErr: false, valErrContains: ""},
		{name: "Raw String With Trailing NL", inputArg: "Hello\nWorld\n", wantResult: 2, wantErr: false, valErrContains: ""},
		{name: "Raw String Empty", inputArg: "", wantResult: 0, wantErr: false, valErrContains: ""},
		{name: "Raw String Just Newline", inputArg: "\n", wantResult: 1, wantErr: false, valErrContains: ""},
		{name: "Raw String Just Newlines", inputArg: "\n\n\n", wantResult: 3, wantErr: false, valErrContains: ""},
		// *** UPDATED Expected Error String and inputArg type ***
		{name: "Validation Wrong Arg Type", inputArg: 123, wantErr: true, valErrContains: "type validation failed for argument 'content' of tool 'LineCountString': expected string, got int"},
		{name: "Validation Wrong Arg Count", inputArg: nil, wantErr: true, valErrContains: "expected exactly 1 arguments"}, // inputArg is nil here, used for arg count test setup
	}
	spec := ToolSpec{Name: "LineCountString", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs() // No args
			} else {
				rawArgs = makeArgs(tt.inputArg) // Pass the defined input arg
			}

			convertedArgs, finalErr := ValidateAndConvertArgs(spec, rawArgs)
			var gotResult interface{}
			var toolErr error
			if finalErr == nil {
				gotResult, toolErr = toolLineCountString(dummyInterp, convertedArgs)
				if toolErr != nil {
					finalErr = toolErr
				}
			}

			// Check error expectation
			if (finalErr != nil) != tt.wantErr {
				t.Errorf("Test %q: Error mismatch. Got error: %v, wantErr: %v", tt.name, finalErr, tt.wantErr)
				return
			}
			// Check error content if error was expected
			if tt.wantErr && tt.valErrContains != "" {
				if finalErr == nil || !strings.Contains(finalErr.Error(), tt.valErrContains) {
					t.Errorf("Test %q: Expected error containing %q, got: %v", tt.name, tt.valErrContains, finalErr)
				}
			}
			// Check result if no error was expected
			if !tt.wantErr {
				gotInt, ok := gotResult.(int64)
				if !ok {
					t.Fatalf("Test %q: Expected int64 result, got %T (%v)", tt.name, gotResult, gotResult)
				}
				if gotInt != tt.wantResult {
					t.Errorf("Test %q: Result mismatch: got %d, want %d", tt.name, gotInt, tt.wantResult)
				}
			}
		})
	}
}
