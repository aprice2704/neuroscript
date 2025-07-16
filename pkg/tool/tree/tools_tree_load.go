// NeuroScript Version: 0.6.5
// File version: 3
// Purpose: Corrected JSON loading to deterministically create nodes and properly use the 'type' field from the JSON object.
// filename: pkg/tool/tree/tools_tree_load.go
// nlines: 130
// risk_rating: HIGH

package tree

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// toolTreeLoadJSON parses a JSON string and returns a handle to the generic tree.
func toolTreeLoadJSON(interp tool.Runtime, args []interface{}) (interface{}, error) {
	toolName := "Tree.LoadJSON"

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

	tree := utils.NewGenericTree()

	var buildNode func(parentID string, keyForParentAttribute string, value interface{}) (string, error)
	buildNode = func(parentID string, keyForParentAttribute string, value interface{}) (string, error) {
		nodeType := ""
		vMap, isMap := value.(map[string]interface{})

		// Prioritize the 'type' field from the JSON object itself.
		if isMap {
			if typeVal, ok := vMap["type"].(string); ok {
				nodeType = typeVal
			}
		}

		// Fallback to Go type if no explicit type field was found.
		if nodeType == "" {
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
					fmt.Sprintf("%s: unsupported JSON type encountered: %T", toolName, value),
					lang.ErrInternal,
				)
			}
		}

		node := tree.NewNode(parentID, nodeType)
		if parentNode, parentExists := tree.NodeMap[parentID]; parentExists && parentNode.Type == "object" {
			node.ParentAttributeKey = keyForParentAttribute
		}

		switch v := value.(type) {
		case map[string]interface{}:
			node.Attributes = make(utils.TreeAttrs)

			// Sort keys for deterministic node ID generation.
			keys := make([]string, 0, len(v))
			for k := range v {
				keys = append(keys, k)
			}
			sort.Strings(keys)

			for _, k := range keys {
				// The 'type' field determines the node's type, it doesn't become a child attribute.
				if k == "type" {
					continue
				}
				val := v[k]
				childID, errBuild := buildNode(node.ID, k, val)
				if errBuild != nil {
					return "", errBuild
				}
				node.Attributes[k] = childID
			}
		case []interface{}:
			node.ChildIDs = make([]string, len(v))
			for i, item := range v {
				childID, errBuild := buildNode(node.ID, strconv.Itoa(i), item)
				if errBuild != nil {
					return "", errBuild
				}
				node.ChildIDs[i] = childID
			}
		default:
			node.Value = v
		}

		if parentID == "" {
			tree.RootID = node.ID
		}
		return node.ID, nil
	}

	_, err = buildNode("", "", data)
	if err != nil {
		var rtErr *lang.RuntimeError
		if errors.As(err, &rtErr) {
			return nil, rtErr
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: failed to build tree from JSON: %v", toolName, err),
			lang.ErrInternal,
		)
	}

	if tree.RootID == "" && data != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: failed to determine root node", toolName),
			lang.ErrInternal,
		)
	}

	handleID, handleErr := interp.RegisterHandle(tree, utils.GenericTreeHandleType)
	if handleErr != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal,
			fmt.Sprintf("%s: failed to register tree handle: %v", toolName, handleErr),
			errors.Join(lang.ErrInternal, handleErr),
		)
	}

	interp.GetLogger().Debug(fmt.Sprintf("%s: Successfully parsed JSON into tree", toolName), "rootId", tree.RootID, "nodeCount", len(tree.NodeMap), "handle", handleID)
	return handleID, nil
}
