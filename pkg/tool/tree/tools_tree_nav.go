// NeuroScript Version: 0.3.1
// File version: 0.1.6
// CRITICAL FIX: Changed "children" in GetNode result to be []interface{} to align with NeuroScript list type conventions.
// nlines: 82 // Approximate
// risk_rating: MEDIUM // Critical for correct tool behavior
// filename: pkg/tool/tree/tools_tree_nav.go

package tree

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolTreeGetNode implements the Tree.GetNode tool.
func toolTreeGetNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetNode"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	// NeuroScript lists are []interface{}, so we must return that type.
	var childrenSlice []interface{}
	if node.ChildIDs != nil {
		childrenSlice = make([]interface{}, len(node.ChildIDs))
		for i, id := range node.ChildIDs {
			childrenSlice[i] = id
		}
	} else {
		childrenSlice = []interface{}{}
	}

	nodeMap := map[string]interface{}{
		"id":			node.ID,
		"type":			node.Type,
		"value":		node.Value,
		"attributes":		node.Attributes,
		"children":		childrenSlice,	// CORRECTED TYPE
		"parent_id":		node.ParentID,
		"parent_attribute_key":	node.ParentAttributeKey,
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Retrieved node information", toolName),
		"handle", treeHandle, "nodeId", nodeID)

	return nodeMap, nil
}

// toolTreeGetChildren implements the Tree.GetChildren tool.
func toolTreeGetChildren(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetChildren"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	if node.Type != "array" {	// This tool is specific to "array" type nodes
		return nil, lang.NewRuntimeError(ErrorCodeNodeWrongType,
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
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), ErrInvalidArgument)
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