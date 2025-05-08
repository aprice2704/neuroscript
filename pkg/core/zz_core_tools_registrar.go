// NeuroScript Version: 0.3.1
// File version: 0.1.3 // Add MakeUnimplementedToolFunc factory.
// Central registrar for a bundle of core tools.
// filename: pkg/core/zz_core_tools_registrar.go

package core

import (
	"fmt"
	"log" // Standard Go logging package
	"strings"
)

// init calls the main registration function for the tool bundle.
// This ensures they are added to the global tool list when the core package is initialized.
func init() {
	registerCoreToolBundle()
}

// MakeUnimplementedToolFunc returns a ToolFunc that logs an error message
// to stderr and returns a standard "not implemented" RuntimeError.
func MakeUnimplementedToolFunc(toolName string) ToolFunc {
	return func(interpreter *Interpreter, args []interface{}) (interface{}, error) {
		errMsg := fmt.Sprintf("TOOL %s NOT IMPLEMENTED", toolName)
		// Log directly to stderr using standard log package as requested
		log.Printf("[ERROR] %s\n", errMsg)
		// Return a standard RuntimeError
		return nil, NewRuntimeError(ErrorCodeNotImplemented, errMsg, ErrNotImplemented)
	}
}

// Tool names: CamelCase. String and math tools need not have *String*Concat just Concat;
//    more specialized tools should be prepended with group and dot:
//     Git.Add, List.Concat etc.
//

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

	// --- ADDED: Appends the File API tools ---
	// fileApiToolsToRegister is defined in tooldefs_file_api.go
	toolsToRegister = append(toolsToRegister, fileApiToolsToRegister...)
	// --- END ADD ---

	// toolsToRegister = append(toolsToRegister, metadataToolsToRegister...)
	// toolsToRegister = append(toolsToRegister, vectorToolsToRegister...)

	if len(toolsToRegister) > 0 {
		// This function adds the implementations to the global list.
		// NewToolRegistry will later use this global list.
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
