package list

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
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

	// Create a new interpreter with the list tools loaded
	interp := interpreter.NewInterpreter()
	// This assumes that the list tools are registered via an init() function
	// in the list package, which is a common pattern. If not, they would
	// need to be registered here manually.
	err := tool.RegisterExtendedTools(interp.ToolRegistry())
	if err != nil {
		t.Fatalf("Failed to register extended tools: %v", err)
	}

	// Get the tool from the registry using its full name
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
					t.Errorf("result does not match expected: got %#v, want %#v", result, tc.Expected)
				}
			}
		})
	}
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
			Name:     "Tail_Empty_List",
			Args:     []interface{}{[]interface{}{}, int64(2)},
			Expected: []interface{}{},
		},
		{
			Name:          "Wrong_Count_Type",
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
