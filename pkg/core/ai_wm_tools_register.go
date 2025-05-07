// NeuroScript Version: 0.3.0
// File version: 0.1.0
// AI Worker Management: Tool Registration
// filename: pkg/core/ai_wm_tools_register.go

package core

import (
	"fmt"
	// "github.com/aprice2704/neuroscript/pkg/logging" // Not directly needed here
)

// RegisterAIWorkerTools registers all AI Worker management tools with the interpreter.
// It also initializes the AIWorkerManager on the interpreter if not already present.
func RegisterAIWorkerTools(i *Interpreter) error {
	if i == nil {
		return fmt.Errorf("interpreter cannot be nil for RegisterAIWorkerTools")
	}
	registry := i.ToolRegistry()
	if registry == nil {
		return fmt.Errorf("ToolRegistry is nil in Interpreter for RegisterAIWorkerTools")
	}

	if i.aiWorkerManager == nil {
		i.Logger().Infof("AIWorkerManager not yet initialized on Interpreter, attempting to create now for AI Worker tools.")
		sandboxDir := i.SandboxDir()
		if sandboxDir == "" {
			return NewRuntimeError(ErrorCodeConfiguration, "cannot initialize AIWorkerManager: Interpreter's sandbox directory is empty", ErrConfiguration)
		}
		if i.llmClient == nil {
			i.Logger().Error("AIWorkerManager initialization failed: Interpreter's LLMClient is nil.")
			return NewRuntimeError(ErrorCodeConfiguration, "cannot initialize AIWorkerManager: Interpreter's LLMClient is nil", ErrConfiguration)
		}

		// Corrected call to NewAIWorkerManager with all required arguments
		// Pass empty strings for initial content, so it defaults to loading from files if they exist, or starts empty.
		manager, managerErr := NewAIWorkerManager(i.Logger(), sandboxDir, i.llmClient, "", "")
		if managerErr != nil {
			if _, ok := managerErr.(*RuntimeError); !ok { // Ensure it's a RuntimeError
				managerErr = NewRuntimeError(ErrorCodeInternal, "failed to initialize AIWorkerManager for tools", managerErr)
			}
			return managerErr
		}
		i.SetAIWorkerManager(manager) // Assumes Interpreter has SetAIWorkerManager
		i.Logger().Infof("AIWorkerManager initialized and set on Interpreter during AI Worker tool registration.")
	} else {
		i.Logger().Infof("AIWorkerManager already initialized on Interpreter for AI Worker tools.")
	}

	toolsToRegister := []ToolImplementation{
		// Definition Tools (from ai_wm_tools_definitions.go)
		toolAIWorkerDefinitionAdd,
		toolAIWorkerDefinitionGet,
		toolAIWorkerDefinitionList,
		toolAIWorkerDefinitionUpdate,
		toolAIWorkerDefinitionRemove,
		// Admin/Load-Save Tools (from ai_wm_tools_admin.go)
		toolAIWorkerDefinitionLoadAll,
		toolAIWorkerDefinitionSaveAll,
		toolAIWorkerSavePerformanceData, // Note: This tool's utility is limited in current design
		toolAIWorkerLoadPerformanceData,
		// Instance Tools (from ai_wm_tools_instances.go)
		toolAIWorkerInstanceSpawn,
		toolAIWorkerInstanceGet,
		toolAIWorkerInstanceListActive,
		toolAIWorkerInstanceRetire,
		toolAIWorkerInstanceUpdateStatus,
		toolAIWorkerInstanceUpdateTokenUsage,
		// Execution Tools (from ai_wm_tools_execution.go)
		toolAIWorkerExecuteStateless,
		// Performance Tools (from ai_wm_tools_performance.go)
		toolAIWorkerLogPerformance,
		toolAIWorkerGetPerformanceRecords,
	}

	for _, t := range toolsToRegister {
		if err := registry.RegisterTool(t); err != nil {
			// Ensure error is wrapped if not already a RuntimeError
			if _, ok := err.(*RuntimeError); !ok {
				err = NewRuntimeError(ErrorCodeInternal, fmt.Sprintf("failed to register AI Worker tool '%s'", t.Spec.Name), err)
			}
			return err
		}
		i.Logger().Debugf("Registered AI Worker tool: %s", t.Spec.Name)
	}
	i.Logger().Infof("Successfully registered %d AI Worker tools.", len(toolsToRegister))
	return nil
}
