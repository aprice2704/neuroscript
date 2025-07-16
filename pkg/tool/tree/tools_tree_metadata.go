// NeuroScript Version: 0.6.5
// File version: 3
// Purpose: Corrected GetNodeMetadata to return a standard map[string]interface{} to align with test expectations.
// filename: pkg/tool/tree/tools_tree_metadata.go
// nlines: 150
// risk_rating: LOW

package tree

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// toolTreeGetNodeMetadata retrieves the metadata attributes of a specific node.
func toolTreeGetNodeMetadata(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.GetNodeMetadata"
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)), lang.ErrArgumentMismatch)
	}

	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), lang.ErrInvalidArgument)
	}
	nodeID, okNodeID := args[1].(string)
	if !okNodeID {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: node_id argument must be a string, got %T", toolName, args[1]), lang.ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	// Convert utils.TreeAttrs to a standard map to be test-friendly.
	attributesMap := make(map[string]interface{})
	if node.Attributes != nil {
		for k, v := range node.Attributes {
			attributesMap[k] = v
		}
	}

	return attributesMap, nil
}

// toolTreeSetNodeMetadata sets or updates a string metadata attribute (key-value pair) on any existing node type.
// Corresponds to ToolSpec "Tree.SetNodeMetadata".
func toolTreeSetNodeMetadata(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.SetNodeMetadata"

	if len(args) != 4 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 4 arguments (tree_handle, node_id, metadata_key, metadata_value), got %d", toolName, len(args)),
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
	metaKey, okMetaKey := args[2].(string)
	if !okMetaKey {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: metadata_key argument must be a string, got %T", toolName, args[2]), lang.ErrInvalidArgument)
	}
	// Metadata value can be of any type for setting, but the spec says string. We'll stick to that.
	metaValue, okMetaValue := args[3].(string)
	if !okMetaValue {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: metadata_value argument must be a string, got %T", toolName, args[3]), lang.ErrInvalidArgument)
	}

	if metaKey == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: metadata_key cannot be empty", toolName), lang.ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	if node.Attributes == nil {
		node.Attributes = make(utils.TreeAttrs)
		interpreter.GetLogger().Warn(fmt.Sprintf("%s: Node attributes map was nil for node '%s', initialized.", toolName, nodeID))
	}
	node.Attributes[metaKey] = metaValue

	interpreter.GetLogger().Debug(fmt.Sprintf("%s: Set node metadata attribute", toolName), "handle", handleID, "nodeId", nodeID, "key", metaKey, "value", metaValue)
	return nil, nil
}

// toolTreeRemoveNodeMetadata removes a metadata attribute (key-value pair) from any node.
// Corresponds to ToolSpec "Tree.RemoveNodeMetadata".
func toolTreeRemoveNodeMetadata(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.RemoveNodeMetadata"

	if len(args) != 3 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 arguments (tree_handle, node_id, metadata_key), got %d", toolName, len(args)),
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
	metaKey, okMetaKey := args[2].(string)
	if !okMetaKey {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("%s: metadata_key argument must be a string, got %T", toolName, args[2]), lang.ErrInvalidArgument)
	}

	if metaKey == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("%s: metadata_key cannot be empty", toolName), lang.ErrInvalidArgument)
	}

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	if node.Attributes == nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: node '%s' has no attributes map (key: %q)", toolName, nodeID, metaKey),
			lang.ErrAttributeNotFound,
		)
	}

	if _, keyExists := node.Attributes[metaKey]; !keyExists {
		return nil, lang.NewRuntimeError(lang.ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: metadata_key '%s' not found on node '%s'", toolName, metaKey, nodeID),
			lang.ErrAttributeNotFound,
		)
	}

	delete(node.Attributes, metaKey)
	interpreter.GetLogger().Debug(fmt.Sprintf("%s: Removed node metadata attribute", toolName), "handle", handleID, "nodeId", nodeID, "key", metaKey)
	return nil, nil
}
