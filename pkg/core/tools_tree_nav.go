// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Include parent_attribute_key in GetNode result.
// nlines: 82 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_tree_nav.go

package core

import "fmt"

// toolTreeGetNode implements the Tree.GetNode tool.
func toolTreeGetNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetNode"
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err // Propagates node not found etc.
	}

	nodeMap := map[string]interface{}{
		"id":                   node.ID,
		"type":                 node.Type,
		"value":                node.Value,              // Will be nil if not applicable
		"attributes":           node.Attributes,         // Will be nil if not applicable
		"child_ids":            node.ChildIDs,           // Will be nil if not applicable
		"parent_id":            node.ParentID,           // Will be empty string if root
		"parent_attribute_key": node.ParentAttributeKey, // Added this field
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved node information", toolName),
		"handle", treeHandle, "nodeId", nodeID)

	return nodeMap, nil
}

// toolTreeGetChildren implements the Tree.GetChildren tool.
func toolTreeGetChildren(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetChildren"
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err // Propagates node not found etc.
	}

	if node.Type != "array" {
		return nil, NewRuntimeError(ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: cannot get children of node type '%s' (expected 'array')", toolName, node.Type),
			ErrNodeWrongType)
	}

	childIDsInterface := make([]interface{}, len(node.ChildIDs))
	for i, id := range node.ChildIDs {
		childIDsInterface[i] = id
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved children IDs", toolName),
		"handle", treeHandle, "nodeId", nodeID, "count", len(childIDsInterface))

	return childIDsInterface, nil
}

// toolTreeGetParent implements the Tree.GetParent tool.
func toolTreeGetParent(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetParent"
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved parent ID", toolName),
		"handle", treeHandle, "nodeId", nodeID, "parentId", node.ParentID)

	if node.ParentID == "" {
		return nil, nil
	}

	return node.ParentID, nil
}
