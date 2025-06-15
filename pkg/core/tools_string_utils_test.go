// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Corrected toolName to "LineCount" to match registry and updated result types to float64.
// filename: pkg/core/tools_string_utils_test.go
// nlines: 60
// risk_rating: LOW

package core

import (
	"errors"
	"reflect"
	"testing"
)

// testStringUtilToolHelper tests a string utility tool implementation directly with primitives.
func testStringUtilToolHelper(t *testing.T, interp *Interpreter, tc struct {
	name       string
	toolName   string
	args       []interface{}
	wantResult interface{}
	wantErrIs  error
}) {
	t.Helper()
	t.Run(tc.name, func(t *testing.T) {
		toolImpl, found := interp.ToolRegistry().GetTool(tc.toolName)
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
			t.Fatalf("Unexpected error during tool execution: %v", toolErr)
		}
		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	})
}

func TestToolLineCountString(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Empty String", toolName: "LineCount", args: MakeArgs(""), wantResult: float64(0)},
		{name: "Single Line No NL", toolName: "LineCount", args: MakeArgs("hello"), wantResult: float64(1)},
		{name: "Single Line With NL", toolName: "LineCount", args: MakeArgs("hello\n"), wantResult: float64(1)},
		{name: "Two Lines No Trailing NL", toolName: "LineCount", args: MakeArgs("hello\nworld"), wantResult: float64(2)},
		{name: "Multiple Blank Lines", toolName: "LineCount", args: MakeArgs("\n\n\n"), wantResult: float64(3)},
		// Per Go's strings.Count, CRLF is not treated as a single newline. This test assumes we are counting '\n'.
		{name: "CRLF Line Endings", toolName: "LineCount", args: MakeArgs("line1\r\nline2\r\n"), wantResult: float64(2)},
		{name: "Validation Wrong Arg Type", toolName: "LineCount", args: MakeArgs(123), wantErrIs: ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringUtilToolHelper(t, interp, tt)
	}
}
