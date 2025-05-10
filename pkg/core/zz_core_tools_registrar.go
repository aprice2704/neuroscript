// NeuroScript Version: 0.3.8
// File version: 0.1.4 // Add registration for metaToolsToRegister
// Central registrar for a bundle of core tools.
// filename: pkg/core/zz_core_tools_registrar.go

package core

import (
	"log" // Standard Go logging package
	"strings"
	// "fmt" // Only needed if error wrapping is used below
)

// init calls the main registration function for the tool bundle.
// This ensures they are added to the global tool list when the core package is initialized.
func init() {
	registerCoreToolBundle()
}

// MakeUnimplementedToolFunc remains the same as provided in your example.
func MakeUnimplementedToolFunc(toolName string) ToolFunc {
	// ... (implementation as in your provided zz_core_tools_registrar.go) ...
	// For brevity, assuming it's the same.
	// If you need it explicitly, I can add it back.
	// It's not directly used by the changes I'm making here but is part of the file.
	return func(interpreter *Interpreter, args []interface{}) (interface{}, error) {
		errMsg := "TOOL " + toolName + " NOT IMPLEMENTED"
		log.Printf("[ERROR] %s\n", errMsg)
		return nil, NewRuntimeError(ErrorCodeNotImplemented, errMsg, ErrNotImplemented)
	}
}

// registerCoreToolBundle defines and registers a collection of core tools.
func registerCoreToolBundle() {
	var toolsToRegister []ToolImplementation

	// Append existing tool groups
	toolsToRegister = append(toolsToRegister, goToolsToRegister...)      // from tooldefs_go.go
	toolsToRegister = append(toolsToRegister, fsToolsToRegister...)      // from tooldefs_fs.go
	toolsToRegister = append(toolsToRegister, gitToolsToRegister...)     // from tooldefs_git.go
	toolsToRegister = append(toolsToRegister, aiWmToolsToRegister...)    // from tooldefs_ai_wm.go (or similar)
	toolsToRegister = append(toolsToRegister, ioToolsToRegister...)      // from tooldefs_io.go
	toolsToRegister = append(toolsToRegister, shellToolsToRegister...)   // from tooldefs_shell.go
	toolsToRegister = append(toolsToRegister, listToolsToRegister...)    // from tooldefs_list.go
	toolsToRegister = append(toolsToRegister, mathToolsToRegister...)    // from tooldefs_math.go
	toolsToRegister = append(toolsToRegister, stringToolsToRegister...)  // from tooldefs_string.go
	toolsToRegister = append(toolsToRegister, treeToolsToRegister...)    // from tooldefs_tree.go
	toolsToRegister = append(toolsToRegister, fileApiToolsToRegister...) // from tooldefs_file_api.go
	toolsToRegister = append(toolsToRegister, metaToolsToRegister...)    // from tooldefs_meta.go

	if len(toolsToRegister) > 0 {
		AddToolImplementations(toolsToRegister...)
		log.Printf("[INFO] zz_core_tools_registrar: Added %d tools to the global registration list via bundle.\n", len(toolsToRegister))

		toolNames := make([]string, len(toolsToRegister))
		for i, tool := range toolsToRegister {
			toolNames[i] = tool.Spec.Name
		}
		log.Printf("[DEBUG] zz_core_tools_registrar: Tools added by bundle: %s\n", strings.Join(toolNames, ", "))
	} else {
		log.Printf("[INFO] zz_core_tools_registrar: No tools were specified in the bundle to register.\n")
	}
}
