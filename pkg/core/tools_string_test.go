package core

import (
	"reflect"
	"testing"
)

// Helper function to create the dummy interpreter needed for tool function signatures
func newDummyInterpreter() *Interpreter {
	// Note: This creates a basic interpreter. If tests needed specific
	// variables or state set in the interpreter, this would need enhancement.
	return NewInterpreter()
}

// Helper function to wrap string slice args in []interface{} for testing ToolFunc
func makeArgs(vals ...interface{}) []interface{} {
	return vals
}

// --- Unit Tests for String Tool Go Functions ---

func TestToolStringLength(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name    string
		args    []interface{} // Now takes []interface{}
		want    interface{}
		wantErr bool
	}{
		{"Empty", makeArgs(""), "0", false},
		{"Simple ASCII", makeArgs("hello"), "5", false},
		{"UTF8", makeArgs("你好"), "2", false},                    // 2 runes
		{"Mixed", makeArgs("hello 你好"), "8", false},             // 5 + 1 space + 2
		{"From Int (via Sprintf)", makeArgs(12345), "5", false}, // Test non-string input
		{"Wrong Arg Count", makeArgs(), nil, true},
		{"Wrong Arg Count 2", makeArgs("a", "b"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pass dummy interpreter and interface{} args
			got, err := toolStringLength(dummyInterp, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("toolStringLength() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolStringLength() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolSubstring(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name    string
		args    []interface{} // Expects string, int, int after validation
		want    interface{}
		wantErr bool
	}{
		// Args passed here assume prior validation/conversion by ValidateAndConvertArgs
		{"Basic", makeArgs("hello", 1, 4), "ell", false},
		{"Start 0", makeArgs("hello", 0, 2), "he", false},
		{"End Len", makeArgs("hello", 3, 5), "lo", false},
		{"Full String", makeArgs("hello", 0, 5), "hello", false},
		{"UTF8", makeArgs("你好世界", 1, 3), "好世", false}, // runes at index 1 and 2
		{"Empty Result Start=End", makeArgs("hello", 2, 2), "", false},
		{"Empty Result Start>End", makeArgs("hello", 3, 1), "", false},
		{"Index Out of Bounds High Start", makeArgs("hello", 10, 12), "", false},
		{"Index Out of Bounds High End", makeArgs("hello", 3, 10), "lo", false}, // Clamps end
		{"Index Out of Bounds Low", makeArgs("hello", -2, 3), "hel", false},     // Clamps start
		{"Wrong Arg Count", makeArgs("a", 1), nil, true},
		// Type errors below assume ValidateAndConvertArgs failed or wasn't called in a direct unit test scenario
		// The function itself now expects ints after validation.
		{"Non-int Start (Error expected)", makeArgs("a", "b", 3), nil, true},
		{"Non-int End (Error expected)", makeArgs("a", 1, "c"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Need to simulate the args *after* ValidateAndConvertArgs would have run
			// For success cases, types match the function's internal expectations (string, int, int)
			// For error cases, we pass invalid types to check the function's internal guards (if any) or expect error from validation phase (tested elsewhere)
			var argsToPass []interface{}
			if !tt.wantErr && len(tt.args) == 3 {
				// Assume validation converted string indices to int for success cases
				_, sOk := tt.args[1].(int) // Check if test case provided int
				_, eOk := tt.args[2].(int)
				if sOk && eOk {
					argsToPass = tt.args
				} else {
					t.Fatalf("Test setup error: success case needs integer indices, got %T, %T", tt.args[1], tt.args[2])
				}
			} else {
				argsToPass = tt.args // Pass as is for error cases or wrong arg counts
			}

			got, err := toolSubstring(dummyInterp, argsToPass)

			if (err != nil) != tt.wantErr {
				t.Errorf("toolSubstring() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolSubstring() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolToUpperLower(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	// Test ToUpper
	upperGot, upperErr := toolToUpper(dummyInterp, makeArgs("Hello World"))
	if upperErr != nil || upperGot != "HELLO WORLD" {
		t.Errorf("toolToUpper failed: got %v, err %v", upperGot, upperErr)
	}
	// Test ToLower
	lowerGot, lowerErr := toolToLower(dummyInterp, makeArgs("Hello World"))
	if lowerErr != nil || lowerGot != "hello world" {
		t.Errorf("toolToLower failed: got %v, err %v", lowerGot, lowerErr)
	}
}

func TestToolTrimSpace(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name string
		args []interface{}
		want interface{}
	}{
		{"Leading/Trailing", makeArgs("  hello world  "), "hello world"},
		{"Only Spaces", makeArgs("   "), ""},
		{"Newlines/Tabs", makeArgs("\n\t hello \t\n"), "hello"},
		{"No Spaces", makeArgs("hello"), "hello"},
		{"Empty", makeArgs(""), ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toolTrimSpace(dummyInterp, tt.args)
			if err != nil {
				t.Errorf("toolTrimSpace() unexpected error: %v", err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolTrimSpace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolSplitString(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name    string
		args    []interface{}
		want    interface{} // Expect []string
		wantErr bool
	}{
		{"Comma Delimiter", makeArgs("a,b,c", ","), []string{"a", "b", "c"}, false},
		{"Space Delimiter", makeArgs("a b c", " "), []string{"a", "b", "c"}, false},
		{"No Delimiter Found", makeArgs("abc", ","), []string{"abc"}, false},
		{"Empty String", makeArgs("", ","), []string{""}, false},
		{"Multi-char Delimiter", makeArgs("axxbxxc", "xx"), []string{"a", "b", "c"}, false},
		{"Empty Delimiter", makeArgs("abc", ""), []string{"a", "b", "c"}, false}, // Go's split behavior
		{"Wrong Arg Count", makeArgs("a"), nil, true},
		{"Non-string Delimiter", makeArgs("a", 1), nil, true}, // Expect error now
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toolSplitString(dummyInterp, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("toolSplitString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			// Compare slices carefully
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				// Use %v for slice comparison output
				t.Errorf("toolSplitString() = %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestToolSplitWords(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name    string
		args    []interface{}
		want    interface{} // Expect []string
		wantErr bool
	}{
		{"Simple Spaces", makeArgs("hello world test"), []string{"hello", "world", "test"}, false},
		{"Multiple Spaces", makeArgs(" hello  world\ttest\n"), []string{"hello", "world", "test"}, false},
		{"Leading/Trailing", makeArgs("  word  "), []string{"word"}, false},
		{"Empty", makeArgs(""), []string{}, false},
		{"Only Whitespace", makeArgs(" \t\n "), []string{}, false},
		{"Wrong Arg Count", makeArgs("a", "b"), nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toolSplitWords(dummyInterp, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("toolSplitWords() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolSplitWords() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolJoinStrings(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name    string
		args    []interface{} // ToolFunc receives []interface{}
		want    interface{}
		wantErr bool
	}{
		// Args passed here should match what the ToolFunc expects *after* validation/conversion
		{"Simple Join", makeArgs([]string{"a", "b", "c"}, "-"), "a-b-c", false},
		{"Empty Separator", makeArgs([]string{"a", "b", "c"}, ""), "abc", false},
		{"Empty Slice", makeArgs([]string{}, "-"), "", false},
		{"Slice with Empty Strings", makeArgs([]string{"a", "", "c"}, ","), "a,,c", false},
		// {"Interface Slice OK", makeArgs([]interface{}{"x", "y", "z"}, ":"), "x:y:z", false}, // ValidateAndConvert should handle this case
		{"Wrong Arg Count", makeArgs([]string{"a"}), nil, true},                          // Only 1 arg
		{"Wrong Arg Count 3", makeArgs([]string{"a"}, "-", "extra"), nil, true},          // 3 args
		{"Non-slice First Arg", makeArgs("abc", "-"), nil, true},                         // Error expected from validation/conversion ideally
		{"Non-string Separator", makeArgs([]string{"a"}, 123), nil, true},                // Error expected from validation/conversion ideally
		{"Interface Slice Bad Content", makeArgs([]interface{}{"a", 1}, "-"), nil, true}, // Error expected from validation/conversion ideally
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Direct unit test - assume args match what the function expects *after* validation
			argsForFunc := tt.args

			// Modify args for specific error cases that the function itself *can* check
			if tt.name == "Non-slice First Arg" {
				// pass the wrong type directly
			} else if tt.name == "Non-string Separator" {
				// pass the wrong type directly
			} else if tt.name == "Interface Slice Bad Content" {
				// This case tests the internal conversion within toolJoinStrings if Validate isn't perfect
				// argsForFunc = makeArgs([]interface{}{"a", 1}, "-") // Pass mixed slice
			} else if len(argsForFunc) == 2 && !tt.wantErr {
				// For valid cases, ensure first arg is specifically []string if possible
				// This might require helper or more complex setup if test args aren't already correct type
				_, ok1 := argsForFunc[0].([]string)
				_, ok2 := argsForFunc[1].(string)
				if !ok1 || !ok2 {
					// Skip if test case args aren't already the precise expected type post-validation
					// Or adjust based on how ValidateAndConvertArgs is expected to work
					t.Skipf("Skipping direct call test %s, requires precise post-validation types", tt.name)
				}
			}

			got, err := toolJoinStrings(dummyInterp, argsForFunc)

			if (err != nil) != tt.wantErr {
				t.Errorf("toolJoinStrings() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toolJoinStrings() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolReplaceAll(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	tests := []struct {
		name    string
		args    []interface{}
		want    string
		wantErr bool
	}{
		{"Basic", makeArgs("hello world hello", "hello", "hi"), "hi world hi", false},
		{"Single Char", makeArgs("aaaa", "a", "b"), "bbbb", false},
		{"Not Found", makeArgs("test", "x", "y"), "test", false},
		{"Empty Old", makeArgs("abc", "", "X"), "XaXbXcX", false}, // Go's ReplaceAll behavior
		{"Empty New", makeArgs("hello", "l", ""), "heo", false},
		{"Wrong Arg Count", makeArgs("a", "b"), "", true},
		{"Non-string Old", makeArgs("a", 1, "b"), "", true},
		{"Non-string New", makeArgs("a", "b", 2), "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toolReplaceAll(dummyInterp, tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("toolReplaceAll() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("toolReplaceAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolContainsPrefixSuffix(t *testing.T) {
	dummyInterp := newDummyInterpreter()
	// Contains
	gotC, errC := toolContains(dummyInterp, makeArgs("hello world", "world"))
	if errC != nil || gotC != "true" {
		t.Errorf("toolContains true failed: %v", errC)
	}
	gotC, errC = toolContains(dummyInterp, makeArgs("hello world", "bye"))
	if errC != nil || gotC != "false" {
		t.Errorf("toolContains false failed: %v", errC)
	}
	_, errC = toolContains(dummyInterp, makeArgs("a")) // Wrong arg count
	if errC == nil {
		t.Errorf("toolContains expected error for wrong arg count")
	}

	// HasPrefix
	gotP, errP := toolHasPrefix(dummyInterp, makeArgs("hello world", "hello"))
	if errP != nil || gotP != "true" {
		t.Errorf("toolHasPrefix true failed: %v", errP)
	}
	gotP, errP = toolHasPrefix(dummyInterp, makeArgs("hello world", "world"))
	if errP != nil || gotP != "false" {
		t.Errorf("toolHasPrefix false failed: %v", errP)
	}
	_, errP = toolHasPrefix(dummyInterp, makeArgs("a")) // Wrong arg count
	if errP == nil {
		t.Errorf("toolHasPrefix expected error for wrong arg count")
	}

	// HasSuffix
	gotS, errS := toolHasSuffix(dummyInterp, makeArgs("hello world", "world"))
	if errS != nil || gotS != "true" {
		t.Errorf("toolHasSuffix true failed: %v", errS)
	}
	gotS, errS = toolHasSuffix(dummyInterp, makeArgs("hello world", "hello"))
	if errS != nil || gotS != "false" {
		t.Errorf("toolHasSuffix false failed: %v", errS)
	}
	_, errS = toolHasSuffix(dummyInterp, makeArgs("a")) // Wrong arg count
	if errS == nil {
		t.Errorf("toolHasSuffix expected error for wrong arg count")
	}
}
