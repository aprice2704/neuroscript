// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 21:35:11 PDT // Register Tree tools
// filename: pkg/core/tools_register.go
package core

import (
	"errors"
	"fmt"
	"strings" // Needed for Join fallback
)

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
func registerCoreTools(registry *ToolRegistry) error {
	var allErrors []error

	// Helper function to append errors
	collectErr := func(name string, err error) {
		if err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed registering %s tools: %w", name, err))
		}
	}

	// Register core tool groups
	collectErr("FS", registerFsTools(registry))
	collectErr("Vector", registerVectorTools(registry))
	collectErr("Git", registerGitTools(registry))
	collectErr("String", registerStringTools(registry))
	collectErr("Shell", registerShellTools(registry)) // Now only registers ExecuteCommand
	collectErr("Go", registerGoTools(registry))
	collectErr("Math", registerMathTools(registry))
	collectErr("Metadata", registerMetadataTools(registry))
	collectErr("List", registerListTools(registry))
	collectErr("IO", registerIOTools(registry))
	collectErr("File API", registerFileAPITools(registry))
	collectErr("LLM", RegisterLLMTools(registry))
	collectErr("Tree", registerTreeTools(registry)) // <<< ADDED Tree Tools registration

	// GoAST tools might be registered elsewhere (e.g., within goast package init or explicitly?)
	// TODO: Clarify GoAST tool registration strategy.

	if len(allErrors) > 0 {
		// Using errors.Join requires Go 1.20+
		// return errors.Join(allErrors...) // Prefer this if >= Go 1.20

		// Fallback for potentially older Go versions:
		errorMessages := make([]string, len(allErrors))
		for i, e := range allErrors {
			errorMessages[i] = e.Error()
		}
		return errors.New(strings.Join(errorMessages, "; "))
	}

	return nil // Success
}

// RegisterCoreTools initializes all tools defined within the core package.
func RegisterCoreTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}
	if err := registerCoreTools(registry); err != nil {
		return err // Propagate error
	}
	return nil // Success
}
