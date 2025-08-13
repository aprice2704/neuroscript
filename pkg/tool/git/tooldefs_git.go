// NeuroScript Version: 0.5.4
// File version: 12
// Purpose: Added 'usesExternal:git' effect to all tools to explicitly mark their reliance on the external Git executable.
// filename: pkg/tool/git/tooldefs_git.go
// nlines: 280 // Approximate
// risk_rating: HIGH

package git

import (
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

const group = "git"

var gitToolsToRegister = []tool.ToolImplementation{
	{
		Spec: tool.ToolSpec{
			Name:        "Status",
			Group:       group,
			Description: "Gets the status of the Git repository.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "repo_path", Type: tool.ArgTypeString, Required: false, Description: "Optional. Relative path to the repository. Defaults to the sandbox root."},
			},
			ReturnType: tool.ArgTypeMap,
		},
		Func:          toolGitStatus,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"readsFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Add",
			Group:       group,
			Description: "Adds file contents to the index.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "paths", Type: tool.ArgTypeSliceAny, Required: true},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitAdd,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Reset",
			Group:       group,
			Description: "Resets the current HEAD to the specified state.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "mode", Type: tool.ArgTypeString, Required: false},
				{Name: "commit", Type: tool.ArgTypeString, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitReset,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Clone",
			Group:       group,
			Description: "Clones a Git repository.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "repository_url", Type: tool.ArgTypeString, Required: true},
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitClone,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"write"}},
			{Resource: "net", Verbs: []string{"read"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "readsNet", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Pull",
			Group:       group,
			Description: "Pulls the latest changes from the remote repository.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "remote_name", Type: tool.ArgTypeString, Required: false},
				{Name: "branch_name", Type: tool.ArgTypeString, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitPull,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "net", Verbs: []string{"read"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "readsNet", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Commit",
			Group:       group,
			Description: "Commits staged changes.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "commit_message", Type: tool.ArgTypeString, Required: true},
				{Name: "allow_empty", Type: tool.ArgTypeBool, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitCommit,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Push",
			Group:       group,
			Description: "Pushes committed changes to a remote repository.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "remote_name", Type: tool.ArgTypeString, Required: false},
				{Name: "branch_name", Type: tool.ArgTypeString, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitPush,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
			{Resource: "net", Verbs: []string{"write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"readsFS", "writesNet", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Branch",
			Group:       group,
			Description: "Manages branches.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "name", Type: tool.ArgTypeString, Required: false},
				{Name: "checkout", Type: tool.ArgTypeBool, Required: false},
				{Name: "list_remote", Type: tool.ArgTypeBool, Required: false},
				{Name: "list_all", Type: tool.ArgTypeBool, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitBranch,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "readsFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Checkout",
			Group:       group,
			Description: "Switches branches or restores working tree files.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "branch", Type: tool.ArgTypeString, Required: true},
				{Name: "create", Type: tool.ArgTypeBool, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitCheckout,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Rm",
			Group:       group,
			Description: "Removes files from the working tree and from the index.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "paths", Type: tool.ArgTypeAny, Required: true},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitRm,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write", "delete"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Merge",
			Group:       group,
			Description: "Joins two or more development histories together.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "branch", Type: tool.ArgTypeString, Required: true},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitMerge,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read", "write"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"writesFS", "usesExternal:git"},
	},
	{
		Spec: tool.ToolSpec{
			Name:        "Diff",
			Group:       group,
			Description: "Shows changes between commits, commit and working tree, etc.",
			Category:    "Git",
			Args: []tool.ArgSpec{
				{Name: "relative_path", Type: tool.ArgTypeString, Required: true},
				{Name: "cached", Type: tool.ArgTypeBool, Required: false},
				{Name: "commit1", Type: tool.ArgTypeString, Required: false},
				{Name: "commit2", Type: tool.ArgTypeString, Required: false},
				{Name: "path", Type: tool.ArgTypeString, Required: false},
			},
			ReturnType: tool.ArgTypeString,
		},
		Func:          toolGitDiff,
		RequiresTrust: true,
		RequiredCaps: []capability.Capability{
			{Resource: "fs", Verbs: []string{"read"}},
			{Resource: "shell", Verbs: []string{"execute"}, Scopes: []string{"git"}},
		},
		Effects: []string{"readsFS", "usesExternal:git"},
	},
}
