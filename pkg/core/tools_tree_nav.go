// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:19:59 PDT // Fix: Remove strict type check in toolTreeGetChildren
// filename: pkg/core/tools_tree_nav.go

// Package core contains core interpreter functionality, including built-in tools.
package core

var toolTreeGetNodeImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeGetNode",
		Description: "Retrieves information about a specific node within a tree handle.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node within the tree."},
		},
		ReturnType: ArgTypeMap,
	},
	Func: toolTreeGetNode,
}

var toolTreeGetChildrenImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeGetChildren",
		Description: "Returns a list of child node IDs for a given node. Returns empty list for non-array nodes.", // Updated description
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node."}, // Removed "must be of type 'array'"
		},
		ReturnType: ArgTypeSliceString, // Returns []string
	},
	Func: toolTreeGetChildren,
}

var toolTreeGetParentImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeGetParent",
		Description: "Returns the parent node ID for a given node (empty string for root).",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node."},
		},
		ReturnType: ArgTypeString,
	},
	Func: toolTreeGetParent,
}

// toolTreeGetNode returns information about a specific node.
func toolTreeGetNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeGetNode"
	// Assumes validation layer handles arg count and type checking.
	handleID := args[0].(string)
	nodeID := args[1].(string)

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Return the detailed error from helper
	}

	// Convert node data to a map for NeuroScript
	// Use anonymous functions to avoid allocating empty maps/slices if not needed
	nodeMap := map[string]interface{}{
		"id":    node.ID,
		"type":  node.Type,
		"value": node.Value, // Will be nil for object/array types
		"attributes": func() map[string]interface{} {
			if len(node.Attributes) == 0 {
				return nil // Return nil instead of empty map
			}
			attrs := make(map[string]interface{}, len(node.Attributes))
			for k, v := range node.Attributes {
				attrs[k] = v // Return child node IDs as strings
			}
			return attrs
		}(),
		"children": func() []interface{} {
			if len(node.ChildIDs) == 0 {
				return nil // Return nil instead of empty slice
			}
			children := make([]interface{}, len(node.ChildIDs))
			for i, id := range node.ChildIDs {
				children[i] = id // Return child node IDs as strings
			}
			return children
		}(),
		"parentId": node.ParentID, // Will be "" for root
	}
	return nodeMap, nil
}

// toolTreeGetChildren returns a list of child node IDs for a given node.
// If the node is not an array type, it correctly returns an empty list.
func toolTreeGetChildren(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeGetChildren"
	// Assumes validation layer handles arg count and type checking.
	handleID := args[0].(string)
	nodeID := args[1].(string)

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}

	// *** REMOVED: Check for node.Type == "array" ***
	// It's valid to ask for children of any node type; non-arrays just have none.

	// Convert []string to []interface{} for return.
	// If node.ChildIDs is empty (because it's not an array or an empty array),
	// this correctly creates and returns an empty []interface{}.
	children := make([]interface{}, len(node.ChildIDs))
	for i, id := range node.ChildIDs {
		children[i] = id
	}
	return children, nil
}

// toolTreeGetParent returns the parent node ID (string).
func toolTreeGetParent(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeGetParent"
	// Assumes validation layer handles arg count and type checking.
	handleID := args[0].(string)
	nodeID := args[1].(string)

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}
	return node.ParentID, nil // ParentID is already a string ("" for root)
}
