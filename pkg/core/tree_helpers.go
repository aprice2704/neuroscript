// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 17:37:26 PDT // Add TreeRemoveNode helpers
// filename: pkg/core/tree_helpers.go

package core

import (
	"fmt"
)

// getTreeFromHandle retrieves the GenericTree from the interpreter's handle registry.
func getTreeFromHandle(interpreter *Interpreter, handleID, toolName string) (*GenericTree, error) {
	if handleID == "" {
		return nil, fmt.Errorf("%w: %s requires non-empty 'tree_handle'", ErrValidationRequiredArgNil, toolName)
	}

	obj, err := interpreter.GetHandleValue(handleID, GenericTreeHandleType)
	if err != nil {
		// Wrap error for context, including the specific handle type expected
		return nil, fmt.Errorf("%s failed getting handle '%s' (type %s): %w", toolName, handleID, GenericTreeHandleType, err)
	}

	tree, ok := obj.(*GenericTree)
	// Check all conditions: type assertion, not nil pointer, and internal map initialized
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle '%s' contains unexpected or uninitialized data type (%T), expected %s", ErrHandleInvalid, toolName, handleID, obj, GenericTreeHandleType)
	}
	return tree, nil
}

// getNodeFromHandle retrieves the GenericTree and the specific GenericTreeNode.
func getNodeFromHandle(interpreter *Interpreter, handleID, nodeID, toolName string) (*GenericTree, *GenericTreeNode, error) {
	if nodeID == "" {
		return nil, nil, fmt.Errorf("%w: %s requires non-empty 'node_id'", ErrValidationRequiredArgNil, toolName)
	}

	// First, get the tree using the helper
	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, nil, err // Error already has context from getTreeFromHandle
	}

	// Then, find the specific node within the retrieved tree
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, nil, fmt.Errorf("%w: %s node ID '%s' not found in tree handle '%s'", ErrNotFound, toolName, nodeID, handleID)
	}

	return tree, node, nil
}

// removeChildFromParent removes a child ID from its parent's ChildIDs or Attributes.
// Returns true if the child was found and removed, false otherwise.
func removeChildFromParent(parent *GenericTreeNode, childID string) bool {
	if parent == nil || childID == "" {
		return false // Invalid input
	}

	removed := false

	// Attempt removal from ChildIDs (typically for arrays, but check anyway)
	if parent.ChildIDs != nil {
		newChildIDs := make([]string, 0, len(parent.ChildIDs))
		for _, id := range parent.ChildIDs {
			if id == childID {
				removed = true // Mark as removed, continue to build list without it
			} else {
				newChildIDs = append(newChildIDs, id)
			}
		}
		if removed {
			parent.ChildIDs = newChildIDs
			return true // Found and removed from ChildIDs
		}
	}

	// Attempt removal from Attributes (for objects)
	if parent.Type == "object" && parent.Attributes != nil {
		for key, valID := range parent.Attributes {
			if valID == childID {
				delete(parent.Attributes, key)
				return true // Found and removed from Attributes
			}
		}
	}

	return false // Not found in either structure
}

// removeNodeRecursive removes a node and all its descendants from the tree's NodeMap.
// It uses a set to track visited nodes during the current removal operation to handle potential cycles (though unlikely for JSON).
func removeNodeRecursive(tree *GenericTree, nodeID string, visited map[string]struct{}) {
	if _, alreadyVisited := visited[nodeID]; alreadyVisited {
		return // Avoid infinite loops in cyclic graphs
	}
	visited[nodeID] = struct{}{} // Mark as visited for this removal operation

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return // Node already removed or never existed
	}

	// Collect all descendant node IDs first
	descendantIDs := make([]string, 0)
	if node.ChildIDs != nil {
		descendantIDs = append(descendantIDs, node.ChildIDs...)
	}
	if node.Attributes != nil {
		for _, childNodeID := range node.Attributes {
			descendantIDs = append(descendantIDs, childNodeID)
		}
	}

	// Recursively remove all descendants
	for _, descendantID := range descendantIDs {
		// Pass the same visited set down to detect cycles within this specific removal call chain
		removeNodeRecursive(tree, descendantID, visited)
	}

	// After all descendants are processed and removed, remove the current node itself
	delete(tree.NodeMap, nodeID)
}
