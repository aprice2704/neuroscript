// NeuroScript Version: 0.3.1
// File version: 3
// Purpose: Corrected type assertions to use the specific `TreeAttrs` type instead of a generic map, resolving the test panic.
// filename: pkg/core/tools_tree_nav_test.go
// nlines: 118
// risk_rating: MEDIUM

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
				childrenArrNodeID := rootMap["attributes"].(TreeAttrs)["children"].(string)
				childrenArrMap, _ := callGetNode(t, interp, rootHandle, childrenArrNodeID)

				childrenVal, valOk := childrenArrMap["children"]
				if !valOk {
					t.Fatalf("GetNode Nested: 'children' key not found in node map for children array. Map: %#v", childrenArrMap)
				}

				var childNodeIDsRaw []string
				childNodeIDsIF, ifOk := childrenVal.([]interface{})
				if !ifOk {
					t.Fatalf("GetNode Nested: 'children' is not []interface{}, but %T. Value: %#v", childrenArrMap["children"], childrenArrMap["children"])
				}
				childNodeIDsRaw = make([]string, len(childNodeIDsIF))
				for i, v := range childNodeIDsIF {
					childNodeIDsRaw[i], _ = v.(string)
				}

				if len(childNodeIDsRaw) == 0 {
					t.Fatalf("GetNode Nested: children list is empty")
				}
				file1ObjNodeID := childNodeIDsRaw[0]

				file1ObjMap, _ := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				fileNameNodeID := file1ObjMap["attributes"].(TreeAttrs)["name"].(string)

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
		{name: "GetNode NonExistent Node", toolName: "Tree.GetNode", args: MakeArgs(rootHandle, "node-999"), wantErr: ErrNotFound},
		{name: "GetNode Invalid Handle", toolName: "Tree.GetNode", args: MakeArgs("bad-handle", "node-1"), wantErr: ErrInvalidArgument},

		{name: "GetChildren of Array Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(TreeAttrs)["children"].(string)
				childrenIDs, errGet := callGetChildren(t, interp, rootHandle, childrenArrNodeID)
				if errGet != nil {
					t.Fatalf("GetChildren for array failed: %v", errGet)
				}
				if len(childrenIDs) != 2 {
					t.Errorf("Expected 2 children, got %d", len(childrenIDs))
				}
			}},
		{name: "GetChildren of Object Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, rootNodeID), wantErr: ErrNodeWrongType},
		{name: "GetChildren of Leaf Node", toolName: "Tree.GetChildren", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(TreeAttrs)["children"].(string)
				childrenArrMap, _ := callGetNode(t, interp, rootHandle, childrenArrNodeID)

				childrenVal, valOk := childrenArrMap["children"]
				if !valOk {
					t.Fatalf("Setup for GetChildren of Leaf: 'children' key not found. Map: %#v", childrenArrMap)
				}
				var childNodeIDsRaw []string
				childNodeIDsIF, ifOk := childrenVal.([]interface{})
				if !ifOk {
					t.Fatalf("Setup for GetChildren of Leaf: could not get 'children' as []interface{}. Got %T", childrenArrMap["children"])
				}
				for _, v := range childNodeIDsIF {
					childNodeIDsRaw = append(childNodeIDsRaw, v.(string))
				}

				if len(childNodeIDsRaw) == 0 {
					t.Fatalf("Setup for GetChildren of Leaf: 'children' list is empty.")
				}

				file1ObjNodeID := childNodeIDsRaw[0]
				file1ObjMap, _ := callGetNode(t, interp, rootHandle, file1ObjNodeID)
				fileNameNodeID := file1ObjMap["attributes"].(TreeAttrs)["name"].(string)

				_, errGet := callGetChildren(t, interp, rootHandle, fileNameNodeID)
				if !errors.Is(errGet, ErrNodeWrongType) {
					t.Fatalf("GetChildren for leaf node expected ErrNodeWrongType, got %v", errGet)
				}
			}},

		{name: "GetParent of Root", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, rootNodeID), checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, setupCtx interface{}) {
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if result != nil {
				t.Errorf("Expected nil result for parent of root, got %v", result)
			}
		}},
		{name: "GetParent of Child", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, ""),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				rootMap, _ := callGetNode(t, interp, rootHandle, rootNodeID)
				childrenArrNodeID := rootMap["attributes"].(TreeAttrs)["children"].(string)

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
		{name: "GetParent NonExistent Node", toolName: "Tree.GetParent", args: MakeArgs(rootHandle, "node-999"), wantErr: ErrNotFound},
	}

	for _, tc := range testCases {
		testTreeToolHelper(t, interp, tc)
	}
}
