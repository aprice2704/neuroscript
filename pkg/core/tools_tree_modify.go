// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Removed local ToolImplementations, standardized errors, aligned functions with tooldefs.
// nlines: 330 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_modify.go

package core

import (
	"errors" // Required for errors.Is/Join
	"fmt"
	"slices" // Used for inserting/removing from slices
	"strconv"
)

// --- Tree.SetValue (was toolTreeModifyNode) ---
// Sets the value of an existing leaf node.
// Corresponds to ToolSpec "Tree.SetValue".
func toolTreeModifyNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.SetValue"

	// Expected args: tree_handle (string), node_id (string), value (any)
	if len(args) != 3 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 arguments (tree_handle, node_id, value), got %d", toolName, len(args)),
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

	newValue := args[2] // Value can be any type, validation of its suitability for the node type happens below.

	// Get Node
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle already returns a RuntimeError
	}

	// Check Node Type Compatibility for setting a direct value
	// checklist_item nodes can have their text value modified.
	// Other types like "string", "number", "boolean", "null" are also fine.
	if node.Type == "object" || node.Type == "array" {
		return nil, NewRuntimeError(ErrorCodeTreeConstraintViolation,
			fmt.Sprintf("%s: cannot set value directly on node '%s' of type '%s'; use object/array modification tools", toolName, nodeID, node.Type),
			ErrCannotSetValueOnType, // Specific sentinel for this case
		)
	}

	// Validate the type of newValue against node.Type if strict type checking is desired here.
	// For now, we allow setting, assuming NeuroScript's dynamic typing handles it,
	// or that specific node types might have implicit conversions or accept various underlying Go types.
	// E.g., a "number" node might accept int64 or float64.
	// If node.Type is "string", newValue should ideally be a string.
	// If node.Type is "number", newValue should be float64 or int64.
	// If node.Type is "boolean", newValue should be bool.
	// If node.Type is "null", newValue should be nil.
	// A more robust implementation might check:
	// switch node.Type {
	// case "string": if _, ok := newValue.(string); !ok { /* error */ }
	// case "number": if _, okF := newValue.(float64); !okF { if _, okI := newValue.(int64); !okI { /* error */ } }
	// ... etc.
	// }
	// For now, simpler assignment:
	node.Value = newValue
	interpreter.Logger().Debug(fmt.Sprintf("%s: Modified node value", toolName), "handle", handleID, "nodeId", nodeID)
	// Avoid logging newValue directly due to potential size/sensitivity.

	return nil, nil // Return nil (NeuroScript null) on success
}

// --- Tree.SetObjectAttribute (was toolTreeSetAttribute) ---
// Sets or updates an attribute on an object node, mapping the attribute key to a child node ID.
// Corresponds to ToolSpec "Tree.SetObjectAttribute".
func toolTreeSetAttribute(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.SetObjectAttribute"

	// Expected args: tree_handle (string), object_node_id (string), attribute_key (string), child_node_id (string)
	if len(args) != 4 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 4 arguments (tree_handle, object_node_id, attribute_key, child_node_id), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string) // Type already validated by spec
	objectNodeID, _ := args[1].(string)
	attrKey, _ := args[2].(string)
	childNodeID, _ := args[3].(string)

	if attrKey == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: attribute_key cannot be empty", toolName), ErrInvalidArgument)
	}
	if childNodeID == "" { // child_node_id must point to an existing node
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: child_node_id cannot be empty", toolName), ErrInvalidArgument)
	}

	tree, objectNode, err := getNodeFromHandle(interpreter, handleID, objectNodeID, toolName)
	if err != nil {
		return nil, err
	}

	if objectNode.Type != "object" {
		return nil, NewRuntimeError(ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: target node '%s' is type '%s', expected 'object'", toolName, objectNodeID, objectNode.Type),
			ErrTreeNodeNotObject,
		)
	}

	// Validate Child Node Existence in the same tree
	if _, childExists := tree.NodeMap[childNodeID]; !childExists {
		return nil, NewRuntimeError(ErrorCodeKeyNotFound, // More specific than generic ErrNotFound
			fmt.Sprintf("%s: specified child_node_id '%s' not found in tree '%s'", toolName, childNodeID, handleID),
			ErrNotFound, // General sentinel
		)
	}

	if objectNode.Attributes == nil { // Should be initialized by NewNode
		objectNode.Attributes = make(map[string]string)
		interpreter.Logger().Warn(fmt.Sprintf("%s: Node attributes map was nil for node '%s', initialized.", toolName, objectNodeID))
	}
	objectNode.Attributes[attrKey] = childNodeID

	interpreter.Logger().Debug(fmt.Sprintf("%s: Set object attribute", toolName), "handle", handleID, "objectNodeId", objectNodeID, "key", attrKey, "childId", childNodeID)
	return nil, nil
}

// --- Tree.RemoveObjectAttribute (was toolTreeRemoveAttribute) ---
// Removes an attribute from an object node.
// Corresponds to ToolSpec "Tree.RemoveObjectAttribute".
func toolTreeRemoveAttribute(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.RemoveObjectAttribute"

	// Expected args: tree_handle (string), object_node_id (string), attribute_key (string)
	if len(args) != 3 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 arguments (tree_handle, object_node_id, attribute_key), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	objectNodeID, _ := args[1].(string)
	attrKey, _ := args[2].(string)

	if attrKey == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: attribute_key cannot be empty", toolName), ErrInvalidArgument)
	}

	_, objectNode, err := getNodeFromHandle(interpreter, handleID, objectNodeID, toolName)
	if err != nil {
		return nil, err
	}

	if objectNode.Type != "object" {
		return nil, NewRuntimeError(ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: target node '%s' is type '%s', expected 'object'", toolName, objectNodeID, objectNode.Type),
			ErrTreeNodeNotObject,
		)
	}

	if objectNode.Attributes == nil {
		return nil, NewRuntimeError(ErrorCodeAttributeNotFound, // No attributes map means key cannot exist
			fmt.Sprintf("%s: node '%s' has no attributes to remove from (key: %q)", toolName, objectNodeID, attrKey),
			ErrAttributeNotFound,
		)
	}

	if _, keyExists := objectNode.Attributes[attrKey]; !keyExists {
		return nil, NewRuntimeError(ErrorCodeAttributeNotFound,
			fmt.Sprintf("%s: attribute_key '%s' not found on node '%s'", toolName, attrKey, objectNodeID),
			ErrAttributeNotFound,
		)
	}

	delete(objectNode.Attributes, attrKey)
	interpreter.Logger().Debug(fmt.Sprintf("%s: Removed object attribute", toolName), "handle", handleID, "objectNodeId", objectNodeID, "key", attrKey)
	return nil, nil
}

// --- Tree.AddChildNode (was toolTreeAddNode) ---
// Adds a new child node to an existing parent node.
// Corresponds to ToolSpec "Tree.AddChildNode".
func toolTreeAddNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.AddChildNode"

	// Expected args: tree_handle, parent_node_id, new_node_id_suggestion (optional), node_type, value (optional), key_for_object_parent (optional)
	if len(args) < 4 || len(args) > 6 { // Min 4 args, max 6
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 4 to 6 arguments, got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	parentID, _ := args[1].(string)
	// arg[2] is new_node_id_suggestion (string, optional)
	// arg[3] is node_type (string)
	// arg[4] is value (any, optional)
	// arg[5] is key_for_object_parent (string, optional)

	nodeType, okType := args[3].(string)
	if !okType {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: node_type argument must be a string, got %T", toolName, args[3]), ErrInvalidArgument)
	}

	var newNodeIDSuggestion string
	if len(args) > 2 && args[2] != nil {
		idSuggestion, ok := args[2].(string)
		if !ok {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: new_node_id_suggestion argument must be a string or null, got %T", toolName, args[2]), ErrInvalidArgument)
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
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: key_for_object_parent argument must be a string or null, got %T", toolName, args[5]), ErrInvalidArgument)
		}
		keyForObjectParent = key
	}

	allowedTypes := []string{"string", "number", "boolean", "null", "object", "array", "checklist_item"}
	if !slices.Contains(allowedTypes, nodeType) {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: invalid node_type specified: %q", toolName, nodeType), ErrInvalidArgument)
	}

	if (nodeType == "object" || nodeType == "array") && nodeValue != nil {
		interpreter.Logger().Warn(fmt.Sprintf("%s: node_value provided but ignored for type '%s'", toolName, nodeType))
		nodeValue = nil
	}
	if nodeType == "checklist_item" && nodeValue != nil {
		if _, ok := nodeValue.(string); !ok {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: node_value must be a string for type 'checklist_item', got %T", toolName, nodeValue), ErrInvalidArgument)
		}
	}

	tree, parentNode, err := getNodeFromHandle(interpreter, handleID, parentID, toolName+" (getting parent)")
	if err != nil {
		return nil, err
	}

	// Determine the new node's ID
	var newNodeID string
	if newNodeIDSuggestion != "" {
		if _, exists := tree.NodeMap[newNodeIDSuggestion]; exists {
			return nil, NewRuntimeError(ErrorCodeTreeConstraintViolation,
				fmt.Sprintf("%s: suggested new_node_id '%s' already exists in tree '%s'", toolName, newNodeIDSuggestion, handleID),
				ErrNodeIDExists,
			)
		}
		newNodeID = newNodeIDSuggestion
	} else {
		// Generate a new ID if no suggestion or suggestion is empty
		// tree.NewNode would generate one like "node-X", but here we need to ensure it's added to map *with this ID*.
		// So, generate ID first, then create node.
		// This deviates from tree.NewNode's internal ID generation if we are to allow user-suggested IDs.
		// For now, if newNodeIDSuggestion is empty, let tree.NewNode generate it.
		// If a suggestion is provided, we need a way to tell NewNode to use it or create a node and set its ID.
		// Let's modify NewNode slightly or handle ID assignment carefully.
		// For now, we'll rely on tree.NewNode for generation if suggestion is empty.
		// If suggestion is provided, we'll need to assign it after creation by tree.NewNode, which is not ideal as NewNode adds to map.
		// Simpler path: if suggestion, check existence. If not, use it. If empty, generate.
		// This requires NewNode to accept an optional ID. Or a new helper.
		// Let's assume `tree.GenerateNodeID()` for now if suggestion is empty, and manually check for collision.
		if newNodeIDSuggestion == "" {
			// This loop is a placeholder for a robust unique ID generation within the tree context.
			// tree.nextID is not directly accessible here in the same way.
			// A method on GenericTree like `GenerateUniqueID()` would be better.
			tempIDCounter := len(tree.NodeMap) + 1 // Simple, potentially colliding in long-running scenarios
			for {
				genID := "node-" + strconv.Itoa(tempIDCounter)
				if _, exists := tree.NodeMap[genID]; !exists {
					newNodeID = genID
					break
				}
				tempIDCounter++
			}
		} else {
			newNodeID = newNodeIDSuggestion // Already checked for existence
		}
	}

	// Create node (without adding to parent's children/attributes yet)
	newNode := &GenericTreeNode{
		ID:         newNodeID,
		Type:       nodeType,
		Value:      nodeValue,
		Attributes: make(map[string]string),
		ChildIDs:   make([]string, 0),
		ParentID:   parentID,
		Tree:       tree,
	}
	tree.NodeMap[newNodeID] = newNode // Add to the tree's central map

	// Attach to parent
	if parentNode.Type == "object" {
		if keyForObjectParent == "" {
			return nil, NewRuntimeError(ErrorCodeArgMismatch,
				fmt.Sprintf("%s: key_for_object_parent is required when adding a child to an 'object' node", toolName),
				ErrInvalidArgument,
			)
		}
		if parentNode.Attributes == nil {
			parentNode.Attributes = make(map[string]string)
		}
		parentNode.Attributes[keyForObjectParent] = newNodeID
	} else if parentNode.Type == "array" {
		if keyForObjectParent != "" {
			interpreter.Logger().Warn(fmt.Sprintf("%s: key_for_object_parent '%s' ignored for array parent '%s'", toolName, keyForObjectParent, parentID))
		}
		if parentNode.ChildIDs == nil {
			parentNode.ChildIDs = make([]string, 0)
		}
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNodeID) // Append by default
	} else {
		// Cannot add keyed or indexed children to simple types
		return nil, NewRuntimeError(ErrorCodeNodeWrongType,
			fmt.Sprintf("%s: parent node '%s' is type '%s', cannot add children in this manner", toolName, parentID, parentNode.Type),
			ErrNodeWrongType,
		)
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Added new node to tree", toolName), "handle", handleID, "parentId", parentID, "newNodeId", newNodeID, "type", nodeType)
	return newNodeID, nil // Return the new node's ID
}

// --- Tree.RemoveNode (was toolTreeRemoveNode) ---
// Removes a node and all its descendants from the tree.
// Corresponds to ToolSpec "Tree.RemoveNode".
func toolTreeRemoveNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.RemoveNode"

	// Expected args: tree_handle (string), node_id (string)
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 2 arguments (tree_handle, node_id), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}

	handleID, _ := args[0].(string)
	nodeIDToRemove, _ := args[1].(string)

	if nodeIDToRemove == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: node_id cannot be empty", toolName), ErrInvalidArgument)
	}

	tree, nodeToRemove, err := getNodeFromHandle(interpreter, handleID, nodeIDToRemove, toolName+" (getting node to remove)")
	if err != nil {
		return nil, err
	}

	if nodeIDToRemove == tree.RootID {
		return nil, NewRuntimeError(ErrorCodeTreeConstraintViolation,
			fmt.Sprintf("%s: cannot remove root node '%s'", toolName, nodeIDToRemove),
			ErrCannotRemoveRoot,
		)
	}

	if nodeToRemove.ParentID == "" {
		// Should not happen if not root, implies inconsistent tree.
		return nil, NewRuntimeError(ErrorCodeInternal,
			fmt.Sprintf("%s: node '%s' is not root but has no ParentID, tree inconsistent", toolName, nodeIDToRemove),
			ErrInternal,
		)
	}

	_, parentNode, parentErr := getNodeFromHandle(interpreter, handleID, nodeToRemove.ParentID, toolName+" (getting parent)")
	if parentErr != nil {
		return nil, NewRuntimeError(ErrorCodeInternal, // Parent of a non-root node must exist
			fmt.Sprintf("%s: parent node '%s' for node '%s' not found: %v", toolName, nodeToRemove.ParentID, nodeIDToRemove, parentErr),
			errors.Join(ErrInternal, parentErr), // Or ErrNotFound if it's considered a lookup failure
		)
	}

	// Remove from parent's ChildIDs or Attributes
	// This uses the existing helper which might need adjustment if parent linking changes.
	if !removeChildFromParent(parentNode, nodeIDToRemove) {
		interpreter.Logger().Warn(fmt.Sprintf("%s: Node '%s' to remove was not found in its parent's (%s) ChildIDs/Attributes list. Tree might be inconsistent.", toolName, nodeIDToRemove, parentNode.ID))
		// Continue to remove from NodeMap regardless, as that's the primary goal.
	}

	// Recursively remove node and its descendants from NodeMap
	removeNodeRecursive(tree, nodeIDToRemove, make(map[string]struct{})) // Use a fresh visited map

	interpreter.Logger().Debug(fmt.Sprintf("%s: Removed node and descendants from tree", toolName), "handle", handleID, "nodeId", nodeIDToRemove)
	return nil, nil
}
