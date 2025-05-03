// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 17:10:00 PM PDT // Fix registration signature, return errors
// filename: pkg/neurodata/checklist/checklist_tool.go
package checklist

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// --- Allowed Statuses ---
var allowedStatuses = map[string]bool{
	"open":       true,
	"done":       true,
	"skipped":    true,
	"partial":    true,
	"inprogress": true,
	"question":   true,
	"blocked":    true,
	"special":    true,
}

// --- Tool Implementation Struct Definitions ---

var toolChecklistLoadTreeImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistLoadTree", // Args, Desc, ReturnType...
		Description: "Parses a checklist string (in Markdown format with :: metadata) and loads it into a GenericTree handle. Returns the handle ID string.",
		Args:        []core.ArgSpec{{Name: "checklist_string", Type: core.ArgTypeString, Required: true, Description: "The checklist content as a string."}},
		ReturnType:  core.ArgTypeString,
	},
	Func: toolChecklistLoadTree,
}

var toolChecklistFormatTreeImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistFormatTree", // Args, Desc, ReturnType...
		Description: "Formats a checklist GenericTree handle back into its Markdown string representation.",
		Args:        []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."}},
		ReturnType:  core.ArgTypeString,
	},
	Func: toolChecklistFormatTree,
}

var toolChecklistSetItemStatusImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistSetItemStatus", // Args, Desc, ReturnType...
		Description: "Sets the status of a specific checklist item node. Does NOT automatically update parent statuses; call Checklist.UpdateStatus explicitly afterwards if needed. Returns nil on success.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."},
			{Name: "node_id", Type: core.ArgTypeString, Required: true, Description: "The unique ID of the checklist item node."},
			{Name: "new_status", Type: core.ArgTypeString, Required: true, Description: "The new status (e.g., 'open', 'done', 'skipped', 'partial', 'inprogress', 'question', 'blocked', 'special')."},
			{Name: "special_symbol", Type: core.ArgTypeString, Required: false, Description: "Required single character if new_status is 'special', otherwise ignored."},
		},
		ReturnType: core.ArgTypeNil,
	},
	Func: toolChecklistSetItemStatus,
}

var toolChecklistSetItemTextImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistSetItemText", // Args, Desc, ReturnType...
		Description: "Sets the text (Value) of a specific checklist item node. Returns nil on success.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."},
			{Name: "node_id", Type: core.ArgTypeString, Required: true, Description: "The unique ID of the checklist item node."},
			{Name: "new_text", Type: core.ArgTypeString, Required: true, Description: "The new text content for the item."},
		},
		ReturnType: core.ArgTypeNil,
	},
	Func: toolChecklistSetItemText,
}

var toolChecklistAddItemImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistAddItem", // Args, Desc, ReturnType...
		Description: "Adds a new checklist item node as a child of the specified parent node ID. Does NOT automatically update parent statuses; call Checklist.UpdateStatus explicitly afterwards. Returns the new node's ID string.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID."},
			{Name: "parent_id", Type: core.ArgTypeString, Required: true, Description: "ID of the parent node (can be root or another item)."},
			{Name: "new_item_text", Type: core.ArgTypeString, Required: true, Description: "Text content for the new item."},
			{Name: "new_item_status", Type: core.ArgTypeString, Required: false, Description: "Initial status (default 'open'). Use allowed leaf statuses only."},
			{Name: "is_automatic", Type: core.ArgTypeBool, Required: false, Description: "Whether the new item is automatic (default false)."},
			{Name: "special_symbol", Type: core.ArgTypeString, Required: false, Description: "Required if status is 'special'."},
			{Name: "index", Type: core.ArgTypeInt, Required: false, Description: "Insertion index in parent's children (-1 or omitted to append)."},
		},
		ReturnType: core.ArgTypeString,
	},
	Func: toolChecklistAddItem,
}

var toolChecklistRemoveItemImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "ChecklistRemoveItem", // Args, Desc, ReturnType...
		Description: "Removes a checklist item node (and all its descendants) from the tree. Does NOT automatically update parent statuses; call Checklist.UpdateStatus explicitly afterwards. Returns nil on success.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."},
			{Name: "node_id", Type: core.ArgTypeString, Required: true, Description: "The unique ID of the checklist item node to remove. Cannot be the root node."},
		},
		ReturnType: core.ArgTypeNil,
	},
	Func: toolChecklistRemoveItem,
}

var toolChecklistUpdateStatusImpl = core.ToolImplementation{
	Spec: core.ToolSpec{Name: "Checklist.UpdateStatus", // Args, Desc, ReturnType...
		Description: "Recursively updates the status of all automatic checklist items based on their children's current statuses.",
		Args:        []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "The GenericTree handle ID for the checklist."}},
		ReturnType:  core.ArgTypeNil,
	},
	Func: toolChecklistUpdateStatus,
}

// --- Registration ---

const ToolsetName = "Checklist"

func init() {
	// <<< FIX: Assign function directly, signatures now match >>>
	toolsets.AddToolsetRegistration(ToolsetName, RegisterChecklistTools)
	fmt.Println("Checklist package init() running...") // Simple log to confirm init
}

// RegisterChecklistTools registers all tools in this package with the interpreter.
// <<< FIX: Add back 'error' return to match toolsets.ToolRegisterFunc >>>
func RegisterChecklistTools(registry core.ToolRegistrar) error {
	fmt.Println("Checklist tools registered via RegisterChecklistTools.") // Simple log
	if registry == nil {
		return errors.New("registry cannot be nil for RegisterChecklistTools")
	}

	tools := []core.ToolImplementation{
		toolChecklistLoadTreeImpl,
		toolChecklistFormatTreeImpl,
		toolChecklistSetItemStatusImpl,
		toolChecklistSetItemTextImpl,
		toolChecklistAddItemImpl,
		toolChecklistRemoveItemImpl,
		toolChecklistUpdateStatusImpl,
	}

	var registrationErrors []error
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			// <<< FIX: Collect errors instead of panicking >>>
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register tool %q in toolset %q: %w", tool.Spec.Name, ToolsetName, err))
		}
	}
	// <<< FIX: Return joined errors (or nil if none) >>>
	return errors.Join(registrationErrors...)
}

// --- Tool Implementations ---

// toolChecklistSetItemStatus implementation (remains here for now)
func toolChecklistSetItemStatus(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemStatus"
	logger := interpreter.Logger()
	// 1. Validate Arguments
	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("%w: %s expected 3 or 4 arguments (handle, nodeId, newStatus, [specialSymbol]), got %d", core.ErrValidationArgCount, toolName, len(args))
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
		return nil, fmt.Errorf("%w: %s 'nodeId' cannot be empty", core.ErrValidationRequiredArgNil, toolName)
	}
	newStatus, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newStatus', got %T", core.ErrValidationTypeMismatch, toolName, args[2])
	}
	if _, ok := allowedStatuses[newStatus]; !ok {
		return nil, fmt.Errorf("%w: %s invalid value for 'newStatus': %q", core.ErrInvalidArgument, toolName, newStatus)
	}
	specialSymbol := ""
	if len(args) == 4 && args[3] != nil {
		symbolStr, ok := args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s expected string or null arg[3] 'specialSymbol', got %T", core.ErrValidationTypeMismatch, toolName, args[3])
		}
		specialSymbol = symbolStr
	}
	if newStatus == "special" {
		if specialSymbol == "" {
			return nil, fmt.Errorf("%w: %s 'newStatus' is 'special' but 'specialSymbol' was not provided or is empty", core.ErrInvalidArgument, toolName)
		}
		if utf8.RuneCountInString(specialSymbol) != 1 {
			return nil, fmt.Errorf("%w: %s 'specialSymbol' must be a single character, got %q", core.ErrInvalidArgument, toolName, specialSymbol)
		}
	}

	// 2. Get Node (using direct access)
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", core.ErrNotFound, toolName, nodeID, handleID)
	}
	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node ID %q is type %q, expected 'checklist_item'", core.ErrInvalidArgument, toolName, nodeID, targetNode.Type)
	}
	if targetNode.Attributes == nil {
		targetNode.Attributes = make(map[string]string)
	}

	// 3. Update Attributes
	oldStatus := targetNode.Attributes["status"]
	targetNode.Attributes["status"] = newStatus
	if newStatus == "special" {
		targetNode.Attributes["special_symbol"] = specialSymbol
	} else {
		delete(targetNode.Attributes, "special_symbol")
	}

	logger.Debug("Manual node status updated", "tool", toolName, "nodeId", nodeID, "oldStatus", oldStatus, "newStatus", newStatus)
	return nil, nil // Success
}

// Implementations for LoadTree, FormatTree, SetItemText, AddItem, RemoveItem, UpdateStatus
// should reside in their respective files (e.g., checklist_tool_load.go, checklist_tool2.go, etc.)
