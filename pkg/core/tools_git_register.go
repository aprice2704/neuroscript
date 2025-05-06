// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Update specs, use toolGitBranch, fix core prefix.
// Registers Git-related tools.
// filename: pkg/core/tools_git_register.go

package core // Ensure we are in the core package

import (
	"errors" // Use errors for joining
	"fmt"
	"strings"
)

// RegisterGitTools adds Git commands to the interpreter's tool registry.
// Assumes toolGitStatus is defined in tools_git_status.go
func RegisterGitTools(registry ToolRegistrar) error {

	// Define ToolImplementations using updated specs and correct function names
	implementations := []ToolImplementation{
		// GitStatus (Spec assumed correct from previous definition)
		{
			Spec: ToolSpec{
				Name: "Git.Status",
				Description: "Gets the current Git repository status using 'git status --porcelain -b --untracked-files=all' and returns a structured map. " +
					"Keys: 'branch', 'remote_branch', 'ahead', 'behind', 'files', 'untracked_files_present', 'is_clean', 'error'.",
				Args:       []ArgSpec{},
				ReturnType: ArgTypeAny, // Returns a map, so ArgTypeAny is appropriate
			},
			Func: toolGitStatus, // Assumes function exists in this package (likely tools_git_status.go)
		},
		// GitAdd (Spec adjusted slightly)
		{
			Spec: ToolSpec{
				Name:        "Git.Add",
				Description: "Add file contents to the index. Accepts a single path or a list of paths.",
				Args: []ArgSpec{
					{Name: "paths", Type: ArgTypeAny, Required: true, Description: "A single file path string or a list of file path strings to stage."},
				},
				ReturnType: ArgTypeString, // Returns success message + output
			},
			Func: toolGitAdd,
		},
		// GitCommit (Updated Spec)
		{
			Spec: ToolSpec{
				Name:        "Git.Commit",
				Description: "Records changes to the repository.",
				Args: []ArgSpec{
					{Name: "message", Type: ArgTypeString, Required: true, Description: "The commit message."},
					{Name: "add_all", Type: ArgTypeBool, Required: false, Description: "If true, stage all tracked, modified files (`git add .`) before committing. Default: false."},
				},
				ReturnType: ArgTypeString, // Success message or specific status like "nothing to commit"
			},
			Func: toolGitCommit,
		},
		// GitBranch (Updated Spec - replaces GitNewBranch)
		{
			Spec: ToolSpec{
				Name:        "Git.Branch",
				Description: "Lists existing branches or creates a new branch. By default lists local branches.",
				Args: []ArgSpec{
					{Name: "name", Type: ArgTypeString, Required: false, Description: "If provided, create a branch with this name. If omitted, list existing branches."},
					{Name: "checkout", Type: ArgTypeBool, Required: false, Description: "If creating a branch (name provided), also check it out immediately (`-b` flag). Default: false."},
					{Name: "list_remote", Type: ArgTypeBool, Required: false, Description: "If listing branches (name omitted), list remote branches (`-r`). Default: false."},
					{Name: "list_all", Type: ArgTypeBool, Required: false, Description: "If listing branches (name omitted), list all branches (`-a`). Default: false."},
				},
				ReturnType: ArgTypeAny, // slice_string for list, string (message) for create
			},
			Func: toolGitBranch, // Use the renamed function
		},
		// GitCheckout (Updated Spec)
		{
			Spec: ToolSpec{
				Name:        "Git.Checkout",
				Description: "Switches branches or restores working tree files. Can also create a new branch before switching.",
				Args: []ArgSpec{
					{Name: "branch", Type: ArgTypeString, Required: true, Description: "The name of the branch or commit to check out."},
					{Name: "create", Type: ArgTypeBool, Required: false, Description: "If true, create the branch if it doesn't exist (`-b` flag). Default: false."},
				},
				ReturnType: ArgTypeString, // Success message
			},
			Func: toolGitCheckout,
		},
		// GitRm (Spec assumed correct)
		{
			Spec: ToolSpec{
				Name:        "Git.Rm",
				Description: "Remove files from the working tree and from the index.",
				Args: []ArgSpec{
					{Name: "paths", Type: ArgTypeAny, Required: true, Description: "A single file path string or a list of file path strings to remove."},
				},
				ReturnType: ArgTypeString, // Success message
			},
			Func: toolGitRm,
		},
		// GitMerge (Spec assumed correct)
		{
			Spec: ToolSpec{
				Name:        "Git.Merge",
				Description: "Join two or more development histories together.",
				Args: []ArgSpec{
					{Name: "branch", Type: ArgTypeString, Required: true, Description: "The name of the branch to merge into the current branch."},
				},
				ReturnType: ArgTypeString, // Success message or error/conflict output
			},
			Func: toolGitMerge,
		},
		// GitPull (Spec assumed correct)
		{
			Spec: ToolSpec{
				Name:        "Git.Pull",
				Description: "Fetch from and integrate with another repository or a local branch.",
				Args:        []ArgSpec{},
				ReturnType:  ArgTypeString, // Success message or error/conflict output
			},
			Func: toolGitPull,
		},
		// GitPush (Updated Spec)
		{
			Spec: ToolSpec{
				Name:        "Git.Push",
				Description: "Updates remote refs along with associated objects.",
				Args: []ArgSpec{
					{Name: "remote", Type: ArgTypeString, Required: false, Description: "The remote repository name. Default: 'origin'."},
					{Name: "branch", Type: ArgTypeString, Required: false, Description: "The local branch name to push. Default: current branch."},
					{Name: "set_upstream", Type: ArgTypeBool, Required: false, Description: "If true, set the upstream tracking configuration (`-u`). Default: false."},
				},
				ReturnType: ArgTypeString, // Success message
			},
			Func: toolGitPush,
		},
		// GitDiff (Updated Spec)
		{
			Spec: ToolSpec{
				Name:        "Git.Diff",
				Description: "Shows changes between commits, commit and working tree, etc. Returns the diff output or a message indicating no changes.",
				Args: []ArgSpec{
					{Name: "cached", Type: ArgTypeBool, Required: false, Description: "Show diff of staged changes against HEAD (`--cached`). Default: false."},
					{Name: "commit1", Type: ArgTypeString, Required: false, Description: "First commit/branch/tree reference. If specified alone, compares commit vs working tree."},
					{Name: "commit2", Type: ArgTypeString, Required: false, Description: "Second commit/branch/tree reference. If commit1 omitted or both omitted, compares index vs working tree."}, // Clarified default behavior
					{Name: "path", Type: ArgTypeString, Required: false, Description: "Limit the diff to the specified file or directory path."},
				},
				ReturnType: ArgTypeString, // The diff output or "GitDiff: No changes detected."
			},
			Func: toolGitDiff,
		},
	}

	// Register all tools
	var errs []error
	for _, impl := range implementations {
		if err := registry.RegisterTool(impl); err != nil {
			errs = append(errs, fmt.Errorf("failed to register tool '%s': %w", impl.Spec.Name, err))
		}
	}

	// Check for registration errors
	if len(errs) > 0 {
		// Use errors.Join (Go 1.20+) for cleaner error aggregation
		combinedError := errors.Join(errs...)
		return fmt.Errorf("encountered %d error(s) registering Git tools: %w", len(errs), combinedError)
	}

	// Optional: Log success if possible (checking if registry is the expected type)
	// Note: This check might be fragile if the interface changes.
	if toolRegistry, ok := registry.(*ToolRegistry); ok {
		if toolRegistry.interpreter != nil && toolRegistry.interpreter.Logger() != nil {
			toolRegistry.interpreter.Logger().Debug("Successfully registered enhanced Git tools.")
		}
	}

	return nil
}

// Helper function to format the output map from toolExec consistently
// This might live in a more general helpers file eventually
func formatExecOutput(outputMap map[string]interface{}) string {
	stdout, _ := outputMap["stdout"].(string)
	stderr, _ := outputMap["stderr"].(string)
	exitCode, _ := outputMap["exit_code"].(int64)
	success, _ := outputMap["success"].(bool)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Success: %v, ExitCode: %d\n", success, exitCode))
	if stdout != "" {
		sb.WriteString(fmt.Sprintf("--- STDOUT ---\n%s\n--------------\n", stdout))
	}
	if stderr != "" {
		sb.WriteString(fmt.Sprintf("--- STDERR ---\n%s\n--------------\n", stderr))
	}
	return sb.String()
}
