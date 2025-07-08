// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to test primitive-based tool implementations directly.
// filename: pkg/tool/maths/tools_math_test.go
// nlines: 110
// risk_rating: LOW

package maths

import (
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// MakeArgs is a convenience function to create a slice of interfaces, useful for constructing tool arguments programmatically.
func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

// testMathToolHelper tests a math tool implementation directly with primitives.
func testMathToolHelper(t *testing.T, interp tool.Runtime, tc struct {
	name       string
	toolName   string
	args       []interface{}
	wantResult interface{}
	wantErrIs  error
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		interpImpl, ok := interp.(interface{ ToolRegistry() tool.ToolRegistry })
		if !ok {
			t.Fatalf("Interpreter does not implement ToolRegistry()")
		}
		fullname := types.MakeFullName(group, tc.toolName)
		toolImpl, found := interpImpl.ToolRegistry().GetTool(fullname)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}

		gotResult, toolErr := toolImpl.Func(interp, tc.args)

		if tc.wantErrIs != nil {
			if toolErr == nil {
				t.Errorf("Expected an error wrapping [%v], but got nil", tc.wantErrIs)
			} else if !errors.Is(toolErr, tc.wantErrIs) {
				t.Errorf("Expected error to wrap [%v], but got: %v", tc.wantErrIs, toolErr)
			}
			return
		}
		if toolErr != nil {
			t.Fatalf("Unexpected error: %v", toolErr)
		}

		// Handle float comparison with tolerance
		if wantFloat, ok := tc.wantResult.(float64); ok {
			gotFloat, ok := lang.ToFloat64(gotResult)
			if !ok {
				t.Errorf("Result mismatch: wanted a float, but got %T", gotResult)
				return
			}
			if math.Abs(wantFloat-gotFloat) > 1e-9 {
				t.Errorf("Float result mismatch:\n  Got:  %v\n  Want: %v", gotFloat, wantFloat)
			}
		} else if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	})
}

func TestToolAdd(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Add Integers", toolName: "Add", args: MakeArgs(float64(5), float64(3)), wantResult: float64(8)},
		{name: "Add Floats", toolName: "Add", args: MakeArgs(float64(2.5), float64(1.5)), wantResult: float64(4.0)},
		{name: "Type Mismatch", toolName: "Add", args: MakeArgs("abc", float64(1)), wantErrIs: lang.ErrInternalTool},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}

func TestToolSubtract(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Subtract Integers", toolName: "Subtract", args: MakeArgs(float64(5), float64(3)), wantResult: float64(2)},
		{name: "Subtract Floats", toolName: "Subtract", args: MakeArgs(float64(2.5), float64(1.5)), wantResult: float64(1.0)},
		{name: "Type Mismatch", toolName: "Subtract", args: MakeArgs(float64(1), "abc"), wantErrIs: lang.ErrInternalTool},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}

func TestToolDivide(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Divide Integers", toolName: "Divide", args: MakeArgs(float64(10), float64(2)), wantResult: float64(5.0)},
		{name: "Divide Floats", toolName: "Divide", args: MakeArgs(float64(5.0), float64(2.0)), wantResult: float64(2.5)},
		{name: "Divide by Zero", toolName: "Divide", args: MakeArgs(float64(10), float64(0)), wantErrIs: lang.ErrDivisionByZero},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}

func TestToolModulo(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Modulo Integers", toolName: "Modulo", args: MakeArgs(int64(10), int64(3)), wantResult: int64(1)},
		{name: "Modulo by Zero", toolName: "Modulo", args: MakeArgs(int64(10), int64(0)), wantErrIs: lang.ErrDivisionByZero},
		{name: "Type Mismatch", toolName: "Modulo", args: MakeArgs(10.5, 3.0), wantErrIs: lang.ErrInternalTool},
	}
	for _, tt := range tests {
		testMathToolHelper(t, interp, tt)
	}
}
