// NeuroScript Version: 0.6.5
// File version: 3
// Purpose: Corrected GetNode to return a standard map[string]interface{} for attributes to prevent test panics.
// filename: pkg/tool/tree/tools_tree_nav.go
// nlines: 200
// risk_rating: MEDIUM

package tree

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// toolTreeGetNode implements the Tree.GetNode tool.
func toolTreeGetNode(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetNode"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), lang.ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	var childrenSlice []interface{}
	if node.ChildIDs != nil {
		childrenSlice = make([]interface{}, len(node.ChildIDs))
		for i, id := range node.ChildIDs {
			childrenSlice[i] = id
		}
	} else {
		childrenSlice = []interface{}{}
	}

	// Convert utils.TreeAttrs to map[string]interface{} to avoid panics in tests.
	attributesMap := make(map[string]interface{})
	if node.Attributes != nil {
		for k, v := range node.Attributes {
			attributesMap[k] = v
		}
	}

	nodeMap := map[string]interface{}{
		"id":                   node.ID,
		"type":                 node.Type,
		"value":                node.Value,
		"attributes":           attributesMap, // CORRECTED TYPE
		"children":             childrenSlice,
		"parent_id":            node.ParentID,
		"parent_attribute_key": node.ParentAttributeKey,
	}

	interpreter.GetLogger().Debug(fmt.Sprintf("%s: Retrieved node information", toolName),
		"handle", treeHandle, "nodeId", nodeID)

	return nodeMap, nil
}

// toolTreeGetChildren implements the Tree.GetChildren tool.
func toolTreeGetChildren(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetChildren"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), lang.ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	if node.Type != "array" { // This tool is specific to "array" type nodes
		return nil, lang.NewRuntimeError(lang.ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: cannot get children of node type '%s' (expected 'array')", toolName, node.Type),
			lang.ErrNodeWrongType)
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

	interpreter.GetLogger().Debug(fmt.Sprintf("%s: Retrieved children IDs", toolName),
		"handle", treeHandle, "nodeId", nodeID, "count", len(childIDsToReturn))

	return childIDsToReturn, nil
}

// toolTreeGetParent implements the Tree.GetParent tool.
func toolTreeGetParent(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetParent"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	nodeID, okNodeID := args[1].(string)
	if !okHandle || !okNodeID || treeHandle == "" || nodeID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid arguments (handle=%T, nodeID=%T)", toolName, args[0], args[1]), lang.ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, treeHandle, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	if node.ParentID == "" {
		return nil, nil
	}

	return toolTreeGetNode(interpreter, []interface{}{treeHandle, node.ParentID})
}

// toolTreeGetRoot retrieves the root node of the tree.
func toolTreeGetRoot(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetRoot"
	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 1 argument (tree_handle), got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	if !okHandle {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), lang.ErrInvalidArgument)
	}

	tree, err := getTreeFromHandle(interpreter, treeHandle, toolName)
	if err != nil {
		return nil, err
	}

	if tree.RootID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "tree has no root ID", lang.ErrInternal)
	}

	return toolTreeGetNode(interpreter, []interface{}{treeHandle, tree.RootID})
}

// toolTreeGetNodeByPath retrieves a node by a path expression (e.g., "key.0.name").
func toolTreeGetNodeByPath(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetNodeByPath"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, path), got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	handleID, okHandle := args[0].(string)
	path, okPath := args[1].(string)
	if !okHandle || !okPath {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: invalid argument types", toolName), lang.ErrInvalidArgument)
	}

	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, err
	}

	currentNode, exists := tree.NodeMap[tree.RootID]
	if !exists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "root node not found in tree", lang.ErrInternal)
	}

	if path == "" {
		return toolTreeGetNode(interpreter, []interface{}{handleID, tree.RootID})
	}

	segments := strings.Split(path, ".")
	for _, segment := range segments {
		if currentNode == nil {
			return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("%s: cannot traverse path, intermediate node is nil", toolName), lang.ErrNotFound)
		}

		switch currentNode.Type {
		case "object":
			childNodeIDUntyped, ok := currentNode.Attributes[segment]
			if !ok {
				return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("%s: key '%s' not found in object node '%s'", toolName, segment, currentNode.ID), lang.ErrNotFound)
			}
			childNodeID, ok := childNodeIDUntyped.(string)
			if !ok {
				return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("%s: attribute value for key '%s' is not a node ID string", toolName, segment), lang.ErrInternal)
			}
			currentNode, exists = tree.NodeMap[childNodeID]
			if !exists {
				return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("%s: child node ID '%s' not found in tree map", toolName, childNodeID), lang.ErrInternal)
			}
		case "array":
			index, err := strconv.Atoi(segment)
			if err != nil {
				return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid array index '%s' in path", toolName, segment), lang.ErrInvalidArgument)
			}
			if index < 0 || index >= len(currentNode.ChildIDs) {
				return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound, fmt.Sprintf("%s: index %d out of bounds for array node '%s'", toolName, index, currentNode.ID), lang.ErrNotFound)
			}
			childNodeID := currentNode.ChildIDs[index]
			currentNode, exists = tree.NodeMap[childNodeID]
			if !exists {
				return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("%s: child node ID '%s' not found in tree map", toolName, childNodeID), lang.ErrInternal)
			}
		default:
			return nil, lang.NewRuntimeError(lang.ErrorCodeNodeWrongType, fmt.Sprintf("%s: cannot traverse path, node '%s' of type '%s' is not an object or array", toolName, currentNode.ID, currentNode.Type), lang.ErrNodeWrongType)
		}
	}

	return toolTreeGetNode(interpreter, []interface{}{handleID, currentNode.ID})
}
