// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Tests for tree modification tools (SetValue, AddChildNode, RemoveNode).
// nlines: 80
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_modify_test.go

package core

import (
	"errors"
	"testing"
)

func TestTreeModificationTools(t *testing.T) {
	jsonSimple := `{"name": "item", "value": 10}` // "name" is string, "value" is number

	setupInitialTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonSimple)
	}
	// Node IDs from this JSON after load (example):
	// node-1 (object for root) Attributes: {"name": "node-2", "value": "node-3"}
	// node-2 (string "item")
	// node-3 (number 10)

	testCases := []treeTestCase{
		// Tree.SetValue
		{name: "SetValue Valid Leaf (string node)", toolName: "Tree.SetValue", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-2", "new_name_value"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("SetValue Valid Leaf failed: %v", err)
				}
				handle := ctx.(string)
				nodeMap, errGet := callGetNode(t, interp, handle, "node-2")
				if errGet != nil {
					t.Fatalf("CheckFunc: Failed to get node-2 after SetValue: %v", errGet)
				}
				if nodeMap["value"] != "new_name_value" {
					t.Errorf("SetValue did not update node. Got: %v, want 'new_name_value'", nodeMap["value"])
				}
			}},
		{name: "SetValue On Object Node", toolName: "Tree.SetValue", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-1", "should_fail"), wantToolErrIs: ErrCannotSetValueOnType},
		{name: "SetValue NonExistent Node", toolName: "Tree.SetValue", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-999", "val"), wantToolErrIs: ErrNotFound},

		// Tree.AddChildNode
		{name: "AddChildNode To Root Object", toolName: "Tree.AddChildNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-1", "newChild1", "string", "hello", "newKeyInRoot"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("AddChildNode failed: %v", err)
				}
				newID, ok := result.(string)
				if !ok || newID == "" {
					t.Fatalf("AddChildNode did not return new node ID string, got %T: %v", result, result)
				}
				// Note: The tool itself might or might not enforce newID == providedNewNodeID.
				// The primary check is that a node was added and the parent points to it.
				// If the tool returns the ID it used (even if it generated one), that's fine.
				// For this test, we assume the provided newID "newChild1" is used if valid.

				handle := ctx.(string)
				parentNodeMap, _ := callGetNode(t, interp, handle, "node-1")
				attrs := parentNodeMap["attributes"].(map[string]string)
				if attrs["newKeyInRoot"] != newID { // Check parent attribute points to the new child's ID
					t.Errorf("AddChildNode: newKeyInRoot not pointing to new child ID %s. Attrs: %v", newID, attrs)
				}
				childNodeMap, _ := callGetNode(t, interp, handle, newID) // Use the ID returned by the tool
				if childNodeMap["value"] != "hello" {
					t.Errorf("Added child has wrong value: %v", childNodeMap["value"])
				}
				if childNodeMap["id"] != newID { // The node itself should have this ID
					t.Errorf("Added child has wrong ID in its map: %s, expected %s", childNodeMap["id"], newID)
				}
			}},
		{name: "AddChildNode ID Exists", toolName: "Tree.AddChildNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-1", "node-2", "string", "fail", "anotherKey"), wantToolErrIs: ErrNodeIDExists},
		{name: "AddChildNode To Leaf Node", toolName: "Tree.AddChildNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-2", "newChild2", "string", "val", "key"), wantToolErrIs: ErrNodeWrongType}, // node-2 is string

		// Tree.RemoveNode
		{name: "RemoveNode Leaf", toolName: "Tree.RemoveNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-3"), // Remove value node (number 10)
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("RemoveNode Leaf failed: %v", err)
				}
				handle := ctx.(string)
				_, errGet := callGetNode(t, interp, handle, "node-3")
				if !errors.Is(errGet, ErrNotFound) {
					t.Errorf("Expected ErrNotFound after removing node-3, got %v", errGet)
				}
				rootMap, _ := callGetNode(t, interp, handle, "node-1")
				attrs := rootMap["attributes"].(map[string]string)
				if _, exists := attrs["value"]; exists { // "value" was the key for node-3
					t.Errorf("Attribute 'value' still exists on root after removing child node-3")
				}
			}},
		{name: "RemoveNode Root", toolName: "Tree.RemoveNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-1"), wantToolErrIs: ErrCannotRemoveRoot},
		{name: "RemoveNode NonExistent", toolName: "Tree.RemoveNode", setupFunc: setupInitialTree, args: MakeArgs("SETUP_HANDLE:tree1", "node-999"), wantToolErrIs: ErrNotFound},
	}
	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t) // Use fresh interp for modification tests
		testTreeToolHelper(t, currentInterp, tc)
	}
}
