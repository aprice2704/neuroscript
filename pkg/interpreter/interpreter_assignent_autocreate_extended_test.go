// NeuroScript Version: 0.5.2
// File version: 2.0.0
// Purpose: Adds advanced, dedicated tests for complex lvalue auto-creation (vivification) scenarios. Corrected helper to wrap test scripts in function definitions.
// filename: pkg/interpreter/interpreter_assignment_autocreate_extended_test.go
// nlines: 100
// risk_rating: MEDIUM

package interpreter

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Helper function to execute a script and check a variable's state against a Value type.
func checkAdvancedAssignment(t *testing.T, script string, initialVars map[string]lang.Value, varName string, expectedValue lang.Value) {
	t.Helper()
	interpreter, err := newLocalTestInterpreter(t, initialVars, nil)
	if err != nil {
		t.Fatalf("Failed to create test interpreter: %v", err)
	}

	// FIX: Wrap the script in a function to make it a valid program.
	scriptName := fmt.Sprintf("test_autocreate_advanced_%s", varName)
	fullScript := fmt.Sprintf("func %s() means\n%s\nendfunc", scriptName, script)

	_, execErr := interpreter.ExecuteScriptString(scriptName, fullScript, nil)

	if execErr != nil {
		t.Fatalf("Script execution failed for '%s':\nScript:\n%s\nError: %s",
			varName, script, execErr.Error())
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

func TestLValueAutoCreation_AdvancedCases(t *testing.T) {
	testCases := []struct {
		name          string
		script        string
		initialVars   map[string]lang.Value
		checkVarName  string
		expectedValue lang.Value
	}{
		{
			name:         "deeply nested mixed map and list",
			script:       `set a.b[0]["c"].d[1] = "deep value"`,
			checkVarName: "a",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"b": lang.NewListValue([]lang.Value{
					lang.NewMapValue(map[string]lang.Value{
						"c": lang.NewMapValue(map[string]lang.Value{
							"d": lang.NewListValue([]lang.Value{
								&lang.NilValue{},
								lang.StringValue{Value: "deep value"},
							}),
						}),
					}),
				}),
			}),
		},
		{
			name:   "overwrite complex structure with simple value",
			script: `set x.y.z = "final"`,
			initialVars: map[string]lang.Value{
				"x": lang.NewMapValue(map[string]lang.Value{
					"y": lang.NewMapValue(map[string]lang.Value{
						"z": lang.NewListValue([]lang.Value{lang.NumberValue{Value: 1}}),
					}),
				}),
			},
			checkVarName: "x",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"y": lang.NewMapValue(map[string]lang.Value{
					"z": lang.StringValue{Value: "final"},
				}),
			}),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkAdvancedAssignment(t, tc.script, tc.initialVars, tc.checkVarName, tc.expectedValue)
		})
	}

	t.Run("access on nil sub-element fails correctly", func(t *testing.T) {
		script := `set a[1].key = "should fail"`
		initialVars := map[string]lang.Value{
			"a": lang.NewListValue([]lang.Value{lang.NumberValue{Value: 1}, &lang.NilValue{}}),
		}
		// This helper is from interpreter_assignment_autocreate_test.go
		checkAssignmentFailure(t, script, initialVars, lang.ErrCollectionIsNil)
	})
}
