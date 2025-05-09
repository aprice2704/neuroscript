// NeuroScript Version: 0.3.1
// File version: 0.1.7 // Add specific handling for "attributes" key in nodeMatchesQuery.
// nlines: 245 // Approximate
// risk_rating: HIGH
// filename: pkg/core/tools_tree_find.go

package core

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
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

		matches, matchErr := nodeMatchesQuery(currentNode, queryMap, tree, toolName) // Pass tree for potential lookups
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
			for _, childNodeID := range currentNode.Attributes {
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
// It now takes the tree to allow dereferencing attribute node IDs if necessary for complex attribute queries.
func nodeMatchesQuery(node *GenericTreeNode, queryMap map[string]interface{}, tree *GenericTree, toolName string) (bool, error) {
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
			if !reflect.DeepEqual(node.Value, expectedQueryValue) {
				// Special handling for numbers that might be int64 in query but float64 in node or vice-versa
				nodeNum, nodeIsNum := ConvertToFloat64(node.Value)
				queryNum, queryIsNum := ConvertToFloat64(expectedQueryValue)
				if !(nodeIsNum && queryIsNum && nodeNum == queryNum) {
					return false, nil
				}
			}
		case "attributes": // Handles queries like {"attributes": {"attrName": "expectedNodeIDValue"}}
			attrQueryMap, ok := expectedQueryValue.(map[string]interface{})
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'attributes' value in query_map must be a map, got %T", toolName, expectedQueryValue), ErrTreeInvalidQuery)
			}

			if node.Attributes == nil && len(attrQueryMap) > 0 {
				return false, nil
			} // Node has no attributes to match against

			for queryAttrKey, queryAttrExpectedValue := range attrQueryMap {
				actualNodeAttrValue, exists := node.Attributes[queryAttrKey] // actualNodeAttrValue is a string (node ID)
				if !exists {
					return false, nil
				} // The queried attribute key does not exist on the node

				// The value of an attribute in node.Attributes is the ID of another node.
				// queryAttrExpectedValue is the expected ID string for that attribute's target node.
				if !compareAttributeValue(actualNodeAttrValue, queryAttrExpectedValue) {
					// This compares if actualNodeAttrValue (string) matches queryAttrExpectedValue (interface{}, likely string)
					return false, nil
				}
			}
			// If we loop through all queryAttrKey and all match, this "attributes" part of the query is satisfied.
		case "metadata": // Handles queries like {"metadata": {"metaKey": "metaValueString"}}
			// Assuming metadata is stored in node.Attributes and values are strings (potentially IDs)
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
				if !compareAttributeValue(actualNodeMetaValue, expectedMetaQueryValue) {
					return false, nil
				}
			}
		default: // This case handles direct attribute name queries like {"myCustomAttribute": "expectedNodeID"}
			actualNodeAttrValue, exists := node.Attributes[key] // key is the attribute name
			if !exists {
				return false, nil
			}
			if !compareAttributeValue(actualNodeAttrValue, expectedQueryValue) {
				return false, nil
			}
		}
	}
	return true, nil // All conditions in queryMap matched
}

// compareAttributeValue compares an actual string value from node.Attributes
// with an expected value (interface{}) from the query.
func compareAttributeValue(actualStringValue string, expectedQueryValue interface{}) bool {
	switch expected := expectedQueryValue.(type) {
	case string:
		return actualStringValue == expected
	case float64:
		parsedActual, err := strconv.ParseFloat(actualStringValue, 64)
		return err == nil && parsedActual == expected
	case int64:
		parsedActual, err := strconv.ParseInt(actualStringValue, 10, 64)
		return err == nil && parsedActual == expected
	case int:
		parsedActual, err := strconv.ParseInt(actualStringValue, 10, 64)
		return err == nil && parsedActual == int64(expected)
	case bool:
		parsedActual, err := strconv.ParseBool(actualStringValue)
		return err == nil && parsedActual == expected
	default:
		// Fallback for other types might be too broad or error-prone.
		// Consider if specific comparisons are needed for other expected types.
		// For now, strict direct comparison or recognized types.
		return false // If expectedQueryValue is not one of the handled types, assume mismatch
	}
}

// ConvertToFloat64 is a helper to handle potential int64/float64 from map[string]interface{}
func ConvertToFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case int64:
		return float64(v), true
	case int:
		return float64(v), true
	// Potentially add string conversion if numbers can be stored as strings in node.Value
	case string:
		fVal, err := strconv.ParseFloat(v, 64)
		if err == nil {
			return fVal, true
		}
	}
	return 0, false
}
