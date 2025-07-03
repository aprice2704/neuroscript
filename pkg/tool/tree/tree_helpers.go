// NeuroScript Version: 0.3.0
// File version: 4
// Purpose: Corrected compiler errors by adding type assertions when reading from the Attributes map in `removeChildFromParent` and `removeNodeRecursive`.
// filename: pkg/tool/tree/tree_helpers.go

package tree

import (
	"errors" // Added for errors.Is and errors.Join
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// getTreeFromHandle retrieves the GenericTree from the interpreter's handle registry.
// If the handle is not found, it returns an error wrapping ErrNotFound.
func getTreeFromHandle(interpreter tool.Runtime, handleID, toolName string) (*utils.GenericTree, error) {
	if handleID == "" {
		// It's better to return an error that can be checked with errors.Is if it's a common case.
		// ErrValidationRequiredArgNil is a sentinel error.
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s requires non-empty 'tree_handle'", toolName),
			lang.ErrValidationRequiredArgNil,
		)
	}

	obj, err := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if err != nil {
		// Check if the error from GetHandleValue is because the handle was not found.
		if errors.Is(err, lang.ErrHandleNotFound) {
			// If so, wrap ErrNotFound for the test and also include the original error details.
			// This makes errors.Is(returnedError, ErrNotFound) true.
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, // Or a more specific tree error code
				fmt.Sprintf("%s: tree handle '%s' not found", toolName, handleID),
				errors.Join(lang.ErrNotFound, err), // Ensure lang.ErrNotFound is in the chain
			)
		}
		// For other errors from GetHandleValue (e.g., wrong type if GetHandleValue checked that, or other internal errors)
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, // Or a more specific tree error code
			fmt.Sprintf("%s: error retrieving handle '%s' (type %s)", toolName, handleID, utils.GenericTreeHandleType),
			err, // Wrap the original error
		)
	}

	tree, ok := obj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		// This indicates the handle existed but contained unexpected data.
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, // Or a more specific tree error code for invalid structure
			fmt.Sprintf("%s: handle '%s' contains unexpected or uninitialized data type (%T), expected %s", toolName, handleID, obj, utils.GenericTreeHandleType),
			lang.ErrHandleInvalid,
		)
	}
	return tree, nil
}

// getNodeFromHandle retrieves the GenericTree and the specific GenericTreeNode.
// If the node is not found within a valid tree, it returns an error wrapping ErrNotFound.
func getNodeFromHandle(interpreter tool.Runtime, handleID, nodeID, toolName string) (*utils.GenericTree, *utils.GenericTreeNode, error) {
	if nodeID == "" {
		return nil, nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s requires non-empty 'node_id'", toolName),
			lang.ErrValidationRequiredArgNil,
		)
	}

	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, nil, err // Error already has context and correct wrapping from getTreeFromHandle
	}

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		// Node not found within a valid tree. Wrap ErrNotFound.
		return nil, nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, // Or a more specific tree node error code
			fmt.Sprintf("%s: node ID '%s' not found in tree handle '%s'", toolName, nodeID, handleID),
			lang.ErrNotFound, // Ensure lang.ErrNotFound is the sentinel error
		)
	}

	return tree, node, nil
}

// removeChildFromParent removes a child ID from its parent's ChildIDs or Attributes.
// Returns true if the child was found and removed, false otherwise.
func removeChildFromParent(parent *utils.GenericTreeNode, childID string) bool {
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
		for key, valIDUntyped := range parent.Attributes {
			// Safely check if the attribute value is the string ID we are looking for.
			if valIDStr, ok := valIDUntyped.(string); ok && valIDStr == childID {
				delete(parent.Attributes, key)
				fmt.Printf("[DEBUG removeChildFromParent] Deleted key '%s'. Parent '%s' attributes now: %v\n", key, parent.ID, parent.Attributes)
				return true // Found and removed from Attributes
			}
		}
	}

	return false // Not found in either structure
}

// removeNodeRecursive removes a node and all its descendants from the tree's NodeMap.
// It uses a set to track visited nodes during the current removal operation to handle potential cycles (though unlikely for JSON).
func removeNodeRecursive(tree *utils.GenericTree, nodeID string, visited map[string]struct{}) {
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
		for _, childNodeIDUntyped := range node.Attributes {
			// An attribute's value is only a descendant link if it's a string.
			if childNodeIDStr, ok := childNodeIDUntyped.(string); ok {
				descendantIDs = append(descendantIDs, childNodeIDStr)
			}
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
