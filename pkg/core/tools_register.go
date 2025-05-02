// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 14:58:13 PDT // Refactor tree tool registration
// filename: pkg/core/tools_register.go
package core

import (
	"errors"
	"fmt"
	// "strings" // No longer needed if using errors.Join
)

// registerCoreTools collects registration calls for all CORE tool groups.
// Extended toolsets (checklist, blocks) are registered elsewhere (e.g., pkg/toolsets).
func registerCoreToolsInternal(registry *ToolRegistry) error {
	var allErrors []error

	// Helper function to append errors
	collectErr := func(name string, err error) {
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed registering %s tools: %w", name, err))
		}
	}

	// Register core tool groups by calling their specific registration functions
	collectErr("FS", registerFsTools(registry))
	collectErr("Vector", registerVectorTools(registry))
	collectErr("Git", registerGitTools(registry))
	collectErr("String", registerStringTools(registry))
	collectErr("Shell", registerShellTools(registry))
	collectErr("Go", registerGoTools(registry))
	collectErr("Math", registerMathTools(registry))
	collectErr("Metadata", registerMetadataTools(registry))
	collectErr("List", registerListTools(registry))
	collectErr("IO", registerIOTools(registry))
	collectErr("File API", registerFileAPITools(registry))
	collectErr("LLM", RegisterLLMTools(registry))

	// --- Updated Tree Tool Registration ---
	collectErr("Tree (Core)", registerTreeTools(registry))         // Registers load, nav, find, etc.
	collectErr("Tree (Render)", registerTreeRenderTools(registry)) // Registers format, render
	// --- End Update ---

	// TODO: Clarify GoAST tool registration strategy.

	if len(allErrors) > 0 {
		return errors.Join(allErrors...) // Use errors.Join (Go 1.20+)
	}

	return nil // Success
}

// RegisterCoreTools is the public entry point for registering all core tools.
// It mainly validates the registry and calls the internal registration logic.
func RegisterCoreTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}
	// Call the internal function that does the actual registration work
	if err := registerCoreToolsInternal(registry); err != nil {
		return err // Propagate error
	}
	return nil // Success
}
