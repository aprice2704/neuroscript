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
		errorContains string // Unified error string check
	}{
		{name: "Empty", args: makeArgs(""), want: int64(0), wantErr: false, errorContains: ""},
		{name: "Simple ASCII", args: makeArgs("hello"), want: int64(5), wantErr: false, errorContains: ""},
		{name: "UTF8", args: makeArgs("你好"), want: int64(2), wantErr: false, errorContains: ""},
		{name: "Mixed", args: makeArgs("hello 你好"), want: int64(8), wantErr: false, errorContains: ""},
		{name: "Wrong Arg Count", args: makeArgs(), want: nil, wantErr: true, errorContains: "expected exactly 1 arguments"},
		// *** UPDATED Expected Error String ***
		{name: "Validation Wrong Type", args: makeArgs(123), want: nil, wantErr: true, errorContains: "type validation failed for argument 'input' of tool 'StringLength': expected string, got int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolSpec := ToolSpec{Name: "StringLength", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeInt}
			convertedArgs, finalErr := ValidateAndConvertArgs(toolSpec, tt.args)
			var got interface{}
			var toolErr error
			if finalErr == nil { // Only execute if validation passed
				got, toolErr = toolStringLength(dummyInterp, convertedArgs)
				if toolErr != nil {
					finalErr = toolErr // Assign tool execution error if it occurred
				}
			}

			// Check error expectation
			if (finalErr != nil) != tt.wantErr {
				t.Errorf("Test %q: Error mismatch. Got error: %v, wantErr: %v", tt.name, finalErr, tt.wantErr)
				return
			}
			// Check error content if error was expected
			if tt.wantErr && tt.errorContains != "" {
				if finalErr == nil || !strings.Contains(finalErr.Error(), tt.errorContains) {
					t.Errorf("Test %q: Expected error containing %q, got: %v", tt.name, tt.errorContains, finalErr)
				}
			}
			// Check result if no error was expected
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
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
		wantErr     bool // Now checks final error (validation or execution)
		errContains string
	}{
		{name: "Basic", args: makeArgs("hello", int64(1), int64(4)), want: "ell", wantErr: false, errContains: ""},
		{name: "Start 0", args: makeArgs("hello", int64(0), int64(2)), want: "he", wantErr: false, errContains: ""},
		{name: "End Len", args: makeArgs("hello", int64(3), int64(5)), want: "lo", wantErr: false, errContains: ""},
		{name: "Full String", args: makeArgs("hello", int64(0), int64(5)), want: "hello", wantErr: false, errContains: ""},
		{name: "UTF8", args: makeArgs("你好世界", int64(1), int64(3)), want: "好世", wantErr: false, errContains: ""},
		{name: "Empty Result Start=End", args: makeArgs("hello", int64(2), int64(2)), want: "", wantErr: false, errContains: ""},
		{name: "Empty Result Start>End", args: makeArgs("hello", int64(3), int64(1)), want: "", wantErr: false, errContains: ""},
		{name: "Index Out of Bounds High Start", args: makeArgs("hello", int64(10), int64(12)), want: "", wantErr: false, errContains: ""},
		{name: "Index Out of Bounds High End", args: makeArgs("hello", int64(3), int64(10)), want: "lo", wantErr: false, errContains: ""},
		{name: "Index Out of Bounds Low", args: makeArgs("hello", int64(-2), int64(3)), want: "hel", wantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs("a", int64(1)), want: nil, wantErr: true, errContains: "expected exactly 3 arguments"},
		// *** UPDATED Expected Error Strings ***
		{name: "Non-string Input (Validation)", args: makeArgs(123, int64(0), int64(1)), want: nil, wantErr: true, errContains: "type validation failed for argument 'input' of tool 'Substring': expected string, got int"},
		{name: "Non-int Start (Validation)", args: makeArgs("a", "b", int64(3)), want: nil, wantErr: true, errContains: "type validation failed for argument 'start' of tool 'Substring': value b (string) cannot be converted to int (int64)"},
		{name: "Non-int End (Validation)", args: makeArgs("a", int64(1), "c"), want: nil, wantErr: true, errContains: "type validation failed for argument 'end' of tool 'Substring': value c (string) cannot be converted to int (int64)"},
	}
	toolSpec := ToolSpec{Name: "Substring", Args: []ArgSpec{
		{Name: "input", Type: ArgTypeString, Required: true}, {Name: "start", Type: ArgTypeInt, Required: true}, {Name: "end", Type: ArgTypeInt, Required: true},
	}, ReturnType: ArgTypeString}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			convertedArgs, finalErr := ValidateAndConvertArgs(toolSpec, tt.args)
			var got interface{}
			var toolErr error
			if finalErr == nil {
				got, toolErr = toolSubstring(dummyInterp, convertedArgs)
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
			if tt.wantErr && tt.errContains != "" {
				if finalErr == nil || !strings.Contains(finalErr.Error(), tt.errContains) {
					t.Errorf("Test %q: Expected error containing %q, got: %v", tt.name, tt.errContains, finalErr)
				}
			}
			// Check result if no error was expected
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Test %q: Result mismatch.\nGot:  %v (%T)\nWant: %v (%T)", tt.name, got, got, tt.want, tt.want)
			}
		})
	}
}

// TestToolToUpperLower (No changes needed, validation already correct)
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

// TestToolTrimSpace (No changes needed, validation already correct)
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

// TestToolReplaceAll (Updated error strings)
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
		// *** UPDATED Expected Error Strings ***
		{name: "Non-string Input", args: makeArgs(1, "b", "c"), want: "", wantErr: true, errContains: "type validation failed for argument 'input' of tool 'ReplaceAll': expected string, got int"},
		{name: "Non-string Old", args: makeArgs("a", 1, "b"), want: "", wantErr: true, errContains: "type validation failed for argument 'old' of tool 'ReplaceAll': expected string, got int"},
		{name: "Non-string New", args: makeArgs("a", "b", 2), want: "", wantErr: true, errContains: "type validation failed for argument 'new' of tool 'ReplaceAll': expected string, got int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ToolSpec{Name: "ReplaceAll", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "old", Type: ArgTypeString, Required: true}, {Name: "new", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
			convArgs, finalErr := ValidateAndConvertArgs(spec, tt.args)
			var got interface{}
			var toolErr error
			if finalErr == nil {
				got, toolErr = toolReplaceAll(dummyInterp, convArgs)
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
			if tt.wantErr && tt.errContains != "" {
				if finalErr == nil || !strings.Contains(finalErr.Error(), tt.errContains) {
					t.Errorf("Test %q: Expected error containing %q, got: %v", tt.name, tt.errContains, finalErr)
				}
			}
			// Check result if no error was expected
			if !tt.wantErr {
				gotStr, ok := got.(string)
				if !ok {
					t.Errorf("Test %q: toolReplaceAll() did not return string, got %T", tt.name, got)
				} else if gotStr != tt.want {
					t.Errorf("Test %q: toolReplaceAll() = %q, want %q", tt.name, gotStr, tt.want)
				}
			}
		})
	}
}
