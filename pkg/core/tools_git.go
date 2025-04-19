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

// --- GitPull Tool Implementation ---
func toolGitPull(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// GitPull takes no arguments, validation ensures len(args) == 0

	interpreter.logger.Printf("[TOOL GitPull] Executing: git pull")
	output, err := toolExec(interpreter, "git", "pull")

	if err != nil {
		// toolExec includes stderr in the error message
		return nil, fmt.Errorf("GitPull failed: %w", err)
	}

	interpreter.logger.Printf("[TOOL GitPull] Success. Output:\n%s", output)
	return fmt.Sprintf("GitPull successful.\nOutput:\n%s", output), nil
}

// --- GitPush Tool Implementation (NEW) ---
func toolGitPush(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// GitPush takes no arguments currently (pushes the current branch to its upstream)
	// Validation ensures len(args) == 0

	interpreter.logger.Printf("[TOOL GitPush] Executing: git push")
	output, err := toolExec(interpreter, "git", "push")

	if err != nil {
		// toolExec includes stderr in the error message
		// Common errors include: rejected push (needs pull), no upstream configured, authentication failure
		return nil, fmt.Errorf("GitPush failed: %w", err)
	}

	interpreter.logger.Printf("[TOOL GitPush] Success. Output:\n%s", output)
	return fmt.Sprintf("GitPush successful.\nOutput:\n%s", output), nil
}

// --- GitDiff Tool Implementation (NEW - Basic) ---
func toolGitDiff(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// GitDiff takes no arguments currently (shows working tree changes vs index)
	// Validation ensures len(args) == 0

	interpreter.logger.Printf("[TOOL GitDiff] Executing: git diff")
	// Note: `git diff` returns a non-zero exit code (which toolExec treats as error)
	// only for fatal errors, not when there are differences found.
	// Differences are simply printed to stdout.
	output, err := toolExec(interpreter, "git", "diff")

	if err != nil {
		// This would likely be an error like 'not a git repository'
		return nil, fmt.Errorf("GitDiff command failed: %w", err)
	}

	interpreter.logger.Printf("[TOOL GitDiff] Success. Output:\n%s", output)
	// If there are no differences, output will be empty.
	if output == "" {
		return "GitDiff: No changes detected in the working tree.", nil
	}
	// Return the diff output directly
	return output, nil
}
