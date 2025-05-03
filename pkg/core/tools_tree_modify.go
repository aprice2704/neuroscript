// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 17:39:03 PDT // Add TreeRemoveNode tool implementation
// filename: pkg/core/tools_tree_modify.go

// Package core contains core interpreter functionality, including built-in tools.
package core

import (
	"fmt"
	"slices" // Used for inserting/removing from slices
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

// --- TreeRemoveAttribute ---

var toolTreeRemoveAttributeImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeRemoveAttribute",
		Description: "Removes an attribute (key-value pair) from an object node. " +
			"The target node must be of type 'object'. Returns nil on success.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the object node to modify."},
			{Name: "attribute_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the attribute to remove."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeRemoveAttribute,
}

func toolTreeRemoveAttribute(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeRemoveAttribute"

	// --- Argument Parsing ---
	// Assumes validation layer handles exact count and base types
	handleID := args[0].(string)
	nodeID := args[1].(string)
	attrKey := args[2].(string)

	if attrKey == "" {
		return nil, fmt.Errorf("%w: %s 'attribute_key' cannot be empty", ErrInvalidArgument, toolName)
	}

	// --- Get Target Node ---
	// We don't need the full tree object here, just the node.
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Error already has context
	}

	// --- Validate Target Node Type ---
	if node.Type != "object" {
		return nil, fmt.Errorf("%w: %s target node '%s' is type '%s'", ErrTreeNodeNotObject, toolName, nodeID, node.Type)
	}

	// --- Check Attribute Existence and Remove ---
	// Check if the map exists and the key is present before deleting
	if node.Attributes == nil {
		// If the map is nil, the key definitely doesn't exist
		return nil, fmt.Errorf("%w: %s node '%s' has no attributes to remove from (key: %q)", ErrAttributeNotFound, toolName, nodeID, attrKey)
	}

	_, keyExists := node.Attributes[attrKey]
	if !keyExists {
		// Key doesn't exist in the map
		return nil, fmt.Errorf("%w: %s key '%s' not found on node '%s'", ErrAttributeNotFound, toolName, attrKey, nodeID)
	}

	// Key exists, remove it
	delete(node.Attributes, attrKey)

	interpreter.Logger().Debug("Removed node attribute", "tool", toolName, "handle", handleID, "nodeId", nodeID, "key", attrKey)

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- TreeAddNode ---

var toolTreeAddNodeImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeAddNode",
		Description: "Adds a new node to the tree as a child of a specified parent. " +
			"The new node ID must be unique within the tree. Returns nil on success.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "parent_node_id", Type: ArgTypeString, Required: true, Description: "ID of the node that will become the parent."},
			{Name: "new_node_id", Type: ArgTypeString, Required: true, Description: "Unique ID for the new node to be created."},
			{Name: "node_type", Type: ArgTypeString, Required: true, Description: `Type of the new node (e.g., "string", "number", "boolean", "null", "object", "array").`},
			{Name: "node_value", Type: ArgTypeAny, Required: false, Description: "Value for simple node types (string, number, boolean, null). Ignored for object/array."},
			// Note: Adding attributes/children directly via this tool is complex.
			// Recommend creating object/array node first, then using TreeSetAttribute/TreeAddNode recursively.
			// {Name: "attributes", Type: ArgTypeMap, Required: false, Description: "Optional map of attribute keys to child node IDs (for type 'object')."},
			// {Name: "child_ids", Type: ArgTypeSliceString, Required: false, Description: "Optional slice of child node IDs (for type 'array')."},
			{Name: "index", Type: ArgTypeInt, Required: false, Description: "Optional insertion index for parent's children list (only for 'array' parents, -1 or omitted to append)."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeAddNode,
}

func toolTreeAddNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeAddNode"

	// --- Argument Parsing ---
	// Base arguments (checked by validation)
	handleID := args[0].(string)
	parentID := args[1].(string)
	newNodeID := args[2].(string)
	nodeType := args[3].(string)

	// Optional arguments
	var nodeValue interface{} = nil // Default
	if len(args) > 4 && args[4] != nil {
		nodeValue = args[4]
	}
	index64 := int64(-1) // Default to append
	if len(args) > 5 && args[5] != nil {
		var err error
		index64, err = ConvertToInt64E(args[5])
		if err != nil {
			return nil, fmt.Errorf("%w: %s invalid 'index' argument: %w", ErrInvalidArgument, toolName, err)
		}
	}
	index := int(index64) // Convert after potential conversion error check

	// Basic input validation
	if newNodeID == "" {
		return nil, fmt.Errorf("%w: %s 'new_node_id' cannot be empty", ErrInvalidArgument, toolName)
	}
	// Validate nodeType is one of the allowed types
	allowedTypes := []string{"string", "number", "boolean", "null", "object", "array"}
	if !slices.Contains(allowedTypes, nodeType) {
		return nil, fmt.Errorf("%w: %s invalid 'node_type' specified: %q", ErrInvalidArgument, toolName, nodeType)
	}
	if (nodeType == "object" || nodeType == "array") && nodeValue != nil {
		interpreter.Logger().Warn("node_value provided but ignored for object/array type", "tool", toolName, "nodeType", nodeType)
		nodeValue = nil // Ensure value is nil for complex types
	}

	// --- Get Tree and Parent Node ---
	tree, parentNode, err := getNodeFromHandle(interpreter, handleID, parentID, toolName)
	if err != nil {
		return nil, fmt.Errorf("failed to get parent node '%s': %w", parentID, err)
	}

	// --- Check for Existing Node ID ---
	if _, exists := tree.NodeMap[newNodeID]; exists {
		return nil, fmt.Errorf("%w: %s node ID '%s' already exists in tree '%s'", ErrNodeIDExists, toolName, newNodeID, handleID)
	}

	// --- Create New Node ---
	newNode := &GenericTreeNode{
		ID:         newNodeID,
		Type:       nodeType,
		Value:      nodeValue,               // Will be nil for object/array
		Attributes: make(map[string]string), // Initialize even if object type
		ChildIDs:   make([]string, 0),       // Initialize even if array type
		ParentID:   parentID,
		Tree:       tree, // Back-pointer
	}

	// Add to tree map
	tree.NodeMap[newNodeID] = newNode

	// --- Attach to Parent's Children List ---
	if parentNode.ChildIDs == nil { // Defensive init
		parentNode.ChildIDs = make([]string, 0)
	}

	if parentNode.Type == "array" && index >= 0 {
		// Insert at index for array parent
		if index > len(parentNode.ChildIDs) {
			// Treat out-of-bounds index as append for robustness? Or error?
			// Let's error for now to be strict.
			return nil, fmt.Errorf("%w: %s index %d out of bounds for parent array (len %d)", ErrListIndexOutOfBounds, toolName, index, len(parentNode.ChildIDs))
			// OR: append instead:
			// parentNode.ChildIDs = append(parentNode.ChildIDs, newNodeID)
			// interpreter.Logger().Warn("Index out of bounds, appending instead.", "tool", toolName, "index", index, "len", len(parentNode.ChildIDs))
		} else {
			// Use slices.Insert
			parentNode.ChildIDs = slices.Insert(parentNode.ChildIDs, index, newNodeID)
		}
	} else {
		// Append for non-array parents or index < 0
		if parentNode.Type != "array" && index >= 0 {
			interpreter.Logger().Warn("'index' argument ignored for non-array parent type", "tool", toolName, "parentType", parentNode.Type)
		}
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNodeID)
	}

	interpreter.Logger().Debug("Added new node to tree", "tool", toolName, "handle", handleID, "parentId", parentID, "newNodeId", newNodeID, "type", nodeType)

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- TreeRemoveNode ---

var toolTreeRemoveNodeImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeRemoveNode",
		Description: "Removes a node and all its descendants from the tree. " +
			"Cannot remove the root node. Returns nil on success.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node to remove."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeRemoveNode,
}

func toolTreeRemoveNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeRemoveNode"

	// --- Argument Parsing ---
	handleID := args[0].(string)
	nodeID := args[1].(string)

	if nodeID == "" {
		return nil, fmt.Errorf("%w: %s 'node_id' cannot be empty", ErrInvalidArgument, toolName)
	}

	// --- Get Tree and Node to Remove ---
	tree, nodeToRemove, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Error already contains context
	}

	// --- Check if Root Node ---
	if nodeID == tree.RootID {
		return nil, fmt.Errorf("%w: %s cannot remove root node '%s'", ErrCannotRemoveRoot, toolName, nodeID)
	}

	// --- Get Parent Node ---
	if nodeToRemove.ParentID == "" {
		// This shouldn't happen if it's not the root, indicates inconsistent tree state
		return nil, fmt.Errorf("%w: %s node '%s' is not root but has no parent ID", ErrInternalTool, toolName, nodeID)
	}
	_, parentNode, parentErr := getNodeFromHandle(interpreter, handleID, nodeToRemove.ParentID, toolName+"(getParent)")
	if parentErr != nil {
		// If parent doesn't exist, tree is inconsistent
		return nil, fmt.Errorf("%w: %s parent node '%s' for node '%s' not found: %w", ErrInternalTool, toolName, nodeToRemove.ParentID, nodeID, parentErr)
	}

	// --- Remove from Parent ---
	removedFromParent := removeChildFromParent(parentNode, nodeID)
	if !removedFromParent {
		// Log this inconsistency, but proceed with removing the node from the map anyway
		interpreter.Logger().Warn("Node to remove was not found in its parent's children/attributes list.",
			"tool", toolName, "nodeId", nodeID, "parentId", parentNode.ID)
	}

	// --- Recursively Remove Node and Descendants from NodeMap ---
	// Use a fresh visited map for each top-level remove call
	visited := make(map[string]struct{})
	removeNodeRecursive(tree, nodeID, visited)

	interpreter.Logger().Debug("Removed node and descendants from tree", "tool", toolName, "handle", handleID, "nodeId", nodeID)

	return nil, nil // Return nil on success
}
