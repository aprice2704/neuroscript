// NeuroScript Version: 0.3.1
// File version: 0.2.0
// Purpose: Updated to use safe type assertions for attributes from  TreeAttrs (map[string]interface{}).
// filename: pkg/neurodata/checklist/checklist_tool2.go

package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/aprice2704/neuroscript/pkg/utils"
)

// toolChecklistFormatTree formats the checklist tree back into a string.
func toolChecklistFormatTree(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "ChecklistFormatTree"
	logger := interpreter.Logger()
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (handle), got %d", lang.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", lang.ErrValidationTypeMismatch, toolName, args[0])
	}

	treeObj, err := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", lang.ErrHandleInvalid, toolName, handleID)
	}

	logger.Debug("Formatting checklist tree to string", "tool", toolName, "handle", handleID)
	formattedString, formatErr := TreeToChecklistString(tree)
	if formatErr != nil {
		logger.Error("Error formatting checklist tree", "tool", toolName, "handle", handleID, "error", formatErr)
		if errors.Is(formatErr, ErrInvalidChecklistTree) || errors.Is(formatErr, ErrMissingStatusAttribute) || errors.Is(formatErr, ErrUnknownStatus) || errors.Is(formatErr, ErrMissingSpecialSymbol) {
			return nil, fmt.Errorf("%w: %s formatting failed: %w", lang.ErrInvalidArgument, toolName, formatErr)
		}
		return nil, fmt.Errorf("%w: %s failed formatting tree: %w", lang.ErrInternal, toolName, formatErr)
	}
	logger.Debug("Successfully formatted checklist tree", "tool", toolName, "handle", handleID)
	return formattedString, nil
}

// toolChecklistSetItemText updates the text value of a checklist item node
// by calling the core Tree.SetValue tool.
func toolChecklistSetItemText(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemText"
	logger := interpreter.Logger()

	if len(args) != 3 {
		return nil, fmt.Errorf("%w: %s expected 3 arguments (handle, nodeId, newText), got %d", lang.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", lang.ErrValidationTypeMismatch, toolName, args[0])
	}
	nodeID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[1] 'nodeId', got %T", lang.ErrValidationTypeMismatch, toolName, args[1])
	}
	if nodeID == "" {
		return nil, fmt.Errorf("%w: %s requires non-empty 'nodeId'", lang.ErrValidationRequiredArgNil, toolName)
	}
	newText, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newText', got %T", lang.ErrValidationTypeMismatch, toolName, args[2])
	}

	treeObj, getHandleErr := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if getHandleErr != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, getHandleErr)
	}
	tree, ok := treeObj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", lang.ErrHandleInvalid, toolName, handleID)
	}
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", lang.ErrNotFound, toolName, nodeID, handleID)
	}

	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node ID %q has type %q, expected type 'checklist_item'",
			lang.ErrInvalidArgument, toolName, nodeID, targetNode.Type)
	}

	coreArgs := tool.MakeArgs(handleID, nodeID, newText)

	modifyToolImpl, found := interpreter.ToolRegistry().GetTool("Tree.SetValue")
	if !found || modifyToolImpl.Func == nil {
		logger.Error("Core tool 'Tree.SetValue' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.SetValue' which was not found", lang.ErrInternal, toolName)
	}

	logger.Debug("Calling core Tree.SetValue tool", "tool", toolName, "handle", handleID, "nodeId", nodeID, "newText", newText)
	result, err := modifyToolImpl.Func(interpreter, coreArgs)

	if err != nil {
		logger.Error("Core Tree.SetValue tool failed", "tool", toolName, "error", err)
		if errors.Is(err, lang.ErrNotFound) || errors.Is(err, lang.ErrInvalidArgument) || errors.Is(err, lang.ErrCannotSetValueOnType) {
			return nil, fmt.Errorf("%w: %s failed: %w", lang.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s internal error calling Tree.SetValue: %w", lang.ErrInternal, toolName, err)
	}

	logger.Debug("Node text updated successfully via Tree.SetValue", "tool", toolName, "nodeId", nodeID)
	return result, nil
}

// toolChecklistUpdateStatus triggers the recursive status update for the entire checklist tree.
func toolChecklistUpdateStatus(interpreter tool.RunTime, args []interface{}) (interface{}, error) {
	toolName := "Checklist.UpdateStatus"
	logger := interpreter.Logger()
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (handle), got %d", lang.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", lang.ErrValidationTypeMismatch, toolName, args[0])
	}

	treeObj, err := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil || tree.RootID == "" {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", lang.ErrHandleInvalid, toolName, handleID)
	}
	rootNode, rootExists := tree.NodeMap[tree.RootID]
	if !rootExists || rootNode == nil || rootNode.Type != "checklist_root" {
		return nil, fmt.Errorf("%w: %s handle %q has invalid root node structure", lang.ErrInternal, toolName, handleID)
	}

	logger.Debug("Starting checklist status update", "tool", toolName, "handle", handleID, "rootId", tree.RootID)
	err = updateChecklistTreeStatus(tree, logger)
	if err != nil {
		logger.Error("Checklist status update failed", "tool", toolName, "handle", handleID, "error", err)
		if errors.Is(err, ErrInvalidChecklistTree) || errors.Is(err, ErrMissingStatusAttribute) || errors.Is(err, ErrInternalParser) {
			return nil, fmt.Errorf("%w: %s update failed due to invalid tree structure or data: %w", lang.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s update failed: %w", lang.ErrInternal, toolName, err)
	}
	logger.Debug("Checklist status update completed successfully", "tool", toolName, "handle", handleID)
	return nil, nil
}

// updateChecklistTreeStatus initiates the recursive update from the root's children.
func updateChecklistTreeStatus(tree *utils.GenericTree, logger interfaces.Logger) error {
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

func updateAutomaticNodeStatus(tree *utils.GenericTree, nodeID string, logger interfaces.Logger) (string, error) {
	node, exists := tree.NodeMap[nodeID]
	if !exists || node == nil {
		logger.Error("Node referenced in tree not found during update", "nodeId", nodeID)
		return "", fmt.Errorf("%w: node %q not found during update", ErrInvalidChecklistTree, nodeID)
	}

	if node.Type != "checklist_item" {
		logger.Warn("Skipping status update for non-checklist item node", "nodeId", nodeID, "type", node.Type)
		// FIX: Safely assert status to a string, provide a default if it doesn't exist or is the wrong type.
		var currentStatus string
		if node.Attributes != nil {
			if status, ok := node.Attributes["status"].(string); ok {
				currentStatus = status
			}
		}
		if currentStatus == "" {
			currentStatus = "open"
		}
		return currentStatus, nil
	}

	if node.Attributes == nil {
		// FIX: Align with the new  TreeAttrs type (map[string]interface{})
		node.Attributes = make(utils.TreeAttrs)
	}

	// FIX: Safely assert status to a string, providing a default.
	var currentStatus string
	if statusVal, ok := node.Attributes["status"]; ok {
		currentStatus, _ = statusVal.(string) // Use blank identifier, default to "" if not a string
	}
	if currentStatus == "" {
		node.Attributes["status"] = "open"
		currentStatus = "open"
		logger.Warn("Checklist item node missing 'status' attribute, defaulting to 'open'", "nodeId", nodeID)
	}

	// FIX: Safely assert is_automatic to a bool.
	isAutomatic, _ := node.Attributes["is_automatic"].(bool)

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
					return "", fmt.Errorf("%w: child node %q (of %q) disappeared during update", lang.ErrInternal, childID, nodeID)
				}
				if childNode.Type == "checklist_item" && childNode.Attributes != nil {
					// FIX: Safely assert special_symbol to a string.
					if symStr, ok := childNode.Attributes["special_symbol"].(string); ok && symStr != "" {
						childSymbols[idx] = symStr
					} else {
						logger.Warn("Special child node missing symbol attribute, using '?'", "parentNodeId", nodeID, "childNodeId", childID)
						childSymbols[idx] = "?"
					}
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

			// FIX: Safely assert current symbol to a string.
			currentSymbol, _ := node.Attributes["special_symbol"].(string)
			statusChanged := currentStatus != finalStatus
			symbolChanged := false

			if statusChanged {
				node.Attributes["status"] = finalStatus
			}
			if finalStatus == "special" {
				if calculatedSymbol == "" {
					logger.Error("Internal inconsistency: node calculated status 'special' but symbol is empty post-calculation", "nodeId", nodeID)
					return "", fmt.Errorf("%w: node %q calculated status 'special' but symbol is empty (post-calculation)", lang.ErrInternal, nodeID)
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
		// FIX: Ensure finalStatus is the string from the attribute map.
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
