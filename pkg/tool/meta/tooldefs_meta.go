// NeuroScript Major Version: 1
// File version: 10
// Purpose: Defines the tool specifications for the 'meta' tool group, linking them to their implementations.
// Latest change: Corrected compiler errors: changed Help to Description, fixed listTools ReturnType, and removed invalid validation block.
// filename: pkg/tool/meta/tooldefs_meta.go
// nlines: 69

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
			Description: "Lists the full specifications of all registered tools.",
			Category:    "Introspection",
			ReturnType:  tool.ArgTypeSlice, // FIX: Changed back from ArrayOf(ObjectOf(...))
			ReturnHelp:  "A list of tool specification objects.",
			Example:     "listTools()",
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
			Description: "Provides a simple, newline-separated list of all available tool signatures.",
			Category:    "Introspection",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeString,
			ReturnHelp:  "A single string, with each tool signature on its own line (e.g., 'tool.group.name(arg:type) -> type').",
			Example:     "listToolNames()",
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
				// FIX: Changed 'Help' field to 'Description'
				{Name: "filter", Type: tool.ArgTypeString, Required: false, Description: "A string to filter tool names. Only tools whose full name contains this text will be included."},
			},
			ReturnType: tool.ArgTypeString,
			ReturnHelp: "A string containing Markdown-formatted help for all matching tools.",
			Example:    `toolsHelp("file")`,
		},
		Func: ToolsHelp,
	},
}

// FIX: Removed the invalid validation block that was causing compiler errors.
// The project convention seen in strtools does not use this.
