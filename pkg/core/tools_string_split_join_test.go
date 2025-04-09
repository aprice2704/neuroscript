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
		wantErr     bool
		errContains string
	}{
		{name: "Comma Delimiter", args: makeArgs("a,b,c", ","), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "Space Delimiter", args: makeArgs("a b c", " "), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "No Delimiter Found", args: makeArgs("abc", ","), want: []string{"abc"}, wantErr: false, errContains: ""},
		{name: "Empty String", args: makeArgs("", ","), want: []string{""}, wantErr: false, errContains: ""},
		{name: "Multi-char Delimiter", args: makeArgs("axxbxxc", "xx"), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "Empty Delimiter", args: makeArgs("abc", ""), want: []string{"a", "b", "c"}, wantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs("a"), want: nil, wantErr: true, errContains: "expected exactly 2 arguments"},
		{name: "Non-string Input", args: makeArgs(1, ","), want: nil, wantErr: true, errContains: "argument 'input' (index 0): expected string"},
		{name: "Non-string Delimiter", args: makeArgs("a", 1), want: nil, wantErr: true, errContains: "argument 'delimiter' (index 1): expected string"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := ToolSpec{Name: "SplitString", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "delimiter", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeSliceString}
			convArgs, valErr := ValidateAndConvertArgs(spec, tt.args)
			if (valErr != nil) != tt.wantErr {
				t.Errorf("SplitString validation error = %v, wantErr %v", valErr, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
				t.Errorf("SplitString validation expected error containing %q, got: %v", tt.errContains, valErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, err := toolSplitString(dummyInterp, convArgs)
			if err != nil {
				t.Errorf("toolSplitString() unexpected tool error: %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolSplitString() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
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
			convArgs, valErr := ValidateAndConvertArgs(spec, tt.args)
			if (valErr != nil) != tt.wantErr {
				t.Errorf("SplitWords validation error = %v, wantErr %v", valErr, tt.wantErr)
				return
			}
			if tt.wantErr && tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
				t.Errorf("SplitWords validation expected error containing %q, got: %v", tt.errContains, valErr)
				return
			}
			if tt.wantErr {
				return
			}
			got, err := toolSplitWords(dummyInterp, convArgs)
			if err != nil {
				t.Errorf("toolSplitWords() unexpected tool error: %v", err)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolSplitWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestToolJoinStrings (Updated errContains)
func TestToolJoinStrings(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool
		valWantErr  bool
		errContains string
	}{
		{name: "Simple Join ([]string)", args: makeArgs([]string{"a", "b", "c"}, "-"), want: "a-b-c", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Empty Separator", args: makeArgs([]string{"a", "b", "c"}, ""), want: "abc", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Empty Slice", args: makeArgs([]string{}, "-"), want: "", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Slice with Empty Strings", args: makeArgs([]string{"a", "", "c"}, ","), want: "a,,c", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Interface Slice OK", args: makeArgs([]interface{}{"x", "y", "z"}, ":"), want: "x:y:z", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Interface Slice Mixed Types", args: makeArgs([]interface{}{"a", 1, true}, "-"), want: "a-1-true", wantErr: false, valWantErr: false, errContains: ""},
		{name: "Wrong Arg Count", args: makeArgs([]string{"a"}), want: nil, wantErr: false, valWantErr: true, errContains: "expected exactly 2 arguments"},
		// *** UPDATED errContains to match EXACT reported error ***
		{name: "Non-slice First Arg (Validation Err)", args: makeArgs("abc", "-"), want: nil, wantErr: false, valWantErr: true, errContains: "received incompatible type string"},
		{name: "Non-string Separator (Validation Err)", args: makeArgs([]string{"a"}, 123), want: nil, wantErr: false, valWantErr: true, errContains: "argument 'separator' (index 1): expected string"},
		// *** END UPDATE ***
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolSpec := ToolSpec{Name: "JoinStrings", Args: []ArgSpec{{Name: "input_slice", Type: ArgTypeSliceAny, Required: true}, {Name: "separator", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
			convertedArgs, valErr := ValidateAndConvertArgs(toolSpec, tt.args)
			if (valErr != nil) != tt.valWantErr {
				t.Errorf("ValidateAndConvertArgs() error = %v, valWantErr %v", valErr, tt.valWantErr)
				return
			}
			if tt.valWantErr && tt.errContains != "" && (valErr == nil || !strings.Contains(valErr.Error(), tt.errContains)) {
				t.Errorf("ValidateAndConvertArgs() expected error containing %q, got: %v", tt.errContains, valErr)
				return
			}
			if tt.valWantErr {
				return
			}
			got, toolErr := toolJoinStrings(dummyInterp, convertedArgs)
			if (toolErr != nil) != tt.wantErr {
				t.Errorf("toolJoinStrings() error = %v, wantErr %v", toolErr, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolJoinStrings() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}
