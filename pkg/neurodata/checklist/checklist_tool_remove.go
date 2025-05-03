// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 15:38:29 PM PDT
// filename: pkg/neurodata/checklist/checklist_tool_remove.go
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// Implementation for ChecklistRemoveItem
func toolChecklistRemoveItem(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistRemoveItem"
	logger := interpreter.Logger()

	// 1. Validate Arguments
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: %s expected 2 arguments (handle, nodeId), got %d", core.ErrValidationArgCount, toolName, len(args))
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

	// 2. Get Tree (needed to check if removing root)
	treeObj, err := interpreter.GetHandleValue(handleID, core.GenericTreeHandleType)
	if err != nil {
		return nil, fmt.Errorf("%s failed getting handle %q: %w", toolName, handleID, err) // Let GetHandleValue handle specifics
	}
	tree, ok := treeObj.(*core.GenericTree)
	if !ok || tree == nil || tree.NodeMap == nil || tree.RootID == "" {
		return nil, fmt.Errorf("%w: %s handle %q did not contain a valid or initialized GenericTree", core.ErrHandleInvalid, toolName, handleID)
	}

	// 3. Prevent removing the root node
	if nodeID == tree.RootID {
		// Use ErrCannotRemoveRoot directly if appropriate, otherwise wrap InvalidArgument
		// Let's wrap InvalidArgument as it's a user-provided ID constraint violation
		return nil, fmt.Errorf("%w: %s cannot remove the root node ('%s') of the checklist tree", core.ErrInvalidArgument, toolName, nodeID)
	}

	// 4. Check if node exists before attempting removal (TreeRemoveNode might do this, but explicit check is clearer)
	targetNode, exists := tree.NodeMap[nodeID]
	if !exists {
		return nil, fmt.Errorf("%w: %s node ID %q not found in tree handle %q", core.ErrNotFound, toolName, nodeID, handleID)
	}
	// Optional: Check if it's actually a checklist_item? TreeRemoveNode doesn't care, but maybe we should?
	if targetNode.Type != "checklist_item" {
		logger.Warn("Attempting to remove a non-checklist_item node from checklist tree", "tool", toolName, "nodeId", nodeID, "nodeType", targetNode.Type)
		// Proceed anyway, TreeRemoveNode will handle it generically.
	}

	// 5. Call the Core TreeRemoveNode Tool
	// We need to get the tool implementation from the registry
	removeToolImpl, found := interpreter.ToolRegistry().GetTool("TreeRemoveNode")
	if !found || removeToolImpl.Func == nil {
		logger.Error("Core tool 'TreeRemoveNode' not found in registry", "tool", toolName)
		return nil, fmt.Errorf("%w: %s requires core tool 'TreeRemoveNode' which was not found", core.ErrInternal, toolName)
	}

	logger.Debug("Calling TreeRemoveNode to remove item", "tool", toolName, "handle", handleID, "nodeId", nodeID)
	_, removeErr := removeToolImpl.Func(interpreter, core.MakeArgs(handleID, nodeID))

	// 6. Handle errors from TreeRemoveNode
	if removeErr != nil {
		// Wrap errors from the core tool
		logger.Error("TreeRemoveNode failed", "tool", toolName, "handle", handleID, "nodeId", nodeID, "error", removeErr)
		// Map specific core errors if needed, otherwise wrap as internal or invalid argument
		if errors.Is(removeErr, core.ErrNotFound) {
			// Should have been caught above, but handle defensively
			return nil, fmt.Errorf("%w: %s node ID %q not found (reported by TreeRemoveNode)", core.ErrNotFound, toolName, nodeID)
		}
		if errors.Is(removeErr, core.ErrCannotRemoveRoot) {
			// Should have been caught above
			return nil, fmt.Errorf("%w: %s cannot remove root node (reported by TreeRemoveNode)", core.ErrInvalidArgument, toolName)
		}
		// Assume other errors might indicate issues with the arguments or internal state
		return nil, fmt.Errorf("%w: %s failed to remove node %q: %w", core.ErrInvalidArgument, toolName, nodeID, removeErr)
	}

	logger.Debug("Successfully removed node using TreeRemoveNode", "tool", toolName, "handle", handleID, "nodeId", nodeID)

	// 7. Trigger status update automatically? Or require explicit call?
	// Let's require an explicit call to Checklist.UpdateStatus afterwards for consistency.
	// If automatic update is desired, call toolChecklistUpdateStatus here.

	return nil, nil // Success
}
