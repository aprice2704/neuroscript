// NeuroScript Major Version: 1
// File version: 12
// Purpose: Defines the tool specifications for the 'meta' tool group.
// Latest change: Added optional 'filter' arg to listToolNames.
// filename: pkg/tool/meta/tooldefs_meta.go
// nlines: 95

package meta

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// metaToolsToRegister holds the definitions for tools that provide information about other tools.
var metaToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "listTools",
			Group:       "meta",
			Description: "Lists the full specifications of all registered tools, optionally filtered by name.",
			Category:    "Introspection",
			Args: []tool.ArgSpec{
				// Added optional filter
				{Name: "filter", Type: tool.ArgTypeString, Required: false, Description: "A string to filter tool names. Only tools whose full name contains this text will be included."},
			},
			ReturnType: tool.ArgTypeSlice,
			ReturnHelp: "A list of tool specification objects.",
			Example:    `listTools("file")`,
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
	{
		Spec: tool.ToolSpec{
			Name:        "listToolNames",
			Group:       "meta",
			Description: "Provides a simple, newline-separated list of all available tool signatures, optionally filtered by name.",
			Category:    "Introspection",
			Args: []tool.ArgSpec{
				{Name: "filter", Type: tool.ArgTypeString, Required: false, Description: "A string to filter tool names. Only tools whose full name contains this text will be included."},
			},
			ReturnType: tool.ArgTypeString,
			ReturnHelp: "A single string, with each tool signature on its own line (e.g., 'tool.group.name(arg:type) -> type').",
			Example:    `listToolNames("file")`,
		},
		Func: ListToolNames,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "toolsHelp",
			Group:       "meta",
			Description: "Provides formatted Markdown help text for tools, optionally filtered by name.",
			Category:    "Introspection",
			Args: []tool.ArgSpec{
				{Name: "filter", Type: tool.ArgTypeString, Required: false, Description: "A string to filter tool names. Only tools whose full name contains this text will be included."},
			},
			ReturnType: tool.ArgTypeString,
			ReturnHelp: "A string containing Markdown-formatted help for all matching tools.",
			Example:    `toolsHelp("file")`,
		},
		Func: ToolsHelp,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "listGlobalConstants",
			Group:       "meta",
			Description: "Lists all global constants visible to the interpreter, optionally filtered by name.",
			Category:    "Introspection",
			Args: []tool.ArgSpec{
				{Name: "filter", Type: tool.ArgTypeString, Required: false, Description: "A string to filter constant names. Case-insensitive."},
			},
			ReturnType: tool.ArgTypeMap, // Returns a map of Name -> Value
			ReturnHelp: "A map containing the names and values of matching global constants.",
			Example:    `listGlobalConstants("FDM_")`,
		},
		Func: ListGlobalConstants,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "listFunctions",
			Group:       "meta",
			Description: "Lists the names of all functions visible to the interpreter, optionally filtered by name.",
			Category:    "Introspection",
			Args: []tool.ArgSpec{
				{Name: "filter", Type: tool.ArgTypeString, Required: false, Description: "A string to filter function names. Case-insensitive."},
			},
			ReturnType: tool.ArgTypeSlice, // Returns a list of strings
			ReturnHelp: "A list of matching function names.",
			Example:    `listFunctions("my_")`,
		},
		Func: ListFunctions,
	},
}
