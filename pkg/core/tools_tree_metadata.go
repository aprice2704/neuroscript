// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Removed local ToolImplementations and registration func, standardized error handling.
// nlines: 100 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_tree_metadata.go

package core

import (
	"fmt"
)

// toolTreeSetNodeMetadata sets or updates a string metadata attribute (key-value pair) on any existing node type.
// Corresponds to ToolSpec "Tree.SetNodeMetadata".
func toolTreeSetNodeMetadata(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.SetNodeMetadata"

	// Expected args: tree_handle (string), node_id (string), metadata_key (string), metadata_value (string)
	if len(args) != 4 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 4 arguments (tree_handle, node_id, metadata_key, metadata_value), got %d", toolName, len(args)),
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
	metaKey, okMetaKey := args[2].(string)
	if !okMetaKey {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: metadata_key argument must be a string, got %T", toolName, args[2]), ErrInvalidArgument)
	}
	metaValue, okMetaValue := args[3].(string)
	if !okMetaValue {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: metadata_value argument must be a string, got %T", toolName, args[3]), ErrInvalidArgument)
	}

	if metaKey == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: metadata_key cannot be empty", toolName), ErrInvalidArgument)
	}
	// metaValue can be an empty string.

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle returns RuntimeError
	}

	// Ensure the Attributes map is initialized (GenericTreeNode.NewNode initializes it)
	if node.Attributes == nil {
		node.Attributes = make(TreeAttrs)
		interpreter.Logger().Warn(fmt.Sprintf("%s: Node attributes map was nil for node '%s', initialized.", toolName, nodeID))
	}
	node.Attributes[metaKey] = metaValue // Set the string value

	interpreter.Logger().Debug(fmt.Sprintf("%s: Set node metadata attribute", toolName), "handle", handleID, "nodeId", nodeID, "key", metaKey, "value", metaValue)
	return nil, nil
}

// toolTreeRemoveNodeMetadata removes a metadata attribute (key-value pair) from any node.
// Corresponds to ToolSpec "Tree.RemoveNodeMetadata".
func toolTreeRemoveNodeMetadata(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.RemoveNodeMetadata"

	// Expected args: tree_handle (string), node_id (string), metadata_key (string)
	if len(args) != 3 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 arguments (tree_handle, node_id, metadata_key), got %d", toolName, len(args)),
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
	metaKey, okMetaKey := args[2].(string)
	if !okMetaKey {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: metadata_key argument must be a string, got %T", toolName, args[2]), ErrInvalidArgument)
	}

	if metaKey == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: metadata_key cannot be empty", toolName), ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle returns RuntimeError
	}

	if node.Attributes == nil {
		// If the map is nil, the key definitely doesn't exist.
		return nil, NewRuntimeError(ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: node '%s' has no attributes map (key: %q)", toolName, nodeID, metaKey),
			ErrAttributeNotFound,
		)
	}

	if _, keyExists := node.Attributes[metaKey]; !keyExists {
		return nil, NewRuntimeError(ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: metadata_key '%s' not found on node '%s'", toolName, metaKey, nodeID),
			ErrAttributeNotFound,
		)
	}

	delete(node.Attributes, metaKey)
	interpreter.Logger().Debug(fmt.Sprintf("%s: Removed node metadata attribute", toolName), "handle", handleID, "nodeId", nodeID, "key", metaKey)
	return nil, nil
}
