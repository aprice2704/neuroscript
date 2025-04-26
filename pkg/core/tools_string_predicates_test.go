// filename: pkg/core/tools_string_predicates_test.go
package core

import (
	"reflect"
	"strings"
	"testing"
)

// Assume newTestInterpreter and MakeArgs are defined in testing_helpers.go

// TestToolContainsPrefixSuffix
func TestToolContainsPrefixSuffix(t *testing.T) {
	dummyInterp, _ := NewDefaultTestInterpreter(t)
	// Contains
	specC := ToolSpec{Name: "Contains", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "substring", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}
	argsC1 := MakeArgs("hello world", "world")
	convArgsC1, _ := ValidateAndConvertArgs(specC, argsC1)
	gotC, errC := toolContains(dummyInterp, convArgsC1)
	if errC != nil || !reflect.DeepEqual(gotC, true) {
		t.Errorf("toolContains true failed: %v, got %v", errC, gotC)
	}
	argsC2 := MakeArgs("hello world", "bye")
	convArgsC2, _ := ValidateAndConvertArgs(specC, argsC2)
	gotC, errC = toolContains(dummyInterp, convArgsC2)
	if errC != nil || !reflect.DeepEqual(gotC, false) {
		t.Errorf("toolContains false failed: %v, got %v", errC, gotC)
	}
	argsC3 := MakeArgs("a")
	_, errC = ValidateAndConvertArgs(specC, argsC3)
	if errC == nil || !strings.Contains(errC.Error(), "expected exactly 2") {
		t.Errorf("toolContains expected validation error, got %v", errC)
	}

	// HasPrefix
	specP := ToolSpec{Name: "HasPrefix", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "prefix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}
	argsP1 := MakeArgs("hello world", "hello")
	convArgsP1, _ := ValidateAndConvertArgs(specP, argsP1)
	gotP, errP := toolHasPrefix(dummyInterp, convArgsP1)
	if errP != nil || !reflect.DeepEqual(gotP, true) {
		t.Errorf("toolHasPrefix true failed: %v", errP)
	}
	argsP2 := MakeArgs("hello world", "world")
	convArgsP2, _ := ValidateAndConvertArgs(specP, argsP2)
	gotP, errP = toolHasPrefix(dummyInterp, convArgsP2)
	if errP != nil || !reflect.DeepEqual(gotP, false) {
		t.Errorf("toolHasPrefix false failed: %v", errP)
	}
	argsP3 := MakeArgs("a")
	_, errP = ValidateAndConvertArgs(specP, argsP3)
	if errP == nil || !strings.Contains(errP.Error(), "expected exactly 2") {
		t.Errorf("toolHasPrefix expected validation error, got %v", errP)
	}

	// HasSuffix
	specS := ToolSpec{Name: "HasSuffix", Args: []ArgSpec{{Name: "input", Type: ArgTypeString, Required: true}, {Name: "suffix", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeBool}
	argsS1 := MakeArgs("hello world", "world")
	convArgsS1, _ := ValidateAndConvertArgs(specS, argsS1)
	gotS, errS := toolHasSuffix(dummyInterp, convArgsS1)
	if errS != nil || !reflect.DeepEqual(gotS, true) {
		t.Errorf("toolHasSuffix true failed: %v", errS)
	}
	argsS2 := MakeArgs("hello world", "hello")
	convArgsS2, _ := ValidateAndConvertArgs(specS, argsS2)
	gotS, errS = toolHasSuffix(dummyInterp, convArgsS2)
	if errS != nil || !reflect.DeepEqual(gotS, false) {
		t.Errorf("toolHasSuffix false failed: %v", errS)
	}
	argsS3 := MakeArgs("a")
	_, errS = ValidateAndConvertArgs(specS, argsS3)
	if errS == nil || !strings.Contains(errS.Error(), "expected exactly 2") {
		t.Errorf("toolHasSuffix expected validation error, got %v", errS)
	}
}
