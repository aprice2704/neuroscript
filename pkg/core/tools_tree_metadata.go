// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 19:06:05 PM PDT // Cleaned whitespace, confirmed implementation
// filename: pkg/core/tools_tree_metadata.go

package core

import (
	"errors"
	"fmt"
)

// --- TreeSetNodeMetadata ---

var toolTreeSetNodeMetadataImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeSetNodeMetadata",
		Description: "Sets or updates a string metadata attribute (key-value pair) on any existing node type. " +
			"This is intended for simple string metadata, not for linking object keys to child nodes (use TreeSetAttribute for that). " +
			"Returns nil on success.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node to modify."},
			{Name: "metadata_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the metadata attribute to set."},
			{Name: "metadata_value", Type: ArgTypeString, Required: true, Description: "The string value to associate with the key."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeSetNodeMetadata,
}

func toolTreeSetNodeMetadata(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeSetNodeMetadata"

	// --- Argument Parsing ---
	handleID := args[0].(string)
	nodeID := args[1].(string)
	metaKey := args[2].(string)
	metaValue := args[3].(string)

	if metaKey == "" {
		return nil, fmt.Errorf("%w: %s 'metadata_key' cannot be empty", ErrInvalidArgument, toolName)
	}
	// metaValue can be empty

	// --- Get Target Node ---
	// No need for the full tree object here
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Error already has context
	}

	// --- Apply Modification ---
	// Ensure the Attributes map is initialized
	if node.Attributes == nil {
		node.Attributes = make(map[string]string)
		interpreter.Logger().Warn("Node attributes map was nil, initialized.", "tool", toolName, "nodeId", nodeID)
	}
	node.Attributes[metaKey] = metaValue // Set the string value

	interpreter.Logger().Debug("Set node metadata attribute", "tool", toolName, "handle", handleID, "nodeId", nodeID, "key", metaKey, "value", metaValue)

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- TreeRemoveNodeMetadata ---

var toolTreeRemoveNodeMetadataImpl = ToolImplementation{
	Spec: ToolSpec{
		Name: "TreeRemoveNodeMetadata",
		Description: "Removes a metadata attribute (key-value pair) from any node that has attributes. " +
			"Returns nil on success, or ErrAttributeNotFound if the key doesn't exist.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node to modify."},
			{Name: "metadata_key", Type: ArgTypeString, Required: true, Description: "The key (name) of the metadata attribute to remove."},
		},
		ReturnType: ArgTypeNil, // Returns nil on success
	},
	Func: toolTreeRemoveNodeMetadata,
}

func toolTreeRemoveNodeMetadata(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeRemoveNodeMetadata"

	// --- Argument Parsing ---
	handleID := args[0].(string)
	nodeID := args[1].(string)
	metaKey := args[2].(string)

	if metaKey == "" {
		return nil, fmt.Errorf("%w: %s 'metadata_key' cannot be empty", ErrInvalidArgument, toolName)
	}

	// --- Get Target Node ---
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Error already has context
	}

	// --- Check Attribute Existence and Remove ---
	if node.Attributes == nil {
		return nil, fmt.Errorf("%w: %s node '%s' has no attributes map (key: %q)", ErrAttributeNotFound, toolName, nodeID, metaKey)
	}

	_, keyExists := node.Attributes[metaKey]
	if !keyExists {
		return nil, fmt.Errorf("%w: %s key '%s' not found on node '%s'", ErrAttributeNotFound, toolName, metaKey, nodeID)
	}

	// Key exists, remove it
	delete(node.Attributes, metaKey)

	interpreter.Logger().Debug("Removed node metadata attribute", "tool", toolName, "handle", handleID, "nodeId", nodeID, "key", metaKey)

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- Registration Function ---

// registerTreeMetadataTools registers the tree metadata manipulation tools.
// This function itself isn't called directly by the interpreter, but is intended
// to be called by a central registration mechanism (e.g., in tools_register.go).
func registerTreeMetadataTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("registerTreeMetadataTools called with nil registry")
	}

	toolsToRegister := []ToolImplementation{
		toolTreeSetNodeMetadataImpl,
		toolTreeRemoveNodeMetadataImpl,
	}

	var registrationErrors []error
	for _, tool := range toolsToRegister {
		if err := registry.RegisterTool(tool); err != nil {
			// Log the error during registration attempt
			fmt.Printf("! Error registering tree metadata tool %s: %v\n", tool.Spec.Name, err)
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register tree metadata tool %q: %w", tool.Spec.Name, err))
		}
	}

	// Combine multiple registration errors if any occurred
	if len(registrationErrors) > 0 {
		return errors.Join(registrationErrors...)
	}
	return nil
}
