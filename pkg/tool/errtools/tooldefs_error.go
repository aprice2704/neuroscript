// NeuroScript Version: 0.4.1
// File version: 1
// Purpose: Defines the ToolImplementation slice for the core Error tool.
// filename: pkg/tool/errtools/tooldefs_error.go
// nlines: 25
// risk_rating: LOW

package errtools

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// errorToolsToRegister contains the ToolImplementation definitions for Error tools.
var errorToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Error.New",
			Description: "Constructs a standard NeuroScript error value map.",
			Category:    "Error Handling",
			Args: []tool.ArgSpec{
				{Name: "code", Type: tool.ArgTypeAny, Required: true, Description: "A string or integer error code."},
				{Name: "message", Type: tool.ArgTypeString, Required: true, Description: "A human-readable error message."},
			},
			ReturnType:      "error",
			ReturnHelp:      "Returns an 'error' type value, which is a map containing 'code' and 'message'.",
			Example:         `set file_err = tool.Error.New("ERR_NOT_FOUND", "The specified file does not exist.")`,
			ErrorConditions: "Returns an error if the argument count is wrong or if arguments have invalid types.",
		},
		Func: toolErrorNew,
	},
}
