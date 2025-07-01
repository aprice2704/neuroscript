// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to use the new primitive-based tree test helper.
// filename: pkg/core/tools_tree_load_test.go
// nlines: 75
// risk_rating: MEDIUM

package core

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTreeLoadJSONAndToJSON(t *testing.T) {
	validJSONSimple := `{"key":"value","num":123}`
	// validJSONNested := `{"a":[1,{"b":null}],"c":true}` // This was unused

	testCases := []treeTestCase{
		// Tree.LoadJSON
		{name: "LoadJSON Simple Object", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONSimple), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if handleStr, ok := result.(string); !ok || !strings.HasPrefix(handleStr, GenericTreeHandleType+"::") {
				t.Errorf("Expected valid handle string, got %T: %v", result, result)
			}
		}},
		{name: "LoadJSON Invalid JSON", toolName: "Tree.LoadJSON", args: MakeArgs(`{"key": "value`), wantErr: ErrTreeJSONUnmarshal},
		{name: "LoadJSON Empty Input", toolName: "Tree.LoadJSON", args: MakeArgs(``), wantErr: ErrTreeJSONUnmarshal},
		{name: "LoadJSON Wrong Arg Type", toolName: "Tree.LoadJSON", args: MakeArgs(123), wantErr: ErrInvalidArgument},

		// Tree.ToJSON
		{name: "ToJSON Simple Object", toolName: "Tree.ToJSON",
			setupFunc: func(t *testing.T, interp *Interpreter) interface{} {
				return setupTreeWithJSON(t, interp, validJSONSimple)
			},
			args: MakeArgs("SETUP_HANDLE:tree1"), // Placeholder replaced by setupFunc result
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
				if err != nil {
					t.Fatalf("ToJSON failed: %v", err)
				}
				jsonStr, ok := result.(string)
				if !ok {
					t.Fatalf("ToJSON did not return a string, got %T: %v", result, result)
				}
				var gotMap, expectedMap map[string]interface{}
				_ = json.Unmarshal([]byte(jsonStr), &gotMap)
				_ = json.Unmarshal([]byte(validJSONSimple), &expectedMap)
				if !reflect.DeepEqual(gotMap, expectedMap) {
					t.Errorf("ToJSON output mismatch after unmarshalling.\nGot:    %#v\nWanted: %#v", gotMap, expectedMap)
				}
			}},
		{name: "ToJSON_Invalid_Handle", toolName: "Tree.ToJSON", args: MakeArgs("invalid-handle"), wantErr: ErrInvalidArgument},
		{name: "ToJSON_Handle_Not_Found", toolName: "Tree.ToJSON", args: MakeArgs(GenericTreeHandleType + "::non-existent-uuid"), wantErr: ErrHandleNotFound},
	}

	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}
