// NeuroScript Version: 0.3.0
// File version: 0.1.2
// Corrected core tool lookup to Tree.SetValue and its argument structure.
// filename: pkg/neurodata/checklist/checklist_tool2.go
// nlines: 290 // Approximate
// risk_rating: MEDIUM
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

	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}

	logger.Debug("Formatting checklist tree to string", "tool", toolName, "handle", handleID)
	formattedString, formatErr := TreeToChecklistString(tree)
	if formatErr != nil {
		logger.Error("Error formatting checklist tree", "tool", toolName, "handle", handleID, "error", formatErr)
		if errors.Is(formatErr, ErrInvalidChecklistTree) || errors.Is(formatErr, ErrMissingStatusAttribute) || errors.Is(formatErr, ErrUnknownStatus) || errors.Is(formatErr, ErrMissingSpecialSymbol) {
			return nil, fmt.Errorf("%w: %s formatting failed: %w", core.ErrInvalidArgument, toolName, formatErr)
		}
		return nil, fmt.Errorf("%w: %s failed formatting tree: %w", core.ErrInternal, toolName, formatErr)
	}
	logger.Debug("Successfully formatted checklist tree", "tool", toolName, "handle", handleID)
	return formattedString, nil
}

// toolChecklistSetItemText updates the text value of a checklist item node
// by calling the core Tree.SetValue tool.
func toolChecklistSetItemText(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemText"
	logger := interpreter.Logger()

	// 1. Validate Arguments
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

	// 2. Get Target Node and Validate Type *before* calling core tool
	treeObj, getHandleErr := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if getHandleErr != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, getHandleErr)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", core.ErrNotFound, toolName, nodeID, handleID)
	}

	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node ID %q has type %q, expected type 'checklist_item'",
			core.ErrInvalidArgument, toolName, nodeID, targetNode.Type)
	}

	// 3. Prepare arguments for Tree.SetValue
	coreArgs := core.MakeArgs(handleID, nodeID, newText) // MODIFIED HERE

	// 4. Get and call the core Tree.SetValue tool
	modifyToolImpl, found := interpreter.ToolRegistry().GetTool("Tree.SetValue") // MODIFIED HERE
	if !found || modifyToolImpl.Func == nil {
		logger.Error("Core tool 'Tree.SetValue' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.SetValue' which was not found", core.ErrInternal, toolName)
	}

	logger.Debug("Calling core Tree.SetValue tool", "tool", toolName, "handle", handleID, "nodeId", nodeID, "newText", newText)
	result, err := modifyToolImpl.Func(interpreter, coreArgs)

	// 5. Handle result/error from Tree.SetValue
	if err != nil {
		logger.Error("Core Tree.SetValue tool failed", "tool", toolName, "error", err)
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArgument) || errors.Is(err, core.ErrCannotSetValueOnType) {
			return nil, fmt.Errorf("%w: %s failed: %w", core.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s internal error calling Tree.SetValue: %w", core.ErrInternal, toolName, err)
	}

	logger.Debug("Node text updated successfully via Tree.SetValue", "tool", toolName, "nodeId", nodeID)
	return result, nil
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

	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil || tree.RootID == "" {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}
	rootNode, rootExists := tree.NodeMap[tree.RootID]
	if !rootExists || rootNode == nil || rootNode.Type != "checklist_root" {
		return nil, fmt.Errorf("%w: %s handle %q has invalid root node structure", core.ErrInternal, toolName, handleID)
	}

	logger.Debug("Starting checklist status update", "tool", toolName, "handle", handleID, "rootId", tree.RootID)
	err = updateChecklistTreeStatus(tree, logger)
	if err != nil {
		logger.Error("Checklist status update failed", "tool", toolName, "handle", handleID, "error", err)
		if errors.Is(err, ErrInvalidChecklistTree) || errors.Is(err, ErrMissingStatusAttribute) || errors.Is(err, ErrInternalParser) {
			return nil, fmt.Errorf("%w: %s update failed due to invalid tree structure or data: %w", core.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s update failed: %w", core.ErrInternal, toolName, err)
	}
	logger.Debug("Checklist status update completed successfully", "tool", toolName, "handle", handleID)
	return nil, nil
}

// updateChecklistTreeStatus initiates the recursive update from the root's children.
func updateChecklistTreeStatus(tree *core.GenericTree, logger logging.Logger) error {
	rootNode, exists := tree.NodeMap[tree.RootID]
	if !exists || rootNode == nil {
		return fmt.Errorf("%w: root node %q not found in provided tree", ErrInvalidChecklistTree, tree.RootID)
	}
	if rootNode.ChildIDs == nil {
		logger.Debug("Root node has no children, nothing to update.")
		return nil
	}
	logger.Debug("Updating children of root", "rootId", tree.RootID, "childIds", rootNode.ChildIDs)
	for _, childID := range rootNode.ChildIDs {
		_, err := updateAutomaticNodeStatus(tree, childID, logger)
		if err != nil {
			return fmt.Errorf("error updating status starting from node %q: %w", childID, err)
		}
	}
	return nil
}

// updateAutomaticNodeStatus recursively traverses the tree, calculates and updates the status
// of automatic nodes based on their children's *final* statuses from the recursive calls.
// It now correctly recurses through non-automatic nodes as well and checks for Type "checklist_item".
func updateAutomaticNodeStatus(tree *core.GenericTree, nodeID string, logger logging.Logger) (string, error) {
	node, exists := tree.NodeMap[nodeID]
	if !exists || node == nil {
		logger.Error("Node referenced in tree not found during update", "nodeId", nodeID)
		return "", fmt.Errorf("%w: node %q not found during update", ErrInvalidChecklistTree, nodeID)
	}

	if node.Type != "checklist_item" {
		logger.Warn("Skipping status update for non-checklist item node", "nodeId", nodeID, "type", node.Type)
		currentStatus := "open" // Default
		if node.Attributes != nil {
			if status, ok := node.Attributes["status"]; ok {
				currentStatus = status
			}
		}
		return currentStatus, nil
	}

	if node.Attributes == nil {
		node.Attributes = make(map[string]string)
	}

	isAutomatic := node.Attributes["is_automatic"] == "true"
	currentStatus, statusExists := node.Attributes["status"]
	if !statusExists {
		logger.Warn("Checklist item node missing 'status' attribute, defaulting to 'open'", "nodeId", nodeID)
		node.Attributes["status"] = "open"
		currentStatus = "open"
	}

	logger.Debug("Entering updateAutomaticNodeStatus", "nodeId", nodeID, "isAutomatic", isAutomatic, "currentStatus", currentStatus)

	childStatuses := make([]string, 0, len(node.ChildIDs))
	childSymbols := make(map[int]string)
	if len(node.ChildIDs) > 0 {
		logger.Debug("Processing children", "parentNodeId", nodeID, "childIds", node.ChildIDs)
		for idx, childID := range node.ChildIDs {
			childFinalStatus, err := updateAutomaticNodeStatus(tree, childID, logger)
			if err != nil {
				return "", err
			}
			childStatuses = append(childStatuses, childFinalStatus)
			if childFinalStatus == "special" {
				childNode, childExists := tree.NodeMap[childID]
				if !childExists || childNode == nil {
					logger.Error("Child node disappeared during parent's update cycle", "parentNodeId", nodeID, "childNodeId", childID)
					return "", fmt.Errorf("%w: child node %q (of %q) disappeared during update", core.ErrInternal, childID, nodeID)
				}
				if childNode.Type == "checklist_item" && childNode.Attributes != nil {
					sym, ok := childNode.Attributes["special_symbol"]
					if !ok || sym == "" {
						logger.Warn("Special child node missing final symbol attribute, using '?'", "parentNodeId", nodeID, "childNodeId", childID)
						sym = "?"
					}
					childSymbols[idx] = sym
				} else {
					logger.Warn("Child node encountered during symbol collection was not a checklist item or lacked attributes", "parentNodeId", nodeID, "childNodeId", childID, "childNodeType", childNode.Type)
				}
			}
		}
		logger.Debug("Collected final child statuses", "parentNodeId", nodeID, "childStatuses", childStatuses, "childSymbols", childSymbols)
	} else {
		logger.Debug("Node has no children", "nodeId", nodeID)
	}

	var finalStatus string
	if isAutomatic {
		if len(node.ChildIDs) == 0 {
			finalStatus = "open"
			logger.Debug("Calculated automatic status (no children)", "nodeId", nodeID, "finalStatus", finalStatus)
		} else {
			calculatedStatus, calculatedSymbol, calcErr := calculateAutomaticStatus(childStatuses, childSymbols)
			if calcErr != nil {
				logger.Error("Error calculating automatic status", "nodeId", nodeID, "error", calcErr)
				return "", fmt.Errorf("calculation failed for node %q: %w", nodeID, calcErr)
			}
			finalStatus = calculatedStatus
			logger.Debug("Calculated automatic status", "nodeId", nodeID, "calculatedStatus", finalStatus, "calculatedSymbol", calculatedSymbol)

			currentSymbol, _ := node.Attributes["special_symbol"]
			statusChanged := currentStatus != finalStatus
			symbolChanged := false

			if statusChanged {
				node.Attributes["status"] = finalStatus
			}
			if finalStatus == "special" {
				if calculatedSymbol == "" {
					logger.Error("Internal inconsistency: node calculated status 'special' but symbol is empty post-calculation", "nodeId", nodeID)
					return "", fmt.Errorf("%w: node %q calculated status 'special' but symbol is empty (post-calculation)", core.ErrInternal, nodeID)
				}
				if currentSymbol != calculatedSymbol {
					node.Attributes["special_symbol"] = calculatedSymbol
					symbolChanged = true
				}
			} else {
				_, hadSymbol := node.Attributes["special_symbol"]
				if hadSymbol {
					delete(node.Attributes, "special_symbol")
					symbolChanged = true
				}
			}
			if statusChanged || symbolChanged {
				logger.Debug("Automatic node status/symbol updated", "nodeId", nodeID, "newStatus", finalStatus)
			} else {
				logger.Debug("No change needed for status or symbol", "nodeId", nodeID)
			}
		}
	} else {
		finalStatus = currentStatus

		if finalStatus != "special" {
			if _, exists := node.Attributes["special_symbol"]; exists {
				delete(node.Attributes, "special_symbol")
				logger.Debug("Removed stale special_symbol attribute from manual node", "nodeId", nodeID)
			}
		}
		logger.Debug("Node is manual, returning its existing status after child processing", "nodeId", nodeID, "status", finalStatus)
	}

	logger.Debug("Exiting updateAutomaticNodeStatus", "nodeId", nodeID, "finalStatusToReturn", finalStatus)
	return finalStatus, nil
}
