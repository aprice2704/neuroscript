// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines ToolImplementation structs for Shell tools.
// filename: pkg/core/tooldefs_shell.go

package core

// shellToolsToRegister contains ToolImplementation definitions for Shell tools.
var shellToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Shell.Execute",
			Description: "Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.",
			Args: []ArgSpec{
				{Name: "command", Type: ArgTypeString, Required: true, Description: "The command or executable path."},
				{Name: "args_list", Type: ArgTypeSliceString, Required: false, Description: "A list of string arguments for the command."},
				{Name: "directory", Type: ArgTypeString, Required: false, Description: "Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root."},
			},
			ReturnType: ArgTypeMap, // Returns map {stdout, stderr, exit_code, success}
		},
		Func: toolExecuteCommand, // Assumes toolExecuteCommand is defined in pkg/core/tools_shell.go
	},
	// If toolExecOutputToFile needs to be registered as a distinct tool, add it here.
}
