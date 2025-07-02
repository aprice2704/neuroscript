// NeuroScript Version: 0.3.0
// File version: 0.1.1 // Changed INFO logs to DEBUG
// AI Worker Management: Tool Registration (currently handles AIWM initialization)
// filename: pkg/core/ai_wm_tools_register.go

package core

import (
	"fmt"

	"github.com/aprice2704/neuroscript/pkg/lang"
	// "github.com/aprice2704/neuroscript/pkg/logging" // Not directly needed here
)

// RegisterAIWorkerTools ensures the AIWorkerManager is initialized on the interpreter.
// If AI Worker tools were to be registered by this specific function (distinct from tooldefs_ai_wm.go),
// that logic would also go here. Currently, its main role is AIWM setup.
func RegisterAIWorkerTools(i *neurogo.Interpreter) error {
	if i == nil {
		return fmt.Errorf("interpreter cannot be nil for RegisterAIWorkerTools")
	}
	registry := i.ToolRegistry()
	if registry == nil {
		// This check might be redundant if i.ToolRegistry() can't return nil,
		// but kept for safety based on original code.
		return fmt.Errorf("ToolRegistry is nil in Interpreter for RegisterAIWorkerTools")
	}

	if i.aiWorkerManager == nil {
		i.Logger().Debugf("AIWorkerManager not yet initialized on Interpreter, attempting to create now for AI Worker tools.") // Changed from Infof
		sandboxDir := i.SandboxDir()
		if sandboxDir == "" {
			return lang.NewRuntimeError(lang.ErrorCodeConfiguration, "cannot initialize AIWorkerManager: Interpreter's sandbox directory is empty", lang.ErrConfiguration)
		}
		if i.llmClient == nil {
			i.Logger().Error("AIWorkerManager initialization failed: Interpreter's LLMClient is nil.")
			return lang.NewRuntimeError(lang.ErrorCodeConfiguration, "cannot initialize AIWorkerManager: Interpreter's LLMClient is nil", lang.ErrConfiguration)
		}

		// Corrected call to NewAIWorkerManager with all required arguments
		// Pass empty strings for initial content, so it defaults to loading from files if they exist, or starts empty.
		manager, managerErr := NewAIWorkerManager(i.Logger(), sandboxDir, i.llmClient, "", "")
		if managerErr != nil {
			if _, ok := managerErr.(*lang.RuntimeError); !ok { // Ensure it's a lang.RuntimeError
				managerErr = lang.NewRuntimeError(lang.ErrorCodeInternal, "failed to initialize AIWorkerManager for tools", managerErr)
			}
			return managerErr
		}
		i.SetAIWorkerManager(manager)                                                                               // Assumes Interpreter has SetAIWorkerManager
		i.Logger().Debugf("AIWorkerManager initialized and set on Interpreter during AI Worker tool registration.") // Changed from Infof
	} else {
		i.Logger().Debugf("AIWorkerManager already initialized on Interpreter for AI Worker tools.") // Changed from Infof
	}
	return nil
}
