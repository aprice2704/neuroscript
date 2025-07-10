// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Added the missing Reset tool definition.
// Purpose: Defines ToolImplementation structs for Git tools.
// filename: pkg/tool/git/tooldefs_git.go
// nlines: 260 // Approximate
// risk_rating: MEDIUM // Interacts with Git, potentially modifying state or exposing info.

package git

import (
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "git"

var gitToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Status",
			Group:       group,
			Description: "Gets the status of the Git repository in the configured sandbox directory.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "repo_path", Type: tool.ArgTypeString, Required: false, Description: "Optional. Relative path to the repository within the sandbox. Defaults to the sandbox root."},
			},
			ReturnType:      tool.ArgTypeMap,
			ReturnHelp:      "Returns a map containing Git status information.",
			Example:         `TOOL.Git.Status()`,
			ErrorConditions: "ErrConfiguration, ErrGitRepositoryNotFound, ErrIOFailed, ErrInvalidArgument.",
		},
		Func: toolGitStatus,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Add",
			Group:       group,
			Description: "Adds file contents to the index.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "The relative path within the sandbox to the Git repository."},
				{Name: "paths", Type: tool.ArgTypeSliceAny, Required: true, Description: "A list of file paths to add to the index."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message upon completion.",
			Example:         `TOOL.Git.Add(relative_path: "my_repo", paths: ["file1.txt", "docs/"])`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitAdd,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Reset",
			Group:       group,
			Description: "Resets the current HEAD to the specified state.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "mode", Type: tool.ArgTypeString, Required: false, Description: "Reset mode: 'soft', 'mixed' (default), 'hard', 'merge', or 'keep'."},
				{Name: "commit", Type: tool.ArgTypeString, Required: false, Description: "Commit to reset to. Defaults to HEAD."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Reset(relative_path: "my_repo", mode: "hard", commit: "HEAD~1")`,
			ErrorConditions: "ErrInvalidArgument, ErrGitOperationFailed.",
		},
		Func: toolGitReset,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Clone",
			Group:       group,
			Description: "Clones a Git repository into the specified relative path within the sandbox.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "repository_url", Type: tool.ArgTypeString, Required: true, Description: "The URL of the Git repository to clone."},
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "The relative path within the sandbox where the repository should be cloned."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Clone(repository_url: "https://github.com/example/repo.git", relative_path: "cloned_repos/my_repo")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrPathExists, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitClone,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Pull",
			Group:       group,
			Description: "Pulls the latest changes from the remote repository.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "remote_name", Type: tool.ArgTypeString, Required: false, Description: "Optional. The remote to pull from. Defaults to 'origin'."},
				{Name: "branch_name", Type: tool.ArgTypeString, Required: false, Description: "Optional. The branch to pull. Defaults to the current branch."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Pull(relative_path: "my_repo")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitPull,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Commit",
			Group:       group,
			Description: "Commits staged changes.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "commit_message", Type: tool.ArgTypeString, Required: true, Description: "The commit message."},
				{Name: "allow_empty", Type: tool.ArgTypeBool, Required: false, Description: "Allow an empty commit. Defaults to false."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Commit(relative_path: "my_repo", commit_message: "Fix: bug #123")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitCommit,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Push",
			Group:       group,
			Description: "Pushes committed changes to a remote repository.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "remote_name", Type: tool.ArgTypeString, Required: false, Description: "Optional. The remote to push to. Defaults to 'origin'."},
				{Name: "branch_name", Type: tool.ArgTypeString, Required: false, Description: "Optional. The branch to push. Defaults to the current branch."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Push(relative_path: "my_repo")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitPush,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Branch",
			Group:       group,
			Description: "Manages branches.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "name", Type: tool.ArgTypeString, Required: false, Description: "The name of the branch to create. If omitted, lists branches."},
				{Name: "checkout", Type: tool.ArgTypeBool, Required: false, Description: "If true, checks out the new branch after creation."},
				{Name: "list_remote", Type: tool.ArgTypeBool, Required: false, Description: "If true, lists remote-tracking branches."},
				{Name: "list_all", Type: tool.ArgTypeBool, Required: false, Description: "If true, lists all branches."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message or a list of branches.",
			Example:         `TOOL.Git.Branch(relative_path: "my_repo", name: "new-feature")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitBranch,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Checkout",
			Group:       group,
			Description: "Switches branches or restores working tree files.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "branch", Type: tool.ArgTypeString, Required: true, Description: "The branch to checkout."},
				{Name: "create", Type: tool.ArgTypeBool, Required: false, Description: "If true, creates and checks out the branch."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Checkout(relative_path: "my_repo", branch: "main")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitCheckout,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Rm",
			Group:       group,
			Description: "Removes files from the working tree and from the index.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "paths", Type: tool.ArgTypeAny, Required: true, Description: "A single path or a list of paths to remove."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Rm(relative_path: "my_repo", paths: "old_file.txt")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitRm,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Merge",
			Group:       group,
			Description: "Joins two or more development histories together.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "branch", Type: tool.ArgTypeString, Required: true, Description: "The branch to merge into the current branch."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a success message.",
			Example:         `TOOL.Git.Merge(relative_path: "my_repo", branch: "feature-branch")`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitMerge,
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Diff",
			Group:       group,
			Description: "Shows changes between commits, commit and working tree, etc.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true, Description: "Path to the repository."},
				{Name: "cached", Type: tool.ArgTypeBool, Required: false, Description: "Show staged changes."},
				{Name: "commit1", Type: tool.ArgTypeString, Required: false, Description: "First commit for diff."},
				{Name: "commit2", Type: tool.ArgTypeString, Required: false, Description: "Second commit for diff."},
				{Name: "path", Type: tool.ArgTypeString, Required: false, Description: "Limit the diff to a specific path."},
			},
			ReturnType:      tool.ArgTypeString,
			ReturnHelp:      "Returns a string containing the diff output.",
			Example:         `TOOL.Git.Diff(relative_path: "my_repo", cached: true)`,
			ErrorConditions: "ErrConfiguration, ErrInvalidArgument, ErrGitRepositoryNotFound, ErrGitOperationFailed, ErrSecurityPath.",
		},
		Func: toolGitDiff,
	},
}
