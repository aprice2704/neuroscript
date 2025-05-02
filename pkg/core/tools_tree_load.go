// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:15:12 PDT // Fix: Use safe type assertion for input
// filename: pkg/core/tools_tree_load.go

// Package core contains core interpreter functionality, including built-in tools.
package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
)

var toolTreeLoadJSONImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeLoadJSON",
		Description: "Parses a JSON string into an internal tree structure. Returns a tree handle.",
		Args: []ArgSpec{
			{Name: "content", Type: ArgTypeString, Required: true, Description: "JSON content as a string."},
		},
		ReturnType: ArgTypeString, // Returns handle ID string
	},
	Func: toolTreeLoadJSON,
}

// toolTreeLoadJSON parses a JSON string and returns a handle to the generic tree.
func toolTreeLoadJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeLoadJSON"
	// Argument validation (Count should be checked by interpreter's validation layer)
	if len(args) != 1 {
		// This check might be redundant if validation layer is robust, but safer to keep.
		return nil, fmt.Errorf("%w: %s expected 1 argument, got %d", ErrValidationArgCount, toolName, len(args))
	}

	// *** FIXED: Use safe type assertion ***
	jsonContent, ok := args[0].(string)
	if !ok {
		// Return a specific type mismatch error if the input is not a string
		return nil, fmt.Errorf("%w: %s requires a string argument for 'content', got %T", ErrValidationTypeMismatch, toolName, args[0])
	}
	// *** END FIX ***

	var data interface{}
	err := json.Unmarshal([]byte(jsonContent), &data)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTreeJSONUnmarshal, err)
	}

	tree := &GenericTree{
		NodeMap: make(map[string]*GenericTreeNode),
		nextID:  1,
	}

	// Recursive function to build the tree (remains the same)
	var buildNode func(parentID string, key string, value interface{}) (string, error)
	buildNode = func(parentID string, key string, value interface{}) (string, error) {
		var node *GenericTreeNode
		nodeType := ""

		switch v := value.(type) {
		case map[string]interface{}:
			nodeType = "object"
			node = tree.newNode(parentID, nodeType)
			for k, val := range v {
				childID, errBuild := buildNode(node.ID, k, val)
				if errBuild != nil {
					return "", errBuild
				}
				node.Attributes[k] = childID
			}
		case []interface{}:
			nodeType = "array"
			node = tree.newNode(parentID, nodeType)
			node.ChildIDs = make([]string, len(v))
			for i, item := range v {
				childID, errBuild := buildNode(node.ID, strconv.Itoa(i), item)
				if errBuild != nil {
					return "", errBuild
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
		// Ensure ErrTreeBuildFailed is included if not already present
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
		interpreter.Logger().Error("Failed to register GenericTree handle", "error", handleErr)
		return nil, fmt.Errorf("%w: failed to register tree handle: %w", ErrInternalTool, handleErr)
	}

	interpreter.Logger().Debug("Successfully parsed JSON into tree", "rootId", tree.RootID, "nodeCount", len(tree.NodeMap), "handle", handleID)
	return handleID, nil
}
