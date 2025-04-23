// filename: pkg/core/tools_register.go
package core

import (
	"errors" // Needed for Join fallback
	"fmt"
	"strings" // Needed for Join fallback
)

// registerCoreTools defines the specs for built-in tools and registers them
// by calling registration functions from specific tool files WITHIN THE CORE PACKAGE.
// UPDATED: Added call to registerGoAstPackageTools
func registerCoreTools(registry *ToolRegistry) error {
	var errs []error // Collect errors

	// Helper function to append errors
	collectErr := func(name string, err error) {
		if err != nil {
			errs = append(errs, fmt.Errorf("failed registering %s tools: %w", name, err))
		}
	}

	// Register core tool groups
	collectErr("FS", registerFsTools(registry))                       // Assumes exists in tools_fs.go (or similar)
	collectErr("Vector", registerVectorTools(registry))               // Assumes exists in tools_vector.go
	collectErr("Git", registerGitTools(registry))                     // Assumes exists in tools_git_register.go
	collectErr("String", registerStringTools(registry))               // Assumes exists in tools_string.go
	collectErr("Shell", registerShellTools(registry))                 // Assumes exists in tools_shell.go
	collectErr("Math", registerMathTools(registry))                   // Assumes exists in tools_math.go
	collectErr("Metadata", registerMetadataTools(registry))           // Assumes exists in tools_metadata.go
	collectErr("List", registerListTools(registry))                   // Assumes exists in tools_list_register.go
	collectErr("Go AST", registerGoAstTools(registry))                // Assumes exists in tools_go_ast.go (Registers basic AST tools)
	collectErr("Go AST Package", registerGoAstPackageTools(registry)) // NEW: Registers package-level refactoring tools
	collectErr("IO", registerIOTools(registry))                       // Assumes exists in tools_io.go (now includes Log)
	collectErr("File API", registerFileAPITools(registry))            // Assumes exists in tools_file_api.go
	collectErr("LLM", registerLLMTools(registry))                     // Assumes exists in llm_tools.go

	if len(errs) > 0 {
		// Combine multiple registration errors if necessary
		// Fallback for older Go versions:
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

// NOTE: The implementations for registerFsTools, registerVectorTools, etc.,
// must exist in their respective files within the core package.
