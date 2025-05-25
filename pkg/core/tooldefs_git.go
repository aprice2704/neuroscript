// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Populated Category, Example, ReturnHelp, ErrorConditions for ToolSpecs.
// nlines: 130 // Approximate
// risk_rating: MEDIUM // Interacts with Git, potentially modifying state or exposing info.
// filename: pkg/core/tooldefs_git.go

package core

var gitToolsToRegister = []ToolImplementation{
	{
		Spec: ToolSpec{
			Name:        "Git.Status",
			Description: "Gets the status of the Git repository in the configured sandbox directory.",
			Category:    "Git",
			Args: []ArgSpec{
				{Name: "repo_path", Type: ArgTypeString, Required: false, Description: "Optional. Relative path to the repository within the sandbox. Defaults to the sandbox root."},
			},
			ReturnType:      ArgTypeMap,
			ReturnHelp:      "Returns a map containing Git status information: 'current_branch' (string), 'is_clean' (bool), 'uncommitted_changes' ([]string of changed file paths), 'untracked_files' ([]string of untracked file paths), and 'error' (string, if any occurred internally). See tools_git_status.go for exact structure.",
			Example:         `TOOL.Git.Status() // For sandbox root\nTOOL.Git.Status(repo_path: "my_sub_repo")`,
			ErrorConditions: "ErrConfiguration if sandbox directory is not set; ErrGitRepositoryNotFound if the specified path is not a Git repository; ErrIOFailed for underlying Git command execution errors or issues reading Git output; ErrInvalidArgument if repo_path is not a string.",
		},
		Func: toolGitStatus, // from tools_git_status.go
	},
	{
		Spec: ToolSpec{
			Name:        "Git.Clone",
			Description: "Clones a Git repository into the specified relative path within the sandbox.",
			Category:    "Git",
			Args: []ArgSpec{
				{Name: "repository_url", Type: ArgTypeString, Required: true, Description: "The URL of the Git repository to clone."},
				{Name: "relative_path", Type: ArgTypeString, Required: true, Description: "The relative path within the sandbox where the repository should be cloned."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a success message string like 'Successfully cloned <URL> to <path>.' on successful clone. Returns nil on error.",
			Example:         `TOOL.Git.Clone(repository_url: "https://github.com/example/repo.git", relative_path: "cloned_repos/my_repo")`,
			ErrorConditions: "ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if repository_url or relative_path are missing or not strings; ErrPathExists if the target relative_path already exists; ErrGitOperationFailed for errors during the 'git clone' command execution (e.g., authentication failure, repository not found, network issues); ErrSecurityPath for invalid relative_path.",
		},
		Func: toolGitClone, // from tools_git.go
	},
	{
		Spec: ToolSpec{
			Name:        "Git.Pull",
			Description: "Pulls the latest changes from the remote repository for the specified Git repository within the sandbox.",
			Category:    "Git",
			Args: []ArgSpec{
				{Name: "relative_path", Type: ArgTypeString, Required: true, Description: "The relative path within the sandbox to the Git repository."},
				{Name: "remote_name", Type: ArgTypeString, Required: false, Description: "Optional. The name of the remote to pull from (e.g., 'origin'). Defaults to 'origin'."},
				{Name: "branch_name", Type: ArgTypeString, Required: false, Description: "Optional. The name of the branch to pull. Defaults to the current branch."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a success message string like 'Successfully pulled from <remote>/<branch> for repository <path>.' or details of the pull. Returns nil on error.",
			Example:         `TOOL.Git.Pull(relative_path: "my_repo")\nTOOL.Git.Pull(relative_path: "my_repo", remote_name: "upstream", branch_name: "main")`,
			ErrorConditions: "ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if relative_path is missing or not a string, or other args are invalid types; ErrGitRepositoryNotFound if the specified relative_path is not a Git repository; ErrGitOperationFailed for errors during the 'git pull' command execution (e.g., merge conflicts, authentication failure, network issues); ErrSecurityPath for invalid relative_path.",
		},
		Func: toolGitPull, // from tools_git.go
	},
	{
		Spec: ToolSpec{
			Name:        "Git.Commit",
			Description: "Commits staged changes in the specified Git repository within the sandbox.",
			Category:    "Git",
			Args: []ArgSpec{
				{Name: "relative_path", Type: ArgTypeString, Required: true, Description: "The relative path within the sandbox to the Git repository."},
				{Name: "commit_message", Type: ArgTypeString, Required: true, Description: "The commit message."},
				{Name: "allow_empty", Type: ArgTypeBool, Required: false, Description: "Optional. Allow an empty commit (no changes). Defaults to false."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a success message string like 'Successfully committed to repository <path>.' or the commit hash. Returns nil on error.",
			Example:         `TOOL.Git.Commit(relative_path: "my_repo", commit_message: "Fix: addressed critical bug #123")`,
			ErrorConditions: "ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if relative_path or commit_message are missing/invalid types; ErrGitRepositoryNotFound if the specified relative_path is not a Git repository; ErrGitOperationFailed for errors during the 'git commit' command (e.g., nothing to commit and allow_empty is false, pre-commit hooks failure); ErrSecurityPath for invalid relative_path.",
		},
		Func: toolGitCommit, // from tools_git.go
	},
	{
		Spec: ToolSpec{
			Name:        "Git.Push",
			Description: "Pushes committed changes to a remote repository.",
			Category:    "Git",
			Args: []ArgSpec{
				{Name: "relative_path", Type: ArgTypeString, Required: true, Description: "The relative path within the sandbox to the Git repository."},
				{Name: "remote_name", Type: ArgTypeString, Required: false, Description: "Optional. The name of the remote to push to (e.g., 'origin'). Defaults to 'origin'."},
				{Name: "branch_name", Type: ArgTypeString, Required: false, Description: "Optional. The name of the local branch to push. Defaults to the current branch."},
			},
			ReturnType:      ArgTypeString,
			ReturnHelp:      "Returns a success message string like 'Successfully pushed to <remote>/<branch> for repository <path>.' Returns nil on error.",
			Example:         `TOOL.Git.Push(relative_path: "my_repo")\nTOOL.Git.Push(relative_path: "my_repo", remote_name: "origin", branch_name: "feature/new-thing")`,
			ErrorConditions: "ErrConfiguration if sandbox directory is not set; ErrInvalidArgument if relative_path is missing/invalid type; ErrGitRepositoryNotFound if the specified relative_path is not a Git repository; ErrGitOperationFailed for errors during the 'git push' command (e.g., authentication failure, non-fast-forward, network issues); ErrSecurityPath for invalid relative_path.",
		},
		Func: toolGitPush, // from tools_git.go
	},
}
