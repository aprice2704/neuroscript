// NeuroScript Version: 0.3.1
// File version: 0.2.0
// Purpose: Aligned with  TreeAttrs by updating map initialization and adding safe type assertions for attribute access.
// filename: pkg/neurodata/checklist/checklist_modify_tool.go
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/parser"
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

var toolChecklistSetItemStatusImpl = runtime.tool.ToolImplementation{
	Spec: runtime.tool.ToolSpec{
		Name: "ChecklistSetItemStatus",
		Description: "Manually sets the status of a non-automatic checklist item. " +
			"Requires a special symbol if status is 'special'. " +
			"Automatically removes the special symbol if status is not 'special'. " +
			"Returns nil on success.",
		Args: []runtime.tool.ArgSpec{
			{Name: "tree_handle", Type: parser.ArgTypeString, Required: true, Description: "Handle for the checklist tree."},
			{Name: "node_id", Type: parser.ArgTypeString, Required: true, Description: "ID of the checklist item node."},
			{Name: "new_status", Type: parser.ArgTypeString, Required: true, Description: "The new status string (e.g., 'open', 'done', 'skipped', 'inprogress', 'blocked', 'question', 'special')."},
			{Name: "special_symbol", Type: parser.ArgTypeString, Required: false, Description: "Required only if new_status is 'special'. The single character symbol."},
		},
		ReturnType: runtime.tool.ArgTypeNil,
	},
	Func: toolChecklistSetItemStatus,
}

func toolChecklistSetItemStatus(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistSetItemStatus"
	logger := interpreter.Logger()

	if len(args) < 3 || len(args) > 4 {
		return nil, fmt.Errorf("%w: %s expected 3 or 4 arguments (handle, nodeId, newStatus, [specialSymbol]), got %d", lang.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s arg[0] 'handle' must be string, got %T", lang.ErrValidationTypeMismatch, toolName, args[0])
	}
	nodeID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s arg[1] 'nodeId' must be string, got %T", lang.ErrValidationTypeMismatch, toolName, args[1])
	}
	if nodeID == "" {
		return nil, fmt.Errorf("%w: %s arg[1] 'nodeId' cannot be empty", lang.ErrValidationRequiredArgNil, toolName)
	}
	newStatus, ok := args[2].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s arg[2] 'newStatus' must be string, got %T", lang.ErrValidationTypeMismatch, toolName, args[2])
	}
	specialSymbol := ""
	if len(args) == 4 && args[3] != nil {
		specialSymbol, ok = args[3].(string)
		if !ok {
			return nil, fmt.Errorf("%w: %s arg[3] 'specialSymbol' must be string, got %T", lang.ErrValidationTypeMismatch, toolName, args[3])
		}
	}

	if _, isValid := validManualStatuses[newStatus]; !isValid {
		return nil, fmt.Errorf("%w: %s invalid 'newStatus' value %q", lang.ErrInvalidArgument, toolName, newStatus)
	}
	if newStatus == "special" && specialSymbol == "" {
		return nil, fmt.Errorf("%w: %s 'special_symbol' argument is required when 'newStatus' is 'special'", lang.ErrValidationRequiredArgNil, toolName)
	}
	if newStatus == "special" && len(specialSymbol) != 1 {
		return nil, fmt.Errorf("%w: %s 'special_symbol' must be a single character, got %q", lang.ErrInvalidArgument, toolName, specialSymbol)
	}
	if newStatus != "special" && specialSymbol != "" {
		logger.Warn("special_symbol provided but ignored as newStatus is not 'special'", "tool", toolName, "newStatus", newStatus)
		specialSymbol = ""
	}

	treeObj, getHandleErr := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if getHandleErr != nil {
		return nil, fmt.Errorf("%s getting handle %q failed: %w", toolName, handleID, getHandleErr)
	}
	tree, ok := treeObj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid GenericTree", lang.ErrHandleInvalid, toolName, handleID)
	}
	node, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", lang.ErrNotFound, toolName, nodeID, handleID)
	}

	if node.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node %q has type %q, expected 'checklist_item'", lang.ErrInvalidArgument, toolName, nodeID, node.Type)
	}

	if node.Attributes == nil {
		// FIX: Use correct type for attribute map initialization.
		node.Attributes = make(utils.TreeAttrs)
		logger.Warn("Node attributes map was nil on read, initialized.", "tool", toolName, "nodeId", nodeID)
	}

	// FIX: Safely check the 'is_automatic' attribute, which is now interface{}.
	var isAutomatic bool
	if autoVal, ok := node.Attributes["is_automatic"]; ok {
		if autoBool, isBool := autoVal.(bool); isBool {
			isAutomatic = autoBool
		} else if autoStr, isStr := autoVal.(string); isStr {
			isAutomatic = (autoStr == "true")
		}
	}

	if isAutomatic {
		return nil, fmt.Errorf("%w: %s cannot manually set status on automatic node %q", lang.ErrInvalidArgument, toolName, nodeID)
	}

	setMetaToolImpl, foundSet := interpreter.ToolRegistry().GetTool("Tree.SetNodeMetadata")
	if !foundSet || setMetaToolImpl.Func == nil {
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.SetNodeMetadata' which was not found", lang.ErrInternal, toolName)
	}
	removeMetaToolImpl, foundRemove := interpreter.ToolRegistry().GetTool("Tree.RemoveNodeMetadata")
	if !foundRemove || removeMetaToolImpl.Func == nil {
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.RemoveNodeMetadata' which was not found", lang.ErrInternal, toolName)
	}

	logger.Debug("Calling Tree.SetNodeMetadata for status", "tool", toolName, "nodeId", nodeID, "status", newStatus)
	setStatusArgs := runtime.tool.MakeArgs(handleID, nodeID, "status", newStatus)
	_, err := setMetaToolImpl.Func(interpreter, setStatusArgs)
	if err != nil {
		logger.Error("Tree.SetNodeMetadata failed for status", "tool", toolName, "nodeId", nodeID, "error", err)
		if errors.Is(err, lang.ErrNotFound) || errors.Is(err, lang.ErrInvalidArgument) || errors.Is(err, lang.ErrHandleInvalid) {
			return nil, fmt.Errorf("%w: %s setting status failed: %w", lang.ErrInvalidArgument, toolName, err)
		}
		return nil, fmt.Errorf("%w: %s internal error setting status: %w", lang.ErrInternal, toolName, err)
	}

	if newStatus == "special" {
		logger.Debug("Calling Tree.SetNodeMetadata for special_symbol", "tool", toolName, "nodeId", nodeID, "symbol", specialSymbol)
		setSymbolArgs := runtime.tool.MakeArgs(handleID, nodeID, "special_symbol", specialSymbol)
		_, err = setMetaToolImpl.Func(interpreter, setSymbolArgs)
		if err != nil {
			logger.Error("Tree.SetNodeMetadata failed for special_symbol", "tool", toolName, "nodeId", nodeID, "error", err)
			if errors.Is(err, lang.ErrNotFound) || errors.Is(err, lang.ErrInvalidArgument) || errors.Is(err, lang.ErrHandleInvalid) {
				return nil, fmt.Errorf("%w: %s setting special_symbol failed: %w", lang.ErrInvalidArgument, toolName, err)
			}
			return nil, fmt.Errorf("%w: %s internal error setting special_symbol: %w", lang.ErrInternal, toolName, err)
		}
	} else {
		logger.Debug("Calling Tree.RemoveNodeMetadata for special_symbol if it exists", "tool", toolName, "nodeId", nodeID)
		removeSymbolArgs := runtime.tool.MakeArgs(handleID, nodeID, "special_symbol")
		_, err = removeMetaToolImpl.Func(interpreter, removeSymbolArgs)
		if err != nil && !errors.Is(err, lang.ErrAttributeNotFound) {
			logger.Error("Tree.RemoveNodeMetadata failed for special_symbol", "tool", toolName, "nodeId", nodeID, "error", err)
			if errors.Is(err, lang.ErrNotFound) || errors.Is(err, lang.ErrInvalidArgument) || errors.Is(err, lang.ErrHandleInvalid) {
				return nil, fmt.Errorf("%w: %s removing special_symbol failed: %w", lang.ErrInvalidArgument, toolName, err)
			}
			return nil, fmt.Errorf("%w: %s internal error removing special_symbol: %w", lang.ErrInternal, toolName, err)
		}
		if err == nil {
			logger.Debug("Removed existing special_symbol attribute", "tool", toolName, "nodeId", nodeID)
		} else {
			logger.Debug("No special_symbol attribute existed to remove or TreeRemoveNodeMetadata call failed gracefully.", "tool", toolName, "nodeId", nodeID)
		}
	}

	logger.Debug("Checklist item status updated successfully", "tool", toolName, "nodeId", nodeID, "newStatus", newStatus)
	return nil, nil
}

func registerChecklistModifyTools(registry runtime.tool.ToolRegistry) error {
	if registry == nil {
		return errors.New("registry cannot be nil for registerChecklistModifyTools")
	}
	tool := toolChecklistSetItemStatusImpl
	if err := registry.RegisterTool(tool); err != nil {
		return fmt.Errorf("failed to register checklist tool %q: %w", tool.Spec.Name, err)
	}
	return nil
}
