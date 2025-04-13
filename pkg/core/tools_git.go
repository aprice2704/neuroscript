// pkg/core/tools_git.go
package core

import (
	"fmt"
	// "os" // No longer need os here
	// "path/filepath" // Not directly needed here, but SecureFilePath uses it
)

// registerGitTools adds Git-related tools to the registry.
func registerGitTools(registry *ToolRegistry) error {
	tools := []ToolImplementation{
		{
			Spec: ToolSpec{
				Name:        "GitAdd",
				Description: "Stages a file for the next Git commit.",
				Args: []ArgSpec{
					{Name: "filepath", Type: ArgTypeString, Required: true, Description: "Relative path to the file to stage (within the sandbox)."},
				},
				ReturnType: ArgTypeString, // Returns "OK" or error message
			},
			Func: toolGitAdd,
		},
		{
			Spec: ToolSpec{
				Name:        "GitCommit",
				Description: "Commits currently staged changes.",
				Args: []ArgSpec{
					{Name: "message", Type: ArgTypeString, Required: true, Description: "The commit message."},
				},
				ReturnType: ArgTypeString, // Returns "OK" or error message
			},
			Func: toolGitCommit,
		},
	}
	for _, tool := range tools {
		if err := registry.RegisterTool(tool); err != nil {
			return fmt.Errorf("failed to register Git tool %s: %w", tool.Spec.Name, err)
		}
	}
	return nil
}

// toolGitAdd stages a file using the git command.
// Assumes git executable is in PATH. Validates path using SecureFilePath.
// *** MODIFIED: Use interpreter.sandboxDir instead of os.Getwd() ***
func toolGitAdd(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs before calling this.
	filePathRel := args[0].(string)

	// *** Get sandbox root directly from the interpreter ***
	sandboxRoot := interpreter.sandboxDir // Use the field name you added
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL GitAdd] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Ensure it's at least relative to CWD if empty
	}

	// Use SecureFilePath to ensure path is safe relative to sandboxDir
	// Note: git add typically wants relative paths from the repo root (which should be sandboxDir).
	// SecureFilePath returns the absolute path, which *usually* works with git add,
	// but let's use the validated *relative* path for git add.
	_, secErr := SecureFilePath(filePathRel, sandboxRoot) // Validate using sandboxRoot
	if secErr != nil {
		// Path validation failed
		errMsg := fmt.Sprintf("GitAdd path error: %s", secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL GitAdd] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		// Return the error message string for NeuroScript, and the actual Go error.
		return errMsg, secErr
	}

	// Use the original (validated) relative path for the git command
	gitPathArg := filePathRel

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL GitAdd] Staging validated relative path: %s (Sandbox: %s)", gitPathArg, sandboxRoot)
	}

	// Call the helper from tools_helpers.go
	// Assumes runGitCommand executes relative to the process CWD.
	// If tests pass with this, it means the interpreter's CWD isn't changed by tests anymore.
	err := runGitCommand("add", gitPathArg)
	if err != nil {
		// Return git error message as the result string, wrap for Go
		errMsg := fmt.Sprintf("GitAdd command failed: %s", err.Error())
		return errMsg, fmt.Errorf("%w: running git add '%s': %w", ErrInternalTool, gitPathArg, err)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL GitAdd] GitAdd successful for %s", gitPathArg)
	}
	return "OK", nil
}

// toolGitCommit commits staged changes using the git command.
// (Implementation unchanged, doesn't use paths)
func toolGitCommit(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation handled by ValidateAndConvertArgs
	message := args[0].(string)

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      Calling TOOL.GitCommit with message: %q", message)
	}

	// Call the helper from tools_helpers.go
	err := runGitCommand("commit", "-m", message)
	if err != nil {
		// Return git error message as the result string, wrap for Go
		errMsg := fmt.Sprintf("GitCommit command failed: %s", err.Error())
		return errMsg, fmt.Errorf("%w: running git commit: %w", ErrInternalTool, err)

	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[DEBUG-INTERP]      GitCommit successful.")
	}
	return "OK", nil
}
