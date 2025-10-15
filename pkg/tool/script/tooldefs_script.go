// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Exports the tool list for use in external test packages.
// filename: pkg/tool/script/tooldefs_script.go
package script

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "script"

// ToolsToRegister holds the definitions for the script-related tools.
// It is exported to allow external test packages to register these tools.
var ToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "LoadScript",
			Group:       group,
			Description: "Parses a string of NeuroScript code and loads its functions and event handlers into the current interpreter's scope. Does not execute any code.",
			Category:    "Scripting",
			Args: []tool.ArgSpec{
				{Name: "script_content", Type: tool.ArgTypeString, Required: true, Description: "A string containing the NeuroScript code to load."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map with keys 'functions_loaded', 'event_handlers_loaded', and 'metadata', which contains the file-level metadata from the script header.",
			Example:         `set result = tool.script.LoadScript(":: purpose: example\nfunc f()means\nendfunc")\nemit result["metadata"]["purpose"]`,
			ErrorConditions: "ErrArgumentMismatch if script_content is not a string or is missing. ErrSyntax if the script has syntax errors. ErrExecutionFailed if a function or event handler conflicts with an existing one (e.g., duplicate function name).",
		},
		Func: toolLoadScript, // from tools_script.go
	},
	{
		Spec: tool.ToolSpec{
			Name:        "ListFunctions",
			Group:       group,
			Description: "Returns a map of all currently loaded function (procedure) names to their signatures.",
			Category:    "Scripting",
			Args:        []tool.ArgSpec{},
			ReturnType:  tool.ArgTypeMap, // Returns a map of strings to strings
			ReturnHelp:  "Returns a map where each key is the name of a known function and the value is its signature.",
			Example:     `set loaded_functions = tool.script.ListFunctions()`,
		},
		Func: toolScriptListFunctions, // from tools_script.go
	},
}
