// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Made RemoveNode Leaf and SetValue Valid Leaf tests robust to non-deterministic node IDs
// Tests for tree modification tools (SetValue, AddChildNode, RemoveNode).
// nlines: ~130
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_modify_test.go

package core

import (
	"errors" // Added for robust SetValue test logging
	"testing"
	// Assuming NewDefaultTestInterpreter is in this package or a test helper package.
	// If it's in 'core' itself, no separate import needed.
	// If it's in a 'testhelpers' sub-package, it would be:
	// "github.com/aprice2704/neuroscript/pkg/core/testhelpers"
)

// Helper function (can be local to TestTreeModificationTools or moved to tools_tree_test_helpers.go if used more widely)
// findNodeIDByRootAttribute finds the ID of a child node linked by a specific attribute key on a given root/parent node.
func findNodeIDByRootAttribute(t *testing.T, interp *Interpreter, handle string, parentNodeID string, attributeKey string) (string, bool) {
	t.Helper()
	parentNodeData, err := callGetNode(t, interp, handle, parentNodeID) // callGetNode is from tools_tree_test_helpers.go
	if err != nil {
		t.Logf("findNodeIDByRootAttribute: Failed to get parent node '%s': %v", parentNodeID, err)
		return "", false
	}

	// toolTreeGetNode returns "attributes" as map[string]string
	attributes, ok := parentNodeData["attributes"].(map[string]string)
	if !ok {
		// This case might happen if attributes is nil or a different type than expected.
		// toolTreeGetNode might return nil for attributes if node.Attributes is empty or nil.
		// The callGetNode helper should ideally return an error or an empty map in such cases
		// rather than letting the type assertion fail here if `attributes` key is missing.
		// For now, assume if attributes field exists, it's map[string]string, or it's absent (nil).
		if parentNodeData["attributes"] == nil {
			t.Logf("findNodeIDByRootAttribute: Parent node '%s' has no 'attributes' field or it's nil.", parentNodeID)
		} else {
			t.Logf("findNodeIDByRootAttribute: Parent node '%s' attributes field is not map[string]string, got %T", parentNodeID, parentNodeData["attributes"])
		}
		return "", false
	}

	nodeID, found := attributes[attributeKey]
	if !found {
		t.Logf("findNodeIDByRootAttribute: Attribute key '%s' not found on parent node '%s'. Attributes: %v", attributeKey, parentNodeID, attributes)
		return "", false
	}
	return nodeID, true
}

func TestTreeModificationTools(t *testing.T) {
	jsonSimple := `{"name": "item", "value": 10}` // "name" is string, "value" is number

	// setupInitialTree loads the simple JSON and returns the tree handle.
	setupInitialTree := func(t *testing.T, interp *Interpreter) interface{} {
		return setupTreeWithJSON(t, interp, jsonSimple) // setupTreeWithJSON is from tools_tree_test_helpers.go
	}

	testCases := []treeTestCase{
		{
			name:      "SetValue Valid Leaf (string node) (Robust)",
			toolName:  "Tree.GetNode",                           // Initial benign call by testTreeToolHelper
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"), // Get root, not strictly used by this checkFunc's core logic
			setupFunc: setupInitialTree,
			checkFunc: func(t *testing.T, interp *Interpreter, initialToolResult interface{}, initialToolErr error, ctx interface{}) {
				handle := ctx.(string)
				newValueToSet := "new_name_value"

				// 1. Dynamically find the node ID associated with the "name" attribute of the root.
				//    This is the node we semantically want to modify (the one originally "item").
				targetNodeID, found := findNodeIDByRootAttribute(t, interp, handle, "node-1", "name")
				if !found {
					rootNodeDataForDebug, _ := callGetNode(t, interp, handle, "node-1") // For debugging
					t.Fatalf("SetValue check: Could not find node linked by 'name' attribute of root. Root data: %v", rootNodeDataForDebug)
				}
				t.Logf("SetValue check: Dynamically found target node ID (for original 'name') as: %s", targetNodeID)

				// 2. Call Tree.SetValue on this dynamically found node ID.
				setValueTool, toolFound := interp.ToolRegistry().GetTool("Tree.SetValue")
				if !toolFound {
					t.Fatalf("Tool Tree.SetValue not found in registry.")
				}
				_, setValueErr := setValueTool.Func(interp, MakeArgs(handle, targetNodeID, newValueToSet))
				if setValueErr != nil {
					t.Fatalf("SetValue tool call failed for node '%s': %v", targetNodeID, setValueErr)
				}

				// 3. Verify the node was updated.
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
			name:          "SetValue On Object Node",
			toolName:      "Tree.SetValue",
			setupFunc:     setupInitialTree,
			args:          MakeArgs("SETUP_HANDLE:tree1", "node-1", "should_fail"), // "node-1" is always root ID
			wantToolErrIs: ErrCannotSetValueOnType,
		},
		{
			name:          "SetValue NonExistent Node",
			toolName:      "Tree.SetValue",
			setupFunc:     setupInitialTree,
			args:          MakeArgs("SETUP_HANDLE:tree1", "node-999", "val"),
			wantToolErrIs: ErrNotFound,
		},

		// Tree.AddChildNode - This test might also need robustness if it relies on specific child IDs for verification.
		// For now, keeping it as is, assuming its checks are based on what AddChildNode returns or broader structure.
		{
			name:      "AddChildNode To Root Object",
			toolName:  "Tree.AddChildNode",
			setupFunc: setupInitialTree,
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1", "newChild1", "string", "hello", "newKeyInRoot"),
			checkFunc: func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("AddChildNode failed: %v", err)
				}
				newID, ok := result.(string) // AddChildNode returns the ID of the new node
				if !ok || newID == "" {
					t.Fatalf("AddChildNode did not return new node ID string, got %T: %v", result, result)
				}
				handle := ctx.(string)
				parentNodeMap, _ := callGetNode(t, interp, handle, "node-1")
				attrs := parentNodeMap["attributes"].(map[string]string)
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
			name:          "AddChildNode ID Exists", // Assumes "node-2" will be created by setupInitialTree and an attempt to reuse it.
			toolName:      "Tree.AddChildNode",      // This test's robustness depends on "node-2" being predictably one of the first few IDs.
			setupFunc:     setupInitialTree,
			args:          MakeArgs("SETUP_HANDLE:tree1", "node-1", "node-2", "string", "fail", "anotherKey"),
			wantToolErrIs: ErrNodeIDExists,
		},
		{
			name:      "AddChildNode To Leaf Node (Robust)",
			toolName:  "Tree.GetNode",                           // Benign call by helper
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"), // Get root
			setupFunc: setupInitialTree,
			checkFunc: func(t *testing.T, interp *Interpreter, initialToolResult interface{}, initialToolErr error, ctx interface{}) {
				handle := ctx.(string)
				// Find the ID of a leaf node (e.g., the one for "item" via "name" attribute)
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

		// Tree.RemoveNode
		{
			name:      "RemoveNode Leaf (Robust)",
			toolName:  "Tree.GetNode",                           // Benign operation for testTreeToolHelper's initial call
			args:      MakeArgs("SETUP_HANDLE:tree1", "node-1"), // Get root node
			setupFunc: setupInitialTree,
			checkFunc: func(t *testing.T, interp *Interpreter, initialGetNodeResult interface{}, initialGetNodeErr error, ctx interface{}) {
				handle := ctx.(string)

				if initialGetNodeErr != nil {
					t.Fatalf("Initial Tree.GetNode call by helper failed for RemoveNode Leaf test: %v", initialGetNodeErr)
				}

				// 1. Find the actual ID of the node linked by attribute "value" on the root.
				//    This is the node we semantically want to remove (the one representing '10').
				nodeIDToRemove, found := findNodeIDByRootAttribute(t, interp, handle, "node-1", "value")
				if !found {
					rootNodeDataForDebug, _ := callGetNode(t, interp, handle, "node-1")
					t.Fatalf("RemoveNode Leaf checkFunc: Attribute 'value' (to find target node ID) not found on root 'node-1'. Root data: %v", rootNodeDataForDebug)
				}
				t.Logf("Dynamically determined nodeID to remove (node linked by root's 'value' attribute): %s", nodeIDToRemove)

				// 2. Now, call the Tree.RemoveNode tool with this dynamically found ID.
				removeTool, toolFound := interp.ToolRegistry().GetTool("Tree.RemoveNode")
				if !toolFound {
					t.Fatalf("Tool Tree.RemoveNode not found in registry.")
				}
				_, removeErr := removeTool.Func(interp, MakeArgs(handle, nodeIDToRemove))
				if removeErr != nil {
					t.Fatalf("RemoveNode Leaf checkFunc: Tree.RemoveNode tool call failed for dynamically found ID '%s': %v", nodeIDToRemove, removeErr)
				}
				t.Logf("Tree.RemoveNode called successfully for node ID: %s", nodeIDToRemove)

				// 3. Check the removed node is actually gone from the tree.
				_, errGetRemoved := callGetNode(t, interp, handle, nodeIDToRemove)
				if !errors.Is(errGetRemoved, ErrNotFound) {
					// Check if it's a RuntimeError wrapping ErrNotFound
					var rtErr *RuntimeError
					if errors.As(errGetRemoved, &rtErr) && errors.Is(rtErr.Wrapped, ErrNotFound) {
						// This is also an acceptable way to get ErrNotFound
						t.Logf("Successfully confirmed node '%s' not found (wrapped ErrNotFound).", nodeIDToRemove)
					} else {
						t.Errorf("Expected ErrNotFound after removing node '%s', but got: %v (type %T)", nodeIDToRemove, errGetRemoved, errGetRemoved)
					}
				} else {
					t.Logf("Successfully confirmed node '%s' not found.", nodeIDToRemove)
				}

				// 4. Check the attribute "value" is now gone from the root node.
				rootNodeDataAfterRemove, errGetRootAfter := callGetNode(t, interp, handle, "node-1")
				if errGetRootAfter != nil {
					t.Fatalf("Failed to get root node 'node-1' after removal: %v", errGetRootAfter)
				}
				attributesAfterRemove, ok := rootNodeDataAfterRemove["attributes"].(map[string]string)
				if !ok {
					if rootNodeDataAfterRemove["attributes"] == nil {
						t.Logf("Root node 'attributes' field is nil after removal, which is fine.")
						// The key "value" won't exist in a nil map.
					} else {
						t.Fatalf("Root node 'attributes' field after removal is not map[string]string, got %T. Value: %v", rootNodeDataAfterRemove["attributes"], rootNodeDataAfterRemove["attributes"])
					}
				}
				// This check is fine even if attributesAfterRemove is nil (then 'exists' will be false)
				if _, exists := attributesAfterRemove["value"]; exists {
					t.Errorf("Attribute 'value' STILL EXISTS on root after removing child node '%s' (which was linked by it). Final attributes: %v", nodeIDToRemove, attributesAfterRemove)
				} else {
					t.Logf("Attribute 'value' successfully removed from root node's attributes.")
				}
			},
		},
		{
			name:          "RemoveNode Root",
			toolName:      "Tree.RemoveNode",
			setupFunc:     setupInitialTree,
			args:          MakeArgs("SETUP_HANDLE:tree1", "node-1"), // "node-1" is always the root from Tree.LoadJSON with current ID gen
			wantToolErrIs: ErrCannotRemoveRoot,
		},
		{
			name:          "RemoveNode NonExistent",
			toolName:      "Tree.RemoveNode",
			setupFunc:     setupInitialTree,
			args:          MakeArgs("SETUP_HANDLE:tree1", "node-999"),
			wantToolErrIs: ErrNotFound,
		},
	}

	for _, tc := range testCases {
		// Using NewDefaultTestInterpreter for a fresh, isolated environment for each test case.
		// This helper should handle logger, NoOpLLM, sandbox, and registering core tools (including tree tools).
		currentInterp, err := NewDefaultTestInterpreter(t)
		if err != "" {
			t.Fatalf("Failed to create default test interpreter for test case '%s': %v", tc.name, err)
		}
		// Ensure tree tools are registered if NewDefaultTestInterpreter doesn't do it.
		// (Typically, a default test interpreter would register all core tools).
		// For explicitness, or if NewDefaultTestInterpreter is minimal:
		// if err := RegisterTreeTools(currentInterp.ToolRegistry()); err != nil {
		//    t.Fatalf("Failed to register tree tools for test case '%s': %v", tc.name, err)
		// }

		testTreeToolHelper(t, currentInterp, tc) // testTreeToolHelper is from tools_tree_test_helpers.go
	}
}
