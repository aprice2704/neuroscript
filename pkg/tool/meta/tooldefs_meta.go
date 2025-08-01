// NeuroScript Version: 0.3.8
// File version: 0.1.6 // Corrected group name to PascalCase ("Meta") to match convention.
// nlines: 70 // Approximate
// risk_rating: LOW

// filename: pkg/tool/meta/tooldefs_meta.go
package meta

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "Meta"

// metaToolsToRegister holds the definitions for tools that provide information about other tools.
// These tools will be registered globally via zz_core_tools_registrar.go.
var metaToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:            "ListTools",
			Group:           group,
			Description:     "Provides a compact text list (sorted alphabetically) of all currently available tools, including basic parameter information. Each tool is listed on a new line, showing its name, parameters (name:type), and return type. Example: FS.Read(filepath:string) -> string",
			Category:        "Introspection",
			Args:            []tool.ArgSpec{}, // No arguments
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "A string containing a newline-separated list of tool names, their parameters (name:type), and return types.",
			Variadic:        false,
			Example:         "ListTools()",
			ErrorConditions: "Generally does not return errors, unless the ToolRegistry is uninitialized (which would be an ErrorCodeConfiguration if an attempt is made to call it in such a state).",
		},
		Func: toolListTools, // Implementation in tools_meta.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ToolsHelp",
			Group:       group,
			Description: "Provides a more extensive, Markdown-formatted list of available tools, including descriptions, parameters, and return types. Can be filtered by providing a partial tool name. Details include parameter names, types, descriptions, and return type with its description.",
			Category:    "Introspection",
			Args: []tool.ArgSpec{
				{
					Name:        "filter",
					Type:        tool.ArgTypeString,
					Description: "An optional string to filter tool names. Only tools whose names contain this substring will be listed. If empty or omitted, all tools are listed.",
					Required:    false,
				},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "A string in Markdown format detailing available tools, their descriptions, parameters, and return types. Output can be filtered by the optional 'filter' argument.",
			Variadic:        false,
			Example:         "ToolsHelp(filter: \"FS\")\nMeta.ToolsHelp()",
			ErrorConditions: "Returns ErrorCodeType if the 'filter' argument is provided but is not a string. Generally does not return other errors, unless the ToolRegistry is uninitialized (ErrorCodeConfiguration).",
		},
		Func: toolToolsHelp, // Implementation in tools_meta.go
	},
	{
		Spec: tool.ToolSpec{
			Name:            "GetToolSpecificationsJSON",
			Group:           group,
			Description:     "Provides a JSON string containing an array of all currently available tool specifications. Each object in the array represents a tool and includes its name, description, category, arguments (with their details), return type, return help, variadic status, example usage, and error conditions.",
			Category:        "Introspection",
			Args:            []tool.ArgSpec{},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "A JSON string representing an array of ToolSpec objects. This is intended for programmatic use or detailed inspection of all tool capabilities.",
			Variadic:        false,
			Example:         "GetToolSpecificationsJSON()",
			ErrorConditions: "Returns an error (ErrorCodeInternal) if JSON marshalling of the tool specifications fails. Generally does not return other errors unless the ToolRegistry is uninitialized (ErrorCodeConfiguration).",
		},
		Func: toolGetToolSpecificationsJSON, // To be implemented in tools_meta.go
	},
}
