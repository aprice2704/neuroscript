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

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// testStringToolHelper tests a tool implementation directly with primitives.
func testStringToolHelper(t *testing.T, interp *neurogo.Interpreter, tc struct {
	name		string
	toolName	string
	args		[]interface{}
	wantResult	interface{}
	wantErrIs	error
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

func TestToolStringLength(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name		string
		toolName	string
		args		[]interface{}
		wantResult	interface{}
		wantErrIs	error
	}{
		{name: "Simple", toolName: "Length", args: tool.MakeArgs("hello"), wantResult: float64(5)},
		{name: "Empty", toolName: "Length", args: tool.MakeArgs(""), wantResult: float64(0)},
		{name: "UTF8", toolName: "Length", args: tool.MakeArgs("你好"), wantResult: float64(2)},
		{name: "Validation Wrong Type", toolName: "Length", args: tool.MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolSubstring(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name		string
		toolName	string
		args		[]interface{}
		wantResult	interface{}
		wantErrIs	error
	}{
		{name: "Simple_Substring", toolName: "Substring", args: tool.MakeArgs("abcdef", int64(1), int64(3)), wantResult: "bcd"},
		{name: "Substring_To_End", toolName: "Substring", args: tool.MakeArgs("abcdef", int64(3), int64(3)), wantResult: "def"},
		{name: "Substring_Negative_Length", toolName: "Substring", args: tool.MakeArgs("abcdef", int64(4), int64(-1)), wantErrIs: lang.ErrListIndexOutOfBounds},
		{name: "Substring_Negative_Start", toolName: "Substring", args: tool.MakeArgs("abcdef", int64(-2), int64(3)), wantErrIs: lang.ErrListIndexOutOfBounds},
		{name: "Substring_UTF8", toolName: "Substring", args: tool.MakeArgs("你好世界", int64(1), int64(2)), wantResult: "好世"},
		{name: "Validation_Non-string_Input", toolName: "Substring", args: tool.MakeArgs(123, int64(0), int64(1)), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-int_Start", toolName: "Substring", args: tool.MakeArgs("abc", "b", int64(1)), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolToUpperLower(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name		string
		toolName	string
		args		[]interface{}
		wantResult	interface{}
		wantErrIs	error
	}{
		{name: "ToUpper Simple", toolName: "ToUpper", args: tool.MakeArgs("hello"), wantResult: "HELLO"},
		{name: "ToUpper Empty", toolName: "ToUpper", args: tool.MakeArgs(""), wantResult: ""},
		{name: "ToUpper Validation Wrong Type", toolName: "ToUpper", args: tool.MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
		{name: "ToLower Simple", toolName: "ToLower", args: tool.MakeArgs("HELLO"), wantResult: "hello"},
		{name: "ToLower Validation Wrong Type", toolName: "ToLower", args: tool.MakeArgs(true), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolTrimSpace(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name		string
		toolName	string
		args		[]interface{}
		wantResult	interface{}
		wantErrIs	error
	}{
		{name: "Trim Both", toolName: "TrimSpace", args: tool.MakeArgs("  hello  "), wantResult: "hello"},
		{name: "Trim Internal Space", toolName: "TrimSpace", args: tool.MakeArgs(" hello world "), wantResult: "hello world"},
		{name: "Validation Wrong Type", toolName: "TrimSpace", args: tool.MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}

func TestToolReplaceAll(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name		string
		toolName	string
		args		[]interface{}
		wantResult	interface{}
		wantErrIs	error
	}{
		{name: "Simple_Replace", toolName: "Replace", args: tool.MakeArgs("hello world", "l", "X", int64(-1)), wantResult: "heXXo worXd"},
		{name: "Replace_With_Count_1", toolName: "Replace", args: tool.MakeArgs("hello world", "l", "X", int64(1)), wantResult: "heXlo world"},
		{name: "Validation_Non-string_Input", toolName: "Replace", args: tool.MakeArgs(123, "a", "b", int64(-1)), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-int_Count", toolName: "Replace", args: tool.MakeArgs("abc", "a", "b", "c"), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringToolHelper(t, interp, tt)
	}
}