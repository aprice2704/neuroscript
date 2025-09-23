// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines the specification for the new 'Inspect' string formatting tool.
// filename: pkg/tool/strtools/tooldefs_string_format.go
// nlines: 32
// risk_rating: LOW

package strtools

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

var stringFormatToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Inspect",
			Group:       group,
			Description: "Returns a human-readable, truncated string representation of any variable. Useful for debugging complex data like maps and lists.",
			Category:    "String Formatting",
			Args: []tool.ArgSpec{
				{Name: "input_variable", Type: tool.ArgTypeAny, Required: true, Description: "The variable to inspect."},
				{Name: "max_length", Type: tool.ArgTypeInt, Required: false, Description: "The maximum length for strings before truncation. Default: 128."},
				{Name: "max_depth", Type: tool.ArgTypeInt, Required: false, Description: "The maximum depth to recurse into nested structures. Default: 5."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a formatted string representing the variable.",
			Example:         `str.Inspect(my_map, max_length: 64, max_depth: 2)`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if the wrong number of arguments is provided or if they have incorrect types.",
		},
		Func: toolInspect,
	},
}
