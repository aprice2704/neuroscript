// filename: pkg/core/tools_fs_read.go
package core

import (
	"fmt"
	"os"
	// No longer need path/filepath directly
)

// toolReadFile reads the content of a specified file.
// *** MODIFIED: Uses interpreter.sandboxDir instead of os.Getwd() ***
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	filePathRel := args[0].(string)
	// *** Get sandbox root directly from the interpreter ***
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		// Fallback or error if sandboxRoot is somehow empty
		// This shouldn't happen if NewInterpreter sets a default (".")
		// Let SecureFilePath handle "." as CWD if it happens.
		// Or return an internal error? Let's rely on SecureFilePath handling "."
		if interpreter.logger != nil {
			interpreter.logger.Printf("[WARN TOOL ReadFile] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Ensure it's at least relative to CWD if empty
	}

	// Use SecureFilePath to validate the relative path is within the interpreter's sandboxDir
	// and get the secure absolute path.
	absPath, secErr := SecureFilePath(filePathRel, sandboxRoot) // *** Use sandboxRoot ***
	if secErr != nil {
		// Path validation failed (absolute, outside sandboxDir, etc.)
		errMsg := fmt.Sprintf("ReadFile path error for '%s': %s", filePathRel, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ReadFile] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		// Return the error message string for NeuroScript, but the actual Go error for context.
		return errMsg, secErr
	}

	// Path is validated and within the sandbox, attempt to read the file using the absolute path
	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ReadFile] Attempting to read validated path: %s (Original Relative: %s, Sandbox: %s)", absPath, filePathRel, sandboxRoot)
	}
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		errMsg := ""
		if os.IsNotExist(readErr) {
			errMsg = fmt.Sprintf("ReadFile failed: File not found at path '%s'", filePathRel)
		} else {
			// Consider making this more specific for directories if needed
			errMsg = fmt.Sprintf("ReadFile failed for '%s': %s", filePathRel, readErr.Error())
		}
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ReadFile] %s", errMsg)
		}
		// Return error message string for script, but wrap the actual os error for Go context.
		return errMsg, fmt.Errorf("%w: reading file '%s': %w", ErrInternalTool, filePathRel, readErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ReadFile] Read %d bytes successfully from %s", len(contentBytes), filePathRel)
	}

	// Return file content as string
	return string(contentBytes), nil
}
