// NeuroScript Version: 0.3.1
// File version: 2
// Purpose: Corrected compiler errors by adding type assertions when accessing child node IDs from the Attributes map.
// nlines: 130 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_tree_render.go

package core

import (
	"encoding/json"
	"errors" // Required for errors.Is/Join
	"fmt"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolTreeFormatJSON serializes the tree structure associated with a handle back into a formatted JSON string.
// Corresponds to ToolSpec "Tree.ToJSON".
func toolTreeFormatJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.ToJSON" // User-facing tool name

	if len(args) != 1 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 1 argument (tree_handle), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}
	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, lang.NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}

	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, err // getTreeFromHandle returns RuntimeError
	}

	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists {
		return nil, lang.NewRuntimeError(ErrorCodeInternal, // Root node missing in a valid tree is an internal inconsistency
			fmt.Sprintf("%s: cannot find root node ID '%s' in tree handle '%s'", toolName, tree.RootID, handleID),
			ErrInternal, // Or a more specific ErrTreeIntegrity sentinel
		)
	}

	var buildOutput func(node *GenericTreeNode) (interface{}, error)
	buildOutput = func(node *GenericTreeNode) (interface{}, error) {
		if node == nil {
			// This indicates a programming error in the traversal logic.
			return nil, lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("%s: attempted to build output from nil node", toolName), ErrInternal)
		}
		switch node.Type {
		case "object":
			objMap := make(map[string]interface{})
			keys := make([]string, 0, len(node.Attributes))
			for k := range node.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys) // For deterministic output
			for _, key := range keys {
				childIDUntyped := node.Attributes[key]
				childID, ok := childIDUntyped.(string)
				if !ok {
					return nil, lang.NewRuntimeError(ErrorCodeInternal,
						fmt.Sprintf("%s: attribute '%s' has non-string value (%T) in node '%s', cannot serialize object", toolName, key, childIDUntyped, node.ID),
						ErrTreeIntegrity,
					)
				}

				childNode, ok := tree.NodeMap[childID]
				if !ok {
					return nil, lang.NewRuntimeError(ErrorCodeInternal, // Child ID in attributes but not in NodeMap
						fmt.Sprintf("%s: child node ID '%s' (key '%s') not found in tree map", toolName, childID, key),
						ErrInternal, // Or ErrTreeIntegrity
					)
				}
				childValue, buildErr := buildOutput(childNode)
				if buildErr != nil {
					return nil, buildErr // Propagate RuntimeError
				}
				objMap[key] = childValue
			}
			return objMap, nil
		case "array":
			arrSlice := make([]interface{}, len(node.ChildIDs))
			for i, childID := range node.ChildIDs {
				childNode, ok := tree.NodeMap[childID]
				if !ok {
					return nil, lang.NewRuntimeError(ErrorCodeInternal, // Child ID in ChildIDs but not in NodeMap
						fmt.Sprintf("%s: child node ID '%s' (index %d) not found in tree map", toolName, childID, i),
						ErrInternal, // Or ErrTreeIntegrity
					)
				}
				childValue, buildErr := buildOutput(childNode)
				if buildErr != nil {
					return nil, buildErr // Propagate RuntimeError
				}
				arrSlice[i] = childValue
			}
			return arrSlice, nil
		case "string", "number", "boolean", "null":
			return node.Value, nil
		default:
			return nil, lang.NewRuntimeError(ErrorCodeInternal, // Unknown node type implies data corruption or bad node creation
				fmt.Sprintf("%s: unknown node type '%s' encountered during JSON serialization", toolName, node.Type),
				ErrInternal, // Or ErrNodeWrongType with a different connotation
			)
		}
	}

	outputData, err := buildOutput(rootNode)
	if err != nil {
		// If err is already RuntimeError, return it, otherwise wrap
		var rtErr *RuntimeError
		if errors.As(err, &rtErr) {
			return nil, rtErr
		}
		return nil, lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("%s: failed to build data for JSON serialization: %v", toolName, err), ErrInternal)
	}

	jsonBytes, marshalErr := json.MarshalIndent(outputData, "", "  ") // Default indent from original code
	if marshalErr != nil {
		return nil, lang.NewRuntimeError(ErrorCodeInternal, // JSON marshalling is an internal operation failure
			fmt.Sprintf("%s: failed to marshal tree data to JSON: %v", toolName, marshalErr),
			errors.Join(ErrTreeJSONMarshal, marshalErr), // Use specific sentinel
		)
	}
	interpreter.Logger().Debug(fmt.Sprintf("%s: Successfully formatted tree to JSON", toolName), "handle", handleID)
	return string(jsonBytes), nil
}

// toolTreeRenderText creates an indented text representation of the tree.
// Corresponds to ToolSpec "Tree.RenderText".
func toolTreeRenderText(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.RenderText"

	if len(args) != 1 {
		return nil, lang.NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 1 argument (tree_handle), got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
	}
	handleID, okHandle := args[0].(string)
	if !okHandle {
		return nil, lang.NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}

	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, err // getTreeFromHandle returns RuntimeError
	}

	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists {
		return nil, lang.NewRuntimeError(ErrorCodeInternal,
			fmt.Sprintf("%s: cannot find root node ID '%s' in tree handle '%s'", toolName, tree.RootID, handleID),
			ErrInternal,
		)
	}

	var builder strings.Builder
	var renderNodeRec func(node *GenericTreeNode, indentLevel int) error
	renderNodeRec = func(node *GenericTreeNode, indentLevel int) error {
		if node == nil {
			return lang.NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("%s: renderNodeRec called with nil node", toolName), ErrInternal)
		}

		indent := strings.Repeat(defaultIndent, indentLevel)
		builder.WriteString(fmt.Sprintf("%s- (%s)", indent, node.Type))

		if node.Type == "object" {
			builder.WriteString(fmt.Sprintf(" (attrs: %d)", len(node.Attributes)))
		} else if node.Type == "array" {
			builder.WriteString(fmt.Sprintf(" (len: %d)", len(node.ChildIDs)))
		}

		switch node.Type {
		case "string":
			builder.WriteString(fmt.Sprintf(": %q", node.Value))
		case "number", "boolean":
			builder.WriteString(fmt.Sprintf(": %v", node.Value))
		case "null":
			builder.WriteString(": null")
		}
		builder.WriteString("\n")

		if node.Type == "object" {
			keys := make([]string, 0, len(node.Attributes))
			for k := range node.Attributes {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			keyIndent := strings.Repeat(defaultIndent, indentLevel+1)
			for _, key := range keys {
				childIDUntyped := node.Attributes[key]
				childID, ok := childIDUntyped.(string)
				builder.WriteString(fmt.Sprintf("%s* Key: %q\n", keyIndent, key))
				if !ok {
					builder.WriteString(fmt.Sprintf("%s<ERROR: attribute value is not a string node ID, but %T>\n", strings.Repeat(defaultIndent, indentLevel+2), childIDUntyped))
					continue
				}

				childNode, childExists := tree.NodeMap[childID]
				if !childExists {
					builder.WriteString(fmt.Sprintf("%s<ERROR: missing node '%s'>\n", strings.Repeat(defaultIndent, indentLevel+2), childID))
					continue // Log or handle as critical error? For rendering, showing error might be best.
				}
				if errRender := renderNodeRec(childNode, indentLevel+2); errRender != nil {
					return errRender
				}
			}
		} else if node.Type == "array" {
			itemIndent := strings.Repeat(defaultIndent, indentLevel+1)
			for i, childID := range node.ChildIDs {
				childNode, childExists := tree.NodeMap[childID]
				if !childExists {
					builder.WriteString(fmt.Sprintf("%s- <ERROR: missing node '%s' at index %d>\n", itemIndent, childID, i))
					continue
				}
				if errRender := renderNodeRec(childNode, indentLevel+1); errRender != nil {
					return errRender
				}
			}
		}
		return nil
	}

	if err := renderNodeRec(rootNode, 0); err != nil {
		var rtErr *RuntimeError
		if errors.As(err, &rtErr) { // If it's already a RuntimeError, pass it through
			return nil, rtErr
		}
		// Wrap other internal find errors
		return nil, lang.NewRuntimeError(ErrorCodeInternal,
			fmt.Sprintf("%s: failed during text rendering: %v", toolName, err),
			errors.Join(ErrInternal, err), // Or ErrTreeFormatFailed
		)
	}
	interpreter.Logger().Debug(fmt.Sprintf("%s: Successfully rendered tree to text", toolName), "handle", handleID)
	return builder.String(), nil
}
