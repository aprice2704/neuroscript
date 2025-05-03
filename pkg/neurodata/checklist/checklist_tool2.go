// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 22:40:10 PDT // Re-fix status rollup logic and priority
package checklist

import (
	"errors"
	"fmt"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// toolChecklistFormatTree - unchanged
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
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		if errors.Is(err, core.ErrHandleNotFound) {
			return nil, fmt.Errorf("%w: %s handle %q not found", core.ErrHandleNotFound, toolName, handleID)
		}
		if errors.Is(err, core.ErrHandleWrongType) {
			return nil, fmt.Errorf("%w: %s handle %q has wrong type", core.ErrHandleWrongType, toolName, handleID)
		}
		// Treat invalid format as InvalidArgument based on test output
		if strings.Contains(err.Error(), "invalid handle format") {
			return nil, fmt.Errorf("%w: %s failed getting handle %q: %w", core.ErrInvalidArgument, toolName, handleID, err)
		}
		return nil, fmt.Errorf("%s failed getting handle %q (type %s): %w", toolName, handleID, core.GenericTreeHandleType, err) // Generic fallback
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrInternalTool, toolName, handleID)
	}
	logger.Debug("Formatting checklist tree to string", "tool", toolName, "handle", handleID)
	formattedString, formatErr := TreeToChecklistString(tree)
	if formatErr != nil {
		logger.Error("Error formatting checklist tree", "tool", toolName, "handle", handleID, "error", formatErr)
		if errors.Is(formatErr, ErrInvalidChecklistTree) || errors.Is(formatErr, ErrMissingStatusAttribute) || errors.Is(formatErr, ErrUnknownStatus) || errors.Is(formatErr, ErrMissingSpecialSymbol) {
			return nil, fmt.Errorf("%w: %s formatting failed: %w", core.ErrInvalidArgument, toolName, formatErr)
		}
		return nil, fmt.Errorf("%w: %s failed formatting tree: %w", core.ErrInternalTool, toolName, formatErr)
	}
	logger.Debug("Successfully formatted checklist tree", "tool", toolName, "handle", handleID)
	return formattedString, nil
}

// toolChecklistSetItemText - unchanged
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
	newText, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newText', got %T", core.ErrValidationTypeMismatch, toolName, args[2])
	}
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		if errors.Is(err, core.ErrHandleNotFound) {
			return nil, fmt.Errorf("%w: %s handle %q not found", core.ErrHandleNotFound, toolName, handleID)
		}
		// Treat invalid format as InvalidArgument based on test output
		if strings.Contains(err.Error(), "invalid handle format") {
			return nil, fmt.Errorf("%w: %s failed getting handle %q: %w", core.ErrInvalidArgument, toolName, handleID, err)
		}
		return nil, fmt.Errorf("%s failed getting handle %q (type %s): %w", toolName, handleID, core.GenericTreeHandleType, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrInternalTool, toolName, handleID)
	}
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", core.ErrNotFound, toolName, nodeID, handleID)
	}
	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node ID %q is type %q, expected 'checklist_item'", core.ErrInvalidArgument, toolName, nodeID, targetNode.Type)
	}
	oldText := ""
	if targetNode.Value != nil {
		oldText = fmt.Sprintf("%v", targetNode.Value)
	}
	targetNode.Value = newText
	logger.Debug("Node text updated", "tool", toolName, "nodeId", nodeID, "oldText", oldText, "newText", newText)
	return nil, nil
}

// toolChecklistUpdateStatus - improved error handling
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
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		// Check specific handle errors first
		if errors.Is(err, core.ErrHandleNotFound) {
			return nil, fmt.Errorf("%w: %s handle %q not found", core.ErrHandleNotFound, toolName, handleID)
		}
		if errors.Is(err, core.ErrHandleWrongType) {
			return nil, fmt.Errorf("%w: %s handle %q has wrong type", core.ErrHandleWrongType, toolName, handleID)
		}
		// Check for invalid format error specifically
		if strings.Contains(err.Error(), "invalid handle format") {
			// Wrap it as ErrInvalidArgument as indicated by test failure message
			return nil, fmt.Errorf("%w: %s handle %q: %w", core.ErrInvalidArgument, toolName, handleID, err)
		}
		// Fallback wrap
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil || tree.RootID == "" {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrInternalTool, toolName, handleID)
	}
	rootNode, rootExists := tree.NodeMap[tree.RootID]
	if !rootExists || rootNode == nil || rootNode.Type != "checklist_root" {
		return nil, fmt.Errorf("%w: %s handle %q has invalid root node", core.ErrInternalTool, toolName, handleID)
	}
	logger.Debug("Starting checklist status update", "tool", toolName, "handle", handleID)
	err = updateChecklistTreeStatus(tree, logger)
	if err != nil {
		logger.Error("Checklist status update failed", "tool", toolName, "handle", handleID, "error", err)
		if errors.Is(err, ErrInvalidChecklistTree) || errors.Is(err, ErrMissingStatusAttribute) || errors.Is(err, ErrInternalParser) {
			return nil, fmt.Errorf("%w: %s update failed due to invalid tree structure or data: %w", core.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s update failed: %w", core.ErrInternalTool, toolName, err)
	}
	logger.Debug("Checklist status update completed successfully", "tool", toolName, "handle", handleID)
	return nil, nil
}

// updateChecklistTreeStatus - unchanged
func updateChecklistTreeStatus(tree *core.GenericTree, logger logging.Logger) error {
	rootNode := tree.NodeMap[tree.RootID]
	for _, childID := range rootNode.ChildIDs {
		_, err := updateAutomaticNodeStatus(tree, childID, logger)
		if err != nil {
			return fmt.Errorf("error updating status starting from node %q: %w", childID, err)
		}
	}
	return nil
}

// updateAutomaticNodeStatus recursively calculates and updates the status of automatic nodes.
// (Logic updated again for correct priority handling and state calculation)
func updateAutomaticNodeStatus(tree *core.GenericTree, nodeID string, logger logging.Logger) (string, error) {
	node, exists := tree.NodeMap[nodeID]
	if !exists || node == nil {
		return "", fmt.Errorf("%w: node %q not found during update", ErrInvalidChecklistTree, nodeID)
	}
	if node.Type != "checklist_item" {
		return "", fmt.Errorf("%w: node %q has unexpected type %q during status update", ErrInvalidChecklistTree, nodeID, node.Type)
	}
	if node.Attributes == nil {
		return "", fmt.Errorf("%w: node %q missing attributes map", ErrInvalidChecklistTree, nodeID)
	}

	isAutomatic := node.Attributes["is_automatic"] == "true"

	// Base Case: Manual Item
	if !isAutomatic {
		currentStatus, statusExists := node.Attributes["status"]
		if !statusExists {
			return "", fmt.Errorf("%w: manual node %q missing 'status' attribute", ErrMissingStatusAttribute, nodeID)
		}
		return currentStatus, nil
	}

	// --- Recursive Step: Automatic Item ---
	// 1. Get statuses of all children first
	childStatuses := make([]string, 0, len(node.ChildIDs))
	childSymbols := make(map[int]string) // Store special symbols by index

	for i, childID := range node.ChildIDs {
		childStatus, err := updateAutomaticNodeStatus(tree, childID, logger)
		if err != nil {
			return "", err // Propagate error up
		}
		childStatuses = append(childStatuses, childStatus)
		if childStatus == "special" {
			childNode, childExists := tree.NodeMap[childID]
			if childExists && childNode != nil && childNode.Attributes != nil {
				childSymbols[i] = childNode.Attributes["special_symbol"]
			}
		}
	}

	// 2. Determine this node's status based on children (Rule 4.2)
	newNodeStatus := "open" // Default: Rule 4 & 5
	newSpecialSymbol := ""

	if len(childStatuses) > 0 {
		// Check priorities in order: !, ?, >, special
		hasBlocked := false
		hasQuestion := false
		hasInProgress := false
		hasSpecial := false
		var firstSpecialSymbol string

		for _, status := range childStatuses {
			if status == "blocked" {
				hasBlocked = true
				break
			} // Highest priority, stop checking
		}

		if !hasBlocked { // Only check lower priorities if higher wasn't found
			for _, status := range childStatuses {
				if status == "question" {
					hasQuestion = true
					break
				}
			}
		}

		if !hasBlocked && !hasQuestion {
			for _, status := range childStatuses {
				if status == "inprogress" {
					hasInProgress = true
					break
				}
			}
		}

		if !hasBlocked && !hasQuestion && !hasInProgress {
			for i, status := range childStatuses {
				if status == "special" {
					hasSpecial = true
					firstSpecialSymbol = childSymbols[i] // Take the symbol from the *first* special child encountered
					break
				}
			}
		}

		// Apply Rule 1
		if hasBlocked {
			newNodeStatus = "blocked"
		} else if hasQuestion {
			newNodeStatus = "question"
		} else if hasInProgress {
			newNodeStatus = "inprogress"
		} else if hasSpecial {
			newNodeStatus = "special"
			newSpecialSymbol = firstSpecialSymbol
		} else {
			// No priority statuses found, check Rules 2, 3, 4
			hasPartialTrigger := false // Tracks if any done/skipped/partial child exists
			allDone := true            // Assume all done unless proven otherwise

			for _, status := range childStatuses {
				if status == "done" || status == "skipped" || status == "partial" {
					hasPartialTrigger = true
				}
				if status != "done" { // Strict check for Rule 3
					allDone = false
				}
			}

			if hasPartialTrigger { // Rule 2
				newNodeStatus = "partial"
			} else if allDone { // Rule 3 (only applies if no partial triggers)
				newNodeStatus = "done"
			}
			// Rule 4 (all open) is the remaining default ("open")
		}
	}
	// Rule 5 (No children) defaults to "open"

	// --- Update Node Attributes ---
	node.Attributes["status"] = newNodeStatus
	if newNodeStatus == "special" {
		if newSpecialSymbol == "" { // Should have been set above if hasSpecial was true
			err := fmt.Errorf("%w: node %q calculated status 'special' but has empty symbol", ErrInternalParser, nodeID)
			logger.Error("Internal error during status update", "error", err, "nodeId", nodeID)
			return "", err // Return internal error
		}
		node.Attributes["special_symbol"] = newSpecialSymbol
	} else {
		delete(node.Attributes, "special_symbol") // Clean up if no longer special
	}

	return newNodeStatus, nil
} // --- End updateAutomaticNodeStatus ---
