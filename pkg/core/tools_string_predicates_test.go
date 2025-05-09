// NeuroScript Version: 0.3.1
// File version: 0.1.1
// Use errors.Is for validation checks. Use correct tool names.
// nlines: 140
// risk_rating: LOW
// filename: pkg/core/tools_string_predicates_test.go
package core

import (
	"errors" // Import errors package
	"reflect"

	// "strings" // No longer needed for error message checking
	"testing"
)

// Assume NewDefaultTestInterpreter and MakeArgs are defined in testing_helpers.go or similar

// TestToolContainsPrefixSuffix
func TestToolContainsPrefixSuffix(t *testing.T) {
	dummyInterp, _ := NewDefaultTestInterpreter(t) // Using NewDefaultTestInterpreter from helpers.go

	// --- String.Contains Tests ---
	toolNameC := "Contains" // Corrected: Use base name from tooldefs_string.go
	toolImplC, foundC := dummyInterp.ToolRegistry().GetTool(toolNameC)
	if !foundC {
		t.Fatalf("Tool %q not found in registry", toolNameC)
	}
	specC := toolImplC.Spec

	t.Run(toolNameC, func(t *testing.T) {
		// Test Case 1: Contains True
		t.Run("True", func(t *testing.T) {
			argsC1 := MakeArgs("hello world", "world")
			convArgsC1, valErrC1 := ValidateAndConvertArgs(specC, argsC1)
			if valErrC1 != nil {
				t.Fatalf("ValidateAndConvertArgs failed: %v", valErrC1)
			}
			gotC, errC := toolStringContains(dummyInterp, convArgsC1)
			if errC != nil || !reflect.DeepEqual(gotC, true) {
				t.Errorf("got (%v, %v), want (true, nil)", gotC, errC)
			}
		})

		// Test Case 2: Contains False
		t.Run("False", func(t *testing.T) {
			argsC2 := MakeArgs("hello world", "bye")
			convArgsC2, valErrC2 := ValidateAndConvertArgs(specC, argsC2)
			if valErrC2 != nil {
				t.Fatalf("ValidateAndConvertArgs failed: %v", valErrC2)
			}
			gotC, errC := toolStringContains(dummyInterp, convArgsC2)
			if errC != nil || !reflect.DeepEqual(gotC, false) {
				t.Errorf("got (%v, %v), want (false, nil)", gotC, errC)
			}
		})

		// Test Case 3: Missing Substring Arg Validation
		t.Run("Missing_Arg", func(t *testing.T) {
			argsC3 := MakeArgs("a") // Not enough arguments
			_, errC := ValidateAndConvertArgs(specC, argsC3)
			// Corrected: Use errors.Is to check for the specific validation error
			if !errors.Is(errC, ErrValidationRequiredArgMissing) {
				t.Errorf("expected error wrapping [%v], but got [%T] %v", ErrValidationRequiredArgMissing, errC, errC)
			}
		})
	})

	// --- String.HasPrefix Tests ---
	toolNameP := "HasPrefix" // Corrected: Use base name
	toolImplP, foundP := dummyInterp.ToolRegistry().GetTool(toolNameP)
	if !foundP {
		t.Fatalf("Tool %q not found in registry", toolNameP)
	}
	specP := toolImplP.Spec

	t.Run(toolNameP, func(t *testing.T) {
		// Test Case 1: HasPrefix True
		t.Run("True", func(t *testing.T) {
			argsP1 := MakeArgs("hello world", "hello")
			convArgsP1, valErrP1 := ValidateAndConvertArgs(specP, argsP1)
			if valErrP1 != nil {
				t.Fatalf("ValidateAndConvertArgs failed: %v", valErrP1)
			}
			gotP, errP := toolStringHasPrefix(dummyInterp, convArgsP1)
			if errP != nil || !reflect.DeepEqual(gotP, true) {
				t.Errorf("got (%v, %v), want (true, nil)", gotP, errP)
			}
		})

		// Test Case 2: HasPrefix False
		t.Run("False", func(t *testing.T) {
			argsP2 := MakeArgs("hello world", "world")
			convArgsP2, valErrP2 := ValidateAndConvertArgs(specP, argsP2)
			if valErrP2 != nil {
				t.Fatalf("ValidateAndConvertArgs failed: %v", valErrP2)
			}
			gotP, errP := toolStringHasPrefix(dummyInterp, convArgsP2)
			if errP != nil || !reflect.DeepEqual(gotP, false) {
				t.Errorf("got (%v, %v), want (false, nil)", gotP, errP)
			}
		})

		// Test Case 3: Missing Prefix Arg Validation
		t.Run("Missing_Arg", func(t *testing.T) {
			argsP3 := MakeArgs("a")
			_, errP := ValidateAndConvertArgs(specP, argsP3)
			// Corrected: Use errors.Is
			if !errors.Is(errP, ErrValidationRequiredArgMissing) {
				t.Errorf("expected error wrapping [%v], but got [%T] %v", ErrValidationRequiredArgMissing, errP, errP)
			}
		})
	})

	// --- String.HasSuffix Tests ---
	toolNameS := "HasSuffix" // Corrected: Use base name
	toolImplS, foundS := dummyInterp.ToolRegistry().GetTool(toolNameS)
	if !foundS {
		t.Fatalf("Tool %q not found in registry", toolNameS)
	}
	specS := toolImplS.Spec

	t.Run(toolNameS, func(t *testing.T) {
		// Test Case 1: HasSuffix True
		t.Run("True", func(t *testing.T) {
			argsS1 := MakeArgs("hello world", "world")
			convArgsS1, valErrS1 := ValidateAndConvertArgs(specS, argsS1)
			if valErrS1 != nil {
				t.Fatalf("ValidateAndConvertArgs failed: %v", valErrS1)
			}
			gotS, errS := toolStringHasSuffix(dummyInterp, convArgsS1)
			if errS != nil || !reflect.DeepEqual(gotS, true) {
				t.Errorf("got (%v, %v), want (true, nil)", gotS, errS)
			}
		})

		// Test Case 2: HasSuffix False
		t.Run("False", func(t *testing.T) {
			argsS2 := MakeArgs("hello world", "hello")
			convArgsS2, valErrS2 := ValidateAndConvertArgs(specS, argsS2)
			if valErrS2 != nil {
				t.Fatalf("ValidateAndConvertArgs failed: %v", valErrS2)
			}
			gotS, errS := toolStringHasSuffix(dummyInterp, convArgsS2)
			if errS != nil || !reflect.DeepEqual(gotS, false) {
				t.Errorf("got (%v, %v), want (false, nil)", gotS, errS)
			}
		})

		// Test Case 3: Missing Suffix Arg Validation
		t.Run("Missing_Arg", func(t *testing.T) {
			argsS3 := MakeArgs("a")
			_, errS := ValidateAndConvertArgs(specS, argsS3)
			// Corrected: Use errors.Is
			if !errors.Is(errS, ErrValidationRequiredArgMissing) {
				t.Errorf("expected error wrapping [%v], but got [%T] %v", ErrValidationRequiredArgMissing, errS, errS)
			}
		})
	})
}
