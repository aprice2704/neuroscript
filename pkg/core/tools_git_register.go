package core

import "fmt"

// --- Registration ---
// This function registers all Git-related tools, including GitStatus implemented elsewhere.
func registerGitTools(registry *ToolRegistry) error {
	var err error

	// Register GitAdd (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitAdd",
			Description: "Stages changes for commit using 'git add'. Accepts one or more file paths relative to the sandbox root.",
			// *** FIX: Use ArgTypeSliceString based on provided tools_types.go ***
			Args:       []ArgSpec{{Name: "paths", Type: ArgTypeSliceString, Required: true, Description: "A list of relative file paths to stage."}},
			ReturnType: ArgTypeString, // Returns success message or error output
		},
		// *** Note: toolGitAdd needs adjustment if validation is stricter now ***
		Func: toolGitAdd,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitAdd: %w", err)
	}

	// Register GitCommit (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitCommit",
			Description: "Commits staged changes using 'git commit -m'.",
			Args:        []ArgSpec{{Name: "message", Type: ArgTypeString, Required: true}},
			ReturnType:  ArgTypeString, // Returns success message or error output
		},
		Func: toolGitCommit,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitCommit: %w", err)
	}

	// Register GitNewBranch (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitNewBranch",
			Description: "Creates and checks out a new branch using 'git checkout -b'.",
			Args:        []ArgSpec{{Name: "branch_name", Type: ArgTypeString, Required: true}},
			ReturnType:  ArgTypeString, // Returns success message or error output
		},
		Func: toolGitNewBranch,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitNewBranch: %w", err)
	}

	// Register GitCheckout (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitCheckout",
			Description: "Checks out an existing branch or commit using 'git checkout'.",
			Args:        []ArgSpec{{Name: "branch_name", Type: ArgTypeString, Required: true}}, // Can be branch, tag, commit hash etc.
			ReturnType:  ArgTypeString,                                                         // Returns success message or error output
		},
		Func: toolGitCheckout,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitCheckout: %w", err)
	}

	// Register GitRm (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitRm",
			Description: "Removes a file from the Git index using 'git rm'. Path must be relative to project root.",
			Args:        []ArgSpec{{Name: "path", Type: ArgTypeString, Required: true}},
			ReturnType:  ArgTypeString, // Returns success message or error output
		},
		Func: toolGitRm,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitRm: %w", err)
	}

	// Register GitMerge (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitMerge",
			Description: "Merges the specified branch into the current branch using 'git merge'. Handles potential conflicts by returning error output.",
			Args:        []ArgSpec{{Name: "branch_name", Type: ArgTypeString, Required: true}},
			ReturnType:  ArgTypeString, // Returns success message or error/conflict output
		},
		Func: toolGitMerge,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitMerge: %w", err)
	}

	// --- Register GitStatus (Implementation is in tools_git_status.go) ---
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "GitStatus",
			Description: "Gets the current Git repository status using 'git status --porcelain -b --untracked-files=all' and returns a structured map. " +
				"Keys: 'branch' (string|null), 'remote_branch' (string|null), 'ahead' (int), 'behind' (int), " +
				"'files' (list[map{'path','index_status','worktree_status','original_path'}]), 'untracked_files_present' (bool), 'is_clean' (bool), 'error' (string|null).",
			Args: []ArgSpec{}, // No arguments
			// *** FIX: Use ArgTypeAny as ArgTypeMap is not defined ***
			ReturnType: ArgTypeAny,
		},
		Func: toolGitStatus, // Reference the function from tools_git_status.go
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GitStatus: %w", err)
	}
	// --- END Register GitStatus ---

	// Register other Git tools here...

	return nil
}
