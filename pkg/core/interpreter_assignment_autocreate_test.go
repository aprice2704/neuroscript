// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Adjusted multi-statement tests for correct parsing.
// Purpose: Tests for interpreter's auto-creation of maps/lists during 'set' assignments.
// filename: pkg/core/interpreter_assignment_autocreate_test.go
// nlines: 320 // Approximate
// risk_rating: MEDIUM

package core

import (
	"fmt"
	"reflect" // Required for reflect.TypeOf and reflect.DeepEqual
	"testing"
)

// testLoggerAdapter provides a interfaces.Logger that writes to t.Logf.
type testLoggerAdapter struct {
	t *testing.T
}

func (tla *testLoggerAdapter) Debug(msg string, args ...interface{}) {
	fullMsg := msg
	if len(args) > 0 {
		fullMsg = fmt.Sprintf(msg, args...)
	}
	tla.t.Logf("DEBUG: %s", fullMsg)
}
func (tla *testLoggerAdapter) Info(msg string, args ...interface{}) {
	fullMsg := msg
	if len(args) > 0 {
		fullMsg = fmt.Sprintf(msg, args...)
	}
	tla.t.Logf("INFO: %s", fullMsg)
}
func (tla *testLoggerAdapter) Warn(msg string, args ...interface{}) {
	fullMsg := msg
	if len(args) > 0 {
		fullMsg = fmt.Sprintf(msg, args...)
	}
	tla.t.Logf("WARN: %s", fullMsg)
}
func (tla *testLoggerAdapter) Error(msg string, args ...interface{}) {
	fullMsg := msg
	if len(args) > 0 {
		fullMsg = fmt.Sprintf(msg, args...)
	}
	tla.t.Logf("ERROR: %s", fullMsg)
}
func (tla *testLoggerAdapter) Fatal(msg string, args ...interface{}) {
	fullMsg := msg
	if len(args) > 0 {
		fullMsg = fmt.Sprintf(msg, args...)
	}
	tla.t.Fatalf("FATAL: %s", fullMsg)
}
func (tla *testLoggerAdapter) Debugf(format string, args ...interface{}) {
	tla.t.Logf("DEBUG: "+format, args...)
}
func (tla *testLoggerAdapter) Infof(format string, args ...interface{}) {
	tla.t.Logf("INFO: "+format, args...)
}
func (tla *testLoggerAdapter) Warnf(format string, args ...interface{}) {
	tla.t.Logf("WARN: "+format, args...)
}
func (tla *testLoggerAdapter) Errorf(format string, args ...interface{}) {
	tla.t.Logf("ERROR: "+format, args...)
}
func (tla *testLoggerAdapter) Fatalf(format string, args ...interface{}) {
	tla.t.Fatalf("FATAL: "+format, args...)
}

// Helper function to execute a script and check a variable's state.
func checkVariableStateAfterSet(t *testing.T, script string, varName string, expectedValue interface{}, expectedType reflect.Type) {
	t.Helper()
	interpreter, _ := NewDefaultTestInterpreter(t)

	// This method is now assumed to exist in pkg/core/interpreter_scriptexec.go
	_, execErr := interpreter.ExecuteScriptString(fmt.Sprintf("test_autocreate_%s", varName), script, nil)

	if execErr != nil {
		// If the test *expects* an error for a specific setup, this check needs adjustment.
		// For auto-creation success tests, no execution error is expected.
		t.Fatalf("Script execution failed for '%s':\nScript:\n%s\nError: %s (Code: %d, Pos: %s, Wrapped: %v)",
			varName, script, execErr.Message, execErr.Code, execErr.Position, execErr.Wrapped)
		return
	}

	val, exists := interpreter.GetVariable(varName)
	if !exists {
		t.Errorf("Expected variable '%s' to exist after script execution, but it does not.\nScript:\n%s", varName, script)
		return
	}

	if expectedType != nil {
		valType := reflect.TypeOf(val)
		if valType != expectedType {
			t.Errorf("Variable '%s': expected type %v, got %v.\nScript:\n%s", varName, expectedType, valType, script)
		}
	}

	if !reflect.DeepEqual(val, expectedValue) {
		if expectedMap, okE := expectedValue.(map[string]interface{}); okE {
			if actualMap, okA := val.(map[string]interface{}); okA {
				if !reflect.DeepEqual(actualMap, expectedMap) {
					t.Errorf("Variable '%s': expected map value %#v, got %#v.\nScript:\n%s", varName, expectedMap, actualMap, script)
				}
			} else {
				t.Errorf("Variable '%s': expected a map, but got type %T with value %#v.\nScript:\n%s", varName, val, val, script)
			}
		} else if expectedSlice, okE := expectedValue.([]interface{}); okE {
			if actualSlice, okA := val.([]interface{}); okA {
				if !reflect.DeepEqual(actualSlice, expectedSlice) {
					t.Errorf("Variable '%s': expected slice value %#v, got %#v.\nScript:\n%s", varName, expectedSlice, actualSlice, script)
				}
			} else {
				t.Errorf("Variable '%s': expected a slice, but got type %T with value %#v.\nScript:\n%s", varName, val, val, script)
			}
		} else {
			t.Errorf("Variable '%s': expected value %#v (type %T), got %#v (type %T).\nScript:\n%s",
				varName, expectedValue, expectedValue, val, val, script)
		}
	}
}

func TestLValueAutoCreation(t *testing.T) {
	var expectedMapType = reflect.TypeOf(map[string]interface{}{})
	var expectedSliceType = reflect.TypeOf([]interface{}{})

	testCases := []struct {
		name          string
		script        string
		checkVarName  string
		expectedValue interface{}
		expectedType  reflect.Type
	}{
		{
			name:          "base map auto-creation with string key",
			script:        `set a["key1"] = "value1"`,
			checkVarName:  "a",
			expectedValue: map[string]interface{}{"key1": "value1"},
			expectedType:  expectedMapType,
		},
		{
			name:          "base list auto-creation with numeric index 0",
			script:        `set b[0] = "value0"`,
			checkVarName:  "b",
			expectedValue: []interface{}{"value0"},
			expectedType:  expectedSliceType,
		},
		{
			name:          "base list auto-creation with numeric index 2 (pads with nil)",
			script:        `set c[2] = "value2"`,
			checkVarName:  "c",
			expectedValue: []interface{}{nil, nil, "value2"},
			expectedType:  expectedSliceType,
		},
		{
			name:         "nested map auto-creation via dot access",
			script:       `set d.level1.level2 = "valueD"`,
			checkVarName: "d",
			expectedValue: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "valueD",
				},
			},
			expectedType: expectedMapType,
		},
		{
			name:         "nested map auto-creation via bracket access",
			script:       `set e["level1"]["level2"] = "valueE"`,
			checkVarName: "e",
			expectedValue: map[string]interface{}{
				"level1": map[string]interface{}{
					"level2": "valueE",
				},
			},
			expectedType: expectedMapType,
		},
		{
			name:         "nested list in map auto-creation",
			script:       `set f.listKey[1] = "item1"`,
			checkVarName: "f",
			expectedValue: map[string]interface{}{
				"listKey": []interface{}{nil, "item1"},
			},
			expectedType: expectedMapType,
		},
		{
			name:         "nested map in list auto-creation",
			script:       `set g[0].mapKey = "itemG"`,
			checkVarName: "g",
			expectedValue: []interface{}{
				map[string]interface{}{"mapKey": "itemG"},
			},
			expectedType: expectedSliceType,
		},
		{
			name:         "deeply nested auto-creation map-list-map-list",
			script:       `set h.maps[0].anotherMap["deepKey"][1] = "finalValue"`,
			checkVarName: "h",
			expectedValue: map[string]interface{}{
				"maps": []interface{}{
					map[string]interface{}{
						"anotherMap": map[string]interface{}{
							"deepKey": []interface{}{nil, "finalValue"},
						},
					},
				},
			},
			expectedType: expectedMapType,
		},
		{
			name:          "overwrite existing string with map on complex assignment",
			script:        "set k = \"i am a string\"\nset k[\"newKey\"] = \"now a map\"",
			checkVarName:  "k",
			expectedValue: map[string]interface{}{"newKey": "now a map"},
			expectedType:  expectedMapType,
		},
		{
			name:          "overwrite existing number with list on complex assignment",
			script:        "set l = 123\nset l[0] = \"now a list\"",
			checkVarName:  "l",
			expectedValue: []interface{}{"now a list"},
			expectedType:  expectedSliceType,
		},
		{
			name:         "dot access creates map then bracket access on it",
			script:       `set m.firstMap["secondKey"] = "valM"`,
			checkVarName: "m",
			expectedValue: map[string]interface{}{
				"firstMap": map[string]interface{}{
					"secondKey": "valM",
				},
			},
			expectedType: expectedMapType,
		},
		{
			name:         "bracket access creates map then dot access on it",
			script:       `set n["firstMap"].secondKey = "valN"`,
			checkVarName: "n",
			expectedValue: map[string]interface{}{
				"firstMap": map[string]interface{}{
					"secondKey": "valN",
				},
			},
			expectedType: expectedMapType,
		},
		{
			name:         "list auto-creation within auto-created map within auto-created list",
			script:       `set p[0]["listInMap"][1] = "complexP"`,
			checkVarName: "p",
			expectedValue: []interface{}{
				map[string]interface{}{
					"listInMap": []interface{}{nil, "complexP"},
				},
			},
			expectedType: expectedSliceType,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			checkVariableStateAfterSet(t, tc.script, tc.checkVarName, tc.expectedValue, tc.expectedType)
		})
	}
}
