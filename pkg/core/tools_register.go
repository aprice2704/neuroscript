// filename: pkg/core/tools_register.go
package core

import (
	"errors" // Needed for Join fallback
	"fmt"
	"strings" // Needed for Join fallback
)

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
func registerCoreTools(registry *ToolRegistry) error {
	var errs []error // Collect errors

	// Helper function to append errors
	collectErr := func(name string, err error) {
		if err != nil {
			errs = append(errs, fmt.Errorf("failed registering %s tools: %w", name, err))
		}
	}

	// Register core tool groups
	collectErr("FS", registerFsTools(registry))
	collectErr("Vector", registerVectorTools(registry))
	collectErr("Git", registerGitTools(registry))
	collectErr("String", registerStringTools(registry))
	collectErr("Shell", registerShellTools(registry))
	collectErr("Math", registerMathTools(registry))
	collectErr("Metadata", registerMetadataTools(registry))
	collectErr("List", registerListTools(registry))
	collectErr("IO", registerIOTools(registry))
	collectErr("File API", registerFileAPITools(registry))
	collectErr("LLM", RegisterLLMTools(registry)) // <<< CORRECTED FUNCTION NAME

	// GoAST tools might be registered elsewhere

	if len(errs) > 0 {
		errorMessages := make([]string, len(errs))
		for i, e := range errs {
			errorMessages[i] = e.Error()
		}
		return errors.New(strings.Join(errorMessages, "; "))
	}

	return nil // Success
}

// RegisterCoreTools initializes all tools defined within the core package.
// This remains the function called by the application setup (e.g., neurogo/app.go)
func RegisterCoreTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}
	if err := registerCoreTools(registry); err != nil {
		return err // Propagate error
	}
	return nil // Success
}
