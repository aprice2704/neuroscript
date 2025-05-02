// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 14:58:13 PDT // Split from tools_tree.go
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
