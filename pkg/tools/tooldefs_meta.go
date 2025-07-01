// NeuroScript Version: 0.3.8
// File version: 0.1.4 // Added Meta.GetToolSpecificationsJSON tool definition.
// Filename: pkg/core/tooldefs_meta.go
// nlines: 70 // Approximate
// risk_rating: LOW

package core

// metaToolsToRegister holds the definitions for tools that provide information about other tools.
// These tools will be registered globally via zz_core_tools_registrar.go.
var metaToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:            "Meta.ListTools",
			Description:     "Provides a compact text list (sorted alphabetically) of all currently available tools, including basic parameter information. Each tool is listed on a new line, showing its name, parameters (name:type), and return type. Example: FS.Read(filepath:string) -> string",
			Category:        "Introspection",
			Args:            []ArgSpec{}, // No arguments
			ReturnType:      ArgTypeString,
			ReturnHelp:      "A string containing a newline-separated list of tool names, their parameters (name:type), and return types.",
			Variadic:        false,
			Example:         "TOOL.Meta.ListTools()",
			ErrorConditions: "Generally does not return errors, unless the ToolRegistry is uninitialized (which would be an ErrorCodeConfiguration if an attempt is made to call it in such a state).",
		},
		Func: toolListTools, // Implementation in tools_meta.go
	},
	{
		Spec: ToolSpec{
			Name:        "Meta.ToolsHelp",
			Description: "Provides a more extensive, Markdown-formatted list of available tools, including descriptions, parameters, and return types. Can be filtered by providing a partial tool name. Details include parameter names, types, descriptions, and return type with its description.",
			Category:    "Introspection",
			Args: []ArgSpec{
				{
					Name:        "filter",
					Type:        ArgTypeString,
					Description: "An optional string to filter tool names. Only tools whose names contain this substring will be listed. If empty or omitted, all tools are listed.",
					Required:    false,
					// DefaultValue: nil, // Handled by tool logic
				},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "A string in Markdown format detailing available tools, their descriptions, parameters, and return types. Output can be filtered by the optional 'filter' argument.",
			Variadic:        false,
			Example:         "TOOL.Meta.ToolsHelp(filter: \"FS\")\nTOOL.Meta.ToolsHelp()",
			ErrorConditions: "Returns ErrorCodeType if the 'filter' argument is provided but is not a string. Generally does not return other errors, unless the ToolRegistry is uninitialized (ErrorCodeConfiguration).",
		},
		Func: toolToolsHelp, // Implementation in tools_meta.go
	},
	{
		Spec: ToolSpec{
			Name:            "Meta.GetToolSpecificationsJSON",
			Description:     "Provides a JSON string containing an array of all currently available tool specifications. Each object in the array represents a tool and includes its name, description, category, arguments (with their details), return type, return help, variadic status, example usage, and error conditions.",
			Category:        "Introspection",
			Args:            []ArgSpec{}, // No arguments
			ReturnType:      ArgTypeString,
			ReturnHelp:      "A JSON string representing an array of ToolSpec objects. This is intended for programmatic use or detailed inspection of all tool capabilities.",
			Variadic:        false,
			Example:         "TOOL.Meta.GetToolSpecificationsJSON()",
			ErrorConditions: "Returns an error (ErrorCodeInternal) if JSON marshalling of the tool specifications fails. Generally does not return other errors unless the ToolRegistry is uninitialized (ErrorCodeConfiguration).",
		},
		Func: toolGetToolSpecificationsJSON, // To be implemented in tools_meta.go
	},
}
