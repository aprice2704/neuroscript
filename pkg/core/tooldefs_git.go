// NeuroScript Version: 0.3.1
// File version: 0.1.0
// Defines ToolImplementation structs for selected Git tools.
// filename: pkg/core/tooldefs_git.go

package core

// gitToolsToRegister contains ToolImplementation definitions for a subset of Git tools.
// This array is intended to be concatenated with other similar arrays in a central
// registrar (e.g., zz_core_tools_registrar.go) to be processed by AddToolImplementations.
//
// Tools that already have their own init() function calling AddToolImplementations
// (e.g., most tools in tools_git.go) should NOT be included here.
var gitToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "GitStatus",
			Description: "Provides a summary of the git repository status, including branch, remote, ahead/behind counts, and modified/untracked files.",
			Args:        []ArgSpec{}, // Expects no arguments
			ReturnType:  ArgTypeMap,  // Returns a map detailing the status
		},
		Func: toolGitStatus, // Assumes toolGitStatus is defined in pkg/core/tools_git_status.go
	},
	// Add other Git tool definitions here if they follow the same pattern
	// (i.e., not registered by an init() in their own file or the main tools_git.go init).
}
