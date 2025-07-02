// NeuroScript Version: 0.3.1
// File version: 0.0.5
// Purpose: Updated tests to expect core.Value types instead of native Go types.
// filename: pkg/runtime/interpreter_assignment_autocreate_test.go
// nlines: 325
// risk_rating: LOW

package runtime

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// Helper function to execute a script and check a variable's state against a Value type.
func checkVariableStateAfterSet(t *testing.T, script string, varName string, expectedValue lang.Value) {
	t.Helper()
	interpreter, _ := llm.NewDefaultTestInterpreter(t)

	_, execErr := interpreter.ExecuteScriptString(fmt.Sprintf("test_autocreate_%s", varName), script, nil)

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
			name:         "base list auto-creation with numeric index 2 (pads with nil)",
			script:       `set c[2] = "value2"`,
			checkVarName: "c",
			expectedValue: lang.NewListValue([]lang.Value{
				lang.NilValue{},
				lang.NilValue{},
				lang.StringValue{Value: "value2"},
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
			name:         "nested map auto-creation via bracket access",
			script:       `set e["level1"]["level2"] = "valueE"`,
			checkVarName: "e",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"level1": lang.NewMapValue(map[string]lang.Value{
					"level2": lang.StringValue{Value: "valueE"},
				}),
			}),
		},
		{
			name:         "nested list in map auto-creation",
			script:       `set f.listKey[1] = "item1"`,
			checkVarName: "f",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"listKey": lang.NewListValue([]lang.Value{
					lang.NilValue{},
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
		{
			name:         "deeply nested auto-creation map-list-map-list",
			script:       `set h.maps[0].anotherMap["deepKey"][1] = "finalValue"`,
			checkVarName: "h",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"maps": lang.NewListValue([]lang.Value{
					lang.NewMapValue(map[string]lang.Value{
						"anotherMap": lang.NewMapValue(map[string]lang.Value{
							"deepKey": lang.NewListValue([]lang.Value{
								lang.NilValue{},
								lang.StringValue{Value: "finalValue"},
							}),
						}),
					}),
				}),
			}),
		},
		{
			name:         "overwrite existing string with map on complex assignment",
			script:       "set k = \"i am a string\"\nset k[\"newKey\"] = \"now a map\"",
			checkVarName: "k",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"newKey": lang.StringValue{Value: "now a map"},
			}),
		},
		{
			name:         "overwrite existing number with list on complex assignment",
			script:       "set l = 123\nset l[0] = \"now a list\"",
			checkVarName: "l",
			expectedValue: lang.NewListValue([]lang.Value{
				lang.StringValue{Value: "now a list"},
			}),
		},
		{
			name:         "dot access creates map then bracket access on it",
			script:       `set m.firstMap["secondKey"] = "valM"`,
			checkVarName: "m",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"firstMap": lang.NewMapValue(map[string]lang.Value{
					"secondKey": lang.StringValue{Value: "valM"},
				}),
			}),
		},
		{
			name:         "bracket access creates map then dot access on it",
			script:       `set n["firstMap"].secondKey = "valN"`,
			checkVarName: "n",
			expectedValue: lang.NewMapValue(map[string]lang.Value{
				"firstMap": lang.NewMapValue(map[string]lang.Value{
					"secondKey": lang.StringValue{Value: "valN"},
				}),
			}),
		},
		{
			name:         "list auto-creation within auto-created map within auto-created list",
			script:       `set p[0]["listInMap"][1] = "complexP"`,
			checkVarName: "p",
			expectedValue: lang.NewListValue([]lang.Value{
				lang.NewMapValue(map[string]lang.Value{
					"listInMap": lang.NewListValue([]lang.Value{
						lang.NilValue{},
						lang.StringValue{Value: "complexP"},
					}),
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
