// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-03 19:39:45 PM PDT // Remove duplicate registration of tree metadata tools
// filename: pkg/core/tools_register.go
package core

import (
	"errors"
	"fmt"
	// "strings" // No longer needed if using errors.Join
)

// registerCoreToolsInternal collects registration calls for all CORE tool groups.
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
	collectErr("Vector", registerVectorTools(registry)) // Assuming registerVectorTools exists
	collectErr("Git", registerGitTools(registry))
	collectErr("String", registerStringTools(registry))
	collectErr("Shell", registerShellTools(registry)) // Assuming registerShellTools exists
	collectErr("Go", registerGoTools(registry))
	collectErr("Math", registerMathTools(registry))
	collectErr("Metadata", registerMetadataTools(registry)) // Assuming registerMetadataTools exists (for non-tree metadata?)
	collectErr("List", registerListTools(registry))
	collectErr("IO", registerIOTools(registry))
	collectErr("File API", registerFileAPITools(registry))
	collectErr("LLM", RegisterLLMTools(registry)) // Assuming RegisterLLMTools exists

	// --- Tree Tool Registration ---
	collectErr("Tree (Core)", registerTreeTools(registry))         // Registers load, nav, find, basic modify, AND metadata
	collectErr("Tree (Render)", registerTreeRenderTools(registry)) // Registers format, render
	// <<< REMOVED redundant call to registerTreeMetadataTools(registry) >>>
	// --- End Tree Tool Registration ---

	// TODO: Clarify GoAST tool registration strategy.

	if len(allErrors) > 0 {
		return errors.Join(allErrors...) // Use errors.Join (Go 1.20+)
	}

	// Log success (assuming logger is accessible, otherwise might need adjustment)
	// If registry has a logger:
	// registry.interpreter.Logger().Info("Core tools registered successfully.")
	// Otherwise, fallback or adjust as needed:
	fmt.Println("Core tools registered successfully.")

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
		// Log the error before returning?
		// fmt.Printf("! Error during core tool registration: %v\n", err)
		return err // Propagate error
	}
	return nil // Success
}

// // Dummy registration functions for groups mentioned but potentially not defined elsewhere
// // These should be replaced with actual calls if the corresponding files exist and define them.
// func registerFsTools(registry *ToolRegistry) error {
//  fmt.Println("Registering FS tools...")
//  return nil
// }
// func registerVectorTools(registry *ToolRegistry) error {
//  fmt.Println("Registering Vector tools...")
//  return nil
// }
// func registerShellTools(registry *ToolRegistry) error {
//  fmt.Println("Registering Shell tools...")
//  return nil
// }
// func registerMetadataTools(registry *ToolRegistry) error {
// 	// This likely refers to non-tree metadata tools if they exist
// 	fmt.Println("Registering (non-tree) Metadata tools...")
// 	return nil
// }
// func RegisterLLMTools(registry *ToolRegistry) error {
//  fmt.Println("Registering LLM tools...")
//  return nil
// }

// Ensure other registration functions like registerStringTools, registerGoTools, registerMathTools etc.
// are defined in other files within the core package as indicated previously.

// Note: registerTreeRenderTools is assumed to be defined in tools_tree_render.go
// Note: registerTreeTools is assumed to be defined in tools_tree_register.go (the *other* one)
