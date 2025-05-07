// NeuroScript Version: 0.3.0
// File version: 0.1.3 // Modified getTreeFromHandle to wrap ErrNotFound when ErrHandleNotFound occurs.
// filename: pkg/core/tree_helpers.go

package core

import (
	"errors" // Added for errors.Is and errors.Join
	"fmt"
)

// getTreeFromHandle retrieves the GenericTree from the interpreter's handle registry.
// If the handle is not found, it returns an error wrapping ErrNotFound.
func getTreeFromHandle(interpreter *Interpreter, handleID, toolName string) (*GenericTree, error) {
	if handleID == "" {
		// It's better to return an error that can be checked with errors.Is if it's a common case.
		// ErrValidationRequiredArgNil is a sentinel error.
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s requires non-empty 'tree_handle'", toolName),
			ErrValidationRequiredArgNil,
		)
	}

	obj, err := interpreter.GetHandleValue(handleID, GenericTreeHandleType)
	if err != nil {
		// Check if the error from GetHandleValue is because the handle was not found.
		if errors.Is(err, ErrHandleNotFound) {
			// If so, wrap ErrNotFound for the test and also include the original error details.
			// This makes errors.Is(returnedError, ErrNotFound) true.
			return nil, NewRuntimeError(ErrorCodeKeyNotFound, // Or a more specific tree error code
				fmt.Sprintf("%s: tree handle '%s' not found", toolName, handleID),
				errors.Join(ErrNotFound, err), // Ensure ErrNotFound is in the chain
			)
		}
		// For other errors from GetHandleValue (e.g., wrong type if GetHandleValue checked that, or other internal errors)
		return nil, NewRuntimeError(ErrorCodeInternal, // Or a more specific tree error code
			fmt.Sprintf("%s: error retrieving handle '%s' (type %s)", toolName, handleID, GenericTreeHandleType),
			err, // Wrap the original error
		)
	}

	tree, ok := obj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		// This indicates the handle existed but contained unexpected data.
		return nil, NewRuntimeError(ErrorCodeInternal, // Or a more specific tree error code for invalid structure
			fmt.Sprintf("%s: handle '%s' contains unexpected or uninitialized data type (%T), expected %s", toolName, handleID, obj, GenericTreeHandleType),
			ErrHandleInvalid,
		)
	}
	return tree, nil
}

// getNodeFromHandle retrieves the GenericTree and the specific GenericTreeNode.
// If the node is not found within a valid tree, it returns an error wrapping ErrNotFound.
func getNodeFromHandle(interpreter *Interpreter, handleID, nodeID, toolName string) (*GenericTree, *GenericTreeNode, error) {
	if nodeID == "" {
		return nil, nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s requires non-empty 'node_id'", toolName),
			ErrValidationRequiredArgNil,
		)
	}

	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, nil, err // Error already has context and correct wrapping from getTreeFromHandle
	}

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		// Node not found within a valid tree. Wrap ErrNotFound.
		return nil, nil, NewRuntimeError(ErrorCodeKeyNotFound, // Or a more specific tree node error code
			fmt.Sprintf("%s: node ID '%s' not found in tree handle '%s'", toolName, nodeID, handleID),
			ErrNotFound, // Ensure ErrNotFound is the sentinel error
		)
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
