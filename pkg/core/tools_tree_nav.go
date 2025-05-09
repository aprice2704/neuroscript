// NeuroScript Version: 0.3.1
// File version: 0.1.5
// CRITICAL FIX: Changed "child_ids" key to "children" in GetNode result to match spec.
// nlines: 82 // Approximate
// risk_rating: MEDIUM // Critical for correct tool behavior
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
		return nil, err
	}

	var childrenSlice []string
	if node.ChildIDs != nil {
		childrenSlice = make([]string, len(node.ChildIDs))
		copy(childrenSlice, node.ChildIDs)
	} else {
		childrenSlice = []string{}
	}

	nodeMap := map[string]interface{}{
		"id":                   node.ID,
		"type":                 node.Type,
		"value":                node.Value,
		"attributes":           node.Attributes,
		"children":             childrenSlice, // CORRECTED KEY
		"parent_id":            node.ParentID,
		"parent_attribute_key": node.ParentAttributeKey,
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
		return nil, err
	}

	if node.Type != "array" { // This tool is specific to "array" type nodes
		return nil, NewRuntimeError(ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: cannot get children of node type '%s' (expected 'array')", toolName, node.Type),
			ErrNodeWrongType)
	}

	var childIDsToReturn []interface{}
	if node.ChildIDs != nil {
		childIDsToReturn = make([]interface{}, len(node.ChildIDs))
		for i, id := range node.ChildIDs {
			childIDsToReturn[i] = id
		}
	} else {
		childIDsToReturn = []interface{}{}
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved children IDs", toolName),
		"handle", treeHandle, "nodeId", nodeID, "count", len(childIDsToReturn))

	return childIDsToReturn, nil
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
