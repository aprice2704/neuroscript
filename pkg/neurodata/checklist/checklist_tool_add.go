// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 01:10:00 AM PDT // Add MORE pre/post insertion debug logs
// pkg/neurodata/checklist/checklist_tool_add.go
package checklist

import (
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/uuid"
)

// Implementation for ChecklistAddItem
func toolChecklistAddItem(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistAddItem"
	logger := interpreter.Logger()
	// Argument parsing and validation (Unchanged)...
	if len(args) != 7 {
		return nil, fmt.Errorf("%w: %s expected 7 arguments (handle, parentId, text, status, isAuto, symbol, index), got %d. Use null for optional args.", core.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}
	parentID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[1] 'parentId', got %T", core.ErrValidationTypeMismatch, toolName, args[1])
	}
	newItemText, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newItemText', got %T", core.ErrValidationTypeMismatch, toolName, args[2])
	}
	newItemStatus := "open"
	if args[3] != nil { /* ... status parsing ... */
		statusStr, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[3] 'newItemStatus', got %T", core.ErrValidationTypeMismatch, toolName, args[3])
		}
		if !allowedLeafStatuses[statusStr] {
			return nil, fmt.Errorf("%w: %s invalid value for 'newItemStatus': %q", core.ErrInvalidArgument, toolName, statusStr)
		}
		newItemStatus = statusStr
	}
	isAutomatic := false
	if args[4] != nil { /* ... isAutomatic parsing ... */
		isAutoBool, ok := args[4].(bool)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected bool or null arg[4] 'isAutomatic', got %T", core.ErrValidationTypeMismatch, toolName, args[4])
		}
		isAutomatic = isAutoBool
	}
	specialSymbol := ""
	if args[5] != nil { /* ... specialSymbol parsing ... */
		symbolStr, ok := args[5].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[5] 'specialSymbol', got %T", core.ErrValidationTypeMismatch, toolName, args[5])
		}
		specialSymbol = symbolStr
	}
	index := -1
	if args[6] != nil { /* ... index parsing ... */
		var indexInt64 int64
		switch v := args[6].(type) {
		case int:
			indexInt64 = int64(v)
		case int64:
			indexInt64 = v
		case float64:
			if v == float64(int64(v)) {
				indexInt64 = int64(v)
			} else {
				return nil, fmt.Errorf("%w: %s expected integer arg[6] 'index', got non-integer float %f", core.ErrValidationTypeMismatch, toolName, v)
			}
		default:
			return nil, fmt.Errorf("%w: %s expected integer or null arg[6] 'index', got %T", core.ErrValidationTypeMismatch, toolName, args[6])
		}
		index = int(indexInt64)
		if index < 0 {
			index = -1
		}
	}
	if newItemStatus == "special" { /* ... validation ... */
		if specialSymbol == "" {
			return nil, fmt.Errorf("%w: %s 'newItemStatus' is 'special' but 'specialSymbol' was not provided or is empty", core.ErrInvalidArgument, toolName)
		}
		if utf8.RuneCountInString(specialSymbol) != 1 {
			return nil, fmt.Errorf("%w: %s 'specialSymbol' must be a single character, got %q", core.ErrInvalidArgument, toolName, specialSymbol)
		}
	}
	if isAutomatic && newItemStatus != "open" {
		logger.Warn("Ignoring specified status for new automatic item, starting as 'open'.", "tool", toolName, "specifiedStatus", newItemStatus)
		newItemStatus = "open"
	}
	// Get Tree and Parent (Unchanged)...
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("failed getting handle %q: %w", handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrInternalTool, toolName, handleID)
	}
	parentNode, exists := tree.NodeMap[parentID]
	if !exists {
		return nil, fmt.Errorf("%w: parent node ID %q not found", core.ErrNotFound, parentID)
	}
	if parentNode.Type != "checklist_root" && parentNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s parent node %q has invalid type %q", core.ErrInvalidArgument, toolName, parentID, parentNode.Type)
	}

	// Create New Node (Unchanged)...
	newNode := &core.GenericTreeNode{ID: uuid.NewString(), Type: "checklist_item", Value: newItemText, Attributes: make(map[string]string), ParentID: parentID, ChildIDs: []string{}}
	newNode.Attributes["status"] = newItemStatus
	if isAutomatic {
		newNode.Attributes["is_automatic"] = "true"
	}
	if newItemStatus == "special" {
		newNode.Attributes["special_symbol"] = specialSymbol
	}

	// Add to tree map (Unchanged)...
	tree.NodeMap[newNode.ID] = newNode

	// Ensure parent ChildIDs slice exists (Unchanged)...
	if parentNode.ChildIDs == nil {
		parentNode.ChildIDs = make([]string, 0, 1)
		logger.Debug("Initialized parent ChildIDs slice", "tool", toolName, "parentId", parentID)
	}

	// *** MORE DEBUG LOGGING ADDED ***
	logger.Debug("Parent node BEFORE modification", "tool", toolName, "parentId", parentID, "type", parentNode.Type, "numChildren", len(parentNode.ChildIDs), "children", parentNode.ChildIDs, "targetIndex", index)

	// Handle insertion index
	if index < 0 || index >= len(parentNode.ChildIDs) {
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNode.ID)
		logger.Debug("Appending new node", "tool", toolName, "newNodeId", newNode.ID, "parentId", parentID, "originalIndex", index)
	} else {
		// Correct insertion logic
		parentNode.ChildIDs = append(parentNode.ChildIDs[:index], append([]string{newNode.ID}, parentNode.ChildIDs[index:]...)...)
		logger.Debug("Inserting new node at index", "tool", toolName, "newNodeId", newNode.ID, "parentId", parentID, "index", index)
	}

	// *** MORE DEBUG LOGGING ADDED ***
	logger.Debug("Parent node AFTER modification", "tool", toolName, "parentId", parentID, "type", parentNode.Type, "numChildren", len(parentNode.ChildIDs), "children", parentNode.ChildIDs)

	logger.Debug("Added new checklist item node successfully", "tool", toolName, "newNodeId", newNode.ID, "parentId", parentID) // Changed log message slightly
	return newNode.ID, nil
}
