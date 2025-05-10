// NeuroScript Version: 0.3.8
// File version: 0.1.1
// Filename: pkg/core/tooldefs_meta.go
// nlines: 40
// risk_rating: LOW

package core

// metaToolsToRegister holds the definitions for tools that provide information about other tools.
// These tools will be registered globally via zz_core_tools_registrar.go.
var metaToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Meta.ListTools",
			Description: "Provides a compact text list (sorted alphabetically) of all currently available tools, including basic parameter information. Each tool is listed on a new line, showing its name, parameters (name:type), and return type. Example: FS.Read(filepath:string) -> string",
			Args:        []ArgSpec{}, // No arguments
			ReturnType:  ArgTypeString,
			// ReturnHelp removed, integrated into Description
		},
		Func: toolListTools, // Implementation in tools_meta.go
	},
	{
		Spec: ToolSpec{
			Name:        "Meta.ToolsHelp",
			Description: "Provides a more extensive, Markdown-formatted list of available tools, including descriptions, parameters, and return types. Can be filtered by providing a partial tool name. Details include parameter names, types, descriptions, and return type with its description.",
			Args: []ArgSpec{
				{
					Name:        "filter",
					Type:        ArgTypeString,
					Description: "An optional string to filter tool names. Only tools whose names contain this substring will be listed. If empty or omitted, all tools are listed.",
					Required:    false,
					// Default removed
				},
			},
			ReturnType: ArgTypeString,
			// ReturnHelp removed, integrated into Description
		},
		Func: toolToolsHelp, // Implementation in tools_meta.go
	},
}
