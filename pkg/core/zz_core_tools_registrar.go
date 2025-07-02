// NeuroScript Version: 0.3.8
// File version: 0.1.6 // Remove all bootstrap log.Printf INFO messages.
// Central registrar for a bundle of core tools.
// filename: pkg/core/zz_core_tools_registrar.go

package core

import (
	"log"	// Standard Go logging package

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// init calls the main registration function for the tool bundle.
// This ensures they are added to the global tool list when the core package is initialized.
func init() {
	registerCoreToolBundle()
}

// MakeUnimplementedToolFunc remains the same.
func MakeUnimplementedToolFunc(toolName string) ToolFunc {
	return func(interpreter *neurogo.Interpreter, args []interface{}) (interface{}, error) {
		errMsg := "TOOL " + toolName + " NOT IMPLEMENTED"
		log.Printf("[ERROR] %s\n", errMsg)	// Standard log for critical missing piece
		return nil, lang.NewRuntimeError(lang.ErrorCodeNotImplemented, errMsg, lang.ErrNotImplemented)
	}
}

// registerCoreToolBundle defines and registers a collection of core tools.
func registerCoreToolBundle() {
	var toolsToRegister []ToolImplementation

	// Append existing tool groups
	toolsToRegister = append(toolsToRegister, gotools.goToolsToRegister...)
	toolsToRegister = append(toolsToRegister, fs.fsToolsToRegister...)
	toolsToRegister = append(toolsToRegister, git.gitToolsToRegister...)
	toolsToRegister = append(toolsToRegister, ai.aiWmToolsToRegister...)
	toolsToRegister = append(toolsToRegister, io.ioToolsToRegister...)
	toolsToRegister = append(toolsToRegister, shell.shellToolsToRegister...)
	toolsToRegister = append(toolsToRegister, list.listToolsToRegister...)
	toolsToRegister = append(toolsToRegister, math.mathToolsToRegister...)
	toolsToRegister = append(toolsToRegister, strtools.stringToolsToRegister...)
	toolsToRegister = append(toolsToRegister, tree.treeToolsToRegister...)
	toolsToRegister = append(toolsToRegister, fileapi.fileApiToolsToRegister...)
	toolsToRegister = append(toolsToRegister, meta.metaToolsToRegister...)
	toolsToRegister = append(toolsToRegister, syntax.syntaxToolsToRegister...)
	toolsToRegister = append(toolsToRegister, time.timeToolsToRegister...)
	toolsToRegister = append(toolsToRegister, errtools.errorToolsToRegister...)
	toolsToRegister = append(toolsToRegister, script.scriptToolsToRegister...)

	if len(toolsToRegister) > 0 {
		AddToolImplementations(toolsToRegister...)
		// REMOVED: log.Printf("[INFO] zz_core_tools_registrar: Added %d tools to the global registration list via bundle.\n", len(toolsToRegister))
	} else {
		// REMOVED: log.Printf("[INFO] zz_core_tools_registrar: No tools were specified in the bundle to register.\n")
	}
}