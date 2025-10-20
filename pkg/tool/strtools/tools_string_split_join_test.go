// NeuroScript Version: 0.4.0
// File version: 3
// Purpose: Corrected test case Validation_Non-string_elements_in_slice to reflect Join's new coercion behavior.
// filename: pkg/tool/strtools/tools_string_split_join_test.go
// nlines: 109
// risk_rating: MEDIUM

package strtools

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
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

func TestToolSplitString(t *testing.T) {
	interp := newStringTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple_Split", toolName: "Split", args: MakeArgs("a,b,c", ","), wantResult: []string{"a", "b", "c"}},
		{name: "Empty_Delimiter", toolName: "Split", args: MakeArgs("abc", ""), wantResult: []string{"a", "b", "c"}},
		{name: "Validation_Non-string_Input", toolName: "Split", args: MakeArgs(123, ","), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-string_Delimiter", toolName: "Split", args: MakeArgs("abc", 1), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolSplitWords(t *testing.T) {
	interp := newStringTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Simple_Words", toolName: "SplitWords", args: MakeArgs("hello world"), wantResult: []string{"hello", "world"}},
		{name: "Multiple_Spaces", toolName: "SplitWords", args: MakeArgs("  hello \t world  \n next"), wantResult: []string{"hello", "world", "next"}},
		{name: "Empty_String", toolName: "SplitWords", args: MakeArgs(""), wantResult: []string{}},
		{name: "Validation_Non-string_Input", toolName: "SplitWords", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}

func TestToolJoinStrings(t *testing.T) {
	interp := newStringTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Join_Simple", toolName: "Join", args: MakeArgs([]string{"a", "b", "c"}, ","), wantResult: "a,b,c"},
		{name: "Join_Empty_Slice", toolName: "Join", args: MakeArgs([]string{}, ","), wantResult: ""},
		{name: "Validation_Non-slice_First_Arg", toolName: "Join", args: MakeArgs("abc", ","), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Validation_Non-string_Separator", toolName: "Join", args: MakeArgs([]string{"a"}, 123), wantErrIs: lang.ErrArgumentMismatch},
		// FIX: Updated test to expect successful coercion, not an error.
		{
			name:       "Validation_Non-string_elements_in_slice",
			toolName:   "Join",
			args:       MakeArgs([]interface{}{"a", 1, true}, ","), // Mixed types
			wantResult: "a,1,true",                                 // Expect coercion to string
			wantErrIs:  nil,                                        // No error expected now
		},
	}
	for _, tt := range tests {
		testStringSplitJoinToolHelper(t, interp, tt)
	}
}
