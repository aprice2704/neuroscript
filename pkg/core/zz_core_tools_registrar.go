// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Added INFO and DEBUG logging for registered tools
// Central registrar for a bundle of core tools.
// filename: pkg/core/zz_core_tools_registrar.go

package core

import (
	"log" // Standard Go logging package
	"strings"
)

// init calls the main registration function for the tool bundle.
// This ensures they are added to the global tool list when the core package is initialized.
func init() {
	registerCoreToolBundle()
}

// registerCoreToolBundle defines and registers a collection of core tools.
// This function is intended to quickly register tools that are being migrated
// to the init-based AddToolImplementations pattern or were previously unhandled.
func registerCoreToolBundle() {
	// Concatenate tool definitions from various tooldefs_*.go files
	// Ensure these variables (goToolsToRegister, fstoolsToRegister, etc.)
	// are defined in their respective tooldefs_*.go files within the core package.
	var toolsToRegister []ToolImplementation
	toolsToRegister = append(toolsToRegister, goToolsToRegister...)
	toolsToRegister = append(toolsToRegister, fsToolsToRegister...)
	toolsToRegister = append(toolsToRegister, gitToolsToRegister...)
	toolsToRegister = append(toolsToRegister, aiWmToolsToRegister...)
	toolsToRegister = append(toolsToRegister, ioToolsToRegister...)
	toolsToRegister = append(toolsToRegister, shellToolsToRegister...)
	toolsToRegister = append(toolsToRegister, listToolsToRegister...)
	toolsToRegister = append(toolsToRegister, mathToolsToRegister...)
	toolsToRegister = append(toolsToRegister, stringToolsToRegister...)
	toolsToRegister = append(toolsToRegister, treeToolsToRegister...)

	// toolsToRegister = append(toolsToRegister, metadataToolsToRegister...)
	// toolsToRegister = append(toolsToRegister, vectorToolsToRegister...)

	if len(toolsToRegister) > 0 {
		AddToolImplementations(toolsToRegister...)

		// Log the number of tools registered by this bundle
		log.Printf("[INFO] zz_core_tools_registrar: Added %d tools to the global registration list via bundle.\n", len(toolsToRegister))

		// Collect tool names for debug logging
		toolNames := make([]string, len(toolsToRegister))
		for i, tool := range toolsToRegister {
			toolNames[i] = tool.Spec.Name
		}
		// Log the list of tool names (conceptually DEBUG level)
		log.Printf("[DEBUG] zz_core_tools_registrar: Tools added by bundle: %s\n", strings.Join(toolNames, ", "))
	} else {
		log.Printf("[INFO] zz_core_tools_registrar: No tools were specified in the bundle to register.\n")
	}
}
