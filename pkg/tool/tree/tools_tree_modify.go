// NeuroScript Version: 0.3.1
// File version: 2
// Purpose: Corrected initialization of node Attributes to use the new TreeAttrs type instead of map[string]string.
// nlines: 335 // Approximate
// risk_rating: LOW
// filename: pkg/tool/tree/tools_tree_modify.go

package tree

import (
	"errors"
	"fmt"
	"slices"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// --- Tree.SetValue (was toolTreeModifyNode) ---
// Sets the value of an existing leaf node.
// Corresponds to ToolSpec "Tree.SetValue".
func toolTreeModifyNode(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "Tree.SetValue"

	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 arguments (tree_handle, node_id, value), got %d", toolName, len(args)),
			lang.ErrArgumentMismatch,
		)
	}

	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), lang.ErrInvalidArgument)
	}

	nodeID, okNodeID := args[1].(string)
	if !okNodeID {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: node_id argument must be a string, got %T", toolName, args[1]), lang.ErrInvalidArgument)
	}

	newValue := args[2]

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	if node.Type == "object" || node.Type == "array" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeTreeConstraintViolation,
			fmt.Sprintf("%s: cannot set value directly on node '%s' of type '%s'; use object/array modification tools", toolName, nodeID, node.Type),
			lang.ErrCannotSetValueOnType,
		)
	}

	node.Value = newValue
	interpreter.Logger().Debug(fmt.Sprintf("%s: Modified node value", toolName), "handle", handleID, "nodeId", nodeID)

	return nil, nil
}

// --- Tree.SetObjectAttribute (was toolTreeSetAttribute) ---
// Sets or updates an attribute on an object node, mapping the attribute key to a child node ID.
// Corresponds to ToolSpec "Tree.SetObjectAttribute".
func toolTreeSetAttribute(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "Tree.SetObjectAttribute"

	if len(args) != 4 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 4 arguments (tree_handle, object_node_id, attribute_key, child_node_id), got %d", toolName, len(args)),
			lang.ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	objectNodeID, _ := args[1].(string)
	attrKey, _ := args[2].(string)
	childNodeID, _ := args[3].(string)

	if attrKey == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: attribute_key cannot be empty", toolName), lang.ErrInvalidArgument)
	}
	if childNodeID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: child_node_id cannot be empty", toolName), lang.ErrInvalidArgument)
	}

	tree, objectNode, err := getNodeFromHandle(interpreter, handleID, objectNodeID, toolName)
	if err != nil {
		return nil, err
	}

	if objectNode.Type != "object" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: target node '%s' is type '%s', expected 'object'", toolName, objectNodeID, objectNode.Type),
			lang.ErrTreeNodeNotObject,
		)
	}

	if _, childExists := tree.NodeMap[childNodeID]; !childExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeKeyNotFound,
			fmt.Sprintf("%s: specified child_node_id '%s' not found in tree '%s'", toolName, childNodeID, handleID),
			lang.ErrNotFound,
		)
	}

	if objectNode.Attributes == nil {
		objectNode.Attributes = make(utils.TreeAttrs)
		interpreter.Logger().Warn(fmt.Sprintf("%s: Node attributes map was nil for node '%s', initialized.", toolName, objectNodeID))
	}
	objectNode.Attributes[attrKey] = childNodeID

	interpreter.Logger().Debug(fmt.Sprintf("%s: Set object attribute", toolName), "handle", handleID, "objectNodeId", objectNodeID, "key", attrKey, "childId", childNodeID)
	return nil, nil
}

// --- Tree.RemoveObjectAttribute (was toolTreeRemoveAttribute) ---
// Removes an attribute from an object node.
// Corresponds to ToolSpec "Tree.RemoveObjectAttribute".
func toolTreeRemoveAttribute(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "Tree.RemoveObjectAttribute"

	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 arguments (tree_handle, object_node_id, attribute_key), got %d", toolName, len(args)),
			lang.ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	objectNodeID, _ := args[1].(string)
	attrKey, _ := args[2].(string)

	if attrKey == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: attribute_key cannot be empty", toolName), lang.ErrInvalidArgument)
	}

	_, objectNode, err := getNodeFromHandle(interpreter, handleID, objectNodeID, toolName)
	if err != nil {
		return nil, err
	}

	if objectNode.Type != "object" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: target node '%s' is type '%s', expected 'object'", toolName, objectNodeID, objectNode.Type),
			lang.ErrTreeNodeNotObject,
		)
	}

	if objectNode.Attributes == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: node '%s' has no attributes to remove from (key: %q)", toolName, objectNodeID, attrKey),
			lang.ErrAttributeNotFound,
		)
	}

	if _, keyExists := objectNode.Attributes[attrKey]; !keyExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: attribute_key '%s' not found on node '%s'", toolName, attrKey, objectNodeID),
			lang.ErrAttributeNotFound,
		)
	}

	delete(objectNode.Attributes, attrKey)
	interpreter.Logger().Debug(fmt.Sprintf("%s: Removed object attribute", toolName), "handle", handleID, "objectNodeId", objectNodeID, "key", attrKey)
	return nil, nil
}

// --- Tree.AddChildNode (was toolTreeAddNode) ---
// Adds a new child node to an existing parent node.
// Corresponds to ToolSpec "Tree.AddChildNode".
func toolTreeAddNode(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "Tree.AddChildNode"

	if len(args) < 4 || len(args) > 6 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 4 to 6 arguments, got %d", toolName, len(args)),
			lang.ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	parentID, _ := args[1].(string)

	nodeType, okType := args[3].(string)
	if !okType {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: node_type argument must be a string, got %T", toolName, args[3]), lang.ErrInvalidArgument)
	}

	var newNodeIDSuggestion string
	if len(args) > 2 && args[2] != nil {
		idSuggestion, ok := args[2].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: new_node_id_suggestion argument must be a string or null, got %T", toolName, args[2]), lang.ErrInvalidArgument)
		}
		newNodeIDSuggestion = idSuggestion
	}

	var nodeValue interface{} = nil
	if len(args) > 4 && args[4] != nil {
		nodeValue = args[4]
	}

	var keyForObjectParent string
	if len(args) > 5 && args[5] != nil {
		key, ok := args[5].(string)
		if !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: key_for_object_parent argument must be a string or null, got %T", toolName, args[5]), lang.ErrInvalidArgument)
		}
		keyForObjectParent = key
	}

	allowedTypes := []string{"string", "number", "boolean", "null", "object", "array", "checklist_item"}
	if !slices.Contains(allowedTypes, nodeType) {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid node_type specified: %q", toolName, nodeType), lang.ErrInvalidArgument)
	}

	if (nodeType == "object" || nodeType == "array") && nodeValue != nil {
		interpreter.Logger().Warn(fmt.Sprintf("%s: node_value provided but ignored for type '%s'", toolName, nodeType))
		nodeValue = nil
	}
	if nodeType == "checklist_item" && nodeValue != nil {
		if _, ok := nodeValue.(string); !ok {
			return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: node_value must be a string for type 'checklist_item', got %T", toolName, nodeValue), lang.ErrInvalidArgument)
		}
	}

	tree, parentNode, err := getNodeFromHandle(interpreter, handleID, parentID, toolName+" (getting parent)")
	if err != nil {
		return nil, err
	}

	var newNodeID string
	if newNodeIDSuggestion != "" {
		if _, exists := tree.NodeMap[newNodeIDSuggestion]; exists {
			return nil, lang.NewRuntimeError(lang.ErrorCodeTreeConstraintViolation,
				fmt.Sprintf("%s: suggested new_node_id '%s' already exists in tree '%s'", toolName, newNodeIDSuggestion, handleID),
				lang.ErrNodeIDExists,
			)
		}
		newNodeID = newNodeIDSuggestion
	} else {
		tempIDCounter := len(tree.NodeMap) + 1
		for {
			genID := "node-" + strconv.Itoa(tempIDCounter)
			if _, exists := tree.NodeMap[genID]; !exists {
				newNodeID = genID
				break
			}
			tempIDCounter++
		}
	}

	newNode := &utils.GenericTreeNode{
		ID:         newNodeID,
		Type:       nodeType,
		Value:      nodeValue,
		Attributes: make(utils.TreeAttrs),
		ChildIDs:   make([]string, 0),
		ParentID:   parentID,
		Tree:       tree,
	}
	tree.NodeMap[newNodeID] = newNode

	// MODIFIED: Allow adding children to 'checklist_root' and 'checklist_item' as if they were arrays.
	if parentNode.Type == "object" {
		if keyForObjectParent == "" {
			return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
				fmt.Sprintf("%s: key_for_object_parent is required when adding a child to an 'object' node", toolName),
				lang.ErrInvalidArgument,
			)
		}
		if parentNode.Attributes == nil {
			parentNode.Attributes = make(utils.TreeAttrs)
		}
		parentNode.Attributes[keyForObjectParent] = newNodeID
	} else if parentNode.Type == "array" || parentNode.Type == "checklist_root" || parentNode.Type == "checklist_item" { // MODIFIED HERE
		// keyForObjectParent (args[5]) should be nil if not an object parent, as handled by ChecklistAddItem.
		// Log a warning if it was somehow provided for these types.
		if keyForObjectParent != "" {
			interpreter.Logger().Warn(fmt.Sprintf("%s: key_for_object_parent '%s' ignored for %s parent '%s'", toolName, keyForObjectParent, parentNode.Type, parentID))
		}
		if parentNode.ChildIDs == nil {
			parentNode.ChildIDs = make([]string, 0)
		}
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNodeID)
	} else {
		return nil, lang.NewRuntimeError(lang.ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: parent node '%s' is type '%s', cannot add children in this manner", toolName, parentID, parentNode.Type),
			lang.ErrNodeWrongType,
		)
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Added new node to tree", toolName), "handle", handleID, "parentId", parentID, "newNodeId", newNodeID, "type", nodeType)
	return newNodeID, nil
}

// --- Tree.RemoveNode (was toolTreeRemoveNode) ---
// Removes a node and all its descendants from the tree.
// Corresponds to ToolSpec "Tree.RemoveNode".
func toolTreeRemoveNode(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "Tree.RemoveNode"

	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)),
			lang.ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	nodeIDToRemove, _ := args[1].(string)

	if nodeIDToRemove == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: node_id cannot be empty", toolName), lang.ErrInvalidArgument)
	}

	tree, nodeToRemove, err := getNodeFromHandle(interpreter, handleID, nodeIDToRemove, toolName+" (getting node to remove)")
	if err != nil {
		return nil, err
	}

	if nodeIDToRemove == tree.RootID {
		return nil, lang.NewRuntimeError(lang.ErrorCodeTreeConstraintViolation,
			fmt.Sprintf("%s: cannot remove root node '%s'", toolName, nodeIDToRemove),
			lang.ErrCannotRemoveRoot,
		)
	}

	if nodeToRemove.ParentID == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: node '%s' is not root but has no ParentID, tree inconsistent", toolName, nodeIDToRemove),
			lang.ErrInternal,
		)
	}

	_, parentNode, parentErr := getNodeFromHandle(interpreter, handleID, nodeToRemove.ParentID, toolName+" (getting parent)")
	if parentErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: parent node '%s' for node '%s' not found: %v", toolName, nodeToRemove.ParentID, nodeIDToRemove, parentErr),
			errors.Join(lang.ErrInternal, parentErr),
		)
	}

	if !removeChildFromParent(parentNode, nodeIDToRemove) {
		interpreter.Logger().Warn(fmt.Sprintf("%s: Node '%s' to remove was not found in its parent's (%s) ChildIDs/Attributes list. Tree might be inconsistent.", toolName, nodeIDToRemove, parentNode.ID))
	}

	removeNodeRecursive(tree, nodeIDToRemove, make(map[string]struct{}))

	interpreter.Logger().Debug(fmt.Sprintf("%s: Removed node and descendants from tree", toolName), "handle", handleID, "nodeId", nodeIDToRemove)
	return nil, nil
}
