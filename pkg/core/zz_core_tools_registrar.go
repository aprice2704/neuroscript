// NeuroScript Version: 0.3.8
// File version: 0.1.6 // Remove all bootstrap log.Printf INFO messages.
// Central registrar for a bundle of core tools.
// filename: pkg/core/zz_core_tools_registrar.go

package core

import (
	"log" // Standard Go logging package
)

// init calls the main registration function for the tool bundle.
// This ensures they are added to the global tool list when the core package is initialized.
func init() {
	registerCoreToolBundle()
}

// MakeUnimplementedToolFunc remains the same.
func MakeUnimplementedToolFunc(toolName string) ToolFunc {
	return func(interpreter *Interpreter, args []interface{}) (interface{}, error) {
		errMsg := "TOOL " + toolName + " NOT IMPLEMENTED"
		log.Printf("[ERROR] %s\n", errMsg) // Standard log for critical missing piece
		return nil, NewRuntimeError(ErrorCodeNotImplemented, errMsg, ErrNotImplemented)
	}
}

// registerCoreToolBundle defines and registers a collection of core tools.
func registerCoreToolBundle() {
	var toolsToRegister []ToolImplementation

	// Append existing tool groups
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
	toolsToRegister = append(toolsToRegister, fileApiToolsToRegister...)
	toolsToRegister = append(toolsToRegister, metaToolsToRegister...)
	toolsToRegister = append(toolsToRegister, syntaxToolsToRegister...)
	toolsToRegister = append(toolsToRegister, timeToolsToRegister...)
	toolsToRegister = append(toolsToRegister, errorToolsToRegister...)
	toolsToRegister = append(toolsToRegister, scriptToolsToRegister...)

	if len(toolsToRegister) > 0 {
		AddToolImplementations(toolsToRegister...)
		// REMOVED: log.Printf("[INFO] zz_core_tools_registrar: Added %d tools to the global registration list via bundle.\n", len(toolsToRegister))
	} else {
		// REMOVED: log.Printf("[INFO] zz_core_tools_registrar: No tools were specified in the bundle to register.\n")
	}
}
