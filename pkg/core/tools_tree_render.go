// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:15:12 PDT // Fix: Match test output format for TreeRenderText
// filename: pkg/core/tools_tree_render.go

package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
)

// --- toolTreeFormatJSON Implementation (No change needed here) ---

var toolTreeFormatJSONImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeFormatJSON",
		Description: "Serializes the tree structure associated with a handle back into a formatted JSON string.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
		},
		ReturnType: ArgTypeString,
	},
	Func: toolTreeFormatJSON,
}

func toolTreeFormatJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeFormatJSON"
	handleID := args[0].(string)

	tree, err := getTreeFromHandle(interpreter, handleID, toolName) // Use helper
	if err != nil {
		return nil, err
	}

	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists {
		return nil, fmt.Errorf("%w: %s cannot find root node ID '%s' in tree handle '%s'", ErrInternalTool, toolName, tree.RootID, handleID)
	}

	var buildOutput func(node *GenericTreeNode) (interface{}, error)
	buildOutput = func(node *GenericTreeNode) (interface{}, error) {
		// ... (implementation unchanged) ...
		if node == nil {
			return nil, fmt.Errorf("%w: attempted to build output from nil node", ErrInternalTool)
		}
		switch node.Type {
		case "object":
			objMap := make(map[string]interface{})
			keys := make([]string, 0, len(node.Attributes))
			for k := range node.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				childID := node.Attributes[key]
				childNode, ok := tree.NodeMap[childID]
				if !ok {
					return nil, fmt.Errorf("%w: %s child node ID '%s' (key '%s') not found", ErrInternalTool, toolName, childID, key)
				}
				childValue, buildErr := buildOutput(childNode)
				if buildErr != nil {
					return nil, buildErr
				}
				objMap[key] = childValue
			}
			return objMap, nil
		case "array":
			arrSlice := make([]interface{}, len(node.ChildIDs))
			for i, childID := range node.ChildIDs {
				childNode, ok := tree.NodeMap[childID]
				if !ok {
					return nil, fmt.Errorf("%w: %s child node ID '%s' (index %d) not found", ErrInternalTool, toolName, childID, i)
				}
				childValue, buildErr := buildOutput(childNode)
				if buildErr != nil {
					return nil, buildErr
				}
				arrSlice[i] = childValue
			}
			return arrSlice, nil
		case "string", "number", "boolean", "null":
			return node.Value, nil
		default:
			return nil, fmt.Errorf("%w: %s unknown node type '%s'", ErrInternalTool, toolName, node.Type)
		}
	}

	outputData, err := buildOutput(rootNode)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTreeFormatFailed, err)
	}
	jsonBytes, err := json.MarshalIndent(outputData, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTreeJSONMarshal, err)
	}
	return string(jsonBytes), nil
}

// --- toolTreeRenderText Implementation (Updated Formatting) ---

var toolTreeRenderTextImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeRenderText",
		Description: "Renders the tree structure associated with a handle as an indented text string, matching test format.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
		},
		ReturnType: ArgTypeString,
	},
	Func: toolTreeRenderText,
}

// toolTreeRenderText creates an indented text representation of the tree.
func toolTreeRenderText(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeRenderText"
	// Assumes validation layer handles arg count and type checking.
	handleID := args[0].(string)

	tree, err := getTreeFromHandle(interpreter, handleID, toolName) // Use helper
	if err != nil {
		return nil, err
	}

	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists {
		return nil, fmt.Errorf("%w: %s cannot find root node ID '%s' in tree handle '%s'", ErrInternalTool, toolName, tree.RootID, handleID)
	}

	var builder strings.Builder

	// Recursive helper function
	var renderNodeRec func(node *GenericTreeNode, indentLevel int) error
	renderNodeRec = func(node *GenericTreeNode, indentLevel int) error {
		if node == nil {
			return fmt.Errorf("%w: %s renderNodeRec called with nil node", ErrInternalTool, toolName)
		}

		indent := strings.Repeat(defaultIndent, indentLevel)

		// --- Build the main node line ---
		builder.WriteString(fmt.Sprintf("%s- (%s)", indent, node.Type)) // Node type

		// Add count for objects/arrays
		if node.Type == "object" {
			builder.WriteString(fmt.Sprintf(" (attrs: %d)", len(node.Attributes)))
		} else if node.Type == "array" {
			builder.WriteString(fmt.Sprintf(" (len: %d)", len(node.ChildIDs)))
		}

		// Add value for simple types
		switch node.Type {
		case "string":
			builder.WriteString(fmt.Sprintf(": %q", node.Value))
		case "number", "boolean":
			builder.WriteString(fmt.Sprintf(": %v", node.Value))
		case "null":
			builder.WriteString(": null")
		}
		builder.WriteString("\n") // End of the main node line

		// --- Recurse for complex types ---
		if node.Type == "object" {
			keys := make([]string, 0, len(node.Attributes))
			for k := range node.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys) // Sort keys for deterministic test output

			keyIndent := strings.Repeat(defaultIndent, indentLevel+1) // Indent for keys

			for _, key := range keys {
				childID := node.Attributes[key]
				childNode, exists := tree.NodeMap[childID]

				// Write the "Key:" line
				builder.WriteString(fmt.Sprintf("%s* Key: %q\n", keyIndent, key))

				if !exists {
					// Indicate missing node clearly, indented under the key
					builder.WriteString(fmt.Sprintf("%s<ERROR: missing node '%s'>\n", strings.Repeat(defaultIndent, indentLevel+2), childID))
					continue
				}
				// Render the child node, further indented under the key
				if err := renderNodeRec(childNode, indentLevel+2); err != nil {
					return err // Propagate error
				}
			}
		} else if node.Type == "array" {
			itemIndent := strings.Repeat(defaultIndent, indentLevel+1) // Indent for array items
			for i, childID := range node.ChildIDs {
				childNode, exists := tree.NodeMap[childID]
				if !exists {
					// Indicate missing node clearly, indented as an item
					builder.WriteString(fmt.Sprintf("%s- <ERROR: missing node '%s' at index %d>\n", itemIndent, childID, i))
					continue
				}
				// Render the child node, indented as an array item
				if err := renderNodeRec(childNode, indentLevel+1); err != nil {
					return err // Propagate error
				}
			}
		}
		return nil // Success for this node
	}

	// Start rendering from the root
	err = renderNodeRec(rootNode, 0)
	if err != nil {
		interpreter.Logger().Error("Error during TreeRenderText execution", "error", err)
		return nil, fmt.Errorf("%w: %s failed during rendering: %w", ErrInternalTool, toolName, err)
	}

	return builder.String(), nil
}

// --- Registration Function (registerTreeRenderTools) ---
// (No changes needed in the registration function itself)
func registerTreeRenderTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("registerTreeRenderTools called with nil registry")
	}
	toolsToRegister := []ToolImplementation{toolTreeFormatJSONImpl, toolTreeRenderTextImpl}
	var registrationErrors []error
	for _, tool := range toolsToRegister {
		if err := registry.RegisterTool(tool); err != nil {
			fmt.Printf("! Error registering tree render tool %s: %v\n", tool.Spec.Name, err)
			registrationErrors = append(registrationErrors, fmt.Errorf("failed register tree render tool %q: %w", tool.Spec.Name, err))
		}
	}
	if len(registrationErrors) > 0 {
		return errors.Join(registrationErrors...)
	}
	return nil
}
