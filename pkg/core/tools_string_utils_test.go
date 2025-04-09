// filename: pkg/core/tools_string_utils_test.go
package core

import (
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

// TestToolLineCountString
func TestToolLineCountString(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name           string
		inputArg       string
		wantResult     int64
		valWantErr     bool
		valErrContains string
	}{
		{name: "Raw String One Line", inputArg: "Hello", wantResult: 1, valWantErr: false, valErrContains: ""},
		{name: "Raw String Multi Line", inputArg: "Hello\nWorld\nTest", wantResult: 3, valWantErr: false, valErrContains: ""},
		{name: "Raw String With Trailing NL", inputArg: "Hello\nWorld\n", wantResult: 2, valWantErr: false, valErrContains: ""},
		{name: "Raw String Empty", inputArg: "", wantResult: 0, valWantErr: false, valErrContains: ""},
		{name: "Raw String Just Newline", inputArg: "\n", wantResult: 1, valWantErr: false, valErrContains: ""},
		{name: "Raw String Just Newlines", inputArg: "\n\n\n", wantResult: 3, valWantErr: false, valErrContains: ""},
		{name: "Validation Wrong Arg Type", inputArg: "", valWantErr: true, valErrContains: "expected string, but received type int"},
		{name: "Validation Wrong Arg Count", inputArg: "", valWantErr: true, valErrContains: "expected exactly 1 arguments"},
	}
	spec := ToolSpec{Name: "LineCountString", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var rawArgs []interface{}
			if tt.name == "Validation Wrong Arg Count" {
				rawArgs = makeArgs()
			} else if tt.name == "Validation Wrong Arg Type" {
				rawArgs = makeArgs(123)
			} else {
				rawArgs = makeArgs(tt.inputArg)
			}
			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("Validate err=%v, wantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.valErrContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.valErrContains)) {
					t.Errorf("Validate expected err %q, got: %v", tt.valErrContains, valErr)
				}
				return
			}
			if valErr != nil && !tt.valWantErr {
				t.Fatalf("Validate unexpected err: %v", valErr)
			}
			gotResult, toolErr := toolLineCountString(dummyInterp, convertedArgs)
			if toolErr != nil {
				t.Fatalf("toolLineCountString unexpected Go err: %v", toolErr)
			}
			gotInt, ok := gotResult.(int64)
			if !ok {
				t.Fatalf("Expected int64 result, got %T (%v)", gotResult, gotResult)
			}
			if gotInt != tt.wantResult {
				t.Errorf("Result mismatch: got %d, want %d", gotInt, tt.wantResult)
			}
		})
	}
}
