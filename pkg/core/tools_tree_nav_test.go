// NeuroScript Version: 0.3.1
// File version: 0.1.1
// Updated tests to expect "children" key from Tree.GetNode output.
// nlines: 115 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_nav_test.go

package core

import (
	"errors" // Added for DeepEqual
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
	rootHandle := setupTreeWithJSON(t, interp, jsonInput)
	rootNodeID := "node-1"

	testCases := []treeTestCase{
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
			if nodeMap["type"] != "object" {
				t.Errorf("GetNode Root: type mismatch, got %v, expected 'object'", nodeMap["type"])
			}
		}},
		{name: "GetNode Nested (file1 name)", toolName: "Tree.GetNode", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"]
				childrenArrMap, _ := callGetNode(t, interp, rootHandle, childrenArrNodeID)

				// MODIFIED: Expect "children" key from childrenArrMap (output of Tree.GetNode)
				childrenVal, valOk := childrenArrMap["children"]
				if !valOk {
					t.Fatalf("GetNode Nested: 'children' key not found in node map for children array. Map: %#v", childrenArrMap)
				}
				childNodeIDsRaw, ok := childrenVal.([]string)
				if !ok {
					// Try []interface{} and convert
					childNodeIDsIF, ifOk := childrenVal.([]interface{})
					if !ifOk {
						t.Fatalf("GetNode Nested: 'children' is not []string or []interface{}, but %T. Value: %#v", childrenArrMap["children"], childrenArrMap["children"])
					}
					childNodeIDsRaw = make([]string, len(childNodeIDsIF))
					for i, v := range childNodeIDsIF {
						childNodeIDsRaw[i], _ = v.(string)
					}
				}

				if len(childNodeIDsRaw) == 0 {
					t.Fatalf("GetNode Nested: children list is empty")
				}
				file1ObjNodeID := childNodeIDsRaw[0]

				file1ObjMap, _ := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				fileNameNodeID := file1ObjMap["attributes"].(map[string]string)["name"]

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
		{name: "GetNode Invalid Handle", toolName: "Tree.GetNode", args: MakeArgs("bad-handle", "node-1"), wantToolErrIs: ErrInvalidArgument}, // Assuming ErrInvalidArgument for bad handle format

		{name: "GetChildren of Array Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"]
				childrenIDs, errGet := callGetChildren(t, interp, rootHandle, childrenArrNodeID)
				if errGet != nil {
					t.Fatalf("GetChildren for array failed: %v", errGet)
				}
				if len(childrenIDs) != 2 {
					t.Errorf("Expected 2 children, got %d", len(childrenIDs))
				}
				// Example of further check:
				// _, nodeErr1 := callGetNode(t, interp, rootHandle, childrenIDs[0].(string))
				// if nodeErr1 != nil { t.Errorf("Could not get child node 0: %v", nodeErr1)}
			}},
		{name: "GetChildren of Object Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, rootNodeID), wantToolErrIs: ErrNodeWrongType},
		{name: "GetChildren of Leaf Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"]
				childrenArrMap, _ := callGetNode(t, interp, rootHandle, childrenArrNodeID)

				// MODIFIED: Expect "children" key
				childrenVal, valOk := childrenArrMap["children"]
				if !valOk {
					t.Fatalf("Setup for GetChildren of Leaf: 'children' key not found. Map: %#v", childrenArrMap)
				}
				childNodeIDsRaw, ok := childrenVal.([]string)
				if !ok {
					childNodeIDsIF, ifOk := childrenVal.([]interface{})
					if !ifOk {
						t.Fatalf("Setup for GetChildren of Leaf: could not get 'children' as []string or []interface{}. Got %T", childrenArrMap["children"])
					}
					childNodeIDsRaw = make([]string, len(childNodeIDsIF))
					for i, v := range childNodeIDsIF {
						childNodeIDsRaw[i], _ = v.(string)
					}
				}
				if len(childNodeIDsRaw) == 0 {
					t.Fatalf("Setup for GetChildren of Leaf: 'children' list is empty.")
				}

				file1ObjNodeID := childNodeIDsRaw[0]
				file1ObjMap, _ := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				fileNameNodeID := file1ObjMap["attributes"].(map[string]string)["name"]

				_, errGet := callGetChildren(t, interp, rootHandle, fileNameNodeID)
				if !errors.Is(errGet, ErrNodeWrongType) {
					t.Fatalf("GetChildren for leaf node expected ErrNodeWrongType, got %v", errGet)
				}
			}},

		{name: "GetParent of Root", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, rootNodeID), wantResult: nil},
		{name: "GetParent of Child", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(map[string]string)["children"]

				getParentTool, _ := interp.ToolRegistry().GetTool("Tree.GetParent")
				parentIDResult, errGet := getParentTool.Func(interp, MakeArgs(rootHandle, childrenArrNodeID))
				if errGet != nil {
					t.Fatalf("GetParent failed: %v", errGet)
				}
				parentID, ok := parentIDResult.(string)
				if !ok && parentIDResult != nil {
					t.Fatalf("GetParent did not return a string or nil, got %T: %v", parentIDResult, parentIDResult)
				}
				if parentID != rootNodeID {
					t.Errorf("Expected parent %s, got %v", rootNodeID, parentID)
				}
			}},
		{name: "GetParent NonExistent Node", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, "node-999"), wantToolErrIs: ErrNotFound},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, interp, tc)
	}
}
