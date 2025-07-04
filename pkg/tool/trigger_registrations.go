// filename: pkg/tool/zz.go
// NeuroScript Version: 0.3.8
// File version: 0.2.0
// Purpose: Simplified to only provide a global registrar entry point.

package tool

import (
	"log"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// MakeUnimplementedToolFunc creates a placeholder function for tools that are defined but not yet implemented.
func MakeUnimplementedToolFunc(toolName string) ToolFunc {
	return func(interpreter Runtime, args []interface{}) (interface{}, error) {
		errMsg := "TOOL " + toolName + " NOT IMPLEMENTED"
		log.Printf("[ERROR] %s\n", errMsg)
		return nil, lang.NewRuntimeError(lang.ErrorCodeNotImplemented, errMsg, lang.ErrNotImplemented)
	}
}

// NOTE: The registerCoreToolBundle function and the init() that calls it have been removed.
// The new pattern relies on individual tool packages registering themselves via their own init() functions.
// The interpreter will now call RegisterExtendedTools to complete the process.
