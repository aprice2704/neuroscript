// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Removed local ToolImplementation, standardized error handling.
// nlines: 90 // Approximate
// risk_rating: LOW
// filename: pkg/core/tools_tree_load.go

package core

import (
	"encoding/json"
	"errors" // Required for errors.Is
	"fmt"
	"strconv"
)

// toolTreeLoadJSON parses a JSON string and returns a handle to the generic tree.
// Corresponds to ToolSpec "Tree.LoadJSON".
func toolTreeLoadJSON(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.LoadJSON" // User-facing tool name for error messages

	// Argument validation is expected to be handled by the validation layer
	// based on ToolSpec. However, a direct call or internal use might bypass it.
	if len(args) != 1 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 1 argument (json_string), got %d", toolName, len(args)),
			ErrArgumentMismatch, // Use the more general ErrArgumentMismatch
		)
	}

	jsonContent, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType,
			fmt.Sprintf("%s: json_string argument must be a string, got %T", toolName, args[0]),
			ErrInvalidArgument, // Use ErrInvalidArgument for type issues post-validation if it slips through
		)
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonContent), &data)
	if err != nil {
		// For JSON parsing errors, ErrorCodeSyntax is appropriate, wrapping ErrTreeJSONUnmarshal
		return nil, NewRuntimeError(ErrorCodeSyntax,
			fmt.Sprintf("%s: failed to unmarshal JSON input: %v", toolName, err),
			errors.Join(ErrTreeJSONUnmarshal, err), // Ensure ErrTreeJSONUnmarshal is in the chain for specific checks
		)
	}

	// Use the NewGenericTree constructor
	tree := NewGenericTree() // Initializes NodeMap and nextID

	var buildNode func(parentID string, key string, value interface{}) (string, error)
	buildNode = func(parentID string, key string, value interface{}) (string, error) {
		var node *GenericTreeNode // Node will be created by tree.NewNode
		nodeType := ""

		switch v := value.(type) {
		case map[string]interface{}:
			nodeType = "object"
			node = tree.NewNode(parentID, nodeType) // NewNode adds to tree.NodeMap
			for k, val := range v {
				childID, errBuild := buildNode(node.ID, k, val)
				if errBuild != nil {
					return "", errBuild // Propagate error directly
				}
				// Attributes map string keys to child node IDs for objects
				if node.Attributes == nil { // Should be initialized by NewNode, but defensive
					node.Attributes = make(map[string]string)
				}
				node.Attributes[k] = childID
			}
		case []interface{}:
			nodeType = "array"
			node = tree.NewNode(parentID, nodeType) // NewNode adds to tree.NodeMap
			node.ChildIDs = make([]string, len(v))  // Pre-allocate ChildIDs
			for i, item := range v {
				childID, errBuild := buildNode(node.ID, strconv.Itoa(i), item) // Key for array items is their index as string
				if errBuild != nil {
					return "", errBuild // Propagate error
				}
				node.ChildIDs[i] = childID
			}
		case string:
			nodeType = "string"
			node = tree.NewNode(parentID, nodeType)
			node.Value = v
		case float64: // JSON numbers are float64
			nodeType = "number"
			node = tree.NewNode(parentID, nodeType)
			node.Value = v
		case bool:
			nodeType = "boolean"
			node = tree.NewNode(parentID, nodeType)
			node.Value = v
		case nil:
			nodeType = "null"
			node = tree.NewNode(parentID, nodeType)
			node.Value = nil // Explicitly set for clarity, though default is nil
		default:
			// This indicates an issue with the JSON unmarshaler or an unexpected type
			return "", NewRuntimeError(ErrorCodeInternal, // Or ErrorCodeSyntax if considered a parsing issue
				fmt.Sprintf("%s: unsupported JSON type encountered during tree build: %T", toolName, value),
				ErrInternal, // Or a more specific tree build error
			)
		}

		// If this is the first node being built (no parentID), it's the root.
		if parentID == "" {
			tree.RootID = node.ID
		}
		return node.ID, nil
	}

	_, err = buildNode("", "", data) // Initial call for the root of the JSON data
	if err != nil {
		// If buildNode returned a RuntimeError, pass it, else wrap it.
		var rtErr *RuntimeError
		if errors.As(err, &rtErr) {
			return nil, rtErr
		}
		return nil, NewRuntimeError(ErrorCodeInternal, // Indicates failure in the tree construction logic
			fmt.Sprintf("%s: failed to build tree from parsed JSON: %v", toolName, err),
			ErrInternal, // Or a specific ErrTreeBuildFailed sentinel if defined and appropriate
		)
	}

	if tree.RootID == "" && jsonContent != "null" && jsonContent != `""` && jsonContent != "[]" && jsonContent != "{}" {
		// This case handles if JSON was valid (e.g. "null") but resulted in no root.
		// Tree.NewNode always creates an ID, so if RootID is empty after buildNode, it's an issue.
		// However, for valid empty structures like `[]` or `{}`, or `null`, RootID will be set.
		// This check is more for an unexpected state.
		// An empty JSON string `""` would fail unmarshal earlier.
		// A simple JSON value like `"hello"` or `123` will also have RootID set.
		// The only problematic case could be an empty input that somehow passes unmarshal but not build.
		// If jsonContent is "null", tree.RootID will be "node-1" of type "null".
		// Check if tree.RootID is empty ONLY if it's not a case that naturally results in one node.
		// A single "null" JSON value will result in a root node. An empty string `""` jsonContent errors earlier.
		if data != nil { // if data was unmarshalled, a root should have been made
			interpreter.Logger().Error(fmt.Sprintf("%s: RootID is empty after successful JSON unmarshal and build for non-empty data", toolName), "json_content", jsonContent, "parsed_data_type", fmt.Sprintf("%T", data))
			return nil, NewRuntimeError(ErrorCodeInternal,
				fmt.Sprintf("%s: failed to determine root node after parsing JSON", toolName),
				ErrInternal, // Or a specific ErrTreeBuildFailed
			)
		}
	}

	handleID, handleErr := interpreter.RegisterHandle(tree, GenericTreeHandleType)
	if handleErr != nil {
		interpreter.Logger().Error(fmt.Sprintf("%s: Failed to register GenericTree handle", toolName), "error", handleErr)
		return nil, NewRuntimeError(ErrorCodeInternal, // Handle registration is an internal system concern
			fmt.Sprintf("%s: failed to register tree handle: %v", toolName, handleErr),
			errors.Join(ErrInternal, handleErr), // Join to keep original context
		)
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Successfully parsed JSON into tree", toolName), "rootId", tree.RootID, "nodeCount", len(tree.NodeMap), "handle", handleID)
	return handleID, nil
}
