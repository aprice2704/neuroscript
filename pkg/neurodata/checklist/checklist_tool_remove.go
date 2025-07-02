// NeuroScript Version: 0.3.0
// File version: 0.1.1
// Corrected core tool lookup for Tree.RemoveNode.
// filename: pkg/neurodata/checklist/checklist_tool_remove.go
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// Implementation for ChecklistRemoveItem
func toolChecklistRemoveItem(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistRemoveItem"
	logger := interpreter.Logger()

	if len(args) != 2 {
		return nil, fmt.Errorf("%w: %s expected 2 arguments (handle, nodeId), got %d", lang.ErrValidationArgCount, toolName, len(args))
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'handle', got %T", lang.ErrValidationTypeMismatch, toolName, args[0])
	}
	nodeID, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[1] 'nodeId', got %T", lang.ErrValidationTypeMismatch, toolName, args[1])
	}
	if nodeID == "" {
		return nil, fmt.Errorf("%w: %s 'nodeId' cannot be empty", lang.ErrValidationRequiredArgNil, toolName)
	}

	treeObj, err := interpreter.GetHandleValue(handleID, utils.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err)
	}
	tree, ok := treeObj.(*GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil || tree.RootID == "" {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", lang.ErrHandleInvalid, toolName, handleID)
	}

	if nodeID == tree.RootID {
		return nil, fmt.Errorf("%w: %s cannot remove the root node ('%s') of the checklist tree", lang.ErrInvalidArgument, toolName, nodeID)
	}

	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", lang.ErrNotFound, toolName, nodeID, handleID)
	}

	if targetNode.Type != "checklist_item" {
		return nil, fmt.Errorf("%w: %s node %q has type %q, expected 'checklist_item'", lang.ErrInvalidArgument, toolName, nodeID, targetNode.Type)
	}

	removeToolImpl, found := interpreter.ToolRegistry().GetTool("Tree.RemoveNode")	// MODIFIED
	if !found || removeToolImpl.Func == nil {
		logger.Error("Core tool 'Tree.RemoveNode' not found in registry", "tool", toolName)	// MODIFIED
		return nil, fmt.Errorf("%w: %s requires core tool 'Tree.RemoveNode' which was not found", lang.ErrInternal, toolName)
	}

	logger.Debug("Calling Tree.RemoveNode to remove item", "tool", toolName, "handle", handleID, "nodeId", nodeID)
	_, removeErr := removeToolImpl.Func(interpreter, tool.MakeArgs(handleID, nodeID))

	if removeErr != nil {
		logger.Error("Tree.RemoveNode failed", "tool", toolName, "handle", handleID, "nodeId", nodeID, "error", removeErr)
		if errors.Is(removeErr, lang.ErrNotFound) {
			return nil, fmt.Errorf("%w: %s node ID %q not found (reported by Tree.RemoveNode)", lang.ErrNotFound, toolName, nodeID)
		}
		if errors.Is(removeErr, lang.ErrCannotRemoveRoot) {
			return nil, fmt.Errorf("%w: %s cannot remove root node (reported by Tree.RemoveNode)", lang.ErrInvalidArgument, toolName)
		}
		if errors.Is(removeErr, lang.ErrInternalTool) {	// Assuming lang.ErrInternalTool is a valid sentinel in your core package
			return nil, fmt.Errorf("%w: %s internal error removing node %q: %w", lang.ErrInternal, toolName, nodeID, removeErr)
		}
		return nil, fmt.Errorf("%w: %s failed to remove node %q: %w", lang.ErrInvalidArgument, toolName, nodeID, removeErr)
	}

	logger.Debug("Successfully removed node using Tree.RemoveNode", "tool", toolName, "handle", handleID, "nodeId", nodeID)
	return nil, nil
}