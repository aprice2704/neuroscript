// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Renamed 'Prompt' tool to 'Input' to match tests and implementation.
// Purpose: Defines ToolImplementation structs for basic I/O tools.
// filename: pkg/tool/io/tooldefs_io.go
// nlines: 70 // Approximate
// risk_rating: LOW // Primarily deals with standard I/O.

package io

// ioToolsToRegister contains ToolImplementation definitions for basic I/O tools.
// Based on the provided pkg/core/tools_io.go.
var ioToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:		"Print",
			Description:	"Prints values to the standard output. If multiple values are passed in a list, they are printed space-separated.",
			Category:	"Input/Output",
			Args: []ArgSpec{
				// The toolPrint implementation takes a single arg, which can be a slice.
				// The NeuroScript 'Print' tool can be variadic, and the interpreter
				// would typically package these into a slice for the 'values' argument.
				{Name: "values", Type: ArgTypeAny, Required: true, Description: "A single value or a list of values to print. List elements will be space-separated."},
			},
			ReturnType:		ArgTypeNil,
			Variadic:		true,	// NeuroScript engine should pack variadic arguments into the 'values' slice.
			ReturnHelp:		"Returns nil. This tool is used for its side effect of printing to standard output.",
			Example:		`TOOL.Print(value: "Hello World")\nTOOL.Print(values: ["Hello", 42, "World!"]) // Prints "Hello 42 World!"`,
			ErrorConditions:	"ErrArgumentMismatch if the internal 'values' argument is not provided as expected by the implementation.",
		},
		Func:	toolPrint,	//
	},
	{
		Spec: ToolSpec{
			Name:		"Input",	// Changed from "Prompt" to "Input"
			Description:	"Displays a message and waits for user input from standard input. Returns the input as a string.",
			Category:	"Input/Output",
			Args: []ArgSpec{
				{Name: "message", Type: ArgTypeString, Required: false, Description: "The message to display to the user before waiting for input. If null or empty, no prompt message is printed."},
			},
			ReturnType:		ArgTypeString,
			ReturnHelp:		"Returns the string entered by the user, with trailing newline characters trimmed. Returns an empty string and an error if reading input fails.",
			Example:		`userName = TOOL.Input(message: "Enter your name: ")`,	// Updated example to use TOOL.Input
			ErrorConditions:	"ErrorCodeType if the prompt message argument is provided but not a string; ErrorCodeIOFailed if reading from standard input fails (e.g., EOF).",
		},
		Func:	toolInput,	// Mapped to toolInput
	},
	// Removed "Error" tool as toolPrintError is not in the provided tools_io.go
	// Removed "Log" tool as toolLogMessage is not in the provided tools_io.go
}