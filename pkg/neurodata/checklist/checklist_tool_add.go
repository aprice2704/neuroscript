// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 18:38:29 PM PDT // Refactor AddItem with Metadata tools
// filename: pkg/neurodata/checklist/checklist_tool_add.go
package checklist

import (
	"errors" // Import errors for Is
	"fmt"    // Import strconv for boolean conversion
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/uuid"
)

// Implementation for ChecklistAddItem using core TreeAddNode and TreeSetNodeMetadata tools.
func toolChecklistAddItem(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistAddItem"
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation (mostly unchanged) ---
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

	newItemStatus := "open" // Default
	if args[3] != nil {
		statusStr, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[3] 'newItemStatus', got %T", core.ErrValidationTypeMismatch, toolName, args[3])
		}
		if !allowedStatuses[statusStr] {
			return nil, fmt.Errorf("%w: %s invalid value for 'newItemStatus': %q", core.ErrInvalidArgument, toolName, statusStr)
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
			return nil, fmt.Errorf("%w: %s expected bool or null arg[4] 'isAutomatic', got %T", core.ErrValidationTypeMismatch, toolName, args[4])
		}
		isAutomatic = isAutoBool
	}

	specialSymbol := ""
	if args[5] != nil {
		symbolStr, ok := args[5].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[5] 'specialSymbol', got %T", core.ErrValidationTypeMismatch, toolName, args[5])
		}
		specialSymbol = symbolStr
	}
	if newItemStatus == "special" {
		if specialSymbol == "" {
			return nil, fmt.Errorf("%w: %s 'newItemStatus' is 'special' but 'specialSymbol' was not provided or is empty", core.ErrInvalidArgument, toolName)
		}
		if utf8.RuneCountInString(specialSymbol) != 1 {
			return nil, fmt.Errorf("%w: %s 'specialSymbol' must be a single character, got %q", core.ErrInvalidArgument, toolName, specialSymbol)
		}
	}

	index := -1 // Default append
	if args[6] != nil {
		var indexInt64 int64
		var err error
		indexInt64, err = core.ConvertToInt64E(args[6]) // Use core helper
		if err != nil {
			return nil, fmt.Errorf("%w: %s invalid 'index' argument: %w", core.ErrInvalidArgument, toolName, err)
		}
		index = int(indexInt64)
		if index < 0 {
			index = -1
		}
	}

	// --- Initial Status Adjustment for Automatic Items ---
	if isAutomatic {
		if newItemStatus != "open" {
			logger.Warn("Ignoring specified status for new automatic item; it will start as 'open' until children are added and UpdateStatus is called.",
				"tool", toolName, "specifiedStatus", newItemStatus)
		}
		newItemStatus = "open"
		specialSymbol = ""
	}

	// --- Refactored Logic using Core Tools ---

	// 1. Get Core Tool Functions
	addNodeImpl, foundAdd := interpreter.ToolRegistry().GetTool("TreeAddNode")
	if !foundAdd || addNodeImpl.Func == nil {
		logger.Error("Core tool 'TreeAddNode' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'TreeAddNode'", core.ErrInternal, toolName)
	}
	setMetaImpl, foundSet := interpreter.ToolRegistry().GetTool("TreeSetNodeMetadata")
	if !foundSet || setMetaImpl.Func == nil {
		logger.Error("Core tool 'TreeSetNodeMetadata' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'TreeSetNodeMetadata'", core.ErrInternal, toolName)
	}

	// 2. Generate New Node ID
	newNodeID := uuid.NewString()

	// 3. Call TreeAddNode to create the basic node with Type: "checklist_item"
	addArgs := core.MakeArgs(handleID, parentID, newNodeID, "checklist_item", newItemText, index)
	logger.Debug("Calling TreeAddNode", "tool", toolName, "parentId", parentID, "newNodeId", newNodeID, "type", "checklist_item", "value", newItemText, "index", index)
	_, addErr := addNodeImpl.Func(interpreter, addArgs)
	if addErr != nil {
		logger.Error("Core TreeAddNode tool failed", "tool", toolName, "error", addErr)
		if errors.Is(addErr, core.ErrNotFound) || errors.Is(addErr, core.ErrInvalidArgument) || errors.Is(addErr, core.ErrNodeIDExists) || errors.Is(addErr, core.ErrListIndexOutOfBounds) {
			return nil, fmt.Errorf("%w: %s failed adding node: %w", core.ErrInvalidArgument, toolName, addErr)
		}
		return nil, fmt.Errorf("%w: %s internal error adding node: %w", core.ErrInternal, toolName, addErr)
	}
	logger.Debug("TreeAddNode call successful", "tool", toolName, "newNodeId", newNodeID)

	// 4. Call TreeSetNodeMetadata for essential checklist attributes
	attributesToSet := map[string]string{
		"status": newItemStatus,
	}
	if isAutomatic {
		attributesToSet["is_automatic"] = "true"
	}
	if newItemStatus == "special" {
		attributesToSet["special_symbol"] = specialSymbol
	}

	// Note: No need to explicitly set "Subtype" as the node Type is now "checklist_item"

	for key, value := range attributesToSet {
		attrArgs := core.MakeArgs(handleID, newNodeID, key, value)
		logger.Debug("Calling TreeSetNodeMetadata", "tool", toolName, "nodeId", newNodeID, "key", key, "value", value)
		_, setErr := setMetaImpl.Func(interpreter, attrArgs)
		if setErr != nil {
			logger.Error("Core TreeSetNodeMetadata tool failed while setting attributes for new node", "tool", toolName, "newNodeId", newNodeID, "attributeKey", key, "error", setErr)
			if errors.Is(setErr, core.ErrNotFound) || errors.Is(setErr, core.ErrInvalidArgument) {
				return nil, fmt.Errorf("%w: %s failed setting attribute '%s': %w", core.ErrInvalidArgument, toolName, key, setErr)
			}
			return nil, fmt.Errorf("%w: %s internal error setting attribute '%s': %w", core.ErrInternal, toolName, key, setErr)
		}
	}

	logger.Debug("Added new checklist item node and attributes successfully", "tool", toolName, "newNodeId", newNodeID, "parentId", parentID)
	return newNodeID, nil // Return the new node's ID
}
