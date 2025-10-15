// NeuroScript Version: 0.4.0
// File version: 2
// Purpose: Refactored to a single, table-driven test using the new primitive-aware helper.
// filename: pkg/tool/strtools/tools_string_predicates_test.go
// nlines: 68
// risk_rating: LOW

package strtools

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// testStringPredicateToolHelper tests a tool implementation directly with primitives.
func testStringPredicateToolHelper(t *testing.T, interp tool.Runtime, tc struct {
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

func TestToolContainsPrefixSuffix(t *testing.T) {
	interp := newStringTestInterpreter(t)
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Contains True", toolName: "Contains", args: MakeArgs("hello world", "world"), wantResult: true},
		{name: "Contains False", toolName: "Contains", args: MakeArgs("hello world", "bye"), wantResult: false},
		{name: "Contains Wrong Type", toolName: "Contains", args: MakeArgs(123, "a"), wantErrIs: lang.ErrArgumentMismatch},

		{name: "HasPrefix True", toolName: "HasPrefix", args: MakeArgs("hello world", "hello"), wantResult: true},
		{name: "HasPrefix False", toolName: "HasPrefix", args: MakeArgs("hello world", "world"), wantResult: false},
		{name: "HasPrefix Wrong Type", toolName: "HasPrefix", args: MakeArgs(123, "a"), wantErrIs: lang.ErrArgumentMismatch},

		{name: "HasSuffix True", toolName: "HasSuffix", args: MakeArgs("hello world", "world"), wantResult: true},
		{name: "HasSuffix False", toolName: "HasSuffix", args: MakeArgs("hello world", "hello"), wantResult: false},
		{name: "HasSuffix Wrong Type", toolName: "HasSuffix", args: MakeArgs(123, "a"), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testStringPredicateToolHelper(t, interp, tt)
	}
}
