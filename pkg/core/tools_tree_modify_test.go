// NeuroScript Version: 0.3.1
// File version: 4
// Purpose: Corrected type assertions to use TreeAttrs instead of a generic map, resolving test failures.
// nlines: ~128
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_modify_test.go

package core

import (
	"errors" // Added for robust SetValue test logging
	"testing"
)

// findNodeIDByRootAttribute finds the ID of a child node linked by a specific attribute key on a given root/parent node.
func findNodeIDByRootAttribute(t *testing.T, interp *Interpreter, handle string, parentNodeID string, attributeKey string) (string, bool) {
	t.Helper()
	parentNodeData, err := callGetNode(t, interp, handle, parentNodeID)
	if err != nil {
		t.Logf("findNodeIDByRootAttribute: Failed to get parent node '%s': %v", parentNodeID, err)
		return "", false
	}

	attributes, ok := parentNodeData["attributes"].(TreeAttrs)
	if !ok {
		if parentNodeData["attributes"] == nil {
			t.Logf("findNodeIDByRootAttribute: Parent node '%s' has no 'attributes' field or it's nil.", parentNodeID)
		} else {
			t.Logf("findNodeIDByRootAttribute: Parent node '%s' attributes field is not TreeAttrs, got %T", parentNodeID, parentNodeData["attributes"])
		}
		return "", false
	}

	nodeIDUntyped, found := attributes[attributeKey]
	if !found {
		t.Logf("findNodeIDByRootAttribute: Attribute key '%s' not found on parent node '%s'. Attributes: %v", attributeKey, parentNodeID, attributes)
		return "", false
	}
	nodeID, ok := nodeIDUntyped.(string)
	if !ok {
		t.Logf("findNodeIDByRootAttribute: attribute '%s' is not a string, but %T", attributeKey, nodeIDUntyped)
		return "", false
	}
	return nodeID, true
}

func TestTreeModificationTools(t *testing.T) {
	jsonSimple := `{"name": "item", "value": 10}`

	setupInitialTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonSimple)
	}

	testCases := []treeTestCase{
		{
			name:      "SetValue Valid Leaf (string node) (Robust)",
			toolName:  "Tree.GetNode",
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"),
			setupFunc: setupInitialTree,
			checkFunc: func(t *testing.T, interp *Interpreter, initialToolResult interface{}, initialToolErr error, ctx interface{}) {
				handle := ctx.(string)
				newValueToSet := "new_name_value"

				targetNodeID, found := findNodeIDByRootAttribute(t, interp, handle, "node-1", "name")
				if !found {
					rootNodeDataForDebug, _ := callGetNode(t, interp, handle, "node-1")
					t.Fatalf("SetValue check: Could not find node linked by 'name' attribute of root. Root data: %v", rootNodeDataForDebug)
				}
				t.Logf("SetValue check: Dynamically found target node ID (for original 'name') as: %s", targetNodeID)

				setValueTool, toolFound := interp.ToolRegistry().GetTool("Tree.SetValue")
				if !toolFound {
					t.Fatalf("Tool Tree.SetValue not found in registry.")
				}
				_, setValueErr := setValueTool.Func(interp, MakeArgs(handle, targetNodeID, newValueToSet))
				if setValueErr != nil {
					t.Fatalf("SetValue tool call failed for node '%s': %v", targetNodeID, setValueErr)
				}

				nodeMap, errGet := callGetNode(t, interp, handle, targetNodeID)
				if errGet != nil {
					t.Fatalf("CheckFunc: Failed to get node '%s' after SetValue: %v", targetNodeID, errGet)
				}
				if nodeMap["value"] != newValueToSet {
					t.Errorf("SetValue did not update node '%s'. Got value: %v, want '%s'", targetNodeID, nodeMap["value"], newValueToSet)
				} else {
					t.Logf("SetValue successfully updated node '%s' to '%s'", targetNodeID, newValueToSet)
				}
			},
		},
		{
			name:      "SetValue On Object Node",
			toolName:  "Tree.SetValue",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1", "should_fail"),
			wantErr:   ErrCannotSetValueOnType,
		},
		{
			name:      "SetValue NonExistent Node",
			toolName:  "Tree.SetValue",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-999", "val"),
			wantErr:   ErrNotFound,
		},
		{
			name:      "AddChildNode To Root Object",
			toolName:  "Tree.AddChildNode",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1", "newChild1", "string", "hello", "newKeyInRoot"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("AddChildNode failed: %v", err)
				}
				newID, ok := result.(string)
				if !ok || newID == "" {
					t.Fatalf("AddChildNode did not return new node ID string, got %T: %v", result, result)
				}
				handle := ctx.(string)
				parentNodeMap, _ := callGetNode(t, interp, handle, "node-1")
				attrs, _ := parentNodeMap["attributes"].(TreeAttrs)
				if attrs["newKeyInRoot"] != newID {
					t.Errorf("AddChildNode: newKeyInRoot not pointing to new child ID %s. Attrs: %v", newID, attrs)
				}
				childNodeMap, _ := callGetNode(t, interp, handle, newID)
				if childNodeMap["value"] != "hello" {
					t.Errorf("Added child has wrong value: %v", childNodeMap["value"])
				}
			},
		},
		{
			name:      "AddChildNode ID Exists",
			toolName:  "Tree.AddChildNode",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1", "node-2", "string", "fail", "anotherKey"),
			wantErr:   ErrNodeIDExists,
		},
		{
			name:      "AddChildNode To Leaf Node (Robust)",
			toolName:  "Tree.GetNode",
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"),
			setupFunc: setupInitialTree,
			checkFunc: func(t *testing.T, interp *Interpreter, initialToolResult interface{}, initialToolErr error, ctx interface{}) {
				handle := ctx.(string)
				leafNodeID, found := findNodeIDByRootAttribute(t, interp, handle, "node-1", "name")
				if !found {
					t.Fatalf("AddChildNode To Leaf Node check: Could not find leaf node (e.g., for 'name' attribute).")
				}

				addTool, _ := interp.ToolRegistry().GetTool("Tree.AddChildNode")
				_, addErr := addTool.Func(interp, MakeArgs(handle, leafNodeID, "newChild2", "string", "val", "key"))

				if !errors.Is(addErr, ErrNodeWrongType) {
					t.Errorf("Expected ErrNodeWrongType when adding child to leaf node '%s', got %v", leafNodeID, addErr)
				}
			},
		},
		{
			name:      "RemoveNode Leaf (Robust)",
			toolName:  "Tree.GetNode",
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"),
			setupFunc: setupInitialTree,
			checkFunc: func(t *testing.T, interp *Interpreter, initialGetNodeResult interface{}, initialGetNodeErr error, ctx interface{}) {
				handle := ctx.(string)

				if initialGetNodeErr != nil {
					t.Fatalf("Initial Tree.GetNode call by helper failed for RemoveNode Leaf test: %v", initialGetNodeErr)
				}

				nodeIDToRemove, found := findNodeIDByRootAttribute(t, interp, handle, "node-1", "value")
				if !found {
					rootNodeDataForDebug, _ := callGetNode(t, interp, handle, "node-1")
					t.Fatalf("RemoveNode Leaf checkFunc: Attribute 'value' (to find target node ID) not found on root 'node-1'. Root data: %v", rootNodeDataForDebug)
				}
				t.Logf("Dynamically determined nodeID to remove (node linked by root's 'value' attribute): %s", nodeIDToRemove)

				removeTool, toolFound := interp.ToolRegistry().GetTool("Tree.RemoveNode")
				if !toolFound {
					t.Fatalf("Tool Tree.RemoveNode not found in registry.")
				}
				_, removeErr := removeTool.Func(interp, MakeArgs(handle, nodeIDToRemove))
				if removeErr != nil {
					t.Fatalf("RemoveNode Leaf checkFunc: Tree.RemoveNode tool call failed for dynamically found ID '%s': %v", nodeIDToRemove, removeErr)
				}
				t.Logf("Tree.RemoveNode called successfully for node ID: %s", nodeIDToRemove)

				_, errGetRemoved := callGetNode(t, interp, handle, nodeIDToRemove)
				if !errors.Is(errGetRemoved, ErrNotFound) {
					var rtErr *RuntimeError
					if errors.As(errGetRemoved, &rtErr) && errors.Is(rtErr.Wrapped, ErrNotFound) {
						t.Logf("Successfully confirmed node '%s' not found (wrapped ErrNotFound).", nodeIDToRemove)
					} else {
						t.Errorf("Expected ErrNotFound after removing node '%s', but got: %v (type %T)", nodeIDToRemove, errGetRemoved, errGetRemoved)
					}
				} else {
					t.Logf("Successfully confirmed node '%s' not found.", nodeIDToRemove)
				}

				rootNodeDataAfterRemove, errGetRootAfter := callGetNode(t, interp, handle, "node-1")
				if errGetRootAfter != nil {
					t.Fatalf("Failed to get root node 'node-1' after removal: %v", errGetRootAfter)
				}
				attributesAfterRemove, ok := rootNodeDataAfterRemove["attributes"].(TreeAttrs)
				if !ok {
					if rootNodeDataAfterRemove["attributes"] == nil {
						t.Logf("Root node 'attributes' field is nil after removal, which is fine.")
					} else {
						t.Fatalf("Root node 'attributes' field after removal is not TreeAttrs, got %T. Value: %v", rootNodeDataAfterRemove["attributes"], rootNodeDataAfterRemove["attributes"])
					}
				}
				if _, exists := attributesAfterRemove["value"]; exists {
					t.Errorf("Attribute 'value' STILL EXISTS on root after removing child node '%s' (which was linked by it). Final attributes: %v", nodeIDToRemove, attributesAfterRemove)
				} else {
					t.Logf("Attribute 'value' successfully removed from root node's attributes.")
				}
			},
		},
		{
			name:      "RemoveNode Root",
			toolName:  "Tree.RemoveNode",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"),
			wantErr:   ErrCannotRemoveRoot,
		},
		{
			name:      "RemoveNode NonExistent",
			toolName:  "Tree.RemoveNode",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-999"),
			wantErr:   ErrNotFound,
		},
	}

	for _, tc := range testCases {
		currentInterp, _ := NewDefaultTestInterpreter(t)
		testTreeToolHelper(t, currentInterp, tc)
	}
}
