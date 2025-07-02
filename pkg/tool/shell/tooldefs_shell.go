// NeuroScript Version: 0.3.1
// File version: 0.1.1
// Purpose: Populated Category, Example, ReturnHelp, and ErrorConditions for Shell.Execute tool spec.
// filename: pkg/tool/shell/tooldefs_shell.go
// nlines: 34
// risk_rating: HIGH

package shell

// shellToolsToRegister contains ToolImplementation definitions for Shell tools.
var shellToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:		"Shell.Execute",
			Description:	"Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.",
			Category:	"Shell Operations",
			Args: []ArgSpec{
				{Name: "command", Type: ArgTypeString, Required: true, Description: "The command or executable path (must not contain path separators like '/' or '\\')."},
				{Name: "args_list", Type: ArgTypeSliceString, Required: false, Description: "A list of string arguments for the command."},
				{Name: "directory", Type: ArgTypeString, Required: false, Description: "Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root."},
			},
			ReturnType:	ArgTypeMap,	// Returns map {stdout, stderr, exit_code, success}
			ReturnHelp:	"Returns a map containing 'stdout' (string), 'stderr' (string), 'exit_code' (int), and 'success' (bool) of the executed command. 'success' is true if the command exits with code 0, false otherwise. The command is executed within the sandboxed environment.",
			Example:	`tool.Shell.Execute("ls", ["-la"], "my_directory")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if an incorrect number of arguments is provided. " +
				"Returns `ErrInvalidArgument` or `ErrorCodeType` if 'command' is not a string, 'args_list' is not a list of strings, or 'directory' is not a string. " +
				"Returns `ErrSecurityViolation` if the 'command' path is deemed suspicious (e.g., contains path separators or shell metacharacters). " +
				"Returns `ErrInternal` if the internal FileAPI is not available. " +
				"May return path-related errors (e.g., `ErrFileNotFound`, `ErrPathNotDirectory`, `ErrPermissionDenied`) if the specified 'directory' is invalid or inaccessible. " +
				"If the command itself executes but fails (non-zero exit code), 'success' in the result map will be false, and 'stderr' may contain error details. OS-level execution errors are also captured in 'stderr'.",
		},
		Func:	toolExecuteCommand,	// Assumes toolExecuteCommand is defined in pkg/core/tools_shell.go
	},
	// If toolExecOutputToFile needs to be registered as a distinct tool, add it here.
}