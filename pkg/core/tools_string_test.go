package core

import (
	"reflect"
	"strings"
	"testing"
	// Import utf8 package
)

// --- Unit Tests for String Tool Go Functions (Part 1) ---

func TestToolStringLength(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name          string
		args          []interface{}
		want          interface{}
		wantErr       bool
		errorContains string
	}{
		{"Empty", makeArgs(""), int64(0), false, ""},
		{"Simple ASCII", makeArgs("hello"), int64(5), false, ""},
		{"UTF8", makeArgs("你好"), int64(2), false, ""},
		{"Mixed", makeArgs("hello 你好"), int64(8), false, ""},
		// Expect validation error now because ArgTypeString is stricter
		{
			name:          "From Int",
			args:          makeArgs(12345),
			want:          nil,
			wantErr:       true,                                     // Expect validation error
			errorContains: "expected string, but received type int", // Validation error msg
		},
		{"Wrong Arg Count", makeArgs(), nil, true, "expected exactly 1 arguments"},           // Corrected expectation
		{"Wrong Arg Count 2", makeArgs("a", "b"), nil, true, "expected exactly 1 arguments"}, // Corrected expectation
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
				if finalErr == nil {
					t.Errorf("Test %q: Expected error containing %q, but got nil error", tt.name, tt.errorContains)
				} else if !strings.Contains(finalErr.Error(), tt.errorContains) {
					t.Errorf("Test %q: Expected error containing %q, got: %v", tt.name, tt.errorContains, finalErr)
				}
			}
			if !tt.wantErr && finalErr == nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Test %q: Result mismatch.\nGot:  %v (%T)\nWant: %v (%T)", tt.name, got, got, tt.want, tt.want)
			}
		})
	}
}

func TestToolSubstring(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name        string
		args        []interface{} // Raw args before validation simulation
		want        interface{}
		wantErr     bool   // Expect error from tool func itself?
		valWantErr  bool   // Expect error from ValidateAndConvertArgs?
		errContains string // Expected validation error substring
	}{
		// Pass int64 now as ValidateAndConvertArgs produces it
		{"Basic", makeArgs("hello", int64(1), int64(4)), "ell", false, false, ""},
		{"Start 0", makeArgs("hello", int64(0), int64(2)), "he", false, false, ""},
		{"End Len", makeArgs("hello", int64(3), int64(5)), "lo", false, false, ""},
		{"Full String", makeArgs("hello", int64(0), int64(5)), "hello", false, false, ""},
		{"UTF8", makeArgs("你好世界", int64(1), int64(3)), "好世", false, false, ""},
		{"Empty Result Start=End", makeArgs("hello", int64(2), int64(2)), "", false, false, ""},
		{"Empty Result Start>End", makeArgs("hello", int64(3), int64(1)), "", false, false, ""},
		{"Index Out of Bounds High Start", makeArgs("hello", int64(10), int64(12)), "", false, false, ""},
		{"Index Out of Bounds High End", makeArgs("hello", int64(3), int64(10)), "lo", false, false, ""},
		{"Index Out of Bounds Low", makeArgs("hello", int64(-2), int64(3)), "hel", false, false, ""},
		// Validation Error cases
		{"Wrong Arg Count", makeArgs("a", int64(1)), nil, false, true, "expected exactly 3 arguments"}, // Corrected expectation
		{"Non-string Input (Validation)", makeArgs(123, int64(0), int64(1)), nil, false, true, "expected string, but received type int"},
		{"Non-int Start (Validation)", makeArgs("a", "b", int64(3)), nil, false, true, "cannot be converted to int"},
		{"Non-int End (Validation)", makeArgs("a", int64(1), "c"), nil, false, true, "cannot be converted to int"},
	}

	toolSpec := ToolSpec{Name: "Substring", Args: []ArgSpec{
		{Name: "input", Type: ArgTypeString, Required: true},
		{Name: "start", Type: ArgTypeInt, Required: true},
		{Name: "end", Type: ArgTypeInt, Required: true},
	}, ReturnType: ArgTypeString,
	}

	// Test runner loop
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			// Call toolSubstring only if validation passed
			got, toolErr := toolSubstring(dummyInterp, convertedArgs)

			// Check tool execution error expectation
			if (toolErr != nil) != tt.wantErr {
				t.Errorf("toolSubstring() error = %v, wantErr %v", toolErr, tt.wantErr)
				return
			}
			// Compare result only if no error expected from tool itself
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolSubstring() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolToUpperLower(t *testing.T) { /* ... as before ... */
	dummyInterp := newDummyInterpreter()
	specUp := ToolSpec{Name: "ToUpper", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeString}
	convArgsUp, valErrUp := ValidateAndConvertArgs(specUp, makeArgs("Hello World"))
	if valErrUp != nil {
		t.Fatalf("ToUpper validation failed: %v", valErrUp)
	}
	upperGot, upperErr := toolToUpper(dummyInterp, convArgsUp)
	if upperErr != nil || upperGot != "HELLO WORLD" {
		t.Errorf("toolToUpper failed: got %v, err %v", upperGot, upperErr)
	}
	specLo := ToolSpec{Name: "ToLower", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeString}
	convArgsLo, valErrLo := ValidateAndConvertArgs(specLo, makeArgs("Hello World"))
	if valErrLo != nil {
		t.Fatalf("ToLower validation failed: %v", valErrLo)
	}
	lowerGot, lowerErr := toolToLower(dummyInterp, convArgsLo)
	if lowerErr != nil || lowerGot != "hello world" {
		t.Errorf("toolToLower failed: got %v, err %v", lowerGot, lowerErr)
	}
}
func TestToolTrimSpace(t *testing.T) { /* ... as before ... */
	dummyInterp := newDummyInterpreter()
	spec := ToolSpec{Name: "TrimSpace", Args: []ArgSpec{{Name: "input", Type: ArgTypeString}}, ReturnType: ArgTypeString}
	tests := []struct {
		name string
		args []interface{}
		want interface{}
	}{{"Leading/Trailing", makeArgs("  hello world  "), "hello world"}, {"Only Spaces", makeArgs("   "), ""}, {"Newlines/Tabs", makeArgs("\n\t hello \t\n"), "hello"}, {"No Spaces", makeArgs("hello"), "hello"}, {"Empty", makeArgs(""), ""}}
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
func TestToolSplitString(t *testing.T) { /* ... as before, with corrected errContains */
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool
		errContains string
	}{{"Comma Delimiter", makeArgs("a,b,c", ","), []string{"a", "b", "c"}, false, ""}, {"Space Delimiter", makeArgs("a b c", " "), []string{"a", "b", "c"}, false, ""}, {"No Delimiter Found", makeArgs("abc", ","), []string{"abc"}, false, ""}, {"Empty String", makeArgs("", ","), []string{""}, false, ""}, {"Multi-char Delimiter", makeArgs("axxbxxc", "xx"), []string{"a", "b", "c"}, false, ""}, {"Empty Delimiter", makeArgs("abc", ""), []string{"a", "b", "c"}, false, ""}, {"Wrong Arg Count", makeArgs("a"), nil, true, "expected exactly 2 arguments"}, {"Non-string Input", makeArgs(1, ","), nil, true, "argument 'input' (index 0): expected string, but received type int"}, {"Non-string Delimiter", makeArgs("a", 1), nil, true, "argument 'delimiter' (index 1): expected string, but received type int"}}
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
func TestToolSplitWords(t *testing.T) { /* ... as before, with corrected errContains */
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        interface{}
		wantErr     bool
		errContains string
	}{{"Simple Spaces", makeArgs("hello world test"), []string{"hello", "world", "test"}, false, ""}, {"Multiple Spaces", makeArgs(" hello  world\ttest\n"), []string{"hello", "world", "test"}, false, ""}, {"Leading/Trailing", makeArgs("  word  "), []string{"word"}, false, ""}, {"Empty", makeArgs(""), []string{}, false, ""}, {"Only Whitespace", makeArgs(" \t\n "), []string{}, false, ""}, {"Wrong Arg Count", makeArgs("a", "b"), nil, true, "expected exactly 1 arguments"}} // Corrected errContains
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
