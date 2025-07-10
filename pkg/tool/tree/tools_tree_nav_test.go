// NeuroScript Version: 0.5.4
// File version: 14
// Purpose: Corrects the final compiler error by handling special test setup directly within the test loop, avoiding the scoping issue.
// filename: pkg/tool/tree/tools_tree_nav_test.go
// nlines: 145
// risk_rating: LOW
package tree

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

func TestTreeNav(t *testing.T) {
	baseJSON := `{
		"root": {
			"children": [
				{"name": "child1", "value": 1},
				{"name": "child2", "value": 2, "children": [
					{"name": "grandchild1", "value": 2.1}
				]}
			],
			"metadata": {"type": "parent"}
		}
	}`

	testCases := []treeTestCase{
		{
			Name:      "Get_Node_Root",
			JSONInput: baseJSON,
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				rootID := getRootID(t, interp, treeHandle)
				rootNode, err := callGetNode(t, interp, treeHandle, rootID)
				if err != nil {
					t.Fatalf("Validation failed: could not get root node: %v", err)
				}
				if !reflect.DeepEqual(rootNode, getRootNode(t, interp, treeHandle)) {
					t.Error("GetNode with path 'root' did not return the root node handle")
				}
			},
		},
		{
			Name:      "Get_Node_Child",
			JSONInput: baseJSON,
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				nodeID, err := getNodeIDByPath(t, interp, treeHandle, "root.children.0")
				if err != nil {
					t.Fatalf("Validation failed: could not get node id for 'root.children.0': %v", err)
				}
				node, err := callGetNode(t, interp, treeHandle, nodeID)
				if err != nil {
					t.Fatalf("Validation failed: could not get value of node: %v", err)
				}
				nodeMap, ok := node.(map[string]interface{})
				if !ok {
					t.Fatalf("node is not a map, but %T", node)
				}
				attributes, ok := nodeMap["attributes"].(utils.TreeAttrs)
				if !ok {
					t.Fatalf("attributes is not utils.TreeAttrs, but %T", nodeMap["attributes"])
				}

				nameNodeID := attributes["name"].(string)
				nameNodeValue, err := callGetValue(t, interp, treeHandle, nameNodeID)
				if err != nil {
					t.Fatalf("could not get name node value: %v", err)
				}

				if nameNodeValue != "child1" {
					t.Errorf("Expected node name to be 'child1', got %v", nameNodeValue)
				}
			},
		},
		{
			Name:        "Get_Node_Invalid_Path",
			JSONInput:   baseJSON,
			ToolName:    "GetNode",
			Args:        []interface{}{nil, "non-existent-id"},
			ExpectedErr: lang.ErrNotFound,
		},
		{
			Name:      "Get_Children",
			JSONInput: baseJSON,
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				nodeID, err := getNodeIDByPath(t, interp, treeHandle, "root.children")
				if err != nil {
					t.Fatalf("could not get node id for %s: %v", "root.children", err)
				}
				children, err := callGetChildren(t, interp, treeHandle, nodeID)
				if err != nil {
					t.Fatalf("GetChildren failed: %v", err)
				}

				childSlice, ok := children.([]interface{})
				if !ok {
					t.Fatalf("GetChildren did not return a slice, got %T", children)
				}
				if len(childSlice) != 2 {
					t.Errorf("Expected 2 children for 'root.children', got %d", len(childSlice))
				}
			},
		},
		{
			Name:      "Get_Parent",
			JSONInput: baseJSON,
			Validation: func(t *testing.T, interp *interpreter.Interpreter, treeHandle string, result interface{}) {
				childID, err := getNodeIDByPath(t, interp, treeHandle, "root.children.0")
				if err != nil {
					t.Fatalf("Setup for Get_Parent failed: could not get child node: %v", err)
				}
				parentHandle, err := runTool(t, interp, "GetParent", treeHandle, childID)
				if err != nil {
					t.Fatalf("GetParent failed: %v", err)
				}

				parentID, ok := parentHandle.(string)
				if !ok {
					t.Fatalf("GetParent did not return a string handle, got %T", parentHandle)
				}
				children, err := callGetChildren(t, interp, treeHandle, parentID)
				if err != nil {
					t.Fatalf("Validation failed: could not get children of parent: %v", err)
				}
				if len(children.([]interface{})) != 2 {
					t.Errorf("Expected parent to have 2 children, got %d", len(children.([]interface{})))
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

			var result interface{}
			if tc.Validation != nil {
				tc.Validation(t, interp, treeHandle, result)
			} else if tc.ToolName != "" {
				args := tc.Args
				if len(args) > 0 {
					if args[0] == nil {
						args[0] = treeHandle
					}
				}
				result, err = runTool(t, interp, tc.ToolName, args...)
				assertResult(t, result, err, tc.Expected, tc.ExpectedErr)
			}
		})
	}
}
