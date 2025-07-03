// NeuroScript Version: 0.5.2
// File version: 9
// Purpose: Corrected the final compiler error by removing an invalid type assertion.
// filename: pkg/interpreter/interpreter_assignment_autocreate_test.go
// nlines: 150
// risk_rating: LOW

package interpreter

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Helper function to execute a script and check a variable's state against a Value type.
// This now uses the local test interpreter helper to avoid import cycles.
func checkVariableStateAfterSet(t *testing.T, script string, varName string, expectedValue lang.Value) {
	t.Helper()
	interpreter, err := newLocalTestInterpreter(t, nil, nil)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	_, execErr := interpreter.ExecuteScriptString(fmt.Sprintf("test_autocreate_%s", varName), script, nil)

	if execErr != nil {
		// FIX: Removed the invalid type assertion. 'execErr' is already the correct type.
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

func TestLValueAutoCreation(t *testing.T) {
	testCases := []struct {
		name          string
		script        string
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
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkVariableStateAfterSet(t, tc.script, tc.checkVarName, tc.expectedValue)
		})
	}
}
