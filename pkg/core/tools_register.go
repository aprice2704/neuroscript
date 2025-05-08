// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Change RegisterCoreTools and internal to accept ToolRegistry interface
// nlines: 70
// risk_rating: MEDIUM
// filename: pkg/core/tools_register.go
package core

import (
	"fmt"
	// "log" // Standard log, if needed for any direct registrations remaining.
)

// registerCoreToolsInternal is called by the public RegisterCoreTools.
// It now accepts the ToolRegistry interface.
// Its role is to register any core tools that *cannot* use the
// init() + AddToolImplementations pattern.
// Processing of globalToolImplementations is handled by NewToolRegistry.
func registerCoreToolsInternal(registry ToolRegistry) error {
	// If this function needs to log, it should obtain a logger via the registry
	// interface if such a method is added to ToolRegistry, or handle logging externally.

	// --- Legacy or Non-Init Registration Section ---
	// Example:
	// specificTool := ToolImplementation{Spec: ToolSpec{Name: "MySpecialTool", ...}, Func: toolMySpecialTool}
	// if err := registry.RegisterTool(specificTool); err != nil { // Uses the interface method
	//     return fmt.Errorf("failed registering MySpecialTool: %w", err)
	// }
	// --- End Legacy or Non-Init Registration Section ---

	// For now, assume NewToolRegistry (called during Interpreter init) handles all core tool registrations
	// that use the init-time pattern. If there are other core tools needing explicit registration
	// after the registry is created, they would be registered here using registry.RegisterTool().
	return nil
}

// RegisterCoreTools is the public entry point for ensuring core tools are available.
// It accepts the ToolRegistry interface.
// With the init-based registration pattern, NewToolRegistry (called by Interpreter)
// handles the primary registration from globalToolImplementations.
func RegisterCoreTools(registry ToolRegistry) error {
	if registry == nil {
		return fmt.Errorf("RegisterCoreTools called with a nil registry")
	}

	// Call the internal function, passing the interface.
	if err := registerCoreToolsInternal(registry); err != nil {
		return err
	}

	// Logging: If the ToolRegistry interface had a Logger() method, we could use it.
	// Example:
	// if logger := registry.Logger(); logger != nil {
	//     logger.Info("RegisterCoreTools: Explicit core tool registration phase complete.")
	// }
	// For now, this specific logging is removed as the interface doesn't expose the logger directly.
	// The creation of the toolRegistryImpl within NewInterpreter already logs.

	return nil
}
