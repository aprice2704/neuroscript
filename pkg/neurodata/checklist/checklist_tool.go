// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 01:19:00 AM PDT // Consolidate all tool var definitions here
package checklist

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// --- ADDED: init() function for self-registration ---
func init() {
	fmt.Println("Checklist package init() running...") // Debug output
	// Register the main registration function with the toolsets package.
	toolsets.AddToolsetRegistration("Checklist", RegisterChecklistTools)
}

// Allowed leaf statuses that can be set directly by the user tool
var allowedLeafStatuses = map[string]bool{
	"open":       true,
	"done":       true,
	"skipped":    true,
	"inprogress": true,
	"blocked":    true,
	"question":   true,
	"special":    true,
}

// --- Tool Definitions (ALL CONSOLIDATED HERE) ---

// ChecklistLoadTree - Parses checklist, returns GenericTree handle
var toolChecklistLoadTreeImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name:        "ChecklistLoadTree",
		Description: "Parses checklist content string into an internal generic tree structure representing the checklist. Returns a tree handle.",
		Args:        []core.ArgSpec{{Name: "content", Type: core.ArgTypeString, Required: true, Description: "The string content containing the checklist."}},
		ReturnType:  core.ArgTypeString,
	},
	Func: toolChecklistLoadTree, // Implementation in this file
}

// ChecklistSetItemStatus - Updates the status of a checklist item node
var toolChecklistSetItemStatusImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "ChecklistSetItemStatus",
		Description: "Sets the status of a specific *manual* checklist item node within a checklist tree handle. " +
			"It's recommended to use Checklist.UpdateStatus afterwards to ensure automatic parent statuses are correct.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle ID of the checklist tree."},
			{Name: "nodeId", Type: core.ArgTypeString, Required: true, Description: "ID of the target checklist_item node."},
			{Name: "newStatus", Type: core.ArgTypeString, Required: true, Description: "The desired new status (e.g., 'open', 'done', 'skipped', 'inprogress', 'blocked', 'question', 'special')."},
			{Name: "specialSymbol", Type: core.ArgTypeString, Required: false, Description: "Required only if newStatus is 'special'. The single character symbol to use."},
		},
		ReturnType: "", // Represents null/no specific return value on success
	},
	Func: toolChecklistSetItemStatus, // Implementation in this file
}

// ChecklistFormatTree - Formats a checklist tree handle back to Markdown string
// (Defined here, Implemented in checklist_tool2.go)
var toolChecklistFormatTreeImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name:        "ChecklistFormatTree",
		Description: "Formats a checklist tree (referenced by handle) back into a NeuroData Checklist Markdown string. Assumes statuses are up-to-date (run Checklist.UpdateStatus first if needed).",
		Args:        []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle ID of the checklist tree (obtained from ChecklistLoadTree)."}},
		ReturnType:  core.ArgTypeString,
	},
	Func: toolChecklistFormatTree, // Implementation in checklist_tool2.go
}

// ChecklistSetItemText - Updates the text description of a checklist item node
// (Defined here, Implemented in checklist_tool2.go)
var toolChecklistSetItemTextImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name:        "ChecklistSetItemText",
		Description: "Sets the text description (Value) of a specific checklist item node within a checklist tree handle.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle ID of the checklist tree."},
			{Name: "nodeId", Type: core.ArgTypeString, Required: true, Description: "ID of the target checklist_item node."},
			{Name: "newText", Type: core.ArgTypeString, Required: true, Description: "The new text description for the item."},
		},
		ReturnType: "", // Represents null/no specific return value on success
	},
	Func: toolChecklistSetItemText, // Implementation in checklist_tool2.go
}

// ChecklistUpdateStatus - Recalculates status for all automatic items
// (Defined here, Implemented in checklist_tool2.go)
var toolChecklistUpdateStatusImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name:        "Checklist.UpdateStatus", // Using dot notation convention
		Description: "Recursively updates the status attribute of all automatic checklist items ('| |') in the tree based on their direct children's statuses, following the rules in checklist.md.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle ID of the checklist tree to update."},
		},
		ReturnType: "", // Represents null/no specific return value on success
	},
	Func: toolChecklistUpdateStatus, // Implementation in checklist_tool2.go
}

// ChecklistAddItem - Adds a new checklist item node
// (Defined here, Implemented in checklist_tool_add.go)
var toolChecklistAddItemImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "ChecklistAddItem",
		Description: "Adds a new checklist item node as a child of a specified parent node within a checklist tree handle. " +
			"Returns the ID of the newly created node.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle ID of the checklist tree."},
			{Name: "parentId", Type: core.ArgTypeString, Required: true, Description: "ID of the parent node to add the new item under."},
			{Name: "newItemText", Type: core.ArgTypeString, Required: true, Description: "The text description for the new item."},
			{Name: "newItemStatus", Type: core.ArgTypeString, Required: false, Description: "Initial status (e.g., 'open', 'done'). Defaults to 'open' if null/omitted."},
			{Name: "isAutomatic", Type: core.ArgTypeBool, Required: false, Description: "Set to true if the new item's status should be automatically calculated. Defaults to false if null/omitted."},
			{Name: "specialSymbol", Type: core.ArgTypeString, Required: false, Description: "Required only if newItemStatus is 'special'. The single character symbol."},
			{Name: "index", Type: core.ArgTypeInt, Required: false, Description: "Optional zero-based index to insert item. Appends if omitted/null."},
		},
		ReturnType: core.ArgTypeString, // Returns the new node ID
	},
	Func: toolChecklistAddItem, // Implementation in checklist_tool_add.go
}

// --- Registration ---

// RegisterChecklistTools adds the checklist tools.
func RegisterChecklistTools(registry core.ToolRegistrar) error {
	if registry == nil {
		return fmt.Errorf("RegisterChecklistTools called with nil registry")
	}
	toolsToRegister := []core.ToolImplementation{
		toolChecklistLoadTreeImpl,      // Defined above
		toolChecklistSetItemStatusImpl, // Defined above
		toolChecklistFormatTreeImpl,    // Defined above
		toolChecklistSetItemTextImpl,   // Defined above
		toolChecklistUpdateStatusImpl,  // Defined above
		toolChecklistAddItemImpl,       // Defined above
	}

	var registrationErrors []error
	for _, tool := range toolsToRegister {
		// Ensure Spec is not nil and Name is not empty before registering
		if tool.Spec.Name == "" {
			// This check prevents the "ToolSpec.Name cannot be empty" error at runtime
			registrationErrors = append(registrationErrors, fmt.Errorf("attempted to register a tool with an empty name (check tool variable definitions)"))
			continue // Skip registration of this invalid tool
		}
		if err := registry.RegisterTool(tool); err != nil {
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register checklist tool %q: %w", tool.Spec.Name, err))
		}
	}

	if len(registrationErrors) > 0 {
		return errors.Join(registrationErrors...) // Use errors.Join (Go 1.20+)
	}
	fmt.Println("Checklist tools registered via RegisterChecklistTools.") // Debug
	return nil
}

// --- Tool Functions (Implementations for tools defined in this file) ---

// toolChecklistLoadTree implements the ChecklistLoadTree tool.
func toolChecklistLoadTree(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistLoadTree"
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (content), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0], got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}
	logger := interpreter.Logger()
	logger.Debug("Parsing checklist content", "tool", toolName)
	parsedData, parseErr := ParseChecklist(content, logger)
	if parseErr != nil {
		if errors.Is(parseErr, ErrNoContent) || errors.Is(parseErr, ErrMalformedItem) || errors.Is(parseErr, ErrScannerFailed) || errors.Is(parseErr, ErrMetadataExtraction) {
			return nil, fmt.Errorf("%w: %s parsing failed: %w", core.ErrInvalidArgument, toolName, parseErr)
		}
		return nil, fmt.Errorf("%w: %s parsing failed unexpectedly: %w", core.ErrInternalTool, toolName, parseErr)
	}
	logger.Debug("Adapting parsed checklist to GenericTree", "tool", toolName, "itemCount", len(parsedData.Items))
	tree, adaptErr := ChecklistToTree(parsedData.Items, parsedData.Metadata)
	if adaptErr != nil {
		return nil, fmt.Errorf("%w: %s failed adapting checklist: %w", core.ErrInternalTool, toolName, adaptErr)
	}
	if tree == nil {
		return nil, fmt.Errorf("%w: %s checklist adapter returned nil tree", core.ErrInternalTool, toolName)
	}
	handleID, handleErr := interpreter.RegisterHandle(tree, core.GenericTreeHandleType)
	if handleErr != nil {
		return nil, fmt.Errorf("%w: %s failed to register checklist tree handle: %w", core.ErrInternalTool, toolName, handleErr)
	}
	logger.Debug("Successfully loaded checklist into tree", "tool", toolName, "handle", handleID)
	return handleID, nil
}

// toolChecklistSetItemStatus implements the ChecklistSetItemStatus tool.
func toolChecklistSetItemStatus(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemStatus"
	logger := interpreter.Logger()
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
	newStatus, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[2] 'newStatus', got %T", core.ErrValidationTypeMismatch, toolName, args[2])
	}
	if !allowedLeafStatuses[newStatus] {
		return nil, fmt.Errorf("%w: %s invalid value for 'newStatus': %q", core.ErrInvalidArgument, toolName, newStatus)
	}
	specialSymbol := ""
	if len(args) > 3 {
		specialSymbolArg := args[3]
		if specialSymbolArg != nil {
			specialSymbol, ok = specialSymbolArg.(string)
			if !ok {
				return nil, fmt.Errorf("%w: %s expected string arg[3] 'specialSymbol', got %T", core.ErrValidationTypeMismatch, toolName, args[3])
			}
		}
	}
	if newStatus == "special" {
		if specialSymbol == "" {
			return nil, fmt.Errorf("%w: %s 'newStatus' is 'special' but 'specialSymbol' argument was not provided or is empty", core.ErrInvalidArgument, toolName)
		}
		if utf8.RuneCountInString(specialSymbol) != 1 {
			return nil, fmt.Errorf("%w: %s 'specialSymbol' must be a single character, got %q", core.ErrInvalidArgument, toolName, specialSymbol)
		}
	}
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("failed getting handle %q: %w", handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrInternalTool, toolName, handleID)
	}
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: node ID %q", core.ErrNotFound, nodeID)
	}
	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: node ID %q is type %q, expected 'checklist_item'", core.ErrInvalidArgument, nodeID, targetNode.Type)
	}
	if targetNode.Attributes == nil {
		targetNode.Attributes = make(map[string]string)
	}
	currentStatus := targetNode.Attributes["status"]
	targetNode.Attributes["status"] = newStatus
	if newStatus == "special" {
		targetNode.Attributes["special_symbol"] = specialSymbol
	} else {
		delete(targetNode.Attributes, "special_symbol")
	} // Fix: Use targetNode
	logger.Debug("Manual node status updated", "tool", toolName, "nodeId", nodeID, "oldStatus", currentStatus, "newStatus", newStatus)
	return nil, nil
}
