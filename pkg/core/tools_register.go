// NeuroScript Version: 0.3.1
// File version: 0.1.2 // Simplified: NewToolRegistry now handles globalToolImplementations.
// filename: pkg/core/tools_register.go
package core

import (
	"fmt"
	// "log" // Standard log, if needed for any direct registrations remaining.
)

// registerCoreToolsInternal is called by the public RegisterCoreTools.
// Its primary role is now to register any core tools that *cannot* use the
// init() + AddToolImplementations pattern.
// Processing of globalToolImplementations is handled by NewToolRegistry.
func registerCoreToolsInternal(registry *ToolRegistry) error {
	// var allErrors []error // Only needed if there are direct registrations below.

	// The globalToolImplementations slice (populated by all init() functions)
	// is now processed by NewToolRegistry when the registry is created.
	// Therefore, this function should no longer iterate over globalToolImplementations.

	// --- Legacy or Non-Init Registration Section ---
	// If there are any core tools that absolutely cannot be registered via init()
	// (e.g., they require a fully initialized interpreter instance for their ToolSpec),
	// their registration would go here.
	// For example:
	// collectErr := func(name string, err error) {
	//  if err != nil {
	//      allErrors = append(allErrors, fmt.Errorf("failed registering core %s tools: %w", name, err))
	//  }
	// }
	//
	// Example: if some specific tools needed direct registration:
	// specificTool := ToolImplementation{Spec: ToolSpec{Name: "MySpecialTool", ...}, Func: toolMySpecialTool}
	// if err := registry.RegisterTool(specificTool); err != nil {
	//     allErrors = append(allErrors, fmt.Errorf("failed registering MySpecialTool: %w", err))
	// }

	// if len(allErrors) > 0 {
	//  finalError := fmt.Errorf("errors during explicit tool registration in RegisterCoreTools")
	//  for _, err := range allErrors {
	//      finalError = fmt.Errorf("%w; %w", finalError, err)
	//  }
	//  return finalError
	// }
	// --- End Legacy or Non-Init Registration Section ---

	return nil // Success, assuming init-based tools are primary and handled.
}

// RegisterCoreTools is the public entry point for ensuring core tools are available.
// With the init-based registration pattern, NewToolRegistry handles the primary
// registration from globalToolImplementations. This function's role is now minimal
// unless there are core tools that cannot use the init() pattern.
func RegisterCoreTools(registry *ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}

	// Call the internal function, which now handles only non-init registrations if any.
	if err := registerCoreToolsInternal(registry); err != nil {
		// This error would now primarily come from any direct/legacy registrations
		// made within registerCoreToolsInternal.
		// The FATAL error from tests was due to re-processing globalToolImplementations here.
		return err
	}

	// Log successful completion of this phase.
	// The interpreter's logger might be available here.
	if registry.interpreter != nil && registry.interpreter.logger != nil {
		registry.interpreter.logger.Info("RegisterCoreTools: Explicit core tool registration phase complete (most tools registered via init).")
	} else {
		// log.Println("[INFO] RegisterCoreTools: Explicit core tool registration phase complete.")
	}

	return nil // Success
}
