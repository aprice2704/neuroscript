// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 16:31:01 PM PDT // Fix RegisterHandle argument order based on user's code
// filename: pkg/neurodata/checklist/checklist_tool_load.go
package checklist

import (
	"errors"
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
)

// Implementation for ChecklistLoadTree
func toolChecklistLoadTree(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	toolName := "ChecklistLoadTree"
	logger := interpreter.Logger()

	// 1. Validate Arguments
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: %s expected 1 argument (checklist_string), got %d", core.ErrValidationArgCount, toolName, len(args))
	}
	checklistString, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: %s expected string arg[0] 'checklist_string', got %T", core.ErrValidationTypeMismatch, toolName, args[0])
	}

	// 2. Parse the Checklist String
	logger.Debug("Parsing checklist content", "tool", toolName)
	parserLogger := logger
	if parserLogger == nil {
		parserLogger = adapters.NewNoOpLogger()
	}
	parsedData, parseErr := ParseChecklist(checklistString, parserLogger)
	if parseErr != nil {
		logger.Error("Failed to parse checklist string", "tool", toolName, "error", parseErr)
		if errors.Is(parseErr, ErrNoContent) || errors.Is(parseErr, ErrMalformedItem) {
			return nil, fmt.Errorf("%w: %s parsing failed: %w", core.ErrInvalidArgument, toolName, parseErr)
		}
		return nil, fmt.Errorf("%w: %s internal parsing error: %w", core.ErrInternal, toolName, parseErr)
	}

	// 3. Adapt to GenericTree
	logger.Debug("Adapting parsed checklist to GenericTree", "tool", toolName, "itemCount", len(parsedData.Items))
	tree, adaptErr := ChecklistToTree(parsedData.Items, parsedData.Metadata)
	if adaptErr != nil {
		logger.Error("Failed to adapt checklist to GenericTree", "tool", toolName, "error", adaptErr)
		return nil, fmt.Errorf("%w: %s failed to create tree structure: %w", core.ErrInternal, toolName, adaptErr)
	}

	// 4. Register Tree Handle
	// <<< FIX: Use argument order (obj interface{}, typePrefix string) >>>
	handleID, handleErr := interpreter.RegisterHandle(tree, core.GenericTreeHandleType) // obj first, then type string
	if handleErr != nil {
		logger.Error("Failed to register GenericTree handle", "tool", toolName, "error", handleErr)
		return nil, fmt.Errorf("%w: %s failed to register tree handle: %w", core.ErrInternal, toolName, handleErr)
	}

	logger.Debug("Successfully loaded checklist into tree", "tool", toolName, "handle", handleID)
	return handleID, nil // Return the handle ID string
}
