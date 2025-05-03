// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 16:35:00 PM PDT // Use defined allowedStatuses
// filename: pkg/neurodata/checklist/checklist_tool_add.go
package checklist

import (
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/uuid" // Assuming this is used for node ID generation
)

// Implementation for ChecklistAddItem
func toolChecklistAddItem(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistAddItem"
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
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

	// --- Status Parsing & Validation ---
	newItemStatus := "open" // Default
	if args[3] != nil {
		statusStr, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[3] 'newItemStatus', got %T", core.ErrValidationTypeMismatch, toolName, args[3])
		}
		// <<< FIX: Use allowedStatuses defined in checklist_tool.go >>>
		// Note: We might want a stricter set for ADDING items (e.g., disallow 'partial')
		// but for now, using the main set.
		if !allowedStatuses[statusStr] {
			return nil, fmt.Errorf("%w: %s invalid value for 'newItemStatus': %q", core.ErrInvalidArgument, toolName, statusStr)
		}
		// Prevent setting calculated statuses like 'partial' directly? For now, allow based on map.
		if statusStr == "partial" {
			logger.Warn("Setting initial status to 'partial' might be overwritten by UpdateStatus.", "tool", toolName)
		}
		newItemStatus = statusStr
	}

	// --- isAutomatic Parsing ---
	isAutomatic := false
	if args[4] != nil {
		isAutoBool, ok := args[4].(bool)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected bool or null arg[4] 'isAutomatic', got %T", core.ErrValidationTypeMismatch, toolName, args[4])
		}
		isAutomatic = isAutoBool
	}

	// --- specialSymbol Parsing & Validation ---
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

	// --- index Parsing ---
	index := -1 // Default append
	if args[6] != nil {
		var indexInt64 int64
		switch v := args[6].(type) {
		case int:
			indexInt64 = int64(v)
		case int64:
			indexInt64 = v
		case float64: // Allow float if it's a whole number
			if v != float64(int64(v)) {
				return nil, fmt.Errorf("%w: %s expected integer arg[6] 'index', got non-integer float %f", core.ErrValidationTypeMismatch, toolName, v)
			}
			indexInt64 = int64(v)
		default:
			return nil, fmt.Errorf("%w: %s expected integer or null arg[6] 'index', got %T", core.ErrValidationTypeMismatch, toolName, args[6])
		}
		index = int(indexInt64)
		// Treat negative index as append
		if index < 0 {
			index = -1
		}
	}

	// --- Initial Status Adjustment for Automatic Items ---
	// An automatic item without children should always start as 'open' regardless of input,
	// as its status will be calculated by UpdateStatus based on having no children.
	if isAutomatic {
		if newItemStatus != "open" {
			logger.Warn("Ignoring specified status for new automatic item; it will start as 'open' until children are added and UpdateStatus is called.",
				"tool", toolName, "specifiedStatus", newItemStatus)
		}
		newItemStatus = "open" // Force initial status to open for new automatic items
		specialSymbol = ""     // Cannot be special if it starts as open
	}

	// --- Get Tree and Parent Node ---
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}
	parentNode, exists := tree.NodeMap[parentID]
	if !exists {
		return nil, fmt.Errorf("%w: parent node ID %q not found in tree handle %q", core.ErrNotFound, parentID, handleID)
	}
	// Allow adding to root or other checklist items
	if parentNode.Type != "checklist_root" && parentNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s parent node %q has invalid type %q, expected 'checklist_root' or 'checklist_item'", core.ErrInvalidArgument, toolName, parentID, parentNode.Type)
	}

	// --- Create and Add New Node ---
	newNodeID := uuid.NewString() // Generate new ID
	newNode := &core.GenericTreeNode{
		ID:         newNodeID,
		Type:       "checklist_item", // Always this type
		Value:      newItemText,
		Attributes: make(map[string]string),
		ParentID:   parentID,
		ChildIDs:   []string{}, // New node has no children initially
		// Tree pointer might not be strictly necessary if not used by node methods
	}
	newNode.Attributes["status"] = newItemStatus
	if isAutomatic {
		newNode.Attributes["is_automatic"] = "true"
	}
	if newItemStatus == "special" {
		// We already validated specialSymbol is a single char if status is special
		newNode.Attributes["special_symbol"] = specialSymbol
	}

	// Add to tree map
	tree.NodeMap[newNodeID] = newNode

	// Attach to parent's children list at the correct index
	if parentNode.ChildIDs == nil {
		parentNode.ChildIDs = make([]string, 0, 1)
	}

	logger.Debug("Parent node BEFORE modification", "tool", toolName, "parentId", parentID, "type", parentNode.Type, "numChildren", len(parentNode.ChildIDs), "children", parentNode.ChildIDs, "targetIndex", index)

	if index == -1 || index >= len(parentNode.ChildIDs) { // Append
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNodeID)
		logger.Debug("Appending new node", "tool", toolName, "newNodeId", newNodeID, "parentId", parentID, "originalIndex", index)
	} else { // Insert
		parentNode.ChildIDs = append(parentNode.ChildIDs[:index], append([]string{newNodeID}, parentNode.ChildIDs[index:]...)...)
		logger.Debug("Inserting new node at index", "tool", toolName, "newNodeId", newNodeID, "parentId", parentID, "index", index)
	}

	logger.Debug("Parent node AFTER modification", "tool", toolName, "parentId", parentID, "type", parentNode.Type, "numChildren", len(parentNode.ChildIDs), "children", parentNode.ChildIDs)

	logger.Debug("Added new checklist item node successfully", "tool", toolName, "newNodeId", newNodeID, "parentId", parentID)
	return newNodeID, nil // Return the new node's ID
}
