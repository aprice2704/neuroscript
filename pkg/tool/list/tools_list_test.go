// NeuroScript Version: 0.4.0
// File version: 6
// Purpose: Corrected error assertion in TestToolListTail/Wrong_Count_Type_Float_NonInt.
package list

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// testCase defines the structure for a single list tool test case.
type testCase struct {
	Name          string
	Args          []interface{}
	Expected      interface{}
	ExpectedErrIs error
}

// helper function to run tests on list tools
func testListTool(t *testing.T, toolName types.ToolName, cases []testCase) {
	t.Helper()

	hostCtx := &interpreter.HostContext{
		Logger: logging.NewTestLogger(t),
		Stdout: &bytes.Buffer{},
		Stdin:  &bytes.Buffer{},
		Stderr: &bytes.Buffer{},
	}
	// Create a new interpreter with the list tools loaded
	interp := interpreter.NewInterpreter(interpreter.WithHostContext(hostCtx))
	// This assumes that the list tools are registered via an init() function
	// in the list package.
	fullName := types.MakeFullName(group, string(toolName))
	toolImpl, found := interp.ToolRegistry().GetTool(fullName)
	if !found {
		t.Fatalf("Tool %q not found in registry", fullName)
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			// Execute the tool with the test case parameters
			result, err := toolImpl.Func(interp, tc.Args)

			// Assert the results using the standard testing package
			if tc.ExpectedErrIs != nil {
				if err == nil {
					t.Errorf("expected an error wrapping '%v', but got nil", tc.ExpectedErrIs)
				} else if !errors.Is(err, tc.ExpectedErrIs) {
					t.Errorf("expected error to wrap '%v', but got: %v", tc.ExpectedErrIs, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if !reflect.DeepEqual(result, tc.Expected) {
					// Add more detail for slice comparison failures
					resSlice, resOk := result.([]interface{})
					expSlice, expOk := tc.Expected.([]interface{})
					if resOk && expOk {
						t.Errorf("result slice does not match expected:\n got: %#v (%d items)\nwant: %#v (%d items)",
							resSlice, len(resSlice), expSlice, len(expSlice))
					} else {
						t.Errorf("result does not match expected:\n got: %#v (%T)\nwant: %#v (%T)",
							result, result, tc.Expected, tc.Expected)
					}
				}
			}
		})
	}
}

func TestToolListAppend(t *testing.T) {
	testCases := []testCase{
		{
			Name:     "Append_Simple",
			Args:     []interface{}{[]interface{}{"a", "b"}, "c"},
			Expected: []interface{}{"a", "b", "c"},
		},
		{
			Name:     "Append_To_Empty",
			Args:     []interface{}{[]interface{}{}, "a"},
			Expected: []interface{}{"a"},
		},
		{
			Name:     "Append_Nil",
			Args:     []interface{}{[]interface{}{"a"}, nil},
			Expected: []interface{}{"a", nil},
		},
		{
			Name:          "Wrong_Type",
			Args:          []interface{}{"not a list", "a"},
			ExpectedErrIs: lang.ErrArgumentMismatch,
		},
	}
	testListTool(t, "Append", testCases)
}

func TestToolListAppendInPlace(t *testing.T) {
	testCases := []testCase{
		{
			Name:     "AppendInPlace_Simple",
			Args:     []interface{}{[]interface{}{"a", "b"}, "c"},
			Expected: []interface{}{"a", "b", "c"},
		},
		{
			Name:     "AppendInPlace_To_Empty",
			Args:     []interface{}{[]interface{}{}, "a"},
			Expected: []interface{}{"a"},
		},
		{
			Name:     "AppendInPlace_Nil",
			Args:     []interface{}{[]interface{}{"a"}, nil},
			Expected: []interface{}{"a", nil},
		},
		{
			Name:          "AppendInPlace_Wrong_Type",
			Args:          []interface{}{"not a list", "a"},
			ExpectedErrIs: lang.ErrArgumentMismatch,
		},
		// Test that it doesn't modify the *original* slice passed in
		// (because Go passes slices by value, the tool gets a copy)
		{
			Name: "AppendInPlace_DoesNotModifyOriginalArg",
			Args: func() []interface{} {
				originalList := []interface{}{"original"}
				// Pass the original list as the first argument
				return []interface{}{originalList, "new"}
			}(),
			Expected: []interface{}{"original", "new"},
			// We can't easily check the original slice *after* the call here,
			// but this test ensures the *returned* value is correct.
		},
	}
	testListTool(t, "AppendInPlace", testCases)
}

func TestToolListHead(t *testing.T) {
	testCases := []testCase{
		{
			Name:     "Head_Simple",
			Args:     []interface{}{[]interface{}{"a", "b"}},
			Expected: "a",
		},
		{
			Name:     "Head_First_is_Nil",
			Args:     []interface{}{[]interface{}{nil, "b"}},
			Expected: nil,
		},
		{
			Name:          "Head_Empty",
			Args:          []interface{}{[]interface{}{}},
			Expected:      nil, // Head of empty list should be nil
			ExpectedErrIs: nil, // and not an error
		},
		{
			Name:          "Wrong_Type",
			Args:          []interface{}{"not a list"},
			ExpectedErrIs: lang.ErrArgumentMismatch,
		},
	}
	testListTool(t, "Head", testCases)
}

func TestToolListRest(t *testing.T) {
	testCases := []testCase{
		{
			Name:     "Rest_Simple",
			Args:     []interface{}{[]interface{}{"a", "b", "c"}},
			Expected: []interface{}{"b", "c"},
		},
		{
			Name:     "Rest_Single",
			Args:     []interface{}{[]interface{}{"a"}},
			Expected: []interface{}{},
		},
		{
			Name:     "Rest_Empty",
			Args:     []interface{}{[]interface{}{}},
			Expected: []interface{}{},
		},
		{
			Name:          "Wrong_Type",
			Args:          []interface{}{"not a list"},
			ExpectedErrIs: lang.ErrArgumentMismatch,
		},
	}
	testListTool(t, "Rest", testCases)
}

func TestToolListTail(t *testing.T) {
	testCases := []testCase{
		{
			Name:     "Tail_Last_2",
			Args:     []interface{}{[]interface{}{"a", "b", "c"}, int64(2)},
			Expected: []interface{}{"b", "c"},
		},
		{
			Name:     "Tail_Last_5_(All)",
			Args:     []interface{}{[]interface{}{"a", "b", "c"}, int64(5)},
			Expected: []interface{}{"a", "b", "c"},
		},
		{
			Name:     "Tail_Count_0",
			Args:     []interface{}{[]interface{}{"a", "b", "c"}, int64(0)},
			Expected: []interface{}{},
		},
		{
			Name:     "Tail_Count_Negative", // Added test for negative count
			Args:     []interface{}{[]interface{}{"a", "b", "c"}, int64(-1)},
			Expected: []interface{}{},
		},
		{
			Name:     "Tail_Empty_List",
			Args:     []interface{}{[]interface{}{}, int64(2)},
			Expected: []interface{}{},
		},
		{
			Name:          "Wrong_Count_Type_Float_NonInt", // Test float count with fraction
			Args:          []interface{}{[]interface{}{"a"}, float64(1.5)},
			ExpectedErrIs: lang.ErrArgumentMismatch, // FIX: Expect ArgumentMismatch (wrapped by type error)
		},
		{
			Name:     "Wrong_Count_Type_Float_Int", // Test float count that is whole number
			Args:     []interface{}{[]interface{}{"a", "b"}, float64(1.0)},
			Expected: []interface{}{"b"}, // Should succeed
		},
		{
			Name:          "Wrong_Count_Type_String",
			Args:          []interface{}{[]interface{}{"a"}, "not an integer"},
			ExpectedErrIs: lang.ErrArgumentMismatch,
		},
		{
			Name:          "Wrong_List_Type",
			Args:          []interface{}{"not a list", int64(1)},
			ExpectedErrIs: lang.ErrArgumentMismatch,
		},
	}
	testListTool(t, "Tail", testCases)
}

// Add other test functions (TestToolListSort, TestToolListGet, etc.) here...
