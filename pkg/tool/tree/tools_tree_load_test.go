// NeuroScript Version: 0.4.0
// File version: 1
// Purpose: Refactored to use the new primitive-based tree test helper.
// filename: pkg/tool/tree/tools_tree_load_test.go
// nlines: 75
// risk_rating: MEDIUM

package tree

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

func TestTreeLoadJSONAndToJSON(t *testing.T) {
	validJSONSimple := `{"key":"value","num":123}`
	// validJSONNested := `{"a":[1,{"b":null}],"c":true}` // This was unused

	testCases := []treeTestCase{
		// Tree.LoadJSON
		{name: "LoadJSON Simple Object", toolName: "Tree.LoadJSON", args: tool.MakeArgs(validJSONSimple), checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, _ interface{}) {
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			} else if handleStr, ok := result.(string); !ok || !strings.HasPrefix(handleStr, utils.GenericTreeHandleType+"::") {
				t.Errorf("Expected valid handle string, got %T: %v", result, result)
			}
		}},
		{name: "LoadJSON Invalid JSON", toolName: "Tree.LoadJSON", args: tool.MakeArgs(`{"key": "value`), wantErr: lang.ErrTreeJSONUnmarshal},
		{name: "LoadJSON Empty Input", toolName: "Tree.LoadJSON", args: tool.MakeArgs(``), wantErr: lang.ErrTreeJSONUnmarshal},
		{name: "LoadJSON Wrong Arg Type", toolName: "Tree.LoadJSON", args: tool.MakeArgs(123), wantErr: lang.ErrInvalidArgument},

		// Tree.ToJSON
		{name: "ToJSON Simple Object", toolName: "Tree.ToJSON",
			setupFunc: func(t *testing.T, interp tool.RunTime) interface{} {
				return setupTreeWithJSON(t, interp, validJSONSimple)
			},
			args: tool.MakeArgs("SETUP_HANDLE:tree1"), // Placeholder replaced by setupFunc result
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, _ interface{}) {
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
		{name: "ToJSON_Invalid_Handle", toolName: "Tree.ToJSON", args: tool.MakeArgs("invalid-handle"), wantErr: lang.ErrInvalidArgument},
		{name: "ToJSON_Handle_Not_Found", toolName: "Tree.ToJSON", args: tool.MakeArgs(utils.GenericTreeHandleType + "::non-existent-uuid"), wantErr: lang.ErrHandleNotFound},
	}

	for _, tc := range testCases {
		currentInterp, _ := llm.NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}
