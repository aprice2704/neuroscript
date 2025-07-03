// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to test primitive-based tool implementations directly.
// filename: pkg/tool/strtools/tools_string_split_join_test.go
// nlines: 107
// risk_rating: MEDIUM

package strtools

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// testStringSplitJoinToolHelper tests a tool implementation directly with primitives.
func testStringSplitJoinToolHelper(t *testing.T, interp tool.Runtime, tc struct {
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

func TestToolSplitString(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple_Split", toolName: "Split", args: tool.MakeArgs("a,b,c", ","), wantResult: []string{"a", "b", "c"}},
		{name: "Empty_Delimiter", toolName: "Split", args: tool.MakeArgs("abc", ""), wantResult: []string{"a", "b", "c"}},
		{name: "Validation_Non-string_Input", toolName: "Split", args: tool.MakeArgs(123, ","), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-string_Delimiter", toolName: "Split", args: tool.MakeArgs("abc", 1), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolSplitWords(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple_Words", toolName: "SplitWords", args: tool.MakeArgs("hello world"), wantResult: []string{"hello", "world"}},
		{name: "Multiple_Spaces", toolName: "SplitWords", args: tool.MakeArgs("  hello \t world  \n next"), wantResult: []string{"hello", "world", "next"}},
		{name: "Empty_String", toolName: "SplitWords", args: tool.MakeArgs(""), wantResult: []string{}},
		{name: "Validation_Non-string_Input", toolName: "SplitWords", args: tool.MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolJoinStrings(t *testing.T) {
	interp, _ := llm.NewDefaultTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Join_Simple", toolName: "Join", args: tool.MakeArgs([]string{"a", "b", "c"}, ","), wantResult: "a,b,c"},
		{name: "Join_Empty_Slice", toolName: "Join", args: tool.MakeArgs([]string{}, ","), wantResult: ""},
		{name: "Validation_Non-slice_First_Arg", toolName: "Join", args: tool.MakeArgs("abc", ","), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-string_Separator", toolName: "Join", args: tool.MakeArgs([]string{"a"}, 123), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-string_elements_in_slice", toolName: "Join", args: tool.MakeArgs([]interface{}{"a", 1}, ","), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}
