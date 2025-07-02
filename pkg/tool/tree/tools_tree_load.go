// NeuroScript Version: 0.3.1
// File version: 0.1.2 // Set ParentAttributeKey for nodes created as object attributes.
// nlines: 95 // Approximate
// risk_rating: MEDIUM
// filename: pkg/tool/tree/tools_tree_load.go

package tree

import (
	"encoding/json"
	"errors"	// Required for errors.Is
	"fmt"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// toolTreeLoadJSON parses a JSON string and returns a handle to the generic tree.
func toolTreeLoadJSON(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.LoadJSON"	// User-facing tool name for error messages

	if len(args) != 1 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 1 argument (json_string), got %d", toolName, len(args)),
			lang.ErrArgumentMismatch,
		)
	}

	jsonContent, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType,
			fmt.Sprintf("%s: json_string argument must be a string, got %T", toolName, args[0]),
			lang.ErrInvalidArgument,
		)
	}

	var data interface{}
	err := json.Unmarshal([]byte(jsonContent), &data)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeSyntax,
			fmt.Sprintf("%s: failed to unmarshal JSON input: %v", toolName, err),
			errors.Join(lang.ErrTreeJSONUnmarshal, err),
		)
	}

	tree := utils.NewGenericTree()	// Initializes NodeMap and nextID

	var buildNode func(parentID string, keyForParentAttribute string, value interface{}) (string, error)
	buildNode = func(parentID string, keyForParentAttribute string, value interface{}) (string, error) {
		var node *utils.GenericTreeNode
		nodeType := ""

		// Determine nodeType first, then create node, then set ParentAttributeKey if applicable
		switch value.(type) {
		case map[string]interface{}:
			nodeType = "object"
		case []interface{}:
			nodeType = "array"
		case string:
			nodeType = "string"
		case float64:
			nodeType = "number"
		case bool:
			nodeType = "boolean"
		case nil:
			nodeType = "null"
		default:
			return "", lang.NewRuntimeError(lang.ErrorCodeInternal,
				fmt.Sprintf("%s: unsupported JSON type encountered during tree build: %T", toolName, value),
				lang.ErrInternal,
			)
		}

		node = tree.NewNode(parentID, nodeType)	// NewNode sets ParentID

		// Set ParentAttributeKey if this node is an attribute of an object parent
		if parentNode, parentExists := tree.NodeMap[parentID]; parentExists && parentNode.Type == "object" {
			node.ParentAttributeKey = keyForParentAttribute
		}
		// For array elements, keyForParentAttribute is its index as a string.
		// If parent is an array, ParentAttributeKey will be like "0", "1", etc. This is fine.

		// Now populate based on type
		switch v := value.(type) {
		case map[string]interface{}:
			// node.Type is "object", node is already created and ParentAttributeKey potentially set
			node.Attributes = make(utils.TreeAttrs)
			for k, val := range v {	// k is the attribute key within this new object node
				childID, errBuild := buildNode(node.ID, k, val)	// Pass k as keyForParentAttribute for children of this object
				if errBuild != nil {
					return "", errBuild
				}
				node.Attributes[k] = childID
			}
		case []interface{}:
			// node.Type is "array", node is already created
			node.ChildIDs = make([]string, len(v))
			for i, item := range v {
				// Pass the index as string for keyForParentAttribute, though it's less semantically critical for array elements
				childID, errBuild := buildNode(node.ID, strconv.Itoa(i), item)
				if errBuild != nil {
					return "", errBuild
				}
				node.ChildIDs[i] = childID
			}
		case string:
			node.Value = v
		case float64:
			node.Value = v
		case bool:
			node.Value = v
		case nil:
			node.Value = nil
		}

		if parentID == "" {	// This is the root node of the entire JSON structure
			tree.RootID = node.ID
		}
		return node.ID, nil
	}

	_, err = buildNode("", "", data)	// Root node has no parentID and no keyForParentAttribute from a JSON perspective
	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			return nil, rtErr
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: failed to build tree from parsed JSON: %v", toolName, err),
			lang.ErrInternal,
		)
	}

	if tree.RootID == "" && data != nil {
		interpreter.Logger().Error(fmt.Sprintf("%s: RootID is empty after successful JSON unmarshal and build for non-empty data", toolName), "json_content_snippet", fmt.Sprintf("%.30s...", jsonContent), "parsed_data_type", fmt.Sprintf("%T", data))
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: failed to determine root node after parsing JSON", toolName),
			lang.ErrInternal,
		)
	}

	handleID, handleErr := interpreter.RegisterHandle(tree, utils.GenericTreeHandleType)
	if handleErr != nil {
		interpreter.Logger().Error(fmt.Sprintf("%s: Failed to register GenericTree handle", toolName), "error", handleErr)
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: failed to register tree handle: %v", toolName, handleErr),
			errors.Join(lang.ErrInternal, handleErr),
		)
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Successfully parsed JSON into tree", toolName), "rootId", tree.RootID, "nodeCount", len(tree.NodeMap), "handle", handleID)
	return handleID, nil
}