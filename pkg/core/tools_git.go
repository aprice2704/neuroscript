// filename: pkg/core/tools_git.go
// UPDATED: Use internal toolExec helper instead of toolExecuteCommand
package core

import (
	"errors"
	"fmt"
	"strings"
)

// --- toolGitAdd implementation ---
// Assumes ValidateAndConvertArgs handles conversion to []interface{} containing strings
func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	pathsRaw := args[0].([]interface{}) // args[0] itself is the list of paths
	paths := make([]string, 0, len(pathsRaw))
	validatedPaths := make([]string, 0, len(pathsRaw))

	for _, pathRaw := range pathsRaw {
		pathStr, ok := pathRaw.(string)
		if !ok {
			return nil, fmt.Errorf("internal error: expected string path arg, got %T", pathRaw)
		}

		_, secErr := SecureFilePath(pathStr, interpreter.sandboxDir)
		if secErr != nil {
			errMsg := fmt.Sprintf("GitAdd path error for '%s': %s", pathStr, secErr.Error())
			return errMsg, errors.Join(ErrValidationArgValue, secErr)
		}
		validatedPaths = append(validatedPaths, pathStr) // Add validated path
		paths = append(paths, pathStr)                   // Collect relative paths for command
	}

	if len(validatedPaths) == 0 {
		return "GitAdd: No valid file paths provided.", nil
	}

	// --- FIX: Call toolExec correctly ---
	// Command is "git", arguments are "add", path1, path2, ...
	cmdArgs := append([]string{"add"}, paths...)
	output, err := toolExec(interpreter, append([]string{"git"}, cmdArgs...)...) // Pass "git" and then the cmdArgs elements
	// --- END FIX ---

	if err != nil {
		// toolExec includes output in error, so just wrap
		return nil, fmt.Errorf("GitAdd failed: %w", err)
	}

	return fmt.Sprintf("GitAdd successful for paths: %v.\nOutput:\n%s", validatedPaths, output), nil
}

// --- toolGitCommit implementation ---
func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	message := args[0].(string)
	if message == "" {
		return nil, fmt.Errorf("commit message cannot be empty: %w", ErrValidationArgValue)
	}

	// --- FIX: Call toolExec correctly ---
	output, err := toolExec(interpreter, "git", "commit", "-m", message)
	// --- END FIX ---

	if err != nil {
		return nil, fmt.Errorf("GitCommit failed: %w", err)
	}

	return fmt.Sprintf("GitCommit successful. Message: %q.\nOutput:\n%s", message, output), nil
}

// --- GitNewBranch Tool Implementation ---
func toolGitNewBranch(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	branchName := args[0].(string) // Assumes validation already happened

	if branchName == "" {
		return nil, fmt.Errorf("branch name cannot be empty: %w", ErrValidationArgValue)
	}
	if strings.ContainsAny(branchName, " \t\n\\/:*?\"<>|~^") {
		return nil, fmt.Errorf("branch name '%s' contains invalid characters: %w", branchName, ErrValidationArgValue)
	}

	interpreter.logger.Printf("[TOOL GitNewBranch] Executing: git checkout -b %s", branchName)
	// --- FIX: Call toolExec correctly ---
	output, err := toolExec(interpreter, "git", "checkout", "-b", branchName)
	// --- END FIX ---

	if err != nil {
		return nil, fmt.Errorf("failed to create new branch '%s': %w", branchName, err)
	}

	interpreter.logger.Printf("[TOOL GitNewBranch] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully created and checked out new branch '%s'.\nOutput:\n%s", branchName, output), nil
}

// --- GitCheckout Tool Implementation ---
func toolGitCheckout(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	branchName := args[0].(string) // Assumes validation already happened

	if branchName == "" {
		return nil, fmt.Errorf("branch name cannot be empty: %w", ErrValidationArgValue)
	}

	interpreter.logger.Printf("[TOOL GitCheckout] Executing: git checkout %s", branchName)
	// --- FIX: Call toolExec correctly ---
	output, err := toolExec(interpreter, "git", "checkout", branchName)
	// --- END FIX ---

	if err != nil {
		return nil, fmt.Errorf("failed to checkout branch/ref '%s': %w", branchName, err)
	}
	interpreter.logger.Printf("[TOOL GitCheckout] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully checked out branch/ref '%s'.\nOutput:\n%s", branchName, output), nil
}

// --- GitRm Tool Implementation ---
func toolGitRm(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	path := args[0].(string) // Assumes validation already happened

	securePath, err := SecureFilePath(path, interpreter.sandboxDir)
	if err != nil {
		return nil, fmt.Errorf("invalid path for GitRm '%s': %w", path, errors.Join(ErrValidationArgValue, err))
	}
	relativePath := path

	interpreter.logger.Printf("[TOOL GitRm] Executing: git rm %s (validated path: %s)", relativePath, securePath)
	// --- FIX: Call toolExec correctly ---
	output, err := toolExec(interpreter, "git", "rm", relativePath)
	// --- END FIX ---

	if err != nil {
		return nil, fmt.Errorf("failed to remove path '%s': %w", relativePath, err)
	}
	interpreter.logger.Printf("[TOOL GitRm] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully removed path '%s' from git index.\nOutput:\n%s", relativePath, output), nil
}

// --- GitMerge Tool Implementation ---
func toolGitMerge(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	branchName := args[0].(string) // Assumes validation already happened

	if branchName == "" {
		return nil, fmt.Errorf("branch name cannot be empty: %w", ErrValidationArgValue)
	}

	interpreter.logger.Printf("[TOOL GitMerge] Executing: git merge %s", branchName)
	// --- FIX: Call toolExec correctly ---
	output, err := toolExec(interpreter, "git", "merge", branchName)
	// --- END FIX ---

	// Merge conflicts will likely result in an error from toolExec
	// The error message from toolExec now includes the output.
	if err != nil {
		return nil, fmt.Errorf("failed to merge branch '%s' (check for conflicts): %w", branchName, err)
	}

	interpreter.logger.Printf("[TOOL GitMerge] Success. Output:\n%s", output)
	return fmt.Sprintf("Successfully merged branch '%s'.\nOutput:\n%s", branchName, output), nil
}

// --- Registration ---
// (Registration logic remains the same)
func registerGitTools(registry *ToolRegistry) error {
	var err error

	// Register GitAdd (Existing)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitAdd",
			Description: "Stages changes for commit using 'git add'. Accepts one or more file paths relative to the sandbox root.",
			Args:        []ArgSpec{{Name: "paths", Type: ArgTypeSliceAny, Required: true, Description: "A list of relative file paths to stage."}},
			ReturnType:  ArgTypeString, // Returns success message or error output
		},
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

	// Register GitNewBranch (New)
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

	// Register GitCheckout (New)
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

	// Register GitRm (New)
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

	// Register GitMerge (New)
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

	// Register other Git tools here...

	return nil
}
