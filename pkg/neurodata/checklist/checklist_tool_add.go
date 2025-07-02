// NeuroScript Version: 0.3.0
// File version: 0.1.2
// Corrected Tree.AddChildNode call and implemented indexed insertion.
// filename: pkg/neurodata/checklist/checklist_tool_add.go
// nlines: 180 // Approximate
// risk_rating: MEDIUM
package checklist

import (
	"errors"
	"fmt"
	"slices"	// For slice manipulation
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
	"github.com/google/uuid"
)

// Implementation for ChecklistAddItem using core Tree.AddChildNode and Tree.SetNodeMetadata tools.
// Handles indexed insertion for array-like parent nodes.
func toolChecklistAddItem(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistAddItem"
	logger := interpreter.Logger()

	if len(args) != 7 {
		return nil, fmt.Errorf("%w: %s expected 7 arguments (handle, parentId, text, status, isAuto, symbol, index), got %d. Use null for optional args.", lang.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", lang.ErrValidationTypeMismatch, toolName, args[0])
	}
	parentID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[1] 'parentId', got %T", lang.ErrValidationTypeMismatch, toolName, args[1])
	}
	newItemText, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newItemText', got %T", lang.ErrValidationTypeMismatch, toolName, args[2])
	}

	newItemStatus := "open"
	if args[3] != nil {
		statusStr, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[3] 'newItemStatus', got %T", lang.ErrValidationTypeMismatch, toolName, args[3])
		}
		if !allowedStatuses[statusStr] {
			return nil, fmt.Errorf("%w: %s invalid value for 'newItemStatus': %q", lang.ErrInvalidArgument, toolName, statusStr)
		}
		if statusStr == "partial" {
			logger.Warn("Setting initial status to 'partial' might be overwritten by UpdateStatus.", "tool", toolName)
		}
		newItemStatus = statusStr
	}

	isAutomatic := false
	if args[4] != nil {
		isAutoBool, ok := args[4].(bool)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected bool or null arg[4] 'isAutomatic', got %T", lang.ErrValidationTypeMismatch, toolName, args[4])
		}
		isAutomatic = isAutoBool
	}

	specialSymbol := ""
	if args[5] != nil {
		symbolStr, ok := args[5].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[5] 'specialSymbol', got %T", lang.ErrValidationTypeMismatch, toolName, args[5])
		}
		specialSymbol = symbolStr
	}
	if newItemStatus == "special" {
		if specialSymbol == "" {
			return nil, fmt.Errorf("%w: %s 'newItemStatus' is 'special' but 'specialSymbol' was not provided or is empty", lang.ErrInvalidArgument, toolName)
		}
		if utf8.RuneCountInString(specialSymbol) != 1 {
			return nil, fmt.Errorf("%w: %s 'specialSymbol' must be a single character, got %q", lang.ErrInvalidArgument, toolName, specialSymbol)
		}
	}

	index := -1
	if args[6] != nil {
		var indexInt64 int64
		var err error
		indexInt64, err = utils.ConvertToInt64E(args[6])
		if err != nil {
			return nil, fmt.Errorf("%w: %s invalid 'index' argument: %w", lang.ErrInvalidArgument, toolName, err)
		}
		index = int(indexInt64)
		// No negative index for actual insertion, -1 means append.
		// If user provides < -1, treat as append.
		if index < -1 {
			index = -1
		}
	}

	if isAutomatic {
		if newItemStatus != "open" {
			logger.Warn("Ignoring specified status for new automatic item; it will start as 'open' until children are added and UpdateStatus is called.",
				"tool", toolName, "specifiedStatus", newItemStatus)
		}
		newItemStatus = "open"
		specialSymbol = ""
	}

	addNodeImpl, foundAdd := interpreter.ToolRegistry().GetTool("Tree.AddChildNode")
	if !foundAdd || addNodeImpl.Func == nil {
		logger.Error("Core tool 'Tree.AddChildNode' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.AddChildNode'", lang.ErrInternal, toolName)
	}
	setMetaImpl, foundSet := interpreter.ToolRegistry().GetTool("Tree.SetNodeMetadata")
	if !foundSet || setMetaImpl.Func == nil {
		logger.Error("Core tool 'Tree.SetNodeMetadata' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.SetNodeMetadata'", lang.ErrInternal, toolName)
	}

	newNodeID := uuid.NewString()

	// MODIFIED: Pass nil for key_for_object_parent.
	// Tree.AddChildNode will append if parent is array-like.
	// Value (newItemText) is appropriate for 'checklist_item' type.
	addArgs := tool.MakeArgs(handleID, parentID, newNodeID, "checklist_item", newItemText, nil)
	logger.Debug("Calling Tree.AddChildNode", "tool", toolName, "parentId", parentID, "newNodeId", newNodeID, "type", "checklist_item", "value", newItemText, "key_for_object_parent", nil)
	_, addErr := addNodeImpl.Func(interpreter, addArgs)
	if addErr != nil {
		logger.Error("Core Tree.AddChildNode tool failed", "tool", toolName, "error", addErr)
		if errors.Is(addErr, lang.ErrNotFound) || errors.Is(addErr, lang.ErrInvalidArgument) || errors.Is(addErr, lang.ErrNodeIDExists) || errors.Is(addErr, lang.ErrListIndexOutOfBounds) || errors.Is(addErr, lang.ErrNodeWrongType) {
			return nil, fmt.Errorf("%w: %s failed adding node: %w", lang.ErrInvalidArgument, toolName, addErr)
		}
		return nil, fmt.Errorf("%w: %s internal error adding node: %w", lang.ErrInternal, toolName, addErr)
	}
	logger.Debug("Tree.AddChildNode call successful", "tool", toolName, "newNodeId", newNodeID)

	// --- Handle indexed insertion manually for array-like parents ---
	// This needs to happen after the node is added to the NodeMap by Tree.AddChildNode.
	if index != -1 {	// -1 means append, which Tree.AddChildNode does by default for array types
		treeObj, getHandleErr := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
		if getHandleErr != nil {
			return nil, fmt.Errorf("%s: failed to get tree handle %q for indexed insertion: %w", toolName, handleID, getHandleErr)
		}
		tree, castOk := treeObj.(*GenericTree)
		if !castOk || tree == nil || tree.NodeMap == nil {
			return nil, fmt.Errorf("%w: %s: handle %q did not contain a valid GenericTree for indexed insertion", lang.ErrHandleInvalid, toolName, handleID)
		}
		parentNode, parentExists := tree.NodeMap[parentID]
		if !parentExists {
			// Should have been caught by Tree.AddChildNode, but defensive check.
			return nil, fmt.Errorf("%w: %s: parent node %q not found for indexed insertion", lang.ErrNotFound, toolName, parentID)
		}

		// Only perform indexed insertion if parent is 'checklist_root' or 'checklist_item' (array-like behavior for ChildIDs)
		// or explicitly an 'array' type.
		if parentNode.Type == "checklist_root" || parentNode.Type == "checklist_item" || parentNode.Type == "array" {
			// Remove the newly added nodeID from the end (where Tree.AddChildNode put it for array-like parents)
			// Ensure ChildIDs is not nil
			if parentNode.ChildIDs == nil {
				parentNode.ChildIDs = []string{}	// Should not happen if Tree.AddChildNode succeeded
			}

			// Find and remove the newNodeID from its current lang.Position
			foundAtIndex := -1
			for i, childNodeID := range parentNode.ChildIDs {
				if childNodeID == newNodeID {
					foundAtIndex = i
					break
				}
			}

			if foundAtIndex != -1 {
				parentNode.ChildIDs = slices.Delete(parentNode.ChildIDs, foundAtIndex, foundAtIndex+1)
			} else {
				// This is unexpected if Tree.AddChildNode succeeded and parent type is correct.
				logger.Warn("Newly added node not found in parent's children list for reordering.", "tool", toolName, "parentNodeId", parentID, "newNodeId", newNodeID)
				// Continue to attempt insertion at index, NodeMap is the source of truth.
			}

			// Clamp index to valid range for insertion
			if index < 0 {	// Should be caught by earlier check, but defensive
				index = 0
			}
			if index > len(parentNode.ChildIDs) {
				index = len(parentNode.ChildIDs)	// Insert at the end
			}

			parentNode.ChildIDs = slices.Insert(parentNode.ChildIDs, index, newNodeID)
			logger.Debug("Reordered children for indexed insertion", "tool", toolName, "parentNodeId", parentID, "newNodeId", newNodeID, "index", index)
		} else {
			logger.Warn("Index parameter provided but parent node is not array-like; index ignored.", "tool", toolName, "parentNodeId", parentID, "parentNodeType", parentNode.Type, "index", index)
		}
	}

	// --- Set Metadata ---
	attributesToSet := map[string]string{
		"status": newItemStatus,
	}
	if isAutomatic {
		attributesToSet["is_automatic"] = "true"
	}
	if newItemStatus == "special" {
		attributesToSet["special_symbol"] = specialSymbol
	}

	for key, value := range attributesToSet {
		attrArgs := tool.MakeArgs(handleID, newNodeID, key, value)
		logger.Debug("Calling TreeSetNodeMetadata", "tool", toolName, "nodeId", newNodeID, "key", key, "value", value)
		_, setErr := setMetaImpl.Func(interpreter, attrArgs)
		if setErr != nil {
			logger.Error("Core TreeSetNodeMetadata tool failed while setting attributes for new node", "tool", toolName, "newNodeId", newNodeID, "attributeKey", key, "error", setErr)
			// It's possible the node was removed due to an earlier error in indexed insertion logic if not handled perfectly,
			// or other concurrent modification (though unlikely in single tool exec).
			if errors.Is(setErr, lang.ErrNotFound) || errors.Is(setErr, lang.ErrInvalidArgument) {
				return nil, fmt.Errorf("%w: %s failed setting attribute '%s': %w", lang.ErrInvalidArgument, toolName, key, setErr)
			}
			return nil, fmt.Errorf("%w: %s internal error setting attribute '%s': %w", lang.ErrInternal, toolName, key, setErr)
		}
	}

	logger.Debug("Added new checklist item node and attributes successfully", "tool", toolName, "newNodeId", newNodeID, "parentId", parentID)
	return newNodeID, nil
}