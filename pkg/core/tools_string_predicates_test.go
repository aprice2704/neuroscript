// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Corrected direct tool function calls to match current names
// nlines: 70 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_string_predicates_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// Assume NewTestInterpreter and MakeArgs are defined in testing_helpers.go or similar

// TestToolContainsPrefixSuffix
func TestToolContainsPrefixSuffix(t *testing.T) {
	dummyInterp, _ := NewDefaultTestInterpreter(t) // Using NewDefaultTestInterpreter from helpers.go

	// Contains
	// Note: The ToolSpec defined here is for local ValidateAndConvertArgs testing.
	// The actual tool registration uses the spec in tooldefs_string.go.
	specC := ToolSpec{
		Name: "StringContains", // Updated to match tooldefs_string.go
		Args: []ArgSpec{
			{Name: "input_string", Type: ArgTypeString, Required: true},
			{Name: "substring", Type: ArgTypeString, Required: true},
		},
		ReturnType: ArgTypeBool,
	}
	argsC1 := MakeArgs("hello world", "world")
	convArgsC1, valErrC1 := ValidateAndConvertArgs(specC, argsC1)
	if valErrC1 != nil {
		t.Fatalf("ValidateAndConvertArgs for Contains ('hello world', 'world') failed: %v", valErrC1)
	}
	gotC, errC := toolStringContains(dummyInterp, convArgsC1) // Corrected function name
	if errC != nil || !reflect.DeepEqual(gotC, true) {
		t.Errorf("toolStringContains true failed: err=%v, got=%v, want=true", errC, gotC)
	}

	argsC2 := MakeArgs("hello world", "bye")
	convArgsC2, valErrC2 := ValidateAndConvertArgs(specC, argsC2)
	if valErrC2 != nil {
		t.Fatalf("ValidateAndConvertArgs for Contains ('hello world', 'bye') failed: %v", valErrC2)
	}
	gotC, errC = toolStringContains(dummyInterp, convArgsC2) // Corrected function name
	if errC != nil || !reflect.DeepEqual(gotC, false) {
		t.Errorf("toolStringContains false failed: err=%v, got=%v, want=false", errC, gotC)
	}

	argsC3 := MakeArgs("a") // Not enough arguments
	_, errC = ValidateAndConvertArgs(specC, argsC3)
	// The error message from ValidateAndConvertArgs would typically be from ErrValidationArgCount
	// For this specific direct test, we check if the error is non-nil as expected.
	// A more precise check would use errors.Is(errC, ErrValidationArgCount).
	if errC == nil {
		t.Errorf("toolStringContains expected validation error for arg count, got nil")
	} else if !strings.Contains(errC.Error(), "incorrect argument count: expected 2, got 1") && !strings.Contains(errC.Error(), "expected exactly 2") {
		// Accommodate slightly different phrasings from ValidateAndConvertArgs
		t.Errorf("toolStringContains validation error message mismatch, got: %v", errC)
	}

	// HasPrefix
	specP := ToolSpec{
		Name: "StringHasPrefix", // Updated to match tooldefs_string.go
		Args: []ArgSpec{
			{Name: "input_string", Type: ArgTypeString, Required: true},
			{Name: "prefix", Type: ArgTypeString, Required: true},
		},
		ReturnType: ArgTypeBool,
	}
	argsP1 := MakeArgs("hello world", "hello")
	convArgsP1, valErrP1 := ValidateAndConvertArgs(specP, argsP1)
	if valErrP1 != nil {
		t.Fatalf("ValidateAndConvertArgs for HasPrefix ('hello world', 'hello') failed: %v", valErrP1)
	}
	gotP, errP := toolStringHasPrefix(dummyInterp, convArgsP1) // Corrected function name
	if errP != nil || !reflect.DeepEqual(gotP, true) {
		t.Errorf("toolStringHasPrefix true failed: err=%v, got=%v, want=true", errP, gotP)
	}

	argsP2 := MakeArgs("hello world", "world")
	convArgsP2, valErrP2 := ValidateAndConvertArgs(specP, argsP2)
	if valErrP2 != nil {
		t.Fatalf("ValidateAndConvertArgs for HasPrefix ('hello world', 'world') failed: %v", valErrP2)
	}
	gotP, errP = toolStringHasPrefix(dummyInterp, convArgsP2) // Corrected function name
	if errP != nil || !reflect.DeepEqual(gotP, false) {
		t.Errorf("toolStringHasPrefix false failed: err=%v, got=%v, want=false", errP, gotP)
	}
	argsP3 := MakeArgs("a")
	_, errP = ValidateAndConvertArgs(specP, argsP3)
	if errP == nil {
		t.Errorf("toolStringHasPrefix expected validation error for arg count, got nil")
	} else if !strings.Contains(errP.Error(), "incorrect argument count: expected 2, got 1") && !strings.Contains(errP.Error(), "expected exactly 2") {
		t.Errorf("toolStringHasPrefix validation error message mismatch, got: %v", errP)
	}

	// HasSuffix
	specS := ToolSpec{
		Name: "StringHasSuffix", // Updated to match tooldefs_string.go
		Args: []ArgSpec{
			{Name: "input_string", Type: ArgTypeString, Required: true},
			{Name: "suffix", Type: ArgTypeString, Required: true},
		},
		ReturnType: ArgTypeBool,
	}
	argsS1 := MakeArgs("hello world", "world")
	convArgsS1, valErrS1 := ValidateAndConvertArgs(specS, argsS1)
	if valErrS1 != nil {
		t.Fatalf("ValidateAndConvertArgs for HasSuffix ('hello world', 'world') failed: %v", valErrS1)
	}
	gotS, errS := toolStringHasSuffix(dummyInterp, convArgsS1) // Corrected function name
	if errS != nil || !reflect.DeepEqual(gotS, true) {
		t.Errorf("toolStringHasSuffix true failed: err=%v, got=%v, want=true", errS, gotS)
	}

	argsS2 := MakeArgs("hello world", "hello")
	convArgsS2, valErrS2 := ValidateAndConvertArgs(specS, argsS2)
	if valErrS2 != nil {
		t.Fatalf("ValidateAndConvertArgs for HasSuffix ('hello world', 'hello') failed: %v", valErrS2)
	}
	gotS, errS = toolStringHasSuffix(dummyInterp, convArgsS2) // Corrected function name
	if errS != nil || !reflect.DeepEqual(gotS, false) {
		t.Errorf("toolStringHasSuffix false failed: err=%v, got=%v, want=false", errS, gotS)
	}

	argsS3 := MakeArgs("a")
	_, errS = ValidateAndConvertArgs(specS, argsS3)
	if errS == nil {
		t.Errorf("toolStringHasSuffix expected validation error for arg count, got nil")
	} else if !strings.Contains(errS.Error(), "incorrect argument count: expected 2, got 1") && !strings.Contains(errS.Error(), "expected exactly 2") {
		t.Errorf("toolStringHasSuffix validation error message mismatch, got: %v", errS)
	}
}
