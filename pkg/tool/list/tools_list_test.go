// NeuroScript Version: 0.4.0
// File version: 5
// Purpose: Refactored to test primitive-based tool implementations directly, per the bridge contract.
// filename: pkg/tool/list/tools_list_test.go
// nlines: 320
// risk_rating: LOW
package list

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/testutil"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// MakeArgs is a convenience function to create a slice of interfaces, useful for constructing tool arguments programmatically.
func MakeArgs(vals ...interface{}) []interface{} {
	if vals == nil {
		return []interface{}{}
	}
	return vals
}

// testListToolHelper encapsulates the logic for executing a list tool implementation test case.
// It calls the tool function directly with primitive arguments and compares primitive results.
func testListToolHelper(t *testing.T, interp tool.Runtime, tc struct {
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
		fullname := tool.MakeFullName(group, tc.toolName)
		toolImpl, found := interpImpl.ToolRegistry().GetTool(fullname)
		if !found {
			t.Fatalf("Tool %q not found in registry", tc.toolName)
		}

		// Call the tool's implementation function directly with primitive Go types.
		gotResult, toolErr := toolImpl.Func(interp, tc.args)

		// --- Error Handling ---
		if tc.wantErrIs != nil {
			if toolErr == nil {
				t.Errorf("Expected an error wrapping [%v], but got nil", tc.wantErrIs)
			} else if !errors.Is(toolErr, tc.wantErrIs) {
				// Check if the actual error is a RuntimeError that wraps the expected error.
				if re, ok := toolErr.(*lang.RuntimeError); ok {
					if !errors.Is(re.Wrapped, tc.wantErrIs) {
						t.Errorf("Expected error to wrap [%v], but got runtime error: %v", tc.wantErrIs, toolErr)
					}
				} else {
					t.Errorf("Expected error to wrap [%v], but got a different error: %v", tc.wantErrIs, toolErr)
				}
			}
			return // Test is complete if an error was expected and received.
		}

		if toolErr != nil {
			t.Fatalf("Unexpected error during tool execution: %v", toolErr)
		}

		// --- Result Comparison ---
		if !reflect.DeepEqual(gotResult, tc.wantResult) {
			t.Errorf("Tool function result mismatch:\n  Got:  %#v (%T)\n  Want: %#v (%T)",
				gotResult, gotResult, tc.wantResult, tc.wantResult)
		}
	})
}

// --- Test Data Generation ---

// --- Test Functions for Each Tool ---

func TestToolListLength(t *testing.T) {
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
		{name: "Simple", toolName: "List.Length", args: MakeArgs([]interface{}{1, "a", true}), wantResult: float64(3)},
		{name: "Wrong Type", toolName: "List.Length", args: MakeArgs("not a list"), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListAppendPrepend(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	baseList := []interface{}{"a", float64(1)}
	tests := []struct {
		name       string
		toolName   string // Append or Prepend
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Append Simple", toolName: "List.Append", args: MakeArgs(baseList, true), wantResult: []interface{}{"a", float64(1), true}},
		{name: "Append To Empty", toolName: "List.Append", args: MakeArgs([]interface{}{}, "new"), wantResult: []interface{}{"new"}},
		{name: "Append Wrong Type", toolName: "List.Append", args: MakeArgs("not list", "el"), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Prepend Simple", toolName: "List.Prepend", args: MakeArgs(baseList, true), wantResult: []interface{}{true, "a", float64(1)}},
		{name: "Prepend To Empty", toolName: "List.Prepend", args: MakeArgs([]interface{}{}, "new"), wantResult: []interface{}{"new"}},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListGet(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	list := []interface{}{"a", float64(1), true, nil}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Get First", toolName: "List.Get", args: MakeArgs(list, int64(0)), wantResult: "a"},
		{name: "Get Middle", toolName: "List.Get", args: MakeArgs(list, int64(1)), wantResult: float64(1)},
		{name: "Get Last", toolName: "List.Get", args: MakeArgs(list, int64(2)), wantResult: true},
		{name: "Get Nil Element", toolName: "List.Get", args: MakeArgs(list, int64(3)), wantResult: nil},
		{name: "OOB High No Default", toolName: "List.Get", args: MakeArgs(list, int64(4)), wantResult: nil},
		{name: "OOB Low No Default", toolName: "List.Get", args: MakeArgs(list, int64(-1)), wantResult: nil},
		{name: "Empty No Default", toolName: "List.Get", args: MakeArgs([]interface{}{}, int64(0)), wantResult: nil},
		{name: "OOB High With Default", toolName: "List.Get", args: MakeArgs(list, int64(5), "default"), wantResult: "default"},
		{name: "Wrong Index Type", toolName: "List.Get", args: MakeArgs(list, "abc"), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Wrong List Type", toolName: "List.Get", args: MakeArgs("not list", int64(0)), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListSlice(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	list := []interface{}{"a", "b", "c", "d", "e"}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Slice Middle", toolName: "List.Slice", args: MakeArgs(list, int64(1), int64(4)), wantResult: []interface{}{"b", "c", "d"}},
		{name: "Slice Clamp High End", toolName: "List.Slice", args: MakeArgs(list, int64(3), int64(10)), wantResult: []interface{}{"d", "e"}},
		{name: "Slice Clamp Both", toolName: "List.Slice", args: MakeArgs(list, int64(-1), int64(10)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Slice Empty List", toolName: "List.Slice", args: MakeArgs([]interface{}{}, int64(0), int64(1)), wantResult: []interface{}{}},
		{name: "Wrong List Type", toolName: "List.Slice", args: MakeArgs("no", int64(0), int64(1)), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Wrong Start Type", toolName: "List.Slice", args: MakeArgs(list, "abc", int64(2)), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListContains(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	list := []interface{}{"a", float64(1), true, nil, []interface{}{"sub"}}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Contains String", toolName: "List.Contains", args: MakeArgs(list, "a"), wantResult: true},
		{name: "Contains Float", toolName: "List.Contains", args: MakeArgs(list, float64(1.0)), wantResult: true},
		{name: "Contains Sub-List", toolName: "List.Contains", args: MakeArgs(list, []interface{}{"sub"}), wantResult: true},
		{name: "Not Contains String", toolName: "List.Contains", args: MakeArgs(list, "b"), wantResult: false},
		{name: "Empty List", toolName: "List.Contains", args: MakeArgs([]interface{}{}, "a"), wantResult: false},
		{name: "Wrong List Type", toolName: "List.Contains", args: MakeArgs("no", "a"), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListReverse(t *testing.T) {
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
		{name: "Reverse Simple", toolName: "List.Reverse", args: MakeArgs([]interface{}{"a", float64(1), true}), wantResult: []interface{}{true, float64(1), "a"}},
		{name: "Reverse Single", toolName: "List.Reverse", args: MakeArgs([]interface{}{"a"}), wantResult: []interface{}{"a"}},
		{name: "Reverse Empty", toolName: "List.Reverse", args: MakeArgs([]interface{}{}), wantResult: []interface{}{}},
		{name: "Wrong Type", toolName: "List.Reverse", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListSort(t *testing.T) {
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
		{name: "Sort Strings", toolName: "List.Sort", args: MakeArgs([]interface{}{"c", "a", "b"}), wantResult: []interface{}{"a", "b", "c"}},
		{name: "Sort Floats", toolName: "List.Sort", args: MakeArgs([]interface{}{3.3, 1.1, 2.2}), wantResult: []interface{}{1.1, 2.2, 3.3}},
		{name: "Sort Mixed Numbers", toolName: "List.Sort", args: MakeArgs([]interface{}{float64(3), 1.1, int64(2)}), wantResult: []interface{}{1.1, float64(2), float64(3)}},
		{name: "Sort Empty", toolName: "List.Sort", args: MakeArgs([]interface{}{}), wantResult: []interface{}{}},
		{name: "Sort Mixed String/Num", toolName: "List.Sort", args: MakeArgs([]interface{}{"a", 1}), wantErrIs: lang.ErrListCannotSortMixedTypes},
		{name: "Sort List With Nil Element", toolName: "List.Sort", args: MakeArgs([]interface{}{"a", nil, "c"}), wantErrIs: lang.ErrListCannotSortMixedTypes},
		{name: "Sort Strings Looking Like Numbers", toolName: "List.Sort", args: MakeArgs([]interface{}{"10", "2", "1"}), wantResult: []interface{}{"1", "10", "2"}},
		{name: "Wrong Type", toolName: "List.Sort", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListIsEmpty(t *testing.T) {
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
		{name: "Is Empty True", toolName: "List.IsEmpty", args: MakeArgs([]interface{}{}), wantResult: true},
		{name: "Is Empty False", toolName: "List.IsEmpty", args: MakeArgs([]interface{}{"a"}), wantResult: false},
		{name: "Wrong Type", toolName: "List.IsEmpty", args: MakeArgs("not a list"), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListHead(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	list := []interface{}{"a", "b", float64(1), nil}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Head Simple", toolName: "List.Head", args: MakeArgs(list), wantResult: "a"},
		{name: "Head First is Nil", toolName: "List.Head", args: MakeArgs([]interface{}{nil, "b"}), wantResult: nil},
		{name: "Head Empty", toolName: "List.Head", args: MakeArgs([]interface{}{}), wantResult: nil},
		{name: "Wrong Type", toolName: "List.Head", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListRest(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	list := []interface{}{"a", "b", float64(1)}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Rest Simple", toolName: "List.Rest", args: MakeArgs(list), wantResult: []interface{}{"b", float64(1)}},
		{name: "Rest Single", toolName: "List.Rest", args: MakeArgs([]interface{}{"only"}), wantResult: []interface{}{}},
		{name: "Rest Empty", toolName: "List.Rest", args: MakeArgs([]interface{}{}), wantResult: []interface{}{}},
		{name: "Wrong Type", toolName: "List.Rest", args: MakeArgs(123), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}

func TestToolListTail(t *testing.T) {
	interp, err := testutil.NewTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("NewTestInterpreter failed: %v", err)
	}
	list := []interface{}{"a", "b", "c", "d", "e"}
	tests := []struct {
		name       string
		toolName   string
		args       []interface{}
		wantResult interface{}
		wantErrIs  error
	}{
		{name: "Tail Last 2", toolName: "List.Tail", args: MakeArgs(list, int64(2)), wantResult: []interface{}{"d", "e"}},
		{name: "Tail Last 5 (All)", toolName: "List.Tail", args: MakeArgs(list, int64(5)), wantResult: []interface{}{"a", "b", "c", "d", "e"}},
		{name: "Tail Count 0", toolName: "List.Tail", args: MakeArgs(list, int64(0)), wantResult: []interface{}{}},
		{name: "Tail Empty List", toolName: "List.Tail", args: MakeArgs([]interface{}{}, int64(2)), wantResult: []interface{}{}},
		{name: "Wrong Count Type", toolName: "List.Tail", args: MakeArgs(list, "abc"), wantErrIs: lang.ErrArgumentMismatch},
		{name: "Wrong List Type", toolName: "List.Tail", args: MakeArgs("no", int64(1)), wantErrIs: lang.ErrArgumentMismatch},
	}
	for _, tt := range tests {
		testListToolHelper(t, interp, tt)
	}
}
