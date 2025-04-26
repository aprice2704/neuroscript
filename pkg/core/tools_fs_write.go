// filename: pkg/core/tools_fs_write.go
package core

import (
	"fmt"
	"os"
	"path/filepath"
)

// toolWriteFile writes content to a specified file.
// *** MODIFIED: Uses interpreter.sandboxDir instead of os.Getwd() ***
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] and args[1] are strings
	filePath := args[0].(string)
	content := args[1].(string)

	// *** Get sandbox root directly from the interpreter ***
	sandboxRoot := interpreter.sandboxDir // Use the field name you added
	if sandboxRoot == "" {
		// Fallback or error if sandboxRoot is somehow empty
		if interpreter.logger != nil {
			interpreter.logger.Warn("TOOL WriteFile] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Ensure it's at least relative to CWD if empty
	}

	// Use SecureFilePath to validate the relative path is within the interpreter's sandboxDir
	// and get the secure absolute path.
	absPath, secErr := SecureFilePath(filePath, sandboxRoot) // *** Use sandboxRoot ***
	if secErr != nil {
		// Path validation failed (absolute, outside sandboxDir, etc.)
		errMsg := fmt.Sprintf("WriteFile path error for '%s': %s", filePath, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: WriteFile] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		// Return the error message string for NeuroScript, but the actual Go error for context.
		return errMsg, secErr
	}

	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: WriteFile] Writing to validated path: %s (Original Relative: %s, Sandbox: %s)", absPath, filePath, sandboxRoot)
	}

	// Ensure directory exists before writing (using the validated absolute path)
	dirPath := filepath.Dir(absPath)
	if dirErr := os.MkdirAll(dirPath, 0755); dirErr != nil {
		errMsg := fmt.Sprintf("WriteFile mkdir failed for dir '%s': %s", dirPath, dirErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: WriteFile] %s", errMsg)
		}
		return errMsg, fmt.Errorf("%w: creating directory '%s': %w", ErrInternalTool, dirPath, dirErr)
	}

	// Write the file (using the validated absolute path)
	writeErr := os.WriteFile(absPath, []byte(content), 0644)
	if writeErr != nil {
		errMsg := fmt.Sprintf("WriteFile failed for '%s': %s", filePath, writeErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: WriteFile] %s", errMsg)
		}
		return errMsg, fmt.Errorf("%w: writing file '%s': %w", ErrInternalTool, filePath, writeErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: WriteFile] Wrote %d bytes successfully to %s", len(content), filePath)
	}
	// Return "OK" on success
	return "OK", nil
}
