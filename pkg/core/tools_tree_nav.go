// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Removed local ToolImplementations, standardized error handling.
// nlines: 100 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_tree_nav.go

package core

import (
	"fmt"
	// "errors" - Not directly needed if helpers handle error wrapping appropriately
)

// toolTreeGetNode returns information about a specific node.
// Corresponds to ToolSpec "Tree.GetNode".
func toolTreeGetNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetNode"

	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}

	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}
	nodeID, okNodeID := args[1].(string)
	if !okNodeID {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: node_id argument must be a string, got %T", toolName, args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle already returns a RuntimeError
	}

	// Prepare attributes map for return
	attributesMap := make(map[string]interface{}) // Always create, even if empty, for consistent return structure
	if node.Attributes != nil {
		for k, v := range node.Attributes {
			attributesMap[k] = v
		}
	}

	// Prepare children IDs slice for return
	childrenSlice := make([]interface{}, len(node.ChildIDs)) // Correctly handles nil or empty node.ChildIDs
	for i, id := range node.ChildIDs {
		childrenSlice[i] = id
	}

	nodeMap := map[string]interface{}{
		"id":         node.ID,
		"type":       node.Type,
		"value":      node.Value,    // Will be nil for object/array types if not explicitly set otherwise
		"attributes": attributesMap, // Contains metadata or object key->childID mappings
		"children":   childrenSlice, // Contains ordered child IDs for arrays, or general children
		"parentId":   node.ParentID, // Will be "" for root
	}
	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved node information", toolName), "handle", handleID, "nodeId", nodeID)
	return nodeMap, nil
}

// toolTreeGetChildren returns a list of child node IDs for a given node.
// If the node is not an object/array or has no children, it returns an empty list.
// Corresponds to ToolSpec "Tree.GetChildren".
func toolTreeGetChildren(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetChildren"

	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}
	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}
	nodeID, okNodeID := args[1].(string)
	if !okNodeID {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: node_id argument must be a string, got %T", toolName, args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle already returns a RuntimeError
	}

	// node.ChildIDs is []string. Convert to []interface{} for return.
	// If node.ChildIDs is nil or empty, this correctly creates and returns an empty []interface{}.
	childrenIDs := make([]interface{}, len(node.ChildIDs))
	for i, id := range node.ChildIDs {
		childrenIDs[i] = id
	}
	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved children IDs", toolName), "handle", handleID, "nodeId", nodeID, "count", len(childrenIDs))
	return childrenIDs, nil
}

// toolTreeGetParent returns the parent node ID (string).
// Returns an empty string if the node is the root or has no parent.
// Corresponds to ToolSpec "Tree.GetParent".
func toolTreeGetParent(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetParent"

	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}
	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}
	nodeID, okNodeID := args[1].(string)
	if !okNodeID {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: node_id argument must be a string, got %T", toolName, args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle already returns a RuntimeError
	}
	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved parent ID", toolName), "handle", handleID, "nodeId", nodeID, "parentId", node.ParentID)
	return node.ParentID, nil // ParentID is already a string ("" for root)
}
