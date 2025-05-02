// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 15:08:32 PDT // Fix: Use ConvertToInt64E helper
// filename: pkg/core/tools_tree_find.go

// Package core contains core interpreter functionality, including built-in tools.
package core

import (
	"fmt"
	"math"
	"reflect"
)

var toolTreeFindNodesImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "TreeFindNodes",
		Description: "Finds nodes within a tree structure based on query criteria. Returns a list of matching node IDs.",
		Args: []ArgSpec{
			{Name: "tree_handle", Type: ArgTypeString, Required: true, Description: "Handle for the tree structure."},
			{Name: "query", Type: ArgTypeMap, Required: true, Description: `Map defining search criteria. Must contain 'type' (string). Can optionally contain 'value' (any) for exact value match. Example: {"type": "string", "value": "hello"}`},
			{Name: "start_node_id", Type: ArgTypeString, Required: false, Description: "Node ID to start searching from (defaults to tree root)."},
			{Name: "max_depth", Type: ArgTypeInt, Required: false, Description: "Maximum depth to search relative to start node (0 means only start node, <= 0 or omitted means infinite)."},
			{Name: "max_results", Type: ArgTypeInt, Required: false, Description: "Maximum number of matching node IDs to return (> 0, omitted means unlimited)."}, // Updated description
		},
		ReturnType: ArgTypeSliceString, // Returns []string containing node IDs
	},
	Func: toolTreeFindNodes,
}

// toolTreeFindNodes finds nodes matching criteria.
func toolTreeFindNodes(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	toolName := "TreeFindNodes"

	// --- Argument Parsing ---
	if len(args) < 2 || len(args) > 5 {
		return nil, fmt.Errorf("%w: %s expected 2 to 5 arguments, got %d", ErrValidationArgCount, toolName, len(args))
	}

	handleID, ok := args[0].(string)
	if !ok || handleID == "" {
		return nil, fmt.Errorf("%w: %s requires a non-empty string 'tree_handle' as the first argument", ErrValidationTypeMismatch, toolName)
	}

	queryMapInterface, ok := args[1].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: %s requires a map 'query' as the second argument", ErrValidationTypeMismatch, toolName)
	}

	// Optional arguments with defaults
	startNodeID := ""
	maxDepth64 := int64(math.MaxInt32)   // Use int64 internally for consistency with helper
	maxResults64 := int64(math.MaxInt32) // Use int64 internally

	if len(args) > 2 && args[2] != nil {
		startNodeID, ok = args[2].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s optional 'start_node_id' must be a string or nil", ErrValidationTypeMismatch, toolName)
		}
	}
	if len(args) > 3 && args[3] != nil {
		// *** FIXED: Use ConvertToInt64E and check error ***
		depthVal, err := ConvertToInt64E(args[3])
		if err != nil {
			// Wrap error for better context
			return nil, fmt.Errorf("%w for 'max_depth': %w", ErrInvalidArgument, err)
		}
		// Allow 0 depth (means only check start node)
		if depthVal >= 0 {
			maxDepth64 = depthVal
		}
		// If negative, keep default (effectively infinite)
	}
	if len(args) > 4 && args[4] != nil {
		// *** FIXED: Use ConvertToInt64E and check error ***
		resultsVal, err := ConvertToInt64E(args[4])
		if err != nil {
			// Wrap error for better context
			return nil, fmt.Errorf("%w for 'max_results': %w", ErrInvalidArgument, err)
		}
		if resultsVal > 0 {
			maxResults64 = resultsVal
		} else {
			// Treat 0 or negative max_results as invalid input
			return nil, fmt.Errorf("%w: %s optional 'max_results' must be a positive integer", ErrInvalidArgument, toolName)
		}
	}

	// Convert validated int64 limits to int for internal use (safe after checks)
	maxDepth := int(maxDepth64)
	maxResults := int(maxResults64)

	// --- Query Map Validation ---
	queryType, typeExists := queryMapInterface["type"].(string)
	if !typeExists || queryType == "" {
		return nil, fmt.Errorf("%w: %s query map must contain a non-empty string value for the 'type' key", ErrTreeInvalidQuery, toolName)
	}

	queryValue, hasQueryValue := queryMapInterface["value"]
	// TODO: Add support for querying by attributes later if needed

	// --- Get Tree and Start Node ---
	tree, err := getTreeFromHandle(interpreter, handleID, toolName)
	if err != nil {
		return nil, err
	}

	var startNode *GenericTreeNode
	if startNodeID == "" || startNodeID == tree.RootID {
		startNodeID = tree.RootID // Ensure startNodeID holds the actual root ID if default
		var exists bool
		startNode, exists = tree.NodeMap[startNodeID]
		if !exists {
			return nil, fmt.Errorf("%w: %s cannot find root node ID '%s' in tree handle '%s'", ErrInternalTool, toolName, startNodeID, handleID)
		}
	} else {
		var nodeErr error
		_, startNode, nodeErr = getNodeFromHandle(interpreter, handleID, startNodeID, toolName)
		if nodeErr != nil {
			return nil, fmt.Errorf("%s: invalid 'start_node_id': %w", toolName, nodeErr)
		}
	}

	// --- Perform Search ---
	results := make([]string, 0)
	// Define recursive helper function using a closure
	var findRecursive func(currentNode *GenericTreeNode, currentDepth int)

	findRecursive = func(currentNode *GenericTreeNode, currentDepth int) {
		// Stop conditions: node is nil, depth exceeded, results full
		if currentNode == nil || currentDepth > maxDepth || len(results) >= maxResults {
			return
		}

		// Check match
		match := currentNode.Type == queryType
		if match && hasQueryValue {
			match = reflect.DeepEqual(currentNode.Value, queryValue)
		}

		if match {
			results = append(results, currentNode.ID)
			if len(results) >= maxResults {
				return // Stop early if limit reached
			}
		}

		// Stop descending if current depth is already max_depth
		// (Check after potentially matching the node at maxDepth itself)
		if currentDepth >= maxDepth {
			return
		}

		// Recurse children and attributes
		// Combine child ID collection for slightly cleaner iteration
		childNodeIDs := make([]string, 0, len(currentNode.ChildIDs)+len(currentNode.Attributes))
		childNodeIDs = append(childNodeIDs, currentNode.ChildIDs...)
		for _, attrChildID := range currentNode.Attributes {
			childNodeIDs = append(childNodeIDs, attrChildID)
		}

		for _, childID := range childNodeIDs {
			childNode, exists := tree.NodeMap[childID]
			if exists {
				findRecursive(childNode, currentDepth+1)
				if len(results) >= maxResults {
					return
				} // Check limit after each recursion
			} else {
				interpreter.Logger().Warn("Node ID not found during recursive search", "tool", toolName, "parentId", currentNode.ID, "missingChildId", childID)
			}
		}
	}

	findRecursive(startNode, 0) // Start search from the determined startNode

	// Convert results ([]string) to []interface{} for NeuroScript return
	interfaceResults := make([]interface{}, len(results))
	for i, id := range results {
		interfaceResults[i] = id
	}

	interpreter.Logger().Debug("Tree search completed", "tool", toolName, "queryType", queryType, "hasQueryValue", hasQueryValue, "resultsFound", len(results))
	return interfaceResults, nil
}
