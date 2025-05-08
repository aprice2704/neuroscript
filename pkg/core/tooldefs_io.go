// NeuroScript Version: 0.3.1
// File version: 0.0.1 // Defines IO tools registration variable.
// nlines: 31
// risk_rating: LOW
// filename: pkg/core/tooldefs_io.go

package core

// ioToolsToRegister defines the ToolImplementation structs for core I/O tools.
// This variable is used by zz_core_tools_registrar.go to register the tools.
var ioToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Input",
			Description: "Reads a single line of text from standard input.",
			Args: []ArgSpec{
				{
					Name:        "prompt",
					Type:        ArgTypeString,
					Description: "Optional prompt message to display to the user.",
					Required:    false,
				},
			},
			ReturnType: ArgTypeString, // Returns the line read from input
		},
		Func: toolInput, // Assumes toolInput is defined in tools_io.go
	},
	{
		Spec: ToolSpec{
			Name:        "Print",
			Description: "Prints the provided arguments to standard output, separated by spaces, followed by a newline.",
			Args: []ArgSpec{
				// Note: Uses VariadicArgs field in the underlying implementation if available,
				// otherwise Args needs to handle different types or expect a list.
				// Keeping Args simple here, toolPrint needs to handle variadic nature.
				{Name: "values", Type: ArgTypeAny, Required: true, Description: "One or more values to print."},
			},
			ReturnType: ArgTypeNil, // Print has no return value
		},
		Func: toolPrint, // Assumes toolPrint is defined in tools_io.go
	},
}
