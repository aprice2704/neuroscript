// NeuroScript Version: 0.4.2
// File version: 1.3.0
// Purpose: Updated LoadScript tool definition to include file metadata in its return value.
// filename: pkg/core/tooldefs_script.go
package core

var scriptToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "LoadScript",
			Description: "Parses a string of NeuroScript code and loads its functions and event handlers into the current interpreter's scope. Does not execute any code.",
			Category:    "Scripting",
			Args: []ArgSpec{
				{Name: "script_content", Type: ArgTypeString, Required: true, Description: "A string containing the NeuroScript code to load."},
			},
			ReturnType:      ArgTypeMap,
			ReturnHelp:      "Returns a map with keys 'functions_loaded', 'event_handlers_loaded', and 'metadata', which contains the file-level metadata from the script header.",
			Example:         `set result = tool.LoadScript(script_content: ":: purpose: example\nfunc f()means\nendfunc")\nemit result["metadata"]["purpose"]`,
			ErrorConditions: "ErrArgumentMismatch if script_content is not a string or is missing. ErrSyntax if the script has syntax errors. ErrExecutionFailed if a function or event handler conflicts with an existing one (e.g., duplicate function name).",
		},
		Func: toolLoadScript, // from tools_script.go
	},
	{
		Spec: ToolSpec{
			Name:        "Script.ListFunctions",
			Description: "Returns a list of the names of all currently loaded functions (procedures) in the interpreter.",
			Category:    "Scripting",
			Args:        []ArgSpec{},
			ReturnType:  ArgTypeSliceAny, // Returns a list of strings
			ReturnHelp:  "Returns a list of strings, where each string is the name of a known function.",
			Example:     `set loaded_functions = call tool.Script.ListFunctions()`,
		},
		Func: toolScriptListFunctions, // from tools_script.go
	},
}
