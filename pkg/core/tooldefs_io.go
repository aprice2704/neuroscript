// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines ToolImplementation structs for IO tools.
// filename: pkg/core/tooldefs_io.go

package core

// ioToolsToRegister contains ToolImplementation definitions for IO tools.
var ioToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "IO.Input",
			Description: "Prompts the user for text input via the console. Enforces max input size. Not allowed in agent mode.",
			Args: []ArgSpec{
				{Name: "prompt", Type: ArgTypeString, Required: true, Description: "The text prompt to display to the user."},
			},
			ReturnType: ArgTypeMap, // Returns map {"input": string|null, "error": string|null}
		},
		Func: toolIOInput, // Assumes toolIOInput is defined in pkg/core/tools_io.go
	},
	{
		Spec: ToolSpec{
			Name:        "Log",
			Description: "Writes a message to the application's internal log stream at a specified level.",
			Args: []ArgSpec{
				{Name: "level", Type: ArgTypeString, Required: true, Description: "Log level (e.g., 'Info', 'Debug', 'Warn', 'Error'). Case-insensitive."},
				{Name: "message", Type: ArgTypeString, Required: true, Description: "The message to log."},
			},
			ReturnType: ArgTypeNil, // No meaningful return value
		},
		Func: toolLog, // Assumes toolLog is defined in pkg/core/tools_io.go
	},
}
