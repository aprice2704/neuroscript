// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 22:19:01 PDT // Fix indentation logic AGAIN in TreeRenderText
// filename: pkg/core/tools_tree_render.go

package core

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// toolTreeFormatJSON remains unchanged
// ... (code omitted for brevity) ...
func toolTreeFormatJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeFormatJSON"
	handleID := args[0].(string)
	obj, err := interpreter.GetHandleValue(handleID, GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", toolName, err)
	}
	tree, ok := obj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle '%s' contains invalid data type (%T)", ErrHandleInvalid, toolName, handleID, obj)
	}
	var buildInterface func(nodeID string) (interface{}, error)
	buildInterface = func(nodeID string) (interface{}, error) {
		node, exists := tree.NodeMap[nodeID]
		if !exists {
			return nil, fmt.Errorf("%w: node ID '%s' not found", ErrTreeFormatFailed, nodeID)
		}
		switch node.Type {
		case "object":
			objMap := make(map[string]interface{})
			for key, childID := range node.Attributes {
				childValue, err := buildInterface(childID)
				if err != nil {
					return nil, err
				}
				objMap[key] = childValue
			}
			return objMap, nil
		case "array":
			arrSlice := make([]interface{}, len(node.ChildIDs))
			for i, childID := range node.ChildIDs {
				childValue, err := buildInterface(childID)
				if err != nil {
					return nil, err
				}
				arrSlice[i] = childValue
			}
			return arrSlice, nil
		case "string", "number", "boolean", "null":
			return node.Value, nil
		default:
			return nil, fmt.Errorf("%w: unknown node type '%s'", ErrTreeFormatFailed, node.Type)
		}
	}
	rootValue, err := buildInterface(tree.RootID)
	if err != nil {
		return nil, err
	}
	jsonBytes, err := json.MarshalIndent(rootValue, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrTreeJSONMarshal, err)
	}
	return string(jsonBytes), nil
}

// toolTreeRenderText creates an indented text representation of the tree.
func toolTreeRenderText(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeRenderText"
	handleID := args[0].(string)
	obj, err := interpreter.GetHandleValue(handleID, GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", toolName, err)
	}
	tree, ok := obj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle '%s' contains invalid data type (%T)", ErrHandleInvalid, toolName, handleID, obj)
	}

	var builder strings.Builder
	var renderNode func(nodeID string, indentLevel int) error
	renderNode = func(nodeID string, indentLevel int) error {
		node, exists := tree.NodeMap[nodeID]
		if !exists {
			return fmt.Errorf("%w: node ID '%s' not found", ErrInternalTool, nodeID)
		}

		indent := strings.Repeat(defaultIndent, indentLevel)
		builder.WriteString(fmt.Sprintf("%s- [%s] (%s)", indent, node.ID, node.Type))

		switch node.Type {
		case "string":
			builder.WriteString(fmt.Sprintf(": %q", node.Value))
		case "number", "boolean":
			builder.WriteString(fmt.Sprintf(": %v", node.Value))
		case "null":
			builder.WriteString(": null")
		case "object":
			builder.WriteString(fmt.Sprintf(" (attrs: %d)", len(node.Attributes)))
		case "array":
			builder.WriteString(fmt.Sprintf(" (len: %d)", len(node.ChildIDs)))
		}
		builder.WriteString("\n")

		var childErr error
		if node.Type == "object" {
			keys := make([]string, 0, len(node.Attributes))
			for k := range node.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, key := range keys {
				childID := node.Attributes[key]
				keyIndent := strings.Repeat(defaultIndent, indentLevel+1) // Key indented one level
				builder.WriteString(fmt.Sprintf("%s* Key: %q\n", keyIndent, key))
				// *** FIXED: Render child node TWO levels deeper than the parent object ***
				childErr = renderNode(childID, indentLevel+2)
				if childErr != nil {
					return childErr
				}
			}
		} else if node.Type == "array" {
			for _, childID := range node.ChildIDs {
				// Render child node ONE level deeper than the parent array
				childErr = renderNode(childID, indentLevel+1)
				if childErr != nil {
					return childErr
				}
			}
		}
		return nil
	}

	err = renderNode(tree.RootID, 0)
	if err != nil {
		return nil, err
	}
	return builder.String(), nil
}
