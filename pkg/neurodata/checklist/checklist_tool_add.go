// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 00:35:00 PDT // Fix compiler errors in AddItem tool definition and implementation
// pkg/neurodata/checklist/checklist_tool_add.go
package checklist

import (
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/google/uuid" // <<< ADDED for UUID generation
	// NOTE: toolsets import is removed as init() is not here
	// "github.com/aprice2704/neuroscript/pkg/toolsets"
)

// --- Tool Definition and Implementation for ChecklistAddItem ---

// Definition for ChecklistAddItem
var toolChecklistAddItemImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "ChecklistAddItem",
		Description: "Adds a new checklist item node as a child of a specified parent node within a checklist tree handle. " +
			"Returns the ID of the newly created node.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle ID of the checklist tree."},
			{Name: "parentId", Type: core.ArgTypeString, Required: true, Description: "ID of the parent node to add the new item under."},
			{Name: "newItemText", Type: core.ArgTypeString, Required: true, Description: "The text description for the new item."},
			// <<< REMOVED Default field >>>
			{Name: "newItemStatus", Type: core.ArgTypeString, Required: false, Description: "Initial status (e.g., 'open', 'done'). Defaults to 'open' if null/omitted."},
			{Name: "isAutomatic", Type: core.ArgTypeBool, Required: false, Description: "Set to true if the new item's status should be automatically calculated. Defaults to false if null/omitted."},
			{Name: "specialSymbol", Type: core.ArgTypeString, Required: false, Description: "Required only if newItemStatus is 'special'. The single character symbol."},
			{Name: "index", Type: core.ArgTypeInt, Required: false, Description: "Optional zero-based index to insert item. Appends if omitted/null."},
		},
		ReturnType: core.ArgTypeString, // Returns the new node ID
	},
	// NOTE: This Func reference assumes checklist_tool_add.go is part of the main package build,
	//       and that RegisterChecklistTools in checklist_tool.go adds this toolChecklistAddItemImpl.
	Func: toolChecklistAddItem,
}

// Implementation for ChecklistAddItem
func toolChecklistAddItem(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistAddItem"
	logger := interpreter.Logger()
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
	if args[3] != nil {
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
	index := -1
	if args[6] != nil {
		// <<< FIXED: Use type assertion for int conversion >>>
		var indexInt64 int64
		switch v := args[6].(type) {
		case int:
			indexInt64 = int64(v)
		case int64:
			indexInt64 = v
		case float64: // Allow float if it's a whole number
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
	if newItemStatus == "special" {
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

	// <<< FIXED: Use direct struct initialization and uuid.NewString() >>>
	newNode := &core.GenericTreeNode{
		ID:         uuid.NewString(), // Use imported uuid package
		Type:       "checklist_item",
		Value:      newItemText,
		Attributes: make(map[string]string), // Initialize map
		ParentID:   parentID,
		ChildIDs:   []string{}, // Initialize as empty slice
	}
	newNode.Attributes["status"] = newItemStatus // Set status attribute

	if isAutomatic {
		newNode.Attributes["is_automatic"] = "true"
		if newItemStatus != "open" && newItemStatus != "partial" && newItemStatus != "done" {
			newNode.Attributes["status"] = "open"
		}
	}
	if newItemStatus == "special" {
		newNode.Attributes["special_symbol"] = specialSymbol
	}
	tree.NodeMap[newNode.ID] = newNode
	if index < 0 || index >= len(parentNode.ChildIDs) {
		parentNode.ChildIDs = append(parentNode.ChildIDs, newNode.ID)
	} else {
		parentNode.ChildIDs = append(parentNode.ChildIDs[:index+1], parentNode.ChildIDs[index:]...)
		parentNode.ChildIDs[index] = newNode.ID
	}
	logger.Debug("Added new checklist item", "tool", toolName, "newNodeId", newNode.ID, "parentId", parentID, "index", index)
	return newNode.ID, nil
}
