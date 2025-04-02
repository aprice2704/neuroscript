// pkg/core/tools_git.go
package core

import (
	"fmt"
	"os"
	// "path/filepath" // Not directly needed here, but secureFilePath uses it
)

// registerGitTools adds Git-related tools to the registry.
func registerGitTools(registry *ToolRegistry) {
	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitAdd",
			Description: "Stages a file for the next Git commit.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file to stage."},
			},
			ReturnType: ArgTypeString, // Returns "OK" or error message
		},
		Func: toolGitAdd,
	})

	registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GitCommit",
			Description: "Commits currently staged changes.",
			Args: []ArgSpec{
				{Name: "message", Type: ArgTypeString, Required: true, Description: "The commit message."},
			},
			ReturnType: ArgTypeString, // Returns "OK" or error message
		},
		Func: toolGitCommit,
	})
}

// toolGitAdd stages a file using the git command.
// Assumes git executable is in PATH and operates in the current working directory.
// Uses secureFilePath for safety.
func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation should be handled by ValidateAndConvertArgs before calling this.
	filePath := args[0].(string)

	// Use secureFilePath to ensure path is safe relative to CWD
	cwd, errWd := os.Getwd()
	if errWd != nil {
		// This is an internal error, not a user path error
		return nil, fmt.Errorf("GitAdd failed to get working directory: %w", errWd)
	}
	absPath, secErr := secureFilePath(filePath, cwd)
	if secErr != nil {
		// Return the security error message as the result string for NeuroScript
		return fmt.Sprintf("GitAdd path error: %s", secErr.Error()), nil
	}

	// Relative path calculation might be needed if git add behaves differently,
	// but usually adding the absolute path within the repo works. Test this assumption.
	// For now, using the secured absolute path.

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GitAdd for %s (Resolved: %s)", filePath, absPath)
	}

	// Call the helper from tools_helpers.go
	err := runGitCommand("add", absPath)
	if err != nil {
		// Return git error message as the result string
		return fmt.Sprintf("GitAdd command failed: %s", err.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      GitAdd successful for %s", filePath)
	}
	return "OK", nil
}

// toolGitCommit commits staged changes using the git command.
// Assumes git executable is in PATH and operates in the current working directory.
func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	message := args[0].(string)

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GitCommit with message: %q", message)
	}

	// Call the helper from tools_helpers.go
	err := runGitCommand("commit", "-m", message)
	if err != nil {
		// Return git error message as the result string
		return fmt.Sprintf("GitCommit command failed: %s", err.Error()), nil
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      GitCommit successful.")
	}
	return "OK", nil
}
