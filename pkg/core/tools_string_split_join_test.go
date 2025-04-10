// filename: pkg/core/tools_string_split_join_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// Assume newTestInterpreter and makeArgs are defined in testing_helpers.go

// TestToolSplitString
func TestToolSplitString(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool // Checks final error (validation or execution)
		errContains string
	}{
		{name: "Comma Delimiter", args: makeArgs("a,b,c", ","), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "Space Delimiter", args: makeArgs("a b c", " "), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "No Delimiter Found", args: makeArgs("abc", ","), want: []string{"abc"}, wantErr: false, errContains: ""},
		{name: "Empty String", args: makeArgs("", ","), want: []string{""}, wantErr: false, errContains: ""},
		{name: "Multi-char Delimiter", args: makeArgs("axxbxxc", "xx"), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "Empty Delimiter", args: makeArgs("abc", ""), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs("a"), want: nil, wantErr: true, errContains: "expected exactly 2 arguments"},
		{name: "Non-string Input", args: makeArgs(1, ","), want: nil, wantErr: true, errContains: "type validation failed for argument 'input' of tool 'SplitString': expected string, got int"},
		{name: "Non-string Delimiter", args: makeArgs("a", 1), want: nil, wantErr: true, errContains: "type validation failed for argument 'delimiter' of tool 'SplitString': expected string, got int"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ToolSpec{Name: "SplitString", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "delimiter", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}
			convArgs, finalErr := ValidateAndConvertArgs(spec, tt.args)
			var got interface{}
			var toolErr error
			if finalErr == nil {
				got, toolErr = toolSplitString(dummyInterp, convArgs)
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
				t.Errorf("Test %q: toolSplitString() = %v (%T), want %v (%T)", tt.name, got, got, tt.want, tt.want)
			}
		})
	}
}

// TestToolSplitWords
func TestToolSplitWords(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool
		errContains string
	}{
		{name: "Simple Spaces", args: makeArgs("hello world test"), want: []string{"hello", "world", "test"}, wantErr: false, errContains: ""},
		{name: "Multiple Spaces", args: makeArgs(" hello  world\ttest\n"), want: []string{"hello", "world", "test"}, wantErr: false, errContains: ""},
		{name: "Leading/Trailing", args: makeArgs("  word  "), want: []string{"word"}, wantErr: false, errContains: ""},
		{name: "Empty", args: makeArgs(""), want: []string{}, wantErr: false, errContains: ""},
		{name: "Only Whitespace", args: makeArgs(" \t\n "), want: []string{}, wantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs("a", "b"), want: nil, wantErr: true, errContains: "expected exactly 1 arguments"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ToolSpec{Name: "SplitWords", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}
			convArgs, finalErr := ValidateAndConvertArgs(spec, tt.args)
			var got interface{}
			var toolErr error
			if finalErr == nil {
				got, toolErr = toolSplitWords(dummyInterp, convArgs)
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
				t.Errorf("Test %q: toolSplitWords() = %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

// TestToolJoinStrings (Updated error strings)
func TestToolJoinStrings(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool // Checks final error (validation or execution)
		errContains string
	}{
		{name: "Simple Join ([]string)", args: makeArgs([]string{"a", "b", "c"}, "-"), want: "a-b-c", wantErr: false, errContains: ""},
		{name: "Empty Separator", args: makeArgs([]string{"a", "b", "c"}, ""), want: "abc", wantErr: false, errContains: ""},
		{name: "Empty Slice", args: makeArgs([]string{}, "-"), want: "", wantErr: false, errContains: ""},
		{name: "Slice with Empty Strings", args: makeArgs([]string{"a", "", "c"}, ","), want: "a,,c", wantErr: false, errContains: ""},
		{name: "Interface Slice OK", args: makeArgs([]interface{}{"x", "y", "z"}, ":"), want: "x:y:z", wantErr: false, errContains: ""},
		{name: "Interface Slice Mixed Types", args: makeArgs([]interface{}{"a", 1, true}, "-"), want: "a-1-true", wantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs([]string{"a"}), want: nil, wantErr: true, errContains: "expected exactly 2 arguments"},
		// *** UPDATED Expected Error String ***
		{name: "Non-slice First Arg (Validation Err)", args: makeArgs("abc", "-"), want: nil, wantErr: true, errContains: "type validation failed for argument 'input_slice' of tool 'JoinStrings': expected a slice (list), got string"},
		{name: "Non-string Separator (Validation Err)", args: makeArgs([]string{"a"}, 123), want: nil, wantErr: true, errContains: "type validation failed for argument 'separator' of tool 'JoinStrings': expected string, got int"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolSpec := ToolSpec{Name: "JoinStrings", Args: []ArgSpec{{Name: "input_slice", Type: ArgTypeSliceAny, Required: true}, {Name: "separator", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
			convertedArgs, finalErr := ValidateAndConvertArgs(toolSpec, tt.args)
			var got interface{}
			var toolErr error
			if finalErr == nil {
				got, toolErr = toolJoinStrings(dummyInterp, convertedArgs)
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
				t.Errorf("Test %q: toolJoinStrings() = %v (%T), want %v (%T)", tt.name, got, got, tt.want, tt.want)
			}
		})
	}
}
