// NeuroScript Version: 0.3.0
// File version: 5
// Purpose: Corrected compiler errors and removed direct import of 'interpreter' package to break import cycle. Now uses tool.Runtime interface.
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
func getTreeFromHandle(rt tool.Runtime, handleID, toolName string) (*utils.GenericTree, error) {
	if handleID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s requires non-empty 'tree_handle'", toolName),
			lang.ErrValidationRequiredArgNil,
		)
	}

	// FIX: Use the runtime interface instead of casting to the interpreter implementation.
	obj, err := rt.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if err != nil {
		if errors.Is(err, lang.ErrHandleNotFound) {
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound,
				fmt.Sprintf("%s: tree handle '%s' not found", toolName, handleID),
				errors.Join(lang.ErrNotFound, err),
			)
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: error retrieving handle '%s' (type %s)", toolName, handleID, utils.GenericTreeHandleType),
			err,
		)
	}

	tree, ok := obj.(*utils.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
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
		return nil, nil, err
	}

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound,
			fmt.Sprintf("%s: node ID '%s' not found in tree handle '%s'", toolName, nodeID, handleID),
			lang.ErrNotFound,
		)
	}

	return tree, node, nil
}

// removeChildFromParent removes a child ID from its parent's ChildIDs or Attributes.
// Returns true if the child was found and removed, false otherwise.
func removeChildFromParent(parent *utils.GenericTreeNode, childID string) bool {
	if parent == nil || childID == "" {
		return false
	}

	removed := false

	if parent.ChildIDs != nil {
		newChildIDs := make([]string, 0, len(parent.ChildIDs))
		for _, id := range parent.ChildIDs {
			if id == childID {
				removed = true
			} else {
				newChildIDs = append(newChildIDs, id)
			}
		}
		if removed {
			parent.ChildIDs = newChildIDs
			return true
		}
	}

	if parent.Type == "object" && parent.Attributes != nil {
		for key, valIDUntyped := range parent.Attributes {
			if valIDStr, ok := valIDUntyped.(string); ok && valIDStr == childID {
				delete(parent.Attributes, key)
				fmt.Printf("[DEBUG removeChildFromParent] Deleted key '%s'. Parent '%s' attributes now: %v\n", key, parent.ID, parent.Attributes)
				return true
			}
		}
	}

	return false
}

// removeNodeRecursive removes a node and all its descendants from the tree's NodeMap.
func removeNodeRecursive(tree *utils.GenericTree, nodeID string, visited map[string]struct{}) {
	if _, alreadyVisited := visited[nodeID]; alreadyVisited {
		return
	}
	visited[nodeID] = struct{}{}

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return
	}

	descendantIDs := make([]string, 0)
	if node.ChildIDs != nil {
		descendantIDs = append(descendantIDs, node.ChildIDs...)
	}
	if node.Attributes != nil {
		for _, childNodeIDUntyped := range node.Attributes {
			if childNodeIDStr, ok := childNodeIDUntyped.(string); ok {
				descendantIDs = append(descendantIDs, childNodeIDStr)
			}
		}
	}

	for _, descendantID := range descendantIDs {
		removeNodeRecursive(tree, descendantID, visited)
	}

	delete(tree.NodeMap, nodeID)
}
