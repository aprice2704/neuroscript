// NeuroScript Version: 0.3.0
// File version: 0.1.2
// Corrected core tool lookups for Tree.SetNodeMetadata and Tree.RemoveNodeMetadata.
// filename: pkg/neurodata/checklist/checklist_modify_tool.go
// nlines: 190 // Approximate
// risk_rating: LOW
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// Define valid manual status transitions/values here or import if defined centrally
var validManualStatuses = map[string]struct{}{
	"open":       {},
	"done":       {},
	"skipped":    {},
	"inprogress": {},
	"blocked":    {},
	"question":   {},
	"special":    {},
}

var toolChecklistSetItemStatusImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "ChecklistSetItemStatus",
		Description: "Manually sets the status of a non-automatic checklist item. " +
			"Requires a special symbol if status is 'special'. " +
			"Automatically removes the special symbol if status is not 'special'. " +
			"Returns nil on success.",
		Args: []core.ArgSpec{
			{Name: "tree_handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the checklist tree."},
			{Name: "node_id", Type: core.ArgTypeString, Required: true, Description: "ID of the checklist item node."},
			{Name: "new_status", Type: core.ArgTypeString, Required: true, Description: "The new status string (e.g., 'open', 'done', 'skipped', 'inprogress', 'blocked', 'question', 'special')."},
			{Name: "special_symbol", Type: core.ArgTypeString, Required: false, Description: "Required only if new_status is 'special'. The single character symbol."},
		},
		ReturnType: core.ArgTypeNil,
	},
	Func: toolChecklistSetItemStatus,
}

func toolChecklistSetItemStatus(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemStatus"
	logger := interpreter.Logger()

	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("%w: %s expected 3 or 4 arguments (handle, nodeId, newStatus, [specialSymbol]), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s arg[0] 'handle' must be string, got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}
	nodeID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s arg[1] 'nodeId' must be string, got %T", core.ErrValidationTypeMismatch, toolName, args[1])
	}
	if nodeID == "" {
		return nil, fmt.Errorf("%w: %s arg[1] 'nodeId' cannot be empty", core.ErrValidationRequiredArgNil, toolName)
	}
	newStatus, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s arg[2] 'newStatus' must be string, got %T", core.ErrValidationTypeMismatch, toolName, args[2])
	}
	specialSymbol := ""
	if len(args) == 4 && args[3] != nil {
		specialSymbol, ok = args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s arg[3] 'specialSymbol' must be string, got %T", core.ErrValidationTypeMismatch, toolName, args[3])
		}
	}

	if _, isValid := validManualStatuses[newStatus]; !isValid {
		return nil, fmt.Errorf("%w: %s invalid 'newStatus' value %q", core.ErrInvalidArgument, toolName, newStatus)
	}
	if newStatus == "special" && specialSymbol == "" {
		return nil, fmt.Errorf("%w: %s 'special_symbol' argument is required when 'newStatus' is 'special'", core.ErrValidationRequiredArgNil, toolName)
	}
	if newStatus == "special" && len(specialSymbol) != 1 { // Assuming UTF-8 length check might be better if symbols can be multi-byte
		return nil, fmt.Errorf("%w: %s 'special_symbol' must be a single character, got %q", core.ErrInvalidArgument, toolName, specialSymbol)
	}
	if newStatus != "special" && specialSymbol != "" {
		logger.Warn("special_symbol provided but ignored as newStatus is not 'special'", "tool", toolName, "newStatus", newStatus)
		specialSymbol = ""
	}

	treeObj, getHandleErr := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if getHandleErr != nil {
		return nil, fmt.Errorf("%s getting handle %q failed: %w", toolName, handleID, getHandleErr)
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", core.ErrNotFound, toolName, nodeID, handleID)
	}

	if node.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node %q has type %q, expected 'checklist_item'", core.ErrInvalidArgument, toolName, nodeID, node.Type)
	}

	if node.Attributes == nil {
		node.Attributes = make(map[string]string)
		logger.Warn("Node attributes map was nil on read, initialized.", "tool", toolName, "nodeId", nodeID)
	}

	if node.Attributes["is_automatic"] == "true" {
		return nil, fmt.Errorf("%w: %s cannot manually set status on automatic node %q", core.ErrInvalidArgument, toolName, nodeID)
	}

	setMetaToolImpl, foundSet := interpreter.ToolRegistry().GetTool("Tree.SetNodeMetadata") // MODIFIED
	if !foundSet || setMetaToolImpl.Func == nil {
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.SetNodeMetadata' which was not found", core.ErrInternal, toolName)
	}
	removeMetaToolImpl, foundRemove := interpreter.ToolRegistry().GetTool("Tree.RemoveNodeMetadata") // MODIFIED
	if !foundRemove || removeMetaToolImpl.Func == nil {
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.RemoveNodeMetadata' which was not found", core.ErrInternal, toolName)
	}

	logger.Debug("Calling Tree.SetNodeMetadata for status", "tool", toolName, "nodeId", nodeID, "status", newStatus)
	setStatusArgs := core.MakeArgs(handleID, nodeID, "status", newStatus)
	_, err := setMetaToolImpl.Func(interpreter, setStatusArgs)
	if err != nil {
		logger.Error("Tree.SetNodeMetadata failed for status", "tool", toolName, "nodeId", nodeID, "error", err)
		if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArgument) || errors.Is(err, core.ErrHandleInvalid) {
			return nil, fmt.Errorf("%w: %s setting status failed: %w", core.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s internal error setting status: %w", core.ErrInternal, toolName, err)
	}

	if newStatus == "special" {
		logger.Debug("Calling Tree.SetNodeMetadata for special_symbol", "tool", toolName, "nodeId", nodeID, "symbol", specialSymbol)
		setSymbolArgs := core.MakeArgs(handleID, nodeID, "special_symbol", specialSymbol)
		_, err = setMetaToolImpl.Func(interpreter, setSymbolArgs)
		if err != nil {
			logger.Error("Tree.SetNodeMetadata failed for special_symbol", "tool", toolName, "nodeId", nodeID, "error", err)
			if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArgument) || errors.Is(err, core.ErrHandleInvalid) {
				return nil, fmt.Errorf("%w: %s setting special_symbol failed: %w", core.ErrInvalidArgument, toolName, err)
			}
			return nil, fmt.Errorf("%w: %s internal error setting special_symbol: %w", core.ErrInternal, toolName, err)
		}
	} else {
		// Only attempt to remove if it might exist (i.e., node.Attributes["special_symbol"] was there)
		// However, TreeRemoveNodeMetadata handles ErrAttributeNotFound gracefully.
		logger.Debug("Calling Tree.RemoveNodeMetadata for special_symbol if it exists", "tool", toolName, "nodeId", nodeID)
		removeSymbolArgs := core.MakeArgs(handleID, nodeID, "special_symbol")
		_, err = removeMetaToolImpl.Func(interpreter, removeSymbolArgs)
		if err != nil && !errors.Is(err, core.ErrAttributeNotFound) { // Ignore ErrAttributeNotFound
			logger.Error("Tree.RemoveNodeMetadata failed for special_symbol", "tool", toolName, "nodeId", nodeID, "error", err)
			if errors.Is(err, core.ErrNotFound) || errors.Is(err, core.ErrInvalidArgument) || errors.Is(err, core.ErrHandleInvalid) {
				return nil, fmt.Errorf("%w: %s removing special_symbol failed: %w", core.ErrInvalidArgument, toolName, err)
			}
			return nil, fmt.Errorf("%w: %s internal error removing special_symbol: %w", core.ErrInternal, toolName, err)
		}
		if err == nil { // Means attribute was found and removed
			logger.Debug("Removed existing special_symbol attribute", "tool", toolName, "nodeId", nodeID)
		} else { // Means attribute was not found
			logger.Debug("No special_symbol attribute existed to remove or TreeRemoveNodeMetadata call failed gracefully.", "tool", toolName, "nodeId", nodeID)
		}
	}

	logger.Debug("Checklist item status updated successfully", "tool", toolName, "nodeId", nodeID, "newStatus", newStatus)
	return nil, nil
}

func registerChecklistModifyTools(registry core.ToolRegistry) error {
	if registry == nil {
		return errors.New("registry cannot be nil for registerChecklistModifyTools")
	}
	tool := toolChecklistSetItemStatusImpl
	if err := registry.RegisterTool(tool); err != nil {
		return fmt.Errorf("failed to register checklist tool %q: %w", tool.Spec.Name, err)
	}
	return nil
}
