// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 21:49:55 PDT // Add TreeRenderText tool
// filename: pkg/core/tools_tree.go

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	// Added for strings.Builder and indentation
	// Assuming ErrInternalTool, ErrValidation*, ErrHandle* are defined in errors.go
	// Assuming HandleRegistry logic exists in interpreter.go or similar
)

// --- Generic Tree Representation ---

const GenericTreeHandleType = "GenericTree"
const defaultIndent = "  " // Default indentation string

// GenericTreeNode represents a node within our generic tree structure.
// --- (GenericTreeNode struct remains unchanged) ---
type GenericTreeNode struct {
	ID         string            `json:"id"`         // Unique ID within the tree instance (e.g., "node-1", "node-2")
	Type       string            `json:"type"`       // Node type (e.g., "object", "array", "string", "number", "boolean", "null")
	Value      interface{}       `json:"value"`      // Holds the value for simple types (string, number, bool, nil)
	Attributes map[string]string `json:"attributes"` // For object keys (maps string keys to child node IDs)
	ChildIDs   []string          `json:"children"`   // Ordered list of child node IDs (for arrays)
	ParentID   string            `json:"-"`          // ID of the parent node ("" for root), excluded from direct node JSON
	Tree       *GenericTree      `json:"-"`          // Back-pointer to the tree (for convenience, excluded)
}

// GenericTree holds the entire tree structure associated with a handle.
// --- (GenericTree struct remains unchanged) ---
type GenericTree struct {
	RootID  string                      `json:"rootId"` // ID of the root node
	NodeMap map[string]*GenericTreeNode `json:"-"`      // Map from node ID to the node struct, excluded from direct tree JSON
	nextID  int                         // Internal counter for generating IDs
}

// newNode creates a new node and adds it to the tree's NodeMap.
// --- (newNode method remains unchanged) ---
func (gt *GenericTree) newNode(parentID string, nodeType string) *GenericTreeNode {
	nodeID := "node-" + strconv.Itoa(gt.nextID)
	gt.nextID++
	node := &GenericTreeNode{
		ID:         nodeID,
		Type:       nodeType,
		Attributes: make(map[string]string), // Initialize maps/slices
		ChildIDs:   make([]string, 0),
		ParentID:   parentID,
		Tree:       gt,
	}
	gt.NodeMap[nodeID] = node
	return node
}

// --- Tool Implementations ---

// toolTreeLoadJSON parses a JSON string and returns a handle to the generic tree.
// --- (toolTreeLoadJSON remains unchanged) ---
func toolTreeLoadJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Safeguard type check (though validation layer should handle this)
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: expected 1 argument (content string), got %d", ErrValidationArgCount, len(args))
	}
	jsonContent, ok := args[0].(string)
	if !ok {
		// If validation layer missed it, return appropriate error here
		return nil, fmt.Errorf("%w: expected content argument to be a string, got %T", ErrValidationTypeMismatch, args[0])
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonContent), &data)
	if err != nil {
		// Wrap with specific ErrTreeJSONUnmarshal
		return nil, fmt.Errorf("%w: %w", ErrTreeJSONUnmarshal, err)
	}

	tree := &GenericTree{
		NodeMap: make(map[string]*GenericTreeNode),
		nextID:  1, // Start ID counter consistently
	}

	// Recursive function to build the tree
	var buildNode func(parentID string, key string, value interface{}) (string, error)
	buildNode = func(parentID string, key string, value interface{}) (string, error) {
		var node *GenericTreeNode
		nodeType := ""

		switch v := value.(type) {
		case map[string]interface{}:
			nodeType = "object"
			node = tree.newNode(parentID, nodeType)
			for k, val := range v {
				childID, err := buildNode(node.ID, k, val)
				if err != nil {
					return "", err
				}
				node.Attributes[k] = childID
			}
		case []interface{}:
			nodeType = "array"
			node = tree.newNode(parentID, nodeType)
			node.ChildIDs = make([]string, len(v))
			for i, item := range v {
				childID, err := buildNode(node.ID, strconv.Itoa(i), item)
				if err != nil {
					return "", err
				}
				node.ChildIDs[i] = childID
			}
		case string:
			nodeType = "string"
			node = tree.newNode(parentID, nodeType)
			node.Value = v
		case float64:
			nodeType = "number"
			node = tree.newNode(parentID, nodeType)
			node.Value = v
		case bool:
			nodeType = "boolean"
			node = tree.newNode(parentID, nodeType)
			node.Value = v
		case nil:
			nodeType = "null"
			node = tree.newNode(parentID, nodeType)
			node.Value = nil
		default:
			return "", fmt.Errorf("%w: unsupported JSON type encountered: %T", ErrTreeBuildFailed, value)
		}

		if parentID == "" {
			tree.RootID = node.ID
		}
		return node.ID, nil
	}

	_, err = buildNode("", "", data)
	if err != nil {
		if !errors.Is(err, ErrTreeBuildFailed) {
			err = fmt.Errorf("%w: %w", ErrTreeBuildFailed, err)
		}
		return nil, err
	}
	if tree.RootID == "" {
		return nil, fmt.Errorf("%w: failed to determine root node after parsing JSON", ErrTreeBuildFailed)
	}

	handleID, handleErr := interpreter.RegisterHandle(tree, GenericTreeHandleType)
	if handleErr != nil {
		interpreter.Logger().Error("Failed to register GenericTree handle: %v", handleErr)
		return nil, fmt.Errorf("%w: failed to register tree handle: %w", ErrInternalTool, handleErr)
	}

	interpreter.Logger().Debug("[TOOL-TreeLoadJSON] Successfully parsed JSON. RootID: %s, Total Nodes: %d. Handle: %s", tree.RootID, len(tree.NodeMap), handleID)
	return handleID, nil
}

// getNodeFromHandle retrieves the GenericTree and the specific GenericTreeNode.
// --- (getNodeFromHandle remains unchanged) ---
func getNodeFromHandle(interpreter *Interpreter, handleID, nodeID, toolName string) (*GenericTree, *GenericTreeNode, error) {
	if handleID == "" || nodeID == "" {
		return nil, nil, fmt.Errorf("%w: %s requires non-empty 'tree_handle' and 'node_id'", ErrValidationRequiredArgNil, toolName)
	}

	obj, err := interpreter.GetHandleValue(handleID, GenericTreeHandleType)
	if err != nil {
		return nil, nil, fmt.Errorf("%s: error getting handle '%s': %w", toolName, handleID, err)
	}

	tree, ok := obj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, nil, fmt.Errorf("%w: %s handle '%s' contains unexpected data type (%T), expected %s", ErrHandleInvalid, toolName, handleID, obj, GenericTreeHandleType)
	}

	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, nil, fmt.Errorf("%w: %s node ID '%s' not found in tree handle '%s'", ErrNotFound, toolName, nodeID, handleID)
	}

	return tree, node, nil
}

// toolTreeGetNode returns information about a specific node.
// --- (toolTreeGetNode remains unchanged) ---
func toolTreeGetNode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeGetNode"
	// Validation ensures 2 string args
	handleID := args[0].(string)
	nodeID := args[1].(string)

	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err // Return the detailed error from helper
	}

	// Convert node data to a map for NeuroScript
	nodeMap := map[string]interface{}{
		"id":    node.ID,
		"type":  node.Type,
		"value": node.Value,
		"attributes": func() map[string]interface{} {
			attrs := make(map[string]interface{})
			for k, v := range node.Attributes {
				attrs[k] = v
			}
			return attrs
		}(),
		"children": func() []interface{} {
			children := make([]interface{}, len(node.ChildIDs))
			for i, id := range node.ChildIDs {
				children[i] = id
			}
			return children
		}(),
		"parentId": node.ParentID,
	}
	return nodeMap, nil
}

// toolTreeGetChildren returns a list of child node IDs.
// --- (toolTreeGetChildren remains unchanged) ---
func toolTreeGetChildren(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeGetChildren"
	// Validation ensures 2 string args
	handleID := args[0].(string)
	nodeID := args[1].(string)
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}
	children := make([]interface{}, len(node.ChildIDs))
	for i, id := range node.ChildIDs {
		children[i] = id
	}
	return children, nil
}

// toolTreeGetParent returns the parent node ID (string).
// --- (toolTreeGetParent remains unchanged) ---
func toolTreeGetParent(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeGetParent"
	// Validation ensures 2 string args
	handleID := args[0].(string)
	nodeID := args[1].(string)
	_, node, err := getNodeFromHandle(interpreter, handleID, nodeID, toolName)
	if err != nil {
		return nil, err
	}
	return node.ParentID, nil
}

// --- Registration ---

func registerTreeTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{ // TreeLoadJSON spec...
			Spec: ToolSpec{Name: "TreeLoadJSON", Description: "Parses a JSON string into an internal tree structure. Returns a tree handle.", Args: []ArgSpec{{Name: "content", Type: ArgTypeString, Required: true, Description: "JSON content as a string."}}, ReturnType: ArgTypeString},
			Func: toolTreeLoadJSON,
		},
		{ // TreeGetNode spec...
			Spec: ToolSpec{Name: "TreeGetNode", Description: "Retrieves information about a specific node within a tree handle.", Args: []ArgSpec{{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."}, {Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node within the tree."}}, ReturnType: ArgTypeMap},
			Func: toolTreeGetNode,
		},
		{ // TreeGetChildren spec...
			Spec: ToolSpec{Name: "TreeGetChildren", Description: "Returns a list of child node IDs for a given node.", Args: []ArgSpec{{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."}, {Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node."}}, ReturnType: ArgTypeSliceString},
			Func: toolTreeGetChildren,
		},
		{ // TreeGetParent spec...
			Spec: ToolSpec{Name: "TreeGetParent", Description: "Returns the parent node ID for a given node (empty string for root).", Args: []ArgSpec{{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."}, {Name: "node_id", Type: ArgTypeString, Required: true, Description: "Unique ID of the node."}}, ReturnType: ArgTypeString},
			Func: toolTreeGetParent,
		},
		{ // TreeFormatJSON spec...
			Spec: ToolSpec{Name: "TreeFormatJSON", Description: "Serializes the tree structure associated with a handle back into a formatted JSON string.", Args: []ArgSpec{{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."}}, ReturnType: ArgTypeString},
			Func: toolTreeFormatJSON,
		},
		// +++ Add TreeRenderText Spec +++
		{
			Spec: ToolSpec{
				Name:        "TreeRenderText",
				Description: "Renders the tree structure associated with a handle as an indented text string.",
				Args: []ArgSpec{
					{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
					// Optional: Add args for indent string, max depth later if needed
				},
				ReturnType: ArgTypeString, // Returns the formatted text tree
			},
			Func: toolTreeRenderText,
		},
		// +++ End TreeRenderText Spec +++
	}

	// --- (Registration loop remains unchanged) ---
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			fmt.Printf("! Error registering tool %s: %v\n", tool.Spec.Name, err)
			// Consider collecting errors instead of returning on first failure
			// return fmt.Errorf("failed to register tree tool %q: %w", tool.Spec.Name, err)
		}
	}
	return nil
}
