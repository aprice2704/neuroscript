// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 21:28:16 PDT // Fix compiler error: use core.ErrValidationArgCount
// pkg/neurodata/checklist/checklist_tool.go
package checklist

import (
	"errors"
	"fmt"
	"unicode/utf8"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets" // <<< ADDED import for init()
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

// --- Tool Implementations ---

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
			"It's recommended to use Checklist.UpdateStatus afterwards to ensure automatic parent statuses are correct.", // Updated description slightly
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

// NEW: ChecklistUpdateStatus - Recalculates status for all automatic items
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

// --- Registration ---

// RegisterChecklistTools adds the checklist tools.
// This function is now called via the init() mechanism.
func RegisterChecklistTools(registry core.ToolRegistrar) error {
	if registry == nil {
		return fmt.Errorf("RegisterChecklistTools called with nil registry")
	}
	toolsToRegister := []core.ToolImplementation{
		toolChecklistLoadTreeImpl,
		toolChecklistSetItemStatusImpl,
		toolChecklistFormatTreeImpl,
		toolChecklistSetItemTextImpl,
		toolChecklistUpdateStatusImpl,
		toolChecklistAddItemImpl,
	}

	var registrationErrors []error
	for _, tool := range toolsToRegister {
		// Use Spec.Name for registration key
		if err := registry.RegisterTool(tool); err != nil {
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register checklist tool %q: %w", tool.Spec.Name, err))
		}
	}

	if len(registrationErrors) > 0 {
		// Consider using errors.Join if available (Go 1.20+)
		return errors.Join(registrationErrors...)
	}
	fmt.Println("Checklist tools registered via RegisterChecklistTools.") // Debug
	return nil
}

// --- Tool Functions ---

// toolChecklistLoadTree implements the ChecklistLoadTree tool.
func toolChecklistLoadTree(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistLoadTree"
	// Argument count validation
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (content), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	content, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0], got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}
	logger := interpreter.Logger()
	logger.Debug("Parsing checklist content", "tool", toolName)
	parsedData, parseErr := ParseChecklist(content, logger) // Pass logger to parser
	if parseErr != nil {
		// Handle specific parsing errors as invalid argument vs internal error
		if errors.Is(parseErr, ErrNoContent) || errors.Is(parseErr, ErrMalformedItem) || errors.Is(parseErr, ErrScannerFailed) || errors.Is(parseErr, ErrMetadataExtraction) {
			return nil, fmt.Errorf("%w: %s parsing failed: %w", core.ErrInvalidArgument, toolName, parseErr)
		}
		// Assume other errors are internal
		return nil, fmt.Errorf("%w: %s parsing failed unexpectedly: %w", core.ErrInternalTool, toolName, parseErr)
	}
	logger.Debug("Adapting parsed checklist to GenericTree", "tool", toolName, "itemCount", len(parsedData.Items))
	tree, adaptErr := ChecklistToTree(parsedData.Items, parsedData.Metadata)
	if adaptErr != nil {
		// Adapter errors are likely internal
		return nil, fmt.Errorf("%w: %s failed adapting checklist: %w", core.ErrInternalTool, toolName, adaptErr)
	}
	if tree == nil {
		return nil, fmt.Errorf("%w: %s checklist adapter returned nil tree", core.ErrInternalTool, toolName)
	}
	handleID, handleErr := interpreter.RegisterHandle(tree, core.GenericTreeHandleType)
	if handleErr != nil {
		// Handle registration error is internal
		return nil, fmt.Errorf("%w: %s failed to register checklist tree handle: %w", core.ErrInternalTool, toolName, handleErr)
	}
	logger.Debug("Successfully loaded checklist into tree", "tool", toolName, "handle", handleID)
	return handleID, nil
}

// toolChecklistSetItemStatus implements the ChecklistSetItemStatus tool.
// NOTE: This tool NO LONGER automatically triggers recalculation.
// Use Checklist.UpdateStatus explicitly after making changes.
func toolChecklistSetItemStatus(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemStatus"
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
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
	var specialSymbolArg interface{}
	if len(args) > 3 {
		specialSymbolArg = args[3]   // Optional 4th argument
		if specialSymbolArg != nil { // Allow explicit nil/omission
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

	// --- Get Tree and Node ---
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		// Propagate errors like ErrHandleNotFound, ErrHandleWrongType, ErrInvalidArgument directly
		return nil, fmt.Errorf("failed getting handle %q: %w", handleID, err)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrInternalTool, toolName, handleID)
	}
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: node ID %q", core.ErrNotFound, nodeID) // Simplified error for ErrNotFound
	}
	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: node ID %q is type %q, expected 'checklist_item'", core.ErrInvalidArgument, nodeID, targetNode.Type)
	}
	if targetNode.Attributes == nil {
		// Initialize attributes map if it doesn't exist
		targetNode.Attributes = make(map[string]string)
	}

	// --- Update Status Attributes ---
	currentStatus := targetNode.Attributes["status"] // Get current status for logging
	targetNode.Attributes["status"] = newStatus
	if newStatus == "special" {
		targetNode.Attributes["special_symbol"] = specialSymbol
	} else {
		// Remove special symbol if status is no longer special
		delete(targetNode.Attributes, "special_symbol")
	}
	// NodeMap already points to the updated targetNode struct (it's a pointer)

	logger.Debug("Manual node status updated", "tool", toolName, "nodeId", nodeID, "oldStatus", currentStatus, "newStatus", newStatus)

	return nil, nil // Return null on success
}
