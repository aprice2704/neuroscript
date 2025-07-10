// NeuroScript Version: 0.5.4
// File version: 9
// Purpose: Corrects tree modification tests to use the precise tool names and handle-based API for all operations and validations.
// filename: pkg/tool/tree/tools_tree_modify_test.go
// nlines: 200
// risk_rating: LOW
package tree

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
)

func TestTreeModify(t *testing.T) {
	baseJSON := `{"a":{"b":{"c":1}},"d":[2,3]}`

	testCases := []treeTestCase{
		{
			Name:      "Add_Node_to_Root",
			JSONInput: baseJSON,
			ToolName:  "AddChildNode",
			// Args: treeHandle, parentNodeId, newNodeId, type, value, key
			Args:     []interface{}{nil, "placeholder_parent", "e", "number", float64(4), "e"},
			Expected: "e",
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				val, err := callGetValue(t, interp, treeHandle, result.(string))
				if err != nil {
					t.Fatalf("Validation failed: could not get value of node 'e': %v", err)
				}
				if val != float64(4) {
					t.Errorf("Expected node 'e' to have value 4, got %v", val)
				}
			},
		},
		{
			Name:      "Remove_Node_from_Child",
			JSONInput: baseJSON,
			ToolName:  "RemoveNode",
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				// To remove a.b, we need its actual ID first
				nodeID, err := getNodeIDByPath(t, interp, treeHandle, "a.b")
				if err != nil {
					t.Fatalf("Setup failed: could not get node 'a.b': %v", err)
				}
				// Now call RemoveNode with the correct ID
				_, err = runTool(t, interp, "RemoveNode", treeHandle, nodeID)
				if err != nil {
					t.Fatalf("Setup failed: RemoveNode failed unexpectedly: %v", err)
				}

				_, err = getNodeIDByPath(t, interp, treeHandle, "a.b")
				if err == nil {
					t.Error("Validation failed: expected error getting removed node 'a.b', but got nil")
				}
			},
		},
		{
			Name:      "Set_Value_on_Child",
			JSONInput: baseJSON,
			ToolName:  "SetValue",
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				// Get the ID for a.b.c to set its value
				nodeID, err := getNodeIDByPath(t, interp, treeHandle, "a.b.c")
				if err != nil {
					t.Fatalf("Setup failed: could not get node 'a.b.c': %v", err)
				}
				_, err = runTool(t, interp, "SetValue", treeHandle, nodeID, "new_value")
				if err != nil {
					t.Fatalf("Setup failed: SetValue failed unexpectedly: %v", err)
				}

				val, err := callGetValue(t, interp, treeHandle, nodeID)
				if err != nil {
					t.Fatalf("Validation failed: could not get value of 'a.b.c': %v", err)
				}
				if val != "new_value" {
					t.Errorf("Expected node 'a.b.c' value to be 'new_value', got %v", val)
				}
			},
		},
		{
			Name:      "Append_Child_to_Array",
			JSONInput: baseJSON,
			ToolName:  "AddChildNode",
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				// Get the ID for the array 'd'
				nodeID, err := getNodeIDByPath(t, interp, treeHandle, "d")
				if err != nil {
					t.Fatalf("Setup failed: could not get node 'd': %v", err)
				}
				// Add a new number node to the array
				_, err = callAddChildNode(t, interp, treeHandle, nodeID, "new_child", "number", float64(4), nil)
				if err != nil {
					t.Fatalf("Setup failed: AddChildNode failed unexpectedly: %v", err)
				}

				children, err := callGetChildren(t, interp, treeHandle, nodeID)
				if err != nil {
					t.Fatalf("Validation failed: could not get children of 'd': %v", err)
				}
				childIDs := children.([]interface{})
				if len(childIDs) != 3 {
					t.Fatalf("Expected 3 children for node 'd', got %d", len(childIDs))
				}
				lastChildVal, err := callGetValue(t, interp, treeHandle, childIDs[2].(string))
				if err != nil {
					t.Fatalf("Validation failed: could not get value of new child: %v", err)
				}
				if lastChildVal != float64(4) {
					t.Errorf("Expected new child to have value 4, got %v", lastChildVal)
				}
			},
		},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, tc.Name, func(t *testing.T, interp *interpreter.Interpreter) {
			treeHandle, err := setupTreeWithJSON(t, interp, tc.JSONInput)
			if err != nil {
				t.Fatalf("Tree setup failed unexpectedly: %v", err)
			}
			rootID := getRootID(t, interp, treeHandle)

			var result interface{}
			if tc.ToolName != "" && len(tc.Args) > 0 {
				// Replace nil placeholder with the handle if needed
				args := tc.Args
				if len(args) > 0 {
					if args[0] == nil {
						args[0] = treeHandle
					}
					if args[1] == "placeholder_parent" {
						args[1] = rootID
					}
				}
				result, err = runTool(t, interp, tc.ToolName, args...)
				assertResult(t, result, err, tc.Expected, tc.ExpectedErr)
			}

			// Run validation if it exists
			if tc.Validation != nil {
				tc.Validation(t, interp, treeHandle, result)
			}
		})
	}
}
