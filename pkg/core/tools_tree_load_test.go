// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Tests for Tree.LoadJSON and Tree.ToJSON tools.
// nlines: 90
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_load_test.go

package core

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestTreeLoadJSONAndToJSON(t *testing.T) {
	validJSONSimple := `{"key":"value","num":123}`
	validJSONNested := `{"a":[1,{"b":null}],"c":true}`
	validJSONArray := `[1,"two",true]`

	testCases := []treeTestCase{
		// Tree.LoadJSON
		{name: "LoadJSON Simple Object", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONSimple), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			handleStr, ok := result.(string)
			if !ok || !strings.HasPrefix(handleStr, GenericTreeHandleType+"::") {
				t.Errorf("Expected valid handle string, got %T: %v", result, result)
			}
		}},
		{name: "LoadJSON Nested Structure", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONNested), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Simple Array", toolName: "Tree.LoadJSON", args: MakeArgs(validJSONArray), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Empty Object", toolName: "Tree.LoadJSON", args: MakeArgs(`{}`), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Empty Array", toolName: "Tree.LoadJSON", args: MakeArgs(`[]`), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}
			if _, ok := result.(string); !ok {
				t.Errorf("Expected string handle, got %T", result)
			}
		}},
		{name: "LoadJSON Invalid JSON", toolName: "Tree.LoadJSON", args: MakeArgs(`{"key": "value`), wantToolErrIs: ErrTreeJSONUnmarshal},
		{name: "LoadJSON Empty Input", toolName: "Tree.LoadJSON", args: MakeArgs(``), wantToolErrIs: ErrTreeJSONUnmarshal},
		{name: "LoadJSON Wrong Arg Type", toolName: "Tree.LoadJSON", args: MakeArgs(123), valWantErrIs: ErrValidationTypeMismatch},

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

				var gotMap map[string]interface{}
				if errUnmarshal := json.Unmarshal([]byte(jsonStr), &gotMap); errUnmarshal != nil {
					t.Fatalf("ToJSON output could not be unmarshalled: %v. Output was: %s", errUnmarshal, jsonStr)
				}

				expectedMap := map[string]interface{}{
					"key": "value",
					"num": float64(123), // JSON numbers are float64 by default when unmarshalled into interface{}
				}
				if !reflect.DeepEqual(gotMap, expectedMap) {
					t.Errorf("ToJSON output mismatch after unmarshalling.\nGot:    %#v\nWanted: %#v\nOriginal JSON string: %s", gotMap, expectedMap, jsonStr)
				}
			}},
		{name: "ToJSON_Invalid_Handle", toolName: "Tree.ToJSON", args: MakeArgs("invalid-handle"), wantToolErrIs: ErrInvalidArgument},
		{name: "ToJSON_Handle_Not_Found", toolName: "Tree.ToJSON", args: MakeArgs(GenericTreeHandleType + "::non-existent-uuid"), wantToolErrIs: ErrHandleNotFound},
	}

	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}
