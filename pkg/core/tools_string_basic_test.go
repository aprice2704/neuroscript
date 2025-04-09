// filename: pkg/core/tools_string_basic_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

// TestToolStringLength
func TestToolStringLength(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name          string
		args          []interface{}
		want          interface{}
		wantErr       bool
		errorContains string
	}{
		{name: "Empty", args: makeArgs(""), want: int64(0), wantErr: false, errorContains: ""},
		{name: "Simple ASCII", args: makeArgs("hello"), want: int64(5), wantErr: false, errorContains: ""},
		{name: "UTF8", args: makeArgs("你好"), want: int64(2), wantErr: false, errorContains: ""},
		{name: "Mixed", args: makeArgs("hello 你好"), want: int64(8), wantErr: false, errorContains: ""},
		{name: "Wrong Arg Count", args: makeArgs(), want: nil, wantErr: true, errorContains: "expected exactly 1 arguments"},
		{name: "Validation Wrong Type", args: makeArgs(123), want: nil, wantErr: true, errorContains: "expected string, but received type int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolSpec := ToolSpec{Name: "StringLength", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}
			convertedArgs, valErr := ValidateAndConvertArgs(toolSpec, tt.args)
			finalErr := valErr
			var got interface{}
			var toolErr error
			if valErr == nil {
				got, toolErr = toolStringLength(dummyInterp, convertedArgs)
				if toolErr != nil {
					finalErr = toolErr
				}
			}
			if (finalErr != nil) != tt.wantErr {
				t.Errorf("Test %q: Error mismatch. Got error: %v, wantErr: %v", tt.name, finalErr, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errorContains != "" {
				if finalErr == nil || !strings.Contains(finalErr.Error(), tt.errorContains) {
					t.Errorf("Test %q: Expected error containing %q, got: %v", tt.name, tt.errorContains, finalErr)
				}
			}
			if !tt.wantErr && finalErr == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Test %q: Result mismatch.\nGot:  %v (%T)\nWant: %v (%T)", tt.name, got, got, tt.want, tt.want)
			}
		})
	}
}

// TestToolSubstring (Updated expected error strings)
func TestToolSubstring(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool
		valWantErr  bool
		errContains string
	}{
		{name: "Basic", args: makeArgs("hello", int64(1), int64(4)), want: "ell", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Start 0", args: makeArgs("hello", int64(0), int64(2)), want: "he", wantErr: false, valWantErr: false, errContains: ""},
		{name: "End Len", args: makeArgs("hello", int64(3), int64(5)), want: "lo", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Full String", args: makeArgs("hello", int64(0), int64(5)), want: "hello", wantErr: false, valWantErr: false, errContains: ""},
		{name: "UTF8", args: makeArgs("你好世界", int64(1), int64(3)), want: "好世", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Empty Result Start=End", args: makeArgs("hello", int64(2), int64(2)), want: "", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Empty Result Start>End", args: makeArgs("hello", int64(3), int64(1)), want: "", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Index Out of Bounds High Start", args: makeArgs("hello", int64(10), int64(12)), want: "", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Index Out of Bounds High End", args: makeArgs("hello", int64(3), int64(10)), want: "lo", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Index Out of Bounds Low", args: makeArgs("hello", int64(-2), int64(3)), want: "hel", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs("a", int64(1)), want: nil, wantErr: false, valWantErr: true, errContains: "expected exactly 3 arguments"},
		{name: "Non-string Input (Validation)", args: makeArgs(123, int64(0), int64(1)), want: nil, wantErr: false, valWantErr: true, errContains: "expected string, but received type int"},
		// *** UPDATED errContains to match EXACT reported error ***
		{name: "Non-int Start (Validation)", args: makeArgs("a", "b", int64(3)), want: nil, wantErr: false, valWantErr: true, errContains: "value b (string) cannot be converted to int (int64)"},
		{name: "Non-int End (Validation)", args: makeArgs("a", int64(1), "c"), want: nil, wantErr: false, valWantErr: true, errContains: "value c (string) cannot be converted to int (int64)"},
		// *** END UPDATE ***
	}
	toolSpec := ToolSpec{Name: "Substring", Args: []ArgSpec{
		{Name: "input", Type: ArgTypeString, Required: true}, {Name: "start", Type: ArgTypeInt, Required: true}, {Name: "end", Type: ArgTypeInt, Required: true},
	}, ReturnType: ArgTypeString}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertedArgs, valErr := ValidateAndConvertArgs(toolSpec, tt.args)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, valWantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr {
				if tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
					t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.errContains, valErr)
					return
				}
				return
			}
			if valErr != nil {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}
			got, toolErr := toolSubstring(dummyInterp, convertedArgs)
			if (toolErr != nil) != tt.wantErr {
				t.Errorf("toolSubstring() execution error = %v, wantErr %v", toolErr, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolSubstring() result = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToolToUpperLower
func TestToolToUpperLower(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	specUp := ToolSpec{Name: "ToUpper", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
	convArgsUp, valErrUp := ValidateAndConvertArgs(specUp, makeArgs("Hello World"))
	if valErrUp != nil {
		t.Fatalf("ToUpper validation failed: %v", valErrUp)
	}
	upperGot, upperErr := toolToUpper(dummyInterp, convArgsUp)
	if upperErr != nil || upperGot != "HELLO WORLD" {
		t.Errorf("toolToUpper failed: got %v, err %v", upperGot, upperErr)
	}
	specLo := ToolSpec{Name: "ToLower", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
	convArgsLo, valErrLo := ValidateAndConvertArgs(specLo, makeArgs("Hello World"))
	if valErrLo != nil {
		t.Fatalf("ToLower validation failed: %v", valErrLo)
	}
	lowerGot, lowerErr := toolToLower(dummyInterp, convArgsLo)
	if lowerErr != nil || lowerGot != "hello world" {
		t.Errorf("toolToLower failed: got %v, err %v", lowerGot, lowerErr)
	}
}

// TestToolTrimSpace
func TestToolTrimSpace(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	spec := ToolSpec{Name: "TrimSpace", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
	tests := []struct {
		name string
		args []interface{}
		want interface{}
	}{
		{name: "Leading/Trailing", args: makeArgs("  hello world  "), want: "hello world"}, {name: "Only Spaces", args: makeArgs("   "), want: ""},
		{name: "Newlines/Tabs", args: makeArgs("\n\t hello \t\n"), want: "hello"}, {name: "No Spaces", args: makeArgs("hello"), want: "hello"}, {name: "Empty", args: makeArgs(""), want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convArgs, valErr := ValidateAndConvertArgs(spec, tt.args)
			if valErr != nil {
				t.Fatalf("TrimSpace validation failed: %v", valErr)
			}
			got, err := toolTrimSpace(dummyInterp, convArgs)
			if err != nil {
				t.Errorf("toolTrimSpace() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolTrimSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToolReplaceAll
func TestToolReplaceAll(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        string
		wantErr     bool
		errContains string
	}{
		{name: "Basic", args: makeArgs("hello world hello", "hello", "hi"), want: "hi world hi", wantErr: false, errContains: ""},
		{name: "Single Char", args: makeArgs("aaaa", "a", "b"), want: "bbbb", wantErr: false, errContains: ""},
		{name: "Not Found", args: makeArgs("test", "x", "y"), want: "test", wantErr: false, errContains: ""},
		{name: "Empty Old", args: makeArgs("abc", "", "X"), want: "XaXbXcX", wantErr: false, errContains: ""},
		{name: "Empty New", args: makeArgs("hello", "l", ""), want: "heo", wantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs("a", "b"), want: "", wantErr: true, errContains: "expected exactly 3 arguments"},
		{name: "Non-string Input", args: makeArgs(1, "b", "c"), want: "", wantErr: true, errContains: "argument 'input' (index 0): expected string"},
		{name: "Non-string Old", args: makeArgs("a", 1, "b"), want: "", wantErr: true, errContains: "argument 'old' (index 1): expected string"},
		{name: "Non-string New", args: makeArgs("a", "b", 2), want: "", wantErr: true, errContains: "argument 'new' (index 2): expected string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ToolSpec{Name: "ReplaceAll", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "old", Type: ArgTypeString, Required: true}, {Name: "new", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
			convArgs, valErr := ValidateAndConvertArgs(spec, tt.args)
			if (valErr != nil) != tt.wantErr {
				t.Errorf("ReplaceAll validation error = %v, wantErr %v", valErr, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
				t.Errorf("ReplaceAll validation expected error containing %q, got: %v", tt.errContains, valErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, err := toolReplaceAll(dummyInterp, convArgs)
			if err != nil {
				t.Errorf("toolReplaceAll() unexpected tool error: %v", err)
				return
			}
			gotStr, ok := got.(string)
			if !ok {
				t.Errorf("toolReplaceAll() did not return string, got %T", got)
				return
			}
			if gotStr != tt.want {
				t.Errorf("toolReplaceAll() = %q, want %q", gotStr, tt.want)
			}
		})
	}
}
