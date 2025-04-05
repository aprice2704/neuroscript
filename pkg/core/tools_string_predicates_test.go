package core

import (
	"reflect"
	"strings"
	"testing"
	//"unicode/utf8" // Not needed in this part
)

// --- Unit Tests for String Tool Go Functions (Part 2) ---

func TestToolJoinStrings(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{} // Raw args before validation simulation
		want        interface{}
		wantErr     bool // Whether the TOOL call itself should error
		valWantErr  bool // Separate flag: whether VALIDATION should error
		errContains string
	}{
		{"Simple Join ([]string)", makeArgs([]string{"a", "b", "c"}, "-"), "a-b-c", false, false, ""},
		{"Empty Separator", makeArgs([]string{"a", "b", "c"}, ""), "abc", false, false, ""},
		{"Empty Slice", makeArgs([]string{}, "-"), "", false, false, ""},
		{"Slice with Empty Strings", makeArgs([]string{"a", "", "c"}, ","), "a,,c", false, false, ""},
		{"Interface Slice OK (Validation Conv.)", makeArgs([]interface{}{"x", "y", "z"}, ":"), "x:y:z", false, false, ""},
		{"Interface Slice Mixed Types (Validation Conv.)", makeArgs([]interface{}{"a", 1, true}, "-"), "a-1-true", false, false, ""},
		{"Wrong Arg Count", makeArgs([]string{"a"}), nil, false, true, "expected exactly 2 arguments"},                 // Corrected expectation
		{"Wrong Arg Count 3", makeArgs([]string{"a"}, "-", "extra"), nil, false, true, "expected exactly 2 arguments"}, // Corrected expectation
		{"Non-slice First Arg (Validation Err)", makeArgs("abc", "-"), nil, false, true, "expected slice_string, but received incompatible type string"},
		{"Non-string Separator (Validation Err)", makeArgs([]string{"a"}, 123), nil, false, true, "argument 'separator' (index 1): expected string, but received type int"},
	}
	// ... Test runner loop as previously corrected ...
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toolSpec := ToolSpec{Name: "JoinStrings", Args: []ArgSpec{{Name: "input_slice", Type: ArgTypeSliceString, Required: true}, {Name: "separator", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString}
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

func TestToolReplaceAll(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()
	tests := []struct {
		name        string
		args        []interface{}
		want        string
		wantErr     bool // Expect validation error
		errContains string
	}{
		{"Basic", makeArgs("hello world hello", "hello", "hi"), "hi world hi", false, ""},
		{"Single Char", makeArgs("aaaa", "a", "b"), "bbbb", false, ""},
		{"Not Found", makeArgs("test", "x", "y"), "test", false, ""},
		{"Empty Old", makeArgs("abc", "", "X"), "XaXbXcX", false, ""},
		{"Empty New", makeArgs("hello", "l", ""), "heo", false, ""},
		{"Wrong Arg Count", makeArgs("a", "b"), "", true, "expected exactly 3 arguments"}, // Corrected expectation
		{"Non-string Old", makeArgs("a", 1, "b"), "", true, "argument 'old' (index 1): expected string, but received type int"},
		{"Non-string New", makeArgs("a", "b", 2), "", true, "argument 'new' (index 2): expected string, but received type int"},
	}
	// ... Test runner loop as previously corrected ...
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
			if got != tt.want {
				t.Errorf("toolReplaceAll() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestToolContainsPrefixSuffix(t *testing.T) {
	dummyInterp := newDefaultTestInterpreter()

	// Contains
	specC := ToolSpec{Name: "Contains", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "substring", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}
	argsC1 := makeArgs("hello world", "world")
	convArgsC1, valErrC1 := ValidateAndConvertArgs(specC, argsC1)
	if valErrC1 != nil {
		t.Fatalf("Contains val failed: %v", valErrC1)
	}
	gotC, errC := toolContains(dummyInterp, convArgsC1)
	if errC != nil || !reflect.DeepEqual(gotC, true) {
		t.Errorf("toolContains true failed: %v, got %v", errC, gotC)
	}

	argsC2 := makeArgs("hello world", "bye")
	convArgsC2, valErrC2 := ValidateAndConvertArgs(specC, argsC2)
	if valErrC2 != nil {
		t.Fatalf("Contains val failed: %v", valErrC2)
	}
	gotC, errC = toolContains(dummyInterp, convArgsC2)
	if errC != nil || !reflect.DeepEqual(gotC, false) {
		t.Errorf("toolContains false failed: %v, got %v", errC, gotC)
	}

	argsC3 := makeArgs("a") // Wrong arg count
	_, errC = ValidateAndConvertArgs(specC, argsC3)
	if errC == nil || !strings.Contains(errC.Error(), "expected exactly 2 arguments") {
		t.Errorf("toolContains expected validation error for wrong arg count, got %v", errC)
	} // Corrected expectation

	// HasPrefix
	specP := ToolSpec{Name: "HasPrefix", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "prefix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}
	argsP1 := makeArgs("hello world", "hello")
	convArgsP1, valErrP1 := ValidateAndConvertArgs(specP, argsP1)
	if valErrP1 != nil {
		t.Fatalf("HasPrefix val failed: %v", valErrP1)
	}
	gotP, errP := toolHasPrefix(dummyInterp, convArgsP1)
	if errP != nil || !reflect.DeepEqual(gotP, true) {
		t.Errorf("toolHasPrefix true failed: %v", errP)
	}

	argsP2 := makeArgs("hello world", "world")
	convArgsP2, valErrP2 := ValidateAndConvertArgs(specP, argsP2)
	if valErrP2 != nil {
		t.Fatalf("HasPrefix val failed: %v", valErrP2)
	}
	gotP, errP = toolHasPrefix(dummyInterp, convArgsP2)
	if errP != nil || !reflect.DeepEqual(gotP, false) {
		t.Errorf("toolHasPrefix false failed: %v", errP)
	}

	argsP3 := makeArgs("a") // Wrong arg count
	_, errP = ValidateAndConvertArgs(specP, argsP3)
	if errP == nil || !strings.Contains(errP.Error(), "expected exactly 2 arguments") {
		t.Errorf("toolHasPrefix expected validation error for wrong arg count, got %v", errP)
	} // Corrected expectation

	// HasSuffix
	specS := ToolSpec{Name: "HasSuffix", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "suffix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}
	argsS1 := makeArgs("hello world", "world")
	convArgsS1, valErrS1 := ValidateAndConvertArgs(specS, argsS1)
	if valErrS1 != nil {
		t.Fatalf("HasSuffix val failed: %v", valErrS1)
	}
	gotS, errS := toolHasSuffix(dummyInterp, convArgsS1)
	if errS != nil || !reflect.DeepEqual(gotS, true) {
		t.Errorf("toolHasSuffix true failed: %v", errS)
	}

	argsS2 := makeArgs("hello world", "hello")
	convArgsS2, valErrS2 := ValidateAndConvertArgs(specS, argsS2)
	if valErrS2 != nil {
		t.Fatalf("HasSuffix val failed: %v", valErrS2)
	}
	gotS, errS = toolHasSuffix(dummyInterp, convArgsS2)
	if errS != nil || !reflect.DeepEqual(gotS, false) {
		t.Errorf("toolHasSuffix false failed: %v", errS)
	}

	argsS3 := makeArgs("a") // Wrong arg count
	_, errS = ValidateAndConvertArgs(specS, argsS3)
	if errS == nil || !strings.Contains(errS.Error(), "expected exactly 2 arguments") {
		t.Errorf("toolHasSuffix expected validation error for wrong arg count, got %v", errS)
	} // Corrected expectation
}
