// NeuroScript Version: 0.3.1
// File version: 0.1.1 // Consolidate all Git tool definitions as literals from tools_git_register.go.
// nlines: 156 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/tooldefs_git.go

package core

var gitToolsToRegister = []ToolImplementation{
	ToolImplementation{
		Spec: ToolSpec{
			Name: "Git.Status",
			Description: "Gets the current Git repository status using 'git status --porcelain -b --untracked-files=all' and returns a structured map. " +
				"Keys: 'branch', 'remote_branch', 'ahead', 'behind', 'files', 'untracked_files_present', 'is_clean', 'error'.",
			Args:       []ArgSpec{},
			ReturnType: ArgTypeMap, // Returns a map
		},
		Func: toolGitStatus, // from tools_git_status.go
	},
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Add",
			Description: "Add file contents to the index. Accepts a single path string or a list of path strings.",
			Args: []ArgSpec{
				{Name: "paths", Type: ArgTypeAny, Required: true, Description: "A single file path string or a list of file path strings to stage."},
			},
			ReturnType: ArgTypeString, // Returns success message + output
		},
		Func: toolGitAdd, // from tools_git.go
	},
	// --- Git.Commit ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Commit",
			Description: "Records changes to the repository.",
			Args: []ArgSpec{
				{Name: "message", Type: ArgTypeString, Required: true, Description: "The commit message."},
				{Name: "add_all", Type: ArgTypeBool, Required: false, Description: "If true, stage all tracked, modified files (`git add .`) before committing. Default: false."},
			},
			ReturnType: ArgTypeString, // Success message or specific status like "nothing to commit"
		},
		Func: toolGitCommit, // from tools_git.go
	},
	// --- Git.Branch ---
	ToolImplementation{
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
		Func: toolGitBranch, // from tools_git.go
	},
	// --- Git.Checkout ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Checkout",
			Description: "Switches branches or restores working tree files. Can also create a new branch before switching.",
			Args: []ArgSpec{
				{Name: "branch", Type: ArgTypeString, Required: true, Description: "The name of the branch or commit to check out."},
				{Name: "create", Type: ArgTypeBool, Required: false, Description: "If true, create the branch if it doesn't exist (`-b` flag). Default: false."},
			},
			ReturnType: ArgTypeString, // Success message
		},
		Func: toolGitCheckout, // from tools_git.go
	},
	// --- Git.Rm ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Rm",
			Description: "Remove files from the working tree and from the index.",
			Args: []ArgSpec{
				{Name: "paths", Type: ArgTypeAny, Required: true, Description: "A single file path string or a list of file path strings to remove."},
			},
			ReturnType: ArgTypeString, // Success message
		},
		Func: toolGitRm, // from tools_git.go
	},
	// --- Git.Merge ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Merge",
			Description: "Join two or more development histories together.",
			Args: []ArgSpec{
				{Name: "branch", Type: ArgTypeString, Required: true, Description: "The name of the branch to merge into the current branch."},
			},
			ReturnType: ArgTypeString, // Success message or error/conflict output
		},
		Func: toolGitMerge, // from tools_git.go
	},
	// --- Git.Pull ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Pull",
			Description: "Fetch from and integrate with another repository or a local branch.",
			Args:        []ArgSpec{},
			ReturnType:  ArgTypeString, // Success message or error/conflict output
		},
		Func: toolGitPull, // from tools_git.go
	},
	// --- Git.Push ---
	ToolImplementation{
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
		Func: toolGitPush, // from tools_git.go
	},
	// --- Git.Diff ---
	ToolImplementation{
		Spec: ToolSpec{
			Name:        "Git.Diff",
			Description: "Shows changes between commits, commit and working tree, etc. Returns the diff output or a message indicating no changes.",
			Args: []ArgSpec{
				{Name: "cached", Type: ArgTypeBool, Required: false, Description: "Show diff of staged changes against HEAD (`--cached`). Default: false."},
				{Name: "commit1", Type: ArgTypeString, Required: false, Description: "First commit/branch/tree reference. Default: Index."},
				{Name: "commit2", Type: ArgTypeString, Required: false, Description: "Second commit/branch/tree reference. Default: Working tree."},
				{Name: "path", Type: ArgTypeString, Required: false, Description: "Limit the diff to the specified file or directory path."},
			},
			ReturnType: ArgTypeString, // The diff output or "Git.Diff: No changes detected."
		},
		Func: toolGitDiff, // from tools_git.go
	},
}
