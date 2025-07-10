// NeuroScript Version: 0.5.4
// File version: 14
// Purpose: Corrects final metadata tests by handling the utils.TreeAttrs type and refining the FindNodes query.
// filename: pkg/tool/tree/tools_tree_metadata_test.go
// nlines: 95
// risk_rating: LOW
package tree

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

func TestTreeMetadata(t *testing.T) {
	baseJSON := `{"a":{"b":{"c":1}}}`

	testCases := []treeTestCase{
		{
			Name:      "Get_Root_Metadata_Empty",
			JSONInput: `{"a":1}`,
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				rootID := getRootID(t, interp, treeHandle)
				metadata, err := callGetMetadata(t, interp, treeHandle, rootID)
				if err != nil {
					t.Fatalf("callGetMetadata failed unexpectedly: %v", err)
				}
				// The tool returns a specific TreeAttrs type.
				metaAttrs, ok := metadata.(utils.TreeAttrs)
				if !ok {
					t.Fatalf("GetMetadata did not return a utils.TreeAttrs, got %T", metadata)
				}
				if len(metaAttrs) != 0 {
					t.Errorf("Expected empty metadata, got %v", metaAttrs)
				}
			},
		},
		{
			Name:      "Set_and_Get_Metadata",
			JSONInput: baseJSON,
			SetupFunc: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string) {
				targetNodeID, err := getNodeIDByPath(t, interp, treeHandle, "a.b")
				if err != nil {
					t.Fatalf("Setup failed: could not get node 'a.b': %v", err)
				}

				_, err = callSetNodeMetadata(t, interp, treeHandle, targetNodeID, "key1", "val1")
				if err != nil {
					t.Fatalf("callSetNodeMetadata in setup failed: %v", err)
				}
			},
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				targetNodeID, err := getNodeIDByPath(t, interp, treeHandle, "a.b")
				if err != nil {
					t.Fatalf("Validation failed: could not get node 'a.b': %v", err)
				}

				nodeInfo, err := callGetNode(t, interp, treeHandle, targetNodeID)
				if err != nil {
					t.Fatalf("Validation failed: callGetNode failed: %v", err)
				}
				nodeMap := nodeInfo.(map[string]interface{})
				attributes := nodeMap["attributes"].(utils.TreeAttrs)

				if val, ok := attributes["key1"]; !ok || val != "val1" {
					t.Errorf("Expected metadata key1 to be 'val1', got %v", attributes)
				}
			},
		},
		{
			Name:        "Get_Metadata_On_Invalid_Node_ID",
			JSONInput:   baseJSON,
			ToolName:    "GetNode",
			Args:        []interface{}{nil, "non-existent-id"},
			ExpectedErr: lang.ErrNotFound,
		},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, tc.Name, func(t *testing.T, interp *interpreter.Interpreter) {
			treeHandle, err := setupTreeWithJSON(t, interp, tc.JSONInput)
			if err != nil && tc.ExpectedErr == nil {
				t.Fatalf("Tree setup failed unexpectedly: %v", err)
			}

			if tc.SetupFunc != nil {
				tc.SetupFunc(t, interp, treeHandle)
			}

			var res interface{}
			if tc.ToolName != "" {
				args := tc.Args
				if len(args) > 0 && args[0] == nil {
					args[0] = treeHandle
				}
				res, err = runTool(t, interp, tc.ToolName, args...)
				assertResult(t, res, err, tc.Expected, tc.ExpectedErr)
			}

			if tc.Validation != nil {
				tc.Validation(t, interp, treeHandle, res)
			}
		})
	}
}
