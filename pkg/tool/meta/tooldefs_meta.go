// NeuroScript Version: 0.8.0
// File version: 7
// Purpose: Defines the tool specifications for the 'meta' tool group, linking them to their implementations.
// filename: pkg/tool/meta/tooldefs.go
// nlines: 55
// risk_rating: LOW

package meta

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// metaToolsToRegister holds the definitions for tools that provide information about other tools.
var metaToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "getTool",
			Group:       "meta",
			Description: "Retrieves the specification of a single registered tool by its full name.",
			Args: []tool.ArgSpec{
				{Name: "fullName", Type: tool.ArgTypeString, Description: "The full canonical name of the tool (e.g., 'tool.fs.readFile').", Required: true},
			},
			ReturnType: tool.ArgTypeMap,
			ReturnHelp: "A map containing a 'found' boolean and the tool 'spec' if found.",
		},
		Func: GetTool,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "listTools",
			Group:       "meta",
			Description: "Lists the specifications of all registered tools.",
			ReturnType:  tool.ArgTypeSlice,
			ReturnHelp:  "A list of tool specification objects.",
		},
		Func: ListTools,
	},
	{
		Spec: tool.ToolSpec{
			Name:            "getToolSpecificationsJson",
			Group:           "meta",
			Description:     "Provides a JSON string containing an array of all currently available tool specifications.",
			Category:        "Introspection",
			Args:            []tool.ArgSpec{},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "A JSON string representing an array of ToolSpec objects.",
			Example:         "getToolSpecificationsJson()",
			ErrorConditions: "Returns an error if JSON marshalling of the tool specifications fails.",
		},
		Func: GetToolSpecificationsJSON,
	},
}
