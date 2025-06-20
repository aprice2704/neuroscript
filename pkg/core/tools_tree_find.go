// NeuroScript Version: 0.3.1
// File version: 8
// Purpose: Corrected compiler errors by adding type assertions and replacing the flawed `compareAttributeValue` with a robust `deepCompareValues` function.
// filename: pkg/core/tools_tree_find.go
// nlines: 247
// risk_rating: MEDIUM

package core

import (
	"errors"
	"fmt"
	"reflect"
)

// toolTreeFindNodes implements the Tree.FindNodes tool.
func toolTreeFindNodes(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.FindNodes"

	if len(args) < 3 || len(args) > 5 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: expected 3 to 5 arguments, got %d", toolName, len(args)), ErrArgumentMismatch)
	}

	treeHandle, okHandle := args[0].(string)
	if !okHandle {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: tree_handle argument must be a string, got %T", toolName, args[0]), ErrInvalidArgument)
	}
	startNodeID, okStartNodeID := args[1].(string)
	if !okStartNodeID {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: start_node_id argument must be a string, got %T", toolName, args[1]), ErrInvalidArgument)
	}
	queryMap, okQueryMap := args[2].(map[string]interface{})
	if !okQueryMap {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: query_map argument must be a map, got %T", toolName, args[2]), ErrInvalidArgument)
	}

	maxDepth := -1
	if len(args) > 3 && args[3] != nil {
		depthRaw, okDepth := args[3].(int64)
		if !okDepth {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: max_depth argument must be an integer or null, got %T", toolName, args[3]), ErrInvalidArgument)
		}
		maxDepth = int(depthRaw)
	}

	maxResults := -1
	if len(args) > 4 && args[4] != nil {
		resultsRaw, okResults := args[4].(int64)
		if !okResults {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: max_results argument must be an integer or null, got %T", toolName, args[4]), ErrInvalidArgument)
		}
		maxResults = int(resultsRaw)
	}

	if treeHandle == "" || startNodeID == "" || len(queryMap) == 0 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: tree_handle, start_node_id, and query_map cannot be empty", toolName), ErrInvalidArgument)
	}

	tree, startNode, err := getNodeFromHandle(interpreter, treeHandle, startNodeID, toolName)
	if err != nil {
		return nil, err
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Executing search", toolName),
		"tree_handle", treeHandle, "start_node_id", startNode.ID,
		"query", queryMap, "max_depth", maxDepth, "max_results", maxResults)

	foundNodeIDs := make([]interface{}, 0)
	visited := make(map[string]bool)

	var findRecursive func(currentNode *GenericTreeNode, currentDepth int) error
	findRecursive = func(currentNode *GenericTreeNode, currentDepth int) error {
		if currentNode == nil || visited[currentNode.ID] {
			return nil
		}
		visited[currentNode.ID] = true

		matches, matchErr := nodeMatchesQuery(currentNode, queryMap, toolName)
		if matchErr != nil {
			return matchErr
		}

		if matches {
			foundNodeIDs = append(foundNodeIDs, currentNode.ID)
			if maxResults != -1 && len(foundNodeIDs) >= maxResults {
				return errors.New("max results reached")
			}
		}

		if maxDepth != -1 && currentDepth >= maxDepth {
			return nil
		}

		// Recurse through children referenced by ChildIDs (arrays)
		if currentNode.ChildIDs != nil {
			for _, childID := range currentNode.ChildIDs {
				childNode, exists := tree.NodeMap[childID]
				if exists {
					if err := findRecursive(childNode, currentDepth+1); err != nil {
						if err.Error() == "max results reached" {
							return err
						}
						return err
					}
				}
			}
		}

		// Recurse through children referenced by Attributes (objects)
		if currentNode.Type == "object" && currentNode.Attributes != nil {
			for _, childNodeIDUntyped := range currentNode.Attributes {
				childNodeID, ok := childNodeIDUntyped.(string)
				if !ok {
					// This attribute's value is not a node ID string.
					// In the context of finding children, we can safely skip it.
					continue
				}
				childNode, exists := tree.NodeMap[childNodeID]
				if exists {
					if err := findRecursive(childNode, currentDepth+1); err != nil {
						if err.Error() == "max results reached" {
							return err
						}
						return err
					}
				}
			}
		}
		return nil
	}

	searchErr := findRecursive(startNode, 0)
	if searchErr != nil && searchErr.Error() != "max results reached" {
		var rtErr *RuntimeError
		if errors.As(searchErr, &rtErr) {
			return nil, rtErr
		}
		return nil, NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("%s: error during node search: %v", toolName, searchErr), ErrInternal)
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Search completed", toolName), "found_count", len(foundNodeIDs))
	return foundNodeIDs, nil
}

// nodeMatchesQuery checks if a single node matches the provided query map.
func nodeMatchesQuery(node *GenericTreeNode, queryMap map[string]interface{}, toolName string) (bool, error) {
	if node == nil {
		return false, nil
	}

	for key, expectedQueryValue := range queryMap {
		switch key {
		case "id":
			idStr, ok := expectedQueryValue.(string)
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'id' in query_map must be a string, got %T", toolName, expectedQueryValue), ErrTreeInvalidQuery)
			}
			if node.ID != idStr {
				return false, nil
			}
		case "type":
			typeStr, ok := expectedQueryValue.(string)
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'type' in query_map must be a string, got %T", toolName, expectedQueryValue), ErrTreeInvalidQuery)
			}
			if node.Type != typeStr {
				return false, nil
			}
		case "value":
			if !deepCompareValues(node.Value, expectedQueryValue) {
				return false, nil
			}
		case "attributes":
			attrQueryMap, ok := expectedQueryValue.(map[string]interface{})
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'attributes' value in query_map must be a map, got %T", toolName, expectedQueryValue), ErrTreeInvalidQuery)
			}
			if node.Attributes == nil && len(attrQueryMap) > 0 {
				return false, nil
			}
			for queryAttrKey, queryAttrExpectedValue := range attrQueryMap {
				actualNodeAttrValue, exists := node.Attributes[queryAttrKey]
				if !exists {
					return false, nil
				}
				if !deepCompareValues(actualNodeAttrValue, queryAttrExpectedValue) {
					return false, nil
				}
			}
		case "metadata":
			metadataQuery, ok := expectedQueryValue.(map[string]interface{})
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'metadata' in query_map must be a map, got %T", toolName, expectedQueryValue), ErrTreeInvalidQuery)
			}
			if node.Attributes == nil && len(metadataQuery) > 0 {
				return false, nil
			}
			for metaKey, expectedMetaQueryValue := range metadataQuery {
				actualNodeMetaValue, exists := node.Attributes[metaKey]
				if !exists {
					return false, nil
				}
				if !deepCompareValues(actualNodeMetaValue, expectedMetaQueryValue) {
					return false, nil
				}
			}
		default: // This case handles direct attribute name queries like {"myCustomAttribute": "expectedValue"}
			actualNodeAttrValue, exists := node.Attributes[key] // key is the attribute name
			if !exists {
				return false, nil
			}
			if !deepCompareValues(actualNodeAttrValue, expectedQueryValue) {
				return false, nil
			}
		}
	}
	return true, nil // All conditions in queryMap matched
}

// deepCompareValues compares two interface{} values, with special handling for numeric types.
func deepCompareValues(actualValue, expectedValue interface{}) bool {
	// First, try a simple deep equal. This covers string, bool, nil, and matching number types.
	if reflect.DeepEqual(actualValue, expectedValue) {
		return true
	}

	// Special handling for numbers of different types (e.g., int64 vs float64).
	actualNum, actualIsNum := ConvertToFloat64(actualValue)
	expectedNum, expectedIsNum := ConvertToFloat64(expectedValue)

	if actualIsNum && expectedIsNum && actualNum == expectedNum {
		return true
	}

	return false
}
