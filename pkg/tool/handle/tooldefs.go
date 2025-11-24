// NeuroScript Version: 0.8.0
// File version: 1
// Purpose: Defines specifications for handle inspection tools.
// filename: pkg/tool/handle/tooldefs.go
// nlines: 45
// risk_rating: LOW

package handle

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "handle"

// handleToolsToRegister contains ToolImplementation definitions for Handle tools.
var handleToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Type",
			Group:       group,
			Description: "Returns the kind/type tag of an opaque handle (e.g., 'fsmeta', 'overlaymeta').",
			Category:    "Handle Inspection",
			Args: []tool.ArgSpec{
				{Name: "h", Type: tool.ArgTypeHandle, Required: true, Description: "The handle to inspect."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "The string identifier for the handle's kind.",
			Example:         `if handle.Type(h) == "fsmeta" { ... }`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the argument is not a handle.",
		},
		Func: toolHandleType,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "IsValid",
			Group:       group,
			Description: "Checks if a handle is valid (exists in the active registry).",
			Category:    "Handle Inspection",
			Args: []tool.ArgSpec{
				{Name: "h", Type: tool.ArgTypeHandle, Required: true, Description: "The handle to check."},
			},
			ReturnType:      tool.ArgTypeBool,
			ReturnHelp:      "Returns true if the handle points to a live object, false otherwise.",
			Example:         `if handle.IsValid(h) { ... }`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the argument is not a handle.",
		},
		Func: toolHandleIsValid,
	},
}
