// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Add Group field to Shell.Execute tool spec and correct example for full name registration.
// filename: pkg/tool/shell/tooldefs_shell.go
// nlines: 35
// risk_rating: HIGH

package shell

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "shell"

// shellToolsToRegister contains ToolImplementation definitions for Shell tools.
var shellToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Execute",
			Group:       group,
			Description: "Executes an arbitrary shell command. WARNING: Use with extreme caution due to security risks. Command path validation is basic. Consider using specific tools (e.g., GoBuild, GitAdd) instead.",
			Category:    "Shell Operations",
			Args: []tool.ArgSpec{
				{Name: "command", Type: tool.ArgTypeString, Required: true, Description: "The command or executable path (must not contain path separators like '/' or '\\')."},
				{Name: "args_list", Type: tool.ArgTypeSliceString, Required: false, Description: "A list of string arguments for the command."},
				{Name: "directory", Type: tool.ArgTypeString, Required: false, Description: "Optional directory (relative to sandbox) to execute the command in. Defaults to sandbox root."},
			},
			ReturnType: tool.ArgTypeMap, // Returns map {stdout, stderr, exit_code, success}
			ReturnHelp: "Returns a map containing 'stdout' (string), 'stderr' (string), 'exit_code' (int), and 'success' (bool) of the executed command. 'success' is true if the command exits with code 0, false otherwise. The command is executed within the sandboxed environment.",
			Example:    `tool.shell.Execute("ls", ["-la"], "my_directory")`,
			ErrorConditions: "Returns `ErrArgumentMismatch` if an incorrect number of arguments is provided. " +
				"Returns `ErrInvalidArgument` or `ErrorCodeType` if 'command' is not a string, 'args_list' is not a list of strings, or 'directory' is not a string. " +
				"Returns `ErrSecurityViolation` if the 'command' path is deemed suspicious (e.g., contains path separators or shell metacharacters). " +
				"Returns `ErrInternal` if the internal FileAPI is not available. " +
				"May return path-related errors (e.g., `ErrFileNotFound`, `ErrPathNotDirectory`, `ErrPermissionDenied`) if the specified 'directory' is invalid or inaccessible. " +
				"If the command itself executes but fails (non-zero exit code), 'success' in the result map will be false, and 'stderr' may contain error details. OS-level execution errors are also captured in 'stderr'.",
		},
		Func: ToolExecuteCommand, // Assumes toolExecuteCommand is defined in pkg/core/tools_shell.go
	},
	// If toolExecOutputToFile needs to be registered as a distinct tool, add it here.
}
