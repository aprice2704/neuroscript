// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to test primitive-based tool implementations directly.
// filename: pkg/tool/strtools/tools_string_basic_test.go
// nlines: 161
// risk_rating: LOW

package strtools

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
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

// testStringToolHelper tests a tool implementation directly with primitives.
func testStringToolHelper(t *testing.T, interp tool.Runtime, tc struct {
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
			t.Fatalf("Unexpected error during tool execution: %v", toolErr)
		}
		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	})
}

func TestToolStringLength(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple", toolName: "Length", args: MakeArgs("hello"), wantResult: float64(5)},
		{name: "Empty", toolName: "Length", args: MakeArgs(""), wantResult: float64(0)},
		{name: "UTF8", toolName: "Length", args: MakeArgs("你好"), wantResult: float64(2)},
		{name: "Validation Wrong Type", toolName: "Length", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolSubstring(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple_Substring", toolName: "Substring", args: MakeArgs("abcdef", int64(1), int64(3)), wantResult: "bcd"},
		{name: "Substring_To_End", toolName: "Substring", args: MakeArgs("abcdef", int64(3), int64(3)), wantResult: "def"},
		{name: "Substring_Negative_Length", toolName: "Substring", args: MakeArgs("abcdef", int64(4), int64(-1)), wantErrIs: lang.ErrListIndexOutOfBounds},
		{name: "Substring_Negative_Start", toolName: "Substring", args: MakeArgs("abcdef", int64(-2), int64(3)), wantErrIs: lang.ErrListIndexOutOfBounds},
		{name: "Substring_UTF8", toolName: "Substring", args: MakeArgs("你好世界", int64(1), int64(2)), wantResult: "好世"},
		{name: "Validation_Non-string_Input", toolName: "Substring", args: MakeArgs(123, int64(0), int64(1)), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-int_Start", toolName: "Substring", args: MakeArgs("abc", "b", int64(1)), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolToUpperLower(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "ToUpper Simple", toolName: "ToUpper", args: MakeArgs("hello"), wantResult: "HELLO"},
		{name: "ToUpper Empty", toolName: "ToUpper", args: MakeArgs(""), wantResult: ""},
		{name: "ToUpper Validation Wrong Type", toolName: "ToUpper", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
		{name: "ToLower Simple", toolName: "ToLower", args: MakeArgs("HELLO"), wantResult: "hello"},
		{name: "ToLower Validation Wrong Type", toolName: "ToLower", args: MakeArgs(true), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolTrimSpace(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Trim Both", toolName: "TrimSpace", args: MakeArgs("  hello  "), wantResult: "hello"},
		{name: "Trim Internal Space", toolName: "TrimSpace", args: MakeArgs(" hello world "), wantResult: "hello world"},
		{name: "Validation Wrong Type", toolName: "TrimSpace", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolReplaceAll(t *testing.T) {
	interp := interpreter.NewInterpreter()
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple_Replace", toolName: "Replace", args: MakeArgs("hello world", "l", "X", int64(-1)), wantResult: "heXXo worXd"},
		{name: "Replace_With_Count_1", toolName: "Replace", args: MakeArgs("hello world", "l", "X", int64(1)), wantResult: "heXlo world"},
		{name: "Validation_Non-string_Input", toolName: "Replace", args: MakeArgs(123, "a", "b", int64(-1)), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-int_Count", toolName: "Replace", args: MakeArgs("abc", "a", "b", "c"), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}
