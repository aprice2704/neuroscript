// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Use init()-based registration list, remove direct sub-package imports/calls.
// filename: pkg/core/tools_register.go
package core

import (
	"fmt"
	// No longer import sub-packages like gosemantic here
	// "github.com/aprice2704/neuroscript/pkg/core/tools/gosemantic"
	// "github.com/aprice2704/neuroscript/pkg/core/tools/goast" // Example if goast had register func
)

// registerCoreToolsInternal registers all tools added via AddToolImplementations.
// It's called by the public RegisterCoreTools.
func registerCoreToolsInternal(registry *ToolRegistry) error {
	var allErrors []error

	// Get the list of tools registered via init()
	// Lock isn't strictly needed here if AddToolImplementations is only called during init,
	// but it's safer if there's any chance of concurrent access later.
	globalRegMutex.Lock()
	implementationsToRegister := make([]ToolImplementation, len(globalToolImplementations))
	copy(implementationsToRegister, globalToolImplementations)
	globalRegMutex.Unlock()

	// Register each tool
	registeredNames := make(map[string]string) // Track names to report duplicates clearly
	for _, impl := range implementationsToRegister {
		toolName := impl.Spec.Name
		if existingPkg, exists := registeredNames[toolName]; exists {
			// This indicates two different packages tried to register the same tool name via init()
			// Or the same package added it twice.
			err := fmt.Errorf("duplicate tool name registration detected via init(): '%s' already registered (potentially by package providing %s)", toolName, existingPkg)
			allErrors = append(allErrors, err)
			// Skip registering the duplicate
			continue
		}

		if err := registry.RegisterTool(impl); err != nil {
			allErrors = append(allErrors, fmt.Errorf("failed registering tool '%s' from global list: %w", toolName, err))
		} else {
			// Record successful registration (value doesn't matter much here, maybe store package later?)
			registeredNames[toolName] = toolName
		}
	}

	// --- Legacy Registration (Example - Keep if some tools are still registered directly within core) ---
	// If some tool groups are still defined *entirely within core* and use the old pattern:
	// collectErr := func(name string, err error) {
	// 	if err != nil {
	// 		allErrors = append(allErrors, fmt.Errorf("failed registering core %s tools: %w", name, err))
	// 	}
	// }
	// collectErr("String", registerStringTools(registry)) // Assuming this only registers tools defined *in* core
	// collectErr("Math", registerMathTools(registry))     // Assuming this only registers tools defined *in* core
	// ... etc for FS, Vector, Git, Shell, Metadata, List, IO, FileAPI, LLM, Tree IF they are defined in core.
	// --- End Legacy Registration ---

	if len(allErrors) > 0 {
		// Use errors.Join for better formatting if Go >= 1.20
		// return errors.Join(allErrors...)
		// Fallback for broader compatibility:
		finalError := fmt.Errorf("errors during tool registration")
		for _, err := range allErrors {
			finalError = fmt.Errorf("%w; %w", finalError, err)
		}
		return finalError
	}

	return nil // Success
}

// RegisterCoreTools is the public entry point for registering all core tools.
// It now relies on tools being added via AddToolImplementations in their respective packages' init() functions.
func RegisterCoreTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}
	// Call the internal function that processes the globally registered tools
	if err := registerCoreToolsInternal(registry); err != nil {
		return err // Propagate error
	}

	// Log success (assuming logger is accessible, otherwise might need adjustment)
	// if registry.interpreter != nil && registry.interpreter.logger != nil {
	// 	registry.interpreter.logger.Info("Tools registered successfully via global list.")
	// }

	return nil // Success
}

// Remove or comment out dummy registration functions if they are not needed or if
// those tools are now registered via init() in their respective packages.
// func registerStringTools(registry *ToolRegistry) error { return nil } // Keep if string tools defined in core
// func registerMathTools(registry *ToolRegistry) error { return nil } // Keep if math tools defined in core
// ... and so on
