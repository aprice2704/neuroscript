// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:40:49 PDT // Add TreeSetAttribute tool
// filename: pkg/core/tools_tree_modify.go

// Package core contains core interpreter functionality, including built-in tools.
package core

import (
	"fmt"
)

// --- TreeModifyNode (Value only) ---

var toolTreeModifyNodeImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeModifyNode",
		Description: "Modifies the 'Value' field of an existing node in a tree. " +
			"Only applicable to nodes with simple types (string, number, boolean, null). " +
			"Returns nil on success.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node to modify."},
			{Name: "modifications", Type: ArgTypeMap, Required: true, Description: "Map containing the modifications. Must have a 'value' key with the new value."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeModifyNode,
}

func toolTreeModifyNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeModifyNode"
	// Assumes validation layer handles arg count and exact types
	handleID := args[0].(string)
	nodeID := args[1].(string)
	modMap := args[2].(map[string]interface{}) // Already validated as map

	// Validate modifications map: must contain 'value' key
	newValue, valueExists := modMap["value"]
	if !valueExists {
		return nil, fmt.Errorf("%w: %s 'modifications' map must contain a 'value' key", ErrInvalidArgument, toolName)
	}

	// Get Node
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	// Check Node Type Compatibility
	if node.Type == "object" || node.Type == "array" {
		return nil, fmt.Errorf("%w: %s (node '%s' is type '%s')", ErrTreeCannotSetValueOnType, toolName, nodeID, node.Type)
	}

	// Apply Modification
	node.Value = newValue
	interpreter.Logger().Debug("Modified node value", "tool", toolName, "handle", handleID, "nodeId", nodeID) // Avoid logging newValue directly

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- TreeSetAttribute ---

var toolTreeSetAttributeImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeSetAttribute",
		Description: "Sets or updates an attribute on an object node, mapping the attribute key to a child node ID. " +
			"The target node must be of type 'object' and the child node must exist. Returns nil on success.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the object node to modify."},
			{Name: "attribute_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the attribute to set."},
			{Name: "child_node_id", Type: ArgTypeString, Required: true, Description: "The ID of the existing node to associate with the key."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeSetAttribute,
}

func toolTreeSetAttribute(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeSetAttribute"

	// --- Argument Parsing ---
	// Assumes validation layer handles exact count and base types
	handleID := args[0].(string)
	nodeID := args[1].(string)
	attrKey := args[2].(string)
	childNodeID := args[3].(string)

	if attrKey == "" {
		return nil, fmt.Errorf("%w: %s 'attribute_key' cannot be empty", ErrInvalidArgument, toolName)
	}
	// childNodeID emptiness is checked by node existence check below

	// --- Get Target Node and Tree ---
	tree, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Error already has context
	}

	// --- Validate Target Node Type ---
	if node.Type != "object" {
		return nil, fmt.Errorf("%w: %s target node '%s' is type '%s'", ErrTreeNodeNotObject, toolName, nodeID, node.Type)
	}

	// --- Validate Child Node Existence ---
	// Check if the child node exists within the *same* tree
	_, childExists := tree.NodeMap[childNodeID]
	if !childExists {
		// Use ErrNotFound to indicate the referenced child is missing
		return nil, fmt.Errorf("%w: %s specified child node ID '%s' not found in tree '%s'", ErrNotFound, toolName, childNodeID, handleID)
	}

	// --- Apply Modification ---
	// Ensure the Attributes map is initialized (should be by newNode, but defensive check)
	if node.Attributes == nil {
		node.Attributes = make(map[string]string)
		// Log this unusual situation
		interpreter.Logger().Warn("Node attributes map was nil, initialized.", "tool", toolName, "nodeId", nodeID)
	}
	node.Attributes[attrKey] = childNodeID

	interpreter.Logger().Debug("Set node attribute", "tool", toolName, "handle", handleID, "nodeId", nodeID, "key", attrKey, "childId", childNodeID)

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- Add TreeRemoveAttribute etc. here later ---
