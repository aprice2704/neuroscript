// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Align with GenericTree, implement basic find, standardize errors.
// nlines: 200 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tools_tree_find.go

package core

import (
	"errors"
	"fmt"
	"reflect" // For deep equality check of values if needed
)

// toolTreeFindNodes implements the Tree.FindNodes tool.
// It searches for nodes within a GenericTree starting from a specific node,
// based on criteria provided in a query map.
func toolTreeFindNodes(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Tree.FindNodes"

	// Args: tree_handle, start_node_id, query_map, [max_depth], [max_results]
	if len(args) < 3 || len(args) > 5 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch,
			fmt.Sprintf("%s: expected 3 to 5 arguments, got %d", toolName, len(args)),
			ErrArgumentMismatch,
		)
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

	maxDepth := -1 // Default: unlimited depth
	if len(args) > 3 && args[3] != nil {
		depthRaw, okDepth := args[3].(int64)
		if !okDepth {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: max_depth argument must be an integer or null, got %T", toolName, args[3]), ErrInvalidArgument)
		}
		maxDepth = int(depthRaw)
	}

	maxResults := -1 // Default: unlimited results
	if len(args) > 4 && args[4] != nil {
		resultsRaw, okResults := args[4].(int64)
		if !okResults {
			return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("%s: max_results argument must be an integer or null, got %T", toolName, args[4]), ErrInvalidArgument)
		}
		maxResults = int(resultsRaw)
	}

	if treeHandle == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: tree_handle cannot be empty", toolName), ErrInvalidArgument)
	}
	if startNodeID == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: start_node_id cannot be empty", toolName), ErrInvalidArgument)
	}
	if len(queryMap) == 0 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: query_map cannot be empty", toolName), ErrTreeInvalidQuery)
	}

	// Get the tree and the starting node using GenericTreeHandleType
	tree, startNode, err := getNodeFromHandle(interpreter, treeHandle, startNodeID, toolName)
	if err != nil {
		return nil, err // getNodeFromHandle returns RuntimeError
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Executing search", toolName),
		"tree_handle", treeHandle, "start_node_id", startNode.ID,
		"query", queryMap, "max_depth", maxDepth, "max_results", maxResults)

	foundNodeIDs := make([]interface{}, 0)
	visited := make(map[string]bool) // To handle potential cycles and avoid reprocessing

	var findRecursive func(currentNode *GenericTreeNode, currentDepth int) error
	findRecursive = func(currentNode *GenericTreeNode, currentDepth int) error {
		if currentNode == nil || visited[currentNode.ID] {
			return nil
		}
		visited[currentNode.ID] = true

		// Check if current node matches query
		matches, matchErr := nodeMatchesQuery(currentNode, queryMap, toolName)
		if matchErr != nil {
			return matchErr // Propagate RuntimeError from matcher
		}
		if matches {
			foundNodeIDs = append(foundNodeIDs, currentNode.ID)
			if maxResults != -1 && len(foundNodeIDs) >= maxResults {
				return errors.New("max results reached") // Signal to stop search
			}
		}

		// Stop if max depth reached for children
		if maxDepth != -1 && currentDepth >= maxDepth {
			return nil
		}

		// Recurse through children (attributes for objects, ChildIDs for arrays/general)
		// For objects, children are linked via attributes
		if currentNode.Type == "object" && currentNode.Attributes != nil {
			for _, childID := range currentNode.Attributes {
				childNode, exists := tree.NodeMap[childID]
				if exists {
					if err := findRecursive(childNode, currentDepth+1); err != nil {
						if err.Error() == "max results reached" {
							return err
						}
						return err // Propagate other errors
					}
				}
			}
		}
		// For arrays (and potentially other types that might use ChildIDs generally)
		if currentNode.ChildIDs != nil {
			for _, childID := range currentNode.ChildIDs {
				childNode, exists := tree.NodeMap[childID]
				if exists {
					if err := findRecursive(childNode, currentDepth+1); err != nil {
						if err.Error() == "max results reached" {
							return err
						}
						return err // Propagate other errors
					}
				}
			}
		}
		return nil
	}

	searchErr := findRecursive(startNode, 0)
	if searchErr != nil && searchErr.Error() != "max results reached" {
		// If it's already a RuntimeError, pass it, else wrap it.
		var rtErr *RuntimeError
		if errors.As(searchErr, &rtErr) {
			return nil, rtErr
		}
		return nil, NewRuntimeError(ErrorCodeInternal, // Catch-all for unexpected search errors
			fmt.Sprintf("%s: error during node search: %v", toolName, searchErr),
			ErrInternal,
		)
	}

	interpreter.Logger().Debug(fmt.Sprintf("%s: Search completed", toolName), "found_count", len(foundNodeIDs))
	return foundNodeIDs, nil
}

// nodeMatchesQuery checks if a single node matches the provided query map.
func nodeMatchesQuery(node *GenericTreeNode, queryMap map[string]interface{}, toolName string) (bool, error) {
	if node == nil {
		return false, nil // Or an error if a nil node here is unexpected
	}

	for key, expectedValue := range queryMap {
		switch key {
		case "type":
			typeStr, ok := expectedValue.(string)
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'type' in query_map must be a string, got %T", toolName, expectedValue), ErrTreeInvalidQuery)
			}
			if node.Type != typeStr {
				return false, nil
			}
		case "value":
			// Using reflect.DeepEqual for value comparison allows matching complex values if stored,
			// though tree nodes primarily store simple values directly or references.
			// For simple scalar values, direct comparison is fine.
			// JSON numbers are float64, so ensure comparison handles that.
			if node.Type == "number" {
				nodeValFloat, nodeOk := node.Value.(float64)
				queryValFloat, queryOk := ConvertToFloat64(expectedValue) // Helper to convert int64/float64
				if !nodeOk || !queryOk || nodeValFloat != queryValFloat {
					return false, nil
				}
			} else if !reflect.DeepEqual(node.Value, expectedValue) {
				return false, nil
			}
		case "metadata":
			metadataQuery, ok := expectedValue.(map[string]interface{})
			if !ok {
				return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: 'metadata' in query_map must be a map, got %T", toolName, expectedValue), ErrTreeInvalidQuery)
			}
			if node.Attributes == nil && len(metadataQuery) > 0 { // Node has no attributes but query expects some
				return false, nil
			}
			for metaKey, metaExpectedValue := range metadataQuery {
				actualMetaValue, exists := node.Attributes[metaKey]
				if !exists {
					return false, nil
				}
				// Assuming metadata values in node.Attributes are strings, as per SetNodeMetadata.
				// Comparison needs to handle if metaExpectedValue is not a string.
				metaExpectedStr, isStr := metaExpectedValue.(string)
				if !isStr { // Query for metadata expects string value
					if metaExpectedValue == nil && actualMetaValue == "" { // Allow matching empty string with nil query, perhaps?
						// This behavior might need refinement based on desired nil/empty string semantics for metadata.
						// For now, strict string match or explicit nil match.
					} else if actualMetaValue != fmt.Sprintf("%v", metaExpectedValue) { // Fallback to string comparison
						return false, nil
					}
				} else if actualMetaValue != metaExpectedStr {
					return false, nil
				}
			}
		default:
			return false, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("%s: unknown key '%s' in query_map", toolName, key), ErrTreeInvalidQuery)
		}
	}
	return true, nil // All query conditions matched
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
	default:
		return 0, false
	}
}
