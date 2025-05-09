// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Tests for tree navigation tools (GetNode, GetChildren, GetParent).
// nlines: 115
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_nav_test.go

package core

import (
	"errors"
	"testing"
)

func TestTreeNavigationTools(t *testing.T) {
	interp, _ := NewDefaultTestInterpreter(t)
	jsonInput := `{
		"name": "root_obj",
		"type": "directory",
		"children": [
			{"name": "file1.txt", "type": "file", "size": 100},
			{"name": "subdir", "type": "directory", "children": [
				{"name": "file2.txt", "type": "file", "size": 50}
			]}
		],
		"metadata": {"owner": "admin"}
	}`
	rootHandle := setupTreeWithJSON(t, interp, jsonInput) // Load once for all nav tests
	rootNodeID := "node-1"                                // Assume root is node-1 based on current LoadJSON behavior

	testCases := []treeTestCase{
		// Tree.GetNode
		{name: "GetNode Root", toolName: "Tree.GetNode", args: MakeArgs(rootHandle, rootNodeID), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
			if err != nil {
				t.Fatalf("GetNode Root failed: %v", err)
			}
			nodeMap, ok := result.(map[string]interface{})
			if !ok {
				t.Fatalf("GetNode Root: expected map, got %T", result)
			}
			if nodeMap["id"] != rootNodeID {
				t.Errorf("GetNode Root: ID mismatch, got %v", nodeMap["id"])
			}
			if nodeMap["type"] != "object" { // The root of the JSON is an object
				t.Errorf("GetNode Root: type mismatch, got %v, expected 'object'", nodeMap["type"])
			}
		}},
		{name: "GetNode Nested (file1 name)", toolName: "Tree.GetNode", args: MakeArgs(rootHandle, ""), // Node ID determined dynamically in checkFunc
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) { // result and err from static args are ignored
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"]
				childrenArrMap, _ := callGetNode(t, interp, rootHandle, childrenArrNodeID)

				childNodeIDs, ok := childrenArrMap["child_ids"].([]string)
				if !ok {
					t.Fatalf("GetNode Nested: child_ids is not []string, but %T. Value: %#v", childrenArrMap["child_ids"], childrenArrMap["child_ids"])
				}
				if len(childNodeIDs) == 0 {
					t.Fatalf("GetNode Nested: child_ids is empty")
				}
				file1ObjNodeID := childNodeIDs[0] // This is the ID of the first child object {"name": "file1.txt", ...}

				file1ObjMap, _ := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				fileNameNodeID := file1ObjMap["attributes"].(map[string]string)["name"]

				// This is the actual node being tested by the test case name
				nodeMap, errGet := callGetNode(t, interp, rootHandle, fileNameNodeID)
				if errGet != nil {
					t.Fatalf("GetNode for file1 name failed: %v", errGet)
				}
				if nodeMap["value"] != "file1.txt" {
					t.Errorf("Expected 'file1.txt', got %v", nodeMap["value"])
				}
				if nodeMap["type"] != "string" {
					t.Errorf("Expected type string, got %v", nodeMap["type"])
				}
			}},
		{name: "GetNode NonExistent Node", toolName: "Tree.GetNode", args: MakeArgs(rootHandle, "node-999"), wantToolErrIs: ErrNotFound},
		{name: "GetNode Invalid Handle", toolName: "Tree.GetNode", args: MakeArgs("bad-handle", "node-1"), wantToolErrIs: ErrInvalidArgument},

		// Tree.GetChildren
		{name: "GetChildren of Array Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, ""), // Node ID determined dynamically
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"] // ID of the array node
				childrenIDs, errGet := callGetChildren(t, interp, rootHandle, childrenArrNodeID)
				if errGet != nil {
					t.Fatalf("GetChildren for array failed: %v", errGet)
				}
				if len(childrenIDs) != 2 { // The "children" array in JSON has 2 elements
					t.Errorf("Expected 2 children, got %d", len(childrenIDs))
				}
			}},
		{name: "GetChildren of Object Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, rootNodeID), wantToolErrIs: ErrNodeWrongType}, // Cannot get children of object
		{name: "GetChildren of Leaf Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, ""), // Node ID for a string leaf node
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"]
				childrenArrMap, _ := callGetNode(t, interp, rootHandle, childrenArrNodeID)
				childNodeIDs, ok := childrenArrMap["child_ids"].([]string)
				if !ok || len(childNodeIDs) == 0 {
					t.Fatalf("Setup for GetChildren of Leaf: could not get child_ids as []string")
				}
				file1ObjNodeID := childNodeIDs[0]
				file1ObjMap, _ := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				fileNameNodeID := file1ObjMap["attributes"].(map[string]string)["name"] // This is the ID of the string node "file1.txt"

				_, errGet := callGetChildren(t, interp, rootHandle, fileNameNodeID) // A string node is not an array
				if !errors.Is(errGet, ErrNodeWrongType) {
					t.Fatalf("GetChildren for leaf node expected ErrNodeWrongType, got %v", errGet)
				}
			}},

		// Tree.GetParent
		{name: "GetParent of Root", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, rootNodeID), wantResult: nil}, // Root has no parent ID
		{name: "GetParent of Child", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, ""), // Node ID determined dynamically
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"] // ID of the array node

				getParentTool, _ := interp.ToolRegistry().GetTool("Tree.GetParent")
				parentIDResult, errGet := getParentTool.Func(interp, MakeArgs(rootHandle, childrenArrNodeID))
				if errGet != nil {
					t.Fatalf("GetParent failed: %v", errGet)
				}
				parentID, ok := parentIDResult.(string)
				if !ok && parentIDResult != nil { // Allow nil if that's what it means, but here we expect rootNodeID
					t.Fatalf("GetParent did not return a string or nil, got %T: %v", parentIDResult, parentIDResult)
				}
				if parentID != rootNodeID {
					t.Errorf("Expected parent %s, got %v", rootNodeID, parentID)
				}
			}},
		{name: "GetParent NonExistent Node", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, "node-999"), wantToolErrIs: ErrNotFound},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, interp, tc) // Use the same interp for nav tests as state is read-only after initial load
	}
}
