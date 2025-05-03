// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 16:15:00 PM PDT // Remove internal core helpers; Use exported errors
// filename: pkg/neurodata/checklist/checklist_tool2.go
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// toolChecklistFormatTree formats the checklist tree back into a string.
func toolChecklistFormatTree(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistFormatTree"
	logger := interpreter.Logger()
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (handle), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}

	// Get handle value directly
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		// Let GetHandleValue's error handling provide details (NotFound, WrongType, InvalidFormat)
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}

	// Perform checks after getting the object
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil { // Added nil check for tree
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}

	logger.Debug("Formatting checklist tree to string", "tool", toolName, "handle", handleID)
	formattedString, formatErr := TreeToChecklistString(tree)
	if formatErr != nil {
		logger.Error("Error formatting checklist tree", "tool", toolName, "handle", handleID, "error", formatErr)
		// Map internal formatting errors to InvalidArgument for user tools
		if errors.Is(formatErr, ErrInvalidChecklistTree) || errors.Is(formatErr, ErrMissingStatusAttribute) || errors.Is(formatErr, ErrUnknownStatus) || errors.Is(formatErr, ErrMissingSpecialSymbol) {
			return nil, fmt.Errorf("%w: %s formatting failed: %w", core.ErrInvalidArgument, toolName, formatErr)
		}
		// Treat other errors as internal (use exported core.ErrInternal)
		return nil, fmt.Errorf("%w: %s failed formatting tree: %w", core.ErrInternal, toolName, formatErr)
	}
	logger.Debug("Successfully formatted checklist tree", "tool", toolName, "handle", handleID)
	return formattedString, nil
}

// toolChecklistSetItemText updates the text value of a checklist item node.
func toolChecklistSetItemText(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemText"
	logger := interpreter.Logger()
	if len(args) != 3 {
		return nil, fmt.Errorf("%w: %s expected 3 arguments (handle, nodeId, newText), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}
	nodeID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[1] 'nodeId', got %T", core.ErrValidationTypeMismatch, toolName, args[1])
	}
	if nodeID == "" {
		return nil, fmt.Errorf("%w: %s requires non-empty 'nodeId'", core.ErrValidationRequiredArgNil, toolName)
	}
	newText, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newText', got %T", core.ErrValidationTypeMismatch, toolName, args[2])
	}

	// Get handle value directly
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err) // Let GetHandleValue handle specifics
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}

	// Find the node within the tree
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", core.ErrNotFound, toolName, nodeID, handleID)
	}

	// Validate node type
	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node ID %q is type %q, expected 'checklist_item'", core.ErrInvalidArgument, toolName, nodeID, targetNode.Type)
	}

	// Update the value
	oldText := ""
	if targetNode.Value != nil {
		oldText = fmt.Sprintf("%v", targetNode.Value)
	}
	targetNode.Value = newText
	logger.Debug("Node text updated", "tool", toolName, "nodeId", nodeID, "oldText", oldText, "newText", newText)
	return nil, nil // Success returns nil
}

// toolChecklistUpdateStatus triggers the recursive status update for the entire checklist tree.
func toolChecklistUpdateStatus(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "Checklist.UpdateStatus"
	logger := interpreter.Logger()
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (handle), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}

	// Get handle value directly
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err) // Let GetHandleValue handle specifics
	}

	// Perform checks after getting the object
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil || tree.RootID == "" {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}
	rootNode, rootExists := tree.NodeMap[tree.RootID]
	if !rootExists || rootNode == nil || rootNode.Type != "checklist_root" {
		// Use exported core.ErrInternal as this indicates corrupted data if the handle was valid
		return nil, fmt.Errorf("%w: %s handle %q has invalid root node structure", core.ErrInternal, toolName, handleID)
	}

	logger.Debug("Starting checklist status update", "tool", toolName, "handle", handleID, "rootId", tree.RootID)
	err = updateChecklistTreeStatus(tree, logger)
	if err != nil {
		logger.Error("Checklist status update failed", "tool", toolName, "handle", handleID, "error", err)
		// Map internal update errors to InvalidArgument for user tools if appropriate
		if errors.Is(err, ErrInvalidChecklistTree) || errors.Is(err, ErrMissingStatusAttribute) || errors.Is(err, ErrInternalParser) {
			return nil, fmt.Errorf("%w: %s update failed due to invalid tree structure or data: %w", core.ErrInvalidArgument, toolName, err)
		}
		// Treat other errors as internal (use exported core.ErrInternal)
		return nil, fmt.Errorf("%w: %s update failed: %w", core.ErrInternal, toolName, err)
	}
	logger.Debug("Checklist status update completed successfully", "tool", toolName, "handle", handleID)
	return nil, nil // Success returns nil
}

// updateChecklistTreeStatus initiates the recursive update from the root's children.
func updateChecklistTreeStatus(tree *core.GenericTree, logger logging.Logger) error {
	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists || rootNode == nil {
		// This indicates an invalid tree state passed into the function
		return fmt.Errorf("%w: root node %q not found in provided tree", ErrInvalidChecklistTree, tree.RootID)
	}
	if rootNode.ChildIDs == nil {
		logger.Debug("Root node has no children, nothing to update.")
		return nil // Nothing to update if root has no children
	}
	logger.Debug("Updating children of root", "rootId", tree.RootID, "childIds", rootNode.ChildIDs)
	// Call the recursive function for each direct child of the root.
	for _, childID := range rootNode.ChildIDs {
		_, err := updateAutomaticNodeStatus(tree, childID, logger)
		if err != nil {
			// Wrap the error to indicate where the failure started
			return fmt.Errorf("error updating status starting from node %q: %w", childID, err)
		}
	}
	return nil
}

// updateAutomaticNodeStatus recursively traverses the tree, calculates and updates the status
// of automatic nodes based on their children's *final* statuses from the recursive calls.
// It now correctly recurses through manual nodes as well.
func updateAutomaticNodeStatus(tree *core.GenericTree, nodeID string, logger logging.Logger) (string, error) {
	node, exists := tree.NodeMap[nodeID]
	if !exists || node == nil {
		logger.Error("Node referenced in tree not found during update", "nodeId", nodeID)
		return "", fmt.Errorf("%w: node %q not found during update", ErrInvalidChecklistTree, nodeID)
	}
	if node.Type != "checklist_item" {
		logger.Error("Encountered non-checklist_item node during status update", "nodeId", nodeID, "type", node.Type)
		return "", fmt.Errorf("%w: node %q has unexpected type %q during status update", ErrInvalidChecklistTree, nodeID, node.Type)
	}
	if node.Attributes == nil {
		node.Attributes = make(map[string]string) // Ensure attributes map exists
	}

	isAutomatic := node.Attributes["is_automatic"] == "true"
	// Get current status *before* recursion, used for comparison and for manual nodes
	currentStatus, statusExists := node.Attributes["status"]
	if !statusExists {
		logger.Warn("Node missing 'status' attribute, defaulting to 'open'", "nodeId", nodeID)
		node.Attributes["status"] = "open" // Fix missing status
		currentStatus = "open"
	}

	logger.Debug("Entering updateAutomaticNodeStatus", "nodeId", nodeID, "isAutomatic", isAutomatic, "currentStatus", currentStatus)

	// --- Step 1: ALWAYS Recurse to children first to get their final calculated statuses ---
	childStatuses := make([]string, 0, len(node.ChildIDs))
	childSymbols := make(map[int]string) // Symbols from children whose FINAL status is 'special'

	if len(node.ChildIDs) > 0 {
		logger.Debug("Processing children", "parentNodeId", nodeID, "childIds", node.ChildIDs)
		for idx, childID := range node.ChildIDs {
			// Get the definitive status of the child AFTER it has been processed recursively
			childFinalStatus, err := updateAutomaticNodeStatus(tree, childID, logger)
			if err != nil {
				return "", err // Propagate error up
			}
			childStatuses = append(childStatuses, childFinalStatus)

			// Check if the child's FINAL status requires us to grab its symbol
			if childFinalStatus == "special" {
				// Re-fetch child node as its attributes might have changed
				childNode, childExists := tree.NodeMap[childID]
				if !childExists {
					logger.Error("Child node disappeared during parent's update cycle", "parentNodeId", nodeID, "childNodeId", childID)
					// Use exported core.ErrInternal for this unexpected state
					return "", fmt.Errorf("%w: child node %q (of %q) disappeared during update", core.ErrInternal, childID, nodeID)
				}
				sym, ok := childNode.Attributes["special_symbol"]
				if !ok || sym == "" {
					logger.Warn("Special child node missing final symbol attribute, using '?'", "parentNodeId", nodeID, "childNodeId", childID)
					sym = "?"
				}
				childSymbols[idx] = sym
			}
		}
		logger.Debug("Collected final child statuses", "parentNodeId", nodeID, "childStatuses", childStatuses, "childSymbols", childSymbols)
	} else {
		logger.Debug("Node has no children", "nodeId", nodeID)
	}

	// --- Step 2: Determine the FINAL status for THIS node ---
	var finalStatus string // The status this function will return

	if isAutomatic {
		// --- Automatic Node: Calculate status based on children's final statuses ---
		if len(node.ChildIDs) == 0 { // Rule 5: No children -> open
			finalStatus = "open"
			logger.Debug("Calculated automatic status (no children)", "nodeId", nodeID, "finalStatus", finalStatus)
		} else {
			// Calculate using collected final child statuses (Rules 1-4)
			calculatedStatus, calculatedSymbol, calcErr := calculateAutomaticStatus(childStatuses, childSymbols)
			if calcErr != nil {
				logger.Error("Error calculating automatic status", "nodeId", nodeID, "childStatuses", childStatuses, "childSymbols", childSymbols, "error", calcErr)
				return "", fmt.Errorf("calculation failed for node %q: %w", nodeID, calcErr)
			}
			finalStatus = calculatedStatus // Use the calculated status
			logger.Debug("Calculated automatic status", "nodeId", nodeID, "calculatedStatus", finalStatus, "calculatedSymbol", calculatedSymbol)

			// Update attributes *only if necessary*
			currentSymbol, _ := node.Attributes["special_symbol"] // Get symbol *before* potential deletion
			statusChanged := currentStatus != finalStatus
			symbolChanged := false

			logger.Debug("Preparing to update attributes for automatic node", "nodeId", nodeID, "currentStatus", currentStatus, "newStatus", finalStatus, "currentSymbol", currentSymbol, "newSymbol", calculatedSymbol)

			if statusChanged {
				node.Attributes["status"] = finalStatus // Update status
				logger.Debug("Attribute status AFTER direct assignment", "nodeId", nodeID, "status_in_map", node.Attributes["status"])
			}

			if finalStatus == "special" {
				// Ensure symbol is valid (should be guaranteed by calculateAutomaticStatus error check)
				if calculatedSymbol == "" {
					// Use exported core.ErrInternal as calculateAutomaticStatus should prevent this
					return "", fmt.Errorf("%w: node %q calculated status 'special' but symbol is empty (post-calculation)", core.ErrInternal, nodeID)
				}
				if currentSymbol != calculatedSymbol {
					node.Attributes["special_symbol"] = calculatedSymbol
					symbolChanged = true
					logger.Debug("Set/Updated special_symbol attribute", "nodeId", nodeID, "symbol", calculatedSymbol)
				}
			} else {
				// If status is not special, ensure symbol is removed
				_, hadSymbol := node.Attributes["special_symbol"]
				if hadSymbol {
					delete(node.Attributes, "special_symbol")
					symbolChanged = true
					logger.Debug("Removed special_symbol attribute", "nodeId", nodeID)
				}
			}

			// Log final update outcome
			if statusChanged || symbolChanged {
				logger.Debug("Automatic node status/symbol updated", "nodeId", nodeID, "oldStatus", currentStatus, "newStatus", finalStatus, "oldSymbol", currentSymbol, "newSymbol", calculatedSymbol)
			} else {
				logger.Debug("No change needed for status or symbol", "nodeId", nodeID, "status", finalStatus)
			}
		}
	} else { // --- Manual Node ---
		finalStatus = currentStatus // Manual node's status is determined by its own attribute, not children
		// Clean up potentially conflicting attributes just in case (these shouldn't be set on manual nodes)
		delete(node.Attributes, "is_automatic")
		if finalStatus != "special" {
			if _, exists := node.Attributes["special_symbol"]; exists {
				delete(node.Attributes, "special_symbol")
				logger.Debug("Removed stale special_symbol attribute from manual node", "nodeId", nodeID)
			}
		}
		logger.Debug("Node is manual, returning its existing status after child processing", "nodeId", nodeID, "status", finalStatus)
	}

	logger.Debug("Exiting updateAutomaticNodeStatus", "nodeId", nodeID, "finalStatusToReturn", finalStatus)
	return finalStatus, nil // Return the determined final status for this node
}
