// NeuroScript Version: 0.5.2
// File version: 12
// Purpose: Corrected test suite helpers to wrap script snippets in a full function definition, fixing parsing errors.
// filename: pkg/interpreter/interpreter_assignment_autocreate_test.go
// nlines: 180
// risk_rating: LOW

package interpreter

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Helper function to execute a script and check a variable's state against a Value type.
func checkVariableStateAfterSet(t *testing.T, script string, initialVars map[string]lang.Value, varName string, expectedValue lang.Value) {
	t.Helper()
	interpreter, err := newLocalTestInterpreter(t, initialVars, nil)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	// FIX: Wrap the script in a function to make it a valid program.
	scriptName := fmt.Sprintf("test_autocreate_%s", varName)
	fullScript := fmt.Sprintf("func %s() means\n%s\nendfunc", scriptName, script)

	_, execErr := interpreter.ExecuteScriptString(scriptName, fullScript, nil)

	if execErr != nil {
		t.Fatalf("Script execution failed for '%s':\nScript:\n%s\nError: %s (Code: %d, Position: %s, Wrapped: %v)",
			varName, script, execErr.Message, execErr.Code, execErr.Position, execErr.Wrapped)
		return
	}

	val, exists := interpreter.GetVariable(varName)
	if !exists {
		t.Errorf("Expected variable '%s' to exist after script execution, but it does not.\nScript:\n%s", varName, script)
		return
	}

	if !reflect.DeepEqual(val, expectedValue) {
		t.Errorf("Variable '%s':\nExpected: %#v (%T)\nGot:      %#v (%T)\nScript:\n%s",
			varName, expectedValue, expectedValue, val, val, script)
	}
}

// Helper function to test for expected failures during assignment.
func checkAssignmentFailure(t *testing.T, script string, initialVars map[string]lang.Value, expectedErrorIs error) {
	t.Helper()
	interpreter, err := newLocalTestInterpreter(t, initialVars, nil)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	// FIX: Wrap the script in a function to make it a valid program.
	scriptName := "test_failure"
	fullScript := fmt.Sprintf("func %s() means\n%s\nendfunc", scriptName, script)

	_, execErr := interpreter.ExecuteScriptString(scriptName, fullScript, nil)

	if execErr == nil {
		t.Fatalf("Expected script to fail, but it succeeded.\nScript:\n%s", script)
		return
	}

	if !errors.Is(execErr, expectedErrorIs) {
		t.Fatalf("Script failed with unexpected error.\nExpected error to wrap: %v\nGot: %v\nScript:\n%s",
			expectedErrorIs, execErr, script)
	}
}

func TestLValueAutoCreation_SuccessCases(t *testing.T) {
	testCases := []struct {
		name          string
		script        string
		initialVars   map[string]lang.Value
		checkVarName  string
		expectedValue lang.Value
	}{
		{
			name:         "base map auto-creation with string key",
			script:       `set a["key1"] = "value1"`,
			checkVarName: "a",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"key1": lang.StringValue{Value: "value1"},
			}),
		},
		{
			name:         "base list auto-creation with numeric index 0",
			script:       `set b[0] = "value0"`,
			checkVarName: "b",
			expectedValue: lang.NewListValue([]lang.Value{
				lang.StringValue{Value: "value0"},
			}),
		},
		{
			name:         "nested map auto-creation via dot access",
			script:       `set d.level1.level2 = "valueD"`,
			checkVarName: "d",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"level1": lang.NewMapValue(map[string]lang.Value{
					"level2": lang.StringValue{Value: "valueD"},
				}),
			}),
		},
		{
			name:         "nested list in map auto-creation",
			script:       `set f.listKey[1] = "item1"`,
			checkVarName: "f",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"listKey": lang.NewListValue([]lang.Value{
					&lang.NilValue{},
					lang.StringValue{Value: "item1"},
				}),
			}),
		},
		{
			name:         "nested map in list auto-creation",
			script:       `set g[0].mapKey = "itemG"`,
			checkVarName: "g",
			expectedValue: lang.NewListValue([]lang.Value{
				lang.NewMapValue(map[string]lang.Value{
					"mapKey": lang.StringValue{Value: "itemG"},
				}),
			}),
		},
		{ // FIX: This test was moved from failures to success cases.
			name:   "auto-create from nil value",
			script: `set myVar[0] = "success"`,
			initialVars: map[string]lang.Value{
				"myVar": &lang.NilValue{},
			},
			checkVarName: "myVar",
			expectedValue: lang.NewListValue([]lang.Value{
				lang.StringValue{Value: "success"},
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkVariableStateAfterSet(t, tc.script, tc.initialVars, tc.checkVarName, tc.expectedValue)
		})
	}
}

func TestLValueAutoCreation_FailureCases(t *testing.T) {
	testCases := []struct {
		name          string
		script        string
		initialVars   map[string]lang.Value
		expectedError error
	}{
		{
			name:   "Assign into a number",
			script: `set myVar[0] = "fail"`,
			initialVars: map[string]lang.Value{
				"myVar": lang.NumberValue{Value: 123},
			},
			expectedError: lang.ErrCannotAccessType,
		},
		{
			name:   "Assign into a string",
			script: `set myVar["key"] = "fail"`,
			initialVars: map[string]lang.Value{
				"myVar": lang.StringValue{Value: "hello"},
			},
			expectedError: lang.ErrCannotAccessType,
		},
		{
			name:   "Assign into a boolean",
			script: `set myVar.field = "fail"`,
			initialVars: map[string]lang.Value{
				"myVar": lang.BoolValue{Value: true},
			},
			expectedError: lang.ErrCannotAccessType,
		},
		{
			name:   "Assign to list with string key",
			script: `set myList["key"] = "fail"`,
			initialVars: map[string]lang.Value{
				"myList": lang.NewListValue(nil),
			},
			expectedError: lang.ErrListInvalidIndexType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkAssignmentFailure(t, tc.script, tc.initialVars, tc.expectedError)
		})
	}
}
